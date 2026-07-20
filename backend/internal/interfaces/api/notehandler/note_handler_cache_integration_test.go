//go:build integration

package notehandler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"knowledge-graph/internal/application/cache"
	"knowledge-graph/internal/config"
	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/testutil"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// NoteHandlerCacheIntegrationTestSuite - интеграционные тесты для кэша графов в NoteHandler
type NoteHandlerCacheIntegrationTestSuite struct {
	suite.Suite
	db          *gorm.DB
	repo        *postgres.NoteRepository
	router      *gin.Engine
	cleanup     func()
	redis       *miniredis.Miniredis
	redisClient *redis.Client
	graphCache  *cache.GraphCache
}

func (s *NoteHandlerCacheIntegrationTestSuite) SetupSuite() {
	// Поднимаем тестовую БД
	s.db, s.cleanup = testutil.SetupTestDB(s.T())

	// Миграция всех моделей
	models := []interface{}{
		&postgres.NoteModel{},
		&postgres.LinkModel{},
		&postgres.NoteKeywordModel{},
		&postgres.UserModel{},
		&postgres.UserRoleModel{},
		&postgres.TagModel{},
		&postgres.NoteTagModel{},
	}
	err := s.db.AutoMigrate(models...)
	s.Require().NoError(err, "failed to migrate models")

	// Поднимаем тестовый Redis
	s.redis, err = miniredis.Run()
	s.Require().NoError(err, "failed to start miniredis")

	s.redisClient = redis.NewClient(&redis.Options{
		Addr: s.redis.Addr(),
	})

	// Создаем репозиторий
	s.repo = postgres.NewNoteRepository(s.db, s.redisClient)

	// Создаем graph cache
	s.graphCache = cache.NewGraphCache(infracache.NewRedisCacheClient(s.redisClient))

	// Создаем хендлер с graph cache
	handler := New(
		s.repo,
		nil, // taskQueue - nil для тестов
		nil, // suggestionsHandler
		nil, // affectedNotesSvc
		0,   // taskDelay
		nil, // recRepo
		nil, // embeddingRepo
		nil,
		&config.Config{
			RecommendationTopN:                    10,
			RecommendationFallbackEnabled:         false,
			RecommendationFallbackSemanticEnabled: false,
		},
		s.graphCache,
		nil, // achievementService
	)

	// Настраиваем Gin
	gin.SetMode(gin.TestMode)
	s.router = gin.New()

	// Регистрируем маршруты
	s.router.POST("/notes", handler.Create)
}

func (s *NoteHandlerCacheIntegrationTestSuite) TearDownSuite() {
	s.redis.Close()
	s.cleanup()
}

func (s *NoteHandlerCacheIntegrationTestSuite) SetupTest() {
	// Очищаем таблицы перед каждым тестом
	err := testutil.TruncateTables(s.db)
	s.Require().NoError(err, "failed to truncate tables")

	// Очищаем Redis
	s.redis.FlushAll()
}

