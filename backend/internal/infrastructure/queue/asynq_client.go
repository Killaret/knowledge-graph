package queue

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/hibiken/asynq"
)

// AsynqClient реализует интерфейс common.TaskQueue
type AsynqClient struct {
	client *asynq.Client
}

// NewAsynqClient создаёт новый клиент asynq.
// redisAddr — адрес Redis, например "localhost:6379".
func NewAsynqClient(redisAddr string) (*AsynqClient, error) {
	redisAddr = strings.TrimPrefix(redisAddr, "redis://")
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	if err := client.Ping(); err != nil {
		return nil, err
	}
	return &AsynqClient{client: client}, nil
}

func (c *AsynqClient) EnqueueExtractKeywords(ctx context.Context, noteID string, topN int) error {
	payload, err := json.Marshal(ExtractKeywordsTaskPayload{NoteID: noteID, TopN: topN})
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypeExtractKeywords, payload)
	_, err = c.client.Enqueue(task)
	return err
}

func (c *AsynqClient) EnqueueComputeEmbedding(ctx context.Context, noteID string) error {
	payload, err := json.Marshal(ComputeEmbeddingTaskPayload{NoteID: noteID})
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypeComputeEmbedding, payload)
	_, err = c.client.Enqueue(task)
	return err
}

// Close закрывает клиент.
func (c *AsynqClient) Close() error {
	return c.client.Close()
}
