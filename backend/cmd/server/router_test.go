package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"knowledge-graph/internal/auth"
	"knowledge-graph/internal/config"
	"knowledge-graph/internal/domain/note"
	drafthandler "knowledge-graph/internal/interfaces/api/handlers/draft"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupRouter(t *testing.T) {
	cfg := &config.Config{}

	jwtConfig := middleware.DefaultJWTConfig(nil, nil)
	apiKeyConfig := middleware.DefaultAPIKeyConfig(nil, false, "")
	skipAuthConfig := middleware.DefaultSkipAuthConfig(false)

	healthHandler := newHealthHandler(nil, nil, nil)
	r := setupRouter(
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil,
		cfg,
		healthHandler,
		newMetricsHandler(nil),
		newWriteLimiter(cfg),
		jwtConfig,
		apiKeyConfig,
		skipAuthConfig,
		nil,
		nil,
	)

	require.NotNil(t, r)
	routes := r.Routes()
	assert.NotEmpty(t, routes)

	// Ensure expected groups/paths are registered
	paths := make(map[string]bool)
	for _, route := range routes {
		paths[route.Path] = true
	}

	assert.True(t, paths["/health"])
	assert.True(t, paths["/swagger/*any"])
	assert.True(t, paths["/api/v1/notes"])
	assert.True(t, paths["/api/v1/tags"])
	assert.True(t, paths["/api/v1/links"])

	// Regression check: user profile and API key management routes must be
	// wired up (previously only GET /users/me was registered, silently
	// breaking profile updates, account deletion, and API key management).
	methodPaths := make(map[string]bool)
	for _, route := range routes {
		methodPaths[route.Method+" "+route.Path] = true
	}

	for _, mp := range []string{
		"GET /api/v1/users/me",
		"PUT /api/v1/users/me",
		"DELETE /api/v1/users/me",
		"GET /api/v1/users/me/api-keys",
		"POST /api/v1/users/me/api-keys",
		"DELETE /api/v1/users/me/api-keys/:id",
	} {
		assert.True(t, methodPaths[mp], "expected route to be registered: %s", mp)
	}
}

// accessGuardNoteRepo is a note.Repository that returns one fixed note for
// every FindByID; other methods are never reached by the access middleware.
type accessGuardNoteRepo struct {
	note *note.Note
}

func (r *accessGuardNoteRepo) Save(ctx context.Context, n *note.Note) error { return nil }
func (r *accessGuardNoteRepo) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	return r.note, nil
}
func (r *accessGuardNoteRepo) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	return r.note, nil
}
func (r *accessGuardNoteRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (r *accessGuardNoteRepo) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
	return nil
}
func (r *accessGuardNoteRepo) Restore(ctx context.Context, id uuid.UUID) error { return nil }
func (r *accessGuardNoteRepo) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*note.Note, int64, error) {
	return nil, 0, nil
}
func (r *accessGuardNoteRepo) Search(ctx context.Context, userID uuid.UUID, query string, limit, offset int) ([]*note.Note, int64, error) {
	return nil, 0, nil
}
func (r *accessGuardNoteRepo) FindAll(ctx context.Context) ([]*note.Note, error) {
	return nil, nil
}
func (r *accessGuardNoteRepo) FindAllPaginated(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*note.Note, int64, error) {
	return nil, 0, nil
}

