//go:build integration

package taghandler

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
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// TagHandlerIntegrationTestSuite - интеграционные тесты для TagHandler
type TagHandlerIntegrationTestSuite struct {
	suite.Suite
	db       *gorm.DB
	noteRepo *postgres.NoteRepository
	tagRepo  *postgres.TagRepository
	router   *gin.Engine
	cleanup  func()
}

func (s *TagHandlerIntegrationTestSuite) SetupSuite() {
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
	s.tagRepo = postgres.NewTagRepository(s.db)

	// Создаем хендлер
	handler := New(s.tagRepo, s.noteRepo)

	// Настраиваем Gin
	gin.SetMode(gin.TestMode)
	s.router = gin.New()

	// Регистрируем маршруты тегов
	s.router.POST("/tags", handler.Create)
	s.router.GET("/tags", handler.List)
	s.router.GET("/tags/:id", handler.Get)
	s.router.PUT("/tags/:id", handler.Update)
	s.router.DELETE("/tags/:id", handler.Delete)

	// Регистрируем маршруты привязки тегов к заметкам
	s.router.POST("/notes/:id/tags", handler.AddTagToNote)
	s.router.DELETE("/notes/:id/tags/:tagId", handler.RemoveTagFromNote)
	s.router.GET("/notes/:id/tags", handler.GetTagsByNote)
}

func (s *TagHandlerIntegrationTestSuite) TearDownSuite() {
	s.cleanup()
}

func (s *TagHandlerIntegrationTestSuite) SetupTest() {
	// Очищаем таблицы перед каждым тестом
	err := testutil.TruncateTables(s.db)
	s.Require().NoError(err, "failed to truncate tables")
}

// createTestNote создает тестовую заметку
func (s *TagHandlerIntegrationTestSuite) createTestNote(title, content, noteType string) *note.Note {
	ctx := s.db.Statement.Context
	noteTitle, _ := note.NewTitle(title)
	noteContent, _ := note.NewContent(content)
	metadata, _ := note.NewMetadata(map[string]interface{}{})
	n := note.NewNote(noteTitle, noteContent, noteType, metadata)
	err := s.noteRepo.Save(ctx, n)
	s.Require().NoError(err, "failed to create test note")
	return n
}

// createTestTag создает тестовый тег
func (s *TagHandlerIntegrationTestSuite) createTestTag(name string) *postgres.TagModel {
	tag := &postgres.TagModel{
		Name: name,
	}
	err := s.tagRepo.Create(s.db.Statement.Context, tag)
	s.Require().NoError(err, "failed to create test tag")
	return tag
}

// TestCreateTag_Success - успешное создание тега
func (s *TagHandlerIntegrationTestSuite) TestCreateTag_Success() {
	reqBody := map[string]interface{}{
		"name": "important",
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/tags", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)

	var response TagResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.NotEmpty(response.ID)
	s.Equal("important", response.Name)
}

// TestCreateTag_Duplicate - дубликат имени тега
func (s *TagHandlerIntegrationTestSuite) TestCreateTag_Duplicate() {
	// Создаем первый тег
	s.createTestTag("unique-tag")

	// Пытаемся создать дубликат
	reqBody := map[string]interface{}{
		"name": "unique-tag",
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/tags", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusConflict, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Contains(response["error"], "already exists")
}

// TestCreateTag_InvalidJSON - невалидный JSON
func (s *TagHandlerIntegrationTestSuite) TestCreateTag_InvalidJSON() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/tags", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
}

// TestCreateTag_EmptyName - пустое имя тега
func (s *TagHandlerIntegrationTestSuite) TestCreateTag_EmptyName() {
	reqBody := map[string]interface{}{
		"name": "",
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/tags", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
}

// TestListTags - получение списка тегов
func (s *TagHandlerIntegrationTestSuite) TestListTags() {
	// Создаем несколько тегов
	s.createTestTag("tag-1")
	s.createTestTag("tag-2")
	s.createTestTag("tag-3")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tags", nil)
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var response []TagResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Len(response, 3)
}

// TestListTags_Empty - пустой список тегов
func (s *TagHandlerIntegrationTestSuite) TestListTags_Empty() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tags", nil)
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var response []TagResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Len(response, 0)
}

// TestGetTag_Success - получение тега по ID
func (s *TagHandlerIntegrationTestSuite) TestGetTag_Success() {
	tag := s.createTestTag("test-tag")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tags/"+tag.ID.String(), nil)
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var response TagResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal(tag.ID.String(), response.ID)
	s.Equal("test-tag", response.Name)
}

// TestGetTag_NotFound - несуществующий тег
func (s *TagHandlerIntegrationTestSuite) TestGetTag_NotFound() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tags/550e8400-e29b-41d4-a716-446655440000", nil)
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Contains(response["error"], "tag not found")
}

// TestGetTag_InvalidID - невалидный UUID
func (s *TagHandlerIntegrationTestSuite) TestGetTag_InvalidID() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tags/invalid-uuid", nil)
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Contains(response["error"], "invalid tag id")
}

