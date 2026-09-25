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

	noteRepo := postgres.NewNoteRepository(database, nil)
	linkRepo := postgres.NewLinkRepository(database)
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

	// Event publisher and task queue so the graph-service refreshes
	// note_links_closure and recommendation caches pick up the new edges.
	var eventPublisher *events.Publisher
	if cfg.EventChannel != "" {
		eventPublisher = events.NewPublisher(cacheClient, cfg.EventChannel)
	} else {
		log.Println("WARNING: EVENT_CHANNEL not set, LinkCreated/LinkDeleted events will not be published")
	}
	taskQueue, err := queue.NewAsynqClient(cfg.RedisURL, cfg.BackupEnabled, cfg.NLPPipelineEnabled)
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

	creators := map[uuid.UUID]string{}
	creatorOf := func(noteID uuid.UUID) string {
		if id, ok := creators[noteID]; ok {
			return id
		}
		n, err := noteRepo.FindByID(ctx, noteID)
		if err != nil || n == nil {
			creators[noteID] = ""
			return ""
		}
		id := ""
		if n.CreatorID() != nil {
			id = n.CreatorID().String()
		}
		creators[noteID] = id
		return id
	}

	// Publish LinkDeleted for every gamma link before removal so caches and
	// the closure view reflect the deletion.
	for _, l := range existing {
		if eventPublisher != nil {
			if err := eventPublisher.PublishLinkDeleted(ctx, l.SourceNoteID().String(), l.TargetNoteID().String(), creatorOf(l.SourceNoteID())); err != nil {
				log.Printf("Failed to publish LinkDeleted %s -> %s: %v", l.SourceNoteID(), l.TargetNoteID(), err)
			}
		}
	}

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
			if eventPublisher != nil {
				if err := eventPublisher.PublishLinkCreated(ctx, l.SourceNoteID().String(), l.TargetNoteID().String(), creatorOf(l.SourceNoteID())); err != nil {
					log.Printf("Failed to publish LinkCreated %s -> %s: %v", l.SourceNoteID(), l.TargetNoteID(), err)
				}
			}
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

	log.Println("==========================")
	log.Printf("Completed: %d deleted, %d created across %d notes, %d refreshes enqueued",
		deleted, createdCount, len(noteIDs), len(refreshed))
}
