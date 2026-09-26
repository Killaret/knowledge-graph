package queue

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/infrastructure/mongo"
	"knowledge-graph/internal/infrastructure/nlp"
)

// contentGuardRepo wraps the worker mock and fails if any write path on
// the note repository is touched — NLP-4 criterion 1: notes.content is
// byte-identical after a normalization pass, the pipeline never writes back.
type contentGuardRepo struct {
	*mockNoteRepoForWorker
	writes int
}

func (r *contentGuardRepo) Save(ctx context.Context, n *note.Note) error {
	r.writes++
	return r.mockNoteRepoForWorker.Save(ctx, n)
}

func (r *contentGuardRepo) Delete(ctx context.Context, id uuid.UUID) error {
	r.writes++
	return r.mockNoteRepoForWorker.Delete(ctx, id)
}

func (r *contentGuardRepo) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
	r.writes++
	return r.mockNoteRepoForWorker.DeleteBatch(ctx, ids)
}

// fakeNlpArtifactsStore records calls; simulates the Mongo repository.
type fakeNlpArtifactsStore struct {
	currentHash      string
	found            bool
	saved            []*mongo.NlpArtifact
	savedWithHistory []bool
	deleted          []uuid.UUID
}

func (f *fakeNlpArtifactsStore) FindCurrentSourceHash(ctx context.Context, noteID uuid.UUID, pipelineVersion string) (string, bool, error) {
	return f.currentHash, f.found, nil
}

func (f *fakeNlpArtifactsStore) SaveCurrent(ctx context.Context, doc *mongo.NlpArtifact, historyEnabled bool) error {
	f.saved = append(f.saved, doc)
	f.savedWithHistory = append(f.savedWithHistory, historyEnabled)
	return nil
}

func (f *fakeNlpArtifactsStore) DeleteByNoteID(ctx context.Context, noteID uuid.UUID) (int64, error) {
	f.deleted = append(f.deleted, noteID)
	return int64(len(f.deleted)), nil
}

func newNlp4Note(t *testing.T) *note.Note {
	title, err := note.NewTitle("Note title")
	require.NoError(t, err)
	content, err := note.NewContent("some meaningful content for normalization")
	require.NoError(t, err)
	metadata, err := note.NewMetadata(nil)
	require.NoError(t, err)
	return note.NewNote(title, content, note.MustType("star"), metadata)
}

