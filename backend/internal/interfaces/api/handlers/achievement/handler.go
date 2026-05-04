// Package achievement provides HTTP handlers for achievements
package achievement

import (
	"net/http"

	"knowledge-graph/internal/application/achievement"
	achievementDomain "knowledge-graph/internal/domain/achievement"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
)

// Handler handles achievement requests
type Handler struct {
	service *achievement.Service
}

// NewHandler creates a new achievement handler
func NewHandler(service *achievement.Service) *Handler {
	return &Handler{service: service}
}

// AchievementResponse represents an achievement in API responses
type AchievementResponse struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Points      int    `json:"points"`
	ObtainedAt  string `json:"obtained_at,omitempty"`
}

// GetMyAchievements returns all achievements earned by the current user
func (h *Handler) GetMyAchievements(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	achievements, err := h.service.GetUserAchievements(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load achievements"})
		return
	}

	response := make([]AchievementResponse, 0, len(achievements))
	for _, a := range achievements {
		response = append(response, AchievementResponse{
			ID:          a.ID().String(),
			Code:        a.Code(),
			Title:       a.Title(),
			Description: a.Description(),
			Icon:        a.Icon(),
			Points:      a.Points(),
			ObtainedAt:  "", // Would be populated from user_achievements table
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"achievements": response,
		"total_points": calculateTotalPoints(achievements),
	})
}

// GetAllAchievements returns all available achievements in the system
func (h *Handler) GetAllAchievements(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)

	allAchievements, err := h.service.GetAllAchievements(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load achievements"})
		return
	}

	// Get user's earned achievements to mark them
	var userAchievementIDs map[string]bool
	if exists {
		userAchievements, _ := h.service.GetUserAchievements(c.Request.Context(), userID)
		userAchievementIDs = make(map[string]bool)
		for _, ua := range userAchievements {
			userAchievementIDs[ua.ID().String()] = true
		}
	}

	response := make([]map[string]interface{}, 0, len(allAchievements))
	for _, a := range allAchievements {
		// Skip hidden achievements that user hasn't earned
		if a.IsHidden() && !userAchievementIDs[a.ID().String()] {
			continue
		}

		achievementData := map[string]interface{}{
			"id":          a.ID().String(),
			"code":        a.Code(),
			"title":       a.Title(),
			"description": a.Description(),
			"icon":        a.Icon(),
			"points":      a.Points(),
			"earned":      userAchievementIDs[a.ID().String()],
		}

		response = append(response, achievementData)
	}

	c.JSON(http.StatusOK, gin.H{
		"achievements": response,
	})
}

// GetStreak returns the current login streak for the user
func (h *Handler) GetStreak(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	streak, err := h.service.GetStreak(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get streak"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"streak":      streak,
		"next_reward": streak >= 7,
	})
}

// calculateTotalPoints sums up the points from all achievements
func calculateTotalPoints(achievements []achievementDomain.Achievement) int {
	total := 0
	for _, a := range achievements {
		total += a.Points()
	}
	return total
}
