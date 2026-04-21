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
	createdAt time.Time
	updatedAt time.Time
}

func NewNote(title Title, content Content, noteType string, metadata Metadata) *Note {
	now := time.Now()
	if noteType == "" {
		noteType = "star"
	}
	return &Note{
		id:        uuid.New(),
		title:     title,
		content:   content,
		type_:     noteType,
		metadata:  metadata,
		createdAt: now,
		updatedAt: now,
	}
}

// ReconstructNote восстанавливает заметку из сохранённых данных (используется репозиторием)
func ReconstructNote(id uuid.UUID, title Title, content Content, noteType string, metadata Metadata, createdAt, updatedAt time.Time) *Note {
	if noteType == "" {
		noteType = "star"
	}
	return &Note{
		id:        id,
		title:     title,
		content:   content,
		type_:     noteType,
		metadata:  metadata,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

// Геттеры
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

func (n *Note) SetType(noteType string) {
	if noteType != "" {
		n.type_ = noteType
		n.updatedAt = time.Now()
	}
}

func (n *Note) CreatedAt() time.Time {
	return n.createdAt
}

func (n *Note) UpdatedAt() time.Time {
	return n.updatedAt
}

// Методы изменения с валидацией
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
