package taghandler

import (
	"net/http"
	"time"

	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/infrastructure/db/postgres"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler — HTTP-хендлер для работы с тегами
type Handler struct {
	tagRepo  *postgres.TagRepository
	noteRepo note.Repository
}

// New создает новый хендлер тегов
func New(tagRepo *postgres.TagRepository, noteRepo note.Repository) *Handler {
	return &Handler{
		tagRepo:  tagRepo,
		noteRepo: noteRepo,
	}
}

// TagResponse — структура ответа с тегом
type TagResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateRequest — запрос на создание тега
type CreateRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

// UpdateRequest — запрос на обновление тега
type UpdateRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

// toTagResponse преобразует модель в ответ
func toTagResponse(tag *postgres.TagModel) TagResponse {
	return TagResponse{
		ID:        tag.ID.String(),
		Name:      tag.Name,
		CreatedAt: tag.CreatedAt,
	}
}

// Create создает новый тег
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Проверяем уникальность имени
	existing, err := h.tagRepo.FindByName(ctx, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check tag existence"})
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "tag with this name already exists"})
		return
	}

	tag := &postgres.TagModel{
		ID:   uuid.New(),
		Name: req.Name,
	}

	if err := h.tagRepo.Create(ctx, tag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create tag"})
		return
	}

	c.JSON(http.StatusCreated, toTagResponse(tag))
}

// List возвращает список всех тегов
func (h *Handler) List(c *gin.Context) {
	ctx := c.Request.Context()

	tags, err := h.tagRepo.FindAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tags"})
		return
	}

	response := make([]TagResponse, len(tags))
	for i, tag := range tags {
		response[i] = toTagResponse(tag)
	}

	c.JSON(http.StatusOK, response)
}

// Get возвращает тег по ID
func (h *Handler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	ctx := c.Request.Context()
	tag, err := h.tagRepo.FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tag"})
		return
	}
	if tag == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		return
	}

	c.JSON(http.StatusOK, toTagResponse(tag))
}

// Update обновляет тег
func (h *Handler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Проверяем существование тега
	tag, err := h.tagRepo.FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tag"})
		return
	}
	if tag == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		return
	}

	// Проверяем уникальность нового имени (если оно изменилось)
	if req.Name != tag.Name {
		existing, err := h.tagRepo.FindByName(ctx, req.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check tag existence"})
			return
		}
		if existing != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "tag with this name already exists"})
			return
		}
	}

	tag.Name = req.Name
	if err := h.tagRepo.Update(ctx, tag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tag"})
		return
	}

	c.JSON(http.StatusOK, toTagResponse(tag))
}

// Delete удаляет тег
func (h *Handler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	ctx := c.Request.Context()

	// Проверяем существование тега
	tag, err := h.tagRepo.FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tag"})
		return
	}
	if tag == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		return
	}

	if err := h.tagRepo.Delete(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete tag"})
		return
	}

	c.Status(http.StatusNoContent)
}

// AddTagToNote привязывает тег к заметке
func (h *Handler) AddTagToNote(c *gin.Context) {
	noteIDStr := c.Param("id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note id"})
		return
	}

	var req struct {
		TagID string `json:"tag_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tagID, err := uuid.Parse(req.TagID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	ctx := c.Request.Context()

	// Проверяем существование заметки
	note, err := h.noteRepo.FindByID(ctx, noteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check note"})
		return
	}
	if note == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}

	// Проверяем существование тега
	tag, err := h.tagRepo.FindByID(ctx, tagID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check tag"})
		return
	}
	if tag == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		return
	}

	// Проверяем, не привязан ли уже тег
	exists, err := h.tagRepo.IsTagAssignedToNote(ctx, noteID, tagID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check assignment"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "tag already assigned to note"})
		return
	}

	if err := h.tagRepo.AddTagToNote(ctx, noteID, tagID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign tag"})
		return
	}

	c.Status(http.StatusCreated)
}

// RemoveTagFromNote отвязывает тег от заметки
func (h *Handler) RemoveTagFromNote(c *gin.Context) {
	noteIDStr := c.Param("id")
	tagIDStr := c.Param("tagId")

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note id"})
		return
	}

	tagID, err := uuid.Parse(tagIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	ctx := c.Request.Context()

	if err := h.tagRepo.RemoveTagFromNote(ctx, noteID, tagID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove tag"})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetTagsByNote возвращает теги заметки
func (h *Handler) GetTagsByNote(c *gin.Context) {
	noteIDStr := c.Param("id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note id"})
		return
	}

	ctx := c.Request.Context()

	// Проверяем существование заметки
	note, err := h.noteRepo.FindByID(ctx, noteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check note"})
		return
	}
	if note == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}

	tags, err := h.tagRepo.GetTagsByNoteID(ctx, noteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tags"})
		return
	}

	response := make([]TagResponse, len(tags))
	for i, tag := range tags {
		response[i] = toTagResponse(tag)
	}

	c.JSON(http.StatusOK, response)
}
