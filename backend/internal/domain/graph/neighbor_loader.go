package graph

import (
	"context"

	"github.com/google/uuid"
)

// Edge — connection in the graph
type Edge struct {
	From   uuid.UUID
	To     uuid.UUID
	Weight float64
}

// NeighborLoader — interface for loading neighbors
type NeighborLoader interface {
	GetNeighbors(ctx context.Context, nodeID uuid.UUID) ([]Edge, error)
	GetNeighborsBatch(ctx context.Context, nodeIDs []uuid.UUID) (map[uuid.UUID][]Edge, error)
}
