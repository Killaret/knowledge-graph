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
	client        *asynq.Client
	backupEnabled bool
}

// NewAsynqClient creates a new asynq client.
// redisAddr is the Redis address, e.g. "localhost:6379".
// backupEnabled controls whether backup tasks are enqueued.
func NewAsynqClient(redisAddr string, backupEnabled bool) (*AsynqClient, error) {
	redisAddr = strings.TrimPrefix(redisAddr, "redis://")
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return &AsynqClient{client: client, backupEnabled: backupEnabled}, nil
}

func (c *AsynqClient) EnqueueBackupToCloud(ctx context.Context, localPath, remoteKey, backupDate string) error {
	task, err := tasks.NewBackupToCloudTask(localPath, remoteKey, backupDate)
	if err != nil {
		return err
	}
	_, err = c.client.EnqueueContext(ctx, task)
	return err
}

func (c *AsynqClient) EnqueueBackupOnNoteChange(ctx context.Context) error {
	if !c.backupEnabled {
		return nil
	}
	task, err := tasks.NewDatabaseBackupTask()
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

func (c *AsynqClient) EnqueueRecalculateLinkWeights(ctx context.Context, noteID uuid.UUID, delay time.Duration) error {
	task, err := tasks.NewRecalculateLinkWeightsTask(noteID, delay)
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

// EnqueueComputeEmbedding schedules embedding computation for a note.
func (c *AsynqClient) EnqueueComputeEmbedding(ctx context.Context, noteID string) error {
	return c.EnqueueComputeEmbeddingDelayed(ctx, noteID, 0)
}

// EnqueueComputeEmbeddingDelayed schedules embedding computation with a delay.
func (c *AsynqClient) EnqueueComputeEmbeddingDelayed(ctx context.Context, noteID string, delay time.Duration) error {
	log.Printf("EnqueueComputeEmbedding called for note %s (delay=%v)", noteID, delay)
	payload, err := json.Marshal(ComputeEmbeddingTaskPayload{NoteID: noteID})
	if err != nil {
		log.Printf("Marshal error: %v", err)
		return err
	}
	var opts []asynq.Option
	if delay > 0 {
		opts = append(opts, asynq.ProcessIn(delay))
	}
	task := asynq.NewTask(TypeComputeEmbedding, payload, opts...)
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

// EnqueueImportBookmarks schedules an async batch import of captured web pages.
func (c *AsynqClient) EnqueueImportBookmarks(ctx context.Context, userID uuid.UUID, taskID string, items []byte) error {
	payload, err := json.Marshal(ImportBookmarksPayload{
		TaskID: taskID,
		UserID: userID.String(),
		Items:  items,
	})
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypeImportBookmarks, payload, asynq.TaskID(taskID), asynq.MaxRetry(3), asynq.Timeout(10*time.Minute))
	_, err = c.client.EnqueueContext(ctx, task)
	return err
}

// Close closes the client.
func (c *AsynqClient) Close() error {
	return c.client.Close()
}
