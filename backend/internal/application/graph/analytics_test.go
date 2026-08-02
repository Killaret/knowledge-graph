package graph

import (
	"context"
	"testing"

	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalytics_EmptyGraph(t *testing.T) {
	linkRepo := newMockLinkRepoForAnalytics()
	noteRepo := newMockNoteRepoForAnalytics()

	a := NewAnalytics(linkRepo, noteRepo)
	result, err := a.ComputeForAll(context.Background())

	require.NoError(t, err)
	assert.Empty(t, result.PageRank)
	assert.Empty(t, result.Clusters)
	assert.Empty(t, result.TopCenters)
}

func TestAnalytics_SimpleGraph(t *testing.T) {
	linkRepo := newMockLinkRepoForAnalytics()
	noteRepo := newMockNoteRepoForAnalytics()

	nodeA := uuid.New()
	nodeB := uuid.New()
	nodeC := uuid.New()

	noteRepo.add(nodeA, "A", "content a")
	noteRepo.add(nodeB, "B", "content b")
	noteRepo.add(nodeC, "C", "content c")

	linkRepo.add(nodeA, nodeB, 0.8)
	linkRepo.add(nodeB, nodeC, 0.7)

	a := NewAnalytics(linkRepo, noteRepo)
	result, err := a.ComputeForAll(context.Background())

	require.NoError(t, err)
	assert.Len(t, result.PageRank, 3)
	assert.Len(t, result.Clusters, 1)
	assert.Len(t, result.TopCenters, 3)
	assert.InDelta(t, 1.0, sum(result.PageRank), 0.001)
}

type mockLinkRepoForAnalytics struct {
	links []*link.Link
}

func newMockLinkRepoForAnalytics() *mockLinkRepoForAnalytics {
	return &mockLinkRepoForAnalytics{}
}

func (m *mockLinkRepoForAnalytics) add(source, target uuid.UUID, weight float64) {
	lt, _ := link.NewLinkType("reference")
	w, _ := link.NewWeight(weight)
	md, _ := link.NewMetadata(nil)
	m.links = append(m.links, link.NewLink(source, target, lt, w, md))
}

func (m *mockLinkRepoForAnalytics) Save(ctx context.Context, l *link.Link) error { return nil }
func (m *mockLinkRepoForAnalytics) FindByID(ctx context.Context, id uuid.UUID) (*link.Link, error) {
	return nil, nil
}
func (m *mockLinkRepoForAnalytics) FindBySource(ctx context.Context, sourceID uuid.UUID) ([]*link.Link, error) {
	return nil, nil
}
func (m *mockLinkRepoForAnalytics) FindByTarget(ctx context.Context, targetID uuid.UUID) ([]*link.Link, error) {
	return nil, nil
}
func (m *mockLinkRepoForAnalytics) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockLinkRepoForAnalytics) DeleteBySource(ctx context.Context, sourceID uuid.UUID) error {
	return nil
}
func (m *mockLinkRepoForAnalytics) List(ctx context.Context) ([]*link.Link, error) {
	return m.links, nil
}
func (m *mockLinkRepoForAnalytics) FindAll(ctx context.Context) ([]*link.Link, error) {
	return m.links, nil
}
func (m *mockLinkRepoForAnalytics) FindAllPaginated(ctx context.Context, limit, offset int) ([]*link.Link, int64, error) {
	return m.links, int64(len(m.links)), nil
}
func (m *mockLinkRepoForAnalytics) Update(ctx context.Context, l *link.Link) error { return nil }

type mockNoteRepoForAnalytics struct {
	notes map[uuid.UUID]*note.Note
}

func newMockNoteRepoForAnalytics() *mockNoteRepoForAnalytics {
	return &mockNoteRepoForAnalytics{notes: make(map[uuid.UUID]*note.Note)}
}

func (m *mockNoteRepoForAnalytics) add(id uuid.UUID, title, content string) {
	t, _ := note.NewTitle(title)
	c, _ := note.NewContent(content)
	md, _ := note.NewMetadata(nil)
	n := note.NewNote(t, c, "star", md)
	m.notes[id] = n
}

func (m *mockNoteRepoForAnalytics) Save(ctx context.Context, n *note.Note) error { return nil }
func (m *mockNoteRepoForAnalytics) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	return m.notes[id], nil
}
func (m *mockNoteRepoForAnalytics) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockNoteRepoForAnalytics) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
	return nil
}
func (m *mockNoteRepoForAnalytics) Restore(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockNoteRepoForAnalytics) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*note.Note, int64, error) {
	return nil, 0, nil
}
func (m *mockNoteRepoForAnalytics) Search(ctx context.Context, userID uuid.UUID, query string, limit, offset int) ([]*note.Note, int64, error) {
	return nil, 0, nil
}
func (m *mockNoteRepoForAnalytics) FindAll(ctx context.Context) ([]*note.Note, error) {
	return nil, nil
}
func (m *mockNoteRepoForAnalytics) FindAllPaginated(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*note.Note, int64, error) {
	return nil, 0, nil
}

func sum(m map[string]float64) float64 {
	s := 0.0
	for _, v := range m {
		s += v
	}
	return s
}
