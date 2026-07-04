// Package testutil provides utilities for integration tests
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

// SetupTestDB spins up a PostgreSQL container and returns a GORM connection
func SetupTestDB(t *testing.T) (*gorm.DB, func()) {
	ctx := context.Background()

	// Start the PostgreSQL container
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

	// Get the connection string
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// Connect via GORM
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silent mode for tests
	})
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	// Cleanup function
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

// TruncateTables truncates all tables (use in SetupTest)
func TruncateTables(db *gorm.DB) error {
	// Get the list of base tables (excluding note_embeddings and recommendations)
	tables := []string{"notes", "links", "note_keywords", "users", "tags", "note_tags"}

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)).Error; err != nil {
			return fmt.Errorf("failed to truncate %s: %w", table, err)
		}
	}

	return nil
}

// MigrateModels applies migrations for the models
func MigrateModels(db *gorm.DB, models ...interface{}) error {
	return db.AutoMigrate(models...)
}
