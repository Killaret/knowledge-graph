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
