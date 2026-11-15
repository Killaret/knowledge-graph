package importer

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"testing"
	"time"

	"knowledge-graph/internal/domain/cache"
	"knowledge-graph/internal/domain/cache/cachetest"
	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// fakeDocumentQueue is an in-memory DocumentQueue for unit tests.
type fakeDocumentQueue struct {
	alive   bool
	pushed  [][]byte
	results [][]byte
}

func (q *fakeDocumentQueue) IsListenerAlive(ctx context.Context) (bool, error) {
	return q.alive, nil
}

func (q *fakeDocumentQueue) PushTask(ctx context.Context, payload []byte) error {
	q.pushed = append(q.pushed, payload)
	return nil
}

func (q *fakeDocumentQueue) PopResult(ctx context.Context, timeout time.Duration) ([]byte, error) {
	if len(q.results) == 0 {
		return nil, ErrResultTimeout
	}
	payload := q.results[0]
	q.results = q.results[1:]
	return payload, nil
}

var _ DocumentQueue = (*fakeDocumentQueue)(nil)

func newDocumentService() (*Service, *fakeNoteRepo, *cachetest.FakeCacheClient, *fakeDocumentQueue) {
	repo := newFakeNoteRepo()
	c := cachetest.NewFakeCacheClient()
	dq := &fakeDocumentQueue{alive: true}
	svc := NewService(repo, c, nil, nil)
	svc.SetDocumentQueue(dq)
	return svc, repo, c, dq
}

// Criterion 1: without the listener key nothing is accepted — the mutation
// "the key is not checked" turns every assertion in this test red.
func TestDocumentAvailability(t *testing.T) {
	svc, _, _, dq := newDocumentService()
	ctx := context.Background()

	dq.alive = false
	available, err := svc.DocumentAvailability(ctx)
	require.NoError(t, err)
	require.False(t, available)

	dq.alive = true
	available, err = svc.DocumentAvailability(ctx)
	require.NoError(t, err)
	require.True(t, available)
}

func TestSubmitDocument_NoListener(t *testing.T) {
	svc, _, _, dq := newDocumentService()
	dq.alive = false

	taskID, err := svc.SubmitDocument(context.Background(), uuid.New(), "jwt", DocumentInput{
		Type:    "TEXT",
		Content: "some text",
	})
	require.ErrorIs(t, err, ErrDocumentUnavailable)
	require.Empty(t, taskID)
	require.Empty(t, dq.pushed, "nothing may be queued while the listener is down")
}

// Criterion 2: the queued message parses into exactly the §3 structure —
// same field set, snake_case, no envelope.
func TestSubmitDocument_TaskContract(t *testing.T) {
	svc, _, _, dq := newDocumentService()
	userID := uuid.New()

	taskID, err := svc.SubmitDocument(context.Background(), userID, "jwt-token-123", DocumentInput{
		Type:    "TEXT",
		Content: "three paragraphs of text",
	})
	require.NoError(t, err)
	require.NotEmpty(t, taskID)
	require.Len(t, dq.pushed, 1)

	var task map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(dq.pushed[0], &task))

	wantKeys := []string{
		"event_id", "correlation_id", "user_id", "jwt", "type",
		"content", "import_options", "metadata",
	}
	for _, key := range wantKeys {
		require.Contains(t, task, key)
	}
	// No extra top-level fields and no asynq envelope.
	require.Len(t, task, len(wantKeys)+0, "unexpected top-level fields: %v", keysOf(task))
	require.NotContains(t, task, "content_type")
	require.NotContains(t, task, "typename")
	require.NotContains(t, task, "payload")

	require.Equal(t, `"TEXT"`, string(task["type"]))
	require.Equal(t, `"three paragraphs of text"`, string(task["content"]))
	require.Equal(t, `"jwt-token-123"`, string(task["jwt"]))
	require.Equal(t, `"`+userID.String()+`"`, string(task["user_id"]))
	require.Equal(t, `"`+taskID+`"`, string(task["correlation_id"]))

	var options map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(task["import_options"], &options))
	for _, key := range []string{
		"chunk_size", "overlap", "min_chunk_length", "max_chunk_length",
		"max_retries", "create_links", "smart_threshold", "smart_weights",
	} {
		require.Contains(t, options, key)
	}
	require.Equal(t, `500`, string(options["chunk_size"]))
	require.Equal(t, `70`, string(options["smart_threshold"]))

	var weights map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(options["smart_weights"], &weights))
	require.Equal(t, `25`, string(weights["has_heading"]))
	require.Equal(t, `5`, string(weights["metadata_present"]))

	var meta map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(task["metadata"], &meta))
	require.Equal(t, `"web"`, string(meta["source"]))
}

