package graphhandler

import (
	"context"
	"log"
	"strconv"

	"knowledge-graph/internal/application/cache"
	"knowledge-graph/internal/config"
	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"
	apicommon "knowledge-graph/internal/interfaces/api/common"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GraphNode struct {
	ID    string  `json:"id"`
	Title string  `json:"title"`
	Type  string  `json:"type"`
	X     float64 `json:"x,omitempty"`
	Y     float64 `json:"y,omitempty"`
}

type GraphLink struct {
	Source   string  `json:"source"`
	Target   string  `json:"target"`
	Weight   float64 `json:"weight"`
	LinkType string  `json:"link_type"`
}

type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Links []GraphLink `json:"links"`
}

type Handler struct {
	noteRepo   note.Repository
	linkRepo   link.Repository
	cfg        *config.Config
	graphCache *cache.GraphCache
}

func New(noteRepo note.Repository, linkRepo link.Repository, cfg *config.Config, graphCache *cache.GraphCache) *Handler {
	return &Handler{
		noteRepo:   noteRepo,
		linkRepo:   linkRepo,
		cfg:        cfg,
		graphCache: graphCache,
	}
}

func (h *Handler) GetGraph(c *gin.Context) {
	idStr := c.Param("id")
	centerID, err := uuid.Parse(idStr)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("id", apicommon.ReasonInvalidFormat, apicommon.MsgInvalidUUID, idStr),
		})
		return
	}

	depth := h.cfg.GraphLoadDepth
	if d := c.Query("depth"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			if parsed > h.cfg.GraphLoadDepth {
				depth = h.cfg.GraphLoadDepth
			} else {
				depth = parsed
			}
		}
	}

	nodes, links := h.loadGraphBFS(c.Request.Context(), centerID, depth)
	apicommon.JSON(c, 200, GraphData{Nodes: nodes, Links: links})
}

func (h *Handler) loadGraphBFS(ctx context.Context, centerID uuid.UUID, maxDepth int) ([]GraphNode, []GraphLink) {
	nodeMap := make(map[uuid.UUID]bool)
	linkMap := make(map[string]GraphLink)

	type queueItem struct {
		id    uuid.UUID
		depth int
	}
	queue := []queueItem{{id: centerID, depth: 0}}
	nodeMap[centerID] = true

	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]

		if item.depth >= maxDepth {
			continue
		}

		outgoing, err := h.linkRepo.FindBySource(ctx, item.id)
		if err != nil {
			log.Printf("Error finding outgoing links for %s: %v", item.id, err)
			continue
		}
		for _, l := range outgoing {
			targetID := l.TargetNoteID()
			linkKey := l.SourceNoteID().String() + "->" + targetID.String()
			if _, exists := linkMap[linkKey]; !exists {
				linkMap[linkKey] = GraphLink{
					Source:   l.SourceNoteID().String(),
					Target:   targetID.String(),
					Weight:   l.Weight().Value(),
					LinkType: l.LinkType().String(),
				}
			}
			if !nodeMap[targetID] {
				nodeMap[targetID] = true
				queue = append(queue, queueItem{id: targetID, depth: item.depth + 1})
			}
		}

		incoming, err := h.linkRepo.FindByTarget(ctx, item.id)
		if err != nil {
			log.Printf("Error finding incoming links for %s: %v", item.id, err)
			continue
		}
		for _, l := range incoming {
			sourceID := l.SourceNoteID()
			linkKey := sourceID.String() + "->" + l.TargetNoteID().String()
			if _, exists := linkMap[linkKey]; !exists {
				linkMap[linkKey] = GraphLink{
					Source:   sourceID.String(),
					Target:   l.TargetNoteID().String(),
					Weight:   l.Weight().Value(),
					LinkType: l.LinkType().String(),
				}
			}
			if !nodeMap[sourceID] {
				nodeMap[sourceID] = true
				queue = append(queue, queueItem{id: sourceID, depth: item.depth + 1})
			}
		}
	}

	nodes := make([]GraphNode, 0, len(nodeMap))
	for id := range nodeMap {
		n, err := h.noteRepo.FindByID(ctx, id)
		if err != nil || n == nil {
			continue
		}
		nodeType := "star"
		if metadata := n.Metadata().Value(); metadata != nil {
			if t, ok := metadata["type"]; ok {
				if ts, ok := t.(string); ok {
					nodeType = ts
				}
			}
		}
		nodes = append(nodes, GraphNode{
			ID:    n.ID().String(),
			Title: n.Title().String(),
			Type:  nodeType,
		})
	}

	links := make([]GraphLink, 0, len(linkMap))
	for _, link := range linkMap {
		links = append(links, link)
	}

	return nodes, links
}

