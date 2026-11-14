package notehandler

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
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
	alive  bool
	pushed [][]byte
}

func (q *stubDocumentQueue) IsListenerAlive(ctx context.Context) (bool, error) {
	return q.alive, nil
}

func (q *stubDocumentQueue) PushTask(ctx context.Context, payload []byte) error {
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

func documentFormRequest(t *testing.T, fields map[string]string, fileField, filename, contentType string, content []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	for k, v := range fields {
		require.NoError(t, w.WriteField(k, v))
	}
	if fileField != "" {
		header := textproto.MIMEHeader{}
		header.Set("Content-Disposition", `form-data; name="`+fileField+`"; filename="`+filename+`"`)
		header.Set("Content-Type", contentType)
		part, err := w.CreatePart(header)
		require.NoError(t, err)
		_, err = part.Write(content)
		require.NoError(t, err)
	}
	require.NoError(t, w.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/import/document", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
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
	r.ServeHTTP(w, documentFormRequest(t, map[string]string{"type": "TEXT", "content": "hello"}, "", "", "", nil))
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "source_text_handler_unavailable")
	assert.Empty(t, dq.pushed)
}

func TestImportDocument_TextAccepted(t *testing.T) {
	r, dq, _ := setupDocumentRouter(true)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, documentFormRequest(t, map[string]string{"type": "TEXT", "content": "hello"}, "", "", "", nil))
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

func TestImportDocument_RejectsPrivateURL(t *testing.T) {
	r, dq, _ := setupDocumentRouter(true)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, documentFormRequest(t, map[string]string{"type": "URL", "content": "http://127.0.0.1/x"}, "", "", "", nil))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, dq.pushed)
}

func TestImportDocument_RejectsEmptyText(t *testing.T) {
	r, dq, _ := setupDocumentRouter(true)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, documentFormRequest(t, map[string]string{"type": "TEXT", "content": "   "}, "", "", "", nil))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, dq.pushed)
}

func TestImportDocument_FileOverLimit(t *testing.T) {
	r, dq, _ := setupDocumentRouter(true)

	big := make([]byte, maxDocumentFileSize+1)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, documentFormRequest(t, map[string]string{"type": "FILE"}, "file", "big.bin", "application/octet-stream", big))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, dq.pushed)
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
	owner.ServeHTTP(w, documentFormRequest(t, map[string]string{"type": "TEXT", "content": "hello"}, "", "", "", nil))
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
