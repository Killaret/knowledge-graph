package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"knowledge-graph/internal/domain/link"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LinkRepository struct {
	db *gorm.DB
}

func NewLinkRepository(db *gorm.DB) *LinkRepository {
	return &LinkRepository{db: db}
}

// LinkAffectedRow is the pair metadata a link write touched, reported via
// RETURNING so the outbox decorator (SYNC-1 A2) can record Link* events with
// the right endpoints and owner.
type LinkAffectedRow struct {
	ID           uuid.UUID
	SourceNoteID uuid.UUID
	TargetNoteID uuid.UUID
	CreatorID    *uuid.UUID
}

func (r *LinkRepository) Save(ctx context.Context, l *link.Link) error {
	_, err := r.SaveReturning(ctx, l)
	return err
}

// SaveReturning behaves like Save and reports whether the row was created
// (true) or updated (false) — the outbox decorator maps that to
// LinkCreated/LinkUpdated.
func (r *LinkRepository) SaveReturning(ctx context.Context, l *link.Link) (bool, error) {
	var existing LinkModel
	err := dbFromContext(ctx, r.db).Where("id = ? AND deleted_at IS NULL", l.ID()).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		model, err := toGormLink(l)
		if err != nil {
			log.Printf("[LinkRepository.Save] toGormLink failed: %v", err)
			return false, err
		}
		if err := dbFromContext(ctx, r.db).Create(&model).Error; err != nil {
			log.Printf("[LinkRepository.Save] Create failed: id=%s source=%s target=%s error=%v",
				model.ID, model.SourceNoteID, model.TargetNoteID, err)
			// Проверяем на нарушение уникального ограничения (PostgreSQL код 23505)
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
				return false, link.ErrDuplicateLink
			}
			return false, err
		}
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, r.updateModel(ctx, &existing, l)
}

func (r *LinkRepository) Update(ctx context.Context, l *link.Link) error {
	var existing LinkModel
	err := dbFromContext(ctx, r.db).Where("id = ? AND deleted_at IS NULL", l.ID()).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return link.ErrLinkNotFound
	}
	if err != nil {
		return err
	}
	return r.updateModel(ctx, &existing, l)
}

func (r *LinkRepository) updateModel(ctx context.Context, existing *LinkModel, l *link.Link) error {
	model, err := toGormLink(l)
	if err != nil {
		return err
	}
	return dbFromContext(ctx, r.db).Model(existing).Updates(model).Error
}

func (r *LinkRepository) FindByID(ctx context.Context, id uuid.UUID) (*link.Link, error) {
	var model LinkModel
	err := dbFromContext(ctx, r.db).Where("id = ? AND deleted_at IS NULL", id).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toDomainLink(&model)
}

func (r *LinkRepository) FindBySource(ctx context.Context, sourceID uuid.UUID) ([]*link.Link, error) {
	var models []LinkModel
	err := dbFromContext(ctx, r.db).Where("source_note_id = ? AND deleted_at IS NULL", sourceID).Find(&models).Error
	if err != nil {
		return nil, err
	}
	return toDomainLinks(models), nil
}

func (r *LinkRepository) FindByTarget(ctx context.Context, targetID uuid.UUID) ([]*link.Link, error) {
	var models []LinkModel
	err := dbFromContext(ctx, r.db).Where("target_note_id = ? AND deleted_at IS NULL", targetID).Find(&models).Error
	if err != nil {
		return nil, err
	}
	return toDomainLinks(models), nil
}

// FindBySourceIDs возвращает связи для нескольких source note ID (batch-запрос)
func (r *LinkRepository) FindBySourceIDs(ctx context.Context, sourceIDs []uuid.UUID) (map[uuid.UUID][]*link.Link, error) {
	if len(sourceIDs) == 0 {
		return make(map[uuid.UUID][]*link.Link), nil
	}

	var models []LinkModel
	err := dbFromContext(ctx, r.db).Where("source_note_id IN ? AND deleted_at IS NULL", sourceIDs).Find(&models).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID][]*link.Link)
	for _, m := range models {
		l, err := toDomainLink(&m)
		if err != nil {
			continue
		}
		result[m.SourceNoteID] = append(result[m.SourceNoteID], l)
	}
	return result, nil
}

