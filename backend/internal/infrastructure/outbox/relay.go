package outbox

import (
	"context"
	"log"
	"time"

	"knowledge-graph/internal/infrastructure/events"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Relayer publishes pending graph_outbox rows to the Redis event channel and
// marks them sent. Rows are locked FOR UPDATE SKIP LOCKED, so two relayers
// (or a flush racing the ticker) never double-publish the same row within a
// pass; a crash between commit and mark leaves the row unsent and the next
// pass re-delivers it — subscribers must tolerate duplicates.
type Relayer struct {
	db        *gorm.DB
	publisher *events.Publisher
	interval  time.Duration
	batchSize int
}

func NewRelayer(db *gorm.DB, publisher *events.Publisher, interval time.Duration, batchSize int) *Relayer {
	if interval <= 0 {
		interval = time.Second
	}
	if batchSize <= 0 {
		batchSize = 100
	}
	return &Relayer{db: db, publisher: publisher, interval: interval, batchSize: batchSize}
}

// Run polls the outbox until ctx is cancelled. Started from cmd/worker; if the
// process dies, rows keep waiting — the next process flushes them.
func (r *Relayer) Run(ctx context.Context) {
	if r.publisher == nil {
		log.Println("[Outbox] no event publisher configured — relay disabled")
		return
	}
	// Flush rows left unsent by a previous crash before the first tick.
	if n, err := r.Flush(ctx); err != nil {
		log.Printf("[Outbox] startup flush failed: %v", err)
	} else if n > 0 {
		log.Printf("[Outbox] startup flush delivered %d pending event(s)", n)
	}
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := r.Flush(ctx); err != nil {
				log.Printf("[Outbox] flush failed: %v", err)
			}
		}
	}
}

// Flush drains the outbox: repeated batches until a pass comes back short.
// CLI tools (gamma-links-regenerate) call it before exiting so their events
// do not wait for the worker's next run.
func (r *Relayer) Flush(ctx context.Context) (int, error) {
	total := 0
	for {
		n, err := r.FlushOnce(ctx)
		total += n
		if err != nil || n < r.batchSize {
			return total, err
		}
	}
}

// FlushOnce publishes up to batchSize unsent rows inside one transaction and
// returns how many it delivered. Exposed for CLI tools (gamma-links-regenerate)
// that must not exit with rows still pending.
func (r *Relayer) FlushOnce(ctx context.Context) (int, error) {
	if r.publisher == nil {
		return 0, nil
	}
	var rows []Model
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("sent_at IS NULL").
			Order("id").
			Limit(r.batchSize).
			Find(&rows).Error; err != nil {
			return err
		}
		now := time.Now()
		for i := range rows {
			if err := r.publisher.PublishEvent(ctx, rows[i].EventType, []byte(rows[i].Payload)); err != nil {
				return err
			}
			if err := tx.Model(&Model{}).Where("id = ?", rows[i].ID).Update("sent_at", now).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}

// PurgeSentBefore deletes sent rows older than cutoff — the retention window
// (default 30 days, owner decision 71) keeps the table bounded.
func PurgeSentBefore(db *gorm.DB, ctx context.Context, cutoff time.Time) (int64, error) {
	res := db.WithContext(ctx).Where("sent_at IS NOT NULL AND sent_at < ?", cutoff).Delete(&Model{})
	return res.RowsAffected, res.Error
}
