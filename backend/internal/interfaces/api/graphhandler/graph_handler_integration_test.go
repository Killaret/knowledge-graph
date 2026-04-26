//go:build integration

package graphhandler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// GraphHandlerIntegrationTestSuite - интеграционные тесты для GraphHandler
type GraphHandlerIntegrationTestSuite struct {
	suite.Suite
	db       *gorm.DB
	noteRepo *postgres.NoteRepository
	linkRepo *postgres.LinkRepository
	router   *gin.Engine
	handler  *Handler
	cleanup  func()
}

func (s *GraphHandlerIntegrationTestSuite) SetupSuite() {
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

	// Создаем хендлер с maxDepth = 3
	s.handler = New(s.noteRepo, s.linkRepo, 3)

	// Настраиваем Gin
	gin.SetMode(gin.TestMode)
	s.router = gin.New()

	// Регистрируем маршруты
	s.router.GET("/notes/:id/graph", s.handler.GetGraph)
	s.router.GET("/graph/all", s.handler.GetFullGraph)
}

func (s *GraphHandlerIntegrationTestSuite) TearDownSuite() {
	s.cleanup()
}

func (s *GraphHandlerIntegrationTestSuite) SetupTest() {
	// Очищаем таблицы перед каждым тестом
	err := testutil.TruncateTables(s.db)
	s.Require().NoError(err, "failed to truncate tables")
}

// createTestNote создает тестовую заметку
func (s *GraphHandlerIntegrationTestSuite) createTestNote(title, content, noteType string) *note.Note {
	ctx := s.db.Statement.Context
	noteTitle, _ := note.NewTitle(title)
	noteContent, _ := note.NewContent(content)
	metadata, _ := note.NewMetadata(map[string]interface{}{"type": noteType})
	n := note.NewNote(noteTitle, noteContent, noteType, metadata)
	err := s.noteRepo.Save(ctx, n)
	s.Require().NoError(err, "failed to create test note")
	return n
}

// createTestLink создает тестовую связь между заметками
func (s *GraphHandlerIntegrationTestSuite) createTestLink(source, target *note.Note, linkType string) *link.Link {
	ctx := s.db.Statement.Context
	lt, _ := link.NewLinkType(linkType)
	w, _ := link.NewWeight(1.0)
	m, _ := link.NewMetadata(map[string]interface{}{})
	l := link.NewLink(source.ID(), target.ID(), lt, w, m)
	err := s.linkRepo.Save(ctx, l)
	s.Require().NoError(err, "failed to create test link")
	return l
}

// TestGetGraph_Success - получение графа вокруг заметки
func (s *GraphHandlerIntegrationTestSuite) TestGetGraph_Success() {
	// Создаем тестовые заметки: center -> child1, child2
	center := s.createTestNote("Center Note", "content", "star")
	child1 := s.createTestNote("Child 1", "content", "planet")
	child2 := s.createTestNote("Child 2", "content", "planet")

	// Создаем связи
	s.createTestLink(center, child1, "reference")
	s.createTestLink(center, child2, "related")

	// Получаем граф
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notes/"+center.ID().String()+"/graph", nil)
	s.router.ServeHTTP(w, req)

	s.Equal(200, w.Code)

	var response GraphData
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)

	// Проверяем структуру ответа
	s.NotEmpty(response.Nodes)
	s.Len(response.Nodes, 3) // center + 2 children
	s.Len(response.Links, 2) // 2 связи

	// Проверяем что center присутствует
	foundCenter := false
	for _, n := range response.Nodes {
		if n.ID == center.ID().String() {
			foundCenter = true
			s.Equal("Center Note", n.Title)
			s.Equal("star", n.Type)
			break
		}
	}
	s.True(foundCenter, "center node should be in response")
}

