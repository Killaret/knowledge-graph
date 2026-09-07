package common

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TaskQueue is an application-level port for enqueueing asynchronous tasks.
// Infrastructure adapters (e.g. asynq) implement this interface.
type TaskQueue interface {
	// EnqueueBackupToCloud schedules a cloud backup task.
	EnqueueBackupToCloud(ctx context.Context, localPath, remoteKey, backupDate string) error

	// EnqueueBackupOnNoteChange schedules a database backup after a note is created or modified.
	EnqueueBackupOnNoteChange(ctx context.Context) error

	// EnqueueRefreshRecommendations schedules a recommendation refresh for the given note.
	EnqueueRefreshRecommendations(ctx context.Context, noteID uuid.UUID, delay time.Duration) error

	// EnqueueExtractKeywords schedules keyword extraction for a note.
	EnqueueExtractKeywords(ctx context.Context, noteID string, topN int) error

	// EnqueueComputeEmbedding schedules embedding computation for a note.
	EnqueueComputeEmbedding(ctx context.Context, noteID string) error

	// EnqueueRecalculateLinkWeights schedules link weight recalculation for a note.
	EnqueueRecalculateLinkWeights(ctx context.Context, noteID uuid.UUID, delay time.Duration) error

	// EnqueueNotification schedules a notification task with the given payload.
	EnqueueNotification(ctx context.Context, payload []byte) error

	// EnqueueImportBookmarks schedules an async batch import of captured web pages.
	EnqueueImportBookmarks(ctx context.Context, userID uuid.UUID, taskID string, items []byte) error
}
