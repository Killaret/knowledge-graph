package graphhandler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type mockNoteRepo struct {
	mock.Mock
}

func (m *mockNoteRepo) Save(ctx context.Context, n *note.Note) error {
	args := m.Called(ctx, n)
	return args.Error(0)
}

func (m *mockNoteRepo) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*note.Note), args.Error(1)
}

func (m *mockNoteRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockNoteRepo) List(ctx context.Context, limit, offset int) ([]*note.Note, int64, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*note.Note), args.Get(1).(int64), args.Error(2)
}

func (m *mockNoteRepo) Search(ctx context.Context, query string, limit, offset int) ([]*note.Note, int64, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]*note.Note), args.Get(1).(int64), args.Error(2)
}

func (m *mockNoteRepo) FindAll(ctx context.Context) ([]*note.Note, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*note.Note), args.Error(1)
}

type mockLinkRepo struct {
	mock.Mock
}

func (m *mockLinkRepo) Save(ctx context.Context, l *link.Link) error {
	args := m.Called(ctx, l)
	return args.Error(0)
}

func (m *mockLinkRepo) FindByID(ctx context.Context, id uuid.UUID) (*link.Link, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*link.Link), args.Error(1)
}

func (m *mockLinkRepo) FindBySource(ctx context.Context, sourceID uuid.UUID) ([]*link.Link, error) {
	args := m.Called(ctx, sourceID)
	return args.Get(0).([]*link.Link), args.Error(1)
}

func (m *mockLinkRepo) FindByTarget(ctx context.Context, targetID uuid.UUID) ([]*link.Link, error) {
	args := m.Called(ctx, targetID)
	return args.Get(0).([]*link.Link), args.Error(1)
}

func (m *mockLinkRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockLinkRepo) FindAll(ctx context.Context) ([]*link.Link, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*link.Link), args.Error(1)
}

func (m *mockLinkRepo) DeleteBySource(ctx context.Context, sourceID uuid.UUID) error {
	args := m.Called(ctx, sourceID)
	return args.Error(0)
}

func setupGraphRouter() (*gin.Engine, *mockNoteRepo, *mockLinkRepo) {
	gin.SetMode(gin.TestMode)
	noteRepo := new(mockNoteRepo)
	linkRepo := new(mockLinkRepo)
	handler := New(noteRepo, linkRepo, 3)
	r := gin.Default()
	r.GET("/graph/:id", handler.GetGraph)
	r.GET("/graph", handler.GetFullGraph)
	return r, noteRepo, linkRepo
}

