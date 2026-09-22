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
)

type LinkRepository struct {
	db *gorm.DB
}

func NewLinkRepository(db *gorm.DB) *LinkRepository {
	return &LinkRepository{db: db}
}

func (r *LinkRepository) Save(ctx context.Context, l *link.Link) error {
	var existing LinkModel
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", l.ID()).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		model, err := toGormLink(l)
		if err != nil {
			log.Printf("[LinkRepository.Save] toGormLink failed: %v", err)
			return err
		}
		if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
			log.Printf("[LinkRepository.Save] Create failed: id=%s source=%s target=%s error=%v",
				model.ID, model.SourceNoteID, model.TargetNoteID, err)
			// Проверяем на нарушение уникального ограничения (PostgreSQL код 23505)
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
				return link.ErrDuplicateLink
			}
			return err
		}
		return nil
	}
	if err != nil {
		return err
	}
	return r.updateModel(ctx, &existing, l)
}

func (r *LinkRepository) Update(ctx context.Context, l *link.Link) error {
	var existing LinkModel
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", l.ID()).First(&existing).Error
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
	return r.db.WithContext(ctx).Model(existing).Updates(model).Error
}

func (r *LinkRepository) FindByID(ctx context.Context, id uuid.UUID) (*link.Link, error) {
	var model LinkModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
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
	err := r.db.WithContext(ctx).Where("source_note_id = ?", sourceID).Find(&models).Error
	if err != nil {
		return nil, err
	}
	return toDomainLinks(models), nil
}

func (r *LinkRepository) FindByTarget(ctx context.Context, targetID uuid.UUID) ([]*link.Link, error) {
	var models []LinkModel
	err := r.db.WithContext(ctx).Where("target_note_id = ?", targetID).Find(&models).Error
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
	err := r.db.WithContext(ctx).Where("source_note_id IN ?", sourceIDs).Find(&models).Error
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
	err := r.db.WithContext(ctx).Where("target_note_id IN ?", targetIDs).Find(&models).Error
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
	err := r.db.WithContext(ctx).
		Where("source_note_id = ? AND target_note_id = ? AND deleted_at IS NULL", sourceID, targetID).
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	return toDomainLinks(models), nil
}

// SaveUserLink applies the manual-create rules atomically (LINKS-2):
//   - a gamma row of the same type on the same directed pair is promoted in
//     place (created=false) — the row keeps its id and created_at;
//   - gamma rows of other types on the pair are removed and their provenance
//     moves into the new row's metadata.gamma;
//   - a manual row of the same type yields ErrDuplicateLink;
//   - rejections (link_suppressions) for the normalized pair with NULL or
//     matching link_type are lifted in the same transaction.
func (r *LinkRepository) SaveUserLink(ctx context.Context, l *link.Link) (*link.Link, bool, error) {
	var result *link.Link
	created := true

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var models []LinkModel
		if err := tx.
			Where("source_note_id = ? AND target_note_id = ? AND deleted_at IS NULL", l.SourceNoteID(), l.TargetNoteID()).
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
				sameType = &models[i]
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
			promoted, err := toGormLink(gamma)
			if err != nil {
				return err
			}
			if err := tx.Model(&LinkModel{}).Where("id = ?", sameType.ID).Updates(map[string]interface{}{
				"link_type":   promoted.LinkType,
				"weight":      promoted.Weight,
				"metadata":    promoted.Metadata,
				"source_type": promoted.SourceType,
				"creator_id":  promoted.CreatorID,
				"updated_at":  promoted.UpdatedAt,
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
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
	return upsertSuppression(r.db.WithContext(ctx), s)
}

// FindSuppressionsForNotes returns all rejections that involve any of the notes.
func (r *LinkRepository) FindSuppressionsForNotes(ctx context.Context, noteIDs []uuid.UUID) ([]*link.Suppression, error) {
	if len(noteIDs) == 0 {
		return nil, nil
	}
	var models []LinkSuppressionModel
	err := r.db.WithContext(ctx).
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
	return r.db.WithContext(ctx).Delete(&LinkModel{}, "id = ?", id).Error
}

func (r *LinkRepository) DeleteBySource(ctx context.Context, sourceID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("source_note_id = ?", sourceID).Delete(&LinkModel{}).Error
}

// FindBySourceType returns all links carrying the given source_type
// (e.g. "gamma"). Used by gamma-link regeneration to publish LinkDeleted
// events with correct source/target pairs.
func (r *LinkRepository) FindBySourceType(ctx context.Context, sourceType string) ([]*link.Link, error) {
	var models []LinkModel
	if err := r.db.WithContext(ctx).Where("source_type = ?", sourceType).Find(&models).Error; err != nil {
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
	res := r.db.WithContext(ctx).Where("source_type = ?", sourceType).Delete(&LinkModel{})
	return res.RowsAffected, res.Error
}

// CountBySourceType returns how many links carry the given source_type.
func (r *LinkRepository) CountBySourceType(ctx context.Context, sourceType string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&LinkModel{}).Where("source_type = ?", sourceType).Count(&count).Error
	return count, err
}

// FindAllPaginated возвращает связи с пагинацией на уровне БД
// limit=0 означает "все записи"
func (r *LinkRepository) FindAllPaginated(ctx context.Context, limit, offset int) ([]*link.Link, int64, error) {
	var total int64

	// Считаем общее количество
	if err := r.db.WithContext(ctx).Model(&LinkModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Запрос с пагинацией
	query := r.db.WithContext(ctx)
	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	var models []LinkModel
	if err := query.Find(&models).Error; err != nil {
		return nil, 0, err
	}

	return toDomainLinks(models), total, nil
}

// FindAll возвращает все связи без пагинации
// DEPRECATED: используйте FindAllPaginated для больших наборов данных
func (r *LinkRepository) FindAll(ctx context.Context) ([]*link.Link, error) {
	var models []LinkModel
	err := r.db.WithContext(ctx).Find(&models).Error
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
