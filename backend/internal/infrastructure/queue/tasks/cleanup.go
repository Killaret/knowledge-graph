package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

// TypeCleanupSoftDeleted is the task type for cleaning up soft-deleted records
const TypeCleanupSoftDeleted = "cleanup:soft_deleted"

// CleanupSoftDeletedPayload contains the data needed for cleanup
// Days: delete records that have been soft-deleted for more than N days
// Tables: list of tables to clean up (empty = all tables)
type CleanupSoftDeletedPayload struct {
	Days   int      `json:"days"`
	Tables []string `json:"tables,omitempty"`
}

// NewCleanupSoftDeletedTask creates a new Asynq task for cleaning up soft-deleted records
// days: delete records that have been soft-deleted for more than N days (default: 90)
func NewCleanupSoftDeletedTask(days int, tables []string) (*asynq.Task, error) {
	if days <= 0 {
		days = 90 // Default: 90 days
	}

	payload, err := json.Marshal(CleanupSoftDeletedPayload{
		Days:   days,
		Tables: tables,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	opts := []asynq.Option{
		asynq.MaxRetry(3),
		asynq.Timeout(5 * time.Minute),
		asynq.Queue("default"),
		asynq.Unique(24 * time.Hour), // Only one cleanup task per day
	}

	return asynq.NewTask(TypeCleanupSoftDeleted, payload, opts...), nil
}

// CleanupServiceInterface defines the interface for cleanup service
type CleanupServiceInterface interface {
	CleanupSoftDeleted(ctx context.Context, days int, tables []string) error
}

// HandleCleanupSoftDeleted is the handler for TypeCleanupSoftDeleted tasks
func HandleCleanupSoftDeleted(ctx context.Context, t *asynq.Task, cleanupSvc CleanupServiceInterface) error {
	var p CleanupSoftDeletedPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	err := cleanupSvc.CleanupSoftDeleted(ctx, p.Days, p.Tables)
	if err != nil {
		log.Printf("[Asynq] Failed to cleanup soft-deleted records: %v", err)
	}
	return err
}

// TypeCleanupExpiredTokens is the task type for cleaning up expired tokens
const TypeCleanupExpiredTokens = "cleanup:expired_tokens"

// CleanupExpiredTokensPayload contains parameters for token cleanup
type CleanupExpiredTokensPayload struct {
	BatchSize int `json:"batch_size"`
}

// NewCleanupExpiredTokensTask creates a new Asynq task for cleaning up expired tokens
func NewCleanupExpiredTokensTask(batchSize int) (*asynq.Task, error) {
	if batchSize <= 0 {
		batchSize = 1000
	}

	payload, err := json.Marshal(CleanupExpiredTokensPayload{
		BatchSize: batchSize,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	opts := []asynq.Option{
		asynq.MaxRetry(3),
		asynq.Timeout(5 * time.Minute),
		asynq.Queue("default"),
		asynq.Unique(24 * time.Hour),
	}

	return asynq.NewTask(TypeCleanupExpiredTokens, payload, opts...), nil
}

// TokenCleanupServiceInterface defines the interface for token cleanup service
type TokenCleanupServiceInterface interface {
	CleanupExpiredTokens(ctx context.Context, batchSize int) error
}

// HandleCleanupExpiredTokens is the handler for TypeCleanupExpiredTokens tasks
func HandleCleanupExpiredTokens(ctx context.Context, t *asynq.Task, tokenCleanupSvc TokenCleanupServiceInterface) error {
	var p CleanupExpiredTokensPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	err := tokenCleanupSvc.CleanupExpiredTokens(ctx, p.BatchSize)
	if err != nil {
		log.Printf("[Asynq] Failed to cleanup expired tokens: %v", err)
	}
	return err
}
