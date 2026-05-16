// Package draft provides HTTP handlers for draft operations
package draft

import (
	"net/http"

	"knowledge-graph/internal/application/draft"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles draft requests
type Handler struct {
	service *draft.Service
}

// NewHandler creates a new draft handler
func NewHandler(service *draft.Service) *Handler {
	return &Handler{service: service}
}

// SaveDraftRequest represents a request to save a draft
type SaveDraftRequest struct {
	Content string `json:"content" binding:"required"`
	Title   string `json:"title"`
}

// SaveDraft saves or updates a draft for a note
func (h *Handler) SaveDraft(c *gin.Context) {
	noteIDStr := c.Param("note_id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note_id"})
		return
	}

	// TODO: Get user ID from authentication context
	userID := uuid.New() // Placeholder

	var req SaveDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	draft, err := h.service.SaveDraft(c.Request.Context(), noteID, userID, req.Content, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save draft"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      draft.ID(),
		"state":   draft.State(),
		"message": "draft saved",
	})
}

// SyncDraft synchronizes a draft with the server
func (h *Handler) SyncDraft(c *gin.Context) {
	draftIDStr := c.Param("draft_id")
	draftID, err := uuid.Parse(draftIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid draft_id"})
		return
	}

	if err := h.service.SyncDraft(c.Request.Context(), draftID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sync draft"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "draft synced"})
}

// ResolveConflict resolves a conflict in a draft
func (h *Handler) ResolveConflict(c *gin.Context) {
	draftIDStr := c.Param("draft_id")
	draftID, err := uuid.Parse(draftIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid draft_id"})
		return
	}

	if err := h.service.ResolveConflict(c.Request.Context(), draftID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve conflict"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "conflict resolved"})
}

// GetDraft retrieves a draft for a note
func (h *Handler) GetDraft(c *gin.Context) {
	noteIDStr := c.Param("note_id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note_id"})
		return
	}

	// TODO: Get user ID from authentication context
	userID := uuid.New() // Placeholder

	draft, err := h.service.GetLatestDraft(c.Request.Context(), noteID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get draft"})
		return
	}

	if draft == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "draft not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         draft.ID(),
		"note_id":    draft.NoteID(),
		"user_id":    draft.UserID(),
		"content":    draft.Content(),
		"title":      draft.Title(),
		"state":      draft.State(),
		"updated_at": draft.UpdatedAt(),
	})
}

// DeleteDraft deletes a draft
func (h *Handler) DeleteDraft(c *gin.Context) {
	draftIDStr := c.Param("draft_id")
	draftID, err := uuid.Parse(draftIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid draft_id"})
		return
	}

	if err := h.service.DeleteDraft(c.Request.Context(), draftID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete draft"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "draft deleted"})
}

// GetActiveDrafts retrieves all active drafts for the current user
func (h *Handler) GetActiveDrafts(c *gin.Context) {
	// TODO: Get user ID from authentication context
	userID := uuid.New() // Placeholder

	drafts, err := h.service.GetActiveDrafts(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get drafts"})
		return
	}

	response := make([]gin.H, len(drafts))
	for i, draft := range drafts {
		response[i] = gin.H{
			"id":         draft.ID(),
			"note_id":    draft.NoteID(),
			"content":    draft.Content(),
			"title":      draft.Title(),
			"state":      draft.State(),
			"updated_at": draft.UpdatedAt(),
		}
	}

	c.JSON(http.StatusOK, gin.H{"drafts": response})
}
