package notehandler

import (
	"knowledge-graph/internal/application/common"
	"knowledge-graph/internal/domain/note"
	"log"
	"strconv"

	graphQueries "knowledge-graph/internal/application/queries/graph"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	repo               note.Repository
	taskQueue          common.TaskQueue
	suggestionsHandler *graphQueries.GetSuggestionsHandler
}

func New(repo note.Repository, taskQueue common.TaskQueue, suggestionsHandler *graphQueries.GetSuggestionsHandler) *Handler {
	return &Handler{
		repo:               repo,
		taskQueue:          taskQueue,
		suggestionsHandler: suggestionsHandler,
	}
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

// GetSuggestions возвращает рекомендации для заметки
func (h *Handler) GetSuggestions(c *gin.Context) {
	// Парсим ID из URL
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	// Парсим параметр limit (по умолчанию 10)
	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// Формируем запрос
	query := graphQueries.GetSuggestionsQuery{
		NoteID: id,
		Limit:  limit,
	}

	// Вызываем обработчик
	suggestions, err := h.suggestionsHandler.Handle(c.Request.Context(), query)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get suggestions"})
		return
	}

	// Отдаём JSON
	c.JSON(200, suggestions)
}

// SearchRequest - search query parameters
type SearchRequest struct {
	Q    string `form:"q"`    // search query
	Page int    `form:"page"` // page number (default 1)
	Size int    `form:"size"` // page size (default 20)
}

// SearchResponse - search response structure
type SearchResponse struct {
	Data       []*note.Note `json:"data"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	Size       int          `json:"size"`
	TotalPages int          `json:"totalPages"`
}

// Search performs full-text search on notes
func (h *Handler) Search(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Default values
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 20
	}

	// Perform search using repository directly (for simplicity, could use service layer)
	notes, total, err := h.repo.Search(c.Request.Context(), req.Q, req.Size, (req.Page-1)*req.Size)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to search notes"})
		return
	}

	// Convert domain notes to JSON-структуры (как в методе List)
	responseNotes := make([]gin.H, len(notes))
	for i, n := range notes {
		responseNotes[i] = gin.H{
			"id":         n.ID(),
			"title":      n.Title().String(),
			"content":    n.Content().String(),
			"metadata":   n.Metadata().Value(),
			"created_at": n.CreatedAt(),
			"updated_at": n.UpdatedAt(),
		}
	}

	// Calculate total pages
	totalPages := int((total + int64(req.Size) - 1) / int64(req.Size))

	c.JSON(200, gin.H{
		"data":       responseNotes,
		"total":      total,
		"page":       req.Page,
		"size":       req.Size,
		"totalPages": totalPages,
	})
}

// List возвращает список заметок с пагинацией
func (h *Handler) List(c *gin.Context) {
	// Получаем параметры пагинации из query
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Получаем заметки из репозитория
	notes, total, err := h.repo.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch notes"})
		return
	}

	// Преобразуем доменные модели в JSON-структуры
	responseNotes := make([]gin.H, len(notes))
	for i, n := range notes {
		responseNotes[i] = gin.H{
			"id":         n.ID(),
			"title":      n.Title().String(),
			"content":    n.Content().String(),
			"metadata":   n.Metadata().Value(),
			"created_at": n.CreatedAt(),
			"updated_at": n.UpdatedAt(),
		}
	}

	c.JSON(200, gin.H{
		"notes":  responseNotes,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}
