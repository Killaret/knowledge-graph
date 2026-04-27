package graph

import (
	"context"

	"github.com/google/uuid"
)

// Edge — связь в графе
type Edge struct {
	From   uuid.UUID
	To     uuid.UUID
	Weight float64
}

// NeighborLoader — интерфейс для загрузки соседей
type NeighborLoader interface {
	GetNeighbors(ctx context.Context, nodeID uuid.UUID) ([]Edge, error)
	GetNeighborsBatch(ctx context.Context, nodeIDs []uuid.UUID) (map[uuid.UUID][]Edge, error)
}
