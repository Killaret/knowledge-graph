//go:build !integration
// +build !integration

package notehandler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	appcache "knowledge-graph/internal/application/cache"
	importer "knowledge-graph/internal/application/import"
	"knowledge-graph/internal/application/recommendation"
	"knowledge-graph/internal/config"
	"knowledge-graph/internal/domain/cache/cachetest"
	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type noteRepoMock struct{ mock.Mock }

func (m *noteRepoMock) Save(ctx context.Context, n *note.Note) error {
	return m.Called(ctx, n).Error(0)
}

func (m *noteRepoMock) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*note.Note), args.Error(1)
}

func (m *noteRepoMock) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *noteRepoMock) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
	return m.Called(ctx, ids).Error(0)
}

func (m *noteRepoMock) Restore(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *noteRepoMock) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*note.Note, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*note.Note), args.Get(1).(int64), args.Error(2)
}

func (m *noteRepoMock) Search(ctx context.Context, userID uuid.UUID, query string, limit, offset int) ([]*note.Note, int64, error) {
	args := m.Called(ctx, userID, query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*note.Note), args.Get(1).(int64), args.Error(2)
}

func (m *noteRepoMock) FindAll(ctx context.Context) ([]*note.Note, error) {
	return nil, nil
}

func (m *noteRepoMock) FindAllPaginated(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*note.Note, int64, error) {
	return nil, int64(0), nil
}

type recRepoMock struct{ mock.Mock }

func (m *recRepoMock) Count(ctx context.Context, noteID uuid.UUID) (int64, error) {
	args := m.Called(ctx, noteID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *recRepoMock) GetNotesThatRecommend(ctx context.Context, recommendedID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, recommendedID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func (m *recRepoMock) ReplaceRecommendations(ctx context.Context, noteID uuid.UUID, recs map[uuid.UUID]float64) error {
	return m.Called(ctx, noteID, recs).Error(0)
}

func (m *recRepoMock) GetRecommendations(ctx context.Context, noteID uuid.UUID, limit int) ([]recommendation.Recommendation, error) {
	args := m.Called(ctx, noteID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]recommendation.Recommendation), args.Error(1)
}

type embeddingRepoMock struct{ mock.Mock }

func (m *embeddingRepoMock) FindSimilarNotes(ctx context.Context, noteID uuid.UUID, limit int) ([]recommendation.SimilarNote, error) {
	args := m.Called(ctx, noteID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]recommendation.SimilarNote), args.Error(1)
}

func (m *embeddingRepoMock) FindSimilarNotesBatch(ctx context.Context, noteIDs []uuid.UUID, limit int) (map[uuid.UUID][]recommendation.SimilarNote, error) {
	return nil, nil
}

type taskQueueMock struct{ mock.Mock }

func (m *taskQueueMock) EnqueueBackupToCloud(ctx context.Context, localPath, remoteKey, backupDate string) error {
	return nil
}

func (m *taskQueueMock) EnqueueRefreshRecommendations(ctx context.Context, noteID uuid.UUID, delay time.Duration) error {
	return m.Called(ctx, noteID, delay).Error(0)
}

func (m *taskQueueMock) EnqueueExtractKeywords(ctx context.Context, noteID string, topN int) error {
	return m.Called(ctx, noteID, topN).Error(0)
}

func (m *taskQueueMock) EnqueueComputeEmbedding(ctx context.Context, noteID string) error {
	return m.Called(ctx, noteID).Error(0)
}

func (m *taskQueueMock) EnqueueRecalculateLinkWeights(ctx context.Context, noteID uuid.UUID, delay time.Duration) error {
	return m.Called(ctx, noteID, delay).Error(0)
}

func (m *taskQueueMock) EnqueueNotification(ctx context.Context, payload []byte) error {
	return nil
}

func (m *taskQueueMock) EnqueueImportBookmarks(ctx context.Context, userID uuid.UUID, taskID string, items []byte) error {
	return m.Called(ctx, userID, taskID, items).Error(0)
}

func newTestConfig() *config.Config {
	return &config.Config{
		PaginationDefaultLimit:                20,
		PaginationMaxLimit:                    100,
		RecommendationTopN:                    5,
		RecommendationFallbackEnabled:         true,
		RecommendationFallbackSemanticEnabled: true,
		RecommendationTaskDelaySeconds:        0,
	}
}

