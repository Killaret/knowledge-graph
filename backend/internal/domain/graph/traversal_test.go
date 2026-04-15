package graph

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNeighborLoader - мок для NeighborLoader интерфейса
type MockNeighborLoader struct {
	mock.Mock
}

func (m *MockNeighborLoader) GetNeighbors(ctx context.Context, nodeID uuid.UUID) ([]Edge, error) {
	args := m.Called(ctx, nodeID)
	return args.Get(0).([]Edge), args.Error(1)
}

func TestTraversalService_runBFS(t *testing.T) {
	ctx := context.Background()

	t.Run("single edge traversal", func(t *testing.T) {
		startID := uuid.New()
		targetID := uuid.New()

		loader := new(MockNeighborLoader)
		loader.On("GetNeighbors", ctx, startID).Return([]Edge{
			{From: startID, To: targetID, Weight: 0.8},
		}, nil)
		loader.On("GetNeighbors", ctx, targetID).Return([]Edge{}, nil)

		svc := NewTraversalService(loader, 3, 0.5)
		result := svc.runBFS(ctx, startID)

		assert.Len(t, result, 1)
		assert.InDelta(t, 0.8*0.5, result[targetID], 0.001)
		loader.AssertExpectations(t)
	})

	t.Run("multiple edges from start node", func(t *testing.T) {
		startID := uuid.New()
		target1 := uuid.New()
		target2 := uuid.New()

		loader := new(MockNeighborLoader)
		loader.On("GetNeighbors", ctx, startID).Return([]Edge{
			{From: startID, To: target1, Weight: 0.9},
			{From: startID, To: target2, Weight: 0.6},
		}, nil)
		loader.On("GetNeighbors", ctx, target1).Return([]Edge{}, nil)
		loader.On("GetNeighbors", ctx, target2).Return([]Edge{}, nil)

		svc := NewTraversalService(loader, 3, 0.5)
		result := svc.runBFS(ctx, startID)

		assert.Len(t, result, 2)
		assert.InDelta(t, 0.9*0.5, result[target1], 0.001)
		assert.InDelta(t, 0.6*0.5, result[target2], 0.001)
		loader.AssertExpectations(t)
	})

	t.Run("two hop traversal", func(t *testing.T) {
		// A -> B -> C
		startID := uuid.New()
		midID := uuid.New()
		targetID := uuid.New()

		loader := new(MockNeighborLoader)
		loader.On("GetNeighbors", ctx, startID).Return([]Edge{
			{From: startID, To: midID, Weight: 0.8},
		}, nil)
		loader.On("GetNeighbors", ctx, midID).Return([]Edge{
			{From: midID, To: targetID, Weight: 0.7},
		}, nil)
		loader.On("GetNeighbors", ctx, targetID).Return([]Edge{}, nil)

		svc := NewTraversalService(loader, 3, 0.5)
		result := svc.runBFS(ctx, startID)

		assert.Len(t, result, 2)
		// B: 0.8 * 0.5 = 0.4
		assert.InDelta(t, 0.4, result[midID], 0.001)
		// C: 0.8 * 0.5 * 0.7 * 0.5 = 0.14
		assert.InDelta(t, 0.14, result[targetID], 0.001)
		loader.AssertExpectations(t)
	})

	t.Run("MAX strategy - keep best path", func(t *testing.T) {
		// A -> B (weight 0.3)
		// A -> C (weight 0.8) -> B (weight 0.9)
		// Best path to B: via C = 0.8 * 0.5 * 0.9 * 0.5 = 0.18
		// Direct path to B: 0.3 * 0.5 = 0.15
		// Should keep 0.18 (MAX)
		startID := uuid.New()
		midID := uuid.New()
		targetID := uuid.New()

		loader := new(MockNeighborLoader)
		loader.On("GetNeighbors", ctx, startID).Return([]Edge{
			{From: startID, To: targetID, Weight: 0.3}, // Direct weak link
			{From: startID, To: midID, Weight: 0.8},    // Strong link to mid
		}, nil)
		loader.On("GetNeighbors", ctx, midID).Return([]Edge{
			{From: midID, To: targetID, Weight: 0.9}, // Strong link from mid to target
		}, nil)
		loader.On("GetNeighbors", ctx, targetID).Return([]Edge{}, nil)

		svc := NewTraversalService(loader, 3, 0.5)
		result := svc.runBFS(ctx, startID)

		assert.Len(t, result, 2)
		// Best weight for B should be via C: 0.8 * 0.5 * 0.9 * 0.5 = 0.18
		assert.InDelta(t, 0.18, result[targetID], 0.001)
		loader.AssertExpectations(t)
	})

	t.Run("re-queue when better path found", func(t *testing.T) {
		// A -> B (0.4) -> D (1.0)
		// A -> C (0.9) -> B (1.0) -> D
		// First B discovered with weight 0.2 (0.4 * 0.5)
		// Later B discovered with weight 0.225 (0.9 * 0.5 * 1.0 * 0.5)
		// When improved B (0.225) adds D: 0.225 * 1.0 * 0.5 (decay) = 0.1125
		startID := uuid.New()
		bID := uuid.New()
		cID := uuid.New()
		dID := uuid.New()

		loader := new(MockNeighborLoader)
		loader.On("GetNeighbors", ctx, startID).Return([]Edge{
			{From: startID, To: bID, Weight: 0.4}, // Weak direct link
			{From: startID, To: cID, Weight: 0.9}, // Strong link to C
		}, nil)
		loader.On("GetNeighbors", ctx, bID).Return([]Edge{
			{From: bID, To: dID, Weight: 1.0},
		}, nil)
		loader.On("GetNeighbors", ctx, cID).Return([]Edge{
			{From: cID, To: bID, Weight: 1.0}, // Strong link C->B
		}, nil)
		loader.On("GetNeighbors", ctx, dID).Return([]Edge{}, nil)

		svc := NewTraversalService(loader, 3, 0.5)
		result := svc.runBFS(ctx, startID)

		// B: best = 0.225 (via C: 0.9*0.5*1.0*0.5), direct would be 0.2
		assert.InDelta(t, 0.225, result[bID], 0.001)
		// D: reached via improved B = 0.225 * 1.0 * 0.5 = 0.1125
		// (vs 0.2 * 0.5 = 0.1 via original B)
		assert.InDelta(t, 0.1125, result[dID], 0.001)
		loader.AssertExpectations(t)
	})

	t.Run("depth limit respected", func(t *testing.T) {
		// A -> B -> C -> D (depth 2 should stop at C, not expand C's neighbors)
		startID := uuid.New()
		bID := uuid.New()
		cID := uuid.New()
		dID := uuid.New()

		loader := new(MockNeighborLoader)
		loader.On("GetNeighbors", ctx, startID).Return([]Edge{
			{From: startID, To: bID, Weight: 1.0},
		}, nil)
		loader.On("GetNeighbors", ctx, bID).Return([]Edge{
			{From: bID, To: cID, Weight: 1.0},
		}, nil)
		// cID won't have GetNeighbors called because depth=2, and cID is at depth 2

		svc := NewTraversalService(loader, 2, 0.5) // depth = 2
		result := svc.runBFS(ctx, startID)

		// Should have B and C, but not D (depth limit)
		assert.Len(t, result, 2)
		assert.Contains(t, result, bID)
		assert.Contains(t, result, cID)
		assert.NotContains(t, result, dID)
		loader.AssertExpectations(t)
	})

	t.Run("start node excluded from results", func(t *testing.T) {
		startID := uuid.New()

		loader := new(MockNeighborLoader)
		loader.On("GetNeighbors", ctx, startID).Return([]Edge{}, nil)

		svc := NewTraversalService(loader, 3, 0.5)
		result := svc.runBFS(ctx, startID)

		assert.Len(t, result, 0)
		assert.NotContains(t, result, startID)
		loader.AssertExpectations(t)
	})

	t.Run("self-loop ignored", func(t *testing.T) {
		startID := uuid.New()
		targetID := uuid.New()

		loader := new(MockNeighborLoader)
		loader.On("GetNeighbors", ctx, startID).Return([]Edge{
			{From: startID, To: startID, Weight: 0.9}, // Self-loop
			{From: startID, To: targetID, Weight: 0.8},
		}, nil)
		loader.On("GetNeighbors", ctx, targetID).Return([]Edge{}, nil)

		svc := NewTraversalService(loader, 3, 0.5)
		result := svc.runBFS(ctx, startID)

		assert.Len(t, result, 1)
		assert.Contains(t, result, targetID)
		assert.NotContains(t, result, startID)
		loader.AssertExpectations(t)
	})

	t.Run("error in GetNeighbors - continues processing", func(t *testing.T) {
		startID := uuid.New()
		targetID := uuid.New()

		loader := new(MockNeighborLoader)
		loader.On("GetNeighbors", ctx, startID).Return([]Edge{
			{From: startID, To: targetID, Weight: 0.8},
		}, nil)
		loader.On("GetNeighbors", ctx, targetID).Return([]Edge{}, assert.AnError)

		svc := NewTraversalService(loader, 3, 0.5)
		result := svc.runBFS(ctx, startID)

		// Should still have result for targetID even if its neighbors failed to load
		assert.Len(t, result, 1)
		assert.InDelta(t, 0.4, result[targetID], 0.001)
		loader.AssertExpectations(t)
	})

	t.Run("empty graph - no neighbors", func(t *testing.T) {
		startID := uuid.New()

		loader := new(MockNeighborLoader)
		loader.On("GetNeighbors", ctx, startID).Return([]Edge{}, nil)

		svc := NewTraversalService(loader, 3, 0.5)
		result := svc.runBFS(ctx, startID)

		assert.Len(t, result, 0)
		loader.AssertExpectations(t)
	})
}

