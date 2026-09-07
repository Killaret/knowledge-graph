package graph

import (
	"context"
	apprec "knowledge-graph/internal/application/recommendation"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockEmbeddingRepository для тестов
type MockEmbeddingRepository struct {
	findSimilarFunc      func(context.Context, uuid.UUID, int) ([]apprec.SimilarNote, error)
	findSimilarBatchFunc func(context.Context, []uuid.UUID, int) (map[uuid.UUID][]apprec.SimilarNote, error)
}

func (m *MockEmbeddingRepository) FindSimilarNotes(ctx context.Context, noteID uuid.UUID, limit int) ([]apprec.SimilarNote, error) {
	if m.findSimilarFunc != nil {
		return m.findSimilarFunc(ctx, noteID, limit)
	}
	return nil, nil
}

func (m *MockEmbeddingRepository) FindSimilarNotesBatch(ctx context.Context, noteIDs []uuid.UUID, limit int) (map[uuid.UUID][]apprec.SimilarNote, error) {
	if m.findSimilarBatchFunc != nil {
		return m.findSimilarBatchFunc(ctx, noteIDs, limit)
	}
	return nil, nil
}

func TestNewEmbeddingNeighborLoader(t *testing.T) {
	mockRepo := &MockEmbeddingRepository{}
	limit := 30

	loader := NewEmbeddingNeighborLoader(mockRepo, limit)

	require.NotNil(t, loader)
}

func TestEmbeddingNeighborLoader_GetNeighbors_Success(t *testing.T) {
	nodeID := uuid.New()

	mockRepo := &MockEmbeddingRepository{
		findSimilarFunc: func(ctx context.Context, id uuid.UUID, limit int) ([]apprec.SimilarNote, error) {
			assert.Equal(t, nodeID, id)
			assert.Equal(t, 30, limit)
			return []apprec.SimilarNote{
				{NoteID: uuid.New(), Score: 0.9},
				{NoteID: uuid.New(), Score: 0.8},
				{NoteID: uuid.New(), Score: 0.7},
			}, nil
		},
	}

	loader := NewEmbeddingNeighborLoader(mockRepo, 30)

	edges, err := loader.GetNeighbors(context.Background(), nodeID)

	require.NoError(t, err)
	require.Len(t, edges, 3)

	assert.Equal(t, nodeID, edges[0].From)
	assert.Equal(t, 0.9, edges[0].Weight)
	assert.Equal(t, 0.8, edges[1].Weight)
	assert.Equal(t, 0.7, edges[2].Weight)
}

func TestEmbeddingNeighborLoader_GetNeighbors_EmptyResult(t *testing.T) {
	nodeID := uuid.New()

	mockRepo := &MockEmbeddingRepository{
		findSimilarFunc: func(ctx context.Context, id uuid.UUID, limit int) ([]apprec.SimilarNote, error) {
			return []apprec.SimilarNote{}, nil
		},
	}

	loader := NewEmbeddingNeighborLoader(mockRepo, 30)

	edges, err := loader.GetNeighbors(context.Background(), nodeID)

	require.NoError(t, err)
	assert.Empty(t, edges)
}

func TestEmbeddingNeighborLoader_GetNeighbors_Error(t *testing.T) {
	nodeID := uuid.New()

	mockRepo := &MockEmbeddingRepository{
		findSimilarFunc: func(ctx context.Context, id uuid.UUID, limit int) ([]apprec.SimilarNote, error) {
			return nil, assert.AnError
		},
	}

	loader := NewEmbeddingNeighborLoader(mockRepo, 30)

	edges, err := loader.GetNeighbors(context.Background(), nodeID)

	assert.Error(t, err)
	assert.Nil(t, edges)
}

func TestEmbeddingNeighborLoader_GetNeighborsBatch_Success(t *testing.T) {
	nodeID1 := uuid.New()
	nodeID2 := uuid.New()

	mockRepo := &MockEmbeddingRepository{
		findSimilarBatchFunc: func(ctx context.Context, nodeIDs []uuid.UUID, limit int) (map[uuid.UUID][]apprec.SimilarNote, error) {
			return map[uuid.UUID][]apprec.SimilarNote{
				nodeID1: {
					{NoteID: uuid.New(), Score: 0.9},
					{NoteID: uuid.New(), Score: 0.8},
				},
				nodeID2: {
					{NoteID: uuid.New(), Score: 0.85},
				},
			}, nil
		},
	}

	loader := NewEmbeddingNeighborLoader(mockRepo, 30)

	result, err := loader.GetNeighborsBatch(context.Background(), []uuid.UUID{nodeID1, nodeID2})

	require.NoError(t, err)
	require.Len(t, result, 2)

	assert.Len(t, result[nodeID1], 2)
	assert.Equal(t, 0.9, result[nodeID1][0].Weight)
	assert.Equal(t, 0.8, result[nodeID1][1].Weight)

	assert.Len(t, result[nodeID2], 1)
	assert.Equal(t, 0.85, result[nodeID2][0].Weight)
}

func TestEmbeddingNeighborLoader_GetNeighborsBatch_EmptyNodeIDs(t *testing.T) {
	mockRepo := &MockEmbeddingRepository{}

	loader := NewEmbeddingNeighborLoader(mockRepo, 30)

	result, err := loader.GetNeighborsBatch(context.Background(), []uuid.UUID{})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestEmbeddingNeighborLoader_GetNeighborsBatch_BatchFails_FallbackToSequential(t *testing.T) {
	nodeID1 := uuid.New()
	nodeID2 := uuid.New()

	mockRepo := &MockEmbeddingRepository{
		findSimilarBatchFunc: func(ctx context.Context, nodeIDs []uuid.UUID, limit int) (map[uuid.UUID][]apprec.SimilarNote, error) {
			return nil, assert.AnError // Batch fails
		},
		findSimilarFunc: func(ctx context.Context, id uuid.UUID, limit int) ([]apprec.SimilarNote, error) {
			return []apprec.SimilarNote{
				{NoteID: uuid.New(), Score: 0.9},
			}, nil
		},
	}

	loader := NewEmbeddingNeighborLoader(mockRepo, 30)

	result, err := loader.GetNeighborsBatch(context.Background(), []uuid.UUID{nodeID1, nodeID2})

	require.NoError(t, err)
	require.Len(t, result, 2)

	assert.Len(t, result[nodeID1], 1)
	assert.Len(t, result[nodeID2], 1)
}

func TestEmbeddingNeighborLoader_GetNeighborsBatch_SomeNodesFail(t *testing.T) {
	nodeID1 := uuid.New()
	nodeID2 := uuid.New()

	mockRepo := &MockEmbeddingRepository{
		findSimilarBatchFunc: func(ctx context.Context, nodeIDs []uuid.UUID, limit int) (map[uuid.UUID][]apprec.SimilarNote, error) {
			// Only return result for nodeID1, skip nodeID2
			return map[uuid.UUID][]apprec.SimilarNote{
				nodeID1: {
					{NoteID: uuid.New(), Score: 0.9},
				},
			}, nil
		},
	}

	loader := NewEmbeddingNeighborLoader(mockRepo, 30)

	result, err := loader.GetNeighborsBatch(context.Background(), []uuid.UUID{nodeID1, nodeID2})

	require.NoError(t, err)
	require.Len(t, result, 2)

	assert.Len(t, result[nodeID1], 1)
	assert.Len(t, result[nodeID2], 0) // Empty but present
}