func setupUnitHandler(t *testing.T) (*Handler, *noteRepoMock, *taskQueueMock, *recRepoMock, *embeddingRepoMock, *cachetest.FakeCacheClient) {
	gin.SetMode(gin.TestMode)
	repo := new(noteRepoMock)
	tq := new(taskQueueMock)
	recRepo := new(recRepoMock)
	embRepo := new(embeddingRepoMock)
	cache := cachetest.NewFakeCacheClient()
	cfg := newTestConfig()
	importSvc := importer.NewService(repo, cache, nil)
	h := New(repo, tq, nil, nil, time.Millisecond, recRepo, embRepo, cache, cfg, appcache.NewGraphCache(cache), nil, importSvc)
	return h, repo, tq, recRepo, embRepo, cache
}

func newTestNote(t *testing.T, title, content, noteType string) *note.Note {
	t.Helper()
	ttl, err := note.NewTitle(title)
	require.NoError(t, err)
	cnt, err := note.NewContent(content)
	require.NoError(t, err)
	meta, err := note.NewMetadata(nil)
	require.NoError(t, err)
	return note.NewNote(ttl, cnt, noteType, meta)
}

func newContext(t *testing.T, method, target, body string, userID ...uuid.UUID) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	if method != http.MethodGet {
		c.Request.Header.Set("Content-Type", "application/json")
	}
	if len(userID) > 0 {
		c.Set(middleware.ContextUserIDKey, userID[0])
	}
	return w, c
}

func withID(c *gin.Context, id uuid.UUID) {
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
}

func TestCreateNote_Success(t *testing.T) {
	h, repo, tq, _, _, _ := setupUnitHandler(t)

	repo.On("Save", mock.Anything, mock.AnythingOfType("*note.Note")).Return(nil)
	tq.On("EnqueueExtractKeywords", mock.Anything, mock.AnythingOfType("string"), 10).Return(nil)
	tq.On("EnqueueComputeEmbedding", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	tq.On("EnqueueRecalculateLinkWeights", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("time.Duration")).Return(nil)
	tq.On("EnqueueRecalculateLinkWeights", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("time.Duration")).Return(nil)

	body := `{"title":"Test Note","content":"Hello","type":"star","metadata":{"key":"value"}}`
	w, c := newContext(t, http.MethodPost, "/notes", body)
	h.Create(c)
	_ = w

	assert.Equal(t, http.StatusCreated, c.Writer.Status())
	assert.Contains(t, w.Body.String(), "Test Note")
	repo.AssertExpectations(t)
	tq.AssertExpectations(t)
}

func TestCreateNote_WithUser(t *testing.T) {
	h, repo, tq, _, _, _ := setupUnitHandler(t)
	userID := uuid.New()

	repo.On("Save", mock.Anything, mock.AnythingOfType("*note.Note")).Return(nil)
	tq.On("EnqueueExtractKeywords", mock.Anything, mock.AnythingOfType("string"), 10).Return(nil)
	tq.On("EnqueueComputeEmbedding", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	tq.On("EnqueueRecalculateLinkWeights", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("time.Duration")).Return(nil)

	body := `{"title":"User Note","content":"content","type":"planet"}`
	w, c := newContext(t, http.MethodPost, "/notes", body, userID)
	h.Create(c)
	_ = w

	assert.Equal(t, http.StatusCreated, c.Writer.Status())
	assert.Contains(t, w.Body.String(), "User Note")
	repo.AssertExpectations(t)
	tq.AssertExpectations(t)
}

func TestCreateNote_ValidationError(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)

	body := `{"title":"","content":"Hello"}`
	w, c := newContext(t, http.MethodPost, "/notes", body)
	h.Create(c)
	_ = w

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	repo.AssertNotCalled(t, "Save")
}

func TestCreateNote_NewTitleError(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)

	body := `{"title":"   ","content":"Hello"}`
	w, c := newContext(t, http.MethodPost, "/notes", body)
	h.Create(c)
	_ = w

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	assert.Contains(t, w.Body.String(), "title")
	repo.AssertNotCalled(t, "Save")
}

func TestCreateNote_NewContentError(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)

	longContent := strings.Repeat("a", 10001)
	body := fmt.Sprintf(`{"title":"T","content":"%s"}`, longContent)
	w, c := newContext(t, http.MethodPost, "/notes", body)
	h.Create(c)
	_ = w

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	repo.AssertNotCalled(t, "Save")
}

func TestCreateNote_SaveError(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)

	repo.On("Save", mock.Anything, mock.AnythingOfType("*note.Note")).Return(assert.AnError)

	body := `{"title":"T","content":"c"}`
	w, c := newContext(t, http.MethodPost, "/notes", body)
	h.Create(c)
	_ = w

	assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())
	repo.AssertExpectations(t)
}

