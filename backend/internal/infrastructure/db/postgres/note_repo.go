package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type NoteRepository struct {
	db *gorm.DB
}

func NewNoteRepository(db *gorm.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

func (r *NoteRepository) Save(ctx context.Context, n *note.Note) error {
	var existing NoteModel
	err := r.db.WithContext(ctx).Where("id = ?", n.ID()).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		model, err := toGormNote(n)
		if err != nil {
			return err
		}
		return r.db.WithContext(ctx).Create(&model).Error
	}
	if err != nil {
		return err
	}
	model, err := toGormNote(n)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&existing).Updates(model).Error
}

func (r *NoteRepository) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	var model NoteModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toDomainNote(&model)
}

func (r *NoteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&NoteModel{}, "id = ?", id).Error
}

// List возвращает заметки с пагинацией
func (r *NoteRepository) List(ctx context.Context, limit, offset int) ([]*note.Note, int64, error) {
	var models []NoteModel
	var total int64

	db := r.db.WithContext(ctx).Model(&NoteModel{})

	// Count total
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	return toDomainNotes(models), total, nil
}

// Search performs multilingual full-text search on notes (Russian + English)
func (r *NoteRepository) Search(ctx context.Context, query string, limit, offset int) ([]*note.Note, int64, error) {
	var models []NoteModel
	var total int64

	db := r.db.WithContext(ctx).Model(&NoteModel{})

	if query != "" {
		// Multilingual search: search in both Russian and English
		// Russian search with stemming
		// English search with simple configuration (no stemming)
		db = db.Where(`
			search_vector @@ plainto_tsquery('russian', ?) OR 
			search_vector @@ plainto_tsquery('simple', ?)
		`, query, query)

		// Ranking: combine both Russian and English rankings
		// Use COALESCE to handle null rankings and give preference to matches
		db = db.Order(`
			COALESCE(ts_rank(search_vector, plainto_tsquery('russian', '" + query + "')), 0) +
			COALESCE(ts_rank(search_vector, plainto_tsquery('simple', '" + query + "')), 0) 
			DESC
		`)
	} else {
		// If query is empty, return all notes ordered by creation date
		db = db.Order("created_at DESC")
	}

	// Count total matching records first
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get the paginated results
	err := db.Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	return toDomainNotes(models), total, nil
}

// toGormNote преобразует доменную заметку в GORM-модель
func toGormNote(n *note.Note) (NoteModel, error) {
	metadataJSON, err := json.Marshal(n.Metadata().Value())
	if err != nil {
		return NoteModel{}, err
	}
	return NoteModel{
		ID:        n.ID(),
		Title:     n.Title().String(),
		Content:   n.Content().String(),
		Metadata:  datatypes.JSON(metadataJSON),
		CreatedAt: n.CreatedAt(),
		UpdatedAt: n.UpdatedAt(),
	}, nil
}

// toDomainNote преобразует GORM-модель в доменную заметку
func toDomainNote(m *NoteModel) (*note.Note, error) {
	title, err := note.NewTitle(m.Title)
	if err != nil {
		return nil, err
	}
	content, err := note.NewContent(m.Content)
	if err != nil {
		return nil, err
	}
	var metadataMap map[string]interface{}
	if len(m.Metadata) > 0 {
		if err := json.Unmarshal(m.Metadata, &metadataMap); err != nil {
			return nil, err
		}
	}
	metadata, err := note.NewMetadata(metadataMap)
	if err != nil {
		return nil, err
	}
	return note.ReconstructNote(m.ID, title, content, metadata, m.CreatedAt, m.UpdatedAt), nil
}

// toDomainNotes преобразует список GORM-моделей в список доменных сущностей
func toDomainNotes(models []NoteModel) []*note.Note {
	result := make([]*note.Note, 0, len(models))
	for _, m := range models {
		n, err := toDomainNote(&m)
		if err != nil {
			// Логируем ошибку, но продолжаем (пропускаем битые записи)
			continue
		}
		result = append(result, n)
	}
	return result
}
