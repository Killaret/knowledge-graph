package graphhandler

import (
	"context"
	"log"
	"strconv"

	"knowledge-graph/internal/config"
	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"
	apicommon "knowledge-graph/internal/interfaces/api/common"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GraphNode struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`
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
	noteRepo note.Repository
	linkRepo link.Repository
	cfg      *config.Config
}

func New(noteRepo note.Repository, linkRepo link.Repository, cfg *config.Config) *Handler {
	return &Handler{
		noteRepo: noteRepo,
		linkRepo: linkRepo,
		cfg:      cfg,
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
