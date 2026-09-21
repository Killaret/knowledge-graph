package queue

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/infrastructure/nlp"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
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
func (m *mockNoteRepoForWorker) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*note.Note, int64, error) {
	return nil, 0, nil
}
func (m *mockNoteRepoForWorker) Search(ctx context.Context, userID uuid.UUID, query string, limit, offset int) ([]*note.Note, int64, error) {
	return nil, 0, nil
}
func (m *mockNoteRepoForWorker) FindAll(ctx context.Context) ([]*note.Note, error) { return nil, nil }
func (m *mockNoteRepoForWorker) FindAllPaginated(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*note.Note, int64, error) {
	return nil, 0, nil
}

func TestWorker_HandleExtractKeywords_InvalidPayload(t *testing.T) {
	w := NewWorker(nil, nil, nil, nil, nil, nil, nil, nil, nil, 0)
	task := asynq.NewTask(TypeExtractKeywords, []byte("not json"))
	err := w.HandleExtractKeywords(context.Background(), task)
	assert.Error(t, err)
}

func TestWorker_HandleExtractKeywords_InvalidNoteID(t *testing.T) {
	w := NewWorker(nil, nil, nil, nil, nil, nil, nil, nil, nil, 0)
	payload := `{"note_id":"invalid-uuid"}`
	task := asynq.NewTask(TypeExtractKeywords, []byte(payload))
	err := w.HandleExtractKeywords(context.Background(), task)
	assert.Error(t, err)
}

func TestWorker_HandleExtractKeywords_NoteNotFound(t *testing.T) {
	repo := new(mockNoteRepoForWorker)
	noteID := uuid.New()
	repo.On("FindByID", mock.Anything, noteID).Return(nil, nil)

	w := NewWorker(repo, nil, nil, nil, nil, nil, nil, nil, nil, 0)
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
	n := note.NewNote(title, content, note.MustType("star"), metadata)

	repo.On("FindByID", mock.Anything, noteID).Return(n, nil)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	nlpClient := nlp.NewNLPClient(server.URL, nil, 0)
	w := NewWorker(repo, nil, nil, nlpClient, nil, nil, nil, nil, nil, 0)
	payload := `{"note_id":"` + noteID.String() + `"}`
	task := asynq.NewTask(TypeExtractKeywords, []byte(payload))
	err := w.HandleExtractKeywords(context.Background(), task)
	assert.Error(t, err)
}

// NLP-2: the worker must persist the lemma in keyword, the surface form and
// the extractor name reported by the NLP service.
func TestWorker_HandleExtractKeywords_PersistsLemmaSurfaceExtractor(t *testing.T) {
	repo := new(mockNoteRepoForWorker)
	noteID := uuid.New()
	title, _ := note.NewTitle("Title")
	content, _ := note.NewContent("some content here")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, note.MustType("star"), metadata)

	repo.On("FindByID", mock.Anything, noteID).Return(n, nil)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{
			"extractor": "keybert-hybrid-0.9",
			"keywords": [
				{"keyword": "дерево", "surface": "деревья", "weight": 0.8}
			]
		}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	sqlDB, sqlMock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()
	gormDB, err := gorm.Open(gormpostgres.New(gormpostgres.Config{Conn: sqlDB}), &gorm.Config{})
	require.NoError(t, err)
	keywordRepo := postgres.NewKeywordRepository(gormDB)

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(`DELETE FROM "note_keywords" WHERE note_id = \$1`).
		WithArgs(noteID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	sqlMock.ExpectExec(`INSERT INTO "note_keywords"`).
		WithArgs(noteID, "дерево", "деревья", "keybert-hybrid-0.9", 0.8).
		WillReturnResult(sqlmock.NewResult(0, 1))
	sqlMock.ExpectCommit()

	nlpClient := nlp.NewNLPClient(server.URL, nil, 0)
	w := NewWorker(repo, keywordRepo, nil, nlpClient, nil, nil, nil, nil, nil, 0)
	payload := `{"note_id":"` + noteID.String() + `"}`
	task := asynq.NewTask(TypeExtractKeywords, []byte(payload))
	require.NoError(t, w.HandleExtractKeywords(context.Background(), task))
	require.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestWorker_HandleComputeEmbedding_InvalidPayload(t *testing.T) {
	w := NewWorker(nil, nil, nil, nil, nil, nil, nil, nil, nil, 0)
	task := asynq.NewTask(TypeComputeEmbedding, []byte("not json"))
	err := w.HandleComputeEmbedding(context.Background(), task)
	assert.Error(t, err)
}

func TestWorker_HandleComputeEmbedding_InvalidNoteID(t *testing.T) {
	w := NewWorker(nil, nil, nil, nil, nil, nil, nil, nil, nil, 0)
	payload := `{"note_id":"invalid-uuid"}`
	task := asynq.NewTask(TypeComputeEmbedding, []byte(payload))
	err := w.HandleComputeEmbedding(context.Background(), task)
	assert.Error(t, err)
}

func TestWorker_HandleComputeEmbedding_NoteNotFound(t *testing.T) {
	repo := new(mockNoteRepoForWorker)
	noteID := uuid.New()
	repo.On("FindByID", mock.Anything, noteID).Return(nil, nil)

	w := NewWorker(repo, nil, nil, nil, nil, nil, nil, nil, nil, 0)
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
	n := note.NewNote(title, content, note.MustType("star"), metadata)

	repo.On("FindByID", mock.Anything, noteID).Return(n, nil)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	nlpClient := nlp.NewNLPClient(server.URL, nil, 0)
	w := NewWorker(repo, nil, nil, nlpClient, nil, nil, nil, nil, nil, 0)
	payload := `{"note_id":"` + noteID.String() + `"}`
	task := asynq.NewTask(TypeComputeEmbedding, []byte(payload))
	err := w.HandleComputeEmbedding(context.Background(), task)
	assert.Error(t, err)
}
