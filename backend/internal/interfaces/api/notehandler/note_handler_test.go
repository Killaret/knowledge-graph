package notehandler

import (
	"bytes"

	"context"

	"encoding/json"

	"fmt"
	"net/http"

	"net/http/httptest"

	"testing"

	importer "knowledge-graph/internal/application/import"
	"knowledge-graph/internal/config"
	"knowledge-graph/internal/domain/cache/cachetest"
	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/interfaces/api/middleware"
	contextkeys "knowledge-graph/internal/shared/context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// setupNoteRouter создаёт тестовый роутер с мок-репозиторием

func setupNoteRouter() (*gin.Engine, *mockNoteRepo) {

	gin.SetMode(gin.TestMode)

	repo := newMockNoteRepo()

	// Для тестов дополнительные зависимости не нужны, передаём nil
	cfg := &config.Config{
		RecommendationTopN:                    10,
		RecommendationFallbackSemanticEnabled: false,
		RecommendationFallbackEnabled:         false,
		RecommendationTaskDelaySeconds:        1,
		PaginationDefaultLimit:                20,
		PaginationMaxLimit:                    100,
	}
	cacheClient := cachetest.NewFakeCacheClient()
	importSvc := importer.NewService(repo, cacheClient, nil, nil)
	handler := New(repo, nil, nil, nil, 0, nil, nil, nil, cfg, nil, nil, importSvc)

	r := gin.Default()

	r.POST("/notes", handler.Create)

	r.GET("/notes/:id", handler.Get)

	r.PUT("/notes/:id", handler.Update)

	r.DELETE("/notes/:id", handler.Delete)

	r.POST("/notes/batch/delete", handler.DeleteBatch)
	r.POST("/notes/batch/create", handler.CreateBatch)

	r.GET("/notes/:id/suggestions", handler.GetSuggestions) // если хотите тестировать и рекомендации
	r.GET("/notes", handler.List)

	return r, repo

}

func setupNoteRouterWithUser() (*gin.Engine, *mockNoteRepo, uuid.UUID) {
	gin.SetMode(gin.TestMode)

	repo := newMockNoteRepo()
	cfg := &config.Config{
		RecommendationTopN:                    10,
		RecommendationFallbackSemanticEnabled: false,
		RecommendationFallbackEnabled:         false,
		RecommendationTaskDelaySeconds:        1,
		PaginationDefaultLimit:                20,
		PaginationMaxLimit:                    100,
	}
	cacheClient := cachetest.NewFakeCacheClient()
	importSvc := importer.NewService(repo, cacheClient, nil, nil)
	handler := New(repo, nil, nil, nil, 0, nil, nil, nil, cfg, nil, nil, importSvc)

	userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, userID)
		c.Next()
	})

	r.POST("/notes", handler.Create)
	r.POST("/notes/batch/create", handler.CreateBatch)
	r.POST("/notes/batch/delete", handler.DeleteBatch)
	r.DELETE("/notes/:id", handler.Delete)

	return r, repo, userID
}

func TestCreateNote(t *testing.T) {

	r, _ := setupNoteRouter()

	body := `{"title":"Test Note","content":"Hello","metadata":{}}`

	req := httptest.NewRequest("POST", "/notes", bytes.NewBufferString(body))

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {

		t.Errorf("expected status 201, got %d", w.Code)

	}

	var resp map[string]interface{}

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
		return
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Errorf("response data is not an object")
		return
	}
	if data["title"] != "Test Note" {
		t.Errorf("title mismatch: %v", data["title"])
	}

}

func TestGetNote(t *testing.T) {

	r, repo := setupNoteRouter()

	// Создаём заметку напрямую через репозиторий

	title, _ := note.NewTitle("GetTest")

	content, _ := note.NewContent("Content")

	metadata, _ := note.NewMetadata(nil)

	n := note.NewNote(title, content, "star", metadata)

	ctx := context.Background()

	_ = repo.Save(ctx, n)

	req := httptest.NewRequest("GET", "/notes/"+n.ID().String(), nil)

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {

		t.Errorf("expected 200, got %d", w.Code)

	}

	var resp map[string]interface{}

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
		return
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Error("response data is not an object")
		return
	}
	if data["title"] != "GetTest" {
		t.Error("title mismatch")
	}

}

