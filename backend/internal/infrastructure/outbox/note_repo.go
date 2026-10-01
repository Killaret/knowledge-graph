package outbox

import (
	"context"
	"time"

	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/infrastructure/db/postgres"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NoteRepository decorates the Postgres note repository: every write method
// runs the repository write and its graph_outbox row in one transaction.
// It satisfies note.Repository, so handlers and services keep the domain
// interface and cannot forget the event.
type NoteRepository struct {
	inner *postgres.NoteRepository
	db    *gorm.DB
}

func NewNoteRepository(inner *postgres.NoteRepository, db *gorm.DB) *NoteRepository {
	return &NoteRepository{inner: inner, db: db}
}

func (o *NoteRepository) withTx(ctx context.Context, fn func(txCtx context.Context, tx *gorm.DB) error) error {
	return o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(postgres.ContextWithTx(ctx, tx), tx)
	})
}

func noteDeletedEvents(pairs []postgres.NoteAffectedRow) []Model {
	rows := make([]Model, 0, len(pairs))
	for _, p := range pairs {
		rows = append(rows, noteEvent(EventNoteDeleted, p.ID, p.CreatorID))
	}
	return rows
}

func (o *NoteRepository) Save(ctx context.Context, n *note.Note) error {
	return o.withTx(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		created, err := o.inner.SaveReturning(txCtx, n)
		if err != nil {
			return err
		}
		eventType := EventNoteUpdated
		if created {
			eventType = EventNoteCreated
		}
		return insertEvents(tx, []Model{noteEvent(eventType, n.ID(), n.CreatorID())})
	})
}

func (o *NoteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return o.DeleteBatch(ctx, []uuid.UUID{id})
}

func (o *NoteRepository) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
	return o.withTx(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		pairs, err := o.inner.DeleteBatchReturning(txCtx, ids)
		if err != nil {
			return err
		}
		return insertEvents(tx, noteDeletedEvents(pairs))
	})
}

func (o *NoteRepository) Restore(ctx context.Context, id uuid.UUID) error {
	return o.withTx(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		pairs, err := o.inner.RestoreReturning(txCtx, id)
		if err != nil {
			return err
		}
		rows := make([]Model, 0, len(pairs))
		for _, p := range pairs {
			rows = append(rows, noteEvent(EventNoteUpdated, p.ID, p.CreatorID))
		}
		return insertEvents(tx, rows)
	})
}

// PurgeDeletedBefore is not part of note.Repository — the cleanup worker calls
// it on the decorator so purged rows leave an event like every other write.
func (o *NoteRepository) PurgeDeletedBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	var purged int64
	err := o.withTx(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		var pairs []postgres.NoteAffectedRow
		var err error
		purged, pairs, err = o.inner.PurgeDeletedBeforeReturning(txCtx, cutoff)
		if err != nil {
			return err
		}
		return insertEvents(tx, noteDeletedEvents(pairs))
	})
	return purged, err
}

func (o *NoteRepository) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	return o.inner.FindByID(ctx, id)
}

func (o *NoteRepository) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	return o.inner.FindByIDIncludingDeleted(ctx, id)
}

func (o *NoteRepository) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*note.Note, int64, error) {
	return o.inner.List(ctx, userID, limit, offset)
}

func (o *NoteRepository) Search(ctx context.Context, userID uuid.UUID, query string, limit, offset int) ([]*note.Note, int64, error) {
	return o.inner.Search(ctx, userID, query, limit, offset)
}

func (o *NoteRepository) FindAll(ctx context.Context) ([]*note.Note, error) {
	return o.inner.FindAll(ctx)
}

func (o *NoteRepository) FindAllPaginated(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*note.Note, int64, error) {
	return o.inner.FindAllPaginated(ctx, userID, limit, offset)
}

func (o *NoteRepository) FindComets(ctx context.Context, userID uuid.UUID) ([]*note.Note, error) {
	return o.inner.FindComets(ctx, userID)
}
