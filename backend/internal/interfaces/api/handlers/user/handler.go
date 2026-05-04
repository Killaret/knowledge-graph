// Package user provides HTTP handlers for user management
package user

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"knowledge-graph/internal/auth"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Handler handles user management requests
type Handler struct {
	db     *gorm.DB
	config *auth.PasswordConfig
	policy *auth.PasswordPolicy
}

// NewHandler creates a new user handler
func NewHandler(db *gorm.DB, passwordConfig *auth.PasswordConfig, passwordPolicy *auth.PasswordPolicy) *Handler {
	return &Handler{
		db:     db,
		config: passwordConfig,
		policy: passwordPolicy,
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

	var user postgres.UserModel
	if result := h.db.Preload("Role").First(&user, userID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	roleName := ""
	if user.Role != nil {
		roleName = user.Role.Name
	}

	c.JSON(http.StatusOK, UserResponse{
		ID:        user.ID,
		Login:     user.Login,
		Email:     user.Email,
		Role:      roleName,
		CreatedAt: user.CreatedAt,
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

	var user postgres.UserModel
	if result := h.db.First(&user, userID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	updates := make(map[string]interface{})

	// Update email if provided
	if req.Email != "" {
		// Check if email is already taken by another user
		var existingUser postgres.UserModel
		if result := h.db.Where("email = ? AND id != ?", req.Email, userID).First(&existingUser); result.Error == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "email already in use"})
			return
		}
		updates["email"] = req.Email
	}

	// Update password if provided
	if req.NewPassword != "" {
		// Validate old password
		if req.OldPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "old password is required"})
			return
		}

		// Verify old password
		valid, err := auth.VerifyPassword(req.OldPassword, user.PasswordHash)
		if err != nil || !valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid old password"})
			return
		}

		// Validate new password policy
		if err := auth.ValidatePassword(req.NewPassword, h.policy); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Hash new password
		passwordHash, err := auth.HashPassword(req.NewPassword, h.config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		updates["password_hash"] = passwordHash
	}

	// Apply updates if any
	if len(updates) > 0 {
		if result := h.db.Model(&user).Updates(updates); result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
			return
		}
	}

	// Reload user data
	h.db.First(&user, userID)

	roleName := ""
	if user.Role != nil {
		roleName = user.Role.Name
	}

	c.JSON(http.StatusOK, UserResponse{
		ID:        user.ID,
		Login:     user.Login,
		Email:     user.Email,
		Role:      roleName,
		CreatedAt: user.CreatedAt,
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

	var user postgres.UserModel
	if result := h.db.First(&user, userID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Verify password
	valid, err := auth.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		return
	}

	// Soft delete
	now := time.Now()
	if result := h.db.Model(&user).Update("deleted_at", &now); result.Error != nil {
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

	var keys []postgres.APIKeyModel
	if result := h.db.Where("user_id = ? AND is_active = ?", userID, true).Find(&keys); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch API keys"})
		return
	}

	response := make([]gin.H, 0, len(keys))
	for _, key := range keys {
		response = append(response, gin.H{
			"id":           key.ID,
			"name":         key.Name,
			"scopes":       key.Scopes,
			"created_at":   key.CreatedAt,
			"expires_at":   key.ExpiresAt,
			"last_used_at": key.LastUsedAt,
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

	// Generate API key
	apiKey, err := auth.GenerateRandomToken(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate API key"})
		return
	}

	// Hash the key for storage
	hash := sha256.Sum256([]byte(apiKey))
	keyHash := hex.EncodeToString(hash[:])

	// Store in database
	key := postgres.APIKeyModel{
		ID:        uuid.New(),
		UserID:    userID,
		KeyHash:   keyHash,
		Name:      req.Name,
		Scopes:    req.Scopes,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	if result := h.db.Create(&key); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save API key"})
		return
	}

	// Return the plain API key (only shown once)
	c.JSON(http.StatusCreated, gin.H{
		"id":         key.ID,
		"api_key":    apiKey, // Only shown once!
		"name":       key.Name,
		"scopes":     key.Scopes,
		"created_at": key.CreatedAt,
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

	// Verify ownership and revoke
	result := h.db.Model(&postgres.APIKeyModel{}).
		Where("id = ? AND user_id = ?", keyID, userID).
		Update("is_active", false)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke API key"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key revoked successfully"})
}
