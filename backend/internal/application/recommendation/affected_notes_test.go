package recommendation

import (
	"context"
	"regexp"
	"testing"

	"knowledge-graph/internal/infrastructure/db/postgres"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupAffectedNotesMock(t *testing.T) (*AffectedNotesService, sqlmock.Sqlmock, func()) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	require.NoError(t, err)

	recRepo := postgres.NewRecommendationRepository(db)
	svc := NewAffectedNotesService(recRepo)

	return svc, mock, func() {
		sqlDB.Close()
	}
}

func TestAffectedNotesService_GetAffectedNotes(t *testing.T) {
	svc, mock, cleanup := setupAffectedNotesMock(t)
	defer cleanup()

	ctx := context.Background()
	targetNoteID := uuid.MustParse("a0000000-0000-0000-0000-000000000001")

	t.Run("returns target note and reverse dependencies", func(t *testing.T) {
		// Mock GetNotesThatRecommend query
		reverseRows := sqlmock.NewRows([]string{"note_id"}).
			AddRow(uuid.MustParse("a0000000-0000-0000-0000-000000000002")).
			AddRow(uuid.MustParse("a0000000-0000-0000-0000-000000000003"))

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT note_id FROM "note_recommendations" WHERE recommended_note_id = $1`,
		)).WithArgs(targetNoteID).WillReturnRows(reverseRows)

		result, err := svc.GetAffectedNotes(ctx, targetNoteID)
		require.NoError(t, err)

		// Should contain target note + 2 reverse dependencies
		assert.Len(t, result, 3)
		assert.Contains(t, result, targetNoteID)
		assert.Contains(t, result, uuid.MustParse("a0000000-0000-0000-0000-000000000002"))
		assert.Contains(t, result, uuid.MustParse("a0000000-0000-0000-0000-000000000003"))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns only target note when no reverse dependencies", func(t *testing.T) {
		// Mock empty result from GetNotesThatRecommend
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT note_id FROM "note_recommendations" WHERE recommended_note_id = $1`,
		)).WithArgs(targetNoteID).WillReturnRows(sqlmock.NewRows([]string{"note_id"}))

		result, err := svc.GetAffectedNotes(ctx, targetNoteID)
		require.NoError(t, err)

		// Should contain only target note
		assert.Len(t, result, 1)
		assert.Equal(t, targetNoteID, result[0])
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("deduplicates duplicate IDs", func(t *testing.T) {
		// Mock query returning duplicate IDs (should not happen in practice, but test dedup)
		duplicateID := uuid.MustParse("a0000000-0000-0000-0000-000000000002")
		reverseRows := sqlmock.NewRows([]string{"note_id"}).
			AddRow(duplicateID).
			AddRow(duplicateID) // Duplicate

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT note_id FROM "note_recommendations" WHERE recommended_note_id = $1`,
		)).WithArgs(targetNoteID).WillReturnRows(reverseRows)

		result, err := svc.GetAffectedNotes(ctx, targetNoteID)
		require.NoError(t, err)

		// Should contain target note + 1 unique reverse dependency
		assert.Len(t, result, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("handles database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT note_id FROM "note_recommendations" WHERE recommended_note_id = $1`,
		)).WithArgs(targetNoteID).WillReturnError(assert.AnError)

		result, err := svc.GetAffectedNotes(ctx, targetNoteID)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDeduplicate(t *testing.T) {
	t.Run("removes duplicates", func(t *testing.T) {
		id1 := uuid.MustParse("a0000000-0000-0000-0000-000000000001")
		id2 := uuid.MustParse("a0000000-0000-0000-0000-000000000002")
		id3 := uuid.MustParse("a0000000-0000-0000-0000-000000000003")

		input := []uuid.UUID{id1, id2, id1, id3, id2, id1}
		result := deduplicate(input)

		assert.Len(t, result, 3)
		assert.Contains(t, result, id1)
		assert.Contains(t, result, id2)
		assert.Contains(t, result, id3)
	})

	t.Run("handles empty slice", func(t *testing.T) {
		result := deduplicate([]uuid.UUID{})
		assert.Empty(t, result)
	})

	t.Run("handles single element", func(t *testing.T) {
		id := uuid.MustParse("a0000000-0000-0000-0000-000000000001")
		result := deduplicate([]uuid.UUID{id})
		assert.Len(t, result, 1)
		assert.Equal(t, id, result[0])
	})

	t.Run("handles all duplicates", func(t *testing.T) {
		id := uuid.MustParse("a0000000-0000-0000-0000-000000000001")
		input := []uuid.UUID{id, id, id, id}
		result := deduplicate(input)
		assert.Len(t, result, 1)
	})
}

func TestReverseCascadeDepth(t *testing.T) {
	// Verify that reverseCascadeDepth is set to 1 to prevent queue explosion
	assert.Equal(t, 1, reverseCascadeDepth, "reverseCascadeDepth should be 1 to prevent queue explosion")
}
