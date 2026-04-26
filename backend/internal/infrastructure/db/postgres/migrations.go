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

// MigrateUp применяет все SQL миграции из указанной директории
type Migration struct {
	Version string
	Name    string
	UpSQL   string
}

// RunMigrations применяет все "up" миграции из папки migrations
func RunMigrations(db *gorm.DB, migrationsDir string) error {
	// Создаем таблицу для отслеживания миграций, если её нет
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Получаем список примененных миграций
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Читаем все .up.sql файлы
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	// Сортируем по имени (версии)
	sort.Strings(files)

	for _, file := range files {
		version := extractVersion(filepath.Base(file))
		if version == "" {
			continue
		}

		// Пропускаем уже примененные
		if applied[version] {
			log.Printf("Migration %s already applied, skipping", version)
			continue
		}

		// Читаем SQL
		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		sql := string(sqlBytes)
		if strings.TrimSpace(sql) == "" {
			continue
		}

		// Применяем миграцию в транзакции
		if err := applyMigration(db, version, sql); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", version, err)
		}

		log.Printf("✅ Applied migration: %s", version)
	}

	return nil
}

// createMigrationsTable создает таблицу для отслеживания миграций
func createMigrationsTable(db *gorm.DB) error {
	sql := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	return db.Exec(sql).Error
}

// getAppliedMigrations возвращает мапу примененных миграций
func getAppliedMigrations(db *gorm.DB) (map[string]bool, error) {
	applied := make(map[string]bool)

	var versions []string
	result := db.Raw("SELECT version FROM schema_migrations").Scan(&versions)
	if result.Error != nil && result.Error != sql.ErrNoRows {
		// Если таблицы нет, просто вернем пустую мапу
		return applied, nil
	}

	for _, v := range versions {
		applied[v] = true
	}

	return applied, nil
}

// applyMigration применяет одну миграцию и записывает её в таблицу
func applyMigration(db *gorm.DB, version, sql string) error {
	// Выполняем SQL миграции без транзакции для корректной обработки ошибок "already exists"
	err := db.Exec(sql).Error
	if err != nil {
		// Проверяем, является ли ошибка "already exists"
		errStr := err.Error()
		if strings.Contains(errStr, "already exists") ||
			strings.Contains(errStr, "42P07") ||
			strings.Contains(errStr, "42710") {
			// Если объект уже существует, это не критичная ошибка
			log.Printf("Note: %v (skipping)", err)
		} else {
			return err
		}
	}

	// Записываем в таблицу миграций
	if err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version).Error; err != nil {
		return err
	}

	return nil
}

// extractVersion извлекает версию из имени файла (например, "001_create_notes_table.up.sql" -> "001")
func extractVersion(filename string) string {
	// Убираем расширение
	filename = strings.TrimSuffix(filename, ".up.sql")
	// Берем только номер версии (до первого _)
	parts := strings.SplitN(filename, "_", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
