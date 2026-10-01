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
	"unicode/utf8"

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

func (m *noteRepoMock) FindComets(ctx context.Context, userID uuid.UUID) ([]*note.Note, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*note.Note), args.Error(1)
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

type cometRemindCall struct {
	noteID   uuid.UUID
	remindAt time.Time
}

type taskQueueMock struct {
	mock.Mock
	normalizeCalls   []string
	cleanupCalls     []string
	cometRemindCalls []cometRemindCall
}

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

func (m *taskQueueMock) EnqueueBackupOnNoteChange(ctx context.Context) error {
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
	importSvc := importer.NewService(repo, cache, nil, nil)
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
	return note.NewNote(ttl, cnt, note.MustType(noteType), meta)
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
	// NLP-4: creating a note enqueues normalization (the queue client
	// itself gates on nlp.pipeline.enabled — the handler always calls).
	assert.Len(t, tq.normalizeCalls, 1)
}

func TestCreateBatch_EnqueuesNormalize(t *testing.T) {
	h, repo, tq, _, _, _ := setupUnitHandler(t)

	repo.On("Save", mock.Anything, mock.AnythingOfType("*note.Note")).Return(nil)
	tq.On("EnqueueExtractKeywords", mock.Anything, mock.AnythingOfType("string"), 10).Return(nil)
	tq.On("EnqueueComputeEmbedding", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	tq.On("EnqueueRecalculateLinkWeights", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("time.Duration")).Return(nil)

	body := `{"notes":[{"title":"Batch 1","content":"one"},{"title":"Batch 2","content":"two"}]}`
	_, c := newContext(t, http.MethodPost, "/notes/batch/create", body)
	h.CreateBatch(c)

	assert.Equal(t, http.StatusCreated, c.Writer.Status())
	repo.AssertExpectations(t)
	// NLP-4: each note created in a batch goes through normalization.
	assert.Len(t, tq.normalizeCalls, 2)
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

	longContent := strings.Repeat("a", 50001)
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
	importSvc := importer.NewService(repo, cache, nil, nil)
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
	importSvc := importer.NewService(repo, cache, nil, nil)
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
	// NLP-4: editing note text re-enqueues normalization — without it the
	// artifact silently goes stale.
	assert.Len(t, tq.normalizeCalls, 1)
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
	assert.Empty(t, tq.normalizeCalls, "no text change — no normalization")
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

	longContent := strings.Repeat("a", 50001)
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
	h, repo, tq, _, _, _ := setupUnitHandler(t)
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
	// NLP-4 criterion 4: deleting a note cascades to its nlp_artifacts docs.
	assert.Equal(t, []string{n.ID().String()}, tq.cleanupCalls)
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
	h, repo, tq, _, _, _ := setupUnitHandler(t)
	owner := uuid.New()
	n1 := newTestNote(t, "T1", "C1", "star")
	n2 := newTestNote(t, "T2", "C2", "star")
	n1.SetCreatorID(owner)
	n2.SetCreatorID(owner)

	repo.On("FindByID", mock.Anything, n1.ID()).Return(n1, nil)
	repo.On("FindByID", mock.Anything, n2.ID()).Return(n2, nil)
	repo.On("DeleteBatch", mock.Anything, mock.AnythingOfType("[]uuid.UUID")).Return(nil)

	body := fmt.Sprintf(`{"ids":["%s","%s"]}`, n1.ID(), n2.ID())
	w, c := newContext(t, http.MethodPost, "/notes/batch/delete", body, owner)
	h.DeleteBatch(c)
	_ = w

	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	repo.AssertExpectations(t)
	// NLP-4 criterion 4: each deleted note gets its own cleanup task.
	assert.ElementsMatch(t, []string{n1.ID().String(), n2.ID().String()}, tq.cleanupCalls)
}

// SEC-1: a batch containing a foreign note must be refused wholesale —
// nothing is deleted, and the answer is 404 rather than a leak.
func TestDeleteBatchNotes_ForeignNote(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	owner := uuid.New()
	stranger := uuid.New()
	mine := newTestNote(t, "mine", "C", "star")
	theirs := newTestNote(t, "theirs", "C", "star")
	mine.SetCreatorID(owner)
	theirs.SetCreatorID(stranger)

	repo.On("FindByID", mock.Anything, mine.ID()).Return(mine, nil)
	repo.On("FindByID", mock.Anything, theirs.ID()).Return(theirs, nil)

	body := fmt.Sprintf(`{"ids":["%s","%s"]}`, mine.ID(), theirs.ID())
	w, c := newContext(t, http.MethodPost, "/notes/batch/delete", body, owner)
	h.DeleteBatch(c)
	_ = w

	assert.Equal(t, http.StatusNotFound, c.Writer.Status())
	repo.AssertNotCalled(t, "DeleteBatch")
}

func TestDeleteBatchNotes_InvalidBody(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)

	w, c := newContext(t, http.MethodPost, "/notes/batch/delete", `{}`)
	h.DeleteBatch(c)
	_ = w

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	repo.AssertNotCalled(t, "DeleteBatch")
}

