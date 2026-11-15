// Command gamma-links-regenerate rebuilds automatic (gamma) links for all
// notes that have an embedding for the currently configured model.
//
// Manual (user) links are never touched: only rows with source_type='gamma'
// are deleted before regeneration. --dry-run prints how many links would be
// deleted and created without writing anything.
package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"knowledge-graph/internal/application/recommendation"
	"knowledge-graph/internal/config"
	"knowledge-graph/internal/infrastructure/db"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/infrastructure/events"
	"knowledge-graph/internal/infrastructure/outbox"
	"knowledge-graph/internal/infrastructure/queue"
	"knowledge-graph/internal/infrastructure/recompute"
)

const gammaSourceType = "gamma"

func main() {
	dryRun := flag.Bool("dry-run", false, "Print how many gamma links would be deleted and created without writing")
	flag.Parse()

	log.Println("Gamma links regenerate CLI")
	log.Println("==========================")

	cfg, err := config.Load()
	if err != nil {
		log.Printf("FATAL: Failed to load configuration: %v", err)
		return
	}
	log.Printf("Configuration loaded: Model=%s, Database=%s, Redis=%s, MinScore=%.2f",
		cfg.NLPModelName, recompute.DBLocation(cfg.DatabaseURL), cfg.RedisURL, cfg.GammaLinkMinScore)

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	log.Println("Database connected successfully")

	ctx := context.Background()

	cacheClient := redis.NewClient(&redis.Options{Addr: cfg.RedisURL})
	defer func() {
		if err := cacheClient.Close(); err != nil {
			log.Printf("Error closing redis client: %v", err)
		}
	}()

	// SYNC-1 A2: writes go through the outbox decorators — LinkDeleted and
	// LinkCreated rows are recorded with each write and flushed to Redis at
	// the end of the run (and by the worker's relayer if this tool crashes).
	linkRepo := outbox.NewLinkRepository(postgres.NewLinkRepository(database), database)
	embeddingRepo := postgres.NewEmbeddingRepository(database, cfg.NLPModelName)
	gammaGen := recommendation.NewGammaLinkGenerator(embeddingRepo, linkRepo, 2, cfg.GammaLinkMinScore)

	// Candidates: notes that already have an embedding for the current model.
	noteIDs, err := embeddingRepo.FindNoteIDsWithModel(ctx)
	if err != nil {
		log.Fatalf("failed to find notes with embeddings: %v", err)
	}
	log.Printf("Found %d notes with embedding for model %s", len(noteIDs), cfg.NLPModelName)

	existing, err := linkRepo.FindBySourceType(ctx, gammaSourceType)
	if err != nil {
		log.Fatalf("failed to list existing gamma links: %v", err)
	}

	planned, suppressedCount, err := gammaGen.PlanForNotes(ctx, noteIDs)
	if err != nil {
		log.Fatalf("failed to plan regeneration: %v", err)
	}
	wouldCreate := 0
	for _, links := range planned {
		wouldCreate += len(links)
	}

	if *dryRun {
		log.Println("DRY RUN MODE - nothing will be written")
		log.Printf("Would delete %d gamma links (manual links untouched)", len(existing))
		log.Printf("Would create %d gamma links across %d notes", wouldCreate, len(noteIDs))
		log.Printf("%d candidates discarded by recorded rejections (link_suppressions)", suppressedCount)
		return
	}

	// Task queue so the graph-service refreshes note_links_closure and
	// recommendation caches pick up the new edges. Events themselves ride
	// the outbox; the publisher is needed only for the final relay flush.
	var eventPublisher *events.Publisher
	if cfg.EventChannel != "" {
		eventPublisher = events.NewPublisher(cacheClient, cfg.EventChannel)
	} else {
		log.Println("WARNING: EVENT_CHANNEL not set, outbox events will wait for the worker relay")
	}
	taskQueue, err := queue.NewAsynqClient(cfg.RedisURL, cfg.BackupEnabled, cfg.NLPPipelineEnabled, cfg.NLPQualityEnabled)
	if err != nil {
		log.Printf("WARNING: failed to create task queue client: %v", err)
		taskQueue = nil
	} else {
		defer func() {
			if err := taskQueue.Close(); err != nil {
				log.Printf("Error closing task queue client: %v", err)
			}
		}()
	}
	taskDelay := time.Duration(cfg.RecommendationTaskDelaySeconds) * time.Second

	// LinkDeleted events are recorded by the repository decorator for every
	// row removed below — no manual publishing here.
	deleted, err := linkRepo.DeleteBySourceType(ctx, gammaSourceType)
	if err != nil {
		log.Fatalf("failed to delete gamma links: %v", err)
	}
	log.Printf("Deleted %d gamma links (manual links preserved)", deleted)

	createdMap, err := gammaGen.GenerateForNotes(ctx, noteIDs)
	if err != nil {
		log.Fatalf("failed to regenerate gamma links: %v", err)
	}

	createdCount := 0
	refreshed := map[uuid.UUID]bool{}
	for sourceID, links := range createdMap {
		for _, l := range links {
			createdCount++
			if !refreshed[l.TargetNoteID()] {
				refreshed[l.TargetNoteID()] = true
			}
		}
		if len(links) > 0 {
			refreshed[sourceID] = true
		}
	}
	if taskQueue != nil {
		for noteID := range refreshed {
			if err := taskQueue.EnqueueRefreshRecommendations(ctx, noteID, taskDelay); err != nil {
				log.Printf("Failed to enqueue refresh for %s: %v", noteID, err)
			}
		}
	}

	// Deliver the events this run recorded — without the flush they would
	// wait in graph_outbox until the next worker relay tick.
	if eventPublisher != nil {
		relayer := outbox.NewRelayer(database, eventPublisher, 0, cfg.OutboxBatchSize)
		if n, err := relayer.Flush(ctx); err != nil {
			log.Printf("WARNING: outbox flush incomplete (%d sent): %v — the worker relay will deliver the rest", n, err)
		} else if n > 0 {
			log.Printf("Outbox flush delivered %d event(s)", n)
		}
	}

	log.Println("==========================")
	log.Printf("Completed: %d deleted, %d created across %d notes, %d refreshes enqueued",
		deleted, createdCount, len(noteIDs), len(refreshed))
}
