package draft

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"knowledge-graph/internal/application/draft"
	noteDomain "knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockDraftRepo struct {
	mock.Mock
}

func (m *mockDraftRepo) Save(ctx context.Context, draft *noteDomain.Draft) error {
	args := m.Called(ctx, draft)
	return args.Error(0)
}

func (m *mockDraftRepo) FindByNoteAndUser(ctx context.Context, noteID, userID uuid.UUID) (*noteDomain.Draft, error) {
	args := m.Called(ctx, noteID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*noteDomain.Draft), args.Error(1)
}

func (m *mockDraftRepo) FindActiveByUser(ctx context.Context, userID uuid.UUID) ([]*noteDomain.Draft, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*noteDomain.Draft), args.Error(1)
}

func (m *mockDraftRepo) FindByID(ctx context.Context, id uuid.UUID) (*noteDomain.Draft, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*noteDomain.Draft), args.Error(1)
}

func (m *mockDraftRepo) DeleteByID(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockDraftRepo) DeleteExpired(ctx context.Context, before time.Time) (int, error) {
	args := m.Called(ctx, before)
	return args.Int(0), args.Error(1)
}

func (m *mockDraftRepo) Update(ctx context.Context, draft *noteDomain.Draft) error {
	args := m.Called(ctx, draft)
	return args.Error(0)
}

type dummyNoteRepo struct{}

func (d *dummyNoteRepo) Save(ctx context.Context, note *noteDomain.Note) error { return nil }
func (d *dummyNoteRepo) FindByID(ctx context.Context, id uuid.UUID) (*noteDomain.Note, error) {
	return nil, nil
}
func (d *dummyNoteRepo) Delete(ctx context.Context, id uuid.UUID) error         { return nil }
func (d *dummyNoteRepo) DeleteBatch(ctx context.Context, ids []uuid.UUID) error { return nil }
func (d *dummyNoteRepo) Restore(ctx context.Context, id uuid.UUID) error        { return nil }
func (d *dummyNoteRepo) List(ctx context.Context, limit, offset int) ([]*noteDomain.Note, int64, error) {
	return nil, 0, nil
}
func (d *dummyNoteRepo) Search(ctx context.Context, query string, limit, offset int) ([]*noteDomain.Note, int64, error) {
	return nil, 0, nil
}
func (d *dummyNoteRepo) FindAll(ctx context.Context) ([]*noteDomain.Note, error) { return nil, nil }
func (d *dummyNoteRepo) FindAllPaginated(ctx context.Context, limit, offset int) ([]*noteDomain.Note, int64, error) {
	return nil, 0, nil
}

func setupDraftHandler() (*gin.Engine, *mockDraftRepo) {
	gin.SetMode(gin.TestMode)
	repo := new(mockDraftRepo)
	service := draft.NewService(repo, &dummyNoteRepo{}, "")
	h := NewHandler(service)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, uuid.MustParse("00000000-0000-0000-0000-000000000001"))
		c.Next()
	})
	r.POST("/notes/:id/draft", h.SaveDraft)
	r.GET("/notes/:id/draft", h.GetDraft)
	r.DELETE("/drafts/:draft_id", h.DeleteDraft)
	r.POST("/drafts/:draft_id/sync", h.SyncDraft)
	r.POST("/drafts/:draft_id/resolve", h.ResolveConflict)
	r.GET("/drafts", h.GetActiveDrafts)
	return r, repo
}

func TestSaveDraft_Create(t *testing.T) {
	r, repo := setupDraftHandler()
	noteID := uuid.New()
	repo.On("FindByNoteAndUser", mock.Anything, noteID, mock.Anything).Return(nil, nil)
	repo.On("Save", mock.Anything, mock.Anything).Return(nil)

	body, _ := json.Marshal(map[string]string{"content": "draft content", "title": "title"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/notes/"+noteID.String()+"/draft", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSaveDraft_Update(t *testing.T) {
	r, repo := setupDraftHandler()
	noteID := uuid.New()
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	d := noteDomain.NewDraft(noteID, uid, "old", "old")

	repo.On("FindByNoteAndUser", mock.Anything, noteID, uid).Return(d, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)

	body, _ := json.Marshal(map[string]string{"content": "new content", "title": "new title"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/notes/"+noteID.String()+"/draft", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetDraft_NotFound(t *testing.T) {
	r, repo := setupDraftHandler()
	noteID := uuid.New()
	repo.On("FindByNoteAndUser", mock.Anything, noteID, mock.Anything).Return(nil, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String()+"/draft", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteDraft(t *testing.T) {
	r, repo := setupDraftHandler()
	draftID := uuid.New()
	repo.On("DeleteByID", mock.Anything, draftID).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/drafts/"+draftID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSyncDraft(t *testing.T) {
	r, repo := setupDraftHandler()
	draftID := uuid.New()
	d := noteDomain.NewDraft(uuid.New(), uuid.New(), "content", "title")

	repo.On("FindByID", mock.Anything, draftID).Return(d, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	repo.On("DeleteByID", mock.Anything, d.ID()).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/drafts/"+draftID.String()+"/sync", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestResolveConflict(t *testing.T) {
	r, repo := setupDraftHandler()
	draftID := uuid.New()
	d := noteDomain.NewDraft(uuid.New(), uuid.New(), "content", "title")
	d.StartPublishing()
	d.MarkAsConflict()

	repo.On("FindByID", mock.Anything, draftID).Return(d, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/drafts/"+draftID.String()+"/resolve", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetActiveDrafts(t *testing.T) {
	r, repo := setupDraftHandler()
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	d := noteDomain.NewDraft(uuid.New(), uid, "content", "title")

	repo.On("FindActiveByUser", mock.Anything, uid).Return([]*noteDomain.Draft{d}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/drafts", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp["drafts"], 1)
}
