//go:build integration

package api

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"knowledge-graph-graph-service/internal/cache"
	"knowledge-graph-graph-service/internal/db"
	graphservice "knowledge-graph-graph-service/proto"
)

type GRPCIntegrationTestSuite struct {
	suite.Suite
	redis    *miniredis.Miniredis
	server   *grpc.Server
	client   graphservice.GraphServiceClient
	conn     *grpc.ClientConn
	listener *bufconn.Listener
	pgPool   *pgxpool.Pool
	cache    *cache.RedisCache
	postgres db.PostgresClient
}

func (s *GRPCIntegrationTestSuite) SetupSuite() {
	// Setup miniredis
	s.redis = miniredis.RunT(s.T())

	// Setup PostgreSQL connection pool
	// In a real integration test, we would use testcontainers-go
	// For now, we'll skip actual PostgreSQL and use a mock
	// This is a placeholder for a proper integration test

	// Create bufconn listener for in-memory gRPC
	s.listener = bufconn.Listen(1024 << 10)

	// Create Redis client
	redisClient := s.redis.RedisClient()
	s.cache = cache.NewRedisCache(redisClient)

	// Create mock PostgreSQL client (placeholder)
	// In a real integration test, this would connect to a test database
	s.postgres = &mockPostgresClient{}

	// Create gRPC server
	s.server = grpc.NewServer()
	service := NewGraphService(s.postgres, s.cache, 1000)
	graphservice.RegisterGraphServiceServer(s.server, service)

	go func() {
		if err := s.server.Serve(s.listener); err != nil {
			s.T().Logf("gRPC server error: %v", err)
		}
	}()

	// Create gRPC client
	dialer := func(context.Context, string) (net.Conn, error) {
		return s.listener.Dial()
	}

	var err error
	s.conn, err = grpc.NewClient("bufnet", grpc.WithContextDialer(dialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)

	s.client = graphservice.NewGraphServiceClient(s.conn)
}

func (s *GRPCIntegrationTestSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.server != nil {
		s.server.Stop()
	}
	if s.redis != nil {
		s.redis.Close()
	}
	if s.pgPool != nil {
		s.pgPool.Close()
	}
}

func (s *GRPCIntegrationTestSuite) TestGetNoteLayout() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &graphservice.NoteLayoutRequest{
		NoteId: "test-note-1",
		Depth:  2,
		UserId: "test-user",
	}

	resp, err := s.client.GetNoteLayout(ctx, req)
	s.Require().NoError(err)
	s.NotNil(resp)
	s.NotEmpty(resp.Hash)
}

func (s *GRPCIntegrationTestSuite) TestGetNoteLayout_CacheHit() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	noteID := "test-note-cache"
	depth := int32(2)
	userID := "test-user-cache"

	// First request - cache miss
	req1 := &graphservice.NoteLayoutRequest{
		NoteId: noteID,
		Depth:  depth,
		UserId: userID,
	}

	resp1, err := s.client.GetNoteLayout(ctx, req1)
	s.Require().NoError(err)
	s.NotNil(resp1)

	// Second request - should hit cache
	req2 := &graphservice.NoteLayoutRequest{
		NoteId: noteID,
		Depth:  depth,
		UserId: userID,
	}

	resp2, err := s.client.GetNoteLayout(ctx, req2)
	s.Require().NoError(err)
	s.NotNil(resp2)
	s.Equal(resp1.Hash, resp2.Hash, "Hash should be the same from cache")
}

func (s *GRPCIntegrationTestSuite) TestGetFullLayout() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &graphservice.FullLayoutRequest{
		UserId: "test-user",
		Limit:  100,
	}

	stream, err := s.client.GetFullLayout(ctx, req)
	s.Require().NoError(err)

	chunkCount := 0
	for {
		chunk, err := stream.Recv()
		if err != nil {
			break
		}
		chunkCount++
		s.NotNil(chunk)
	}

	s.Greater(chunkCount, 0, "Should receive at least one chunk")
}

func (s *GRPCIntegrationTestSuite) TestGetDelta() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userID := "test-user-delta"

	// First, ensure we have a cached layout
	req := &graphservice.FullLayoutRequest{
		UserId: userID,
		Limit:  10,
	}
	stream, err := s.client.GetFullLayout(ctx, req)
	s.Require().NoError(err)

	// Consume the stream
	for {
		_, err := stream.Recv()
		if err != nil {
			break
		}
	}

	// Now get delta
	deltaReq := &graphservice.DeltaRequest{
		UserId:   userID,
		LastHash: "some-previous-hash",
	}

	deltaResp, err := s.client.GetDelta(ctx, deltaReq)
	s.Require().NoError(err)
	s.NotNil(deltaResp)
	s.NotEmpty(deltaResp.CurrentHash)
}

func (s *GRPCIntegrationTestSuite) TestInvalidRequest() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test nil request
	_, err := s.client.GetNoteLayout(ctx, nil)
	s.Error(err)

	// Test delta without last_hash
	deltaReq := &graphservice.DeltaRequest{
		UserId: "test-user",
	}
	_, err = s.client.GetDelta(ctx, deltaReq)
	s.Error(err)
}

// mockPostgresClient is a mock implementation for testing
type mockPostgresClient struct{}

func (m *mockPostgresClient) GetNotes(ctx context.Context, filter db.NotesFilter) ([]*db.Note, []*db.Link, error) {
	notes := []*db.Note{
		{ID: "note-1", Title: "Note 1"},
		{ID: "note-2", Title: "Note 2"},
		{ID: "note-3", Title: "Note 3"},
	}

	links := []*db.Link{
		{Source: "note-1", Target: "note-2", LinkType: "related", Weight: 0.5},
		{Source: "note-2", Target: "note-3", LinkType: "related", Weight: 0.7},
	}

	return notes, links, nil
}

func (m *mockPostgresClient) GetEmbeddings(ctx context.Context, noteIDs []string) (map[string][]float32, error) {
	return make(map[string][]float32), nil
}

func TestGRPCIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(GRPCIntegrationTestSuite))
}
