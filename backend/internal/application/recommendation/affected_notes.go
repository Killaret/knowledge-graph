package recommendation

import (
	"context"

	"knowledge-graph/internal/infrastructure/db/postgres"

	"github.com/google/uuid"
)

// reverseCascadeDepth limits the depth of reverse dependency lookup
// to prevent queue explosion. Currently only direct reverse dependencies (depth = 1).
const reverseCascadeDepth = 1

// AffectedNotesService determines which notes should have their recommendations recalculated
// when a specific note changes.
type AffectedNotesService struct {
	recRepo *postgres.RecommendationRepository
}

// NewAffectedNotesService creates a new affected notes service
func NewAffectedNotesService(recRepo *postgres.RecommendationRepository) *AffectedNotesService {
	return &AffectedNotesService{recRepo: recRepo}
}

// GetAffectedNotes returns note IDs that should be recalculated due to changes on targetNoteID.
// Currently limits to direct reverse dependencies (depth = 1) to prevent queue explosion.
// This includes:
// - The target note itself
// - Notes that directly recommend the target note
func (s *AffectedNotesService) GetAffectedNotes(ctx context.Context, targetNoteID uuid.UUID) ([]uuid.UUID, error) {
	affected := []uuid.UUID{targetNoteID}

	// Only get direct reverse dependencies (notes that recommend targetNoteID)
	// This prevents exponential growth in queue size
	if reverseCascadeDepth >= 1 {
		reverse, err := s.recRepo.GetNotesThatRecommend(ctx, targetNoteID)
		if err != nil {
			return nil, err
		}
		affected = append(affected, reverse...)
	}

	return deduplicate(affected), nil
}

// deduplicate removes duplicate UUIDs from a slice
func deduplicate(ids []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]bool, len(ids))
	result := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	return result
}
