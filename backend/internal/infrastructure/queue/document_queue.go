package queue

import (
	"context"
	"errors"
	"time"

	importer "knowledge-graph/internal/application/import"

	"github.com/redis/go-redis/v9"
)

// DocumentQueue adapts *redis.Client to the importer.DocumentQueue port:
// the two plain-JSON lists of the source-text-handler contract
// (TZ-Java-source-text-handler-2026-08-30.md §3/§4) and its liveness key.
type DocumentQueue struct {
	client *redis.Client
}

// NewDocumentQueue creates the adapter over an existing Redis client.
func NewDocumentQueue(client *redis.Client) *DocumentQueue {
	return &DocumentQueue{client: client}
}

// IsListenerAlive reports whether the Java listener wrote its liveness key.
func (q *DocumentQueue) IsListenerAlive(ctx context.Context) (bool, error) {
	n, err := q.client.Exists(ctx, importer.DocumentListenerKey).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// PushTask appends one §3 task payload to the pending list.
func (q *DocumentQueue) PushTask(ctx context.Context, payload []byte) error {
	return q.client.RPush(ctx, importer.DocumentQueuePending, payload).Err()
}

// PopResult blocks for one §4 result payload. An empty wait within timeout is
// reported as importer.ErrResultTimeout.
func (q *DocumentQueue) PopResult(ctx context.Context, timeout time.Duration) ([]byte, error) {
	res, err := q.client.BLPop(ctx, timeout, importer.DocumentQueueResponses).Result()
	if errors.Is(err, redis.Nil) {
		return nil, importer.ErrResultTimeout
	}
	if err != nil {
		return nil, err
	}
	if len(res) < 2 {
		return nil, importer.ErrResultTimeout
	}
	return []byte(res[1]), nil
}

var _ importer.DocumentQueue = (*DocumentQueue)(nil)
