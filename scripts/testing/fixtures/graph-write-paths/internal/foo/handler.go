package foo

import "gorm.io/gorm"

// ManualPublish bypasses the outbox — the guard must flag this call.
func (h *Handler) ManualPublish() {
	h.publisher.PublishNoteCreated(ctx, noteID, userID)
}

// UnwrappedRepo constructs a raw repository — writes through it would emit
// no graph event, so the guard must flag the construction site.
func UnwrappedRepo(db *gorm.DB) {
	linkRepo := postgres.NewLinkRepository(db)
	_ = linkRepo
}