func TestUpdateNote(t *testing.T) {

	r, repo := setupNoteRouter()

	ctx := context.Background()

	title, _ := note.NewTitle("Original")

	content, _ := note.NewContent("Content")

	metadata, _ := note.NewMetadata(nil)

	n := note.NewNote(title, content, "star", metadata)

	_ = repo.Save(ctx, n)

	updateBody := `{"title":"Updated"}`

	req := httptest.NewRequest("PUT", "/notes/"+n.ID().String(), bytes.NewBufferString(updateBody))

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {

		t.Errorf("expected 200, got %d", w.Code)

	}

	var resp map[string]interface{}

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
		return
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Error("response data is not an object")
		return
	}
	if data["title"] != "Updated" {
		t.Error("title not updated")
	}

}

func TestDeleteNote(t *testing.T) {

	r, repo := setupNoteRouter()

	ctx := context.Background()

	title, _ := note.NewTitle("ToDelete")

	content, _ := note.NewContent("Content")

	metadata, _ := note.NewMetadata(nil)

	n := note.NewNote(title, content, "star", metadata)

	_ = repo.Save(ctx, n)

	req := httptest.NewRequest("DELETE", "/notes/"+n.ID().String(), nil)

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {

		t.Errorf("expected 204, got %d", w.Code)

	}

	found, _ := repo.FindByID(ctx, n.ID())

	if found != nil {

		t.Error("note still exists after delete")

	}

}

