//go:build !integration
// +build !integration

package share

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
func (m *mockNoteRepo) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domainnote.Note, int64, error) {
	return nil, 0, nil
}
func (m *mockNoteRepo) Search(ctx context.Context, userID uuid.UUID, query string, limit, offset int) ([]*domainnote.Note, int64, error) {
	return nil, 0, nil
}
func (m *mockNoteRepo) FindAll(ctx context.Context) ([]*domainnote.Note, error) { return nil, nil }
func (m *mockNoteRepo) FindAllPaginated(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domainnote.Note, int64, error) {
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

func TestAccessSharedNote_Branches(t *testing.T) {
	ownerID := uuid.New()
	noteID := uuid.New()
	token := `test-token`
	linkID := uuid.New()
	var emptyToken string
	note := newTestNoteWithCreator(t, ownerID)

	activeLink, err := domainshare.NewShareLink(linkID, noteID, ownerID, token, `read`, nil, nil, 0)
	require.NoError(t, err)

	expiredTime := time.Now().Add(-time.Hour)
	expiredLink, err := domainshare.NewShareLink(uuid.New(), noteID, ownerID, token, `read`, &expiredTime, nil, 0)
	require.NoError(t, err)

	maxUses := 1
	maxUsesLink, err := domainshare.NewShareLink(uuid.New(), noteID, ownerID, token, `read`, nil, &maxUses, 1)
	require.NoError(t, err)

	tests := []struct {
		name       string
		token      string
		setupMocks func(*mockNoteRepo, *mockShareRepo)
		wantStatus int
		wantSub    string
	}{
		{`empty token`, emptyToken, nil, http.StatusBadRequest, `share token required`},
		{`link find error`, token, func(nr *mockNoteRepo, sr *mockShareRepo) {
			sr.On(`FindActiveShareLinkByToken`, mock.Anything, token).Return(nil, errors.New(`db error`))
		}, http.StatusInternalServerError, `failed to fetch share link`},
		{`link not found`, token, func(nr *mockNoteRepo, sr *mockShareRepo) {
			sr.On(`FindActiveShareLinkByToken`, mock.Anything, token).Return(nil, nil)
		}, http.StatusNotFound, `invalid or expired share link`},
		{`expired link`, token, func(nr *mockNoteRepo, sr *mockShareRepo) {
			sr.On(`FindActiveShareLinkByToken`, mock.Anything, token).Return(expiredLink, nil)
		}, http.StatusGone, `share link has expired`},
		{`max uses reached`, token, func(nr *mockNoteRepo, sr *mockShareRepo) {
			sr.On(`FindActiveShareLinkByToken`, mock.Anything, token).Return(maxUsesLink, nil)
		}, http.StatusGone, `maximum uses`},
		{`note find error`, token, func(nr *mockNoteRepo, sr *mockShareRepo) {
			sr.On(`FindActiveShareLinkByToken`, mock.Anything, token).Return(activeLink, nil)
			sr.On(`IncrementShareLinkUsage`, mock.Anything, linkID).Return(nil)
			nr.On(`FindByID`, mock.Anything, noteID).Return(nil, errors.New(`db error`))
		}, http.StatusInternalServerError, `failed to fetch note`},
		{`note not found`, token, func(nr *mockNoteRepo, sr *mockShareRepo) {
			sr.On(`FindActiveShareLinkByToken`, mock.Anything, token).Return(activeLink, nil)
			sr.On(`IncrementShareLinkUsage`, mock.Anything, linkID).Return(nil)
			nr.On(`FindByID`, mock.Anything, noteID).Return(nil, nil)
		}, http.StatusNotFound, `note not found`},
		{`success`, token, func(nr *mockNoteRepo, sr *mockShareRepo) {
			sr.On(`FindActiveShareLinkByToken`, mock.Anything, token).Return(activeLink, nil)
			sr.On(`IncrementShareLinkUsage`, mock.Anything, linkID).Return(nil)
			nr.On(`FindByID`, mock.Anything, noteID).Return(note, nil)
		}, http.StatusOK, `Test Note`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, nr, _, sr := setupShareHandler(t)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, `/shares/`+tt.token, nil)
			c.Params = gin.Params{{Key: `token`, Value: tt.token}}
			if tt.setupMocks != nil {
				tt.setupMocks(nr, sr)
			}
			h.AccessSharedNote(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if len(tt.wantSub) > 0 {
				assert.Contains(t, w.Body.String(), tt.wantSub)
			}
		})
	}
}