func TestNormalizeWeights(t *testing.T) {
	t.Run("normal case - scales to max 1.0", func(t *testing.T) {
		input := map[uuid.UUID]float64{
			uuid.New(): 0.5,
			uuid.New(): 1.0,
			uuid.New(): 0.25,
		}

		result := normalizeWeights(input)

		// Max should be 1.0
		maxVal := 0.0
		for _, v := range result {
			if v > maxVal {
				maxVal = v
			}
		}
		assert.InDelta(t, 1.0, maxVal, 0.001)
	})

	t.Run("single value becomes 1.0", func(t *testing.T) {
		id := uuid.New()
		input := map[uuid.UUID]float64{
			id: 0.3,
		}

		result := normalizeWeights(input)

		assert.InDelta(t, 1.0, result[id], 0.001)
	})

	t.Run("empty map returns empty", func(t *testing.T) {
		input := map[uuid.UUID]float64{}

		result := normalizeWeights(input)

		assert.Len(t, result, 0)
	})

	t.Run("all zeros - returns as is", func(t *testing.T) {
		input := map[uuid.UUID]float64{
			uuid.New(): 0.0,
			uuid.New(): 0.0,
		}

		result := normalizeWeights(input)

		// All values should remain 0
		for _, v := range result {
			assert.InDelta(t, 0.0, v, 0.001)
		}
	})

	t.Run("preserves relative ratios", func(t *testing.T) {
		id1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
		id2 := uuid.MustParse("22222222-2222-2222-2222-222222222222")
		id3 := uuid.MustParse("33333333-3333-3333-3333-333333333333")

		input := map[uuid.UUID]float64{
			id1: 0.8,
			id2: 0.4,
			id3: 0.2,
		}

		result := normalizeWeights(input)

		// After normalization: 0.8/0.8=1.0, 0.4/0.8=0.5, 0.2/0.8=0.25
		assert.InDelta(t, 1.0, result[id1], 0.001)
		assert.InDelta(t, 0.5, result[id2], 0.001)
		assert.InDelta(t, 0.25, result[id3], 0.001)
	})
}

