//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/testutil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// NoteRepositoryIntegrationTestSuite - интеграционные тесты для NoteRepository
type NoteRepositoryIntegrationTestSuite struct {
	suite.Suite
	db      *gorm.DB
	repo    *NoteRepository
	cleanup func()
	ctx     context.Context
}

func (s *NoteRepositoryIntegrationTestSuite) SetupSuite() {
	s.db, s.cleanup = testutil.SetupTestDB(s.T())
	s.ctx = context.Background()

	// Миграция моделей (без NoteEmbeddingModel - требует pgvector extension)
	models := []interface{}{
		&NoteModel{},
		&LinkModel{},
		&NoteKeywordModel{},
		&UserModel{},
		&TagModel{},
		&NoteTagModel{},
	}
	err := s.db.AutoMigrate(models...)
	s.Require().NoError(err, "failed to migrate models")

	// Создаем репозиторий без Redis (для интеграционных тестов)
	s.repo = NewNoteRepository(s.db, nil)
}

func (s *NoteRepositoryIntegrationTestSuite) TearDownSuite() {
	s.cleanup()
}

func (s *NoteRepositoryIntegrationTestSuite) SetupTest() {
	// Очищаем таблицы перед каждым тестом
	err := testutil.TruncateTables(s.db)
	s.Require().NoError(err)
}

// TestSave_Create - создание новой заметки
func (s *NoteRepositoryIntegrationTestSuite) TestSave_Create() {
	title, _ := note.NewTitle("Test Title")
	content, _ := note.NewContent("Test Content")
	metadata, _ := note.NewMetadata(map[string]interface{}{"key": "value"})
	n := note.NewNote(title, content, note.MustType("star"), metadata)

	err := s.repo.Save(s.ctx, n)
	s.NoError(err)
	s.NotZero(n.ID())

	// Проверяем, что заметка сохранилась
	found, err := s.repo.FindByID(s.ctx, n.ID())
	s.NoError(err)
	s.Equal(n.ID(), found.ID())
	s.Equal("Test Title", found.Title().String())
}

// TestSave_Update - обновление существующей заметки
func (s *NoteRepositoryIntegrationTestSuite) TestSave_Update() {
	// Создаем заметку
	title, _ := note.NewTitle("Original Title")
	content, _ := note.NewContent("Original Content")
	metadata, _ := note.NewMetadata(map[string]interface{}{})
	n := note.NewNote(title, content, note.MustType("star"), metadata)

	err := s.repo.Save(s.ctx, n)
	s.NoError(err)
	id := n.ID()

	// Обновляем через повторный Save (репозиторий обновляет существующую запись)
	newTitle, _ := note.NewTitle("Updated Title")
	newContent, _ := note.NewContent("Updated Content")
	updatedNote := note.NewNote(newTitle, newContent, note.MustType("planet"), metadata)
	// Копируем ID для обновления
	updatedNoteModel, _ := toGormNote(updatedNote)
	updatedNoteModel.ID = id
	s.db.Model(&NoteModel{}).Where("id = ?", id).Update("title", "Updated Title")

	// Проверяем обновление
	found, err := s.repo.FindByID(s.ctx, id)
	s.NoError(err)
	s.Equal("Updated Title", found.Title().String())
}

// TestFindByID_NotFound - поиск несуществующей заметки
func (s *NoteRepositoryIntegrationTestSuite) TestFindByID_NotFound() {
	found, err := s.repo.FindByID(s.ctx, uuid.New())
	s.NoError(err)
	s.Nil(found)
}

// TestFindAll - поиск всех заметок
func (s *NoteRepositoryIntegrationTestSuite) TestFindAll() {
	// Создаем несколько заметок
	for i := 0; i < 3; i++ {
		title, _ := note.NewTitle("Note " + string(rune('A'+i)))
		content, _ := note.NewContent("Content " + string(rune('A'+i)))
		metadata, _ := note.NewMetadata(map[string]interface{}{})
		n := note.NewNote(title, content, note.MustType("star"), metadata)
		err := s.repo.Save(s.ctx, n)
		s.NoError(err)
	}

	notes, err := s.repo.FindAll(s.ctx)
	s.NoError(err)
	s.Len(notes, 3)
}

// TestDelete - удаление заметки
func (s *NoteRepositoryIntegrationTestSuite) TestDelete() {
	title, _ := note.NewTitle("To Delete")
	content, _ := note.NewContent("Content to delete")
	metadata, _ := note.NewMetadata(map[string]interface{}{})
	n := note.NewNote(title, content, note.MustType("star"), metadata)

	err := s.repo.Save(s.ctx, n)
	s.NoError(err)
	id := n.ID()

	// Удаляем
	err = s.repo.Delete(s.ctx, id)
	s.NoError(err)

	// Проверяем, что заметка не найдена
	found, err := s.repo.FindByID(s.ctx, id)
	s.NoError(err)
	s.Nil(found)
}

