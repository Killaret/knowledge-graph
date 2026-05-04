package postgres

import (
	"time"

	"github.com/google/uuid"
)

// UserSettingModel — модель настроек пользователя
type UserSettingModel struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_user_settings_user_key,unique"`
	Key       string    `gorm:"type:varchar(100);not null;index:idx_user_settings_user_key,unique"`
	Value     []byte    `gorm:"type:jsonb;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (UserSettingModel) TableName() string {
	return "user_settings"
}
