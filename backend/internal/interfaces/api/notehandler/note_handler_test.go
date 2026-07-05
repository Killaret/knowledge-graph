package notehandler

import (
	"bytes"

	"context"

	"encoding/json"

	"fmt"
	"net/http"

	"net/http/httptest"

	"testing"

	"knowledge-graph/internal/config"
	"knowledge-graph/internal/domain/note"

	"github.com/gin-gonic/gin"
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
	handler := New(repo, nil, nil, nil, 0, nil, nil, nil, cfg, nil, nil)

	r := gin.Default()

	r.POST("/notes", handler.Create)

	r.GET("/notes/:id", handler.Get)

	r.PUT("/notes/:id", handler.Update)

	r.DELETE("/notes/:id", handler.Delete)

	r.POST("/notes/batch", handler.DeleteBatch)

	r.GET("/notes/:id/suggestions", handler.GetSuggestions) // если хотите тестировать и рекомендации
	r.GET("/notes", handler.List)

	return r, repo

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

	req := httptest.NewRequest("POST", "/notes/batch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

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

	req := httptest.NewRequest("POST", "/notes/batch", bytes.NewBufferString(`{"ids":[]}`))
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
