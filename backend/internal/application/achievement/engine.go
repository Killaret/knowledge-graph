// Package achievement provides the achievement engine
package achievement

import (
	"context"
	"fmt"

	achievementDomain "knowledge-graph/internal/domain/achievement"

	"github.com/google/uuid"
)

// Engine evaluates achievement conditions
type Engine struct {
	counter achievementDomain.Counter
}

// NewEngine creates a new achievement engine
func NewEngine(counter achievementDomain.Counter) *Engine {
	return &Engine{counter: counter}
}

// Evaluate evaluates a condition for a user
func (e *Engine) Evaluate(ctx context.Context, condition achievementDomain.Condition, userID uuid.UUID) (bool, error) {
	if err := condition.Validate(); err != nil {
		return false, fmt.Errorf("invalid condition: %w", err)
	}

	switch condition.Type {
	case "count":
		return e.evaluateCount(ctx, condition, userID)
	case "streak":
		return e.evaluateStreak(ctx, condition, userID)
	default:
		return false, fmt.Errorf("unknown condition type: %s", condition.Type)
	}
}

// evaluateCount evaluates count-based conditions
func (e *Engine) evaluateCount(ctx context.Context, condition achievementDomain.Condition, userID uuid.UUID) (bool, error) {
	var count int64
	var err error

	noteType := ""
	if condition.Filter != nil {
		if t, ok := condition.Filter["type"].(string); ok && t != "" {
			noteType = t
		}
	}

	switch condition.Entity {
	case "note":
		count, err = e.counter.CountNotes(ctx, userID, noteType)
	case "link":
		count, err = e.counter.CountLinks(ctx, userID)
	case "search":
		// Search history is not persisted yet
		count = 0
	case "share":
		count, err = e.counter.CountShares(ctx, userID)
	default:
		return false, fmt.Errorf("unknown entity: %s", condition.Entity)
	}

	if err != nil {
		return false, fmt.Errorf("failed to count %s: %w", condition.Entity, err)
	}

	return count >= int64(condition.Threshold), nil
}

// evaluateStreak evaluates streak-based conditions
func (e *Engine) evaluateStreak(ctx context.Context, condition achievementDomain.Condition, userID uuid.UUID) (bool, error) {
	// Streak tracking is handled by the service layer with Redis.
	// The engine itself does not query persistence for streaks.
	return false, nil
}

// TrackLogin updates the login streak for a user
func (e *Engine) TrackLogin(ctx context.Context, userID uuid.UUID, redisClient interface{}) error {
	// This would be implemented with Redis for streak tracking
	// Key: login:streak:{userID}
	// Value: streak count
	// TTL: 48 hours (to allow for missed days)
	return nil
}
