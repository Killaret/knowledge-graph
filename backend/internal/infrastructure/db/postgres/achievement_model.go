package postgres

import (
	"time"

	"github.com/google/uuid"
)

// AchievementModel — модель достижения
type AchievementModel struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Code        string    `gorm:"uniqueIndex;not null"`
	Title       string    `gorm:"not null"`
	Description string
	Icon        string
	Condition   []byte `gorm:"type:jsonb;not null"`
	Points      int    `gorm:"default:0"`
	IsHidden    bool   `gorm:"default:false"`
	CreatedAt   time.Time
}

func (AchievementModel) TableName() string {
	return "achievements"
}

// UserAchievementModel — модель полученного достижения
type UserAchievementModel struct {
	UserID        uuid.UUID `gorm:"type:uuid;not null;primaryKey"`
	AchievementID uuid.UUID `gorm:"type:uuid;not null;primaryKey"`
	ObtainedAt    time.Time `gorm:"not null;default:now()"`
	Metadata      []byte    `gorm:"type:jsonb;default:'{}'"`
}

func (UserAchievementModel) TableName() string {
	return "user_achievements"
}
