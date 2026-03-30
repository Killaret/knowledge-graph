//go:build integration
// +build integration

package postgres

import (
	"context"
	"testing"

	"knowledge-graph/internal/domain/link"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDBForLink(t *testing.T) *gorm.DB {
	dsn := "host=localhost user=kb_user password=kb_password dbname=knowledge_base_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect test db: %v", err)
	}
	// Создаём таблицы notes и links (связь зависит от notes)
	if err := db.AutoMigrate(&NoteModel{}, &LinkModel{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	// Очищаем таблицы перед каждым тестом
	db.Exec("DELETE FROM links")
	db.Exec("DELETE FROM notes")
	return db
}

func TestLinkRepository_SaveAndFind(t *testing.T) {
	db := setupTestDBForLink(t)
	repo := NewLinkRepository(db)

	// Сначала создадим заметки, чтобы ссылаться на них
	note1 := NoteModel{ID: uuid.New(), Title: "Source", Content: "src"}
	note2 := NoteModel{ID: uuid.New(), Title: "Target", Content: "tgt"}
	db.Create(&note1)
	db.Create(&note2)

	linkType, _ := link.NewLinkType("reference")
	weight, _ := link.NewWeight(0.7)
	metadata, _ := link.NewMetadata(nil)

	l := link.NewLink(note1.ID, note2.ID, linkType, weight, metadata)

	ctx := context.Background()
	err := repo.Save(ctx, l)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	found, err := repo.FindByID(ctx, l.ID())
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("link not found")
	}
	if found.SourceNoteID() != note1.ID {
		t.Error("source ID mismatch")
	}
	if found.TargetNoteID() != note2.ID {
		t.Error("target ID mismatch")
	}
	if found.LinkType().String() != "reference" {
		t.Error("link type mismatch")
	}
	if found.Weight().Value() != 0.7 {
		t.Error("weight mismatch")
	}
}

func TestLinkRepository_FindBySource(t *testing.T) {
	db := setupTestDBForLink(t)
	repo := NewLinkRepository(db)

	// Создаём две заметки
	note1 := NoteModel{ID: uuid.New(), Title: "Source1", Content: ""}
	note2 := NoteModel{ID: uuid.New(), Title: "Target1", Content: ""}
	note3 := NoteModel{ID: uuid.New(), Title: "Target2", Content: ""}
	db.Create(&note1)
	db.Create(&note2)
	db.Create(&note3)

	linkType, _ := link.NewLinkType("reference")
	weight, _ := link.NewWeight(1.0)
	metadata, _ := link.NewMetadata(nil)

	l1 := link.NewLink(note1.ID, note2.ID, linkType, weight, metadata)
	l2 := link.NewLink(note1.ID, note3.ID, linkType, weight, metadata)
	ctx := context.Background()
	_ = repo.Save(ctx, l1)
	_ = repo.Save(ctx, l2)

	links, err := repo.FindBySource(ctx, note1.ID)
	if err != nil {
		t.Fatalf("FindBySource failed: %v", err)
	}
	if len(links) != 2 {
		t.Errorf("expected 2 links, got %d", len(links))
	}
}

func TestLinkRepository_DeleteBySource(t *testing.T) {
	db := setupTestDBForLink(t)
	repo := NewLinkRepository(db)

	note1 := NoteModel{ID: uuid.New(), Title: "SourceDel", Content: ""}
	note2 := NoteModel{ID: uuid.New(), Title: "TargetDel", Content: ""}
	db.Create(&note1)
	db.Create(&note2)

	linkType, _ := link.NewLinkType("reference")
	weight, _ := link.NewWeight(1.0)
	metadata, _ := link.NewMetadata(nil)

	l := link.NewLink(note1.ID, note2.ID, linkType, weight, metadata)
	ctx := context.Background()
	_ = repo.Save(ctx, l)

	err := repo.DeleteBySource(ctx, note1.ID)
	if err != nil {
		t.Fatalf("DeleteBySource failed: %v", err)
	}

	links, _ := repo.FindBySource(ctx, note1.ID)
	if len(links) != 0 {
		t.Error("links still exist after DeleteBySource")
	}
}
