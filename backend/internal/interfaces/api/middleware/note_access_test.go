package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"knowledge-graph/internal/domain/note"
	contextkeys "knowledge-graph/internal/shared/context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// stubNoteRepo is a minimal note.Repository implementation for
// access-control tests: it stores one note per id and ignores the rest.
type stubNoteRepo struct {
	byID map[uuid.UUID]*note.Note
}

func (s *stubNoteRepo) Save(context.Context, *note.Note) error { return nil }
func (s *stubNoteRepo) FindByID(_ context.Context, id uuid.UUID) (*note.Note, error) {
	return s.byID[id], nil
}
func (s *stubNoteRepo) Delete(context.Context, uuid.UUID) error        { return nil }
func (s *stubNoteRepo) DeleteBatch(context.Context, []uuid.UUID) error { return nil }
func (s *stubNoteRepo) Restore(context.Context, uuid.UUID) error       { return nil }
func (s *stubNoteRepo) FindAll(context.Context) ([]*note.Note, error)  { return nil, nil }
func (s *stubNoteRepo) List(context.Context, uuid.UUID, int, int) ([]*note.Note, int64, error) {
	return nil, 0, nil
}
func (s *stubNoteRepo) Search(context.Context, uuid.UUID, string, int, int) ([]*note.Note, int64, error) {
	return nil, 0, nil
}
func (s *stubNoteRepo) FindAllPaginated(context.Context, uuid.UUID, int, int) ([]*note.Note, int64, error) {
	return nil, 0, nil
}

func makeNote(t *testing.T, creatorID uuid.UUID, public bool) *note.Note {
	t.Helper()
	title, err := note.NewTitle("owned")
	assert.NoError(t, err)
	content, err := note.NewContent("secret")
	assert.NoError(t, err)
	metadata, err := note.NewMetadata(map[string]interface{}{})
	assert.NoError(t, err)
	return note.NewNoteWithCreator(title, content, note.MustType("star"), metadata, creatorID, note.WithIsPublic(public))
}

// exercise builds a Gin engine with one route guarded by the access
// middleware and performs the request as callerID (uuid.Nil = anonymous).
func exercise(t *testing.T, repo *stubNoteRepo, level NoteAccessLevel, method, targetID string, callerID uuid.UUID, skipAuth bool) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if skipAuth {
			c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), contextkeys.SkipAuthKey, true))
		}
		if callerID != uuid.Nil {
			c.Set(ContextUserIDKey, callerID)
		}
		c.Next()
	})
	handler := func(c *gin.Context) { c.Status(http.StatusOK) }
	switch method {
	case http.MethodGet:
		r.GET("/notes/:id", RequireNoteAccess(repo, level), handler)
	case http.MethodPut:
		r.PUT("/notes/:id", RequireNoteAccess(repo, level), handler)
	case http.MethodDelete:
		r.DELETE("/notes/:id", RequireNoteAccess(repo, level), handler)
	case http.MethodPost:
		r.POST("/notes/:id", RequireNoteAccess(repo, level), handler)
	}
	req := httptest.NewRequest(method, "/notes/"+targetID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRequireNoteAccess_ReadRoutes(t *testing.T) {
	owner := uuid.New()
	stranger := uuid.New()
	private := makeNote(t, owner, false)
	public := makeNote(t, owner, true)
	repo := &stubNoteRepo{byID: map[uuid.UUID]*note.Note{
		private.ID(): private,
		public.ID():  public,
	}}

	tests := []struct {
		name     string
		targetID uuid.UUID
		caller   uuid.UUID
		want     int
	}{
		{"owner reads own private", private.ID(), owner, http.StatusOK},
		{"stranger reads foreign private → 404", private.ID(), stranger, http.StatusNotFound},
		{"anonymous reads foreign private → 404", private.ID(), uuid.Nil, http.StatusNotFound},
		{"stranger reads foreign public → 200", public.ID(), stranger, http.StatusOK},
		{"anonymous reads foreign public → 200", public.ID(), uuid.Nil, http.StatusOK},
		{"missing note → 404", uuid.New(), owner, http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := exercise(t, repo, NoteAccessRead, http.MethodGet, tt.targetID.String(), tt.caller, false)
			assert.Equal(t, tt.want, w.Code)
		})
	}
}

