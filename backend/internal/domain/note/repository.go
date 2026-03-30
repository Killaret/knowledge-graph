package note

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, note *Note) error
	FindByID(ctx context.Context, id uuid.UUID) (*Note, error)
	Delete(ctx context.Context, id uuid.UUID) error
	// FindBySpecification — позже добавим
}
