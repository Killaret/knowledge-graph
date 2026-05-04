// Package auth provides HTTP handlers for authentication
package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"strings"
	"time"

	"knowledge-graph/internal/auth"
	"knowledge-graph/internal/config"
	"knowledge-graph/internal/infrastructure/db/postgres"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Handler handles authentication requests
type Handler struct {
	db             *gorm.DB
	jwtManager     *auth.JWTManager
	tokenStore     *auth.RedisTokenStore
	passwordConfig *auth.PasswordConfig
	passwordPolicy *auth.PasswordPolicy
	cfg            *config.Config
}

// NewHandler creates a new auth handler
func NewHandler(db *gorm.DB, jwtManager *auth.JWTManager, tokenStore *auth.RedisTokenStore, cfg *config.Config) *Handler {
	return &Handler{
		db:         db,
		jwtManager: jwtManager,
		tokenStore: tokenStore,
		passwordConfig: &auth.PasswordConfig{
			Time:    cfg.Argon2Time,
			Memory:  cfg.Argon2Memory,
			Threads: cfg.Argon2Threads,
			KeyLen:  32,
		},
		passwordPolicy: &auth.PasswordPolicy{
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
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate password policy
	if err := auth.ValidatePassword(req.Password, h.passwordPolicy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if login already exists
	var existingUser postgres.UserModel
	if result := h.db.Where("login = ?", req.Login).First(&existingUser); result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "login already exists"})
		return
	}

	// Check if email already exists
	if result := h.db.Where("email = ?", req.Email).First(&existingUser); result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password, h.passwordConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Get default role
	var defaultRole postgres.UserRoleModel
	if result := h.db.Where("name = ?", "user").First(&defaultRole); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get default role"})
		return
	}

	// Create user
	user := postgres.UserModel{
		ID:           uuid.New(),
		Login:        req.Login,
		Email:        req.Email,
		PasswordHash: passwordHash,
		RoleID:       &defaultRole.ID,
		CreatedAt:    time.Now(),
	}

	if result := h.db.Create(&user); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// Generate tokens
	tokens, err := h.jwtManager.GenerateTokenPair(user.ID, user.Login, defaultRole.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	// Store refresh token
	h.storeRefreshToken(c, user.ID, tokens.RefreshToken)

	c.JSON(http.StatusCreated, TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresAt:    tokens.ExpiresAt,
		User: UserInfo{
			ID:        user.ID,
			Login:     user.Login,
			Email:     user.Email,
			Role:      defaultRole.Name,
			CreatedAt: user.CreatedAt,
		},
	})
}

// Login handles user login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by login
	var user postgres.UserModel
	if result := h.db.Where("login = ?", req.Login).First(&user); result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Check if user is deleted
	if user.DeletedAt != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "account has been deleted"})
		return
	}

	// Verify password
	valid, err := auth.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Get user role
	roleName := "user"
	if user.RoleID != nil {
		var role postgres.UserRoleModel
		if result := h.db.First(&role, user.RoleID); result.Error == nil {
			roleName = role.Name
		}
	}

	// Generate tokens
	tokens, err := h.jwtManager.GenerateTokenPair(user.ID, user.Login, roleName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	// Store refresh token
	h.storeRefreshToken(c, user.ID, tokens.RefreshToken)

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresAt:    tokens.ExpiresAt,
		User: UserInfo{
			ID:        user.ID,
			Login:     user.Login,
			Email:     user.Email,
			Role:      roleName,
			CreatedAt: user.CreatedAt,
		},
	})
}

