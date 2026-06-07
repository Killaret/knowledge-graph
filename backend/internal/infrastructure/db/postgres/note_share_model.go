package postgres

import (
	"time"

	"github.com/google/uuid"
)

// NoteShareModel — direct sharing of notes between users
type NoteShareModel struct {
	ID               uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	NoteID           uuid.UUID `gorm:"type:uuid;not null;index"`
	Note             NoteModel `gorm:"foreignKey:NoteID"`
	SharedByUserID   uuid.UUID `gorm:"type:uuid;not null;index"`
	SharedByUser     UserModel `gorm:"foreignKey:SharedByUserID"`
	SharedWithUserID uuid.UUID `gorm:"type:uuid;not null;index"`
	SharedWithUser   UserModel `gorm:"foreignKey:SharedWithUserID"`
	Permission       string    `gorm:"not null;default:'read'"` // 'read' or 'write'
	CreatedAt        time.Time
	ExpiresAt        *time.Time `gorm:"index"`
}

func (NoteShareModel) TableName() string { return "note_shares" }
