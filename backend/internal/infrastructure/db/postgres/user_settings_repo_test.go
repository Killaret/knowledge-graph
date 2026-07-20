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

func TestUserSettingsRepository_FindByUserID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewUserSettingsRepository(db)
	userID := uuid.New()
	id := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "user_id", "key", "value", "created_at", "updated_at"}).
		AddRow(id, userID, "galactic_mode", []byte(`{"value":true}`), time.Now(), time.Now())

	mock.ExpectQuery(`SELECT \* FROM "user_settings" WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(rows)

	settings, err := repo.FindByUserID(context.Background(), userID)
	require.NoError(t, err)
	assert.Len(t, settings, 1)
	assert.Equal(t, "galactic_mode", settings[0].Key().String())
}

func TestUserSettingsRepository_FindByUserIDAndKey(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewUserSettingsRepository(db)
	userID := uuid.New()
	id := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "user_id", "key", "value", "created_at", "updated_at"}).
		AddRow(id, userID, "preferred_language", []byte(`{"value":"ru"}`), time.Now(), time.Now())

	mock.ExpectQuery(`SELECT \* FROM "user_settings" WHERE user_id = \$1 AND key = \$2 ORDER BY "user_settings"."id" LIMIT \$3`).
		WithArgs(userID, "preferred_language", 1).
		WillReturnRows(rows)

	setting, err := repo.FindByUserIDAndKey(context.Background(), userID, "preferred_language")
	require.NoError(t, err)
	require.NotNil(t, setting)
	value, err := setting.GetValue()
	require.NoError(t, err)
	assert.Equal(t, "ru", value["value"])
}

func TestUserSettingsRepository_FindByUserIDAndKey_NotFound(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewUserSettingsRepository(db)
	userID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "user_settings" WHERE user_id = \$1 AND key = \$2 ORDER BY "user_settings"."id" LIMIT \$3`).
		WithArgs(userID, "preferred_language", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	setting, err := repo.FindByUserIDAndKey(context.Background(), userID, "preferred_language")
	require.NoError(t, err)
	assert.Nil(t, setting)
}

func TestUserSettingsRepository_Delete(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewUserSettingsRepository(db)
	userID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "user_settings" WHERE user_id = \$1 AND key = \$2`).
		WithArgs(userID, "galactic_mode").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Delete(context.Background(), userID, "galactic_mode")
	require.NoError(t, err)
}
