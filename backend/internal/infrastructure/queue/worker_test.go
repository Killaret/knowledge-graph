package queue

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/infrastructure/nlp"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockNoteRepoForWorker struct{ mock.Mock }

func (m *mockNoteRepoForWorker) Save(ctx context.Context, note *note.Note) error { return nil }
func (m *mockNoteRepoForWorker) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*note.Note), args.Error(1)
}
func (m *mockNoteRepoForWorker) Delete(ctx context.Context, id uuid.UUID) error         { return nil }
func (m *mockNoteRepoForWorker) DeleteBatch(ctx context.Context, ids []uuid.UUID) error { return nil }
func (m *mockNoteRepoForWorker) Restore(ctx context.Context, id uuid.UUID) error        { return nil }
func (m *mockNoteRepoForWorker) List(ctx context.Context, limit, offset int) ([]*note.Note, int64, error) {
	return nil, 0, nil
}
func (m *mockNoteRepoForWorker) Search(ctx context.Context, query string, limit, offset int) ([]*note.Note, int64, error) {
	return nil, 0, nil
}
func (m *mockNoteRepoForWorker) FindAll(ctx context.Context) ([]*note.Note, error) { return nil, nil }
func (m *mockNoteRepoForWorker) FindAllPaginated(ctx context.Context, limit, offset int) ([]*note.Note, int64, error) {
	return nil, 0, nil
}

func TestWorker_HandleExtractKeywords_InvalidPayload(t *testing.T) {
	w := NewWorker(nil, nil, nil, nil)
	task := asynq.NewTask(TypeExtractKeywords, []byte("not json"))
	err := w.HandleExtractKeywords(context.Background(), task)
	assert.Error(t, err)
}

func TestWorker_HandleExtractKeywords_InvalidNoteID(t *testing.T) {
	w := NewWorker(nil, nil, nil, nil)
	payload := `{"note_id":"invalid-uuid"}`
	task := asynq.NewTask(TypeExtractKeywords, []byte(payload))
	err := w.HandleExtractKeywords(context.Background(), task)
	assert.Error(t, err)
}

func TestWorker_HandleExtractKeywords_NoteNotFound(t *testing.T) {
	repo := new(mockNoteRepoForWorker)
	noteID := uuid.New()
	repo.On("FindByID", mock.Anything, noteID).Return(nil, nil)

	w := NewWorker(repo, nil, nil, nil)
	payload := `{"note_id":"` + noteID.String() + `"}`
	task := asynq.NewTask(TypeExtractKeywords, []byte(payload))
	err := w.HandleExtractKeywords(context.Background(), task)
	assert.NoError(t, err)
}

func TestWorker_HandleExtractKeywords_NLPError(t *testing.T) {
	repo := new(mockNoteRepoForWorker)
	noteID := uuid.New()
	title, _ := note.NewTitle("Title")
	content, _ := note.NewContent("some content here")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, "star", metadata)

	repo.On("FindByID", mock.Anything, noteID).Return(n, nil)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	nlpClient := nlp.NewNLPClient(server.URL, nil, 0)
	w := NewWorker(repo, nil, nil, nlpClient)
	payload := `{"note_id":"` + noteID.String() + `"}`
	task := asynq.NewTask(TypeExtractKeywords, []byte(payload))
	err := w.HandleExtractKeywords(context.Background(), task)
	assert.Error(t, err)
}

func TestWorker_HandleComputeEmbedding_InvalidPayload(t *testing.T) {
	w := NewWorker(nil, nil, nil, nil)
	task := asynq.NewTask(TypeComputeEmbedding, []byte("not json"))
	err := w.HandleComputeEmbedding(context.Background(), task)
	assert.Error(t, err)
}

func TestWorker_HandleComputeEmbedding_InvalidNoteID(t *testing.T) {
	w := NewWorker(nil, nil, nil, nil)
	payload := `{"note_id":"invalid-uuid"}`
	task := asynq.NewTask(TypeComputeEmbedding, []byte(payload))
	err := w.HandleComputeEmbedding(context.Background(), task)
	assert.Error(t, err)
}

func TestWorker_HandleComputeEmbedding_NoteNotFound(t *testing.T) {
	repo := new(mockNoteRepoForWorker)
	noteID := uuid.New()
	repo.On("FindByID", mock.Anything, noteID).Return(nil, nil)

	w := NewWorker(repo, nil, nil, nil)
	payload := `{"note_id":"` + noteID.String() + `"}`
	task := asynq.NewTask(TypeComputeEmbedding, []byte(payload))
	err := w.HandleComputeEmbedding(context.Background(), task)
	assert.NoError(t, err)
}

func TestWorker_HandleComputeEmbedding_NLPError(t *testing.T) {
	repo := new(mockNoteRepoForWorker)
	noteID := uuid.New()
	title, _ := note.NewTitle("Title")
	content, _ := note.NewContent("some content here")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, "star", metadata)

	repo.On("FindByID", mock.Anything, noteID).Return(n, nil)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	nlpClient := nlp.NewNLPClient(server.URL, nil, 0)
	w := NewWorker(repo, nil, nil, nlpClient)
	payload := `{"note_id":"` + noteID.String() + `"}`
	task := asynq.NewTask(TypeComputeEmbedding, []byte(payload))
	err := w.HandleComputeEmbedding(context.Background(), task)
	assert.Error(t, err)
}
