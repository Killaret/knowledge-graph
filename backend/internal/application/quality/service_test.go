package quality

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- fakes ---

type fakeNoteReader struct{ n *Note }

func (f *fakeNoteReader) FindByID(_ context.Context, _ uuid.UUID) (*Note, error) {
	return f.n, nil
}

type fakeLog struct{ entries []LogEntry }

func (f *fakeLog) Append(_ context.Context, e LogEntry) error {
	f.entries = append([]LogEntry{e}, f.entries...) // newest first
	return nil
}

func (f *fakeLog) List(_ context.Context, _ uuid.UUID, limit int) ([]LogEntry, error) {
	if len(f.entries) > limit {
		return f.entries[:limit], nil
	}
	return f.entries, nil
}

func (f *fakeLog) MarkManualReview(_ context.Context, _ uuid.UUID, hash string) (bool, error) {
	for i := range f.entries {
		if f.entries[i].SourceHash == hash {
			f.entries[i].Record.NeedsManualReview = true
			return true, nil
		}
	}
	return false, nil
}

type fakeWriter struct{ calls int }

func (f *fakeWriter) SetQuality(_ context.Context, _ uuid.UUID, _ string, _ Record) (bool, error) {
	f.calls++
	return true, nil
}

type fakeStats struct {
	keywords, links int
	embedding       bool
}

func (f *fakeStats) KeywordCount(context.Context, uuid.UUID) (int, error) { return f.keywords, nil }
func (f *fakeStats) LinkCount(context.Context, uuid.UUID) (int, error)    { return f.links, nil }
func (f *fakeStats) HasEmbedding(context.Context, uuid.UUID) (bool, error) {
	return f.embedding, nil
}

func newAssessor(n *Note, log *fakeLog, w *fakeWriter) *Assessor {
	return NewAssessor(&fakeNoteReader{n: n}, nil, w, log,
		&fakeStats{}, nil, DefaultThresholds())
}

func TestAssess_FirstRunStoresRecord(t *testing.T) {
	log, w := &fakeLog{}, &fakeWriter{}
	a := newAssessor(&Note{ID: uuid.New(), Title: "T", Content: "Нормальный текст из нескольких слов."}, log, w)

	rec, err := a.Assess(context.Background(), uuid.New(), TriggerAuto)
	require.NoError(t, err)
	require.NotNil(t, rec)
	assert.Equal(t, VerdictCreate, rec.Verdict)
	assert.Equal(t, 1, rec.Attempt)
	assert.Equal(t, PipelineVersion, rec.PipelineVersion)
	assert.Len(t, log.entries, 1)
	assert.Equal(t, 1, w.calls)
}

func TestAssess_StopRule_UnchangedSignals(t *testing.T) {
	log, w := &fakeLog{}, &fakeWriter{}
	n := &Note{ID: uuid.New(), Title: "T", Content: "текст"}
	a := newAssessor(n, log, w)

	noteID := uuid.New()
	rec, err := a.Assess(context.Background(), noteID, TriggerAuto)
	require.NoError(t, err)
	require.NotNil(t, rec)

	// Second auto pass with identical signals — suppressed.
	rec, err = a.Assess(context.Background(), noteID, TriggerAuto)
	require.NoError(t, err)
	assert.Nil(t, rec)
	assert.Len(t, log.entries, 1)
}

func TestAssess_AttemptCapAndManualReview(t *testing.T) {
	log, w := &fakeLog{}, &fakeWriter{}
	n := &Note{ID: uuid.New(), Title: "x", Content: "тело", SourceURL: "https://x.example",
		Metadata: map[string]any{"import_truncated": map[string]any{"sections_dropped": 2}},
	}
	noteID := uuid.New()

	// Signals change each pass via the injected keyword counter.
	stats := &fakeStats{keywords: 0}
	a2 := NewAssessor(&fakeNoteReader{n: n}, nil, w, log, stats, nil, DefaultThresholds())

	for i := 0; i < 3; i++ {
		stats.keywords = i // new signals each pass
		rec, err := a2.Assess(context.Background(), noteID, TriggerAuto)
		require.NoError(t, err)
		require.NotNil(t, rec)
		assert.Equal(t, i+1, rec.Attempt)
	}
	// Fourth auto pass — capped; the persistent gate flags manual review.
	rec, err := a2.Assess(context.Background(), noteID, TriggerAuto)
	require.NoError(t, err)
	require.NotNil(t, rec)
	assert.True(t, rec.NeedsManualReview)
	assert.Len(t, log.entries, 3)

	// Manual trigger is not capped.
	rec, err = a2.Assess(context.Background(), noteID, TriggerManual)
	require.NoError(t, err)
	require.NotNil(t, rec)
	assert.Len(t, log.entries, 4)
	assert.Equal(t, TriggerManual, log.entries[0].Trigger)
}

