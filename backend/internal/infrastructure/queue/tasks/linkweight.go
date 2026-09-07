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

// TypeRecalculateLinkWeights is the task type for recalculating link weights.
const TypeRecalculateLinkWeights = "links:recalculate_weights"

// RecalculateLinkWeightsPayload contains the data needed for weight recalculation.
type RecalculateLinkWeightsPayload struct {
	NoteID uuid.UUID `json:"note_id"`
}

// NewRecalculateLinkWeightsTask creates a new Asynq task for recalculating link weights.
func NewRecalculateLinkWeightsTask(noteID uuid.UUID, delay time.Duration) (*asynq.Task, error) {
	payload, err := json.Marshal(RecalculateLinkWeightsPayload{NoteID: noteID})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	uniqueKey := fmt.Sprintf("linkweight:%s", noteID)

	opts := []asynq.Option{
		asynq.MaxRetry(3),
		asynq.Timeout(120 * time.Second),
		asynq.Queue("default"),
		asynq.TaskID(uniqueKey),
	}

	if delay > 0 {
		opts = append(opts, asynq.ProcessIn(delay))
	}

	return asynq.NewTask(TypeRecalculateLinkWeights, payload, opts...), nil
}

// LinkWeightRecalculator defines the interface for the link weight recalculation service.
type LinkWeightRecalculator interface {
	RecalculateForNote(ctx context.Context, noteID uuid.UUID) error
}

// HandleRecalculateLinkWeights is the handler for TypeRecalculateLinkWeights tasks.
func HandleRecalculateLinkWeights(ctx context.Context, t *asynq.Task, svc LinkWeightRecalculator) error {
	var p RecalculateLinkWeightsPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("[Asynq] Recalculating link weights for note %s", p.NoteID)
	if err := svc.RecalculateForNote(ctx, p.NoteID); err != nil {
		log.Printf("[Asynq] Failed to recalculate link weights for note %s: %v", p.NoteID, err)
		return err
	}
	log.Printf("[Asynq] Finished recalculating link weights for note %s", p.NoteID)
	return nil
}
