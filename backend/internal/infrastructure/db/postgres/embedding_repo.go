package postgres

import (
	"context"
	"time"

	apprec "knowledge-graph/internal/application/recommendation"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type EmbeddingRepository struct {
	db        *gorm.DB
	modelName string
}

// NewEmbeddingRepository creates an embedding repository for a specific model.
// All reads and writes use modelName to avoid mixing vectors from different
// embedding spaces.
func NewEmbeddingRepository(db *gorm.DB, modelName string) *EmbeddingRepository {
	if modelName == "" {
		modelName = "paraphrase-multilingual-MiniLM-L12-v2"
	}
	return &EmbeddingRepository{db: db, modelName: modelName}
}

// Upsert creates or updates an embedding for a note.
// The current model name is written explicitly because the column has no default.
func (r *EmbeddingRepository) Upsert(ctx context.Context, noteID uuid.UUID, embedding pgvector.Vector) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "note_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"embedding", "model_name", "updated_at"}),
	}).Create(&NoteEmbeddingModel{
		NoteID:    noteID,
		Embedding: embedding,
		ModelName: r.modelName,
		UpdatedAt: time.Now(),
	}).Error
}

// Delete removes an embedding for a note
func (r *EmbeddingRepository) Delete(ctx context.Context, noteID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("note_id = ?", noteID).Delete(&NoteEmbeddingModel{}).Error
}

// FindSimilarNotes returns up to limit notes that are semantically similar to the given note.
// Only vectors produced by the same model are compared to avoid mixing incompatible spaces.
func (r *EmbeddingRepository) FindSimilarNotes(ctx context.Context, noteID uuid.UUID, limit int) ([]apprec.SimilarNote, error) {
	var results []apprec.SimilarNote

	err := r.db.WithContext(ctx).Raw(`
        SELECT e2.note_id, (1 - (e1.embedding <=> e2.embedding)) / 2.0 as similarity
        FROM note_embeddings e1
        JOIN note_embeddings e2 ON e1.note_id != e2.note_id AND e1.model_name = e2.model_name
        WHERE e1.note_id = ? AND e1.model_name = ? AND e2.model_name = ?
        ORDER BY similarity DESC
        LIMIT ?
    `, noteID, r.modelName, r.modelName, limit).Scan(&results).Error

	if err != nil {
		return nil, err
	}
	return results, nil
}

// BatchSimilarNote represents a similar note in a batch query
type BatchSimilarNote struct {
	SourceID uuid.UUID `gorm:"column:source_id"`
	NoteID   uuid.UUID `gorm:"column:note_id"`
	Score    float64   `gorm:"column:similarity"`
}

// FindSimilarNotesBatch returns similar notes for multiple note IDs (batch query).
// Only vectors produced by the same model are compared.
func (r *EmbeddingRepository) FindSimilarNotesBatch(ctx context.Context, noteIDs []uuid.UUID, limit int) (map[uuid.UUID][]apprec.SimilarNote, error) {
	if len(noteIDs) == 0 {
		return make(map[uuid.UUID][]apprec.SimilarNote), nil
	}

	var results []BatchSimilarNote

	// DISTINCT ON is used to get top-N for each source_id
	err := r.db.WithContext(ctx).Raw(`
        SELECT DISTINCT ON (e1.note_id, e2.note_id) 
            e1.note_id as source_id,
            e2.note_id,
            (1 - (e1.embedding <=> e2.embedding)) / 2.0 as similarity
        FROM note_embeddings e1
        JOIN note_embeddings e2 ON e1.note_id != e2.note_id AND e1.model_name = e2.model_name
        WHERE e1.note_id = ANY(?) AND e1.model_name = ? AND e2.model_name = ?
        ORDER BY e1.note_id, e2.note_id, similarity DESC
    `, noteIDs, r.modelName, r.modelName).Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Group results by source_id
	grouped := make(map[uuid.UUID][]apprec.SimilarNote)
	for _, res := range results {
		// Limit number of results for each source
		if len(grouped[res.SourceID]) < limit {
			grouped[res.SourceID] = append(grouped[res.SourceID], apprec.SimilarNote{
				NoteID: res.NoteID,
				Score:  res.Score,
			})
		}
	}

	return grouped, nil
}

// FindNoteIDsMissingModel returns note IDs that do not have an embedding for the
// current model. A note with an embedding from a different model is included,
// because the PK is `note_id` and a single note only stores one vector at a time.
func (r *EmbeddingRepository) FindNoteIDsMissingModel(ctx context.Context) ([]uuid.UUID, error) {
	var noteIDs []string

	err := r.db.WithContext(ctx).Raw(`
        SELECT n.id
        FROM notes n
        LEFT JOIN note_embeddings e ON n.id = e.note_id AND e.model_name = ?
        WHERE e.note_id IS NULL
          AND n.deleted_at IS NULL
    `, r.modelName).Scan(&noteIDs).Error

	if err != nil {
		return nil, err
	}

	parsed := make([]uuid.UUID, 0, len(noteIDs))
	for _, id := range noteIDs {
		uid, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, uid)
	}

	return parsed, nil
}
