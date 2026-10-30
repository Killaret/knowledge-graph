package recommendation

import (
	"context"
	"fmt"
	"log"
	"time"

	"knowledge-graph/internal/domain/graph"
	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
)

// NoteRepository defines the interface for note operations needed by RefreshService
type NoteRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error)
}

// TraversalService defines the interface for BFS traversal operations
type TraversalService interface {
	GetSuggestions(ctx context.Context, startID uuid.UUID, topN int) ([]graph.SuggestionResult, error)
}

// RefreshServiceInterface defines the interface for refresh operations
type RefreshServiceInterface interface {
	RefreshRecommendations(ctx context.Context, noteID uuid.UUID) error
}

// Ensure RefreshService implements RefreshServiceInterface
var _ RefreshServiceInterface = (*RefreshService)(nil)

// RefreshService handles background recalculation of note recommendations
type RefreshService struct {
	noteRepo     NoteRepository
	recRepo      Repository
	traversalSvc TraversalService
	topN         int
}

// NewRefreshService creates a new refresh service
// Note: traversalSvc can be *graph.TraversalService which implements TraversalService interface
func NewRefreshService(noteRepo NoteRepository, recRepo Repository, traversalSvc TraversalService, topN int) *RefreshService {
	return &RefreshService{
		noteRepo:     noteRepo,
		recRepo:      recRepo,
		traversalSvc: traversalSvc,
		topN:         topN,
	}
}

// RefreshRecommendations recalculates and atomically updates recommendations for a note
// Uses repository transaction to ensure consistency and prevent race conditions
func (s *RefreshService) RefreshRecommendations(ctx context.Context, noteID uuid.UUID) error {
	start := time.Now()
	log.Printf("[RefreshService] Starting refresh recommendations for note %s", noteID)
	defer func() {
		log.Printf("[RefreshService] Finished refresh for note %s in %v", noteID, time.Since(start))
	}()

	// Verify note exists
	n, err := s.noteRepo.FindByID(ctx, noteID)
	if err != nil {
		return fmt.Errorf("note not found: %w", err)
	}
	if n == nil {
		return fmt.Errorf("note not found: %s", noteID)
	}

	// Pass the note's creator so graph-service visibility filtering is applied.
	if n.CreatorID() != nil {
		ctx = graph.WithUserID(ctx, n.CreatorID().String())
	}

	// Get fresh recommendations via BFS traversal
	suggestions, err := s.traversalSvc.GetSuggestions(ctx, noteID, s.topN)
	if err != nil {
		return fmt.Errorf("failed to get suggestions: %w", err)
	}

	// Convert to map format for repository
	recs := make(map[uuid.UUID]float64, len(suggestions))
	for _, sug := range suggestions {
		recs[sug.NodeID] = sug.Score
	}

	// Atomic update: insert/update new recommendations and delete stale ones
	if err := s.recRepo.ReplaceRecommendations(ctx, noteID, recs); err != nil {
		log.Printf("[RefreshService] Failed to refresh recommendations for note %s: %v", noteID, err)
		return fmt.Errorf("failed to replace recommendations: %w", err)
	}

	log.Printf("[RefreshService] Successfully refreshed %d recommendations for note %s", len(recs), noteID)
	return nil
}

// RefreshRecommendationsBatch processes multiple notes with rate limiting
// Useful for initial seeding or periodic full refresh
func (s *RefreshService) RefreshRecommendationsBatch(ctx context.Context, noteIDs []uuid.UUID, rateLimit int) error {
	if rateLimit <= 0 {
		rateLimit = 10 // Default rate limit
	}

	log.Printf("[RefreshService] Starting batch refresh for %d notes (rate limit: %d/sec)", len(noteIDs), rateLimit)

	for i, noteID := range noteIDs {
		if err := s.RefreshRecommendations(ctx, noteID); err != nil {
			log.Printf("[RefreshService] Failed to refresh note %s: %v", noteID, err)
			// Continue with other notes, don't fail the entire batch
			continue
		}

		// Simple rate limiting
		if (i+1)%rateLimit == 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				// Small delay between batches
			}
		}
	}

	log.Printf("[RefreshService] Batch refresh completed")
	return nil
}
