package taghandler

import (
	"context"
	"strings"
	"sync"

	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/domain/tag"

	"github.com/google/uuid"
)

// mockTagRepo is a thread-safe in-memory implementation of tag.Repository.
type mockTagRepo struct {
	mu       sync.RWMutex
	tags     map[uuid.UUID]*tag.Tag
	byName   map[string]uuid.UUID
	noteTags map[uuid.UUID]map[uuid.UUID]bool
}

func newMockTagRepo() *mockTagRepo {
	return &mockTagRepo{
		tags:     make(map[uuid.UUID]*tag.Tag),
		byName:   make(map[string]uuid.UUID),
		noteTags: make(map[uuid.UUID]map[uuid.UUID]bool),
	}
}

func (m *mockTagRepo) Create(ctx context.Context, t *tag.Tag) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tags[t.ID()] = t
	m.byName[t.Name()] = t.ID()
	return nil
}

func (m *mockTagRepo) FindByID(ctx context.Context, id uuid.UUID) (*tag.Tag, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tags[id], nil
}

func (m *mockTagRepo) FindByName(ctx context.Context, name string) (*tag.Tag, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if id, ok := m.byName[name]; ok {
		return m.tags[id], nil
	}
	return nil, nil
}

func (m *mockTagRepo) FindAll(ctx context.Context) ([]*tag.Tag, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*tag.Tag, 0, len(m.tags))
	for _, t := range m.tags {
		result = append(result, t)
	}
	return result, nil
}

func (m *mockTagRepo) Update(ctx context.Context, t *tag.Tag) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	old, ok := m.tags[t.ID()]
	if !ok {
		return nil
	}
	delete(m.byName, old.Name())
	m.tags[t.ID()] = t
	m.byName[t.Name()] = t.ID()
	return nil
}

func (m *mockTagRepo) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.tags[id]; ok {
		delete(m.byName, t.Name())
		delete(m.tags, id)
	}
	return nil
}

func (m *mockTagRepo) AddTagToNote(ctx context.Context, noteID, tagID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.noteTags[noteID] == nil {
		m.noteTags[noteID] = make(map[uuid.UUID]bool)
	}
	m.noteTags[noteID][tagID] = true
	return nil
}

func (m *mockTagRepo) RemoveTagFromNote(ctx context.Context, noteID, tagID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.noteTags[noteID] != nil {
		delete(m.noteTags[noteID], tagID)
	}
	return nil
}

func (m *mockTagRepo) GetTagsByNoteID(ctx context.Context, noteID uuid.UUID) ([]*tag.Tag, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tagIDs, ok := m.noteTags[noteID]
	if !ok {
		return []*tag.Tag{}, nil
	}

	result := make([]*tag.Tag, 0, len(tagIDs))
	for id := range tagIDs {
		if t, ok := m.tags[id]; ok {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockTagRepo) IsTagAssignedToNote(ctx context.Context, noteID, tagID uuid.UUID) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if tags, ok := m.noteTags[noteID]; ok {
		return tags[tagID], nil
	}
	return false, nil
}

// mockNoteRepo is a minimal in-memory implementation of note.Repository.
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
	return m.notes[id], nil
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
	return nil
}

func (m *mockNoteRepo) List(ctx context.Context, limit, offset int) ([]*note.Note, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*note.Note, 0, len(m.notes))
	for _, n := range m.notes {
		result = append(result, n)
	}
	return result, int64(len(result)), nil
}

func (m *mockNoteRepo) Search(ctx context.Context, query string, limit, offset int) ([]*note.Note, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*note.Note
	for _, n := range m.notes {
		if strings.Contains(n.Title().String(), query) || strings.Contains(n.Content().String(), query) {
			result = append(result, n)
		}
	}
	return result, int64(len(result)), nil
}

func (m *mockNoteRepo) FindAll(ctx context.Context) ([]*note.Note, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*note.Note, 0, len(m.notes))
	for _, n := range m.notes {
		result = append(result, n)
	}
	return result, nil
}

func (m *mockNoteRepo) FindAllPaginated(ctx context.Context, limit, offset int) ([]*note.Note, int64, error) {
	return m.List(ctx, limit, offset)
}
