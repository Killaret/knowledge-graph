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
	"knowledge-graph-graph-service/internal/engine"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func TestGetPublicGraphHandler_LinkIdentityAndGammaOrigin(t *testing.T) {
	mockDB := &mockPostgresClient{}
	c, mr := newTestCache(t)
	defer mr.Close()

	server := NewHTTPServer(mockDB, c, 1000, 2)

	mockDB.On("GetNotes", mock.Anything, db.NotesFilter{IsPublic: true}).Return([]*db.Note{
		{ID: "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", Title: "A", Type: "star", Public: true},
		{ID: "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12", Title: "B", Type: "star", Public: true},
	}, []*db.Link{
		{
			ID:          "660e8400-e29b-41d4-a716-446655440002",
			Source:      "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
			Target:      "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
			LinkType:    "related",
			Weight:      0.9,
			SourceType:  "user",
			GammaOrigin: true,
		},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/graph/public", nil)
	rr := httptest.NewRecorder()

	server.GetPublicGraphHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp GraphApiResponse
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Len(t, resp.Data.Links, 1)
	assert.Equal(t, "660e8400-e29b-41d4-a716-446655440002", resp.Data.Links[0].ID)
	assert.True(t, resp.Data.Links[0].GammaOrigin)

	var raw map[string]interface{}
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &raw))
	rawLink := raw["data"].(map[string]interface{})["links"].([]interface{})[0].(map[string]interface{})
	assert.Contains(t, rawLink, "id")
	assert.Equal(t, true, rawLink["gamma_origin"])
}

func TestGetPublicGraphHandlerNoCache(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		expectNew bool
		wantCalls int
	}{
		{"cached empty", "", false, 1},
		{"cached false", "nocache=false", false, 1},
		{"bypass 1", "nocache=1", true, 2},
		{"bypass true", "nocache=true", true, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &mockPostgresClient{}
			c, mr := newTestCache(t)
			defer mr.Close()

			server := NewHTTPServer(mockDB, c, 1000, 2)

			mockDB.On("GetNotes", mock.Anything, db.NotesFilter{IsPublic: true}).Return([]*db.Note{
				{ID: "old", Title: "Cached Note", Type: "star", Public: true},
			}, []*db.Link{}, nil).Once()
			if tt.expectNew {
				mockDB.On("GetNotes", mock.Anything, db.NotesFilter{IsPublic: true}).Return([]*db.Note{
					{ID: "new", Title: "Fresh Note", Type: "star", Public: true},
				}, []*db.Link{}, nil).Once()
			}

			url := "/api/v1/graph/public"
			if tt.query != "" {
				url = url + "?" + tt.query
			}

			first := httptest.NewRequest(http.MethodGet, url, nil)
			rr1 := httptest.NewRecorder()
			server.GetPublicGraphHandler(rr1, first)
			assert.Equal(t, http.StatusOK, rr1.Code)

			var resp1 GraphApiResponse
			assert.NoError(t, json.Unmarshal(rr1.Body.Bytes(), &resp1))
			assert.Len(t, resp1.Data.Nodes, 1)
			assert.Equal(t, "old", resp1.Data.Nodes[0].ID)
			assert.NotNil(t, resp1.Meta)

			second := httptest.NewRequest(http.MethodGet, url, nil)
			rr2 := httptest.NewRecorder()
			server.GetPublicGraphHandler(rr2, second)
			assert.Equal(t, http.StatusOK, rr2.Code)

			var resp2 GraphApiResponse
			assert.NoError(t, json.Unmarshal(rr2.Body.Bytes(), &resp2))
			assert.Len(t, resp2.Data.Nodes, 1)
			assert.NotNil(t, resp2.Meta)

			if tt.expectNew {
				assert.Equal(t, "new", resp2.Data.Nodes[0].ID)
				assert.NotEqual(t, resp1.Meta.Hash, resp2.Meta.Hash)
			} else {
				assert.Equal(t, "old", resp2.Data.Nodes[0].ID)
				assert.Equal(t, resp1.Meta.Hash, resp2.Meta.Hash)
			}

			mockDB.AssertNumberOfCalls(t, "GetNotes", tt.wantCalls)
		})
	}
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

// ── SYNC-1 stage A: delta must be computed against the client's own ────────
// snapshot, removals must reach the client, and an unknown version must be
// answered with resync — never with "everything was added".

