// Package testutil предоставляет утилиты для интеграционных тестов
package testutil

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB поднимает контейнер PostgreSQL и возвращает подключение GORM
func SetupTestDB(t *testing.T) (*gorm.DB, func()) {
	ctx := context.Background()

	// Запускаем контейнер PostgreSQL
	pgContainer, err := pgcontainer.Run(ctx, "postgres:15-alpine",
		pgcontainer.WithDatabase("testdb"),
		pgcontainer.WithUsername("test"),
		pgcontainer.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	// Получаем строку подключения
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// Подключаемся через GORM
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Тихий режим для тестов
	})
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	// Функция очистки
	cleanup := func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
		if err := pgContainer.Terminate(ctx); err != nil {
			log.Printf("failed to terminate container: %v", err)
		}
	}

	return db, cleanup
}

// TruncateTables очищает все таблицы (использовать в SetupTest)
func TruncateTables(db *gorm.DB) error {
	// Получаем список базовых таблиц (без note_embeddings и recommendations)
	tables := []string{"notes", "links", "note_keywords", "users", "tags", "note_tags"}

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)).Error; err != nil {
			return fmt.Errorf("failed to truncate %s: %w", table, err)
		}
	}

	return nil
}

// MigrateModels применяет миграции для моделей
func MigrateModels(db *gorm.DB, models ...interface{}) error {
	return db.AutoMigrate(models...)
}
