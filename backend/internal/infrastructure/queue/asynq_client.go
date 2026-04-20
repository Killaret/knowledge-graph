package queue

import (
	"context"
	"encoding/json"
	"log"
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
	return &AsynqClient{client: client}, nil
}

func (c *AsynqClient) Enqueue(ctx context.Context, task *asynq.Task) error {
	_, err := c.client.EnqueueContext(ctx, task)
	return err
}

func (c *AsynqClient) EnqueueExtractKeywords(ctx context.Context, noteID string, topN int) error {
	log.Printf("EnqueueExtractKeywords called for note %s", noteID)
	payload, err := json.Marshal(ExtractKeywordsTaskPayload{NoteID: noteID, TopN: topN})
	if err != nil {
		log.Printf("Marshal error: %v", err)
		return err
	}
	task := asynq.NewTask(TypeExtractKeywords, payload)
	info, err := c.client.Enqueue(task)
	if err != nil {
		log.Printf("Enqueue error: %v", err)
	} else {
		log.Printf("Task enqueued: %+v", info)
	}
	return err
}

func (c *AsynqClient) EnqueueComputeEmbedding(ctx context.Context, noteID string) error {
	log.Printf("EnqueueComputeEmbedding called for note %s", noteID)
	payload, err := json.Marshal(ComputeEmbeddingTaskPayload{NoteID: noteID})
	if err != nil {
		log.Printf("Marshal error: %v", err)
		return err
	}
	task := asynq.NewTask(TypeComputeEmbedding, payload)
	info, err := c.client.Enqueue(task)
	if err != nil {
		log.Printf("Enqueue error: %v", err)
	} else {
		log.Printf("Task enqueued: %+v", info)
	}
	return err
}

// Close закрывает клиент.
func (c *AsynqClient) Close() error {
	return c.client.Close()
}
