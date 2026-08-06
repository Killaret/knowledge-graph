package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

// TypeBackupToCloud is the task type for uploading backup to cloud storage
const TypeBackupToCloud = "backup:cloud"

// BackupToCloudPayload contains the data needed for cloud backup
type BackupToCloudPayload struct {
	LocalPath  string `json:"local_path"`
	RemoteKey  string `json:"remote_key"`
	BackupDate string `json:"backup_date"`
}

// NewBackupToCloudTask creates a new Asynq task for uploading backup to cloud
func NewBackupToCloudTask(localPath, remoteKey, backupDate string) (*asynq.Task, error) {
	if localPath == "" {
		return nil, fmt.Errorf("local_path is required")
	}
	if remoteKey == "" {
		return nil, fmt.Errorf("remote_key is required")
	}

	payload, err := json.Marshal(BackupToCloudPayload{
		LocalPath:  localPath,
		RemoteKey:  remoteKey,
		BackupDate: backupDate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	opts := []asynq.Option{
		asynq.MaxRetry(3),
		asynq.Timeout(30 * time.Minute),
		asynq.Queue("default"),
		asynq.Unique(24 * time.Hour),
	}

	return asynq.NewTask(TypeBackupToCloud, payload, opts...), nil
}

// BackupServiceInterface defines the interface for backup service
type BackupServiceInterface interface {
	UploadBackup(ctx context.Context, localPath, remoteKey string) error
}

// HandleBackupToCloud is the handler for TypeBackupToCloud tasks
func HandleBackupToCloud(ctx context.Context, t *asynq.Task, backupSvc BackupServiceInterface) error {
	var p BackupToCloudPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("[Asynq] Starting cloud backup: %s -> %s", p.LocalPath, p.RemoteKey)

	err := backupSvc.UploadBackup(ctx, p.LocalPath, p.RemoteKey)
	if err != nil {
		log.Printf("[Asynq] Failed to upload backup to cloud: %v", err)
		return err
	}

	log.Printf("[Asynq] Cloud backup completed successfully: %s", p.RemoteKey)
	return nil
}

// TypeDatabaseBackup is the task type for taking a full database backup.
const TypeDatabaseBackup = "backup:database"

// DatabaseBackupPayload contains data for the database backup task.
type DatabaseBackupPayload struct {
	TriggeredAt string `json:"triggered_at"`
}

// NewDatabaseBackupTask creates a new Asynq task for full database backup.
// It is unique for 5 minutes to avoid backing up on every note change.
func NewDatabaseBackupTask() (*asynq.Task, error) {
	payload, err := json.Marshal(DatabaseBackupPayload{
		TriggeredAt: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal database backup payload: %w", err)
	}

	opts := []asynq.Option{
		asynq.MaxRetry(2),
		asynq.Timeout(30 * time.Minute),
		asynq.Queue("default"),
		asynq.Unique(5 * time.Minute),
		asynq.ProcessIn(30 * time.Second),
	}

	return asynq.NewTask(TypeDatabaseBackup, payload, opts...), nil
}

// DatabaseBackupRunner defines the interface for running a full database backup.
type DatabaseBackupRunner interface {
	Run(ctx context.Context) (string, error)
}

// HandleDatabaseBackup is the handler for TypeDatabaseBackup tasks.
func HandleDatabaseBackup(ctx context.Context, t *asynq.Task, runner DatabaseBackupRunner) error {
	var p DatabaseBackupPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal database backup payload: %w", err)
	}

	log.Printf("[Asynq] Starting database backup (triggered at %s)", p.TriggeredAt)

	backupPath, err := runner.Run(ctx)
	if err != nil {
		log.Printf("[Asynq] Database backup failed: %v", err)
		return err
	}

	log.Printf("[Asynq] Database backup completed: %s", backupPath)
	return nil
}
