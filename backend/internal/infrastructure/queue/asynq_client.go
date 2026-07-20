package queue

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"knowledge-graph/internal/infrastructure/queue/tasks"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// AsynqClient implements the common.TaskQueue port using asynq.
type AsynqClient struct {
	client *asynq.Client
}

// NewAsynqClient creates a new asynq client.
// redisAddr is the Redis address, e.g. "localhost:6379".
func NewAsynqClient(redisAddr string) (*AsynqClient, error) {
	redisAddr = strings.TrimPrefix(redisAddr, "redis://")
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return &AsynqClient{client: client}, nil
}

func (c *AsynqClient) EnqueueBackupToCloud(ctx context.Context, localPath, remoteKey, backupDate string) error {
	task, err := tasks.NewBackupToCloudTask(localPath, remoteKey, backupDate)
	if err != nil {
		return err
	}
	_, err = c.client.EnqueueContext(ctx, task)
	return err
}

func (c *AsynqClient) EnqueueRefreshRecommendations(ctx context.Context, noteID uuid.UUID, delay time.Duration) error {
	task, err := tasks.NewRefreshRecommendationsTask(noteID, delay)
	if err != nil {
		return err
	}
	_, err = c.client.EnqueueContext(ctx, task)
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
	info, err := c.client.EnqueueContext(ctx, task)
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
	info, err := c.client.EnqueueContext(ctx, task)
	if err != nil {
		log.Printf("Enqueue error: %v", err)
	} else {
		log.Printf("Task enqueued: %+v", info)
	}
	return err
}

func (c *AsynqClient) EnqueueNotification(ctx context.Context, payload []byte) error {
	task := asynq.NewTask(TypeNotificationAchievement, payload)
	_, err := c.client.EnqueueContext(ctx, task)
	return err
}

// Close closes the client.
func (c *AsynqClient) Close() error {
	return c.client.Close()
}
