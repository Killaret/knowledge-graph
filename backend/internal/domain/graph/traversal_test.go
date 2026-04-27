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

func (m *MockNeighborLoader) GetNeighborsBatch(ctx context.Context, nodeIDs []uuid.UUID) (map[uuid.UUID][]Edge, error) {
	args := m.Called(ctx, nodeIDs)
	return args.Get(0).(map[uuid.UUID][]Edge), args.Error(1)
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

		// normalize=false чтобы проверить raw вес
		svc := NewTraversalService(loader, 3, 0.5, "max", false)
		result := svc.RunBFSWeights(ctx, startID)

		assert.Len(t, result, 1)
		// Decay не применяется на первом хопе (depth=0)
		assert.InDelta(t, 0.8, result[targetID], 0.001)
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

		// normalize=false чтобы проверить raw веса
		svc := NewTraversalService(loader, 3, 0.5, "max", false)
		result := svc.RunBFSWeights(ctx, startID)

		assert.Len(t, result, 2)
		// Decay не применяется на первом хопе (depth=0)
		assert.InDelta(t, 0.9, result[target1], 0.001)
		assert.InDelta(t, 0.6, result[target2], 0.001)
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

		// normalize=false чтобы проверить raw веса
		svc := NewTraversalService(loader, 3, 0.5, "max", false)
		result := svc.RunBFSWeights(ctx, startID)

		assert.Len(t, result, 2)
		// B: 0.8 (без decay на первом хопе)
		assert.InDelta(t, 0.8, result[midID], 0.001)
		// C: 0.8 * 0.7 * 0.5 (decay на втором хопе) = 0.28
		assert.InDelta(t, 0.28, result[targetID], 0.001)
		loader.AssertExpectations(t)
	})

	t.Run("MAX strategy - keep best path", func(t *testing.T) {
		// A -> B (weight 0.3)
		// A -> C (weight 0.8) -> B (weight 0.9)
		// Пути к B:
		// - Direct: 0.3 (без decay)
		// - Via C: 0.8 * 0.9 * 0.5 (decay на C->B) = 0.36
		// MAX должен выбрать 0.36
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

		// normalize=false чтобы проверить raw веса
		svc := NewTraversalService(loader, 3, 0.5, "max", false)
		result := svc.RunBFSWeights(ctx, startID)

		assert.Len(t, result, 2)
		// Best weight for B via C: 0.8 * 0.9 * 0.5 = 0.36
		assert.InDelta(t, 0.36, result[targetID], 0.001)
		loader.AssertExpectations(t)
	})

	t.Run("re-queue when better path found", func(t *testing.T) {
		// A -> B (0.4) -> D (1.0)
		// A -> C (0.9) -> B (1.0) -> D
		// Пути к B:
		// - Direct: 0.4 (без decay на первом хопе)
		// - Via C: 0.9 * 1.0 * 0.5 (decay на C->B) = 0.45
		// Пути к D:
		// - Via direct B: 0.4 * 1.0 * 0.5 (decay) = 0.2
		// - Via improved B: 0.45 * 1.0 * 0.5 (decay) = 0.225
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

		// normalize=false чтобы проверить raw веса
		svc := NewTraversalService(loader, 3, 0.5, "max", false)
		result := svc.RunBFSWeights(ctx, startID)

		// B: best = 0.45 (via C: 0.9*1.0*0.5), direct would be 0.4
		assert.InDelta(t, 0.45, result[bID], 0.001)
		// D: best via improved B = 0.45 * 1.0 * 0.5 = 0.225
		assert.InDelta(t, 0.225, result[dID], 0.001)
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

		svc := NewTraversalService(loader, 2, 0.5, "max", true) // depth = 2
		result := svc.RunBFSWeights(ctx, startID)

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

		svc := NewTraversalService(loader, 3, 0.5, "max", true)
		result := svc.RunBFSWeights(ctx, startID)

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

		svc := NewTraversalService(loader, 3, 0.5, "max", true)
		result := svc.RunBFSWeights(ctx, startID)

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

		// normalize=false чтобы проверить raw вес
		svc := NewTraversalService(loader, 3, 0.5, "max", false)
		result := svc.RunBFSWeights(ctx, startID)

		// Should still have result for targetID even if its neighbors failed to load
		// Decay не применяется на первом хопе, поэтому вес = 0.8
		assert.Len(t, result, 1)
		assert.InDelta(t, 0.8, result[targetID], 0.001)
		loader.AssertExpectations(t)
	})

	t.Run("empty graph - no neighbors", func(t *testing.T) {
		startID := uuid.New()

		loader := new(MockNeighborLoader)
		loader.On("GetNeighbors", ctx, startID).Return([]Edge{}, nil)

		svc := NewTraversalService(loader, 3, 0.5, "max", true)
		result := svc.RunBFSWeights(ctx, startID)

		assert.Len(t, result, 0)
		loader.AssertExpectations(t)
	})
}