func TestTraversalService_GetSuggestions(t *testing.T) {
	ctx := context.Background()

	t.Run("returns top N results sorted by score", func(t *testing.T) {
		startID := uuid.New()
		target1 := uuid.New()
		target2 := uuid.New()
		target3 := uuid.New()

		loader := new(MockNeighborLoader)
		loader.On("GetNeighbors", ctx, startID).Return([]Edge{
			{From: startID, To: target1, Weight: 0.9},
			{From: startID, To: target2, Weight: 0.6},
			{From: startID, To: target3, Weight: 0.3},
		}, nil)
		loader.On("GetNeighbors", ctx, target1).Return([]Edge{}, nil)
		loader.On("GetNeighbors", ctx, target2).Return([]Edge{}, nil)
		loader.On("GetNeighbors", ctx, target3).Return([]Edge{}, nil)

		svc := NewTraversalService(loader, 3, 0.5)
		results, err := svc.GetSuggestions(ctx, startID, 2)

		assert.NoError(t, err)
		assert.Len(t, results, 2)
		// Should be sorted by normalized score (descending)
		// Raw: 0.45, 0.3, 0.15 -> Normalized: 1.0, 0.667, 0.333
		assert.Equal(t, target1, results[0].NodeID)
		assert.Equal(t, target2, results[1].NodeID)
		assert.InDelta(t, 1.0, results[0].Score, 0.001)
		assert.InDelta(t, 0.667, results[1].Score, 0.001)
		loader.AssertExpectations(t)
	})

	t.Run("returns all when less than topN", func(t *testing.T) {
		startID := uuid.New()
		target1 := uuid.New()

		loader := new(MockNeighborLoader)
		loader.On("GetNeighbors", ctx, startID).Return([]Edge{
			{From: startID, To: target1, Weight: 0.8},
		}, nil)
		loader.On("GetNeighbors", ctx, target1).Return([]Edge{}, nil)

		svc := NewTraversalService(loader, 3, 0.5)
		results, err := svc.GetSuggestions(ctx, startID, 5)

		assert.NoError(t, err)
		assert.Len(t, results, 1)
		loader.AssertExpectations(t)
	})

	t.Run("empty results when no neighbors", func(t *testing.T) {
		startID := uuid.New()

		loader := new(MockNeighborLoader)
		loader.On("GetNeighbors", ctx, startID).Return([]Edge{}, nil)

		svc := NewTraversalService(loader, 3, 0.5)
		results, err := svc.GetSuggestions(ctx, startID, 10)

		assert.NoError(t, err)
		assert.Len(t, results, 0)
		loader.AssertExpectations(t)
	})
}

