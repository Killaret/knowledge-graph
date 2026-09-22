package link

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, link *Link) error
	Update(ctx context.Context, link *Link) error
	FindByID(ctx context.Context, id uuid.UUID) (*Link, error)
	FindBySource(ctx context.Context, sourceID uuid.UUID) ([]*Link, error)
	FindByTarget(ctx context.Context, targetID uuid.UUID) ([]*Link, error)
	// FindByPair возвращает все связи направленной пары (source → target).
	FindByPair(ctx context.Context, sourceID, targetID uuid.UUID) ([]*Link, error)
	// SaveUserLink применяет правила ручного создания атомарно: gamma-связь
	// того же типа на той же паре повышается до пользовательской (created=false),
	// gamma-связи других типов на паре переносят происхождение в новую строку
	// и удаляются, отказы пары снимаются. Ручная связь того же типа →
	// ErrDuplicateLink.
	SaveUserLink(ctx context.Context, link *Link) (saved *Link, created bool, err error)
	Delete(ctx context.Context, id uuid.UUID) error
	// DeleteAndSuppress удаляет связь и записывает отказ пары в одной транзакции.
	DeleteAndSuppress(ctx context.Context, link *Link, suppression *Suppression) error
	DeleteBySource(ctx context.Context, sourceID uuid.UUID) error
	FindAll(ctx context.Context) ([]*Link, error)
	// FindAllPaginated возвращает все связи с пагинацией (limit=0 для всех записей)
	FindAllPaginated(ctx context.Context, limit, offset int) ([]*Link, int64, error)
}
