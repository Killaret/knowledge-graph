package db

import (
	"fmt"
	"time"

	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect creates a PostgreSQL connection and returns *gorm.DB.
// DSN is passed explicitly, not read from env.
func Connect(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("database DSN is empty")
	}
	database, err := gorm.Open(pgdriver.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := database.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Connection pool settings for main backend
	sqlDB.SetMaxOpenConns(25)                 // Maximum number of open connections
	sqlDB.SetMaxIdleConns(5)                  // Maximum number of idle connections
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // Maximum lifetime of a connection
	sqlDB.SetConnMaxIdleTime(1 * time.Minute) // Maximum idle time of a connection

	return database, nil
}

// GetPoolStats returns connection pool statistics for monitoring
func GetPoolStats(database *gorm.DB) map[string]interface{} {
	sqlDB, err := database.DB()
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}
