package postgres

import (
	"context"

	"gorm.io/gorm"
)

// txContextKey carries an open transaction through ctx so a decorator can run
// a repository write and a companion write (graph_outbox row, SYNC-1 stage A2)
// in a single transaction without changing the domain repository interfaces.
type txContextKey struct{}

// ContextWithTx returns ctx bound to tx. Repository write methods resolve the
// bound transaction via dbFromContext; reads never consult it.
func ContextWithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

// dbFromContext returns the transaction bound to ctx, or fallback when none is
// bound. A bound tx already carries a session, so callers must not wrap it in
// WithContext again — tx.WithContext(ctx) is a no-op-safe call anyway.
func dbFromContext(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txContextKey{}).(*gorm.DB); ok && tx != nil {
		return tx
	}
	return fallback.WithContext(ctx)
}
