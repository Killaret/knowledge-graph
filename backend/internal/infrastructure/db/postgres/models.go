package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"gorm.io/datatypes"
)

// NoteModel — модель заметки
type NoteModel struct {
	ID           uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title        string         `gorm:"not null"`
	Content      string         `gorm:"type:text"`
	Metadata     datatypes.JSON `gorm:"type:jsonb"`
	SearchVector string         `gorm:"column:search_vector;type:tsvector;->"` // read-only, updated by trigger
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (NoteModel) TableName() string {
	return "notes"
}

// LinkModel — связь между заметками
type LinkModel struct {
	ID           uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SourceNoteID uuid.UUID      `gorm:"type:uuid;not null;index"`
	TargetNoteID uuid.UUID      `gorm:"type:uuid;not null;index"`
	LinkType     string         `gorm:"default:'reference'"`
	Weight       float64        `gorm:"default:1.0"`
	Metadata     datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt    time.Time

	SourceNote NoteModel `gorm:"foreignKey:SourceNoteID"`
	TargetNote NoteModel `gorm:"foreignKey:TargetNoteID"`
}

func (LinkModel) TableName() string {
	return "links"
}

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

// NoteEmbeddingModel — векторное представление заметки
type NoteEmbeddingModel struct {
	NoteID    uuid.UUID       `gorm:"type:uuid;primaryKey"`
	Embedding pgvector.Vector `gorm:"type:vector(384)"`
	UpdatedAt time.Time

	Note NoteModel `gorm:"foreignKey:NoteID"`
}

func (NoteEmbeddingModel) TableName() string {
	return "note_embeddings"
}

// NoteTagModel — теги заметки
type NoteTagModel struct {
	NoteID uuid.UUID `gorm:"type:uuid;primaryKey"`
	Tag    string    `gorm:"primaryKey"`
	Note   NoteModel `gorm:"foreignKey:NoteID"`
}

func (NoteTagModel) TableName() string { return "note_tags" }

// UserModel — пользователь
type UserModel struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Login        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	Role         string    `gorm:"default:'user'"`
	CreatedAt    time.Time
}

func (UserModel) TableName() string { return "users" }

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

// ShareLinkModel — расшаривание
type ShareLinkModel struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	NoteID         uuid.UUID `gorm:"type:uuid;not null;index"`
	SharedByUserID uuid.UUID `gorm:"type:uuid;not null"`
	ShareToken     string    `gorm:"uniqueIndex;not null"`
	ExpiresAt      *time.Time
	CreatedAt      time.Time
	Note           NoteModel `gorm:"foreignKey:NoteID"`
	SharedBy       UserModel `gorm:"foreignKey:SharedByUserID"`
}

func (ShareLinkModel) TableName() string { return "share_links" }
