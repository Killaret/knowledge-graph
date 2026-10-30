package middleware

import (
	"net/http"
	"strings"
	"time"

	"knowledge-graph/internal/auth"
	"knowledge-graph/internal/domain/user"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// APIKeyConfig holds API key middleware configuration
type APIKeyConfig struct {
	Repo         user.APIKeyRepository
	HeaderName   string
	Enabled      bool
	StaticAPIKey string
	SkipPaths    []string
}

// DefaultAPIKeyConfig returns default API key configuration
func DefaultAPIKeyConfig(repo user.APIKeyRepository, enabled bool, staticAPIKey string) *APIKeyConfig {
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
			adminUUID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
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

		// API key token format: <id>:<secret>
		parts := strings.SplitN(apiKey, ":", 2)
		if len(parts) != 2 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			c.Abort()
			return
		}

		keyID, err := uuid.Parse(parts[0])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			c.Abort()
			return
		}

		// Look up the API key in the database
		key, err := config.Repo.FindActiveByID(c.Request.Context(), keyID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			c.Abort()
			return
		}
		if key == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			c.Abort()
			return
		}

		// Verify the secret against the stored Argon2 hash
		secret := parts[1]
		ok, err := auth.VerifyPassword(secret, key.KeyHash())
		if err != nil || !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			c.Abort()
			return
		}

		// Check if expired
		if expiresAt := key.ExpiresAt(); expiresAt != nil && expiresAt.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key has expired"})
			c.Abort()
			return
		}

		// Update last used time
		_ = config.Repo.UpdateLastUsed(c.Request.Context(), key.ID())

		// Set context values
		c.Set(ContextUserIDKey, key.UserID())
		c.Set(ContextRoleKey, "api_key")
		c.Set("api_key_id", key.ID())

		c.Next()
	}
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
