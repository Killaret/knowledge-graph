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
	metadata  Metadata
	createdAt time.Time
	updatedAt time.Time
}

func NewNote(title Title, content Content, metadata Metadata) *Note {
	now := time.Now()
	return &Note{
		id:        uuid.New(),
		title:     title,
		content:   content,
		metadata:  metadata,
		createdAt: now,
		updatedAt: now,
	}
}

// ReconstructNote восстанавливает заметку из сохранённых данных (используется репозиторием)
func ReconstructNote(id uuid.UUID, title Title, content Content, metadata Metadata, createdAt, updatedAt time.Time) *Note {
	return &Note{
		id:        id,
		title:     title,
		content:   content,
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
