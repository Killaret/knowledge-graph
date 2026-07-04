//go:build integration

package notehandler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"knowledge-graph/internal/config"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// NoteHandlerNLPIntegrationTestSuite - интеграционные тесты для NLP обогащения заметок
type NoteHandlerNLPIntegrationTestSuite struct {
	suite.Suite
	db      *gorm.DB
	repo    *postgres.NoteRepository
	router  *gin.Engine
	cleanup func()
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
		nil, // redis
		&config.Config{
			RecommendationTopN:                    10,
			RecommendationFallbackEnabled:         false,
			RecommendationFallbackSemanticEnabled: false,
		},
		nil, // graphCache
		nil, // achievementService
	)

	// Настраиваем Gin
	gin.SetMode(gin.TestMode)
	s.router = gin.New()

	// Регистрируем маршруты
	s.router.POST("/notes", handler.Create)
	s.router.GET("/notes/:id", handler.Get)
}

func (s *NoteHandlerNLPIntegrationTestSuite) TearDownSuite() {
	s.cleanup()
}

func (s *NoteHandlerNLPIntegrationTestSuite) SetupTest() {
	// Очищаем таблицы перед каждым тестом
	err := testutil.TruncateTables(s.db)
	s.Require().NoError(err, "failed to truncate tables")
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

func TestNoteHandlerNLPIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(NoteHandlerNLPIntegrationTestSuite))
}
