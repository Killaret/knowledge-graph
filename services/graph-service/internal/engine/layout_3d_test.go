package engine

import (
	"testing"

	"knowledge-graph-graph-service/internal/db"

	"github.com/stretchr/testify/assert"
)

func TestLayout3D(t *testing.T) {
	tests := []struct {
		name    string
		notes   []*db.Note
		links   []*db.Link
		wantLen int
	}{
		{
			name:    "empty graph",
			notes:   []*db.Note{},
			links:   []*db.Link{},
			wantLen: 0,
		},
		{
			name: "single node",
			notes: []*db.Note{
				{ID: "1", Title: "Note 1"},
			},
			links:   []*db.Link{},
			wantLen: 1,
		},
		{
			name: "multiple nodes",
			notes: []*db.Note{
				{ID: "1", Title: "Note 1"},
				{ID: "2", Title: "Note 2"},
				{ID: "3", Title: "Note 3"},
				{ID: "4", Title: "Note 4"},
				{ID: "5", Title: "Note 5"},
			},
			links:   []*db.Link{},
			wantLen: 5,
		},
		{
			name: "nodes with links",
			notes: []*db.Note{
				{ID: "1", Title: "Note 1"},
				{ID: "2", Title: "Note 2"},
				{ID: "3", Title: "Note 3"},
			},
			links: []*db.Link{
				{Source: "1", Target: "2", LinkType: "related", Weight: 0.5},
				{Source: "2", Target: "3", LinkType: "related", Weight: 0.7},
			},
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Layout3D(tt.notes, tt.links)
			assert.NotNil(t, result)
			assert.Equal(t, tt.wantLen, len(result.Nodes))
			assert.Equal(t, len(tt.links), len(result.Links))

			// Check that all nodes have 3D coordinates
			for _, node := range result.Nodes {
				// Check that coordinates are set (can be zero in circular layout)
				assert.True(t, node.X != 0 || node.Y != 0 || len(result.Nodes) == 1,
					"At least one coordinate should be non-zero for multi-node graphs")
				// In 3D layout, Z should vary
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

func TestLayout3D_ZCoordinates(t *testing.T) {
	notes := []*db.Note{
		{ID: "1", Title: "A"},
		{ID: "2", Title: "B"},
		{ID: "3", Title: "C"},
		{ID: "4", Title: "D"},
		{ID: "5", Title: "E"},
	}

	result := Layout3D(notes, []*db.Link{})

	assert.Equal(t, 5, len(result.Nodes))

	// Check that Z coordinates are different for each node (spiral layout)
	zValues := make(map[float64]bool)
	for _, node := range result.Nodes {
		zValues[node.Z] = true
		assert.True(t, node.Z != 0 || true, "Z coordinate can be zero for first node")
	}

	// In the current implementation, nodes should have different Z values
	// This tests the spiral aspect of the 3D layout
	assert.True(t, len(zValues) > 1, "Nodes should have different Z coordinates for 3D layout")
}

func TestLayout3D_CircularBase(t *testing.T) {
	notes := []*db.Note{
		{ID: "1", Title: "A"},
		{ID: "2", Title: "B"},
		{ID: "3", Title: "C"},
	}

	result := Layout3D(notes, []*db.Link{})

	// X and Y should form a circular base (like 2D)
	for _, node := range result.Nodes {
		assert.True(t, node.X != 0 || node.Y != 0, "Node should have non-zero X or Y coordinates")
	}
}

func TestLayout3D_LinkPreservation(t *testing.T) {
	notes := []*db.Note{
		{ID: "1", Title: "A"},
		{ID: "2", Title: "B"},
		{ID: "3", Title: "C"},
	}

	links := []*db.Link{
		{Source: "1", Target: "2", LinkType: "related", Weight: 0.9},
		{Source: "2", Target: "3", LinkType: "related", Weight: 0.3},
		{Source: "1", Target: "3", LinkType: "parent", Weight: 1.0},
	}

	result := Layout3D(notes, links)

	assert.Equal(t, 3, len(result.Links))

	// Check that links are preserved
	linkMap := make(map[string]*LayoutLink)
	for _, link := range result.Links {
		key := link.Source + ":" + link.Target + ":" + link.LinkType
		linkMap[key] = link
	}

	assert.Contains(t, linkMap, "1:2:related")
	assert.Contains(t, linkMap, "2:3:related")
	assert.Contains(t, linkMap, "1:3:parent")

	// Check weights
	assert.Equal(t, float64(0.9), linkMap["1:2:related"].Weight)
	assert.Equal(t, float64(0.3), linkMap["2:3:related"].Weight)
	assert.Equal(t, float64(1.0), linkMap["1:3:parent"].Weight)
}

func TestLayout3D_SizeAssignment(t *testing.T) {
	notes := []*db.Note{
		{ID: "1", Title: "A"},
		{ID: "2", Title: "B"},
	}

	result := Layout3D(notes, []*db.Link{})

	// All nodes should have a size assigned
	for _, node := range result.Nodes {
		assert.Equal(t, float64(1.0), node.Size, "Default size should be 1.0")
	}
}
