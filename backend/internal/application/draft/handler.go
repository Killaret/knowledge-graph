package draft

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles draft-related HTTP requests
type Handler struct {
	service *Service
}

// NewHandler creates a new draft handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// SaveDraftRequest represents the request to save a draft
type SaveDraftRequest struct {
	NoteID  uuid.UUID `json:"note_id" binding:"required"`
	Content string    `json:"content" binding:"required"`
	Title   string    `json:"title"`
}

// SaveDraft saves or updates a draft
func (h *Handler) SaveDraft(c *gin.Context) {
	var req SaveDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	draft, err := h.service.SaveDraft(c.Request.Context(), req.NoteID, userUUID, req.Content, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         draft.ID().String(),
		"note_id":    draft.NoteID().String(),
		"state":      draft.State(),
		"updated_at": draft.UpdatedAt(),
	})
}

// GetDraft gets a draft by note ID
func (h *Handler) GetDraft(c *gin.Context) {
	noteIDStr := c.Param("note_id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note id"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	draft, err := h.service.GetLatestDraft(c.Request.Context(), noteID, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if draft == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "draft not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         draft.ID().String(),
		"note_id":    draft.NoteID().String(),
		"content":    draft.Content(),
		"title":      draft.Title(),
		"state":      draft.State(),
		"updated_at": draft.UpdatedAt(),
		"created_at": draft.CreatedAt(),
	})
}

// DeleteDraft deletes a draft
func (h *Handler) DeleteDraft(c *gin.Context) {
	noteIDStr := c.Param("note_id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note id"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	draft, err := h.service.GetLatestDraft(c.Request.Context(), noteID, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if draft == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "draft not found"})
		return
	}

	if err := h.service.DeleteDraft(c.Request.Context(), draft.ID()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "draft deleted"})
}

// SyncDraft synchronizes a draft with the server
func (h *Handler) SyncDraft(c *gin.Context) {
	draftIDStr := c.Param("draft_id")
	draftID, err := uuid.Parse(draftIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid draft id"})
		return
	}

	if err := h.service.SyncDraft(c.Request.Context(), draftID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "sync completed"})
}

// ResolveConflict resolves a conflict
func (h *Handler) ResolveConflict(c *gin.Context) {
	draftIDStr := c.Param("draft_id")
	draftID, err := uuid.Parse(draftIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid draft id"})
		return
	}

	if err := h.service.ResolveConflict(c.Request.Context(), draftID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "conflict resolved"})
}

// GetActiveDrafts gets all active drafts for the current user
func (h *Handler) GetActiveDrafts(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	drafts, err := h.service.GetActiveDrafts(c.Request.Context(), userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]gin.H, len(drafts))
	for i, draft := range drafts {
		result[i] = gin.H{
			"id":         draft.ID().String(),
			"note_id":    draft.NoteID().String(),
			"content":    draft.Content(),
			"title":      draft.Title(),
			"state":      draft.State(),
			"updated_at": draft.UpdatedAt(),
			"created_at": draft.CreatedAt(),
		}
	}

	c.JSON(http.StatusOK, gin.H{"drafts": result})
}

// RegisterRoutes registers draft routes
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	drafts := r.Group("/notes/:note_id/draft")
	{
		drafts.POST("", h.SaveDraft)
		drafts.GET("", h.GetDraft)
		drafts.DELETE("", h.DeleteDraft)
	}

	draftActions := r.Group("/drafts/:draft_id")
	{
		draftActions.POST("/sync", h.SyncDraft)
		draftActions.POST("/resolve", h.ResolveConflict)
	}

	r.GET("/drafts", h.GetActiveDrafts)
}
