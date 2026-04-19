package postgres

import (
	"time"

	"github.com/google/uuid"
)

// RecommendationModel represents a precomputed note recommendation in the database
type RecommendationModel struct {
	NoteID            uuid.UUID `gorm:"primaryKey"`
	RecommendedNoteID uuid.UUID `gorm:"primaryKey"`
	Score             float64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// TableName returns the table name for the recommendation model
func (RecommendationModel) TableName() string {
	return "note_recommendations"
}
