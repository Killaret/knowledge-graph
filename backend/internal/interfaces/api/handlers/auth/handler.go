// Package auth provides HTTP handlers for authentication
package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	authpkg "knowledge-graph/internal/auth"
	"knowledge-graph/internal/config"
	domainuser "knowledge-graph/internal/domain/user"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles authentication requests
type Handler struct {
	userRepo         domainuser.Repository
	refreshTokenRepo authpkg.RefreshTokenRepository
	jwtManager       *authpkg.JWTManager
	tokenStore       authpkg.TokenStore
	passwordConfig   *authpkg.PasswordConfig
	passwordPolicy   *authpkg.PasswordPolicy
	cfg              *config.Config
}

// NewHandler creates a new auth handler
func NewHandler(
	userRepo domainuser.Repository,
	refreshTokenRepo authpkg.RefreshTokenRepository,
	tokenStore authpkg.TokenStore,
	jwtManager *authpkg.JWTManager,
	cfg *config.Config,
) *Handler {
	return &Handler{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtManager:       jwtManager,
		tokenStore:       tokenStore,
		passwordConfig: &authpkg.PasswordConfig{
			Time:    cfg.Argon2Time,
			Memory:  cfg.Argon2Memory,
			Threads: cfg.Argon2Threads,
			KeyLen:  32,
		},
		passwordPolicy: &authpkg.PasswordPolicy{
			MinLength:      cfg.PasswordPolicyMinLength,
			RequireUpper:   cfg.PasswordPolicyRequireUpper,
			RequireLower:   cfg.PasswordPolicyRequireLower,
			RequireDigit:   cfg.PasswordPolicyRequireDigit,
			RequireSpecial: cfg.PasswordPolicyRequireSpecial,
		},
		cfg: cfg,
	}
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Login    string `json:"login" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=10"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// TokenResponse represents a token response
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         UserInfo  `json:"user"`
}

// UserInfo represents user information in responses
type UserInfo struct {
	ID        uuid.UUID `json:"id"`
	Login     string    `json:"login"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// RefreshRequest represents a refresh token request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents a reset password request
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=10"`
}

