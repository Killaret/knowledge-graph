//go:build !integration
// +build !integration

package queue

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Criterion 6: pipeline disabled → EnqueueAssessQuality returns without
// touching Redis (client built against an unreachable address).
func TestAsynqClient_QualityGatedByFlag(t *testing.T) {
	client, err := NewAsynqClient("127.0.0.1:1", false, false, false)
	require.NoError(t, err)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	require.NoError(t, client.EnqueueAssessQuality(ctx, uuid.New().String(), "auto"))
}

// Flag off → nothing reaches the queue even though Redis is up.
func TestAsynqClient_QualityNotEnqueuedWhenDisabled(t *testing.T) {
	mr := miniredis.RunT(t)
	client, err := NewAsynqClient(mr.Addr(), false, false, false)
	require.NoError(t, err)
	defer client.Close()

	require.NoError(t, client.EnqueueAssessQuality(context.Background(), uuid.New().String(), "manual"))

	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: mr.Addr()})
	defer inspector.Close()
	// The disabled path must not even create the queue.
	queues, err := inspector.Queues()
	require.NoError(t, err)
	assert.NotContains(t, queues, "default")
}

// Flag on → the assess task actually lands in the queue.
func TestAsynqClient_QualityEnqueuedWhenEnabled(t *testing.T) {
	mr := miniredis.RunT(t)
	client, err := NewAsynqClient(mr.Addr(), false, false, true)
	require.NoError(t, err)
	defer client.Close()

	noteID := uuid.New().String()
	require.NoError(t, client.EnqueueAssessQuality(context.Background(), noteID, "manual"))

	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: mr.Addr()})
	defer inspector.Close()
	tasks, err := inspector.ListPendingTasks("default")
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, TypeAssessQuality, tasks[0].Type)
	var payload AssessQualityPayload
	require.NoError(t, json.Unmarshal(tasks[0].Payload, &payload))
	assert.Equal(t, noteID, payload.NoteID)
	assert.Equal(t, "manual", payload.Trigger)
}
