package recommendation

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAffectedNotesService_GetAffectedNotes(t *testing.T) {
	ctx := context.Background()
	targetNoteID := uuid.MustParse("a0000000-0000-0000-0000-000000000001")

	t.Run("returns target note and reverse dependencies", func(t *testing.T) {
		mockRepo := new(mockRecommendationRepository)
		svc := NewAffectedNotesService(mockRepo)

		id2 := uuid.MustParse("a0000000-0000-0000-0000-000000000002")
		id3 := uuid.MustParse("a0000000-0000-0000-0000-000000000003")
		mockRepo.On("GetNotesThatRecommend", ctx, targetNoteID).Return([]uuid.UUID{id2, id3}, nil).Once()

		result, err := svc.GetAffectedNotes(ctx, targetNoteID)

		assert.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Contains(t, result, targetNoteID)
		assert.Contains(t, result, id2)
		assert.Contains(t, result, id3)
		mockRepo.AssertExpectations(t)
	})

	t.Run("returns only target note when no reverse dependencies", func(t *testing.T) {
		mockRepo := new(mockRecommendationRepository)
		svc := NewAffectedNotesService(mockRepo)

		mockRepo.On("GetNotesThatRecommend", ctx, targetNoteID).Return([]uuid.UUID{}, nil).Once()

		result, err := svc.GetAffectedNotes(ctx, targetNoteID)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, targetNoteID, result[0])
		mockRepo.AssertExpectations(t)
	})

	t.Run("deduplicates duplicate IDs", func(t *testing.T) {
		mockRepo := new(mockRecommendationRepository)
		svc := NewAffectedNotesService(mockRepo)

		duplicateID := uuid.MustParse("a0000000-0000-0000-0000-000000000002")
		mockRepo.On("GetNotesThatRecommend", ctx, targetNoteID).Return([]uuid.UUID{duplicateID, duplicateID}, nil).Once()

		result, err := svc.GetAffectedNotes(ctx, targetNoteID)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("handles repository error", func(t *testing.T) {
		mockRepo := new(mockRecommendationRepository)
		svc := NewAffectedNotesService(mockRepo)

		mockRepo.On("GetNotesThatRecommend", ctx, targetNoteID).Return(nil, errors.New("db error")).Once()

		result, err := svc.GetAffectedNotes(ctx, targetNoteID)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestDeduplicate(t *testing.T) {
	t.Run("removes duplicates", func(t *testing.T) {
		id1 := uuid.MustParse("a0000000-0000-0000-0000-000000000001")
		id2 := uuid.MustParse("a0000000-0000-0000-0000-000000000002")
		id3 := uuid.MustParse("a0000000-0000-0000-0000-000000000003")

		input := []uuid.UUID{id1, id2, id1, id3, id2, id1}
		result := deduplicate(input)

		assert.Len(t, result, 3)
		assert.Contains(t, result, id1)
		assert.Contains(t, result, id2)
		assert.Contains(t, result, id3)
	})

	t.Run("handles empty slice", func(t *testing.T) {
		result := deduplicate([]uuid.UUID{})
		assert.Empty(t, result)
	})

	t.Run("handles single element", func(t *testing.T) {
		id := uuid.MustParse("a0000000-0000-0000-0000-000000000001")
		result := deduplicate([]uuid.UUID{id})
		assert.Len(t, result, 1)
		assert.Equal(t, id, result[0])
	})

	t.Run("handles all duplicates", func(t *testing.T) {
		id := uuid.MustParse("a0000000-0000-0000-0000-000000000001")
		input := []uuid.UUID{id, id, id, id}
		result := deduplicate(input)
		assert.Len(t, result, 1)
	})
}

func TestReverseCascadeDepth(t *testing.T) {
	// Verify that reverseCascadeDepth is set to 1 to prevent queue explosion
	assert.Equal(t, 1, reverseCascadeDepth, "reverseCascadeDepth should be 1 to prevent queue explosion")
}
