//go:build integration

package linkhandler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// LinkHandlerIntegrationTestSuite - интеграционные тесты для LinkHandler
type LinkHandlerIntegrationTestSuite struct {
	suite.Suite
	db       *gorm.DB
	noteRepo *postgres.NoteRepository
	linkRepo *postgres.LinkRepository
	router   *gin.Engine
	cleanup  func()
}

func (s *LinkHandlerIntegrationTestSuite) SetupSuite() {
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

	// Создаем репозитории
	s.noteRepo = postgres.NewNoteRepository(s.db, nil)
	s.linkRepo = postgres.NewLinkRepository(s.db)

	// Создаем хендлер
	handler := New(s.linkRepo, s.noteRepo, nil, nil, 0)

	// Настраиваем Gin
	gin.SetMode(gin.TestMode)
	s.router = gin.New()

	// Регистрируем маршруты
	s.router.POST("/links", handler.Create)
	s.router.GET("/links/:id", handler.Get)
	s.router.DELETE("/links/:id", handler.Delete)
	s.router.GET("/notes/:id/links", handler.GetByNote)
	s.router.DELETE("/notes/:id/links", handler.DeleteByNote)
}

func (s *LinkHandlerIntegrationTestSuite) TearDownSuite() {
	s.cleanup()
}

func (s *LinkHandlerIntegrationTestSuite) SetupTest() {
	// Очищаем таблицы перед каждым тестом
	err := testutil.TruncateTables(s.db)
	s.Require().NoError(err, "failed to truncate tables")
}

// createTestNote создает тестовую заметку через репозиторий
func (s *LinkHandlerIntegrationTestSuite) createTestNote(title, content, noteType string) *note.Note {
	ctx := s.db.Statement.Context
	noteTitle, _ := note.NewTitle(title)
	noteContent, _ := note.NewContent(content)
	metadata, _ := note.NewMetadata(map[string]interface{}{})
	n := note.NewNote(noteTitle, noteContent, noteType, metadata)
	err := s.noteRepo.Save(ctx, n)
	s.Require().NoError(err, "failed to create test note")
	return n
}

