package middleware

import (
	"context"
	contextkeys "knowledge-graph/internal/shared/context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// IsSkipAuth returns true if the context was created with SKIP_AUTH enabled.
func IsSkipAuth(ctx context.Context) bool {
	v, _ := ctx.Value(contextkeys.SkipAuthKey).(bool)
	return v
}

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

		if !config.Enabled {
			c.Next()
			return
		}

		// Set test user ID for all requests in test mode.
		// Test user is created by the test seeder (backend/cmd/seed) when APP_ENV=test.
		c.Set(ContextUserIDKey, config.DefaultUserID)
		c.Set(ContextRoleKey, config.DefaultRole)
		c.Set(ContextLoginKey, config.DefaultLogin)
		// Propagate skip-auth flag to request context so repositories can opt-out of public-only scoping
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), contextkeys.SkipAuthKey, true))
		c.Next()
	}
}
