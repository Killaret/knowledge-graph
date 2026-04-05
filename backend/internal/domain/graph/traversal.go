package graph

import (
	"context"
	"sort"

	"github.com/google/uuid"
)

// Node представляет заметку в графе (минимальная информация для расчётов)
type Node struct {
	ID uuid.UUID
}

// Edge представляет связь между заметками
type Edge struct {
	From   uuid.UUID
	To     uuid.UUID
	Weight float64
}

// NeighborLoader — интерфейс для загрузки соседей узла (реализуется в Application/Infrastructure)
type NeighborLoader interface {
	GetNeighbors(ctx context.Context, nodeID uuid.UUID) ([]Edge, error)
}

// TraversalService выполняет BFS-обход графа с распространением весов
type TraversalService struct {
	loader NeighborLoader
}

func NewTraversalService(loader NeighborLoader) *TraversalService {
	return &TraversalService{loader: loader}
}

// SuggestionResult содержит результат рекомендации
type SuggestionResult struct {
	NodeID uuid.UUID
	Score  float64
}

// GetSuggestions возвращает топ N заметок, релевантных для заданной, на основе BFS глубиной depth с затуханием decay
func (s *TraversalService) GetSuggestions(ctx context.Context, startID uuid.UUID, depth int, decay float64, topN int) ([]SuggestionResult, error) {
	if depth < 1 {
		depth = 1
	}
	if decay <= 0 || decay > 1 {
		decay = 0.5
	}

	// scores хранит накопленный вес для каждого узла
	scores := make(map[uuid.UUID]float64)

	// BFS очередь: хранит (nodeID, cumulativeWeight, currentDepth)
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
			// Не возвращаемся к стартовой ноде
			if edge.To == startID {
				continue
			}
			// Вычисляем новый вес: текущий * вес ребра * затухание на глубине
			newWeight := current.weight * edge.Weight * decay
			if _, ok := scores[edge.To]; !ok {
				scores[edge.To] = newWeight
			} else {
				scores[edge.To] += newWeight
			}
			// Добавляем в очередь, если не посещали ранее (чтобы избежать циклов)
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

	// Преобразуем в срез и сортируем по убыванию веса
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
