package outbox

import (
	"context"

	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/infrastructure/db/postgres"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LinkRepository decorates the Postgres link repository the same way
// NoteRepository decorates the note repository: link writes and their
// graph_outbox rows commit or roll back together. It satisfies both
// link.Repository and link.SuppressionRepository.
type LinkRepository struct {
	inner *postgres.LinkRepository
	db    *gorm.DB
}

func NewLinkRepository(inner *postgres.LinkRepository, db *gorm.DB) *LinkRepository {
	return &LinkRepository{inner: inner, db: db}
}

func (o *LinkRepository) withTx(ctx context.Context, fn func(txCtx context.Context, tx *gorm.DB) error) error {
	return o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(postgres.ContextWithTx(ctx, tx), tx)
	})
}

// linkOwner resolves the event's user_id: the link's own creator when set
// (manual links), else the source note's creator — gamma rows carry no
// creator but still belong to the note owner's graph.
func linkOwner(tx *gorm.DB, creatorID *uuid.UUID, sourceID uuid.UUID) *uuid.UUID {
	if creatorID != nil {
		return creatorID
	}
	var owner uuid.NullUUID
	if err := tx.Unscoped().Raw("SELECT creator_id FROM notes WHERE id = ?", sourceID).Scan(&owner).Error; err != nil {
		return nil
	}
	if !owner.Valid {
		return nil
	}
	id := owner.UUID
	return &id
}

func linkDeletedEvents(tx *gorm.DB, rows []postgres.LinkAffectedRow) []Model {
	events := make([]Model, 0, len(rows))
	for _, r := range rows {
		events = append(events, linkEvent(EventLinkDeleted, r.ID, r.SourceNoteID, r.TargetNoteID,
			linkOwner(tx, r.CreatorID, r.SourceNoteID)))
	}
	return events
}

func (o *LinkRepository) Save(ctx context.Context, l *link.Link) error {
	return o.withTx(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		created, err := o.inner.SaveReturning(txCtx, l)
		if err != nil {
			return err
		}
		eventType := EventLinkUpdated
		if created {
			eventType = EventLinkCreated
		}
		return insertEvents(tx, []Model{linkEvent(eventType, l.ID(), l.SourceNoteID(), l.TargetNoteID(),
			linkOwner(tx, l.CreatorID(), l.SourceNoteID()))})
	})
}

func (o *LinkRepository) Update(ctx context.Context, l *link.Link) error {
	return o.withTx(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		if err := o.inner.Update(txCtx, l); err != nil {
			return err
		}
		return insertEvents(tx, []Model{linkEvent(EventLinkUpdated, l.ID(), l.SourceNoteID(), l.TargetNoteID(),
			linkOwner(tx, l.CreatorID(), l.SourceNoteID()))})
	})
}

func (o *LinkRepository) SaveUserLink(ctx context.Context, l *link.Link) (*link.Link, bool, error) {
	var saved *link.Link
	created := false
	err := o.withTx(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		var err error
		saved, created, err = o.inner.SaveUserLink(txCtx, l)
		if err != nil {
			return err
		}
		return insertEvents(tx, []Model{linkEvent(EventLinkCreated, saved.ID(), saved.SourceNoteID(), saved.TargetNoteID(),
			linkOwner(tx, saved.CreatorID(), saved.SourceNoteID()))})
	})
	return saved, created, err
}

func (o *LinkRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return o.withTx(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		rows, err := o.inner.DeleteReturning(txCtx, id)
		if err != nil {
			return err
		}
		return insertEvents(tx, linkDeletedEvents(tx, rows))
	})
}

func (o *LinkRepository) DeleteAndSuppress(ctx context.Context, l *link.Link, s *link.Suppression) error {
	return o.withTx(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		if err := o.inner.DeleteAndSuppress(txCtx, l, s); err != nil {
			return err
		}
		return insertEvents(tx, []Model{linkEvent(EventLinkDeleted, l.ID(), l.SourceNoteID(), l.TargetNoteID(),
			linkOwner(tx, l.CreatorID(), l.SourceNoteID()))})
	})
}

func (o *LinkRepository) DeleteBySource(ctx context.Context, sourceID uuid.UUID) error {
	return o.withTx(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		rows, err := o.inner.DeleteBySourceReturning(txCtx, sourceID)
		if err != nil {
			return err
		}
		return insertEvents(tx, linkDeletedEvents(tx, rows))
	})
}

// DeleteBySourceType is not part of link.Repository — gamma-link regeneration
// calls it on the decorator so the wipe still publishes LinkDeleted per row.
func (o *LinkRepository) DeleteBySourceType(ctx context.Context, sourceType string) (int64, error) {
	var count int64
	err := o.withTx(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		var rows []postgres.LinkAffectedRow
		var err error
		count, rows, err = o.inner.DeleteBySourceTypeReturning(txCtx, sourceType)
		if err != nil {
			return err
		}
		return insertEvents(tx, linkDeletedEvents(tx, rows))
	})
	return count, err
}

// SaveSuppression records a pair rejection. The pair's link state changed, so
// the owner's caches invalidate through a LinkUpdated event.
func (o *LinkRepository) SaveSuppression(ctx context.Context, s *link.Suppression) error {
	return o.withTx(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		if err := o.inner.SaveSuppression(txCtx, s); err != nil {
			return err
		}
		return insertEvents(tx, []Model{linkEvent(EventLinkUpdated, s.ID(), s.NoteAID(), s.NoteBID(), s.CreatorID())})
	})
}

func (o *LinkRepository) FindByID(ctx context.Context, id uuid.UUID) (*link.Link, error) {
	return o.inner.FindByID(ctx, id)
}

func (o *LinkRepository) FindBySource(ctx context.Context, sourceID uuid.UUID) ([]*link.Link, error) {
	return o.inner.FindBySource(ctx, sourceID)
}

func (o *LinkRepository) FindByTarget(ctx context.Context, targetID uuid.UUID) ([]*link.Link, error) {
	return o.inner.FindByTarget(ctx, targetID)
}

func (o *LinkRepository) FindBySourceIDs(ctx context.Context, sourceIDs []uuid.UUID) (map[uuid.UUID][]*link.Link, error) {
	return o.inner.FindBySourceIDs(ctx, sourceIDs)
}

func (o *LinkRepository) FindByTargetIDs(ctx context.Context, targetIDs []uuid.UUID) (map[uuid.UUID][]*link.Link, error) {
	return o.inner.FindByTargetIDs(ctx, targetIDs)
}

func (o *LinkRepository) FindByPair(ctx context.Context, sourceID, targetID uuid.UUID) ([]*link.Link, error) {
	return o.inner.FindByPair(ctx, sourceID, targetID)
}

func (o *LinkRepository) FindAll(ctx context.Context) ([]*link.Link, error) {
	return o.inner.FindAll(ctx)
}

func (o *LinkRepository) FindAllPaginated(ctx context.Context, limit, offset int) ([]*link.Link, int64, error) {
	return o.inner.FindAllPaginated(ctx, limit, offset)
}

func (o *LinkRepository) FindBySourceType(ctx context.Context, sourceType string) ([]*link.Link, error) {
	return o.inner.FindBySourceType(ctx, sourceType)
}

func (o *LinkRepository) CountBySourceType(ctx context.Context, sourceType string) (int64, error) {
	return o.inner.CountBySourceType(ctx, sourceType)
}

func (o *LinkRepository) FindSuppressionsForNotes(ctx context.Context, noteIDs []uuid.UUID) ([]*link.Suppression, error) {
	return o.inner.FindSuppressionsForNotes(ctx, noteIDs)
}
