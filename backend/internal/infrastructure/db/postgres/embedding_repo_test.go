//go:build integration
// +build integration

package postgres

import (
	"context"
	"testing"

	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/testutil"

	"github.com/pgvector/pgvector-go"
)

func TestEmbeddingRepository_UpsertAndFind(t *testing.T) {
	db, cleanup := testutil.SetupTestVectorDB(t)
	defer cleanup()

	// Включаем расширение pgvector и мигрируем зависимые модели
	db.Exec("CREATE EXTENSION IF NOT EXISTS vector")
	if err := db.AutoMigrate(&UserModel{}, &NoteModel{}, &NoteEmbeddingModel{}); err != nil {
		t.Fatalf("failed to migrate models: %v", err)
	}

	repo := NewEmbeddingRepository(db, "all-MiniLM-L6-v2")

	// Создаем заметку сначала (для foreign key)
	noteRepo := NewNoteRepository(db, nil)
	title, _ := note.NewTitle("Embedding Test")
	content, _ := note.NewContent("Test content for embedding")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, "star", metadata)

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
	db, cleanup := testutil.SetupTestVectorDB(t)
	defer cleanup()

	// Включаем расширение pgvector и мигрируем зависимые модели
	db.Exec("CREATE EXTENSION IF NOT EXISTS vector")
	if err := db.AutoMigrate(&UserModel{}, &NoteModel{}, &NoteEmbeddingModel{}); err != nil {
		t.Fatalf("failed to migrate models: %v", err)
	}

	repo := NewEmbeddingRepository(db, "all-MiniLM-L6-v2")

	// Создаем заметку
	noteRepo := NewNoteRepository(db, nil)
	title, _ := note.NewTitle("Update Test")
	content, _ := note.NewContent("Test content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, "star", metadata)

	ctx := context.Background()
	if err := noteRepo.Save(ctx, n); err != nil {
		t.Fatalf("Save note failed: %v", err)
	}

	// Первый эмбеддинг
	embedding1 := make([]float32, 384)
	for i := range embedding1 {
		embedding1[i] = float32(i) / 100.0
	}
	vec1 := pgvector.NewVector(embedding1)

	if err := repo.Upsert(ctx, n.ID(), vec1); err != nil {
		t.Fatalf("Upsert embedding failed: %v", err)
	}

	// Обновляем эмбеддинг
	embedding2 := make([]float32, 384)
	for i := range embedding2 {
		embedding2[i] = float32(i+1) / 100.0
	}
	vec2 := pgvector.NewVector(embedding2)

	if err := repo.Upsert(ctx, n.ID(), vec2); err != nil {
		t.Fatalf("Upsert update failed: %v", err)
	}

	// Проверяем что остался 1 эмбеддинг
	var count int64
	db.Model(&NoteEmbeddingModel{}).Where("note_id = ?", n.ID()).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 embedding after update, got %d", count)
	}

	// After an Upsert the model_name column matches the repo's configured model
	var modelName string
	db.Model(&NoteEmbeddingModel{}).Where("note_id = ?", n.ID()).Select("model_name").Scan(&modelName)
	if modelName != "all-MiniLM-L6-v2" {
		t.Errorf("expected model_name to be 'all-MiniLM-L6-v2', got %q", modelName)
	}
}

func TestEmbeddingRepository_ModelFiltering(t *testing.T) {
	db, cleanup := testutil.SetupTestVectorDB(t)
	defer cleanup()

	db.Exec("CREATE EXTENSION IF NOT EXISTS vector")
	if err := db.AutoMigrate(&UserModel{}, &NoteModel{}, &NoteEmbeddingModel{}); err != nil {
		t.Fatalf("failed to migrate models: %v", err)
	}

	currentRepo := NewEmbeddingRepository(db, "paraphrase-multilingual-MiniLM-L12-v2")
	oldRepo := NewEmbeddingRepository(db, "all-MiniLM-L6-v2")
	noteRepo := NewNoteRepository(db, nil)
	ctx := context.Background()

	createNote := func(title string) *note.Note {
		titleV, _ := note.NewTitle(title)
		content, _ := note.NewContent("content")
		metadata, _ := note.NewMetadata(nil)
		n := note.NewNote(titleV, content, "star", metadata)
		if err := noteRepo.Save(ctx, n); err != nil {
			t.Fatalf("Save note failed: %v", err)
		}
		return n
	}

	// Two notes with vectors from the current (multilingual) model.
	n1 := createNote("Note one")
	n2 := createNote("Note two")
	vec := func() pgvector.Vector {
		v := make([]float32, 384)
		for i := range v {
			v[i] = float32(i) / 100.0
		}
		return pgvector.NewVector(v)
	}
	if err := currentRepo.Upsert(ctx, n1.ID(), vec()); err != nil {
		t.Fatalf("Upsert n1 current failed: %v", err)
	}
	if err := currentRepo.Upsert(ctx, n2.ID(), vec()); err != nil {
		t.Fatalf("Upsert n2 current failed: %v", err)
	}

	// A third note still has the old English-only vector.
	// This models a partially migrated state.
	n3 := createNote("Old model note")
	if err := oldRepo.Upsert(ctx, n3.ID(), vec()); err != nil {
		t.Fatalf("Upsert n3 old failed: %v", err)
	}

	// Similar-note search for the current model must not mix in the old vector.
	similar, err := currentRepo.FindSimilarNotes(ctx, n1.ID(), 10)
	if err != nil {
		t.Fatalf("FindSimilarNotes failed: %v", err)
	}
	if len(similar) != 1 {
		t.Errorf("expected 1 current-model similar note, got %d", len(similar))
	}
	if len(similar) > 0 && similar[0].NoteID != n2.ID() {
		t.Errorf("expected n2 as the only similar note, got %v", similar[0].NoteID)
	}

	// Missing model query should flag the note that only has the old vector.
	missing, err := currentRepo.FindNoteIDsMissingModel(ctx)
	if err != nil {
		t.Fatalf("FindNoteIDsMissingModel failed: %v", err)
	}
	if len(missing) != 1 || missing[0] != n3.ID() {
		t.Errorf("expected [%v] missing for current model, got %v", n3.ID(), missing)
	}
}
