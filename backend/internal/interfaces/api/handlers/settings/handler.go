// Package settings provides HTTP handlers for user settings
package settings

import (
	"net/http"

	"knowledge-graph/internal/application/user"
	userDomain "knowledge-graph/internal/domain/user"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
)

// Handler handles settings requests
type Handler struct {
	service *user.SettingsService
}

// NewHandler creates a new settings handler
func NewHandler(service *user.SettingsService) *Handler {
	return &Handler{service: service}
}

// SettingResponse represents a setting in API responses
type SettingResponse struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	UpdatedAt string      `json:"updated_at"`
}

// GetSettingsRequest represents a request to get settings
type GetSettingsRequest struct {
	Keys []string `json:"keys,omitempty"`
}

// UpdateSettingRequest represents a request to update a setting
type UpdateSettingRequest struct {
	Key   string      `json:"key" binding:"required"`
	Value interface{} `json:"value" binding:"required"`
}

// GetMySettings returns all settings for the current user
func (h *Handler) GetMySettings(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	settings, err := h.service.GetSettings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load settings"})
		return
	}

	response := make([]SettingResponse, 0, len(settings))
	for _, s := range settings {
		value, _ := s.GetValue()
		response = append(response, SettingResponse{
			Key:       s.Key().String(),
			Value:     value["value"],
			UpdatedAt: s.UpdatedAt().Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, gin.H{"settings": response})
}

// GetSetting returns a specific setting value
func (h *Handler) GetSetting(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	key := userDomain.SettingKey(c.Param("key"))
	if err := key.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid setting key"})
		return
	}

	value, err := h.service.GetSettingValue(c.Request.Context(), userID, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get setting"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key":   key.String(),
		"value": value["value"],
	})
}

// UpdateSetting updates a setting value
func (h *Handler) UpdateSetting(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key := userDomain.SettingKey(req.Key)
	if err := key.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid setting key"})
		return
	}

	if err := h.service.SetValue(c.Request.Context(), userID, key, req.Value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update setting"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "setting updated",
		"key":     key.String(),
		"value":   req.Value,
	})
}

// DeleteSetting deletes a setting
func (h *Handler) DeleteSetting(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	key := userDomain.SettingKey(c.Param("key"))
	if err := key.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid setting key"})
		return
	}

	if err := h.service.DeleteSetting(c.Request.Context(), userID, key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete setting"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "setting deleted"})
}

// GetGalacticMode returns the galactic mode setting (convenience endpoint)
func (h *Handler) GetGalacticMode(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	value, err := h.service.GetBool(c.Request.Context(), userID, userDomain.SettingKeyGalacticMode)
	if err != nil {
		// Return default value
		value = false
	}

	c.JSON(http.StatusOK, gin.H{"galactic_mode": value})
}

// ToggleGalacticMode toggles the galactic mode setting
func (h *Handler) ToggleGalacticMode(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Get current value
	current, _ := h.service.GetBool(c.Request.Context(), userID, userDomain.SettingKeyGalacticMode)

	// Toggle
	newValue := !current
	if err := h.service.SetBool(c.Request.Context(), userID, userDomain.SettingKeyGalacticMode, newValue); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update setting"})
		return
	}

	message := "galactic mode enabled"
	if !newValue {
		message = "galactic mode disabled"
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       message,
		"galactic_mode": newValue,
	})
}
