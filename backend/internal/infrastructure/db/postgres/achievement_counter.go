package postgres

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AchievementCounter provides count queries used by the achievement engine.
type AchievementCounter struct {
	db *gorm.DB
}

// NewAchievementCounter creates a new achievement counter.
func NewAchievementCounter(db *gorm.DB) *AchievementCounter {
	return &AchievementCounter{db: db}
}

// CountNotes returns the number of non-deleted notes created by userID.
// If noteType is non-empty, the count is filtered by note type.
func (c *AchievementCounter) CountNotes(ctx context.Context, userID uuid.UUID, noteType string) (int64, error) {
	var count int64
	query := c.db.WithContext(ctx).
		Model(&NoteModel{}).
		Where("creator_id = ? AND deleted_at IS NULL", userID)

	if noteType != "" {
		query = query.Where("type = ?", noteType)
	}

	err := query.Count(&count).Error
	return count, err
}

// CountLinks returns the number of non-deleted links created by userID.
func (c *AchievementCounter) CountLinks(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := c.db.WithContext(ctx).
		Model(&LinkModel{}).
		Where("creator_id = ? AND deleted_at IS NULL", userID).
		Count(&count).Error
	return count, err
}

// CountShares returns the number of note shares initiated by userID.
func (c *AchievementCounter) CountShares(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := c.db.WithContext(ctx).
		Model(&NoteShareModel{}).
		Where("shared_by_user_id = ?", userID).
		Count(&count).Error
	return count, err
}
