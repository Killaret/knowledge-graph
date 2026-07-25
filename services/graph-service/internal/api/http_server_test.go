package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"knowledge-graph-graph-service/internal/cache"
	"knowledge-graph-graph-service/internal/config"
	"knowledge-graph-graph-service/internal/db"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPostgresClient struct {
	mock.Mock
}

func (m *mockPostgresClient) GetNotes(ctx context.Context, filter db.NotesFilter) ([]*db.Note, []*db.Link, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*db.Note), args.Get(1).([]*db.Link), args.Error(2)
}

func (m *mockPostgresClient) GetEmbeddings(ctx context.Context, noteIDs []string) (map[string][]float32, error) {
	args := m.Called(ctx, noteIDs)
	return args.Get(0).(map[string][]float32), args.Error(1)
}

func (m *mockPostgresClient) GetNoteNeighbors(ctx context.Context, filter db.NotesFilter, noteID string, depth int) ([]*db.Neighbor, error) {
	args := m.Called(ctx, filter, noteID, depth)
	return args.Get(0).([]*db.Neighbor), args.Error(1)
}

func (m *mockPostgresClient) GetShortestPath(ctx context.Context, filter db.NotesFilter, fromID, toID string) ([]string, int, float64, error) {
	args := m.Called(ctx, filter, fromID, toID)
	return args.Get(0).([]string), args.Get(1).(int), args.Get(2).(float64), args.Error(3)
}

func (m *mockPostgresClient) GetRecommendationCandidates(ctx context.Context, filter db.NotesFilter, noteID string, depth, limit int) ([]*db.RecommendationCandidate, error) {
	args := m.Called(ctx, filter, noteID, depth, limit)
	return args.Get(0).([]*db.RecommendationCandidate), args.Error(1)
}

func (m *mockPostgresClient) RefreshClosureView(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func newTestCache(t *testing.T) (*cache.RedisCache, *miniredis.Miniredis) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	r := redis.NewClient(&redis.Options{Addr: s.Addr()})
	cfg := &config.Config{
		NoteLayoutTTL: 5 * time.Minute,
		FullLayoutTTL: 5 * time.Minute,
		DeltaTTL:      1 * time.Minute,
	}
	return cache.NewRedisCacheWithConfig(r, cfg), s
}

func TestGetPublicGraphHandler(t *testing.T) {
	mockDB := &mockPostgresClient{}
	c, mr := newTestCache(t)
	defer mr.Close()

	server := NewHTTPServer(mockDB, c, 1000, 2)

	mockDB.On("GetNotes", mock.Anything, db.NotesFilter{IsPublic: true}).Return([]*db.Note{
		{ID: "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", Title: "Public Note", Type: "star", Public: true},
	}, []*db.Link{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/graph/public", nil)
	rr := httptest.NewRecorder()

	server.GetPublicGraphHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp GraphApiResponse
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Len(t, resp.Data.Nodes, 1)
	assert.Equal(t, "Public Note", resp.Data.Nodes[0].Title)
}

func TestGetNoteGraphHandlerWithUser(t *testing.T) {
	mockDB := &mockPostgresClient{}
	c, mr := newTestCache(t)
	defer mr.Close()

	server := NewHTTPServer(mockDB, c, 1000, 2)

	noteID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
	mockDB.On("GetNotes", mock.Anything, db.NotesFilter{UserID: "user-1", RootID: noteID, Depth: 2}).Return([]*db.Note{
		{ID: noteID, Title: "My Note", Type: "star"},
	}, []*db.Link{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/graph/note/"+noteID, nil)
	req = req.WithContext(withUserID(req.Context(), "user-1"))
	rr := httptest.NewRecorder()

	server.GetNoteGraphHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp GraphApiResponse
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Len(t, resp.Data.Nodes, 1)
}

func TestGetNoteGraphHandlerPublicContext(t *testing.T) {
	mockDB := &mockPostgresClient{}
	c, mr := newTestCache(t)
	defer mr.Close()

	server := NewHTTPServer(mockDB, c, 1000, 2)

	noteID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
	mockDB.On("GetNotes", mock.Anything, db.NotesFilter{IsPublic: true, RootID: noteID, Depth: 2}).Return([]*db.Note{
		{ID: noteID, Title: "Public Note", Type: "star", Public: true},
	}, []*db.Link{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/graph/note/"+noteID, nil)
	req = req.WithContext(withPublic(req.Context()))
	rr := httptest.NewRecorder()

	server.GetNoteGraphHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp GraphApiResponse
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Len(t, resp.Data.Nodes, 1)
}
