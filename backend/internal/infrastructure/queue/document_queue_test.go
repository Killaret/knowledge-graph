package queue

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	importer "knowledge-graph/internal/application/import"
	"knowledge-graph/internal/domain/cache/cachetest"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func newDocumentQueueFixture(t *testing.T) (*miniredis.Miniredis, *importer.Service) {
	t.Helper()
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	svc := importer.NewService(nil, cachetest.NewFakeCacheClient(), nil, nil)
	svc.SetDocumentQueue(NewDocumentQueue(client))
	return s, svc
}

// SMART-CREATE-1 criterion 2 on real Redis: the pushed message parses into the
// same structure as the §3 example — same field set, snake_case, no envelope.
func TestDocumentQueue_TaskContractOnRedis(t *testing.T) {
	s, svc := newDocumentQueueFixture(t)
	ctx := context.Background()
	userID := uuid.New()

	// The §3 fixture verbatim — field names and import_options values come
	// from the contract, not from a hand-typed list (review blocker: a renamed
	// smart_weights field and a changed overlap value passed the old check).
	fixture := `{
		"event_id": "uuid-task-1",
		"correlation_id": "uuid-correlation-1",
		"user_id": "uuid-user-1",
		"jwt": "eyJhbGci...",
		"type": "FILE|URL|TEXT",
		"content": "base64-string | URL | plain text",
		"content_type": "application/pdf",
		"import_options": {
			"chunk_size": 500,
			"overlap": 50,
			"min_chunk_length": 100,
			"max_chunk_length": 8000,
			"max_retries": 3,
			"create_links": false,
			"smart_threshold": 70,
			"smart_weights": {
				"has_heading": 25,
				"section_boundary": 10,
				"intra_chunk_cohesion": 20,
				"punctuation_completeness": 15,
				"chunk_ending_weight": 15,
				"length_ok": 10,
				"keyword_trend_stable": 5,
				"punctuation_balance": 5,
				"metadata_present": 5
			}
		},
		"metadata": {}
	}`
	var fixtureKeys map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(fixture), &fixtureKeys))

	require.False(t, s.Exists(importer.DocumentListenerKey))
	_, err := svc.SubmitDocument(ctx, userID, "jwt", importer.DocumentInput{Type: "TEXT", Content: "x"})
	require.ErrorIs(t, err, importer.ErrDocumentUnavailable)
	list, _ := s.List(importer.DocumentQueuePending)
	require.Empty(t, list)

	require.NoError(t, s.Set(importer.DocumentListenerKey, "2026-10-18T12:00:00Z"))

	available, err := svc.DocumentAvailability(ctx)
	require.NoError(t, err)
	require.True(t, available)

	_, err = svc.SubmitDocument(ctx, userID, "jwt", importer.DocumentInput{Type: "FILE",
		FileContent: []byte("data"), ContentType: "text/plain", Filename: "a.txt"})
	require.NoError(t, err)

	pending, err := s.List(importer.DocumentQueuePending)
	require.NoError(t, err)
	require.Len(t, pending, 1)

	var task map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(pending[0]), &task))
	for key := range fixtureKeys {
		require.Contains(t, task, key)
	}
	require.Len(t, task, len(fixtureKeys), "task has fields the contract does not define: %v", task)

	// import_options on the wire — every key and value equal to §3.
	var wantOptions, gotOptions map[string]interface{}
	require.NoError(t, json.Unmarshal(fixtureKeys["import_options"], &wantOptions))
	require.NoError(t, json.Unmarshal(task["import_options"], &gotOptions))
	require.Equal(t, wantOptions, gotOptions)
}

// BLPOP drains the §4 response list; an empty list answers ErrResultTimeout.
func TestDocumentQueue_PopResult(t *testing.T) {
	s, svc := newDocumentQueueFixture(t)
	ctx := context.Background()

	_ = svc
	dq := NewDocumentQueue(redis.NewClient(&redis.Options{Addr: s.Addr()}))

	_, err := dq.PopResult(ctx, 50*time.Millisecond)
	require.ErrorIs(t, err, importer.ErrResultTimeout)

	_, err = s.RPush(importer.DocumentQueueResponses, `{"correlation_id":"c1","status":"COMPLETED"}`)
	require.NoError(t, err)
	payload, err := dq.PopResult(ctx, time.Second)
	require.NoError(t, err)
	require.JSONEq(t, `{"correlation_id":"c1","status":"COMPLETED"}`, string(payload))
}
