package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"knowledge-graph/internal/auth"
	"knowledge-graph/internal/infrastructure/db"
	"knowledge-graph/internal/infrastructure/db/postgres"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// rotatedKey holds the metadata for a newly created API key, including the
// one-time token that must be delivered to the user.
type rotatedKey struct {
	UserID    uuid.UUID `json:"user_id"`
	KeyName   string    `json:"name"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	dryRun := flag.Bool("dry-run", false, "Show what would be rotated without writing to the database")
	out := flag.String("out", "api-keys-rotated-"+time.Now().UTC().Format("20060102-150405")+".json", "File to write new tokens to")
	dsn := flag.String("dsn", os.Getenv("DATABASE_URL"), "PostgreSQL DSN (defaults to DATABASE_URL)")
	flag.Parse()

	if *dsn == "" {
		log.Fatalf("DATABASE_URL environment variable or -dsn flag is required")
	}

	database, err := db.Connect(*dsn)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	ctx := context.Background()

	var keys []postgres.APIKeyModel
	if err := database.WithContext(ctx).Where("is_active = ?", true).Find(&keys).Error; err != nil {
		log.Fatalf("Failed to fetch active API keys: %v", err)
	}

	if len(keys) == 0 {
		log.Println("No active API keys found. Nothing to rotate.")
		return
	}

	log.Printf("Found %d active API key(s)", len(keys))

	if *dryRun {
		for _, k := range keys {
			log.Printf("Would rotate: id=%s user_id=%s name=%q", k.ID, k.UserID, k.Name)
		}
		return
	}

	var newKeys []rotatedKey
	passwordConfig := auth.DefaultPasswordConfig()

	err = database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, old := range keys {
			keyID := uuid.New()
			secret, err := auth.GenerateRandomToken(32)
			if err != nil {
				return fmt.Errorf("generate secret for user %s: %w", old.UserID, err)
			}

			keyHash, err := auth.HashPassword(secret, passwordConfig)
			if err != nil {
				return fmt.Errorf("hash secret for user %s: %w", old.UserID, err)
			}

			newName := old.Name
			if newName == "" {
				newName = "rotated"
			} else {
				newName = newName + " (rotated)"
			}

			newKey := postgres.APIKeyModel{
				ID:        keyID,
				UserID:    old.UserID,
				KeyHash:   keyHash,
				Name:      newName,
				Scopes:    old.Scopes,
				CreatedAt: time.Now(),
				ExpiresAt: old.ExpiresAt,
				IsActive:  true,
			}
			if err := tx.Create(&newKey).Error; err != nil {
				return fmt.Errorf("create new key for user %s: %w", old.UserID, err)
			}

			if err := tx.Model(&old).Update("is_active", false).Error; err != nil {
				return fmt.Errorf("revoke old key %s: %w", old.ID, err)
			}

			token := keyID.String() + ":" + secret
			newKeys = append(newKeys, rotatedKey{
				UserID:    old.UserID,
				KeyName:   newKey.Name,
				Token:     token,
				CreatedAt: newKey.CreatedAt,
			})
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Rotation failed: %v", err)
	}

	file, err := os.OpenFile(*out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Failed to open output file %q: %v", *out, err)
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(newKeys); err != nil {
		log.Fatalf("Failed to write new tokens to %q: %v", *out, err)
	}

	log.Printf("Rotated %d API key(s). New tokens written to %s", len(newKeys), *out)
	log.Println("WARNING: the output file contains secrets. Keep it secure and delete it after updating your clients.")
}
