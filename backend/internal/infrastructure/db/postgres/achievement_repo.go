package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	achievementDomain "knowledge-graph/internal/domain/achievement"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AchievementRepository — репозиторий для работы с достижениями
type AchievementRepository struct {
	db *gorm.DB
}

// NewAchievementRepository создает новый репозиторий
func NewAchievementRepository(db *gorm.DB) *AchievementRepository {
	return &AchievementRepository{db: db}
}

// FindAll находит все достижения
func (r *AchievementRepository) FindAll(ctx context.Context) ([]achievementDomain.Achievement, error) {
	var models []AchievementModel
	err := r.db.WithContext(ctx).Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find achievements: %w", err)
	}

	achievements := make([]achievementDomain.Achievement, 0, len(models))
	for _, m := range models {
		achievement, err := toDomainAchievement(&m)
		if err != nil {
			continue // Skip corrupted data
		}
		achievements = append(achievements, *achievement)
	}

	return achievements, nil
}

// FindByCode находит достижение по коду
func (r *AchievementRepository) FindByCode(ctx context.Context, code string) (*achievementDomain.Achievement, error) {
	var model AchievementModel
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&model).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find achievement: %w", err)
	}

	return toDomainAchievement(&model)
}

// FindByUserID находит все достижения пользователя
func (r *AchievementRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]achievementDomain.Achievement, error) {
	var models []AchievementModel
	err := r.db.WithContext(ctx).
		Joins("JOIN user_achievements ON user_achievements.achievement_id = achievements.id").
		Where("user_achievements.user_id = ?", userID).
		Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find user achievements: %w", err)
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

// SaveUserAchievement сохраняет полученное достижение
func (r *AchievementRepository) SaveUserAchievement(ctx context.Context, ua achievementDomain.UserAchievement) error {
	model := UserAchievementModel{
		UserID:        ua.UserID(),
		AchievementID: ua.AchievementID(),
		ObtainedAt:    ua.ObtainedAt(),
		Metadata:      []byte(`{}`),
	}

	return r.db.WithContext(ctx).Create(&model).Error
}

// UserHasAchievement проверяет, есть ли у пользователя достижение
func (r *AchievementRepository) UserHasAchievement(ctx context.Context, userID uuid.UUID, achievementID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&UserAchievementModel{}).
		Where("user_id = ? AND achievement_id = ?", userID, achievementID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// toDomainAchievement преобразует модель в доменную сущность
func toDomainAchievement(m *AchievementModel) (*achievementDomain.Achievement, error) {
	return achievementDomain.ReconstructAchievement(
		m.ID,
		m.Code,
		m.Title,
		m.Description,
		m.Icon,
		json.RawMessage(m.Condition),
		m.Points,
		m.IsHidden,
		m.CreatedAt,
	)
}
