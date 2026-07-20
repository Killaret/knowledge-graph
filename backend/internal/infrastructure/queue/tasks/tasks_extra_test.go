package tasks

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
)

type mockBackupService struct{}

func (m *mockBackupService) UploadBackup(ctx context.Context, localPath, remoteKey string) error {
	return nil
}

type mockDraftCleanupService struct{}

func (m *mockDraftCleanupService) DeleteExpired(ctx context.Context, before time.Time) (int, error) {
	return 5, nil
}

type mockCleanupService struct{}

func (m *mockCleanupService) CleanupSoftDeleted(ctx context.Context, days int, tables []string) error {
	return nil
}

type mockTokenCleanupService struct{}

func (m *mockTokenCleanupService) CleanupExpiredTokens(ctx context.Context, batchSize int) error {
	return nil
}

type mockRefreshService struct{}

func (m *mockRefreshService) RefreshRecommendations(ctx context.Context, noteID uuid.UUID) error {
	return nil
}

func TestNewBackupToCloudTask(t *testing.T) {
	task, err := NewBackupToCloudTask("/tmp/backup.sql", "backups/2026-01-01.sql", "2026-01-01")
	assert.NoError(t, err)
	assert.Equal(t, TypeBackupToCloud, task.Type())

	_, err = NewBackupToCloudTask("", "remote", "")
	assert.Error(t, err)
	_, err = NewBackupToCloudTask("local", "", "")
	assert.Error(t, err)
}

func TestHandleBackupToCloud(t *testing.T) {
	task, _ := NewBackupToCloudTask("/tmp/backup.sql", "backups/2026-01-01.sql", "2026-01-01")
	err := HandleBackupToCloud(context.Background(), task, &mockBackupService{})
	assert.NoError(t, err)

	errTask := asynq.NewTask(TypeBackupToCloud, []byte("bad"))
	err = HandleBackupToCloud(context.Background(), errTask, &mockBackupService{})
	assert.Error(t, err)
}

func TestNewCleanupExpiredDraftsTask(t *testing.T) {
	task, err := NewCleanupExpiredDraftsTask(0)
	assert.NoError(t, err)
	assert.Equal(t, TypeCleanupExpiredDrafts, task.Type())
}

func TestHandleCleanupExpiredDrafts(t *testing.T) {
	task, _ := NewCleanupExpiredDraftsTask(24)
	err := HandleCleanupExpiredDrafts(context.Background(), task, &mockDraftCleanupService{})
	assert.NoError(t, err)

	errTask := asynq.NewTask(TypeCleanupExpiredDrafts, []byte("bad"))
	err = HandleCleanupExpiredDrafts(context.Background(), errTask, &mockDraftCleanupService{})
	assert.Error(t, err)
}

func TestNewCleanupSoftDeletedTask(t *testing.T) {
	task, err := NewCleanupSoftDeletedTask(0, nil)
	assert.NoError(t, err)
	assert.Equal(t, TypeCleanupSoftDeleted, task.Type())
}

func TestHandleCleanupSoftDeleted(t *testing.T) {
	task, _ := NewCleanupSoftDeletedTask(30, []string{"notes"})
	err := HandleCleanupSoftDeleted(context.Background(), task, &mockCleanupService{})
	assert.NoError(t, err)

	errTask := asynq.NewTask(TypeCleanupSoftDeleted, []byte("bad"))
	err = HandleCleanupSoftDeleted(context.Background(), errTask, &mockCleanupService{})
	assert.Error(t, err)
}

func TestNewCleanupExpiredTokensTask(t *testing.T) {
	task, err := NewCleanupExpiredTokensTask(0)
	assert.NoError(t, err)
	assert.Equal(t, TypeCleanupExpiredTokens, task.Type())
}

func TestHandleCleanupExpiredTokens(t *testing.T) {
	task, _ := NewCleanupExpiredTokensTask(100)
	err := HandleCleanupExpiredTokens(context.Background(), task, &mockTokenCleanupService{})
	assert.NoError(t, err)

	errTask := asynq.NewTask(TypeCleanupExpiredTokens, []byte("bad"))
	err = HandleCleanupExpiredTokens(context.Background(), errTask, &mockTokenCleanupService{})
	assert.Error(t, err)
}

func TestNewRefreshRecommendationsTask_Extra(t *testing.T) {
	noteID := uuid.New()
	task, err := NewRefreshRecommendationsTask(noteID, 0)
	assert.NoError(t, err)
	assert.Equal(t, TypeRefreshRecommendations, task.Type())

	taskDelayed, err := NewRefreshRecommendationsTask(noteID, time.Minute)
	assert.NoError(t, err)
	assert.NotNil(t, taskDelayed)
}

func TestHandleRefreshRecommendations_Extra(t *testing.T) {
	noteID := uuid.New()
	task, _ := NewRefreshRecommendationsTask(noteID, 0)
	err := HandleRefreshRecommendations(context.Background(), task, &mockRefreshService{})
	assert.NoError(t, err)

	errTask := asynq.NewTask(TypeRefreshRecommendations, []byte("bad"))
	err = HandleRefreshRecommendations(context.Background(), errTask, &mockRefreshService{})
	assert.Error(t, err)
}

type failingRefreshService struct{}

func (m *failingRefreshService) RefreshRecommendations(ctx context.Context, noteID uuid.UUID) error {
	return errors.New("refresh failed")
}

func TestHandleRefreshRecommendations_ServiceError(t *testing.T) {
	noteID := uuid.New()
	task, _ := NewRefreshRecommendationsTask(noteID, 0)
	err := HandleRefreshRecommendations(context.Background(), task, &failingRefreshService{})
	assert.Error(t, err)
}
