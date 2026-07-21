//go:build !integration
// +build !integration

package share

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domainnote "knowledge-graph/internal/domain/note"
	domainshare "knowledge-graph/internal/domain/share"
	domainuser "knowledge-graph/internal/domain/user"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockNoteRepo struct {
	mock.Mock
}

func (m *mockNoteRepo) Save(ctx context.Context, note *domainnote.Note) error {
	return m.Called(ctx, note).Error(0)
}

func (m *mockNoteRepo) FindByID(ctx context.Context, id uuid.UUID) (*domainnote.Note, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainnote.Note), args.Error(1)
}

func (m *mockNoteRepo) Delete(ctx context.Context, id uuid.UUID) error         { return nil }
func (m *mockNoteRepo) DeleteBatch(ctx context.Context, ids []uuid.UUID) error { return nil }
func (m *mockNoteRepo) Restore(ctx context.Context, id uuid.UUID) error        { return nil }
func (m *mockNoteRepo) List(ctx context.Context, limit, offset int) ([]*domainnote.Note, int64, error) {
	return nil, 0, nil
}
func (m *mockNoteRepo) Search(ctx context.Context, query string, limit, offset int) ([]*domainnote.Note, int64, error) {
	return nil, 0, nil
}
func (m *mockNoteRepo) FindAll(ctx context.Context) ([]*domainnote.Note, error) { return nil, nil }
func (m *mockNoteRepo) FindAllPaginated(ctx context.Context, limit, offset int) ([]*domainnote.Note, int64, error) {
	return nil, 0, nil
}

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domainuser.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainuser.User), args.Error(1)
}
func (m *mockUserRepo) FindByLogin(ctx context.Context, login string) (*domainuser.User, error) {
	return nil, nil
}
func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domainuser.User, error) {
	return nil, nil
}
func (m *mockUserRepo) Create(ctx context.Context, u *domainuser.User) error { return nil }
func (m *mockUserRepo) Update(ctx context.Context, u *domainuser.User) error { return nil }
func (m *mockUserRepo) SoftDelete(ctx context.Context, id uuid.UUID) error   { return nil }
func (m *mockUserRepo) EmailExists(ctx context.Context, email string, excludeID uuid.UUID) (bool, error) {
	return false, nil
}

type mockShareRepo struct {
	mock.Mock
}

func (m *mockShareRepo) FindShareByNoteAndUser(ctx context.Context, noteID, sharedWithUserID uuid.UUID) (*domainshare.NoteShare, error) {
	args := m.Called(ctx, noteID, sharedWithUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainshare.NoteShare), args.Error(1)
}
func (m *mockShareRepo) CreateShare(ctx context.Context, share *domainshare.NoteShare) error {
	return m.Called(ctx, share).Error(0)
}
func (m *mockShareRepo) UpdateShare(ctx context.Context, share *domainshare.NoteShare) error {
	return m.Called(ctx, share).Error(0)
}
func (m *mockShareRepo) ListSharesByNote(ctx context.Context, noteID uuid.UUID) ([]domainshare.NoteShare, error) {
	args := m.Called(ctx, noteID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domainshare.NoteShare), args.Error(1)
}
func (m *mockShareRepo) RevokeShare(ctx context.Context, noteID, shareID, sharedByUserID uuid.UUID) (bool, error) {
	args := m.Called(ctx, noteID, shareID, sharedByUserID)
	return args.Bool(0), args.Error(1)
}
func (m *mockShareRepo) CreateShareLink(ctx context.Context, link *domainshare.ShareLink) error {
	return m.Called(ctx, link).Error(0)
}
func (m *mockShareRepo) FindActiveShareLinkByToken(ctx context.Context, token string) (*domainshare.ShareLink, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainshare.ShareLink), args.Error(1)
}
func (m *mockShareRepo) RevokeShareLink(ctx context.Context, linkID, sharedByUserID uuid.UUID) (bool, error) {
	args := m.Called(ctx, linkID, sharedByUserID)
	return args.Bool(0), args.Error(1)
}
func (m *mockShareRepo) ListShareLinksByNote(ctx context.Context, noteID uuid.UUID) ([]domainshare.ShareLink, error) {
	args := m.Called(ctx, noteID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domainshare.ShareLink), args.Error(1)
}
func (m *mockShareRepo) IncrementShareLinkUsage(ctx context.Context, linkID uuid.UUID) error {
	return m.Called(ctx, linkID).Error(0)
}

