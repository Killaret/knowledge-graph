package postgres

import (
	"github.com/google/uuid"
)

// NoteKeywordModel — ключевые слова заметки
type NoteKeywordModel struct {
	NoteID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	Keyword   string    `gorm:"primaryKey"` // canonical lemma
	Surface   string    `gorm:"default:''"` // surface form as written in the text
	Extractor string    // e.g. "keybert-hybrid-0.9", "yake-0.4.8" for legacy rows
	Weight    float64

	Note NoteModel `gorm:"foreignKey:NoteID"`
}

func (NoteKeywordModel) TableName() string {
	return "note_keywords"
}
