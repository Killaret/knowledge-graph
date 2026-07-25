package engine

// Layout type/node constants used by both 2D and 3D engines.
const (
	NodeTypeNote = "note"

	DefaultNodeSize = 1.0

	// Layout2DRadius is the circle radius for the 2D layout.
	Layout2DRadius = 100.0

	// Layout3DRadius is the helix radius for the 3D layout.
	Layout3DRadius = 120.0

	// Layout3DZStep is the vertical step between successive nodes in 3D.
	Layout3DZStep = 5.0
)

type LayoutNode struct {
	ID    string  `json:"id"`
	Title string  `json:"title"`
	Type  string  `json:"type"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Z     float64 `json:"z"`
	Size  float64 `json:"size"`
}

type LayoutLink struct {
	Source   string  `json:"source"`
	Target   string  `json:"target"`
	Weight   float64 `json:"weight"`
	LinkType string  `json:"link_type"`
}

type LayoutResponse struct {
	Nodes []*LayoutNode `json:"nodes"`
	Links []*LayoutLink `json:"links"`
}

type DeltaResponse struct {
	AddedNodes   []*LayoutNode `json:"added_nodes,omitempty"`
	RemovedNodes []string      `json:"removed_nodes,omitempty"`
	UpdatedNodes []*LayoutNode `json:"updated_nodes,omitempty"`
	AddedLinks   []*LayoutLink `json:"added_links,omitempty"`
	RemovedLinks []*LayoutLink `json:"removed_links,omitempty"`
	CurrentHash  string        `json:"current_hash,omitempty"`
}

func ComputeDelta(oldLayout, newLayout *LayoutResponse) *DeltaResponse {
	if oldLayout == nil {
		return &DeltaResponse{AddedNodes: newLayout.Nodes, AddedLinks: newLayout.Links}
	}

	oldNodeMap := make(map[string]*LayoutNode, len(oldLayout.Nodes))
	for _, node := range oldLayout.Nodes {
		oldNodeMap[node.ID] = node
	}

	newNodeMap := make(map[string]*LayoutNode, len(newLayout.Nodes))
	for _, node := range newLayout.Nodes {
		newNodeMap[node.ID] = node
	}

	addedNodes := make([]*LayoutNode, 0)
	updatedNodes := make([]*LayoutNode, 0)
	removedNodes := make([]string, 0)

	for id, newNode := range newNodeMap {
		if oldNode, exists := oldNodeMap[id]; !exists {
			addedNodes = append(addedNodes, newNode)
		} else if oldNode.Title != newNode.Title || oldNode.X != newNode.X || oldNode.Y != newNode.Y || oldNode.Z != newNode.Z {
			updatedNodes = append(updatedNodes, newNode)
		}
	}
	for id := range oldNodeMap {
		if _, exists := newNodeMap[id]; !exists {
			removedNodes = append(removedNodes, id)
		}
	}

	oldLinkMap := make(map[string]*LayoutLink, len(oldLayout.Links))
	for _, link := range oldLayout.Links {
		oldLinkMap[fmtLinkKey(link)] = link
	}
	newLinkMap := make(map[string]*LayoutLink, len(newLayout.Links))
	for _, link := range newLayout.Links {
		newLinkMap[fmtLinkKey(link)] = link
	}

	addedLinks := make([]*LayoutLink, 0)
	removedLinks := make([]*LayoutLink, 0)
	for key, link := range newLinkMap {
		if _, exists := oldLinkMap[key]; !exists {
			addedLinks = append(addedLinks, link)
		}
	}
	for key := range oldLinkMap {
		if _, exists := newLinkMap[key]; !exists {
			removedLink := oldLinkMap[key]
			removedLinks = append(removedLinks, removedLink)
		}
	}

	return &DeltaResponse{
		AddedNodes:   addedNodes,
		RemovedNodes: removedNodes,
		UpdatedNodes: updatedNodes,
		AddedLinks:   addedLinks,
		RemovedLinks: removedLinks,
	}
}

func fmtLinkKey(link *LayoutLink) string {
	return link.Source + ":" + link.Target + ":" + link.LinkType
}