// FindByTargetIDs возвращает связи для нескольких target note ID (batch-запрос)
func (r *LinkRepository) FindByTargetIDs(ctx context.Context, targetIDs []uuid.UUID) (map[uuid.UUID][]*link.Link, error) {
	if len(targetIDs) == 0 {
		return make(map[uuid.UUID][]*link.Link), nil
	}

	var models []LinkModel
	err := dbFromContext(ctx, r.db).Where("target_note_id IN ? AND deleted_at IS NULL", targetIDs).Find(&models).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID][]*link.Link)
	for _, m := range models {
		l, err := toDomainLink(&m)
		if err != nil {
			continue
		}
		result[m.TargetNoteID] = append(result[m.TargetNoteID], l)
	}
	return result, nil
}

// FindByPair возвращает все связи направленной пары (source → target).
func (r *LinkRepository) FindByPair(ctx context.Context, sourceID, targetID uuid.UUID) ([]*link.Link, error) {
	var models []LinkModel
	err := dbFromContext(ctx, r.db).
		Where("source_note_id = ? AND target_note_id = ? AND deleted_at IS NULL", sourceID, targetID).
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	return toDomainLinks(models), nil
}

// SaveUserLink applies the manual-create rules atomically (LINKS-2):
//   - a gamma row of the same type on the same pair — in either direction —
//     is promoted in place (created=false) — the row keeps its id and
//     created_at;
//   - gamma rows of other types on the pair are removed and their provenance
//     moves into the new row's metadata.gamma;
//   - a manual row of the same type on the pair (either direction) yields
//     ErrDuplicateLink — decision 53: one edge per pair;
//   - rejections (link_suppressions) for the normalized pair with NULL or
//     matching link_type are lifted in the same transaction.
func (r *LinkRepository) SaveUserLink(ctx context.Context, l *link.Link) (*link.Link, bool, error) {
	var result *link.Link
	created := true

	err := dbFromContext(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		var models []LinkModel
		if err := tx.
			Where("(source_note_id = ? AND target_note_id = ? OR source_note_id = ? AND target_note_id = ?) AND deleted_at IS NULL",
				l.SourceNoteID(), l.TargetNoteID(), l.TargetNoteID(), l.SourceNoteID()).
			Find(&models).Error; err != nil {
			return err
		}

		var sameType *LinkModel
		var gammas []LinkModel
		for i := range models {
			m := models[i]
			if m.LinkType == l.LinkType().String() {
				if m.SourceType == "user" {
					return link.ErrDuplicateLink
				}
				// Gamma can sit on the pair in both directions; promote the row
				// already matching the request — rewriting a reverse-direction
				// row would collide with the same-direction one.
				if sameType == nil ||
					(m.SourceNoteID == l.SourceNoteID() && m.TargetNoteID == l.TargetNoteID()) {
					sameType = &models[i]
				}
			}
			if m.SourceType == "gamma" {
				gammas = append(gammas, m)
			}
		}

		if sameType != nil {
			// Promotion: keep row id and created_at, take request fields.
			gamma, err := toDomainLink(sameType)
			if err != nil {
				return err
			}
			gamma.PromoteToUser(l.CreatorID(), l.LinkType(), l.Weight(), l.Metadata())
			if gamma.SourceNoteID() != l.SourceNoteID() || gamma.TargetNoteID() != l.TargetNoteID() {
				gamma = link.ReconstructLinkWithCreator(gamma.ID(), l.SourceNoteID(), l.TargetNoteID(),
					gamma.LinkType(), gamma.Weight(), gamma.Metadata(), gamma.SourceType(),
					gamma.CreatorID(), gamma.CreatedAt(), gamma.UpdatedAt(), gamma.LastWeightUpdate())
			}
			promoted, err := toGormLink(gamma)
			if err != nil {
				return err
			}
			// When the gamma row lives in the opposite direction, promotion
			// adopts the user's direction — one edge per pair (decision 53).
			if err := tx.Model(&LinkModel{}).Where("id = ?", sameType.ID).Updates(map[string]interface{}{
				"source_note_id": l.SourceNoteID(),
				"target_note_id": l.TargetNoteID(),
				"link_type":      promoted.LinkType,
				"weight":         promoted.Weight,
				"metadata":       promoted.Metadata,
				"source_type":    promoted.SourceType,
				"creator_id":     promoted.CreatorID,
				"updated_at":     promoted.UpdatedAt,
			}).Error; err != nil {
				return err
			}
			// Other gamma rows on the pair do not survive confirmation.
			for _, g := range gammas {
				if g.ID == sameType.ID {
					continue
				}
				if err := tx.Delete(&LinkModel{}, "id = ?", g.ID).Error; err != nil {
					return err
				}
			}
			result = gamma
			created = false
		} else {
			// New row: consume provenance of gamma rows on the pair.
			if len(gammas) > 0 {
				donor, err := toDomainLink(&gammas[0])
				if err != nil {
					return err
				}
				l.InheritGammaProvenance(donor)
				for _, g := range gammas {
					if err := tx.Delete(&LinkModel{}, "id = ?", g.ID).Error; err != nil {
						return err
					}
				}
			}
			model, err := toGormLink(l)
			if err != nil {
				return err
			}
			if err := tx.Create(&model).Error; err != nil {
				if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
					return link.ErrDuplicateLink
				}
				return err
			}
			result = l
		}

		// A manual link on a rejected pair lifts the rejection (NULL or same type).
		return liftSuppressions(tx, l.SourceNoteID(), l.TargetNoteID(), l.LinkType().String())
	})
	if err != nil {
		return nil, false, err
	}
	return result, created, nil
}

