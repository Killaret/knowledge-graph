package graphhandler

import (
	"context"
	"log"
	"strconv"

	"knowledge-graph/internal/config"
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
	Source   string  `json:"source"`
	Target   string  `json:"target"`
	Weight   float64 `json:"weight"`
	LinkType string  `json:"link_type"` // "reference", "dependency", "related", "custom"
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
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	// Парсим параметр depth из query (с ограничением по maxDepth)
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

// GetFullGraph возвращает полный граф всех заметок и связей с пагинацией
// Query params:
//   - limit: максимальное количество заметок (default: cfg.GraphDefaultLimit, max: cfg.GraphMaxLimit)
//   - offset: смещение для пагинации (default: 0)
//   - link_limit: максимальное количество связей (default: cfg.GraphLinkDefaultLimit, max: cfg.GraphLinkMaxLimit)
func (h *Handler) GetFullGraph(c *gin.Context) {
	ctx := c.Request.Context()

	// Парсим параметры пагинации для заметок
	limit := h.cfg.GraphDefaultLimit // default from config
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed >= 0 {
			// Ограничиваем максимум для защиты
			if parsed > h.cfg.GraphMaxLimit {
				limit = h.cfg.GraphMaxLimit
			} else if parsed == 0 {
				limit = 0 // все записи
			} else {
				limit = parsed
			}
		}
	}

	offset := 0 // default
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Парсим параметры пагинации для связей
	linkLimit := h.cfg.GraphLinkDefaultLimit // default from config
	if linkLimitStr := c.Query("link_limit"); linkLimitStr != "" {
		if parsed, err := strconv.Atoi(linkLimitStr); err == nil && parsed >= 0 {
			if parsed > h.cfg.GraphLinkMaxLimit {
				linkLimit = h.cfg.GraphLinkMaxLimit
			} else if parsed == 0 {
				linkLimit = 0 // все записи
			} else {
				linkLimit = parsed
			}
		}
	}

	linkOffset := 0 // default
	if linkOffsetStr := c.Query("link_offset"); linkOffsetStr != "" {
		if parsed, err := strconv.Atoi(linkOffsetStr); err == nil && parsed >= 0 {
			linkOffset = parsed
		}
	}

	// Загружаем заметки с пагинацией на уровне БД
	notes, totalNotes, err := h.noteRepo.FindAllPaginated(ctx, limit, offset)
	if err != nil {
		log.Printf("Error loading notes: %v", err)
		c.JSON(500, gin.H{"error": "failed to load notes"})
		return
	}

	// Загружаем связи с пагинацией на уровне БД
	links, totalLinks, err := h.linkRepo.FindAllPaginated(ctx, linkLimit, linkOffset)
	if err != nil {
		log.Printf("Error loading links: %v", err)
		c.JSON(500, gin.H{"error": "failed to load links"})
		return
	}

	// Формируем ответ
	nodes := make([]GraphNode, 0, len(notes))
	debugTypes := make(map[string]int)

	// Все возможные типы узлов для разнообразия
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
		// Если тип не задан в metadata, генерируем на основе индекса для разнообразия
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
	for _, l := range links {
		graphLinks = append(graphLinks, GraphLink{
			Source:   l.SourceNoteID().String(),
			Target:   l.TargetNoteID().String(),
			Weight:   l.Weight().Value(),
			LinkType: l.LinkType().String(),
		})
	}

	// Возвращаем с метаданными пагинации
	c.JSON(200, gin.H{
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
