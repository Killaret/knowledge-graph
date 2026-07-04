package graph

import (
	"context"
	"log"

	"knowledge-graph/internal/domain/graph"
	"knowledge-graph/internal/infrastructure/db/postgres"

	"github.com/google/uuid"
)

// embeddingRepositoryWithBatch is an interface for batch loading of embeddings
type embeddingRepositoryWithBatch interface {
	FindSimilarNotes(ctx context.Context, noteID uuid.UUID, limit int) ([]postgres.SimilarNote, error)
	FindSimilarNotesBatch(ctx context.Context, noteIDs []uuid.UUID, limit int) (map[uuid.UUID][]postgres.SimilarNote, error)
}

// embeddingNeighborLoader loads neighbors based on semantic similarity of embeddings.
type embeddingNeighborLoader struct {
	embeddingRepo embeddingRepositoryWithBatch
	limit         int
}

// NewEmbeddingNeighborLoader creates a loader for embeddings.
// limit is how many of the most similar notes to return (20-50 recommended).
func NewEmbeddingNeighborLoader(embeddingRepo *postgres.EmbeddingRepository, limit int) graph.NeighborLoader {
	return &embeddingNeighborLoader{
		embeddingRepo: embeddingRepo,
		limit:         limit,
	}
}

// GetNeighbors returns edges to semantically similar notes.
// The edge weight is the cosine similarity of the embeddings (from 0 to 1).
func (l *embeddingNeighborLoader) GetNeighbors(ctx context.Context, nodeID uuid.UUID) ([]graph.Edge, error) {
	similar, err := l.embeddingRepo.FindSimilarNotes(ctx, nodeID, l.limit)
	if err != nil {
		log.Printf("embeddingNeighborLoader: failed to find similar notes for %s: %v", nodeID, err)
		return nil, err
	}

	edges := make([]graph.Edge, len(similar))
	for i, s := range similar {
		edges[i] = graph.Edge{
			From:   nodeID,
			To:     s.NoteID,
			Weight: s.Score,
		}
	}
	return edges, nil
}

// GetNeighborsBatch returns neighbors for multiple nodes based on embeddings (batch query)
func (l *embeddingNeighborLoader) GetNeighborsBatch(ctx context.Context, nodeIDs []uuid.UUID) (map[uuid.UUID][]graph.Edge, error) {
	if len(nodeIDs) == 0 {
		return make(map[uuid.UUID][]graph.Edge), nil
	}

	// Try to assert the repository to the interface with a batch method
	batchRepo, ok := l.embeddingRepo.(interface {
		FindSimilarNotesBatch(ctx context.Context, noteIDs []uuid.UUID, limit int) (map[uuid.UUID][]postgres.SimilarNote, error)
	})

	if !ok {
		// Fall back to sequential queries
		result := make(map[uuid.UUID][]graph.Edge, len(nodeIDs))
		for _, nodeID := range nodeIDs {
			edges, err := l.GetNeighbors(ctx, nodeID)
			if err != nil {
				continue
			}
			result[nodeID] = edges
		}
		return result, nil
	}

	// Batch query
	similarMap, err := batchRepo.FindSimilarNotesBatch(ctx, nodeIDs, l.limit)
	if err != nil {
		// Fall back to sequential queries on error
		result := make(map[uuid.UUID][]graph.Edge, len(nodeIDs))
		for _, nodeID := range nodeIDs {
			edges, err := l.GetNeighbors(ctx, nodeID)
			if err != nil {
				continue
			}
			result[nodeID] = edges
		}
		return result, nil
	}

	// Convert the result into an Edge
	result := make(map[uuid.UUID][]graph.Edge, len(nodeIDs))
	for nodeID, similar := range similarMap {
		edges := make([]graph.Edge, len(similar))
		for i, s := range similar {
			edges[i] = graph.Edge{
				From:   nodeID,
				To:     s.NoteID,
				Weight: s.Score,
			}
		}
		result[nodeID] = edges
	}

	// Ensure that all requested nodes are present in the result (even if empty)
	for _, nodeID := range nodeIDs {
		if _, ok := result[nodeID]; !ok {
			result[nodeID] = []graph.Edge{}
		}
	}

	return result, nil
}
