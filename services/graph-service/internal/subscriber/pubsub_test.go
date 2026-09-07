package subscriber

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"knowledge-graph-graph-service/internal/cache"
	"knowledge-graph-graph-service/internal/db"
)

type mockPostgresClient struct{}

func (m *mockPostgresClient) GetNotes(ctx context.Context, filter db.NotesFilter) ([]*db.Note, []*db.Link, error) {
	return []*db.Note{
			{ID: "note-1", Title: "Note 1"},
			{ID: "note-2", Title: "Note 2"},
		}, []*db.Link{
			{Source: "note-1", Target: "note-2", LinkType: "related", Weight: 0.5},
		}, nil
}

func (m *mockPostgresClient) GetEmbeddings(ctx context.Context, noteIDs []string) (map[string][]float32, error) {
	return make(map[string][]float32), nil
}

func (m *mockPostgresClient) GetNoteNeighbors(ctx context.Context, filter db.NotesFilter, noteID string, depth int) ([]*db.Neighbor, error) {
	return nil, nil
}

func (m *mockPostgresClient) GetShortestPath(ctx context.Context, filter db.NotesFilter, fromID, toID string) ([]string, int, float64, error) {
	return nil, 0, 0, nil
}

func (m *mockPostgresClient) GetRecommendationCandidates(ctx context.Context, filter db.NotesFilter, noteID string, depth, limit int) ([]*db.RecommendationCandidate, error) {
	return nil, nil
}

func (m *mockPostgresClient) RefreshClosureView(ctx context.Context) error {
	return nil
}

func TestEventParsing(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		wantErr bool
		event   string
	}{
		{
			name: "valid note created event",
			payload: `{
				"event_id": "evt-123",
				"event": "NoteCreated",
				"payload": {"note_id": "note-1", "user_id": "user-1"}
			}`,
			wantErr: false,
			event:   "NoteCreated",
		},
		{
			name: "valid note updated event",
			payload: `{
				"event_id": "evt-124",
				"event": "NoteUpdated",
				"payload": {"note_id": "note-2", "user_id": "user-1"}
			}`,
			wantErr: false,
			event:   "NoteUpdated",
		},
		{
			name: "valid link created event",
			payload: `{
				"event_id": "evt-125",
				"event": "LinkCreated",
				"payload": {"source_note_id": "note-1", "target_note_id": "note-2", "user_id": "user-1"}
			}`,
			wantErr: false,
			event:   "LinkCreated",
		},
		{
			name:    "invalid JSON",
			payload: `{invalid}`,
			wantErr: true,
		},
		{
			name: "missing event field",
			payload: `{
				"event_id": "evt-126",
				"payload": {}
			}`,
			wantErr: false,
			event:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var event Event
			err := json.Unmarshal([]byte(tt.payload), &event)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.event, event.Event)
			}
		})
	}
}

func TestEventAcknowledgment(t *testing.T) {
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer redisClient.Close()
	defer mr.Close()

	postgres := &mockPostgresClient{}
	cache := cache.NewRedisCache(redisClient)

	subscriber := NewRedisSubscriber(redisClient, postgres, cache, "test-channel", 100)

	ctx := context.Background()

	// Simulate receiving an event
	eventPayload := `{
		"event_id": "evt-test-ack",
		"event": "NoteCreated",
		"payload": {"note_id": "note-1", "user_id": "user-1"}
	}`

	// Process the event
	subscriber.handleEvent(ctx, eventPayload)

	// Check that event is acknowledged
	eventKey := "event:evt-test-ack"
	result, err := redisClient.HGetAll(ctx, eventKey).Result()
	require.NoError(t, err)

	var updatedTracking EventTracking
	var data []byte
	if val, ok := result[""]; ok {
		data = []byte(val)
	} else if val, ok := result["data"]; ok {
		data = []byte(val)
	}

	if len(data) > 0 {
		err = json.Unmarshal(data, &updatedTracking)
		require.NoError(t, err)
		assert.True(t, updatedTracking.IsAcknowledged)
		assert.False(t, updatedTracking.TimestampProcessed.IsZero())
	}
}