func TestTraversalService_MaxAggregation(t *testing.T) {
	// Граф: A→B(0.8), A→C(0.5), B→D(0.9), C→D(0.9)
	// При depth=2, decay=0.5:
	// Путь A→B: 0.8 (без decay на первом хопе)
	// Путь A→C: 0.5 (без decay на первом хопе)
	// Путь A→B→D: 0.8 * 0.9 * 0.5 (decay на хопе B→D) = 0.36
	// Путь A→C→D: 0.5 * 0.9 * 0.5 (decay на хопе C→D) = 0.225
	// MAX должен выбрать 0.36 (путь через B)
	ctx := context.Background()

	aID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	bID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	cID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	dID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")

	loader := new(MockNeighborLoader)
	loader.On("GetNeighbors", ctx, aID).Return([]Edge{
		{From: aID, To: bID, Weight: 0.8},
		{From: aID, To: cID, Weight: 0.5},
	}, nil)
	loader.On("GetNeighbors", ctx, bID).Return([]Edge{
		{From: bID, To: dID, Weight: 0.9},
	}, nil)
	loader.On("GetNeighbors", ctx, cID).Return([]Edge{
		{From: cID, To: dID, Weight: 0.9},
	}, nil)
	// GetNeighbors не будет вызван для dID, т.к. depth=2 и dID на глубине 2

	svc := NewTraversalService(loader, 2, 0.5, "max", true)
	result := svc.RunBFSWeights(ctx, aID)

	// D должен иметь вес 0.36 (максимальный путь через B)
	// Нормализация: максимум = 0.8 (B), D = 0.36/0.8 = 0.45
	assert.InDelta(t, 0.45, result[dID], 0.001)
	loader.AssertExpectations(t)
}

func TestTraversalService_SumAggregation(t *testing.T) {
	// Граф: A→B(0.8), A→C(0.5), B→D(0.9), C→D(0.9)
	// При depth=2, decay=0.5:
	// Путь A→B: 0.8 (без decay на первом хопе)
	// Путь A→C: 0.5 (без decay на первом хопе)
	// Путь A→B→D: 0.8 * 0.9 * 0.5 (decay на хопе B→D) = 0.36
	// Путь A→C→D: 0.5 * 0.9 * 0.5 (decay на хопе C→D) = 0.225
	// SUM для D должен дать 0.36 + 0.225 = 0.585
	ctx := context.Background()

	aID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	bID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	cID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	dID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")

	loader := new(MockNeighborLoader)
	loader.On("GetNeighbors", ctx, aID).Return([]Edge{
		{From: aID, To: bID, Weight: 0.8},
		{From: aID, To: cID, Weight: 0.5},
	}, nil)
	loader.On("GetNeighbors", ctx, bID).Return([]Edge{
		{From: bID, To: dID, Weight: 0.9},
	}, nil)
	loader.On("GetNeighbors", ctx, cID).Return([]Edge{
		{From: cID, To: dID, Weight: 0.9},
	}, nil)
	// GetNeighbors не будет вызван для dID, т.к. depth=2 и dID на глубине 2

	svc := NewTraversalService(loader, 2, 0.5, "sum", true)
	result := svc.RunBFSWeights(ctx, aID)

	// D должен иметь вес 0.585 (сумма путей через B и C: 0.36 + 0.225)
	// Нормализация: максимум = 0.8 (B), D = 0.585/0.8 = 0.73125
	assert.InDelta(t, 0.73125, result[dID], 0.001)
	loader.AssertExpectations(t)
}

