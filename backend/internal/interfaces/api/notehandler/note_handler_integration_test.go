//go:build integration

package notehandler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"knowledge-graph/internal/config"
	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// NoteHandlerIntegrationTestSuite - интеграционные тесты для NoteHandler
type NoteHandlerIntegrationTestSuite struct {
	suite.Suite
	db      *gorm.DB
	repo    *postgres.NoteRepository
	router  *gin.Engine
	cleanup func()
}

func (s *NoteHandlerIntegrationTestSuite) SetupSuite() {
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

	// Создаем репозиторий
	s.repo = postgres.NewNoteRepository(s.db, nil)

	// Создаем хендлер
	handler := New(
		s.repo,
		nil, // taskQueue - nil для тестов
		nil, // suggestionsHandler
		nil, // affectedNotesSvc
		0,   // taskDelay
		nil, // recRepo
		nil, // embeddingRepo
		nil, // redis
		&config.Config{
			RecommendationTopN:                    10,
			RecommendationFallbackEnabled:         false,
			RecommendationFallbackSemanticEnabled: false,
		},
	)

	// Настраиваем Gin
	gin.SetMode(gin.TestMode)
	s.router = gin.New()

	// Регистрируем маршруты
	s.router.POST("/notes", handler.Create)
	s.router.GET("/notes", handler.List)
	s.router.GET("/notes/:id", handler.Get)
	s.router.PUT("/notes/:id", handler.Update)
	s.router.DELETE("/notes/:id", handler.Delete)
	s.router.GET("/notes/:id/suggestions", handler.GetSuggestions)
	s.router.GET("/notes/search", handler.Search)
}

func (s *NoteHandlerIntegrationTestSuite) TearDownSuite() {
	s.cleanup()
}

func (s *NoteHandlerIntegrationTestSuite) SetupTest() {
	// Очищаем таблицы перед каждым тестом
	err := testutil.TruncateTables(s.db)
	s.Require().NoError(err, "failed to truncate tables")
}

// createTestNote создает тестовую заметку через API
func (s *NoteHandlerIntegrationTestSuite) createTestNote(title, content, noteType string) string {
	reqBody := map[string]interface{}{
		"title":    title,
		"content":  content,
		"type":     noteType,
		"metadata": map[string]interface{}{},
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	s.Equal(201, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)

	return response["id"].(string)
}

// TestCreateNote - создание заметки
func (s *NoteHandlerIntegrationTestSuite) TestCreateNote() {
	reqBody := map[string]interface{}{
		"title":   "Test Note",
		"content": "Test content",
		"type":    "star",
		"metadata": map[string]interface{}{
			"source": "test",
		},
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, req)

	s.Equal(201, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.NotEmpty(response["id"])
	s.Equal("Test Note", response["title"])
	s.Equal("Test content", response["content"])
	s.Equal("star", response["type"])
}

// TestCreateNote_InvalidJSON - невалидный JSON
func (s *NoteHandlerIntegrationTestSuite) TestCreateNote_InvalidJSON() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notes", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, req)

	s.Equal(400, w.Code)
}

// TestCreateNote_EmptyTitle - пустой заголовок
func (s *NoteHandlerIntegrationTestSuite) TestCreateNote_EmptyTitle() {
	reqBody := map[string]interface{}{
		"title":   "",
		"content": "content",
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, req)

	s.Equal(400, w.Code)
}

// TestGetNote - получение заметки
func (s *NoteHandlerIntegrationTestSuite) TestGetNote() {
	id := s.createTestNote("Get Me", "content", "star")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notes/"+id, nil)
	s.router.ServeHTTP(w, req)

	s.Equal(200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal(id, response["id"])
	s.Equal("Get Me", response["title"])
}

// TestGetNote_NotFound - получение несуществующей заметки
func (s *NoteHandlerIntegrationTestSuite) TestGetNote_NotFound() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notes/"+uuid.New().String(), nil)
	s.router.ServeHTTP(w, req)

	s.Equal(404, w.Code)
}

// TestGetNote_InvalidID - невалидный ID
func (s *NoteHandlerIntegrationTestSuite) TestGetNote_InvalidID() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notes/invalid-uuid", nil)
	s.router.ServeHTTP(w, req)

	s.Equal(400, w.Code)
}

