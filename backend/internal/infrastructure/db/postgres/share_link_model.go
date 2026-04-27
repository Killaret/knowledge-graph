package postgres

import (
	"time"

	"github.com/google/uuid"
)

// ShareLinkModel — расшаривание
type ShareLinkModel struct {
	ID             uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	NoteID         uuid.UUID  `gorm:"type:uuid;not null;index"`
	SharedByUserID uuid.UUID  `gorm:"type:uuid;not null"`
	ShareToken     string     `gorm:"uniqueIndex;not null"`
	ExpiresAt      *time.Time
	CreatedAt      time.Time
	Note           NoteModel  `gorm:"foreignKey:NoteID"`
	SharedBy       UserModel  `gorm:"foreignKey:SharedByUserID"`
}

func (ShareLinkModel) TableName() string { return "share_links" }