// Register handles user registration
func (h *Handler) Register(c *gin.Context) {
	middleware.SetDBEntity(c, "users")
	middleware.SetDBOperation(c, "create")

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate password policy
	if err := authpkg.ValidatePassword(req.Password, h.passwordPolicy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Check if login already exists
	existing, err := h.userRepo.FindByLogin(ctx, req.Login)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check login"})
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "login already exists"})
		return
	}

	// Check if email already exists
	existing, err = h.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check email"})
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		return
	}

	// Hash password
	passwordHash, err := authpkg.HashPassword(req.Password, h.passwordConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Create user with the default role "user"
	newUser, err := domainuser.NewUser(uuid.New(), req.Login, req.Email, passwordHash, "user", time.Now(), time.Time{}, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	if err := h.userRepo.Create(ctx, newUser); err != nil {
		if errors.Is(err, domainuser.ErrRoleNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get default role"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// Generate tokens
	tokens, err := h.jwtManager.GenerateTokenPair(newUser.ID(), newUser.Login(), newUser.Role())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	// Store refresh token
	if err := h.storeRefreshToken(c, newUser.ID(), tokens.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store refresh token"})
		return
	}

	c.JSON(http.StatusCreated, TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresAt:    tokens.ExpiresAt,
		User: UserInfo{
			ID:        newUser.ID(),
			Login:     newUser.Login(),
			Email:     newUser.Email(),
			Role:      newUser.Role(),
			CreatedAt: newUser.CreatedAt(),
		},
	})
}

// Login handles user login
func (h *Handler) Login(c *gin.Context) {
	middleware.SetDBEntity(c, "users")
	middleware.SetDBOperation(c, "read")

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by login
	u, err := h.userRepo.FindByLogin(c.Request.Context(), req.Login)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if u == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Verify password
	valid, err := authpkg.VerifyPassword(req.Password, u.PasswordHash())
	if err != nil || !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	roleName := u.Role()
	if roleName == "" {
		roleName = "user"
	}

	// Generate tokens
	tokens, err := h.jwtManager.GenerateTokenPair(u.ID(), u.Login(), roleName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	// Store refresh token
	if err := h.storeRefreshToken(c, u.ID(), tokens.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store refresh token"})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresAt:    tokens.ExpiresAt,
		User: UserInfo{
			ID:        u.ID(),
			Login:     u.Login(),
			Email:     u.Email(),
			Role:      roleName,
			CreatedAt: u.CreatedAt(),
		},
	})
}

// Refresh handles token refresh
func (h *Handler) Refresh(c *gin.Context) {
	middleware.SetDBEntity(c, "refresh_tokens")
	middleware.SetDBOperation(c, "read")

	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate refresh token
	claims, err := h.jwtManager.ValidateToken(req.RefreshToken, "refresh")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	ctx := c.Request.Context()

	// Check if blacklisted
	if h.tokenStore != nil {
		blacklisted, err := h.tokenStore.IsTokenBlacklisted(ctx, req.RefreshToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate token"})
			return
		}
		if blacklisted {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token has been revoked"})
			return
		}

		// Validate in cache
		_, err = h.tokenStore.ValidateRefreshToken(ctx, req.RefreshToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		}
	}

	// Get user and role
	u, err := h.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	if u == nil || u.IsDeleted() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	roleName := u.Role()
	if roleName == "" {
		roleName = "user"
	}

	// Generate new token pair (token rotation)
	tokens, err := h.jwtManager.GenerateTokenPair(u.ID(), u.Login(), roleName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	// Store new refresh token before revoking the old one to avoid invalidating
	// the session if the storage step fails.
	if err := h.storeRefreshToken(c, u.ID(), tokens.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store refresh token"})
		return
	}

	// Revoke old refresh token
	if h.tokenStore != nil {
		_ = h.tokenStore.RevokeRefreshToken(ctx, req.RefreshToken, h.cfg.JWTRefreshTTL)
	}

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresAt:    tokens.ExpiresAt,
		User: UserInfo{
			ID:        u.ID(),
			Login:     u.Login(),
			Email:     u.Email(),
			Role:      roleName,
			CreatedAt: u.CreatedAt(),
		},
	})
}

// Logout handles user logout
func (h *Handler) Logout(c *gin.Context) {
	middleware.SetDBEntity(c, "refresh_tokens")
	middleware.SetDBOperation(c, "delete")

	// Get refresh token from request body or header
	refreshToken := c.GetHeader("X-Refresh-Token")
	if refreshToken == "" {
		var req RefreshRequest
		if err := c.ShouldBindJSON(&req); err == nil {
			refreshToken = req.RefreshToken
		}
	}

	if refreshToken != "" && h.tokenStore != nil {
		// Revoke refresh token
		_ = h.tokenStore.RevokeRefreshToken(c.Request.Context(), refreshToken, h.cfg.JWTRefreshTTL)
	}

	// Get access token and blacklist it
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && h.tokenStore != nil {
		parts := strings.SplitN(authHeader, " ", 2)
		// Try to extract and blacklist access token
		if len(parts) == 2 {
			_ = h.tokenStore.BlacklistToken(c.Request.Context(), parts[1], h.cfg.JWTAccessTTL)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// ForgotPassword handles password reset request
func (h *Handler) ForgotPassword(c *gin.Context) {
	middleware.SetDBEntity(c, "users")
	middleware.SetDBOperation(c, "read")

	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	u, err := h.userRepo.FindByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find user"})
		return
	}
	if u == nil {
		// Return success even if email not found (security best practice)
		c.JSON(http.StatusOK, gin.H{"message": "if the email exists, a reset link has been sent"})
		return
	}

	// Generate reset token
	token, err := authpkg.GenerateRandomToken(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate reset token"})
		return
	}

	// Store token in Redis
	if h.tokenStore != nil {
		err = h.tokenStore.StorePasswordResetToken(c.Request.Context(), u.ID().String(), token, h.cfg.PasswordResetTTL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store reset token"})
			return
		}
	}

	// TODO: Send email with reset link
	// For now, just return the token in development mode
	c.JSON(http.StatusOK, gin.H{
		"message":     "if the email exists, a reset link has been sent",
		"debug_token": token, // Remove in production
	})
}

// ResetPassword handles password reset
func (h *Handler) ResetPassword(c *gin.Context) {
	middleware.SetDBEntity(c, "users")
	middleware.SetDBOperation(c, "update")

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate password policy
	if err := authpkg.ValidatePassword(req.NewPassword, h.passwordPolicy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate reset token
	if h.tokenStore == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token store not available"})
		return
	}

	userID, err := h.tokenStore.ValidatePasswordResetToken(c.Request.Context(), req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired reset token"})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired reset token"})
		return
	}

	u, err := h.userRepo.FindByID(c.Request.Context(), userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find user"})
		return
	}
	if u == nil || u.IsDeleted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired reset token"})
		return
	}

	// Hash new password
	passwordHash, err := authpkg.HashPassword(req.NewPassword, h.passwordConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	u.SetPasswordHash(passwordHash)
	if err := h.userRepo.Update(c.Request.Context(), u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	// Delete reset token
	_ = h.tokenStore.DeletePasswordResetToken(c.Request.Context(), req.Token)

	c.JSON(http.StatusOK, gin.H{"message": "password has been reset successfully"})
}

// YandexLogin initiates Yandex OAuth login
func (h *Handler) YandexLogin(c *gin.Context) {
	if h.cfg.YandexClientID == "" {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Yandex OAuth not configured"})
		return
	}

	// Generate state and PKCE if enabled
	state, _ := authpkg.GenerateRandomToken(32)

	var codeChallenge string
	if h.cfg.PKCEEnabled {
		pkce, _ := authpkg.GeneratePKCE(h.cfg.PKCECodeChallengeLength)
		codeChallenge = pkce.CodeChallenge
		// Store PKCE data in Redis
		if h.tokenStore != nil {
			_ = h.tokenStore.StorePKCE(c.Request.Context(), state, pkce, 10*time.Minute)
		}
	}

	// Build authorization URL
	authURL := "https://oauth.yandex.com/authorize"
	params := url.Values{
		"client_id":     []string{string(h.cfg.YandexClientID)},
		"response_type": []string{"code"},
		"state":         []string{string(state)},
		"scope":         []string{"login:email login:info"},
	}

	if h.cfg.PKCEEnabled && codeChallenge != "" {
		params.Set("code_challenge", codeChallenge)
		params.Set("code_challenge_method", "plain")
	}

	redirectURL := authURL + "?" + params.Encode()
	c.Redirect(http.StatusFound, redirectURL)
}

// YandexCallback handles Yandex OAuth callback
func (h *Handler) YandexCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "authorization code not provided"})
		return
	}

	// Verify state and get PKCE verifier if enabled
	var codeVerifier string
	if h.cfg.PKCEEnabled && h.tokenStore != nil {
		pkce, err := h.tokenStore.GetPKCE(c.Request.Context(), state)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
			return
		}
		codeVerifier = pkce.CodeVerifier
	}

	// TODO: Exchange code for access token with Yandex
	// This would require making an HTTP request to Yandex OAuth token endpoint
	// For now, return a placeholder

	_ = codeVerifier // Use in actual implementation

	c.JSON(http.StatusOK, gin.H{
		"message": "Yandex OAuth callback received",
		"code":    code,
		"state":   state,
	})
}

// storeRefreshToken stores refresh token hash in database and Redis cache
func (h *Handler) storeRefreshToken(c *gin.Context, userID uuid.UUID, token string) error {
	if h.tokenStore == nil || h.refreshTokenRepo == nil {
		return nil
	}

	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	refreshToken := &authpkg.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		IPAddress: c.ClientIP(),
		ExpiresAt: time.Now().Add(h.cfg.JWTRefreshTTL),
		CreatedAt: time.Now(),
	}

	if err := h.refreshTokenRepo.Create(c.Request.Context(), refreshToken); err != nil {
		return err
	}

	// Also store in Redis so the refresh endpoint can validate it
	return h.tokenStore.StoreRefreshToken(c.Request.Context(), userID.String(), token, refreshToken.ExpiresAt)
}
