package achievement

import (
	"context"
	"testing"

	achievementDomain "knowledge-graph/internal/domain/achievement"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	return gormDB, mock
}

func TestEngine_Evaluate_Count_Condition(t *testing.T) {
	db, mock := setupMockDB(t)
	engine := NewEngine(db)
	userID := uuid.New()
	ctx := context.Background()

	t.Run("evaluates note count condition", func(t *testing.T) {
		condition := achievementDomain.Condition{
			Type:      "count",
			Entity:    "note",
			Action:    "create",
			Threshold: 5,
		}

		// Mock count query returning 10 (above threshold)
		// Use simple pattern that matches the actual query structure
		mock.ExpectQuery(`SELECT count\(\*\) FROM "notes" WHERE .*deleted_at IS NULL`).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

		result, err := engine.Evaluate(ctx, condition, userID)

		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("returns false when count below threshold", func(t *testing.T) {
		condition := achievementDomain.Condition{
			Type:      "count",
			Entity:    "note",
			Action:    "create",
			Threshold: 10,
		}

		// Mock count query returning 5 (below threshold)
		mock.ExpectQuery(`SELECT count\(\*\) FROM "notes" WHERE .*deleted_at IS NULL`).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

		result, err := engine.Evaluate(ctx, condition, userID)

		assert.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("evaluates note count with type filter", func(t *testing.T) {
		condition := achievementDomain.Condition{
			Type:      "count",
			Entity:    "note",
			Action:    "create",
			Filter:    map[string]interface{}{"type": "galaxy"},
			Threshold: 3,
		}

		// Mock count query with type filter
		mock.ExpectQuery(`SELECT count\(\*\) FROM "notes" WHERE .*AND type =`).
			WithArgs(userID, "galaxy").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

		result, err := engine.Evaluate(ctx, condition, userID)

		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("evaluates link count condition", func(t *testing.T) {
		condition := achievementDomain.Condition{
			Type:      "count",
			Entity:    "link",
			Action:    "create",
			Threshold: 5,
		}

		// Mock count query
		mock.ExpectQuery(`SELECT count\(\*\) FROM "links" WHERE .*deleted_at IS NULL`).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

		result, err := engine.Evaluate(ctx, condition, userID)

		assert.NoError(t, err)
		assert.True(t, result)
	})
}

func TestEngine_Evaluate_Invalid_Condition(t *testing.T) {
	db, _ := setupMockDB(t)
	engine := NewEngine(db)
	userID := uuid.New()
	ctx := context.Background()

	t.Run("returns error for invalid condition type", func(t *testing.T) {
		condition := achievementDomain.Condition{
			Type:      "invalid",
			Entity:    "note",
			Action:    "create",
			Threshold: 5,
		}

		result, err := engine.Evaluate(ctx, condition, userID)

		assert.Error(t, err)
		assert.False(t, result)
	})

	t.Run("returns error for invalid entity", func(t *testing.T) {
		condition := achievementDomain.Condition{
			Type:      "count",
			Entity:    "invalid",
			Action:    "create",
			Threshold: 5,
		}

		result, err := engine.Evaluate(ctx, condition, userID)

		assert.Error(t, err)
		assert.False(t, result)
	})

	t.Run("returns error for zero threshold", func(t *testing.T) {
		condition := achievementDomain.Condition{
			Type:      "count",
			Entity:    "note",
			Action:    "create",
			Threshold: 0,
		}

		result, err := engine.Evaluate(ctx, condition, userID)

		assert.Error(t, err)
		assert.False(t, result)
	})
}

func TestEngine_Evaluate_Streak_Condition(t *testing.T) {
	db, _ := setupMockDB(t)
	engine := NewEngine(db)
	userID := uuid.New()
	ctx := context.Background()

	t.Run("returns false for streak (not implemented in engine)", func(t *testing.T) {
		condition := achievementDomain.Condition{
			Type:      "streak",
			Action:    "login",
			Threshold: 7,
		}

		result, err := engine.Evaluate(ctx, condition, userID)

		assert.NoError(t, err) // Streak returns false without error
		assert.False(t, result)
	})
}
