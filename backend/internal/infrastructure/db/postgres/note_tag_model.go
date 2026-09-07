package postgres

import (
	"github.com/google/uuid"
)

// NoteTagModel — связь заметки с тегом (многие-ко-многим)
type NoteTagModel struct {
	NoteID uuid.UUID `gorm:"type:uuid;primaryKey"`
	TagID  uuid.UUID `gorm:"type:uuid;primaryKey"`
	Note   NoteModel `gorm:"foreignKey:NoteID"`
	Tag    TagModel  `gorm:"foreignKey:TagID"`
}

func (NoteTagModel) TableName() string { return "note_tags" }
