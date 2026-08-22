package postgres

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AchievementModel maps the achievements table.
type AchievementModel struct {
	ID            uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Code          string          `gorm:"uniqueIndex;not null"`
	NameRu        *string         `gorm:"column:name_ru"`
	NameEn        *string         `gorm:"column:name_en"`
	DescriptionRu *string         `gorm:"column:description_ru"`
	DescriptionEn *string         `gorm:"column:description_en"`
	IconEmoji     *string         `gorm:"column:icon_emoji"`
	Category      *string         `gorm:"column:category"`
	Points        int             `gorm:"column:points;default:0"`
	ConditionJSON json.RawMessage `gorm:"column:condition_json;type:jsonb;not null;default:'{}'"`
	Hidden        bool            `gorm:"column:hidden;default:false"`
	CreatedAt     time.Time       `gorm:"column:created_at"`
	UpdatedAt     time.Time       `gorm:"column:updated_at"`
}

func (AchievementModel) TableName() string {
	return "achievements"
}

// UserAchievementModel maps the user_achievements table.
type UserAchievementModel struct {
	ID               uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID           uuid.UUID `gorm:"type:uuid;not null;index"`
	AchievementID    uuid.UUID `gorm:"type:uuid;not null;index"`
	UnlockedAt       time.Time `gorm:"column:unlocked_at;not null;default:now()"`
	NotificationSeen bool      `gorm:"column:notification_seen;not null;default:false"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

func (UserAchievementModel) TableName() string {
	return "user_achievements"
}
