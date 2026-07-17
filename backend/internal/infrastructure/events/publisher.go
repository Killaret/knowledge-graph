package events

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Event represents the structure for graph service events
type Event struct {
	EventID string          `json:"event_id"`
	Event   string          `json:"event"` // NoteCreated, NoteUpdated, NoteDeleted, LinkCreated, LinkUpdated, LinkDeleted
	Payload json.RawMessage `json:"payload"`
}

// NoteEventPayload represents the payload for note events
type NoteEventPayload struct {
	NoteID string `json:"note_id"`
	UserID string `json:"user_id"`
}

// LinkEventPayload represents the payload for link events
type LinkEventPayload struct {
	SourceNoteID string `json:"source_note_id"`
	TargetNoteID string `json:"target_note_id"`
	UserID       string `json:"user_id"`
}

// Publisher handles publishing events to Redis Pub/Sub
type Publisher struct {
	redis   *redis.Client
	channel string
}

// NewPublisher creates a new event publisher
func NewPublisher(redis *redis.Client, channel string) *Publisher {
	return &Publisher{
		redis:   redis,
		channel: channel,
	}
}

// PublishNoteCreated publishes a NoteCreated event
func (p *Publisher) PublishNoteCreated(ctx context.Context, noteID, userID string) error {
	payload := NoteEventPayload{
		NoteID: noteID,
		UserID: userID,
	}
	return p.publishEvent(ctx, "NoteCreated", payload)
}

// PublishNoteUpdated publishes a NoteUpdated event
func (p *Publisher) PublishNoteUpdated(ctx context.Context, noteID, userID string) error {
	payload := NoteEventPayload{
		NoteID: noteID,
		UserID: userID,
	}
	return p.publishEvent(ctx, "NoteUpdated", payload)
}

// PublishNoteDeleted publishes a NoteDeleted event
func (p *Publisher) PublishNoteDeleted(ctx context.Context, noteID, userID string) error {
	payload := NoteEventPayload{
		NoteID: noteID,
		UserID: userID,
	}
	return p.publishEvent(ctx, "NoteDeleted", payload)
}

// PublishLinkCreated publishes a LinkCreated event
func (p *Publisher) PublishLinkCreated(ctx context.Context, sourceNoteID, targetNoteID, userID string) error {
	payload := LinkEventPayload{
		SourceNoteID: sourceNoteID,
		TargetNoteID: targetNoteID,
		UserID:       userID,
	}
	return p.publishEvent(ctx, "LinkCreated", payload)
}

// PublishLinkDeleted publishes a LinkDeleted event
func (p *Publisher) PublishLinkDeleted(ctx context.Context, sourceNoteID, targetNoteID, userID string) error {
	payload := LinkEventPayload{
		SourceNoteID: sourceNoteID,
		TargetNoteID: targetNoteID,
		UserID:       userID,
	}
	return p.publishEvent(ctx, "LinkDeleted", payload)
}

// publishEvent is the internal method to publish events
func (p *Publisher) publishEvent(ctx context.Context, eventType string, payload interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[EventPublisher] Failed to marshal payload for %s: %v", eventType, err)
		return err
	}

	event := Event{
		EventID: uuid.New().String(),
		Event:   eventType,
		Payload: payloadBytes,
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		log.Printf("[EventPublisher] Failed to marshal event %s: %v", eventType, err)
		return err
	}

	// Fire-and-forget: publish without waiting for acknowledgment
	// Use a background context to ensure the publish completes even if the request context is cancelled
	bgCtx := context.Background()
	if err := p.redis.Publish(bgCtx, p.channel, eventBytes).Err(); err != nil {
		log.Printf("[EventPublisher] Failed to publish event %s to channel %s: %v", eventType, p.channel, err)
		return err
	}

	log.Printf("[EventPublisher] Published event %s (ID: %s) to channel %s", eventType, event.EventID, p.channel)
	return nil
}
