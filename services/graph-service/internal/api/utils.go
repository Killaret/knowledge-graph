package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"knowledge-graph-graph-service/internal/engine"
)

// computeLayoutHash generates a hash for the layout
func computeLayoutHash(layout *engine.LayoutResponse) string {
	payload, _ := json.Marshal(layout)
	h := sha256.Sum256(payload)
	return hex.EncodeToString(h[:])
}

// Convert engine.LayoutNode to ProtoLayoutNode
func convertLayoutNode(node *engine.LayoutNode) *ProtoLayoutNode {
	return &ProtoLayoutNode{
		Id:    node.ID,
		Title: node.Title,
		Type:  node.Type,
		X:     node.X,
		Y:     node.Y,
		Z:     node.Z,
		Size:  node.Size,
	}
}

// Convert engine.LayoutLink to ProtoLayoutLink
func convertLayoutLink(link *engine.LayoutLink) *ProtoLayoutLink {
	return &ProtoLayoutLink{
		Source:   link.Source,
		Target:   link.Target,
		Weight:   link.Weight,
		LinkType: link.LinkType,
	}
}

// Convert []*engine.LayoutNode to []*ProtoLayoutNode
func convertLayoutNodes(nodes []*engine.LayoutNode) []*ProtoLayoutNode {
	result := make([]*ProtoLayoutNode, len(nodes))
	for i, node := range nodes {
		result[i] = convertLayoutNode(node)
	}
	return result
}

// Convert []*engine.LayoutLink to []*ProtoLayoutLink
func convertLayoutLinks(links []*engine.LayoutLink) []*ProtoLayoutLink {
	result := make([]*ProtoLayoutLink, len(links))
	for i, link := range links {
		result[i] = convertLayoutLink(link)
	}
	return result
}
