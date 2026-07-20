package graph

import (
	"context"
	"testing"
	"time"

	domainGraph "knowledge-graph/internal/domain/graph"
	"knowledge-graph/internal/domain/note"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockNeighborLoader struct {
	mock.Mock
}

func (m *mockNeighborLoader) GetNeighbors(ctx context.Context, nodeID uuid.UUID) ([]domainGraph.Edge, error) {
	args := m.Called(ctx, nodeID)
	return args.Get(0).([]domainGraph.Edge), args.Error(1)
}

func (m *mockNeighborLoader) GetNeighborsBatch(ctx context.Context, nodeIDs []uuid.UUID) (map[uuid.UUID][]domainGraph.Edge, error) {
	args := m.Called(ctx, nodeIDs)
	return args.Get(0).(map[uuid.UUID][]domainGraph.Edge), args.Error(1)
}

type mockNoteRepoForQueries struct{}

func (m *mockNoteRepoForQueries) Save(ctx context.Context, note *note.Note) error                   { return nil }
func (m *mockNoteRepoForQueries) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	title, _ := note.NewTitle("Neighbor")
	content, _ := note.NewContent("content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, "star", metadata)
	return n, nil
}
func (m *mockNoteRepoForQueries) Delete(ctx context.Context, id uuid.UUID) error                   { return nil }
func (m *mockNoteRepoForQueries) DeleteBatch(ctx context.Context, ids []uuid.UUID) error           { return nil }
func (m *mockNoteRepoForQueries) Restore(ctx context.Context, id uuid.UUID) error                  { return nil }
func (m *mockNoteRepoForQueries) List(ctx context.Context, limit, offset int) ([]*note.Note, int64, error) {
	return nil, 0, nil
}
func (m *mockNoteRepoForQueries) Search(ctx context.Context, query string, limit, offset int) ([]*note.Note, int64, error) {
	return nil, 0, nil
}
func (m *mockNoteRepoForQueries) FindAll(ctx context.Context) ([]*note.Note, error)                 { return nil, nil }
func (m *mockNoteRepoForQueries) FindAllPaginated(ctx context.Context, limit, offset int) ([]*note.Note, int64, error) {
	return nil, 0, nil
}

func TestGetSuggestionsHandler_Handle(t *testing.T) {
	ctx := context.Background()

	loader := new(mockNeighborLoader)
	noteID := uuid.New()
	neighborID := uuid.New()

	loader.On("GetNeighbors", ctx, noteID).Return([]domainGraph.Edge{
		{From: noteID, To: neighborID, Weight: 0.9},
	}, nil)
	loader.On("GetNeighbors", ctx, neighborID).Return([]domainGraph.Edge{}, nil)

	traversalSvc := domainGraph.NewTraversalService(loader, 3, 0.5, "max", false)
	rdb := redis.NewClient(&redis.Options{Addr: miniredis.RunT(t).Addr()})

	handler := NewGetSuggestionsHandler(traversalSvc, &mockNoteRepoForQueries{}, rdb, time.Minute)

	result, err := handler.Handle(ctx, GetSuggestionsQuery{NoteID: noteID, Limit: 5})
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, neighborID, result[0].NoteID)

	// Second call should hit cache
	cached, err := handler.Handle(ctx, GetSuggestionsQuery{NoteID: noteID, Limit: 5})
	require.NoError(t, err)
	assert.Equal(t, result, cached)
}
