package postgres

import (
	"context"
	"errors"
	"time"

	"knowledge-graph/internal/domain/share"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ShareRepository implements share.Repository using GORM.
type ShareRepository struct {
	db *gorm.DB
}

// NewShareRepository creates a new share repository.
func NewShareRepository(db *gorm.DB) *ShareRepository {
	return &ShareRepository{db: db}
}

func toDomainNoteShare(m *NoteShareModel) *share.NoteShare {
	s, _ := share.NewNoteShare(m.ID, m.NoteID, m.SharedByUserID, m.SharedWithUserID, m.Permission, m.ExpiresAt)
	if s != nil && m.SharedWithUser.ID != uuid.Nil {
		s.SetSharedWithLogin(m.SharedWithUser.Login)
	}
	return s
}

func fromDomainNoteShare(s *share.NoteShare) *NoteShareModel {
	return &NoteShareModel{
		ID:               s.ID(),
		NoteID:           s.NoteID(),
		SharedByUserID:   s.SharedByUserID(),
		SharedWithUserID: s.SharedWithUserID(),
		Permission:       s.Permission(),
		CreatedAt:        s.CreatedAt(),
		ExpiresAt:        s.ExpiresAt(),
	}
}

func toDomainShareLink(m *ShareLinkModel) *share.ShareLink {
	l, _ := share.NewShareLink(m.ID, m.NoteID, m.SharedByUserID, m.Token, m.Permission, m.ExpiresAt, m.MaxUses, m.UsesCount)
	if l != nil && !m.IsActive {
		l.Deactivate()
	}
	return l
}

func fromDomainShareLink(l *share.ShareLink) *ShareLinkModel {
	return &ShareLinkModel{
		ID:             l.ID(),
		NoteID:         l.NoteID(),
		SharedByUserID: l.SharedByUserID(),
		Token:          l.Token(),
		Permission:     l.Permission(),
		CreatedAt:      l.CreatedAt(),
		ExpiresAt:      l.ExpiresAt(),
		MaxUses:        l.MaxUses(),
		UsesCount:      l.UsesCount(),
		IsActive:       l.IsActive(),
	}
}

// FindShareByNoteAndUser finds an existing share for the note and user.
func (r *ShareRepository) FindShareByNoteAndUser(ctx context.Context, noteID, sharedWithUserID uuid.UUID) (*share.NoteShare, error) {
	var model NoteShareModel
	err := r.db.WithContext(ctx).
		Where("note_id = ? AND shared_with_user_id = ?", noteID, sharedWithUserID).
		First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toDomainNoteShare(&model), nil
}

// CreateShare inserts a new note share.
func (r *ShareRepository) CreateShare(ctx context.Context, s *share.NoteShare) error {
	model := fromDomainNoteShare(s)
	model.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(model).Error
}

// UpdateShare saves changes to an existing note share.
func (r *ShareRepository) UpdateShare(ctx context.Context, s *share.NoteShare) error {
	model := fromDomainNoteShare(s)
	return r.db.WithContext(ctx).Save(model).Error
}

// ListSharesByNote returns all direct shares for a note.
func (r *ShareRepository) ListSharesByNote(ctx context.Context, noteID uuid.UUID) ([]share.NoteShare, error) {
	var models []NoteShareModel
	err := r.db.WithContext(ctx).Preload("SharedWithUser").Where("note_id = ?", noteID).Find(&models).Error
	if err != nil {
		return nil, err
	}

	shares := make([]share.NoteShare, 0, len(models))
	for _, m := range models {
		s := toDomainNoteShare(&m)
		if s != nil {
			shares = append(shares, *s)
		}
	}
	return shares, nil
}

// RevokeShare removes a note share if it belongs to the shared-by user.
func (r *ShareRepository) RevokeShare(ctx context.Context, noteID, shareID, sharedByUserID uuid.UUID) (bool, error) {
	result := r.db.WithContext(ctx).
		Where("id = ? AND note_id = ? AND shared_by_user_id = ?", shareID, noteID, sharedByUserID).
		Delete(&NoteShareModel{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// CreateShareLink inserts a new public share link.
func (r *ShareRepository) CreateShareLink(ctx context.Context, l *share.ShareLink) error {
	model := fromDomainShareLink(l)
	model.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(model).Error
}

// FindActiveShareLinkByToken finds an active share link by its token.
func (r *ShareRepository) FindActiveShareLinkByToken(ctx context.Context, token string) (*share.ShareLink, error) {
	var model ShareLinkModel
	err := r.db.WithContext(ctx).
		Where("token = ? AND is_active = ?", token, true).
		First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toDomainShareLink(&model), nil
}

// RevokeShareLink marks a share link as inactive if owned by the user.
func (r *ShareRepository) RevokeShareLink(ctx context.Context, linkID, sharedByUserID uuid.UUID) (bool, error) {
	result := r.db.WithContext(ctx).
		Model(&ShareLinkModel{}).
		Where("id = ? AND shared_by_user_id = ?", linkID, sharedByUserID).
		Update("is_active", false)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// ListShareLinksByNote returns all active share links for a note.
func (r *ShareRepository) ListShareLinksByNote(ctx context.Context, noteID uuid.UUID) ([]share.ShareLink, error) {
	var models []ShareLinkModel
	err := r.db.WithContext(ctx).
		Where("note_id = ? AND is_active = ?", noteID, true).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	links := make([]share.ShareLink, 0, len(models))
	for _, m := range models {
		l := toDomainShareLink(&m)
		if l != nil {
			links = append(links, *l)
		}
	}
	return links, nil
}

// IncrementShareLinkUsage increments the usage counter for a share link.
func (r *ShareRepository) IncrementShareLinkUsage(ctx context.Context, linkID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&ShareLinkModel{}).
		Where("id = ?", linkID).
		UpdateColumn("uses_count", gorm.Expr("uses_count + 1")).Error
}
