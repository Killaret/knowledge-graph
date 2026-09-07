package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// TypeRefreshRecommendations is the task type for refreshing note recommendations
const TypeRefreshRecommendations = "recommendation:refresh"

// RefreshRecommendationsPayload contains the data needed to refresh recommendations
type RefreshRecommendationsPayload struct {
	NoteID uuid.UUID `json:"note_id"`
}

// NewRefreshRecommendationsTask creates a new Asynq task for refreshing recommendations
// delay: time to wait before processing (for deduplication)
func NewRefreshRecommendationsTask(noteID uuid.UUID, delay time.Duration) (*asynq.Task, error) {
	payload, err := json.Marshal(RefreshRecommendationsPayload{NoteID: noteID})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create unique task ID based on noteID to prevent duplicate tasks
	uniqueKey := fmt.Sprintf("rec:%s", noteID)

	// Use unique key for deduplication - multiple tasks for same note will be merged
	opts := []asynq.Option{
		asynq.MaxRetry(3),
		asynq.Timeout(30 * time.Second),
		asynq.Queue("default"),
		asynq.TaskID(uniqueKey),
	}

	if delay > 0 {
		opts = append(opts, asynq.ProcessIn(delay))
	}

	return asynq.NewTask(TypeRefreshRecommendations, payload, opts...), nil
}

// RefreshServiceInterface defines the interface for refresh service
type RefreshServiceInterface interface {
	RefreshRecommendations(ctx context.Context, noteID uuid.UUID) error
}

// HandleRefreshRecommendations is the handler for TypeRefreshRecommendations tasks
func HandleRefreshRecommendations(ctx context.Context, t *asynq.Task, refreshSvc RefreshServiceInterface) error {
	var p RefreshRecommendationsPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	err := refreshSvc.RefreshRecommendations(ctx, p.NoteID)
	if err != nil {
		log.Printf("[Asynq] Failed to refresh recommendations for note %s: %v", p.NoteID, err)
	}
	return err
}
