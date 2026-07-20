package achievement

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	userApp "knowledge-graph/internal/application/user"
	achievementDomain "knowledge-graph/internal/domain/achievement"
	dcache "knowledge-graph/internal/domain/cache"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// Service handles achievement operations
type Service struct {
	engine          achievementDomain.Engine
	achievementRepo achievementDomain.Repository
	settingsService *userApp.SettingsService
	taskQueue       interface{}
	cacheClient     dcache.CacheClient
}

// NewService creates a new achievement service
func NewService(
	engine achievementDomain.Engine,
	achievementRepo achievementDomain.Repository,
	settingsService *userApp.SettingsService,
	cacheClient dcache.CacheClient,
) *Service {
	return &Service{
		engine:          engine,
		achievementRepo: achievementRepo,
		settingsService: settingsService,
		cacheClient:     cacheClient,
	}
}

// CheckTrigger checks if a user should earn any achievements for a trigger
func (s *Service) CheckTrigger(ctx context.Context, userID uuid.UUID, triggerKey string) error {
	// Find all achievements that might be triggered by this action
	allAchievements, err := s.achievementRepo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load achievements: %w", err)
	}

	// Parse trigger key (e.g., "note.create" -> entity: note, action: create)
	triggerParts := parseTriggerKey(triggerKey)
	if triggerParts == nil {
		return fmt.Errorf("invalid trigger key: %s", triggerKey)
	}

	// Check each achievement
	for _, achievement := range allAchievements {
		condition := achievement.Condition()

		// Check if this achievement matches the trigger
		if !matchesTrigger(condition, triggerParts) {
			continue
		}

		// Check if user already has this achievement
		hasIt, err := s.achievementRepo.UserHasAchievement(ctx, userID, achievement.ID())
		if err != nil {
			log.Printf("[AchievementService] Failed to check user achievement: %v", err)
			continue
		}
		if hasIt {
			continue // Already earned
		}

		// Evaluate the condition
		fulfilled, err := s.engine.Evaluate(ctx, condition, userID)
		if err != nil {
			log.Printf("[AchievementService] Failed to evaluate condition: %v", err)
			continue
		}

		if fulfilled {
			// Grant the achievement
			ua := achievementDomain.NewUserAchievement(userID, achievement.ID())
			if err := s.achievementRepo.SaveUserAchievement(ctx, *ua); err != nil {
				log.Printf("[AchievementService] Failed to save user achievement: %v", err)
				continue
			}

			// Send notification if enabled
			if err := s.sendNotification(ctx, userID, achievement); err != nil {
				log.Printf("[AchievementService] Failed to send notification: %v", err)
			}
		}
	}

	return nil
}

// parseTriggerKey parses a trigger key like "note.create"
func parseTriggerKey(triggerKey string) map[string]string {
	// Simple parsing for now
	parts := make(map[string]string)
	switch triggerKey {
	case "note.create":
		parts["entity"] = "note"
		parts["action"] = "create"
	case "link.create":
		parts["entity"] = "link"
		parts["action"] = "create"
	case "login":
		parts["action"] = "login"
	case "search.execute":
		parts["entity"] = "search"
		parts["action"] = "execute"
	case "share.create":
		parts["entity"] = "share"
		parts["action"] = "create"
	default:
		return nil
	}
	return parts
}

// matchesTrigger checks if a condition matches a trigger
func matchesTrigger(condition achievementDomain.Condition, trigger map[string]string) bool {
	if condition.Type != "count" {
		return true // Streaks are handled separately
	}

	// Check entity match
	if entity, ok := trigger["entity"]; ok {
		if condition.Entity != entity {
			return false
		}
	}

	// Check action match
	if action, ok := trigger["action"]; ok {
		if condition.Action != action {
			return false
		}
	}

	return true
}

