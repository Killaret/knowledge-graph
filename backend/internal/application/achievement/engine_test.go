package achievement

import (
	"context"
	"errors"
	"testing"

	achievementDomain "knowledge-graph/internal/domain/achievement"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type fakeCounter struct {
	noteCount  int64
	linkCount  int64
	shareCount int64
	err        error
}

func (f *fakeCounter) CountNotes(ctx context.Context, userID uuid.UUID, noteType string) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	if noteType != "" {
		// Simulate filtering by returning a fixed filtered count.
		// Real tests can use a more elaborate fake if needed.
		return 3, nil
	}
	return f.noteCount, nil
}

func (f *fakeCounter) CountLinks(ctx context.Context, userID uuid.UUID) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	return f.linkCount, nil
}

func (f *fakeCounter) CountShares(ctx context.Context, userID uuid.UUID) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	return f.shareCount, nil
}

func TestEngine_Evaluate_Count_Condition(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("evaluates note count condition above threshold", func(t *testing.T) {
		engine := NewEngine(&fakeCounter{noteCount: 10})
		condition := achievementDomain.Condition{
			Type:      "count",
			Entity:    "note",
			Action:    "create",
			Threshold: 5,
		}
		result, err := engine.Evaluate(ctx, condition, userID)
		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("returns false when count below threshold", func(t *testing.T) {
		engine := NewEngine(&fakeCounter{noteCount: 5})
		condition := achievementDomain.Condition{
			Type:      "count",
			Entity:    "note",
			Action:    "create",
			Threshold: 10,
		}
		result, err := engine.Evaluate(ctx, condition, userID)
		assert.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("evaluates note count with type filter", func(t *testing.T) {
		engine := NewEngine(&fakeCounter{})
		condition := achievementDomain.Condition{
			Type:      "count",
			Entity:    "note",
			Action:    "create",
			Filter:    map[string]interface{}{"type": "galaxy"},
			Threshold: 2,
		}
		result, err := engine.Evaluate(ctx, condition, userID)
		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("evaluates link count condition", func(t *testing.T) {
		engine := NewEngine(&fakeCounter{linkCount: 10})
		condition := achievementDomain.Condition{
			Type:      "count",
			Entity:    "link",
			Action:    "create",
			Threshold: 5,
		}
		result, err := engine.Evaluate(ctx, condition, userID)
		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("evaluates share count condition", func(t *testing.T) {
		engine := NewEngine(&fakeCounter{shareCount: 7})
		condition := achievementDomain.Condition{
			Type:      "count",
			Entity:    "share",
			Action:    "create",
			Threshold: 5,
		}
		result, err := engine.Evaluate(ctx, condition, userID)
		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("returns error when counter fails", func(t *testing.T) {
		engine := NewEngine(&fakeCounter{err: errors.New("db down")})
		condition := achievementDomain.Condition{
			Type:      "count",
			Entity:    "note",
			Action:    "create",
			Threshold: 5,
		}
		result, err := engine.Evaluate(ctx, condition, userID)
		assert.Error(t, err)
		assert.False(t, result)
	})
}

func TestEngine_Evaluate_Invalid_Condition(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	engine := NewEngine(&fakeCounter{})

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
	ctx := context.Background()
	userID := uuid.New()
	engine := NewEngine(&fakeCounter{})

	t.Run("returns false for streak (not implemented in engine)", func(t *testing.T) {
		condition := achievementDomain.Condition{
			Type:      "streak",
			Action:    "login",
			Threshold: 7,
		}
		result, err := engine.Evaluate(ctx, condition, userID)
		assert.NoError(t, err)
		assert.False(t, result)
	})
}