func TestCreateNote_AffectedNotesSvc(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := new(noteRepoMock)
	tq := new(taskQueueMock)
	recRepo := new(recRepoMock)
	cache := cachetest.NewFakeCacheClient()
	cfg := newTestConfig()
	affectedSvc := recommendation.NewAffectedNotesService(recRepo)
	importSvc := importer.NewService(repo, cache, nil)
	h := New(repo, tq, nil, affectedSvc, time.Millisecond, recRepo, nil, cache, cfg, nil, nil, importSvc)

	repo.On("Save", mock.Anything, mock.AnythingOfType("*note.Note")).Return(nil)
	tq.On("EnqueueExtractKeywords", mock.Anything, mock.AnythingOfType("string"), 10).Return(nil)
	tq.On("EnqueueComputeEmbedding", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	tq.On("EnqueueRecalculateLinkWeights", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("time.Duration")).Return(nil)
	recRepo.On("GetNotesThatRecommend", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return([]uuid.UUID{uuid.New()}, nil)
	tq.On("EnqueueRefreshRecommendations", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("time.Duration")).Return(nil).Twice()

	body := `{"title":"Affected","content":"note"}`
	w, c := newContext(t, http.MethodPost, "/notes", body)
	h.Create(c)
	_ = w

	assert.Equal(t, http.StatusCreated, c.Writer.Status())
	tq.AssertExpectations(t)
	recRepo.AssertExpectations(t)
}

func TestCreateNote_AffectedNotesSvcError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := new(noteRepoMock)
	tq := new(taskQueueMock)
	recRepo := new(recRepoMock)
	cache := cachetest.NewFakeCacheClient()
	cfg := newTestConfig()
	affectedSvc := recommendation.NewAffectedNotesService(recRepo)
	importSvc := importer.NewService(repo, cache, nil)
	h := New(repo, tq, nil, affectedSvc, time.Millisecond, recRepo, nil, cache, cfg, nil, nil, importSvc)

	repo.On("Save", mock.Anything, mock.AnythingOfType("*note.Note")).Return(nil)
	tq.On("EnqueueExtractKeywords", mock.Anything, mock.AnythingOfType("string"), 10).Return(nil)
	tq.On("EnqueueComputeEmbedding", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	tq.On("EnqueueRecalculateLinkWeights", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("time.Duration")).Return(nil)
	recRepo.On("GetNotesThatRecommend", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, assert.AnError)

	body := `{"title":"Affected","content":"note"}`
	w, c := newContext(t, http.MethodPost, "/notes", body)
	h.Create(c)
	_ = w

	assert.Equal(t, http.StatusCreated, c.Writer.Status())
	tq.AssertExpectations(t)
	recRepo.AssertExpectations(t)
}

func TestUpdateNote_Success(t *testing.T) {
	h, repo, tq, _, _, _ := setupUnitHandler(t)
	userID := uuid.New()
	n := newTestNote(t, "Old Title", "Old Content", "star")

	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)
	repo.On("Save", mock.Anything, mock.AnythingOfType("*note.Note")).Return(nil)
	tq.On("EnqueueExtractKeywords", mock.Anything, mock.AnythingOfType("string"), 10).Return(nil)
	tq.On("EnqueueComputeEmbedding", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	tq.On("EnqueueRecalculateLinkWeights", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("time.Duration")).Return(nil)

	body := `{"title":"New Title","content":"New Content","type":"planet"}`
	w, c := newContext(t, http.MethodPut, "/notes/"+n.ID().String(), body, userID)
	withID(c, n.ID())
	h.Update(c)
	_ = w

	assert.Equal(t, http.StatusOK, c.Writer.Status())
	assert.Contains(t, w.Body.String(), "New Title")
	repo.AssertExpectations(t)
	tq.AssertExpectations(t)
}

func TestUpdateNote_NoTextChange(t *testing.T) {
	h, repo, tq, _, _, _ := setupUnitHandler(t)
	n := newTestNote(t, "Title", "Content", "star")

	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)
	repo.On("Save", mock.Anything, mock.AnythingOfType("*note.Note")).Return(nil)

	body := `{"type":"planet","metadata":{"key":"value"}}`
	w, c := newContext(t, http.MethodPut, "/notes/"+n.ID().String(), body)
	withID(c, n.ID())
	h.Update(c)
	_ = w

	assert.Equal(t, http.StatusOK, c.Writer.Status())
	tq.AssertNotCalled(t, "EnqueueExtractKeywords")
	tq.AssertNotCalled(t, "EnqueueComputeEmbedding")
}

func TestUpdateNote_InvalidID(t *testing.T) {
	h, _, _, _, _, _ := setupUnitHandler(t)

	w, c := newContext(t, http.MethodPut, "/notes/bad-uuid", `{"title":"T"}`)
	c.Params = gin.Params{{Key: "id", Value: "bad-uuid"}}
	h.Update(c)
	_ = w

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
}

func TestUpdateNote_NotFound(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	id := uuid.New()

	repo.On("FindByID", mock.Anything, id).Return(nil, nil)

	w, c := newContext(t, http.MethodPut, "/notes/"+id.String(), `{"title":"T"}`)
	withID(c, id)
	h.Update(c)
	_ = w

	assert.Equal(t, http.StatusNotFound, c.Writer.Status())
}

func TestUpdateNote_FindByIDError(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	id := uuid.New()

	repo.On("FindByID", mock.Anything, id).Return(nil, assert.AnError)

	w, c := newContext(t, http.MethodPut, "/notes/"+id.String(), `{"title":"T"}`)
	withID(c, id)
	h.Update(c)
	_ = w

	assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())
}

