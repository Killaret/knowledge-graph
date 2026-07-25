package note

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Note struct {
	id        uuid.UUID
	title     Title
	content   Content
	type_     string
	metadata  Metadata
	creatorID *uuid.UUID
	isPublic  bool
	createdAt time.Time
	updatedAt time.Time
}

// NoteOption configures a Note during construction.
type NoteOption func(*Note)

// WithIsPublic sets the public visibility flag on a reconstructed note.
func WithIsPublic(isPublic bool) NoteOption {
	return func(n *Note) {
		n.isPublic = isPublic
	}
}

func NewNote(title Title, content Content, noteType string, metadata Metadata, opts ...NoteOption) *Note {
	now := time.Now()
	if noteType == "" {
		noteType = "star"
	}
	n := &Note{
		id:        uuid.New(),
		title:     title,
		content:   content,
		type_:     noteType,
		metadata:  metadata,
		creatorID: nil,
		createdAt: now,
		updatedAt: now,
	}
	for _, opt := range opts {
		opt(n)
	}
	return n
}

// NewNoteWithCreator creates a new note with a creator ID
func NewNoteWithCreator(title Title, content Content, noteType string, metadata Metadata, creatorID uuid.UUID, opts ...NoteOption) *Note {
	now := time.Now()
	if noteType == "" {
		noteType = "star"
	}
	n := &Note{
		id:        uuid.New(),
		title:     title,
		content:   content,
		type_:     noteType,
		metadata:  metadata,
		creatorID: &creatorID,
		createdAt: now,
		updatedAt: now,
	}
	for _, opt := range opts {
		opt(n)
	}
	return n
}

// ReconstructNote reconstructs a note from saved data (used by repository)
func ReconstructNote(id uuid.UUID, title Title, content Content, noteType string, metadata Metadata, createdAt, updatedAt time.Time, opts ...NoteOption) *Note {
	if noteType == "" {
		noteType = "star"
	}
	n := &Note{
		id:        id,
		title:     title,
		content:   content,
		type_:     noteType,
		metadata:  metadata,
		creatorID: nil,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
	for _, opt := range opts {
		opt(n)
	}
	return n
}

// ReconstructNoteWithCreator reconstructs a note with creator ID
func ReconstructNoteWithCreator(id uuid.UUID, title Title, content Content, noteType string, metadata Metadata, creatorID *uuid.UUID, createdAt, updatedAt time.Time, opts ...NoteOption) *Note {
	if noteType == "" {
		noteType = "star"
	}
	n := &Note{
		id:        id,
		title:     title,
		content:   content,
		type_:     noteType,
		metadata:  metadata,
		creatorID: creatorID,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
	for _, opt := range opts {
		opt(n)
	}
	return n
}

// Getters
func (n *Note) ID() uuid.UUID {
	return n.id
}

func (n *Note) Title() Title {
	return n.title
}

func (n *Note) Content() Content {
	return n.content
}

func (n *Note) Metadata() Metadata {
	return n.metadata
}

func (n *Note) Type() string {
	return n.type_
}

func (n *Note) CreatorID() *uuid.UUID {
	return n.creatorID
}

func (n *Note) IsPublic() bool {
	return n.isPublic
}

func (n *Note) SetCreatorID(creatorID uuid.UUID) {
	n.creatorID = &creatorID
	n.updatedAt = time.Now()
}

func (n *Note) SetType(noteType string) {
	if noteType != "" {
		n.type_ = noteType
		n.updatedAt = time.Now()
	}
}

func (n *Note) SetIsPublic(isPublic bool) {
	n.isPublic = isPublic
	n.updatedAt = time.Now()
}

func (n *Note) CreatedAt() time.Time {
	return n.createdAt
}

func (n *Note) UpdatedAt() time.Time {
	return n.updatedAt
}

// Mutation methods with validation
func (n *Note) UpdateTitle(newTitle Title) error {
	if newTitle.String() == "" {
		return fmt.Errorf("cannot update with empty title")
	}
	n.title = newTitle
	n.updatedAt = time.Now()
	return nil
}

func (n *Note) UpdateContent(newContent Content) error {
	if newContent.String() == "" {
		return fmt.Errorf("cannot update with empty content")
	}
	n.content = newContent
	n.updatedAt = time.Now()
	return nil
}

func (n *Note) UpdateMetadata(newMetadata Metadata) error {
	if newMetadata.Value() == nil {
		return fmt.Errorf("cannot update with nil metadata")
	}
	n.metadata = newMetadata
	n.updatedAt = time.Now()
	return nil
}
