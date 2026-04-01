package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type EmbeddingRepository struct {
	db *gorm.DB
}

func NewEmbeddingRepository(db *gorm.DB) *EmbeddingRepository {
	return &EmbeddingRepository{db: db}
}

// Upsert создаёт или обновляет эмбеддинг для заметки
func (r *EmbeddingRepository) Upsert(ctx context.Context, noteID uuid.UUID, embedding pgvector.Vector) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "note_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"embedding", "updated_at"}),
	}).Create(&NoteEmbeddingModel{
		NoteID:    noteID,
		Embedding: embedding,
		UpdatedAt: time.Now(),
	}).Error
}

// Delete удаляет эмбеддинг для заметки
func (r *EmbeddingRepository) Delete(ctx context.Context, noteID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("note_id = ?", noteID).Delete(&NoteEmbeddingModel{}).Error
}
