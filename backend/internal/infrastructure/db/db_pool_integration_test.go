//go:build integration

package db

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// DB-POOL-1: the pool values from config must actually reach sql.DB —
// replacing `sqlDB.SetMaxOpenConns(pool.MaxOpenConns)` with a constant must
// fail here, not pass silently.
func TestConnectWithPool_AppliesConfiguredValues_Integration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	var cleanup func()
	if dsn == "" {
		ctx := context.Background()
		pgContainer, err := pgcontainer.Run(ctx, "postgres:15-alpine",
			pgcontainer.WithDatabase("pooldb"),
			pgcontainer.WithUsername("test"),
			pgcontainer.WithPassword("test"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(120*time.Second),
			),
		)
		require.NoError(t, err)
		cleanup = func() {
			_ = pgContainer.Terminate(context.Background())
		}
		dsn, err = pgContainer.ConnectionString(ctx, "sslmode=disable")
		require.NoError(t, err)
	} else {
		cleanup = func() {}
	}
	defer cleanup()

	database, err := ConnectWithPool(dsn, PoolConfig{
		MaxOpenConns:           7,
		MaxIdleConns:           3,
		ConnMaxLifetimeSeconds: 111,
		ConnMaxIdleTimeSeconds: 22,
	})
	require.NoError(t, err)

	sqlDB, err := database.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	stats := sqlDB.Stats()
	assert.Equal(t, 7, stats.MaxOpenConnections)
}
