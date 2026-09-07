package note

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DraftState represents the state of a draft
type DraftState string

const (
	// DraftStateActive is the initial state when a draft is created
	DraftStateActive DraftState = "active"
	// DraftStatePublishing when the draft is being synchronized with the server
	DraftStatePublishing DraftState = "publishing"
	// DraftStatePublished when the draft has been successfully published
	DraftStatePublished DraftState = "published"
	// DraftStateConflict when there was a conflict during publishing
	DraftStateConflict DraftState = "conflict"
)

// DraftPublished is a domain event emitted when a draft is published
type DraftPublished struct {
	DraftID  uuid.UUID
	NoteID   uuid.UUID
	UserID   uuid.UUID
	Occurred time.Time
}

// Draft represents a note draft stored in MongoDB
type Draft struct {
	id        uuid.UUID
	noteID    uuid.UUID
	userID    uuid.UUID
	content   string
	title     string
	state     DraftState
	updatedAt time.Time
	createdAt time.Time
	events    []interface{}
}

// NewDraft creates a new draft in active state
func NewDraft(noteID, userID uuid.UUID, content, title string) *Draft {
	now := time.Now()
	return &Draft{
		id:        uuid.New(),
		noteID:    noteID,
		userID:    userID,
		content:   content,
		title:     title,
		state:     DraftStateActive,
		createdAt: now,
		updatedAt: now,
		events:    make([]interface{}, 0),
	}
}

// ReconstructDraft reconstructs a draft from stored data
func ReconstructDraft(id, noteID, userID uuid.UUID, content, title string, state DraftState, createdAt, updatedAt time.Time) *Draft {
	return &Draft{
		id:        id,
		noteID:    noteID,
		userID:    userID,
		content:   content,
		title:     title,
		state:     state,
		createdAt: createdAt,
		updatedAt: updatedAt,
		events:    make([]interface{}, 0),
	}
}

// Getters
func (d *Draft) ID() uuid.UUID {
	return d.id
}

func (d *Draft) NoteID() uuid.UUID {
	return d.noteID
}

func (d *Draft) UserID() uuid.UUID {
	return d.userID
}

func (d *Draft) Content() string {
	return d.content
}

func (d *Draft) Title() string {
	return d.title
}

func (d *Draft) State() DraftState {
	return d.state
}

func (d *Draft) UpdatedAt() time.Time {
	return d.updatedAt
}

func (d *Draft) CreatedAt() time.Time {
	return d.createdAt
}

// Events returns domain events
func (d *Draft) Events() []interface{} {
	return d.events
}

// ClearEvents clears domain events
func (d *Draft) ClearEvents() {
	d.events = make([]interface{}, 0)
}

// State transitions

// StartPublishing transitions the draft to publishing state
func (d *Draft) StartPublishing() error {
	if d.state != DraftStateActive {
		return fmt.Errorf("cannot start publishing from state %s, only from active", d.state)
	}
	d.state = DraftStatePublishing
	d.updatedAt = time.Now()
	return nil
}

// MarkAsPublished transitions the draft to published state and emits event
func (d *Draft) MarkAsPublished() error {
	if d.state != DraftStatePublishing {
		return fmt.Errorf("cannot mark as published from state %s, only from publishing", d.state)
	}
	d.state = DraftStatePublished
	d.updatedAt = time.Now()

	// Emit domain event
	event := DraftPublished{
		DraftID:  d.id,
		NoteID:   d.noteID,
		UserID:   d.userID,
		Occurred: time.Now(),
	}
	d.events = append(d.events, event)

	return nil
}

// MarkAsConflict transitions the draft to conflict state
func (d *Draft) MarkAsConflict() error {
	if d.state != DraftStatePublishing {
		return fmt.Errorf("cannot mark as conflict from state %s, only from publishing", d.state)
	}
	d.state = DraftStateConflict
	d.updatedAt = time.Now()
	return nil
}

// ResolveConflict transitions the draft back to active state for retry
func (d *Draft) ResolveConflict() error {
	if d.state != DraftStateConflict {
		return fmt.Errorf("cannot resolve conflict from state %s, only from conflict", d.state)
	}
	d.state = DraftStateActive
	d.updatedAt = time.Now()
	return nil
}

// UpdateContent updates the draft content
func (d *Draft) UpdateContent(content string) error {
	if d.state != DraftStateActive {
		return fmt.Errorf("cannot update content from state %s, only from active", d.state)
	}
	d.content = content
	d.updatedAt = time.Now()
	return nil
}

// UpdateTitle updates the draft title
func (d *Draft) UpdateTitle(title string) error {
	if d.state != DraftStateActive {
		return fmt.Errorf("cannot update title from state %s, only from active", d.state)
	}
	d.title = title
	d.updatedAt = time.Now()
	return nil
}
