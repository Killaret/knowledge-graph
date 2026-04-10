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

func TestGetAllNotes(t *testing.T) {
	r, repo := setupNoteRouter()
	ctx := context.Background()

	// Create multiple notes
	title1, _ := note.NewTitle("Note1")
	content1, _ := note.NewContent("Content1")
	metadata1, _ := note.NewMetadata(nil)
	n1 := note.NewNote(title1, content1, metadata1)
	_ = repo.Save(ctx, n1)

	title2, _ := note.NewTitle("Note2")
	content2, _ := note.NewContent("Content2")
	metadata2, _ := note.NewMetadata(nil)
	n2 := note.NewNote(title2, content2, metadata2)
	_ = repo.Save(ctx, n2)

	notes, err := repo.FindAll(ctx)
	if err != nil {
		t.Errorf("FindAll failed: %v", err)
	}
	if len(notes) != 2 {
		t.Errorf("expected 2 notes, got %d", len(notes))
	}
}

func TestSearchNotes(t *testing.T) {
	r, repo := setupNoteRouter()
	ctx := context.Background()

	title, _ := note.NewTitle("SearchTest")
	content, _ := note.NewContent("This is a searchable content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, metadata)
	_ = repo.Save(ctx, n)

	// Test search by title
	results, err := repo.Search(ctx, "Search")
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	// Test search by content
	results, err = repo.Search(ctx, "searchable")
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	// Test search with no results
	results, err = repo.Search(ctx, "nonexistent")
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestFindByKeywords(t *testing.T) {
	r, repo := setupNoteRouter()
	ctx := context.Background()

	title, _ := note.NewTitle("KeywordTest")
	content, _ := note.NewContent("This contains keywords like test and example")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, metadata)
	_ = repo.Save(ctx, n)

	keywords := []string{"test", "example"}
	results, err := repo.FindByKeywords(ctx, keywords)
	if err != nil {
		t.Errorf("FindByKeywords failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	// Test with no matching keywords
	keywords = []string{"nonexistent"}
	results, err = repo.FindByKeywords(ctx, keywords)
	if err != nil {
		t.Errorf("FindByKeywords failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestUpdateNote(t *testing.T) {
	r, repo := setupNoteRouter()
	ctx := context.Background()

	title, _ := note.NewTitle("Original")
	content, _ := note.NewContent("Original content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, metadata)
	_ = repo.Save(ctx, n)

	// Update note
	newTitle, _ := note.NewTitle("Updated")
	err := n.UpdateTitle(newTitle)
	if err != nil {
		t.Errorf("UpdateTitle failed: %v", err)
	}

	err = repo.Update(ctx, n)
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	// Verify update
	updated, err := repo.FindByID(ctx, n.ID())
	if err != nil {
		t.Errorf("FindByID failed: %v", err)
	}
	if updated.Title().String() != "Updated" {
		t.Errorf("title not updated, expected 'Updated', got '%s'", updated.Title().String())
	}
}

func TestUpdateNonExistentNote(t *testing.T) {
	repo := newMockNoteRepo()
	ctx := context.Background()

	title, _ := note.NewTitle("NonExistent")
	content, _ := note.NewContent("Content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, metadata)

	err := repo.Update(ctx, n)
	if err == nil {
		t.Error("expected error when updating non-existent note")
	}
}
