// Package tag defines the repository contract for Tag persistence.
package tag

import (
	"context"

	"github.com/google/uuid"
)

// Repository is the persistence contract for tags.
// Implementations live in the infrastructure layer.
type Repository interface {
	Create(ctx context.Context, tag *Tag) error
	FindByID(ctx context.Context, id uuid.UUID) (*Tag, error)
	FindByName(ctx context.Context, name string) (*Tag, error)
	FindAll(ctx context.Context) ([]*Tag, error)
	Update(ctx context.Context, tag *Tag) error
	Delete(ctx context.Context, id uuid.UUID) error
	AddTagToNote(ctx context.Context, noteID, tagID uuid.UUID) error
	RemoveTagFromNote(ctx context.Context, noteID, tagID uuid.UUID) error
	GetTagsByNoteID(ctx context.Context, noteID uuid.UUID) ([]*Tag, error)
	IsTagAssignedToNote(ctx context.Context, noteID, tagID uuid.UUID) (bool, error)
}
