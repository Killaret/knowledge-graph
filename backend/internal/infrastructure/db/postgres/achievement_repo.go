package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	achievementDomain "knowledge-graph/internal/domain/achievement"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AchievementRepository struct {
	db *gorm.DB
}

type UserAchievementWithStatus struct {
	AchievementModel
	UnlockedAt       time.Time `gorm:"column:unlocked_at"`
	NotificationSeen bool      `gorm:"column:notification_seen"`
}

func NewAchievementRepository(db *gorm.DB) *AchievementRepository {
	return &AchievementRepository{db: db}
}

func (r *AchievementRepository) FindAll(ctx context.Context) ([]achievementDomain.Achievement, error) {
	var models []AchievementModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to load achievements: %w", err)
	}

	achievements := make([]achievementDomain.Achievement, 0, len(models))
	for _, m := range models {
		achievement, err := toDomainAchievement(&m)
		if err != nil {
			continue
		}
		achievements = append(achievements, *achievement)
	}

	return achievements, nil
}

func (r *AchievementRepository) FindByCode(ctx context.Context, code string) (*achievementDomain.Achievement, error) {
	var model AchievementModel
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load achievement by code: %w", err)
	}

	return toDomainAchievement(&model)
}

func (r *AchievementRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]achievementDomain.Achievement, error) {
	var models []AchievementModel
	if err := r.db.WithContext(ctx).
		Joins("JOIN user_achievements ON user_achievements.achievement_id = achievements.id").
		Where("user_achievements.user_id = ?", userID).
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to load user achievements: %w", err)
	}

	achievements := make([]achievementDomain.Achievement, 0, len(models))
	for _, m := range models {
		achievement, err := toDomainAchievement(&m)
		if err != nil {
			continue
		}
		achievements = append(achievements, *achievement)
	}

	return achievements, nil
}

func (r *AchievementRepository) FindUserAchievementsWithStatus(ctx context.Context, userID uuid.UUID) ([]achievementDomain.UserAchievementWithStatus, error) {
	var models []UserAchievementWithStatus
	if err := r.db.WithContext(ctx).
		Table("achievements").
		Select("achievements.*, user_achievements.unlocked_at, user_achievements.notification_seen").
		Joins("JOIN user_achievements ON user_achievements.achievement_id = achievements.id").
		Where("user_achievements.user_id = ?", userID).
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to load user achievements with status: %w", err)
	}

	results := make([]achievementDomain.UserAchievementWithStatus, 0, len(models))
	for _, m := range models {
		results = append(results, achievementDomain.UserAchievementWithStatus{
			ID:               m.ID.String(),
			Code:             m.Code,
			NameRu:           m.NameRu,
			NameEn:           m.NameEn,
			DescriptionRu:    m.DescriptionRu,
			DescriptionEn:    m.DescriptionEn,
			IconEmoji:        m.IconEmoji,
			Category:         m.Category,
			Points:           m.Points,
			UnlockedAt:       m.UnlockedAt,
			NotificationSeen: m.NotificationSeen,
		})
	}

	return results, nil
}

func (r *AchievementRepository) SaveUserAchievement(ctx context.Context, ua achievementDomain.UserAchievement) error {
	model := UserAchievementModel{
		UserID:           ua.UserID(),
		AchievementID:    ua.AchievementID(),
		UnlockedAt:       ua.ObtainedAt(),
		NotificationSeen: false,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *AchievementRepository) UserHasAchievement(ctx context.Context, userID uuid.UUID, achievementID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&UserAchievementModel{}).
		Where("user_id = ? AND achievement_id = ?", userID, achievementID).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *AchievementRepository) MarkNotificationSeen(ctx context.Context, userID uuid.UUID, achievementID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&UserAchievementModel{}).
		Where("user_id = ? AND achievement_id = ?", userID, achievementID).
		Updates(map[string]interface{}{"notification_seen": true, "updated_at": time.Now().UTC()}).Error
}

func toDomainAchievement(m *AchievementModel) (*achievementDomain.Achievement, error) {
	title := ""
	if m.NameEn != nil && *m.NameEn != "" {
		title = *m.NameEn
	} else if m.NameRu != nil {
		title = *m.NameRu
	}

	description := ""
	if m.DescriptionEn != nil && *m.DescriptionEn != "" {
		description = *m.DescriptionEn
	} else if m.DescriptionRu != nil {
		description = *m.DescriptionRu
	}

	icon := ""
	if m.IconEmoji != nil {
		icon = *m.IconEmoji
	}

	return achievementDomain.ReconstructAchievement(
		m.ID,
		m.Code,
		title,
		description,
		icon,
		json.RawMessage(m.ConditionJSON),
		0,
		m.Hidden,
		m.CreatedAt,
	)
}
