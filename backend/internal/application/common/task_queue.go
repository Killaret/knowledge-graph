package common

import (
	"context"

	"github.com/hibiken/asynq"
)

// TaskQueue is the interface for enqueuing asynchronous tasks.
type TaskQueue interface {
	// Enqueue puts an arbitrary task into the queue.
	Enqueue(ctx context.Context, task *asynq.Task) error
	// EnqueueExtractKeywords enqueues a keyword extraction task for a note.
	EnqueueExtractKeywords(ctx context.Context, noteID string, topN int) error
	// EnqueueComputeEmbedding enqueues an embedding computation task for a note.
	EnqueueComputeEmbedding(ctx context.Context, noteID string) error
}
