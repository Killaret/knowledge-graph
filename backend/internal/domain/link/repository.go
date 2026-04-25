package link

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, link *Link) error
	FindByID(ctx context.Context, id uuid.UUID) (*Link, error)
	FindBySource(ctx context.Context, sourceID uuid.UUID) ([]*Link, error)
	FindByTarget(ctx context.Context, targetID uuid.UUID) ([]*Link, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteBySource(ctx context.Context, sourceID uuid.UUID) error
	FindAll(ctx context.Context) ([]*Link, error)
	// FindAllPaginated возвращает все связи с пагинацией (limit=0 для всех записей)
	FindAllPaginated(ctx context.Context, limit, offset int) ([]*Link, int64, error)
}
