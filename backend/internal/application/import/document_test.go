package importer

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"sort"
	"testing"
	"time"

	"knowledge-graph/internal/domain/cache"
	"knowledge-graph/internal/domain/cache/cachetest"
	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// fakeDocumentQueue is an in-memory DocumentQueue for unit tests. endless is
// answered on every poll once results are drained — it simulates a continuous
// response stream (the sweep must not starve behind it).
type fakeDocumentQueue struct {
	alive   bool
	pushed  [][]byte
	results [][]byte
	endless []byte
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
		if q.endless != nil {
			return q.endless, nil
		}
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

// tz3TaskFixture is the §3 example of TZ-Java-source-text-handler-2026-08-30.md
// verbatim. The contract is defined by this literal, not by a key list typed
// from memory — the review mutation (a renamed smart_weights field and an
// overlap value change) went unnoticed under the old check because it verified
// only two of nine weights and two of eight options.
const tz3TaskFixture = `{
	"event_id": "uuid-task-1",
	"correlation_id": "uuid-correlation-1",
	"user_id": "uuid-user-1",
	"jwt": "eyJhbGci...",
	"type": "FILE|URL|TEXT",
	"content": "base64-string | URL | plain text",
	"content_type": "application/pdf",
	"import_options": {
		"chunk_size": 500,
		"overlap": 50,
		"min_chunk_length": 100,
		"max_chunk_length": 8000,
		"max_retries": 3,
		"create_links": false,
		"smart_threshold": 70,
		"smart_weights": {
			"has_heading": 25,
			"section_boundary": 10,
			"intra_chunk_cohesion": 20,
			"punctuation_completeness": 15,
			"chunk_ending_weight": 15,
			"length_ok": 10,
			"keyword_trend_stable": 5,
			"punctuation_balance": 5,
			"metadata_present": 5
		}
	},
	"metadata": {
		"filename": "doc.pdf",
		"source": "telegram"
	}
}`

// Criterion 2: the queued message parses into exactly the §3 structure —
// the same top-level field set, import_options deep-equal to the §3 literal,
// snake_case everywhere, no envelope.
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

	var fixture map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(tz3TaskFixture), &fixture))
	// §3 marks content_type as FILE-only; a TEXT task must not carry it.
	delete(fixture, "content_type")

	var task map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(dq.pushed[0], &task))

	// Top-level field set: exactly the §3 keys — no missing field, no extra
	// field, no asynq envelope.
	require.Equal(t, sortedKeys(fixture), sortedKeys(task),
		"top-level fields differ from the §3 contract")

	require.Equal(t, `"TEXT"`, string(task["type"]))
	require.Equal(t, `"three paragraphs of text"`, string(task["content"]))
	require.Equal(t, `"jwt-token-123"`, string(task["jwt"]))
	require.Equal(t, `"`+userID.String()+`"`, string(task["user_id"]))
	require.Equal(t, `"`+taskID+`"`, string(task["correlation_id"]))

	// import_options — the whole object, not sampled keys: every key and
	// every value must equal the §3 literal.
	var wantOptions, gotOptions map[string]interface{}
	require.NoError(t, json.Unmarshal(fixture["import_options"], &wantOptions))
	require.NoError(t, json.Unmarshal(task["import_options"], &gotOptions))
	require.Equal(t, wantOptions, gotOptions)

	var meta map[string]interface{}
	require.NoError(t, json.Unmarshal(task["metadata"], &meta))
	require.Equal(t, "web", meta["source"])
}

func sortedKeys(m map[string]json.RawMessage) []string {
	keys := keysOf(m)
	sort.Strings(keys)
	return keys
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

// The timeout sweep must not starve behind a continuous result stream: when
// every BLPOP answers instantly, pending tasks past their deadline still fail
// — the mutation "sweep only on an empty pop" turns this red.
func TestRunDocumentResultLoop_SweepsDuringResultStream(t *testing.T) {
	svc, _, _, dq := newDocumentService()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// The endless results all carry unknown correlation ids — the drop log
	// line would fire thousands of times otherwise.
	old := log.Writer()
	log.SetOutput(io.Discard)
	defer log.SetOutput(old)

	staleID, err := svc.SubmitDocument(ctx, uuid.New(), "jwt", DocumentInput{Type: "TEXT", Content: "x"})
	require.NoError(t, err)
	stale, err := svc.loadStatus(ctx, staleID)
	require.NoError(t, err)
	stale.CreatedAt = time.Now().Add(-time.Hour)
	require.NoError(t, svc.storeStatus(ctx, stale))

	// An endless stream of unknown-correlation results keeps every pop hot.
	dq.endless = []byte(`{"correlation_id":"` + uuid.New().String() + `","status":"COMPLETED"}`)

	go svc.RunDocumentResultLoop(ctx, time.Millisecond, 2*time.Millisecond, 10*time.Minute)

	require.Eventually(t, func() bool {
		st, err := svc.GetTaskStatus(ctx, staleID)
		return err == nil && st.Status == statusFailed && st.Reason == reasonTimeout
	}, 5*time.Second, 10*time.Millisecond)
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
