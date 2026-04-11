package graph

import (
	"context"
	"fmt"
	"log"
	"strings"

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
// и объединяет результаты. Если хотя бы один загрузчик успешно вернул результат — возвращаем объединённый список.
// Если все загрузчики завершились с ошибкой — возвращаем агрегированную ошибку.
func (c *compositeNeighborLoader) GetNeighbors(ctx context.Context, nodeID uuid.UUID) ([]graph.Edge, error) {
	var allEdges []graph.Edge
	var errs []string
	anySuccess := false

	for i, loader := range c.loaders {
		edges, err := loader.GetNeighbors(ctx, nodeID)
		if err != nil {
			errMsg := fmt.Sprintf("loader %T: %v", loader, err)
			log.Printf("compositeNeighborLoader: %s", errMsg)
			errs = append(errs, errMsg)
			continue
		}
		anySuccess = true

		weight := 1.0
		if i < len(c.weights) {
			weight = c.weights[i]
		}
		for _, e := range edges {
			e.Weight *= weight
			allEdges = append(allEdges, e)
		}
	}

	if !anySuccess && len(errs) > 0 {
		return nil, fmt.Errorf("compositeNeighborLoader: all loaders failed: %s", strings.Join(errs, "; "))
	}

	return allEdges, nil
}