func TestUpdateNote_ValidationError(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	n := newTestNote(t, "T", "C", "star")

	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)

	longTitle := strings.Repeat("a", 201)
	body := fmt.Sprintf(`{"title":"%s"}`, longTitle)
	w, c := newContext(t, http.MethodPut, "/notes/"+n.ID().String(), body)
	withID(c, n.ID())
	h.Update(c)
	_ = w

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	repo.AssertNotCalled(t, "Save")
}

func TestUpdateNote_NewTitleError(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	n := newTestNote(t, "T", "C", "star")

	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)

	body := `{"title":"   "}`
	w, c := newContext(t, http.MethodPut, "/notes/"+n.ID().String(), body)
	withID(c, n.ID())
	h.Update(c)
	_ = w

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	repo.AssertNotCalled(t, "Save")
}

func TestUpdateNote_NewContentError(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	n := newTestNote(t, "T", "C", "star")

	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)

	longContent := strings.Repeat("a", 10001)
	body := fmt.Sprintf(`{"content":"%s"}`, longContent)
	w, c := newContext(t, http.MethodPut, "/notes/"+n.ID().String(), body)
	withID(c, n.ID())
	h.Update(c)
	_ = w

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	repo.AssertNotCalled(t, "Save")
}

func TestUpdateNote_SaveError(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	n := newTestNote(t, "T", "C", "star")

	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)
	repo.On("Save", mock.Anything, mock.AnythingOfType("*note.Note")).Return(assert.AnError)

	body := `{"title":"New"}`
	w, c := newContext(t, http.MethodPut, "/notes/"+n.ID().String(), body)
	withID(c, n.ID())
	h.Update(c)
	_ = w

	assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())
}

func TestDeleteNote_Success(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	userID := uuid.New()
	n := newTestNote(t, "ToDelete", "Content", "star")

	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)
	repo.On("Delete", mock.Anything, n.ID()).Return(nil)

	w, c := newContext(t, http.MethodDelete, "/notes/"+n.ID().String(), "", userID)
	withID(c, n.ID())
	h.Delete(c)
	_ = w

	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	repo.AssertExpectations(t)
}

func TestDeleteNote_InvalidID(t *testing.T) {
	h, _, _, _, _, _ := setupUnitHandler(t)

	w, c := newContext(t, http.MethodDelete, "/notes/bad-uuid", "")
	c.Params = gin.Params{{Key: "id", Value: "bad-uuid"}}
	h.Delete(c)
	_ = w

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
}

func TestDeleteNote_NotFound(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	id := uuid.New()

	repo.On("FindByID", mock.Anything, id).Return(nil, nil)

	w, c := newContext(t, http.MethodDelete, "/notes/"+id.String(), "")
	withID(c, id)
	h.Delete(c)
	_ = w

	assert.Equal(t, http.StatusNotFound, c.Writer.Status())
}

func TestDeleteNote_FindByIDError(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	id := uuid.New()

	repo.On("FindByID", mock.Anything, id).Return(nil, assert.AnError)

	w, c := newContext(t, http.MethodDelete, "/notes/"+id.String(), "")
	withID(c, id)
	h.Delete(c)
	_ = w

	assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())
}

func TestDeleteNote_DeleteError(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	n := newTestNote(t, "T", "C", "star")

	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)
	repo.On("Delete", mock.Anything, n.ID()).Return(assert.AnError)

	w, c := newContext(t, http.MethodDelete, "/notes/"+n.ID().String(), "")
	withID(c, n.ID())
	h.Delete(c)
	_ = w

	assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())
}

func TestDeleteBatchNotes_Success(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	id1 := uuid.New()
	id2 := uuid.New()

	repo.On("DeleteBatch", mock.Anything, mock.AnythingOfType("[]uuid.UUID")).Return(nil)

	body := fmt.Sprintf(`{"ids":["%s","%s"]}`, id1, id2)
	w, c := newContext(t, http.MethodPost, "/notes/batch", body)
	h.DeleteBatch(c)
	_ = w

	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	repo.AssertExpectations(t)
}

