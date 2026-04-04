package db

import (
	"log"
	"os"

	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"knowledge-graph/internal/infrastructure/db/postgres"
)

var DB *gorm.DB

func Init() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	var err error
	DB, err = gorm.Open(pgdriver.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	log.Println("Connected to PostgreSQL")

	// Run migrations
	if err := DB.AutoMigrate(
		&postgres.NoteModel{},
		&postgres.LinkModel{},
		&postgres.NoteKeywordModel{},
		&postgres.NoteEmbeddingModel{},
		&postgres.NoteTagModel{},
		&postgres.UserModel{},
		&postgres.NoteLikeModel{},
		&postgres.SuggestionFeedbackModel{},
		&postgres.ShareLinkModel{},
	); err != nil {
		log.Fatal("failed to run migrations:", err)
	}
	log.Println("Database migrations completed successfully")
}
