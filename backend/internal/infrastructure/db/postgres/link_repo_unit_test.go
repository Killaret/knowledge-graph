package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"knowledge-graph/internal/domain/link"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupMockDBForLink создаёт mock базы данных для тестирования связей
func setupMockDBForLink(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}

	dialector := postgres.New(postgres.Config{
		Conn: sqlDB,
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm: %v", err)
	}

	cleanup := func() {
		sqlDB.Close()
	}

	return db, mock, cleanup
}

// TestLinkRepository_Save_Create тестирует создание новой связи
func TestLinkRepository_Save_Create(t *testing.T) {
	db, mock, cleanup := setupMockDBForLink(t)
	defer cleanup()

	repo := NewLinkRepository(db)

	sourceID := uuid.New()
	targetID := uuid.New()
	linkType, _ := link.NewLinkType("reference")
	weight, _ := link.NewWeight(0.8)
	metadata, _ := link.NewMetadata(map[string]interface{}{"strength": "high"})

	l := link.NewLink(sourceID, targetID, linkType, weight, metadata)

	// Ожидаем запрос на проверку существования
	mock.ExpectQuery(`SELECT \* FROM "links" WHERE id = \$1 ORDER BY "links"."id" LIMIT \$2`).
		WithArgs(l.ID(), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// Ожидаем INSERT — GORM использует Query с RETURNING для PostgreSQL
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "links" \("source_id","target_id","link_type","weight","metadata","created_at","id"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7\) RETURNING "id"`).
		WithArgs(
			sourceID,
			targetID,
			"reference",
			0.8,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(l.ID()))
	mock.ExpectCommit()

	ctx := context.Background()
	err := repo.Save(ctx, l)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestLinkRepository_FindByID_Found тестирует поиск существующей связи
func TestLinkRepository_FindByID_Found(t *testing.T) {
	db, mock, cleanup := setupMockDBForLink(t)
	defer cleanup()

	repo := NewLinkRepository(db)

	id := uuid.New()
	sourceID := uuid.New()
	targetID := uuid.New()
	now := time.Now()

	// Ожидаем SELECT
	mock.ExpectQuery(`SELECT \* FROM "links" WHERE id = \$1 ORDER BY "links"."id" LIMIT \$2`).
		WithArgs(id, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "source_id", "target_id", "link_type", "weight", "metadata", "created_at"}).
			AddRow(id, sourceID, targetID, "reference", 0.7, `{}`, now))

	ctx := context.Background()
	found, err := repo.FindByID(ctx, id)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if found == nil {
		t.Error("expected link to be found, got nil")
	}
	if found.SourceNoteID() != sourceID {
		t.Error("source ID mismatch")
	}
	if found.TargetNoteID() != targetID {
		t.Error("target ID mismatch")
	}
	if found.LinkType().String() != "reference" {
		t.Errorf("expected link type 'reference', got %s", found.LinkType().String())
	}
	if found.Weight().Value() != 0.7 {
		t.Errorf("expected weight 0.7, got %f", found.Weight().Value())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestLinkRepository_FindByID_NotFound тестирует поиск несуществующей связи
func TestLinkRepository_FindByID_NotFound(t *testing.T) {
	db, mock, cleanup := setupMockDBForLink(t)
	defer cleanup()

	repo := NewLinkRepository(db)

	id := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "links" WHERE id = \$1 ORDER BY "links"."id" LIMIT \$2`).
		WithArgs(id, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	ctx := context.Background()
	found, err := repo.FindByID(ctx, id)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if found != nil {
		t.Error("expected nil for non-existing link")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestLinkRepository_FindBySource тестирует поиск связей по source ID
func TestLinkRepository_FindBySource(t *testing.T) {
	db, mock, cleanup := setupMockDBForLink(t)
	defer cleanup()

	repo := NewLinkRepository(db)

	sourceID := uuid.New()
	targetID1 := uuid.New()
	targetID2 := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM "links" WHERE source_id = \$1`).
		WithArgs(sourceID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "source_id", "target_id", "link_type", "weight", "metadata", "created_at"}).
			AddRow(uuid.New(), sourceID, targetID1, "reference", 0.8, `{}`, now).
			AddRow(uuid.New(), sourceID, targetID2, "related", 0.6, `{}`, now))

	ctx := context.Background()
	links, err := repo.FindBySource(ctx, sourceID)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(links) != 2 {
		t.Errorf("expected 2 links, got %d", len(links))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestLinkRepository_FindByTarget тестирует поиск связей по target ID
func TestLinkRepository_FindByTarget(t *testing.T) {
	db, mock, cleanup := setupMockDBForLink(t)
	defer cleanup()

	repo := NewLinkRepository(db)

	targetID := uuid.New()
	sourceID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM "links" WHERE target_id = \$1`).
		WithArgs(targetID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "source_id", "target_id", "link_type", "weight", "metadata", "created_at"}).
			AddRow(uuid.New(), sourceID, targetID, "reference", 0.9, `{}`, now))

	ctx := context.Background()
	links, err := repo.FindByTarget(ctx, targetID)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(links) != 1 {
		t.Errorf("expected 1 link, got %d", len(links))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestLinkRepository_FindBySourceIDs_Batch тестирует batch-запрос связей
func TestLinkRepository_FindBySourceIDs_Batch(t *testing.T) {
	db, mock, cleanup := setupMockDBForLink(t)
	defer cleanup()

	repo := NewLinkRepository(db)

	sourceID1 := uuid.New()
	sourceID2 := uuid.New()
	targetID := uuid.New()
	now := time.Now()

	sourceIDs := []uuid.UUID{sourceID1, sourceID2}

	mock.ExpectQuery(`SELECT \* FROM "links" WHERE source_id IN \(\$1,\$2\)`).
		WithArgs(sourceID1, sourceID2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "source_id", "target_id", "link_type", "weight", "metadata", "created_at"}).
			AddRow(uuid.New(), sourceID1, targetID, "reference", 0.8, `{}`, now).
			AddRow(uuid.New(), sourceID2, targetID, "related", 0.7, `{}`, now))

	ctx := context.Background()
	result, err := repo.FindBySourceIDs(ctx, sourceIDs)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 entries in result, got %d", len(result))
	}
	if len(result[sourceID1]) != 1 {
		t.Errorf("expected 1 link for sourceID1, got %d", len(result[sourceID1]))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestLinkRepository_FindBySourceIDs_EmptyInput тестирует batch-запрос с пустым входом
func TestLinkRepository_FindBySourceIDs_EmptyInput(t *testing.T) {
	db, _, cleanup := setupMockDBForLink(t)
	defer cleanup()

	repo := NewLinkRepository(db)

	ctx := context.Background()
	result, err := repo.FindBySourceIDs(ctx, []uuid.UUID{})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d", len(result))
	}
}