// Refresh handles token refresh
func (h *Handler) Refresh(c *gin.Context) {
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

	// Check if blacklisted
	if h.tokenStore != nil {
		blacklisted, err := h.tokenStore.IsTokenBlacklisted(c.Request.Context(), req.RefreshToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate token"})
			return
		}
		if blacklisted {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token has been revoked"})
			return
		}

		// Validate in database
		_, err = h.tokenStore.ValidateRefreshToken(c.Request.Context(), req.RefreshToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		}
	}

	// Get user and role
	var user postgres.UserModel
	if result := h.db.First(&user, claims.UserID); result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	roleName := "user"
	if user.RoleID != nil {
		var role postgres.UserRoleModel
		if result := h.db.First(&role, user.RoleID); result.Error == nil {
			roleName = role.Name
		}
	}

	// Generate new token pair (token rotation)
	tokens, err := h.jwtManager.GenerateTokenPair(user.ID, user.Login, roleName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	// Revoke old refresh token
	if h.tokenStore != nil {
		h.tokenStore.RevokeRefreshToken(c.Request.Context(), req.RefreshToken, h.cfg.JWTRefreshTTL)
	}

	// Store new refresh token
	h.storeRefreshToken(c, user.ID, tokens.RefreshToken)

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresAt:    tokens.ExpiresAt,
		User: UserInfo{
			ID:        user.ID,
			Login:     user.Login,
			Email:     user.Email,
			Role:      roleName,
			CreatedAt: user.CreatedAt,
		},
	})
}

// Logout handles user logout
func (h *Handler) Logout(c *gin.Context) {
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
		h.tokenStore.RevokeRefreshToken(c.Request.Context(), refreshToken, h.cfg.JWTRefreshTTL)
	}

	// Get access token and blacklist it
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && h.tokenStore != nil {
		parts := strings.SplitN(authHeader, " ", 2)
		// Try to extract and blacklist access token
		if len(parts) == 2 {
			h.tokenStore.BlacklistToken(c.Request.Context(), parts[1], h.cfg.JWTAccessTTL)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// ForgotPassword handles password reset request
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	var user postgres.UserModel
	if result := h.db.Where("email = ?", req.Email).First(&user); result.Error != nil {
		// Return success even if email not found (security best practice)
		c.JSON(http.StatusOK, gin.H{"message": "if the email exists, a reset link has been sent"})
		return
	}

	// Generate reset token
	token, err := auth.GenerateRandomToken(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate reset token"})
		return
	}

	// Store token in Redis
	if h.tokenStore != nil {
		err = h.tokenStore.StorePasswordResetToken(c.Request.Context(), user.ID.String(), token, h.cfg.PasswordResetTTL)
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
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate password policy
	if err := auth.ValidatePassword(req.NewPassword, h.passwordPolicy); err != nil {
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

	// Hash new password
	passwordHash, err := auth.HashPassword(req.NewPassword, h.passwordConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Update user password
	userUUID, _ := uuid.Parse(userID)
	if result := h.db.Model(&postgres.UserModel{}).Where("id = ?", userUUID).Update("password_hash", passwordHash); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	// Delete reset token
	h.tokenStore.DeletePasswordResetToken(c.Request.Context(), req.Token)

	c.JSON(http.StatusOK, gin.H{"message": "password has been reset successfully"})
}

// YandexLogin initiates Yandex OAuth login
func (h *Handler) YandexLogin(c *gin.Context) {
	if h.cfg.YandexClientID == "" {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Yandex OAuth not configured"})
		return
	}

	// Generate state and PKCE if enabled
	state, _ := auth.GenerateRandomToken(32)

	var codeChallenge string
	if h.cfg.PKCEEnabled {
		pkce, _ := auth.GeneratePKCE(h.cfg.PKCECodeChallengeLength)
		codeChallenge = pkce.CodeChallenge
		// Store PKCE data in Redis
		if h.tokenStore != nil {
			h.tokenStore.StorePKCE(c.Request.Context(), state, pkce, 10*time.Minute)
		}
	}

	// Build authorization URL
	authURL := "https://oauth.yandex.com/authorize"
	params := url.Values{
		"client_id":     {h.cfg.YandexClientID},
		"response_type": {"code"},
		"state":         {state},
		"scope":         {"login:email login:info"},
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

// storeRefreshToken stores refresh token hash in database
func (h *Handler) storeRefreshToken(c *gin.Context, userID uuid.UUID, token string) {
	if h.tokenStore == nil {
		return
	}

	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	refreshToken := postgres.RefreshTokenModel{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		IPAddress: c.ClientIP(),
		ExpiresAt: time.Now().Add(h.cfg.JWTRefreshTTL),
		CreatedAt: time.Now(),
	}

	h.db.Create(&refreshToken)
}
