package postgres

import (
	"time"

	"github.com/google/uuid"
)

// TagModel — тег
type TagModel struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string    `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time
}

func (TagModel) TableName() string { return "tags" }
