package postgres

import (
	"context"
	"errors"

	"knowledge-graph/internal/domain/link"

	"github.com/google/uuid"
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
	err := r.db.WithContext(ctx).Where("id = ?", l.ID()).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		model := toGormLink(l)
		return r.db.WithContext(ctx).Create(&model).Error
	}
	if err != nil {
		return err
	}
	model := toGormLink(l)
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

func (r *LinkRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&LinkModel{}, "id = ?", id).Error
}

func (r *LinkRepository) DeleteBySource(ctx context.Context, sourceID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("source_note_id = ?", sourceID).Delete(&LinkModel{}).Error
}

// Преобразование домен → GORM
func toGormLink(l *link.Link) LinkModel {
	return LinkModel{
		ID:           l.ID(),
		SourceNoteID: l.SourceNoteID(),
		TargetNoteID: l.TargetNoteID(),
		LinkType:     l.LinkType().String(),
		Weight:       l.Weight().Value(),
		Metadata:     l.Metadata().Value(),
		CreatedAt:    l.CreatedAt(),
	}
}

// Преобразование GORM → домен (один)
func toDomainLink(m *LinkModel) (*link.Link, error) {
	linkType, err := link.NewLinkType(m.LinkType)
	if err != nil {
		return nil, err
	}
	weight, err := link.NewWeight(m.Weight)
	if err != nil {
		return nil, err
	}
	metadata, err := link.NewMetadata(m.Metadata)
	if err != nil {
		return nil, err
	}
	return link.ReconstructLink(m.ID, m.SourceNoteID, m.TargetNoteID, linkType, weight, metadata, m.CreatedAt), nil
}

// Преобразование GORM → домен (список)
func toDomainLinks(models []LinkModel) []*link.Link {
	result := make([]*link.Link, 0, len(models))
	for _, m := range models {
		l, err := toDomainLink(&m)
		if err != nil {
			// Логируем, но продолжаем (пропускаем битые записи)
			continue
		}
		result = append(result, l)
	}
	return result
}
