//go:build integration

package db

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestPostgresClient_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start PostgreSQL container
	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("pgvector/pgvector:pg16"),
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test_user"),
		postgres.WithPassword("test_pass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)
	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %v", err)
		}
	}()

	// Get connection string
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Create connection pool
	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	defer pool.Close()

	// Enable pgvector extension and create tables
	_, err = pool.Exec(ctx, `
		CREATE EXTENSION IF NOT EXISTS vector;

		CREATE TABLE IF NOT EXISTS notes (
			id UUID PRIMARY KEY,
			title TEXT NOT NULL,
			type TEXT NOT NULL DEFAULT 'star',
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW(),
			deleted_at TIMESTAMP,
			creator_id UUID,
			is_public BOOLEAN NOT NULL DEFAULT false
		);

		CREATE TABLE IF NOT EXISTS links (
			id UUID PRIMARY KEY,
			source_note_id UUID NOT NULL REFERENCES notes(id),
			target_note_id UUID NOT NULL REFERENCES notes(id),
			link_type TEXT NOT NULL,
			weight FLOAT NOT NULL DEFAULT 0.5,
			source_type TEXT DEFAULT 'user',
			created_at TIMESTAMP DEFAULT NOW(),
			deleted_at TIMESTAMP,
			creator_id UUID
		);

		CREATE TABLE IF NOT EXISTS note_embeddings (
			id UUID PRIMARY KEY,
			note_id UUID NOT NULL REFERENCES notes(id),
			embedding vector(3),
			updated_at TIMESTAMP DEFAULT NOW()
		);
	`)
	require.NoError(t, err)

	// Insert test data (include required visibility columns)
	_, err = pool.Exec(ctx, `
		INSERT INTO notes (id, title, creator_id, is_public) VALUES 
			('550e8400-e29b-41d4-a716-446655440000', 'Note 1', NULL, true),
			('550e8400-e29b-41d4-a716-446655440001', 'Note 2', NULL, true),
			('550e8400-e29b-41d4-a716-446655440002', 'Note 3', NULL, true);
	`)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO links (id, source_note_id, target_note_id, link_type, weight) VALUES
			('660e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440001', 'related', 0.5),
			('660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', 'related', 0.7);
	`)
	require.NoError(t, err)

	// Create client
	client := NewPostgresClient(pool)

	t.Run("GetNotes_All", func(t *testing.T) {
		notes, links, err := client.GetNotes(ctx, NotesFilter{})
		require.NoError(t, err)
		assert.Equal(t, 3, len(notes))
		assert.Equal(t, 2, len(links))
	})

	t.Run("GetNotes_Public", func(t *testing.T) {
		notes, links, err := client.GetNotes(ctx, NotesFilter{IsPublic: true})
		require.NoError(t, err)
		assert.Equal(t, 3, len(notes))
		assert.Equal(t, 2, len(links))
	})

	t.Run("GetNotes_WithRootAndDepth", func(t *testing.T) {
		notes, links, err := client.GetNotes(ctx, NotesFilter{RootID: "550e8400-e29b-41d4-a716-446655440000", Depth: 2})
		require.NoError(t, err)
		assert.Greater(t, len(notes), 0)
		assert.Greater(t, len(links), 0)
	})

	t.Run("GetEmbeddings", func(t *testing.T) {
		// Insert test embedding
		_, err := pool.Exec(ctx, `
			INSERT INTO note_embeddings (id, note_id, embedding) VALUES
				('770e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440000', '[0.1, 0.2, 0.3]'::vector(3));
		`)
		require.NoError(t, err)

		noteIDs := []string{"550e8400-e29b-41d4-a716-446655440000"}
		embeddings, err := client.GetEmbeddings(ctx, noteIDs)
		require.NoError(t, err)
		assert.Contains(t, embeddings, "550e8400-e29b-41d4-a716-446655440000")
		assert.NotEmpty(t, embeddings["550e8400-e29b-41d4-a716-446655440000"])
	})
}
