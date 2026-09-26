//go:build !integration
// +build !integration

package notehandler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	importer "knowledge-graph/internal/application/import"
	"knowledge-graph/internal/config"
	"knowledge-graph/internal/domain/note"
)

type fakeExtractor struct {
	page  *importer.ExtractedPage
	err   error
	calls int
}

func (f *fakeExtractor) Extract(_ context.Context, _ string) (*importer.ExtractedPage, error) {
	f.calls++
	return f.page, f.err
}

func refetchNote(t *testing.T, content, sourceURL string) *note.Note {
	meta, err := note.NewMetadata(map[string]any{"source_url": sourceURL})
	require.NoError(t, err)
	title, err := note.NewTitle("старая заметка")
	require.NoError(t, err)
	c, err := note.NewContent(content)
	require.NoError(t, err)
	return note.NewNote(title, c, note.MustType("star"), meta)
}

func refetchRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/notes/:id/refetch/preview", h.RefetchPreview)
	r.POST("/notes/:id/refetch", h.RefetchApply)
	r.POST("/notes/:id/refetch/restore", h.RefetchRestore)
	return r
}

func refetchHandler(repo *noteRepoMock, x RefetchExtractor) *Handler {
	h := New(repo, nil, nil, nil, 0, nil, nil, nil, &config.Config{}, nil, nil, nil)
	h.SetRefetch(x)
	return h
}

func TestRefetchPreview_NoMutation(t *testing.T) {
	n := refetchNote(t, "старый текст", "https://example.com/a")
	repo := &noteRepoMock{}
	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)
	// Save must never be called — preview is a pure read.
	repo.On("Save", mock.Anything, mock.Anything).Return(nil)

	x := &fakeExtractor{page: &importer.ExtractedPage{
		Title:           "fallback",
		TitleCandidates: []string{"Свежее название"},
		TitleSource:     "rule",
		Text:            "Новый текст страницы со структурой.",
		Outline:         []importer.OutlineEntry{{Level: 2, Text: "Раздел"}},
		NoiseDropped:    3,
	}}
	h := refetchHandler(repo, x)
	r := refetchRouter(h)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/notes/"+n.ID().String()+"/refetch/preview", nil))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Contains(t, rec.Body.String(), "Свежее название")
	assert.Equal(t, 1, x.calls)
	repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	// Note object itself is untouched.
	assert.Equal(t, "старый текст", n.Content().String())
}

func TestRefetchPreview_NoSourceURL(t *testing.T) {
	n := refetchNote(t, "текст", "")
	repo := &noteRepoMock{}
	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)
	h := refetchHandler(repo, &fakeExtractor{})
	r := refetchRouter(h)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/notes/"+n.ID().String()+"/refetch/preview", nil))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRefetchApply_SavesPreviousAndReplaces(t *testing.T) {
	oldContent := "## [Старое](https://example.com/a)\n\nстарая версия"
	n := refetchNote(t, oldContent, "https://example.com/a")
	repo := &noteRepoMock{}
	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)
	repo.On("Save", mock.Anything, n).Return(nil)

	x := &fakeExtractor{page: &importer.ExtractedPage{
		TitleCandidates: []string{"Новое название"},
		Text:            "Полный новый текст.",
	}}
	h := refetchHandler(repo, x)
	r := refetchRouter(h)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/notes/"+n.ID().String()+"/refetch", nil))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	// Content replaced by the fresh extraction.
	assert.Contains(t, n.Content().String(), "Полный новый текст.")
	assert.Equal(t, "Новое название", n.Title().String())

	// Previous version preserved byte-for-byte.
	prev, ok := n.Metadata().Value()["previous_content"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, oldContent, prev["content"])
	assert.Equal(t, "старая заметка", prev["title"])
	repo.AssertCalled(t, "Save", mock.Anything, n)
}

func TestRefetchApply_TitleOverride(t *testing.T) {
	n := refetchNote(t, "старое", "https://example.com/a")
	repo := &noteRepoMock{}
	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)
	repo.On("Save", mock.Anything, n).Return(nil)
	x := &fakeExtractor{page: &importer.ExtractedPage{Title: "t", Text: "x"}}
	h := refetchHandler(repo, x)
	r := refetchRouter(h)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/notes/"+n.ID().String()+"/refetch",
		strings.NewReader(`{"title":"Моё название"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Моё название", n.Title().String())
}

func TestRefetchRestore_ByteExact(t *testing.T) {
	oldContent := "## [Старое](https://example.com/a)\n\nстарая версия — побайтово"
	n := refetchNote(t, oldContent, "https://example.com/a")
	repo := &noteRepoMock{}
	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)
	repo.On("Save", mock.Anything, n).Return(nil)
	x := &fakeExtractor{page: &importer.ExtractedPage{TitleCandidates: []string{"Н"}, Text: "новое"}}
	h := refetchHandler(repo, x)
	r := refetchRouter(h)

	// apply then restore
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/notes/"+n.ID().String()+"/refetch", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotEqual(t, oldContent, n.Content().String())

	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/notes/"+n.ID().String()+"/refetch/restore", nil))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, oldContent, n.Content().String())
	_, stillThere := n.Metadata().Value()["previous_content"]
	assert.False(t, stillThere)

	// second restore — nothing left to restore
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/notes/"+n.ID().String()+"/refetch/restore", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestRefetchApply_FetchError(t *testing.T) {
	n := refetchNote(t, "старое", "https://example.com/a")
	repo := &noteRepoMock{}
	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)
	x := &fakeExtractor{err: errors.New("network down")}
	h := refetchHandler(repo, x)
	r := refetchRouter(h)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/notes/"+n.ID().String()+"/refetch", nil))
	assert.Equal(t, http.StatusBadGateway, rec.Code)
	repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}