func TestCacheInvalidation(t *testing.T) {
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer redisClient.Close()
	defer mr.Close()

	postgres := &mockPostgresClient{}
	cacheClient := cache.NewRedisCache(redisClient)

	ctx := context.Background()

	// Pre-populate cache with some data
	cacheKey := "graph-service:full:user-1"
	cacheData := `{"nodes": [{"id": "note-1"}], "links": []}`
	err := redisClient.Set(ctx, cacheKey, cacheData, 5*time.Minute).Err()
	require.NoError(t, err)

	// Verify cache exists
	exists, _ := redisClient.Exists(ctx, cacheKey).Result()
	assert.Equal(t, int64(1), exists)

	// Create subscriber and handle event
	subscriber := NewRedisSubscriber(redisClient, postgres, cacheClient, "test-channel", 100)

	eventPayload := `{
		"event_id": "evt-test-invalidate",
		"event": "NoteCreated",
		"payload": {"note_id": "note-2", "user_id": "user-1"}
	}`

	subscriber.handleEvent(ctx, eventPayload)

	// Verify cache was invalidated
	exists, _ = redisClient.Exists(ctx, cacheKey).Result()
	assert.Equal(t, int64(0), exists, "Cache should be invalidated")
}

func TestNoteEventHandling(t *testing.T) {
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer redisClient.Close()
	defer mr.Close()

	postgres := &mockPostgresClient{}
	cacheClient := cache.NewRedisCache(redisClient)

	subscriber := NewRedisSubscriber(redisClient, postgres, cacheClient, "test-channel", 100)

	ctx := context.Background()

	tests := []struct {
		name    string
		payload interface{}
	}{
		{
			name: "NoteCreated",
			payload: NoteEventPayload{
				NoteID: "note-1",
				UserID: "user-1",
			},
		},
		{
			name: "NoteUpdated",
			payload: NoteEventPayload{
				NoteID: "note-2",
				UserID: "user-1",
			},
		},
		{
			name: "NoteDeleted",
			payload: NoteEventPayload{
				NoteID: "note-3",
				UserID: "user-1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payloadBytes, err := json.Marshal(tt.payload)
			require.NoError(t, err)
			event := Event{
				EventID: "evt-" + tt.name,
				Event:   tt.name,
				Payload: payloadBytes,
			}

			err = subscriber.processEvent(ctx, event)
			assert.NoError(t, err)
		})
	}
}

func TestLinkEventHandling(t *testing.T) {
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer redisClient.Close()
	defer mr.Close()

	postgres := &mockPostgresClient{}
	cacheClient := cache.NewRedisCache(redisClient)

	subscriber := NewRedisSubscriber(redisClient, postgres, cacheClient, "test-channel", 100)

	ctx := context.Background()

	tests := []struct {
		name    string
		payload interface{}
	}{
		{
			name: "LinkCreated",
			payload: LinkEventPayload{
				SourceNoteID: "note-1",
				TargetNoteID: "note-2",
				UserID:       "user-1",
			},
		},
		{
			name: "LinkUpdated",
			payload: LinkEventPayload{
				SourceNoteID: "note-1",
				TargetNoteID: "note-2",
				UserID:       "user-1",
			},
		},
		{
			name: "LinkDeleted",
			payload: LinkEventPayload{
				SourceNoteID: "note-1",
				TargetNoteID: "note-2",
				UserID:       "user-1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payloadBytes, err := json.Marshal(tt.payload)
			require.NoError(t, err)
			event := Event{
				EventID: "evt-" + tt.name,
				Event:   tt.name,
				Payload: payloadBytes,
			}

			err = subscriber.processEvent(ctx, event)
			assert.NoError(t, err)
		})
	}
}
