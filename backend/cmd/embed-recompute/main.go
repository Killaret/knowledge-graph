package main

import (
	"context"
	"flag"
	"log"
	"time"

	"knowledge-graph/internal/config"
	"knowledge-graph/internal/infrastructure/db"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/infrastructure/queue"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "Print tasks that would be enqueued without actually enqueuing them")
	batchDelay := flag.Int("batch-delay", 30, "Delay in seconds for batch processing when more than 1000 notes")
	flag.Parse()

	log.Println("Embedding recompute CLI")
	log.Println("=======================")

	cfg, err := config.Load()
	if err != nil {
		log.Printf("FATAL: Failed to load configuration: %v", err)
		return
	}
	log.Printf("Configuration loaded: Model=%s, Database=%s, Redis=%s",
		cfg.NLPModelName, cfg.DatabaseURL, cfg.RedisURL)

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	log.Println("Database connected successfully")

	ctx := context.Background()
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
}