func TestRevokeShare_Branches(t *testing.T) {
	ownerID := uuid.New()
	otherID := uuid.New()
	noteID := uuid.New()
	shareID := uuid.New()
	note := newTestNoteWithCreator(t, ownerID)

	tests := []struct {
		name       string
		userID     uuid.UUID
		paramID    string
		shareParam string
		setupMocks func(*mockNoteRepo, *mockShareRepo)
		wantStatus int
		wantSub    string
		noAuth     bool
	}{
		{`unauthorized`, uuid.Nil, noteID.String(), shareID.String(), nil, http.StatusUnauthorized, `authentication required`, true},
		{`invalid note ID`, ownerID, `bad-uuid`, shareID.String(), nil, http.StatusBadRequest, `invalid note ID`, false},
		{`invalid share ID`, ownerID, noteID.String(), `bad-uuid`, nil, http.StatusBadRequest, `invalid share ID`, false},
		{`note find error`, ownerID, noteID.String(), shareID.String(), func(nr *mockNoteRepo, _ *mockShareRepo) {
			nr.On(`FindByID`, mock.Anything, noteID).Return(nil, errors.New(`db error`))
		}, http.StatusInternalServerError, `failed to fetch note`, false},
		{`note not found`, ownerID, noteID.String(), shareID.String(), func(nr *mockNoteRepo, _ *mockShareRepo) {
			nr.On(`FindByID`, mock.Anything, noteID).Return(nil, nil)
		}, http.StatusNotFound, `note not found`, false},
		{`not creator`, otherID, noteID.String(), shareID.String(), func(nr *mockNoteRepo, _ *mockShareRepo) {
			nr.On(`FindByID`, mock.Anything, noteID).Return(note, nil)
		}, http.StatusForbidden, `only the creator can revoke shares`, false},
		{`revoke error`, ownerID, noteID.String(), shareID.String(), func(nr *mockNoteRepo, sr *mockShareRepo) {
			nr.On(`FindByID`, mock.Anything, noteID).Return(note, nil)
			sr.On(`RevokeShare`, mock.Anything, noteID, shareID, ownerID).Return(false, errors.New(`db error`))
		}, http.StatusInternalServerError, `failed to revoke share`, false},
		{`share not found`, ownerID, noteID.String(), shareID.String(), func(nr *mockNoteRepo, sr *mockShareRepo) {
			nr.On(`FindByID`, mock.Anything, noteID).Return(note, nil)
			sr.On(`RevokeShare`, mock.Anything, noteID, shareID, ownerID).Return(false, nil)
		}, http.StatusNotFound, `share not found`, false},
		{`success`, ownerID, noteID.String(), shareID.String(), func(nr *mockNoteRepo, sr *mockShareRepo) {
			nr.On(`FindByID`, mock.Anything, noteID).Return(note, nil)
			sr.On(`RevokeShare`, mock.Anything, noteID, shareID, ownerID).Return(true, nil)
		}, http.StatusOK, ``, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, nr, _, sr := setupShareHandler(t)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, `/notes/`+tt.paramID+`/shares/`+tt.shareParam, nil)
			c.Params = gin.Params{
				{Key: `id`, Value: tt.paramID},
				{Key: `shareId`, Value: tt.shareParam},
			}
			if !tt.noAuth {
				c.Set(middleware.ContextUserIDKey, tt.userID)
			}
			if tt.setupMocks != nil {
				tt.setupMocks(nr, sr)
			}
			h.RevokeShare(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if len(tt.wantSub) > 0 {
				assert.Contains(t, w.Body.String(), tt.wantSub)
			}
		})
	}
}

