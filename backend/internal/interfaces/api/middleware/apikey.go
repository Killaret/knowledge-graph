package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// APIKeyConfig holds API key middleware configuration
type APIKeyConfig struct {
	DB            *gorm.DB
	HeaderName    string
	Enabled       bool
	SkipPaths     []string
}

// DefaultAPIKeyConfig returns default API key configuration
func DefaultAPIKeyConfig(db *gorm.DB, enabled bool) *APIKeyConfig {
	return &APIKeyConfig{
		DB:         db,
		HeaderName: "X-API-Key",
		Enabled:    enabled,
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
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID `gorm:"type:uuid;not null"`
	KeyHash    string    `gorm:"not null;index"`
	IsActive   bool      `gorm:"default:true"`
	ExpiresAt  *string
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

		// Hash the API key for lookup
		hash := hashAPIKey(apiKey)

		// Look up the API key in the database
		var keyModel APIKeyModel
		result := config.DB.Where("key_hash = ? AND is_active = ?", hash, true).First(&keyModel)
		
		if result.Error != nil {
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
		config.DB.Model(&keyModel).Update("last_used_at", "now()")

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
