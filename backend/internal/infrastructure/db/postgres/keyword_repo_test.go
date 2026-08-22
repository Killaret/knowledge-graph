package postgres

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeywordRepository_GetKeywordsWithWeights(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewKeywordRepository(db)
	noteID := uuid.New()

	rows := sqlmock.NewRows([]string{"note_id", "keyword", "weight"}).
		AddRow(noteID, "go", 1.0).
		AddRow(noteID, "test", 0.5)

	mock.ExpectQuery(`SELECT \* FROM "note_keywords" WHERE note_id = \$1`).
		WithArgs(noteID).
		WillReturnRows(rows)

	result, err := repo.GetKeywordsWithWeights(context.Background(), noteID)
	require.NoError(t, err)
	assert.Equal(t, 1.0, result["go"])
	assert.Equal(t, 0.5, result["test"])
}

func TestKeywordRepository_GetKeywordsBatchWithWeights(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewKeywordRepository(db)
	noteID1 := uuid.New()
	noteID2 := uuid.New()

	rows := sqlmock.NewRows([]string{"note_id", "keyword", "weight"}).
		AddRow(noteID1, "go", 1.0).
		AddRow(noteID2, "test", 0.5)

	mock.ExpectQuery(`SELECT \* FROM "note_keywords" WHERE note_id IN \(\$1,\$2\)`).
		WithArgs(noteID1, noteID2).
		WillReturnRows(rows)

	result, err := repo.GetKeywordsBatchWithWeights(context.Background(), []uuid.UUID{noteID1, noteID2})
	require.NoError(t, err)
	assert.Equal(t, 1.0, result[noteID1]["go"])
	assert.Equal(t, 0.5, result[noteID2]["test"])
}

func TestKeywordRepository_DeleteAll(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewKeywordRepository(db)
	noteID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "note_keywords" WHERE note_id = \$1`).
		WithArgs(noteID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.DeleteAll(context.Background(), noteID)
	require.NoError(t, err)
}
