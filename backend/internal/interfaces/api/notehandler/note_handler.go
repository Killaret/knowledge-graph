package notehandler

import (
	"knowledge-graph/internal/application/common"
	"knowledge-graph/internal/domain/note"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	repo      note.Repository
	taskQueue common.TaskQueue
}

func New(repo note.Repository, taskQueue common.TaskQueue) *Handler {
	return &Handler{repo: repo, taskQueue: taskQueue}
}

type createNoteRequest struct {
	Title    string                 `json:"title" binding:"required"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
}

func (h *Handler) Create(c *gin.Context) {
	var req createNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	title, err := note.NewTitle(req.Title)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	content, err := note.NewContent(req.Content)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	metadata, err := note.NewMetadata(req.Metadata)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	newNote := note.NewNote(title, content, metadata)

	if err := h.repo.Save(c.Request.Context(), newNote); err != nil {
		c.JSON(500, gin.H{"error": "failed to save note"})
		return
	}

	// Ставим задачи в очередь
	log.Printf("taskQueue is nil? %v", h.taskQueue == nil)
	if h.taskQueue != nil {
		noteID := newNote.ID().String()
		log.Printf("Enqueuing tasks for note %s", noteID)
		if err := h.taskQueue.EnqueueExtractKeywords(c.Request.Context(), noteID, 10); err != nil {
			log.Printf("Failed to enqueue extract keywords: %v", err)
		}
		if err := h.taskQueue.EnqueueComputeEmbedding(c.Request.Context(), noteID); err != nil {
			log.Printf("Failed to enqueue compute embedding: %v", err)
		}
	} else {
		log.Println("taskQueue is nil, tasks not enqueued")
	}

	c.JSON(201, gin.H{
		"id":         newNote.ID(),
		"title":      newNote.Title().String(),
		"content":    newNote.Content().String(),
		"metadata":   newNote.Metadata().Value(),
		"created_at": newNote.CreatedAt(),
		"updated_at": newNote.UpdatedAt(),
	})
}

type updateNoteRequest struct {
	Title    string                 `json:"title"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
}

func (h *Handler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	existing, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch note"})
		return
	}
	if existing == nil {
		c.JSON(404, gin.H{"error": "note not found"})
		return
	}

	var req updateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Флаг, изменился ли текст (чтобы ставить задачи)
	textChanged := false

	if req.Title != "" {
		title, err := note.NewTitle(req.Title)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if err := existing.UpdateTitle(title); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		textChanged = true
	}
	if req.Content != "" {
		content, err := note.NewContent(req.Content)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if err := existing.UpdateContent(content); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		textChanged = true
	}
	if req.Metadata != nil {
		metadata, err := note.NewMetadata(req.Metadata)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if err := existing.UpdateMetadata(metadata); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		// изменение метаданных не влияет на текст, можно не ставить задачи
	}

	if err := h.repo.Save(c.Request.Context(), existing); err != nil {
		c.JSON(500, gin.H{"error": "failed to update note"})
		return
	}

	// Если текст изменился, ставим задачи заново
	if textChanged && h.taskQueue != nil {
		noteID := existing.ID().String()
		_ = h.taskQueue.EnqueueExtractKeywords(c.Request.Context(), noteID, 10)
		_ = h.taskQueue.EnqueueComputeEmbedding(c.Request.Context(), noteID)
	}

	c.JSON(200, gin.H{
		"id":         existing.ID(),
		"title":      existing.Title().String(),
		"content":    existing.Content().String(),
		"metadata":   existing.Metadata().Value(),
		"created_at": existing.CreatedAt(),
		"updated_at": existing.UpdatedAt(),
	})
}

func (h *Handler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	existing, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch note"})
		return
	}
	if existing == nil {
		c.JSON(404, gin.H{"error": "note not found"})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(500, gin.H{"error": "failed to delete note"})
		return
	}
	c.JSON(204, nil)
}

func (h *Handler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	n, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch note"})
		return
	}
	if n == nil {
		c.JSON(404, gin.H{"error": "note not found"})
		return
	}

	c.JSON(200, gin.H{
		"id":         n.ID(),
		"title":      n.Title().String(),
		"content":    n.Content().String(),
		"metadata":   n.Metadata().Value(),
		"created_at": n.CreatedAt(),
		"updated_at": n.UpdatedAt(),
	})
}