// DeleteAndSuppress removes the link and records the pair rejection atomically.
func (r *LinkRepository) DeleteAndSuppress(ctx context.Context, l *link.Link, s *link.Suppression) error {
	return dbFromContext(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if s != nil {
			if err := upsertSuppression(tx, s); err != nil {
				return err
			}
		}
		return tx.Delete(&LinkModel{}, "id = ?", l.ID()).Error
	})
}

// SaveSuppression upserts a pair rejection.
func (r *LinkRepository) SaveSuppression(ctx context.Context, s *link.Suppression) error {
	return upsertSuppression(dbFromContext(ctx, r.db), s)
}

// FindSuppressionsForNotes returns all rejections that involve any of the notes.
func (r *LinkRepository) FindSuppressionsForNotes(ctx context.Context, noteIDs []uuid.UUID) ([]*link.Suppression, error) {
	if len(noteIDs) == 0 {
		return nil, nil
	}
	var models []LinkSuppressionModel
	err := dbFromContext(ctx, r.db).
		Where("note_a_id IN ? OR note_b_id IN ?", noteIDs, noteIDs).
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	result := make([]*link.Suppression, 0, len(models))
	for _, m := range models {
		result = append(result, link.ReconstructSuppression(m.ID, m.NoteAID, m.NoteBID, m.LinkType, m.CreatorID, m.CreatedAt))
	}
	return result, nil
}

// upsertSuppression inserts or refreshes a rejection row. The unique index on
// (note_a_id, note_b_id, COALESCE(link_type,”)) cannot be a GORM conflict
// target, so deduplication is done inside the transaction.
func upsertSuppression(tx *gorm.DB, s *link.Suppression) error {
	linkType := s.LinkType()
	var existing LinkSuppressionModel
	q := tx.Where("note_a_id = ? AND note_b_id = ?", s.NoteAID(), s.NoteBID())
	if linkType == nil {
		q = q.Where("link_type IS NULL")
	} else {
		q = q.Where("link_type = ?", *linkType)
	}
	err := q.First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return tx.Create(&LinkSuppressionModel{
		ID:        s.ID(),
		NoteAID:   s.NoteAID(),
		NoteBID:   s.NoteBID(),
		LinkType:  linkType,
		CreatorID: s.CreatorID(),
		CreatedAt: s.CreatedAt(),
	}).Error
}

