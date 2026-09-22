package postgres

import (
	"time"

	"github.com/google/uuid"
)

// LinkSuppressionModel — запись отказа «эти заметки не связаны».
// Пара нормализована (note_a_id < note_b_id); link_type NULL означает
// «никакой связи между ними». Удаление заметки уносит её отказы каскадом.
type LinkSuppressionModel struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	NoteAID   uuid.UUID  `gorm:"type:uuid;not null;column:note_a_id"`
	NoteBID   uuid.UUID  `gorm:"type:uuid;not null;column:note_b_id"`
	LinkType  *string    `gorm:"type:text;column:link_type"`
	CreatorID *uuid.UUID `gorm:"type:uuid;column:creator_id"`
	CreatedAt time.Time  `gorm:"column:created_at"`

	NoteA   NoteModel  `gorm:"foreignKey:NoteAID;references:ID;constraint:OnDelete:CASCADE"`
	NoteB   NoteModel  `gorm:"foreignKey:NoteBID;references:ID;constraint:OnDelete:CASCADE"`
	Creator *UserModel `gorm:"foreignKey:CreatorID;references:ID"`
}

func (LinkSuppressionModel) TableName() string {
	return "link_suppressions"
}
