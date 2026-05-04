package postgres

import (
	"context"
	"fmt"

	userDomain "knowledge-graph/internal/domain/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserSettingsRepository — репозиторий для работы с настройками пользователя
type UserSettingsRepository struct {
	db *gorm.DB
}

// NewUserSettingsRepository создает новый репозиторий
func NewUserSettingsRepository(db *gorm.DB) *UserSettingsRepository {
	return &UserSettingsRepository{db: db}
}

// FindByUserID находит все настройки пользователя
func (r *UserSettingsRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]userDomain.UserSetting, error) {
	var models []UserSettingModel
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find settings: %w", err)
	}

	settings := make([]userDomain.UserSetting, 0, len(models))
	for _, m := range models {
		setting, err := userDomain.ReconstructUserSetting(
			m.ID, m.UserID, m.Key, m.Value, m.CreatedAt, m.UpdatedAt,
		)
		if err != nil {
			continue // Skip corrupted data
		}
		settings = append(settings, *setting)
	}

	return settings, nil
}

// FindByUserIDAndKey находит настройку по пользователю и ключу
func (r *UserSettingsRepository) FindByUserIDAndKey(ctx context.Context, userID uuid.UUID, key userDomain.SettingKey) (*userDomain.UserSetting, error) {
	var model UserSettingModel
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND key = ?", userID, key.String()).
		First(&model).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find setting: %w", err)
	}

	return userDomain.ReconstructUserSetting(
		model.ID, model.UserID, model.Key, model.Value, model.CreatedAt, model.UpdatedAt,
	)
}

// Upsert создает или обновляет настройку
func (r *UserSettingsRepository) Upsert(ctx context.Context, setting userDomain.UserSetting) error {
	model := UserSettingModel{
		ID:        setting.ID(),
		UserID:    setting.UserID(),
		Key:       setting.Key().String(),
		Value:     setting.Value(),
		CreatedAt: setting.CreatedAt(),
		UpdatedAt: setting.UpdatedAt(),
	}

	return r.db.WithContext(ctx).Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
		},
	).Create(&model).Error
}

// Delete удаляет настройку
func (r *UserSettingsRepository) Delete(ctx context.Context, userID uuid.UUID, key userDomain.SettingKey) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND key = ?", userID, key.String()).
		Delete(&UserSettingModel{}).Error
}