func TestRequireNoteAccess_WriteRoutes(t *testing.T) {
	owner := uuid.New()
	stranger := uuid.New()
	private := makeNote(t, owner, false)
	public := makeNote(t, owner, true)
	repo := &stubNoteRepo{byID: map[uuid.UUID]*note.Note{
		private.ID(): private,
		public.ID():  public,
	}}

	tests := []struct {
		name     string
		method   string
		targetID uuid.UUID
		caller   uuid.UUID
		want     int
	}{
		// The three SEC-1 criteria: a stranger must not read, modify, or
		// delete a foreign private note — each guarded, each 404.
		{"stranger PUT foreign private → 404", http.MethodPut, private.ID(), stranger, http.StatusNotFound},
		{"stranger DELETE foreign private → 404", http.MethodDelete, private.ID(), stranger, http.StatusNotFound},
		{"stranger PUT foreign public → 404 (read ≠ write)", http.MethodPut, public.ID(), stranger, http.StatusNotFound},
		{"stranger DELETE foreign public → 404", http.MethodDelete, public.ID(), stranger, http.StatusNotFound},
		{"anonymous PUT → 404", http.MethodPut, private.ID(), uuid.Nil, http.StatusNotFound},
		{"anonymous DELETE → 404", http.MethodDelete, public.ID(), uuid.Nil, http.StatusNotFound},
		{"owner PUT own private → 200", http.MethodPut, private.ID(), owner, http.StatusOK},
		{"owner DELETE own private → 200", http.MethodDelete, private.ID(), owner, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := exercise(t, repo, NoteAccessWrite, tt.method, tt.targetID.String(), tt.caller, false)
			assert.Equal(t, tt.want, w.Code)
		})
	}
}

// A route registered with NoteAccessRead must not turn a write method
// into a stranger-allowed call.
func TestRequireNoteAccess_ReadLevelDoesNotPermitWrites(t *testing.T) {
	owner := uuid.New()
	stranger := uuid.New()
	public := makeNote(t, owner, true)
	repo := &stubNoteRepo{byID: map[uuid.UUID]*note.Note{public.ID(): public}}

	w := exercise(t, repo, NoteAccessRead, http.MethodPost, public.ID().String(), stranger, false)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// SKIP_AUTH keeps full access for the seeded test user so E2E suites
// running with the bypass stay green.
func TestRequireNoteAccess_SkipAuthBypasses(t *testing.T) {
	owner := uuid.New()
	private := makeNote(t, owner, false)
	repo := &stubNoteRepo{byID: map[uuid.UUID]*note.Note{private.ID(): private}}

	w := exercise(t, repo, NoteAccessWrite, http.MethodDelete, private.ID().String(), uuid.Nil, true)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNoteAccessHelpers(t *testing.T) {
	owner := uuid.New()
	stranger := uuid.New()
	private := makeNote(t, owner, false)
	public := makeNote(t, owner, true)
	creatorless := makeNote(t, uuid.Nil, false)

	assert.True(t, private.IsOwnedBy(owner))
	assert.False(t, private.IsOwnedBy(stranger))
	// Pure id equality: the seeded test user authenticates as uuid.Nil and
	// must own its Nil-created notes. Anonymous requests are excluded by the
	// middleware's `authed` gate, not by this method.
	assert.True(t, creatorless.IsOwnedBy(uuid.Nil))
	assert.True(t, public.IsPublic())
	assert.False(t, private.IsPublic())
}

// A Nil-id note belongs to the authenticated Nil-id test user: in the test
// stack testuser IS uuid.Nil, so it must reach its own notes.
func TestRequireNoteAccess_NilIdentityOwner(t *testing.T) {
	n := makeNote(t, uuid.Nil, false)
	repo := &stubNoteRepo{byID: map[uuid.UUID]*note.Note{n.ID(): n}}

	w := exercise(t, repo, NoteAccessWrite, http.MethodDelete, n.ID().String(), uuid.Nil, false)
	assert.Equal(t, http.StatusNotFound, w.Code, "anonymous (no identity) must not own Nil-created notes")

	// Authenticated Nil identity (the test stack's testuser) — owner.
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(ContextUserIDKey, uuid.Nil)
		c.Next()
	})
	r.DELETE("/notes/:id", RequireNoteAccess(repo, NoteAccessWrite), func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest(http.MethodDelete, "/notes/"+n.ID().String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}
