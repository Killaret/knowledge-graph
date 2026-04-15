package graphhandler

import (
	"context"
	"log"
	"strconv"

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
	maxDepth int // максимальная глубина загрузки графа
}

func New(noteRepo note.Repository, linkRepo link.Repository, maxDepth int) *Handler {
	return &Handler{
		noteRepo: noteRepo,
		linkRepo: linkRepo,
		maxDepth: maxDepth,
	}
}

func (h *Handler) GetGraph(c *gin.Context) {
	idStr := c.Param("id")
	centerID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	// Парсим параметр depth из query (с ограничением по maxDepth)
	depth := h.maxDepth
	if d := c.Query("depth"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			if parsed > h.maxDepth {
				depth = h.maxDepth
			} else {
				depth = parsed
			}
		}
	}

	// BFS для загрузки графа с заданной глубиной
	nodes, links := h.loadGraphBFS(c.Request.Context(), centerID, depth)

	c.JSON(200, GraphData{Nodes: nodes, Links: links})
}

// loadGraphBFS загружает граф с помощью BFS до заданной глубины
func (h *Handler) loadGraphBFS(ctx context.Context, centerID uuid.UUID, maxDepth int) ([]GraphNode, []GraphLink) {
	nodeMap := make(map[uuid.UUID]bool)
	linkMap := make(map[string]GraphLink) // ключ: "source->target"

	// BFS очередь: [noteID, depth]
	type queueItem struct {
		id    uuid.UUID
		depth int
	}
	queue := []queueItem{{id: centerID, depth: 0}}
	nodeMap[centerID] = true

	for len(queue) > 0 {
		// Извлекаем из очереди
		item := queue[0]
		queue = queue[1:]

		if item.depth >= maxDepth {
			continue
		}

		// Загружаем исходящие связи
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
					Source: l.SourceNoteID().String(),
					Target: targetID.String(),
					Weight: l.Weight().Value(),
				}
			}
			if !nodeMap[targetID] {
				nodeMap[targetID] = true
				queue = append(queue, queueItem{id: targetID, depth: item.depth + 1})
			}
		}

		// Загружаем входящие связи
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
					Source: sourceID.String(),
					Target: l.TargetNoteID().String(),
					Weight: l.Weight().Value(),
				}
			}
			if !nodeMap[sourceID] {
				nodeMap[sourceID] = true
				queue = append(queue, queueItem{id: sourceID, depth: item.depth + 1})
			}
		}
	}

	// Формируем список узлов
	nodes := make([]GraphNode, 0, len(nodeMap))
	for id := range nodeMap {
		n, err := h.noteRepo.FindByID(ctx, id)
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

	// Формируем список связей
	links := make([]GraphLink, 0, len(linkMap))
	for _, link := range linkMap {
		links = append(links, link)
	}

	return nodes, links
}

// GetFullGraph возвращает полный граф всех заметок и связей
func (h *Handler) GetFullGraph(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем лимит из query параметра (по умолчанию 0 - все заметки)
	limit := 0
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	// Загружаем все заметки
	notes, err := h.noteRepo.FindAll(ctx)
	if err != nil {
		log.Printf("Error loading all notes: %v", err)
		c.JSON(500, gin.H{"error": "failed to load notes"})
		return
	}

	// Применяем лимит если задан
	if limit > 0 && len(notes) > limit {
		notes = notes[:limit]
		log.Printf("[GetFullGraph] Limited to %d notes (total available: more)", limit)
	}

	// Загружаем все связи
	links, err := h.linkRepo.FindAll(ctx)
	if err != nil {
		log.Printf("Error loading all links: %v", err)
		c.JSON(500, gin.H{"error": "failed to load links"})
		return
	}

	// Формируем ответ
	nodes := make([]GraphNode, 0, len(notes))
	for _, n := range notes {
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

	graphLinks := make([]GraphLink, 0, len(links))
	for _, l := range links {
		graphLinks = append(graphLinks, GraphLink{
			Source: l.SourceNoteID().String(),
			Target: l.TargetNoteID().String(),
			Weight: l.Weight().Value(),
		})
	}

	c.JSON(200, GraphData{Nodes: nodes, Links: graphLinks})
}
