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
// Entries marked in `deleted` are invisible to FindByID but visible to
// FindByIDIncludingDeleted — like the real repository's trash.
type stubNoteRepo struct {
	byID    map[uuid.UUID]*note.Note
	deleted map[uuid.UUID]bool
}

func (s *stubNoteRepo) Save(context.Context, *note.Note) error { return nil }
func (s *stubNoteRepo) FindByID(_ context.Context, id uuid.UUID) (*note.Note, error) {
	if s.deleted[id] {
		return nil, nil
	}
	return s.byID[id], nil
}
func (s *stubNoteRepo) FindByIDIncludingDeleted(_ context.Context, id uuid.UUID) (*note.Note, error) {
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

// NOTE-DELETE-1 regression: the restore route must reach a note that sits in
// the trash. Plain NoteAccessWrite answers 404 because FindByID cannot see
// deleted rows; the trash-aware level resolves ownership on the trashed row.
func TestRequireNoteAccess_WriteIncludeDeleted(t *testing.T) {
	owner := uuid.New()
	stranger := uuid.New()
	trashed := makeNote(t, owner, false)
	repo := &stubNoteRepo{
		byID:    map[uuid.UUID]*note.Note{trashed.ID(): trashed},
		deleted: map[uuid.UUID]bool{trashed.ID(): true},
	}

	tests := []struct {
		name   string
		caller uuid.UUID
		want   int
	}{
		{"owner reaches own trashed note → 200", owner, http.StatusOK},
		{"stranger on foreign trashed note → 404", stranger, http.StatusNotFound},
		{"anonymous on trashed note → 404", uuid.Nil, http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := exercise(t, repo, NoteAccessWriteIncludeDeleted, http.MethodPost, trashed.ID().String(), tt.caller, false)
			assert.Equal(t, tt.want, w.Code)
		})
	}

	// The original defect shape: same request under the plain write level.
	w := exercise(t, repo, NoteAccessWrite, http.MethodPost, trashed.ID().String(), owner, false)
	assert.Equal(t, http.StatusNotFound, w.Code, "without the trash-aware level even the owner gets 404")
}

// A repository without FindByIDIncludingDeleted must fail loudly — a silent
// FindByID fallback would resurrect the 404-on-restore defect.
func TestRequireNoteAccess_WriteIncludeDeletedWithoutSupport(t *testing.T) {
	type bareRepo struct{ note.Repository }
	repo := &bareRepo{}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(ContextUserIDKey, uuid.New())
		c.Next()
	})
	r.POST("/notes/:id/restore", RequireNoteAccess(repo, NoteAccessWriteIncludeDeleted),
		func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest(http.MethodPost, "/notes/"+uuid.New().String()+"/restore", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
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
