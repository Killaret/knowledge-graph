package postgres

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// NoteModel — модель заметки с привязкой к создателю
type NoteModel struct {
	ID           uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title        string         `gorm:"not null"`
	Content      string         `gorm:"type:text"`
	Type         string         `gorm:"type:varchar(50);default:'star'"`
	Metadata     datatypes.JSON `gorm:"type:jsonb"`
	SearchVector string         `gorm:"column:search_vector;type:tsvector;->"` // read-only, updated by trigger
	CreatorID    *uuid.UUID     `gorm:"type:uuid;index"`
	Creator      *UserModel     `gorm:"foreignKey:CreatorID"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time `gorm:"index"`
}

func (NoteModel) TableName() string {
	return "notes"
}
