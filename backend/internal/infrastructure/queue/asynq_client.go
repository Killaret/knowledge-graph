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
	client             *asynq.Client
	backupEnabled      bool
	nlpPipelineEnabled bool
}

// NewAsynqClient creates a new asynq client.
// redisAddr is the Redis address, e.g. "localhost:6379".
// backupEnabled controls whether backup tasks are enqueued.
// nlpPipelineEnabled gates nlp:normalize tasks (NLP-4 switch, default off).
func NewAsynqClient(redisAddr string, backupEnabled bool, nlpPipelineEnabled bool) (*AsynqClient, error) {
	redisAddr = strings.TrimPrefix(redisAddr, "redis://")
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return &AsynqClient{client: client, backupEnabled: backupEnabled, nlpPipelineEnabled: nlpPipelineEnabled}, nil
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
	return c.EnqueueExtractKeywordsDelayed(ctx, noteID, topN, 0)
}

// EnqueueExtractKeywordsDelayed schedules keyword extraction with a delay.
func (c *AsynqClient) EnqueueExtractKeywordsDelayed(ctx context.Context, noteID string, topN int, delay time.Duration) error {
	log.Printf("EnqueueExtractKeywords called for note %s (delay=%v)", noteID, delay)
	payload, err := json.Marshal(ExtractKeywordsTaskPayload{NoteID: noteID, TopN: topN})
	if err != nil {
		log.Printf("Marshal error: %v", err)
		return err
	}
	var opts []asynq.Option
	if delay > 0 {
		opts = append(opts, asynq.ProcessIn(delay))
	}
	task := asynq.NewTask(TypeExtractKeywords, payload, opts...)
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

// EnqueueNormalizeNote schedules NLP-4 normalization for a note.
// When the pipeline flag is off the call is a no-op — prod behaves exactly
// as before, no task enters the queue.
func (c *AsynqClient) EnqueueNormalizeNote(ctx context.Context, noteID string) error {
	if !c.nlpPipelineEnabled {
		return nil
	}
	payload, err := json.Marshal(NormalizeNotePayload{NoteID: noteID})
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypeNormalizeNote, payload)
	_, err = c.client.EnqueueContext(ctx, task)
	return err
}

// EnqueueNlpArtifactsCleanup schedules removal of a note's nlp_artifacts.
// Not gated by the pipeline flag: cleanup also removes artifacts written
// while the flag was on, so it enqueues whenever the queue is available.
// The worker no-ops when the artifacts store is absent.
func (c *AsynqClient) EnqueueNlpArtifactsCleanup(ctx context.Context, noteID string) error {
	payload, err := json.Marshal(NlpArtifactsCleanupPayload{NoteID: noteID})
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypeNlpArtifactsCleanup, payload)
	_, err = c.client.EnqueueContext(ctx, task)
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
