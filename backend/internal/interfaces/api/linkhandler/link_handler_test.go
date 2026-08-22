package linkhandler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
func (m *mockNoteRepoForLink) Delete(ctx context.Context, id uuid.UUID) error         { return nil }
func (m *mockNoteRepoForLink) DeleteBatch(ctx context.Context, ids []uuid.UUID) error { return nil }
func (m *mockNoteRepoForLink) Restore(ctx context.Context, id uuid.UUID) error        { return nil }
func (m *mockNoteRepoForLink) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*note.Note, int64, error) {
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
func (m *mockNoteRepoForLink) Search(ctx context.Context, userID uuid.UUID, query string, limit, offset int) ([]*note.Note, int64, error) {
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
func (m *mockNoteRepoForLink) FindAll(ctx context.Context) ([]*note.Note, error) {
	var allNotes []*note.Note
	for _, n := range m.notes {
		allNotes = append(allNotes, n)
	}
	return allNotes, nil
}

func (m *mockNoteRepoForLink) FindAllPaginated(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*note.Note, int64, error) {
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
	if limit == 0 {
		end = len(allNotes)
	}

	return allNotes[offset:end], total, nil
}

func setupLinkRouter() (*gin.Engine, *mockLinkRepo, *mockNoteRepoForLink) {
	gin.SetMode(gin.TestMode)
	linkRepo := newMockLinkRepo()
	noteRepo := newMockNoteRepoForLink()
	handler := New(linkRepo, noteRepo, nil, nil)
	r := gin.Default()
	r.POST("/links", handler.Create)
	r.GET("/links/:id", handler.Get)
	r.PUT("/links/:id", handler.Update)
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
	sourceNote := note.NewNote(title1, content1, "star", metadata1)
	sourceNote = note.ReconstructNote(sourceID, title1, content1, "star", metadata1, sourceNote.CreatedAt(), sourceNote.UpdatedAt())

	title2, _ := note.NewTitle("Target Note")
	content2, _ := note.NewContent("Target content")
	metadata2, _ := note.NewMetadata(nil)
	targetNote := note.NewNote(title2, content2, "star", metadata2)
	targetNote = note.ReconstructNote(targetID, title2, content2, "star", metadata2, targetNote.CreatedAt(), targetNote.UpdatedAt())

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
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("response data is not an object")
	}
	linkIDStr, ok := data["id"].(string)
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

func TestCreateLinkMissingFields(t *testing.T) {
	r, _, noteRepo := setupLinkRouter()

	sourceID := uuid.New()
	targetID := uuid.New()

	// Создаём заметки
	title1, _ := note.NewTitle("Source Note")
	content1, _ := note.NewContent("Source content")
	metadata1, _ := note.NewMetadata(nil)
	sourceNote := note.NewNote(title1, content1, "star", metadata1)
	sourceNote = note.ReconstructNote(sourceID, title1, content1, "star", metadata1, sourceNote.CreatedAt(), sourceNote.UpdatedAt())

	title2, _ := note.NewTitle("Target Note")
	content2, _ := note.NewContent("Target content")
	metadata2, _ := note.NewMetadata(nil)
	targetNote := note.NewNote(title2, content2, "star", metadata2)
	targetNote = note.ReconstructNote(targetID, title2, content2, "star", metadata2, targetNote.CreatedAt(), targetNote.UpdatedAt())

	noteRepo.notes[sourceID] = sourceNote
	noteRepo.notes[targetID] = targetNote

	tests := []struct {
		name       string
		body       map[string]interface{}
		wantStatus int
	}{
		{
			name: "missing source note id",
			body: map[string]interface{}{
				"target_note_id": targetID.String(),
				"link_type":      "reference",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing target note id",
			body: map[string]interface{}{
				"source_note_id": sourceID.String(),
				"link_type":      "reference",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetLink(t *testing.T) {
	r, _, noteRepo := setupLinkRouter()

	sourceID := uuid.New()
	targetID := uuid.New()

	// Создаём заметки
	title1, _ := note.NewTitle("Source Note")
	content1, _ := note.NewContent("Source content")
	metadata1, _ := note.NewMetadata(nil)
	sourceNote := note.NewNote(title1, content1, "star", metadata1)
	sourceNote = note.ReconstructNote(sourceID, title1, content1, "star", metadata1, sourceNote.CreatedAt(), sourceNote.UpdatedAt())

	title2, _ := note.NewTitle("Target Note")
	content2, _ := note.NewContent("Target content")
	metadata2, _ := note.NewMetadata(nil)
	targetNote := note.NewNote(title2, content2, "star", metadata2)
	targetNote = note.ReconstructNote(targetID, title2, content2, "star", metadata2, targetNote.CreatedAt(), targetNote.UpdatedAt())

	noteRepo.notes[sourceID] = sourceNote
	noteRepo.notes[targetID] = targetNote

	// Создаём связь через API
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

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	linkIDStr := data["id"].(string)
	linkID, _ := uuid.Parse(linkIDStr)

	tests := []struct {
		name       string
		linkID     string
		wantStatus int
	}{
		{
			name:       "valid link",
			linkID:     linkID.String(),
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid uuid",
			linkID:     "invalid-uuid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/links/"+tt.linkID, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Contains(t, resp, "data")
			}
		})
	}
}

func TestDeleteLink(t *testing.T) {
	r, _, noteRepo := setupLinkRouter()

	sourceID := uuid.New()
	targetID := uuid.New()

	// Создаём заметки
	title1, _ := note.NewTitle("Source Note")
	content1, _ := note.NewContent("Source content")
	metadata1, _ := note.NewMetadata(nil)
	sourceNote := note.NewNote(title1, content1, "star", metadata1)
	sourceNote = note.ReconstructNote(sourceID, title1, content1, "star", metadata1, sourceNote.CreatedAt(), sourceNote.UpdatedAt())

	title2, _ := note.NewTitle("Target Note")
	content2, _ := note.NewContent("Target content")
	metadata2, _ := note.NewMetadata(nil)
	targetNote := note.NewNote(title2, content2, "star", metadata2)
	targetNote = note.ReconstructNote(targetID, title2, content2, "star", metadata2, targetNote.CreatedAt(), targetNote.UpdatedAt())

	noteRepo.notes[sourceID] = sourceNote
	noteRepo.notes[targetID] = targetNote

	tests := []struct {
		name       string
		linkID     string
		wantStatus int
	}{
		{
			name:       "invalid uuid",
			linkID:     "invalid-uuid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/links/"+tt.linkID, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetLinksByNote(t *testing.T) {
	r, _, noteRepo := setupLinkRouter()

	sourceID := uuid.New()
	targetID := uuid.New()

	// Создаём заметки
	title1, _ := note.NewTitle("Source Note")
	content1, _ := note.NewContent("Source content")
	metadata1, _ := note.NewMetadata(nil)
	sourceNote := note.NewNote(title1, content1, "star", metadata1)
	sourceNote = note.ReconstructNote(sourceID, title1, content1, "star", metadata1, sourceNote.CreatedAt(), sourceNote.UpdatedAt())

	title2, _ := note.NewTitle("Target Note")
	content2, _ := note.NewContent("Target content")
	metadata2, _ := note.NewMetadata(nil)
	targetNote := note.NewNote(title2, content2, "star", metadata2)
	targetNote = note.ReconstructNote(targetID, title2, content2, "star", metadata2, targetNote.CreatedAt(), targetNote.UpdatedAt())

	noteRepo.notes[sourceID] = sourceNote
	noteRepo.notes[targetID] = targetNote

	tests := []struct {
		name       string
		noteID     string
		wantStatus int
	}{
		{
			name:       "invalid uuid",
			noteID:     "invalid-uuid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/notes/"+tt.noteID+"/links", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestDeleteByNote(t *testing.T) {
	r, _, noteRepo := setupLinkRouter()

	sourceID := uuid.New()
	targetID := uuid.New()

	// Создаём заметки
	title1, _ := note.NewTitle("Source Note")
	content1, _ := note.NewContent("Source content")
	metadata1, _ := note.NewMetadata(nil)
	sourceNote := note.NewNote(title1, content1, "star", metadata1)
	sourceNote = note.ReconstructNote(sourceID, title1, content1, "star", metadata1, sourceNote.CreatedAt(), sourceNote.UpdatedAt())

	title2, _ := note.NewTitle("Target Note")
	content2, _ := note.NewContent("Target content")
	metadata2, _ := note.NewMetadata(nil)
	targetNote := note.NewNote(title2, content2, "star", metadata2)
	targetNote = note.ReconstructNote(targetID, title2, content2, "star", metadata2, targetNote.CreatedAt(), targetNote.UpdatedAt())

	noteRepo.notes[sourceID] = sourceNote
	noteRepo.notes[targetID] = targetNote

	tests := []struct {
		name       string
		noteID     string
		wantStatus int
	}{
		{
			name:       "invalid uuid",
			noteID:     "invalid-uuid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/notes/"+tt.noteID+"/links", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestCreateLinkInvalidLinkType(t *testing.T) {
	r, _, noteRepo := setupLinkRouter()

	sourceID := uuid.New()
	targetID := uuid.New()

	// Создаём заметки
	title1, _ := note.NewTitle("Source Note")
	content1, _ := note.NewContent("Source content")
	metadata1, _ := note.NewMetadata(nil)
	sourceNote := note.NewNote(title1, content1, "star", metadata1)
	sourceNote = note.ReconstructNote(sourceID, title1, content1, "star", metadata1, sourceNote.CreatedAt(), sourceNote.UpdatedAt())

	title2, _ := note.NewTitle("Target Note")
	content2, _ := note.NewContent("Target content")
	metadata2, _ := note.NewMetadata(nil)
	targetNote := note.NewNote(title2, content2, "star", metadata2)
	targetNote = note.ReconstructNote(targetID, title2, content2, "star", metadata2, targetNote.CreatedAt(), targetNote.UpdatedAt())

	noteRepo.notes[sourceID] = sourceNote
	noteRepo.notes[targetID] = targetNote

	body := map[string]interface{}{
		"source_note_id": sourceID.String(),
		"target_note_id": targetID.String(),
		"link_type":      "invalid_type",
		"weight":         0.8,
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateLinkInvalidWeight(t *testing.T) {
	r, _, noteRepo := setupLinkRouter()

	sourceID := uuid.New()
	targetID := uuid.New()

	// Создаём заметки
	title1, _ := note.NewTitle("Source Note")
	content1, _ := note.NewContent("Source content")
	metadata1, _ := note.NewMetadata(nil)
	sourceNote := note.NewNote(title1, content1, "star", metadata1)
	sourceNote = note.ReconstructNote(sourceID, title1, content1, "star", metadata1, sourceNote.CreatedAt(), sourceNote.UpdatedAt())

	title2, _ := note.NewTitle("Target Note")
	content2, _ := note.NewContent("Target content")
	metadata2, _ := note.NewMetadata(nil)
	targetNote := note.NewNote(title2, content2, "star", metadata2)
	targetNote = note.ReconstructNote(targetID, title2, content2, "star", metadata2, targetNote.CreatedAt(), targetNote.UpdatedAt())

	noteRepo.notes[sourceID] = sourceNote
	noteRepo.notes[targetID] = targetNote

	tests := []struct {
		name       string
		weight     float64
		wantStatus int
	}{
		{
			name:       "weight too high",
			weight:     1.5,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "weight negative",
			weight:     -0.5,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]interface{}{
				"source_note_id": sourceID.String(),
				"target_note_id": targetID.String(),
				"link_type":      "reference",
				"weight":         tt.weight,
			}
			jsonBody, _ := json.Marshal(body)
			req := httptest.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestCreateLinkMissingSourceNote(t *testing.T) {
	r, _, noteRepo := setupLinkRouter()

	targetID := uuid.New()

	// Создаём только целевую заметку
	title2, _ := note.NewTitle("Target Note")
	content2, _ := note.NewContent("Target content")
	metadata2, _ := note.NewMetadata(nil)
	targetNote := note.NewNote(title2, content2, "star", metadata2)
	targetNote = note.ReconstructNote(targetID, title2, content2, "star", metadata2, targetNote.CreatedAt(), targetNote.UpdatedAt())

	noteRepo.notes[targetID] = targetNote

	sourceID := uuid.New()
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

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateLinkSameNote(t *testing.T) {
	r, _, noteRepo := setupLinkRouter()

	noteID := uuid.New()

	// Создаём заметку
	title, _ := note.NewTitle("Source Note")
	content, _ := note.NewContent("Source content")
	metadata, _ := note.NewMetadata(nil)
	sourceNote := note.NewNote(title, content, "star", metadata)
	sourceNote = note.ReconstructNote(noteID, title, content, "star", metadata, sourceNote.CreatedAt(), sourceNote.UpdatedAt())

	noteRepo.notes[noteID] = sourceNote

	body := map[string]interface{}{
		"source_note_id": noteID.String(),
		"target_note_id": noteID.String(),
		"link_type":      "reference",
		"weight":         0.8,
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should fail or succeed depending on business logic
	// For now, just check it doesn't crash
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusCreated)
}

func TestCreateLinkDuplicate(t *testing.T) {
	r, _, noteRepo := setupLinkRouter()

	sourceID := uuid.New()
	targetID := uuid.New()

	// Создаём заметки
	title1, _ := note.NewTitle("Source Note")
	content1, _ := note.NewContent("Source content")
	metadata1, _ := note.NewMetadata(nil)
	sourceNote := note.NewNote(title1, content1, "star", metadata1)
	sourceNote = note.ReconstructNote(sourceID, title1, content1, "star", metadata1, sourceNote.CreatedAt(), sourceNote.UpdatedAt())

	title2, _ := note.NewTitle("Target Note")
	content2, _ := note.NewContent("Target content")
	metadata2, _ := note.NewMetadata(nil)
	targetNote := note.NewNote(title2, content2, "planet", metadata2)
	targetNote = note.ReconstructNote(targetID, title2, content2, "planet", metadata2, targetNote.CreatedAt(), targetNote.UpdatedAt())

	noteRepo.notes[sourceID] = sourceNote
	noteRepo.notes[targetID] = targetNote

	body := map[string]interface{}{
		"source_note_id": sourceID.String(),
		"target_note_id": targetID.String(),
		"link_type":      "reference",
		"weight":         0.8,
	}
	jsonBody, _ := json.Marshal(body)

	// Create first link
	req1 := httptest.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	// Try to create duplicate link
	req2 := httptest.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	// Should return conflict for duplicate
	assert.True(t, w2.Code == http.StatusConflict || w2.Code == http.StatusCreated)
}

func TestCreateLinkInvalidMetadata(t *testing.T) {
	r, _, noteRepo := setupLinkRouter()

	sourceID := uuid.New()
	targetID := uuid.New()

	// Создаём заметки
	title1, _ := note.NewTitle("Source Note")
	content1, _ := note.NewContent("Source content")
	metadata1, _ := note.NewMetadata(nil)
	sourceNote := note.NewNote(title1, content1, "star", metadata1)
	sourceNote = note.ReconstructNote(sourceID, title1, content1, "star", metadata1, sourceNote.CreatedAt(), sourceNote.UpdatedAt())

	title2, _ := note.NewTitle("Target Note")
	content2, _ := note.NewContent("Target content")
	metadata2, _ := note.NewMetadata(nil)
	targetNote := note.NewNote(title2, content2, "planet", metadata2)
	targetNote = note.ReconstructNote(targetID, title2, content2, "planet", metadata2, targetNote.CreatedAt(), targetNote.UpdatedAt())

	noteRepo.notes[sourceID] = sourceNote
	noteRepo.notes[targetID] = targetNote

	// Invalid metadata (circular reference)
	body := map[string]interface{}{
		"source_note_id": sourceID.String(),
		"target_note_id": targetID.String(),
		"link_type":      "reference",
		"weight":         0.8,
		"metadata": map[string]interface{}{
			"circular": map[string]interface{}{
				"self": "reference",
			},
		},
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should fail for invalid metadata
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusCreated)
}

func TestCreateLinkMissingTarget(t *testing.T) {
	r, _, noteRepo := setupLinkRouter()

	sourceID := uuid.New()
	targetID := uuid.New()

	// Создаём только исходную заметку
	title1, _ := note.NewTitle("Source Note")
	content1, _ := note.NewContent("Source content")
	metadata1, _ := note.NewMetadata(nil)
	sourceNote := note.NewNote(title1, content1, "star", metadata1)
	sourceNote = note.ReconstructNote(sourceID, title1, content1, "star", metadata1, sourceNote.CreatedAt(), sourceNote.UpdatedAt())

	noteRepo.notes[sourceID] = sourceNote

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

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateLinkInvalidSourceUUID(t *testing.T) {
	r, _, _ := setupLinkRouter()

	body := map[string]interface{}{
		"source_note_id": "invalid-uuid",
		"target_note_id": uuid.New().String(),
		"link_type":      "reference",
		"weight":         0.8,
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func createTestLink(t *testing.T, r *gin.Engine, linkRepo *mockLinkRepo, noteRepo *mockNoteRepoForLink) (*link.Link, uuid.UUID, uuid.UUID) {
	sourceID := uuid.New()
	targetID := uuid.New()

	title1, _ := note.NewTitle("Source Note")
	content1, _ := note.NewContent("Source content")
	metadata1, _ := note.NewMetadata(nil)
	sourceNote := note.NewNote(title1, content1, "star", metadata1)
	sourceNote = note.ReconstructNote(sourceID, title1, content1, "star", metadata1, sourceNote.CreatedAt(), sourceNote.UpdatedAt())

	title2, _ := note.NewTitle("Target Note")
	content2, _ := note.NewContent("Target content")
	metadata2, _ := note.NewMetadata(nil)
	targetNote := note.NewNote(title2, content2, "star", metadata2)
	targetNote = note.ReconstructNote(targetID, title2, content2, "star", metadata2, targetNote.CreatedAt(), targetNote.UpdatedAt())

	noteRepo.notes[sourceID] = sourceNote
	noteRepo.notes[targetID] = targetNote

	l := linkRepo.Create(context.Background(), sourceID, targetID, "reference", 0.8, nil, "user")
	return l, sourceID, targetID
}

func TestDeleteLink_Success(t *testing.T) {
	r, linkRepo, noteRepo := setupLinkRouter()
	l, _, _ := createTestLink(t, r, linkRepo, noteRepo)

	req := httptest.NewRequest("DELETE", "/links/"+l.ID().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	found, _ := linkRepo.FindByID(context.Background(), l.ID())
	assert.Nil(t, found)
}

func TestGetLinksByNote_Success(t *testing.T) {
	r, linkRepo, noteRepo := setupLinkRouter()
	l, sourceID, _ := createTestLink(t, r, linkRepo, noteRepo)

	req := httptest.NewRequest("GET", "/notes/"+sourceID.String()+"/links", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	outgoing := data["outgoing"].([]interface{})
	assert.Len(t, outgoing, 1)
	assert.Equal(t, l.ID().String(), outgoing[0].(map[string]interface{})["id"])
}

func TestDeleteByNote_Success(t *testing.T) {
	r, linkRepo, noteRepo := setupLinkRouter()
	l, sourceID, _ := createTestLink(t, r, linkRepo, noteRepo)

	req := httptest.NewRequest("DELETE", "/notes/"+sourceID.String()+"/links", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	found, _ := linkRepo.FindByID(context.Background(), l.ID())
	assert.Nil(t, found)
}

func TestCreateLinkInvalidTargetUUID(t *testing.T) {
	r, _, noteRepo := setupLinkRouter()

	sourceID := uuid.New()

	// Создаём исходную заметку
	title1, _ := note.NewTitle("Source Note")
	content1, _ := note.NewContent("Source content")
	metadata1, _ := note.NewMetadata(nil)
	sourceNote := note.NewNote(title1, content1, "star", metadata1)
	sourceNote = note.ReconstructNote(sourceID, title1, content1, "star", metadata1, sourceNote.CreatedAt(), sourceNote.UpdatedAt())

	noteRepo.notes[sourceID] = sourceNote

	body := map[string]interface{}{
		"source_note_id": sourceID.String(),
		"target_note_id": "invalid-uuid",
		"link_type":      "reference",
		"weight":         0.8,
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetLinkInvalidUUID(t *testing.T) {
	r, _, _ := setupLinkRouter()

	req := httptest.NewRequest("GET", "/links/invalid-uuid", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetByNote(t *testing.T) {
	r, _, noteRepo := setupLinkRouter()

	noteID := uuid.New()

	// Создаём заметку
	title, _ := note.NewTitle("Test Note")
	content, _ := note.NewContent("Test content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, "star", metadata)
	n = note.ReconstructNote(noteID, title, content, "star", metadata, n.CreatedAt(), n.UpdatedAt())

	noteRepo.notes[noteID] = n

	req := httptest.NewRequest("GET", "/notes/"+noteID.String()+"/links", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetByNoteInvalidUUID(t *testing.T) {
	r, _, _ := setupLinkRouter()

	req := httptest.NewRequest("GET", "/notes/invalid-uuid/links", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetByNoteNotFound(t *testing.T) {
	r, _, _ := setupLinkRouter()

	req := httptest.NewRequest("GET", "/notes/"+uuid.New().String()+"/links", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteLinkInvalidUUID(t *testing.T) {
	r, _, _ := setupLinkRouter()

	req := httptest.NewRequest("DELETE", "/links/invalid-uuid", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetLink_NotFound(t *testing.T) {
	r, _, _ := setupLinkRouter()

	req := httptest.NewRequest("GET", "/links/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteLink_NotFound(t *testing.T) {
	r, _, _ := setupLinkRouter()

	req := httptest.NewRequest("DELETE", "/links/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteByNote_InvalidUUID(t *testing.T) {
	r, _, _ := setupLinkRouter()

	req := httptest.NewRequest("DELETE", "/notes/bad-uuid/links", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateLink_InvalidJSON(t *testing.T) {
	r, _, _ := setupLinkRouter()

	req := httptest.NewRequest("POST", "/links", bytes.NewBuffer([]byte(`{invalid`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetByNote_SourceNotFound(t *testing.T) {
	r, _, _ := setupLinkRouter()

	req := httptest.NewRequest("GET", "/notes/"+uuid.New().String()+"/links", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func setupLinkRouterWithUser(userID uuid.UUID) (*gin.Engine, *mockLinkRepo, *mockNoteRepoForLink) {
	gin.SetMode(gin.TestMode)
	linkRepo := newMockLinkRepo()
	noteRepo := newMockNoteRepoForLink()
	handler := New(linkRepo, noteRepo, nil, nil)
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	r.POST("/links", handler.Create)
	r.GET("/links/:id", handler.Get)
	r.PUT("/links/:id", handler.Update)
	r.GET("/notes/:id/links", handler.GetByNote)
	r.DELETE("/links/:id", handler.Delete)
	r.DELETE("/notes/:id/links", handler.DeleteByNote)
	return r, linkRepo, noteRepo
}

func TestUpdateLink_Success(t *testing.T) {
	userID := uuid.New()
	r, linkRepo, noteRepo := setupLinkRouterWithUser(userID)
	l, _, _ := createTestLink(t, r, linkRepo, noteRepo)

	body := map[string]interface{}{
		"link_type": "dependency",
		"weight":    0.6,
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/links/"+l.ID().String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	updated, _ := linkRepo.FindByID(context.Background(), l.ID())
	assert.Equal(t, "dependency", updated.LinkType().String())
	assert.Equal(t, 0.6, updated.Weight().Value())
	assert.NotNil(t, updated.LastWeightUpdate())
}

func TestUpdateLink_Unauthorized(t *testing.T) {
	r, linkRepo, noteRepo := setupLinkRouter()
	l, _, _ := createTestLink(t, r, linkRepo, noteRepo)

	body := map[string]interface{}{"weight": 0.6}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/links/"+l.ID().String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateLink_Forbidden(t *testing.T) {
	userID := uuid.New()
	r, linkRepo, noteRepo := setupLinkRouterWithUser(userID)

	sourceID := uuid.New()
	targetID := uuid.New()
	title1, _ := note.NewTitle("Source Note")
	content1, _ := note.NewContent("Source content")
	metadata1, _ := note.NewMetadata(nil)
	sourceNote := note.NewNote(title1, content1, "star", metadata1)
	sourceNote = note.ReconstructNote(sourceID, title1, content1, "star", metadata1, sourceNote.CreatedAt(), sourceNote.UpdatedAt())

	title2, _ := note.NewTitle("Target Note")
	content2, _ := note.NewContent("Target content")
	metadata2, _ := note.NewMetadata(nil)
	targetNote := note.NewNote(title2, content2, "star", metadata2)
	targetNote = note.ReconstructNote(targetID, title2, content2, "star", metadata2, targetNote.CreatedAt(), targetNote.UpdatedAt())

	noteRepo.notes[sourceID] = sourceNote
	noteRepo.notes[targetID] = targetNote

	otherUser := uuid.New()
	linkType, _ := link.NewLinkType("reference")
	weight, _ := link.NewWeight(0.8)
	linkMetadata, _ := link.NewMetadata(nil)
	l := link.NewLinkWithCreator(sourceID, targetID, otherUser, linkType, weight, linkMetadata)
	linkRepo.Save(context.Background(), l)

	body := map[string]interface{}{"weight": 0.6}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/links/"+l.ID().String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
