// Package user provides HTTP handlers for user management
package user

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"knowledge-graph/internal/auth"
	"knowledge-graph/internal/domain/user"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles user management requests
type Handler struct {
	repo           user.Repository
	apiKeyRepo     user.APIKeyRepository
	passwordConfig *auth.PasswordConfig
	passwordPolicy *auth.PasswordPolicy
}

// NewHandler creates a new user handler
func NewHandler(
	repo user.Repository,
	apiKeyRepo user.APIKeyRepository,
	passwordConfig *auth.PasswordConfig,
	passwordPolicy *auth.PasswordPolicy,
) *Handler {
	return &Handler{
		repo:           repo,
		apiKeyRepo:     apiKeyRepo,
		passwordConfig: passwordConfig,
		passwordPolicy: passwordPolicy,
	}
}

// UserResponse represents a user response
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Login     string    `json:"login"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// UpdateUserRequest represents a request to update user data
type UpdateUserRequest struct {
	Email       string `json:"email,omitempty"`
	OldPassword string `json:"old_password,omitempty"`
	NewPassword string `json:"new_password,omitempty"`
}

// GetMe returns the current authenticated user's data
func (h *Handler) GetMe(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	u, err := h.repo.FindByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}
	if u == nil || u.IsDeleted() {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, UserResponse{
		ID:        u.ID(),
		Login:     u.Login(),
		Email:     u.Email(),
		Role:      u.Role(),
		CreatedAt: u.CreatedAt(),
	})
}

// UpdateMe updates the current authenticated user's data
func (h *Handler) UpdateMe(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.repo.FindByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}
	if u == nil || u.IsDeleted() {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if req.Email != "" {
		taken, err := h.repo.EmailExists(c.Request.Context(), req.Email, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check email"})
			return
		}
		if taken {
			c.JSON(http.StatusConflict, gin.H{"error": "email already in use"})
			return
		}
		u.SetEmail(req.Email)
	}

	if req.NewPassword != "" {
		if req.OldPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "old password is required"})
			return
		}
		valid, err := auth.VerifyPassword(req.OldPassword, u.PasswordHash())
		if err != nil || !valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid old password"})
			return
		}
		if err := auth.ValidatePassword(req.NewPassword, h.passwordPolicy); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		hash, err := auth.HashPassword(req.NewPassword, h.passwordConfig)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		u.SetPasswordHash(hash)
	}

	if err := h.repo.Update(c.Request.Context(), u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	c.JSON(http.StatusOK, UserResponse{
		ID:        u.ID(),
		Login:     u.Login(),
		Email:     u.Email(),
		Role:      u.Role(),
		CreatedAt: u.CreatedAt(),
	})
}

// DeleteMe performs soft delete of the current authenticated user
func (h *Handler) DeleteMe(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.repo.FindByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}
	if u == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	valid, err := auth.VerifyPassword(req.Password, u.PasswordHash())
	if err != nil || !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		return
	}

	if err := h.repo.SoftDelete(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account deleted successfully"})
}

// ListAPIKeys returns API keys for the current user
func (h *Handler) ListAPIKeys(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	keys, err := h.apiKeyRepo.FindByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch API keys"})
		return
	}

	response := make([]gin.H, 0, len(keys))
	for _, key := range keys {
		response = append(response, gin.H{
			"id":           key.ID(),
			"name":         key.Name(),
			"scopes":       key.Scopes(),
			"created_at":   key.CreatedAt(),
			"expires_at":   key.ExpiresAt(),
			"last_used_at": key.LastUsedAt(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"api_keys": response})
}

// CreateAPIKeyRequest represents a request to create an API key
type CreateAPIKeyRequest struct {
	Name   string   `json:"name" binding:"required"`
	Scopes []string `json:"scopes"`
}

// CreateAPIKey creates a new API key for the current user
func (h *Handler) CreateAPIKey(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiKey, err := auth.GenerateRandomToken(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate API key"})
		return
	}

	hash := sha256.Sum256([]byte(apiKey))
	keyHash := hex.EncodeToString(hash[:])

	key, err := user.NewAPIKey(uuid.New(), userID, keyHash, req.Name, req.Scopes, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create API key"})
		return
	}

	if err := h.apiKeyRepo.Create(c.Request.Context(), key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save API key"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         key.ID(),
		"api_key":    apiKey, // Only shown once!
		"name":       key.Name(),
		"scopes":     key.Scopes(),
		"created_at": key.CreatedAt(),
	})
}

// RevokeAPIKey revokes an API key
func (h *Handler) RevokeAPIKey(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	keyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid API key ID"})
		return
	}

	revoked, err := h.apiKeyRepo.Revoke(c.Request.Context(), keyID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke API key"})
		return
	}
	if !revoked {
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key revoked successfully"})
}
