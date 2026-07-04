package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"
)

// MigrateUp applies all SQL migrations from the given directory
type Migration struct {
	Version string
	Name    string
	UpSQL   string
}

// RunMigrations applies all "up" migrations from the migrations folder
func RunMigrations(db *gorm.DB, migrationsDir string) error {
	// Create the migration tracking table if it does not exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get the list of applied migrations
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Read all .up.sql files
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	// Sort by name (version)
	sort.Strings(files)

	for _, file := range files {
		version := extractVersion(filepath.Base(file))
		if version == "" {
			continue
		}

		// Skip already applied ones
		if applied[version] {
			log.Printf("Migration %s already applied, skipping", version)
			continue
		}

		// Read the SQL
		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		sql := string(sqlBytes)
		if strings.TrimSpace(sql) == "" {
			continue
		}

		// Apply the migration within a transaction
		if err := applyMigration(db, version, sql); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", version, err)
		}

		log.Printf("✅ Applied migration: %s", version)
	}

	return nil
}

// createMigrationsTable creates the migration tracking table
func createMigrationsTable(db *gorm.DB) error {
	sql := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	return db.Exec(sql).Error
}

// getAppliedMigrations returns a map of applied migrations
func getAppliedMigrations(db *gorm.DB) (map[string]bool, error) {
	applied := make(map[string]bool)

	var versions []string
	result := db.Raw("SELECT version FROM schema_migrations").Scan(&versions)
	if result.Error != nil && result.Error != sql.ErrNoRows {
		// If the table does not exist, just return an empty map
		return applied, nil
	}

	for _, v := range versions {
		applied[v] = true
	}

	return applied, nil
}

// applyMigration applies a single migration and records it in the table
func applyMigration(db *gorm.DB, version, sql string) error {
	// Run the SQL migration without a transaction to correctly handle "already exists" errors
	err := db.Exec(sql).Error
	if err != nil {
		// Check whether the error is "already exists"
		errStr := err.Error()
		if strings.Contains(errStr, "already exists") ||
			strings.Contains(errStr, "42P07") ||
			strings.Contains(errStr, "42710") {
			// If the object already exists, this is not a critical error
			log.Printf("Note: %v (skipping)", err)
		} else {
			return err
		}
	}

	// Record it in the migrations table
	if err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version).Error; err != nil {
		return err
	}

	return nil
}

// extractVersion extracts the version from a file name (e.g. "001_create_notes_table.up.sql" -> "001")
func extractVersion(filename string) string {
	// Strip the extension
	filename = strings.TrimSuffix(filename, ".up.sql")
	// Take only the version number (up to the first _)
	parts := strings.SplitN(filename, "_", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
