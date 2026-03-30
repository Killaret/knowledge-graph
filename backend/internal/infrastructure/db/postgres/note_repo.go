package postgres

import (
	"context"
	"errors"

	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
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
		// Создание
		model := toGormNote(n)
		return r.db.WithContext(ctx).Create(&model).Error
	}
	if err != nil {
		return err
	}
	// Обновление
	model := toGormNote(n)
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

// Преобразование домен → GORM
func toGormNote(n *note.Note) NoteModel {
	return NoteModel{
		ID:        n.ID(),
		Title:     n.Title().String(),
		Content:   n.Content().String(),
		Metadata:  n.Metadata().Value(),
		CreatedAt: n.CreatedAt(),
		UpdatedAt: n.UpdatedAt(),
	}
}

// Преобразование GORM → домен
func toDomainNote(m *NoteModel) (*note.Note, error) {
	title, err := note.NewTitle(m.Title)
	if err != nil {
		return nil, err
	}
	content, err := note.NewContent(m.Content)
	if err != nil {
		return nil, err
	}
	metadata, err := note.NewMetadata(m.Metadata)
	if err != nil {
		return nil, err
	}
	return note.ReconstructNote(m.ID, title, content, metadata, m.CreatedAt, m.UpdatedAt), nil
}
