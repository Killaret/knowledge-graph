package main

import (
	"context"
	"log"
	"time"

	"knowledge-graph/internal/infrastructure/outbox"

	"gorm.io/gorm"
)

// trashPurger is the port the NOTE-DELETE-1 cleanup needs from the note
// repository — kept narrow on purpose.
type trashPurger interface {
	PurgeDeletedBefore(ctx context.Context, cutoff time.Time) (int64, error)
}

// softDeletedCleanup adapts the note repository's trash purge to
// tasks.CleanupServiceInterface: rows soft-deleted more than `days` ago are
// removed for good (links go with their notes through the FK cascade).
// The same daily pass purges sent graph_outbox rows past their retention
// window (SYNC-1 A2, default 30 days).
type softDeletedCleanup struct {
	purger         trashPurger
	db             *gorm.DB
	outboxKeepDays int
}

func (s softDeletedCleanup) CleanupSoftDeleted(ctx context.Context, days int, _ []string) error {
	cutoff := time.Now().Add(-24 * time.Hour * time.Duration(days))
	n, err := s.purger.PurgeDeletedBefore(ctx, cutoff)
	if err != nil {
		return err
	}
	log.Printf("[Worker] purged %d trashed notes older than %d days", n, days)
	if s.db != nil && s.outboxKeepDays > 0 {
		sentCutoff := time.Now().Add(-24 * time.Hour * time.Duration(s.outboxKeepDays))
		m, err := outbox.PurgeSentBefore(s.db, ctx, sentCutoff)
		if err != nil {
			return err
		}
		if m > 0 {
			log.Printf("[Worker] purged %d sent outbox events older than %d days", m, s.outboxKeepDays)
		}
	}
	return nil
}
