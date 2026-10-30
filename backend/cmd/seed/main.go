package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"knowledge-graph/internal/auth"
	"knowledge-graph/internal/config"
	"knowledge-graph/internal/infrastructure/db"
)

const testUserID = "00000000-0000-0000-0000-000000000000"

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	if !cfg.IsTest() {
		log.Fatalf("seeder can only run with APP_ENV=test, got %q", cfg.AppEnv)
	}

	password := os.Getenv("SEED_TEST_USER_PASSWORD")
	if password == "" {
		log.Fatalf("SEED_TEST_USER_PASSWORD is required to create the test user")
	}

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		log.Fatalf("failed to get sql db: %v", err)
	}
	defer sqlDB.Close()

	if err := runSeed(ctx, cfg, database, password); err != nil {
		log.Fatalf("failed to seed test user: %v", err)
	}

	fmt.Println("Test user seeded successfully")
}

func runSeed(ctx context.Context, cfg *config.Config, database *gorm.DB, password string) error {
	hash, err := auth.HashPassword(password, &auth.PasswordConfig{
		Time:    cfg.Argon2Time,
		Memory:  cfg.Argon2Memory,
		Threads: cfg.Argon2Threads,
		KeyLen:  32,
	})
	if err != nil {
		return fmt.Errorf("hash test user password: %w", err)
	}

	var roleIDStr string
	if err := database.WithContext(ctx).Raw("SELECT id FROM user_roles WHERE name = 'user' LIMIT 1").Scan(&roleIDStr).Error; err != nil {
		return fmt.Errorf("look up 'user' role: %w", err)
	}

	if roleIDStr == "" {
		return fmt.Errorf("'user' role not found in database; migrations must be applied before seeding")
	}

	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		return fmt.Errorf("parse 'user' role id: %w", err)
	}
	if roleID == uuid.Nil {
		return fmt.Errorf("'user' role id is nil; migrations must be applied before seeding")
	}

	testID, err := uuid.Parse(testUserID)
	if err != nil {
		return fmt.Errorf("invalid test user id: %w", err)
	}

	if err := database.WithContext(ctx).Exec(`
		INSERT INTO users (id, login, email, password_hash, role_id, created_at)
		VALUES (?, ?, ?, ?, ?, NOW())
		ON CONFLICT (id) DO UPDATE SET
			login = EXCLUDED.login,
			email = EXCLUDED.email,
			password_hash = EXCLUDED.password_hash,
			role_id = EXCLUDED.role_id,
			deleted_at = NULL
	`, testID, "testuser", "testuser@example.com", hash, roleID).Error; err != nil {
		return fmt.Errorf("upsert test user: %w", err)
	}

	return nil
}
