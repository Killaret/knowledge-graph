package queue

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	importer "knowledge-graph/internal/application/import"
	"knowledge-graph/internal/domain/cache/cachetest"
	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type importNoteRepo struct {
	mu    sync.RWMutex
	notes map[uuid.UUID]*note.Note
}

func newImportNoteRepo() *importNoteRepo {
	return &importNoteRepo{notes: make(map[uuid.UUID]*note.Note)}
}

func (r *importNoteRepo) Save(ctx context.Context, n *note.Note) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.notes[n.ID()] = n
	return nil
}

func (r *importNoteRepo) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.notes[id], nil
}

func (r *importNoteRepo) Delete(ctx context.Context, id uuid.UUID) error         { return nil }
func (r *importNoteRepo) DeleteBatch(ctx context.Context, ids []uuid.UUID) error { return nil }
func (r *importNoteRepo) Restore(ctx context.Context, id uuid.UUID) error        { return nil }

func (r *importNoteRepo) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*note.Note, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var all []*note.Note
	for _, n := range r.notes {
		all = append(all, n)
	}
	total := int64(len(all))
	if offset >= len(all) {
		return nil, total, nil
	}
	end := offset + limit
	if limit <= 0 || end > len(all) || end < offset {
		end = len(all)
	}
	return all[offset:end], total, nil
}

func (r *importNoteRepo) Search(ctx context.Context, userID uuid.UUID, query string, limit, offset int) ([]*note.Note, int64, error) {
	return r.List(ctx, userID, limit, offset)
}

func (r *importNoteRepo) FindAll(ctx context.Context) ([]*note.Note, error) {
	notes, _, err := r.List(ctx, uuid.Nil, 0, 0)
	return notes, err
}

func (r *importNoteRepo) FindAllPaginated(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*note.Note, int64, error) {
	return r.List(ctx, userID, limit, offset)
}

func TestWorker_HandleImportBookmarks_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	repo := newImportNoteRepo()
	cache := cachetest.NewFakeCacheClient()
	importSvc := importer.NewService(repo, cache, nil)
	w := NewWorker(nil, nil, nil, nil, cache, importSvc)

	items := []importer.Item{
		{Title: "One", URL: "https://example.com/one", Type: "asteroid"},
		{Title: "Two", URL: "https://example.com/two", Type: "planet"},
	}
	itemsJSON, err := json.Marshal(items)
	require.NoError(t, err)

	payload := ImportBookmarksPayload{
		TaskID: uuid.New().String(),
		UserID: userID.String(),
		Items:  itemsJSON,
	}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(TypeImportBookmarks, payloadBytes)
	err = w.HandleImportBookmarks(ctx, task)
	require.NoError(t, err)

	require.Len(t, repo.notes, 2)

	status, err := importSvc.GetTaskStatus(ctx, payload.TaskID)
	require.NoError(t, err)
	assert.Equal(t, payload.TaskID, status.TaskID)
	assert.Equal(t, "done", status.Status)
	assert.Equal(t, 2, status.Progress.Created)
}

func TestWorker_HandleImportBookmarks_ImportServiceNil(t *testing.T) {
	w := NewWorker(nil, nil, nil, nil, nil, nil)
	task := asynq.NewTask(TypeImportBookmarks, []byte("{}"))
	err := w.HandleImportBookmarks(context.Background(), task)
	assert.Error(t, err)
}

func TestWorker_HandleImportBookmarks_InvalidPayload(t *testing.T) {
	w := NewWorker(nil, nil, nil, nil, nil, nil)
	task := asynq.NewTask(TypeImportBookmarks, []byte("not json"))
	err := w.HandleImportBookmarks(context.Background(), task)
	assert.Error(t, err)
}
