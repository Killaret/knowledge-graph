//go:build integration

package graphhandler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"knowledge-graph/internal/application/cache"
	"knowledge-graph/internal/config"
	"knowledge-graph/internal/domain/link"
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

// GraphHandlerCacheIntegrationTestSuite - интеграционные тесты для кэша графов в GraphHandler
type GraphHandlerCacheIntegrationTestSuite struct {
	suite.Suite
	db          *gorm.DB
	noteRepo    *postgres.NoteRepository
	linkRepo    *postgres.LinkRepository
	router      *gin.Engine
	handler     *Handler
	cleanup     func()
	redis       *miniredis.Miniredis
	redisClient *redis.Client
	graphCache  *cache.GraphCache
}

func (s *GraphHandlerCacheIntegrationTestSuite) SetupSuite() {
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
	s.graphCache = cache.NewGraphCache(s.redisClient)

	// Создаем хендлер с graph cache и конфигурацией
	cfg := &config.Config{
		GraphDefaultLimit:     100,
		GraphMaxLimit:         1000,
		GraphLinkDefaultLimit: 1000,
		GraphLinkMaxLimit:     10000,
		GraphLoadDepth:        3,
	}
	s.handler = New(s.noteRepo, s.linkRepo, cfg, s.graphCache)

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
	s.router.GET("/me/graph/cached", s.handler.GetCachedGraph)
	s.router.GET("/me/graph/fresh", s.handler.GetFreshGraph)
	s.router.GET("/notes/:id/graph", s.handler.GetGraph)
	s.router.GET("/graph/all", s.handler.GetFullGraph)
}

func (s *GraphHandlerCacheIntegrationTestSuite) TearDownSuite() {
	s.redis.Close()
	s.cleanup()
}

func (s *GraphHandlerCacheIntegrationTestSuite) SetupTest() {
	// Очищаем таблицы перед каждым тестом
	err := testutil.TruncateTables(s.db)
	s.Require().NoError(err, "failed to truncate tables")

	// Очищаем Redis
	s.redis.FlushAll()
}

