//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	apprec "knowledge-graph/internal/application/recommendation"
	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/testutil"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

// ISOLATION-1: the gamma link generator must never create an auto-link
// between notes of different owners, even when the foreign note is the
// nearest neighbour in vector space.
func TestGammaLinkGenerator_OwnerIsolation(t *testing.T) {
	db, cleanup := testutil.SetupTestVectorDB(t)
	defer cleanup()

	db.Exec("CREATE EXTENSION IF NOT EXISTS vector")
	models := []interface{}{
		&UserModel{}, &NoteModel{}, &LinkModel{}, &LinkSuppressionModel{},
		&NoteKeywordModel{}, &TagModel{}, &NoteTagModel{}, &NoteEmbeddingModel{},
	}
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	ctx := context.Background()
	noteRepo := NewNoteRepository(db, nil)
	embRepo := NewEmbeddingRepository(db, "paraphrase-multilingual-MiniLM-L12-v2")
	linkRepo := NewLinkRepository(db)
	gen := apprec.NewGammaLinkGenerator(embRepo, linkRepo, 3, 0.0)

	userA := uuid.New()
	userB := uuid.New()
	for i, u := range []uuid.UUID{userA, userB} {
		if err := db.Exec(`INSERT INTO users (id, login, email, password_hash, created_at)
			VALUES (?, ?, ?, 'x', ?)`, u, "iso"+string(rune('a'+i)), "iso"+u.String()[:8]+"@t.t", time.Now()).Error; err != nil {
			t.Fatalf("insert user: %v", err)
		}
	}

	mk := func(title string, owner uuid.UUID) uuid.UUID {
		ti, _ := note.NewTitle(title)
		c, _ := note.NewContent("content")
		m, _ := note.NewMetadata(nil)
		n := note.NewNoteWithCreator(ti, c, note.MustType("star"), m, owner)
		if err := noteRepo.Save(ctx, n); err != nil {
			t.Fatalf("save note: %v", err)
		}
		return n.ID()
	}

	noteA := mk("gamma owner A", userA)
	foreign := mk("gamma owner B", userB)

	base := make([]float32, 384)
	for i := range base {
		base[i] = float32(i) / 384.0
	}
	near := make([]float32, 384)
	copy(near, base)
	near[0] += 0.0001

	if err := embRepo.Upsert(ctx, noteA, pgvector.NewVector(base)); err != nil {
		t.Fatalf("upsert A: %v", err)
	}
	if err := embRepo.Upsert(ctx, foreign, pgvector.NewVector(near)); err != nil {
		t.Fatalf("upsert B: %v", err)
	}

	created, err := gen.GenerateForNote(ctx, noteA)
	if err != nil {
		t.Fatalf("GenerateForNote: %v", err)
	}
	for _, l := range created {
		if l.TargetNoteID() == foreign || l.SourceNoteID() == foreign {
			t.Fatalf("gamma link created to another user's note %s", foreign)
		}
	}

	// Direct check on persisted links too.
	links, err := linkRepo.FindBySource(ctx, noteA)
	if err != nil {
		t.Fatalf("FindBySource: %v", err)
	}
	for _, l := range links {
		if l.TargetNoteID() == foreign {
			t.Fatalf("persisted gamma link crosses owners: %s -> %s", noteA, foreign)
		}
	}
}
