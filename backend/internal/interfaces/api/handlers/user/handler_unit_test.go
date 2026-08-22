//go:build !integration
// +build !integration

package user

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"knowledge-graph/internal/auth"
	domainuser "knowledge-graph/internal/domain/user"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

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
	args := m.Called(ctx, login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainuser.User), args.Error(1)
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domainuser.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainuser.User), args.Error(1)
}

func (m *mockUserRepo) Create(ctx context.Context, u *domainuser.User) error {
	return m.Called(ctx, u).Error(0)
}

func (m *mockUserRepo) Update(ctx context.Context, u *domainuser.User) error {
	return m.Called(ctx, u).Error(0)
}

func (m *mockUserRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockUserRepo) EmailExists(ctx context.Context, email string, excludeID uuid.UUID) (bool, error) {
	args := m.Called(ctx, email, excludeID)
	return args.Bool(0), args.Error(1)
}

type mockAPIKeyRepo struct {
	mock.Mock
}

func (m *mockAPIKeyRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]domainuser.APIKey, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domainuser.APIKey), args.Error(1)
}

func (m *mockAPIKeyRepo) Create(ctx context.Context, key *domainuser.APIKey) error {
	return m.Called(ctx, key).Error(0)
}

func (m *mockAPIKeyRepo) Revoke(ctx context.Context, keyID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, keyID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *mockAPIKeyRepo) FindActiveByHash(ctx context.Context, hash string) (*domainuser.APIKey, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainuser.APIKey), args.Error(1)
}

func (m *mockAPIKeyRepo) UpdateLastUsed(ctx context.Context, keyID uuid.UUID) error {
	return m.Called(ctx, keyID).Error(0)
}

func setupUserHandler(t *testing.T) (*Handler, *mockUserRepo, *mockAPIKeyRepo) {
	gin.SetMode(gin.TestMode)
	userRepo := new(mockUserRepo)
	apiKeyRepo := new(mockAPIKeyRepo)
	cfg := &auth.PasswordConfig{
		Time:    1,
		Memory:  64 * 1024,
		Threads: 4,
		KeyLen:  32,
	}
	policy := auth.DefaultPasswordPolicy()
	return NewHandler(userRepo, apiKeyRepo, cfg, policy), userRepo, apiKeyRepo
}

func newTestUser(t *testing.T) *domainuser.User {
	now := time.Now()
	hash, err := auth.HashPassword("TestPass123!", &auth.PasswordConfig{
		Time: 1, Memory: 64 * 1024, Threads: 4, KeyLen: 32,
	})
	require.NoError(t, err)
	u, err := domainuser.NewUser(uuid.New(), "login", "user@example.com", hash, "user", now, now, nil)
	require.NoError(t, err)
	return u
}

