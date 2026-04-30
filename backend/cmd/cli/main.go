package main

import (
	"context"
	"flag"
	"log"
	"time"

	"knowledge-graph/internal/config"
	"knowledge-graph/internal/infrastructure/db"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/infrastructure/queue/tasks"

	"github.com/hibiken/asynq"
)

func main() {
	// Parse command line flags
	dryRun := flag.Bool("dry-run", false, "Print tasks that would be enqueued without actually enqueuing them")
	batchDelay := flag.Int("batch-delay", 30, "Delay in seconds for batch processing when more than 1000 notes")
	flag.Parse()

	log.Println("Recommendation Precomputation CLI")
	log.Println("=================================")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Printf("FATAL: Failed to load configuration: %v", err)
		return
	}
	log.Printf("Configuration loaded: RedisURL=%s, TopN=%d, TaskDelay=%ds",
		cfg.RedisURL, cfg.RecommendationTopN, cfg.RecommendationTaskDelaySeconds)

	// Initialize database
	db.Init()
	if db.DB == nil {
		log.Fatal("Database connection failed")
	}
	log.Println("Database connected successfully")

	// Create repositories
	noteRepo := postgres.NewNoteRepository(db.DB, nil)
	ctx := context.Background()

	// Fetch all notes
	log.Println("Fetching all notes from database...")
	notes, err := noteRepo.FindAll(ctx)
	if err != nil {
		log.Fatalf("Failed to fetch notes: %v", err)
	}
	log.Printf("Found %d notes to process", len(notes))

	if len(notes) == 0 {
		log.Println("No notes found. Exiting.")
		return
	}

	// Determine delay based on batch size
	delay := time.Duration(cfg.RecommendationTaskDelaySeconds) * time.Second
	if len(notes) > 1000 {
		delay = time.Duration(*batchDelay) * time.Second
		log.Printf("Large batch detected (%d notes). Using increased delay: %v", len(notes), delay)
	}

	if *dryRun {
		log.Println("DRY RUN MODE - No tasks will be enqueued")
		for _, note := range notes {
			log.Printf("Would enqueue: note_id=%s, delay=%v", note.ID(), delay)
		}
		log.Printf("Total tasks that would be enqueued: %d", len(notes))
		return
	}

	// Create Asynq client
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.RedisURL})
	defer client.Close()
	log.Println("Asynq client connected to Redis")

	// Enqueue tasks
	enqueued := 0
	failed := 0

	for i, note := range notes {
		task, err := tasks.NewRefreshRecommendationsTask(note.ID(), delay)
		if err != nil {
			log.Printf("Failed to create task for note %s: %v", note.ID(), err)
			failed++
			continue
		}

		info, err := client.Enqueue(task)
		if err != nil {
			log.Printf("Failed to enqueue task for note %s: %v", note.ID(), err)
			failed++
			continue
		}

		enqueued++
		if (i+1)%100 == 0 || i == len(notes)-1 {
			log.Printf("Progress: %d/%d tasks enqueued (ID: %s, Queue: %s)",
				i+1, len(notes), info.ID, info.Queue)
		}

		// Small sleep to avoid overwhelming Redis
		if (i+1)%50 == 0 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	log.Println("=================================")
	log.Printf("Completed: %d tasks enqueued, %d failed", enqueued, failed)
	log.Println("Recommendations will be computed in the background by workers")
	log.Println("Monitor the queue using Redis or asynqmon for progress")
}
