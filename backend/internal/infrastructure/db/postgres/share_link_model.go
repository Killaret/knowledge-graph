package postgres

import (
	"time"

	"github.com/google/uuid"
)

// ShareLinkModel — публичные ссылки для шаринга заметок
type ShareLinkModel struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	NoteID         uuid.UUID `gorm:"type:uuid;not null;index"`
	Note           NoteModel `gorm:"foreignKey:NoteID"`
	SharedByUserID uuid.UUID `gorm:"type:uuid;not null;index"`
	SharedBy       UserModel `gorm:"foreignKey:SharedByUserID"`
	Token          string    `gorm:"uniqueIndex;not null;index"`
	Permission     string    `gorm:"not null;default:'read'"` // 'read' или 'write'
	CreatedAt      time.Time
	ExpiresAt      *time.Time `gorm:"index"`
	MaxUses        *int
	UsesCount      int  `gorm:"default:0"`
	IsActive       bool `gorm:"default:true"`
}

func (ShareLinkModel) TableName() string { return "share_links" }
