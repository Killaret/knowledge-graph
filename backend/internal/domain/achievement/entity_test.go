package achievement

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCondition_Validate_ValidCountCondition(t *testing.T) {
	condition := Condition{
		Type:      "count",
		Entity:    "note",
		Action:    "create",
		Threshold: 10,
	}

	err := condition.Validate()
	assert.NoError(t, err)
}

func TestCondition_Validate_ValidStreakCondition(t *testing.T) {
	condition := Condition{
		Type:      "streak",
		Entity:    "login",
		Action:    "login",
		Threshold: 7,
	}

	err := condition.Validate()
	assert.NoError(t, err)
}

func TestCondition_Validate_InvalidType(t *testing.T) {
	condition := Condition{
		Type:      "invalid",
		Entity:    "note",
		Action:    "create",
		Threshold: 10,
	}

	err := condition.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid condition type")
}

func TestCondition_Validate_InvalidEntityForCount(t *testing.T) {
	condition := Condition{
		Type:      "count",
		Entity:    "invalid",
		Action:    "create",
		Threshold: 10,
	}

	err := condition.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid entity")
}

func TestCondition_Validate_NegativeThreshold(t *testing.T) {
	condition := Condition{
		Type:      "count",
		Entity:    "note",
		Action:    "create",
		Threshold: -5,
	}

	err := condition.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "threshold must be positive")
}

func TestCondition_Validate_ZeroThreshold(t *testing.T) {
	condition := Condition{
		Type:      "count",
		Entity:    "note",
		Action:    "create",
		Threshold: 0,
	}

	err := condition.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "threshold must be positive")
}

func TestNewAchievement_Success(t *testing.T) {
	condition := Condition{
		Type:      "count",
		Entity:    "note",
		Action:    "create",
		Threshold: 100,
	}

	achievement, err := NewAchievement("first_note", "First Note", "Create your first note", "📝", condition, 50, false)

	require.NoError(t, err)
	assert.NotNil(t, achievement)
	assert.NotEqual(t, uuid.Nil, achievement.ID())
	assert.Equal(t, "first_note", achievement.Code())
	assert.Equal(t, "First Note", achievement.Title())
	assert.Equal(t, "Create your first note", achievement.Description())
	assert.Equal(t, "📝", achievement.Icon())
	assert.Equal(t, condition, achievement.Condition())
	assert.Equal(t, 50, achievement.Points())
	assert.False(t, achievement.IsHidden())
	assert.False(t, achievement.CreatedAt().IsZero())
}