// liftSuppressions removes rejections of a normalized pair contradicted by a
// manual link: NULL-type ("no link at all") and same-type rows.
func liftSuppressions(tx *gorm.DB, sourceID, targetID uuid.UUID, linkType string) error {
	noteA, noteB := link.NormalizePair(sourceID, targetID)
	return tx.
		Where("note_a_id = ? AND note_b_id = ? AND (link_type IS NULL OR link_type = ?)", noteA, noteB, linkType).
		Delete(&LinkSuppressionModel{}).Error
}

func (r *LinkRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.DeleteReturning(ctx, id)
	return err
}

// DeleteReturning is Delete plus the deleted row's endpoint/owner metadata —
// collected via RETURNING for the outbox decorator.
func (r *LinkRepository) DeleteReturning(ctx context.Context, id uuid.UUID) ([]LinkAffectedRow, error) {
	var links []LinkModel
	err := dbFromContext(ctx, r.db).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}, {Name: "source_note_id"}, {Name: "target_note_id"}, {Name: "creator_id"}}}).
		Where("id = ?", id).Delete(&links).Error
	return linkAffectedRows(links), err
}

func (r *LinkRepository) DeleteBySource(ctx context.Context, sourceID uuid.UUID) error {
	_, err := r.DeleteBySourceReturning(ctx, sourceID)
	return err
}

// DeleteBySourceReturning is DeleteBySource plus RETURNING metadata for every
// deleted link — the decorator emits one LinkDeleted per row.
func (r *LinkRepository) DeleteBySourceReturning(ctx context.Context, sourceID uuid.UUID) ([]LinkAffectedRow, error) {
	var links []LinkModel
	err := dbFromContext(ctx, r.db).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}, {Name: "source_note_id"}, {Name: "target_note_id"}, {Name: "creator_id"}}}).
		Where("source_note_id = ?", sourceID).Delete(&links).Error
	return linkAffectedRows(links), err
}

func linkAffectedRows(models []LinkModel) []LinkAffectedRow {
	rows := make([]LinkAffectedRow, 0, len(models))
	for i := range models {
		rows = append(rows, LinkAffectedRow{
			ID:           models[i].ID,
			SourceNoteID: models[i].SourceNoteID,
			TargetNoteID: models[i].TargetNoteID,
			CreatorID:    models[i].CreatorID,
		})
	}
	return rows
}

// FindBySourceType returns all links carrying the given source_type
// (e.g. "gamma"). Used by gamma-link regeneration to publish LinkDeleted
// events with correct source/target pairs.
func (r *LinkRepository) FindBySourceType(ctx context.Context, sourceType string) ([]*link.Link, error) {
	var models []LinkModel
	if err := dbFromContext(ctx, r.db).Where("source_type = ? AND deleted_at IS NULL", sourceType).Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]*link.Link, 0, len(models))
	for _, m := range models {
		l, err := toDomainLink(&m)
		if err != nil {
			continue
		}
		result = append(result, l)
	}
	return result, nil
}

// DeleteBySourceType removes all links with the given source_type (e.g. "gamma")
// and returns how many rows were deleted. Manual links have a different
// source_type and are not touched.
func (r *LinkRepository) DeleteBySourceType(ctx context.Context, sourceType string) (int64, error) {
	count, _, err := r.DeleteBySourceTypeReturning(ctx, sourceType)
	return count, err
}

