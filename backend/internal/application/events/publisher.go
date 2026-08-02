package events

import "context"

// Publisher publishes graph-relevant events to the graph-service cache invalidation channel.
type Publisher interface {
	PublishNoteCreated(ctx context.Context, noteID, userID string) error
	PublishNoteUpdated(ctx context.Context, noteID, userID string) error
	PublishNoteDeleted(ctx context.Context, noteID, userID string) error
	PublishLinkCreated(ctx context.Context, sourceNoteID, targetNoteID, userID string) error
	PublishLinkUpdated(ctx context.Context, sourceNoteID, targetNoteID, userID string) error
	PublishLinkDeleted(ctx context.Context, sourceNoteID, targetNoteID, userID string) error
}
