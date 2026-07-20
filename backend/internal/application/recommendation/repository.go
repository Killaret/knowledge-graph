package recommendation

import (
	"context"

	"github.com/google/uuid"
)

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
}
