package note

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// Common repository errors
var (
	ErrNoteNotFound = errors.New("note not found")
)

type Repository interface {
	Save(ctx context.Context, note *Note) error
	FindByID(ctx context.Context, id uuid.UUID) (*Note, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Note, int64, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*Note, int64, error)
	// FindBySpecification — позже добавим
}
