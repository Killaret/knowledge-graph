package notehandler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"knowledge-graph/internal/application/common"
	importer "knowledge-graph/internal/application/import"
	"knowledge-graph/internal/config"
	"knowledge-graph/internal/domain/cache/cachetest"
	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubTaskQueue is a no-op task queue for handler tests.
type stubTaskQueue struct{}

func (q *stubTaskQueue) EnqueueBackupToCloud(ctx context.Context, localPath, remoteKey, backupDate string) error {
	return nil
}

func (q *stubTaskQueue) EnqueueRefreshRecommendations(ctx context.Context, noteID uuid.UUID, delay time.Duration) error {
	return nil
}

func (q *stubTaskQueue) EnqueueExtractKeywords(ctx context.Context, noteID string, topN int) error {
	return nil
}

func (q *stubTaskQueue) EnqueueComputeEmbedding(ctx context.Context, noteID string) error {
	return nil
}

func (q *stubTaskQueue) EnqueueRecalculateLinkWeights(ctx context.Context, noteID uuid.UUID, delay time.Duration) error {
	return nil
}

func (q *stubTaskQueue) EnqueueNotification(ctx context.Context, payload []byte) error {
	return nil
}

func (q *stubTaskQueue) EnqueueBackupOnNoteChange(ctx context.Context) error {
	return nil
}

func (q *stubTaskQueue) EnqueueImportBookmarks(ctx context.Context, userID uuid.UUID, taskID string, items []byte) error {
	return nil
}

var _ common.TaskQueue = (*stubTaskQueue)(nil)

func setupImportRouter() (*gin.Engine, *mockNoteRepo, *mockLinkRepo, uuid.UUID) {
	gin.SetMode(gin.TestMode)

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	repo := newMockNoteRepo()
	linkRepo := newMockLinkRepo()
	cache := cachetest.NewFakeCacheClient()
	importSvc := importer.NewService(repo, cache, &stubTaskQueue{}, nil)

	cfg := &config.Config{
		RecommendationTopN:                    10,
		RecommendationFallbackSemanticEnabled: false,
		RecommendationFallbackEnabled:         false,
		RecommendationTaskDelaySeconds:        1,
		PaginationDefaultLimit:                20,
		PaginationMaxLimit:                    100,
	}
	handler := New(repo, nil, nil, nil, 0, nil, nil, nil, cfg, nil, nil, importSvc)
	handler.SetLinkRepository(linkRepo)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, userID)
		c.Next()
	})
	r.POST("/api/v1/import/bookmarks/preview", handler.ImportBookmarksPreview)
	r.POST("/api/v1/import/bookmarks", handler.ImportBookmarks)
	r.POST("/api/v1/import/batch", handler.ImportBatch)
	r.GET("/api/v1/import/:task_id/status", handler.ImportBookmarksStatus)

	return r, repo, linkRepo, userID
}

func TestImportBookmarksPreview(t *testing.T) {
	r, repo, _, userID := setupImportRouter()
	ctx := context.Background()

	// Seed an existing note with the same source URL.
	title, _ := note.NewTitle("Existing")
	content, _ := note.NewContent("content")
	meta, _ := note.NewMetadata(map[string]interface{}{"source_url": "https://example.com/existing"})
	existing := note.NewNoteWithCreator(title, content, "asteroid", meta, userID)
	require.NoError(t, repo.Save(ctx, existing))

	body := `{
		"items": [
			{"title": "Existing", "url": "https://example.com/existing", "type": "asteroid"},
			{"title": "New", "url": "https://example.com/new", "type": "asteroid"},
			{"title": "Bad URL", "url": "http://localhost"}
		],
		"options": {"default_type": "planet"}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/import/bookmarks/preview", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	items, ok := data["items"].([]interface{})
	require.True(t, ok)
	require.Len(t, items, 3)

	first := items[0].(map[string]interface{})
	assert.False(t, first["is_new"].(bool))
	assert.Equal(t, existing.ID().String(), first["existing_note_id"])

	second := items[1].(map[string]interface{})
	assert.True(t, second["is_new"].(bool))
	assert.Empty(t, second["existing_note_id"])

	third := items[2].(map[string]interface{})
	assert.NotEmpty(t, third["error"])
}

func TestImportBookmarksPreview_BatchTooLarge(t *testing.T) {
	r, _, _, _ := setupImportRouter()

	items := make([]map[string]string, 51)
	for i := range items {
		items[i] = map[string]string{"title": fmt.Sprintf("T%d", i), "url": fmt.Sprintf("https://example.com/%d", i)}
	}
	bodyBytes, _ := json.Marshal(map[string]interface{}{"items": items})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/import/bookmarks/preview", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestImportBookmarks_CreateAndStatus(t *testing.T) {
	r, _, _, _ := setupImportRouter()

	body := `{
		"items": [
			{"title": "One", "url": "https://example.com/one", "type": "asteroid"},
			{"title": "Two", "url": "https://example.com/two", "type": "planet"}
		]
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/import/bookmarks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	taskID, ok := data["task_id"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, taskID)

	// Check status endpoint.
	statusReq := httptest.NewRequest(http.MethodGet, "/api/v1/import/"+taskID+"/status", nil)
	statusW := httptest.NewRecorder()
	r.ServeHTTP(statusW, statusReq)

	assert.Equal(t, http.StatusOK, statusW.Code)
	var statusResp map[string]interface{}
	require.NoError(t, json.Unmarshal(statusW.Body.Bytes(), &statusResp))
	statusData, ok := statusResp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, taskID, statusData["task_id"])
	assert.Equal(t, "pending", statusData["status"])
}

