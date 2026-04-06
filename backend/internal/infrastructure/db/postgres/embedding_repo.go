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

// SimilarNote — структура для результата поиска похожих заметок
type SimilarNote struct {
	NoteID uuid.UUID `gorm:"column:note_id"`
	Score  float64   `gorm:"column:similarity"`
}

// FindSimilarNotes возвращает до limit заметок, семантически похожих на данную.
func (r *EmbeddingRepository) FindSimilarNotes(ctx context.Context, noteID uuid.UUID, limit int) ([]SimilarNote, error) {
	var results []SimilarNote

	err := r.db.WithContext(ctx).Raw(`
        SELECT e2.note_id, (1 - (e1.embedding <=> e2.embedding)) / 2.0 as similarity
        FROM note_embeddings e1
        JOIN note_embeddings e2 ON e1.note_id != e2.note_id
        WHERE e1.note_id = ?
        ORDER BY similarity DESC
        LIMIT ?
    `, noteID, limit).Scan(&results).Error

	if err != nil {
		return nil, err
	}
	return results, nil
}
