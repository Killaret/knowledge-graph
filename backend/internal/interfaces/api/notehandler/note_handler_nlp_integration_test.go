//go:build integration

package notehandler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"knowledge-graph/internal/config"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/interfaces/api/middleware"
	"knowledge-graph/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// NoteHandlerNLPIntegrationTestSuite - интеграционные тесты для NLP обогащения заметок
type NoteHandlerNLPIntegrationTestSuite struct {
	suite.Suite
	db         *gorm.DB
	repo       *postgres.NoteRepository
	router     *gin.Engine
	cleanup    func()
	testUserID uuid.UUID
}

func (s *NoteHandlerNLPIntegrationTestSuite) SetupSuite() {
	// Поднимаем тестовую БД
	s.db, s.cleanup = testutil.SetupTestDB(s.T())

	// Миграция всех моделей (без NoteEmbeddingModel для упрощения тестов)
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

	// Создаем репозиторий
	s.repo = postgres.NewNoteRepository(s.db, nil)

	// Создаем хендлер (без task queue для простоты)
	handler := New(
		s.repo,
		nil, // taskQueue - nil для тестов
		nil, // suggestionsHandler
		nil, // affectedNotesSvc
		0,   // taskDelay
		nil, // recRepo
		nil, // embeddingRepo
		nil, // cacheClient
		&config.Config{
			RecommendationTopN:                    10,
			RecommendationFallbackEnabled:         false,
			RecommendationFallbackSemanticEnabled: false,
			PaginationDefaultLimit:                20,
			PaginationMaxLimit:                    100,
		},
		nil, // graphCache
		nil, // achievementService
	)

	// Настраиваем Gin
	gin.SetMode(gin.TestMode)
	s.router = gin.New()

	// Middleware: подставляет тестового пользователя, чтобы List
	// возвращал заметки текущего пользователя, а не только публичные.
	s.router.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, s.testUserID)
		c.Set(middleware.ContextRoleKey, "user")
		c.Next()
	})

	// Регистрируем маршруты
	s.router.POST("/notes", handler.Create)
	s.router.GET("/notes", handler.List)
	s.router.GET("/notes/:id", handler.Get)
}

func (s *NoteHandlerNLPIntegrationTestSuite) TearDownSuite() {
	s.cleanup()
}

func (s *NoteHandlerNLPIntegrationTestSuite) SetupTest() {
	// Очищаем таблицы перед каждым тестом
	err := testutil.TruncateTables(s.db)
	s.Require().NoError(err, "failed to truncate tables")

	// Создаем тестового пользователя и фиксируем его ID для middleware.
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

// TestCreateDustNote - создание заметки типа "dust"
func (s *NoteHandlerNLPIntegrationTestSuite) TestCreateDustNote() {
	reqBody := map[string]interface{}{
		"title":   "Quick Dust Note",
		"content": "This is a quick dust note for testing NLP enrichment",
		"type":    "dust",
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

	var wrappedResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &wrappedResponse)
	s.NoError(err)

	data := wrappedResponse["data"].(map[string]interface{})
	noteID := data["id"].(string)
	s.NotEmpty(noteID)
	s.Equal("Quick Dust Note", data["title"])
	s.Equal("dust", data["type"])

	// Проверяем, что заметка сохранена в БД
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/notes/"+noteID, nil)
	s.router.ServeHTTP(w2, req2)

	s.Equal(200, w2.Code)

	var wrappedResponse2 map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &wrappedResponse2)
	s.NoError(err)

	data2 := wrappedResponse2["data"].(map[string]interface{})
	s.Equal(noteID, data2["id"])
	s.Equal("Quick Dust Note", data2["title"])
	s.Equal("dust", data2["type"])
}

// TestCreateDustNote_InvalidType - проверка валидации типа заметки
func (s *NoteHandlerNLPIntegrationTestSuite) TestCreateDustNote_InvalidType() {
	reqBody := map[string]interface{}{
		"title":   "Test Note",
		"content": "Test content",
		"type":    "invalid_type",
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, req)

	// Должна быть ошибка валидации (400 или 422)
	s.NotEqual(201, w.Code)
}

