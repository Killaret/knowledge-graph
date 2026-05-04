package middleware

import (
	"net/http"
	"time"

	"knowledge-graph/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PermissionConfig holds permission middleware configuration
type PermissionConfig struct {
	DB         *gorm.DB
	TokenStore *auth.RedisTokenStore
	CacheTTL   time.Duration
}

// DefaultPermissionConfig returns default permission configuration
func DefaultPermissionConfig(db *gorm.DB, tokenStore *auth.RedisTokenStore) *PermissionConfig {
	return &PermissionConfig{
		DB:         db,
		TokenStore: tokenStore,
		CacheTTL:   5 * time.Minute,
	}
}

// RolePermissionModel for database queries
type RolePermissionModel struct {
	Resource string
	Action   string
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

		// Check permissions in database
		allowed, err := checkPermission(config.DB, userID, resource, action)
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

		// Get the owner of the resource
		ownerID, err := getOwnerFunc(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check resource ownership"})
			c.Abort()
			return
		}

		// If user is the owner, allow access
		if ownerID == userID {
			c.Next()
			return
		}

		// Check if user has 'manage' or specific permission
		allowed, err := checkPermission(config.DB, userID, resource, "manage")
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

// checkPermission checks if user has a specific permission
func checkPermission(db *gorm.DB, userID uuid.UUID, resource, action string) (bool, error) {
	// Query using the user_permissions_view created in migration
	var count int64

	err := db.Raw(`
		SELECT COUNT(*) FROM user_permissions_view 
		WHERE user_id = ? AND resource = ? AND action = ?
	`, userID, resource, action).Scan(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
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

// GetOwnerFromNote extracts note owner from database
func GetNoteOwner(db *gorm.DB, noteID uuid.UUID) func(*gin.Context) (uuid.UUID, error) {
	return func(c *gin.Context) (uuid.UUID, error) {
		var creatorID uuid.UUID
		err := db.Raw(`SELECT creator_id FROM notes WHERE id = ? AND deleted_at IS NULL`, noteID).Scan(&creatorID).Error
		if err != nil {
			return uuid.Nil, err
		}
		return creatorID, nil
	}
}

// CheckNoteAccess checks if user has access to a note (owner or shared)
func CheckNoteAccess(db *gorm.DB, tokenStore *auth.RedisTokenStore, noteID uuid.UUID, userID uuid.UUID) (bool, string, error) {
	// Check if owner
	var creatorID uuid.UUID
	err := db.Raw(`SELECT creator_id FROM notes WHERE id = ? AND deleted_at IS NULL`, noteID).Scan(&creatorID).Error
	if err != nil {
		return false, "", err
	}

	if creatorID == userID {
		return true, "owner", nil
	}

	// Check if shared directly
	var sharePermission string
	err = db.Raw(`
		SELECT permission FROM note_shares 
		WHERE note_id = ? AND shared_with_user_id = ? 
		AND (expires_at IS NULL OR expires_at > now())
	`, noteID, userID).Scan(&sharePermission).Error

	if err == nil && sharePermission != "" {
		return true, sharePermission, nil
	}

	return false, "", nil
}

// NoteAccessMiddleware middleware to check note access
func NoteAccessMiddleware(db *gorm.DB, tokenStore *auth.RedisTokenStore) gin.HandlerFunc {
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

		hasAccess, permission, err := CheckNoteAccess(db, tokenStore, noteID, userID)
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

		c.Set("note_permission", permission)
		c.Next()
	}
}
