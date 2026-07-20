package share

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"knowledge-graph/internal/domain/note"
	domainuser "knowledge-graph/internal/domain/user"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/interfaces/api/middleware"
	"knowledge-graph/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestHandler(t *testing.T) (*Handler, *postgres.NoteRepository, *postgres.UserRepository, *postgres.ShareRepository, func()) {
	db, cleanup := testutil.SetupTestDB(t)

	err := db.AutoMigrate(
		&postgres.UserModel{},
		&postgres.UserRoleModel{},
		&postgres.NoteModel{},
		&postgres.NoteShareModel{},
		&postgres.ShareLinkModel{},
	)
	require.NoError(t, err)

	defaultRole := postgres.UserRoleModel{Name: "user", CreatedAt: time.Now()}
	require.NoError(t, db.Create(&defaultRole).Error)

	noteRepo := postgres.NewNoteRepository(db, nil)
	userRepo := postgres.NewUserRepository(db)
	shareRepo := postgres.NewShareRepository(db)

	handler := NewHandler(noteRepo, userRepo, shareRepo)
	return handler, noteRepo, userRepo, shareRepo, cleanup
}

func createTestUser(t *testing.T, userRepo *postgres.UserRepository, login, email string) uuid.UUID {
	ctx := context.Background()
	u, err := domainuser.NewUser(uuid.New(), login, email, "password-hash", "user", time.Now(), time.Time{}, nil)
	require.NoError(t, err)
	require.NoError(t, userRepo.Create(ctx, u))
	return u.ID()
}

func createTestNote(t *testing.T, noteRepo *postgres.NoteRepository, creatorID uuid.UUID) uuid.UUID {
	ctx := context.Background()
	title, err := note.NewTitle("Test Note")
	require.NoError(t, err)
	content, err := note.NewContent("Test content")
	require.NoError(t, err)
	meta, err := note.NewMetadata(map[string]interface{}{"foo": "bar"})
	require.NoError(t, err)
	n := note.NewNoteWithCreator(title, content, "star", meta, creatorID)
	require.NoError(t, noteRepo.Save(ctx, n))
	return n.ID()
}

func withAuth(userID uuid.UUID) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, userID)
		c.Next()
	}
}

func TestShareHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, noteRepo, userRepo, _, cleanup := setupTestHandler(t)
	defer cleanup()

	ownerID := createTestUser(t, userRepo, "owner", "owner@example.com")
	targetID := createTestUser(t, userRepo, "target", "target@example.com")
	otherID := createTestUser(t, userRepo, "other", "other@example.com")
	noteID := createTestNote(t, noteRepo, ownerID)

	t.Run("ShareNote success", func(t *testing.T) {
		authRouter := gin.New()
		authRouter.Use(withAuth(ownerID))
		authRouter.POST("/notes/:id/share", handler.ShareNote)
		body, _ := json.Marshal(map[string]string{
			"user_id":    targetID.String(),
			"permission": "read",
		})

		req := httptest.NewRequest(http.MethodPost, "/notes/"+noteID.String()+"/share", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		authRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp ShareNoteResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, noteID, resp.NoteID)
		assert.Equal(t, targetID, resp.SharedWithUserID)
		assert.Equal(t, "target", resp.SharedWithLogin)
	})

	t.Run("ShareNote updates existing", func(t *testing.T) {
		authRouter := gin.New()
		authRouter.Use(withAuth(ownerID))
		authRouter.POST("/notes/:id/share", handler.ShareNote)
		body, _ := json.Marshal(map[string]string{
			"user_id":    targetID.String(),
			"permission": "write",
		})

		req := httptest.NewRequest(http.MethodPost, "/notes/"+noteID.String()+"/share", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		authRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("ShareNote no auth", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"user_id":    targetID.String(),
			"permission": "read",
		})
		req := httptest.NewRequest(http.MethodPost, "/notes/"+noteID.String()+"/share", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		plainRouter := gin.New()
		plainRouter.POST("/notes/:id/share", handler.ShareNote)
		plainRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("ShareNote note not found", func(t *testing.T) {
		authRouter := gin.New()
		authRouter.Use(withAuth(ownerID))
		authRouter.POST("/notes/:id/share", handler.ShareNote)
		body, _ := json.Marshal(map[string]string{
			"user_id":    targetID.String(),
			"permission": "read",
		})

		req := httptest.NewRequest(http.MethodPost, "/notes/"+uuid.New().String()+"/share", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		authRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("ShareNote forbidden", func(t *testing.T) {
		authRouter := gin.New()
		authRouter.Use(withAuth(otherID))
		authRouter.POST("/notes/:id/share", handler.ShareNote)
		body, _ := json.Marshal(map[string]string{
			"user_id":    targetID.String(),
			"permission": "read",
		})

		req := httptest.NewRequest(http.MethodPost, "/notes/"+noteID.String()+"/share", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		authRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("CreateShareLink success", func(t *testing.T) {
		authRouter := gin.New()
		authRouter.Use(withAuth(ownerID))
		authRouter.POST("/notes/:id/share-links", handler.CreateShareLink)
		body, _ := json.Marshal(map[string]string{"permission": "read"})

		req := httptest.NewRequest(http.MethodPost, "/notes/"+noteID.String()+"/share-links", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		authRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp ShareLinkResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.NotEmpty(t, resp.Token)
		assert.Equal(t, "read", resp.Permission)
	})

	t.Run("CreateShareLink forbidden", func(t *testing.T) {
		authRouter := gin.New()
		authRouter.Use(withAuth(otherID))
		authRouter.POST("/notes/:id/share-links", handler.CreateShareLink)
		body, _ := json.Marshal(map[string]string{"permission": "read"})

		req := httptest.NewRequest(http.MethodPost, "/notes/"+noteID.String()+"/share-links", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		authRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("ListNoteShares", func(t *testing.T) {
		authRouter := gin.New()
		authRouter.Use(withAuth(ownerID))
		authRouter.GET("/notes/:id/shares", handler.ListNoteShares)

		req := httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String()+"/shares", nil)
		w := httptest.NewRecorder()
		authRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Contains(t, resp, "user_shares")
		assert.Contains(t, resp, "share_links")
	})

	t.Run("AccessSharedNote", func(t *testing.T) {
		authRouter := gin.New()
		authRouter.Use(withAuth(ownerID))
		authRouter.POST("/notes/:id/share-links", handler.CreateShareLink)
		body, _ := json.Marshal(map[string]string{"permission": "read"})

		req := httptest.NewRequest(http.MethodPost, "/notes/"+noteID.String()+"/share-links", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		authRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)
		var linkResp ShareLinkResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &linkResp))

		publicRouter := gin.New()
		publicRouter.GET("/shared/:token", handler.AccessSharedNote)

		req = httptest.NewRequest(http.MethodGet, "/shared/"+linkResp.Token, nil)
		w = httptest.NewRecorder()
		publicRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Contains(t, resp, "note")
	})

	t.Run("RevokeShare not found", func(t *testing.T) {
		authRouter := gin.New()
		authRouter.Use(withAuth(ownerID))
		authRouter.DELETE("/notes/:id/shares/:shareId", handler.RevokeShare)

		req := httptest.NewRequest(http.MethodDelete, "/notes/"+noteID.String()+"/shares/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		authRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("RevokeShareLink not found", func(t *testing.T) {
		authRouter := gin.New()
		authRouter.Use(withAuth(ownerID))
		authRouter.DELETE("/share-links/:id", handler.RevokeShareLink)

		req := httptest.NewRequest(http.MethodDelete, "/share-links/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		authRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
