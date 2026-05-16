// Package backup provides HTTP handlers for backup operations
package backup

import (
	"net/http"
	"time"

	"knowledge-graph/internal/config"
	"knowledge-graph/internal/infrastructure/cloud"
	"knowledge-graph/internal/infrastructure/queue/tasks"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

// Handler handles backup requests
type Handler struct {
	cfg         *config.Config
	r2Service   *cloud.R2BackupService
	asynqClient *asynq.Client
}

// NewHandler creates a new backup handler
func NewHandler(cfg *config.Config, r2Service *cloud.R2BackupService, asynqClient *asynq.Client) *Handler {
	return &Handler{
		cfg:         cfg,
		r2Service:   r2Service,
		asynqClient: asynqClient,
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

	// Enqueue task
	_, err = h.asynqClient.Enqueue(task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue backup task"})
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
		"r2_bucket":      h.cfg.BackupR2Bucket,
		"r2_region":      h.cfg.BackupR2Region,
		"r2_account_id":  maskSensitive(h.cfg.BackupR2AccountID),
	})
}

// maskSensitive masks sensitive information for display
func maskSensitive(s string) string {
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + "****" + s[len(s)-4:]
}