func TestTraversalService_Normalization(t *testing.T) {
	ctx := context.Background()
	startID := uuid.New()
	target1 := uuid.New()
	target2 := uuid.New()

	loader := new(MockNeighborLoader)
	loader.On("GetNeighbors", ctx, startID).Return([]Edge{
		{From: startID, To: target1, Weight: 0.9},
		{From: startID, To: target2, Weight: 0.3},
	}, nil)
	loader.On("GetNeighbors", ctx, target1).Return([]Edge{}, nil)
	loader.On("GetNeighbors", ctx, target2).Return([]Edge{}, nil)

	svc := NewTraversalService(loader, 3, 0.5, "max", true)
	result := svc.RunBFSWeights(ctx, startID)

	// С нормализацией: максимальный вес должен быть 1.0
	maxVal := 0.0
	for _, v := range result {
		if v > maxVal {
			maxVal = v
		}
	}
	assert.InDelta(t, 1.0, maxVal, 0.001)
	// target2 должен быть 0.3/0.9 = 0.333
	assert.InDelta(t, 0.333, result[target2], 0.01)
}

func TestTraversalService_NoNormalization(t *testing.T) {
	ctx := context.Background()
	startID := uuid.New()
	target1 := uuid.New()
	target2 := uuid.New()

	loader := new(MockNeighborLoader)
	loader.On("GetNeighbors", ctx, startID).Return([]Edge{
		{From: startID, To: target1, Weight: 0.9},
		{From: startID, To: target2, Weight: 0.3},
	}, nil)
	loader.On("GetNeighbors", ctx, target1).Return([]Edge{}, nil)
	loader.On("GetNeighbors", ctx, target2).Return([]Edge{}, nil)

	svc := NewTraversalService(loader, 3, 0.5, "max", false)
	result := svc.RunBFSWeights(ctx, startID)

	// Без нормализации: веса без изменений (decay не применяется на первом хопе)
	assert.InDelta(t, 0.9, result[target1], 0.001) // 0.9 (без decay на depth=0)
	assert.InDelta(t, 0.3, result[target2], 0.001) // 0.3 (без decay на depth=0)
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

		svc := NewTraversalService(loader, 3, 0.5, "max", true)
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

		svc := NewTraversalService(loader, 3, 0.5, "max", true)
		results, err := svc.GetSuggestions(ctx, startID, 5)

		assert.NoError(t, err)
		assert.Len(t, results, 1)
		loader.AssertExpectations(t)
	})

	t.Run("empty results when no neighbors", func(t *testing.T) {
		startID := uuid.New()

		loader := new(MockNeighborLoader)
		loader.On("GetNeighbors", ctx, startID).Return([]Edge{}, nil)

		svc := NewTraversalService(loader, 3, 0.5, "max", true)
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
		// normalize=false чтобы проверить raw веса
		svc := NewTraversalService(loader, 0, 1.5, "max", false)
		result := svc.RunBFSWeights(ctx, startID)

		// With default decay 0.5, но decay не применяется на первом хопе (depth=0)
		// Поэтому вес = 1.0 * 1.0 = 1.0
		assert.InDelta(t, 1.0, result[targetID], 0.001)
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
		svc := NewTraversalService(loader, 1, 0.5, "max", true)
		result := svc.RunBFSWeights(ctx, startID)

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
		// normalize=false чтобы проверить raw веса
		svc := NewTraversalService(loader, 3, 0.25, "max", false)
		result := svc.RunBFSWeights(ctx, startID)

		// Decay не применяется на первом хопе (depth=0), поэтому вес = 1.0 * 1.0 = 1.0
		assert.InDelta(t, 1.0, result[targetID], 0.001)
		loader.AssertExpectations(t)
	})
}