// sendNotification sends an achievement notification if enabled
func (s *Service) sendNotification(ctx context.Context, userID uuid.UUID, achievement achievementDomain.Achievement) error {
	// Check if notifications are enabled
	showNotifications := true
	if s.settingsService != nil {
		var err error
		showNotifications, err = s.settingsService.GetBool(ctx, userID, "show_achievement_notifications")
		if err != nil {
			// Default to true on error
			showNotifications = true
		}
	}

	if !showNotifications {
		return nil
	}

	// Create notification payload
	payload := map[string]interface{}{
		"user_id":          userID.String(),
		"achievement_id":   achievement.ID().String(),
		"achievement_code": achievement.Code(),
		"title":            achievement.Title(),
		"description":      achievement.Description(),
		"icon":             achievement.Icon(),
		"points":           achievement.Points(),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Queue notification task if task queue is available
	if s.taskQueue != nil {
		task := asynq.NewTask("notification:achievement", data)
		_, err := s.taskQueue.(*asynq.Client).Enqueue(task)
		if err != nil {
			return err
		}
	}

	// Also log for debugging
	log.Printf("[AchievementService] User %s earned achievement: %s (%s)",
		userID.String(), achievement.Code(), achievement.Title())

	return nil
}

// CheckStreaks checks streak-based achievements for all users
func (s *Service) CheckStreaks(ctx context.Context) error {
	// Load all streak-based achievements
	allAchievements, err := s.achievementRepo.FindAll(ctx)
	if err != nil {
		return err
	}

	for _, achievement := range allAchievements {
		if achievement.Condition().Type != "streak" {
			continue
		}

		// For streak achievements, we'd need to check all active users
		// This is a simplified version - in production, you'd batch this
		_ = achievement
	}

	return nil
}

// GetUserAchievements returns all achievements for a user
func (s *Service) GetUserAchievements(ctx context.Context, userID uuid.UUID) ([]achievementDomain.Achievement, error) {
	return s.achievementRepo.FindByUserID(ctx, userID)
}

type UserAchievementWithStatus struct {
	ID               string  `json:"id"`
	Code             string  `json:"code"`
	NameRu           *string `json:"name_ru,omitempty"`
	NameEn           *string `json:"name_en,omitempty"`
	DescriptionRu    *string `json:"description_ru,omitempty"`
	DescriptionEn    *string `json:"description_en,omitempty"`
	IconEmoji        *string `json:"icon_emoji,omitempty"`
	Category         *string `json:"category,omitempty"`
	Points           int     `json:"points"`
	UnlockedAt       string  `json:"unlocked_at"`
	NotificationSeen bool    `json:"notification_seen"`
}

// GetUserAchievementsWithStatus returns all user-earned achievements with unlock metadata.
func (s *Service) GetUserAchievementsWithStatus(ctx context.Context, userID uuid.UUID) ([]UserAchievementWithStatus, error) {
	results, err := s.achievementRepo.FindUserAchievementsWithStatus(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]UserAchievementWithStatus, 0, len(results))
	for _, r := range results {
		responses = append(responses, UserAchievementWithStatus{
			ID:               r.ID,
			Code:             r.Code,
			NameRu:           r.NameRu,
			NameEn:           r.NameEn,
			DescriptionRu:    r.DescriptionRu,
			DescriptionEn:    r.DescriptionEn,
			IconEmoji:        r.IconEmoji,
			Category:         r.Category,
			Points:           r.Points,
			UnlockedAt:       r.UnlockedAt.Format(time.RFC3339),
			NotificationSeen: r.NotificationSeen,
		})
	}

	return responses, nil
}

// GetAllAchievements returns all available achievements
func (s *Service) GetAllAchievements(ctx context.Context) ([]achievementDomain.Achievement, error) {
	return s.achievementRepo.FindAll(ctx)
}

// TrackLogin records a user login for streak tracking
func (s *Service) TrackLogin(ctx context.Context, userID uuid.UUID) error {
	if s.cacheClient == nil {
		return nil
	}

	key := fmt.Sprintf("login:streak:%s", userID.String())

	// Check if there's an existing streak
	exists, err := s.cacheClient.Exists(ctx, key)
	if err != nil {
		return err
	}

	if exists == 0 {
		// New streak
		_ = s.cacheClient.Set(ctx, key, "1", 48*time.Hour)
	} else {
		// Increment existing streak
		_, _ = s.cacheClient.Incr(ctx, key)
		// Extend TTL
		_ = s.cacheClient.Expire(ctx, key, 48*time.Hour)
	}

	// Check for streak-based achievements
	return s.CheckTrigger(ctx, userID, "login")
}

// GetStreak returns the current login streak for a user
func (s *Service) GetStreak(ctx context.Context, userID uuid.UUID) (int, error) {
	if s.cacheClient == nil {
		return 0, nil
	}

	key := fmt.Sprintf("login:streak:%s", userID.String())
	streakStr, err := s.cacheClient.Get(ctx, key)
	if err != nil {
		if err == dcache.ErrCacheMiss {
			return 0, nil
		}
		return 0, err
	}

	streak, err := strconv.Atoi(streakStr)
	if err != nil {
		return 0, err
	}

	return streak, nil
}

// MarkNotificationSeen marks the notification as read for a user's achievement.
func (s *Service) MarkNotificationSeen(ctx context.Context, userID uuid.UUID, achievementID string) error {
	if achievementID == "" {
		return nil
	}

	achievementUUID, err := uuid.Parse(achievementID)
	if err != nil {
		return err
	}

	return s.achievementRepo.MarkNotificationSeen(ctx, userID, achievementUUID)
}
