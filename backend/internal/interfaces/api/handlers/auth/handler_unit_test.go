//go:build !integration
// +build !integration

package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authpkg "knowledge-graph/internal/auth"
	"knowledge-graph/internal/config"
	domainuser "knowledge-graph/internal/domain/user"
	oauthpkg "knowledge-graph/internal/infrastructure/oauth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) Create(ctx context.Context, u *domainuser.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
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

func (m *mockUserRepo) Update(ctx context.Context, u *domainuser.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *mockUserRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserRepo) EmailExists(ctx context.Context, email string, excludeID uuid.UUID) (bool, error) {
	args := m.Called(ctx, email, excludeID)
	return args.Bool(0), args.Error(1)
}

type mockRefreshTokenRepo struct {
	mock.Mock
}

func (m *mockRefreshTokenRepo) Create(ctx context.Context, token *authpkg.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

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

type mockEmailSender struct {
	mock.Mock
}

func (m *mockEmailSender) SendPasswordReset(ctx context.Context, to, resetLink string) error {
	return m.Called(ctx, to, resetLink).Error(0)
}

type mockOAuthProvider struct {
	mock.Mock
}

func (m *mockOAuthProvider) Exchange(ctx context.Context, code, codeVerifier string) (string, error) {
	args := m.Called(ctx, code, codeVerifier)
	return args.String(0), args.Error(1)
}

func (m *mockOAuthProvider) UserInfo(ctx context.Context, accessToken string) (*oauthpkg.UserInfo, error) {
	args := m.Called(ctx, accessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*oauthpkg.UserInfo), args.Error(1)
}

func setupUnitHandler(t *testing.T) (*Handler, *mockUserRepo, *mockRefreshTokenRepo, *mockTokenStore, *mockEmailSender) {
	gin.SetMode(gin.TestMode)
	jwtManager := authpkg.NewJWTManager("test-secret-key", time.Hour, 7*24*time.Hour)
	cfg := &config.Config{
		Argon2Time:                   1,
		Argon2Memory:                 64 * 1024,
		Argon2Threads:                4,
		PasswordPolicyMinLength:      8,
		PasswordPolicyRequireUpper:   false,
		PasswordPolicyRequireLower:   false,
		PasswordPolicyRequireDigit:   false,
		PasswordPolicyRequireSpecial: false,
		JWTAccessTTL:                 time.Hour,
		JWTRefreshTTL:                7 * 24 * time.Hour,
		YandexClientID:               "client-id",
		YandexClientSecret:           "client-secret",
	}

	userRepo := new(mockUserRepo)
	refreshRepo := new(mockRefreshTokenRepo)
	tokenStore := new(mockTokenStore)
	emailSender := new(mockEmailSender)

	h := NewHandler(userRepo, refreshRepo, tokenStore, jwtManager, cfg, emailSender)
	return h, userRepo, refreshRepo, tokenStore, emailSender
}

func TestForgotPassword_UserExists(t *testing.T) {
	h, userRepo, _, tokenStore, emailSender := setupUnitHandler(t)

	now := time.Now()
	u, err := domainuser.NewUser(uuid.New(), "login", "user@example.com", "hash", "user", now, now, nil)
	require.NoError(t, err)

	userRepo.On("FindByEmail", mock.Anything, "user@example.com").Return(u, nil)
	tokenStore.On("StorePasswordResetToken", mock.Anything, u.ID().String(), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil)
	emailSender.On("SendPasswordReset", mock.Anything, "user@example.com", mock.AnythingOfType("string")).Return(nil)

	body, _ := json.Marshal(ForgotPasswordRequest{Email: "user@example.com"})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/forgot-password", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Host = "example.com"
	h.ForgotPassword(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), "debug_token")
	emailSender.AssertExpectations(t)
}

