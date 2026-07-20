package recommendation

import (
	"context"
	"errors"
	"testing"

	"knowledge-graph/internal/domain/graph"
	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRefreshService_RefreshRecommendations(t *testing.T) {
	ctx := context.Background()
	noteID := uuid.MustParse("a0000000-0000-0000-0000-000000000001")
	targetID := uuid.MustParse("a0000000-0000-0000-0000-000000000002")

	makeNote := func() *note.Note {
		title, _ := note.NewTitle("Test Note")
		content, _ := note.NewContent("Content")
		meta, _ := note.NewMetadata(map[string]interface{}{})
		n := note.NewNoteWithCreator(title, content, "star", meta, uuid.New())
		return n
	}

	t.Run("successful refresh", func(t *testing.T) {
		noteRepo := new(mockNoteRepository)
		recRepo := new(mockRecommendationRepository)
		traversalSvc := new(MockTraversalService)

		svc := NewRefreshService(noteRepo, recRepo, traversalSvc, 10)

		n := makeNote()
		noteRepo.On("FindByID", ctx, noteID).Return(n, nil).Once()

		suggestions := []graph.SuggestionResult{{NodeID: targetID, Score: 0.9}}
		traversalSvc.On("GetSuggestions", ctx, noteID, 10).Return(suggestions, nil).Once()

		recRepo.On("ReplaceRecommendations", ctx, noteID, map[uuid.UUID]float64{targetID: 0.9}).Return(nil).Once()

		err := svc.RefreshRecommendations(ctx, noteID)

		assert.NoError(t, err)
		noteRepo.AssertExpectations(t)
		recRepo.AssertExpectations(t)
		traversalSvc.AssertExpectations(t)
	})

	t.Run("note not found", func(t *testing.T) {
		noteRepo := new(mockNoteRepository)
		recRepo := new(mockRecommendationRepository)
		traversalSvc := new(MockTraversalService)

		svc := NewRefreshService(noteRepo, recRepo, traversalSvc, 10)

		noteRepo.On("FindByID", ctx, noteID).Return(nil, nil).Once()

		err := svc.RefreshRecommendations(ctx, noteID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "note not found")
		noteRepo.AssertExpectations(t)
	})

	t.Run("note repository error", func(t *testing.T) {
		noteRepo := new(mockNoteRepository)
		recRepo := new(mockRecommendationRepository)
		traversalSvc := new(MockTraversalService)

		svc := NewRefreshService(noteRepo, recRepo, traversalSvc, 10)

		noteRepo.On("FindByID", ctx, noteID).Return(nil, errors.New("db error")).Once()

		err := svc.RefreshRecommendations(ctx, noteID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "note not found")
		noteRepo.AssertExpectations(t)
	})

	t.Run("traversal service error", func(t *testing.T) {
		noteRepo := new(mockNoteRepository)
		recRepo := new(mockRecommendationRepository)
		traversalSvc := new(MockTraversalService)

		svc := NewRefreshService(noteRepo, recRepo, traversalSvc, 10)

		n := makeNote()
		noteRepo.On("FindByID", ctx, noteID).Return(n, nil).Once()
		traversalSvc.On("GetSuggestions", ctx, noteID, 10).Return(nil, errors.New("traversal failed")).Once()

		err := svc.RefreshRecommendations(ctx, noteID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get suggestions")
		noteRepo.AssertExpectations(t)
		traversalSvc.AssertExpectations(t)
	})

	t.Run("replace recommendations error", func(t *testing.T) {
		noteRepo := new(mockNoteRepository)
		recRepo := new(mockRecommendationRepository)
		traversalSvc := new(MockTraversalService)

		svc := NewRefreshService(noteRepo, recRepo, traversalSvc, 10)

		n := makeNote()
		noteRepo.On("FindByID", ctx, noteID).Return(n, nil).Once()

		suggestions := []graph.SuggestionResult{{NodeID: targetID, Score: 0.9}}
		traversalSvc.On("GetSuggestions", ctx, noteID, 10).Return(suggestions, nil).Once()

		recRepo.On("ReplaceRecommendations", ctx, noteID, map[uuid.UUID]float64{targetID: 0.9}).Return(errors.New("db error")).Once()

		err := svc.RefreshRecommendations(ctx, noteID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to replace recommendations")
		noteRepo.AssertExpectations(t)
		recRepo.AssertExpectations(t)
		traversalSvc.AssertExpectations(t)
	})
}

func TestRefreshService_RefreshRecommendationsBatch(t *testing.T) {
	ctx := context.Background()
	noteID := uuid.MustParse("a0000000-0000-0000-0000-000000000001")
	targetID := uuid.MustParse("a0000000-0000-0000-0000-000000000002")

	title, _ := note.NewTitle("Test Note")
	content, _ := note.NewContent("Content")
	meta, _ := note.NewMetadata(map[string]interface{}{})
	n := note.NewNoteWithCreator(title, content, "star", meta, uuid.New())

	noteRepo := new(mockNoteRepository)
	recRepo := new(mockRecommendationRepository)
	traversalSvc := new(MockTraversalService)

	svc := NewRefreshService(noteRepo, recRepo, traversalSvc, 10)

	noteRepo.On("FindByID", ctx, noteID).Return(n, nil).Once()
	traversalSvc.On("GetSuggestions", ctx, noteID, 10).Return([]graph.SuggestionResult{{NodeID: targetID, Score: 0.9}}, nil).Once()
	recRepo.On("ReplaceRecommendations", ctx, noteID, map[uuid.UUID]float64{targetID: 0.9}).Return(nil).Once()

	err := svc.RefreshRecommendationsBatch(ctx, []uuid.UUID{noteID}, 1)

	assert.NoError(t, err)
	noteRepo.AssertExpectations(t)
	recRepo.AssertExpectations(t)
	traversalSvc.AssertExpectations(t)
}
