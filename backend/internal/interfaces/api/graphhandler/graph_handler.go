package graphhandler

import (
	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GraphNode struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"` // "star", "planet", "comet", "galaxy"
}

type GraphLink struct {
	Source string  `json:"source"`
	Target string  `json:"target"`
	Weight float64 `json:"weight"`
}

type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Links []GraphLink `json:"links"`
}

type Handler struct {
	noteRepo note.Repository
	linkRepo link.Repository
}

func New(noteRepo note.Repository, linkRepo link.Repository) *Handler {
	return &Handler{noteRepo: noteRepo, linkRepo: linkRepo}
}

func (h *Handler) GetGraph(c *gin.Context) {
	idStr := c.Param("id")
	centerID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	outgoing, err := h.linkRepo.FindBySource(c.Request.Context(), centerID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	incoming, err := h.linkRepo.FindByTarget(c.Request.Context(), centerID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	nodeMap := make(map[uuid.UUID]bool)
	nodeMap[centerID] = true
	for _, l := range outgoing {
		nodeMap[l.TargetNoteID()] = true
	}
	for _, l := range incoming {
		nodeMap[l.SourceNoteID()] = true
	}

	nodes := make([]GraphNode, 0, len(nodeMap))
	for id := range nodeMap {
		n, err := h.noteRepo.FindByID(c.Request.Context(), id)
		if err != nil || n == nil {
			continue
		}
		// Определяем тип небесного тела
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

	links := make([]GraphLink, 0, len(outgoing)+len(incoming))
	for _, l := range outgoing {
		links = append(links, GraphLink{
			Source: l.SourceNoteID().String(),
			Target: l.TargetNoteID().String(),
			Weight: l.Weight().Value(),
		})
	}
	for _, l := range incoming {
		links = append(links, GraphLink{
			Source: l.SourceNoteID().String(),
			Target: l.TargetNoteID().String(),
			Weight: l.Weight().Value(),
		})
	}

	c.JSON(200, GraphData{Nodes: nodes, Links: links})
}
