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
	GetNeighborsBatch(ctx context.Context, nodeIDs []uuid.UUID) (map[uuid.UUID][]Edge, error)
}

type TraversalService struct {
	loader      NeighborLoader
	depth       int
	decay       float64
	aggregation string // "max" или "sum"
	normalize   bool
}

func NewTraversalService(loader NeighborLoader, depth int, decay float64, aggregation string, normalize bool) *TraversalService {
	if aggregation != "max" && aggregation != "sum" {
		aggregation = "max"
	}
	return &TraversalService{
		loader:      loader,
		depth:       depth,
		decay:       decay,
		aggregation: aggregation,
		normalize:   normalize,
	}
}

type SuggestionResult struct {
	NodeID uuid.UUID
	Score  float64
}

type bfsItem struct {
	nodeID uuid.UUID
	weight float64
	depth  int
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
	queue := []bfsItem{{nodeID: startID, weight: 1.0, depth: 0}}
	bestWeight[startID] = 1.0

	// Для sum-агрегации отслеживаем visited, чтобы не обрабатывать узлы повторно
	var visited map[uuid.UUID]bool
	if s.aggregation == "sum" {
		visited = make(map[uuid.UUID]bool)
		visited[startID] = true
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Для max-агрегации: пропускаем, если нашли лучший путь
		if s.aggregation == "max" && current.weight < bestWeight[current.nodeID] {
			continue
		}

		if current.depth >= depth {
			continue
		}

		neighbors, err := s.loader.GetNeighbors(ctx, current.nodeID)
		if err != nil {
			continue
		}

		for _, edge := range neighbors {
			if edge.To == startID {
				continue
			}

			newWeight := current.weight * edge.Weight
			if current.depth > 0 {
				newWeight *= decay
			}

			if s.aggregation == "max" {
				if newWeight > bestWeight[edge.To] {
					bestWeight[edge.To] = newWeight
					queue = append(queue, bfsItem{nodeID: edge.To, weight: newWeight, depth: current.depth + 1})
				}
			} else { // "sum"
				bestWeight[edge.To] += newWeight
				if !visited[edge.To] {
					visited[edge.To] = true
					queue = append(queue, bfsItem{nodeID: edge.To, weight: newWeight, depth: current.depth + 1})
				}
			}
		}
	}

	delete(bestWeight, startID)

	// Нормализация
	if s.normalize {
		maxW := 0.0
		for _, w := range bestWeight {
			if w > maxW {
				maxW = w
			}
		}
		if maxW > 0 {
			for id := range bestWeight {
				bestWeight[id] /= maxW
			}
		}
	}

	return bestWeight
}

func (s *TraversalService) GetSuggestions(ctx context.Context, startID uuid.UUID, topN int) ([]SuggestionResult, error) {
	bestWeight := s.runBFS(ctx, startID)

	results := make([]SuggestionResult, 0, len(bestWeight))
	for id, score := range bestWeight {
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
