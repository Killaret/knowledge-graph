// Package tag provides the Tag domain entity and value objects.
package tag

import (
	"time"

	"github.com/google/uuid"
)

const maxNameLength = 50

// Tag represents a label that can be attached to notes.
type Tag struct {
	id        uuid.UUID
	name      string
	createdAt time.Time
}

// New creates a new active Tag.
func New(name string) (*Tag, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Tag{
		id:        uuid.New(),
		name:      name,
		createdAt: now,
	}, nil
}

// Reconstruct rebuilds a Tag from persisted data.
func Reconstruct(id uuid.UUID, name string, createdAt time.Time) (*Tag, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}
	return &Tag{
		id:        id,
		name:      name,
		createdAt: createdAt,
	}, nil
}

// ID returns the tag identifier.
func (t *Tag) ID() uuid.UUID { return t.id }

// Name returns the tag name.
func (t *Tag) Name() string { return t.name }

// CreatedAt returns the tag creation timestamp.
func (t *Tag) CreatedAt() time.Time { return t.createdAt }

// Rename changes the tag name after validation.
func (t *Tag) Rename(name string) error {
	if err := validateName(name); err != nil {
		return err
	}
	t.name = name
	return nil
}

func validateName(name string) error {
	if name == "" {
		return ErrEmptyName
	}
	if len(name) > maxNameLength {
		return ErrNameTooLong
	}
	return nil
}
