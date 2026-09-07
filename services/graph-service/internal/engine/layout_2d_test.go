package engine

import (
	"testing"

	"knowledge-graph-graph-service/internal/db"

	"github.com/stretchr/testify/assert"
)

func TestLayout2D(t *testing.T) {
	tests := []struct {
		name    string
		notes   []*db.Note
		links   []*db.Link
		rootID  string
		wantLen int
	}{
		{
			name:    "empty graph",
			notes:   []*db.Note{},
			links:   []*db.Link{},
			rootID:  "",
			wantLen: 0,
		},
		{
			name: "single node",
			notes: []*db.Note{
				{ID: "1", Title: "Note 1"},
			},
			links:   []*db.Link{},
			rootID:  "",
			wantLen: 1,
		},
		{
			name: "multiple nodes",
			notes: []*db.Note{
				{ID: "1", Title: "Note 1"},
				{ID: "2", Title: "Note 2"},
				{ID: "3", Title: "Note 3"},
			},
			links:   []*db.Link{},
			rootID:  "",
			wantLen: 3,
		},
		{
			name: "nodes with links",
			notes: []*db.Note{
				{ID: "1", Title: "Note 1"},
				{ID: "2", Title: "Note 2"},
			},
			links: []*db.Link{
				{Source: "1", Target: "2", LinkType: "related", Weight: 0.5},
			},
			rootID:  "",
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Layout2D(tt.notes, tt.links, tt.rootID)
			assert.NotNil(t, result)
			assert.Equal(t, tt.wantLen, len(result.Nodes))
			assert.Equal(t, len(tt.links), len(result.Links))

			// Check that all nodes have coordinates
			for _, node := range result.Nodes {
				// Check that coordinates are set (can be zero in circular layout)
				assert.True(t, node.X != 0 || node.Y != 0 || len(result.Nodes) == 1,
					"At least one coordinate should be non-zero for multi-node graphs")
				assert.Equal(t, 0.0, node.Z) // 2D layout should have Z=0
			}

			// Check that node IDs match
			resultIDs := make(map[string]bool)
			for _, node := range result.Nodes {
				resultIDs[node.ID] = true
			}
			for _, note := range tt.notes {
				assert.True(t, resultIDs[note.ID], "Note ID %s should be in result", note.ID)
			}
		})
	}
}

func TestLayout2D_CircularLayout(t *testing.T) {
	notes := []*db.Note{
		{ID: "1", Title: "A"},
		{ID: "2", Title: "B"},
		{ID: "3", Title: "C"},
		{ID: "4", Title: "D"},
	}

	result := Layout2D(notes, []*db.Link{}, "")

	// Check that nodes are arranged in a circle
	assert.Equal(t, 4, len(result.Nodes))

	// All nodes should be at the same distance from origin (approximately)
	// This is a basic check for circular layout
	// More sophisticated tests would check actual coordinates
	for _, node := range result.Nodes {
		assert.True(t, node.X != 0 || node.Y != 0, "Node should have non-zero coordinates")
	}
}

func TestLayout2D_RootNode(t *testing.T) {
	notes := []*db.Note{
		{ID: "1", Title: "Root"},
		{ID: "2", Title: "Child 1"},
		{ID: "3", Title: "Child 2"},
	}

	result := Layout2D(notes, []*db.Link{}, "1")

	assert.Equal(t, 3, len(result.Nodes))

	// Root node should be in the result
	rootFound := false
	for _, node := range result.Nodes {
		if node.ID == "1" {
			rootFound = true
			break
		}
	}
	assert.True(t, rootFound, "Root node should be in the layout")
}

func TestLayout2D_LinkPreservation(t *testing.T) {
	notes := []*db.Note{
		{ID: "1", Title: "A"},
		{ID: "2", Title: "B"},
		{ID: "3", Title: "C"},
	}

	links := []*db.Link{
		{Source: "1", Target: "2", LinkType: "related", Weight: 0.8},
		{Source: "2", Target: "3", LinkType: "related", Weight: 0.5},
	}

	result := Layout2D(notes, links, "")

	assert.Equal(t, 2, len(result.Links))

	// Check that links are preserved
	linkMap := make(map[string]*LayoutLink)
	for _, link := range result.Links {
		key := link.Source + ":" + link.Target + ":" + link.LinkType
		linkMap[key] = link
	}

	assert.Contains(t, linkMap, "1:2:related")
	assert.Contains(t, linkMap, "2:3:related")

	// Check weights
	assert.Equal(t, float64(0.8), linkMap["1:2:related"].Weight)
	assert.Equal(t, float64(0.5), linkMap["2:3:related"].Weight)
}
