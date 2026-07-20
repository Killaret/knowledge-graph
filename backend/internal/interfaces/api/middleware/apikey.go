package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// APIKeyRepository abstracts persistence for API keys.
// It lives in the middleware package because APIKeyModel is local.
type APIKeyRepository interface {
	FindActiveByHash(ctx context.Context, hash string) (*APIKeyModel, error)
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
}

// APIKeyConfig holds API key middleware configuration
type APIKeyConfig struct {
	Repo         APIKeyRepository
	HeaderName   string
	Enabled      bool
	StaticAPIKey string
	SkipPaths    []string
}

// DefaultAPIKeyConfig returns default API key configuration
type gormAPIKeyRepository struct {
	db *gorm.DB
}

func (r *gormAPIKeyRepository) FindActiveByHash(ctx context.Context, hash string) (*APIKeyModel, error) {
	var key APIKeyModel
	err := r.db.WithContext(ctx).Where("key_hash = ? AND is_active = ?", hash, true).First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *gormAPIKeyRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&APIKeyModel{}).Where("id = ?", id).Update("last_used_at", "now()").Error
}

func DefaultAPIKeyConfig(db *gorm.DB, enabled bool, staticAPIKey string) *APIKeyConfig {
	var repo APIKeyRepository
	if db != nil {
		repo = &gormAPIKeyRepository{db: db}
	}

	return &APIKeyConfig{
		Repo:         repo,
		HeaderName:   "X-API-Key",
		Enabled:      enabled,
		StaticAPIKey: staticAPIKey,
		SkipPaths: []string{
			"/api/v1/auth/*",
			"/health",
			"/swagger/*",
			"/openapi.yaml",
		},
	}
}

// APIKeyModel represents the database model for API keys
// This is a minimal representation for the middleware query
type APIKeyModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	KeyHash   string    `gorm:"not null;index"`
	IsActive  bool      `gorm:"default:true"`
	ExpiresAt *string
}

func (APIKeyModel) TableName() string { return "api_keys" }

// APIKey middleware validates API keys
func APIKey(config *APIKeyConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if API key auth is disabled
		if !config.Enabled {
			c.Next()
			return
		}

		// Skip for certain paths
		for _, path := range config.SkipPaths {
			if strings.HasSuffix(path, "/*") {
				prefix := strings.TrimSuffix(path, "/*")
				if strings.HasPrefix(c.Request.URL.Path, prefix) {
					c.Next()
					return
				}
			}
			if c.FullPath() == path || c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		// Check if already authenticated via JWT
		_, authenticated := GetUserID(c)
		if authenticated {
			c.Next()
			return
		}

		// Extract API key
		apiKey := c.GetHeader(config.HeaderName)
		if apiKey == "" {
			// No API key provided, continue to next auth method
			c.Next()
			return
		}

		// Check static API key first
		if config.StaticAPIKey != "" && apiKey == config.StaticAPIKey {
			// Static API key authenticated - set admin context
			adminUUID := uuid.MustParse("00000000-0000-0000-0000-000000000000") // Test user (matches migration 019)
			c.Set(ContextUserIDKey, adminUUID)
			c.Set(ContextRoleKey, "admin")
			c.Set("api_key_id", adminUUID)
			c.Next()
			return
		}

		if config.Repo == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			c.Abort()
			return
		}

		// Hash the API key for lookup
		hash := hashAPIKey(apiKey)

		// Look up the API key in the database
		keyModel, err := config.Repo.FindActiveByHash(c.Request.Context(), hash)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			c.Abort()
			return
		}

		// Check if expired
		if keyModel.ExpiresAt != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key has expired"})
			c.Abort()
			return
		}

		// Update last used time
		_ = config.Repo.UpdateLastUsed(c.Request.Context(), keyModel.ID)

		// Set context values
		c.Set(ContextUserIDKey, keyModel.UserID)
		c.Set(ContextRoleKey, "api_key") // Special role for API key access
		c.Set("api_key_id", keyModel.ID)

		c.Next()
	}
}

// hashAPIKey creates a SHA256 hash of the API key
func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// GetAPIKeyID extracts API key ID from context
func GetAPIKeyID(c *gin.Context) (uuid.UUID, bool) {
	keyID, exists := c.Get("api_key_id")
	if !exists {
		return uuid.Nil, false
	}

	id, ok := keyID.(uuid.UUID)
	return id, ok
}
