package notification

import (
	"time"

	"github.com/google/uuid"
)

// TypeCometReminder marks a notification produced by a comet reminder task.
const TypeCometReminder = "comet_reminder"

// Notification is an in-app message addressed to a single user.
// COMET-1: created by the comet:remind worker task; other kinds may reuse the
// table later via the type field.
type Notification struct {
	id        uuid.UUID
	userID    uuid.UUID
	noteID    *uuid.UUID
	type_     string
	title     string
	body      string
	dedupeKey string
	createdAt time.Time
	readAt    *time.Time
}

// NewNotification creates an unread notification. dedupeKey may be empty —
// when set, the repository refuses duplicates.
func NewNotification(userID uuid.UUID, noteID *uuid.UUID, typ, title, body, dedupeKey string) *Notification {
	return &Notification{
		id:        uuid.New(),
		userID:    userID,
		noteID:    noteID,
		type_:     typ,
		title:     title,
		body:      body,
		dedupeKey: dedupeKey,
		createdAt: time.Now(),
	}
}

// Reconstruct rebuilds a notification from persistence.
func Reconstruct(id, userID uuid.UUID, noteID *uuid.UUID, typ, title, body, dedupeKey string, createdAt time.Time, readAt *time.Time) *Notification {
	return &Notification{id: id, userID: userID, noteID: noteID, type_: typ, title: title, body: body, dedupeKey: dedupeKey, createdAt: createdAt, readAt: readAt}
}

func (n *Notification) ID() uuid.UUID        { return n.id }
func (n *Notification) UserID() uuid.UUID    { return n.userID }
func (n *Notification) NoteID() *uuid.UUID   { return n.noteID }
func (n *Notification) Type() string         { return n.type_ }
func (n *Notification) Title() string        { return n.title }
func (n *Notification) Body() string         { return n.body }
func (n *Notification) DedupeKey() string    { return n.dedupeKey }
func (n *Notification) CreatedAt() time.Time { return n.createdAt }
func (n *Notification) ReadAt() *time.Time   { return n.readAt }
func (n *Notification) IsRead() bool         { return n.readAt != nil }

// MarkRead stamps the notification as read (idempotent).
func (n *Notification) MarkRead() {
	if n.readAt != nil {
		return
	}
	now := time.Now()
	n.readAt = &now
}