func TestNewAchievement_InvalidCondition(t *testing.T) {
	condition := Condition{
		Type:      "invalid",
		Entity:    "note",
		Action:    "create",
		Threshold: 100,
	}

	_, err := NewAchievement("invalid_achievement", "Invalid", "Description", "🚫", condition, 50, false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid condition")
}

func TestNewAchievement_Hidden(t *testing.T) {
	condition := Condition{
		Type:      "count",
		Entity:    "link",
		Action:    "create",
		Threshold: 50,
	}

	achievement, err := NewAchievement("secret_link", "Secret Link", "Hidden achievement", "🔮", condition, 100, true)

	require.NoError(t, err)
	assert.True(t, achievement.IsHidden())
}

func TestReconstructAchievement_Success(t *testing.T) {
	id := uuid.New()
	condition := Condition{
		Type:      "count",
		Entity:    "note",
		Action:    "create",
		Threshold: 10,
	}
	conditionJSON, _ := json.Marshal(condition)
	createdAt := time.Now().Add(-24 * time.Hour)

	achievement, err := ReconstructAchievement(id, "test_code", "Test Title", "Test Description", "🎯", conditionJSON, 25, false, createdAt)

	require.NoError(t, err)
	assert.Equal(t, id, achievement.ID())
	assert.Equal(t, "test_code", achievement.Code())
	assert.Equal(t, "Test Title", achievement.Title())
	assert.Equal(t, "Test Description", achievement.Description())
	assert.Equal(t, "🎯", achievement.Icon())
	assert.Equal(t, condition, achievement.Condition())
	assert.Equal(t, 25, achievement.Points())
	assert.False(t, achievement.IsHidden())
	assert.Equal(t, createdAt, achievement.CreatedAt())
}

func TestReconstructAchievement_InvalidConditionJSON(t *testing.T) {
	id := uuid.New()
	invalidJSON := json.RawMessage(`{invalid json}`)
	createdAt := time.Now()

	_, err := ReconstructAchievement(id, "test_code", "Test Title", "Test Description", "🎯", invalidJSON, 25, false, createdAt)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal condition")
}

func TestAchievement_Getters(t *testing.T) {
	condition := Condition{
		Type:      "count",
		Entity:    "search",
		Action:    "execute",
		Threshold: 1000,
	}

	achievement, err := NewAchievement("search_master", "Search Master", "Execute 1000 searches", "🔍", condition, 200, false)
	require.NoError(t, err)

	assert.NotEqual(t, uuid.Nil, achievement.ID())
	assert.Equal(t, "search_master", achievement.Code())
	assert.Equal(t, "Search Master", achievement.Title())
	assert.Equal(t, "Execute 1000 searches", achievement.Description())
	assert.Equal(t, "🔍", achievement.Icon())
	assert.Equal(t, condition, achievement.Condition())
	assert.Equal(t, 200, achievement.Points())
	assert.False(t, achievement.IsHidden())
	assert.False(t, achievement.CreatedAt().IsZero())
}

func TestNewUserAchievement(t *testing.T) {
	userID := uuid.New()
	achievementID := uuid.New()

	userAchievement := NewUserAchievement(userID, achievementID)

	assert.NotNil(t, userAchievement)
	assert.Equal(t, userID, userAchievement.UserID())
	assert.Equal(t, achievementID, userAchievement.AchievementID())
	assert.False(t, userAchievement.ObtainedAt().IsZero())
	assert.NotNil(t, userAchievement.metadata)
}

func TestReconstructUserAchievement(t *testing.T) {
	userID := uuid.New()
	achievementID := uuid.New()
	obtainedAt := time.Now().Add(-1 * time.Hour)
	metadata := json.RawMessage(`{"key": "value"}`)

	userAchievement := ReconstructUserAchievement(userID, achievementID, obtainedAt, metadata)

	assert.Equal(t, userID, userAchievement.UserID())
	assert.Equal(t, achievementID, userAchievement.AchievementID())
	assert.Equal(t, obtainedAt, userAchievement.ObtainedAt())
	assert.Equal(t, metadata, userAchievement.metadata)
}

func TestUserAchievement_Getters(t *testing.T) {
	userID := uuid.New()
	achievementID := uuid.New()

	userAchievement := NewUserAchievement(userID, achievementID)

	assert.Equal(t, userID, userAchievement.UserID())
	assert.Equal(t, achievementID, userAchievement.AchievementID())
	assert.False(t, userAchievement.ObtainedAt().IsZero())
}

func TestCondition_AllValidEntities(t *testing.T) {
	validEntities := []string{"note", "link", "search", "share", "login"}

	for _, entity := range validEntities {
		condition := Condition{
			Type:      "count",
			Entity:    entity,
			Action:    "create",
			Threshold: 10,
		}

		err := condition.Validate()
		assert.NoError(t, err, "Entity %s should be valid", entity)
	}
}

func TestCondition_AllValidTypes(t *testing.T) {
	validTypes := []string{"count", "streak"}

	for _, typeValue := range validTypes {
		condition := Condition{
			Type:      typeValue,
			Entity:    "note",
			Action:    "create",
			Threshold: 10,
		}

		err := condition.Validate()
		assert.NoError(t, err, "Type %s should be valid", typeValue)
	}
}

func TestNewAchievement_WithFilter(t *testing.T) {
	condition := Condition{
		Type:      "count",
		Entity:    "note",
		Action:    "create",
		Filter:    map[string]interface{}{"type": "star"},
		Threshold: 10,
	}

	achievement, err := NewAchievement("star_notes", "Star Notes", "Create 10 star notes", "⭐", condition, 75, false)

	require.NoError(t, err)
	assert.Equal(t, condition.Filter, achievement.Condition().Filter)
}
