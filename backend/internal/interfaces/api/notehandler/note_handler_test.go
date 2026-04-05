package notehandler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"knowledge-graph/internal/domain/note"

	"github.com/gin-gonic/gin"
)

// setupNoteRouter создаёт тестовый роутер с мок-репозиторием
func setupNoteRouter() (*gin.Engine, *mockNoteRepo) {
	gin.SetMode(gin.TestMode)
	repo := newMockNoteRepo()
	// Для тестов taskQueue и suggestionsHandler не нужны, передаём nil
	handler := New(repo, nil, nil)
	r := gin.Default()
	r.POST("/notes", handler.Create)
	r.GET("/notes/:id", handler.Get)
	r.PUT("/notes/:id", handler.Update)
	r.DELETE("/notes/:id", handler.Delete)
	r.GET("/notes/:id/suggestions", handler.GetSuggestions) // если хотите тестировать и рекомендации
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
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["title"] != "Test Note" {
		t.Errorf("title mismatch: %v", resp["title"])
	}
}

func TestGetNote(t *testing.T) {
	r, repo := setupNoteRouter()
	// Создаём заметку напрямую через репозиторий
	title, _ := note.NewTitle("GetTest")
	content, _ := note.NewContent("Content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, metadata)
	ctx := context.Background()
	_ = repo.Save(ctx, n)

	req := httptest.NewRequest("GET", "/notes/"+n.ID().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["title"] != "GetTest" {
		t.Error("title mismatch")
	}
}

func TestUpdateNote(t *testing.T) {
	r, repo := setupNoteRouter()
	ctx := context.Background()
	title, _ := note.NewTitle("Original")
	content, _ := note.NewContent("Content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, metadata)
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
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["title"] != "Updated" {
		t.Error("title not updated")
	}
}

func TestDeleteNote(t *testing.T) {
	r, repo := setupNoteRouter()
	ctx := context.Background()
	title, _ := note.NewTitle("ToDelete")
	content, _ := note.NewContent("Content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, metadata)
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
