package db

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewPostgresClient tests the constructor
func TestNewPostgresClient(t *testing.T) {
	// This is a basic unit test that doesn't require a real DB
	// In a real scenario, you'd use a mock or test container

	// For now, just test that nil is handled gracefully
	var pool *pgxpool.Pool
	client := NewPostgresClient(pool)
	assert.NotNil(t, client)
}

// MockPostgresClient is a simple mock for testing
type MockPostgresClient struct {
	notes []*Note
	links []*Link
}

func NewMockPostgresClient() *MockPostgresClient {
	return &MockPostgresClient{
		notes: make([]*Note, 0),
		links: make([]*Link, 0),
	}
}

func (m *MockPostgresClient) GetNotes(ctx context.Context, rootID string, depth int) ([]*Note, []*Link, error) {
	if rootID == "" || depth <= 0 {
		return m.notes, m.links, nil
	}
	// Simple mock implementation
	return m.notes[:1], m.links, nil
}

func (m *MockPostgresClient) AddNote(note *Note) {
	m.notes = append(m.notes, note)
}

func (m *MockPostgresClient) AddLink(link *Link) {
	m.links = append(m.links, link)
}

// TestMockPostgresClient tests the mock implementation
func TestMockPostgresClient(t *testing.T) {
	ctx := context.Background()
	mock := NewMockPostgresClient()

	// Add test data
	mock.AddNote(&Note{ID: "1", Title: "Test Note"})
	mock.AddNote(&Note{ID: "2", Title: "Another Note"})
	mock.AddLink(&Link{Source: "1", Target: "2", LinkType: "reference", Weight: 1.0})

	// Test GetNotes with no root ID (should return all)
	notes, links, err := mock.GetNotes(ctx, "", 0)
	require.NoError(t, err)
	assert.Len(t, notes, 2)
	assert.Len(t, links, 1)

	// Test GetNotes with root ID (depth limited)
	notes, links, err = mock.GetNotes(ctx, "1", 1)
	require.NoError(t, err)
	assert.Len(t, notes, 1)
}
