package graph

import (
	"context"

	"github.com/google/uuid"
)

// GraphServiceClient is a domain port for delegating graph analytics to the
// dedicated graph-service. Implementations live in the infrastructure layer.
type GraphServiceClient interface {
	// GetRecommendations returns ranked note recommendations for the source note.
	GetRecommendations(ctx context.Context, noteID uuid.UUID, limit int) ([]SuggestionResult, error)

	// GetNeighbors returns a map of neighbor note IDs to their combined weights.
	GetNeighbors(ctx context.Context, noteID uuid.UUID, depth int) (map[uuid.UUID]float64, error)

	// GetPath returns the shortest path (list of note IDs) between two notes.
	GetPath(ctx context.Context, from, to uuid.UUID) ([]uuid.UUID, error)
}
