//go:build integration
// +build integration

package postgres

import (
	"context"
	"testing"

	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDBForEmbedding(t *testing.T) *gorm.DB {
	dsn := "host=localhost user=kb_user password=kb_password dbname=knowledge_base_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect test db: %v", err)
	}
	// Автомиграция (создаст таблицы notes и note_embeddings)
	if err := db.AutoMigrate(&NoteModel{}, &NoteEmbeddingModel{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	// Очистка таблиц перед каждым тестом
	db.Exec("DELETE FROM note_embeddings")
	db.Exec("DELETE FROM notes")
	return db
}

func TestEmbeddingRepository_UpsertAndFind(t *testing.T) {
	db := setupTestDBForEmbedding(t)
	repo := NewEmbeddingRepository(db)

	// Создаем заметку сначала (для foreign key)
	noteRepo := NewNoteRepository(db)
	title, _ := note.NewTitle("Embedding Test")
	content, _ := note.NewContent("Test content for embedding")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, metadata)

	ctx := context.Background()
	if err := noteRepo.Save(ctx, n); err != nil {
		t.Fatalf("Save note failed: %v", err)
	}

	// Создаем эмбеддинг (384 dimensions - стандарт для all-MiniLM-L6-v2)
	embedding := make([]float32, 384)
	for i := range embedding {
		embedding[i] = float32(i) / 100.0
	}
	vec := pgvector.NewVector(embedding)

	// Upsert эмбеддинг
	err := repo.Upsert(ctx, n.ID(), vec)
	if err != nil {
		t.Fatalf("Upsert embedding failed: %v", err)
	}

	// Проверяем что эмбеддинг сохранен (через прямой запрос)
	var count int64
	db.Model(&NoteEmbeddingModel{}).Where("note_id = ?", n.ID()).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 embedding, got %d", count)
	}
}

func TestEmbeddingRepository_UpsertUpdate(t *testing.T) {
	db := setupTestDBForEmbedding(t)
	repo := NewEmbeddingRepository(db)

	// Создаем заметку
	noteRepo := NewNoteRepository(db)
	title, _ := note.NewTitle("Update Test")
	content, _ := note.NewContent("Test content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, metadata)

	ctx := context.Background()
	if err := noteRepo.Save(ctx, n); err != nil {
		t.Fatalf("Save note failed: %v", err)
	}

	// Первый upsert
	embedding1 := make([]float32, 384)
	for i := range embedding1 {
		embedding1[i] = 0.1
	}
	vec1 := pgvector.NewVector(embedding1)
	if err := repo.Upsert(ctx, n.ID(), vec1); err != nil {
		t.Fatalf("First upsert failed: %v", err)
	}

	// Второй upsert (должен обновить)
	embedding2 := make([]float32, 384)
	for i := range embedding2 {
		embedding2[i] = 0.9
	}
	vec2 := pgvector.NewVector(embedding2)
	if err := repo.Upsert(ctx, n.ID(), vec2); err != nil {
		t.Fatalf("Second upsert failed: %v", err)
	}

	// Проверяем что все еще 1 запись (не 2)
	var count int64
	db.Model(&NoteEmbeddingModel{}).Where("note_id = ?", n.ID()).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 embedding after update, got %d", count)
	}
}

func TestEmbeddingRepository_Delete(t *testing.T) {
	db := setupTestDBForEmbedding(t)
	repo := NewEmbeddingRepository(db)

	// Создаем заметку
	noteRepo := NewNoteRepository(db)
	title, _ := note.NewTitle("Delete Test")
	content, _ := note.NewContent("Test content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, metadata)

	ctx := context.Background()
	if err := noteRepo.Save(ctx, n); err != nil {
		t.Fatalf("Save note failed: %v", err)
	}

	// Создаем эмбеддинг
	embedding := make([]float32, 384)
	vec := pgvector.NewVector(embedding)
	if err := repo.Upsert(ctx, n.ID(), vec); err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	// Удаляем
	if err := repo.Delete(ctx, n.ID()); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Проверяем что удалено
	var count int64
	db.Model(&NoteEmbeddingModel{}).Where("note_id = ?", n.ID()).Count(&count)
	if count != 0 {
		t.Errorf("expected 0 embeddings after delete, got %d", count)
	}
}

func TestEmbeddingRepository_FindSimilarNotes(t *testing.T) {
	db := setupTestDBForEmbedding(t)
	repo := NewEmbeddingRepository(db)
	noteRepo := NewNoteRepository(db)

	ctx := context.Background()

	// Создаем 3 заметки
	notes := make([]*note.Note, 3)
	for i := 0; i < 3; i++ {
		title, _ := note.NewTitle("Note " + string(rune('A'+i)))
		content, _ := note.NewContent("Content " + string(rune('A'+i)))
		metadata, _ := note.NewMetadata(nil)
		notes[i] = note.NewNote(title, content, metadata)
		if err := noteRepo.Save(ctx, notes[i]); err != nil {
			t.Fatalf("Save note %d failed: %v", i, err)
		}
	}

	// Создаем эмбеддинги
	// Note A - [1, 0, 0, ...]
	// Note B - [0.9, 0.1, 0, ...] (похож на A)
	// Note C - [0, 1, 0, ...] (не похож на A)
	for i, n := range notes {
		embedding := make([]float32, 384)
		if i == 0 {
			embedding[0] = 1.0
		} else if i == 1 {
			embedding[0] = 0.9
			embedding[1] = 0.1
		} else {
			embedding[1] = 1.0
		}
		vec := pgvector.NewVector(embedding)
		if err := repo.Upsert(ctx, n.ID(), vec); err != nil {
			t.Fatalf("Upsert embedding %d failed: %v", i, err)
		}
	}

	// Ищем похожие на Note A
	similar, err := repo.FindSimilarNotes(ctx, notes[0].ID(), 10)
	if err != nil {
		t.Fatalf("FindSimilarNotes failed: %v", err)
	}

	// Должны найти Note B и Note C
	if len(similar) != 2 {
		t.Errorf("expected 2 similar notes, got %d", len(similar))
	}

	// Note B должен быть более похож чем Note C
	if len(similar) >= 2 {
		// Note B имеет большее сходство (0.9 по первому измерению)
		// Note C имеет 0 по первому измерению
		if similar[0].Score < similar[1].Score {
			t.Error("expected first result to have higher similarity")
		}
	}
}

func TestEmbeddingRepository_FindSimilarNotesNoEmbedding(t *testing.T) {
	db := setupTestDBForEmbedding(t)
	repo := NewEmbeddingRepository(db)
	noteRepo := NewNoteRepository(db)

	ctx := context.Background()

	// Создаем заметку без эмбеддинга
	title, _ := note.NewTitle("No Embedding")
	content, _ := note.NewContent("Test")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, metadata)
	if err := noteRepo.Save(ctx, n); err != nil {
		t.Fatalf("Save note failed: %v", err)
	}

	// Ищем похожие - должно вернуть пустой результат без ошибки
	similar, err := repo.FindSimilarNotes(ctx, n.ID(), 10)
	if err != nil {
		// Может вернуть ошибку или пустой результат - оба валидно
		t.Logf("FindSimilarNotes returned error (expected): %v", err)
		return
	}

	if len(similar) != 0 {
		t.Errorf("expected 0 similar notes for note without embedding, got %d", len(similar))
	}
}
