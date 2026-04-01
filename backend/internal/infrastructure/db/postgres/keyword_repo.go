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

// SaveAll сохраняет ключевые слова для заметки (удаляет старые, вставляет новые)
func (r *KeywordRepository) SaveAll(ctx context.Context, noteID uuid.UUID, keywords []NoteKeywordModel) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Удаляем старые
		if err := tx.Where("note_id = ?", noteID).Delete(&NoteKeywordModel{}).Error; err != nil {
			return err
		}
		// Вставляем новые
		if len(keywords) > 0 {
			return tx.Create(&keywords).Error
		}
		return nil
	})
}

// DeleteAll удаляет все ключевые слова заметки
func (r *KeywordRepository) DeleteAll(ctx context.Context, noteID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("note_id = ?", noteID).Delete(&NoteKeywordModel{}).Error
}