func TestDeleteBatchNotes_InvalidBody(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)

	w, c := newContext(t, http.MethodPost, "/notes/batch", `{}`)
	h.DeleteBatch(c)
	_ = w

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	repo.AssertNotCalled(t, "DeleteBatch")
}

func TestDeleteBatchNotes_EmptyIDs(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)

	w, c := newContext(t, http.MethodPost, "/notes/batch", `{"ids":[]}`)
	h.DeleteBatch(c)
	_ = w

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	repo.AssertNotCalled(t, "DeleteBatch")
}

func TestDeleteBatchNotes_InvalidUUID(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)

	w, c := newContext(t, http.MethodPost, "/notes/batch", `{"ids":["not-a-uuid"]}`)
	h.DeleteBatch(c)
	_ = w

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	repo.AssertNotCalled(t, "DeleteBatch")
}

func TestDeleteBatchNotes_RepoError(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	id := uuid.New()

	repo.On("DeleteBatch", mock.Anything, mock.AnythingOfType("[]uuid.UUID")).Return(assert.AnError)

	body := fmt.Sprintf(`{"ids":["%s"]}`, id)
	w, c := newContext(t, http.MethodPost, "/notes/batch", body)
	h.DeleteBatch(c)
	_ = w

	assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())
}

func TestRestoreNote_Success(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	userID := uuid.New()
	id := uuid.New()

	repo.On("Restore", mock.Anything, id).Return(nil)

	w, c := newContext(t, http.MethodPost, "/notes/"+id.String()+"/restore", "", userID)
	withID(c, id)
	h.Restore(c)
	_ = w

	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	repo.AssertExpectations(t)
}

func TestRestoreNote_InvalidID(t *testing.T) {
	h, _, _, _, _, _ := setupUnitHandler(t)

	w, c := newContext(t, http.MethodPost, "/notes/bad-uuid/restore", "")
	c.Params = gin.Params{{Key: "id", Value: "bad-uuid"}}
	h.Restore(c)
	_ = w

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
}

func TestRestoreNote_NotFoundUnit(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	id := uuid.New()

	repo.On("Restore", mock.Anything, id).Return(note.ErrNoteNotFound)

	w, c := newContext(t, http.MethodPost, "/notes/"+id.String()+"/restore", "")
	withID(c, id)
	h.Restore(c)
	_ = w

	assert.Equal(t, http.StatusNotFound, c.Writer.Status())
}

func TestRestoreNote_RepoError(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	id := uuid.New()

	repo.On("Restore", mock.Anything, id).Return(assert.AnError)

	w, c := newContext(t, http.MethodPost, "/notes/"+id.String()+"/restore", "")
	withID(c, id)
	h.Restore(c)
	_ = w

	assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())
}

func TestGetNote_Success(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	n := newTestNote(t, "GetTest", "Content", "star")

	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)

	w, c := newContext(t, http.MethodGet, "/notes/"+n.ID().String(), "")
	withID(c, n.ID())
	h.Get(c)
	_ = w

	assert.Equal(t, http.StatusOK, c.Writer.Status())
	assert.Contains(t, w.Body.String(), "GetTest")
}

func TestGetNote_InvalidID(t *testing.T) {
	h, _, _, _, _, _ := setupUnitHandler(t)

	w, c := newContext(t, http.MethodGet, "/notes/bad-uuid", "")
	c.Params = gin.Params{{Key: "id", Value: "bad-uuid"}}
	h.Get(c)
	_ = w

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
}

func TestGetNote_NotFound(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	id := uuid.New()

	repo.On("FindByID", mock.Anything, id).Return(nil, nil)

	w, c := newContext(t, http.MethodGet, "/notes/"+id.String(), "")
	withID(c, id)
	h.Get(c)
	_ = w

	assert.Equal(t, http.StatusNotFound, c.Writer.Status())
}

func TestGetNote_FindByIDError(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	id := uuid.New()

	repo.On("FindByID", mock.Anything, id).Return(nil, assert.AnError)

	w, c := newContext(t, http.MethodGet, "/notes/"+id.String(), "")
	withID(c, id)
	h.Get(c)
	_ = w

	assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())
}

