package common

import "context"

// TaskQueue интерфейс для постановки асинхронных задач.
type TaskQueue interface {
	// EnqueueExtractKeywords ставит задачу извлечения ключевых слов для заметки.
	EnqueueExtractKeywords(ctx context.Context, noteID string, topN int) error
	// EnqueueComputeEmbedding ставит задачу вычисления эмбеддинга для заметки.
	EnqueueComputeEmbedding(ctx context.Context, noteID string) error
}
