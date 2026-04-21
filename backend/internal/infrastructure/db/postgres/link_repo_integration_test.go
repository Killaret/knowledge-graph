//go:build integration

package postgres

import (
	"context"
	"testing"

	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/testutil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// LinkRepositoryIntegrationTestSuite - интеграционные тесты для LinkRepository
type LinkRepositoryIntegrationTestSuite struct {
	suite.Suite
	db       *gorm.DB
	repo     *LinkRepository
	noteRepo *NoteRepository
	cleanup  func()
	ctx      context.Context

	// Тестовые заметки
	sourceNote *note.Note
	targetNote *note.Note
}

func (s *LinkRepositoryIntegrationTestSuite) SetupSuite() {
	s.db, s.cleanup = testutil.SetupTestDB(s.T())
	s.ctx = context.Background()

	// Миграция моделей
	models := []interface{}{
		&NoteModel{},
		&LinkModel{},
		&NoteKeywordModel{},
	}
	err := s.db.AutoMigrate(models...)
	s.Require().NoError(err, "failed to migrate models")

	// Создаем репозитории
	s.repo = NewLinkRepository(s.db)
	s.noteRepo = NewNoteRepository(s.db, nil)
}

func (s *LinkRepositoryIntegrationTestSuite) TearDownSuite() {
	s.cleanup()
}

func (s *LinkRepositoryIntegrationTestSuite) SetupTest() {
	// Очищаем таблицы перед каждым тестом
	err := testutil.TruncateTables(s.db)
	s.Require().NoError(err, "failed to truncate tables")

	// Создаем тестовые заметки
	sourceTitle, _ := note.NewTitle("Source Note")
	sourceContent, _ := note.NewContent("Source content for testing")
	metadata, _ := note.NewMetadata(map[string]interface{}{})
	s.sourceNote = note.NewNote(sourceTitle, sourceContent, "star", metadata)

	targetTitle, _ := note.NewTitle("Target Note")
	targetContent, _ := note.NewContent("Target content for testing")
	s.targetNote = note.NewNote(targetTitle, targetContent, "star", metadata)

	// Сохраняем заметки
	err = s.noteRepo.Save(s.ctx, s.sourceNote)
	s.Require().NoError(err, "failed to save source note")
	err = s.noteRepo.Save(s.ctx, s.targetNote)
	s.Require().NoError(err, "failed to save target note")
}

// TestSave_Create - создание новой связи
func (s *LinkRepositoryIntegrationTestSuite) TestSave_Create() {
	linkType, _ := link.NewLinkType("reference")
	weight, _ := link.NewWeight(0.8)
	metadata, _ := link.NewMetadata(map[string]interface{}{"reason": "test"})

	l := link.NewLink(s.sourceNote.ID(), s.targetNote.ID(), linkType, weight, metadata)

	err := s.repo.Save(s.ctx, l)
	s.NoError(err)

	// Проверяем что связь сохранена
	found, err := s.repo.FindByID(s.ctx, l.ID())
	s.NoError(err)
	s.NotNil(found)
	s.Equal(s.sourceNote.ID(), found.SourceNoteID())
	s.Equal(s.targetNote.ID(), found.TargetNoteID())
	s.Equal("reference", found.LinkType().String())
	s.Equal(0.8, found.Weight().Value())
}

// TestSave_Update - обновление существующей связи
func (s *LinkRepositoryIntegrationTestSuite) TestSave_Update() {
	// Создаем связь
	linkType, _ := link.NewLinkType("reference")
	weight, _ := link.NewWeight(0.5)
	metadata, _ := link.NewMetadata(map[string]interface{}{})

	l := link.NewLink(s.sourceNote.ID(), s.targetNote.ID(), linkType, weight, metadata)
	err := s.repo.Save(s.ctx, l)
	s.NoError(err)

	id := l.ID()

	// Обновляем через прямое изменение в БД (имитация обновления)
	err = s.db.Model(&LinkModel{}).Where("id = ?", id).Update("weight", 0.9).Error
	s.NoError(err)

	// Проверяем обновление
	found, err := s.repo.FindByID(s.ctx, id)
	s.NoError(err)
	s.Equal(0.9, found.Weight().Value())
}

// TestFindByID_NotFound - поиск несуществующей связи
func (s *LinkRepositoryIntegrationTestSuite) TestFindByID_NotFound() {
	found, err := s.repo.FindByID(s.ctx, uuid.New())
	s.NoError(err)
	s.Nil(found)
}

// TestFindBySource - поиск по source_id
func (s *LinkRepositoryIntegrationTestSuite) TestFindBySource() {
	// Создаем несколько связей от одного source
	linkType, _ := link.NewLinkType("reference")
	weight, _ := link.NewWeight(0.5)
	metadata, _ := link.NewMetadata(map[string]interface{}{})

	// Создаем вторую target заметку
	targetTitle2, _ := note.NewTitle("Target Note 2")
	targetContent2, _ := note.NewContent("Content 2")
	noteMetadata, _ := note.NewMetadata(map[string]interface{}{})
	targetNote2 := note.NewNote(targetTitle2, targetContent2, "star", noteMetadata)
	err := s.noteRepo.Save(s.ctx, targetNote2)
	s.NoError(err)

	// Создаем две связи от sourceNote
	l1 := link.NewLink(s.sourceNote.ID(), s.targetNote.ID(), linkType, weight, metadata)
	l2 := link.NewLink(s.sourceNote.ID(), targetNote2.ID(), linkType, weight, metadata)

	err = s.repo.Save(s.ctx, l1)
	s.NoError(err)
	err = s.repo.Save(s.ctx, l2)
	s.NoError(err)

	// Ищем по source
	links, err := s.repo.FindBySource(s.ctx, s.sourceNote.ID())
	s.NoError(err)
	s.Len(links, 2)
}

// TestFindByTarget - поиск по target_id
func (s *LinkRepositoryIntegrationTestSuite) TestFindByTarget() {
	// Создаем связь
	linkType, _ := link.NewLinkType("reference")
	weight, _ := link.NewWeight(0.5)
	metadata, _ := link.NewMetadata(map[string]interface{}{})

	l := link.NewLink(s.sourceNote.ID(), s.targetNote.ID(), linkType, weight, metadata)
	err := s.repo.Save(s.ctx, l)
	s.NoError(err)

	// Ищем по target
	links, err := s.repo.FindByTarget(s.ctx, s.targetNote.ID())
	s.NoError(err)
	s.Len(links, 1)
	s.Equal(s.sourceNote.ID(), links[0].SourceNoteID())
}

// TestFindBySourceIDs - batch поиск по нескольким source_id
func (s *LinkRepositoryIntegrationTestSuite) TestFindBySourceIDs() {
	// Создаем вторую source заметку
	sourceTitle2, _ := note.NewTitle("Source Note 2")
	sourceContent2, _ := note.NewContent("Content 2")
	noteMetadata, _ := note.NewMetadata(map[string]interface{}{})
	sourceNote2 := note.NewNote(sourceTitle2, sourceContent2, "star", noteMetadata)
	err := s.noteRepo.Save(s.ctx, sourceNote2)
	s.NoError(err)

	// Создаем связи
	linkType, _ := link.NewLinkType("reference")
	weight, _ := link.NewWeight(0.5)
	metadata, _ := link.NewMetadata(map[string]interface{}{})

	l1 := link.NewLink(s.sourceNote.ID(), s.targetNote.ID(), linkType, weight, metadata)
	l2 := link.NewLink(sourceNote2.ID(), s.targetNote.ID(), linkType, weight, metadata)

	err = s.repo.Save(s.ctx, l1)
	s.NoError(err)
	err = s.repo.Save(s.ctx, l2)
	s.NoError(err)

	// Batch поиск
	result, err := s.repo.FindBySourceIDs(s.ctx, []uuid.UUID{s.sourceNote.ID(), sourceNote2.ID()})
	s.NoError(err)
	s.Len(result, 2)
	s.Len(result[s.sourceNote.ID()], 1)
	s.Len(result[sourceNote2.ID()], 1)
}

// TestFindByTargetIDs - batch поиск по нескольким target_id
func (s *LinkRepositoryIntegrationTestSuite) TestFindByTargetIDs() {
	// Создаем вторую target заметку
	targetTitle2, _ := note.NewTitle("Target Note 2")
	targetContent2, _ := note.NewContent("Content 2")
	noteMetadata, _ := note.NewMetadata(map[string]interface{}{})
	targetNote2 := note.NewNote(targetTitle2, targetContent2, "star", noteMetadata)
	err := s.noteRepo.Save(s.ctx, targetNote2)
	s.NoError(err)

	// Создаем связи к разным target
	linkType, _ := link.NewLinkType("reference")
	weight, _ := link.NewWeight(0.5)
	metadata, _ := link.NewMetadata(map[string]interface{}{})

	l1 := link.NewLink(s.sourceNote.ID(), s.targetNote.ID(), linkType, weight, metadata)
	l2 := link.NewLink(s.sourceNote.ID(), targetNote2.ID(), linkType, weight, metadata)

	err = s.repo.Save(s.ctx, l1)
	s.NoError(err)
	err = s.repo.Save(s.ctx, l2)
	s.NoError(err)

	// Batch поиск по target
	result, err := s.repo.FindByTargetIDs(s.ctx, []uuid.UUID{s.targetNote.ID(), targetNote2.ID()})
	s.NoError(err)
	s.Len(result, 2)
}

// TestFindAll - поиск всех связей
func (s *LinkRepositoryIntegrationTestSuite) TestFindAll() {
	// Создаем несколько связей
	linkType, _ := link.NewLinkType("reference")
	weight, _ := link.NewWeight(0.5)
	metadata, _ := link.NewMetadata(map[string]interface{}{})

	l1 := link.NewLink(s.sourceNote.ID(), s.targetNote.ID(), linkType, weight, metadata)

	// Создаем вторую target заметку и связь
	targetTitle2, _ := note.NewTitle("Target Note 2")
	targetContent2, _ := note.NewContent("Content 2")
	noteMetadata, _ := note.NewMetadata(map[string]interface{}{})
	targetNote2 := note.NewNote(targetTitle2, targetContent2, "star", noteMetadata)
	err := s.noteRepo.Save(s.ctx, targetNote2)
	s.NoError(err)

	l2 := link.NewLink(s.sourceNote.ID(), targetNote2.ID(), linkType, weight, metadata)

	err = s.repo.Save(s.ctx, l1)
	s.NoError(err)
	err = s.repo.Save(s.ctx, l2)
	s.NoError(err)

	// Проверяем что обе связи сохранены
	allLinks, err := s.repo.FindAll(s.ctx)
	s.NoError(err)
	s.Len(allLinks, 2)
}

// TestDelete - удаление связи
func (s *LinkRepositoryIntegrationTestSuite) TestDelete() {
	// Создаем связь
	linkType, _ := link.NewLinkType("reference")
	weight, _ := link.NewWeight(0.5)
	metadata, _ := link.NewMetadata(map[string]interface{}{})

	l := link.NewLink(s.sourceNote.ID(), s.targetNote.ID(), linkType, weight, metadata)
	err := s.repo.Save(s.ctx, l)
	s.NoError(err)
	id := l.ID()

	// Удаляем
	err = s.repo.Delete(s.ctx, id)
	s.NoError(err)

	// Проверяем что не найдена
	found, err := s.repo.FindByID(s.ctx, id)
	s.NoError(err)
	s.Nil(found)
}

// TestDeleteBySource - удаление всех связей от source
func (s *LinkRepositoryIntegrationTestSuite) TestDeleteBySource() {
	// Создаем вторую target заметку
	targetTitle2, _ := note.NewTitle("Target Note 2")
	targetContent2, _ := note.NewContent("Content 2")
	noteMetadata, _ := note.NewMetadata(map[string]interface{}{})
	targetNote2 := note.NewNote(targetTitle2, targetContent2, "star", noteMetadata)
	err := s.noteRepo.Save(s.ctx, targetNote2)
	s.NoError(err)

	// Создаем две связи от одного source
	linkType, _ := link.NewLinkType("reference")
	weight, _ := link.NewWeight(0.5)
	metadata, _ := link.NewMetadata(map[string]interface{}{})

	l1 := link.NewLink(s.sourceNote.ID(), s.targetNote.ID(), linkType, weight, metadata)
	l2 := link.NewLink(s.sourceNote.ID(), targetNote2.ID(), linkType, weight, metadata)

	err = s.repo.Save(s.ctx, l1)
	s.NoError(err)
	err = s.repo.Save(s.ctx, l2)
	s.NoError(err)

	// Удаляем все связи от source
	err = s.repo.DeleteBySource(s.ctx, s.sourceNote.ID())
	s.NoError(err)

	// Проверяем что связей больше нет
	links, err := s.repo.FindBySource(s.ctx, s.sourceNote.ID())
	s.NoError(err)
	s.Len(links, 0)
}

// TestConstraintProtection - проверка что FK constraint защищает от удаления
func (s *LinkRepositoryIntegrationTestSuite) TestConstraintProtection() {
	// Создаем связь
	linkType, _ := link.NewLinkType("reference")
	weight, _ := link.NewWeight(0.5)
	metadata, _ := link.NewMetadata(map[string]interface{}{})

	l := link.NewLink(s.sourceNote.ID(), s.targetNote.ID(), linkType, weight, metadata)
	err := s.repo.Save(s.ctx, l)
	s.NoError(err)

	// Пытаемся удалить source заметку - должно упасть из-за FK constraint
	err = s.noteRepo.Delete(s.ctx, s.sourceNote.ID())
	s.Error(err, "should fail due to foreign key constraint")
	s.Contains(err.Error(), "violates foreign key constraint")

	// Корректный порядок: сначала удаляем связи
	err = s.repo.DeleteBySource(s.ctx, s.sourceNote.ID())
	s.NoError(err)

	// Теперь можно удалить заметку
	err = s.noteRepo.Delete(s.ctx, s.sourceNote.ID())
	s.NoError(err)
}

// TestDifferentLinkTypes - разные типы связей
func (s *LinkRepositoryIntegrationTestSuite) TestDifferentLinkTypes() {
	metadata, _ := link.NewMetadata(map[string]interface{}{})

	// Создаем связи разных типов
	referenceType, _ := link.NewLinkType("reference")
	dependencyType, _ := link.NewLinkType("dependency")
	weight, _ := link.NewWeight(0.5)

	l1 := link.NewLink(s.sourceNote.ID(), s.targetNote.ID(), referenceType, weight, metadata)
	l2 := link.NewLink(s.targetNote.ID(), s.sourceNote.ID(), dependencyType, weight, metadata)

	err := s.repo.Save(s.ctx, l1)
	s.NoError(err)
	err = s.repo.Save(s.ctx, l2)
	s.NoError(err)

	// Проверяем типы
	allLinks, err := s.repo.FindAll(s.ctx)
	s.NoError(err)
	s.Len(allLinks, 2)

	types := make(map[string]bool)
	for _, l := range allLinks {
		types[l.LinkType().String()] = true
	}
	s.True(types["reference"])
	s.True(types["dependency"])
}

// Запускаем тесты
func TestLinkRepositoryIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	suite.Run(t, new(LinkRepositoryIntegrationTestSuite))
}
