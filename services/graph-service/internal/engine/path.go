package engine

import (
	"context"

	"knowledge-graph-graph-service/internal/db"
)

// GetPath returns the shortest directed path between two notes using the
// materialized transitive closure. If no directed path exists, it attempts the
// reverse direction and returns the reversed path.
func GetPath(ctx context.Context, client db.PostgresClient, filter db.NotesFilter, fromID, toID string) (*AnalyticsPath, error) {
	noteIDs, distance, weight, err := client.GetShortestPath(ctx, filter, fromID, toID)
	if err != nil {
		return nil, err
	}

	return &AnalyticsPath{
		NoteIDs:  noteIDs,
		Distance: distance,
		Weight:   weight,
	}, nil
}
