package postgres

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// QualityStatsRepository supplies the stage-1 readiness counters for
// NOTE-QUALITY-1: keyword count, link count (manual + auto), embedding flag.
type QualityStatsRepository struct {
	db *gorm.DB
}

func NewQualityStatsRepository(db *gorm.DB) *QualityStatsRepository {
	return &QualityStatsRepository{db: db}
}

func (r *QualityStatsRepository) KeywordCount(ctx context.Context, noteID uuid.UUID) (int, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Model(&NoteKeywordModel{}).
		Where("note_id = ?", noteID).
		Count(&n).Error
	return int(n), err
}

func (r *QualityStatsRepository) LinkCount(ctx context.Context, noteID uuid.UUID) (int, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Model(&LinkModel{}).
		Where("source_id = ? OR target_id = ?", noteID, noteID).
		Count(&n).Error
	return int(n), err
}

func (r *QualityStatsRepository) HasEmbedding(ctx context.Context, noteID uuid.UUID) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Model(&NoteEmbeddingModel{}).
		Where("note_id = ?", noteID).
		Limit(1).
		Count(&n).Error
	return n > 0, err
}