func TestHandler_GetGraph(t *testing.T) {
	t.Run("successful graph load", func(t *testing.T) {
		r, noteRepo, linkRepo := setupGraphRouter()

		centerID := uuid.New()
		title, _ := note.NewTitle("Center Node")
		content, _ := note.NewContent("Center content")
		metadata, _ := note.NewMetadata(map[string]interface{}{"type": "star"})
		centerNote := note.NewNote(title, content, "star", metadata)

		noteRepo.On("FindByID", mock.Anything, centerID).Return(centerNote, nil)
		linkRepo.On("FindBySource", mock.Anything, centerID).Return([]*link.Link{}, nil)
		linkRepo.On("FindByTarget", mock.Anything, centerID).Return([]*link.Link{}, nil)

		req := httptest.NewRequest("GET", "/graph/"+centerID.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response GraphData
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Nodes, 1)
		assert.Equal(t, "Center Node", response.Nodes[0].Title)
		assert.Equal(t, "star", response.Nodes[0].Type)
	})

	t.Run("invalid id - should return 400", func(t *testing.T) {
		r, _, _ := setupGraphRouter()

		req := httptest.NewRequest("GET", "/graph/invalid-uuid", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("graph with connected nodes", func(t *testing.T) {
		r, noteRepo, linkRepo := setupGraphRouter()

		centerID := uuid.New()
		neighborID := uuid.New()

		title1, _ := note.NewTitle("Center")
		content1, _ := note.NewContent("Center content")
		metadata1, _ := note.NewMetadata(map[string]interface{}{"type": "star"})
		centerNote := note.NewNote(title1, content1, "star", metadata1)

		title2, _ := note.NewTitle("Neighbor")
		content2, _ := note.NewContent("Neighbor content")
		metadata2, _ := note.NewMetadata(map[string]interface{}{"type": "planet"})
		neighborNote := note.NewNote(title2, content2, "planet", metadata2)

		linkType, _ := link.NewLinkType("reference")
		weight, _ := link.NewWeight(0.8)
		linkMetadata, _ := link.NewMetadata(nil)
		l := link.NewLink(centerID, neighborID, linkType, weight, linkMetadata)

		noteRepo.On("FindByID", mock.Anything, centerID).Return(centerNote, nil)
		noteRepo.On("FindByID", mock.Anything, neighborID).Return(neighborNote, nil)
		linkRepo.On("FindBySource", mock.Anything, centerID).Return([]*link.Link{l}, nil)
		linkRepo.On("FindByTarget", mock.Anything, centerID).Return([]*link.Link{}, nil)
		linkRepo.On("FindBySource", mock.Anything, neighborID).Return([]*link.Link{}, nil)
		linkRepo.On("FindByTarget", mock.Anything, neighborID).Return([]*link.Link{}, nil)

		req := httptest.NewRequest("GET", "/graph/"+centerID.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response GraphData
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Nodes, 2)
		assert.Len(t, response.Links, 1)
		// Check that both node types exist (order-independent)
		types := make([]string, len(response.Nodes))
		for i, node := range response.Nodes {
			types[i] = node.Type
		}
		assert.Contains(t, types, "star")
		assert.Contains(t, types, "planet")
	})

	t.Run("custom depth parameter", func(t *testing.T) {
		r, noteRepo, linkRepo := setupGraphRouter()

		centerID := uuid.New()
		title, _ := note.NewTitle("Center")
		content, _ := note.NewContent("Content")
		metadata, _ := note.NewMetadata(nil)
		centerNote := note.NewNote(title, content, "star", metadata)

		noteRepo.On("FindByID", mock.Anything, centerID).Return(centerNote, nil)
		linkRepo.On("FindBySource", mock.Anything, centerID).Return([]*link.Link{}, nil)
		linkRepo.On("FindByTarget", mock.Anything, centerID).Return([]*link.Link{}, nil)

		req := httptest.NewRequest("GET", "/graph/"+centerID.String()+"?depth=2", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("depth exceeding max should be capped", func(t *testing.T) {
		r, noteRepo, linkRepo := setupGraphRouter()

		centerID := uuid.New()
		title, _ := note.NewTitle("Center")
		content, _ := note.NewContent("Content")
		metadata, _ := note.NewMetadata(nil)
		centerNote := note.NewNote(title, content, "star", metadata)

		noteRepo.On("FindByID", mock.Anything, centerID).Return(centerNote, nil)
		linkRepo.On("FindBySource", mock.Anything, centerID).Return([]*link.Link{}, nil)
		linkRepo.On("FindByTarget", mock.Anything, centerID).Return([]*link.Link{}, nil)

		// Request depth=10 but max is 3, should be capped at 3
		req := httptest.NewRequest("GET", "/graph/"+centerID.String()+"?depth=10", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHandler_GetFullGraph(t *testing.T) {
	t.Run("successful full graph load", func(t *testing.T) {
		r, noteRepo, linkRepo := setupGraphRouter()

		note1ID := uuid.New()
		note2ID := uuid.New()

		title1, _ := note.NewTitle("Note 1")
		content1, _ := note.NewContent("Content 1")
		metadata1, _ := note.NewMetadata(map[string]interface{}{"type": "star"})
		n1 := note.NewNote(title1, content1, "star", metadata1)

		title2, _ := note.NewTitle("Note 2")
		content2, _ := note.NewContent("Content 2")
		metadata2, _ := note.NewMetadata(map[string]interface{}{"type": "planet"})
		n2 := note.NewNote(title2, content2, "planet", metadata2)

		linkType, _ := link.NewLinkType("reference")
		weight, _ := link.NewWeight(1.0)
		linkMetadata, _ := link.NewMetadata(nil)
		l := link.NewLink(note1ID, note2ID, linkType, weight, linkMetadata)

		noteRepo.On("FindAll", mock.Anything).Return([]*note.Note{n1, n2}, nil)
		linkRepo.On("FindAll", mock.Anything).Return([]*link.Link{l}, nil)

		req := httptest.NewRequest("GET", "/graph", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response GraphData
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Nodes, 2)
		assert.Len(t, response.Links, 1)
	})

	t.Run("with limit parameter", func(t *testing.T) {
		r, noteRepo, linkRepo := setupGraphRouter()

		title1, _ := note.NewTitle("Note 1")
		content1, _ := note.NewContent("Content 1")
		metadata1, _ := note.NewMetadata(nil)
		n1 := note.NewNote(title1, content1, "star", metadata1)

		noteRepo.On("FindAll", mock.Anything).Return([]*note.Note{n1}, nil)
		linkRepo.On("FindAll", mock.Anything).Return([]*link.Link{}, nil)

		req := httptest.NewRequest("GET", "/graph?limit=1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response GraphData
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Nodes, 1)
	})

	t.Run("database error - should return 500", func(t *testing.T) {
		r, noteRepo, _ := setupGraphRouter()

		noteRepo.On("FindAll", mock.Anything).Return([]*note.Note{}, errors.New("db error"))

		req := httptest.NewRequest("GET", "/graph", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestNew(t *testing.T) {
	noteRepo := new(mockNoteRepo)
	linkRepo := new(mockLinkRepo)

	handler := New(noteRepo, linkRepo, 5)

	assert.NotNil(t, handler)
}