func TestTraversalService_Configuration(t *testing.T) {
	ctx := context.Background()
	startID := uuid.New()
	targetID := uuid.New()

	t.Run("default depth and decay when invalid", func(t *testing.T) {
		loader := new(MockNeighborLoader)
		loader.On("GetNeighbors", ctx, startID).Return([]Edge{
			{From: startID, To: targetID, Weight: 1.0},
		}, nil)
		// Note: targetID won't be processed because depth defaults to 1,
		// and when we process targetID (depth=1), 1 >= 1 so we skip

		// depth=0, decay=1.5 (invalid) -> should use defaults (depth=1, decay=0.5)
		svc := NewTraversalService(loader, 0, 1.5)
		result := svc.runBFS(ctx, startID)

		// With default decay 0.5: 1.0 * 0.5 = 0.5
		assert.InDelta(t, 0.5, result[targetID], 0.001)
		loader.AssertExpectations(t)
	})

	t.Run("custom depth limits traversal", func(t *testing.T) {
		bID := uuid.New()
		cID := uuid.New()

		loader := new(MockNeighborLoader)
		loader.On("GetNeighbors", ctx, startID).Return([]Edge{
			{From: startID, To: bID, Weight: 1.0},
		}, nil)
		// bID won't be expanded because depth=1, and bID is at depth 1
		// cID won't be reached at all

		// depth=1: should only get B, not C
		svc := NewTraversalService(loader, 1, 0.5)
		result := svc.runBFS(ctx, startID)

		assert.Len(t, result, 1)
		assert.Contains(t, result, bID)
		assert.NotContains(t, result, cID)
		loader.AssertExpectations(t)
	})

	t.Run("custom decay affects weights", func(t *testing.T) {
		loader := new(MockNeighborLoader)
		loader.On("GetNeighbors", ctx, startID).Return([]Edge{
			{From: startID, To: targetID, Weight: 1.0},
		}, nil)
		loader.On("GetNeighbors", ctx, targetID).Return([]Edge{}, nil)

		// decay=0.25
		svc := NewTraversalService(loader, 3, 0.25)
		result := svc.runBFS(ctx, startID)

		// 1.0 * 0.25 = 0.25
		assert.InDelta(t, 0.25, result[targetID], 0.001)
		loader.AssertExpectations(t)
	})
}
