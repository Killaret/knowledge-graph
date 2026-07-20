package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestUserRepository_FindByID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewUserRepository(db)
	id := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "login", "email", "password_hash", "role_id", "created_at", "deleted_at"}).
		AddRow(id, "testuser", "test@example.com", "hash", nil, now, nil)

	mock.ExpectQuery(`SELECT \* FROM "users" WHERE id = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(id, 1).
		WillReturnRows(rows)

	user, err := repo.FindByID(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "testuser", user.Login)
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewUserRepository(db)
	id := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "users" WHERE id = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs(id, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	user, err := repo.FindByID(context.Background(), id)
	require.NoError(t, err)
	assert.Nil(t, user)
}

func TestUserRepository_FindByLogin(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewUserRepository(db)
	id := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "login", "email", "password_hash", "role_id", "created_at", "deleted_at"}).
		AddRow(id, "testuser", "test@example.com", "hash", nil, now, nil)

	mock.ExpectQuery(`SELECT \* FROM "users" WHERE login = \$1 ORDER BY "users"."id" LIMIT \$2`).
		WithArgs("testuser", 1).
		WillReturnRows(rows)

	user, err := repo.FindByLogin(context.Background(), "testuser")
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, id, user.ID)
}

func TestUserRepository_Exists(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewUserRepository(db)
	id := uuid.New()

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "users" WHERE id = \$1`).
		WithArgs(id).
		WillReturnRows(rows)

	exists, err := repo.Exists(context.Background(), id)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestUserRepository_Delete(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewUserRepository(db)
	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "users" WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Delete(context.Background(), id)
	require.NoError(t, err)
}
