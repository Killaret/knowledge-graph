package graph

import (
	"context"

	"github.com/google/uuid"
)

// weightedPath — результат BFS с весом и глубиной
type weightedPath struct {
	weight float64
	depth  int
}

// bfsItem — элемент очереди BFS
type bfsItem struct {
	nodeID uuid.UUID
	weight float64
	depth  int
}

// runBFS — алгоритм обхода в ширину для графовых рекомендаций
func runBFS(ctx context.Context, startID uuid.UUID, loader NeighborLoader, depth int, decay float64, aggregation string) map[uuid.UUID]weightedPath {
	if depth < 1 {
		depth = 1
	}
	if decay <= 0 || decay > 1 {
		decay = 0.5
	}

	bestWeight := make(map[uuid.UUID]weightedPath)
	queue := []bfsItem{{nodeID: startID, weight: 1.0, depth: 0}}
	bestWeight[startID] = weightedPath{weight: 1.0, depth: 0}

	// Для sum-агрегации отслеживаем visited, чтобы не обрабатывать узлы повторно
	var visited map[uuid.UUID]bool
	if aggregation == "sum" {
		visited = make(map[uuid.UUID]bool)
		visited[startID] = true
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Для max-агрегации: пропускаем, если нашли лучший путь
		if aggregation == "max" {
			if current.weight < bestWeight[current.nodeID].weight {
				continue
			}
		}

		if current.depth >= depth {
			continue
		}

		neighbors, err := loader.GetNeighbors(ctx, current.nodeID)
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

			if aggregation == "max" {
				if newWeight > bestWeight[edge.To].weight {
					bestWeight[edge.To] = weightedPath{weight: newWeight, depth: current.depth + 1}
					queue = append(queue, bfsItem{nodeID: edge.To, weight: newWeight, depth: current.depth + 1})
				}
			} else { // "sum"
				existing := bestWeight[edge.To]
				bestWeight[edge.To] = weightedPath{weight: existing.weight + newWeight, depth: current.depth + 1}
				if !visited[edge.To] {
					visited[edge.To] = true
					queue = append(queue, bfsItem{nodeID: edge.To, weight: newWeight, depth: current.depth + 1})
				}
			}
		}
	}

	delete(bestWeight, startID)

	return bestWeight
}
