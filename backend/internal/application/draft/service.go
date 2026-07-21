// Package draft provides application services for note drafts
package draft

import (
	"context"
	"fmt"
	"time"

	noteDomain "knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
)

// Service provides draft management operations
type Service struct {
	repo         noteDomain.DraftRepository
	noteRepo     noteDomain.Repository
	syncEndpoint string
	maxRetries   int
}

// NewService creates a new draft service
func NewService(repo noteDomain.DraftRepository, noteRepo noteDomain.Repository, syncEndpoint string) *Service {
	return &Service{
		repo:         repo,
		noteRepo:     noteRepo,
		syncEndpoint: syncEndpoint,
		maxRetries:   3,
	}
}

// SetMaxRetries is used by tests to avoid long retry sleeps.
func (s *Service) SetMaxRetries(n int) {
	s.maxRetries = n
}

// SaveDraft saves a draft for a note
func (s *Service) SaveDraft(ctx context.Context, noteID, userID uuid.UUID, content, title string) (*noteDomain.Draft, error) {
	// Check if draft already exists
	existing, err := s.repo.FindByNoteAndUser(ctx, noteID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing draft: %w", err)
	}

	if existing != nil {
		// Update existing draft
		if err := existing.UpdateContent(content); err != nil {
			return nil, fmt.Errorf("failed to update draft content: %w", err)
		}
		if err := existing.UpdateTitle(title); err != nil {
			return nil, fmt.Errorf("failed to update draft title: %w", err)
		}
		if err := s.repo.Update(ctx, existing); err != nil {
			return nil, fmt.Errorf("failed to update draft: %w", err)
		}
		return existing, nil
	}

	// Create new draft
	draft := noteDomain.NewDraft(noteID, userID, content, title)
	if err := s.repo.Save(ctx, draft); err != nil {
		return nil, fmt.Errorf("failed to save draft: %w", err)
	}

	return draft, nil
}

// SyncDraft synchronizes a draft with the server
func (s *Service) SyncDraft(ctx context.Context, draftID uuid.UUID) error {
	draft, err := s.repo.FindByID(ctx, draftID)
	if err != nil {
		return fmt.Errorf("failed to find draft: %w", err)
	}
	if draft == nil {
		return fmt.Errorf("draft not found")
	}

	// Transition to publishing state
	if err := draft.StartPublishing(); err != nil {
		return fmt.Errorf("failed to start publishing: %w", err)
	}

	if err := s.repo.Update(ctx, draft); err != nil {
		return fmt.Errorf("failed to update draft state: %w", err)
	}

	// Attempt to sync with retry logic
	var lastErr error
	for i := 0; i < s.maxRetries; i++ {
		err := s.syncWithServer(ctx, draft)
		if err == nil {
			// Success - mark as published
			if err := draft.MarkAsPublished(); err != nil {
				return fmt.Errorf("failed to mark as published: %w", err)
			}
			if err := s.repo.Update(ctx, draft); err != nil {
				return fmt.Errorf("failed to update draft after publish: %w", err)
			}

			// Delete draft after successful publish
			if err := s.repo.DeleteByID(ctx, draft.ID()); err != nil {
				return fmt.Errorf("failed to delete published draft: %w", err)
			}

			return nil
		}
		lastErr = err
		time.Sleep(time.Second * time.Duration(i+1))
	}

	// All retries failed - mark as conflict
	if err := draft.MarkAsConflict(); err != nil {
		return fmt.Errorf("failed to mark as conflict: %w", err)
	}
	if err := s.repo.Update(ctx, draft); err != nil {
		return fmt.Errorf("failed to update draft after conflict: %w", err)
	}

	return fmt.Errorf("sync failed after %d retries: %w", s.maxRetries, lastErr)
}

// syncWithServer performs the actual sync of a draft to the note repository.
// When a remote syncEndpoint is configured it will be used; otherwise the sync
// is performed locally through the note repository.
func (s *Service) syncWithServer(ctx context.Context, draft *noteDomain.Draft) error {
	if s.syncEndpoint != "" {
		// Remote sync over HTTP is not implemented yet.
		return fmt.Errorf("remote sync endpoint not implemented: %s", s.syncEndpoint)
	}

	// Local sync: update the target note from the draft contents.
	note, err := s.noteRepo.FindByID(ctx, draft.NoteID())
	if err != nil {
		return fmt.Errorf("failed to find note: %w", err)
	}
	if note == nil {
		return fmt.Errorf("note not found")
	}

	if note.CreatorID() == nil || *note.CreatorID() != draft.UserID() {
		return fmt.Errorf("draft does not belong to the note owner")
	}

	title, err := noteDomain.NewTitle(draft.Title())
	if err != nil {
		return fmt.Errorf("invalid draft title: %w", err)
	}
	content, err := noteDomain.NewContent(draft.Content())
	if err != nil {
		return fmt.Errorf("invalid draft content: %w", err)
	}

	if err := note.UpdateTitle(title); err != nil {
		return fmt.Errorf("failed to update note title: %w", err)
	}
	if err := note.UpdateContent(content); err != nil {
		return fmt.Errorf("failed to update note content: %w", err)
	}

	if err := s.noteRepo.Save(ctx, note); err != nil {
		return fmt.Errorf("failed to save note: %w", err)
	}

	return nil
}

// ResolveConflict resolves a conflict by transitioning back to active state
func (s *Service) ResolveConflict(ctx context.Context, draftID uuid.UUID) error {
	draft, err := s.repo.FindByID(ctx, draftID)
	if err != nil {
		return fmt.Errorf("failed to find draft: %w", err)
	}
	if draft == nil {
		return fmt.Errorf("draft not found")
	}

	if err := draft.ResolveConflict(); err != nil {
		return fmt.Errorf("failed to resolve conflict: %w", err)
	}

	if err := s.repo.Update(ctx, draft); err != nil {
		return fmt.Errorf("failed to update draft: %w", err)
	}

	return nil
}

// GetLatestDraft gets the latest draft for a note
func (s *Service) GetLatestDraft(ctx context.Context, noteID, userID uuid.UUID) (*noteDomain.Draft, error) {
	return s.repo.FindByNoteAndUser(ctx, noteID, userID)
}

// GetActiveDrafts gets all active drafts for a user
func (s *Service) GetActiveDrafts(ctx context.Context, userID uuid.UUID) ([]*noteDomain.Draft, error) {
	return s.repo.FindActiveByUser(ctx, userID)
}

// DeleteDraft deletes a draft
func (s *Service) DeleteDraft(ctx context.Context, draftID uuid.UUID) error {
	return s.repo.DeleteByID(ctx, draftID)
}