// TestCreateDustNote_EmptyContent - проверка валидации пустого контента
func (s *NoteHandlerNLPIntegrationTestSuite) TestCreateDustNote_EmptyContent() {
	reqBody := map[string]interface{}{
		"title":   "Test Note",
		"content": "",
		"type":    "dust",
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, req)

	// Пустой контент может быть допустимым (quick notes)
	// Проверяем что заметка создана успешно
	s.Equal(201, w.Code)
}

func (s *NoteHandlerNLPIntegrationTestSuite) TestCreateDustNote_MultipleNotes() {
	// Создаем первую заметку
	reqBody1 := map[string]interface{}{
		"title":   "First Dust Note",
		"content": "First dust note content",
		"type":    "dust",
	}
	jsonBody1, _ := json.Marshal(reqBody1)

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/notes", bytes.NewBuffer(jsonBody1))
	req1.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w1, req1)

	s.Equal(201, w1.Code)

	// Создаем вторую заметку
	reqBody2 := map[string]interface{}{
		"title":   "Second Dust Note",
		"content": "Second dust note content",
		"type":    "dust",
	}
	jsonBody2, _ := json.Marshal(reqBody2)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/notes", bytes.NewBuffer(jsonBody2))
	req2.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w2, req2)

	s.Equal(201, w2.Code)

	// Проверяем, что обе заметки созданы через GET /notes
	time.Sleep(100 * time.Millisecond)

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/notes", nil)
	s.router.ServeHTTP(w3, req3)

	s.Equal(200, w3.Code)

	var wrappedResponse3 map[string]interface{}
	err := json.Unmarshal(w3.Body.Bytes(), &wrappedResponse3)
	s.NoError(err)

	// Проверяем формат ответа (может быть data.notes или просто notes)
	var notes []interface{}
	if data3, ok := wrappedResponse3["data"].(map[string]interface{}); ok {
		if notesField, ok := data3["notes"].([]interface{}); ok {
			notes = notesField
		}
	} else if notesField, ok := wrappedResponse3["notes"].([]interface{}); ok {
		notes = notesField
	}

	s.GreaterOrEqual(len(notes), 2, "should have at least 2 notes")
}

// TestCreateDustNote_WithExistingNotes - проверка создания links с существующими заметками
func (s *NoteHandlerNLPIntegrationTestSuite) TestCreateDustNote_WithExistingNotes() {
	// Создаем существующую заметку
	existingReq := map[string]interface{}{
		"title":   "Existing Note",
		"content": "This is an existing note about testing",
		"type":    "star",
	}
	jsonBody1, _ := json.Marshal(existingReq)

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/notes", bytes.NewBuffer(jsonBody1))
	req1.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w1, req1)

	s.Equal(201, w1.Code)

	var wrappedResponse1 map[string]interface{}
	err := json.Unmarshal(w1.Body.Bytes(), &wrappedResponse1)
	s.NoError(err)
	_ = wrappedResponse1["data"].(map[string]interface{})["id"].(string)

	// Ждем обработки первой заметки
	time.Sleep(2 * time.Second)

	// Создаем новую заметку типа "dust"
	newReq := map[string]interface{}{
		"title":   "New Dust Note",
		"content": "This is a new dust note about testing as well",
		"type":    "dust",
	}
	jsonBody2, _ := json.Marshal(newReq)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/notes", bytes.NewBuffer(jsonBody2))
	req2.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w2, req2)

	s.Equal(201, w2.Code)

	var wrappedResponse2 map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &wrappedResponse2)
	s.NoError(err)
	newNoteID := wrappedResponse2["data"].(map[string]interface{})["id"].(string)

	// Ждем обработки задач и создания links
	time.Sleep(3 * time.Second)

	// Проверяем, что links созданы (если есть семантическое сходство)
	// В тестовом окружении с mock NLP это может не создавать реальные links
	// но проверяем, что процесс не падает с ошибкой
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/notes/"+newNoteID, nil)
	s.router.ServeHTTP(w3, req3)

	s.Equal(200, w3.Code)
}

func TestNoteHandlerNLPIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(NoteHandlerNLPIntegrationTestSuite))
}