// TestGetGraph_WithDepthParam - проверка query-параметра depth
func (s *GraphHandlerIntegrationTestSuite) TestGetGraph_WithDepthParam() {
	// Создаем цепочку: A -> B -> C -> D
	noteA := s.createTestNote("Note A", "content", "star")
	noteB := s.createTestNote("Note B", "content", "planet")
	noteC := s.createTestNote("Note C", "content", "planet")
	noteD := s.createTestNote("Note D", "content", "comet")

	s.createTestLink(noteA, noteB, "reference")
	s.createTestLink(noteB, noteC, "reference")
	s.createTestLink(noteC, noteD, "reference")

	// Запрос с depth=1 (только A -> B)
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/notes/"+noteA.ID().String()+"/graph?depth=1", nil)
	s.router.ServeHTTP(w1, req1)

	s.Equal(200, w1.Code)

	var resp1 GraphData
	err := json.Unmarshal(w1.Body.Bytes(), &resp1)
	s.NoError(err)

	// С depth=1 должны получить только A и B
	s.Len(resp1.Nodes, 2)
	s.Len(resp1.Links, 1)

	// Запрос с depth=2 (A -> B -> C)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/notes/"+noteA.ID().String()+"/graph?depth=2", nil)
	s.router.ServeHTTP(w2, req2)

	s.Equal(200, w2.Code)

	var resp2 GraphData
	err = json.Unmarshal(w2.Body.Bytes(), &resp2)
	s.NoError(err)

	// С depth=2 должны получить A, B, C
	s.Len(resp2.Nodes, 3)
	s.Len(resp2.Links, 2)
}

// TestGetGraph_WithDepthExceedMax - depth не может превысить maxDepth хендлера
func (s *GraphHandlerIntegrationTestSuite) TestGetGraph_WithDepthExceedMax() {
	// Создаем цепочку: A -> B -> C -> D -> E
	noteA := s.createTestNote("Note A", "content", "star")
	noteB := s.createTestNote("Note B", "content", "planet")
	noteC := s.createTestNote("Note C", "content", "planet")
	noteD := s.createTestNote("Note D", "content", "comet")
	noteE := s.createTestNote("Note E", "content", "comet")

	s.createTestLink(noteA, noteB, "reference")
	s.createTestLink(noteB, noteC, "reference")
	s.createTestLink(noteC, noteD, "reference")
	s.createTestLink(noteD, noteE, "reference")

	// Запрос с depth=5, но maxDepth хендлера = 3
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notes/"+noteA.ID().String()+"/graph?depth=5", nil)
	s.router.ServeHTTP(w, req)

	s.Equal(200, w.Code)

	var response GraphData
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)

	// Глубина должна быть ограничена maxDepth=3, поэтому получаем A, B, C, D
	s.Len(response.Nodes, 4)
	s.Len(response.Links, 3)
}

// TestGetGraph_EmptyGraph - заметка без связей
func (s *GraphHandlerIntegrationTestSuite) TestGetGraph_EmptyGraph() {
	note := s.createTestNote("Isolated Note", "content", "star")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notes/"+note.ID().String()+"/graph", nil)
	s.router.ServeHTTP(w, req)

	s.Equal(200, w.Code)

	var response GraphData
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)

	// Должна вернуться только сама заметка
	s.Len(response.Nodes, 1)
	s.Len(response.Links, 0)
	s.Equal(note.ID().String(), response.Nodes[0].ID)
}

// TestGetGraph_InvalidID - невалидный UUID
func (s *GraphHandlerIntegrationTestSuite) TestGetGraph_InvalidID() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notes/invalid-uuid/graph", nil)
	s.router.ServeHTTP(w, req)

	s.Equal(400, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Contains(response["error"], "invalid id")
}

// TestGetGraph_InvalidDepth - невалидный depth
func (s *GraphHandlerIntegrationTestSuite) TestGetGraph_InvalidDepth() {
	note := s.createTestNote("Note", "content", "star")

	// Невалидный depth (строка) - должен использовать значение по умолчанию
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notes/"+note.ID().String()+"/graph?depth=invalid", nil)
	s.router.ServeHTTP(w, req)

	s.Equal(200, w.Code) // Должен вернуть 200 с fallback на maxDepth

	var response GraphData
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Len(response.Nodes, 1)

	// Отрицательный depth
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/notes/"+note.ID().String()+"/graph?depth=-1", nil)
	s.router.ServeHTTP(w2, req2)

	s.Equal(200, w2.Code) // Отрицательный игнорируется, используется maxDepth
}

