package note

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Note struct {
	id                  uuid.UUID
	title               Title
	content             Content
	type_               NoteType
	metadata            Metadata
	creatorID           *uuid.UUID
	isPublic            bool
	dueAt               *time.Time
	remindBeforeSeconds *int64
	doneAt              *time.Time
	createdAt           time.Time
	updatedAt           time.Time
}

// NoteOption configures a Note during construction.
type NoteOption func(*Note)

// WithIsPublic sets the public visibility flag on a reconstructed note.
func WithIsPublic(isPublic bool) NoteOption {
	return func(n *Note) {
		n.isPublic = isPublic
	}
}

// WithID sets a specific id on a newly created note.
// Used by importers that generate note IDs ahead of time so links can reference them.
func WithID(id uuid.UUID) NoteOption {
	return func(n *Note) {
		n.id = id
	}
}

// WithCometFields restores comet scheduling fields when reconstructing a note.
// The fields are kept for any type so re-typing a comet does not lose the date.
func WithCometFields(dueAt *time.Time, remindBeforeSeconds *int64, doneAt *time.Time) NoteOption {
	return func(n *Note) {
		n.dueAt = dueAt
		n.remindBeforeSeconds = remindBeforeSeconds
		n.doneAt = doneAt
	}
}

func NewNote(title Title, content Content, noteType NoteType, metadata Metadata, opts ...NoteOption) *Note {
	now := time.Now()
	if !noteType.IsValid() {
		noteType = Unknown()
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
func NewNoteWithCreator(title Title, content Content, noteType NoteType, metadata Metadata, creatorID uuid.UUID, opts ...NoteOption) *Note {
	now := time.Now()
	if !noteType.IsValid() {
		noteType = Unknown()
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
func ReconstructNote(id uuid.UUID, title Title, content Content, noteType NoteType, metadata Metadata, createdAt, updatedAt time.Time, opts ...NoteOption) *Note {
	if !noteType.IsValid() {
		noteType = Unknown()
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
func ReconstructNoteWithCreator(id uuid.UUID, title Title, content Content, noteType NoteType, metadata Metadata, creatorID *uuid.UUID, createdAt, updatedAt time.Time, opts ...NoteOption) *Note {
	if !noteType.IsValid() {
		noteType = Unknown()
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
	return n.type_.String()
}

func (n *Note) CreatorID() *uuid.UUID {
	return n.creatorID
}

func (n *Note) IsPublic() bool {
	return n.isPublic
}

// IsOwnedBy reports whether userID matches the note's creator. It is a
// pure id comparison: callers MUST combine it with an "identity present"
// check — anonymous callers carry uuid.Nil, and the seeded test user of
// the test stack legitimately authenticates as uuid.Nil (migration 019),
// so Nil-vs-Nil alone cannot mean "not owned".
func (n *Note) IsOwnedBy(userID uuid.UUID) bool {
	return n.creatorID != nil && *n.creatorID == userID
}

func (n *Note) SetCreatorID(creatorID uuid.UUID) {
	n.creatorID = &creatorID
	n.updatedAt = time.Now()
}

func (n *Note) SetType(noteType NoteType) {
	if noteType.IsValid() {
		n.type_ = noteType
		n.updatedAt = time.Now()
	}
}

func (n *Note) SetIsPublic(isPublic bool) {
	n.isPublic = isPublic
	n.updatedAt = time.Now()
}

func (n *Note) DueAt() *time.Time {
	return n.dueAt
}

func (n *Note) RemindBeforeSeconds() *int64 {
	return n.remindBeforeSeconds
}

func (n *Note) DoneAt() *time.Time {
	return n.doneAt
}

// SetCometFields replaces the comet scheduling fields. A reminder offset is
// meaningless without a due date, so clearing dueAt clears the offset too.
func (n *Note) SetCometFields(dueAt *time.Time, remindBeforeSeconds *int64, doneAt *time.Time) error {
	if dueAt == nil && remindBeforeSeconds != nil {
		return fmt.Errorf("remind_before requires due_at")
	}
	n.dueAt = dueAt
	n.remindBeforeSeconds = remindBeforeSeconds
	n.doneAt = doneAt
	n.updatedAt = time.Now()
	return nil
}

// RemindAt returns the moment a reminder is due, or nil when the note has no
// date or no reminder offset. Done notes are not reminded.
func (n *Note) RemindAt() *time.Time {
	if n.dueAt == nil || n.remindBeforeSeconds == nil || n.doneAt != nil {
		return nil
	}
	t := n.dueAt.Add(-time.Duration(*n.remindBeforeSeconds) * time.Second)
	return &t
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
