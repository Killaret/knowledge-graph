//go:build integration

package notehandler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"knowledge-graph/internal/config"
	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/interfaces/api/middleware"
	"knowledge-graph/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// NoteHandlerSemanticIntegrationTestSuite verifies the note-card
// "semantically similar" fallback against a real pgvector database.
// This is a regression test for two defects found manually:
//   1. SQL alias mismatch in EmbeddingRepository.FindSimilarNotes
//      ("similarity" alias while GORM expected "score").
//   2. Missing title lookup in the semantic fallback of GetSuggestions.
type NoteHandlerSemanticIntegrationTestSuite struct {
	suite.Suite
	db            *gorm.DB
	noteRepo      *postgres.NoteRepository
	embeddingRepo *postgres.EmbeddingRepository
	router        *gin.Engine
	cleanup       func()
	testUserID    uuid.UUID
}

func (s *NoteHandlerSemanticIntegrationTestSuite) SetupSuite() {
	// Use the vector-enabled test database because we need pgvector.
	s.db, s.cleanup = testutil.SetupTestVectorDB(s.T())

	// Ensure vector extension and all dependent tables exist.
	s.db.Exec("CREATE EXTENSION IF NOT EXISTS vector")
	models := []interface{}{
		&postgres.NoteModel{},
		&postgres.LinkModel{},
		&postgres.NoteKeywordModel{},
		&postgres.UserModel{},
		&postgres.UserRoleModel{},
		&postgres.TagModel{},
		&postgres.NoteTagModel{},
		&postgres.NoteEmbeddingModel{},
	}
	err := s.db.AutoMigrate(models...)
	s.Require().NoError(err, "failed to migrate models")

	s.noteRepo = postgres.NewNoteRepository(s.db, nil)
	s.embeddingRepo = postgres.NewEmbeddingRepository(s.db, "paraphrase-multilingual-MiniLM-L12-v2")

	cfg := &config.Config{
		RecommendationTopN:                    10,
		RecommendationFallbackEnabled:         false,
		RecommendationFallbackSemanticEnabled: true, // force the branch under test
		PaginationDefaultLimit:                20,
		PaginationMaxLimit:                    100,
	}

	handler := New(
		s.noteRepo,
		nil, // taskQueue
		nil, // suggestionsHandler
		nil, // affectedNotesSvc
		0,   // taskDelay
		nil, // recRepo
		s.embeddingRepo,
		nil, // cacheClient
		cfg,
		nil, // graphCache
		nil, // achievementService
		nil, // importSvc
	)

	gin.SetMode(gin.TestMode)
	s.router = gin.New()

	// Inject a test user so the handler can read a real user_id from context.
	s.router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, s.testUserID)
		c.Set(middleware.ContextRoleKey, "user")
		c.Next()
	})

	s.router.GET("/notes/:id/suggestions", handler.GetSuggestions)
}

func (s *NoteHandlerSemanticIntegrationTestSuite) TearDownSuite() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s *NoteHandlerSemanticIntegrationTestSuite) SetupTest() {
	err := testutil.TruncateTables(s.db)
	s.Require().NoError(err, "failed to truncate tables")

	// TruncateTables does not list note_embeddings, clean it explicitly.
	s.Require().NoError(s.db.Exec("TRUNCATE TABLE note_embeddings RESTART IDENTITY CASCADE").Error)

	s.testUserID = uuid.New()
	err = s.db.Create(&postgres.UserModel{
		ID:           s.testUserID,
		Login:        "testuser",
		Email:        "test@example.com",
		PasswordHash: "test-hash",
		CreatedAt:    time.Now(),
	}).Error
	s.Require().NoError(err, "failed to create test user")
}

func (s *NoteHandlerSemanticIntegrationTestSuite) createNote(title, content, noteType string) *note.Note {
	t, err := note.NewTitle(title)
	s.Require().NoError(err)
	c, err := note.NewContent(content)
	s.Require().NoError(err)
	metadata, err := note.NewMetadata(map[string]interface{}{})
	s.Require().NoError(err)
	n := note.NewNoteWithCreator(t, c, noteType, metadata, s.testUserID)

	ctx := context.Background()
	s.Require().NoError(s.noteRepo.Save(ctx, n), "failed to save note")
	return n
}

func (s *NoteHandlerSemanticIntegrationTestSuite) TestGetSuggestions_SemanticFallback_PopulatesTitleAndScore() {
	ctx := context.Background()

	source := s.createNote(
		"Grand Blue anime",
		"A comedy anime about a college diving club and beach life",
		"star",
	)
	similar := s.createNote(
		"Tongari Boushi no Atelier",
		"A fantasy anime about a girl learning magic in an atelier",
		"star",
	)

	// Build two close but non-identical 384-dimensional vectors.
	v1 := make([]float32, 384)
	v2 := make([]float32, 384)
	for i := range v1 {
		base := float32(i) / 100.0
		v1[i] = base
		if i%2 == 0 {
			v2[i] = base
		} else {
			v2[i] = base + 0.01
		}
	}

	s.Require().NoError(s.embeddingRepo.Upsert(ctx, source.ID(), pgvector.NewVector(v1)))
	s.Require().NoError(s.embeddingRepo.Upsert(ctx, similar.ID(), pgvector.NewVector(v2)))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/notes/%s/suggestions?limit=5", source.ID().String()), nil)
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "semantic fallback should return 200")
	s.Equal("semantic", w.Header().Get("X-Recommendations-Source"))
	s.Equal("true", w.Header().Get("X-Recommendations-Stale"))

	var resp SuggestionsResponse
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Len(resp.Suggestions, 1, "should return exactly one similar note")

	suggestion := resp.Suggestions[0]
	s.Equal(similar.ID().String(), suggestion.NoteID)
	s.Equal("Tongari Boushi no Atelier", suggestion.Title)
	s.Greater(suggestion.Score, 0.0, "returned score must be a positive real value, not a default zero from alias mismatch")
}

func TestNoteHandlerSemanticIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	suite.Run(t, new(NoteHandlerSemanticIntegrationTestSuite))
}