// TestGetFullGraph_Success - получение полного графа
func (s *GraphHandlerIntegrationTestSuite) TestGetFullGraph_Success() {
	// Создаем несколько заметок и связей
	note1 := s.createTestNote("Note 1", "content", "star")
	note2 := s.createTestNote("Note 2", "content", "planet")
	note3 := s.createTestNote("Note 3", "content", "comet")

	s.createTestLink(note1, note2, "reference")
	s.createTestLink(note2, note3, "related")

	// Получаем полный граф
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/graph/all", nil)
	s.router.ServeHTTP(w, req)

	s.Equal(200, w.Code)

	var response GraphData
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)

	// Должны получить все заметки и связи
	s.Len(response.Nodes, 3)
	s.Len(response.Links, 2)

	// Проверяем типы узлов
	for _, n := range response.Nodes {
		s.NotEmpty(n.Type)
		s.Contains([]string{"star", "planet", "comet"}, n.Type)
	}
}

// TestGetFullGraph_Empty - пустой граф
func (s *GraphHandlerIntegrationTestSuite) TestGetFullGraph_Empty() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/graph/all", nil)
	s.router.ServeHTTP(w, req)

	s.Equal(200, w.Code)

	var response GraphData
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)

	s.Len(response.Nodes, 0)
	s.Len(response.Links, 0)
}

// TestGetFullGraph_WithLimit - проверка query-параметра limit
func (s *GraphHandlerIntegrationTestSuite) TestGetFullGraph_WithLimit() {
	// Создаем 5 заметок
	for i := 1; i <= 5; i++ {
		s.createTestNote("Note "+string(rune('0'+i)), "content", "star")
	}

	// Запрос с limit=3
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/graph/all?limit=3", nil)
	s.router.ServeHTTP(w, req)

	s.Equal(200, w.Code)

	var response GraphData
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)

	// Должны получить только 3 заметки
	s.Len(response.Nodes, 3)
}

// TestGetFullGraph_InvalidLimit - невалидный limit
func (s *GraphHandlerIntegrationTestSuite) TestGetFullGraph_InvalidLimit() {
	s.createTestNote("Note 1", "content", "star")
	s.createTestNote("Note 2", "content", "planet")

	// Невалидный limit (строка)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/graph/all?limit=invalid", nil)
	s.router.ServeHTTP(w, req)

	s.Equal(200, w.Code)

	var response GraphData
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)

	// Должны получить все заметки (limit игнорируется)
	s.Len(response.Nodes, 2)
}

// TestGraphBidirectionalLinks - проверка загрузки связей в обоих направлениях
func (s *GraphHandlerIntegrationTestSuite) TestGraphBidirectionalLinks() {
	// Создаем треугольник: A <-> B, B <-> C, C <-> A
	noteA := s.createTestNote("Note A", "content", "star")
	noteB := s.createTestNote("Note B", "content", "planet")
	noteC := s.createTestNote("Note C", "content", "comet")

	// A -> B, B -> C, C -> A
	s.createTestLink(noteA, noteB, "reference")
	s.createTestLink(noteB, noteC, "reference")
	s.createTestLink(noteC, noteA, "reference")

	// Получаем граф от A с depth=2
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notes/"+noteA.ID().String()+"/graph?depth=2", nil)
	s.router.ServeHTTP(w, req)

	s.Equal(200, w.Code)

	var response GraphData
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)

	// Должны получить все 3 узла и 3 связи
	s.Len(response.Nodes, 3)
	s.Len(response.Links, 3)

	// Проверяем что все узлы присутствуют
	ids := make(map[string]bool)
	for _, n := range response.Nodes {
		ids[n.ID] = true
	}
	s.True(ids[noteA.ID().String()])
	s.True(ids[noteB.ID().String()])
	s.True(ids[noteC.ID().String()])
}

// Запускаем тесты
func TestGraphHandlerIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	suite.Run(t, new(GraphHandlerIntegrationTestSuite))
}
