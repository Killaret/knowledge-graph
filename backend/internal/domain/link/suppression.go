package link

import (
	"bytes"
	"context"
	"time"

	"github.com/google/uuid"
)

// Suppression is a recorded human rejection "these notes are not related".
// The pair is normalized (smaller uuid first): a rejection applies to the
// pair regardless of the direction the gamma link was proposed in.
// LinkType nil means "no link between them at all".
type Suppression struct {
	id        uuid.UUID
	noteAID   uuid.UUID
	noteBID   uuid.UUID
	linkType  *string
	creatorID *uuid.UUID
	createdAt time.Time
}

// NormalizePair returns the pair ordered so the smaller uuid comes first.
func NormalizePair(a, b uuid.UUID) (uuid.UUID, uuid.UUID) {
	if bytes.Compare(a[:], b[:]) <= 0 {
		return a, b
	}
	return b, a
}

// NewSuppression creates a rejection record for a pair. linkType nil means
// "no link at all"; a non-nil value rejects that link type only.
func NewSuppression(a, b uuid.UUID, linkType *string, creatorID *uuid.UUID) *Suppression {
	noteA, noteB := NormalizePair(a, b)
	return &Suppression{
		id:        uuid.New(),
		noteAID:   noteA,
		noteBID:   noteB,
		linkType:  linkType,
		creatorID: creatorID,
		createdAt: time.Now(),
	}
}

// ReconstructSuppression rebuilds a suppression from stored data.
func ReconstructSuppression(id, noteA, noteB uuid.UUID, linkType *string, creatorID *uuid.UUID, createdAt time.Time) *Suppression {
	return &Suppression{
		id:        id,
		noteAID:   noteA,
		noteBID:   noteB,
		linkType:  linkType,
		creatorID: creatorID,
		createdAt: createdAt,
	}
}

func (s *Suppression) ID() uuid.UUID         { return s.id }
func (s *Suppression) NoteAID() uuid.UUID    { return s.noteAID }
func (s *Suppression) NoteBID() uuid.UUID    { return s.noteBID }
func (s *Suppression) LinkType() *string     { return s.linkType }
func (s *Suppression) CreatorID() *uuid.UUID { return s.creatorID }
func (s *Suppression) CreatedAt() time.Time  { return s.createdAt }

// Blocks reports whether this suppression rejects a link of the given type
// on the given pair (any direction).
func (s *Suppression) Blocks(a, b uuid.UUID, linkType string) bool {
	na, nb := NormalizePair(a, b)
	if s.noteAID != na || s.noteBID != nb {
		return false
	}
	return s.linkType == nil || *s.linkType == linkType
}

// SuppressionRepository stores and queries pair rejections. The method names
// carry the "Suppression" prefix so a store that also implements Repository
// (links) can satisfy both interfaces.
type SuppressionRepository interface {
	// SaveSuppression upserts a rejection (unique on normalized pair + link_type).
	SaveSuppression(ctx context.Context, s *Suppression) error
	// FindSuppressionsForNotes returns all rejections that involve any of the notes.
	FindSuppressionsForNotes(ctx context.Context, noteIDs []uuid.UUID) ([]*Suppression, error)
}
