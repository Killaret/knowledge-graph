package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

// Note — модель заметки
type Note struct {
	ID        uuid.UUID              `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title     string                 `gorm:"not null"`
	Content   string                 `gorm:"type:text"`
	Metadata  map[string]interface{} `gorm:"type:jsonb"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Note) TableName() string {
	return "notes"
}

// Link — связь между заметками
type Link struct {
	ID           uuid.UUID              `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SourceNoteID uuid.UUID              `gorm:"type:uuid;not null;index"`
	TargetNoteID uuid.UUID              `gorm:"type:uuid;not null;index"`
	LinkType     string                 `gorm:"default:'reference'"`
	Weight       float64                `gorm:"default:1.0"`
	Metadata     map[string]interface{} `gorm:"type:jsonb"`
	CreatedAt    time.Time

	// Внешние ключи (GORM будет автоматически использовать)
	SourceNote Note `gorm:"foreignKey:SourceNoteID"`
	TargetNote Note `gorm:"foreignKey:TargetNoteID"`
}

func (Link) TableName() string {
	return "links"
}

// NoteKeyword — ключевые слова заметки (ядро идеи)
type NoteKeyword struct {
	NoteID  uuid.UUID `gorm:"type:uuid;primaryKey"`
	Keyword string    `gorm:"primaryKey"`
	Weight  float64

	Note Note `gorm:"foreignKey:NoteID"`
}

func (NoteKeyword) TableName() string {
	return "note_keywords"
}

// NoteEmbedding — векторное представление заметки
type NoteEmbedding struct {
	NoteID    uuid.UUID       `gorm:"type:uuid;primaryKey"`
	Embedding pgvector.Vector `gorm:"type:vector(384)"`
	UpdatedAt time.Time

	Note Note `gorm:"foreignKey:NoteID"`
}

func (NoteEmbedding) TableName() string {
	return "note_embeddings"
}

// NoteTag — теги заметки
type NoteTag struct {
	NoteID uuid.UUID `gorm:"type:uuid;primaryKey"`
	Tag    string    `gorm:"primaryKey"`

	Note Note `gorm:"foreignKey:NoteID"`
}

func (NoteTag) TableName() string {
	return "note_tags"
}

// User — пользователь (IAM)
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Login        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	Role         string    `gorm:"default:'user'"`
	CreatedAt    time.Time
}

func (User) TableName() string {
	return "users"
}

// NoteLike — лайк/дизлайк на заметку
type NoteLike struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	NoteID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	LikeType  string    `gorm:"not null"` // 'like' или 'dislike'
	CreatedAt time.Time

	User User `gorm:"foreignKey:UserID"`
	Note Note `gorm:"foreignKey:NoteID"`
}

func (NoteLike) TableName() string {
	return "note_likes"
}

// SuggestionFeedback — обратная связь по рекомендации
type SuggestionFeedback struct {
	UserID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	SourceNoteID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	SuggestedNoteID uuid.UUID `gorm:"type:uuid;primaryKey"`
	FeedbackType    string    `gorm:"not null"` // 'like' или 'dislike'
	CreatedAt       time.Time

	User          User `gorm:"foreignKey:UserID"`
	SourceNote    Note `gorm:"foreignKey:SourceNoteID"`
	SuggestedNote Note `gorm:"foreignKey:SuggestedNoteID"`
}

func (SuggestionFeedback) TableName() string {
	return "suggestion_feedback"
}

// ShareLink — расшаривание заметки
type ShareLink struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	NoteID         uuid.UUID `gorm:"type:uuid;not null;index"`
	SharedByUserID uuid.UUID `gorm:"type:uuid;not null"`
	ShareToken     string    `gorm:"uniqueIndex;not null"`
	ExpiresAt      *time.Time
	CreatedAt      time.Time

	Note     Note `gorm:"foreignKey:NoteID"`
	SharedBy User `gorm:"foreignKey:SharedByUserID"`
}

func (ShareLink) TableName() string {
	return "share_links"
}
