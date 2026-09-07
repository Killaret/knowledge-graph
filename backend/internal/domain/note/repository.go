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
	DeleteBatch(ctx context.Context, ids []uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID) error
	// List возвращает заметки пользователя. userID = uuid.Nil — только публичные.
	List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Note, int64, error)
	// Search ищет по заметкам пользователя. userID = uuid.Nil — только публичные.
	Search(ctx context.Context, userID uuid.UUID, query string, limit, offset int) ([]*Note, int64, error)
	FindAll(ctx context.Context) ([]*Note, error)
	// FindAllPaginated возвращает все заметки с пагинацией (limit=0 для всех записей)
	FindAllPaginated(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Note, int64, error)
	// FindBySpecification — позже добавим
}
