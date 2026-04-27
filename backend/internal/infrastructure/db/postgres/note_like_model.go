package postgres

import (
	"time"

	"github.com/google/uuid"
)

// NoteLikeModel — лайк/дизлайк
type NoteLikeModel struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	NoteID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	LikeType  string    `gorm:"not null"`
	CreatedAt time.Time
	User      UserModel `gorm:"foreignKey:UserID"`
	Note      NoteModel `gorm:"foreignKey:NoteID"`
}

func (NoteLikeModel) TableName() string { return "note_likes" }