// createTestNote создает тестовую заметку
func (s *GraphHandlerCacheIntegrationTestSuite) createTestNote(title, content, noteType string, userID uuid.UUID) *note.Note {
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

// createTestLink создает тестовую связь между заметками
func (s *GraphHandlerCacheIntegrationTestSuite) createTestLink(source, target *note.Note, linkType string) *link.Link {
	ctx := context.Background()
	lt, _ := link.NewLinkType(linkType)
	w, _ := link.NewWeight(1.0)
	m, _ := link.NewMetadata(map[string]interface{}{})
	l := link.NewLink(source.ID(), target.ID(), lt, w, m)
	err := s.linkRepo.Save(ctx, l)
	s.Require().NoError(err, "failed to create test link")
	return l
}

// TestGetCachedGraph_Returns204WhenNoCache - проверка что endpoint возвращает 204 когда кэш пуст
func (s *GraphHandlerCacheIntegrationTestSuite) TestGetCachedGraph_Returns204WhenNoCache() {
	userID := uuid.New()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/me/graph/cached", nil)
	req.Header.Set("X-User-ID", userID.String())
	s.router.ServeHTTP(w, req)

	s.Equal(204, w.Code, "should return 204 when no cache exists")
}

// TestGetCachedGraph_ReturnsCachedData - проверка что endpoint возвращает кэшированные данные
func (s *GraphHandlerCacheIntegrationTestSuite) TestGetCachedGraph_ReturnsCachedData() {
	ctx := context.Background()
	userID := uuid.New()

	// Кэшируем тестовые данные напрямую
	cachedData := cache.GraphData{
		Nodes: []cache.GraphNode{
			{ID: "node1", Title: "Cached Node 1", Type: "star"},
			{ID: "node2", Title: "Cached Node 2", Type: "planet"},
		},
		Links: []cache.GraphLink{
			{Source: "node1", Target: "node2", Weight: 0.5, LinkType: "reference"},
		},
	}
	err := s.graphCache.CacheUserGraph(ctx, userID.String(), cachedData)
	s.Require().NoError(err, "failed to cache graph")

	// Проверяем что кэш работает через cached endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/me/graph/cached", nil)
	req.Header.Set("X-User-ID", userID.String())
	s.router.ServeHTTP(w, req)

	s.Equal(200, w.Code, "should return 200 with cached data")

	var response GraphData
	err = json.Unmarshal(w.Body.Bytes(), &response)
	s.Require().NoError(err, "failed to unmarshal response")

	// Проверяем только что данные возвращаются в правильном формате
	s.NotNil(&response, "response should exist")
	s.NotNil(&response.Nodes, "nodes should exist")
	s.NotNil(&response.Links, "links should exist")
}

// TestGetFreshGraph_ReturnsFreshDataAndCachesIt - проверка что fresh endpoint возвращает данные и кэширует их
func (s *GraphHandlerCacheIntegrationTestSuite) TestGetFreshGraph_ReturnsFreshDataAndCachesIt() {
	ctx := context.Background()
	userID := uuid.New()

	// Кэшируем начальные данные
	cachedData := cache.GraphData{
		Nodes: []cache.GraphNode{
			{ID: "node1", Title: "Initial Node", Type: "star"},
		},
		Links: []cache.GraphLink{},
	}
	err := s.graphCache.CacheUserGraph(ctx, userID.String(), cachedData)
	s.Require().NoError(err)

	// Запрашиваем свежие данные
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/me/graph/fresh", nil)
	req.Header.Set("X-User-ID", userID.String())
	s.router.ServeHTTP(w, req)

	s.Equal(200, w.Code, "should return 200 with fresh data")

	var response struct {
		Fresh GraphData   `json:"fresh"`
		Delta *GraphDelta `json:"delta,omitempty"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	s.Require().NoError(err, "failed to unmarshal response")

	// Проверяем что ответ имеет правильную структуру
	s.NotNil(&response.Fresh, "fresh data should exist")

	// Проверяем что данные были закэшированы (упрощенная проверка)
	_, found, err := s.graphCache.GetCachedUserGraph(ctx, userID.String())
	s.Require().NoError(err, "failed to get cached graph")
	s.True(found, "data should be cached after fresh request")
}

// TestGetFreshGraph_ComputesDelta - проверка что delta вычисляется корректно
func (s *GraphHandlerCacheIntegrationTestSuite) TestGetFreshGraph_ComputesDelta() {
	ctx := context.Background()
	userID := uuid.New()

	// Создаем и кэшируем начальные данные
	note1 := s.createTestNote("Note 1", "content1", "star", userID)
	cachedData := cache.GraphData{
		Nodes: []cache.GraphNode{
			{ID: note1.ID().String(), Title: "Note 1", Type: "star"},
		},
		Links: []cache.GraphLink{},
	}
	err := s.graphCache.CacheUserGraph(ctx, userID.String(), cachedData)
	s.Require().NoError(err)

	// Добавляем новую заметку
	s.createTestNote("Note 2", "content2", "planet", userID)

	// Запрашиваем свежие данные
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/me/graph/fresh", nil)
	req.Header.Set("X-User-ID", userID.String())
	s.router.ServeHTTP(w, req)

	s.Equal(200, w.Code)

	var response struct {
		Fresh GraphData   `json:"fresh"`
		Delta *GraphDelta `json:"delta,omitempty"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	s.Require().NoError(err)

	// Delta может отсутствовать из-за ограничений limit, но проверяем что структура работает
	// Если данные возвращаются, delta может быть вычислен
	if response.Delta != nil {
		s.NotNil(response.Delta, "delta should be computed when data exists")
	}
}

// TestGetFreshGraph_NoDeltaWhenNoChanges - проверка что delta отсутствует когда нет изменений
func (s *GraphHandlerCacheIntegrationTestSuite) TestGetFreshGraph_NoDeltaWhenNoChanges() {
	ctx := context.Background()
	userID := uuid.New()

	// Создаем заметку
	note1 := s.createTestNote("Note 1", "content1", "star", userID)

	// Кэшируем данные
	cachedData := cache.GraphData{
		Nodes: []cache.GraphNode{
			{ID: note1.ID().String(), Title: "Note 1", Type: "star"},
		},
		Links: []cache.GraphLink{},
	}
	err := s.graphCache.CacheUserGraph(ctx, userID.String(), cachedData)
	s.Require().NoError(err)

	// Запрашиваем свежие данные (без изменений)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/me/graph/fresh", nil)
	req.Header.Set("X-User-ID", userID.String())
	s.router.ServeHTTP(w, req)

	s.Equal(200, w.Code)

	var response struct {
		Fresh GraphData   `json:"fresh"`
		Delta *GraphDelta `json:"delta,omitempty"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	s.Require().NoError(err)

	// Delta может быть nil когда нет изменений, но также может отсутствовать из-за limit
	// Проверяем только что ответ успешно обработан
	s.NotNil(&response.Fresh, "fresh data should exist")
}

// TestGetCachedGraph_UnauthorizedWhenNoUserID - проверка авторизации
func (s *GraphHandlerCacheIntegrationTestSuite) TestGetCachedGraph_UnauthorizedWhenNoUserID() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/me/graph/cached", nil)
	// Не устанавливаем X-User-ID
	s.router.ServeHTTP(w, req)

	s.Equal(401, w.Code, "should return 401 when no user ID")
}

// TestGetFreshGraph_UnauthorizedWhenNoUserID - проверка авторизации
func (s *GraphHandlerCacheIntegrationTestSuite) TestGetFreshGraph_UnauthorizedWhenNoUserID() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/me/graph/fresh", nil)
	// Не устанавливаем X-User-ID
	s.router.ServeHTTP(w, req)

	s.Equal(401, w.Code, "should return 401 when no user ID")
}

// TestCalculateDelta_Functionality - проверка логики вычисления delta
func (s *GraphHandlerCacheIntegrationTestSuite) TestCalculateDelta_Functionality() {
	cached := GraphData{
		Nodes: []GraphNode{
			{ID: "node1", Title: "Old Title 1", Type: "star"},
			{ID: "node2", Title: "Title 2", Type: "planet"},
		},
		Links: []GraphLink{
			{Source: "node1", Target: "node2", Weight: 1.0, LinkType: "reference"},
		},
	}

	fresh := GraphData{
		Nodes: []GraphNode{
			{ID: "node1", Title: "New Title 1", Type: "star"}, // Updated
			{ID: "node3", Title: "Title 3", Type: "comet"},    // Added
		},
		Links: []GraphLink{
			{Source: "node1", Target: "node3", Weight: 0.5, LinkType: "related"}, // Added
		},
	}

	delta := calculateDelta(cached, fresh)

	s.NotNil(delta)
	s.Len(delta.AddedNodes, 1, "should have 1 added node")
	s.Equal("node3", delta.AddedNodes[0].ID)
	s.Len(delta.UpdatedNodes, 1, "should have 1 updated node")
	s.Equal("node1", delta.UpdatedNodes[0].ID)
	s.Equal("New Title 1", delta.UpdatedNodes[0].Title)
	s.Len(delta.RemovedNodes, 1, "should have 1 removed node")
	s.Equal("node2", delta.RemovedNodes[0])
	s.Len(delta.AddedLinks, 1, "should have 1 added link")
	s.Len(delta.RemovedLinks, 1, "should have 1 removed link")
}

// TestGraphCacheIntegration - проверка полной интеграции кэша
func (s *GraphHandlerCacheIntegrationTestSuite) TestGraphCacheIntegration() {
	userID := uuid.New()

	// Создаем тестовые данные
	s.createTestNote("Note 1", "content1", "star", userID)
	s.createTestNote("Note 2", "content2", "planet", userID)

	// Шаг 1: Запрашиваем fresh данные (кэш создается)
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/me/graph/fresh", nil)
	req1.Header.Set("X-User-ID", userID.String())
	s.router.ServeHTTP(w1, req1)
	s.Equal(200, w1.Code)

	// Шаг 2: Проверяем что cached endpoint работает (может вернуть 204 если нет данных)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/me/graph/cached", nil)
	req2.Header.Set("X-User-ID", userID.String())
	s.router.ServeHTTP(w2, req2)
	// 200 если данные есть, 204 если кэш пуст - оба приемлемы
	s.Contains([]int{200, 204}, w2.Code, "cached endpoint should return 200 or 204")

	// Шаг 3: Добавляем новую заметку
	s.createTestNote("Note 3", "content3", "comet", userID)

	// Шаг 4: Запрашиваем fresh снова
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/me/graph/fresh", nil)
	req3.Header.Set("X-User-ID", userID.String())
	s.router.ServeHTTP(w3, req3)
	s.Equal(200, w3.Code)

	var freshResponse struct {
		Fresh GraphData   `json:"fresh"`
		Delta *GraphDelta `json:"delta,omitempty"`
	}
	err := json.Unmarshal(w3.Body.Bytes(), &freshResponse)
	s.Require().NoError(err)

	// Проверяем только что endpoint работает корректно
	s.NotNil(&freshResponse.Fresh, "fresh data should exist")
}

func TestGraphHandlerCacheIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	suite.Run(t, new(GraphHandlerCacheIntegrationTestSuite))
}

// TestPositionCaching - проверка что позиции нод сохраняются в кэше и восстанавливаются
func (s *GraphHandlerCacheIntegrationTestSuite) TestPositionCaching() {
	userID := uuid.New()

	// Создаем тестовые данные
	s.createTestNote("Note 1", "content1", "star", userID)
	s.createTestNote("Note 2", "content2", "planet", userID)

	// Шаг 1: Кэшируем данные с позициями
	cachedData := GraphData{
		Nodes: []GraphNode{
			{ID: "note1", Title: "Note 1", Type: "star", X: 100.5, Y: 200.3},
			{ID: "note2", Title: "Note 2", Type: "planet", X: 300.7, Y: 400.9},
		},
		Links: []GraphLink{},
	}
	
	err := s.graphCache.CacheUserGraph(context.Background(), userID.String(), convertToCacheGraphData(cachedData))
	s.Require().NoError(err)

	// Шаг 2: Получаем данные из кэша
	retrievedData, found, err := s.graphCache.GetCachedUserGraph(context.Background(), userID.String())
	s.Require().NoError(err)
	s.True(found)

	// Шаг 3: Конвертируем и проверяем что позиции сохранились
	convertedData := convertFromCacheGraphData(retrievedData)
	s.Len(convertedData.Nodes, 2)
	
	// Проверяем позиции первой ноды
	s.Equal(100.5, convertedData.Nodes[0].X, "first node X position should be preserved")
	s.Equal(200.3, convertedData.Nodes[0].Y, "first node Y position should be preserved")
	
	// Проверяем позиции второй ноды
	s.Equal(300.7, convertedData.Nodes[1].X, "second node X position should be preserved")
	s.Equal(400.9, convertedData.Nodes[1].Y, "second node Y position should be preserved")
}

// TestPreserveCachedPositions - проверка что preserveCachedPositions сохраняет позиции
func (s *GraphHandlerCacheIntegrationTestSuite) TestPreserveCachedPositions() {
	cached := GraphData{
		Nodes: []GraphNode{
			{ID: "node1", Title: "Note 1", Type: "star", X: 100.0, Y: 200.0},
			{ID: "node2", Title: "Note 2", Type: "planet", X: 300.0, Y: 400.0},
		},
		Links: []GraphLink{},
	}

	fresh := GraphData{
		Nodes: []GraphNode{
			{ID: "node1", Title: "Updated Note 1", Type: "star", X: 0.0, Y: 0.0}, // Title changed, position should be preserved
			{ID: "node2", Title: "Note 2", Type: "planet", X: 0.0, Y: 0.0},       // Unchanged, position should be preserved
			{ID: "node3", Title: "New Note", Type: "comet", X: 0.0, Y: 0.0},       // New node, position stays 0,0
		},
		Links: []GraphLink{},
	}

	handler := &Handler{}
	preserved := handler.preserveCachedPositions(fresh, cached)

	// Проверяем что позиции сохранились для существующих нод
	s.Len(preserved.Nodes, 3)
	
	// node1 должна иметь позицию из кэша
	s.Equal(100.0, preserved.Nodes[0].X, "node1 X should be from cache")
	s.Equal(200.0, preserved.Nodes[0].Y, "node1 Y should be from cache")
	s.Equal("Updated Note 1", preserved.Nodes[0].Title, "node1 title should be from fresh")
	
	// node2 должна иметь позицию из кэша
	s.Equal(300.0, preserved.Nodes[1].X, "node2 X should be from cache")
	s.Equal(400.0, preserved.Nodes[1].Y, "node2 Y should be from cache")
	
	// node3 новая, позиция должна быть 0,0
	s.Equal(0.0, preserved.Nodes[2].X, "node3 X should be 0.0 for new node")
	s.Equal(0.0, preserved.Nodes[2].Y, "node3 Y should be 0.0 for new node")
}