func newNormalizeServer(t *testing.T, normalizeCalls *int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/normalize", r.URL.Path)
		*normalizeCalls++
		var req struct {
			Text  string `json:"text"`
			Title string `json:"title"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		w.Header().Set("Content-Type", "application/json")
		resp := `{
			"normalized_text": "cleaned text",
			"chunks": [{"idx":0,"text":"cleaned text","heading_path":[],"char_span":[0,12],"token_count":2,"kind":"prose","forced_split":false}],
			"metrics": {"raw_tokens":10,"norm_tokens":2,"compression":0.2,"iterations":1,"stop_reason":"single_pass","emb_cosine":0.95},
			"rolled_back": false,
			"rollback_reason": null,
			"skipped": false,
			"pipeline_version": "norm-v1"
		}`
		_, err := w.Write([]byte(resp))
		require.NoError(t, err)
	}))
}

func TestWorker_HandleNormalizeNote_SavesArtifact(t *testing.T) {
	repo := new(mockNoteRepoForWorker)
	n := newNlp4Note(t)
	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)

	normalizeCalls := 0
	server := newNormalizeServer(t, &normalizeCalls)
	defer server.Close()

	store := &fakeNlpArtifactsStore{}
	nlpClient := nlp.NewNLPClient(server.URL, nil, 0)
	w := NewWorker(repo, nil, nil, nlpClient, nil, nil, nil, nil, nil, 0, store, true, "test-model")

	payload, _ := json.Marshal(NormalizeNotePayload{NoteID: n.ID().String()})
	require.NoError(t, w.HandleNormalizeNote(context.Background(), asynq.NewTask(TypeNormalizeNote, payload)))

	require.Equal(t, 1, normalizeCalls)
	require.Len(t, store.saved, 1)
	doc := store.saved[0]
	assert.Equal(t, n.ID(), doc.NoteID)
	assert.Equal(t, NlpPipelineVersion, doc.PipelineVersion)
	assert.Equal(t, "test-model", doc.ModelVersion)
	assert.Equal(t, "cleaned text", doc.NormalizedText)
	assert.Equal(t, nlpArtifactsSourceHash(n.Title().String(), n.Content().String()), doc.SourceHash)
	require.Len(t, doc.Chunks, 1)
	assert.Equal(t, "cleaned text", doc.Chunks[0].Text)
	assert.Equal(t, "single_pass", doc.Metrics.StopReason)
	require.NotNil(t, doc.Metrics.EmbCosine)
	assert.InDelta(t, 0.95, *doc.Metrics.EmbCosine, 1e-6)
	assert.True(t, store.savedWithHistory[0])
}

// NLP-4 criterion 1: a normalization pass must not write to notes at all —
// notes.content stays byte-identical because the worker never persists it.
func TestWorker_HandleNormalizeNote_NeverWritesNoteContent(t *testing.T) {
	base := new(mockNoteRepoForWorker)
	n := newNlp4Note(t)
	base.On("FindByID", mock.Anything, n.ID()).Return(n, nil)
	repo := &contentGuardRepo{mockNoteRepoForWorker: base}

	normalizeCalls := 0
	server := newNormalizeServer(t, &normalizeCalls)
	defer server.Close()

	store := &fakeNlpArtifactsStore{}
	nlpClient := nlp.NewNLPClient(server.URL, nil, 0)
	w := NewWorker(repo, nil, nil, nlpClient, nil, nil, nil, nil, nil, 0, store, true, "test-model")

	payload, _ := json.Marshal(NormalizeNotePayload{NoteID: n.ID().String()})
	require.NoError(t, w.HandleNormalizeNote(context.Background(), asynq.NewTask(TypeNormalizeNote, payload)))

	assert.Zero(t, repo.writes, "normalization pipeline must not write notes.content")
	assert.Equal(t, "some meaningful content for normalization", n.Content().String())
}

// nlp.history.enabled=false must reach the store — the repository decides
// supersede-vs-delete from this flag, so dropping it silently accrues
// superseded documents (mutation guard for criterion 7).
func TestWorker_HandleNormalizeNote_PropagatesHistoryFlag(t *testing.T) {
	repo := new(mockNoteRepoForWorker)
	n := newNlp4Note(t)
	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)

	normalizeCalls := 0
	server := newNormalizeServer(t, &normalizeCalls)
	defer server.Close()

	store := &fakeNlpArtifactsStore{}
	nlpClient := nlp.NewNLPClient(server.URL, nil, 0)
	w := NewWorker(repo, nil, nil, nlpClient, nil, nil, nil, nil, nil, 0, store, false, "test-model")

	payload, _ := json.Marshal(NormalizeNotePayload{NoteID: n.ID().String()})
	require.NoError(t, w.HandleNormalizeNote(context.Background(), asynq.NewTask(TypeNormalizeNote, payload)))

	require.Len(t, store.savedWithHistory, 1)
	assert.False(t, store.savedWithHistory[0], "historyEnabled=false must reach SaveCurrent")
}

func TestWorker_HandleNormalizeNote_SkipsUnchangedSource(t *testing.T) {
	repo := new(mockNoteRepoForWorker)
	n := newNlp4Note(t)
	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)

	normalizeCalls := 0
	server := newNormalizeServer(t, &normalizeCalls)
	defer server.Close()

	store := &fakeNlpArtifactsStore{
		currentHash: nlpArtifactsSourceHash(n.Title().String(), n.Content().String()),
		found:       true,
	}
	nlpClient := nlp.NewNLPClient(server.URL, nil, 0)
	w := NewWorker(repo, nil, nil, nlpClient, nil, nil, nil, nil, nil, 0, store, true, "test-model")

	payload, _ := json.Marshal(NormalizeNotePayload{NoteID: n.ID().String()})
	require.NoError(t, w.HandleNormalizeNote(context.Background(), asynq.NewTask(TypeNormalizeNote, payload)))

	assert.Equal(t, 0, normalizeCalls, "unchanged source_hash must not call /normalize")
	assert.Empty(t, store.saved)
}

func TestWorker_HandleNormalizeNote_DeletedNoteCleansArtifacts(t *testing.T) {
	repo := new(mockNoteRepoForWorker)
	noteID := uuid.New()
	repo.On("FindByID", mock.Anything, noteID).Return(nil, nil)

	normalizeCalls := 0
	server := newNormalizeServer(t, &normalizeCalls)
	defer server.Close()

	store := &fakeNlpArtifactsStore{}
	nlpClient := nlp.NewNLPClient(server.URL, nil, 0)
	w := NewWorker(repo, nil, nil, nlpClient, nil, nil, nil, nil, nil, 0, store, true, "test-model")

	payload, _ := json.Marshal(NormalizeNotePayload{NoteID: noteID.String()})
	require.NoError(t, w.HandleNormalizeNote(context.Background(), asynq.NewTask(TypeNormalizeNote, payload)))

	assert.Equal(t, 0, normalizeCalls)
	require.Equal(t, []uuid.UUID{noteID}, store.deleted)
}

func TestWorker_HandleNormalizeNote_NoStoreIsNoOp(t *testing.T) {
	normalizeCalls := 0
	server := newNormalizeServer(t, &normalizeCalls)
	defer server.Close()

	nlpClient := nlp.NewNLPClient(server.URL, nil, 0)
	w := NewWorker(nil, nil, nil, nlpClient, nil, nil, nil, nil, nil, 0, nil, true, "test-model")

	payload, _ := json.Marshal(NormalizeNotePayload{NoteID: uuid.New().String()})
	require.NoError(t, w.HandleNormalizeNote(context.Background(), asynq.NewTask(TypeNormalizeNote, payload)))
	assert.Equal(t, 0, normalizeCalls)
}

func TestWorker_HandleNormalizeNote_InvalidPayload(t *testing.T) {
	w := NewWorker(nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, nil, false, "")
	assert.Error(t, w.HandleNormalizeNote(context.Background(), asynq.NewTask(TypeNormalizeNote, []byte("not json"))))
	assert.Error(t, w.HandleNormalizeNote(context.Background(), asynq.NewTask(TypeNormalizeNote, []byte(`{"note_id":"bad"}`))))
}

func TestWorker_HandleNlpArtifactsCleanup(t *testing.T) {
	store := &fakeNlpArtifactsStore{}
	w := NewWorker(nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, store, false, "")

	noteID := uuid.New()
	payload, _ := json.Marshal(NlpArtifactsCleanupPayload{NoteID: noteID.String()})
	require.NoError(t, w.HandleNlpArtifactsCleanup(context.Background(), asynq.NewTask(TypeNlpArtifactsCleanup, payload)))
	assert.Equal(t, []uuid.UUID{noteID}, store.deleted)
}

func TestWorker_HandleNlpArtifactsCleanup_NoStore(t *testing.T) {
	w := NewWorker(nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, nil, false, "")
	payload, _ := json.Marshal(NlpArtifactsCleanupPayload{NoteID: uuid.New().String()})
	require.NoError(t, w.HandleNlpArtifactsCleanup(context.Background(), asynq.NewTask(TypeNlpArtifactsCleanup, payload)))
}

// Criterion 5: pipeline disabled → EnqueueNormalizeNote returns without
// touching Redis (client built against an unreachable address).
func TestAsynqClient_NormalizeGatedByFlag(t *testing.T) {
	client, err := NewAsynqClient("127.0.0.1:1", false, false, false)
	require.NoError(t, err)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	require.NoError(t, client.EnqueueNormalizeNote(ctx, uuid.New().String()))
}

// Pipeline enabled → the normalize task actually lands in the queue
// (miniredis + asynq Inspector read it back).
func TestAsynqClient_NormalizeEnqueuedWhenEnabled(t *testing.T) {
	mr := miniredis.RunT(t)
	client, err := NewAsynqClient(mr.Addr(), false, true, false)
	require.NoError(t, err)
	defer client.Close()

	noteID := uuid.New().String()
	require.NoError(t, client.EnqueueNormalizeNote(context.Background(), noteID))

	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: mr.Addr()})
	defer inspector.Close()
	tasks, err := inspector.ListPendingTasks("default")
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, TypeNormalizeNote, tasks[0].Type)
	var payload NormalizeNotePayload
	require.NoError(t, json.Unmarshal(tasks[0].Payload, &payload))
	assert.Equal(t, noteID, payload.NoteID)
}