func TestForgotPassword_UserNotFound(t *testing.T) {
	h, userRepo, _, _, _ := setupUnitHandler(t)

	userRepo.On("FindByEmail", mock.Anything, "missing@example.com").Return(nil, nil)

	body, _ := json.Marshal(ForgotPasswordRequest{Email: "missing@example.com"})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/forgot-password", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Host = "example.com"
	h.ForgotPassword(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestYandexCallback_NewUser(t *testing.T) {
	h, userRepo, refreshRepo, tokenStore, _ := setupUnitHandler(t)
	h.SetOAuthProvider(&mockOAuthProvider{})
	provider := h.oauthProvider.(*mockOAuthProvider)

	provider.On("Exchange", mock.Anything, "code", "").Return("access-token", nil)
	provider.On("UserInfo", mock.Anything, "access-token").Return(&oauthpkg.UserInfo{
		ID:    "yandex-1",
		Login: "yandexuser",
		Email: "oauth@example.com",
	}, nil)

	userRepo.On("FindByEmail", mock.Anything, "oauth@example.com").Return(nil, nil)
	userRepo.On("FindByLogin", mock.Anything, "yandexuser").Return(nil, nil)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
	refreshRepo.On("Create", mock.Anything, mock.AnythingOfType("*auth.RefreshToken")).Return(nil)
	tokenStore.On("StoreRefreshToken", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/auth/yandex/callback?code=code&state=state", nil)
	c.Request.Host = "example.com"
	h.YandexCallback(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "authenticated via Yandex")
	userRepo.AssertExpectations(t)
	refreshRepo.AssertExpectations(t)
	tokenStore.AssertExpectations(t)
}

func TestYandexCallback_ExistingUser(t *testing.T) {
	h, userRepo, refreshRepo, tokenStore, _ := setupUnitHandler(t)
	h.SetOAuthProvider(&mockOAuthProvider{})
	provider := h.oauthProvider.(*mockOAuthProvider)

	now := time.Now()
	u, err := domainuser.NewUser(uuid.New(), "existing", "oauth@example.com", "hash", "user", now, now, nil)
	require.NoError(t, err)

	provider.On("Exchange", mock.Anything, "code", "").Return("access-token", nil)
	provider.On("UserInfo", mock.Anything, "access-token").Return(&oauthpkg.UserInfo{
		ID:    "yandex-2",
		Login: "existing",
		Email: "oauth@example.com",
	}, nil)

	userRepo.On("FindByEmail", mock.Anything, "oauth@example.com").Return(u, nil)
	refreshRepo.On("Create", mock.Anything, mock.AnythingOfType("*auth.RefreshToken")).Return(nil)
	tokenStore.On("StoreRefreshToken", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/auth/yandex/callback?code=code&state=state", nil)
	c.Request.Host = "example.com"
	h.YandexCallback(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func newJSONContext(method, path string, body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, bytes.NewBuffer(body))
	if body != nil {
		c.Request.Header.Set("Content-Type", "application/json")
	}
	return c, w
}

func newTestUser(t *testing.T, login, email, password string) *domainuser.User {
	hash, err := authpkg.HashPassword(password, &authpkg.PasswordConfig{
		Time:    1,
		Memory:  64 * 1024,
		Threads: 4,
		KeyLen:  32,
	})
	require.NoError(t, err)
	u, err := domainuser.NewUser(uuid.New(), login, email, hash, "user", time.Now(), time.Time{}, nil)
	require.NoError(t, err)
	return u
}

func TestRegister(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		h, userRepo, refreshRepo, tokenStore, _ := setupUnitHandler(t)

		userRepo.On("FindByLogin", mock.Anything, "newuser").Return(nil, nil)
		userRepo.On("FindByEmail", mock.Anything, "new@example.com").Return(nil, nil)
		userRepo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
		refreshRepo.On("Create", mock.Anything, mock.AnythingOfType("*auth.RefreshToken")).Return(nil)
		tokenStore.On("StoreRefreshToken", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)

		body, _ := json.Marshal(RegisterRequest{Login: "newuser", Email: "new@example.com", Password: "Password123!"})
		c, w := newJSONContext(http.MethodPost, "/auth/register", body)
		h.Register(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "access_token")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		h, _, _, _, _ := setupUnitHandler(t)
		c, w := newJSONContext(http.MethodPost, "/auth/register", []byte("{\"login\":\""))
		h.Register(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing required fields", func(t *testing.T) {
		h, _, _, _, _ := setupUnitHandler(t)
		body, _ := json.Marshal(RegisterRequest{})
		c, w := newJSONContext(http.MethodPost, "/auth/register", body)
		h.Register(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("password policy failure", func(t *testing.T) {
		h, _, _, _, _ := setupUnitHandler(t)
		h.passwordPolicy.MinLength = 15
		body, _ := json.Marshal(RegisterRequest{Login: "newuser", Email: "new@example.com", Password: "1234567890"})
		c, w := newJSONContext(http.MethodPost, "/auth/register", body)
		h.Register(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "password must be at least")
	})

	t.Run("login already exists", func(t *testing.T) {
		h, userRepo, _, _, _ := setupUnitHandler(t)
		existing := newTestUser(t, "existing", "existing@example.com", "Password123!")
		userRepo.On("FindByLogin", mock.Anything, "existing").Return(existing, nil)

		body, _ := json.Marshal(RegisterRequest{Login: "existing", Email: "new@example.com", Password: "Password123!"})
		c, w := newJSONContext(http.MethodPost, "/auth/register", body)
		h.Register(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "login already exists")
	})

	t.Run("email already exists", func(t *testing.T) {
		h, userRepo, _, _, _ := setupUnitHandler(t)
		userRepo.On("FindByLogin", mock.Anything, "newuser2").Return(nil, nil)
		existing := newTestUser(t, "existing", "existing@example.com", "Password123!")
		userRepo.On("FindByEmail", mock.Anything, "existing@example.com").Return(existing, nil)

		body, _ := json.Marshal(RegisterRequest{Login: "newuser2", Email: "existing@example.com", Password: "Password123!"})
		c, w := newJSONContext(http.MethodPost, "/auth/register", body)
		h.Register(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "email already exists")
	})

	t.Run("find by login error", func(t *testing.T) {
		h, userRepo, _, _, _ := setupUnitHandler(t)
		userRepo.On("FindByLogin", mock.Anything, "newuser").Return(nil, errors.New("db error"))

		body, _ := json.Marshal(RegisterRequest{Login: "newuser", Email: "new@example.com", Password: "Password123!"})
		c, w := newJSONContext(http.MethodPost, "/auth/register", body)
		h.Register(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed to check login")
	})

	t.Run("find by email error", func(t *testing.T) {
		h, userRepo, _, _, _ := setupUnitHandler(t)
		userRepo.On("FindByLogin", mock.Anything, "newuser").Return(nil, nil)
		userRepo.On("FindByEmail", mock.Anything, "new@example.com").Return(nil, errors.New("db error"))

		body, _ := json.Marshal(RegisterRequest{Login: "newuser", Email: "new@example.com", Password: "Password123!"})
		c, w := newJSONContext(http.MethodPost, "/auth/register", body)
		h.Register(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed to check email")
	})

	t.Run("create role not found", func(t *testing.T) {
		h, userRepo, _, _, _ := setupUnitHandler(t)
		userRepo.On("FindByLogin", mock.Anything, "newuser").Return(nil, nil)
		userRepo.On("FindByEmail", mock.Anything, "new@example.com").Return(nil, nil)
		userRepo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(domainuser.ErrRoleNotFound)

		body, _ := json.Marshal(RegisterRequest{Login: "newuser", Email: "new@example.com", Password: "Password123!"})
		c, w := newJSONContext(http.MethodPost, "/auth/register", body)
		h.Register(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed to get default role")
	})

	t.Run("create generic error", func(t *testing.T) {
		h, userRepo, _, _, _ := setupUnitHandler(t)
		userRepo.On("FindByLogin", mock.Anything, "newuser").Return(nil, nil)
		userRepo.On("FindByEmail", mock.Anything, "new@example.com").Return(nil, nil)
		userRepo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(errors.New("db error"))

		body, _ := json.Marshal(RegisterRequest{Login: "newuser", Email: "new@example.com", Password: "Password123!"})
		c, w := newJSONContext(http.MethodPost, "/auth/register", body)
		h.Register(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed to create user")
	})

	t.Run("store refresh token error", func(t *testing.T) {
		h, userRepo, refreshRepo, tokenStore, _ := setupUnitHandler(t)
		userRepo.On("FindByLogin", mock.Anything, "newuser").Return(nil, nil)
		userRepo.On("FindByEmail", mock.Anything, "new@example.com").Return(nil, nil)
		userRepo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
		refreshRepo.On("Create", mock.Anything, mock.AnythingOfType("*auth.RefreshToken")).Return(nil)
		tokenStore.On("StoreRefreshToken", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(errors.New("redis down"))

		body, _ := json.Marshal(RegisterRequest{Login: "newuser", Email: "new@example.com", Password: "Password123!"})
		c, w := newJSONContext(http.MethodPost, "/auth/register", body)
		h.Register(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed to store refresh token")
	})
}

func TestLogin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		h, userRepo, refreshRepo, tokenStore, _ := setupUnitHandler(t)
		u := newTestUser(t, "validuser", "valid@example.com", "Password123!")
		userRepo.On("FindByLogin", mock.Anything, "validuser").Return(u, nil)
		refreshRepo.On("Create", mock.Anything, mock.AnythingOfType("*auth.RefreshToken")).Return(nil)
		tokenStore.On("StoreRefreshToken", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)

		body, _ := json.Marshal(LoginRequest{Login: "validuser", Password: "Password123!"})
		c, w := newJSONContext(http.MethodPost, "/auth/login", body)
		h.Login(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "access_token")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		h, _, _, _, _ := setupUnitHandler(t)
		c, w := newJSONContext(http.MethodPost, "/auth/login", []byte("{\"login\":\""))
		h.Login(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing required fields", func(t *testing.T) {
		h, _, _, _, _ := setupUnitHandler(t)
		body, _ := json.Marshal(LoginRequest{})
		c, w := newJSONContext(http.MethodPost, "/auth/login", body)
		h.Login(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("find by login error", func(t *testing.T) {
		h, userRepo, _, _, _ := setupUnitHandler(t)
		userRepo.On("FindByLogin", mock.Anything, "validuser").Return(nil, errors.New("db error"))

		body, _ := json.Marshal(LoginRequest{Login: "validuser", Password: "Password123!"})
		c, w := newJSONContext(http.MethodPost, "/auth/login", body)
		h.Login(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("user not found", func(t *testing.T) {
		h, userRepo, _, _, _ := setupUnitHandler(t)
		userRepo.On("FindByLogin", mock.Anything, "validuser").Return(nil, nil)

		body, _ := json.Marshal(LoginRequest{Login: "validuser", Password: "Password123!"})
		c, w := newJSONContext(http.MethodPost, "/auth/login", body)
		h.Login(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid credentials")
	})

	t.Run("invalid password", func(t *testing.T) {
		h, userRepo, _, _, _ := setupUnitHandler(t)
		u := newTestUser(t, "validuser", "valid@example.com", "Password123!")
		userRepo.On("FindByLogin", mock.Anything, "validuser").Return(u, nil)

		body, _ := json.Marshal(LoginRequest{Login: "validuser", Password: "WrongPassword123!"})
		c, w := newJSONContext(http.MethodPost, "/auth/login", body)
		h.Login(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid credentials")
	})

	t.Run("store refresh token error", func(t *testing.T) {
		h, userRepo, refreshRepo, tokenStore, _ := setupUnitHandler(t)
		u := newTestUser(t, "validuser", "valid@example.com", "Password123!")
		userRepo.On("FindByLogin", mock.Anything, "validuser").Return(u, nil)
		refreshRepo.On("Create", mock.Anything, mock.AnythingOfType("*auth.RefreshToken")).Return(nil)
		tokenStore.On("StoreRefreshToken", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(errors.New("redis down"))

		body, _ := json.Marshal(LoginRequest{Login: "validuser", Password: "Password123!"})
		c, w := newJSONContext(http.MethodPost, "/auth/login", body)
		h.Login(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed to store refresh token")
	})
}

func TestLogout(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		h, _, _, tokenStore, _ := setupUnitHandler(t)
		tokenStore.On("RevokeRefreshToken", mock.Anything, "refresh-token", mock.AnythingOfType("time.Duration")).Return(nil)
		tokenStore.On("BlacklistToken", mock.Anything, "access-token", mock.AnythingOfType("time.Duration")).Return(nil)

		c, w := newJSONContext(http.MethodPost, "/auth/logout", nil)
		c.Request.Header.Set("X-Refresh-Token", "refresh-token")
		c.Request.Header.Set("Authorization", "Bearer access-token")
		h.Logout(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "logged out successfully")
	})

	t.Run("missing token", func(t *testing.T) {
		h, _, _, _, _ := setupUnitHandler(t)
		c, w := newJSONContext(http.MethodPost, "/auth/logout", nil)
		h.Logout(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "logged out successfully")
	})
}

func TestRefresh(t *testing.T) {
	t.Run("success from body", func(t *testing.T) {
		h, userRepo, refreshRepo, tokenStore, _ := setupUnitHandler(t)
		uid := uuid.New()
		pair, err := h.jwtManager.GenerateTokenPair(uid, "refreshuser", "user")
		require.NoError(t, err)

		tokenStore.On("IsTokenBlacklisted", mock.Anything, pair.RefreshToken).Return(false, nil)
		tokenStore.On("ValidateRefreshToken", mock.Anything, pair.RefreshToken).Return(uid.String(), nil)

		u, err := domainuser.NewUser(uid, "refreshuser", "refresh@example.com", "hash", "user", time.Now(), time.Time{}, nil)
		require.NoError(t, err)
		userRepo.On("FindByID", mock.Anything, uid).Return(u, nil)

		refreshRepo.On("Create", mock.Anything, mock.AnythingOfType("*auth.RefreshToken")).Return(nil)
		tokenStore.On("StoreRefreshToken", mock.Anything, uid.String(), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)
		tokenStore.On("RevokeRefreshToken", mock.Anything, pair.RefreshToken, mock.AnythingOfType("time.Duration")).Return(nil)

		body, _ := json.Marshal(RefreshRequest{RefreshToken: pair.RefreshToken})
		c, w := newJSONContext(http.MethodPost, "/auth/refresh", body)
		h.Refresh(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "access_token")
	})

	t.Run("success from cookie", func(t *testing.T) {
		h, userRepo, refreshRepo, tokenStore, _ := setupUnitHandler(t)
		uid := uuid.New()
		pair, err := h.jwtManager.GenerateTokenPair(uid, "refreshuser", "user")
		require.NoError(t, err)

		tokenStore.On("IsTokenBlacklisted", mock.Anything, pair.RefreshToken).Return(false, nil)
		tokenStore.On("ValidateRefreshToken", mock.Anything, pair.RefreshToken).Return(uid.String(), nil)

		u, err := domainuser.NewUser(uid, "refreshuser", "refresh@example.com", "hash", "user", time.Now(), time.Time{}, nil)
		require.NoError(t, err)
		userRepo.On("FindByID", mock.Anything, uid).Return(u, nil)

		refreshRepo.On("Create", mock.Anything, mock.AnythingOfType("*auth.RefreshToken")).Return(nil)
		tokenStore.On("StoreRefreshToken", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)
		tokenStore.On("RevokeRefreshToken", mock.Anything, pair.RefreshToken, mock.AnythingOfType("time.Duration")).Return(nil)

		c, w := newJSONContext(http.MethodPost, "/auth/refresh", nil)
		c.Request.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: pair.RefreshToken})
		h.Refresh(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("missing token", func(t *testing.T) {
		h, _, _, _, _ := setupUnitHandler(t)
		c, w := newJSONContext(http.MethodPost, "/auth/refresh", nil)
		h.Refresh(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "refresh token is required")
	})

	t.Run("invalid token", func(t *testing.T) {
		h, _, _, _, _ := setupUnitHandler(t)
		body, _ := json.Marshal(RefreshRequest{RefreshToken: "invalid-token"})
		c, w := newJSONContext(http.MethodPost, "/auth/refresh", body)
		h.Refresh(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid refresh token")
	})

	t.Run("blacklisted token", func(t *testing.T) {
		h, _, _, tokenStore, _ := setupUnitHandler(t)
		uid := uuid.New()
		pair, err := h.jwtManager.GenerateTokenPair(uid, "refreshuser", "user")
		require.NoError(t, err)

		tokenStore.On("IsTokenBlacklisted", mock.Anything, pair.RefreshToken).Return(true, nil)

		body, _ := json.Marshal(RefreshRequest{RefreshToken: pair.RefreshToken})
		c, w := newJSONContext(http.MethodPost, "/auth/refresh", body)
		h.Refresh(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "token has been revoked")
	})

	t.Run("is token blacklisted error", func(t *testing.T) {
		h, _, _, tokenStore, _ := setupUnitHandler(t)
		uid := uuid.New()
		pair, err := h.jwtManager.GenerateTokenPair(uid, "refreshuser", "user")
		require.NoError(t, err)

		tokenStore.On("IsTokenBlacklisted", mock.Anything, pair.RefreshToken).Return(false, errors.New("redis down"))

		body, _ := json.Marshal(RefreshRequest{RefreshToken: pair.RefreshToken})
		c, w := newJSONContext(http.MethodPost, "/auth/refresh", body)
		h.Refresh(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed to validate token")
	})

	t.Run("validate refresh token error", func(t *testing.T) {
		h, _, _, tokenStore, _ := setupUnitHandler(t)
		uid := uuid.New()
		pair, err := h.jwtManager.GenerateTokenPair(uid, "refreshuser", "user")
		require.NoError(t, err)

		tokenStore.On("IsTokenBlacklisted", mock.Anything, pair.RefreshToken).Return(false, nil)
		tokenStore.On("ValidateRefreshToken", mock.Anything, pair.RefreshToken).Return("", errors.New("not found"))

		body, _ := json.Marshal(RefreshRequest{RefreshToken: pair.RefreshToken})
		c, w := newJSONContext(http.MethodPost, "/auth/refresh", body)
		h.Refresh(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid refresh token")
	})

	t.Run("user not found", func(t *testing.T) {
		h, userRepo, _, tokenStore, _ := setupUnitHandler(t)
		uid := uuid.New()
		pair, err := h.jwtManager.GenerateTokenPair(uid, "refreshuser", "user")
		require.NoError(t, err)

		tokenStore.On("IsTokenBlacklisted", mock.Anything, pair.RefreshToken).Return(false, nil)
		tokenStore.On("ValidateRefreshToken", mock.Anything, pair.RefreshToken).Return(uid.String(), nil)
		userRepo.On("FindByID", mock.Anything, uid).Return(nil, nil)

		body, _ := json.Marshal(RefreshRequest{RefreshToken: pair.RefreshToken})
		c, w := newJSONContext(http.MethodPost, "/auth/refresh", body)
		h.Refresh(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "user not found")
	})

	t.Run("find by ID error", func(t *testing.T) {
		h, userRepo, _, tokenStore, _ := setupUnitHandler(t)
		uid := uuid.New()
		pair, err := h.jwtManager.GenerateTokenPair(uid, "refreshuser", "user")
		require.NoError(t, err)

		tokenStore.On("IsTokenBlacklisted", mock.Anything, pair.RefreshToken).Return(false, nil)
		tokenStore.On("ValidateRefreshToken", mock.Anything, pair.RefreshToken).Return(uid.String(), nil)
		userRepo.On("FindByID", mock.Anything, uid).Return(nil, errors.New("db error"))

		body, _ := json.Marshal(RefreshRequest{RefreshToken: pair.RefreshToken})
		c, w := newJSONContext(http.MethodPost, "/auth/refresh", body)
		h.Refresh(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("user deleted", func(t *testing.T) {
		h, userRepo, _, tokenStore, _ := setupUnitHandler(t)
		uid := uuid.New()
		pair, err := h.jwtManager.GenerateTokenPair(uid, "refreshuser", "user")
		require.NoError(t, err)

		tokenStore.On("IsTokenBlacklisted", mock.Anything, pair.RefreshToken).Return(false, nil)
		tokenStore.On("ValidateRefreshToken", mock.Anything, pair.RefreshToken).Return(uid.String(), nil)

		deletedAt := time.Now()
		u, err := domainuser.NewUser(uid, "refreshuser", "refresh@example.com", "hash", "user", time.Now(), time.Time{}, &deletedAt)
		require.NoError(t, err)
		userRepo.On("FindByID", mock.Anything, uid).Return(u, nil)

		body, _ := json.Marshal(RefreshRequest{RefreshToken: pair.RefreshToken})
		c, w := newJSONContext(http.MethodPost, "/auth/refresh", body)
		h.Refresh(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("store refresh token error", func(t *testing.T) {
		h, userRepo, refreshRepo, tokenStore, _ := setupUnitHandler(t)
		uid := uuid.New()
		pair, err := h.jwtManager.GenerateTokenPair(uid, "refreshuser", "user")
		require.NoError(t, err)

		tokenStore.On("IsTokenBlacklisted", mock.Anything, pair.RefreshToken).Return(false, nil)
		tokenStore.On("ValidateRefreshToken", mock.Anything, pair.RefreshToken).Return(uid.String(), nil)

		u, err := domainuser.NewUser(uid, "refreshuser", "refresh@example.com", "hash", "user", time.Now(), time.Time{}, nil)
		require.NoError(t, err)
		userRepo.On("FindByID", mock.Anything, uid).Return(u, nil)

		refreshRepo.On("Create", mock.Anything, mock.AnythingOfType("*auth.RefreshToken")).Return(errors.New("db error"))

		body, _ := json.Marshal(RefreshRequest{RefreshToken: pair.RefreshToken})
		c, w := newJSONContext(http.MethodPost, "/auth/refresh", body)
		h.Refresh(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed to store refresh token")
	})
}

func TestResetPassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		h, userRepo, _, tokenStore, _ := setupUnitHandler(t)
		uid := uuid.New()
		u, err := domainuser.NewUser(uid, "resetuser", "reset@example.com", "hash", "user", time.Now(), time.Time{}, nil)
		require.NoError(t, err)

		tokenStore.On("ValidatePasswordResetToken", mock.Anything, "valid-token").Return(uid.String(), nil)
		userRepo.On("FindByID", mock.Anything, uid).Return(u, nil)
		userRepo.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
		tokenStore.On("DeletePasswordResetToken", mock.Anything, "valid-token").Return(nil)

		body, _ := json.Marshal(ResetPasswordRequest{Token: "valid-token", NewPassword: "NewPass123!"})
		c, w := newJSONContext(http.MethodPost, "/auth/reset-password", body)
		h.ResetPassword(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "password has been reset successfully")
	})

	t.Run("missing fields", func(t *testing.T) {
		h, _, _, _, _ := setupUnitHandler(t)
		body, _ := json.Marshal(ResetPasswordRequest{})
		c, w := newJSONContext(http.MethodPost, "/auth/reset-password", body)
		h.ResetPassword(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("new password too short", func(t *testing.T) {
		h, _, _, _, _ := setupUnitHandler(t)
		body, _ := json.Marshal(ResetPasswordRequest{Token: "valid-token", NewPassword: "short"})
		c, w := newJSONContext(http.MethodPost, "/auth/reset-password", body)
		h.ResetPassword(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("password policy failure", func(t *testing.T) {
		h, _, _, _, _ := setupUnitHandler(t)
		h.passwordPolicy.MinLength = 15
		body, _ := json.Marshal(ResetPasswordRequest{Token: "valid-token", NewPassword: "1234567890"})
		c, w := newJSONContext(http.MethodPost, "/auth/reset-password", body)
		h.ResetPassword(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "password must be at least")
	})

	t.Run("token store not available", func(t *testing.T) {
		h, _, _, _, _ := setupUnitHandler(t)
		h.tokenStore = nil
		body, _ := json.Marshal(ResetPasswordRequest{Token: "valid-token", NewPassword: "NewPass123!"})
		c, w := newJSONContext(http.MethodPost, "/auth/reset-password", body)
		h.ResetPassword(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "token store not available")
	})

	t.Run("invalid or expired token", func(t *testing.T) {
		h, _, _, tokenStore, _ := setupUnitHandler(t)
		tokenStore.On("ValidatePasswordResetToken", mock.Anything, "bad-token").Return("", errors.New("expired"))

		body, _ := json.Marshal(ResetPasswordRequest{Token: "bad-token", NewPassword: "NewPass123!"})
		c, w := newJSONContext(http.MethodPost, "/auth/reset-password", body)
		h.ResetPassword(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid or expired reset token")
	})

	t.Run("user ID parse error", func(t *testing.T) {
		h, _, _, tokenStore, _ := setupUnitHandler(t)
		tokenStore.On("ValidatePasswordResetToken", mock.Anything, "bad-token").Return("not-a-uuid", nil)

		body, _ := json.Marshal(ResetPasswordRequest{Token: "bad-token", NewPassword: "NewPass123!"})
		c, w := newJSONContext(http.MethodPost, "/auth/reset-password", body)
		h.ResetPassword(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid or expired reset token")
	})

	t.Run("user not found", func(t *testing.T) {
		h, userRepo, _, tokenStore, _ := setupUnitHandler(t)
		uid := uuid.New()
		tokenStore.On("ValidatePasswordResetToken", mock.Anything, "valid-token").Return(uid.String(), nil)
		userRepo.On("FindByID", mock.Anything, uid).Return(nil, nil)

		body, _ := json.Marshal(ResetPasswordRequest{Token: "valid-token", NewPassword: "NewPass123!"})
		c, w := newJSONContext(http.MethodPost, "/auth/reset-password", body)
		h.ResetPassword(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("find by ID error", func(t *testing.T) {
		h, userRepo, _, tokenStore, _ := setupUnitHandler(t)
		uid := uuid.New()
		tokenStore.On("ValidatePasswordResetToken", mock.Anything, "valid-token").Return(uid.String(), nil)
		userRepo.On("FindByID", mock.Anything, uid).Return(nil, errors.New("db error"))

		body, _ := json.Marshal(ResetPasswordRequest{Token: "valid-token", NewPassword: "NewPass123!"})
		c, w := newJSONContext(http.MethodPost, "/auth/reset-password", body)
		h.ResetPassword(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed to find user")
	})

	t.Run("update password error", func(t *testing.T) {
		h, userRepo, _, tokenStore, _ := setupUnitHandler(t)
		uid := uuid.New()
		u, err := domainuser.NewUser(uid, "resetuser", "reset@example.com", "hash", "user", time.Now(), time.Time{}, nil)
		require.NoError(t, err)

		tokenStore.On("ValidatePasswordResetToken", mock.Anything, "valid-token").Return(uid.String(), nil)
		userRepo.On("FindByID", mock.Anything, uid).Return(u, nil)
		userRepo.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(errors.New("db error"))

		body, _ := json.Marshal(ResetPasswordRequest{Token: "valid-token", NewPassword: "NewPass123!"})
		c, w := newJSONContext(http.MethodPost, "/auth/reset-password", body)
		h.ResetPassword(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed to update password")
	})
}
