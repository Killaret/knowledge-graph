package note

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// DraftRepository defines the interface for draft persistence
type DraftRepository interface {
	// Save saves a draft to the repository
	Save(ctx context.Context, draft *Draft) error

	// FindByNoteAndUser finds a draft by note ID and user ID
	FindByNoteAndUser(ctx context.Context, noteID, userID uuid.UUID) (*Draft, error)

	// FindActiveByUser finds all active drafts for a user
	FindActiveByUser(ctx context.Context, userID uuid.UUID) ([]*Draft, error)

	// FindByID finds a draft by its ID
	FindByID(ctx context.Context, id uuid.UUID) (*Draft, error)

	// DeleteByID deletes a draft by its ID
	DeleteByID(ctx context.Context, id uuid.UUID) error

	// DeleteExpired deletes drafts that haven't been updated since the given time
	DeleteExpired(ctx context.Context, before time.Time) (int, error)

	// Update updates an existing draft
	Update(ctx context.Context, draft *Draft) error
}
