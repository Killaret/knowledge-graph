package notehandler

import (
	"context"
	"sync"

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

func (m *mockNoteRepo) FindAll(ctx context.Context) ([]*note.Note, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	notes := make([]*note.Note, 0, len(m.notes))
	for _, n := range m.notes {
		notes = append(notes, n)
	}
	return notes, nil
}

func (m *mockNoteRepo) Search(ctx context.Context, query string, limit, offset int) ([]*note.Note, int64, error) {
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

func (m *mockNoteRepo) List(ctx context.Context, limit, offset int) ([]*note.Note, int64, error) {
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
	if end > len(allNotes) {
		end = len(allNotes)
	}

	return allNotes[offset:end], total, nil
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
