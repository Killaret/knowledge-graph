package notehandler

import (
	"context"
	"sync"

	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
)

type mockNoteRepo struct {
	mu    sync.RWMutex
	notes map[uuid.UUID]*note.Note
}

func newMockNoteRepo() *mockNoteRepo {
	return &mockNoteRepo{
		notes: make(map[uuid.UUID]*note.Note),
	}
}

func (m *mockNoteRepo) Save(ctx context.Context, n *note.Note) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notes[n.ID()] = n
	return nil
}

func (m *mockNoteRepo) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n, ok := m.notes[id]
	if !ok {
		return nil, nil
	}
	return n, nil
}

func (m *mockNoteRepo) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.notes, id)
	return nil
}

func (m *mockNoteRepo) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, id := range ids {
		delete(m.notes, id)
	}
	return nil
}

func (m *mockNoteRepo) Restore(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.notes[id]; !ok {
		return note.ErrNoteNotFound
	}
	return nil
}

func (m *mockNoteRepo) FindAll(ctx context.Context) ([]*note.Note, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	notes := make([]*note.Note, 0, len(m.notes))
	for _, n := range m.notes {
		notes = append(notes, n)
	}
	return notes, nil
}

func (m *mockNoteRepo) Search(ctx context.Context, userID uuid.UUID, query string, limit, offset int) ([]*note.Note, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*note.Note
	for _, n := range m.notes {
		if contains(n.Title().String(), query) || contains(n.Content().String(), query) {
			results = append(results, n)
		}
	}

	// Apply pagination
	total := int64(len(results))
	if offset >= len(results) {
		return []*note.Note{}, total, nil
	}

	end := offset + limit
	if end > len(results) {
		end = len(results)
	}

	return results[offset:end], total, nil
}

func (m *mockNoteRepo) FindByKeywords(ctx context.Context, keywords []string) ([]*note.Note, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*note.Note
	for _, n := range m.notes {
		for _, keyword := range keywords {
			if contains(n.Title().String(), keyword) || contains(n.Content().String(), keyword) {
				results = append(results, n)
				break
			}
		}
	}
	return results, nil
}

func (m *mockNoteRepo) Update(ctx context.Context, n *note.Note) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.notes[n.ID()]; !exists {
		return note.ErrNoteNotFound
	}

	m.notes[n.ID()] = n
	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				findSubstring(s, substr))))
}

func (m *mockNoteRepo) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*note.Note, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var allNotes []*note.Note
	for _, n := range m.notes {
		allNotes = append(allNotes, n)
	}

	total := int64(len(allNotes))

	if offset >= len(allNotes) {
		return []*note.Note{}, total, nil
	}

	end := offset + limit
	if limit <= 0 || end > len(allNotes) || end < offset {
		end = len(allNotes)
	}

	return allNotes[offset:end], total, nil
}

func (m *mockNoteRepo) FindAllPaginated(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*note.Note, int64, error) {
	return m.List(ctx, userID, limit, offset)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

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
	for _, existing := range m.links {
		if existing.SourceNoteID() == l.SourceNoteID() &&
			existing.TargetNoteID() == l.TargetNoteID() &&
			existing.LinkType().String() == l.LinkType().String() {
			return link.ErrDuplicateLink
		}
	}
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

func (m *mockLinkRepo) Update(ctx context.Context, l *link.Link) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.links[l.ID()]; !ok {
		return link.ErrLinkNotFound
	}
	m.links[l.ID()] = l
	return nil
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

func (m *mockLinkRepo) FindAll(ctx context.Context) ([]*link.Link, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*link.Link
	for _, l := range m.links {
		result = append(result, l)
	}
	return result, nil
}

func (m *mockLinkRepo) FindAllPaginated(ctx context.Context, limit, offset int) ([]*link.Link, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	all := make([]*link.Link, 0, len(m.links))
	for _, l := range m.links {
		all = append(all, l)
	}
	total := int64(len(all))
	if offset >= len(all) {
		return []*link.Link{}, total, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], total, nil
}
