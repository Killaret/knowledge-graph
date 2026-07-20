// Package backup provides HTTP handlers for backup operations
package backup

import (
	"context"
	"net/http"
	"time"

	"knowledge-graph/internal/application/common"
	"knowledge-graph/internal/config"
	"knowledge-graph/internal/infrastructure/queue/tasks"

	"github.com/gin-gonic/gin"
)

// Handler handles backup requests
type Handler struct {
	cfg       *config.Config
	taskQueue common.TaskQueue
}

// NewHandler creates a new backup handler
func NewHandler(cfg *config.Config, taskQueue common.TaskQueue) *Handler {
	return &Handler{
		cfg:       cfg,
		taskQueue: taskQueue,
	}
}

// TriggerCloudBackupRequest represents a request to trigger cloud backup
type TriggerCloudBackupRequest struct {
	LocalPath string `json:"local_path" binding:"required"`
}

// TriggerCloudBackup triggers a cloud backup task
func (h *Handler) TriggerCloudBackup(c *gin.Context) {
	// Check if cloud backup is enabled
	if !h.cfg.BackupCloudEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cloud backup is not enabled"})
		return
	}

	var req TriggerCloudBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate remote key with timestamp
	timestamp := time.Now().Format("2006-01-02")
	remoteKey := "backups/backup-personal-" + timestamp + ".sql.gz"

	// Create Asynq task
	task, err := tasks.NewBackupToCloudTask(req.LocalPath, remoteKey, timestamp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create backup task"})
		return
	}

	// Enqueue task using task queue
	if h.taskQueue != nil {
		err = h.taskQueue.Enqueue(context.Background(), task)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue backup task"})
			return
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "task queue not available"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":    "cloud backup task enqueued",
		"local_path": req.LocalPath,
		"remote_key": remoteKey,
	})
}

// GetBackupStatus returns the current backup configuration status
func (h *Handler) GetBackupStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"cloud_enabled":  h.cfg.BackupCloudEnabled,
		"cloud_provider": h.cfg.BackupCloudProvider,
		"local_path":     h.cfg.BackupLocalPath,
		"schedule":       h.cfg.BackupSchedule,
		"retention_days": h.cfg.BackupRetentionDays,
		"yandex_folder":  h.cfg.BackupYandexFolder,
		"yandex_token":   maskSensitive(h.cfg.BackupYandexOAuthToken),
		"max_backups":    h.cfg.BackupYandexMaxBackups,
	})
}

// maskSensitive masks sensitive information for display
func maskSensitive(s string) string {
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + "****" + s[len(s)-4:]
}
