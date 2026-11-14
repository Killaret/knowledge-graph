package notehandler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	importer "knowledge-graph/internal/application/import"
	"knowledge-graph/internal/config"
	"knowledge-graph/internal/domain/cache/cachetest"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubDocumentQueue is an in-memory importer.DocumentQueue for handler tests.
type stubDocumentQueue struct {
	alive   bool
	pushErr error
	pushed  [][]byte
}

func (q *stubDocumentQueue) IsListenerAlive(ctx context.Context) (bool, error) {
	return q.alive, nil
}

func (q *stubDocumentQueue) PushTask(ctx context.Context, payload []byte) error {
	if q.pushErr != nil {
		return q.pushErr
	}
	q.pushed = append(q.pushed, payload)
	return nil
}

func (q *stubDocumentQueue) PopResult(ctx context.Context, timeout time.Duration) ([]byte, error) {
	return nil, importer.ErrResultTimeout
}

var documentTestOwner = uuid.MustParse("11111111-1111-1111-1111-111111111111")

func setupDocumentRouter(alive bool) (*gin.Engine, *stubDocumentQueue, *cachetest.FakeCacheClient) {
	gin.SetMode(gin.TestMode)

	cache := cachetest.NewFakeCacheClient()
	dq := &stubDocumentQueue{alive: alive}
	importSvc := importer.NewService(newMockNoteRepo(), cache, &stubTaskQueue{}, nil)
	importSvc.SetDocumentQueue(dq)

	r := documentRouterFor(importSvc, documentTestOwner)
	return r, dq, cache
}

// documentRouterFor mounts the document routes with a fixed request user.
func documentRouterFor(importSvc *importer.Service, userID uuid.UUID) *gin.Engine {
	handler := New(newMockNoteRepo(), nil, nil, nil, 0, nil, nil, nil, &config.Config{}, nil, nil, importSvc)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, userID)
		c.Next()
	})
	r.GET("/api/v1/import/document/availability", handler.ImportDocumentAvailability)
	r.POST("/api/v1/import/document", handler.ImportDocument)
	r.GET("/api/v1/import/:task_id/status", handler.ImportBookmarksStatus)
	return r
}

// documentRequest builds the raw-body upload the frontend sends: metadata in
// the query string, the document itself as application/octet-stream.
func documentRequest(t *testing.T, query url.Values, body []byte) *http.Request {
	t.Helper()
	target := "/api/v1/import/document"
	if len(query) > 0 {
		target += "?" + query.Encode()
	}
	req := httptest.NewRequest(http.MethodPost, target, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Authorization", "Bearer test-jwt")
	return req
}

// Criterion 1: no listener key → availability false, POST answers 503 with the
// contract code and queues nothing.
func TestImportDocument_NoListener(t *testing.T) {
	r, dq, _ := setupDocumentRouter(false)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/import/document/availability", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"available":false`)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, documentRequest(t, url.Values{"type": {"TEXT"}}, []byte("hello")))
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "source_text_handler_unavailable")
	assert.Empty(t, dq.pushed)
}

func TestImportDocument_TextAccepted(t *testing.T) {
	r, dq, _ := setupDocumentRouter(true)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, documentRequest(t, url.Values{"type": {"TEXT"}}, []byte("hello")))
	assert.Equal(t, http.StatusAccepted, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.NotEmpty(t, data["task_id"])
	require.Len(t, dq.pushed, 1)

	var task map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(dq.pushed[0], &task))
	assert.Equal(t, `"TEXT"`, string(task["type"]))
	assert.Equal(t, `"test-jwt"`, string(task["jwt"]))
}

func TestImportDocument_FileAccepted(t *testing.T) {
	r, dq, _ := setupDocumentRouter(true)

	q := url.Values{"type": {"FILE"}, "filename": {"doc.pdf"}, "content_type": {"application/pdf"}}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, documentRequest(t, q, []byte("pdf-bytes")))
	assert.Equal(t, http.StatusAccepted, w.Code)
	require.Len(t, dq.pushed, 1)

	var task map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(dq.pushed[0], &task))
	assert.Equal(t, `"FILE"`, string(task["type"]))
	assert.Equal(t, `"application/pdf"`, string(task["content_type"]))
	var meta map[string]interface{}
	require.NoError(t, json.Unmarshal(task["metadata"], &meta))
	assert.Equal(t, "doc.pdf", meta["filename"])
}

func TestImportDocument_RejectsPrivateURL(t *testing.T) {
	r, dq, _ := setupDocumentRouter(true)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, documentRequest(t, url.Values{"type": {"URL"}}, []byte("http://127.0.0.1/x")))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, dq.pushed)
}

