package linkhandler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"knowledge-graph/internal/domain/note"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// mockNoteRepo для linkHandler (упрощённый, только FindByID)
type mockNoteRepoForLink struct {
	notes map[uuid.UUID]*note.Note
}

func newMockNoteRepoForLink() *mockNoteRepoForLink {
	return &mockNoteRepoForLink{
		notes: make(map[uuid.UUID]*note.Note),
	}
}

func (m *mockNoteRepoForLink) Save(ctx context.Context, n *note.Note) error { return nil }
func (m *mockNoteRepoForLink) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	n, ok := m.notes[id]
	if !ok {
		return nil, nil
	}
	return n, nil
}
func (m *mockNoteRepoForLink) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockNoteRepoForLink) List(ctx context.Context, limit, offset int) ([]*note.Note, int64, error) {
	var allNotes []*note.Note
	for _, n := range m.notes {
		allNotes = append(allNotes, n)
	}

	total := int64(len(allNotes))

	if offset >= len(allNotes) {
		return []*note.Note{}, total, nil
	}

	end := offset + limit
	if end > len(allNotes) {
		end = len(allNotes)
	}

	return allNotes[offset:end], total, nil
}
func (m *mockNoteRepoForLink) Search(ctx context.Context, query string, limit, offset int) ([]*note.Note, int64, error) {
	var results []*note.Note
	for _, n := range m.notes {
		// Simple string matching for mock
		if len(query) == 0 ||
			strings.Contains(strings.ToLower(n.Title().String()), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(n.Content().String()), strings.ToLower(query)) {
			results = append(results, n)
		}
	}

	// Apply pagination
	total := int64(len(results))
	if offset >= len(results) {
		return []*note.Note{}, total, nil
	}

	end := offset + limit
	if end > len(results) {
		end = len(results)
	}

	return results[offset:end], total, nil
}

func setupLinkRouter() (*gin.Engine, *mockLinkRepo, *mockNoteRepoForLink) {
	gin.SetMode(gin.TestMode)
	linkRepo := newMockLinkRepo()
	noteRepo := newMockNoteRepoForLink()
	handler := New(linkRepo, noteRepo)
	r := gin.Default()
	r.POST("/links", handler.Create)
	r.GET("/links/:id", handler.Get)
	r.GET("/notes/:id/links", handler.GetByNote)
	r.DELETE("/links/:id", handler.Delete)
	r.DELETE("/notes/:id/links", handler.DeleteByNote)
	return r, linkRepo, noteRepo
}

func TestCreateLink(t *testing.T) {
	r, linkRepo, noteRepo := setupLinkRouter()

	sourceID := uuid.New()
	targetID := uuid.New()

	// Создаём заметки через доменные конструкторы, затем восстанавливаем с нужными ID
	title1, _ := note.NewTitle("Source Note")
	content1, _ := note.NewContent("Source content")
	metadata1, _ := note.NewMetadata(nil)
	sourceNote := note.NewNote(title1, content1, metadata1)
	sourceNote = note.ReconstructNote(sourceID, title1, content1, metadata1, sourceNote.CreatedAt(), sourceNote.UpdatedAt())

	title2, _ := note.NewTitle("Target Note")
	content2, _ := note.NewContent("Target content")
	metadata2, _ := note.NewMetadata(nil)
	targetNote := note.NewNote(title2, content2, metadata2)
	targetNote = note.ReconstructNote(targetID, title2, content2, metadata2, targetNote.CreatedAt(), targetNote.UpdatedAt())

	noteRepo.notes[sourceID] = sourceNote
	noteRepo.notes[targetID] = targetNote

	body := map[string]interface{}{
		"source_note_id": sourceID.String(),
		"target_note_id": targetID.String(),
		"link_type":      "reference",
		"weight":         0.8,
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}

	// Проверяем, что связь сохранилась
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	linkIDStr, ok := resp["id"].(string)
	if !ok {
		t.Fatal("no id in response")
	}
	linkID, _ := uuid.Parse(linkIDStr)

	saved, err := linkRepo.FindByID(context.Background(), linkID)
	if err != nil {
		t.Fatalf("failed to find saved link: %v", err)
	}
	if saved == nil {
		t.Fatal("link not saved")
	}
	if saved.SourceNoteID() != sourceID {
		t.Error("source note id mismatch")
	}
	if saved.TargetNoteID() != targetID {
		t.Error("target note id mismatch")
	}
}
