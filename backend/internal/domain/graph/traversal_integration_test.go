//go:build integration
// +build integration

package graph

import (
	"context"
	"os"
	"testing"

	"knowledge-graph/internal/application/graph"
	"knowledge-graph/internal/infrastructure/db/postgres"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestIntegration_BFSWithRealRepository тестирует BFS с реальными репозиториями
func TestIntegration_BFSWithRealRepository(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://kb_user:kb_pass@localhost:5432/knowledge_base?sslmode=disable"
	}

	db, err := gorm.Open(pgdriver.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	ctx := context.Background()

	// Создаем реальные репозитории
	linkRepo := postgres.NewLinkRepository(db)
	
	// Создаем загрузчик
	loader := graph.NewNeighborLoader(linkRepo)

	// Тестовый граф:
	// A (a0000000-0000-0000-0000-000000000001)
	// ├── B (a0000000-0000-0000-0000-000000000002) weight=0.8
	// └── C (a0000000-0000-0000-0000-000000000003) weight=0.5
	// B └── D (a0000000-0000-0000-0000-000000000004) weight=0.9
	// C └── D (a0000000-0000-0000-0000-000000000004) weight=0.9

	startID := uuid.MustParse("a0000000-0000-0000-0000-000000000001")
	targetD := uuid.MustParse("a0000000-0000-0000-0000-000000000004")

	t.Run("MAX aggregation with real repository", func(t *testing.T) {
		svc := NewTraversalService(loader, 3, 0.5, "max", false)
		result := svc.runBFS(ctx, startID)

		t.Logf("MAX aggregation results: %v", result)

		// При MAX агрегации D должен иметь вес через лучший путь (через B):
		// A->B: 0.8 (без decay)
		// B->D: 0.8 * 0.9 * 0.5 (decay на втором хопе) = 0.36
		require.Contains(t, result, targetD)
		assert.InDelta(t, 0.36, result[targetD], 0.01)
	})

	t.Run("SUM aggregation with real repository", func(t *testing.T) {
		svc := NewTraversalService(loader, 3, 0.5, "sum", false)
		result := svc.runBFS(ctx, startID)

		t.Logf("SUM aggregation results: %v", result)

		// При SUM агрегации D должен иметь сумму весов всех путей:
		// A->B->D: 0.8 * 0.9 * 0.5 = 0.36
		// A->C->D: 0.5 * 0.9 * 0.5 = 0.225
		// Сумма: 0.36 + 0.225 = 0.585
		require.Contains(t, result, targetD)
		assert.InDelta(t, 0.585, result[targetD], 0.01)
	})

	t.Run("MAX with normalization", func(t *testing.T) {
		svc := NewTraversalService(loader, 3, 0.5, "max", true)
		result := svc.runBFS(ctx, startID)

		t.Logf("MAX with normalization results: %v", result)

		// С нормализацией, максимальный вес должен быть 1.0
		maxWeight := 0.0
		for _, w := range result {
			if w > maxWeight {
				maxWeight = w
			}
		}
		assert.InDelta(t, 1.0, maxWeight, 0.01)

		// D должен иметь вес 0.36 / 0.8 = 0.45
		require.Contains(t, result, targetD)
		assert.InDelta(t, 0.45, result[targetD], 0.01)
	})
}

// TestIntegration_BatchLoading проверяет batch-загрузку соседей
func TestIntegration_BatchLoading(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://kb_user:kb_pass@localhost:5432/knowledge_base?sslmode=disable"
	}

	db, err := gorm.Open(pgdriver.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	ctx := context.Background()

	linkRepo := postgres.NewLinkRepository(db)
	loader := graph.NewNeighborLoader(linkRepo)

	// Тестируем batch-загрузку
	nodeIDs := []uuid.UUID{
		uuid.MustParse("a0000000-0000-0000-0000-000000000001"),
		uuid.MustParse("a0000000-0000-0000-0000-000000000002"),
	}

	result, err := loader.GetNeighborsBatch(ctx, nodeIDs)
	require.NoError(t, err)

	t.Logf("Batch loading results: %v", result)

	// Должны получить соседей для обоих узлов
	assert.Len(t, result, 2)
	assert.Contains(t, result, nodeIDs[0])
	assert.Contains(t, result, nodeIDs[1])

	// Узел A должен иметь соседей B и C
	neighborsA := result[nodeIDs[0]]
	assert.Len(t, neighborsA, 2)

	// Узел B должен иметь соседа D и обратную связь с A
	neighborsB := result[nodeIDs[1]]
	assert.NotEmpty(t, neighborsB)
}