func TestRevokeShareLink_Branches(t *testing.T) {
	userID := uuid.New()
	linkID := uuid.New()

	tests := []struct {
		name       string
		paramID    string
		setupMocks func(*mockShareRepo)
		wantStatus int
		wantSub    string
		noAuth     bool
	}{
		{`unauthorized`, linkID.String(), nil, http.StatusUnauthorized, `authentication required`, true},
		{`invalid link ID`, `bad-uuid`, nil, http.StatusBadRequest, `invalid share link ID`, false},
		{`revoke error`, linkID.String(), func(sr *mockShareRepo) {
			sr.On(`RevokeShareLink`, mock.Anything, linkID, userID).Return(false, errors.New(`db error`))
		}, http.StatusInternalServerError, `failed to revoke share link`, false},
		{`not found`, linkID.String(), func(sr *mockShareRepo) {
			sr.On(`RevokeShareLink`, mock.Anything, linkID, userID).Return(false, nil)
		}, http.StatusNotFound, `share link not found`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, _, _, sr := setupShareHandler(t)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, `/share-links/`+tt.paramID, nil)
			c.Params = gin.Params{{Key: `id`, Value: tt.paramID}}
			if !tt.noAuth {
				c.Set(middleware.ContextUserIDKey, userID)
			}
			if tt.setupMocks != nil {
				tt.setupMocks(sr)
			}
			h.RevokeShareLink(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if len(tt.wantSub) > 0 {
				assert.Contains(t, w.Body.String(), tt.wantSub)
			}
		})
	}
}

func TestShareNote_Branches(t *testing.T) {
	ownerID := uuid.New()
	otherID := uuid.New()
	noteID := uuid.New()
	targetID := uuid.New()
	note := newTestNoteWithCreator(t, ownerID)
	targetUser := newTestUser(t, targetID, `target`)

	validBody, _ := json.Marshal(map[string]string{`user_id`: targetID.String(), `permission`: `read`})
	updateBody, _ := json.Marshal(map[string]string{`user_id`: targetID.String(), `permission`: `write`})
	invalidPermBody, _ := json.Marshal(map[string]string{`user_id`: targetID.String(), `permission`: `admin`})
	invalidBody := []byte(`{}`)

	existingShare, err := domainshare.NewNoteShare(uuid.New(), noteID, ownerID, targetID, `read`, nil)
	require.NoError(t, err)

	tests := []struct {
		name       string
		userID     uuid.UUID
		paramID    string
		body       []byte
		setupMocks func(*mockNoteRepo, *mockUserRepo, *mockShareRepo)
		wantStatus int
		wantSub    string
		noAuth     bool
	}{
		{`unauthorized`, uuid.Nil, noteID.String(), validBody, nil, http.StatusUnauthorized, `authentication required`, true},
		{`invalid note ID`, ownerID, `bad-uuid`, validBody, nil, http.StatusBadRequest, `invalid note ID`, false},
		{`invalid request body`, ownerID, noteID.String(), invalidBody, nil, http.StatusBadRequest, ``, false},
		{`invalid permission`, ownerID, noteID.String(), invalidPermBody, nil, http.StatusBadRequest, ``, false},
		{`note find error`, ownerID, noteID.String(), validBody, func(nr *mockNoteRepo, _ *mockUserRepo, _ *mockShareRepo) {
			nr.On(`FindByID`, mock.Anything, noteID).Return(nil, errors.New(`db error`))
		}, http.StatusInternalServerError, `failed to fetch note`, false},
		{`note not found`, ownerID, noteID.String(), validBody, func(nr *mockNoteRepo, _ *mockUserRepo, _ *mockShareRepo) {
			nr.On(`FindByID`, mock.Anything, noteID).Return(nil, nil)
		}, http.StatusNotFound, `note not found`, false},
		{`not creator`, otherID, noteID.String(), validBody, func(nr *mockNoteRepo, _ *mockUserRepo, _ *mockShareRepo) {
			nr.On(`FindByID`, mock.Anything, noteID).Return(note, nil)
		}, http.StatusForbidden, `only the creator can share`, false},
		{`target user not found`, ownerID, noteID.String(), validBody, func(nr *mockNoteRepo, ur *mockUserRepo, _ *mockShareRepo) {
			nr.On(`FindByID`, mock.Anything, noteID).Return(note, nil)
			ur.On(`FindByID`, mock.Anything, targetID).Return(nil, nil)
		}, http.StatusNotFound, `target user not found`, false},
		{`target user find error`, ownerID, noteID.String(), validBody, func(nr *mockNoteRepo, ur *mockUserRepo, _ *mockShareRepo) {
			nr.On(`FindByID`, mock.Anything, noteID).Return(note, nil)
			ur.On(`FindByID`, mock.Anything, targetID).Return(nil, errors.New(`db error`))
		}, http.StatusInternalServerError, `failed to fetch target user`, false},
		{`existing share check error`, ownerID, noteID.String(), validBody, func(nr *mockNoteRepo, ur *mockUserRepo, sr *mockShareRepo) {
			nr.On(`FindByID`, mock.Anything, noteID).Return(note, nil)
			ur.On(`FindByID`, mock.Anything, targetID).Return(targetUser, nil)
			sr.On(`FindShareByNoteAndUser`, mock.Anything, noteID, targetID).Return(nil, errors.New(`db error`))
		}, http.StatusInternalServerError, `failed to check existing share`, false},
		{`create share error`, ownerID, noteID.String(), validBody, func(nr *mockNoteRepo, ur *mockUserRepo, sr *mockShareRepo) {
			nr.On(`FindByID`, mock.Anything, noteID).Return(note, nil)
			ur.On(`FindByID`, mock.Anything, targetID).Return(targetUser, nil)
			sr.On(`FindShareByNoteAndUser`, mock.Anything, noteID, targetID).Return(nil, nil)
			sr.On(`CreateShare`, mock.Anything, mock.Anything).Return(errors.New(`db error`))
		}, http.StatusInternalServerError, `failed to create share`, false},
		{`update existing share`, ownerID, noteID.String(), updateBody, func(nr *mockNoteRepo, ur *mockUserRepo, sr *mockShareRepo) {
			nr.On(`FindByID`, mock.Anything, noteID).Return(note, nil)
			ur.On(`FindByID`, mock.Anything, targetID).Return(targetUser, nil)
			sr.On(`FindShareByNoteAndUser`, mock.Anything, noteID, targetID).Return(existingShare, nil)
			sr.On(`UpdateShare`, mock.Anything, mock.Anything).Return(nil)
		}, http.StatusOK, `target`, false},
		{`update share error`, ownerID, noteID.String(), updateBody, func(nr *mockNoteRepo, ur *mockUserRepo, sr *mockShareRepo) {
			nr.On(`FindByID`, mock.Anything, noteID).Return(note, nil)
			ur.On(`FindByID`, mock.Anything, targetID).Return(targetUser, nil)
			sr.On(`FindShareByNoteAndUser`, mock.Anything, noteID, targetID).Return(existingShare, nil)
			sr.On(`UpdateShare`, mock.Anything, mock.Anything).Return(errors.New(`db error`))
		}, http.StatusInternalServerError, `failed to update share`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, nr, ur, sr := setupShareHandler(t)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, `/notes/`+tt.paramID+`/share`, bytes.NewBuffer(tt.body))
			if tt.body != nil {
				c.Request.Header.Set(`Content-Type`, `application/json`)
			}
			c.Params = gin.Params{{Key: `id`, Value: tt.paramID}}
			if !tt.noAuth {
				c.Set(middleware.ContextUserIDKey, tt.userID)
			}
			if tt.setupMocks != nil {
				tt.setupMocks(nr, ur, sr)
			}
			h.ShareNote(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if len(tt.wantSub) > 0 {
				assert.Contains(t, w.Body.String(), tt.wantSub)
			}
		})
	}
}

