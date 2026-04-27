package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

// NoteEmbeddingModel — векторное представление заметки
type NoteEmbeddingModel struct {
	NoteID    uuid.UUID       `gorm:"type:uuid;primaryKey"`
	Embedding pgvector.Vector `gorm:"type:vector(384)"`
	UpdatedAt time.Time

	Note NoteModel `gorm:"foreignKey:NoteID"`
}

func (NoteEmbeddingModel) TableName() string {
	return "note_embeddings"
}
