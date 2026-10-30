package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authpkg "knowledge-graph/internal/auth"
	"knowledge-graph/internal/domain/permission"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPermissionRepo struct {
	mock.Mock
}

func (m *mockPermissionRepo) HasPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	args := m.Called(ctx, userID, resource, action)
	return args.Bool(0), args.Error(1)
}

func (m *mockPermissionRepo) CheckNoteAccess(ctx context.Context, noteID, userID uuid.UUID) (bool, string, error) {
	args := m.Called(ctx, noteID, userID)
	return args.Bool(0), args.String(1), args.Error(2)
}

func (m *mockPermissionRepo) GetNoteOwner(ctx context.Context, noteID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(ctx, noteID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

var _ permission.Repository = (*mockPermissionRepo)(nil)

type mockTokenStore struct {
	mock.Mock
}

func (m *mockTokenStore) BlacklistToken(ctx context.Context, token string, ttl time.Duration) error {
	return m.Called(ctx, token, ttl).Error(0)
}

func (m *mockTokenStore) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

func (m *mockTokenStore) StoreRefreshToken(ctx context.Context, userID string, token string, expiresAt time.Time) error {
	return m.Called(ctx, userID, token, expiresAt).Error(0)
}

func (m *mockTokenStore) ValidateRefreshToken(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}

func (m *mockTokenStore) RevokeRefreshToken(ctx context.Context, token string, ttl time.Duration) error {
	return m.Called(ctx, token, ttl).Error(0)
}

func (m *mockTokenStore) StorePasswordResetToken(ctx context.Context, userID string, token string, ttl time.Duration) error {
	return m.Called(ctx, userID, token, ttl).Error(0)
}

func (m *mockTokenStore) ValidatePasswordResetToken(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}

func (m *mockTokenStore) DeletePasswordResetToken(ctx context.Context, token string) error {
	return m.Called(ctx, token).Error(0)
}

func (m *mockTokenStore) StorePKCE(ctx context.Context, state string, pkce *authpkg.PKCE, ttl time.Duration) error {
	return m.Called(ctx, state, pkce, ttl).Error(0)
}

func (m *mockTokenStore) GetPKCE(ctx context.Context, state string) (*authpkg.PKCE, error) {
	args := m.Called(ctx, state)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authpkg.PKCE), args.Error(1)
}

func (m *mockTokenStore) StoreState(ctx context.Context, state string, ttl time.Duration) error {
	return m.Called(ctx, state, ttl).Error(0)
}

func (m *mockTokenStore) GetState(ctx context.Context, state string) (string, error) {
	args := m.Called(ctx, state)
	return args.String(0), args.Error(1)
}

func (m *mockTokenStore) CachePermission(ctx context.Context, userID, resource, action string, allowed bool, ttl time.Duration) error {
	return m.Called(ctx, userID, resource, action, allowed, ttl).Error(0)
}

func (m *mockTokenStore) CheckCachedPermission(ctx context.Context, userID, resource, action string) (bool, bool, error) {
	args := m.Called(ctx, userID, resource, action)
	return args.Bool(0), args.Bool(1), args.Error(2)
}

func (m *mockTokenStore) InvalidatePermissionCache(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

func setupPermissionContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestCan_AllowedViaRepo(t *testing.T) {
	repo := new(mockPermissionRepo)
	store := new(mockTokenStore)
	cfg := DefaultPermissionConfig(repo, store)

	userID := uuid.New()
	c, w := setupPermissionContext()
	c.Set(ContextUserIDKey, userID)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	store.On("CheckCachedPermission", mock.Anything, userID.String(), "notes", "read").Return(false, false, nil)
	repo.On("HasPermission", mock.Anything, userID, "notes", "read").Return(true, nil)
	store.On("CachePermission", mock.Anything, userID.String(), "notes", "read", true, cfg.CacheTTL).Return(nil)

	called := false
	Can(cfg, "notes", "read")(c)
	c.Next()
	_ = called

	assert.Equal(t, http.StatusOK, w.Code)
	repo.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestCan_Denied(t *testing.T) {
	repo := new(mockPermissionRepo)
	store := new(mockTokenStore)
	cfg := DefaultPermissionConfig(repo, store)

	userID := uuid.New()
	c, w := setupPermissionContext()
	c.Set(ContextUserIDKey, userID)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	store.On("CheckCachedPermission", mock.Anything, userID.String(), "notes", "read").Return(false, false, nil)
	repo.On("HasPermission", mock.Anything, userID, "notes", "read").Return(false, nil)
	store.On("CachePermission", mock.Anything, userID.String(), "notes", "read", false, cfg.CacheTTL).Return(nil)

	Can(cfg, "notes", "read")(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCan_Unauthorized(t *testing.T) {
	repo := new(mockPermissionRepo)
	cfg := DefaultPermissionConfig(repo, nil)

	c, w := setupPermissionContext()
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	Can(cfg, "notes", "read")(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCan_CacheDenied(t *testing.T) {
	repo := new(mockPermissionRepo)
	store := new(mockTokenStore)
	cfg := DefaultPermissionConfig(repo, store)

	userID := uuid.New()
	c, w := setupPermissionContext()
	c.Set(ContextUserIDKey, userID)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	store.On("CheckCachedPermission", mock.Anything, userID.String(), "notes", "read").Return(false, true, nil)

	Can(cfg, "notes", "read")(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCan_RepoError(t *testing.T) {
	repo := new(mockPermissionRepo)
	store := new(mockTokenStore)
	cfg := DefaultPermissionConfig(repo, store)

	userID := uuid.New()
	c, w := setupPermissionContext()
	c.Set(ContextUserIDKey, userID)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	store.On("CheckCachedPermission", mock.Anything, userID.String(), "notes", "read").Return(false, false, nil)
	repo.On("HasPermission", mock.Anything, userID, "notes", "read").Return(false, errors.New("db error"))

	Can(cfg, "notes", "read")(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCanOwn_OwnerAllowed(t *testing.T) {
	repo := new(mockPermissionRepo)
	cfg := DefaultPermissionConfig(repo, nil)

	userID := uuid.New()
	ownerID := userID
	c, w := setupPermissionContext()
	c.Set(ContextUserIDKey, userID)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	CanOwn(cfg, "notes", func(c *gin.Context) (uuid.UUID, error) { return ownerID, nil })(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCanOwn_NonOwnerWithPermission(t *testing.T) {
	repo := new(mockPermissionRepo)
	cfg := DefaultPermissionConfig(repo, nil)

	userID := uuid.New()
	ownerID := uuid.New()
	c, w := setupPermissionContext()
	c.Set(ContextUserIDKey, userID)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	repo.On("HasPermission", mock.Anything, userID, "notes", "manage").Return(true, nil)

	CanOwn(cfg, "notes", func(c *gin.Context) (uuid.UUID, error) { return ownerID, nil })(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCanOwn_NonOwnerWithoutPermission(t *testing.T) {
	repo := new(mockPermissionRepo)
	cfg := DefaultPermissionConfig(repo, nil)

	userID := uuid.New()
	ownerID := uuid.New()
	c, w := setupPermissionContext()
	c.Set(ContextUserIDKey, userID)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	repo.On("HasPermission", mock.Anything, userID, "notes", "manage").Return(false, nil)

	CanOwn(cfg, "notes", func(c *gin.Context) (uuid.UUID, error) { return ownerID, nil })(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCanOwn_Unauthorized(t *testing.T) {
	repo := new(mockPermissionRepo)
	cfg := DefaultPermissionConfig(repo, nil)

	c, w := setupPermissionContext()
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	CanOwn(cfg, "notes", func(c *gin.Context) (uuid.UUID, error) { return uuid.New(), nil })(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireRole_Allowed(t *testing.T) {
	c, w := setupPermissionContext()
	c.Set(ContextRoleKey, "admin")
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	RequireRole("admin")(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_Forbidden(t *testing.T) {
	c, w := setupPermissionContext()
	c.Set(ContextRoleKey, "user")
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	RequireRole("admin")(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireRole_Unauthorized(t *testing.T) {
	c, w := setupPermissionContext()
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	RequireRole("admin")(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestIsOwner(t *testing.T) {
	userID := uuid.New()
	c, _ := setupPermissionContext()
	c.Set(ContextUserIDKey, userID)

	assert.True(t, IsOwner(c, userID))
	assert.False(t, IsOwner(c, uuid.New()))
}

func TestNoteAccessMiddleware_AccessGranted(t *testing.T) {
	repo := new(mockPermissionRepo)
	noteID := uuid.New()
	userID := uuid.New()

	c, w := setupPermissionContext()
	c.Request = httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: noteID.String()}}
	c.Set(ContextUserIDKey, userID)

	repo.On("CheckNoteAccess", mock.Anything, noteID, userID).Return(true, "write", nil)

	NoteAccessMiddleware(repo, nil)(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "write", c.GetString("note_permission"))
}

func TestNoteAccessMiddleware_AccessDenied(t *testing.T) {
	repo := new(mockPermissionRepo)
	noteID := uuid.New()
	userID := uuid.New()

	c, w := setupPermissionContext()
	c.Request = httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: noteID.String()}}
	c.Set(ContextUserIDKey, userID)

	repo.On("CheckNoteAccess", mock.Anything, noteID, userID).Return(false, "", nil)

	NoteAccessMiddleware(repo, nil)(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestNoteAccessMiddleware_InvalidNoteID(t *testing.T) {
	repo := new(mockPermissionRepo)

	c, w := setupPermissionContext()
	c.Request = httptest.NewRequest(http.MethodGet, "/notes/not-a-uuid", nil)
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}
	c.Set(ContextUserIDKey, uuid.New())

	NoteAccessMiddleware(repo, nil)(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
