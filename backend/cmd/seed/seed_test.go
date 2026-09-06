//go:build integration

package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"knowledge-graph/internal/config"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/testutil"
)

func TestSeedCreatesTestUser(t *testing.T) {
	if testing.Short() {
		t.Skip("seeder integration test requires a database")
	}

	ctx := context.Background()
	db, cleanup := testutil.SetupTestVectorDB(t)
	defer cleanup()

	require.NoError(t, postgres.RunMigrations(db, "../../migrations"),
		"migrations must be applied before the seeder can create the test user")

	cfg := &config.Config{
		AppEnv:        "test",
		Argon2Time:    3,
		Argon2Memory:  65536,
		Argon2Threads: 4,
	}

	err := runSeed(ctx, cfg, db, "TestPassword123!")
	require.NoError(t, err)

	var login string
	require.NoError(t, db.Raw(
		"SELECT login FROM users WHERE id = '00000000-0000-0000-0000-000000000000'",
	).Scan(&login).Error)
	assert.Equal(t, "testuser", login)
}
