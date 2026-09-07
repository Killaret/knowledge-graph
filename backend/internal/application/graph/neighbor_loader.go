package graph

import (
	"context"

	"knowledge-graph/internal/domain/graph"
	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
)

// neighborLoader — struct that implements the graph.NeighborLoader interface.
// It knows how to load note neighbors from the database through repositories.
type neighborLoader struct {
	linkRepo link.Repository // repository for working with links
	noteRepo note.Repository // repository for notes (not currently used, but may be useful)
}

// NewNeighborLoader — constructor that returns an implementation of the graph.NeighborLoader interface.
// Accepts repositories (dependencies) needed for loading data.
func NewNeighborLoader(linkRepo link.Repository, noteRepo note.Repository) graph.NeighborLoader {
	return &neighborLoader{
		linkRepo: linkRepo,
		noteRepo: noteRepo,
	}
}

// linkRepositoryWithBatch — interface that extends link.Repository with batch methods
type linkRepositoryWithBatch interface {
	link.Repository
	FindBySourceIDs(ctx context.Context, sourceIDs []uuid.UUID) (map[uuid.UUID][]*link.Link, error)
	FindByTargetIDs(ctx context.Context, targetIDs []uuid.UUID) (map[uuid.UUID][]*link.Link, error)
}

// GetNeighbors — method that returns a list of edges (connections) to neighboring notes for a given note (nodeID).
// For recommendations, we make the graph undirected: we consider both outgoing and incoming connections.
// Each edge contains:
//   - From: ID of the source note (always nodeID, but for incoming connections we reverse the direction)
//   - To:   ID of the neighboring note
//   - Weight: connection weight (already calculated and stored in the database)
func (l *neighborLoader) GetNeighbors(ctx context.Context, nodeID uuid.UUID) ([]graph.Edge, error) {
	// 1. Get outgoing connections: where our note is the source
	outgoing, err := l.linkRepo.FindBySource(ctx, nodeID)
	if err != nil {
		// If there's an error during the database query, return it so the upper layer can handle it
		return nil, err
	}

	// 2. Get incoming connections: where our note is the target
	incoming, err := l.linkRepo.FindByTarget(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	// 3. Prepare slice for edges (approximate size = sum of outgoing and incoming)
	edges := make([]graph.Edge, 0, len(outgoing)+len(incoming))

	// 4. Add outgoing connections as-is (direction from nodeID to target)
	for _, ln := range outgoing {
		edges = append(edges, graph.Edge{
			From:   ln.SourceNoteID(), // this is nodeID
			To:     ln.TargetNoteID(),
			Weight: ln.Weight().Value(), // weight is stored in Value Object Weight, take float64
		})
	}

	// 5. Add incoming connections, but reverse the direction (to make the graph undirected)
	//    For recommendations, it doesn't matter who references whom – the connection itself is important.
	for _, ln := range incoming {
		edges = append(edges, graph.Edge{
			From:   ln.TargetNoteID(), // in reversed form: from target note to source
			To:     ln.SourceNoteID(),
			Weight: ln.Weight().Value(),
		})
	}

	// 6. Return the list of edges
	return edges, nil
}

// GetNeighborsBatch returns neighbors for multiple nodes (batch query)
func (l *neighborLoader) GetNeighborsBatch(ctx context.Context, nodeIDs []uuid.UUID) (map[uuid.UUID][]graph.Edge, error) {
	if len(nodeIDs) == 0 {
		return make(map[uuid.UUID][]graph.Edge), nil
	}

	// Try to cast linkRepo to interface with batch methods
	batchRepo, ok := l.linkRepo.(linkRepositoryWithBatch)
	if !ok {
		// If batch methods are unavailable, fallback to sequential queries
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

	// Batch queries for outgoing and incoming connections
	outgoingMap, err := batchRepo.FindBySourceIDs(ctx, nodeIDs)
	if err != nil {
		outgoingMap = make(map[uuid.UUID][]*link.Link)
	}

	incomingMap, err := batchRepo.FindByTargetIDs(ctx, nodeIDs)
	if err != nil {
		incomingMap = make(map[uuid.UUID][]*link.Link)
	}

	// Merge results
	result := make(map[uuid.UUID][]graph.Edge, len(nodeIDs))
	for _, nodeID := range nodeIDs {
		edges := make([]graph.Edge, 0)

		// Outgoing connections
		if outgoing, ok := outgoingMap[nodeID]; ok {
			for _, ln := range outgoing {
				edges = append(edges, graph.Edge{
					From:   ln.SourceNoteID(),
					To:     ln.TargetNoteID(),
					Weight: ln.Weight().Value(),
				})
			}
		}

		// Incoming connections (reverse direction)
		if incoming, ok := incomingMap[nodeID]; ok {
			for _, ln := range incoming {
				edges = append(edges, graph.Edge{
					From:   ln.TargetNoteID(),
					To:     ln.SourceNoteID(),
					Weight: ln.Weight().Value(),
				})
			}
		}

		result[nodeID] = edges
	}

	return result, nil
}
