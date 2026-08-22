//go:build integration

package linkhandler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"knowledge-graph/internal/application/cache"
	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"
	infracache "knowledge-graph/internal/infrastructure/cache"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/testutil"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// LinkHandlerCacheIntegrationTestSuite - интеграционные тесты для кэша графов в LinkHandler
type LinkHandlerCacheIntegrationTestSuite struct {
	suite.Suite
	db          *gorm.DB
	noteRepo    *postgres.NoteRepository
	linkRepo    *postgres.LinkRepository
	router      *gin.Engine
	cleanup     func()
	redis       *miniredis.Miniredis
	redisClient *redis.Client
	graphCache  *cache.GraphCache
}

func (s *LinkHandlerCacheIntegrationTestSuite) SetupSuite() {
	// Поднимаем тестовую БД
	s.db, s.cleanup = testutil.SetupTestDB(s.T())

	// Миграция всех моделей
	models := []interface{}{
		&postgres.NoteModel{},
		&postgres.LinkModel{},
		&postgres.NoteKeywordModel{},
		&postgres.UserModel{},
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

	// Создаем репозитории
	s.noteRepo = postgres.NewNoteRepository(s.db, nil)
	s.linkRepo = postgres.NewLinkRepository(s.db)

	// Создаем graph cache
	s.graphCache = cache.NewGraphCache(infracache.NewRedisCacheClient(s.redisClient))

	// Создаем хендлер с graph cache
	handler := New(
		s.linkRepo,
		s.noteRepo,
		nil, // achievementService
		s.graphCache,
	)

	// Настраиваем Gin
	gin.SetMode(gin.TestMode)
	s.router = gin.New()

	// Middleware для установки user ID в контекст
	s.router.Use(func(c *gin.Context) {
		// Для тестов получаем user ID из заголовка
		if userID := c.GetHeader("X-User-ID"); userID != "" {
			c.Set("user_id", uuid.MustParse(userID))
		}
		c.Next()
	})

	// Регистрируем маршруты
	s.router.POST("/links", handler.Create)
	s.router.DELETE("/links/:id", handler.Delete)
}

func (s *LinkHandlerCacheIntegrationTestSuite) TearDownSuite() {
	s.redis.Close()
	s.cleanup()
}

func (s *LinkHandlerCacheIntegrationTestSuite) SetupTest() {
	// Очищаем таблицы перед каждым тестом
	err := testutil.TruncateTables(s.db)
	s.Require().NoError(err, "failed to truncate tables")

	// Очищаем Redis
	s.redis.FlushAll()
}

// createTestNote создает тестовую заметку
func (s *LinkHandlerCacheIntegrationTestSuite) createTestNote(title, content, noteType string, userID uuid.UUID) *note.Note {
	ctx := context.Background()

	// Сначала создаем пользователя в базе с обязательными полями
	userModel := &postgres.UserModel{
		ID:           userID,
		Login:        "testuser_" + userID.String(),
		Email:        "test_" + userID.String() + "@example.com",
		PasswordHash: "dummy_hash", // Обязательное поле
	}
	s.db.WithContext(ctx).Save(userModel)

	noteTitle, _ := note.NewTitle(title)
	noteContent, _ := note.NewContent(content)
	metadata, _ := note.NewMetadata(map[string]interface{}{"type": noteType})
	n := note.NewNoteWithCreator(noteTitle, noteContent, noteType, metadata, userID)
	err := s.noteRepo.Save(ctx, n)
	s.Require().NoError(err, "failed to create test note")
	return n
}

// TestLinkCreateInvalidatesGraphCache - проверка инвалидации кэша при создании связи
func (s *LinkHandlerCacheIntegrationTestSuite) TestLinkCreateInvalidatesGraphCache() {
	ctx := context.Background()
	userID := uuid.New()

	// Создаем заметки
	note1 := s.createTestNote("Note 1", "content1", "star", userID)
	note2 := s.createTestNote("Note 2", "content2", "planet", userID)

	// Кэшируем граф для пользователя
	cachedData := cache.GraphData{
		Nodes: []cache.GraphNode{
			{ID: note1.ID().String(), Title: "Note 1", Type: "star"},
			{ID: note2.ID().String(), Title: "Note 2", Type: "planet"},
		},
		Links: []cache.GraphLink{},
	}

	err := s.graphCache.CacheUserGraph(ctx, userID.String(), cachedData)
	s.Require().NoError(err)

	// Проверяем что кэш существует
	_, found, err := s.graphCache.GetCachedUserGraph(ctx, userID.String())
	s.Require().NoError(err)
	s.Require().True(found, "cache should exist before link creation")

	// Создаем связь через API
	reqBody := map[string]interface{}{
		"source_note_id": note1.ID().String(),
		"target_note_id": note2.ID().String(),
		"link_type":      "reference",
		"weight":         0.5,
		"metadata":       map[string]interface{}{},
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID.String())
	s.router.ServeHTTP(w, req)

	s.Equal(201, w.Code, "link creation should succeed")

	// Проверяем что кэш был инвалидирован
	_, found, err = s.graphCache.GetCachedUserGraph(ctx, userID.String())
	s.Require().NoError(err)
	s.Require().False(found, "cache should be invalidated after link creation")
}

// TestLinkDeleteInvalidatesGraphCache - проверка инвалидации кэша при удалении связи
func (s *LinkHandlerCacheIntegrationTestSuite) TestLinkDeleteInvalidatesGraphCache() {
	ctx := context.Background()
	userID := uuid.New()

	// Создаем заметки и связь
	note1 := s.createTestNote("Note 1", "content1", "star", userID)
	note2 := s.createTestNote("Note 2", "content2", "planet", userID)

	lt, _ := link.NewLinkType("reference")
	weight, _ := link.NewWeight(1.0)
	m, _ := link.NewMetadata(map[string]interface{}{})
	l := link.NewLinkWithCreator(note1.ID(), note2.ID(), userID, lt, weight, m)
	err := s.linkRepo.Save(ctx, l)
	s.Require().NoError(err)

	// Кэшируем граф для пользователя
	cachedData := cache.GraphData{
		Nodes: []cache.GraphNode{
			{ID: note1.ID().String(), Title: "Note 1", Type: "star"},
			{ID: note2.ID().String(), Title: "Note 2", Type: "planet"},
		},
		Links: []cache.GraphLink{
			{Source: note1.ID().String(), Target: note2.ID().String(), Weight: 1.0, LinkType: "reference"},
		},
	}

	err = s.graphCache.CacheUserGraph(ctx, userID.String(), cachedData)
	s.Require().NoError(err)

	// Проверяем что кэш существует
	_, found, err := s.graphCache.GetCachedUserGraph(ctx, userID.String())
	s.Require().NoError(err)
	s.Require().True(found)

	// Удаляем связь через API
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/links/"+l.ID().String(), nil)
	req.Header.Set("X-User-ID", userID.String())
	s.router.ServeHTTP(w, req)

	s.Equal(204, w.Code)

	// Проверяем что кэш был инвалидирован
	_, found, err = s.graphCache.GetCachedUserGraph(ctx, userID.String())
	s.Require().NoError(err)
	s.Require().False(found, "cache should be invalidated after link deletion")
}

// TestLinkCreateWithoutAuth_NoInvalidation - проверка что без авторизации инвалидации не происходит
func (s *LinkHandlerCacheIntegrationTestSuite) TestLinkCreateWithoutAuth_NoInvalidation() {
	ctx := context.Background()
	userID := uuid.New()

	// Создаем заметки
	note1 := s.createTestNote("Note 1", "content1", "star", userID)
	note2 := s.createTestNote("Note 2", "content2", "planet", userID)

	// Кэшируем граф для пользователя
	cachedData := cache.GraphData{
		Nodes: []cache.GraphNode{
			{ID: note1.ID().String(), Title: "Note 1", Type: "star"},
		},
		Links: []cache.GraphLink{},
	}

	err := s.graphCache.CacheUserGraph(ctx, userID.String(), cachedData)
	s.Require().NoError(err)

	// Создаем связь без user ID в заголовке
	reqBody := map[string]interface{}{
		"source_note_id": note1.ID().String(),
		"target_note_id": note2.ID().String(),
		"link_type":      "reference",
		"weight":         0.5,
		"metadata":       map[string]interface{}{},
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	// Не устанавливаем X-User-ID
	s.router.ServeHTTP(w, req)

	s.Equal(201, w.Code, "link creation should succeed even without auth")

	// Кэш должен остаться нетронутым (так как user ID не был установлен)
	_, found, err := s.graphCache.GetCachedUserGraph(ctx, userID.String())
	s.Require().NoError(err)
	s.Require().True(found, "cache should NOT be invalidated when no user ID is provided")
}

// TestLinkDeleteWithoutAuth_NoInvalidation - проверка что без авторизации инвалидации не происходит
func (s *LinkHandlerCacheIntegrationTestSuite) TestLinkDeleteWithoutAuth_NoInvalidation() {
	ctx := context.Background()
	userID := uuid.New()

	// Создаем заметки и связь
	note1 := s.createTestNote("Note 1", "content1", "star", userID)
	note2 := s.createTestNote("Note 2", "content2", "planet", userID)

	lt, _ := link.NewLinkType("reference")
	weight, _ := link.NewWeight(1.0)
	m, _ := link.NewMetadata(map[string]interface{}{})
	l := link.NewLink(note1.ID(), note2.ID(), lt, weight, m)
	err := s.linkRepo.Save(ctx, l)
	s.Require().NoError(err)

	// Кэшируем граф для пользователя
	cachedData := cache.GraphData{
		Nodes: []cache.GraphNode{
			{ID: note1.ID().String(), Title: "Note 1", Type: "star"},
		},
		Links: []cache.GraphLink{},
	}

	err = s.graphCache.CacheUserGraph(ctx, userID.String(), cachedData)
	s.Require().NoError(err)

	// Удаляем связь без user ID в заголовке
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/links/"+l.ID().String(), nil)
	// Не устанавливаем X-User-ID
	s.router.ServeHTTP(w, req)

	s.Equal(204, w.Code)

	// Кэш должен остаться нетронутым
	_, found, err := s.graphCache.GetCachedUserGraph(ctx, userID.String())
	s.Require().NoError(err)
	s.Require().True(found, "cache should NOT be invalidated when no user ID is provided")
}

// TestLinkCreateError_NoInvalidation - проверка что при ошибке создания связи инвалидация не происходит
func (s *LinkHandlerCacheIntegrationTestSuite) TestLinkCreateError_NoInvalidation() {
	ctx := context.Background()
	userID := uuid.New()

	// Создаем только одну заметку (вторая не существует)
	note1 := s.createTestNote("Note 1", "content1", "star", userID)
	fakeNoteID := uuid.New()

	// Кэшируем граф для пользователя
	cachedData := cache.GraphData{
		Nodes: []cache.GraphNode{
			{ID: note1.ID().String(), Title: "Note 1", Type: "star"},
		},
		Links: []cache.GraphLink{},
	}

	err := s.graphCache.CacheUserGraph(ctx, userID.String(), cachedData)
	s.Require().NoError(err)

	// Пытаемся создать связь с несуществующей заметкой
	reqBody := map[string]interface{}{
		"source_note_id": note1.ID().String(),
		"target_note_id": fakeNoteID.String(),
		"link_type":      "reference",
		"weight":         0.5,
		"metadata":       map[string]interface{}{},
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID.String())
	s.router.ServeHTTP(w, req)

	s.Equal(404, w.Code, "should return 404 when target note not found")

	// Кэш должен остаться нетронутым
	_, found, err := s.graphCache.GetCachedUserGraph(ctx, userID.String())
	s.Require().NoError(err)
	s.Require().True(found, "cache should NOT be invalidated when link creation fails")
}

func TestLinkHandlerCacheIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	suite.Run(t, new(LinkHandlerCacheIntegrationTestSuite))
}
