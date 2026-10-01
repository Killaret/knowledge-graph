package notification

import (
	"context"

	"github.com/google/uuid"
)

// Repository handles persistence for notifications.
type Repository interface {
	// CreateIfAbsent inserts the notification; returns created=false when a
	// notification with the same non-empty dedupe key already exists.
	CreateIfAbsent(ctx context.Context, n *Notification) (created bool, err error)
	// ListByUser returns the user's notifications newest first.
	ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*Notification, error)
	// CountUnread returns the number of unread notifications of the user.
	CountUnread(ctx context.Context, userID uuid.UUID) (int64, error)
	// MarkRead marks a notification as read; only the owner may do so.
	MarkRead(ctx context.Context, id, userID uuid.UUID) error
}
