package engine

import (
	"math"

	"knowledge-graph-graph-service/internal/db"
)

// Layout2D computes a 2D circular layout for the given notes.
// Nodes are placed evenly around a circle of radius Layout2DRadius.
// rootID is currently used as a hint for future BFS-aware layouts.
func Layout2D(notes []*db.Note, links []*db.Link, rootID string) *LayoutResponse {
	nodes := make([]*LayoutNode, 0, len(notes))
	count := len(notes)
	for i, note := range notes {
		theta := 2.0 * math.Pi * float64(i) / math.Max(1, float64(count))
		nodes = append(nodes, &LayoutNode{
			ID:    note.ID,
			Title: note.Title,
			Type:  NodeTypeNote,
			X:     math.Cos(theta) * Layout2DRadius,
			Y:     math.Sin(theta) * Layout2DRadius,
			Z:     0,
			Size:  DefaultNodeSize,
		})
	}

	linksOut := make([]*LayoutLink, 0, len(links))
	for _, link := range links {
		linksOut = append(linksOut, &LayoutLink{
			Source:   link.Source,
			Target:   link.Target,
			Weight:   link.Weight,
			LinkType: link.LinkType,
		})
	}

	return &LayoutResponse{Nodes: nodes, Links: linksOut}
}
