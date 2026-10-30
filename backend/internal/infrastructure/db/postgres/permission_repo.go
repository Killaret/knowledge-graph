package postgres

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PermissionRepository provides permission and access checks.
type PermissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository creates a new permission repository.
func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// HasPermission checks whether a user has a specific permission on a resource.
func (r *PermissionRepository) HasPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM user_permissions_view
		WHERE user_id = ? AND resource = ? AND action = ?
	`, userID, resource, action).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetNoteOwner returns the creator of a note.
func (r *PermissionRepository) GetNoteOwner(ctx context.Context, noteID uuid.UUID) (uuid.UUID, error) {
	var creatorID uuid.UUID
	err := r.db.WithContext(ctx).Raw(`
		SELECT creator_id FROM notes WHERE id = ? AND deleted_at IS NULL
	`, noteID).Scan(&creatorID).Error
	if err != nil {
		return uuid.Nil, err
	}
	return creatorID, nil
}

// CheckNoteAccess verifies whether a user can access a note.
func (r *PermissionRepository) CheckNoteAccess(ctx context.Context, noteID, userID uuid.UUID) (bool, string, error) {
	// Check if owner
	var creatorID uuid.UUID
	err := r.db.WithContext(ctx).Raw(`
		SELECT creator_id FROM notes WHERE id = ? AND deleted_at IS NULL
	`, noteID).Scan(&creatorID).Error
	if err != nil {
		return false, "", err
	}

	if creatorID == userID {
		return true, "owner", nil
	}

	// Check if shared directly
	var sharePermission string
	err = r.db.WithContext(ctx).Raw(`
		SELECT permission FROM note_shares
		WHERE note_id = ? AND shared_with_user_id = ?
		AND (expires_at IS NULL OR expires_at > now())
	`, noteID, userID).Scan(&sharePermission).Error

	if err == nil && sharePermission != "" {
		return true, sharePermission, nil
	}

	return false, "", nil
}
