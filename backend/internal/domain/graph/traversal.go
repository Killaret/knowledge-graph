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

func (s *TraversalService) GetSuggestions(ctx context.Context, startID uuid.UUID, topN int) ([]SuggestionResult, error) {
	depth := s.depth
	decay := s.decay
	if depth < 1 {
		depth = 1
	}
	if decay <= 0 || decay > 1 {
		decay = 0.5
	}

	scores := make(map[uuid.UUID]float64)
	type bfsItem struct {
		nodeID uuid.UUID
		weight float64
		depth  int
	}
	queue := []bfsItem{{nodeID: startID, weight: 1.0, depth: 0}}
	visited := map[uuid.UUID]bool{startID: true}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.depth >= depth {
			continue
		}

		neighbors, err := s.loader.GetNeighbors(ctx, current.nodeID)
		if err != nil {
			return nil, err
		}

		for _, edge := range neighbors {
			if edge.To == startID {
				continue
			}
			newWeight := current.weight * edge.Weight
			if current.depth > 0 {
				newWeight *= decay
			}
			if _, ok := scores[edge.To]; !ok {
				scores[edge.To] = newWeight
			} else {
				scores[edge.To] += newWeight
			}
			if !visited[edge.To] {
				visited[edge.To] = true
				queue = append(queue, bfsItem{
					nodeID: edge.To,
					weight: newWeight,
					depth:  current.depth + 1,
				})
			}
		}
	}

	results := make([]SuggestionResult, 0, len(scores))
	for id, score := range scores {
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
