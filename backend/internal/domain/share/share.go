package share

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidPermission = errors.New("permission must be read or write")

// NoteShare represents a direct user-to-user note share.
type NoteShare struct {
	id               uuid.UUID
	noteID           uuid.UUID
	sharedByUserID   uuid.UUID
	sharedWithUserID uuid.UUID
	sharedWithLogin  string
	permission       string
	createdAt        time.Time
	expiresAt        *time.Time
}

// NewNoteShare creates a validated note share.
func NewNoteShare(id, noteID, sharedByUserID, sharedWithUserID uuid.UUID, permission string, expiresAt *time.Time) (*NoteShare, error) {
	if permission != "read" && permission != "write" {
		return nil, ErrInvalidPermission
	}
	return &NoteShare{
		id:               id,
		noteID:           noteID,
		sharedByUserID:   sharedByUserID,
		sharedWithUserID: sharedWithUserID,
		permission:       permission,
		createdAt:        time.Now(),
		expiresAt:        expiresAt,
	}, nil
}

func (s *NoteShare) ID() uuid.UUID               { return s.id }
func (s *NoteShare) NoteID() uuid.UUID           { return s.noteID }
func (s *NoteShare) SharedByUserID() uuid.UUID   { return s.sharedByUserID }
func (s *NoteShare) SharedWithUserID() uuid.UUID { return s.sharedWithUserID }
func (s *NoteShare) SharedWithLogin() string     { return s.sharedWithLogin }
func (s *NoteShare) Permission() string          { return s.permission }
func (s *NoteShare) CreatedAt() time.Time        { return s.createdAt }
func (s *NoteShare) ExpiresAt() *time.Time       { return s.expiresAt }

// SetSharedWithLogin is a read-only projection helper used by repositories.
func (s *NoteShare) SetSharedWithLogin(login string) {
	s.sharedWithLogin = login
}

// UpdatePermission changes the access level.
func (s *NoteShare) UpdatePermission(permission string) error {
	if permission != "read" && permission != "write" {
		return ErrInvalidPermission
	}
	s.permission = permission
	return nil
}

// ShareLink represents a public or token-based share link.
type ShareLink struct {
	id             uuid.UUID
	noteID         uuid.UUID
	sharedByUserID uuid.UUID
	token          string
	permission     string
	expiresAt      *time.Time
	maxUses        *int
	usesCount      int
	isActive       bool
	createdAt      time.Time
}

// NewShareLink creates a validated share link.
func NewShareLink(id, noteID, sharedByUserID uuid.UUID, token, permission string, expiresAt *time.Time, maxUses *int, usesCount int) (*ShareLink, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}
	if permission != "read" && permission != "write" {
		return nil, ErrInvalidPermission
	}
	return &ShareLink{
		id:             id,
		noteID:         noteID,
		sharedByUserID: sharedByUserID,
		token:          token,
		permission:     permission,
		expiresAt:      expiresAt,
		maxUses:        maxUses,
		usesCount:      usesCount,
		isActive:       true,
		createdAt:      time.Now(),
	}, nil
}

func (l *ShareLink) ID() uuid.UUID             { return l.id }
func (l *ShareLink) NoteID() uuid.UUID         { return l.noteID }
func (l *ShareLink) SharedByUserID() uuid.UUID { return l.sharedByUserID }
func (l *ShareLink) Token() string             { return l.token }
func (l *ShareLink) Permission() string        { return l.permission }
func (l *ShareLink) ExpiresAt() *time.Time     { return l.expiresAt }
func (l *ShareLink) MaxUses() *int             { return l.maxUses }
func (l *ShareLink) UsesCount() int            { return l.usesCount }
func (l *ShareLink) IsActive() bool            { return l.isActive }
func (l *ShareLink) CreatedAt() time.Time      { return l.createdAt }

// IncrementUsage increases the usage counter.
func (l *ShareLink) IncrementUsage() {
	l.usesCount++
}

// Deactivate marks the link as revoked.
func (l *ShareLink) Deactivate() {
	l.isActive = false
}
