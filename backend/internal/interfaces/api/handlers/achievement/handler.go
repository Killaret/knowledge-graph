package achievementhandler

import (
	"net/http"

	achievementApp "knowledge-graph/internal/application/achievement"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
)

// Handler handles achievement API requests.
type Handler struct {
	service *achievementApp.Service
}

// NewHandler creates a new achievement handler.
func NewHandler(service *achievementApp.Service) *Handler {
	return &Handler{service: service}
}

// AchievementResponse represents an achievement in API responses.
type AchievementResponse struct {
	ID               string `json:"id"`
	Code             string `json:"code"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	Icon             string `json:"icon"`
	Points           int    `json:"points"`
	UnlockedAt       string `json:"unlocked_at,omitempty"`
	NotificationSeen bool   `json:"notification_seen"`
}

// ListAchievements returns all available achievements.
func (h *Handler) ListAchievements(c *gin.Context) {
	allAchievements, err := h.service.GetAllAchievements(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch achievements"})
		return
	}

	response := make([]AchievementResponse, 0, len(allAchievements))
	for _, a := range allAchievements {
		response = append(response, AchievementResponse{
			ID:          a.ID().String(),
			Code:        a.Code(),
			Title:       a.Title(),
			Description: a.Description(),
			Icon:        a.Icon(),
			Points:      a.Points(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"achievements": response})
}

// GetUserAchievements returns achievements earned by the current user.
func (h *Handler) GetUserAchievements(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	achievements, err := h.service.GetUserAchievementsWithStatus(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user achievements"})
		return
	}

	response := make([]AchievementResponse, 0, len(achievements))
	for _, a := range achievements {
		title := ""
		if a.NameEn != nil && *a.NameEn != "" {
			title = *a.NameEn
		} else if a.NameRu != nil {
			title = *a.NameRu
		}

		description := ""
		if a.DescriptionEn != nil && *a.DescriptionEn != "" {
			description = *a.DescriptionEn
		} else if a.DescriptionRu != nil {
			description = *a.DescriptionRu
		}

		icon := ""
		if a.IconEmoji != nil {
			icon = *a.IconEmoji
		}

		response = append(response, AchievementResponse{
			ID:               a.ID,
			Code:             a.Code,
			Title:            title,
			Description:      description,
			Icon:             icon,
			Points:           a.Points,
			UnlockedAt:       a.UnlockedAt,
			NotificationSeen: a.NotificationSeen,
		})
	}

	c.JSON(http.StatusOK, gin.H{"achievements": response})
}

// MarkSeen marks a user's achievement notification as seen.
func (h *Handler) MarkSeen(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	achievementID := c.Param("id")
	if achievementID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing achievement id"})
		return
	}

	if err := h.service.MarkNotificationSeen(c.Request.Context(), userID, achievementID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark notification as seen"})
		return
	}

	c.Status(http.StatusNoContent)
}
