package engine

import (
	"math"

	"knowledge-graph-graph-service/internal/db"
)

// Layout3D computes a 3D helical layout for the given notes.
// Nodes are placed along a helix with radius Layout3DRadius and
// vertical step Layout3DZStep.
func Layout3D(notes []*db.Note, links []*db.Link) *LayoutResponse {
	nodes := make([]*LayoutNode, 0, len(notes))
	for i, note := range notes {
		theta := 2.0 * math.Pi * float64(i) / math.Max(1, float64(len(notes)))
		nodeType := note.Type
		if nodeType == "" {
			nodeType = NodeTypeNote
		}
		nodes = append(nodes, &LayoutNode{
			ID:    note.ID,
			Title: note.Title,
			Type:  nodeType,
			X:     math.Cos(theta) * Layout3DRadius,
			Y:     math.Sin(theta) * Layout3DRadius,
			Z:     float64(i) * Layout3DZStep,
			Size:  DefaultNodeSize,
		})
	}

	linksOut := make([]*LayoutLink, 0, len(links))
	for _, link := range links {
		linksOut = append(linksOut, &LayoutLink{
			Source:     link.Source,
			Target:     link.Target,
			Weight:     link.Weight,
			LinkType:   link.LinkType,
			SourceType: link.SourceType,
		})
	}

	return &LayoutResponse{Nodes: nodes, Links: linksOut}
}
