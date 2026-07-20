package recommendation

import (
	"context"

	"knowledge-graph/internal/domain/graph"
	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type mockNoteRepository struct {
	mock.Mock
}

func (m *mockNoteRepository) Save(ctx context.Context, n *note.Note) error {
	return m.Called(ctx, n).Error(0)
}

func (m *mockNoteRepository) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*note.Note), args.Error(1)
}

func (m *mockNoteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockNoteRepository) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
	return m.Called(ctx, ids).Error(0)
}

func (m *mockNoteRepository) Restore(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockNoteRepository) List(ctx context.Context, limit, offset int) ([]*note.Note, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*note.Note), args.Get(1).(int64), args.Error(2)
}

func (m *mockNoteRepository) Search(ctx context.Context, query string, limit, offset int) ([]*note.Note, int64, error) {
	args := m.Called(ctx, query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*note.Note), args.Get(1).(int64), args.Error(2)
}

func (m *mockNoteRepository) FindAll(ctx context.Context) ([]*note.Note, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*note.Note), args.Error(1)
}

func (m *mockNoteRepository) FindAllPaginated(ctx context.Context, limit, offset int) ([]*note.Note, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*note.Note), args.Get(1).(int64), args.Error(2)
}

type mockRecommendationRepository struct {
	mock.Mock
}

func (m *mockRecommendationRepository) Count(ctx context.Context, noteID uuid.UUID) (int64, error) {
	args := m.Called(ctx, noteID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRecommendationRepository) GetNotesThatRecommend(ctx context.Context, recommendedID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, recommendedID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func (m *mockRecommendationRepository) ReplaceRecommendations(ctx context.Context, noteID uuid.UUID, recs map[uuid.UUID]float64) error {
	return m.Called(ctx, noteID, recs).Error(0)
}

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
