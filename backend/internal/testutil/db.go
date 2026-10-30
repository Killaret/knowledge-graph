// Package testutil предоставляет утилиты для интеграционных тестов
package testutil

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
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
	return setupTestDB(t, "postgres:15-alpine", false)
}

// SetupTestVectorDB поднимает контейнер PostgreSQL с pgvector и возвращает подключение GORM
func SetupTestVectorDB(t *testing.T) (*gorm.DB, func()) {
	return setupTestDB(t, "pgvector/pgvector:pg16", true)
}

func setupTestDB(t *testing.T, image string, needsVector bool) (*gorm.DB, func()) {
	// В CI используем общий сервис-контейнер и создаём изолированную БД на тест.
	// Это позволяет избежать OOM от десятков testcontainers.
	if connStr := os.Getenv("TEST_DATABASE_URL"); connStr != "" {
		return setupTestDBFromEnv(t, connStr, needsVector)
	}
	if connStr := os.Getenv("DATABASE_URL"); connStr != "" {
		return setupTestDBFromEnv(t, connStr, needsVector)
	}
	return setupTestDBWithContainer(t, image, needsVector)
}

func setupTestDBFromEnv(t *testing.T, baseConnStr string, needsVector bool) (*gorm.DB, func()) {
	ctx := context.Background()

	u, err := url.Parse(baseConnStr)
	if err != nil {
		t.Fatalf("failed to parse database URL: %v", err)
	}

	dbName := uniqueDBName(t.Name())

	// Подключаемся к существующей БД (той же, что в URL) для CREATE/DROP DATABASE.
	adminURL := replaceDBName(u, strings.TrimPrefix(u.Path, "/"))
	adminDB, err := openGORM(adminURL.String())
	if err != nil {
		t.Fatalf("failed to connect to admin database: %v", err)
	}

	if err := adminDB.Exec("CREATE DATABASE " + dbName).Error; err != nil {
		_ = closeDB(adminDB)
		t.Fatalf("failed to create test database %s: %v", dbName, err)
	}

	testURL := replaceDBName(u, dbName)
	db, err := openGORM(testURL.String())
	if err != nil {
		_ = dropDatabase(adminDB, dbName)
		_ = closeDB(adminDB)
		t.Fatalf("failed to connect to test database %s: %v", dbName, err)
	}

	if err := createExtensions(db, needsVector); err != nil {
		_ = closeDB(db)
		_ = dropDatabase(adminDB, dbName)
		_ = closeDB(adminDB)
		t.Fatalf("failed to create extensions: %v", err)
	}

	var cleanupOnce sync.Once
	cleanup := func() {
		cleanupOnce.Do(func() {
			if err := closeDB(db); err != nil {
				log.Printf("failed to close test db connection: %v", err)
			}
			if err := dropDatabase(adminDB, dbName); err != nil {
				log.Printf("failed to drop test database %s: %v", dbName, err)
			}
			if err := closeDB(adminDB); err != nil {
				log.Printf("failed to close admin db connection: %v", err)
			}
		})
	}

	// Автоматически очищаем, если тест забыл вызвать cleanup.
	t.Cleanup(cleanup)

	_ = ctx
	return db, cleanup
}

func setupTestDBWithContainer(t *testing.T, image string, needsVector bool) (*gorm.DB, func()) {
	ctx := context.Background()

	pgContainer, err := pgcontainer.Run(ctx, image,
		pgcontainer.WithDatabase("testdb"),
		pgcontainer.WithUsername("test"),
		pgcontainer.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(120*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	db, err := openGORM(connStr)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	if err := createExtensions(db, needsVector); err != nil {
		t.Fatalf("failed to create extensions: %v", err)
	}

	var cleanupOnce sync.Once
	cleanup := func() {
		cleanupOnce.Do(func() {
			if err := closeDB(db); err != nil {
				log.Printf("failed to close db connection: %v", err)
			}
			if err := pgContainer.Terminate(ctx); err != nil {
				log.Printf("failed to terminate container: %v", err)
			}
		})
	}

	t.Cleanup(cleanup)
	return db, cleanup
}

func openGORM(connStr string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}

func closeDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func replaceDBName(u *url.URL, dbName string) *url.URL {
	newURL := *u
	newURL.Path = "/" + dbName
	return &newURL
}

var dbNameSanitizer = regexp.MustCompile(`[^a-z0-9_]+`)

func uniqueDBName(testName string) string {
	sanitized := strings.ToLower(testName)
	sanitized = dbNameSanitizer.ReplaceAllString(sanitized, "_")
	sanitized = strings.Trim(sanitized, "_")
	if sanitized == "" {
		sanitized = "test"
	}
	if len(sanitized) > 20 {
		sanitized = sanitized[:20]
	}
	randomSuffix, _ := rand.Int(rand.Reader, big.NewInt(10000))
	return fmt.Sprintf("test_%s_%d_%04d", sanitized, time.Now().UnixNano(), randomSuffix.Int64())
}

func createExtensions(db *gorm.DB, needsVector bool) error {
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto").Error; err != nil {
		return fmt.Errorf("create pgcrypto: %w", err)
	}
	if needsVector {
		if err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
			return fmt.Errorf("create vector: %w", err)
		}
	}
	return nil
}

func dropDatabase(adminDB *gorm.DB, dbName string) error {
	return adminDB.Exec("DROP DATABASE IF EXISTS " + dbName + " WITH (FORCE)").Error
}

// TruncateTables очищает все базовые таблицы (использовать в SetupTest)
func TruncateTables(db *gorm.DB) error {
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
