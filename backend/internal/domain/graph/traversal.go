package graph

import (
	"context"
	"sort"

	"github.com/google/uuid"
)

type Node struct {
	ID uuid.UUID
}

type Edge struct {
	From   uuid.UUID
	To     uuid.UUID
	Weight float64
}

type NeighborLoader interface {
	GetNeighbors(ctx context.Context, nodeID uuid.UUID) ([]Edge, error)
}

type TraversalService struct {
	loader NeighborLoader
	depth  int
	decay  float64
}

func NewTraversalService(loader NeighborLoader, depth int, decay float64) *TraversalService {
	return &TraversalService{
		loader: loader,
		depth:  depth,
		decay:  decay,
	}
}

type SuggestionResult struct {
	NodeID uuid.UUID
	Score  float64
}

type bfsNode struct {
	ID     uuid.UUID
	Weight float64
	Depth  int
}

func normalizeWeights(weights map[uuid.UUID]float64) map[uuid.UUID]float64 {
	if len(weights) == 0 {
		return weights
	}
	maxW := 0.0
	for _, w := range weights {
		if w > maxW {
			maxW = w
		}
	}
	if maxW == 0 {
		return weights
	}
	normalized := make(map[uuid.UUID]float64, len(weights))
	for id, w := range weights {
		normalized[id] = w / maxW
	}
	return normalized
}

func (s *TraversalService) runBFS(ctx context.Context, startID uuid.UUID) map[uuid.UUID]float64 {
	depth := s.depth
	decay := s.decay
	if depth < 1 {
		depth = 1
	}
	if decay <= 0 || decay > 1 {
		decay = 0.5
	}

	bestWeight := make(map[uuid.UUID]float64)
	queue := []bfsNode{{ID: startID, Weight: 1.0, Depth: 0}}
	bestWeight[startID] = 1.0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.Weight < bestWeight[current.ID] {
			continue
		}

		if current.Depth >= depth {
			continue
		}

		neighbors, err := s.loader.GetNeighbors(ctx, current.ID)
		if err != nil {
			continue
		}

		for _, edge := range neighbors {
			if edge.To == startID {
				continue
			}
			candidateWeight := current.Weight * edge.Weight * decay

			if candidateWeight > bestWeight[edge.To] {
				bestWeight[edge.To] = candidateWeight
				queue = append(queue, bfsNode{
					ID:     edge.To,
					Weight: candidateWeight,
					Depth:  current.Depth + 1,
				})
			}
		}
	}

	delete(bestWeight, startID)
	return bestWeight
}

func (s *TraversalService) GetSuggestions(ctx context.Context, startID uuid.UUID, topN int) ([]SuggestionResult, error) {
	bestWeight := s.runBFS(ctx, startID)
	normalized := normalizeWeights(bestWeight)

	results := make([]SuggestionResult, 0, len(normalized))
	for id, score := range normalized {
		results = append(results, SuggestionResult{NodeID: id, Score: score})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > topN {
		results = results[:topN]
	}
	return results, nil
}
