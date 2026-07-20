package taghandler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"knowledge-graph/internal/domain/note"
	tagDomain "knowledge-graph/internal/domain/tag"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestNote(t *testing.T, repo *mockNoteRepo, title string) *note.Note {
	noteTitle, err := note.NewTitle(title)
	require.NoError(t, err)
	content, err := note.NewContent("Test content for " + title)
	require.NoError(t, err)
	metadata, err := note.NewMetadata(map[string]interface{}{})
	require.NoError(t, err)
	n := note.NewNote(noteTitle, content, "star", metadata)
	err = repo.Save(nil, n)
	require.NoError(t, err)
	return n
}

func createTestTag(t *testing.T, repo *mockTagRepo, name string) *tagDomain.Tag {
	tag, err := tagDomain.New(name)
	require.NoError(t, err)
	err = repo.Create(nil, tag)
	require.NoError(t, err)
	return tag
}

func setupTagRouter() (*gin.Engine, *mockTagRepo, *mockNoteRepo) {
	gin.SetMode(gin.TestMode)
	tagRepo := newMockTagRepo()
	noteRepo := newMockNoteRepo()
	handler := New(tagRepo, noteRepo)

	r := gin.New()
	r.POST("/tags", handler.Create)
	r.GET("/tags", handler.List)
	r.GET("/tags/:id", handler.Get)
	r.PUT("/tags/:id", handler.Update)
	r.DELETE("/tags/:id", handler.Delete)
	r.POST("/notes/:id/tags", handler.AddTagToNote)
	r.DELETE("/notes/:id/tags/:tagId", handler.RemoveTagFromNote)
	r.GET("/notes/:id/tags", handler.GetTagsByNote)

	return r, tagRepo, noteRepo
}

func TestHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       map[string]interface{}
		wantStatus int
		wantName   string
		wantErr    bool
	}{
		{
			name:       "valid tag",
			body:       map[string]interface{}{"name": "golang"},
			wantStatus: http.StatusCreated,
			wantName:   "golang",
			wantErr:    false,
		},
		{
			name:       "tag with dashes",
			body:       map[string]interface{}{"name": "my-awesome_tag"},
			wantStatus: http.StatusCreated,
			wantName:   "my-awesome_tag",
			wantErr:    false,
		},
		{
			name:       "empty name",
			body:       map[string]interface{}{"name": ""},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name:       "invalid json",
			body:       map[string]interface{}{},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name:       "name too long",
			body:       map[string]interface{}{"name": string(make([]byte, 51))},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _, _ := setupTagRouter()

			jsonBody, err := json.Marshal(tt.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/tags", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if !tt.wantErr {
				var resp struct {
					Data TagResponse `json:"data"`
				}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.NotEmpty(t, resp.Data.ID)
				assert.Equal(t, tt.wantName, resp.Data.Name)
			}
		})
	}

	t.Run("duplicate name", func(t *testing.T) {
		r, tagRepo, _ := setupTagRouter()
		_ = createTestTag(t, tagRepo, "unique")

		body := map[string]interface{}{"name": "unique"}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/tags", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestHandler_List(t *testing.T) {
	r, tagRepo, _ := setupTagRouter()
	createTestTag(t, tagRepo, "go")
	createTestTag(t, tagRepo, "python")

	req := httptest.NewRequest(http.MethodGet, "/tags", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Data []TagResponse `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 2)
}

func TestHandler_Get(t *testing.T) {
	r, tagRepo, _ := setupTagRouter()
	tag := createTestTag(t, tagRepo, "my-tag")

	tests := []struct {
		name       string
		id         string
		wantStatus int
		wantName   string
	}{
		{"found", tag.ID().String(), http.StatusOK, "my-tag"},
		{"not found", uuid.New().String(), http.StatusNotFound, ""},
		{"invalid id", "not-uuid", http.StatusBadRequest, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/tags/"+tt.id, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var resp struct {
					Data TagResponse `json:"data"`
				}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, tt.wantName, resp.Data.Name)
			}
		})
	}
}

func TestHandler_Update(t *testing.T) {
	r, tagRepo, _ := setupTagRouter()
	tag := createTestTag(t, tagRepo, "old-name")

	t.Run("success", func(t *testing.T) {
		body := map[string]interface{}{"name": "new-name"}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/tags/"+tag.ID().String(), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data TagResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "new-name", resp.Data.Name)
	})

	t.Run("duplicate", func(t *testing.T) {
		_ = createTestTag(t, tagRepo, "taken")

		body := map[string]interface{}{"name": "taken"}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/tags/"+tag.ID().String(), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		body := map[string]interface{}{"name": "name"}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/tags/"+uuid.New().String(), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestHandler_Delete(t *testing.T) {
	r, tagRepo, _ := setupTagRouter()
	tag := createTestTag(t, tagRepo, "delete-me")

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/tags/"+tag.ID().String(), nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		found, err := tagRepo.FindByID(nil, tag.ID())
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/tags/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestHandler_AddTagToNote(t *testing.T) {
	r, tagRepo, noteRepo := setupTagRouter()
	n := createTestNote(t, noteRepo, "note")
	tag := createTestTag(t, tagRepo, "important")

	t.Run("success", func(t *testing.T) {
		body := map[string]interface{}{"tag_id": tag.ID().String()}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/notes/"+n.ID().String()+"/tags", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("already assigned", func(t *testing.T) {
		err := tagRepo.AddTagToNote(nil, n.ID(), tag.ID())
		require.NoError(t, err)

		body := map[string]interface{}{"tag_id": tag.ID().String()}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/notes/"+n.ID().String()+"/tags", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("note not found", func(t *testing.T) {
		body := map[string]interface{}{"tag_id": tag.ID().String()}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/notes/"+uuid.New().String()+"/tags", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("tag not found", func(t *testing.T) {
		body := map[string]interface{}{"tag_id": uuid.New().String()}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/notes/"+n.ID().String()+"/tags", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestHandler_RemoveTagFromNote(t *testing.T) {
	r, tagRepo, noteRepo := setupTagRouter()
	n := createTestNote(t, noteRepo, "note")
	tag := createTestTag(t, tagRepo, "removable")
	err := tagRepo.AddTagToNote(nil, n.ID(), tag.ID())
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/notes/"+n.ID().String()+"/tags/"+tag.ID().String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestHandler_GetTagsByNote(t *testing.T) {
	r, tagRepo, noteRepo := setupTagRouter()
	n := createTestNote(t, noteRepo, "note")
	tag1 := createTestTag(t, tagRepo, "t1")
	tag2 := createTestTag(t, tagRepo, "t2")
	err := tagRepo.AddTagToNote(nil, n.ID(), tag1.ID())
	require.NoError(t, err)
	err = tagRepo.AddTagToNote(nil, n.ID(), tag2.ID())
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/notes/"+n.ID().String()+"/tags", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data []TagResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Len(t, resp.Data, 2)
	})

	t.Run("note not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/notes/"+uuid.New().String()+"/tags", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
