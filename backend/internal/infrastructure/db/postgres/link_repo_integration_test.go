//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

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
		&LinkSuppressionModel{},
		&NoteKeywordModel{},
		&UserModel{},
		&TagModel{},
		&NoteTagModel{},
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
	s.sourceNote = note.NewNote(sourceTitle, sourceContent, note.MustType("star"), metadata)

	targetTitle, _ := note.NewTitle("Target Note")
	targetContent, _ := note.NewContent("Target content for testing")
	s.targetNote = note.NewNote(targetTitle, targetContent, note.MustType("star"), metadata)

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
	targetNote2 := note.NewNote(targetTitle2, targetContent2, note.MustType("star"), noteMetadata)
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
	sourceNote2 := note.NewNote(sourceTitle2, sourceContent2, note.MustType("star"), noteMetadata)
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
	targetNote2 := note.NewNote(targetTitle2, targetContent2, note.MustType("star"), noteMetadata)
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
	targetNote2 := note.NewNote(targetTitle2, targetContent2, note.MustType("star"), noteMetadata)
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
	targetNote2 := note.NewNote(targetTitle2, targetContent2, note.MustType("star"), noteMetadata)
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

// TestConstraintProtection - удаление заметки каскадно удаляет связи (ON DELETE CASCADE, миграция 002)
func (s *LinkRepositoryIntegrationTestSuite) TestConstraintProtection() {
	// Создаем связь
	linkType, _ := link.NewLinkType("reference")
	weight, _ := link.NewWeight(0.5)
	metadata, _ := link.NewMetadata(map[string]interface{}{})

	l := link.NewLink(s.sourceNote.ID(), s.targetNote.ID(), linkType, weight, metadata)
	err := s.repo.Save(s.ctx, l)
	s.NoError(err)

	// Удаляем source заметку - каскад сносит связь, ошибки нет
	err = s.noteRepo.Delete(s.ctx, s.sourceNote.ID())
	s.NoError(err, "delete should cascade through ON DELETE CASCADE")

	links, err := s.repo.FindBySource(s.ctx, s.sourceNote.ID())
	s.NoError(err)
	s.Len(links, 0, "cascade must remove outgoing links")

	incoming, err := s.repo.FindByTarget(s.ctx, s.sourceNote.ID())
	s.NoError(err)
	s.Len(incoming, 0, "cascade must remove incoming links")
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

// TestDeleteBySourceType — LINKS-1: регенерация удаляет только gamma-связи,
// ручные (source_type='user') сохраняются.
func (s *LinkRepositoryIntegrationTestSuite) TestDeleteBySourceType() {
	linkType, _ := link.NewLinkType("related")
	weight, _ := link.NewWeight(0.9)
	metadata, _ := link.NewMetadata(map[string]interface{}{"source": "gamma"})

	gamma := link.NewGammaLink(s.sourceNote.ID(), s.targetNote.ID(), linkType, weight, metadata)
	s.Require().NoError(s.repo.Save(s.ctx, gamma))

	// Ручная связь той же пары — другой link_type, чтобы не упираться в
	// uniqueIndex (source, target, type).
	depType, _ := link.NewLinkType("dependency")
	manual := link.NewLink(s.sourceNote.ID(), s.targetNote.ID(), depType, weight, metadata)
	s.Require().NoError(s.repo.Save(s.ctx, manual))

	deleted, err := s.repo.DeleteBySourceType(s.ctx, "gamma")
	s.Require().NoError(err)
	s.Equal(int64(1), deleted)

	remaining, err := s.repo.FindBySource(s.ctx, s.sourceNote.ID())
	s.Require().NoError(err)
	s.Require().Len(remaining, 1)
	s.Equal("user", remaining[0].SourceType().String(), "manual link must survive gamma regeneration")
}

// --- LINKS-2: promotion and rejections on a real database ---

// TestSaveUserLink_PromotesGamma — POST семантика повышения: одна строка,
// source_type='user', metadata.gamma с происхождением, created_at сохранён.
func (s *LinkRepositoryIntegrationTestSuite) TestSaveUserLink_PromotesGamma() {
	linkType, _ := link.NewLinkType("related")
	gammaWeight, _ := link.NewWeight(0.6)
	md, _ := link.NewMetadata(map[string]interface{}{"source": "gamma"})
	gamma := link.NewGammaLink(s.sourceNote.ID(), s.targetNote.ID(), linkType, gammaWeight, md)
	s.Require().NoError(s.repo.Save(s.ctx, gamma))

	userWeight, _ := link.NewWeight(0.9)
	reqMD, _ := link.NewMetadata(map[string]interface{}{"note": "confirmed"})
	creator := uuid.New()
	s.Require().NoError(s.db.Create(&UserModel{ID: creator, Login: "promoter", PasswordHash: "x"}).Error)
	manual := link.NewLinkWithCreator(s.sourceNote.ID(), s.targetNote.ID(), creator, linkType, userWeight, reqMD)

	saved, created, err := s.repo.SaveUserLink(s.ctx, manual)
	s.Require().NoError(err)
	s.False(created, "promotion must not insert a new row")
	s.Equal(gamma.ID(), saved.ID())
	s.Equal("user", saved.SourceType().String())
	s.Equal(0.9, saved.Weight().Value())
	s.Equal(creator, *saved.CreatorID())
	s.WithinDuration(gamma.CreatedAt(), saved.CreatedAt(), time.Microsecond, "created_at must be preserved (postgres stores µs precision)")

	prov, ok := saved.Metadata().Value()["gamma"].(map[string]interface{})
	s.Require().True(ok, "metadata.gamma must record the origin")
	s.InDelta(0.6, prov["score"], 0.0001)
	s.NotEmpty(prov["generated_at"])

	// Проверяем, что в базе ровно одна строка на паре.
	pair, err := s.repo.FindByPair(s.ctx, s.sourceNote.ID(), s.targetNote.ID())
	s.Require().NoError(err)
	s.Require().Len(pair, 1)
	s.Equal("user", pair[0].SourceType().String())
}

// TestSaveUserLink_PromotesGammaReverseDirection — ручная связь Y→X поверх
// gamma X→Y повышает существующую строку, принимая направление запроса:
// одно ребро на пару (решение 53), id и created_at сохранены.
func (s *LinkRepositoryIntegrationTestSuite) TestSaveUserLink_PromotesGammaReverseDirection() {
	linkType, _ := link.NewLinkType("related")
	gammaWeight, _ := link.NewWeight(0.6)
	md, _ := link.NewMetadata(map[string]interface{}{"source": "gamma"})
	gamma := link.NewGammaLink(s.sourceNote.ID(), s.targetNote.ID(), linkType, gammaWeight, md)
	s.Require().NoError(s.repo.Save(s.ctx, gamma))

	userWeight, _ := link.NewWeight(0.9)
	reqMD, _ := link.NewMetadata(nil)
	// Ручная связь в обратную сторону.
	manual := link.NewLink(s.targetNote.ID(), s.sourceNote.ID(), linkType, userWeight, reqMD)

	saved, created, err := s.repo.SaveUserLink(s.ctx, manual)
	s.Require().NoError(err)
	s.False(created, "reverse-direction manual link must promote the gamma row")
	s.Equal(gamma.ID(), saved.ID())
	s.Equal("user", saved.SourceType().String())
	s.Equal(s.targetNote.ID(), saved.SourceNoteID(), "promoted row must adopt the requested direction")
	s.Equal(s.sourceNote.ID(), saved.TargetNoteID())
	s.True(saved.HasGammaProvenance(), "promotion keeps gamma origin")

	pairForward, err := s.repo.FindByPair(s.ctx, s.sourceNote.ID(), s.targetNote.ID())
	s.Require().NoError(err)
	s.Empty(pairForward, "no X→Y row must remain")
	pairBack, err := s.repo.FindByPair(s.ctx, s.targetNote.ID(), s.sourceNote.ID())
	s.Require().NoError(err)
	s.Require().Len(pairBack, 1)
}

// TestSaveUserLink_ManualConflictReverse — ручная поверх ручной во встречном
// направлении тоже конфликт: одно ребро на пару.
func (s *LinkRepositoryIntegrationTestSuite) TestSaveUserLink_ManualConflictReverse() {
	linkType, _ := link.NewLinkType("related")
	weight, _ := link.NewWeight(0.8)
	md, _ := link.NewMetadata(nil)
	s.Require().NoError(s.repo.Save(s.ctx, link.NewLink(s.sourceNote.ID(), s.targetNote.ID(), linkType, weight, md)))

	back := link.NewLink(s.targetNote.ID(), s.sourceNote.ID(), linkType, weight, md)
	_, _, err := s.repo.SaveUserLink(s.ctx, back)
	s.ErrorIs(err, link.ErrDuplicateLink)
}

// TestSaveUserLink_ManualConflict — ручная поверх ручной остаётся конфликтом.
func (s *LinkRepositoryIntegrationTestSuite) TestSaveUserLink_ManualConflict() {
	linkType, _ := link.NewLinkType("related")
	weight, _ := link.NewWeight(0.8)
	md, _ := link.NewMetadata(nil)
	s.Require().NoError(s.repo.Save(s.ctx, link.NewLink(s.sourceNote.ID(), s.targetNote.ID(), linkType, weight, md)))

	again := link.NewLink(s.sourceNote.ID(), s.targetNote.ID(), linkType, weight, md)
	_, _, err := s.repo.SaveUserLink(s.ctx, again)
	s.ErrorIs(err, link.ErrDuplicateLink)
}

// TestSaveUserLink_DifferentTypeGamma — gamma другого типа удаляется,
// происхождение переезжает в новую строку.
func (s *LinkRepositoryIntegrationTestSuite) TestSaveUserLink_DifferentTypeGamma() {
	relType, _ := link.NewLinkType("related")
	gammaWeight, _ := link.NewWeight(0.55)
	md, _ := link.NewMetadata(map[string]interface{}{"source": "gamma"})
	s.Require().NoError(s.repo.Save(s.ctx,
		link.NewGammaLink(s.sourceNote.ID(), s.targetNote.ID(), relType, gammaWeight, md)))

	depType, _ := link.NewLinkType("dependency")
	userWeight, _ := link.NewWeight(0.8)
	reqMD, _ := link.NewMetadata(nil)
	manual := link.NewLink(s.sourceNote.ID(), s.targetNote.ID(), depType, userWeight, reqMD)

	saved, created, err := s.repo.SaveUserLink(s.ctx, manual)
	s.Require().NoError(err)
	s.True(created, "a different-type manual link is a new row")
	s.Equal("dependency", saved.LinkType().String())

	pair, err := s.repo.FindByPair(s.ctx, s.sourceNote.ID(), s.targetNote.ID())
	s.Require().NoError(err)
	s.Require().Len(pair, 1, "the gamma row must not remain next to the manual one")
	prov, ok := pair[0].Metadata().Value()["gamma"].(map[string]interface{})
	s.Require().True(ok)
	s.InDelta(0.55, prov["score"], 0.0001)
}

// TestDeleteAndSuppress_Persists — удаление gamma-связи пишет отказ,
// который виден через FindSuppressionsForNotes в обе стороны.
func (s *LinkRepositoryIntegrationTestSuite) TestDeleteAndSuppress_Persists() {
	linkType, _ := link.NewLinkType("related")
	weight, _ := link.NewWeight(0.6)
	md, _ := link.NewMetadata(map[string]interface{}{"source": "gamma"})
	gamma := link.NewGammaLink(s.sourceNote.ID(), s.targetNote.ID(), linkType, weight, md)
	s.Require().NoError(s.repo.Save(s.ctx, gamma))

	supType := "related"
	sup := link.NewSuppression(s.sourceNote.ID(), s.targetNote.ID(), &supType, nil)
	s.Require().NoError(s.repo.DeleteAndSuppress(s.ctx, gamma, sup))

	gone, err := s.repo.FindByID(s.ctx, gamma.ID())
	s.Require().NoError(err)
	s.Nil(gone)

	sups, err := s.repo.FindSuppressionsForNotes(s.ctx, []uuid.UUID{s.targetNote.ID()})
	s.Require().NoError(err)
	s.Require().Len(sups, 1, "the rejection must be visible from the other note too (symmetric pair)")
	s.True(sups[0].Blocks(s.targetNote.ID(), s.sourceNote.ID(), "related"),
		"rejection A→B must block the B→A proposal")
	s.False(sups[0].Blocks(s.targetNote.ID(), s.sourceNote.ID(), "dependency"),
		"a typed rejection must not block other types")
}

// TestSaveUserLink_LiftsSuppression — ручное создание на отвергнутой паре
// снимает отказ в той же транзакции.
func (s *LinkRepositoryIntegrationTestSuite) TestSaveUserLink_LiftsSuppression() {
	supType := "related"
	s.Require().NoError(s.repo.SaveSuppression(s.ctx,
		link.NewSuppression(s.sourceNote.ID(), s.targetNote.ID(), &supType, nil)))

	linkType, _ := link.NewLinkType("related")
	weight, _ := link.NewWeight(0.8)
	md, _ := link.NewMetadata(nil)
	manual := link.NewLink(s.sourceNote.ID(), s.targetNote.ID(), linkType, weight, md)

	_, created, err := s.repo.SaveUserLink(s.ctx, manual)
	s.Require().NoError(err)
	s.True(created)

	sups, err := s.repo.FindSuppressionsForNotes(s.ctx, []uuid.UUID{s.sourceNote.ID()})
	s.Require().NoError(err)
	s.Empty(sups, "the rejection must be lifted by manual creation")
}

// TestSuppressionCascadeOnNoteDelete — удаление заметки уносит её отказы.
func (s *LinkRepositoryIntegrationTestSuite) TestSuppressionCascadeOnNoteDelete() {
	supType := "related"
	s.Require().NoError(s.repo.SaveSuppression(s.ctx,
		link.NewSuppression(s.sourceNote.ID(), s.targetNote.ID(), &supType, nil)))

	// Каскад живёт в FK notes→link_suppressions, который создаёт миграция 034.
	// В тесте AutoMigrate создаёт тот же FK через теги модели.
	s.Require().NoError(s.db.Exec("DELETE FROM notes WHERE id = ?", s.sourceNote.ID()).Error)

	sups, err := s.repo.FindSuppressionsForNotes(s.ctx, []uuid.UUID{s.targetNote.ID()})
	s.Require().NoError(err)
	s.Empty(sups, "note deletion must cascade to its rejections")
}

// TestDeleteAndSuppress_AlreadyDeleted — повторный delete/suppress той же
// строки идемпотентен: отказ записывается, ошибки нет (гонка DELETE-запросов).
func (s *LinkRepositoryIntegrationTestSuite) TestDeleteAndSuppress_AlreadyDeleted() {
	linkType, _ := link.NewLinkType("related")
	weight, _ := link.NewWeight(0.6)
	md, _ := link.NewMetadata(nil)
	gamma := link.NewGammaLink(s.sourceNote.ID(), s.targetNote.ID(), linkType, weight, md)
	s.Require().NoError(s.repo.Save(s.ctx, gamma))

	supType := "related"
	sup := link.NewSuppression(s.sourceNote.ID(), s.targetNote.ID(), &supType, nil)
	s.Require().NoError(s.repo.DeleteAndSuppress(s.ctx, gamma, sup))
	// The row is already gone — a second request must still succeed.
	s.Require().NoError(s.repo.DeleteAndSuppress(s.ctx, gamma, sup))

	sups, err := s.repo.FindSuppressionsForNotes(s.ctx, []uuid.UUID{s.sourceNote.ID()})
	s.Require().NoError(err)
	s.Len(sups, 1, "duplicate delete must not duplicate the rejection")
}

// TestSuppression_NullTypeBlocksEverything — link_type NULL переживает
// round-trip через БД и блокирует любой тип предложения на паре.
func (s *LinkRepositoryIntegrationTestSuite) TestSuppression_NullTypeBlocksEverything() {
	s.Require().NoError(s.repo.SaveSuppression(s.ctx,
		link.NewSuppression(s.sourceNote.ID(), s.targetNote.ID(), nil, nil)))

	sups, err := s.repo.FindSuppressionsForNotes(s.ctx, []uuid.UUID{s.sourceNote.ID()})
	s.Require().NoError(err)
	s.Require().Len(sups, 1)
	s.Nil(sups[0].LinkType(), "NULL link_type must round-trip as nil")
	s.True(sups[0].Blocks(s.sourceNote.ID(), s.targetNote.ID(), "related"))
	s.True(sups[0].Blocks(s.targetNote.ID(), s.sourceNote.ID(), "dependency"),
		"untyped rejection must block every type in either direction")
}

// Запускаем тесты
func TestLinkRepositoryIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	suite.Run(t, new(LinkRepositoryIntegrationTestSuite))
}
