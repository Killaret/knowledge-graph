package engine

import (
	"context"
	"testing"

	"knowledge-graph-graph-service/internal/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// depthCapturingClient records the depth Neighbors asks the DB for.
type depthCapturingClient struct {
	depth int
}

func (c *depthCapturingClient) GetNotes(ctx context.Context, filter db.NotesFilter) ([]*db.Note, []*db.Link, error) {
	return nil, nil, nil
}

func (c *depthCapturingClient) GetEmbeddings(ctx context.Context, noteIDs []string) (map[string][]float32, error) {
	return nil, nil
}

func (c *depthCapturingClient) GetNoteNeighbors(ctx context.Context, filter db.NotesFilter, noteID string, depth int) ([]*db.Neighbor, error) {
	c.depth = depth
	return nil, nil
}

func (c *depthCapturingClient) GetShortestPath(ctx context.Context, filter db.NotesFilter, fromID, toID string) ([]string, int, float64, error) {
	return nil, 0, 0, nil
}

func (c *depthCapturingClient) GetRecommendationCandidates(ctx context.Context, filter db.NotesFilter, noteID string, depth, limit int) ([]*db.RecommendationCandidate, error) {
	return nil, nil
}

func (c *depthCapturingClient) RefreshClosureView(ctx context.Context) error {
	return nil
}

// The closure view stores distances only up to maxRecDepth; asking deeper
// must be clamped instead of silently returning a truncated neighbourhood.
func TestNeighbors_DepthClamp(t *testing.T) {
	tests := []struct {
		name      string
		depth     int
		wantDepth int
	}{
		{"zero depth falls back to 2", 0, 2},
		{"negative depth falls back to 2", -1, 2},
		{"normal depth passes through", 3, 3},
		{"depth at the cap passes through", maxRecDepth, maxRecDepth},
		{"depth beyond the cap is clamped", maxRecDepth + 1, maxRecDepth},
		{"very deep request is clamped", 100, maxRecDepth},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &depthCapturingClient{}
			_, err := Neighbors(context.Background(), client, db.NotesFilter{}, "n1", tt.depth)
			require.NoError(t, err)
			assert.Equal(t, tt.wantDepth, client.depth)
		})
	}
}