func TestAssess_ManualAlwaysWrites(t *testing.T) {
	log, w := &fakeLog{}, &fakeWriter{}
	a := newAssessor(&Note{ID: uuid.New(), Title: "T", Content: "текст"}, log, w)
	noteID := uuid.New()

	_, err := a.Assess(context.Background(), noteID, TriggerAuto)
	require.NoError(t, err)
	// Manual trigger with identical signals still lands (user asked).
	rec, err := a.Assess(context.Background(), noteID, TriggerManual)
	require.NoError(t, err)
	require.NotNil(t, rec)
	assert.Len(t, log.entries, 2)
	assert.False(t, log.entries[0].Changed)
}

func TestAssess_NoNote_Nil(t *testing.T) {
	a := newAssessor(nil, &fakeLog{}, &fakeWriter{})
	rec, err := a.Assess(context.Background(), uuid.New(), TriggerAuto)
	require.NoError(t, err)
	assert.Nil(t, rec)
}

func TestAssess_UsesArtifactWhenHashMatches(t *testing.T) {
	n := &Note{ID: uuid.New(), Title: "T", Content: "сырой текст заметки"}
	hash := SourceHash(n.Title, n.Content)
	artifact := &Artifact{SourceHash: hash, NormalizedText: "НОРМАЛИЗОВАННЫЙ ТЕКСТ ДЛЯ ЗАМЕТКИ"}

	a := NewAssessor(&fakeNoteReader{n: n},
		&fakeArtifacts{a: artifact}, &fakeWriter{}, &fakeLog{}, &fakeStats{}, nil,
		DefaultThresholds())
	rec, err := a.Assess(context.Background(), uuid.New(), TriggerAuto)
	require.NoError(t, err)
	require.NotNil(t, rec)
	// words counted from normalized text, not the raw body
	assert.Equal(t, 4, rec.Signals.Words)
}

type fakeArtifacts struct{ a *Artifact }

func (f *fakeArtifacts) CurrentArtifact(context.Context, uuid.UUID) (*Artifact, error) {
	return f.a, nil
}

// recordingEmbedder satisfies the Embedder port and logs every text it is
// asked to embed — used to prove the assessor never sends the source URL
// anywhere.
type recordingEmbedder struct{ texts []string }

func (r *recordingEmbedder) Embed(_ context.Context, text string, _ string) ([]float32, error) {
	r.texts = append(r.texts, text)
	return []float32{0.1, 0.2, 0.3}, nil
}

// Criterion 4: the assess task never touches the network — source_url is
// data, not a fetch instruction. A counting HTTP server stands in for any
// outbound call; the embedder records what it is fed.
func TestAssess_NeverFetchesSourceURL(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := &Note{
		ID:        uuid.New(),
		Title:     "Imported note",
		Content:   "First chunk of prose. Second chunk of prose.",
		SourceURL: server.URL + "/article",
	}
	hash := SourceHash(n.Title, n.Content)
	artifact := &Artifact{
		SourceHash:     hash,
		NormalizedText: n.Content,
		Chunks:         []string{"First chunk of prose.", "Second chunk of prose."},
	}
	emb := &recordingEmbedder{}
	a := NewAssessor(&fakeNoteReader{n: n}, &fakeArtifacts{a: artifact}, &fakeWriter{},
		&fakeLog{}, &fakeStats{}, emb, DefaultThresholds())

	rec, err := a.Assess(context.Background(), uuid.New(), TriggerAuto)
	require.NoError(t, err)
	require.NotNil(t, rec)

	assert.Equal(t, 0, hits, "assessor must not fetch source_url")
	for _, text := range emb.texts {
		assert.False(t, strings.Contains(text, server.URL), "embedder must never see the URL")
	}
}
