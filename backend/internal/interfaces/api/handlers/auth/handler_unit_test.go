//go:build !integration
// +build !integration

package auth

import (
	"bytes"
	"context"
	"encoding/json"
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