func TestGetMe_Success(t *testing.T) {
	h, repo, _ := setupUserHandler(t)
	u := newTestUser(t)
	repo.On("FindByID", mock.Anything, u.ID()).Return(u, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/me", nil)
	c.Set(middleware.ContextUserIDKey, u.ID())
	h.GetMe(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "user@example.com")
}

func TestGetMe_NotFound(t *testing.T) {
	h, repo, _ := setupUserHandler(t)
	userID := uuid.New()
	repo.On("FindByID", mock.Anything, userID).Return(nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/me", nil)
	c.Set(middleware.ContextUserIDKey, userID)
	h.GetMe(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateMe_Email(t *testing.T) {
	h, repo, _ := setupUserHandler(t)
	u := newTestUser(t)

	repo.On("FindByID", mock.Anything, u.ID()).Return(u, nil)
	repo.On("EmailExists", mock.Anything, "new@example.com", u.ID()).Return(false, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)

	body, _ := json.Marshal(UpdateUserRequest{Email: "new@example.com"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/me", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(middleware.ContextUserIDKey, u.ID())
	h.UpdateMe(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "new@example.com")
}

func TestUpdateMe_Password(t *testing.T) {
	h, repo, _ := setupUserHandler(t)
	u := newTestUser(t)

	repo.On("FindByID", mock.Anything, u.ID()).Return(u, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)

	body, _ := json.Marshal(UpdateUserRequest{
		OldPassword: "TestPass123!",
		NewPassword: "NewPass123!",
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/me", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(middleware.ContextUserIDKey, u.ID())
	h.UpdateMe(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteMe_Success(t *testing.T) {
	h, repo, _ := setupUserHandler(t)
	u := newTestUser(t)

	repo.On("FindByID", mock.Anything, u.ID()).Return(u, nil)
	repo.On("SoftDelete", mock.Anything, u.ID()).Return(nil)

	body, _ := json.Marshal(map[string]string{"password": "TestPass123!"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/me", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(middleware.ContextUserIDKey, u.ID())
	h.DeleteMe(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListAPIKeys(t *testing.T) {
	h, _, keyRepo := setupUserHandler(t)
	userID := uuid.New()
	now := time.Now()
	key, err := domainuser.NewAPIKey(uuid.New(), userID, "hash", "test-key", []string{"read"}, now)
	require.NoError(t, err)

	keyRepo.On("FindByUserID", mock.Anything, userID).Return([]domainuser.APIKey{*key}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api-keys", nil)
	c.Set(middleware.ContextUserIDKey, userID)
	h.ListAPIKeys(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test-key")
}

func TestCreateAPIKey(t *testing.T) {
	h, _, keyRepo := setupUserHandler(t)
	userID := uuid.New()
	keyRepo.On("Create", mock.Anything, mock.AnythingOfType("*user.APIKey")).Return(nil)

	body, _ := json.Marshal(CreateAPIKeyRequest{Name: "new-key", Scopes: []string{"read"}})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api-keys", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(middleware.ContextUserIDKey, userID)
	h.CreateAPIKey(c)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestRevokeAPIKey(t *testing.T) {
	h, _, keyRepo := setupUserHandler(t)
	userID := uuid.New()
	keyID := uuid.New()

	keyRepo.On("Revoke", mock.Anything, keyID, userID).Return(true, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api-keys/"+keyID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: keyID.String()}}
	c.Set(middleware.ContextUserIDKey, userID)
	h.RevokeAPIKey(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetMe_Unauthorized(t *testing.T) {
	h, _, _ := setupUserHandler(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/me", nil)
	h.GetMe(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateMe_EmailTaken(t *testing.T) {
	h, repo, _ := setupUserHandler(t)
	u := newTestUser(t)

	repo.On("FindByID", mock.Anything, u.ID()).Return(u, nil)
	repo.On("EmailExists", mock.Anything, "taken@example.com", u.ID()).Return(true, nil)

	body, _ := json.Marshal(UpdateUserRequest{Email: "taken@example.com"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/me", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(middleware.ContextUserIDKey, u.ID())
	h.UpdateMe(c)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestUpdateMe_InvalidEmailFormat(t *testing.T) {
	h, repo, _ := setupUserHandler(t)
	u := newTestUser(t)

	body, _ := json.Marshal(UpdateUserRequest{Email: "not-an-email"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/me", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(middleware.ContextUserIDKey, u.ID())
	h.UpdateMe(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	repo.AssertNotCalled(t, "FindByID", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "EmailExists", mock.Anything, mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestUpdateMe_WrongOldPassword(t *testing.T) {
	h, repo, _ := setupUserHandler(t)
	u := newTestUser(t)

	repo.On("FindByID", mock.Anything, u.ID()).Return(u, nil)

	body, _ := json.Marshal(UpdateUserRequest{
		OldPassword: "WrongPass123!",
		NewPassword: "NewPass123!",
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/me", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(middleware.ContextUserIDKey, u.ID())
	h.UpdateMe(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDeleteMe_WrongPassword(t *testing.T) {
	h, repo, _ := setupUserHandler(t)
	u := newTestUser(t)

	repo.On("FindByID", mock.Anything, u.ID()).Return(u, nil)

	body, _ := json.Marshal(map[string]string{"password": "WrongPass123!"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/me", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(middleware.ContextUserIDKey, u.ID())
	h.DeleteMe(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateAPIKey_Invalid(t *testing.T) {
	h, _, _ := setupUserHandler(t)
	userID := uuid.New()

	body, _ := json.Marshal(map[string]string{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api-keys", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(middleware.ContextUserIDKey, userID)
	h.CreateAPIKey(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRevokeAPIKey_InvalidID(t *testing.T) {
	h, _, _ := setupUserHandler(t)
	userID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api-keys/not-a-uuid", nil)
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}
	c.Set(middleware.ContextUserIDKey, userID)
	h.RevokeAPIKey(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetMe_UserNotFound(t *testing.T) {
	h, repo, _ := setupUserHandler(t)
	uid := uuid.New()

	repo.On("FindByID", mock.Anything, uid).Return(nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/me", nil)
	c.Set(middleware.ContextUserIDKey, uid)
	h.GetMe(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetMe_FindByIDError(t *testing.T) {
	h, repo, _ := setupUserHandler(t)
	uid := uuid.New()

	repo.On("FindByID", mock.Anything, uid).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/me", nil)
	c.Set(middleware.ContextUserIDKey, uid)
	h.GetMe(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateMe_UserNotFound(t *testing.T) {
	h, repo, _ := setupUserHandler(t)
	uid := uuid.New()

	repo.On("FindByID", mock.Anything, uid).Return(nil, nil)

	body, _ := json.Marshal(map[string]string{"email": "new@example.com"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/me", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(middleware.ContextUserIDKey, uid)
	h.UpdateMe(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateMe_UpdateError(t *testing.T) {
	h, repo, _ := setupUserHandler(t)
	u := newTestUser(t)

	repo.On("FindByID", mock.Anything, u.ID()).Return(u, nil)
	repo.On("EmailExists", mock.Anything, "new@example.com", u.ID()).Return(false, nil)
	repo.On("Update", mock.Anything, u).Return(assert.AnError)

	body, _ := json.Marshal(map[string]string{"email": "new@example.com"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/me", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(middleware.ContextUserIDKey, u.ID())
	h.UpdateMe(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteMe_MissingBody(t *testing.T) {
	h, _, _ := setupUserHandler(t)
	uid := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/me", bytes.NewBuffer([]byte(`{}`)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(middleware.ContextUserIDKey, uid)
	h.DeleteMe(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteMe_SoftDeleteError(t *testing.T) {
	h, repo, _ := setupUserHandler(t)
	u := newTestUser(t)

	repo.On("FindByID", mock.Anything, u.ID()).Return(u, nil)
	repo.On("SoftDelete", mock.Anything, u.ID()).Return(assert.AnError)

	body, _ := json.Marshal(map[string]string{"password": "TestPass123!"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/me", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(middleware.ContextUserIDKey, u.ID())
	h.DeleteMe(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestListAPIKeys_Error(t *testing.T) {
	h, _, keyRepo := setupUserHandler(t)
	uid := uuid.New()

	keyRepo.On("FindByUserID", mock.Anything, uid).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api-keys", nil)
	c.Set(middleware.ContextUserIDKey, uid)
	h.ListAPIKeys(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateAPIKey_SaveError(t *testing.T) {
	h, _, keyRepo := setupUserHandler(t)
	uid := uuid.New()

	keyRepo.On("Create", mock.Anything, mock.AnythingOfType("*user.APIKey")).Return(assert.AnError)

	body, _ := json.Marshal(CreateAPIKeyRequest{Name: "new-key", Scopes: []string{"read"}})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api-keys", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(middleware.ContextUserIDKey, uid)
	h.CreateAPIKey(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRevokeAPIKey_NotFound(t *testing.T) {
	h, _, keyRepo := setupUserHandler(t)
	uid := uuid.New()
	keyID := uuid.New()

	keyRepo.On("Revoke", mock.Anything, keyID, uid).Return(false, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api-keys/"+keyID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: keyID.String()}}
	c.Set(middleware.ContextUserIDKey, uid)
	h.RevokeAPIKey(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRevokeAPIKey_RepoError(t *testing.T) {
	h, _, keyRepo := setupUserHandler(t)
	uid := uuid.New()
	keyID := uuid.New()

	keyRepo.On("Revoke", mock.Anything, keyID, uid).Return(false, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api-keys/"+keyID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: keyID.String()}}
	c.Set(middleware.ContextUserIDKey, uid)
	h.RevokeAPIKey(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