func TestImportBookmarksStatus_NotFound(t *testing.T) {
	r, _, _, _ := setupImportRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/import/"+uuid.New().String()+"/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestImportBookmarksStatus_InvalidTaskID(t *testing.T) {
	r, _, _, _ := setupImportRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/import/not-a-uuid/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestImportBatch_NotesOnly(t *testing.T) {
	r, repo, _, _ := setupImportRouter()
	ctx := context.Background()

	body := `{"notes":[{"title":"Note A","content":"content A","type":"star"},{"title":"Note B","content":"content B","type":"planet"}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/import/batch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	createdNotes := data["created_notes"].([]interface{})
	failedNotes := data["failed_notes"].([]interface{})
	assert.Len(t, createdNotes, 2)
	assert.Len(t, failedNotes, 0)

	all, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 2)
}

func TestImportBatch_NotesAndLinks(t *testing.T) {
	r, repo, linkRepo, _ := setupImportRouter()
	ctx := context.Background()

	idA := uuid.New().String()
	idB := uuid.New().String()
	body := fmt.Sprintf(`{
		"notes": [
			{"id": "%s", "title": "Note A", "content": "content A", "type": "star"},
			{"id": "%s", "title": "Note B", "content": "content B", "type": "planet"}
		],
		"links": [
			{"source_note_id": "%s", "target_note_id": "%s", "link_type": "reference", "weight": 0.75}
		]
	}`, idA, idB, idA, idB)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/import/batch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	createdNotes := data["created_notes"].([]interface{})
	createdLinks := data["created_links"].([]interface{})
	assert.Len(t, createdNotes, 2)
	assert.Len(t, createdLinks, 1)

	all, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 2)

	links, err := linkRepo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, links, 1)
	assert.Equal(t, idA, links[0].SourceNoteID().String())
	assert.Equal(t, idB, links[0].TargetNoteID().String())
}

func TestImportBatch_WithExistingNoteLink(t *testing.T) {
	r, repo, linkRepo, userID := setupImportRouter()
	ctx := context.Background()

	title, _ := note.NewTitle("Existing")
	content, _ := note.NewContent("content")
	meta, _ := note.NewMetadata(nil)
	existing := note.NewNoteWithCreator(title, content, "star", meta, userID)
	require.NoError(t, repo.Save(ctx, existing))

	newID := uuid.New().String()
	body := fmt.Sprintf(`{
		"notes": [
			{"id": "%s", "title": "New Note", "content": "content", "type": "planet"}
		],
		"links": [
			{"source_note_id": "%s", "target_note_id": "%s", "link_type": "related", "weight": 0.9}
		]
	}`, newID, newID, existing.ID().String())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/import/batch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	createdNotes := data["created_notes"].([]interface{})
	createdLinks := data["created_links"].([]interface{})
	assert.Len(t, createdNotes, 1)
	assert.Len(t, createdLinks, 1)

	links, err := linkRepo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, links, 1)
}

func TestImportBatch_ClientIDCollisionWithExisting(t *testing.T) {
	r, repo, _, userID := setupImportRouter()
	ctx := context.Background()

	title, _ := note.NewTitle("Original")
	content, _ := note.NewContent("content")
	meta, _ := note.NewMetadata(nil)
	existing := note.NewNoteWithCreator(title, content, "star", meta, userID)
	require.NoError(t, repo.Save(ctx, existing))

	body := fmt.Sprintf(`{"notes":[{"id":"%s","title":"Overwritten","content":"new","type":"planet"}]}`, existing.ID().String())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/import/batch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	createdNotes := data["created_notes"].([]interface{})
	failedNotes := data["failed_notes"].([]interface{})
	assert.Len(t, createdNotes, 0)
	assert.Len(t, failedNotes, 1)

	saved, err := repo.FindByID(ctx, existing.ID())
	require.NoError(t, err)
	assert.Equal(t, "Original", saved.Title().String())
}

func TestImportBatch_ClientIDCollisionWithinRequest(t *testing.T) {
	r, repo, _, _ := setupImportRouter()
	ctx := context.Background()

	sharedID := uuid.New().String()
	body := fmt.Sprintf(`{
		"notes": [
			{"id": "%s", "title": "First", "content": "content", "type": "star"},
			{"id": "%s", "title": "Second", "content": "content", "type": "planet"}
		]
	}`, sharedID, sharedID)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/import/batch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	createdNotes := data["created_notes"].([]interface{})
	failedNotes := data["failed_notes"].([]interface{})
	assert.Len(t, createdNotes, 1)
	assert.Len(t, failedNotes, 1)

	all, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)
}

func TestImportBatch_LinkToMissingNote(t *testing.T) {
	r, repo, linkRepo, _ := setupImportRouter()
	ctx := context.Background()

	body := `{
		"notes": [{"title": "Only Note", "content": "content", "type": "star"}],
		"links": [{"source_note_id": "11111111-1111-1111-1111-111111111111", "target_note_id": "22222222-2222-2222-2222-222222222222", "link_type": "reference", "weight": 0.5}]
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/import/batch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	createdNotes := data["created_notes"].([]interface{})
	failedLinks := data["failed_links"].([]interface{})
	assert.Len(t, createdNotes, 1)
	assert.Len(t, failedLinks, 1)

	all, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)

	links, err := linkRepo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, links, 0)
}

func TestImportBatch_LinkToFailedNote(t *testing.T) {
	r, repo, linkRepo, _ := setupImportRouter()
	ctx := context.Background()

	missingID := uuid.New().String()
	body := fmt.Sprintf(`{
		"notes": [
			{"title": "Valid Note", "content": "content", "type": "star"},
			{"title": "   ", "content": "whitespace title", "type": "planet"}
		],
		"links": [
			{"source_note_id": "%s", "target_note_id": "%s", "link_type": "reference", "weight": 0.5}
		]
	}`, missingID, missingID)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/import/batch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	createdNotes := data["created_notes"].([]interface{})
	failedNotes := data["failed_notes"].([]interface{})
	failedLinks := data["failed_links"].([]interface{})
	assert.Len(t, createdNotes, 1)
	assert.Len(t, failedNotes, 1)
	assert.Len(t, failedLinks, 1)

	all, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)

	links, err := linkRepo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, links, 0)
}

