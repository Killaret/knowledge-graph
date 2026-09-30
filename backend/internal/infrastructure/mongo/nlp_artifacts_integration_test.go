//go:build integration

package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	mongocontainer "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/wait"
)

// NLP-4 criteria 3-4: unique (note_id, pipeline_version) for current,
// supersede on re-save, history flag, cascade delete. Runs in the
// integration phase (check-all non-quick) with a real MongoDB container.

func setupArtifactsRepo(t *testing.T) (*NlpArtifactsRepository, func()) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	container, err := mongocontainer.Run(ctx, "mongo:7", tc.WithWaitStrategy(
		wait.ForLog("Waiting for connections").WithStartupTimeout(60*time.Second),
	))
	require.NoError(t, err)

	connStr, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	client, err := NewClient(ctx, connStr, "nlp_artifacts_test")
	require.NoError(t, err)

	repo := NewNlpArtifactsRepository(client)
	require.NoError(t, repo.EnsureIndexes(ctx))

	return repo, func() {
		_ = client.Close(ctx)
		_ = container.Terminate(context.Background())
		cancel()
	}
}

func newArtifact(noteID uuid.UUID, hash, text string) *NlpArtifact {
	return &NlpArtifact{
		NoteID:          noteID,
		SourceHash:      hash,
		PipelineVersion: "norm-v1",
		ModelVersion:    "test-model",
		NormalizedText:  text,
		Chunks: []NlpArtifactChunk{
			{Idx: 0, Text: text, HeadingPath: []string{"H1"}, CharSpan: [2]int{0, 5}, TokenCount: 2, Kind: "prose"},
		},
		Metrics: NlpArtifactMetrics{RawTokens: 10, NormTokens: 2, Compression: 0.2, Iterations: 1, StopReason: "single_pass"},
	}
}

func TestNlpArtifacts_SaveSupersedesPrevious(t *testing.T) {
	repo, cleanup := setupArtifactsRepo(t)
	defer cleanup()
	ctx := context.Background()
	noteID := uuid.New()

	require.NoError(t, repo.SaveCurrent(ctx, newArtifact(noteID, "hash-a", "v1 text"), true))
	require.NoError(t, repo.SaveCurrent(ctx, newArtifact(noteID, "hash-b", "v2 text"), true))

	current, err := repo.FindCurrent(ctx, noteID, "norm-v1")
	require.NoError(t, err)
	require.NotNil(t, current)
	require.Equal(t, "hash-b", current.SourceHash)
	require.Equal(t, "v2 text", current.NormalizedText)

	superseded, err := repo.CountByStatus(ctx, NlpArtifactSuperseded)
	require.NoError(t, err)
	require.Equal(t, int64(1), superseded)
}

func TestNlpArtifacts_HistoryDisabledDeletesInstead(t *testing.T) {
	repo, cleanup := setupArtifactsRepo(t)
	defer cleanup()
	ctx := context.Background()
	noteID := uuid.New()

	require.NoError(t, repo.SaveCurrent(ctx, newArtifact(noteID, "hash-a", "v1"), false))
	require.NoError(t, repo.SaveCurrent(ctx, newArtifact(noteID, "hash-b", "v2"), false))

	superseded, err := repo.CountByStatus(ctx, NlpArtifactSuperseded)
	require.NoError(t, err)
	require.Equal(t, int64(0), superseded, "history disabled must not accrue superseded docs")

	current, err := repo.FindCurrent(ctx, noteID, "norm-v1")
	require.NoError(t, err)
	require.Equal(t, "hash-b", current.SourceHash)
}

func TestNlpArtifacts_DeleteByNoteIDCascades(t *testing.T) {
	repo, cleanup := setupArtifactsRepo(t)
	defer cleanup()
	ctx := context.Background()
	noteID, other := uuid.New(), uuid.New()

	require.NoError(t, repo.SaveCurrent(ctx, newArtifact(noteID, "h1", "cur"), true))
	require.NoError(t, repo.SaveCurrent(ctx, newArtifact(noteID, "h2", "newer"), true))
	require.NoError(t, repo.SaveCurrent(ctx, newArtifact(other, "hx", "other"), true))

	deleted, err := repo.DeleteByNoteID(ctx, noteID)
	require.NoError(t, err)
	require.Equal(t, int64(2), deleted, "current + superseded must both go")

	current, err := repo.FindCurrent(ctx, noteID, "norm-v1")
	require.NoError(t, err)
	require.Nil(t, current)

	stillThere, err := repo.FindCurrent(ctx, other, "norm-v1")
	require.NoError(t, err)
	require.NotNil(t, stillThere)
}

func TestNlpArtifacts_UniqueCurrentIndex(t *testing.T) {
	repo, cleanup := setupArtifactsRepo(t)
	defer cleanup()
	ctx := context.Background()
	noteID := uuid.New()

	require.NoError(t, repo.SaveCurrent(ctx, newArtifact(noteID, "h1", "one"), true))
	require.NoError(t, repo.SaveCurrent(ctx, newArtifact(noteID, "h2", "two"), true))

	// Two current docs for the same (note_id, pipeline_version) must not exist.
	counts := map[string]int64{}
	current, err := repo.CountByStatus(ctx, NlpArtifactCurrent)
	require.NoError(t, err)
	counts["current"] = current
	require.Equal(t, int64(1), counts["current"])
}
