package foo

// UncoveredSave writes a note without publishing a graph event — the guard
// must flag exactly this call site.
func (h *Handler) UncoveredSave() {
	if err := h.repo.Save(ctx, n); err != nil {
		return
	}
}

// CoveredSave publishes after writing — the guard must stay green on it.
func (h *Handler) CoveredSave() {
	if err := h.repo.Save(ctx, n); err != nil {
		return
	}
	h.eventPublisher.PublishNoteCreated(ctx, n.ID().String(), userID)
}

// CoveredViaHelper delegates the publish to a same-file helper.
func (h *Handler) CoveredViaHelper() {
	if err := h.repo.Save(ctx, n); err != nil {
		return
	}
	h.postprocessCreatedNote(n)
}

func (h *Handler) postprocessCreatedNote(n *Note) {
	h.eventPublisher.PublishNoteCreated(ctx, n.ID().String(), userID)
}
