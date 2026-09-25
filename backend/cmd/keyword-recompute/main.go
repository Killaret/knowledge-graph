package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"knowledge-graph/internal/config"
	"knowledge-graph/internal/infrastructure/db"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/infrastructure/queue"
	"knowledge-graph/internal/infrastructure/recompute"
)

// healthResponse is the subset of the nlp-service /health payload we need:
// the name of the extractor currently deployed, so the recompute filter
// targets exactly the rows it would produce.
type healthResponse struct {
	Extractor string `json:"extractor"`
}

func currentExtractor(nlpURL string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(nlpURL + "/health")
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("nlp-service health returned %s", resp.Status)
	}
	var h healthResponse
	if err := json.NewDecoder(resp.Body).Decode(&h); err != nil {
		return "", err
	}
	if h.Extractor == "" {
		return "", fmt.Errorf("nlp-service did not report an extractor name")
	}
	return h.Extractor, nil
}

func main() {
	dryRun := flag.Bool("dry-run", false, "Print tasks that would be enqueued without actually enqueuing them")
	batchDelay := flag.Int("batch-delay", 30, "Delay in seconds for batch processing when more than 1000 notes")
	post := flag.Bool("post", false, "Run post-recompute steps: reindex ivfflat, clear stale recommendations, invalidate Redis cache, recalculate link weights")
	flag.Parse()

	log.Println("Keyword recompute CLI")
	log.Println("=====================")

	cfg, err := config.Load()
	if err != nil {
		log.Printf("FATAL: Failed to load configuration: %v", err)
		return
	}
	log.Printf("Configuration loaded: NLPService=%s, Database=%s, Redis=%s",
		cfg.NLPServiceURL, recompute.DBLocation(cfg.DatabaseURL), cfg.RedisURL)

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	log.Println("Database connected successfully")

	ctx := context.Background()

	if *post {
		recompute.RunPostSteps(ctx, database, cfg, *dryRun)
		return
	}

	extractor, err := currentExtractor(cfg.NLPServiceURL)
	if err != nil {
		log.Fatalf("failed to query nlp-service extractor: %v", err)
	}
	log.Printf("nlp-service reports extractor %q", extractor)

	keywordRepo := postgres.NewKeywordRepository(database)
	missing, err := keywordRepo.FindNoteIDsMissingExtractor(ctx, extractor)
	if err != nil {
		log.Fatalf("failed to find notes missing keywords: %v", err)
	}
	log.Printf("Found %d notes without keywords for extractor %s", len(missing), extractor)

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

	taskQueue, err := queue.NewAsynqClient(cfg.RedisURL, cfg.BackupEnabled, cfg.NLPPipelineEnabled)
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
		err := taskQueue.EnqueueExtractKeywordsDelayed(ctx, noteID.String(), 0, delay)
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

	log.Println("=====================")
	log.Printf("Completed: %d tasks enqueued, %d failed", enqueued, failed)
	log.Println("Keywords will be extracted in the background by workers")
	log.Println("After workers finish, run again with -post to rebuild the index, refresh recommendations and recalculate link weights")
}
