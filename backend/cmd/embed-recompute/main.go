package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"time"

	"knowledge-graph/internal/config"
	"knowledge-graph/internal/infrastructure/db"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/infrastructure/queue"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "Print tasks that would be enqueued without actually enqueuing them")
	batchDelay := flag.Int("batch-delay", 30, "Delay in seconds for batch processing when more than 1000 notes")
	post := flag.Bool("post", false, "Run post-recompute steps: reindex ivfflat, clear stale recommendations, invalidate Redis cache, recalculate link weights")
	flag.Parse()

	log.Println("Embedding recompute CLI")
	log.Println("=======================")

	cfg, err := config.Load()
	if err != nil {
		log.Printf("FATAL: Failed to load configuration: %v", err)
		return
	}
	log.Printf("Configuration loaded: Model=%s, Database=%s, Redis=%s",
		cfg.NLPModelName, dbLocation(cfg.DatabaseURL), cfg.RedisURL)

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	log.Println("Database connected successfully")

	ctx := context.Background()

	if *post {
		runPostSteps(ctx, database, cfg, *dryRun)
		return
	}

	embeddingRepo := postgres.NewEmbeddingRepository(database, cfg.NLPModelName)

	missing, err := embeddingRepo.FindNoteIDsMissingModel(ctx)
	if err != nil {
		log.Fatalf("failed to find notes missing embeddings: %v", err)
	}
	log.Printf("Found %d notes without embedding for model %s", len(missing), cfg.NLPModelName)

	if len(missing) == 0 {
		log.Println("Nothing to do. Exiting.")
		return
	}

	delay := time.Duration(cfg.RecommendationTaskDelaySeconds) * time.Second
	if len(missing) > 1000 {
		delay = time.Duration(*batchDelay) * time.Second
		log.Printf("Large batch detected (%d notes). Using increased delay: %v", len(missing), delay)
	}

	if *dryRun {
		log.Println("DRY RUN MODE - No tasks will be enqueued")
		for _, noteID := range missing {
			log.Printf("Would enqueue: note_id=%s, delay=%v", noteID, delay)
		}
		log.Printf("Total tasks that would be enqueued: %d", len(missing))
		return
	}

	taskQueue, err := queue.NewAsynqClient(cfg.RedisURL, cfg.BackupEnabled)
	if err != nil {
		log.Fatalf("Failed to create task queue client: %v", err)
	}
	defer func() {
		if err := taskQueue.Close(); err != nil {
			log.Printf("Error closing task queue client: %v", err)
		}
	}()
	log.Println("Task queue client connected to Redis")

	enqueued := 0
	failed := 0

	for i, noteID := range missing {
		err := taskQueue.EnqueueComputeEmbeddingDelayed(ctx, noteID.String(), delay)
		if err != nil {
			log.Printf("Failed to enqueue task for note %s: %v", noteID, err)
			failed++
			continue
		}

		enqueued++
		if (i+1)%100 == 0 || i == len(missing)-1 {
			log.Printf("Progress: %d/%d tasks enqueued (note_id: %s)",
				i+1, len(missing), noteID)
		}

		if (i+1)%50 == 0 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	log.Println("=======================")
	log.Printf("Completed: %d tasks enqueued, %d failed", enqueued, failed)
	log.Println("Embeddings will be computed in the background by workers")
	log.Println("After workers finish, run again with -post to rebuild the ivfflat index, refresh recommendations and recalculate link weights")
}

// dbLocation returns host and database name from a connection URL for logging,
// never including credentials.
func dbLocation(databaseURL string) string {
	u, err := url.Parse(databaseURL)
	if err != nil || u.Host == "" {
		return "(unparsable)"
	}
	return u.Host + u.Path
}

// runPostSteps performs the follow-up work required after embeddings were
// recomputed for a new model: it is a separate invocation because the actual
// recompute is asynchronous and only finishes once the queue drains.
func runPostSteps(ctx context.Context, database *gorm.DB, cfg *config.Config, dryRun bool) {
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
		taskQueue, err := queue.NewAsynqClient(cfg.RedisURL, cfg.BackupEnabled)
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
