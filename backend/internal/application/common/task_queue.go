package common

import (
	"context"

	"github.com/hibiken/asynq"
)

// TaskQueue интерфейс для постановки асинхронных задач.
type TaskQueue interface {
	// Enqueue ставит произвольную задачу в очередь.
	Enqueue(ctx context.Context, task *asynq.Task) error
	// EnqueueExtractKeywords ставит задачу извлечения ключевых слов для заметки.
	EnqueueExtractKeywords(ctx context.Context, noteID string, topN int) error
	// EnqueueComputeEmbedding ставит задачу вычисления эмбеддинга для заметки.
	EnqueueComputeEmbedding(ctx context.Context, noteID string) error
}
