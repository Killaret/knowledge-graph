package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TagRepository — репозиторий для работы с тегами
type TagRepository struct {
	db *gorm.DB
}

// NewTagRepository создает новый репозиторий
func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

// Create сохраняет новый тег
func (r *TagRepository) Create(ctx context.Context, tag *TagModel) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

// FindByID ищет тег по ID
func (r *TagRepository) FindByID(ctx context.Context, id uuid.UUID) (*TagModel, error) {
	var tag TagModel
	err := r.db.WithContext(ctx).First(&tag, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// FindByName ищет тег по имени
func (r *TagRepository) FindByName(ctx context.Context, name string) (*TagModel, error) {
	var tag TagModel
	err := r.db.WithContext(ctx).First(&tag, "name = ?", name).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// Update обновляет тег
func (r *TagRepository) Update(ctx context.Context, tag *TagModel) error {
	return r.db.WithContext(ctx).Save(tag).Error
}

// Delete удаляет тег
func (r *TagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&TagModel{}, "id = ?", id).Error
}

// AddTagToNote добавляет тег к заметке
func (r *TagRepository) AddTagToNote(ctx context.Context, noteID, tagID uuid.UUID) error {
	noteTag := &NoteTagModel{
		NoteID: noteID,
		TagID:  tagID,
	}
	return r.db.WithContext(ctx).Create(noteTag).Error
}

// RemoveTagFromNote удаляет тег от заметки
func (r *TagRepository) RemoveTagFromNote(ctx context.Context, noteID, tagID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("note_id = ? AND tag_id = ?", noteID, tagID).
		Delete(&NoteTagModel{}).Error
}

// GetTagsForNote возвращает все теги заметки
func (r *TagRepository) GetTagsForNote(ctx context.Context, noteID uuid.UUID) ([]*TagModel, error) {
	var tags []*TagModel
	err := r.db.WithContext(ctx).
		Table("tags").
		Select("tags.*").
		Joins("JOIN note_tags ON note_tags.tag_id = tags.id").
		Where("note_tags.note_id = ?", noteID).
		Find(&tags).Error
	return tags, err
}

// GetNotesForTag возвращает все заметки с тегом
func (r *TagRepository) GetNotesForTag(ctx context.Context, tagID uuid.UUID) ([]*NoteModel, error) {
	var notes []*NoteModel
	err := r.db.WithContext(ctx).
		Table("notes").
		Select("notes.*").
		Joins("JOIN note_tags ON note_tags.note_id = notes.id").
		Where("note_tags.tag_id = ?", tagID).
		Find(&notes).Error
	return notes, err
}

// Exists проверяет существование тега
func (r *TagRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&TagModel{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// FindAll возвращает все теги
func (r *TagRepository) FindAll(ctx context.Context) ([]*TagModel, error) {
	var tags []*TagModel
	err := r.db.WithContext(ctx).Find(&tags).Error
	return tags, err
}
