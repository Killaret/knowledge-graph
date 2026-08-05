package notehandler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	importer "knowledge-graph/internal/application/import"
	"knowledge-graph/internal/config"
	"knowledge-graph/internal/domain/cache/cachetest"
	"knowledge-graph/internal/domain/note"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupFullNoteRouter() (*gin.Engine, *mockNoteRepo) {
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
	importSvc := importer.NewService(repo, cacheClient, nil)
	handler := New(repo, nil, nil, nil, 0, nil, nil, nil, cfg, nil, nil, importSvc)

	r := gin.Default()
	r.POST("/notes", handler.Create)
	r.GET("/notes/:id", handler.Get)
	r.PUT("/notes/:id", handler.Update)
	r.DELETE("/notes/:id", handler.Delete)
	r.POST("/notes/:id/restore", handler.Restore)
	r.POST("/notes/batch", handler.DeleteBatch)
	r.GET("/notes/:id/suggestions", handler.GetSuggestions)
	r.GET("/notes/search", handler.Search)
	r.GET("/notes", handler.List)
	return r, repo
}

func TestRestoreNote(t *testing.T) {
	r, repo := setupFullNoteRouter()

	title, _ := note.NewTitle("Restored")
	content, _ := note.NewContent("Content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, "star", metadata)
	ctx := context.Background()
	_ = repo.Save(ctx, n)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/notes/"+n.ID().String()+"/restore", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestSearchNotes(t *testing.T) {
	r, repo := setupFullNoteRouter()

	title, _ := note.NewTitle("Searchable Note")
	content, _ := note.NewContent("find me here")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, "star", metadata)
	ctx := context.Background()
	_ = repo.Save(ctx, n)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/notes/search?q=find", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSearchNotes_EmptyQuery(t *testing.T) {
	r, repo := setupFullNoteRouter()

	title, _ := note.NewTitle("Note")
	content, _ := note.NewContent("content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, "star", metadata)
	_ = repo.Save(context.Background(), n)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/notes/search?q=", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSearchNotes_LongQuery(t *testing.T) {
	r, _ := setupFullNoteRouter()

	longQ := bytes.Repeat([]byte("a"), 300)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/notes/search?q="+string(longQ), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRestoreNote_InvalidUUID(t *testing.T) {
	r, _ := setupFullNoteRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/notes/invalid-uuid/restore", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRestoreNote_NotFound(t *testing.T) {
	r, _ := setupFullNoteRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/notes/"+uuid.New().String()+"/restore", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