func TestImportDocument_RejectsEmptyText(t *testing.T) {
	r, dq, _ := setupDocumentRouter(true)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, documentRequest(t, url.Values{"type": {"TEXT"}}, []byte("   ")))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, dq.pushed)
}

func TestImportDocument_FileOverLimit(t *testing.T) {
	r, dq, _ := setupDocumentRouter(true)

	big := make([]byte, maxDocumentFileSize+1)
	q := url.Values{"type": {"FILE"}, "filename": {"big.bin"}}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, documentRequest(t, q, big))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, dq.pushed)
}

func TestImportDocument_FileEmpty(t *testing.T) {
	r, dq, _ := setupDocumentRouter(true)

	q := url.Values{"type": {"FILE"}, "filename": {"empty.bin"}}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, documentRequest(t, q, nil))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, dq.pushed)
}

func TestImportDocument_FileWithoutFilename(t *testing.T) {
	r, dq, _ := setupDocumentRouter(true)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, documentRequest(t, url.Values{"type": {"FILE"}}, []byte("bytes")))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, dq.pushed)
}

func TestImportDocument_FileWithoutContentType(t *testing.T) {
	r, dq, _ := setupDocumentRouter(true)

	q := url.Values{"type": {"FILE"}, "filename": {"mystery.bin"}}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, documentRequest(t, q, []byte("bytes")))
	assert.Equal(t, http.StatusAccepted, w.Code)
	require.Len(t, dq.pushed, 1)

	var task map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(dq.pushed[0], &task))
	// §3 requires content_type for FILE; the lawful default fills in.
	assert.Equal(t, `"application/octet-stream"`, string(task["content_type"]))
}

func TestImportDocument_RejectsMissingType(t *testing.T) {
	r, dq, _ := setupDocumentRouter(true)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, documentRequest(t, nil, []byte("hello")))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, dq.pushed)
}

func TestImportDocument_RejectsUnknownType(t *testing.T) {
	r, dq, _ := setupDocumentRouter(true)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, documentRequest(t, url.Values{"type": {"IMAGE"}}, []byte("hello")))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, dq.pushed)
}

// A leftover multipart client must not smuggle its envelope into the document
// body — the queue would get boundary junk as task content.
func TestImportDocument_RejectsMultipartEnvelope(t *testing.T) {
	r, dq, _ := setupDocumentRouter(true)

	body := []byte("--boundary\r\nContent-Disposition: form-data; name=\"type\"\r\n\r\nTEXT\r\n--boundary--")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/import/document?type=TEXT", bytes.NewReader(body))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")
	req.Header.Set("Authorization", "Bearer test-jwt")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, dq.pushed)
}

// Internal failures are 500 without internals — not a 400 carrying the wrapped
// error text to the client (review finding).
func TestImportDocument_InternalErrorIs500(t *testing.T) {
	r, dq, _ := setupDocumentRouter(true)
	dq.pushErr = errors.New("redis exploded internally")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, documentRequest(t, url.Values{"type": {"TEXT"}}, []byte("hello")))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NotContains(t, w.Body.String(), "redis exploded internally")
}

// Criterion 4: a task owned by another user answers 404 on the status route.
func TestImportDocumentStatus_Foreign(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache := cachetest.NewFakeCacheClient()
	dq := &stubDocumentQueue{alive: true}
	importSvc := importer.NewService(newMockNoteRepo(), cache, &stubTaskQueue{}, nil)
	importSvc.SetDocumentQueue(dq)

	owner := documentRouterFor(importSvc, documentTestOwner)
	stranger := documentRouterFor(importSvc, uuid.New())

	// Owner creates a task.
	w := httptest.NewRecorder()
	owner.ServeHTTP(w, documentRequest(t, url.Values{"type": {"TEXT"}}, []byte("hello")))
	require.Equal(t, http.StatusAccepted, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	taskID := resp["data"].(map[string]interface{})["task_id"].(string)

	// Owner reads it.
	statusW := httptest.NewRecorder()
	owner.ServeHTTP(statusW, httptest.NewRequest(http.MethodGet, "/api/v1/import/"+taskID+"/status", nil))
	assert.Equal(t, http.StatusOK, statusW.Code)

	// A different user gets 404.
	strangerW := httptest.NewRecorder()
	stranger.ServeHTTP(strangerW, httptest.NewRequest(http.MethodGet, "/api/v1/import/"+taskID+"/status", nil))
	assert.Equal(t, http.StatusNotFound, strangerW.Code)
}