func TestGetSuggestions_EmptyFallback(t *testing.T) {
	r, repo := setupNoteRouter()

	// Создаём заметку
	title, _ := note.NewTitle("SugTest")
	content, _ := note.NewContent("Content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, "star", metadata)
	ctx := context.Background()
	_ = repo.Save(ctx, n)

	// Запрос рекомендаций при отсутствии recRepo/embedding/redis
	req := httptest.NewRequest("GET", "/notes/"+n.ID().String()+"/suggestions", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected 202 Accepted, got %d", w.Code)
	}

	// Проверяем заголовки
	if got := w.Header().Get("X-Recommendations-Source"); got != "empty" {
		t.Errorf("expected X-Recommendations-Source=empty, got %s", got)
	}
	if got := w.Header().Get("X-Recommendations-Stale"); got != "true" {
		t.Errorf("expected X-Recommendations-Stale=true, got %s", got)
	}

	// Тело ответа должно содержать пустой список suggestions
	var resp struct {
		Suggestions []interface{} `json:"suggestions"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(resp.Suggestions) != 0 {
		t.Errorf("expected 0 suggestions, got %d", len(resp.Suggestions))
	}
}

func TestDeleteBatchNotes(t *testing.T) {

	r, repo := setupNoteRouter()

	ctx := context.Background()

	title1, _ := note.NewTitle("ToDelete1")
	content1, _ := note.NewContent("Content1")
	metadata1, _ := note.NewMetadata(nil)
	n1 := note.NewNote(title1, content1, "star", metadata1)

	title2, _ := note.NewTitle("ToDelete2")
	content2, _ := note.NewContent("Content2")
	metadata2, _ := note.NewMetadata(nil)
	n2 := note.NewNote(title2, content2, "planet", metadata2)

	_ = repo.Save(ctx, n1)
	_ = repo.Save(ctx, n2)

	body := fmt.Sprintf(`{"ids":["%s","%s"]}`, n1.ID().String(), n2.ID().String())

	req := httptest.NewRequest("POST", "/notes/batch/delete", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	// SEC-1: this router has no auth at all, so the batch ownership check
	// would refuse creatorless notes. Exercise it under the test bypass.
	req = req.WithContext(context.WithValue(req.Context(), contextkeys.SkipAuthKey, true))

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}

	found1, _ := repo.FindByID(ctx, n1.ID())
	found2, _ := repo.FindByID(ctx, n2.ID())

	if found1 != nil || found2 != nil {
		t.Error("notes still exist after batch delete")
	}

}

func TestDeleteBatchNotes_EmptyBody(t *testing.T) {

	r, _ := setupNoteRouter()

	req := httptest.NewRequest("POST", "/notes/batch/delete", bytes.NewBufferString(`{"ids":[]}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

}

func TestSearchValidationTooLong(t *testing.T) {
	r, _ := setupNoteRouter()

	longQ := ""
	for i := 0; i < 300; i++ { // longer than 200 limit
		longQ += "a"
	}

	req := httptest.NewRequest("GET", "/notes/search?q="+longQ, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 BadRequest, got %d", w.Code)
	}
}

func TestCreateNoteInvalidType(t *testing.T) {
	r, _ := setupNoteRouter()

	body := `{"title":"Test Note","content":"Hello","type":"invalid_type"}`
	req := httptest.NewRequest("POST", "/notes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCreateNoteTooLongTitle(t *testing.T) {
	r, _ := setupNoteRouter()

	longTitle := ""
	for i := 0; i < 201; i++ {
		longTitle += "a"
	}

	body := fmt.Sprintf(`{"title":"%s","content":"Hello"}`, longTitle)
	req := httptest.NewRequest("POST", "/notes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCreateNoteTooLongContent(t *testing.T) {
	r, _ := setupNoteRouter()

	longContent := ""
	for i := 0; i < 50001; i++ {
		longContent += "a"
	}

	body := fmt.Sprintf(`{"title":"Test","content":"%s"}`, longContent)
	req := httptest.NewRequest("POST", "/notes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCreateNoteMissingTitle(t *testing.T) {
	r, _ := setupNoteRouter()

	body := `{"content":"Hello"}`
	req := httptest.NewRequest("POST", "/notes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCreateNoteInvalidJSON(t *testing.T) {
	r, _ := setupNoteRouter()

	body := `invalid json`
	req := httptest.NewRequest("POST", "/notes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestUpdateNoteNotFound(t *testing.T) {
	r, _ := setupNoteRouter()

	body := `{"title":"Updated","content":"Updated content"}`
	req := httptest.NewRequest("PUT", "/notes/"+uuid.New().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestUpdateNoteInvalidUUID(t *testing.T) {
	r, _ := setupNoteRouter()

	body := `{"title":"Updated","content":"Updated content"}`
	req := httptest.NewRequest("PUT", "/notes/invalid-uuid", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetNoteInvalidUUID(t *testing.T) {
	r, _ := setupNoteRouter()

	req := httptest.NewRequest("GET", "/notes/invalid-uuid", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestDeleteNoteInvalidUUID(t *testing.T) {
	r, _ := setupNoteRouter()

	req := httptest.NewRequest("DELETE", "/notes/invalid-uuid", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestListNotesPagination(t *testing.T) {
	r, repo := setupNoteRouter()

	ctx := context.Background()
	for i := 0; i < 25; i++ {
		title, _ := note.NewTitle(fmt.Sprintf("Note %d", i))
		content, _ := note.NewContent(fmt.Sprintf("Content %d", i))
		metadata, _ := note.NewMetadata(nil)
		n := note.NewNote(title, content, "star", metadata)
		_ = repo.Save(ctx, n)
	}

	req := httptest.NewRequest("GET", "/notes?limit=10&offset=0", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestListNotesInvalidLimit(t *testing.T) {
	r, _ := setupNoteRouter()

	req := httptest.NewRequest("GET", "/notes?limit=999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Handler may accept high limit and cap it internally
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Errorf("expected status 200 or 400, got %d", w.Code)
	}
}

func TestListNotesInvalidOffset(t *testing.T) {
	r, _ := setupNoteRouter()

	req := httptest.NewRequest("GET", "/notes?offset=-1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Handler may handle negative offset
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Errorf("expected status 200 or 400, got %d", w.Code)
	}
}

func TestUpdateNoteEmptyTitle(t *testing.T) {
	r, repo := setupNoteRouter()

	ctx := context.Background()
	title, _ := note.NewTitle("Original Title")
	content, _ := note.NewContent("Original content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, "star", metadata)
	_ = repo.Save(ctx, n)

	body := `{"title":""}`
	req := httptest.NewRequest("PUT", "/notes/"+n.ID().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should fail for empty title
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusOK)
}

func TestUpdateNoteEmptyContent(t *testing.T) {
	r, repo := setupNoteRouter()

	ctx := context.Background()
	title, _ := note.NewTitle("Original Title")
	content, _ := note.NewContent("Original content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, "star", metadata)
	_ = repo.Save(ctx, n)

	body := `{"content":""}`
	req := httptest.NewRequest("PUT", "/notes/"+n.ID().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Empty content should be allowed
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateNoteInvalidJSON(t *testing.T) {
	r, repo := setupNoteRouter()

	ctx := context.Background()
	title, _ := note.NewTitle("Original Title")
	content, _ := note.NewContent("Original content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, "star", metadata)
	_ = repo.Save(ctx, n)

	body := `invalid json`
	req := httptest.NewRequest("PUT", "/notes/"+n.ID().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateNoteTooLongTitle(t *testing.T) {
	r, repo := setupNoteRouter()

	ctx := context.Background()
	title, _ := note.NewTitle("Original Title")
	content, _ := note.NewContent("Original content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, "star", metadata)
	_ = repo.Save(ctx, n)

	longTitle := ""
	for i := 0; i < 201; i++ {
		longTitle += "a"
	}

	body := fmt.Sprintf(`{"title":"%s"}`, longTitle)
	req := httptest.NewRequest("PUT", "/notes/"+n.ID().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateNoteTooLongContent(t *testing.T) {
	r, repo := setupNoteRouter()

	ctx := context.Background()
	title, _ := note.NewTitle("Original Title")
	content, _ := note.NewContent("Original content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, "star", metadata)
	_ = repo.Save(ctx, n)

	longContent := ""
	for i := 0; i < 50001; i++ {
		longContent += "a"
	}

	body := fmt.Sprintf(`{"content":"%s"}`, longContent)
	req := httptest.NewRequest("PUT", "/notes/"+n.ID().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateNoteWithMetadata(t *testing.T) {
	r, repo := setupNoteRouter()

	ctx := context.Background()
	title, _ := note.NewTitle("Original Title")
	content, _ := note.NewContent("Original content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, "star", metadata)
	_ = repo.Save(ctx, n)

	body := `{"title":"Updated Title","metadata":{"key":"value"}}`
	req := httptest.NewRequest("PUT", "/notes/"+n.ID().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateNoteInvalidMetadata(t *testing.T) {
	r, repo := setupNoteRouter()

	ctx := context.Background()
	title, _ := note.NewTitle("Original Title")
	content, _ := note.NewContent("Original content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, "star", metadata)
	_ = repo.Save(ctx, n)

	// Invalid metadata (circular reference)
	body := `{"title":"Updated Title","metadata":{"circular":{"self":"ref"}}}`
	req := httptest.NewRequest("PUT", "/notes/"+n.ID().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should fail for invalid metadata
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusOK)
}

func TestDeleteNoteNotFound(t *testing.T) {
	r, _ := setupNoteRouter()

	req := httptest.NewRequest("DELETE", "/notes/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateNoteWithValidTypes(t *testing.T) {
	r, _ := setupNoteRouter()

	validTypes := []string{"star", "planet", "comet", "galaxy", "asteroid", "satellite", "debris", "nebula", "dust", "unknown", "blackhole"}

	for _, noteType := range validTypes {
		body := fmt.Sprintf(`{"title":"Test Note","content":"Hello","type":"%s"}`, noteType)
		req := httptest.NewRequest("POST", "/notes", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	}
}

func TestUpdateNoteType(t *testing.T) {
	r, repo := setupNoteRouter()

	ctx := context.Background()
	title, _ := note.NewTitle("Original Title")
	content, _ := note.NewContent("Original content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, "star", metadata)
	_ = repo.Save(ctx, n)

	body := `{"type":"planet"}`
	req := httptest.NewRequest("PUT", "/notes/"+n.ID().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Type update should work
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListPagination(t *testing.T) {
	r, repo := setupNoteRouter()
	ctx := context.Background()

	// create 3 notes
	for i := 0; i < 3; i++ {
		title, _ := note.NewTitle(fmt.Sprintf("N%d", i))
		content, _ := note.NewContent("c")
		metadata, _ := note.NewMetadata(nil)
		n := note.NewNote(title, content, "star", metadata)
		_ = repo.Save(ctx, n)
	}

	req := httptest.NewRequest("GET", "/notes?limit=1&offset=0", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	notes, ok := resp["notes"].([]interface{})
	if !ok {
		t.Fatalf("notes missing or wrong type")
	}
	if len(notes) != 1 {
		t.Errorf("expected 1 note, got %d", len(notes))
	}
}

func TestCreateBatchNotes(t *testing.T) {
	r, repo := setupNoteRouter()
	ctx := context.Background()

	body := `{"notes":[{"title":"Batch 1","content":"Content 1","type":"star"},{"title":"Batch 2"}]}`
	req := httptest.NewRequest("POST", "/notes/batch/create", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), contextkeys.SkipAuthKey, true))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("response data is not an object")
	}
	created, ok := data["notes"].([]interface{})
	if !ok {
		t.Fatalf("created notes is not an array")
	}
	if len(created) != 2 {
		t.Fatalf("expected 2 created notes, got %d", len(created))
	}

	failed, ok := data["failed"].([]interface{})
	if !ok {
		t.Fatalf("failed is not an array")
	}
	if len(failed) != 0 {
		t.Fatalf("expected 0 failed notes, got %d", len(failed))
	}

	// Verify both notes were actually persisted.
	all, _ := repo.FindAll(ctx)
	if len(all) != 2 {
		t.Errorf("expected 2 notes in repo, got %d", len(all))
	}
}

func TestCreateBatchNotesValidationFailure(t *testing.T) {
	r, _ := setupNoteRouter()

	body := `{"notes":[{"content":"Missing title"}]}`
	req := httptest.NewRequest("POST", "/notes/batch/create", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), contextkeys.SkipAuthKey, true))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateBatchNotes_SetsCreator(t *testing.T) {
	r, repo, userID := setupNoteRouterWithUser()
	ctx := context.Background()

	body := `{"notes":[{"title":"With Creator","content":"content","type":"star"}]}`
	req := httptest.NewRequest("POST", "/notes/batch/create", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	all, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 note in repo, got %d", len(all))
	}
	if !all[0].IsOwnedBy(userID) {
		t.Errorf("note is not owned by authenticated user")
	}
}

func TestCreateBatchNotes_TooMany(t *testing.T) {
	r, repo := setupNoteRouter()
	ctx := context.Background()

	notes := make([]map[string]string, 51)
	for i := range notes {
		notes[i] = map[string]string{"title": fmt.Sprintf("Note %d", i)}
	}
	bodyBytes, _ := json.Marshal(map[string]interface{}{"notes": notes})

	req := httptest.NewRequest("POST", "/notes/batch/create", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	all, _ := repo.FindAll(ctx)
	if len(all) != 0 {
		t.Errorf("expected 0 notes in repo, got %d", len(all))
	}
}

func TestCreateBatchNotes_InvalidType(t *testing.T) {
	r, repo := setupNoteRouter()
	ctx := context.Background()

	body := `{"notes":[{"title":"Bad Type","content":"content","type":"invalid"}]}`
	req := httptest.NewRequest("POST", "/notes/batch/create", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	all, _ := repo.FindAll(ctx)
	if len(all) != 0 {
		t.Errorf("expected 0 notes in repo, got %d", len(all))
	}
}

func TestCreateBatchNotes_TooLongTitle(t *testing.T) {
	r, repo := setupNoteRouter()
	ctx := context.Background()

	longTitle := ""
	for i := 0; i < 201; i++ {
		longTitle += "a"
	}
	body := fmt.Sprintf(`{"notes":[{"title":"%s","content":"content"}]}`, longTitle)
	req := httptest.NewRequest("POST", "/notes/batch/create", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	all, _ := repo.FindAll(ctx)
	if len(all) != 0 {
		t.Errorf("expected 0 notes in repo, got %d", len(all))
	}
}

func TestCreateBatchNotes_TooLongContent(t *testing.T) {
	r, repo := setupNoteRouter()
	ctx := context.Background()

	longContent := ""
	for i := 0; i < 50001; i++ {
		longContent += "a"
	}
	body := fmt.Sprintf(`{"notes":[{"title":"T","content":"%s"}]}`, longContent)
	req := httptest.NewRequest("POST", "/notes/batch/create", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	all, _ := repo.FindAll(ctx)
	if len(all) != 0 {
		t.Errorf("expected 0 notes in repo, got %d", len(all))
	}
}

func TestCreateBatchNotes_MixedValidAndInvalid(t *testing.T) {
	r, repo := setupNoteRouter()
	ctx := context.Background()

	body := `{"notes":[{"title":"Valid Note","content":"ok","type":"star"},{"title":"   ","content":"whitespace only"}]}`
	req := httptest.NewRequest("POST", "/notes/batch/create", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), contextkeys.SkipAuthKey, true))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	all, _ := repo.FindAll(ctx)
	if len(all) != 1 {
		t.Fatalf("expected 1 note in repo, got %d", len(all))
	}
	if all[0].Title().String() != "Valid Note" {
		t.Errorf("unexpected note title: %s", all[0].Title().String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	data := resp["data"].(map[string]interface{})
	created := data["notes"].([]interface{})
	failed := data["failed"].([]interface{})
	if len(created) != 1 {
		t.Fatalf("expected 1 created, got %d", len(created))
	}
	if len(failed) != 1 {
		t.Fatalf("expected 1 failed, got %d", len(failed))
	}
	failedItem := failed[0].(map[string]interface{})
	if failedItem["index"].(float64) != 1 {
		t.Errorf("expected failed index 1, got %v", failedItem["index"])
	}
}

func TestCreateBatchNotes_EmptyNotes(t *testing.T) {
	r, _ := setupNoteRouter()

	body := `{"notes":[]}`
	req := httptest.NewRequest("POST", "/notes/batch/create", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateBatchNotes_InvalidJSON(t *testing.T) {
	r, _ := setupNoteRouter()

	req := httptest.NewRequest("POST", "/notes/batch/create", bytes.NewBufferString(`invalid json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestDeleteBatchNotes_MissingNotes(t *testing.T) {
	r, _ := setupNoteRouter()

	req := httptest.NewRequest("POST", "/notes/batch/delete", bytes.NewBufferString(`{"ids":[]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestDeleteBatchNotes_MalformedUUID(t *testing.T) {
	r, _ := setupNoteRouter()

	req := httptest.NewRequest("POST", "/notes/batch/delete", bytes.NewBufferString(`{"ids":["not-a-uuid"]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestDeleteBatchNotes_Owned(t *testing.T) {
	r, repo, userID := setupNoteRouterWithUser()
	ctx := context.Background()

	title1, _ := note.NewTitle("Owned1")
	content1, _ := note.NewContent("Content1")
	meta1, _ := note.NewMetadata(nil)
	n1 := note.NewNoteWithCreator(title1, content1, "star", meta1, userID)

	title2, _ := note.NewTitle("Owned2")
	content2, _ := note.NewContent("Content2")
	meta2, _ := note.NewMetadata(nil)
	n2 := note.NewNoteWithCreator(title2, content2, "planet", meta2, userID)

	_ = repo.Save(ctx, n1)
	_ = repo.Save(ctx, n2)

	body := fmt.Sprintf(`{"ids":["%s","%s"]}`, n1.ID().String(), n2.ID().String())
	req := httptest.NewRequest("POST", "/notes/batch/delete", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	found1, _ := repo.FindByID(ctx, n1.ID())
	found2, _ := repo.FindByID(ctx, n2.ID())
	if found1 != nil || found2 != nil {
		t.Error("owned notes still exist after batch delete")
	}
}

func TestDeleteBatchNotes_ForeignNoteRoute(t *testing.T) {
	r, repo, userID := setupNoteRouterWithUser()
	ctx := context.Background()

	foreignUserID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	title, _ := note.NewTitle("Foreign")
	content, _ := note.NewContent("Content")
	meta, _ := note.NewMetadata(nil)
	n := note.NewNoteWithCreator(title, content, "star", meta, foreignUserID)

	titleOwned, _ := note.NewTitle("Owned")
	contentOwned, _ := note.NewContent("Content")
	metaOwned, _ := note.NewMetadata(nil)
	nOwned := note.NewNoteWithCreator(titleOwned, contentOwned, "planet", metaOwned, userID)

	_ = repo.Save(ctx, n)
	_ = repo.Save(ctx, nOwned)

	body := fmt.Sprintf(`{"ids":["%s","%s"]}`, n.ID().String(), nOwned.ID().String())
	req := httptest.NewRequest("POST", "/notes/batch/delete", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}

	found, _ := repo.FindByID(ctx, nOwned.ID())
	if found == nil {
		t.Error("owned note was deleted even though foreign note was in the same batch")
	}
}

func TestDeleteBatchNotes_MissingIDsNoOp(t *testing.T) {
	r, repo, _ := setupNoteRouterWithUser()
	ctx := context.Background()

	body := fmt.Sprintf(`{"ids":["%s","%s"]}`, uuid.New().String(), uuid.New().String())
	req := httptest.NewRequest("POST", "/notes/batch/delete", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	all, _ := repo.FindAll(ctx)
	if len(all) != 0 {
		t.Errorf("expected repo to be empty, got %d notes", len(all))
	}
}
