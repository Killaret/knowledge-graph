package postgres

import (
	"time"

	"github.com/google/uuid"
)

// UserModel — пользователь
type UserModel struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Login        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	Role         string    `gorm:"default:'user'"`
	CreatedAt    time.Time
}

func (UserModel) TableName() string { return "users" }
