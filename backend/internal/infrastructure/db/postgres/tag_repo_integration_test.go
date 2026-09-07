//go:build integration

package postgres

import (
	"context"
	"strings"
	"testing"

	"knowledge-graph/internal/domain/note"
	tagDomain "knowledge-graph/internal/domain/tag"
	"knowledge-graph/internal/testutil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// TagRepositoryIntegrationTestSuite - интеграционные тесты для TagRepository
type TagRepositoryIntegrationTestSuite struct {
	suite.Suite
	db       *gorm.DB
	repo     *TagRepository
	noteRepo *NoteRepository
	cleanup  func()
	ctx      context.Context
}

func (s *TagRepositoryIntegrationTestSuite) SetupSuite() {
	s.db, s.cleanup = testutil.SetupTestDB(s.T())
	s.ctx = context.Background()

	// Миграция всех моделей
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

	// Создаем репозитории
	s.repo = NewTagRepository(s.db)
	s.noteRepo = NewNoteRepository(s.db, nil)
}

func (s *TagRepositoryIntegrationTestSuite) TearDownSuite() {
	s.cleanup()
}

func (s *TagRepositoryIntegrationTestSuite) SetupTest() {
	// Очищаем таблицы перед каждым тестом
	err := testutil.TruncateTables(s.db)
	s.Require().NoError(err, "failed to truncate tables")
	// Очищаем дополнительно tags и note_tags (не в списке TruncateTables)
	s.db.Exec("TRUNCATE TABLE tags, note_tags RESTART IDENTITY CASCADE")
}

// createTestNote создает тестовую заметку
func (s *TagRepositoryIntegrationTestSuite) createTestNote(title string) *note.Note {
	noteTitle, _ := note.NewTitle(title)
	noteContent, _ := note.NewContent("Test content for " + title)
	metadata, _ := note.NewMetadata(map[string]interface{}{})
	n := note.NewNote(noteTitle, noteContent, "star", metadata)
	err := s.noteRepo.Save(s.ctx, n)
	s.Require().NoError(err, "failed to create test note")
	return n
}

// createTestTag создает тестовый тег
func (s *TagRepositoryIntegrationTestSuite) createTestTag(name string) *tagDomain.Tag {
	t, err := tagDomain.New(name)
	s.Require().NoError(err, "failed to create test tag")
	err = s.repo.Create(s.ctx, t)
	s.Require().NoError(err, "failed to save test tag")
	return t
}

// TestCreate - создание тега
func (s *TagRepositoryIntegrationTestSuite) TestCreate() {
	t := s.createTestTag("golang")

	// Проверяем что тег создан
	found, err := s.repo.FindByID(s.ctx, t.ID())
	s.NoError(err)
	s.NotNil(found)
	s.Equal("golang", found.Name())
}

// TestFindByID - поиск по ID
func (s *TagRepositoryIntegrationTestSuite) TestFindByID() {
	t := s.createTestTag("find_me")

	found, err := s.repo.FindByID(s.ctx, t.ID())
	s.NoError(err)
	s.NotNil(found)
	s.Equal(t.ID(), found.ID())
	s.Equal("find_me", found.Name())
}

// TestFindByID_NotFound - поиск несуществующего
func (s *TagRepositoryIntegrationTestSuite) TestFindByID_NotFound() {
	found, err := s.repo.FindByID(s.ctx, uuid.New())
	s.NoError(err)
	s.Nil(found)
}

// TestFindByName - поиск по имени
func (s *TagRepositoryIntegrationTestSuite) TestFindByName() {
	t := s.createTestTag("specific_tag")

	found, err := s.repo.FindByName(s.ctx, "specific_tag")
	s.NoError(err)
	s.NotNil(found)
	s.Equal(t.ID(), found.ID())
}

// TestFindByName_NotFound - поиск несуществующего имени
func (s *TagRepositoryIntegrationTestSuite) TestFindByName_NotFound() {
	found, err := s.repo.FindByName(s.ctx, "nonexistent")
	s.NoError(err)
	s.Nil(found)
}

// TestDuplicateName - проверка уникальности имени
func (s *TagRepositoryIntegrationTestSuite) TestDuplicateName() {
	s.createTestTag("unique_tag")

	tag2, err := tagDomain.New("unique_tag")
	s.Require().NoError(err)

	err = s.repo.Create(s.ctx, tag2)
	s.Error(err)
	s.Contains(strings.ToLower(err.Error()), "duplicate")
}

// TestUpdate - обновление тега
func (s *TagRepositoryIntegrationTestSuite) TestUpdate() {
	t := s.createTestTag("old_name")

	// Обновляем через прямое изменение в БД
	err := s.db.Model(&TagModel{}).Where("id = ?", t.ID()).Update("name", "new_name").Error
	s.NoError(err)

	// Проверяем обновление
	found, err := s.repo.FindByID(s.ctx, t.ID())
	s.NoError(err)
	s.Equal("new_name", found.Name())
}

// TestDelete - удаление тега
func (s *TagRepositoryIntegrationTestSuite) TestDelete() {
	t := s.createTestTag("delete_me")

	// Проверяем что существует
	exists, err := s.repo.Exists(s.ctx, t.ID())
	s.NoError(err)
	s.True(exists)

	// Удаляем
	err = s.repo.Delete(s.ctx, t.ID())
	s.NoError(err)

	// Проверяем что не существует
	exists, err = s.repo.Exists(s.ctx, t.ID())
	s.NoError(err)
	s.False(exists)
}

// TestAddTagToNote - добавление тега к заметке
func (s *TagRepositoryIntegrationTestSuite) TestAddTagToNote() {
	// Создаем тег и заметку
	t := s.createTestTag("important")
	n := s.createTestNote("Tagged Note")

	// Добавляем тег к заметке
	err := s.repo.AddTagToNote(s.ctx, n.ID(), t.ID())
	s.NoError(err)

	// Проверяем что тег добавлен
	tags, err := s.repo.GetTagsForNote(s.ctx, n.ID())
	s.NoError(err)
	s.Len(tags, 1)
	s.Equal(t.ID(), tags[0].ID)
}

// TestRemoveTagFromNote - удаление тега от заметки
func (s *TagRepositoryIntegrationTestSuite) TestRemoveTagFromNote() {
	// Создаем тег и заметку
	t := s.createTestTag("removable")
	n := s.createTestNote("Untag Me")

	// Добавляем и удаляем тег
	err := s.repo.AddTagToNote(s.ctx, n.ID(), t.ID())
	s.NoError(err)

	err = s.repo.RemoveTagFromNote(s.ctx, n.ID(), t.ID())
	s.NoError(err)

	// Проверяем что тег удален
	tags, err := s.repo.GetTagsForNote(s.ctx, n.ID())
	s.NoError(err)
	s.Len(tags, 0)
}

// TestGetTagsForNote - получение всех тегов заметки
func (s *TagRepositoryIntegrationTestSuite) TestGetTagsForNote() {
	n := s.createTestNote("Multi Tagged")

	// Создаем несколько тегов
	for _, name := range []string{"tag1", "tag2", "tag3"} {
		t := s.createTestTag(name)
		err := s.repo.AddTagToNote(s.ctx, n.ID(), t.ID())
		s.NoError(err)
	}

	// Получаем теги
	tags, err := s.repo.GetTagsForNote(s.ctx, n.ID())
	s.NoError(err)
	s.Len(tags, 3)
}

// TestGetNotesForTag - получение всех заметок с тегом
func (s *TagRepositoryIntegrationTestSuite) TestGetNotesForTag() {
	// Создаем тег
	t := s.createTestTag("shared_tag")

	// Создаем несколько заметок с этим тегом
	for i := 0; i < 3; i++ {
		n := s.createTestNote("Note " + string(rune('A'+i)))
		err := s.repo.AddTagToNote(s.ctx, n.ID(), t.ID())
		s.NoError(err)
		_ = i // используем i чтобы избежать warning
	}

	// Получаем заметки
	notes, err := s.repo.GetNotesForTag(s.ctx, t.ID())
	s.NoError(err)
	s.Len(notes, 3)
}

// TestConstraintProtection - проверка FK constraint защищает от удаления тега с связями
func (s *TagRepositoryIntegrationTestSuite) TestConstraintProtection() {
	// Создаем тег и заметку
	t := s.createTestTag("protected_tag")
	n := s.createTestNote("Protected Note")

	// Добавляем тег
	err := s.repo.AddTagToNote(s.ctx, n.ID(), t.ID())
	s.NoError(err)

	// Пытаемся удалить тег - должно упасть из-за FK constraint
	err = s.repo.Delete(s.ctx, t.ID())
	s.Error(err)
	s.Contains(err.Error(), "violates foreign key constraint")

	// Корректный порядок: сначала удаляем связи
	err = s.repo.RemoveTagFromNote(s.ctx, n.ID(), t.ID())
	s.NoError(err)

	// Теперь можно удалить тег
	err = s.repo.Delete(s.ctx, t.ID())
	s.NoError(err)

	// Проверяем что связей больше нет
	var count int64
	s.db.Model(&NoteTagModel{}).Where("tag_id = ?", t.ID()).Count(&count)
	s.Equal(int64(0), count)
}

// TestFindAll - получение всех тегов
func (s *TagRepositoryIntegrationTestSuite) TestFindAll() {
	// Создаем несколько тегов
	for _, name := range []string{"go", "python", "java"} {
		s.createTestTag(name)
	}

	tags, err := s.repo.FindAll(s.ctx)
	s.NoError(err)
	s.Len(tags, 3)
}

// Запускаем тесты
func TestTagRepositoryIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	suite.Run(t, new(TagRepositoryIntegrationTestSuite))
}
