// Package notificationhandler exposes in-app notifications (COMET-1 stage C).
package notificationhandler

import (
	"errors"
	"net/http"
	"time"

	"knowledge-graph/internal/domain/notification"
	"knowledge-graph/internal/interfaces/api/common"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Handler serves the authenticated user's notifications.
type Handler struct {
	repo notification.Repository
}

// New builds the handler.
func New(repo notification.Repository) *Handler {
	return &Handler{repo: repo}
}

type notificationResponse struct {
	ID        string     `json:"id"`
	NoteID    *string    `json:"note_id"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	CreatedAt time.Time  `json:"created_at"`
	ReadAt    *time.Time `json:"read_at"`
}

func toResponse(n *notification.Notification) notificationResponse {
	var noteID *string
	if n.NoteID() != nil {
		s := n.NoteID().String()
		noteID = &s
	}
	return notificationResponse{
		ID:        n.ID().String(),
		NoteID:    noteID,
		Type:      n.Type(),
		Title:     n.Title(),
		Body:      n.Body(),
		CreatedAt: n.CreatedAt(),
		ReadAt:    n.ReadAt(),
	}
}

// List returns the user's notifications newest first plus the unread count.
func (h *Handler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		common.Error(c, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	limit := 50
	items, err := h.repo.ListByUser(c.Request.Context(), userID, limit)
	if err != nil {
		common.InternalError(c)
		return
	}
	unread, err := h.repo.CountUnread(c.Request.Context(), userID)
	if err != nil {
		common.InternalError(c)
		return
	}
	out := make([]notificationResponse, 0, len(items))
	for _, n := range items {
		out = append(out, toResponse(n))
	}
	c.JSON(http.StatusOK, gin.H{"items": out, "unread_count": unread})
}

// MarkRead marks a notification as read. Only the owner may do so.
func (h *Handler) MarkRead(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		common.Error(c, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		common.Error(c, http.StatusBadRequest, "invalid_id", "invalid notification id")
		return
	}
	err = h.repo.MarkRead(c.Request.Context(), id, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		common.Error(c, http.StatusNotFound, "not_found", "notification not found")
		return
	}
	if err != nil {
		common.InternalError(c)
		return
	}
	c.Status(http.StatusNoContent)
}
