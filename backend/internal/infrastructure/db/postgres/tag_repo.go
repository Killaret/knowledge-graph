package postgres

import (
	"context"
	"errors"

	"knowledge-graph/internal/domain/tag"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TagRepository implements tag.Repository using PostgreSQL/GORM.
type TagRepository struct {
	db *gorm.DB
}

// NewTagRepository creates a new tag repository.
func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

// Create persists a new tag.
func (r *TagRepository) Create(ctx context.Context, t *tag.Tag) error {
	model := tagModelFromDomain(t)
	return r.db.WithContext(ctx).Create(model).Error
}

// FindByID retrieves a tag by ID.
func (r *TagRepository) FindByID(ctx context.Context, id uuid.UUID) (*tag.Tag, error) {
	var model TagModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return tag.Reconstruct(model.ID, model.Name, model.CreatedAt)
}

// FindByName retrieves a tag by name.
func (r *TagRepository) FindByName(ctx context.Context, name string) (*tag.Tag, error) {
	var model TagModel
	err := r.db.WithContext(ctx).First(&model, "name = ?", name).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return tag.Reconstruct(model.ID, model.Name, model.CreatedAt)
}

// FindAll returns all tags.
func (r *TagRepository) FindAll(ctx context.Context) ([]*tag.Tag, error) {
	var models []*TagModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}

	tags := make([]*tag.Tag, 0, len(models))
	for _, m := range models {
		t, err := tag.Reconstruct(m.ID, m.Name, m.CreatedAt)
		if err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}

// Update persists tag changes.
func (r *TagRepository) Update(ctx context.Context, t *tag.Tag) error {
	model := tagModelFromDomain(t)
	return r.db.WithContext(ctx).Save(model).Error
}

// Delete removes a tag by ID.
func (r *TagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&TagModel{}, "id = ?", id).Error
}

// AddTagToNote links a tag to a note.
func (r *TagRepository) AddTagToNote(ctx context.Context, noteID, tagID uuid.UUID) error {
	noteTag := &NoteTagModel{
		NoteID: noteID,
		TagID:  tagID,
	}
	return r.db.WithContext(ctx).Create(noteTag).Error
}

// RemoveTagFromNote removes a tag link from a note.
func (r *TagRepository) RemoveTagFromNote(ctx context.Context, noteID, tagID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("note_id = ? AND tag_id = ?", noteID, tagID).
		Delete(&NoteTagModel{}).Error
}

// GetTagsByNoteID returns all tags linked to a note (domain model).
func (r *TagRepository) GetTagsByNoteID(ctx context.Context, noteID uuid.UUID) ([]*tag.Tag, error) {
	models, err := r.getTagsForNoteModels(ctx, noteID)
	if err != nil {
		return nil, err
	}

	tags := make([]*tag.Tag, 0, len(models))
	for _, m := range models {
		t, err := tag.Reconstruct(m.ID, m.Name, m.CreatedAt)
		if err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}

// GetTagsForNote returns raw persistence models for the note.
// Kept for backward compatibility with integration tests.
func (r *TagRepository) GetTagsForNote(ctx context.Context, noteID uuid.UUID) ([]*TagModel, error) {
	return r.getTagsForNoteModels(ctx, noteID)
}

func (r *TagRepository) getTagsForNoteModels(ctx context.Context, noteID uuid.UUID) ([]*TagModel, error) {
	var tags []*TagModel
	err := r.db.WithContext(ctx).
		Table("tags").
		Select("tags.*").
		Joins("JOIN note_tags ON note_tags.tag_id = tags.id").
		Where("note_tags.note_id = ?", noteID).
		Find(&tags).Error
	return tags, err
}

// GetNotesForTag returns all notes linked to a tag.
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

// Exists checks whether a tag exists.
func (r *TagRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&TagModel{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// IsTagAssignedToNote reports whether a tag is linked to a note.
func (r *TagRepository) IsTagAssignedToNote(ctx context.Context, noteID, tagID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&NoteTagModel{}).
		Where("note_id = ? AND tag_id = ?", noteID, tagID).
		Count(&count).Error
	return count > 0, err
}

func tagModelFromDomain(t *tag.Tag) *TagModel {
	return &TagModel{
		ID:        t.ID(),
		Name:      t.Name(),
		CreatedAt: t.CreatedAt(),
	}
}
