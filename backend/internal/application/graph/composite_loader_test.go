package graph

import (
	"context"
	"testing"

	"knowledge-graph/internal/domain/graph"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNeighborLoader - мок для NeighborLoader
type MockNeighborLoader struct {
	mock.Mock
}

func (m *MockNeighborLoader) GetNeighbors(ctx context.Context, nodeID uuid.UUID) ([]graph.Edge, error) {
	args := m.Called(ctx, nodeID)
	return args.Get(0).([]graph.Edge), args.Error(1)
}

func TestCompositeNeighborLoader_GetNeighbors(t *testing.T) {
	ctx := context.Background()
	nodeID := uuid.New()

	t.Run("single loader with weight 1.0", func(t *testing.T) {
		loader1 := new(MockNeighborLoader)
		loader1.On("GetNeighbors", ctx, nodeID).Return([]graph.Edge{
			{From: nodeID, To: uuid.New(), Weight: 0.5},
			{From: nodeID, To: uuid.New(), Weight: 0.8},
		}, nil)

		composite := NewCompositeNeighborLoaderWithWeights(
			[]graph.NeighborLoader{loader1},
			[]float64{1.0},
		)

		edges, err := composite.GetNeighbors(ctx, nodeID)

		assert.NoError(t, err)
		assert.Len(t, edges, 2)
		assert.InDelta(t, 0.5, edges[0].Weight, 0.001)
		assert.InDelta(t, 0.8, edges[1].Weight, 0.001)
		loader1.AssertExpectations(t)
	})

	t.Run("two loaders with different weights", func(t *testing.T) {
		loader1 := new(MockNeighborLoader)
		loader2 := new(MockNeighborLoader)

		loader1.On("GetNeighbors", ctx, nodeID).Return([]graph.Edge{
			{From: nodeID, To: uuid.New(), Weight: 0.5},
		}, nil)

		loader2.On("GetNeighbors", ctx, nodeID).Return([]graph.Edge{
			{From: nodeID, To: uuid.New(), Weight: 0.6},
		}, nil)

		// Веса: 0.7 для первого, 0.3 для второго
		composite := NewCompositeNeighborLoaderWithWeights(
			[]graph.NeighborLoader{loader1, loader2},
			[]float64{0.7, 0.3},
		)

		edges, err := composite.GetNeighbors(ctx, nodeID)

		assert.NoError(t, err)
		assert.Len(t, edges, 2)

		// Первый loader: 0.5 * 0.7 = 0.35
		assert.InDelta(t, 0.35, edges[0].Weight, 0.001)

		// Второй loader: 0.6 * 0.3 = 0.18
		assert.InDelta(t, 0.18, edges[1].Weight, 0.001)

		loader1.AssertExpectations(t)
		loader2.AssertExpectations(t)
	})

	t.Run("loader returns error - should continue with others", func(t *testing.T) {
		loader1 := new(MockNeighborLoader)
		loader2 := new(MockNeighborLoader)

		loader1.On("GetNeighbors", ctx, nodeID).Return([]graph.Edge{}, assert.AnError)

		loader2.On("GetNeighbors", ctx, nodeID).Return([]graph.Edge{
			{From: nodeID, To: uuid.New(), Weight: 0.6},
		}, nil)

		composite := NewCompositeNeighborLoaderWithWeights(
			[]graph.NeighborLoader{loader1, loader2},
			[]float64{0.5, 0.5},
		)

		edges, err := composite.GetNeighbors(ctx, nodeID)

		assert.NoError(t, err) // Ошибка не прокидывается
		assert.Len(t, edges, 1)
		assert.InDelta(t, 0.3, edges[0].Weight, 0.001) // 0.6 * 0.5 = 0.3

		loader1.AssertExpectations(t)
		loader2.AssertExpectations(t)
	})

	t.Run("missing weight defaults to 1.0", func(t *testing.T) {
		loader1 := new(MockNeighborLoader)

		loader1.On("GetNeighbors", ctx, nodeID).Return([]graph.Edge{
			{From: nodeID, To: uuid.New(), Weight: 0.5},
		}, nil)

		// Не передаем веса - должно использоваться 1.0
		composite := NewCompositeNeighborLoaderWithWeights(
			[]graph.NeighborLoader{loader1},
			[]float64{},
		)

		edges, err := composite.GetNeighbors(ctx, nodeID)

		assert.NoError(t, err)
		assert.Len(t, edges, 1)
		assert.InDelta(t, 0.5, edges[0].Weight, 0.001) // 0.5 * 1.0 = 0.5

		loader1.AssertExpectations(t)
	})

	t.Run("weight multiplication with zero", func(t *testing.T) {
		loader1 := new(MockNeighborLoader)

		loader1.On("GetNeighbors", ctx, nodeID).Return([]graph.Edge{
			{From: nodeID, To: uuid.New(), Weight: 0.8},
		}, nil)

		// Вес 0 - ребро должно иметь вес 0
		composite := NewCompositeNeighborLoaderWithWeights(
			[]graph.NeighborLoader{loader1},
			[]float64{0.0},
		)

		edges, err := composite.GetNeighbors(ctx, nodeID)

		assert.NoError(t, err)
		assert.Len(t, edges, 1)
		assert.InDelta(t, 0.0, edges[0].Weight, 0.001) // 0.8 * 0 = 0

		loader1.AssertExpectations(t)
	})

	t.Run("multiple edges from same loader", func(t *testing.T) {
		loader1 := new(MockNeighborLoader)

		loader1.On("GetNeighbors", ctx, nodeID).Return([]graph.Edge{
			{From: nodeID, To: uuid.New(), Weight: 0.2},
			{From: nodeID, To: uuid.New(), Weight: 0.4},
			{From: nodeID, To: uuid.New(), Weight: 0.6},
		}, nil)

		composite := NewCompositeNeighborLoaderWithWeights(
			[]graph.NeighborLoader{loader1},
			[]float64{0.5},
		)

		edges, err := composite.GetNeighbors(ctx, nodeID)

		assert.NoError(t, err)
		assert.Len(t, edges, 3)
		assert.InDelta(t, 0.1, edges[0].Weight, 0.001) // 0.2 * 0.5
		assert.InDelta(t, 0.2, edges[1].Weight, 0.001) // 0.4 * 0.5
		assert.InDelta(t, 0.3, edges[2].Weight, 0.001) // 0.6 * 0.5

		loader1.AssertExpectations(t)
	})
}

func TestNewCompositeNeighborLoaderWithWeights(t *testing.T) {
	t.Run("create loader with valid inputs", func(t *testing.T) {
		loader1 := new(MockNeighborLoader)
		loader2 := new(MockNeighborLoader)

		loaders := []graph.NeighborLoader{loader1, loader2}
		weights := []float64{0.6, 0.4}

		composite := NewCompositeNeighborLoaderWithWeights(loaders, weights)

		assert.NotNil(t, composite)
		_, ok := composite.(*compositeNeighborLoader)
		assert.True(t, ok)
	})
}
