package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"

	appquality "knowledge-graph/internal/application/quality"
	"knowledge-graph/internal/config"
	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/infrastructure/db"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/infrastructure/mongo"
	"knowledge-graph/internal/infrastructure/queue"
)

// latestReader is the slice of quality_log the planner needs — an interface
// so the plan is unit-testable without Mongo.
type latestReader interface {
	LatestQuality(ctx context.Context, noteID uuid.UUID) (*appquality.LogEntry, error)
}

// exportRow is what --export writes. Deliberately text-free: no title, no
// content, no chunk text — ids, hashes and signals only.
type exportRow struct {
	NoteID            uuid.UUID          `json:"note_id"`
	SourceHash        string             `json:"source_hash"`
	Verdict           string             `json:"verdict"`
	Gates             []string           `json:"gates"`
	Attempt           int                `json:"attempt"`
	NeedsManualReview bool               `json:"needs_manual_review"`
	ComputedAt        string             `json:"computed_at"`
	Signals           appquality.Signals `json:"signals"`
}

// planRecompute lists notes whose current text version has no assessment —
// a note is due when the latest log entry covers a different source_hash.
func planRecompute(ctx context.Context, notes []*note.Note, logStore latestReader) (due []uuid.UUID, assessed int, err error) {
	for _, n := range notes {
		hash := appquality.SourceHash(n.Title().String(), n.Content().String())
		entry, err := logStore.LatestQuality(ctx, n.ID())
		if err != nil {
			return nil, 0, fmt.Errorf("failed to read quality log for note %s: %w", n.ID(), err)
		}
		if entry != nil && entry.SourceHash == hash {
			assessed++
			continue
		}
		due = append(due, n.ID())
	}
	return due, assessed, nil
}

// exportSnapshot writes one JSONL row per note from the latest log entry.
func exportSnapshot(ctx context.Context, notes []*note.Note, logStore latestReader, path string) (int, error) {
	f, err := os.Create(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	written := 0
	enc := json.NewEncoder(f)
	for _, n := range notes {
		entry, err := logStore.LatestQuality(ctx, n.ID())
		if err != nil {
			return written, fmt.Errorf("failed to read quality log for note %s: %w", n.ID(), err)
		}
		if entry == nil {
			continue
		}
		row := exportRow{
			NoteID:            n.ID(),
			SourceHash:        entry.SourceHash,
			Verdict:           entry.Record.Verdict,
			Gates:             entry.Record.Gates,
			Attempt:           entry.Record.Attempt,
			NeedsManualReview: entry.Record.NeedsManualReview,
			ComputedAt:        entry.Record.ComputedAt.Format("2006-01-02T15:04:05Z07:00"),
			Signals:           entry.Record.Signals,
		}
		if err := enc.Encode(row); err != nil {
			return written, err
		}
		written++
	}
	return written, nil
}

// quality-recompute backfills NOTE-QUALITY-1 assessments for existing notes.
// --dry-run prints due/assessed counts without touching the queue.
// --export <file.jsonl> dumps the latest assessment per note (no text).
// A real run enqueues quality:assess manual tasks; the worker computes the
// signals — the command itself never fetches or calls the NLP service.
func main() {
	dryRun := flag.Bool("dry-run", false, "Print how many notes would be assessed")
	exportPath := flag.String("export", "", "Write the latest assessment per note as JSONL (no titles/content)")
	flag.Parse()

	log.Println("Quality recompute (NOTE-QUALITY-1 backfill)")
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
	logStore := mongo.NewQualityLogRepository(mongoClient)
	if err := logStore.EnsureIndexes(context.Background()); err != nil {
		log.Printf("WARNING: failed to ensure quality_log indexes: %v", err)
	}

	ctx := context.Background()
	notes, err := noteRepo.FindAll(ctx)
	if err != nil {
		log.Fatalf("failed to list notes: %v", err)
	}
	log.Printf("Found %d notes", len(notes))

	if *exportPath != "" {
		written, err := exportSnapshot(ctx, notes, logStore, *exportPath)
		if err != nil {
			log.Fatalf("export failed: %v", err)
		}
		log.Printf("Exported %d assessments to %s", written, *exportPath)
	}

	due, assessed, err := planRecompute(ctx, notes, logStore)
	if err != nil {
		log.Fatalf("failed to plan recompute: %v", err)
	}

	if *dryRun {
		log.Println("DRY RUN — nothing enqueued")
		log.Printf("Would assess: %d, already assessed: %d", len(due), assessed)
		return
	}

	if !cfg.NLPQualityEnabled {
		log.Println("NOTE: nlp.quality.enabled is off — nothing will be enqueued. Set NLP_QUALITY_ENABLED=1 first.")
	}

	// The client gate stays authoritative: disabled quality means no
	// quality:assess task ever enters Redis, even from this command.
	taskQueue, err := queue.NewAsynqClient(cfg.RedisURL, cfg.BackupEnabled, cfg.NLPPipelineEnabled, cfg.NLPQualityEnabled)
	if err != nil {
		log.Fatalf("Failed to create task queue client: %v", err)
	}
	defer func() {
		if err := taskQueue.Close(); err != nil {
			log.Printf("Error closing task queue client: %v", err)
		}
	}()

	enqueued, failed := 0, 0
	for _, noteID := range due {
		if err := taskQueue.EnqueueAssessQuality(ctx, noteID.String(), appquality.TriggerManual); err != nil {
			log.Printf("Failed to enqueue %s: %v", noteID, err)
			failed++
			continue
		}
		enqueued++
	}

	log.Println("================================================")
	log.Printf("Done: %d enqueued, %d already assessed, %d failed", enqueued, assessed, failed)
	if enqueued > 0 {
		log.Println("Assessments will be written to MongoDB quality_log by workers")
	}
}
