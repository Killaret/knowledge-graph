package postgres

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type KeywordRepository struct {
	db *gorm.DB
}

func NewKeywordRepository(db *gorm.DB) *KeywordRepository {
	return &KeywordRepository{db: db}
}

// SaveAll saves keywords for a note (deletes old ones, inserts new ones)
func (r *KeywordRepository) SaveAll(ctx context.Context, noteID uuid.UUID, keywords []NoteKeywordModel) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete the old ones
		if err := tx.Where("note_id = ?", noteID).Delete(&NoteKeywordModel{}).Error; err != nil {
			return err
		}
		// Insert the new ones
		if len(keywords) > 0 {
			return tx.Create(&keywords).Error
		}
		return nil
	})
}

// DeleteAll removes all keywords of a note
func (r *KeywordRepository) DeleteAll(ctx context.Context, noteID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("note_id = ?", noteID).Delete(&NoteKeywordModel{}).Error
}