// TestCreateLink_Success - успешное создание связи
func (s *LinkHandlerIntegrationTestSuite) TestCreateLink_Success() {
	source := s.createTestNote("Source Note", "Source content", "star")
	target := s.createTestNote("Target Note", "Target content", "planet")

	reqBody := map[string]interface{}{
		"source_note_id": source.ID().String(),
		"target_note_id": target.ID().String(),
		"link_type":      "reference",
		"weight":         0.8,
		"metadata":       map[string]interface{}{"importance": "high"},
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	s.Equal(201, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.NotEmpty(response["id"])
	s.Equal(source.ID().String(), response["source_note_id"])
	s.Equal(target.ID().String(), response["target_note_id"])
	s.Equal("reference", response["link_type"])
	s.Equal(0.8, response["weight"])
}

// TestCreateLink_MissingSourceNote - несуществующая исходная заметка
func (s *LinkHandlerIntegrationTestSuite) TestCreateLink_MissingSourceNote() {
	target := s.createTestNote("Target Note", "content", "star")

	reqBody := map[string]interface{}{
		"source_note_id": uuid.New().String(),
		"target_note_id": target.ID().String(),
		"link_type":      "reference",
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	s.Equal(404, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Contains(response["error"], "source note not found")
}

// TestCreateLink_MissingTargetNote - несуществующая целевая заметка
func (s *LinkHandlerIntegrationTestSuite) TestCreateLink_MissingTargetNote() {
	source := s.createTestNote("Source Note", "content", "star")

	reqBody := map[string]interface{}{
		"source_note_id": source.ID().String(),
		"target_note_id": uuid.New().String(),
		"link_type":      "reference",
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	s.Equal(404, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Contains(response["error"], "target note not found")
}

// TestCreateLink_MultipleSamePair - создание нескольких связей между одними заметками
// В текущей схеме БД дубликаты разрешены (нет unique constraint на source_id + target_id)
func (s *LinkHandlerIntegrationTestSuite) TestCreateLink_MultipleSamePair() {
	source := s.createTestNote("Source Note", "content", "star")
	target := s.createTestNote("Target Note", "content", "planet")

	reqBody := map[string]interface{}{
		"source_note_id": source.ID().String(),
		"target_note_id": target.ID().String(),
		"link_type":      "reference",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Создаем первую связь
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w1, req1)
	s.Equal(201, w1.Code)

	var resp1 map[string]interface{}
	err := json.Unmarshal(w1.Body.Bytes(), &resp1)
	s.NoError(err)
	linkID1 := resp1["id"].(string)

	// Создаем вторую связь (в текущей схеме дубликаты разрешены)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
	req2.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w2, req2)

	// В текущей реализации API позволяет создавать несколько связей
	// между одними заметками (нет unique constraint)
	s.Equal(201, w2.Code)

	var resp2 map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &resp2)
	s.NoError(err)
	linkID2 := resp2["id"].(string)

	// ID должны быть разными
	s.NotEqual(linkID1, linkID2)
}

// TestCreateLink_InvalidJSON - невалидный JSON
func (s *LinkHandlerIntegrationTestSuite) TestCreateLink_InvalidJSON() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/links", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	s.Equal(400, w.Code)
}

// TestCreateLink_InvalidUUID - невалидный UUID
func (s *LinkHandlerIntegrationTestSuite) TestCreateLink_InvalidUUID() {
	reqBody := map[string]interface{}{
		"source_note_id": "not-a-uuid",
		"target_note_id": "also-not-uuid",
		"link_type":      "reference",
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	s.Equal(400, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Contains(response["error"], "invalid source_note_id")
}

// TestGetLink_Success - получение связи по ID
func (s *LinkHandlerIntegrationTestSuite) TestGetLink_Success() {
	source := s.createTestNote("Source", "content", "star")
	target := s.createTestNote("Target", "content", "planet")

	// Создаем связь
	reqBody := map[string]interface{}{
		"source_note_id": source.ID().String(),
		"target_note_id": target.ID().String(),
		"link_type":      "reference",
	}
	jsonBody, _ := json.Marshal(reqBody)

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w1, req1)
	s.Equal(201, w1.Code)

	var createResponse map[string]interface{}
	err := json.Unmarshal(w1.Body.Bytes(), &createResponse)
	s.NoError(err)
	linkID := createResponse["id"].(string)

	// Получаем связь
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/links/"+linkID, nil)
	s.router.ServeHTTP(w2, req2)

	s.Equal(200, w2.Code)

	var getResponse map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &getResponse)
	s.NoError(err)
	s.Equal(linkID, getResponse["id"])
	s.Equal(source.ID().String(), getResponse["source_note_id"])
	s.Equal(target.ID().String(), getResponse["target_note_id"])
}

// TestGetLink_NotFound - получение несуществующей связи
func (s *LinkHandlerIntegrationTestSuite) TestGetLink_NotFound() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/links/"+uuid.New().String(), nil)
	s.router.ServeHTTP(w, req)

	s.Equal(404, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Contains(response["error"], "link not found")
}

// TestGetLink_InvalidID - невалидный ID
func (s *LinkHandlerIntegrationTestSuite) TestGetLink_InvalidID() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/links/invalid-uuid", nil)
	s.router.ServeHTTP(w, req)

	s.Equal(400, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Contains(response["error"], "invalid id")
}

// TestDeleteLink_Success - удаление связи
func (s *LinkHandlerIntegrationTestSuite) TestDeleteLink_Success() {
	source := s.createTestNote("Source", "content", "star")
	target := s.createTestNote("Target", "content", "planet")

	// Создаем связь
	reqBody := map[string]interface{}{
		"source_note_id": source.ID().String(),
		"target_note_id": target.ID().String(),
		"link_type":      "reference",
	}
	jsonBody, _ := json.Marshal(reqBody)

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w1, req1)
	s.Equal(201, w1.Code)

	var createResponse map[string]interface{}
	err := json.Unmarshal(w1.Body.Bytes(), &createResponse)
	s.NoError(err)
	linkID := createResponse["id"].(string)

	// Удаляем связь
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("DELETE", "/links/"+linkID, nil)
	s.router.ServeHTTP(w2, req2)

	s.Equal(204, w2.Code)

	// Проверяем что связь удалена
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/links/"+linkID, nil)
	s.router.ServeHTTP(w3, req3)
	s.Equal(404, w3.Code)
}

// TestDeleteLink_NotFound - удаление несуществующей связи
func (s *LinkHandlerIntegrationTestSuite) TestDeleteLink_NotFound() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/links/"+uuid.New().String(), nil)
	s.router.ServeHTTP(w, req)

	s.Equal(404, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Contains(response["error"], "link not found")
}

// TestGetLinksByNote - получение связей заметки
func (s *LinkHandlerIntegrationTestSuite) TestGetLinksByNote() {
	note1 := s.createTestNote("Note 1", "content", "star")
	note2 := s.createTestNote("Note 2", "content", "planet")
	note3 := s.createTestNote("Note 3", "content", "comet")

	// Создаем связи
	// note1 -> note2
	reqBody1 := map[string]interface{}{
		"source_note_id": note1.ID().String(),
		"target_note_id": note2.ID().String(),
		"link_type":      "reference",
	}
	jsonBody1, _ := json.Marshal(reqBody1)
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody1))
	req1.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w1, req1)
	s.Equal(201, w1.Code)

	// note3 -> note1 (incoming для note1)
	reqBody2 := map[string]interface{}{
		"source_note_id": note3.ID().String(),
		"target_note_id": note1.ID().String(),
		"link_type":      "related",
	}
	jsonBody2, _ := json.Marshal(reqBody2)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody2))
	req2.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w2, req2)
	s.Equal(201, w2.Code)

	// Получаем связи для note1
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/notes/"+note1.ID().String()+"/links", nil)
	s.router.ServeHTTP(w3, req3)

	s.Equal(200, w3.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w3.Body.Bytes(), &response)
	s.NoError(err)

	outgoing := response["outgoing"].([]interface{})
	incoming := response["incoming"].([]interface{})

	s.Len(outgoing, 1) // note1 -> note2
	s.Len(incoming, 1) // note3 -> note1
}

