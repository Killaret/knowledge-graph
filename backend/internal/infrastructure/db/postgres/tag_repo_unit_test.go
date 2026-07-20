package postgres

import (
	"context"
	"testing"
	"time"

	tagDomain "knowledge-graph/internal/domain/tag"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestTagRepository_Create(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewTagRepository(db)

	tag, err := tagDomain.New("golang")
	require.NoError(t, err)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "tags" \("name","created_at","id"\) VALUES \(\$1,\$2,\$3\) RETURNING "id"`).
		WithArgs(tag.Name(), sqlmock.AnyArg(), tag.ID()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(tag.ID()))
	mock.ExpectCommit()

	ctx := context.Background()
	require.NoError(t, repo.Create(ctx, tag))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTagRepository_FindByID_Found(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewTagRepository(db)

	id := uuid.New()
	now := time.Now()
	mock.ExpectQuery(`SELECT \* FROM "tags" WHERE id = \$1 ORDER BY "tags"."id" LIMIT \$2`).
		WithArgs(id, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at"}).AddRow(id, "golang", now))

	ctx := context.Background()
	found, err := repo.FindByID(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "golang", found.Name())
}

func TestTagRepository_FindByID_NotFound(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewTagRepository(db)

	id := uuid.New()
	mock.ExpectQuery(`SELECT \* FROM "tags" WHERE id = \$1 ORDER BY "tags"."id" LIMIT \$2`).
		WithArgs(id, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	ctx := context.Background()
	found, err := repo.FindByID(ctx, id)
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestTagRepository_FindByName_Found(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewTagRepository(db)

	id := uuid.New()
	now := time.Now()
	mock.ExpectQuery(`SELECT \* FROM "tags" WHERE name = \$1 ORDER BY "tags"."id" LIMIT \$2`).
		WithArgs("golang", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at"}).AddRow(id, "golang", now))

	ctx := context.Background()
	found, err := repo.FindByName(ctx, "golang")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, id, found.ID())
}

func TestTagRepository_FindAll(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewTagRepository(db)

	now := time.Now()
	id1 := uuid.New()
	id2 := uuid.New()
	mock.ExpectQuery(`SELECT \* FROM "tags"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at"}).
			AddRow(id1, "go", now).
			AddRow(id2, "python", now))

	ctx := context.Background()
	tags, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, tags, 2)
}

func TestTagRepository_Update(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewTagRepository(db)

	tag, err := tagDomain.New("go")
	require.NoError(t, err)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "tags" SET`).
		WithArgs(tag.Name(), sqlmock.AnyArg(), tag.ID()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ctx := context.Background()
	require.NoError(t, repo.Update(ctx, tag))
}

func TestTagRepository_Delete(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewTagRepository(db)

	id := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "tags" WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ctx := context.Background()
	require.NoError(t, repo.Delete(ctx, id))
}

func TestTagRepository_Exists(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewTagRepository(db)

	id := uuid.New()
	mock.ExpectQuery(`SELECT count\(\*\) FROM "tags" WHERE id = \$1`).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	ctx := context.Background()
	exists, err := repo.Exists(ctx, id)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestTagRepository_IsTagAssignedToNote(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewTagRepository(db)

	noteID := uuid.New()
	tagID := uuid.New()
	mock.ExpectQuery(`SELECT count\(\*\) FROM "note_tags" WHERE note_id = \$1 AND tag_id = \$2`).
		WithArgs(noteID, tagID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	ctx := context.Background()
	assigned, err := repo.IsTagAssignedToNote(ctx, noteID, tagID)
	require.NoError(t, err)
	assert.True(t, assigned)
}

func TestTagRepository_GetTagsByNoteID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewTagRepository(db)

	noteID := uuid.New()
	tagID := uuid.New()
	now := time.Now()
	mock.ExpectQuery(`SELECT tags\.\* FROM "tags" JOIN note_tags ON note_tags\.tag_id = tags\.id WHERE note_tags\.note_id = \$1`).
		WithArgs(noteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at"}).AddRow(tagID, "go", now))

	ctx := context.Background()
	tags, err := repo.GetTagsByNoteID(ctx, noteID)
	require.NoError(t, err)
	assert.Len(t, tags, 1)
}
