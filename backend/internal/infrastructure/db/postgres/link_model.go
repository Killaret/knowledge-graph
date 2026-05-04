package postgres

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// LinkModel — связь между заметками с привязкой к создателю
type LinkModel struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey"`
	SourceNoteID uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_links_source_target_type;column:source_note_id"`
	TargetNoteID uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_links_source_target_type;column:target_note_id"`
	LinkType     string         `gorm:"default:'reference';uniqueIndex:idx_links_source_target_type;column:link_type"`
	Weight       float64        `gorm:"default:1.0;column:weight"`
	Metadata     datatypes.JSON `gorm:"type:jsonb;column:metadata"`
	CreatorID    *uuid.UUID     `gorm:"type:uuid;index"`
	Creator      *UserModel     `gorm:"foreignKey:CreatorID"`
	CreatedAt    time.Time      `gorm:"column:created_at"`
	DeletedAt    *time.Time     `gorm:"column:deleted_at;index"`

	SourceNote NoteModel `gorm:"foreignKey:SourceNoteID;references:ID"`
	TargetNote NoteModel `gorm:"foreignKey:TargetNoteID;references:ID"`
}

func (LinkModel) TableName() string {
	return "links"
}
