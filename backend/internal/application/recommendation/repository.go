package recommendation

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Recommendation represents a precomputed recommendation stored for a note.
type Recommendation struct {
	NoteID            uuid.UUID
	RecommendedNoteID uuid.UUID
	Score             float64
	UpdatedAt         time.Time
}

// SimilarNote represents a semantically similar note returned by an embedding search.
type SimilarNote struct {
	NoteID uuid.UUID
	Score  float64
}

// Repository provides persistence operations needed by recommendation services.
// Implemented by infrastructure adapters (e.g. postgres.RecommendationRepository).
type Repository interface {
	// Count returns the number of recommendations stored for a note.
	Count(ctx context.Context, noteID uuid.UUID) (int64, error)

	// GetNotesThatRecommend returns IDs of notes that recommend the given note.
	GetNotesThatRecommend(ctx context.Context, recommendedID uuid.UUID) ([]uuid.UUID, error)

	// ReplaceRecommendations atomically replaces recommendations for a note
	// with the given set, removing stale entries.
	ReplaceRecommendations(ctx context.Context, noteID uuid.UUID, recs map[uuid.UUID]float64) error

	// GetRecommendations returns precomputed recommendations for a note sorted by score descending.
	GetRecommendations(ctx context.Context, noteID uuid.UUID, limit int) ([]Recommendation, error)
}

// EmbeddingRepository provides embedding-based semantic similarity search.
// Implemented by infrastructure adapters (e.g. postgres.EmbeddingRepository).
type EmbeddingRepository interface {
	// FindSimilarNotes returns up to limit notes that are semantically similar to the given note.
	FindSimilarNotes(ctx context.Context, noteID uuid.UUID, limit int) ([]SimilarNote, error)

	// FindSimilarNotesBatch returns similar notes for multiple source notes in a single query.
	FindSimilarNotesBatch(ctx context.Context, noteIDs []uuid.UUID, limit int) (map[uuid.UUID][]SimilarNote, error)
}
