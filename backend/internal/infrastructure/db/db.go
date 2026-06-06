package db

import (
	"log"
	"os"
	"time"

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

	// Configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("failed to get database connection:", err)
	}

	// Connection pool settings for main backend
	sqlDB.SetMaxOpenConns(25)                 // Maximum number of open connections
	sqlDB.SetMaxIdleConns(5)                  // Maximum number of idle connections
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // Maximum lifetime of a connection
	sqlDB.SetConnMaxIdleTime(1 * time.Minute) // Maximum idle time of a connection

	log.Println("Connected to PostgreSQL with connection pool")
	log.Println("Pool settings: MaxOpen=25, MaxIdle=5, MaxLifetime=5m, MaxIdleTime=1m")
	log.Println("Note: Database migrations should be applied via SQL files in migrations/ directory")
}

// GetPoolStats returns connection pool statistics for monitoring
func GetPoolStats() map[string]interface{} {
	sqlDB, err := DB.DB()
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
