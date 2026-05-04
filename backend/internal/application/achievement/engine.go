// Package achievement provides the achievement engine
package achievement

import (
	"context"
	"fmt"

	achievementDomain "knowledge-graph/internal/domain/achievement"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Engine evaluates achievement conditions
type Engine struct {
	db *gorm.DB
}

// NewEngine creates a new achievement engine
func NewEngine(db *gorm.DB) *Engine {
	return &Engine{db: db}
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

	switch condition.Entity {
	case "note":
		query := e.db.WithContext(ctx).Model(&struct {
			CreatorID *uuid.UUID
			Type      string
			DeletedAt interface{}
		}{}).
			Table("notes").
			Where("creator_id = ? AND deleted_at IS NULL", userID)

		// Apply type filter if specified
		if condition.Filter != nil {
			if noteType, ok := condition.Filter["type"].(string); ok && noteType != "" {
				query = query.Where("type = ?", noteType)
			}
		}

		err = query.Count(&count).Error

	case "link":
		query := e.db.WithContext(ctx).Model(&struct {
			CreatorID *uuid.UUID
			DeletedAt interface{}
		}{}).
			Table("links").
			Where("creator_id = ? AND deleted_at IS NULL", userID)

		err = query.Count(&count).Error

	case "search":
		// For search, we'd need a search history table
		// For now, return false
		count = 0

	case "share":
		// Count note shares where user is the sharer
		err = e.db.WithContext(ctx).
			Model(&struct{ SharedByUserID uuid.UUID }{}).
			Table("note_shares").
			Where("shared_by_user_id = ?", userID).
			Count(&count).Error

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
	// For streaks, we'd use Redis to track consecutive days
	// For now, return false as this requires Redis integration
	// The streak tracking is handled by the service layer
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
