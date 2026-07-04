package taghandler

import (
	"net/http"
	"time"

	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/infrastructure/db/postgres"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler is the HTTP handler for working with tags
type Handler struct {
	tagRepo  *postgres.TagRepository
	noteRepo note.Repository
}

// New creates a new tag handler
func New(tagRepo *postgres.TagRepository, noteRepo note.Repository) *Handler {
	return &Handler{
		tagRepo:  tagRepo,
		noteRepo: noteRepo,
	}
}

// TagResponse is the response structure for a tag
type TagResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateRequest is the request to create a tag
type CreateRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

// UpdateRequest is the request to update a tag
type UpdateRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

// toTagResponse converts a model into a response
func toTagResponse(tag *postgres.TagModel) TagResponse {
	return TagResponse{
		ID:        tag.ID.String(),
		Name:      tag.Name,
		CreatedAt: tag.CreatedAt,
	}
}

// Create creates a new tag
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Check name uniqueness
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

// List returns the list of all tags
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

// Get returns a tag by ID
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

// Update updates a tag
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

	// Check that the tag exists
	tag, err := h.tagRepo.FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tag"})
		return
	}
	if tag == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		return
	}

	// Check uniqueness of the new name (if it changed)
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

// Delete removes a tag
func (h *Handler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	ctx := c.Request.Context()

	// Check that the tag exists
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

// AddTagToNote attaches a tag to a note
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

	// Check that the note exists
	note, err := h.noteRepo.FindByID(ctx, noteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check note"})
		return
	}
	if note == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}

	// Check that the tag exists
	tag, err := h.tagRepo.FindByID(ctx, tagID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check tag"})
		return
	}
	if tag == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		return
	}

	// Check whether the tag is already attached
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

// RemoveTagFromNote detaches a tag from a note
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

// GetTagsByNote returns a note's tags
func (h *Handler) GetTagsByNote(c *gin.Context) {
	noteIDStr := c.Param("id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note id"})
		return
	}

	ctx := c.Request.Context()

	// Check that the note exists
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
