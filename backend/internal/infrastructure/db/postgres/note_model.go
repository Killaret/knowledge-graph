package postgres

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// NoteModel — note model with creator binding
type NoteModel struct {
	ID           uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title        string         `gorm:"not null"`
	Content      string         `gorm:"type:text"`
	Type         string         `gorm:"type:varchar(50);default:'star'"`
	Metadata     datatypes.JSON `gorm:"type:jsonb"`
	SearchVector string         `gorm:"column:search_vector;type:tsvector;->"` // read-only, updated by trigger
	CreatorID    *uuid.UUID     `gorm:"type:uuid;index"`
	Creator      *UserModel     `gorm:"foreignKey:CreatorID"`
	IsPublic     bool           `gorm:"column:is_public;default:false;index"`
	// COMET-1: scheduling fields — real columns, not metadata, so the upcoming
	// list can index due_at (migration 038).
	DueAt               *time.Time `gorm:"column:due_at"`
	RemindBeforeSeconds *int64     `gorm:"column:remind_before_seconds"`
	DoneAt              *time.Time `gorm:"column:done_at"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           gorm.DeletedAt `gorm:"index"`
}

func (NoteModel) TableName() string {
	return "notes"
}
