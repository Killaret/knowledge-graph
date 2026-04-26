package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"knowledge-graph/internal/domain/note"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupMockDB создаёт mock базы данных для тестирования
func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
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

// TestNoteRepository_Save_Create тестирует создание новой заметки
func TestNoteRepository_Save_Create(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewNoteRepository(db, nil)

	title, _ := note.NewTitle("Test Title")
	content, _ := note.NewContent("Test Content")
	metadata, _ := note.NewMetadata(map[string]interface{}{"key": "value"})
	n := note.NewNote(title, content, "star", metadata)

	// Ожидаем запрос на проверку существования
	mock.ExpectQuery(`SELECT \* FROM "notes" WHERE id = \$1 ORDER BY "notes"."id" LIMIT \$2`).
		WithArgs(n.ID(), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// Ожидаем INSERT — GORM использует Query с RETURNING для PostgreSQL
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "notes" \("title","content","type","metadata","created_at","updated_at","id"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7\) RETURNING "id"`).
		WithArgs(
			"Test Title",
			"Test Content",
			"star",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(n.ID()))
	mock.ExpectCommit()

	ctx := context.Background()
	err := repo.Save(ctx, n)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestNoteRepository_FindByID_Found тестирует поиск существующей заметки
func TestNoteRepository_FindByID_Found(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewNoteRepository(db, nil)

	id := uuid.New()
	now := time.Now()

	// Ожидаем SELECT
	mock.ExpectQuery(`SELECT \* FROM "notes" WHERE id = \$1 ORDER BY "notes"."id" LIMIT \$2`).
		WithArgs(id, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "content", "type", "metadata", "created_at", "updated_at"}).
			AddRow(id, "Test", "Content", "star", `{}`, now, now))

	ctx := context.Background()
	found, err := repo.FindByID(ctx, id)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if found == nil {
		t.Error("expected note to be found, got nil")
	}
	if found.Title().String() != "Test" {
		t.Errorf("expected title 'Test', got %s", found.Title().String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestNoteRepository_FindByID_NotFound тестирует поиск несуществующей заметки
func TestNoteRepository_FindByID_NotFound(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewNoteRepository(db, nil)

	id := uuid.New()

	// Ожидаем SELECT с возвратом ErrRecordNotFound
	mock.ExpectQuery(`SELECT \* FROM "notes" WHERE id = \$1 ORDER BY "notes"."id" LIMIT \$2`).
		WithArgs(id, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	ctx := context.Background()
	found, err := repo.FindByID(ctx, id)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if found != nil {
		t.Error("expected nil for non-existing note")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestNoteRepository_Delete тестирует удаление заметки
func TestNoteRepository_Delete_Unit(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewNoteRepository(db, nil)

	id := uuid.New()

	// Ожидаем DELETE
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "notes" WHERE id = \$1`).
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

// TestNoteRepository_FindAll тестирует получение всех заметок
func TestNoteRepository_FindAll(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewNoteRepository(db, nil)

	now := time.Now()

	// Ожидаем SELECT с ORDER BY
	mock.ExpectQuery(`SELECT \* FROM "notes" ORDER BY created_at DESC`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "content", "type", "metadata", "created_at", "updated_at"}).
			AddRow(uuid.New(), "Note 1", "Content 1", "star", `{}`, now, now).
			AddRow(uuid.New(), "Note 2", "Content 2", "planet", `{}`, now, now))

	ctx := context.Background()
	notes, err := repo.FindAll(ctx)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(notes) != 2 {
		t.Errorf("expected 2 notes, got %d", len(notes))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestNoteRepository_List тестирует пагинацию
func TestNoteRepository_List(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewNoteRepository(db, nil)

	now := time.Now()

	// Ожидаем COUNT
	mock.ExpectQuery(`SELECT count\(\*\) FROM "notes"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

	// Ожидаем SELECT с LIMIT и OFFSET (GORM использует параметризованные запросы)
	mock.ExpectQuery(`SELECT \* FROM "notes" ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(5, 10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "content", "type", "metadata", "created_at", "updated_at"}).
			AddRow(uuid.New(), "Note 1", "Content 1", "star", `{}`, now, now).
			AddRow(uuid.New(), "Note 2", "Content 2", "planet", `{}`, now, now))

	ctx := context.Background()
	notes, total, err := repo.List(ctx, 5, 10)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if total != 10 {
		t.Errorf("expected total 10, got %d", total)
	}
	if len(notes) != 2 {
		t.Errorf("expected 2 notes, got %d", len(notes))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestNoteRepository_Save_Update тестирует обновление существующей заметки
func TestNoteRepository_Save_Update(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewNoteRepository(db, nil)

	title, _ := note.NewTitle("Updated Title")
	content, _ := note.NewContent("Updated Content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, "star", metadata)

	now := time.Now()

	// Ожидаем запрос на проверку существования - запись найдена
	mock.ExpectQuery(`SELECT \* FROM "notes" WHERE id = \$1 ORDER BY "notes"."id" LIMIT \$2`).
		WithArgs(n.ID(), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "content", "type", "metadata", "created_at", "updated_at"}).
			AddRow(n.ID(), "Old Title", "Old Content", "star", `{}`, now, now))

	// Ожидаем UPDATE
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "notes" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ctx := context.Background()
	err := repo.Save(ctx, n)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestNoteRepository_FindByID_DBError тестирует обработку ошибок БД
func TestNoteRepository_FindByID_DBError(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewNoteRepository(db, nil)

	id := uuid.New()

	// Ожидаем SELECT с ошибкой БД
	mock.ExpectQuery(`SELECT \* FROM "notes" WHERE id = \$1 ORDER BY "notes"."id" LIMIT \$2`).
		WithArgs(id, 1).
		WillReturnError(errors.New("database connection failed"))

	ctx := context.Background()
	_, err := repo.FindByID(ctx, id)

	if err == nil {
		t.Error("expected error for DB failure, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestToGormNote тестирует конвертацию доменной заметки в GORM модель
func TestToGormNote(t *testing.T) {
	title, _ := note.NewTitle("Test")
	content, _ := note.NewContent("Content")
	metadata, _ := note.NewMetadata(map[string]interface{}{"key": "value"})
	n := note.NewNote(title, content, "star", metadata)

	model, err := toGormNote(n)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if model.ID != n.ID() {
		t.Error("ID mismatch")
	}
	if model.Title != "Test" {
		t.Errorf("expected title 'Test', got %s", model.Title)
	}
	if model.Content != "Content" {
		t.Errorf("expected content 'Content', got %s", model.Content)
	}
	if model.Type != "star" {
		t.Errorf("expected type 'star', got %s", model.Type)
	}
}

// TestToDomainNote тестирует конвертацию GORM модели в доменную заметку
func TestToDomainNote(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	model := &NoteModel{
		ID:        id,
		Title:     "Test",
		Content:   "Content",
		Type:      "star",
		Metadata:  []byte(`{"key":"value"}`),
		CreatedAt: now,
		UpdatedAt: now,
	}

	n, err := toDomainNote(model)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if n.ID() != id {
		t.Error("ID mismatch")
	}
	if n.Title().String() != "Test" {
		t.Errorf("expected title 'Test', got %s", n.Title().String())
	}
}

// TestToDomainNotes тестирует конвертацию списка моделей
func TestToDomainNotes(t *testing.T) {
	now := time.Now()
	models := []NoteModel{
		{ID: uuid.New(), Title: "Note 1", Content: "Content 1", Type: "star", CreatedAt: now, UpdatedAt: now},
		{ID: uuid.New(), Title: "Note 2", Content: "Content 2", Type: "planet", CreatedAt: now, UpdatedAt: now},
	}

	notes := toDomainNotes(models)

	if len(notes) != 2 {
		t.Errorf("expected 2 notes, got %d", len(notes))
	}
}

// TestToDomainNotes_WithInvalidData тестирует пропуск невалидных данных
func TestToDomainNotes_WithInvalidData(t *testing.T) {
	now := time.Now()
	// Создаём модель с очень длинным title (больше 255 символов), что вызовет ошибку валидации
	longTitle := ""
	for i := 0; i < 300; i++ {
		longTitle += "a"
	}

	models := []NoteModel{
		{ID: uuid.New(), Title: "Valid", Content: "Content", Type: "star", CreatedAt: now, UpdatedAt: now},
		{ID: uuid.New(), Title: longTitle, Content: "Content", Type: "star", CreatedAt: now, UpdatedAt: now},
	}

	notes := toDomainNotes(models)

	// Должна остаться только одна валидная заметка
	if len(notes) != 1 {
		t.Errorf("expected 1 valid note, got %d", len(notes))
	}
	if notes[0].Title().String() != "Valid" {
		t.Errorf("expected 'Valid', got %s", notes[0].Title().String())
	}
}
