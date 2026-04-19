package recommendation

import (
	"context"
	"errors"
	"regexp"
	"sync"
	"testing"

	"knowledge-graph/internal/domain/graph"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// MockTraversalService is a mock for TraversalService interface
type MockTraversalService struct {
	mock.Mock
}

func (m *MockTraversalService) GetSuggestions(ctx context.Context, startID uuid.UUID, topN int) ([]graph.SuggestionResult, error) {
	args := m.Called(ctx, startID, topN)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]graph.SuggestionResult), args.Error(1)
}

// Ensure MockTraversalService implements TraversalService interface
var _ TraversalService = (*MockTraversalService)(nil)

func setupRefreshServiceMock(t *testing.T) (*RefreshService, sqlmock.Sqlmock, *MockTraversalService, func()) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	require.NoError(t, err)

	mockSvc := new(MockTraversalService)
	svc := NewRefreshService(db, nil, mockSvc, 10)

	return svc, mock, mockSvc, func() {
		sqlDB.Close()
	}
}

func TestRefreshService_RefreshRecommendations(t *testing.T) {
	svc, mock, mockSvc, cleanup := setupRefreshServiceMock(t)
	defer cleanup()

	ctx := context.Background()
	noteID := uuid.MustParse("a0000000-0000-0000-0000-000000000001")
	targetID := uuid.MustParse("a0000000-0000-0000-0000-000000000002")

	t.Run("successful refresh", func(t *testing.T) {
		// Mock note lookup
		noteRows := sqlmock.NewRows([]string{"id", "title", "content", "created_at", "updated_at"}).
			AddRow(noteID, "Test Note", "Content", "2024-01-01", "2024-01-01")
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "notes" WHERE id = $1 ORDER BY "notes"."id" LIMIT $2`,
		)).WithArgs(noteID, 1).WillReturnRows(noteRows)

		// Mock traversal service returning suggestions
		suggestions := []graph.SuggestionResult{
			{NodeID: targetID, Score: 0.9},
		}
		mockSvc.On("GetSuggestions", ctx, noteID, 10).Return(suggestions, nil).Once()

		// Mock transaction begin
		mock.ExpectBegin()

		// Mock upsert (ON CONFLICT DO UPDATE)
		mock.ExpectExec(regexp.QuoteMeta(
			`INSERT INTO "note_recommendations"`,
		)).WillReturnResult(sqlmock.NewResult(1, 1))

		// Mock delete not in batch
		mock.ExpectExec(regexp.QuoteMeta(
			`DELETE FROM "note_recommendations"`,
		)).WithArgs(noteID, targetID).WillReturnResult(sqlmock.NewResult(0, 0))

		// Mock transaction commit
		mock.ExpectCommit()

		err := svc.RefreshRecommendations(ctx, noteID)
		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		mockSvc.AssertExpectations(t)
	})

	t.Run("note not found", func(t *testing.T) {
		// Mock note lookup returning error
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "notes" WHERE id = $1 ORDER BY "notes"."id" LIMIT $2`,
		)).WithArgs(noteID, 1).WillReturnError(gorm.ErrRecordNotFound)

		err := svc.RefreshRecommendations(ctx, noteID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "note not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("traversal service error", func(t *testing.T) {
		// Mock note lookup
		noteRows := sqlmock.NewRows([]string{"id", "title", "content", "created_at", "updated_at"}).
			AddRow(noteID, "Test Note", "Content", "2024-01-01", "2024-01-01")
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "notes" WHERE id = $1 ORDER BY "notes"."id" LIMIT $2`,
		)).WithArgs(noteID, 1).WillReturnRows(noteRows)

		// Mock traversal service returning error
		mockSvc.On("GetSuggestions", ctx, noteID, 10).Return(nil, errors.New("traversal failed")).Once()

		err := svc.RefreshRecommendations(ctx, noteID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get suggestions")
		assert.NoError(t, mock.ExpectationsWereMet())
		mockSvc.AssertExpectations(t)
	})

	t.Run("database transaction rollback on save error", func(t *testing.T) {
		// Mock note lookup
		noteRows := sqlmock.NewRows([]string{"id", "title", "content", "created_at", "updated_at"}).
			AddRow(noteID, "Test Note", "Content", "2024-01-01", "2024-01-01")
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "notes" WHERE id = $1 ORDER BY "notes"."id" LIMIT $2`,
		)).WithArgs(noteID, 1).WillReturnRows(noteRows)

		// Mock traversal service
		suggestions := []graph.SuggestionResult{
			{NodeID: targetID, Score: 0.9},
		}
		mockSvc.On("GetSuggestions", ctx, noteID, 10).Return(suggestions, nil).Once()

		// Mock transaction begin
		mock.ExpectBegin()

		// Mock upsert failing
		mock.ExpectExec(regexp.QuoteMeta(
			`INSERT INTO "note_recommendations"`,
		)).WillReturnError(errors.New("database error"))

		// Mock transaction rollback
		mock.ExpectRollback()

		err := svc.RefreshRecommendations(ctx, noteID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save recommendations")
		assert.NoError(t, mock.ExpectationsWereMet())
		mockSvc.AssertExpectations(t)
	})
}

func TestRefreshService_ConcurrentRefresh(t *testing.T) {
	// This test verifies that concurrent refreshes don't cause data corruption
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	require.NoError(t, err)

	mockSvc := new(MockTraversalService)
	svc := NewRefreshService(db, nil, mockSvc, 10)

	ctx := context.Background()
	noteID := uuid.MustParse("a0000000-0000-0000-0000-000000000001")
	targetID1 := uuid.MustParse("a0000000-0000-0000-0000-000000000002")
	targetID2 := uuid.MustParse("a0000000-0000-0000-0000-000000000003")

	// Setup expectations for multiple concurrent calls
	numGoroutines := 5

	for i := 0; i < numGoroutines; i++ {
		// Each goroutine will look up the note
		noteRows := sqlmock.NewRows([]string{"id", "title", "content", "created_at", "updated_at"}).
			AddRow(noteID, "Test Note", "Content", "2024-01-01", "2024-01-01")
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "notes" WHERE id = $1`,
		)).WithArgs(noteID, 1).WillReturnRows(noteRows)

		// Each call returns slightly different results (simulating race)
		suggestions := []graph.SuggestionResult{
			{NodeID: targetID1, Score: 0.9 - float64(i)*0.01},
			{NodeID: targetID2, Score: 0.8 - float64(i)*0.01},
		}
		mockSvc.On("GetSuggestions", ctx, noteID, 10).Return(suggestions, nil).Once()

		// Transaction expectations
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "note_recommendations"`)).
			WillReturnResult(sqlmock.NewResult(2, 2))
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "note_recommendations"`)).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()
	}

	// Run concurrent refreshes
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := svc.RefreshRecommendations(ctx, noteID); err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check that no errors occurred
	errCount := 0
	for err := range errors {
		t.Logf("Error: %v", err)
		errCount++
	}
	assert.Equal(t, 0, errCount, "Concurrent refreshes should not produce errors")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
	mockSvc.AssertExpectations(t)
}