// TestUpdateNote - обновление заметки
func (s *NoteHandlerIntegrationTestSuite) TestUpdateNote() {
	id := s.createTestNote("Original", "content", "star")

	reqBody := map[string]interface{}{
		"title":   "Updated Title",
		"content": "Updated content",
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/notes/"+id, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	s.Equal(200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("Updated Title", response["title"])
	s.Equal("Updated content", response["content"])
}

// TestUpdateNote_NotFound - обновление несуществующей
func (s *NoteHandlerIntegrationTestSuite) TestUpdateNote_NotFound() {
	reqBody := map[string]interface{}{
		"title": "Updated",
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/notes/"+uuid.New().String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	s.Equal(404, w.Code)
}

// TestDeleteNote - удаление заметки
func (s *NoteHandlerIntegrationTestSuite) TestDeleteNote() {
	id := s.createTestNote("Delete Me", "content", "star")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/notes/"+id, nil)
	s.router.ServeHTTP(w, req)

	s.Equal(204, w.Code)

	// Проверяем что удалена
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/notes/"+id, nil)
	s.router.ServeHTTP(w2, req2)
	s.Equal(404, w2.Code)
}

// TestDeleteNote_NotFound - удаление несуществующей
func (s *NoteHandlerIntegrationTestSuite) TestDeleteNote_NotFound() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/notes/"+uuid.New().String(), nil)
	s.router.ServeHTTP(w, req)

	s.Equal(404, w.Code)
}

// TestListNotes - список заметок
func (s *NoteHandlerIntegrationTestSuite) TestListNotes() {
	// Создаем несколько заметок
	s.createTestNote("Note 1", "content 1", "star")
	s.createTestNote("Note 2", "content 2", "planet")
	s.createTestNote("Note 3", "content 3", "comet")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notes?limit=2&offset=0", nil)
	s.router.ServeHTTP(w, req)

	s.Equal(200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.NotNil(response["notes"])
	s.Equal(float64(3), response["total"])
	s.Equal(float64(2), response["limit"])
}

// TestSearchNotes - поиск заметок
func (s *NoteHandlerIntegrationTestSuite) TestSearchNotes() {
	// Создаем заметки
	ctx := s.db.Statement.Context
	title, _ := note.NewTitle("Golang Tutorial")
	content, _ := note.NewContent("Learn Go programming")
	metadata, _ := note.NewMetadata(map[string]interface{}{})
	n := note.NewNote(title, content, "star", metadata)
	s.repo.Save(ctx, n)

	title2, _ := note.NewTitle("Python Guide")
	content2, _ := note.NewContent("Learn Python programming")
	n2 := note.NewNote(title2, content2, "star", metadata)
	s.repo.Save(ctx, n2)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notes/search?q=golang&page=1&size=10", nil)
	s.router.ServeHTTP(w, req)

	s.Equal(200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.NotNil(response["data"])
	s.NotNil(response["total"])
}

// TestGetSuggestions - получение рекомендаций
func (s *NoteHandlerIntegrationTestSuite) TestGetSuggestions() {
	id := s.createTestNote("Suggestions Test", "content", "star")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notes/"+id+"/suggestions?limit=5", nil)
	s.router.ServeHTTP(w, req)

	// Должен вернуть 202 Accepted (нет данных, но запрос принят)
	s.Equal(202, w.Code)

	// Проверяем заголовок
	s.Equal("true", w.Header().Get("X-Recommendations-Stale"))
}

// TestGetSuggestions_InvalidLimit - невалидный limit
func (s *NoteHandlerIntegrationTestSuite) TestGetSuggestions_InvalidLimit() {
	id := s.createTestNote("Test", "content", "star")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notes/"+id+"/suggestions?limit=abc", nil)
	s.router.ServeHTTP(w, req)

	// Должен использовать дефолтное значение
	s.Equal(202, w.Code)
}

// TestCreateAndGet_FullFlow - полный цикл создания и получения
func (s *NoteHandlerIntegrationTestSuite) TestCreateAndGet_FullFlow() {
	// Создаем
	reqBody := map[string]interface{}{
		"title":   "Full Flow Test",
		"content": "Testing complete flow",
		"type":    "galaxy",
		"metadata": map[string]interface{}{
			"priority": "high",
		},
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	s.Equal(201, w.Code)

	var createResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &createResponse)
	s.NoError(err)

	id := createResponse["id"].(string)

	// Получаем
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/notes/"+id, nil)
	s.router.ServeHTTP(w2, req2)

	s.Equal(200, w2.Code)

	var getResponse map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &getResponse)
	s.NoError(err)
	s.Equal("Full Flow Test", getResponse["title"])
	s.Equal("galaxy", getResponse["type"])
}

// Запускаем тесты
func TestNoteHandlerIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	suite.Run(t, new(NoteHandlerIntegrationTestSuite))
}