// TestUpdateTag_Success - успешное обновление тега
func (s *TagHandlerIntegrationTestSuite) TestUpdateTag_Success() {
	tag := s.createTestTag("old-name")

	reqBody := map[string]interface{}{
		"name": "new-name",
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/tags/"+tag.ID.String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var response TagResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal(tag.ID.String(), response.ID)
	s.Equal("new-name", response.Name)
}

// TestUpdateTag_Duplicate - обновление на существующее имя
func (s *TagHandlerIntegrationTestSuite) TestUpdateTag_Duplicate() {
	tag1 := s.createTestTag("first-tag")
	s.createTestTag("second-tag")

	// Пытаемся переименовать tag1 в "second-tag"
	reqBody := map[string]interface{}{
		"name": "second-tag",
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/tags/"+tag1.ID.String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusConflict, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Contains(response["error"], "already exists")
}

// TestUpdateTag_NotFound - обновление несуществующего тега
func (s *TagHandlerIntegrationTestSuite) TestUpdateTag_NotFound() {
	reqBody := map[string]interface{}{
		"name": "new-name",
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/tags/550e8400-e29b-41d4-a716-446655440000", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

// TestDeleteTag_Success - успешное удаление тега
func (s *TagHandlerIntegrationTestSuite) TestDeleteTag_Success() {
	tag := s.createTestTag("delete-me")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/tags/"+tag.ID.String(), nil)
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNoContent, w.Code)

	// Проверяем что тег удален
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/tags/"+tag.ID.String(), nil)
	s.router.ServeHTTP(w2, req2)
	s.Equal(http.StatusNotFound, w2.Code)
}

// TestDeleteTag_NotFound - удаление несуществующего тега
func (s *TagHandlerIntegrationTestSuite) TestDeleteTag_NotFound() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/tags/550e8400-e29b-41d4-a716-446655440000", nil)
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

// TestAddTagToNote_Success - успешная привязка тега к заметке
func (s *TagHandlerIntegrationTestSuite) TestAddTagToNote_Success() {
	note := s.createTestNote("Test Note", "content", "star")
	tag := s.createTestTag("important")

	reqBody := map[string]interface{}{
		"tag_id": tag.ID.String(),
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notes/"+note.ID().String()+"/tags", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)
}

// TestAddTagToNote_NoteNotFound - несуществующая заметка
func (s *TagHandlerIntegrationTestSuite) TestAddTagToNote_NoteNotFound() {
	tag := s.createTestTag("important")

	reqBody := map[string]interface{}{
		"tag_id": tag.ID.String(),
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notes/550e8400-e29b-41d4-a716-446655440000/tags", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

// TestAddTagToNote_TagNotFound - несуществующий тег
func (s *TagHandlerIntegrationTestSuite) TestAddTagToNote_TagNotFound() {
	note := s.createTestNote("Test Note", "content", "star")

	reqBody := map[string]interface{}{
		"tag_id": "550e8400-e29b-41d4-a716-446655440000",
	}
	jsonBody, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notes/"+note.ID().String()+"/tags", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

// TestAddTagToNote_AlreadyAssigned - повторная привязка того же тега
func (s *TagHandlerIntegrationTestSuite) TestAddTagToNote_AlreadyAssigned() {
	note := s.createTestNote("Test Note", "content", "star")
	tag := s.createTestTag("important")

	// Первая привязка
	reqBody := map[string]interface{}{
		"tag_id": tag.ID.String(),
	}
	jsonBody, _ := json.Marshal(reqBody)

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/notes/"+note.ID().String()+"/tags", bytes.NewBuffer(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w1, req1)
	s.Equal(http.StatusCreated, w1.Code)

	// Вторая попытка привязки - конфликт
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/notes/"+note.ID().String()+"/tags", bytes.NewBuffer(jsonBody))
	req2.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(w2, req2)

	s.Equal(http.StatusConflict, w2.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w2.Body.Bytes(), &response)
	s.NoError(err)
	s.Contains(response["error"], "already assigned")
}

// TestRemoveTagFromNote_Success - успешное удаление привязки
func (s *TagHandlerIntegrationTestSuite) TestRemoveTagFromNote_Success() {
	note := s.createTestNote("Test Note", "content", "star")
	tag := s.createTestTag("important")

	// Привязываем тег
	err := s.tagRepo.AddTagToNote(s.db.Statement.Context, note.ID(), tag.ID)
	s.Require().NoError(err)

	// Отвязываем тег
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/notes/"+note.ID().String()+"/tags/"+tag.ID.String(), nil)
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNoContent, w.Code)
}

// TestGetTagsByNote - получение тегов заметки
func (s *TagHandlerIntegrationTestSuite) TestGetTagsByNote() {
	note := s.createTestNote("Test Note", "content", "star")
	tag1 := s.createTestTag("important")
	tag2 := s.createTestTag("urgent")

	// Привязываем оба тега
	err := s.tagRepo.AddTagToNote(s.db.Statement.Context, note.ID(), tag1.ID)
	s.Require().NoError(err)
	err = s.tagRepo.AddTagToNote(s.db.Statement.Context, note.ID(), tag2.ID)
	s.Require().NoError(err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notes/"+note.ID().String()+"/tags", nil)
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var response []TagResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Len(response, 2)

	// Проверяем имена тегов
	names := make(map[string]bool)
	for _, t := range response {
		names[t.Name] = true
	}
	s.True(names["important"])
	s.True(names["urgent"])
}

// TestGetTagsByNote_NotFound - несуществующая заметка
func (s *TagHandlerIntegrationTestSuite) TestGetTagsByNote_NotFound() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notes/550e8400-e29b-41d4-a716-446655440000/tags", nil)
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

// TestGetTagsByNote_Empty - заметка без тегов
func (s *TagHandlerIntegrationTestSuite) TestGetTagsByNote_Empty() {
	note := s.createTestNote("Test Note", "content", "star")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notes/"+note.ID().String()+"/tags", nil)
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var response []TagResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Len(response, 0)
}

// Запускаем тесты
func TestTagHandlerIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	suite.Run(t, new(TagHandlerIntegrationTestSuite))
}
