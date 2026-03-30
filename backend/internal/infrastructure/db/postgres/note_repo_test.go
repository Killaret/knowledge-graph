//go:build integration
// +build integration

package postgres

import (
	"context"
	"testing"

	"knowledge-graph/internal/domain/note"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost user=kb_user password=kb_password dbname=knowledge_base_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect test db: %v", err)
	}
	// Автомиграция (создаст таблицу notes)
	if err := db.AutoMigrate(&NoteModel{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	// Очистка таблицы перед каждым тестом (опционально)
	db.Exec("DELETE FROM notes")
	return db
}

func TestNoteRepository_SaveAndFind(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNoteRepository(db)

	title, _ := note.NewTitle("Test")
	content, _ := note.NewContent("Content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, metadata)

	ctx := context.Background()
	err := repo.Save(ctx, n)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	found, err := repo.FindByID(ctx, n.ID())
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("note not found")
	}
	if found.Title().String() != n.Title().String() {
		t.Error("title mismatch")
	}
	if found.Content().String() != n.Content().String() {
		t.Error("content mismatch")
	}
}

func TestNoteRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNoteRepository(db)

	title, _ := note.NewTitle("Original")
	content, _ := note.NewContent("Content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, metadata)
	ctx := context.Background()
	if err := repo.Save(ctx, n); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Обновляем
	newTitle, _ := note.NewTitle("Updated")
	_ = n.UpdateTitle(newTitle)
	if err := repo.Save(ctx, n); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.FindByID(ctx, n.ID())
	if found.Title().String() != "Updated" {
		t.Error("title not updated in DB")
	}
}

func TestNoteRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNoteRepository(db)

	title, _ := note.NewTitle("ToDelete")
	content, _ := note.NewContent("Content")
	metadata, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, metadata)
	ctx := context.Background()
	if err := repo.Save(ctx, n); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if err := repo.Delete(ctx, n.ID()); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	found, _ := repo.FindByID(ctx, n.ID())
	if found != nil {
		t.Error("note still exists after delete")
	}
}
