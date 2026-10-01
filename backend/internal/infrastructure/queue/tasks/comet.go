package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// CometRemindPayload carries the note and the exact reminder moment the task
// was scheduled for; a mismatch at fire time means the note was rescheduled.
type CometRemindPayload struct {
	NoteID   string    `json:"note_id"`
	RemindAt time.Time `json:"remind_at"`
}

// NewCometRemindTask builds a comet:remind task (COMET-1 stage C).
func NewCometRemindTask(noteID uuid.UUID, remindAt time.Time) (*asynq.Task, error) {
	payload, err := json.Marshal(CometRemindPayload{NoteID: noteID.String(), RemindAt: remindAt})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask("comet:remind", payload), nil
}

// CometReminderService is the application service invoked by the comet:remind
// task (COMET-1 stage C).
type CometReminderService interface {
	HandleReminder(ctx context.Context, noteID uuid.UUID, expectedRemindAt time.Time) error
}

// HandleCometRemind unmarshals the payload and dispatches to the service.
// Staleness and dedupe live in the service — a stale task is a no-op.
func HandleCometRemind(ctx context.Context, t *asynq.Task, svc CometReminderService) error {
	var p CometRemindPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal comet remind payload: %w", err)
	}
	noteID, err := uuid.Parse(p.NoteID)
	if err != nil {
		return fmt.Errorf("comet remind note_id %q: %w", p.NoteID, err)
	}
	return svc.HandleReminder(ctx, noteID, p.RemindAt)
}
