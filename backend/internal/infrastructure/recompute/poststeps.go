// Package recompute holds the shared post-recompute steps used by the
// embed-recompute and keyword-recompute CLI commands (NLP-2): reindex the
// vector index, drop stale recommendations and enqueue their refresh,
// recalculate link weights, invalidate the Redis recommendation cache.
package recompute

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"knowledge-graph/internal/config"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/infrastructure/queue"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// DBLocation returns host and database name from a connection URL for
// logging, never including credentials.
func DBLocation(databaseURL string) string {
	u, err := url.Parse(databaseURL)
	if err != nil || u.Host == "" {
		return "(unparsable)"
	}
	return u.Host + u.Path
}

// RunPostSteps performs the follow-up work required after embeddings or
// keywords were recomputed: it is a separate invocation because the actual
// recompute is asynchronous and only finishes once the queue drains.
func RunPostSteps(ctx context.Context, database *gorm.DB, cfg *config.Config, dryRun bool) {
	log.Println("Post-recompute steps")
	if dryRun {
		log.Println("DRY RUN MODE - nothing will be executed")
	}

	// 1. Rebuild the ivfflat index: it was built over the old vector
	// distribution and degrades after a full recompute.
	if dryRun {
		log.Println("Would reindex idx_note_embeddings_vector")
	} else {
		log.Println("Reindexing idx_note_embeddings_vector ...")
		if err := database.WithContext(ctx).Exec("REINDEX INDEX idx_note_embeddings_vector").Error; err != nil {
			log.Fatalf("reindex failed: %v", err)
		}
		log.Println("Index rebuilt")
	}

	// Collect all live notes once; both follow-up passes need them.
	noteRepo := postgres.NewNoteRepository(database, nil)
	notes, err := noteRepo.FindAll(ctx)
	if err != nil {
		log.Fatalf("failed to list notes: %v", err)
	}
	log.Printf("Post-recompute covers %d notes", len(notes))

	delay := time.Duration(cfg.RecommendationTaskDelaySeconds) * time.Second

	// 2. Drop stale recommendations and enqueue their recompute.
	if dryRun {
		var recCount int64
		if err := database.WithContext(ctx).Raw("SELECT COUNT(*) FROM note_recommendations").Scan(&recCount).Error; err != nil {
			log.Fatalf("failed to count note_recommendations: %v", err)
		}
		log.Printf("Would delete %d rows from note_recommendations and enqueue refresh for %d notes", recCount, len(notes))
	} else {
		taskQueue, err := queue.NewAsynqClient(cfg.RedisURL, cfg.BackupEnabled, cfg.NLPPipelineEnabled, cfg.NLPQualityEnabled)
		if err != nil {
			log.Fatalf("Failed to create task queue client: %v", err)
		}
		defer func() {
			if err := taskQueue.Close(); err != nil {
				log.Printf("Error closing task queue client: %v", err)
			}
		}()

		log.Println("Clearing note_recommendations ...")
		if err := database.WithContext(ctx).Exec("DELETE FROM note_recommendations").Error; err != nil {
			log.Fatalf("failed to clear note_recommendations: %v", err)
		}
		failed := 0
		for _, n := range notes {
			if err := taskQueue.EnqueueRefreshRecommendations(ctx, n.ID(), delay); err != nil {
				log.Printf("Failed to enqueue recommendation refresh for note %s: %v", n.ID(), err)
				failed++
			}
		}
		log.Printf("Recommendations cleared; refresh enqueued for %d notes (%d failed)", len(notes)-failed, failed)

		// 3. Recalculate link weights: they encode semantic similarity computed
		// by the old model.
		failed = 0
		for _, n := range notes {
			if err := taskQueue.EnqueueRecalculateLinkWeights(ctx, n.ID(), delay); err != nil {
				log.Printf("Failed to enqueue link weight recalculation for note %s: %v", n.ID(), err)
				failed++
			}
		}
		log.Printf("Link weight recalculation enqueued for %d notes (%d failed)", len(notes)-failed, failed)
	}
	if dryRun {
		log.Printf("Would enqueue link weight recalculation for %d notes", len(notes))
	}

	// 4. Invalidate the cached recommendation payloads in Redis. The scan is
	// read-only, so it runs in both modes; only the delete is skipped.
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisURL})
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing redis client: %v", err)
		}
	}()
	var cursor uint64
	deleted := 0
	for {
		keys, next, err := redisClient.Scan(ctx, cursor, "recommendations:*", 100).Result()
		if err != nil {
			log.Fatalf("redis scan failed: %v", err)
		}
		if len(keys) > 0 && !dryRun {
			if err := redisClient.Del(ctx, keys...).Err(); err != nil {
				log.Fatalf("redis del failed: %v", err)
			}
		}
		deleted += len(keys)
		cursor = next
		if cursor == 0 {
			break
		}
	}
	if dryRun {
		log.Printf("Would invalidate %d cached recommendation keys", deleted)
	} else {
		log.Printf("Invalidated %d cached recommendation keys", deleted)
	}

	fmt.Println("Post-recompute steps completed")
}