// TestLinkRepository_FindByTargetIDs_Batch тестирует batch-запрос по target IDs
func TestLinkRepository_FindByTargetIDs_Batch(t *testing.T) {
	db, mock, cleanup := setupMockDBForLink(t)
	defer cleanup()

	repo := NewLinkRepository(db)

	targetID1 := uuid.New()
	targetID2 := uuid.New()
	sourceID := uuid.New()
	now := time.Now()

	targetIDs := []uuid.UUID{targetID1, targetID2}

	mock.ExpectQuery(`SELECT \* FROM "links" WHERE target_id IN \(\$1,\$2\)`).
		WithArgs(targetID1, targetID2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "source_id", "target_id", "link_type", "weight", "metadata", "created_at"}).
			AddRow(uuid.New(), sourceID, targetID1, "reference", 0.8, `{}`, now).
			AddRow(uuid.New(), sourceID, targetID2, "related", 0.7, `{}`, now))

	ctx := context.Background()
	result, err := repo.FindByTargetIDs(ctx, targetIDs)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 entries in result, got %d", len(result))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestLinkRepository_Delete тестирует удаление связи
func TestLinkRepository_Delete(t *testing.T) {
	db, mock, cleanup := setupMockDBForLink(t)
	defer cleanup()

	repo := NewLinkRepository(db)

	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "links" WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ctx := context.Background()
	err := repo.Delete(ctx, id)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestLinkRepository_DeleteBySource тестирует удаление всех связей по source ID
func TestLinkRepository_DeleteBySource(t *testing.T) {
	db, mock, cleanup := setupMockDBForLink(t)
	defer cleanup()

	repo := NewLinkRepository(db)

	sourceID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "links" WHERE source_id = \$1`).
		WithArgs(sourceID).
		WillReturnResult(sqlmock.NewResult(0, 3)) // Удалено 3 связи
	mock.ExpectCommit()

	ctx := context.Background()
	err := repo.DeleteBySource(ctx, sourceID)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestLinkRepository_FindAll тестирует получение всех связей
func TestLinkRepository_FindAll(t *testing.T) {
	db, mock, cleanup := setupMockDBForLink(t)
	defer cleanup()

	repo := NewLinkRepository(db)

	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM "links"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "source_id", "target_id", "link_type", "weight", "metadata", "created_at"}).
			AddRow(uuid.New(), uuid.New(), uuid.New(), "reference", 0.8, `{}`, now).
			AddRow(uuid.New(), uuid.New(), uuid.New(), "related", 0.6, `{}`, now))

	ctx := context.Background()
	links, err := repo.FindAll(ctx)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(links) != 2 {
		t.Errorf("expected 2 links, got %d", len(links))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestLinkRepository_Save_Update тестирует обновление существующей связи
func TestLinkRepository_Save_Update(t *testing.T) {
	db, mock, cleanup := setupMockDBForLink(t)
	defer cleanup()

	repo := NewLinkRepository(db)

	sourceID := uuid.New()
	targetID := uuid.New()
	linkType, _ := link.NewLinkType("reference")
	weight, _ := link.NewWeight(0.9)
	metadata, _ := link.NewMetadata(nil)

	l := link.NewLink(sourceID, targetID, linkType, weight, metadata)
	now := time.Now()

	// Ожидаем запрос на проверку существования - запись найдена
	mock.ExpectQuery(`SELECT \* FROM "links" WHERE id = \$1 ORDER BY "links"."id" LIMIT \$2`).
		WithArgs(l.ID(), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "source_id", "target_id", "link_type", "weight", "metadata", "created_at"}).
			AddRow(l.ID(), sourceID, targetID, "reference", 0.5, `{}`, now))

	// Ожидаем UPDATE
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "links" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ctx := context.Background()
	err := repo.Save(ctx, l)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestToGormLink тестирует конвертацию доменной связи в GORM модель
func TestToGormLink(t *testing.T) {
	sourceID := uuid.New()
	targetID := uuid.New()
	linkType, _ := link.NewLinkType("reference")
	weight, _ := link.NewWeight(0.8)
	metadata, _ := link.NewMetadata(map[string]interface{}{"strength": "high"})

	l := link.NewLink(sourceID, targetID, linkType, weight, metadata)

	model, err := toGormLink(l)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if model.ID != l.ID() {
		t.Error("ID mismatch")
	}
	if model.SourceNoteID != sourceID {
		t.Error("source ID mismatch")
	}
	if model.TargetNoteID != targetID {
		t.Error("target ID mismatch")
	}
	if model.LinkType != "reference" {
		t.Errorf("expected link type 'reference', got %s", model.LinkType)
	}
	if model.Weight != 0.8 {
		t.Errorf("expected weight 0.8, got %f", model.Weight)
	}
}

// TestToDomainLink тестирует конвертацию GORM модели в доменную связь
func TestToDomainLink(t *testing.T) {
	id := uuid.New()
	sourceID := uuid.New()
	targetID := uuid.New()
	now := time.Now()

	model := &LinkModel{
		ID:           id,
		SourceNoteID: sourceID,
		TargetNoteID: targetID,
		LinkType:     "reference",
		Weight:       0.75,
		Metadata:     []byte(`{"key":"value"}`),
		CreatedAt:    now,
	}

	l, err := toDomainLink(model)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if l.ID() != id {
		t.Error("ID mismatch")
	}
	if l.SourceNoteID() != sourceID {
		t.Error("source ID mismatch")
	}
	if l.TargetNoteID() != targetID {
		t.Error("target ID mismatch")
	}
	if l.LinkType().String() != "reference" {
		t.Errorf("expected link type 'reference', got %s", l.LinkType().String())
	}
	if l.Weight().Value() != 0.75 {
		t.Errorf("expected weight 0.75, got %f", l.Weight().Value())
	}
}

// TestToDomainLinks тестирует конвертацию списка моделей
func TestToDomainLinks(t *testing.T) {
	now := time.Now()
	models := []LinkModel{
		{ID: uuid.New(), SourceNoteID: uuid.New(), TargetNoteID: uuid.New(), LinkType: "reference", Weight: 0.8, CreatedAt: now},
		{ID: uuid.New(), SourceNoteID: uuid.New(), TargetNoteID: uuid.New(), LinkType: "related", Weight: 0.6, CreatedAt: now},
	}

	links := toDomainLinks(models)

	if len(links) != 2 {
		t.Errorf("expected 2 links, got %d", len(links))
	}
}

// TestToDomainLinks_WithInvalidData тестирует пропуск невалидных данных
func TestToDomainLinks_WithInvalidData(t *testing.T) {
	now := time.Now()
	// Создаём модель с невалидным весом (больше 1.0)
	models := []LinkModel{
		{ID: uuid.New(), SourceNoteID: uuid.New(), TargetNoteID: uuid.New(), LinkType: "reference", Weight: 0.8, CreatedAt: now},
		{ID: uuid.New(), SourceNoteID: uuid.New(), TargetNoteID: uuid.New(), LinkType: "invalid", Weight: 0.5, CreatedAt: now}, // Невалидный тип
	}

	links := toDomainLinks(models)

	// Должна остаться только одна валидная связь
	if len(links) != 1 {
		t.Errorf("expected 1 valid link, got %d", len(links))
	}
	if links[0].Weight().Value() != 0.8 {
		t.Errorf("expected weight 0.8, got %f", links[0].Weight().Value())
	}
}

// TestLinkRepository_FindBySource_DBError тестирует обработку ошибок БД
func TestLinkRepository_FindBySource_DBError(t *testing.T) {
	db, mock, cleanup := setupMockDBForLink(t)
	defer cleanup()

	repo := NewLinkRepository(db)

	sourceID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "links" WHERE source_id = \$1`).
		WithArgs(sourceID).
		WillReturnError(errors.New("database connection failed"))

	ctx := context.Background()
	_, err := repo.FindBySource(ctx, sourceID)

	if err == nil {
		t.Error("expected error for DB failure, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestLinkRepository_FindByTarget_EmptyResult тестирует пустой результат
func TestLinkRepository_FindByTarget_EmptyResult(t *testing.T) {
	db, mock, cleanup := setupMockDBForLink(t)
	defer cleanup()

	repo := NewLinkRepository(db)

	targetID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "links" WHERE target_id = \$1`).
		WithArgs(targetID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "source_id", "target_id", "link_type", "weight", "metadata", "created_at"}))

	ctx := context.Background()
	links, err := repo.FindByTarget(ctx, targetID)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(links) != 0 {
		t.Errorf("expected 0 links, got %d", len(links))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
