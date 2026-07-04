package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TagRepository is the repository for working with tags
type TagRepository struct {
	db *gorm.DB
}

// NewTagRepository creates a new repository
func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

// Create saves a new tag
func (r *TagRepository) Create(ctx context.Context, tag *TagModel) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

// FindByID looks up a tag by ID
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

// FindByName looks up a tag by name
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

// Update updates a tag
func (r *TagRepository) Update(ctx context.Context, tag *TagModel) error {
	return r.db.WithContext(ctx).Save(tag).Error
}

// Delete removes a tag
func (r *TagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&TagModel{}, "id = ?", id).Error
}

// AddTagToNote adds a tag to a note
func (r *TagRepository) AddTagToNote(ctx context.Context, noteID, tagID uuid.UUID) error {
	noteTag := &NoteTagModel{
		NoteID: noteID,
		TagID:  tagID,
	}
	return r.db.WithContext(ctx).Create(noteTag).Error
}

// RemoveTagFromNote removes a tag from a note
func (r *TagRepository) RemoveTagFromNote(ctx context.Context, noteID, tagID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("note_id = ? AND tag_id = ?", noteID, tagID).
		Delete(&NoteTagModel{}).Error
}

// GetTagsForNote returns all tags of a note
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

// GetNotesForTag returns all notes that have the tag
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

// Exists checks whether a tag exists
func (r *TagRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&TagModel{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// FindAll returns all tags
func (r *TagRepository) FindAll(ctx context.Context) ([]*TagModel, error) {
	var tags []*TagModel
	err := r.db.WithContext(ctx).Find(&tags).Error
	return tags, err
}

// IsTagAssignedToNote checks whether a tag is assigned to a note
func (r *TagRepository) IsTagAssignedToNote(ctx context.Context, noteID, tagID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&NoteTagModel{}).
		Where("note_id = ? AND tag_id = ?", noteID, tagID).
		Count(&count).Error
	return count > 0, err
}

// GetTagsByNoteID returns a note's tags (alias for GetTagsForNote)
func (r *TagRepository) GetTagsByNoteID(ctx context.Context, noteID uuid.UUID) ([]*TagModel, error) {
	return r.GetTagsForNote(ctx, noteID)
}