// TestNoteCreateInvalidatesGraphCache - проверка инвалидации кэша при создании заметки
func (s *NoteHandlerCacheIntegrationTestSuite) TestNoteCreateInvalidatesGraphCache() {
	ctx := context.Background()
	userID := uuid.New().String()

	// Кэшируем граф для пользователя
	cachedData := cache.GraphData{
		Nodes: []cache.GraphNode{
			{ID: "node1", Title: "Cached Node", Type: "star"},
		},
		Links: []cache.GraphLink{},
	}

	err := s.graphCache.CacheUserGraph(ctx, userID, cachedData)
	s.Require().NoError(err)

	// Проверяем что кэш существует
	_, found, err := s.graphCache.GetCachedUserGraph(ctx, userID)
	s.Require().NoError(err)
	s.Require().True(found, "cache should exist before note creation")

	// Создаем заметку через API (симулируем аутентифицированного пользователя)
	reqBody := map[string]interface{}{
		"title":    "Test Note",
		"content":  "Test content",
		"type":     "star",
		"metadata": map[string]interface{}{},
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Устанавливаем user ID в контекст (обычно это делает middleware)
	req.Header.Set("X-User-ID", userID)

	// Создаем middleware для установки user ID в контекст
	testRouter := gin.New()
	testRouter.Use(func(c *gin.Context) {
		c.Set("user_id", uuid.MustParse(userID))
		c.Next()
	})

	// Пересоздаем хендлер с middleware
	handler := New(
		s.repo,
		nil,
		nil,
		nil,
		0,
		nil,
		nil,
		nil,
		&config.Config{
			RecommendationTopN:                    10,
			RecommendationFallbackEnabled:         false,
			RecommendationFallbackSemanticEnabled: false,
		},
		s.graphCache,
		nil, // achievementService
	)
	testRouter.POST("/notes", handler.Create)
	testRouter.ServeHTTP(w, req)

	s.Equal(201, w.Code, "note creation should succeed")

	// Проверяем что кэш был инвалидирован
	_, found, err = s.graphCache.GetCachedUserGraph(ctx, userID)
	s.Require().NoError(err)
	s.Require().False(found, "cache should be invalidated after note creation")
}

// TestNoteUpdateInvalidatesGraphCache - проверка инвалидации кэша при обновлении заметки
func (s *NoteHandlerCacheIntegrationTestSuite) TestNoteUpdateInvalidatesGraphCache() {
	ctx := context.Background()
	userID := uuid.New().String()

	// Сначала создаем заметку
	title, _ := note.NewTitle("Original Title")
	content, _ := note.NewContent("Original content")
	metadata, _ := note.NewMetadata(map[string]interface{}{})
	n := note.NewNoteWithCreator(title, content, "star", metadata, uuid.MustParse(userID))
	err := s.repo.Save(ctx, n)
	s.Require().NoError(err)

	// Кэшируем граф для пользователя
	cachedData := cache.GraphData{
		Nodes: []cache.GraphNode{
			{ID: n.ID().String(), Title: "Cached Node", Type: "star"},
		},
		Links: []cache.GraphLink{},
	}

	err = s.graphCache.CacheUserGraph(ctx, userID, cachedData)
	s.Require().NoError(err)

	// Проверяем что кэш существует
	_, found, err := s.graphCache.GetCachedUserGraph(ctx, userID)
	s.Require().NoError(err)
	s.Require().True(found)

	// Обновляем заметку через API
	reqBody := map[string]interface{}{
		"title": "Updated Title",
	}
	jsonBody, _ := json.Marshal(reqBody)

	handler := New(
		s.repo,
		nil,
		nil,
		nil,
		0,
		nil,
		nil,
		nil,
		&config.Config{},
		s.graphCache,
		nil, // achievementService
	)

	testRouter := gin.New()
	testRouter.Use(func(c *gin.Context) {
		c.Set("user_id", uuid.MustParse(userID))
		c.Next()
	})
	testRouter.PUT("/notes/:id", handler.Update)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/notes/"+n.ID().String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	testRouter.ServeHTTP(w, req)

	s.Equal(200, w.Code)

	// Проверяем что кэш был инвалидирован
	_, found, err = s.graphCache.GetCachedUserGraph(ctx, userID)
	s.Require().NoError(err)
	s.Require().False(found, "cache should be invalidated after note update")
}

// TestNoteDeleteInvalidatesGraphCache - проверка инвалидации кэша при удалении заметки
func (s *NoteHandlerCacheIntegrationTestSuite) TestNoteDeleteInvalidatesGraphCache() {
	ctx := context.Background()
	userID := uuid.New().String()

	// Создаем заметку
	title, _ := note.NewTitle("To Delete")
	content, _ := note.NewContent("Content")
	metadata, _ := note.NewMetadata(map[string]interface{}{})
	n := note.NewNoteWithCreator(title, content, "star", metadata, uuid.MustParse(userID))
	err := s.repo.Save(ctx, n)
	s.Require().NoError(err)

	// Кэшируем граф для пользователя
	cachedData := cache.GraphData{
		Nodes: []cache.GraphNode{
			{ID: n.ID().String(), Title: "Cached Node", Type: "star"},
		},
		Links: []cache.GraphLink{},
	}

	err = s.graphCache.CacheUserGraph(ctx, userID, cachedData)
	s.Require().NoError(err)

	// Проверяем что кэш существует
	_, found, err := s.graphCache.GetCachedUserGraph(ctx, userID)
	s.Require().NoError(err)
	s.Require().True(found)

	// Удаляем заметку через API
	handler := New(
		s.repo,
		nil,
		nil,
		nil,
		0,
		nil,
		nil,
		nil,
		&config.Config{},
		s.graphCache,
		nil, // achievementService
	)

	testRouter := gin.New()
	testRouter.Use(func(c *gin.Context) {
		c.Set("user_id", uuid.MustParse(userID))
		c.Next()
	})
	testRouter.DELETE("/notes/:id", handler.Delete)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/notes/"+n.ID().String(), nil)
	testRouter.ServeHTTP(w, req)

	s.Equal(204, w.Code)

	// Проверяем что кэш был инвалидирован
	_, found, err = s.graphCache.GetCachedUserGraph(ctx, userID)
	s.Require().NoError(err)
	s.Require().False(found, "cache should be invalidated after note deletion")
}

// Запускаем тесты
func TestNoteHandlerCacheIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	suite.Run(t, new(NoteHandlerCacheIntegrationTestSuite))
}
