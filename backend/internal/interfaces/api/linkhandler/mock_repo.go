package linkhandler

import (
	"context"
	"sync"

	"knowledge-graph/internal/domain/link"

	"github.com/google/uuid"
)

type mockLinkRepo struct {
	mu    sync.RWMutex
	links map[uuid.UUID]*link.Link
}

func newMockLinkRepo() *mockLinkRepo {
	return &mockLinkRepo{
		links: make(map[uuid.UUID]*link.Link),
	}
}

func (m *mockLinkRepo) Save(ctx context.Context, l *link.Link) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.links[l.ID()] = l
	return nil
}

func (m *mockLinkRepo) FindByID(ctx context.Context, id uuid.UUID) (*link.Link, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	l, ok := m.links[id]
	if !ok {
		return nil, nil
	}
	return l, nil
}

func (m *mockLinkRepo) FindBySource(ctx context.Context, sourceID uuid.UUID) ([]*link.Link, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*link.Link
	for _, l := range m.links {
		if l.SourceNoteID() == sourceID {
			result = append(result, l)
		}
	}
	return result, nil
}

func (m *mockLinkRepo) FindByTarget(ctx context.Context, targetID uuid.UUID) ([]*link.Link, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*link.Link
	for _, l := range m.links {
		if l.TargetNoteID() == targetID {
			result = append(result, l)
		}
	}
	return result, nil
}

func (m *mockLinkRepo) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.links, id)
	return nil
}

func (m *mockLinkRepo) DeleteBySource(ctx context.Context, sourceID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, l := range m.links {
		if l.SourceNoteID() == sourceID {
			delete(m.links, id)
		}
	}
	return nil
}

func (m *mockLinkRepo) List(ctx context.Context) ([]*link.Link, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*link.Link
	for _, l := range m.links {
		result = append(result, l)
	}
	return result, nil
}

func (m *mockLinkRepo) FindAll(ctx context.Context) ([]*link.Link, error) {
	return m.List(ctx)
}

func (m *mockLinkRepo) FindAllPaginated(ctx context.Context, limit, offset int) ([]*link.Link, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var allLinks []*link.Link
	for _, l := range m.links {
		allLinks = append(allLinks, l)
	}

	total := int64(len(allLinks))

	if offset >= len(allLinks) {
		return []*link.Link{}, total, nil
	}

	end := offset + limit
	if end > len(allLinks) {
		end = len(allLinks)
	}
	if limit == 0 {
		end = len(allLinks)
	}

	return allLinks[offset:end], total, nil
}