func TestGetDeltaHandler_UsesClientSnapshotNotCurrentCache(t *testing.T) {
	mockDB := &mockPostgresClient{}
	c, mr := newTestCache(t)
	defer mr.Close()
	rc := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rc.Close()
	ctx := context.Background()

	server := NewHTTPServer(mockDB, c, 1000, 2)

	noteA := &db.Note{ID: "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", Title: "A", Type: "star"}
	noteB := &db.Note{ID: "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12", Title: "B", Type: "star"}

	// Tab A loads the graph when it has one note.
	mockDB.On("GetNotes", mock.Anything, db.NotesFilter{UserID: "user-1"}).
		Return([]*db.Note{noteA}, []*db.Link{}, nil).Once()
	fullReq := httptest.NewRequest(http.MethodGet, "/api/v1/graph/full", nil)
	fullReq = fullReq.WithContext(withUserID(fullReq.Context(), "user-1"))
	rr := httptest.NewRecorder()
	server.GetFullGraphHandler(rr, fullReq)
	require.Equal(t, http.StatusOK, rr.Code)
	var resp1 GraphApiResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp1))
	hashA := resp1.Meta.Hash
	require.NotEmpty(t, hashA)

	// A second note is created; the event clears the "full" pointer (and
	// deltas), but the snapshot the client holds must survive.
	mockDB.On("GetNotes", mock.Anything, db.NotesFilter{UserID: "user-1"}).
		Return([]*db.Note{noteA, noteB}, []*db.Link{}, nil).Twice()
	require.NoError(t, rc.Del(ctx, "graph-service:full:user-1").Err())

	// Another tab reloads: the full-layout cache now holds the newer version.
	rr = httptest.NewRecorder()
	server.GetFullGraphHandler(rr, fullReq)
	require.Equal(t, http.StatusOK, rr.Code)

	// Tab A polls its delta — it must contain note B, not an empty delta
	// computed against the other tab's cached layout.
	deltaReq := httptest.NewRequest(http.MethodGet, "/api/v1/graph/delta?last_hash="+hashA, nil)
	deltaReq = deltaReq.WithContext(withUserID(deltaReq.Context(), "user-1"))
	drr := httptest.NewRecorder()
	server.GetDeltaHandler(drr, deltaReq)
	require.Equal(t, http.StatusOK, drr.Code)

	var delta engine.DeltaResponse
	require.NoError(t, json.Unmarshal(drr.Body.Bytes(), &delta))
	assert.False(t, delta.Resync)
	require.Len(t, delta.AddedNodes, 1)
	assert.Equal(t, noteB.ID, delta.AddedNodes[0].ID)
	assert.NotEmpty(t, delta.CurrentHash)
	assert.NotEqual(t, hashA, delta.CurrentHash)
}

func TestGetDeltaHandler_RemovedLinkReachesClient(t *testing.T) {
	mockDB := &mockPostgresClient{}
	c, mr := newTestCache(t)
	defer mr.Close()
	rc := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rc.Close()
	ctx := context.Background()

	server := NewHTTPServer(mockDB, c, 1000, 2)

	noteA := &db.Note{ID: "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", Title: "A", Type: "star"}
	noteB := &db.Note{ID: "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12", Title: "B", Type: "star"}
	linkAB := &db.Link{
		ID:         "660e8400-e29b-41d4-a716-446655440002",
		Source:     noteA.ID,
		Target:     noteB.ID,
		LinkType:   "related",
		Weight:     0.9,
		SourceType: "user",
	}

	// Client holds the version with the link.
	mockDB.On("GetNotes", mock.Anything, db.NotesFilter{UserID: "user-1"}).
		Return([]*db.Note{noteA, noteB}, []*db.Link{linkAB}, nil).Once()
	fullReq := httptest.NewRequest(http.MethodGet, "/api/v1/graph/full", nil)
	fullReq = fullReq.WithContext(withUserID(fullReq.Context(), "user-1"))
	rr := httptest.NewRecorder()
	server.GetFullGraphHandler(rr, fullReq)
	require.Equal(t, http.StatusOK, rr.Code)
	var resp GraphApiResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	hashA := resp.Meta.Hash

	// The link is deleted; the event clears the full-layout cache.
	mockDB.On("GetNotes", mock.Anything, db.NotesFilter{UserID: "user-1"}).
		Return([]*db.Note{noteA, noteB}, []*db.Link{}, nil).Once()
	require.NoError(t, rc.Del(ctx, "graph-service:full:user-1").Err())

	deltaReq := httptest.NewRequest(http.MethodGet, "/api/v1/graph/delta?last_hash="+hashA, nil)
	deltaReq = deltaReq.WithContext(withUserID(deltaReq.Context(), "user-1"))
	drr := httptest.NewRecorder()
	server.GetDeltaHandler(drr, deltaReq)
	require.Equal(t, http.StatusOK, drr.Code)

	var delta engine.DeltaResponse
	require.NoError(t, json.Unmarshal(drr.Body.Bytes(), &delta))
	assert.False(t, delta.Resync)
	assert.Empty(t, delta.AddedNodes)
	require.Len(t, delta.RemovedLinks, 1)
	assert.Equal(t, linkAB.Source, delta.RemovedLinks[0].Source)
	assert.Equal(t, linkAB.Target, delta.RemovedLinks[0].Target)
}

func TestGetDeltaHandler_UnknownSnapshotAnswersResync(t *testing.T) {
	mockDB := &mockPostgresClient{}
	c, mr := newTestCache(t)
	defer mr.Close()

	server := NewHTTPServer(mockDB, c, 1000, 2)

	deltaReq := httptest.NewRequest(http.MethodGet, "/api/v1/graph/delta?last_hash=never-served", nil)
	deltaReq = deltaReq.WithContext(withUserID(deltaReq.Context(), "user-1"))
	drr := httptest.NewRecorder()
	server.GetDeltaHandler(drr, deltaReq)
	require.Equal(t, http.StatusOK, drr.Code)

	var delta engine.DeltaResponse
	require.NoError(t, json.Unmarshal(drr.Body.Bytes(), &delta))
	assert.True(t, delta.Resync)
	assert.Empty(t, delta.AddedNodes)
	assert.Empty(t, delta.AddedLinks)
	// No snapshot, no work: the handler must not even query Postgres.
	mockDB.AssertNotCalled(t, "GetNotes", mock.Anything, mock.Anything)
}