func TestGetSuggestions_Precomputed(t *testing.T) {
	h, repo, _, recRepo, _, _ := setupUnitHandler(t)
	n := newTestNote(t, "Sug", "Content", "star")

	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)
	recRepo.On("GetRecommendations", mock.Anything, n.ID(), 5).Return([]recommendation.Recommendation{
		{NoteID: n.ID(), RecommendedNoteID: uuid.New(), Score: 0.9, UpdatedAt: time.Now().Add(time.Hour)},
	}, nil)

	w, c := newContext(t, http.MethodGet, "/notes/"+n.ID().String()+"/suggestions", "")
	withID(c, n.ID())
	h.GetSuggestions(c)
	_ = w

	assert.Equal(t, http.StatusOK, c.Writer.Status())
	assert.Equal(t, "table", w.Header().Get("X-Recommendations-Source"))
	assert.Empty(t, w.Header().Get("X-Recommendations-Stale"))
}

func TestGetSuggestions_PrecomputedStale(t *testing.T) {
	h, repo, tq, recRepo, _, _ := setupUnitHandler(t)
	n := newTestNote(t, "Sug", "Content", "star")

	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)
	recRepo.On("GetRecommendations", mock.Anything, n.ID(), 5).Return([]recommendation.Recommendation{
		{NoteID: n.ID(), RecommendedNoteID: uuid.New(), Score: 0.9, UpdatedAt: time.Now().Add(-time.Hour)},
	}, nil)
	tq.On("EnqueueRefreshRecommendations", mock.Anything, n.ID(), mock.AnythingOfType("time.Duration")).Return(nil)

	w, c := newContext(t, http.MethodGet, "/notes/"+n.ID().String()+"/suggestions", "")
	withID(c, n.ID())
	h.GetSuggestions(c)
	_ = w

	assert.Equal(t, http.StatusOK, c.Writer.Status())
	assert.Equal(t, "table", w.Header().Get("X-Recommendations-Source"))
	assert.Equal(t, "true", w.Header().Get("X-Recommendations-Stale"))
	tq.AssertExpectations(t)
}

func TestGetSuggestions_LimitParam(t *testing.T) {
	h, repo, _, recRepo, _, _ := setupUnitHandler(t)
	n := newTestNote(t, "Sug", "Content", "star")

	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)
	recRepo.On("GetRecommendations", mock.Anything, n.ID(), 2).Return([]recommendation.Recommendation{
		{NoteID: n.ID(), RecommendedNoteID: uuid.New(), Score: 0.9, UpdatedAt: time.Now().Add(time.Hour)},
	}, nil)

	w, c := newContext(t, http.MethodGet, "/notes/"+n.ID().String()+"/suggestions?limit=2", "")
	withID(c, n.ID())
	h.GetSuggestions(c)
	_ = w

	assert.Equal(t, http.StatusOK, c.Writer.Status())
	recRepo.AssertExpectations(t)
}

func TestGetSuggestions_SemanticFallback(t *testing.T) {
	h, _, tq, recRepo, embRepo, _ := setupUnitHandler(t)
	n := newTestNote(t, "Sug", "Content", "star")
	similarID := uuid.New()

	recRepo.On("GetRecommendations", mock.Anything, n.ID(), 5).Return([]recommendation.Recommendation{}, nil)
	embRepo.On("FindSimilarNotes", mock.Anything, n.ID(), 5).Return([]recommendation.SimilarNote{
		{NoteID: similarID, Score: 0.85},
	}, nil)
	tq.On("EnqueueRefreshRecommendations", mock.Anything, n.ID(), mock.AnythingOfType("time.Duration")).Return(nil)

	w, c := newContext(t, http.MethodGet, "/notes/"+n.ID().String()+"/suggestions", "")
	withID(c, n.ID())
	h.GetSuggestions(c)
	_ = w

	assert.Equal(t, http.StatusOK, c.Writer.Status())
	assert.Equal(t, "semantic", w.Header().Get("X-Recommendations-Source"))
	assert.Equal(t, "true", w.Header().Get("X-Recommendations-Stale"))
	assert.Contains(t, w.Body.String(), similarID.String())
	tq.AssertExpectations(t)
}

func TestGetSuggestions_CacheFallback(t *testing.T) {
	h, _, tq, recRepo, embRepo, cache := setupUnitHandler(t)
	n := newTestNote(t, "Sug", "Content", "star")
	suggestionID := uuid.New()

	recRepo.On("GetRecommendations", mock.Anything, n.ID(), 5).Return([]recommendation.Recommendation{}, nil)
	embRepo.On("FindSimilarNotes", mock.Anything, n.ID(), 5).Return([]recommendation.SimilarNote{}, nil)

	cached := fmt.Sprintf(`[{"note_id":%q,"score":0.7}]`, suggestionID.String())
	_ = cache.Set(context.Background(), "recommendations:"+n.ID().String(), cached, 0)
	tq.On("EnqueueRefreshRecommendations", mock.Anything, n.ID(), mock.AnythingOfType("time.Duration")).Return(nil)

	w, c := newContext(t, http.MethodGet, "/notes/"+n.ID().String()+"/suggestions", "")
	withID(c, n.ID())
	h.GetSuggestions(c)
	_ = w

	assert.Equal(t, http.StatusOK, c.Writer.Status())
	assert.Equal(t, "redis", w.Header().Get("X-Recommendations-Source"))
	assert.Contains(t, w.Body.String(), suggestionID.String())
	tq.AssertExpectations(t)
}