func (h *Handler) GetFullGraph(c *gin.Context) {
	ctx := c.Request.Context()

	limit := h.cfg.GraphDefaultLimit
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed >= 0 {
			if parsed > h.cfg.GraphMaxLimit {
				limit = h.cfg.GraphMaxLimit
			} else if parsed == 0 {
				limit = 0
			} else {
				limit = parsed
			}
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	linkLimit := h.cfg.GraphLinkDefaultLimit
	if linkLimitStr := c.Query("link_limit"); linkLimitStr != "" {
		if parsed, err := strconv.Atoi(linkLimitStr); err == nil && parsed >= 0 {
			if parsed > h.cfg.GraphLinkMaxLimit {
				linkLimit = h.cfg.GraphLinkMaxLimit
			} else if parsed == 0 {
				linkLimit = 0
			} else {
				linkLimit = parsed
			}
		}
	}

	linkOffset := 0
	if linkOffsetStr := c.Query("link_offset"); linkOffsetStr != "" {
		if parsed, err := strconv.Atoi(linkOffsetStr); err == nil && parsed >= 0 {
			linkOffset = parsed
		}
	}

	notes, totalNotes, err := h.noteRepo.FindAllPaginated(ctx, limit, offset)
	if err != nil {
		log.Printf("Error loading notes: %v", err)
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedLoadGraph)
		return
	}

	links, totalLinks, err := h.linkRepo.FindAllPaginated(ctx, linkLimit, linkOffset)
	if err != nil {
		log.Printf("Error loading links: %v", err)
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedLoadGraph)
		return
	}

	nodes := make([]GraphNode, 0, len(notes))
	debugTypes := make(map[string]int)
	celestialTypes := []string{"star", "planet", "moon", "asteroid", "nebula", "satellite", "comet", "blackhole", "galaxy"}

	for i, n := range notes {
		nodeType := "star"
		hasTypeFromMetadata := false
		if metadata := n.Metadata().Value(); metadata != nil {
			if t, ok := metadata["type"]; ok {
				if ts, ok := t.(string); ok && ts != "" {
					nodeType = ts
					hasTypeFromMetadata = true
				}
			}
		}
		if !hasTypeFromMetadata {
			nodeType = celestialTypes[i%len(celestialTypes)]
		}
		debugTypes[nodeType]++
		nodes = append(nodes, GraphNode{
			ID:    n.ID().String(),
			Title: n.Title().String(),
			Type:  nodeType,
		})
	}
	log.Printf("[GraphHandler] Node types distribution: %v", debugTypes)

	graphLinks := make([]GraphLink, 0, len(links))
	log.Printf("[GraphHandler] Raw links from DB: %d, totalLinks: %d", len(links), totalLinks)
	for i, l := range links {
		if i < 3 {
			log.Printf("[GraphHandler] Link %d: Source=%s, Target=%s, Type=%s", i, l.SourceNoteID().String(), l.TargetNoteID().String(), l.LinkType().String())
		}
		graphLinks = append(graphLinks, GraphLink{
			Source:   l.SourceNoteID().String(),
			Target:   l.TargetNoteID().String(),
			Weight:   l.Weight().Value(),
			LinkType: l.LinkType().String(),
		})
	}
	log.Printf("[GraphHandler] Converted graphLinks: %d", len(graphLinks))

	apicommon.JSON(c, 200, gin.H{
		"nodes": nodes,
		"links": graphLinks,
		"pagination": gin.H{
			"notes": gin.H{
				"total":  totalNotes,
				"limit":  limit,
				"offset": offset,
			},
			"links": gin.H{
				"total":  totalLinks,
				"limit":  linkLimit,
				"offset": linkOffset,
			},
		},
	})
}

// GetCachedGraph returns the user's cached graph data
func (h *Handler) GetCachedGraph(c *gin.Context) {
	ctx := c.Request.Context()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		apicommon.Unauthorized(c)
		return
	}

	if h.graphCache == nil {
		apicommon.InternalErrorWithMessage(c, "cache not available")
		return
	}

	cachedData, found, err := h.graphCache.GetCachedUserGraph(ctx, userID.String())
	if err != nil {
		log.Printf("Error getting cached graph: %v", err)
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedLoadGraph)
		return
	}

	if !found {
		c.Status(204)
		return
	}

	apicommon.JSON(c, 200, convertFromCacheGraphData(cachedData))
}