func TestSubmitDocument_FileBase64(t *testing.T) {
	svc, _, _, dq := newDocumentService()

	_, err := svc.SubmitDocument(context.Background(), uuid.New(), "jwt", DocumentInput{
		Type:        "FILE",
		FileContent: []byte("pdf-bytes"),
		ContentType: "application/pdf",
		Filename:    "doc.pdf",
	})
	require.NoError(t, err)
	require.Len(t, dq.pushed, 1)

	var task DocumentTask
	require.NoError(t, json.Unmarshal(dq.pushed[0], &task))
	require.Equal(t, "FILE", task.Type)
	require.Equal(t, "application/pdf", task.ContentType)
	require.Equal(t, "doc.pdf", task.Metadata["filename"])
	require.Equal(t, []byte("pdf-bytes"), mustDecodeBase64(t, task.Content))
}

func mustDecodeBase64(t *testing.T, s string) []byte {
	t.Helper()
	raw, err := base64.StdEncoding.DecodeString(s)
	require.NoError(t, err)
	return raw
}

func TestSubmitDocument_StoresPendingStatus(t *testing.T) {
	svc, _, _, _ := newDocumentService()
	userID := uuid.New()

	taskID, err := svc.SubmitDocument(context.Background(), userID, "jwt", DocumentInput{
		Type:    "URL",
		Content: "https://example.com/article",
	})
	require.NoError(t, err)

	status, err := svc.GetTaskStatus(context.Background(), taskID)
	require.NoError(t, err)
	require.Equal(t, "pending", status.Status)
	require.Equal(t, documentKind, status.Kind)
	require.Equal(t, userID.String(), status.UserID)
	require.False(t, status.CreatedAt.IsZero())
}

// Criterion 3: §4 fixtures map onto the cached status.
func TestApplyDocumentResult_Completed(t *testing.T) {
	svc, repo, _, _ := newDocumentService()
	ctx := context.Background()
	userID := uuid.New()

	meta, err := note.NewMetadata(nil)
	require.NoError(t, err)
	noteA := note.NewNoteWithCreator(mustTitle(t, "Alpha"), mustContent(t, "a"), note.MustType("unknown"), meta, userID)
	noteB := note.NewNoteWithCreator(mustTitle(t, "Beta"), mustContent(t, "b"), note.MustType("unknown"), meta, userID)
	require.NoError(t, repo.Save(ctx, noteA))
	require.NoError(t, repo.Save(ctx, noteB))

	taskID, err := svc.SubmitDocument(ctx, userID, "jwt", DocumentInput{Type: "TEXT", Content: "text"})
	require.NoError(t, err)

	payload := []byte(`{
		"correlation_id": "` + taskID + `",
		"event_id": "` + uuid.New().String() + `",
		"user_id": "` + userID.String() + `",
		"status": "COMPLETED",
		"note_ids": ["` + noteA.ID().String() + `", "` + noteB.ID().String() + `"],
		"raw_chunks": [],
		"links": [],
		"errors": [],
		"processed_at": "2026-10-18T12:00:00Z"
	}`)
	svc.ApplyDocumentResult(ctx, payload)

	status, err := svc.GetTaskStatus(ctx, taskID)
	require.NoError(t, err)
	require.Equal(t, "done", status.Status)
	require.Len(t, status.Notes, 2)
	require.Equal(t, noteA.ID().String(), status.Notes[0].ID)
	require.Equal(t, "Alpha", status.Notes[0].Title)
	require.Equal(t, "Beta", status.Notes[1].Title)
	require.Zero(t, status.RawChunksCount)
}

func TestApplyDocumentResult_Partial(t *testing.T) {
	svc, _, _, _ := newDocumentService()
	ctx := context.Background()

	taskID, err := svc.SubmitDocument(ctx, uuid.New(), "jwt", DocumentInput{Type: "TEXT", Content: "text"})
	require.NoError(t, err)

	payload := []byte(`{
		"correlation_id": "` + taskID + `",
		"status": "PARTIAL",
		"note_ids": [],
		"raw_chunks": [{"index": 0, "content": "raw"}, {"index": 1, "content": "raw2"}],
		"errors": []
	}`)
	svc.ApplyDocumentResult(ctx, payload)

	status, err := svc.GetTaskStatus(ctx, taskID)
	require.NoError(t, err)
	require.Equal(t, "done", status.Status)
	require.Equal(t, 2, status.RawChunksCount)
}

func TestApplyDocumentResult_Failed(t *testing.T) {
	svc, _, _, _ := newDocumentService()
	ctx := context.Background()

	taskID, err := svc.SubmitDocument(ctx, uuid.New(), "jwt", DocumentInput{Type: "TEXT", Content: "text"})
	require.NoError(t, err)

	payload := []byte(`{
		"correlation_id": "` + taskID + `",
		"status": "FAILED",
		"errors": ["parse failed"]
	}`)
	svc.ApplyDocumentResult(ctx, payload)

	status, err := svc.GetTaskStatus(ctx, taskID)
	require.NoError(t, err)
	require.Equal(t, "failed", status.Status)
	require.Equal(t, []string{"parse failed"}, status.Errors)
}

