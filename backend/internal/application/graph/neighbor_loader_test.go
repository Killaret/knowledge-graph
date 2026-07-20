package graph

import (
	"context"
	"testing"

	"knowledge-graph/internal/domain/link"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLinkRepo struct {
	bySource map[uuid.UUID][]*link.Link
	byTarget map[uuid.UUID][]*link.Link
}

func newMockLinkRepo() *mockLinkRepo {
	return &mockLinkRepo{
		bySource: make(map[uuid.UUID][]*link.Link),
		byTarget: make(map[uuid.UUID][]*link.Link),
	}
}

func (m *mockLinkRepo) addLink(l *link.Link) {
	m.bySource[l.SourceNoteID()] = append(m.bySource[l.SourceNoteID()], l)
	m.byTarget[l.TargetNoteID()] = append(m.byTarget[l.TargetNoteID()], l)
}

func (m *mockLinkRepo) Save(ctx context.Context, l *link.Link) error { return nil }
func (m *mockLinkRepo) FindByID(ctx context.Context, id uuid.UUID) (*link.Link, error) {
	return nil, nil
}

func (m *mockLinkRepo) FindBySource(ctx context.Context, sourceID uuid.UUID) ([]*link.Link, error) {
	return m.bySource[sourceID], nil
}

func (m *mockLinkRepo) FindByTarget(ctx context.Context, targetID uuid.UUID) ([]*link.Link, error) {
	return m.byTarget[targetID], nil
}

func (m *mockLinkRepo) Delete(ctx context.Context, id uuid.UUID) error               { return nil }
func (m *mockLinkRepo) DeleteBySource(ctx context.Context, sourceID uuid.UUID) error { return nil }
func (m *mockLinkRepo) FindAll(ctx context.Context) ([]*link.Link, error)            { return nil, nil }
func (m *mockLinkRepo) FindAllPaginated(ctx context.Context, limit, offset int) ([]*link.Link, int64, error) {
	return nil, 0, nil
}

func (m *mockLinkRepo) FindBySourceIDs(ctx context.Context, sourceIDs []uuid.UUID) (map[uuid.UUID][]*link.Link, error) {
	result := make(map[uuid.UUID][]*link.Link)
	for _, id := range sourceIDs {
		result[id] = m.bySource[id]
	}
	return result, nil
}

func (m *mockLinkRepo) FindByTargetIDs(ctx context.Context, targetIDs []uuid.UUID) (map[uuid.UUID][]*link.Link, error) {
	result := make(map[uuid.UUID][]*link.Link)
	for _, id := range targetIDs {
		result[id] = m.byTarget[id]
	}
	return result, nil
}

func newLink(source, target uuid.UUID, weight float64) *link.Link {
	lt, _ := link.NewLinkType("reference")
	w, _ := link.NewWeight(weight)
	md, _ := link.NewMetadata(nil)
	return link.NewLink(source, target, lt, w, md)
}

func TestNeighborLoader_GetNeighbors(t *testing.T) {
	ctx := context.Background()
	repo := newMockLinkRepo()

	nodeA := uuid.New()
	nodeB := uuid.New()
	nodeC := uuid.New()

	repo.addLink(newLink(nodeA, nodeB, 0.8))
	repo.addLink(newLink(nodeC, nodeA, 0.6))

	loader := NewNeighborLoader(repo, nil)
	edges, err := loader.GetNeighbors(ctx, nodeA)
	require.NoError(t, err)
	assert.Len(t, edges, 2)
}

func TestNeighborLoader_GetNeighborsBatch(t *testing.T) {
	ctx := context.Background()
	repo := newMockLinkRepo()

	nodeA := uuid.New()
	nodeB := uuid.New()

	repo.addLink(newLink(nodeA, nodeB, 0.8))

	loader := NewNeighborLoader(repo, nil)
	result, err := loader.GetNeighborsBatch(ctx, []uuid.UUID{nodeA, nodeB})
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Len(t, result[nodeA], 1)
	assert.Len(t, result[nodeB], 1)
}

func TestNeighborLoader_GetNeighborsBatch_Empty(t *testing.T) {
	ctx := context.Background()
	repo := newMockLinkRepo()

	loader := NewNeighborLoader(repo, nil)
	result, err := loader.GetNeighborsBatch(ctx, nil)
	require.NoError(t, err)
	assert.Empty(t, result)
}
