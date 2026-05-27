package events

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublisher_PublishNoteCreated(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	publisher := NewPublisher(client, "graph:events")
	ctx := context.Background()

	err := publisher.PublishNoteCreated(ctx, "note-123", "user-456")
	require.NoError(t, err)

	// Verify the message was published
	subs := s.Subscriptions()
	require.Len(t, subs, 1)
	assert.Equal(t, "graph:events", subs[0])

	msgs := s.Messages("graph:events")
	require.Len(t, msgs, 1)

	var event Event
	err = json.Unmarshal([]byte(msgs[0]), &event)
	require.NoError(t, err)
	assert.Equal(t, "NoteCreated", event.Event)
	assert.NotEmpty(t, event.EventID)

	var payload NoteEventPayload
	err = json.Unmarshal(event.Payload, &payload)
	require.NoError(t, err)
	assert.Equal(t, "note-123", payload.NoteID)
	assert.Equal(t, "user-456", payload.UserID)
}

func TestPublisher_PublishNoteUpdated(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	publisher := NewPublisher(client, "graph:events")
	ctx := context.Background()

	err := publisher.PublishNoteUpdated(ctx, "note-123", "user-456")
	require.NoError(t, err)

	var event Event
	msgs := s.Messages("graph:events")
	require.Len(t, msgs, 1)

	err = json.Unmarshal([]byte(msgs[0]), &event)
	require.NoError(t, err)
	assert.Equal(t, "NoteUpdated", event.Event)

	var payload NoteEventPayload
	err = json.Unmarshal(event.Payload, &payload)
	require.NoError(t, err)
	assert.Equal(t, "note-123", payload.NoteID)
}

func TestPublisher_PublishNoteDeleted(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	publisher := NewPublisher(client, "graph:events")
	ctx := context.Background()

	err := publisher.PublishNoteDeleted(ctx, "note-123", "user-456")
	require.NoError(t, err)

	var event Event
	msgs := s.Messages("graph:events")
	require.Len(t, msgs, 1)

	err = json.Unmarshal([]byte(msgs[0]), &event)
	require.NoError(t, err)
	assert.Equal(t, "NoteDeleted", event.Event)

	var payload NoteEventPayload
	err = json.Unmarshal(event.Payload, &payload)
	require.NoError(t, err)
	assert.Equal(t, "note-123", payload.NoteID)
}

func TestPublisher_PublishLinkCreated(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	publisher := NewPublisher(client, "graph:events")
	ctx := context.Background()

	err := publisher.PublishLinkCreated(ctx, "source-123", "target-456", "user-789")
	require.NoError(t, err)

	var event Event
	msgs := s.Messages("graph:events")
	require.Len(t, msgs, 1)

	err = json.Unmarshal([]byte(msgs[0]), &event)
	require.NoError(t, err)
	assert.Equal(t, "LinkCreated", event.Event)

	var payload LinkEventPayload
	err = json.Unmarshal(event.Payload, &payload)
	require.NoError(t, err)
	assert.Equal(t, "source-123", payload.SourceNoteID)
	assert.Equal(t, "target-456", payload.TargetNoteID)
	assert.Equal(t, "user-789", payload.UserID)
}

func TestPublisher_PublishLinkDeleted(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	publisher := NewPublisher(client, "graph:events")
	ctx := context.Background()

	err := publisher.PublishLinkDeleted(ctx, "source-123", "target-456", "user-789")
	require.NoError(t, err)

	var event Event
	msgs := s.Messages("graph:events")
	require.Len(t, msgs, 1)

	err = json.Unmarshal([]byte(msgs[0]), &event)
	require.NoError(t, err)
	assert.Equal(t, "LinkDeleted", event.Event)

	var payload LinkEventPayload
	err = json.Unmarshal(event.Payload, &payload)
	require.NoError(t, err)
	assert.Equal(t, "source-123", payload.SourceNoteID)
	assert.Equal(t, "target-456", payload.TargetNoteID)
}
