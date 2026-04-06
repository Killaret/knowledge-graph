package graph

import (
	"context"
	"log"

	"knowledge-graph/internal/domain/graph"

	"github.com/google/uuid"
)

// compositeNeighborLoader объединяет несколько загрузчиков и применяет веса к их рёбрам.
type compositeNeighborLoader struct {
	loaders []graph.NeighborLoader
	weights []float64 // веса для каждого загрузчика (должны совпадать по длине)
}

// NewCompositeNeighborLoaderWithWeights создаёт композитный загрузчик с весами.
// Пример: loaders = [linkLoader, embeddingLoader], weights = [0.7, 0.3]
func NewCompositeNeighborLoaderWithWeights(loaders []graph.NeighborLoader, weights []float64) graph.NeighborLoader {
	return &compositeNeighborLoader{
		loaders: loaders,
		weights: weights,
	}
}

// GetNeighbors опрашивает все внутренние загрузчики, умножает веса их рёбер на соответствующие коэффициенты
// и объединяет результаты. Если один из загрузчиков вернул ошибку, она логируется, но остальные продолжают работу.
func (c *compositeNeighborLoader) GetNeighbors(ctx context.Context, nodeID uuid.UUID) ([]graph.Edge, error) {
	var allEdges []graph.Edge
	for i, loader := range c.loaders {
		edges, err := loader.GetNeighbors(ctx, nodeID)
		if err != nil {
			log.Printf("compositeNeighborLoader: error from loader %T: %v", loader, err)
			continue
		}
		weight := 1.0
		if i < len(c.weights) {
			weight = c.weights[i]
		}
		for _, e := range edges {
			e.Weight *= weight
			allEdges = append(allEdges, e)
		}
	}
	return allEdges, nil
}