func TestGetSuggestions_Empty(t *testing.T) {
	h, _, tq, recRepo, embRepo, _ := setupUnitHandler(t)
	n := newTestNote(t, "Sug", "Content", "star")

	recRepo.On("GetRecommendations", mock.Anything, n.ID(), 5).Return([]recommendation.Recommendation{}, nil)
	embRepo.On("FindSimilarNotes", mock.Anything, n.ID(), 5).Return([]recommendation.SimilarNote{}, nil)
	tq.On("EnqueueRefreshRecommendations", mock.Anything, n.ID(), mock.AnythingOfType("time.Duration")).Return(nil)

	w, c := newContext(t, http.MethodGet, "/notes/"+n.ID().String()+"/suggestions", "")
	withID(c, n.ID())
	h.GetSuggestions(c)
	_ = w

	assert.Equal(t, http.StatusAccepted, c.Writer.Status())
	assert.Equal(t, "empty", w.Header().Get("X-Recommendations-Source"))
	assert.Equal(t, "true", w.Header().Get("X-Recommendations-Stale"))
	tq.AssertExpectations(t)
}

func TestGetSuggestions_TaskQueueNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := new(noteRepoMock)
	recRepo := new(recRepoMock)
	embRepo := new(embeddingRepoMock)
	cache := cachetest.NewFakeCacheClient()
	cfg := newTestConfig()
	importSvc := importer.NewService(repo, cache, nil)
	h := New(repo, nil, nil, nil, 0, recRepo, embRepo, cache, cfg, nil, nil, importSvc)

	n := newTestNote(t, "Sug", "Content", "star")
	recRepo.On("GetRecommendations", mock.Anything, n.ID(), 5).Return([]recommendation.Recommendation{}, nil)
	embRepo.On("FindSimilarNotes", mock.Anything, n.ID(), 5).Return([]recommendation.SimilarNote{}, nil)

	w, c := newContext(t, http.MethodGet, "/notes/"+n.ID().String()+"/suggestions", "")
	withID(c, n.ID())
	h.GetSuggestions(c)
	_ = w

	assert.Equal(t, http.StatusAccepted, c.Writer.Status())
}

func TestGetSuggestions_InvalidID(t *testing.T) {
	h, _, _, _, _, _ := setupUnitHandler(t)

	w, c := newContext(t, http.MethodGet, "/notes/bad-uuid/suggestions", "")
	c.Params = gin.Params{{Key: "id", Value: "bad-uuid"}}
	h.GetSuggestions(c)
	_ = w

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
}

func TestSearchNotes_Success(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	n := newTestNote(t, "Searchable", "find me here", "star")

	repo.On("Search", mock.Anything, mock.Anything, "find", 20, 0).Return([]*note.Note{n}, int64(1), nil)

	w, c := newContext(t, http.MethodGet, "/notes/search?q=find&page=1&size=20", "")
	h.Search(c)
	_ = w

	assert.Equal(t, http.StatusOK, c.Writer.Status())
	assert.Contains(t, w.Body.String(), "Searchable")
}

func TestSearchNotes_TooLong(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)

	longQ := strings.Repeat("a", 201)
	w, c := newContext(t, http.MethodGet, "/notes/search?q="+longQ, "")
	h.Search(c)
	_ = w

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	repo.AssertNotCalled(t, "Search")
}

func TestSearchNotes_InvalidPage(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)

	w, c := newContext(t, http.MethodGet, "/notes/search?q=test&page=abc", "")
	h.Search(c)
	_ = w

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	repo.AssertNotCalled(t, "Search")
}

func TestSearchNotes_RepoError(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)

	repo.On("Search", mock.Anything, mock.Anything, "find", 20, 0).Return(nil, int64(0), assert.AnError)

	w, c := newContext(t, http.MethodGet, "/notes/search?q=find", "")
	h.Search(c)
	_ = w

	assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())
}

func TestListNotes_Success(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	n := newTestNote(t, "Listed", "Content", "star")

	repo.On("List", mock.Anything, mock.Anything, 20, 0).Return([]*note.Note{n}, int64(1), nil)

	w, c := newContext(t, http.MethodGet, "/notes?limit=20&offset=0", "")
	h.List(c)
	_ = w

	assert.Equal(t, http.StatusOK, c.Writer.Status())
	assert.Contains(t, w.Body.String(), "Listed")
}

