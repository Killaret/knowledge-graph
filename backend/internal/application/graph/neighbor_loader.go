package graph

import (
	"context"

	"knowledge-graph/internal/domain/graph"
	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
)

// neighborLoader is a struct that implements the graph.NeighborLoader interface.
// It knows how to load a note's neighbors from the database via repositories.
type neighborLoader struct {
	linkRepo link.Repository // repository for working with links
	noteRepo note.Repository // repository for notes (not used here, but may be useful)
}

// NewNeighborLoader is the constructor; it returns an implementation of the graph.NeighborLoader interface.
// It accepts the repositories (dependencies) needed to load data.
func NewNeighborLoader(linkRepo link.Repository, noteRepo note.Repository) graph.NeighborLoader {
	return &neighborLoader{
		linkRepo: linkRepo,
		noteRepo: noteRepo,
	}
}

// linkRepositoryWithBatch is an interface that extends link.Repository with batch methods
type linkRepositoryWithBatch interface {
	link.Repository
	FindBySourceIDs(ctx context.Context, sourceIDs []uuid.UUID) (map[uuid.UUID][]*link.Link, error)
	FindByTargetIDs(ctx context.Context, targetIDs []uuid.UUID) (map[uuid.UUID][]*link.Link, error)
}

// GetNeighbors returns, for a given note (nodeID), the list of edges (links) to neighboring notes.
// For recommendations we treat the graph as undirected: both outgoing and incoming links are considered.
// Each edge contains:
//   - From: ID of the source note (always nodeID, but for incoming links we reverse the direction)
//   - To:   ID of the neighboring note
//   - Weight: the link weight (already computed and stored in the database)
func (l *neighborLoader) GetNeighbors(ctx context.Context, nodeID uuid.UUID) ([]graph.Edge, error) {
	// 1. Get outgoing links: where our note is the source
	outgoing, err := l.linkRepo.FindBySource(ctx, nodeID)
	if err != nil {
		// On a database query error, return it so the upper layer can handle it
		return nil, err
	}

	// 2. Get incoming links: where our note is the target
	incoming, err := l.linkRepo.FindByTarget(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	// 3. Prepare a slice for the edges (approximate size = outgoing + incoming)
	edges := make([]graph.Edge, 0, len(outgoing)+len(incoming))

	// 4. Add outgoing links as is (direction from nodeID to target)
	for _, ln := range outgoing {
		edges = append(edges, graph.Edge{
			From:   ln.SourceNoteID(), // this is nodeID
			To:     ln.TargetNoteID(),
			Weight: ln.Weight().Value(), // the weight is stored in the Weight value object; take the float64
		})
	}

	// 5. Add incoming links but reverse the direction (so the graph is undirected)
	//    For recommendations it does not matter who references whom — only the link itself matters.
	for _, ln := range incoming {
		edges = append(edges, graph.Edge{
			From:   ln.TargetNoteID(), // reversed: from the target note to the source
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

	// Try to assert linkRepo to the interface with batch methods
	batchRepo, ok := l.linkRepo.(linkRepositoryWithBatch)
	if !ok {
		// If batch methods are unavailable, fall back to sequential queries
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

	// Batch queries for outgoing and incoming links
	outgoingMap, err := batchRepo.FindBySourceIDs(ctx, nodeIDs)
	if err != nil {
		outgoingMap = make(map[uuid.UUID][]*link.Link)
	}

	incomingMap, err := batchRepo.FindByTargetIDs(ctx, nodeIDs)
	if err != nil {
		incomingMap = make(map[uuid.UUID][]*link.Link)
	}

	// Merge the results
	result := make(map[uuid.UUID][]graph.Edge, len(nodeIDs))
	for _, nodeID := range nodeIDs {
		edges := make([]graph.Edge, 0)

		// Outgoing links
		if outgoing, ok := outgoingMap[nodeID]; ok {
			for _, ln := range outgoing {
				edges = append(edges, graph.Edge{
					From:   ln.SourceNoteID(),
					To:     ln.TargetNoteID(),
					Weight: ln.Weight().Value(),
				})
			}
		}

		// Incoming links (reverse the direction)
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
