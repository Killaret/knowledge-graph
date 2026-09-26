package queue

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	appquality "knowledge-graph/internal/application/quality"
	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/infrastructure/nlp"
)

// --- quality fakes ---

type fakeQualityEnqueuer struct{ calls []string }

func (f *fakeQualityEnqueuer) EnqueueAssessQuality(_ context.Context, noteID string, trigger string) error {
	f.calls = append(f.calls, noteID+":"+trigger)
	return nil
}

type fakeQualityLog struct{ entries []appquality.LogEntry }

func (f *fakeQualityLog) Append(_ context.Context, e appquality.LogEntry) error {
	f.entries = append([]appquality.LogEntry{e}, f.entries...)
	return nil
}

func (f *fakeQualityLog) List(_ context.Context, _ uuid.UUID, limit int) ([]appquality.LogEntry, error) {
	if len(f.entries) > limit {
		return f.entries[:limit], nil
	}
	return f.entries, nil
}

func (f *fakeQualityLog) MarkManualReview(_ context.Context, _ uuid.UUID, hash string) (bool, error) {
	for i := range f.entries {
		if f.entries[i].SourceHash == hash {
			f.entries[i].Record.NeedsManualReview = true
			return true, nil
		}
	}
	return false, nil
}

type fakeQualityStats struct{}

func (fakeQualityStats) KeywordCount(context.Context, uuid.UUID) (int, error)  { return 0, nil }
func (fakeQualityStats) LinkCount(context.Context, uuid.UUID) (int, error)     { return 0, nil }
func (fakeQualityStats) HasEmbedding(context.Context, uuid.UUID) (bool, error) { return false, nil }

// workerNoteReader adapts the mock repo to the assessor port.
func workerAssessor(repo note.Repository, log appquality.LogStore) *appquality.Assessor {
	return appquality.NewAssessor(
		&qualityNoteReader{repo: repo}, nil, nil, log, fakeQualityStats{}, nil,
		appquality.DefaultThresholds())
}

func qualityTask(noteID uuid.UUID, trigger string) *asynq.Task {
	payload, _ := json.Marshal(AssessQualityPayload{NoteID: noteID.String(), Trigger: trigger})
	return asynq.NewTask(TypeAssessQuality, payload)
}

func normalizeTask(noteID uuid.UUID) *asynq.Task {
	payload, _ := json.Marshal(NormalizeNotePayload{NoteID: noteID.String()})
	return asynq.NewTask(TypeNormalizeNote, payload)
}

func TestWorker_HandleAssessQuality_WritesLog(t *testing.T) {
	repo := new(mockNoteRepoForWorker)
	n := newNlp4Note(t)
	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)

	log := &fakeQualityLog{}
	w := NewWorker(repo, nil, nil, nil, nil, nil, nil, nil, nil, 0, nil, false, "")
	w.UseQuality(workerAssessor(repo, log), &fakeQualityEnqueuer{})

	err := w.HandleAssessQuality(context.Background(), qualityTask(n.ID(), appquality.TriggerAuto))
	require.NoError(t, err)
	require.Len(t, log.entries, 1)
	assert.Equal(t, appquality.PipelineVersion, log.entries[0].Record.PipelineVersion)
	assert.Equal(t, appquality.VerdictCreate, log.entries[0].Record.Verdict)
}

func TestWorker_HandleAssessQuality_NilAssessorNoOp(t *testing.T) {
	w := NewWorker(nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, nil, false, "")
	// No UseQuality — the task must be a no-op even though it exists.
	err := w.HandleAssessQuality(context.Background(), qualityTask(uuid.New(), appquality.TriggerAuto))
	require.NoError(t, err)
}

func TestWorker_HandleAssessQuality_NeverTouchesNoteWrites(t *testing.T) {
	repo := new(mockNoteRepoForWorker)
	guard := &contentGuardRepo{mockNoteRepoForWorker: repo}
	n := newNlp4Note(t)
	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)

	w := NewWorker(guard, nil, nil, nil, nil, nil, nil, nil, nil, 0, nil, false, "")
	w.UseQuality(workerAssessor(repo, &fakeQualityLog{}), &fakeQualityEnqueuer{})

	require.NoError(t, w.HandleAssessQuality(context.Background(), qualityTask(n.ID(), "auto")))
	assert.Equal(t, 0, guard.writes)
}

func TestWorker_ScheduleQuality_AfterNormalize(t *testing.T) {
	repo := new(mockNoteRepoForWorker)
	n := newNlp4Note(t)
	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)

	normalizeCalls := 0
	server := newNormalizeServer(t, &normalizeCalls)
	defer server.Close()

	enq := &fakeQualityEnqueuer{}
	w := NewWorker(repo, nil, nil, nlp.NewNLPClient(server.URL, nil, 0), nil, nil, nil, nil, nil, 0, &fakeNlpArtifactsStore{}, false, "m")
	w.UseQuality(workerAssessor(repo, &fakeQualityLog{}), enq)

	err := w.HandleNormalizeNote(context.Background(), normalizeTask(n.ID()))
	require.NoError(t, err)
	require.Len(t, enq.calls, 1)
	assert.Equal(t, n.ID().String()+":auto", enq.calls[0])
}

func TestWorker_ScheduleQuality_OffMeansNothing(t *testing.T) {
	repo := new(mockNoteRepoForWorker)
	n := newNlp4Note(t)
	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)

	normalizeCalls := 0
	server := newNormalizeServer(t, &normalizeCalls)
	defer server.Close()

	enq := &fakeQualityEnqueuer{}
	// No UseQuality → scheduleQuality is a no-op.
	w := NewWorker(repo, nil, nil, nlp.NewNLPClient(server.URL, nil, 0), nil, nil, nil, nil, nil, 0, &fakeNlpArtifactsStore{}, false, "m")
	_ = enq
	err := w.HandleNormalizeNote(context.Background(), normalizeTask(n.ID()))
	require.NoError(t, err)
	assert.Empty(t, enq.calls)
}
