package events

import (
	"context"
	"encoding/json"
	"testing"
	"time"

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

	sub := client.Subscribe(ctx, "graph:events")
	defer sub.Close()

	// Wait for subscription confirmation
	for {
		msg, err := sub.ReceiveTimeout(ctx, 1*time.Second)
		if err != nil {
			break
		}
		if _, ok := msg.(*redis.Subscription); ok {
			break
		}
	}

	err := publisher.PublishNoteCreated(ctx, "note-123", "user-456")
	require.NoError(t, err)

	// Wait for actual message
	msg, err := sub.ReceiveTimeout(ctx, 2*time.Second)
	require.NoError(t, err)

	// Handle the message correctly
	if pubSubMsg, ok := msg.(*redis.Message); ok {
		var parsedEvent Event
		err = json.Unmarshal([]byte(pubSubMsg.Payload), &parsedEvent)
		require.NoError(t, err)
		assert.Equal(t, "NoteCreated", parsedEvent.Event)
		assert.NotEmpty(t, parsedEvent.EventID)

		var payload NoteEventPayload
		err = json.Unmarshal(parsedEvent.Payload, &payload)
		require.NoError(t, err)
		assert.Equal(t, "note-123", payload.NoteID)
		assert.Equal(t, "user-456", payload.UserID)
	} else {
		t.Fatalf("Expected Message, got %T", msg)
	}
}

func TestPublisher_PublishNoteUpdated(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	publisher := NewPublisher(client, "graph:events")
	ctx := context.Background()

	sub := client.Subscribe(ctx, "graph:events")
	defer sub.Close()

	// Wait for subscription confirmation
	for {
		msg, err := sub.ReceiveTimeout(ctx, 1*time.Second)
		if err != nil {
			break
		}
		if _, ok := msg.(*redis.Subscription); ok {
			break
		}
	}

	err := publisher.PublishNoteUpdated(ctx, "note-123", "user-456")
	require.NoError(t, err)

	msg, err := sub.ReceiveTimeout(ctx, 2*time.Second)
	require.NoError(t, err)

	if pubSubMsg, ok := msg.(*redis.Message); ok {
		var parsedEvent Event
		err = json.Unmarshal([]byte(pubSubMsg.Payload), &parsedEvent)
		require.NoError(t, err)
		assert.Equal(t, "NoteUpdated", parsedEvent.Event)

		var payload NoteEventPayload
		err = json.Unmarshal(parsedEvent.Payload, &payload)
		require.NoError(t, err)
		assert.Equal(t, "note-123", payload.NoteID)
	} else {
		t.Fatalf("Expected Message, got %T", msg)
	}
}

func TestPublisher_PublishNoteDeleted(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	publisher := NewPublisher(client, "graph:events")
	ctx := context.Background()

	sub := client.Subscribe(ctx, "graph:events")
	defer sub.Close()

	// Wait for subscription confirmation
	for {
		msg, err := sub.ReceiveTimeout(ctx, 1*time.Second)
		if err != nil {
			break
		}
		if _, ok := msg.(*redis.Subscription); ok {
			break
		}
	}

	err := publisher.PublishNoteDeleted(ctx, "note-123", "user-456")
	require.NoError(t, err)

	msg, err := sub.ReceiveTimeout(ctx, 2*time.Second)
	require.NoError(t, err)

	if pubSubMsg, ok := msg.(*redis.Message); ok {
		var parsedEvent Event
		err = json.Unmarshal([]byte(pubSubMsg.Payload), &parsedEvent)
		require.NoError(t, err)
		assert.Equal(t, "NoteDeleted", parsedEvent.Event)

		var payload NoteEventPayload
		err = json.Unmarshal(parsedEvent.Payload, &payload)
		require.NoError(t, err)
		assert.Equal(t, "note-123", payload.NoteID)
	} else {
		t.Fatalf("Expected Message, got %T", msg)
	}
}

func TestPublisher_PublishLinkCreated(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	publisher := NewPublisher(client, "graph:events")
	ctx := context.Background()

	sub := client.Subscribe(ctx, "graph:events")
	defer sub.Close()

	// Wait for subscription confirmation
	for {
		msg, err := sub.ReceiveTimeout(ctx, 1*time.Second)
		if err != nil {
			break
		}
		if _, ok := msg.(*redis.Subscription); ok {
			break
		}
	}

	err := publisher.PublishLinkCreated(ctx, "source-123", "target-456", "user-789")
	require.NoError(t, err)

	msg, err := sub.ReceiveTimeout(ctx, 2*time.Second)
	require.NoError(t, err)

	if pubSubMsg, ok := msg.(*redis.Message); ok {
		var parsedEvent Event
		err = json.Unmarshal([]byte(pubSubMsg.Payload), &parsedEvent)
		require.NoError(t, err)
		assert.Equal(t, "LinkCreated", parsedEvent.Event)

		var payload LinkEventPayload
		err = json.Unmarshal(parsedEvent.Payload, &payload)
		require.NoError(t, err)
		assert.Equal(t, "source-123", payload.SourceNoteID)
		assert.Equal(t, "target-456", payload.TargetNoteID)
		assert.Equal(t, "user-789", payload.UserID)
	} else {
		t.Fatalf("Expected Message, got %T", msg)
	}
}

func TestPublisher_PublishLinkDeleted(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	publisher := NewPublisher(client, "graph:events")
	ctx := context.Background()

	sub := client.Subscribe(ctx, "graph:events")
	defer sub.Close()

	// Wait for subscription confirmation
	for {
		msg, err := sub.ReceiveTimeout(ctx, 1*time.Second)
		if err != nil {
			break
		}
		if _, ok := msg.(*redis.Subscription); ok {
			break
		}
	}

	err := publisher.PublishLinkDeleted(ctx, "source-123", "target-456", "user-789")
	require.NoError(t, err)

	msg, err := sub.ReceiveTimeout(ctx, 2*time.Second)
	require.NoError(t, err)

	if pubSubMsg, ok := msg.(*redis.Message); ok {
		var parsedEvent Event
		err = json.Unmarshal([]byte(pubSubMsg.Payload), &parsedEvent)
		require.NoError(t, err)
		assert.Equal(t, "LinkDeleted", parsedEvent.Event)

		var payload LinkEventPayload
		err = json.Unmarshal(parsedEvent.Payload, &payload)
		require.NoError(t, err)
		assert.Equal(t, "source-123", payload.SourceNoteID)
		assert.Equal(t, "target-456", payload.TargetNoteID)
	} else {
		t.Fatalf("Expected Message, got %T", msg)
	}
}
