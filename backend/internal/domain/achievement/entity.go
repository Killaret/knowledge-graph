// Package achievement provides domain types for the achievement system
package achievement

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Achievement represents a system-wide achievement definition
type Achievement struct {
	id          uuid.UUID
	code        string
	title       string
	description string
	icon        string
	condition   Condition
	points      int
	isHidden    bool
	createdAt   time.Time
}

// Condition represents the condition for earning an achievement
type Condition struct {
	Type      string                 `json:"type"`   // "count", "streak"
	Entity    string                 `json:"entity"` // "note", "link", "search", "share"
	Action    string                 `json:"action"` // "create", "execute", "login"
	Filter    map[string]interface{} `json:"filter,omitempty"`
	Threshold int                    `json:"threshold"` // Required count
}

// Validate checks if the condition is valid
func (c *Condition) Validate() error {
	validTypes := map[string]bool{"count": true, "streak": true}
	validEntities := map[string]bool{"note": true, "link": true, "search": true, "share": true, "login": true}

	if !validTypes[c.Type] {
		return fmt.Errorf("invalid condition type: %s", c.Type)
	}

	if c.Type == "count" && !validEntities[c.Entity] {
		return fmt.Errorf("invalid entity: %s", c.Entity)
	}

	if c.Threshold <= 0 {
		return fmt.Errorf("threshold must be positive")
	}

	return nil
}

// NewAchievement creates a new achievement
func NewAchievement(code, title, description, icon string, condition Condition, points int, isHidden bool) (*Achievement, error) {
	if err := condition.Validate(); err != nil {
		return nil, fmt.Errorf("invalid condition: %w", err)
	}

	return &Achievement{
		id:          uuid.New(),
		code:        code,
		title:       title,
		description: description,
		icon:        icon,
		condition:   condition,
		points:      points,
		isHidden:    isHidden,
		createdAt:   time.Now(),
	}, nil
}

// ReconstructAchievement reconstructs an achievement from stored data
func ReconstructAchievement(id uuid.UUID, code, title, description, icon string, conditionJSON json.RawMessage, points int, isHidden bool, createdAt time.Time) (*Achievement, error) {
	var condition Condition
	if err := json.Unmarshal(conditionJSON, &condition); err != nil {
		return nil, fmt.Errorf("failed to unmarshal condition: %w", err)
	}

	return &Achievement{
		id:          id,
		code:        code,
		title:       title,
		description: description,
		icon:        icon,
		condition:   condition,
		points:      points,
		isHidden:    isHidden,
		createdAt:   createdAt,
	}, nil
}

// ID returns the achievement ID
func (a *Achievement) ID() uuid.UUID {
	return a.id
}

// Code returns the unique code
func (a *Achievement) Code() string {
	return a.code
}

// Title returns the title
func (a *Achievement) Title() string {
	return a.title
}

// Description returns the description
func (a *Achievement) Description() string {
	return a.description
}

// Icon returns the icon identifier
func (a *Achievement) Icon() string {
	return a.icon
}

// Condition returns the condition
func (a *Achievement) Condition() Condition {
	return a.condition
}

// Points returns the points value
func (a *Achievement) Points() int {
	return a.points
}

// IsHidden returns true if the achievement is hidden
func (a *Achievement) IsHidden() bool {
	return a.isHidden
}

// CreatedAt returns creation time
func (a *Achievement) CreatedAt() time.Time {
	return a.createdAt
}

// UserAchievement represents an achievement earned by a user
type UserAchievement struct {
	userID        uuid.UUID
	achievementID uuid.UUID
	obtainedAt    time.Time
	metadata      json.RawMessage
}

// NewUserAchievement creates a new user achievement
func NewUserAchievement(userID, achievementID uuid.UUID) *UserAchievement {
	return &UserAchievement{
		userID:        userID,
		achievementID: achievementID,
		obtainedAt:    time.Now(),
		metadata:      json.RawMessage(`{}`),
	}
}

// ReconstructUserAchievement reconstructs from stored data
func ReconstructUserAchievement(userID, achievementID uuid.UUID, obtainedAt time.Time, metadata json.RawMessage) *UserAchievement {
	return &UserAchievement{
		userID:        userID,
		achievementID: achievementID,
		obtainedAt:    obtainedAt,
		metadata:      metadata,
	}
}

// UserID returns the user ID
func (ua *UserAchievement) UserID() uuid.UUID {
	return ua.userID
}

// AchievementID returns the achievement ID
func (ua *UserAchievement) AchievementID() uuid.UUID {
	return ua.achievementID
}

// ObtainedAt returns when the achievement was earned
func (ua *UserAchievement) ObtainedAt() time.Time {
	return ua.obtainedAt
}

// Repository defines the interface for achievement storage
type Repository interface {
	FindAll(ctx context.Context) ([]Achievement, error)
	FindByCode(ctx context.Context, code string) (*Achievement, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]Achievement, error)
	SaveUserAchievement(ctx context.Context, ua UserAchievement) error
	UserHasAchievement(ctx context.Context, userID uuid.UUID, achievementID uuid.UUID) (bool, error)
}

// Engine defines the interface for evaluating achievement conditions
type Engine interface {
	Evaluate(ctx context.Context, condition Condition, userID uuid.UUID) (bool, error)
}
