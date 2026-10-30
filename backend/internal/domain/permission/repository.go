package permission

import (
	"context"

	"github.com/google/uuid"
)

// Repository provides permission and access checks.
// Implemented by infrastructure adapters (e.g. postgres.PermissionRepository).
type Repository interface {
	// HasPermission checks whether a user has a specific permission on a resource.
	HasPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)

	// GetNoteOwner returns the creator of a note.
	GetNoteOwner(ctx context.Context, noteID uuid.UUID) (uuid.UUID, error)

	// CheckNoteAccess verifies whether a user can access a note and returns the granted permission level.
	// The second return value is the permission string (e.g. "owner", "read", "write") when access is granted.
	CheckNoteAccess(ctx context.Context, noteID, userID uuid.UUID) (bool, string, error)
}
