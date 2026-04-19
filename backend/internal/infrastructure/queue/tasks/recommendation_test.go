package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRefreshService is a mock for refresh operations
type MockRefreshService struct {
	mock.Mock
}

func (m *MockRefreshService) RefreshRecommendations(ctx context.Context, noteID uuid.UUID) error {
	args := m.Called(ctx, noteID)
	return args.Error(0)
}

func TestNewRefreshRecommendationsTask(t *testing.T) {
	noteID := uuid.MustParse("a0000000-0000-0000-0000-000000000001")

	t.Run("create task without delay", func(t *testing.T) {
		task, err := NewRefreshRecommendationsTask(noteID, 0)
		require.NoError(t, err)
		assert.Equal(t, TypeRefreshRecommendations, task.Type())

		var payload RefreshRecommendationsPayload
		err = json.Unmarshal(task.Payload(), &payload)
		require.NoError(t, err)
		assert.Equal(t, noteID, payload.NoteID)
	})

	t.Run("create task with delay", func(t *testing.T) {
		delay := 5 * time.Second
		task, err := NewRefreshRecommendationsTask(noteID, delay)
		require.NoError(t, err)
		assert.Equal(t, TypeRefreshRecommendations, task.Type())

		// Check that ProcessIn option is set (we can't directly verify, but task should be valid)
		assert.NotNil(t, task)
	})
}

func TestHandleRefreshRecommendations(t *testing.T) {
	ctx := context.Background()
	noteID := uuid.MustParse("a0000000-0000-0000-0000-000000000001")

	t.Run("successful handling", func(t *testing.T) {
		mockSvc := new(MockRefreshService)
		mockSvc.On("RefreshRecommendations", ctx, noteID).Return(nil).Once()

		payload := RefreshRecommendationsPayload{NoteID: noteID}
		payloadBytes, _ := json.Marshal(payload)
		task := asynq.NewTask(TypeRefreshRecommendations, payloadBytes)

		// Create a wrapper function that matches the signature
		handler := func(ctx context.Context, t *asynq.Task) error {
			return HandleRefreshRecommendations(ctx, t, mockSvc)
		}

		err := handler(ctx, task)
		require.NoError(t, err)
		mockSvc.AssertExpectations(t)
	})

	t.Run("invalid payload", func(t *testing.T) {
		task := asynq.NewTask(TypeRefreshRecommendations, []byte("invalid json"))

		handler := func(ctx context.Context, t *asynq.Task) error {
			return HandleRefreshRecommendations(ctx, t, nil)
		}

		err := handler(ctx, task)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal")
	})

	t.Run("service error", func(t *testing.T) {
		mockSvc := new(MockRefreshService)
		serviceErr := errors.New("service error")
		mockSvc.On("RefreshRecommendations", ctx, noteID).Return(serviceErr).Once()

		payload := RefreshRecommendationsPayload{NoteID: noteID}
		payloadBytes, _ := json.Marshal(payload)
		task := asynq.NewTask(TypeRefreshRecommendations, payloadBytes)

		handler := func(ctx context.Context, t *asynq.Task) error {
			return HandleRefreshRecommendations(ctx, t, mockSvc)
		}

		err := handler(ctx, task)
		require.Error(t, err)
		assert.Equal(t, serviceErr, err)
		mockSvc.AssertExpectations(t)
	})
}

func TestHandleRefreshRecommendations_Concurrent(t *testing.T) {
	// Test concurrent task processing
	ctx := context.Background()

	mockSvc := new(MockRefreshService)

	// Setup expectations for concurrent calls
	for i := 0; i < 10; i++ {
		noteID := uuid.MustParse("a0000000-0000-0000-0000-000000000001")
		mockSvc.On("RefreshRecommendations", ctx, noteID).Return(nil)
	}

	// Process multiple tasks concurrently
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			noteID := uuid.MustParse("a0000000-0000-0000-0000-000000000001")
			payload := RefreshRecommendationsPayload{NoteID: noteID}
			payloadBytes, _ := json.Marshal(payload)
			task := asynq.NewTask(TypeRefreshRecommendations, payloadBytes)

			handler := func(ctx context.Context, t *asynq.Task) error {
				return HandleRefreshRecommendations(ctx, t, mockSvc)
			}

			if err := handler(ctx, task); err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	errCount := 0
	for err := range errors {
		t.Logf("Error: %v", err)
		errCount++
	}
	assert.Equal(t, 0, errCount, "Concurrent handling should not produce errors")
}