func TestCreateShareLink_Branches(t *testing.T) {
	ownerID := uuid.New()
	otherID := uuid.New()
	noteID := uuid.New()
	note := newTestNoteWithCreator(t, ownerID)

	validBody, _ := json.Marshal(map[string]string{`permission`: `read`})
	invalidPermBody, _ := json.Marshal(map[string]string{`permission`: `admin`})

	tests := []struct {
		name       string
		userID     uuid.UUID
		paramID    string
		body       []byte
		setupMocks func(*mockNoteRepo, *mockShareRepo)
		wantStatus int
		wantSub    string
		noAuth     bool
	}{
		{`unauthorized`, uuid.Nil, noteID.String(), validBody, nil, http.StatusUnauthorized, `authentication required`, true},
		{`invalid note ID`, ownerID, `bad-uuid`, validBody, nil, http.StatusBadRequest, `invalid note ID`, false},
		{`invalid permission`, ownerID, noteID.String(), invalidPermBody, nil, http.StatusBadRequest, ``, false},
		{`note find error`, ownerID, noteID.String(), validBody, func(nr *mockNoteRepo, _ *mockShareRepo) {
			nr.On(`FindByID`, mock.Anything, noteID).Return(nil, errors.New(`db error`))
		}, http.StatusInternalServerError, `failed to fetch note`, false},
		{`note not found`, ownerID, noteID.String(), validBody, func(nr *mockNoteRepo, _ *mockShareRepo) {
			nr.On(`FindByID`, mock.Anything, noteID).Return(nil, nil)
		}, http.StatusNotFound, `note not found`, false},
		{`not creator`, otherID, noteID.String(), validBody, func(nr *mockNoteRepo, _ *mockShareRepo) {
			nr.On(`FindByID`, mock.Anything, noteID).Return(note, nil)
		}, http.StatusForbidden, `only the creator can create share links`, false},
		{`create share link error`, ownerID, noteID.String(), validBody, func(nr *mockNoteRepo, sr *mockShareRepo) {
			nr.On(`FindByID`, mock.Anything, noteID).Return(note, nil)
			sr.On(`CreateShareLink`, mock.Anything, mock.Anything).Return(errors.New(`db error`))
		}, http.StatusInternalServerError, `failed to create share link`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, nr, _, sr := setupShareHandler(t)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, `/notes/`+tt.paramID+`/share-link`, bytes.NewBuffer(tt.body))
			if tt.body != nil {
				c.Request.Header.Set(`Content-Type`, `application/json`)
			}
			c.Params = gin.Params{{Key: `id`, Value: tt.paramID}}
			if !tt.noAuth {
				c.Set(middleware.ContextUserIDKey, tt.userID)
			}
			if tt.setupMocks != nil {
				tt.setupMocks(nr, sr)
			}
			h.CreateShareLink(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if len(tt.wantSub) > 0 {
				assert.Contains(t, w.Body.String(), tt.wantSub)
			}
		})
	}
}