func TestImportBatch_DuplicateLink(t *testing.T) {
	r, repo, linkRepo, _ := setupImportRouter()
	ctx := context.Background()

	idA := uuid.New().String()
	idB := uuid.New().String()
	body := fmt.Sprintf(`{
		"notes": [
			{"id": "%s", "title": "Note A", "content": "content", "type": "star"},
			{"id": "%s", "title": "Note B", "content": "content", "type": "planet"}
		],
		"links": [
			{"source_note_id": "%s", "target_note_id": "%s", "link_type": "reference", "weight": 0.8},
			{"source_note_id": "%s", "target_note_id": "%s", "link_type": "reference", "weight": 0.9}
		]
	}`, idA, idB, idA, idB, idA, idB)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/import/batch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	createdLinks := data["created_links"].([]interface{})
	failedLinks := data["failed_links"].([]interface{})
	assert.Len(t, createdLinks, 1)
	assert.Len(t, failedLinks, 1)

	links, err := linkRepo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, links, 1)

	all, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 2)
}

func TestImportBatch_EmptyNotes(t *testing.T) {
	r, _, _, _ := setupImportRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/import/batch", bytes.NewBufferString(`{"notes":[]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestImportBatch_TooManyNotes(t *testing.T) {
	r, _, _, _ := setupImportRouter()

	notes := make([]map[string]string, 51)
	for i := range notes {
		notes[i] = map[string]string{"title": fmt.Sprintf("Note %d", i)}
	}
	bodyBytes, _ := json.Marshal(map[string]interface{}{"notes": notes})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/import/batch", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestImportBatch_InvalidLinkType(t *testing.T) {
	r, _, _, _ := setupImportRouter()

	body := `{
		"notes": [{"title": "Note", "content": "content", "type": "star"}],
		"links": [{"source_note_id": "11111111-1111-1111-1111-111111111111", "target_note_id": "22222222-2222-2222-2222-222222222222", "link_type": "invalid", "weight": 0.5}]
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/import/batch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestImportBatch_InvalidLinkWeight(t *testing.T) {
	r, _, _, _ := setupImportRouter()

	body := `{
		"notes": [{"title": "Note", "content": "content", "type": "star"}],
		"links": [{"source_note_id": "11111111-1111-1111-1111-111111111111", "target_note_id": "22222222-2222-2222-2222-222222222222", "link_type": "reference", "weight": 2.5}]
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/import/batch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestImportBatch_MissingNoteTitle(t *testing.T) {
	r, repo, _, _ := setupImportRouter()
	ctx := context.Background()

	body := `{"notes":[{"content":"Missing title"}]}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/import/batch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	all, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 0)
}