// SEC-1 regression: every route carrying /notes/:id must refuse a stranger
// with 404 through the wired RequireNoteAccess middleware. A route that loses
// its guard falls through to a nil handler and answers 500 — not 404 — so the
// mutation "remove noteRead/noteWrite from one route" turns this test red.
func TestNoteIDRoutesRequireAccessGuard(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{}
	jwtManager := auth.NewJWTManager("test-secret-key-for-router-guard-test", time.Minute, time.Hour)
	jwtConfig := middleware.DefaultJWTConfig(jwtManager, nil)
	apiKeyConfig := middleware.DefaultAPIKeyConfig(nil, false, "")
	skipAuthConfig := middleware.DefaultSkipAuthConfig(false)

	strangerID := uuid.New()
	pair, err := jwtManager.GenerateTokenPair(strangerID, "stranger", "user")
	require.NoError(t, err)

	title, err := note.NewTitle("Foreign private note")
	require.NoError(t, err)
	content, err := note.NewContent("not yours")
	require.NoError(t, err)
	ownerID := uuid.New()
	noteID := uuid.New()
	foreignNote := note.ReconstructNoteWithCreator(noteID, title, content, note.MustType("star"),
		note.Metadata{}, &ownerID, time.Now(), time.Now())

	r := setupRouter(
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, &drafthandler.Handler{},
		cfg,
		newHealthHandler(nil, nil, nil),
		newMetricsHandler(nil),
		newWriteLimiter(cfg),
		jwtConfig,
		apiKeyConfig,
		skipAuthConfig,
		&accessGuardNoteRepo{note: foreignNote},
		nil,
	)

	type noteRoute struct {
		method string
		path   string
	}
	guarded := make([]noteRoute, 0)
	for _, route := range r.Routes() {
		if strings.Contains(route.Path, "/notes/:id") {
			guarded = append(guarded, noteRoute{method: route.Method, path: route.Path})
		}
	}
	require.GreaterOrEqual(t, len(guarded), 15,
		"expected the full /notes/:id route set; got %d — did routes move?", len(guarded))

	for _, route := range guarded {
		url := route.path
		for {
			start := strings.Index(url, "/:")
			if start < 0 {
				break
			}
			end := strings.Index(url[start+1:], "/")
			if end < 0 {
				end = len(url)
			} else {
				end += start + 1
			}
			replacement := uuid.New().String()
			if strings.HasPrefix(url[start:], "/:id") {
				replacement = noteID.String()
			}
			url = url[:start+1] + replacement + url[end:]
		}

		req := httptest.NewRequest(route.method, url, strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code,
			"%s %s answered %d for a stranger; a route without RequireNoteAccess falls through to the handler",
			route.method, route.path, w.Code)
	}

	// PUB-1: GET /notes/:id is JWT-exempt for anonymous readers, so
	// concealment of private notes must come from noteRead, not from the 401
	// barrier. Anonymous must still get 404, never the content.
	anonReq := httptest.NewRequest(http.MethodGet, "/api/v1/notes/"+noteID.String(), nil)
	anonW := httptest.NewRecorder()
	r.ServeHTTP(anonW, anonReq)
	assert.Equal(t, http.StatusNotFound, anonW.Code,
		"anonymous GET on a private note must be concealed with 404, got %d", anonW.Code)
}

// trashAwareNoteRepo mirrors the real repository's split: FindByID never sees
// soft-deleted notes, FindByIDIncludingDeleted does. The stored note pretends
// to sit in the trash.
type trashAwareNoteRepo struct {
	note *note.Note
}

func (r *trashAwareNoteRepo) Save(ctx context.Context, n *note.Note) error { return nil }
func (r *trashAwareNoteRepo) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	return nil, nil
}
func (r *trashAwareNoteRepo) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	return r.note, nil
}
func (r *trashAwareNoteRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (r *trashAwareNoteRepo) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
	return nil
}
func (r *trashAwareNoteRepo) Restore(ctx context.Context, id uuid.UUID) error { return nil }
func (r *trashAwareNoteRepo) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*note.Note, int64, error) {
	return nil, 0, nil
}
func (r *trashAwareNoteRepo) Search(ctx context.Context, userID uuid.UUID, query string, limit, offset int) ([]*note.Note, int64, error) {
	return nil, 0, nil
}
func (r *trashAwareNoteRepo) FindAll(ctx context.Context) ([]*note.Note, error) {
	return nil, nil
}
func (r *trashAwareNoteRepo) FindAllPaginated(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*note.Note, int64, error) {
	return nil, 0, nil
}

// NOTE-DELETE-1-TAIL: POST /notes/:id/restore must look the note up through
// FindByIDIncludingDeleted — its target sits in the trash. The router wires
// noteWriteTrash for that; the mutation "route behind plain noteWrite" makes
// the guard answer 404 before the handler runs. With the correct guard the
// request reaches the (nil in this harness) handler, which the recovery
// middleware turns into a 500 — anything but 404 proves the guard passed.
func TestRestoreRouteSeesTrashedNote(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{}
	jwtManager := auth.NewJWTManager("test-secret-key-for-restore-guard-test", time.Minute, time.Hour)
	jwtConfig := middleware.DefaultJWTConfig(jwtManager, nil)
	apiKeyConfig := middleware.DefaultAPIKeyConfig(nil, false, "")
	skipAuthConfig := middleware.DefaultSkipAuthConfig(false)

	ownerID := uuid.New()
	pair, err := jwtManager.GenerateTokenPair(ownerID, "owner", "user")
	require.NoError(t, err)

	title, err := note.NewTitle("Trashed note")
	require.NoError(t, err)
	content, err := note.NewContent("in the trash")
	require.NoError(t, err)
	noteID := uuid.New()
	trashed := note.ReconstructNoteWithCreator(noteID, title, content, note.MustType("star"),
		note.Metadata{}, &ownerID, time.Now(), time.Now())

	r := setupRouter(
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, &drafthandler.Handler{},
		cfg,
		newHealthHandler(nil, nil, nil),
		newMetricsHandler(nil),
		newWriteLimiter(cfg),
		jwtConfig,
		apiKeyConfig,
		skipAuthConfig,
		&trashAwareNoteRepo{note: trashed},
		nil,
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/notes/"+noteID.String()+"/restore", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusNotFound, w.Code,
		"owner restoring a trashed note must pass the guard; 404 means the route lost noteWriteTrash")
}
