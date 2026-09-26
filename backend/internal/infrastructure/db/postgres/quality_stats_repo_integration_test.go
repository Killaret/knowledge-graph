//go:build integration

package postgres

import (
	"context"
	"testing"

	"knowledge-graph/internal/testutil"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Integration check for the NOTE-QUALITY-1 readiness counters against the
// real PostgreSQL schema. The live-stack regression this guards: LinkCount
// read source_id/target_id while the links table has source_note_id and
// target_note_id, so every live assessment failed on SQLSTATE 42703 while
// the mocked unit tests stayed green.
func setupQualityStatsDB(t *testing.T) (*gorm.DB, *QualityStatsRepository, func()) {
	t.Helper()
	db, cleanup := testutil.SetupTestVectorDB(t)
	require.NoError(t, db.AutoMigrate(
		&UserModel{},
		&NoteModel{},
		&LinkModel{},
		&NoteKeywordModel{},
		&NoteEmbeddingModel{},
	))
	return db, NewQualityStatsRepository(db), cleanup
}

func createQualityStatsNote(t *testing.T, db *gorm.DB, title string) uuid.UUID {
	t.Helper()
	m := &NoteModel{Title: title, Content: "content"}
	require.NoError(t, db.Create(m).Error)
	return m.ID
}

func TestQualityStatsRepository_LinkCountUsesNoteColumns(t *testing.T) {
	db, repo, cleanup := setupQualityStatsDB(t)
	defer cleanup()
	ctx := context.Background()

	a := createQualityStatsNote(t, db, "note a")
	b := createQualityStatsNote(t, db, "note b")
	c := createQualityStatsNote(t, db, "note c")
	other := createQualityStatsNote(t, db, "unrelated")

	require.NoError(t, db.Create(&LinkModel{ID: uuid.New(), SourceNoteID: a, TargetNoteID: b}).Error)
	require.NoError(t, db.Create(&LinkModel{ID: uuid.New(), SourceNoteID: c, TargetNoteID: a}).Error)
	require.NoError(t, db.Create(&LinkModel{ID: uuid.New(), SourceNoteID: b, TargetNoteID: c}).Error)

	n, err := repo.LinkCount(ctx, a)
	require.NoError(t, err)
	require.Equal(t, 2, n, "note a participates in two links: one outgoing, one incoming")

	n, err = repo.LinkCount(ctx, other)
	require.NoError(t, err)
	require.Equal(t, 0, n)
}

func TestQualityStatsRepository_KeywordCountAndHasEmbedding(t *testing.T) {
	db, repo, cleanup := setupQualityStatsDB(t)
	defer cleanup()
	ctx := context.Background()

	withKeywords := createQualityStatsNote(t, db, "with keywords")
	plain := createQualityStatsNote(t, db, "plain")

	require.NoError(t, db.Create(&NoteKeywordModel{NoteID: withKeywords, Keyword: "k1", Extractor: "test", Weight: 1}).Error)
	require.NoError(t, db.Create(&NoteKeywordModel{NoteID: withKeywords, Keyword: "k2", Extractor: "test", Weight: 1}).Error)

	n, err := repo.KeywordCount(ctx, withKeywords)
	require.NoError(t, err)
	require.Equal(t, 2, n)

	has, err := repo.HasEmbedding(ctx, withKeywords)
	require.NoError(t, err)
	require.False(t, has)

	vec := make([]float32, 384)
	vec[0] = 0.5
	require.NoError(t, db.Create(&NoteEmbeddingModel{
		NoteID:    plain,
		Embedding: pgvector.NewVector(vec),
		ModelName: "test",
	}).Error)
	has, err = repo.HasEmbedding(ctx, plain)
	require.NoError(t, err)
	require.True(t, has)

	n, err = repo.KeywordCount(ctx, plain)
	require.NoError(t, err)
	require.Equal(t, 0, n)
}
