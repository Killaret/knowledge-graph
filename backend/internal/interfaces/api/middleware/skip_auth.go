package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SkipAuthConfig holds configuration for skipping authentication
type SkipAuthConfig struct {
	Enabled bool
	// DefaultUserID is used when auth is skipped (optional)
	DefaultUserID uuid.UUID
	// DefaultLogin is used when auth is skipped (optional)
	DefaultLogin string
	// DefaultRole is used when auth is skipped (optional)
	DefaultRole string
}

// DefaultSkipAuthConfig returns default configuration
func DefaultSkipAuthConfig(enabled bool) *SkipAuthConfig {
	return &SkipAuthConfig{
		Enabled:       enabled,
		DefaultUserID: uuid.MustParse("00000000-0000-0000-0000-000000000000"), // System/test user (matches migration 019)
		DefaultLogin:  "testuser",
		DefaultRole:   "test",
	}
}

// SkipAuth middleware allows all requests without authentication
// Sets test user ID for consistent note creation in test mode
func SkipAuth(config *SkipAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config == nil {
			c.Next()
			return
		}

		fmt.Printf("[DEBUG] SkipAuth: Enabled=%t, Path=%s\n", config.Enabled, c.Request.URL.Path)
		if !config.Enabled {
			c.Next()
			return
		}

		// Set test user ID for all requests in test mode
		// Test user must exist in DB (migration 019)
		c.Set(ContextUserIDKey, config.DefaultUserID)
		c.Set(ContextRoleKey, config.DefaultRole)
		c.Set(ContextLoginKey, config.DefaultLogin)
		fmt.Printf("[DEBUG] SkipAuth: Set user ID=%s\n", config.DefaultUserID.String())
		c.Next()
	}
}
