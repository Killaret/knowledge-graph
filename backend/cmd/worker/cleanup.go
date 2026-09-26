package main

import (
	"context"
	"log"
	"time"
)

// trashPurger is the port the NOTE-DELETE-1 cleanup needs from the note
// repository — kept narrow on purpose.
type trashPurger interface {
	PurgeDeletedBefore(ctx context.Context, cutoff time.Time) (int64, error)
}

// softDeletedCleanup adapts the note repository's trash purge to
// tasks.CleanupServiceInterface: rows soft-deleted more than `days` ago are
// removed for good (links go with their notes through the FK cascade).
type softDeletedCleanup struct {
	purger trashPurger
}

func (s softDeletedCleanup) CleanupSoftDeleted(ctx context.Context, days int, _ []string) error {
	cutoff := time.Now().Add(-24 * time.Hour * time.Duration(days))
	n, err := s.purger.PurgeDeletedBefore(ctx, cutoff)
	if err != nil {
		return err
	}
	log.Printf("[Worker] purged %d trashed notes older than %d days", n, days)
	return nil
}
