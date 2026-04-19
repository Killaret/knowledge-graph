package db

import (
	"log"
	"os"

	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	log.Println("Note: Database migrations should be applied via SQL files in migrations/ directory")
}