func TestListNoteShares_Branches(t *testing.T) {
	ownerID := uuid.New()
	otherID := uuid.New()
	noteID := uuid.New()
	note := newTestNoteWithCreator(t, ownerID)

	tests := []struct {
		name       string
		userID     uuid.UUID
		paramID    string
		setupMocks func(*mockNoteRepo, *mockShareRepo)
		wantStatus int
		wantSub    string
		noAuth     bool
	}{
		{`unauthorized`, uuid.Nil, noteID.String(), nil, http.StatusUnauthorized, `authentication required`, true},
		{`invalid note ID`, ownerID, `bad-uuid`, nil, http.StatusBadRequest, `invalid note ID`, false},
		{`note find error`, ownerID, noteID.String(), func(nr *mockNoteRepo, _ *mockShareRepo) {
			nr.On(`FindByID`, mock.Anything, noteID).Return(nil, errors.New(`db error`))
		}, http.StatusInternalServerError, `failed to fetch note`, false},
		{`note not found`, ownerID, noteID.String(), func(nr *mockNoteRepo, _ *mockShareRepo) {
			nr.On(`FindByID`, mock.Anything, noteID).Return(nil, nil)
		}, http.StatusNotFound, `note not found`, false},
		{`not creator`, otherID, noteID.String(), func(nr *mockNoteRepo, _ *mockShareRepo) {
			nr.On(`FindByID`, mock.Anything, noteID).Return(note, nil)
		}, http.StatusForbidden, `only the creator can view shares`, false},
		{`list shares error`, ownerID, noteID.String(), func(nr *mockNoteRepo, sr *mockShareRepo) {
			nr.On(`FindByID`, mock.Anything, noteID).Return(note, nil)
			sr.On(`ListSharesByNote`, mock.Anything, noteID).Return(nil, errors.New(`db error`))
		}, http.StatusInternalServerError, `failed to list shares`, false},
		{`list share links error`, ownerID, noteID.String(), func(nr *mockNoteRepo, sr *mockShareRepo) {
			nr.On(`FindByID`, mock.Anything, noteID).Return(note, nil)
			sr.On(`ListSharesByNote`, mock.Anything, noteID).Return([]domainshare.NoteShare{}, nil)
			sr.On(`ListShareLinksByNote`, mock.Anything, noteID).Return(nil, errors.New(`db error`))
		}, http.StatusInternalServerError, `failed to list share links`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, nr, _, sr := setupShareHandler(t)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, `/notes/`+tt.paramID+`/shares`, nil)
			c.Params = gin.Params{{Key: `id`, Value: tt.paramID}}
			if !tt.noAuth {
				c.Set(middleware.ContextUserIDKey, tt.userID)
			}
			if tt.setupMocks != nil {
				tt.setupMocks(nr, sr)
			}
			h.ListNoteShares(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if len(tt.wantSub) > 0 {
				assert.Contains(t, w.Body.String(), tt.wantSub)
			}
		})
	}
}