// GetFreshGraph returns fresh graph data with optional delta
func (h *Handler) GetFreshGraph(c *gin.Context) {
	ctx := c.Request.Context()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		apicommon.Unauthorized(c)
		return
	}

	// Get fresh data from database
	freshData, err := h.loadFullGraph(ctx)
	if err != nil {
		log.Printf("Error loading fresh graph: %v", err)
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedLoadGraph)
		return
	}

	// Try to get cached data for delta calculation
	var delta *GraphDelta
	if h.graphCache != nil {
		cachedData, found, err := h.graphCache.GetCachedUserGraph(ctx, userID.String())
		if err == nil && found {
			// Convert cache.GraphData to handler GraphData
			handlerCachedData := convertFromCacheGraphData(cachedData)
			delta = calculateDelta(handlerCachedData, freshData)
			// Preserve cached positions in fresh data for visual stability
			freshData = h.preserveCachedPositions(freshData, handlerCachedData)
		}

		// Cache the fresh data
		if err := h.graphCache.CacheUserGraph(ctx, userID.String(), convertToCacheGraphData(freshData)); err != nil {
			log.Printf("Warning: failed to cache graph: %v", err)
		}
	}

	response := gin.H{
		"fresh": freshData,
	}

	if delta != nil {
		response["delta"] = delta
	}

	apicommon.JSON(c, 200, response)
}

// loadFullGraph loads the full graph data from database
func (h *Handler) loadFullGraph(ctx context.Context) (GraphData, error) {
	limit := h.cfg.GraphDefaultLimit
	linkLimit := h.cfg.GraphLinkDefaultLimit

	notes, _, err := h.noteRepo.FindAllPaginated(ctx, limit, 0)
	if err != nil {
		return GraphData{}, err
	}

	links, _, err := h.linkRepo.FindAllPaginated(ctx, linkLimit, 0)
	if err != nil {
		return GraphData{}, err
	}

	nodes := make([]GraphNode, 0, len(notes))
	celestialTypes := []string{"star", "planet", "moon", "asteroid", "nebula", "satellite", "comet", "blackhole", "galaxy"}

	for i, n := range notes {
		nodeType := "star"
		hasTypeFromMetadata := false
		if metadata := n.Metadata().Value(); metadata != nil {
			if t, ok := metadata["type"]; ok {
				if ts, ok := t.(string); ok && ts != "" {
					nodeType = ts
					hasTypeFromMetadata = true
				}
			}
		}
		if !hasTypeFromMetadata {
			nodeType = celestialTypes[i%len(celestialTypes)]
		}
		nodes = append(nodes, GraphNode{
			ID:    n.ID().String(),
			Title: n.Title().String(),
			Type:  nodeType,
		})
	}

	graphLinks := make([]GraphLink, 0, len(links))
	for _, l := range links {
		graphLinks = append(graphLinks, GraphLink{
			Source:   l.SourceNoteID().String(),
			Target:   l.TargetNoteID().String(),
			Weight:   l.Weight().Value(),
			LinkType: l.LinkType().String(),
		})
	}

	return GraphData{Nodes: nodes, Links: graphLinks}, nil
}

// GraphDelta represents changes between cached and fresh graph
type GraphDelta struct {
	AddedNodes   []GraphNode `json:"added_nodes,omitempty"`
	RemovedNodes []string    `json:"removed_nodes,omitempty"`
	UpdatedNodes []GraphNode `json:"updated_nodes,omitempty"`
	AddedLinks   []GraphLink `json:"added_links,omitempty"`
	RemovedLinks []GraphLink `json:"removed_links,omitempty"`
}

