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

// GetKeywordsWithWeights возвращает ключевые слова с весами для одной заметки
func (r *KeywordRepository) GetKeywordsWithWeights(ctx context.Context, noteID uuid.UUID) (map[string]float64, error) {
	var keywords []NoteKeywordModel
	if err := r.db.WithContext(ctx).
		Where("note_id = ?", noteID).
		Find(&keywords).Error; err != nil {
		return nil, err
	}

	result := make(map[string]float64, len(keywords))
	for _, kw := range keywords {
		result[kw.Keyword] = kw.Weight
	}
	return result, nil
}

// GetKeywordsBatchWithWeights возвращает ключевые слова с весами для нескольких заметок
func (r *KeywordRepository) GetKeywordsBatchWithWeights(ctx context.Context, noteIDs []uuid.UUID) (map[uuid.UUID]map[string]float64, error) {
	if len(noteIDs) == 0 {
		return make(map[uuid.UUID]map[string]float64), nil
	}

	var keywords []NoteKeywordModel
	if err := r.db.WithContext(ctx).
		Where("note_id IN ?", noteIDs).
		Find(&keywords).Error; err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]map[string]float64)
	for _, kw := range keywords {
		if result[kw.NoteID] == nil {
			result[kw.NoteID] = make(map[string]float64)
		}
		result[kw.NoteID][kw.Keyword] = kw.Weight
	}
	return result, nil
}

// FindNoteIDsMissingExtractor returns IDs of notes that have non-empty
// content but no note_keywords row written by the given extractor.
// Notes for which the extractor honestly returned an empty list will keep
// matching — acceptable for a one-shot recompute (see NLP-2 spec).
func (r *KeywordRepository) FindNoteIDsMissingExtractor(ctx context.Context, extractor string) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Raw(`
		SELECT n.id FROM notes n
		WHERE n.deleted_at IS NULL AND btrim(n.content) <> ''
		  AND NOT EXISTS (
			SELECT 1 FROM note_keywords k
			WHERE k.note_id = n.id AND k.extractor = ?
		)`, extractor).Scan(&ids).Error
	return ids, err
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
