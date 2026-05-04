package middleware

import (
	"net/http"
	"strings"

	"knowledge-graph/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Context keys
const (
	ContextUserIDKey = "user_id"
	ContextRoleKey   = "user_role"
	ContextLoginKey  = "user_login"
)

// JWTConfig holds JWT middleware configuration
type JWTConfig struct {
	JWTManager     *auth.JWTManager
	TokenStore     *auth.RedisTokenStore
	SkipPaths      []string
	HeaderName     string
	TokenLookup    string
}

// DefaultJWTConfig returns default JWT configuration
func DefaultJWTConfig(jwtManager *auth.JWTManager, tokenStore *auth.RedisTokenStore) *JWTConfig {
	return &JWTConfig{
		JWTManager:  jwtManager,
		TokenStore:  tokenStore,
		HeaderName:  "Authorization",
		TokenLookup: "header",
		SkipPaths: []string{
			"/api/v1/auth/login",
			"/api/v1/auth/register",
			"/api/v1/auth/refresh",
			"/api/v1/auth/forgot-password",
			"/api/v1/auth/reset-password",
			"/api/v1/auth/yandex/login",
			"/api/v1/auth/yandex/callback",
			"/health",
			"/swagger/*any",
			"/openapi.yaml",
		},
	}
}

// JWTAuth middleware validates JWT tokens
func JWTAuth(config *JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for certain paths
		for _, path := range config.SkipPaths {
			if c.FullPath() == path || (path == "/swagger/*any" && strings.HasPrefix(c.Request.URL.Path, "/swagger/")) {
				c.Next()
				return
			}
		}

		// Extract token
		token, err := extractToken(c, config)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or malformed token"})
			c.Abort()
			return
		}

		// Check if blacklisted
		if config.TokenStore != nil {
			blacklisted, err := config.TokenStore.IsTokenBlacklisted(c.Request.Context(), token)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate token"})
				c.Abort()
				return
			}
			if blacklisted {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "token has been revoked"})
				c.Abort()
				return
			}
		}

		// Validate token
		claims, err := config.JWTManager.ValidateToken(token, "access")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Set context values
		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextRoleKey, claims.Role)
		c.Set(ContextLoginKey, claims.Login)

		c.Next()
	}
}

// extractToken extracts JWT token from request
func extractToken(c *gin.Context, config *JWTConfig) (string, error) {
	// Try header first
	authHeader := c.GetHeader(config.HeaderName)
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return parts[1], nil
		}
		// If no Bearer prefix, assume it's the token itself
		return authHeader, nil
	}

	// Try query parameter
	token := c.Query("token")
	if token != "" {
		return token, nil
	}

	return "", nil
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get(ContextUserIDKey)
	if !exists {
		return uuid.Nil, false
	}
	
	id, ok := userID.(uuid.UUID)
	return id, ok
}

// GetUserRole extracts user role from context
func GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(ContextRoleKey)
	if !exists {
		return "", false
	}
	
	r, ok := role.(string)
	return r, ok
}

// RequireAuth middleware ensures user is authenticated
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := GetUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