// calculateDelta computes the difference between cached and fresh graph
func calculateDelta(cached, fresh GraphData) *GraphDelta {
	delta := &GraphDelta{}

	cachedNodeMap := make(map[string]GraphNode)
	for _, node := range cached.Nodes {
		cachedNodeMap[node.ID] = node
	}

	freshNodeMap := make(map[string]GraphNode)
	for _, node := range fresh.Nodes {
		freshNodeMap[node.ID] = node
	}

	// Find added nodes
	for id, freshNode := range freshNodeMap {
		if _, exists := cachedNodeMap[id]; !exists {
			delta.AddedNodes = append(delta.AddedNodes, freshNode)
		}
	}

	// Find removed nodes
	for id := range cachedNodeMap {
		if _, exists := freshNodeMap[id]; !exists {
			delta.RemovedNodes = append(delta.RemovedNodes, id)
		}
	}

	// Find updated nodes (title or type changed)
	for id, freshNode := range freshNodeMap {
		if cachedNode, exists := cachedNodeMap[id]; exists {
			if cachedNode.Title != freshNode.Title || cachedNode.Type != freshNode.Type {
				delta.UpdatedNodes = append(delta.UpdatedNodes, freshNode)
			}
		}
	}

	cachedLinkMap := make(map[string]GraphLink)
	for _, link := range cached.Links {
		key := link.Source + "->" + link.Target
		cachedLinkMap[key] = link
	}

	freshLinkMap := make(map[string]GraphLink)
	for _, link := range fresh.Links {
		key := link.Source + "->" + link.Target
		freshLinkMap[key] = link
	}

	// Find added links
	for key, freshLink := range freshLinkMap {
		if _, exists := cachedLinkMap[key]; !exists {
			delta.AddedLinks = append(delta.AddedLinks, freshLink)
		}
	}

	// Find removed links
	for key, cachedLink := range cachedLinkMap {
		if _, exists := freshLinkMap[key]; !exists {
			delta.RemovedLinks = append(delta.RemovedLinks, cachedLink)
		}
	}

	// Return nil if no changes
	if len(delta.AddedNodes) == 0 && len(delta.RemovedNodes) == 0 &&
		len(delta.UpdatedNodes) == 0 && len(delta.AddedLinks) == 0 &&
		len(delta.RemovedLinks) == 0 {
		return nil
	}

	return delta
}

// convertToCacheGraphData converts handler GraphData to cache GraphData
func convertToCacheGraphData(data GraphData) cache.GraphData {
	cacheNodes := make([]cache.GraphNode, len(data.Nodes))
	for i, node := range data.Nodes {
		cacheNodes[i] = cache.GraphNode{
			ID:    node.ID,
			Title: node.Title,
			Type:  node.Type,
			X:     node.X,
			Y:     node.Y,
		}
	}

	cacheLinks := make([]cache.GraphLink, len(data.Links))
	for i, link := range data.Links {
		cacheLinks[i] = cache.GraphLink{
			Source:   link.Source,
			Target:   link.Target,
			Weight:   link.Weight,
			LinkType: link.LinkType,
		}
	}

	return cache.GraphData{
		Nodes: cacheNodes,
		Links: cacheLinks,
	}
}

// convertFromCacheGraphData converts cache GraphData to handler GraphData
func convertFromCacheGraphData(data cache.GraphData) GraphData {
	handlerNodes := make([]GraphNode, len(data.Nodes))
	for i, node := range data.Nodes {
		handlerNodes[i] = GraphNode{
			ID:    node.ID,
			Title: node.Title,
			Type:  node.Type,
			X:     node.X,
			Y:     node.Y,
		}
	}

	handlerLinks := make([]GraphLink, len(data.Links))
	for i, link := range data.Links {
		handlerLinks[i] = GraphLink{
			Source:   link.Source,
			Target:   link.Target,
			Weight:   link.Weight,
			LinkType: link.LinkType,
		}
	}

	return GraphData{
		Nodes: handlerNodes,
		Links: handlerLinks,
	}
}

// preserveCachedPositions preserves node positions from cached data in fresh data
func (h *Handler) preserveCachedPositions(fresh, cached GraphData) GraphData {
	cachedPositionMap := make(map[string]struct{ x, y float64 })
	for _, node := range cached.Nodes {
		cachedPositionMap[node.ID] = struct{ x, y float64 }{x: node.X, y: node.Y}
	}

	preservedNodes := make([]GraphNode, len(fresh.Nodes))
	for i, node := range fresh.Nodes {
		x, y := node.X, node.Y
		if pos, exists := cachedPositionMap[node.ID]; exists {
			x, y = pos.x, pos.y
		}
		preservedNodes[i] = GraphNode{
			ID:    node.ID,
			Title: node.Title,
			Type:  node.Type,
			X:     x,
			Y:     y,
		}
	}

	return GraphData{
		Nodes: preservedNodes,
		Links: fresh.Links,
	}
}