// DeleteBySourceTypeReturning additionally reports every deleted link's
// endpoints and owner so the outbox decorator can emit LinkDeleted per row.
func (r *LinkRepository) DeleteBySourceTypeReturning(ctx context.Context, sourceType string) (int64, []LinkAffectedRow, error) {
	var links []LinkModel
	res := dbFromContext(ctx, r.db).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}, {Name: "source_note_id"}, {Name: "target_note_id"}, {Name: "creator_id"}}}).
		Where("source_type = ?", sourceType).Delete(&links)
	return res.RowsAffected, linkAffectedRows(links), res.Error
}

// CountBySourceType returns how many links carry the given source_type.
func (r *LinkRepository) CountBySourceType(ctx context.Context, sourceType string) (int64, error) {
	var count int64
	err := dbFromContext(ctx, r.db).Model(&LinkModel{}).Where("source_type = ? AND deleted_at IS NULL", sourceType).Count(&count).Error
	return count, err
}

// FindAllPaginated возвращает связи с пагинацией на уровне БД
// limit=0 означает "все записи"
func (r *LinkRepository) FindAllPaginated(ctx context.Context, limit, offset int) ([]*link.Link, int64, error) {
	var total int64

	// Считаем общее количество
	if err := dbFromContext(ctx, r.db).Model(&LinkModel{}).Where("deleted_at IS NULL").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Запрос с пагинацией
	query := dbFromContext(ctx, r.db)
	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	var models []LinkModel
	if err := query.Where("deleted_at IS NULL").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	return toDomainLinks(models), total, nil
}

// FindAll возвращает все связи без пагинации
// DEPRECATED: используйте FindAllPaginated для больших наборов данных
func (r *LinkRepository) FindAll(ctx context.Context) ([]*link.Link, error) {
	var models []LinkModel
	err := dbFromContext(ctx, r.db).Where("deleted_at IS NULL").Find(&models).Error
	if err != nil {
		return nil, err
	}
	return toDomainLinks(models), nil
}

// toGormLink преобразует доменную связь в GORM-модель
func toGormLink(l *link.Link) (LinkModel, error) {
	metadataJSON, err := json.Marshal(l.Metadata().Value())
	if err != nil {
		return LinkModel{}, err
	}
	return LinkModel{
		ID:               l.ID(),
		SourceNoteID:     l.SourceNoteID(),
		TargetNoteID:     l.TargetNoteID(),
		LinkType:         l.LinkType().String(),
		Weight:           l.Weight().Value(),
		Metadata:         datatypes.JSON(metadataJSON),
		SourceType:       l.SourceType().String(),
		CreatorID:        l.CreatorID(),
		CreatedAt:        l.CreatedAt(),
		UpdatedAt:        l.UpdatedAt(),
		LastWeightUpdate: l.LastWeightUpdate(),
	}, nil
}

// toDomainLink преобразует GORM-модель в доменную связь
func toDomainLink(m *LinkModel) (*link.Link, error) {
	linkType, err := link.NewLinkType(m.LinkType)
	if err != nil {
		return nil, err
	}
	weight, err := link.NewWeight(m.Weight)
	if err != nil {
		return nil, err
	}
	sourceType, err := link.NewSourceType(m.SourceType)
	if err != nil {
		// Fallback to default if source_type is missing (for backward compatibility)
		sourceType = link.DefaultSourceType()
	}
	var metadataMap map[string]interface{}
	if len(m.Metadata) > 0 {
		if err := json.Unmarshal(m.Metadata, &metadataMap); err != nil {
			return nil, err
		}
	}
	metadata, err := link.NewMetadata(metadataMap)
	if err != nil {
		return nil, err
	}
	return link.ReconstructLinkWithCreator(m.ID, m.SourceNoteID, m.TargetNoteID, linkType, weight, metadata, sourceType, m.CreatorID, m.CreatedAt, m.UpdatedAt, m.LastWeightUpdate), nil
}

// toDomainLinks преобразует список GORM-моделей в список доменных связей
func toDomainLinks(models []LinkModel) []*link.Link {
	result := make([]*link.Link, 0, len(models))
	for _, m := range models {
		l, err := toDomainLink(&m)
		if err != nil {
			continue
		}
		result = append(result, l)
	}
	return result
}
