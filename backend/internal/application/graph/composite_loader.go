package graph

import (
	"context"
	"fmt"
	"log"
	"strings"

	"knowledge-graph/internal/domain/graph"

	"github.com/google/uuid"
)

// compositeNeighborLoader combines several loaders and applies weights to their edges.
type compositeNeighborLoader struct {
	loaders []graph.NeighborLoader
	weights []float64 // weights for each loader (must match in length)
}

// NewCompositeNeighborLoaderWithWeights creates a composite loader with weights.
// Example: loaders = [linkLoader, embeddingLoader], weights = [0.7, 0.3]
func NewCompositeNeighborLoaderWithWeights(loaders []graph.NeighborLoader, weights []float64) graph.NeighborLoader {
	return &compositeNeighborLoader{
		loaders: loaders,
		weights: weights,
	}
}

// GetNeighbors queries all inner loaders, multiplies their edge weights by the corresponding coefficients
// and merges the results. If at least one loader succeeds, the merged list is returned.
// If all loaders fail, an aggregated error is returned.
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

// GetNeighborsBatch returns neighbors for multiple nodes by merging the results of the inner loaders
func (c *compositeNeighborLoader) GetNeighborsBatch(ctx context.Context, nodeIDs []uuid.UUID) (map[uuid.UUID][]graph.Edge, error) {
	result := make(map[uuid.UUID][]graph.Edge)
	var errs []string
	anySuccess := false

	// Initialize empty slices for all nodeIDs
	for _, nodeID := range nodeIDs {
		result[nodeID] = []graph.Edge{}
	}

	for i, loader := range c.loaders {
		batchResult, err := loader.GetNeighborsBatch(ctx, nodeIDs)
		if err != nil {
			errMsg := fmt.Sprintf("loader %T batch: %v", loader, err)
			log.Printf("compositeNeighborLoader: %s", errMsg)
			errs = append(errs, errMsg)
			continue
		}
		anySuccess = true

		weight := 1.0
		if i < len(c.weights) {
			weight = c.weights[i]
		}

		// Merge results applying the weight
		for nodeID, edges := range batchResult {
			for _, e := range edges {
				e.Weight *= weight
				result[nodeID] = append(result[nodeID], e)
			}
		}
	}

	if !anySuccess && len(errs) > 0 {
		return nil, fmt.Errorf("compositeNeighborLoader: all loaders failed: %s", strings.Join(errs, "; "))
	}

	return result, nil
}