func TestApplyDocumentResult_UnknownCorrelationDropped(t *testing.T) {
	svc, _, _, _ := newDocumentService()
	svc.ApplyDocumentResult(context.Background(), []byte(`{
		"correlation_id": "`+uuid.New().String()+`",
		"status": "COMPLETED"
	}`))
}

// Criterion 3b: a document task whose answer never arrives fails as timeout.
func TestFailTimedOutTasks(t *testing.T) {
	svc, _, c, _ := newDocumentService()
	ctx := context.Background()

	freshID, err := svc.SubmitDocument(ctx, uuid.New(), "jwt", DocumentInput{Type: "TEXT", Content: "x"})
	require.NoError(t, err)
	staleID, err := svc.SubmitDocument(ctx, uuid.New(), "jwt", DocumentInput{Type: "TEXT", Content: "y"})
	require.NoError(t, err)

	// Age the second task past the deadline.
	stale, err := svc.loadStatus(ctx, staleID)
	require.NoError(t, err)
	stale.CreatedAt = time.Now().Add(-time.Hour)
	require.NoError(t, svc.storeStatus(ctx, stale))
	_ = c // statuses go through the service cache already

	svc.failTimedOutTasks(ctx, 10*time.Minute)

	fresh, err := svc.GetTaskStatus(ctx, freshID)
	require.NoError(t, err)
	require.Equal(t, "pending", fresh.Status)

	staleAfter, err := svc.GetTaskStatus(ctx, staleID)
	require.NoError(t, err)
	require.Equal(t, "failed", staleAfter.Status)
	require.Equal(t, "timeout", staleAfter.Reason)
}

// The sweep must not touch bookmark tasks — they own their lifecycle.
func TestFailTimedOutTasks_SkipsBookmarkTasks(t *testing.T) {
	svc, _, _, _ := newDocumentService()
	ctx := context.Background()

	taskID, err := svc.StartImport(ctx, uuid.New(), []Item{{URL: "https://example.com/x"}})
	require.Error(t, err) // no task queue configured in this fixture
	_ = taskID

	// Store a bookmark-kind stale status directly.
	stale := TaskStatus{
		TaskID:    uuid.New().String(),
		Status:    statusPending,
		UserID:    uuid.New().String(),
		CreatedAt: time.Now().Add(-time.Hour),
	}
	require.NoError(t, svc.storeStatus(ctx, stale))

	svc.failTimedOutTasks(ctx, 10*time.Minute)

	after, err := svc.GetTaskStatus(ctx, stale.TaskID)
	require.NoError(t, err)
	require.Equal(t, "pending", after.Status)
}

// Criterion 4: a foreign status answers with the cache-miss sentinel the
// handler maps to 404.
func TestGetTaskStatusForUser_Foreign(t *testing.T) {
	svc, _, _, _ := newDocumentService()
	ctx := context.Background()
	owner := uuid.New()

	taskID, err := svc.SubmitDocument(ctx, owner, "jwt", DocumentInput{Type: "TEXT", Content: "x"})
	require.NoError(t, err)

	_, err = svc.GetTaskStatusForUser(ctx, taskID, owner)
	require.NoError(t, err)

	_, err = svc.GetTaskStatusForUser(ctx, taskID, uuid.New())
	require.ErrorIs(t, err, cache.ErrCacheMiss)
}

// Criterion 5: result logging must not leak JWT or document content — the
// reader logs correlation ids and error text only.
func TestApplyDocumentResult_LogsNoSecrets(t *testing.T) {
	svc, _, _, _ := newDocumentService()

	var buf bytes.Buffer
	old := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(old)

	svc.ApplyDocumentResult(context.Background(), []byte(`{
		"correlation_id": "`+uuid.New().String()+`",
		"status": "COMPLETED",
		"jwt": "super-secret-jwt",
		"note_ids": []
	}`))
	svc.ApplyDocumentResult(context.Background(), []byte(`{not json — secret-doc-content`))

	out := buf.String()
	require.NotContains(t, out, "super-secret-jwt")
	require.NotContains(t, out, "secret-doc-content")
}

func keysOf(m map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func mustTitle(t *testing.T, s string) note.Title {
	t.Helper()
	title, err := note.NewTitle(s)
	require.NoError(t, err)
	return title
}

func mustContent(t *testing.T, s string) note.Content {
	t.Helper()
	content, err := note.NewContent(s)
	require.NoError(t, err)
	return content
}
