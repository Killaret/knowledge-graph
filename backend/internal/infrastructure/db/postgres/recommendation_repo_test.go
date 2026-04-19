package postgres

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupRecommendationRepoMock(t *testing.T) (*RecommendationRepository, sqlmock.Sqlmock, func()) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := NewRecommendationRepository(db)
	return repo, mock, func() {
		sqlDB.Close()
	}
}

func TestRecommendationRepository_Get(t *testing.T) {
	repo, mock, cleanup := setupRecommendationRepoMock(t)
	defer cleanup()

	ctx := context.Background()
	noteID := uuid.MustParse("a0000000-0000-0000-0000-000000000001")
	limit := 5

	t.Run("successful retrieval", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"note_id", "recommended_note_id", "score", "created_at", "updated_at"}).
			AddRow(noteID, uuid.MustParse("a0000000-0000-0000-0000-000000000002"), 0.9, time.Now(), time.Now()).
			AddRow(noteID, uuid.MustParse("a0000000-0000-0000-0000-000000000003"), 0.8, time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "note_recommendations" WHERE note_id = $1 ORDER BY score DESC LIMIT $2`,
		)).WithArgs(noteID, limit).WillReturnRows(rows)

		recs, err := repo.Get(ctx, noteID, limit)
		require.NoError(t, err)
		assert.Len(t, recs, 2)
		assert.Equal(t, 0.9, recs[0].Score)
		assert.Equal(t, 0.8, recs[1].Score)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "note_recommendations" WHERE note_id = $1 ORDER BY score DESC LIMIT $2`,
		)).WithArgs(noteID, limit).WillReturnRows(sqlmock.NewRows([]string{"note_id", "recommended_note_id", "score", "created_at", "updated_at"}))

		recs, err := repo.Get(ctx, noteID, limit)
		require.NoError(t, err)
		assert.Empty(t, recs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRecommendationRepository_SaveBatch(t *testing.T) {
	repo, mock, cleanup := setupRecommendationRepoMock(t)
	defer cleanup()

	ctx := context.Background()
	noteID := uuid.MustParse("a0000000-0000-0000-0000-000000000001")

	t.Run("successful upsert", func(t *testing.T) {
		recs := map[uuid.UUID]float64{
			uuid.MustParse("a0000000-0000-0000-0000-000000000002"): 0.9,
			uuid.MustParse("a0000000-0000-0000-0000-000000000003"): 0.8,
		}

		// Expect INSERT with ON CONFLICT DO UPDATE
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`INSERT INTO "note_recommendations" ("note_id","recommended_note_id","score","created_at","updated_at") VALUES ($1,$2,$3,$4,$5),($6,$7,$8,$9,$10) ON CONFLICT ("note_id","recommended_note_id") DO UPDATE SET "score"="excluded"."score","updated_at"="excluded"."updated_at"`,
		)).WillReturnResult(sqlmock.NewResult(2, 2))
		mock.ExpectCommit()

		err := repo.SaveBatch(ctx, noteID, recs)
		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty map returns nil", func(t *testing.T) {
		err := repo.SaveBatch(ctx, noteID, map[uuid.UUID]float64{})
		require.NoError(t, err)
		// No queries expected
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRecommendationRepository_DeleteByNote(t *testing.T) {
	repo, mock, cleanup := setupRecommendationRepoMock(t)
	defer cleanup()

	ctx := context.Background()
	noteID := uuid.MustParse("a0000000-0000-0000-0000-000000000001")

	t.Run("successful deletion", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`DELETE FROM "note_recommendations" WHERE note_id = $1`,
		)).WithArgs(noteID).WillReturnResult(sqlmock.NewResult(0, 3))
		mock.ExpectCommit()

		err := repo.DeleteByNote(ctx, noteID)
		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRecommendationRepository_DeleteNotInBatch(t *testing.T) {
	repo, mock, cleanup := setupRecommendationRepoMock(t)
	defer cleanup()

	ctx := context.Background()
	noteID := uuid.MustParse("a0000000-0000-0000-0000-000000000001")
	keepID := uuid.MustParse("a0000000-0000-0000-0000-000000000002")

	t.Run("delete not in batch", func(t *testing.T) {
		keep := map[uuid.UUID]float64{
			keepID: 0.9,
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`DELETE FROM "note_recommendations" WHERE note_id = $1 AND recommended_note_id NOT IN ($2)`,
		)).WithArgs(noteID, keepID).WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectCommit()

		err := repo.DeleteNotInBatch(ctx, noteID, keep)
		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty keep map deletes all", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(
			`DELETE FROM "note_recommendations" WHERE note_id = $1`,
		)).WithArgs(noteID).WillReturnResult(sqlmock.NewResult(0, 5))
		mock.ExpectCommit()

		err := repo.DeleteNotInBatch(ctx, noteID, map[uuid.UUID]float64{})
		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRecommendationRepository_GetNotesThatRecommend(t *testing.T) {
	repo, mock, cleanup := setupRecommendationRepoMock(t)
	defer cleanup()

	ctx := context.Background()
	recommendedID := uuid.MustParse("a0000000-0000-0000-0000-000000000004")

	t.Run("successful retrieval", func(t *testing.T) {
		note1 := uuid.MustParse("a0000000-0000-0000-0000-000000000002")
		note2 := uuid.MustParse("a0000000-0000-0000-0000-000000000003")

		rows := sqlmock.NewRows([]string{"note_id"}).
			AddRow(note1).
			AddRow(note2)

		mock.ExpectQuery(`SELECT "note_id" FROM "note_recommendations" WHERE recommended_note_id = \$1`).
			WithArgs(recommendedID).WillReturnRows(rows)

		result, err := repo.GetNotesThatRecommend(ctx, recommendedID)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Contains(t, result, note1)
		assert.Contains(t, result, note2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		mock.ExpectQuery(`SELECT "note_id" FROM "note_recommendations" WHERE recommended_note_id = \$1`).
			WithArgs(recommendedID).WillReturnRows(sqlmock.NewRows([]string{"note_id"}))

		result, err := repo.GetNotesThatRecommend(ctx, recommendedID)
		require.NoError(t, err)
		assert.Empty(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRecommendationRepository_Count(t *testing.T) {
	repo, mock, cleanup := setupRecommendationRepoMock(t)
	defer cleanup()

	ctx := context.Background()
	noteID := uuid.MustParse("a0000000-0000-0000-0000-000000000001")

	t.Run("successful count", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"count"}).AddRow(5)

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT count(*) FROM "note_recommendations" WHERE note_id = $1`,
		)).WithArgs(noteID).WillReturnRows(rows)

		count, err := repo.Count(ctx, noteID)
		require.NoError(t, err)
		assert.Equal(t, int64(5), count)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
