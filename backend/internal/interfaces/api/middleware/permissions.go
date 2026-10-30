package middleware

import (
	"net/http"
	"time"

	"knowledge-graph/internal/auth"
	"knowledge-graph/internal/domain/permission"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PermissionConfig holds permission middleware configuration
type PermissionConfig struct {
	Repo       permission.Repository
	TokenStore auth.TokenStore
	CacheTTL   time.Duration
}

// DefaultPermissionConfig returns default permission configuration
func DefaultPermissionConfig(repo permission.Repository, tokenStore auth.TokenStore) *PermissionConfig {
	return &PermissionConfig{
		Repo:       repo,
		TokenStore: tokenStore,
		CacheTTL:   5 * time.Minute,
	}
}

// Can checks if the user has permission for a specific resource and action
func Can(config *PermissionConfig, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		// Check cache first
		if config.TokenStore != nil {
			allowed, cached, err := config.TokenStore.CheckCachedPermission(c.Request.Context(), userID.String(), resource, action)
			if err == nil && cached {
				if !allowed {
					c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
					c.Abort()
					return
				}
				c.Next()
				return
			}
		}

		// Check permissions via repository
		allowed, err := config.Repo.HasPermission(c.Request.Context(), userID, resource, action)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check permissions"})
			c.Abort()
			return
		}

		// Cache the result
		if config.TokenStore != nil {
			_ = config.TokenStore.CachePermission(c.Request.Context(), userID.String(), resource, action, allowed, config.CacheTTL)
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CanOwn checks ownership or permission for resources
func CanOwn(config *PermissionConfig, resource string, getOwnerFunc func(*gin.Context) (uuid.UUID, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		ownerID, err := getOwnerFunc(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check resource ownership"})
			c.Abort()
			return
		}

		if ownerID == userID {
			c.Next()
			return
		}

		allowed, err := config.Repo.HasPermission(c.Request.Context(), userID, resource, "manage")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check permissions"})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole ensures the user has one of the specified roles
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := GetUserRole(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		for _, role := range roles {
			if userRole == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient privileges"})
		c.Abort()
	}
}

// RequireAdmin ensures the user is an admin
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin")
}

// IsOwner checks if the authenticated user owns the resource
func IsOwner(c *gin.Context, ownerID uuid.UUID) bool {
	userID, exists := GetUserID(c)
	if !exists {
		return false
	}
	return userID == ownerID
}

// GetNoteOwner returns a function that extracts note owner from the repository.
func GetNoteOwner(repo permission.Repository, noteID uuid.UUID) func(*gin.Context) (uuid.UUID, error) {
	return func(c *gin.Context) (uuid.UUID, error) {
		return repo.GetNoteOwner(c.Request.Context(), noteID)
	}
}

// NoteAccessMiddleware middleware to check note access
func NoteAccessMiddleware(repo permission.Repository, tokenStore *auth.RedisTokenStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		noteIDStr := c.Param("id")
		if noteIDStr == "" {
			c.Next()
			return
		}

		noteID, err := uuid.Parse(noteIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note ID"})
			c.Abort()
			return
		}

		userID, exists := GetUserID(c)
		if !exists {
			c.Next()
			return
		}

		hasAccess, perm, err := repo.CheckNoteAccess(c.Request.Context(), noteID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check access"})
			c.Abort()
			return
		}

		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			c.Abort()
			return
		}

		c.Set("note_permission", perm)
		c.Next()
	}
}