func TestListNotes_DefaultPagination(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)

	repo.On("List", mock.Anything, mock.Anything, 20, 0).Return([]*note.Note{}, int64(0), nil)

	w, c := newContext(t, http.MethodGet, "/notes", "")
	h.List(c)
	_ = w

	assert.Equal(t, http.StatusOK, c.Writer.Status())
}

func TestListNotes_MaxLimit(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)

	repo.On("List", mock.Anything, mock.Anything, 100, 0).Return([]*note.Note{}, int64(0), nil)

	w, c := newContext(t, http.MethodGet, "/notes?limit=200&offset=0", "")
	h.List(c)
	_ = w

	assert.Equal(t, http.StatusOK, c.Writer.Status())
}

func TestListNotes_InvalidLimitOffset(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)

	repo.On("List", mock.Anything, mock.Anything, 20, 0).Return([]*note.Note{}, int64(0), nil)

	w, c := newContext(t, http.MethodGet, "/notes?limit=abc&offset=-5", "")
	h.List(c)
	_ = w

	assert.Equal(t, http.StatusOK, c.Writer.Status())
}

func TestListNotes_RepoError(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)

	repo.On("List", mock.Anything, mock.Anything, 20, 0).Return(nil, int64(0), assert.AnError)

	w, c := newContext(t, http.MethodGet, "/notes", "")
	h.List(c)
	_ = w

	assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())
}

func TestBookmarklet_Success(t *testing.T) {
	h, repo, tq, _, _, _ := setupUnitHandler(t)
	userID := uuid.New()

	repo.On("Save", mock.Anything, mock.MatchedBy(func(n *note.Note) bool {
		return n.Title().String() == "Example Page" &&
			n.Type() == "asteroid" &&
			strings.Contains(n.Content().String(), "## [Example Page](https://example.com)") &&
			strings.Contains(n.Content().String(), "selected text")
	})).Return(nil)
	tq.On("EnqueueExtractKeywords", mock.Anything, mock.AnythingOfType("string"), 10).Return(nil)
	tq.On("EnqueueComputeEmbedding", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	tq.On("EnqueueRecalculateLinkWeights", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("time.Duration")).Return(nil)

	body := `{"title":"Example Page","url":"https://example.com","text":"selected text"}`
	w, c := newContext(t, http.MethodPost, "/import/bookmarklet", body, userID)
	h.Bookmarklet(c)

	assert.Equal(t, http.StatusCreated, c.Writer.Status())
	assert.Contains(t, w.Body.String(), "Example Page")
	assert.Contains(t, w.Body.String(), "asteroid")
	repo.AssertExpectations(t)
	tq.AssertExpectations(t)
}

func TestBookmarklet_DefaultTypeAndTruncation(t *testing.T) {
	h, repo, tq, _, _, _ := setupUnitHandler(t)
	userID := uuid.New()

	hugeText := strings.Repeat("x", 20000)

	repo.On("Save", mock.Anything, mock.MatchedBy(func(n *note.Note) bool {
		return n.Type() == "asteroid" && len(n.Content().String()) <= 10000
	})).Return(nil)
	tq.On("EnqueueExtractKeywords", mock.Anything, mock.AnythingOfType("string"), 10).Return(nil)
	tq.On("EnqueueComputeEmbedding", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	tq.On("EnqueueRecalculateLinkWeights", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("time.Duration")).Return(nil)

	body := fmt.Sprintf(`{"title":"Long Page","url":"https://example.com","text":"%s"}`, hugeText)
	w, c := newContext(t, http.MethodPost, "/import/bookmarklet", body, userID)
	h.Bookmarklet(c)

	assert.Equal(t, http.StatusCreated, c.Writer.Status())
	repo.AssertExpectations(t)
	tq.AssertExpectations(t)
	_ = w
}

func TestBookmarklet_Unauthorized(t *testing.T) {
	h, _, _, _, _, _ := setupUnitHandler(t)

	body := `{"title":"Example Page","url":"https://example.com","text":"text"}`
	w, c := newContext(t, http.MethodPost, "/import/bookmarklet", body)
	h.Bookmarklet(c)

	assert.Equal(t, http.StatusUnauthorized, c.Writer.Status())
	_ = w
}

func TestBookmarklet_ValidationError(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	userID := uuid.New()

	body := `{"title":"","url":"not-a-url","text":""}`
	w, c := newContext(t, http.MethodPost, "/import/bookmarklet", body, userID)
	h.Bookmarklet(c)

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	repo.AssertNotCalled(t, "Save")
	_ = w
}
