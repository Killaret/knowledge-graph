package graph

import "context"

type contextKey int

const graphUserIDKey contextKey = iota

// WithUserID returns a context that carries the user ID for downstream graph
// service calls.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, graphUserIDKey, userID)
}

// UserIDFromContext extracts the user ID previously stored by WithUserID.
func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(graphUserIDKey).(string)
	return v, ok
}
