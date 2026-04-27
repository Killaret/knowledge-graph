package postgres

import (
	"github.com/google/uuid"
)

// NoteKeywordModel — ключевые слова заметки
type NoteKeywordModel struct {
	NoteID  uuid.UUID `gorm:"type:uuid;primaryKey"`
	Keyword string    `gorm:"primaryKey"`
	Weight  float64

	Note NoteModel `gorm:"foreignKey:NoteID"`
}

func (NoteKeywordModel) TableName() string {
	return "note_keywords"
}
