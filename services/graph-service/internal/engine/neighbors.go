package engine

import (
	"context"

	"knowledge-graph-graph-service/internal/db"
)

// Neighbors returns notes reachable from noteID within the given depth.
// The graph is treated as undirected: both outgoing and incoming links are followed.
func Neighbors(ctx context.Context, client db.PostgresClient, filter db.NotesFilter, noteID string, depth int) ([]*AnalyticsNode, error) {
	if depth <= 0 {
		depth = 2
	}

	rows, err := client.GetNoteNeighbors(ctx, filter, noteID, depth)
	if err != nil {
		return nil, err
	}

	nodes := make([]*AnalyticsNode, 0, len(rows))
	for _, r := range rows {
		nodes = append(nodes, &AnalyticsNode{
			ID:       r.ID,
			Title:    r.Title,
			Type:     r.Type,
			Weight:   r.Weight,
			Distance: r.Distance,
		})
	}

	return nodes, nil
}
