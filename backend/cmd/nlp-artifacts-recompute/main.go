package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log"

	"knowledge-graph/internal/config"
	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/infrastructure/db"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/infrastructure/mongo"
	"knowledge-graph/internal/infrastructure/queue"

	"github.com/google/uuid"
)

// artifactHashReader is the slice of the artifacts repository the scan needs;
// an interface so the plan can be unit-tested without Mongo.
type artifactHashReader interface {
	FindCurrentSourceHash(ctx context.Context, noteID uuid.UUID, pipelineVersion string) (string, bool, error)
}

// sourceHash mirrors queue.nlpArtifactsSourceHash: the same bytes the worker
// hashes before calling /normalize, so a re-run skips what the worker wrote.
func sourceHash(title, content string) string {
	sum := sha256.Sum256([]byte(title + "\x00" + content))
	return hex.EncodeToString(sum[:])
}

// planRecompute lists note IDs whose current artifact is missing or stale.
func planRecompute(ctx context.Context, notes []*note.Note, store artifactHashReader) (toCreate []string, skipped int, err error) {
	for _, n := range notes {
		hash := sourceHash(n.Title().String(), n.Content().String())
		existing, found, err := store.FindCurrentSourceHash(ctx, n.ID(), queue.NlpPipelineVersion)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to read artifact for note %s: %w", n.ID(), err)
		}
		if found && existing == hash {
			skipped++
			continue
		}
		toCreate = append(toCreate, n.ID().String())
	}
	return toCreate, skipped, nil
}

// nlp-artifacts-recompute fills Mongo nlp_artifacts for existing notes.
// Skips notes whose current artifact already matches the source hash —
// a second run creates nothing. --dry-run prints create/skip counts.
//
// The pipeline flag gates only *automatic* enqueue on note save; this
// command is an explicit operator action, so the asynq client is built
// with the flag forced on.
func main() {
	dryRun := flag.Bool("dry-run", false, "Print how many artifacts would be created and how many skipped")
	flag.Parse()

	log.Println("NLP artifacts recompute (nlp:normalize backfill)")
	log.Println("================================================")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("FATAL: Failed to load configuration: %v", err)
	}

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	noteRepo := postgres.NewNoteRepository(database, nil)

	mongoClient, err := mongo.NewClient(context.Background(), cfg.MongoDBURL, cfg.MongoDBDatabase)
	if err != nil {
		log.Fatalf("mongodb connection failed: %v", err)
	}
	defer func() {
		if err := mongoClient.Close(context.Background()); err != nil {
			log.Printf("[MongoDB] Error closing client: %v", err)
		}
	}()
	artifactsRepo := mongo.NewNlpArtifactsRepository(mongoClient)
	if err := artifactsRepo.EnsureIndexes(context.Background()); err != nil {
		log.Fatalf("failed to ensure nlp_artifacts indexes: %v", err)
	}

	ctx := context.Background()
	notes, err := noteRepo.FindAll(ctx)
	if err != nil {
		log.Fatalf("failed to list notes: %v", err)
	}
	log.Printf("Found %d notes", len(notes))

	toCreate, skipped, err := planRecompute(ctx, notes, artifactsRepo)
	if err != nil {
		log.Fatalf("failed to read artifacts: %v", err)
	}

	if *dryRun {
		log.Println("DRY RUN — nothing enqueued")
		log.Printf("Would create/refresh: %d, skipped (unchanged): %d", len(toCreate), skipped)
		return
	}

	// Manual recompute must enqueue regardless of nlp.pipeline.enabled —
	// that flag governs automatic enqueue on note save, not this command.
	taskQueue, err := queue.NewAsynqClient(cfg.RedisURL, cfg.BackupEnabled, true, cfg.NLPQualityEnabled)
	if err != nil {
		log.Fatalf("Failed to create task queue client: %v", err)
	}
	defer func() {
		if err := taskQueue.Close(); err != nil {
			log.Printf("Error closing task queue client: %v", err)
		}
	}()

	enqueued, failed := 0, 0
	for _, noteID := range toCreate {
		if err := taskQueue.EnqueueNormalizeNote(ctx, noteID); err != nil {
			log.Printf("Failed to enqueue %s: %v", noteID, err)
			failed++
			continue
		}
		enqueued++
	}

	log.Println("================================================")
	log.Printf("Done: %d enqueued, %d skipped (unchanged), %d failed", enqueued, skipped, failed)
	log.Println("Artifacts will be written to MongoDB nlp_artifacts by workers")
}