func setupShareHandler(t *testing.T) (*Handler, *mockNoteRepo, *mockUserRepo, *mockShareRepo) {
	gin.SetMode(gin.TestMode)
	noteRepo := new(mockNoteRepo)
	userRepo := new(mockUserRepo)
	shareRepo := new(mockShareRepo)
	return NewHandler(noteRepo, userRepo, shareRepo), noteRepo, userRepo, shareRepo
}

func newTestNoteWithCreator(t *testing.T, creatorID uuid.UUID) *domainnote.Note {
	title, err := domainnote.NewTitle("Test Note")
	require.NoError(t, err)
	content, err := domainnote.NewContent("content")
	require.NoError(t, err)
	metadata, err := domainnote.NewMetadata(nil)
	require.NoError(t, err)
	return domainnote.NewNoteWithCreator(title, content, "star", metadata, creatorID)
}

func newTestUser(t *testing.T, id uuid.UUID, login string) *domainuser.User {
	now := time.Now()
	u, err := domainuser.NewUser(id, login, login+"@example.com", "hash", "user", now, now, nil)
	require.NoError(t, err)
	return u
}

func TestShareNote_Success(t *testing.T) {
	h, noteRepo, userRepo, shareRepo := setupShareHandler(t)

	ownerID := uuid.New()
	targetID := uuid.New()
	noteID := uuid.New()
	note := newTestNoteWithCreator(t, ownerID)
	targetUser := newTestUser(t, targetID, "target")

	body, _ := json.Marshal(ShareNoteRequest{UserID: targetID.String(), Permission: "read"})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/notes/"+noteID.String()+"/share", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: noteID.String()}}
	c.Set(middleware.ContextUserIDKey, ownerID)

	noteRepo.On("FindByID", mock.Anything, noteID).Return(note, nil)
	userRepo.On("FindByID", mock.Anything, targetID).Return(targetUser, nil)
	shareRepo.On("FindShareByNoteAndUser", mock.Anything, noteID, targetID).Return(nil, nil)
	shareRepo.On("CreateShare", mock.Anything, mock.AnythingOfType("*share.NoteShare")).Return(nil)

	h.ShareNote(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "target")
}

func TestCreateShareLink_Success(t *testing.T) {
	h, noteRepo, _, shareRepo := setupShareHandler(t)

	ownerID := uuid.New()
	noteID := uuid.New()
	note := newTestNoteWithCreator(t, ownerID)

	body, _ := json.Marshal(CreateShareLinkRequest{Permission: "read"})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/notes/"+noteID.String()+"/share-link", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: noteID.String()}}
	c.Set(middleware.ContextUserIDKey, ownerID)

	noteRepo.On("FindByID", mock.Anything, noteID).Return(note, nil)
	shareRepo.On("CreateShareLink", mock.Anything, mock.AnythingOfType("*share.ShareLink")).Return(nil)

	h.CreateShareLink(c)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestRevokeShareLink_Success(t *testing.T) {
	h, _, _, shareRepo := setupShareHandler(t)

	userID := uuid.New()
	linkID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/share-links/"+linkID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: linkID.String()}}
	c.Set(middleware.ContextUserIDKey, userID)

	shareRepo.On("RevokeShareLink", mock.Anything, linkID, userID).Return(true, nil)

	h.RevokeShareLink(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListNoteShares_Success(t *testing.T) {
	h, noteRepo, _, shareRepo := setupShareHandler(t)

	ownerID := uuid.New()
	noteID := uuid.New()
	note := newTestNoteWithCreator(t, ownerID)

	share, err := domainshare.NewNoteShare(uuid.New(), noteID, ownerID, uuid.New(), "read", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String()+"/shares", nil)
	c.Params = gin.Params{{Key: "id", Value: noteID.String()}}
	c.Set(middleware.ContextUserIDKey, ownerID)

	noteRepo.On("FindByID", mock.Anything, noteID).Return(note, nil)
	shareRepo.On("ListSharesByNote", mock.Anything, noteID).Return([]domainshare.NoteShare{*share}, nil)
	shareRepo.On("ListShareLinksByNote", mock.Anything, noteID).Return([]domainshare.ShareLink{}, nil)

	h.ListNoteShares(c)

	assert.Equal(t, http.StatusOK, w.Code)
}
