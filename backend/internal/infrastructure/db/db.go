package db

import (
	"fmt"
	"time"

	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PoolConfig holds PostgreSQL connection pool settings.
type PoolConfig struct {
	MaxOpenConns           int
	MaxIdleConns           int
	ConnMaxLifetimeSeconds int
	ConnMaxIdleTimeSeconds int
}

// DefaultPoolConfig returns the pool settings historically used by all binaries.
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxOpenConns:           25,
		MaxIdleConns:           5,
		ConnMaxLifetimeSeconds: 300,
		ConnMaxIdleTimeSeconds: 60,
	}
}

func normalizePoolConfig(pool PoolConfig) PoolConfig {
	defaults := DefaultPoolConfig()
	if pool.MaxOpenConns <= 0 {
		pool.MaxOpenConns = defaults.MaxOpenConns
	}
	if pool.MaxIdleConns <= 0 {
		pool.MaxIdleConns = defaults.MaxIdleConns
	}
	if pool.ConnMaxLifetimeSeconds <= 0 {
		pool.ConnMaxLifetimeSeconds = defaults.ConnMaxLifetimeSeconds
	}
	if pool.ConnMaxIdleTimeSeconds <= 0 {
		pool.ConnMaxIdleTimeSeconds = defaults.ConnMaxIdleTimeSeconds
	}
	return pool
}

// Connect creates a PostgreSQL connection with the default pool and returns *gorm.DB.
// DSN is passed explicitly, not read from env.
func Connect(dsn string) (*gorm.DB, error) {
	return ConnectWithPool(dsn, DefaultPoolConfig())
}

// ConnectWithPool creates a PostgreSQL connection with the given pool settings.
// Non-positive pool values fall back to the defaults.
func ConnectWithPool(dsn string, pool PoolConfig) (*gorm.DB, error) {
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

	pool = normalizePoolConfig(pool)
	sqlDB.SetMaxOpenConns(pool.MaxOpenConns)
	sqlDB.SetMaxIdleConns(pool.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(pool.ConnMaxLifetimeSeconds) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(pool.ConnMaxIdleTimeSeconds) * time.Second)

	return database, nil
}

// GetPoolStats returns connection pool statistics for monitoring
func GetPoolStats(database *gorm.DB) map[string]interface{} {
	if database == nil {
		return map[string]interface{}{"error": "database is nil"}
	}
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
