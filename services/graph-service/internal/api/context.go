package api

import "context"

// contextKey is a private type for context keys to avoid collisions.
type contextKey int

const (
	contextUserIDKey contextKey = iota
	contextIsPublicKey
	contextIsSkipAuthKey
)

// withUserID returns a new context with the authenticated user ID.
func withUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, contextUserIDKey, userID)
}

// userIDFromContext returns the user ID stored in the context, if any.
func userIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(contextUserIDKey).(string)
	return v, ok
}

// withPublic returns a new context marked as a public/anonymous request.
func withPublic(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextIsPublicKey, true)
}

// isPublicRequest returns true if the request was explicitly marked as public.
func isPublicRequest(ctx context.Context) bool {
	v, ok := ctx.Value(contextIsPublicKey).(bool)
	return ok && v
}

// withSkipAuth returns a new context marked as a SKIP_AUTH request.
func withSkipAuth(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextIsSkipAuthKey, true)
}

// isSkipAuthRequest returns true if the request came from a SKIP_AUTH environment.
func isSkipAuthRequest(ctx context.Context) bool {
	v, ok := ctx.Value(contextIsSkipAuthKey).(bool)
	return ok && v
}
