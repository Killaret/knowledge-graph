package postgres

import (
	"time"

	"github.com/google/uuid"
)

// SuggestionFeedbackModel — обратная связь
type SuggestionFeedbackModel struct {
	UserID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	SourceNoteID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	SuggestedNoteID uuid.UUID `gorm:"type:uuid;primaryKey"`
	FeedbackType    string    `gorm:"not null"`
	CreatedAt       time.Time
	User            UserModel `gorm:"foreignKey:UserID"`
	SourceNote      NoteModel `gorm:"foreignKey:SourceNoteID"`
	SuggestedNote   NoteModel `gorm:"foreignKey:SuggestedNoteID"`
}

func (SuggestionFeedbackModel) TableName() string { return "suggestion_feedback" }
