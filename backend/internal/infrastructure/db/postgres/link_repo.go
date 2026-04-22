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

// ErrDuplicateLink возвращается при попытке создать дубликат связи
var ErrDuplicateLink = errors.New("link of this type already exists between these notes")

type LinkRepository struct {
	db *gorm.DB
}

func NewLinkRepository(db *gorm.DB) *LinkRepository {
	return &LinkRepository{db: db}
}

func (r *LinkRepository) Save(ctx context.Context, l *link.Link) error {
	var existing LinkModel
	err := r.db.WithContext(ctx).Where("id = ?", l.ID()).First(&existing).Error
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
				return ErrDuplicateLink
			}
			return err
		}
		return nil
	}
	if err != nil {
		return err
	}
	model, err := toGormLink(l)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&existing).Updates(model).Error
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
	err := r.db.WithContext(ctx).Where("source_id = ?", sourceID).Find(&models).Error
	if err != nil {
		return nil, err
	}
	return toDomainLinks(models), nil
}

func (r *LinkRepository) FindByTarget(ctx context.Context, targetID uuid.UUID) ([]*link.Link, error) {
	var models []LinkModel
	err := r.db.WithContext(ctx).Where("target_id = ?", targetID).Find(&models).Error
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
	err := r.db.WithContext(ctx).Where("source_id IN ?", sourceIDs).Find(&models).Error
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
	err := r.db.WithContext(ctx).Where("target_id IN ?", targetIDs).Find(&models).Error
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

func (r *LinkRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&LinkModel{}, "id = ?", id).Error
}

func (r *LinkRepository) DeleteBySource(ctx context.Context, sourceID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("source_id = ?", sourceID).Delete(&LinkModel{}).Error
}

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
		ID:           l.ID(),
		SourceNoteID: l.SourceNoteID(),
		TargetNoteID: l.TargetNoteID(),
		LinkType:     l.LinkType().String(),
		Weight:       l.Weight().Value(),
		Metadata:     datatypes.JSON(metadataJSON),
		CreatedAt:    l.CreatedAt(),
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
	return link.ReconstructLink(m.ID, m.SourceNoteID, m.TargetNoteID, linkType, weight, metadata, m.CreatedAt), nil
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