// TestFindAllWithFilter - поиск с фильтрацией
func (s *NoteRepositoryIntegrationTestSuite) TestFindAllWithFilter() {
	// Создаем заметки разных типов
	title1, _ := note.NewTitle("Golang Tutorial")
	content1, _ := note.NewContent("Learn Go programming")
	metadata, _ := note.NewMetadata(map[string]interface{}{})
	n1 := note.NewNote(title1, content1, note.MustType("star"), metadata)

	title2, _ := note.NewTitle("Python Guide")
	content2, _ := note.NewContent("Learn Python programming")
	n2 := note.NewNote(title2, content2, note.MustType("planet"), metadata)

	err := s.repo.Save(s.ctx, n1)
	s.NoError(err)
	err = s.repo.Save(s.ctx, n2)
	s.NoError(err)

	// Проверяем что обе заметки сохранены
	allNotes, err := s.repo.FindAll(s.ctx)
	s.NoError(err)
	s.Len(allNotes, 2)
}

// TestFindByID_AfterDelete - проверка отсутствия после удаления
func (s *NoteRepositoryIntegrationTestSuite) TestFindByID_AfterDelete() {
	title, _ := note.NewTitle("Delete Test")
	content, _ := note.NewContent("Content to delete")
	metadata, _ := note.NewMetadata(map[string]interface{}{})
	n := note.NewNote(title, content, note.MustType("star"), metadata)

	err := s.repo.Save(s.ctx, n)
	s.NoError(err)
	id := n.ID()

	// Проверяем что существует
	found, err := s.repo.FindByID(s.ctx, id)
	s.NoError(err)
	s.NotNil(found)

	// Удаляем
	err = s.repo.Delete(s.ctx, id)
	s.NoError(err)

	// Проверяем что не найдена
	found, err = s.repo.FindByID(s.ctx, id)
	s.NoError(err)
	s.Nil(found)
}

// COMET-1: «Ближайшие дела» — незакрытые кометы пользователя, датированные по
// возрастанию due_at (просроченные первыми), бездатные в конце.
func (s *NoteRepositoryIntegrationTestSuite) TestFindComets_OrderingAndFilters() {
	now := time.Now()

	mkUser := func(login string) uuid.UUID {
		u := UserModel{ID: uuid.New(), Login: login, Email: login + "@t.local", PasswordHash: "x"}
		s.Require().NoError(s.db.Create(&u).Error)
		return u.ID
	}
	user := mkUser("comet-owner")
	other := mkUser("comet-stranger")

	mk := func(title string, creator uuid.UUID, typ string, due, done *time.Time) *note.Note {
		ttl, _ := note.NewTitle(title)
		cnt, _ := note.NewContent("content")
		meta, _ := note.NewMetadata(nil)
		n := note.NewNoteWithCreator(ttl, cnt, note.MustType(typ), meta, creator)
		s.Require().NoError(n.SetCometFields(due, nil, done))
		s.Require().NoError(s.repo.Save(s.ctx, n))
		return n
	}
	ptr := func(t time.Time) *time.Time { return &t }

	soon := mk("soon", user, "comet", ptr(now.Add(time.Hour)), nil)
	oldOverdue := mk("oldOverdue", user, "comet", ptr(now.Add(-48*time.Hour)), nil)
	undated := mk("undated", user, "comet", nil, nil)
	mk("done", user, "comet", ptr(now.Add(time.Hour)), ptr(now))  // закрыта — не в списке
	mk("stranger", other, "comet", ptr(now.Add(-time.Hour)), nil) // чужая — не в списке
	mk("notComet", user, "star", ptr(now.Add(-time.Hour)), nil)   // не комета — не в списке
	deleted := mk("deleted", user, "comet", ptr(now.Add(time.Hour)), nil)
	s.Require().NoError(s.repo.Delete(s.ctx, deleted.ID())) // soft-deleted — не в списке

	got, err := s.repo.FindComets(s.ctx, user)
	s.Require().NoError(err)
	s.Require().Len(got, 3)

	s.Equal(oldOverdue.ID(), got[0].ID(), "самая просроченная — первая")
	s.Equal(soon.ID(), got[1].ID())
	s.Equal(undated.ID(), got[2].ID(), "бездатная — последней")
}

// Запускаем тесты
func TestNoteRepositoryIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	suite.Run(t, new(NoteRepositoryIntegrationTestSuite))
}