func TestDeleteBatchNotes_EmptyIDs(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)

	w, c := newContext(t, http.MethodPost, "/notes/batch/delete", `{"ids":[]}`)
	h.DeleteBatch(c)
	_ = w

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	repo.AssertNotCalled(t, "DeleteBatch")
}

func TestDeleteBatchNotes_InvalidUUID(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)

	w, c := newContext(t, http.MethodPost, "/notes/batch/delete", `{"ids":["not-a-uuid"]}`)
	h.DeleteBatch(c)
	_ = w

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	repo.AssertNotCalled(t, "DeleteBatch")
}

func TestDeleteBatchNotes_RepoError(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	id := uuid.New()

	// The id resolves to nothing — missing ids keep their no-op semantics.
	repo.On("FindByID", mock.Anything, id).Return(nil, nil)
	repo.On("DeleteBatch", mock.Anything, mock.AnythingOfType("[]uuid.UUID")).Return(assert.AnError)

	body := fmt.Sprintf(`{"ids":["%s"]}`, id)
	w, c := newContext(t, http.MethodPost, "/notes/batch/delete", body)
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
	rec := newTestNote(t, "Rec", "Content", "planet")

	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)
	repo.On("FindByID", mock.Anything, rec.ID()).Return(rec, nil)
	recRepo.On("GetRecommendations", mock.Anything, n.ID(), 5).Return([]recommendation.Recommendation{
		{NoteID: n.ID(), RecommendedNoteID: rec.ID(), Score: 0.9, UpdatedAt: time.Now().Add(time.Hour)},
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
	rec := newTestNote(t, "Rec", "Content", "planet")

	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)
	repo.On("FindByID", mock.Anything, rec.ID()).Return(rec, nil)
	recRepo.On("GetRecommendations", mock.Anything, n.ID(), 5).Return([]recommendation.Recommendation{
		{NoteID: n.ID(), RecommendedNoteID: rec.ID(), Score: 0.9, UpdatedAt: time.Now().Add(-time.Hour)},
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
	rec := newTestNote(t, "Rec", "Content", "planet")

	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)
	repo.On("FindByID", mock.Anything, rec.ID()).Return(rec, nil)
	recRepo.On("GetRecommendations", mock.Anything, n.ID(), 2).Return([]recommendation.Recommendation{
		{NoteID: n.ID(), RecommendedNoteID: rec.ID(), Score: 0.9, UpdatedAt: time.Now().Add(time.Hour)},
	}, nil)

	w, c := newContext(t, http.MethodGet, "/notes/"+n.ID().String()+"/suggestions?limit=2", "")
	withID(c, n.ID())
	h.GetSuggestions(c)
	_ = w

	assert.Equal(t, http.StatusOK, c.Writer.Status())
	recRepo.AssertExpectations(t)
}

func TestGetSuggestions_SemanticFallback(t *testing.T) {
	h, repo, tq, recRepo, embRepo, _ := setupUnitHandler(t)
	n := newTestNote(t, "Sug", "Content", "star")
	similarID := uuid.New()

	recRepo.On("GetRecommendations", mock.Anything, n.ID(), 5).Return([]recommendation.Recommendation{}, nil)
	embRepo.On("FindSimilarNotes", mock.Anything, n.ID(), 5).Return([]recommendation.SimilarNote{
		{NoteID: similarID, Score: 0.85},
	}, nil)
	repo.On("FindByID", mock.Anything, similarID).Return(newTestNote(t, "Similar", "Body", "planet"), nil)
	tq.On("EnqueueRefreshRecommendations", mock.Anything, n.ID(), mock.AnythingOfType("time.Duration")).Return(nil)

	w, c := newContext(t, http.MethodGet, "/notes/"+n.ID().String()+"/suggestions", "")
	withID(c, n.ID())
	h.GetSuggestions(c)
	_ = w

	assert.Equal(t, http.StatusOK, c.Writer.Status())
	assert.Equal(t, "semantic", w.Header().Get("X-Recommendations-Source"))
	assert.Equal(t, "true", w.Header().Get("X-Recommendations-Stale"))
	assert.Contains(t, w.Body.String(), similarID.String())
	assert.Contains(t, w.Body.String(), "Similar")
	tq.AssertExpectations(t)
}

func TestGetSuggestions_CacheFallback(t *testing.T) {
	h, repo, tq, recRepo, embRepo, cache := setupUnitHandler(t)
	n := newTestNote(t, "Sug", "Content", "star")
	suggestionID := uuid.New()

	recRepo.On("GetRecommendations", mock.Anything, n.ID(), 5).Return([]recommendation.Recommendation{}, nil)
	embRepo.On("FindSimilarNotes", mock.Anything, n.ID(), 5).Return([]recommendation.SimilarNote{}, nil)
	// The cache path re-resolves each entry — deleted/gone notes drop out.
	repo.On("FindByID", mock.Anything, suggestionID).Return(newTestNote(t, "Cached", "Body", "planet"), nil)

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

// NOTE-DELETE-1 regression: a trashed note must not be suggested — the
// precomputed table and the cache both carry stale rows. Mutation "keep
// appending when FindByID returns nil" turns these red.
func TestGetSuggestions_PrecomputedSkipsTrashed(t *testing.T) {
	h, repo, _, recRepo, _, _ := setupUnitHandler(t)
	n := newTestNote(t, "Sug", "Content", "star")
	live := newTestNote(t, "Live", "Content", "planet")
	trashedID := uuid.New()

	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)
	repo.On("FindByID", mock.Anything, live.ID()).Return(live, nil)
	repo.On("FindByID", mock.Anything, trashedID).Return(nil, nil) // in the trash
	recRepo.On("GetRecommendations", mock.Anything, n.ID(), 5).Return([]recommendation.Recommendation{
		{NoteID: n.ID(), RecommendedNoteID: live.ID(), Score: 0.9, UpdatedAt: time.Now().Add(time.Hour)},
		{NoteID: n.ID(), RecommendedNoteID: trashedID, Score: 0.8, UpdatedAt: time.Now().Add(time.Hour)},
	}, nil)

	w, c := newContext(t, http.MethodGet, "/notes/"+n.ID().String()+"/suggestions", "")
	withID(c, n.ID())
	h.GetSuggestions(c)
	_ = w

	assert.Equal(t, http.StatusOK, c.Writer.Status())
	assert.Contains(t, w.Body.String(), live.ID().String())
	assert.NotContains(t, w.Body.String(), trashedID.String(), "trashed note must not be suggested")
}

func TestGetSuggestions_CacheSkipsTrashed(t *testing.T) {
	h, repo, tq, recRepo, embRepo, cache := setupUnitHandler(t)
	n := newTestNote(t, "Sug", "Content", "star")
	liveID := uuid.New()
	trashedID := uuid.New()

	recRepo.On("GetRecommendations", mock.Anything, n.ID(), 5).Return([]recommendation.Recommendation{}, nil)
	embRepo.On("FindSimilarNotes", mock.Anything, n.ID(), 5).Return([]recommendation.SimilarNote{}, nil)
	repo.On("FindByID", mock.Anything, liveID).Return(newTestNote(t, "Live", "Body", "planet"), nil)
	repo.On("FindByID", mock.Anything, trashedID).Return(nil, nil)

	cached := fmt.Sprintf(`[{"note_id":%q,"score":0.7},{"note_id":%q,"score":0.6}]`, liveID.String(), trashedID.String())
	_ = cache.Set(context.Background(), "recommendations:"+n.ID().String(), cached, 0)
	tq.On("EnqueueRefreshRecommendations", mock.Anything, n.ID(), mock.AnythingOfType("time.Duration")).Return(nil)

	w, c := newContext(t, http.MethodGet, "/notes/"+n.ID().String()+"/suggestions", "")
	withID(c, n.ID())
	h.GetSuggestions(c)
	_ = w

	assert.Equal(t, http.StatusOK, c.Writer.Status())
	assert.Equal(t, "redis", w.Header().Get("X-Recommendations-Source"))
	assert.Contains(t, w.Body.String(), liveID.String())
	assert.NotContains(t, w.Body.String(), trashedID.String(), "cached suggestion for a trashed note must be dropped")
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
	importSvc := importer.NewService(repo, cache, nil, nil)
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
	// NLP-4: bookmarklet-created notes go through normalization too.
	assert.Len(t, tq.normalizeCalls, 1)
}

func TestBookmarklet_DefaultTypeAndTruncation(t *testing.T) {
	h, repo, tq, _, _, _ := setupUnitHandler(t)
	userID := uuid.New()

	// URL-HEADING-1: the domain Content limit is 50 000 runes and the request
	// DTO rejects anything above 50 000 chars — so the old 10 000-byte cut no
	// longer fires here; truncation itself stays covered by TestBuildContent.
	hugeText := strings.Repeat("x", 40000)

	repo.On("Save", mock.Anything, mock.MatchedBy(func(n *note.Note) bool {
		return n.Type() == "asteroid" && utf8.RuneCountInString(n.Content().String()) <= 50000
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

// NLP-4 criteria 4-5: normalize enqueue on content-changing writes and the
// deletion cascade are observable — the mock records both so removing the
// call sites reddens the delete/create tests below.
func (m *taskQueueMock) EnqueueNormalizeNote(ctx context.Context, noteID string) error {
	m.normalizeCalls = append(m.normalizeCalls, noteID)
	return nil
}

func (m *taskQueueMock) EnqueueNlpArtifactsCleanup(ctx context.Context, noteID string) error {
	m.cleanupCalls = append(m.cleanupCalls, noteID)
	return nil
}

// COMET-1 stage C: recorded so tests can assert a comet with a reminder
// enqueues comet:remind and a note without one does not.
func (m *taskQueueMock) EnqueueCometRemind(ctx context.Context, noteID uuid.UUID, remindAt time.Time) error {
	m.cometRemindCalls = append(m.cometRemindCalls, cometRemindCall{noteID: noteID, remindAt: remindAt})
	return nil
}

// --- COMET-1: поля планирования и «Ближайшие дела» ---

func TestCreateNote_CometFields(t *testing.T) {
	h, repo, tq, _, _, _ := setupUnitHandler(t)
	userID := uuid.New()

	var saved *note.Note
	repo.On("Save", mock.Anything, mock.AnythingOfType("*note.Note")).
		Run(func(args mock.Arguments) { saved = args.Get(1).(*note.Note) }).Return(nil)
	tq.On("EnqueueExtractKeywords", mock.Anything, mock.AnythingOfType("string"), 10).Return(nil)
	tq.On("EnqueueComputeEmbedding", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	tq.On("EnqueueRecalculateLinkWeights", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("time.Duration")).Return(nil)
	tq.On("EnqueueRefreshRecommendations", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	body := `{"title":"Дело","content":"Сходить к врачу","type":"comet","due_at":"2026-10-05T10:00:00Z","remind_before_seconds":3600}`
	w, c := newContext(t, http.MethodPost, "/notes", body, userID)
	h.Create(c)

	assert.Equal(t, http.StatusCreated, c.Writer.Status())
	require.NotNil(t, saved)
	require.NotNil(t, saved.DueAt())
	assert.Equal(t, int64(3600), *saved.RemindBeforeSeconds())
	assert.Nil(t, saved.DoneAt())
	assert.Contains(t, w.Body.String(), "due_at")
}

func TestCreateNote_RemindWithoutDue_BadRequest(t *testing.T) {
	h, repo, tq, _, _, _ := setupUnitHandler(t)
	_ = tq
	userID := uuid.New()

	body := `{"title":"Дело","content":"x","type":"comet","remind_before_seconds":3600}`
	_, c := newContext(t, http.MethodPost, "/notes", body, userID)
	h.Create(c)

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	repo.AssertNotCalled(t, "Save")
	tq.AssertNotCalled(t, "EnqueueExtractKeywords")
}

func TestCreateNote_NegativeRemind_BadRequest(t *testing.T) {
	h, _, _, _, _, _ := setupUnitHandler(t)
	userID := uuid.New()

	body := `{"title":"Дело","content":"x","type":"comet","due_at":"2026-10-05T10:00:00Z","remind_before_seconds":-5}`
	_, c := newContext(t, http.MethodPost, "/notes", body, userID)
	h.Create(c)

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
}

// COMET-1 stage C: creating a comet with due_at + reminder enqueues
// comet:remind for due_at − remind_before.
func TestCreateNote_CometEnqueuesReminder(t *testing.T) {
	h, repo, tq, _, _, _ := setupUnitHandler(t)
	userID := uuid.New()

	repo.On("Save", mock.Anything, mock.AnythingOfType("*note.Note")).Return(nil)
	tq.On("EnqueueExtractKeywords", mock.Anything, mock.AnythingOfType("string"), 10).Return(nil)
	tq.On("EnqueueComputeEmbedding", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	tq.On("EnqueueNormalizeNote", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	tq.On("EnqueueRecalculateLinkWeights", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("time.Duration")).Return(nil)
	tq.On("EnqueueRefreshRecommendations", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	due := time.Date(2026, 10, 5, 10, 0, 0, 0, time.UTC)
	body := fmt.Sprintf(`{"title":"Дело","content":"x","type":"comet","due_at":%q,"remind_before_seconds":3600}`, due.Format(time.RFC3339))
	_, c := newContext(t, http.MethodPost, "/notes", body, userID)
	h.Create(c)

	assert.Equal(t, http.StatusCreated, c.Writer.Status())
	require.Len(t, tq.cometRemindCalls, 1, "comet with due_at + reminder must enqueue comet:remind")
	assert.Equal(t, due.Add(-time.Hour), tq.cometRemindCalls[0].remindAt)
}

// A comet without a reminder must not enqueue comet:remind.
func TestCreateNote_CometWithoutReminderNoEnqueue(t *testing.T) {
	h, repo, tq, _, _, _ := setupUnitHandler(t)
	userID := uuid.New()

	repo.On("Save", mock.Anything, mock.AnythingOfType("*note.Note")).Return(nil)
	tq.On("EnqueueExtractKeywords", mock.Anything, mock.AnythingOfType("string"), 10).Return(nil)
	tq.On("EnqueueComputeEmbedding", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	tq.On("EnqueueNormalizeNote", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	tq.On("EnqueueRecalculateLinkWeights", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("time.Duration")).Return(nil)
	tq.On("EnqueueRefreshRecommendations", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	body := `{"title":"Дело","content":"x","type":"comet","due_at":"2026-10-05T10:00:00Z"}`
	_, c := newContext(t, http.MethodPost, "/notes", body, userID)
	h.Create(c)

	assert.Equal(t, http.StatusCreated, c.Writer.Status())
	assert.Empty(t, tq.cometRemindCalls)
}

func TestUpdateNote_CometFieldsMergeAndClear(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	userID := uuid.New()
	n := newTestNote(t, "Comet", "Content", "comet")
	due := time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second)
	var remind int64 = 900
	require.NoError(t, n.SetCometFields(&due, &remind, nil))

	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)
	repo.On("Save", mock.Anything, mock.AnythingOfType("*note.Note")).Return(nil)

	// Only done_at present — due_at and remind keep their stored values.
	body := `{"done_at":"` + time.Now().UTC().Format(time.RFC3339) + `"}`
	_, c := newContext(t, http.MethodPut, "/notes/"+n.ID().String(), body, userID)
	withID(c, n.ID())
	h.Update(c)

	assert.Equal(t, http.StatusOK, c.Writer.Status())
	assert.NotNil(t, n.DoneAt(), "done_at must be applied")
	assert.Equal(t, due, *n.DueAt(), "absent due_at keeps the stored value")
	assert.Equal(t, int64(900), *n.RemindBeforeSeconds())
}

func TestUpdateNote_ClearDueAt(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	userID := uuid.New()
	n := newTestNote(t, "Comet", "Content", "comet")
	due := time.Now().Add(24 * time.Hour)
	var remind int64 = 900
	require.NoError(t, n.SetCometFields(&due, &remind, nil))

	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)
	repo.On("Save", mock.Anything, mock.AnythingOfType("*note.Note")).Return(nil)

	// null clears both date and reminder — a reminder without a date is rejected.
	body := `{"due_at":null,"remind_before_seconds":null}`
	_, c := newContext(t, http.MethodPut, "/notes/"+n.ID().String(), body, userID)
	withID(c, n.ID())
	h.Update(c)

	assert.Equal(t, http.StatusOK, c.Writer.Status())
	assert.Nil(t, n.DueAt())
	assert.Nil(t, n.RemindBeforeSeconds())
}

func TestUpdateNote_RemindWithoutDue_BadRequest(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	userID := uuid.New()
	n := newTestNote(t, "Comet", "Content", "comet")

	repo.On("FindByID", mock.Anything, n.ID()).Return(n, nil)

	// remind_before on a note without due_at → 400, nothing saved.
	body := `{"remind_before_seconds":3600}`
	_, c := newContext(t, http.MethodPut, "/notes/"+n.ID().String(), body, userID)
	withID(c, n.ID())
	h.Update(c)

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	repo.AssertNotCalled(t, "Save")
}

func TestComets_GroupsOverdueUpcomingUndated(t *testing.T) {
	h, repo, _, _, _, _ := setupUnitHandler(t)
	userID := uuid.New()
	now := time.Now()

	mk := func(title string, dueAt *time.Time) *note.Note {
		n := newTestNote(t, title, "c", "comet")
		n.SetCreatorID(userID)
		require.NoError(t, n.SetCometFields(dueAt, nil, nil))
		return n
	}
	overdue := mk("overdue", ptrTime(now.Add(-time.Hour)))
	upcoming := mk("upcoming", ptrTime(now.Add(time.Hour)))
	undated := mk("undated", nil)

	// Repo returns due_at ASC NULLS LAST — overdue naturally first.
	repo.On("FindComets", mock.Anything, userID).Return([]*note.Note{overdue, upcoming, undated}, nil)

	w, c := newContext(t, http.MethodGet, "/notes/comets", "", userID)
	h.Comets(c)

	assert.Equal(t, http.StatusOK, c.Writer.Status())
	body := w.Body.String()
	ioverdue := strings.Index(body, "overdue")
	iupcoming := strings.Index(body, "upcoming")
	iundated := strings.Index(body, "undated")
	require.GreaterOrEqual(t, ioverdue, 0)
	assert.Contains(t, body[ioverdue:], `"`+overdue.ID().String()+`"`)
	assert.Contains(t, body[iupcoming:], `"`+upcoming.ID().String()+`"`)
	assert.Contains(t, body[iundated:], `"`+undated.ID().String()+`"`)
}

func ptrTime(t time.Time) *time.Time { return &t }