// TestGetLinksByNote_NotFound - несуществующая заметка
func (s *LinkHandlerIntegrationTestSuite) TestGetLinksByNote_NotFound() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notes/"+uuid.New().String()+"/links", nil)
	s.router.ServeHTTP(w, req)

	s.Equal(404, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Contains(response["error"], "note not found")
}

// TestDeleteLinksByNote - удаление всех связей заметки
func (s *LinkHandlerIntegrationTestSuite) TestDeleteLinksByNote() {
	note1 := s.createTestNote("Note 1", "content", "star")
	note2 := s.createTestNote("Note 2", "content", "planet")
	note3 := s.createTestNote("Note 3", "content", "comet")

	// Создаем несколько связей от note1
	for _, target := range []*note.Note{note2, note3} {
		reqBody := map[string]interface{}{
			"source_note_id": note1.ID().String(),
			"target_note_id": target.ID().String(),
			"link_type":      "reference",
		}
		jsonBody, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		s.router.ServeHTTP(w, req)
		s.Equal(201, w.Code)
	}

	// Удаляем все связи для note1
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/notes/"+note1.ID().String()+"/links", nil)
	s.router.ServeHTTP(w, req)

	s.Equal(204, w.Code)

	// Проверяем что связи удалены
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/notes/"+note1.ID().String()+"/links", nil)
	s.router.ServeHTTP(w2, req2)
	s.Equal(200, w2.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w2.Body.Bytes(), &response)
	s.NoError(err)

	outgoing := response["outgoing"].([]interface{})
	s.Len(outgoing, 0)
}

// TestFullLinkLifecycle - полный цикл: создание → получение → удаление
func (s *LinkHandlerIntegrationTestSuite) TestFullLinkLifecycle() {
	source := s.createTestNote("Source", "content", "star")
	target := s.createTestNote("Target", "content", "planet")

	// 1. Создаем связь
	reqBody := map[string]interface{}{
		"source_note_id": source.ID().String(),
		"target_note_id": target.ID().String(),
		"link_type":      "reference",
		"weight":         0.9,
	}
	jsonBody, _ := json.Marshal(reqBody)

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/links", bytes.NewBuffer(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w1, req1)

	s.Equal(201, w1.Code)

	var createResponse map[string]interface{}
	err := json.Unmarshal(w1.Body.Bytes(), &createResponse)
	s.NoError(err)
	linkID := createResponse["id"].(string)
	s.Equal("reference", createResponse["link_type"])
	s.Equal(0.9, createResponse["weight"])

	// 2. Получаем связь
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/links/"+linkID, nil)
	s.router.ServeHTTP(w2, req2)

	s.Equal(200, w2.Code)

	var getResponse map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &getResponse)
	s.NoError(err)
	s.Equal(linkID, getResponse["id"])

	// 3. Удаляем связь
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("DELETE", "/links/"+linkID, nil)
	s.router.ServeHTTP(w3, req3)

	s.Equal(204, w3.Code)

	// 4. Проверяем что связь удалена
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("GET", "/links/"+linkID, nil)
	s.router.ServeHTTP(w4, req4)

	s.Equal(404, w4.Code)
}

// Запускаем тесты
func TestLinkHandlerIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	suite.Run(t, new(LinkHandlerIntegrationTestSuite))
}
