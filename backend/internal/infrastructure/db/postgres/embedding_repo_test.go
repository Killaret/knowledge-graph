//go:build integration
// +build integration

package postgres

import (
	"context"
	"testing"

	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/testutil"

	"github.com/google/uuid"
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
	n := note.NewNote(title, content, note.MustType("star"), metadata)

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
	n := note.NewNote(title, content, note.MustType("star"), metadata)

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
		n := note.NewNote(titleV, content, note.MustType("star"), metadata)
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

func TestEmbeddingRepository_FindSimilarNotesBatch(t *testing.T) {
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

	// Fixed UUIDs let us reason about PostgreSQL's ORDER BY note_id.
	// UUIDs sort by byte value, so the most-significant nibble determines order.
	source1ID := uuid.MustParse("60000000-0000-0000-0000-000000000001")
	source2ID := uuid.MustParse("70000000-0000-0000-0000-000000000001")
	targetCloseSmallID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	targetFarID := uuid.MustParse("20000000-0000-0000-0000-000000000001")
	targetCloseLargeID := uuid.MustParse("30000000-0000-0000-0000-000000000001")
	oldModelID := uuid.MustParse("40000000-0000-0000-0000-000000000001")
	noEmbeddingID := uuid.MustParse("50000000-0000-0000-0000-000000000001")

	closeVec := func() pgvector.Vector {
		v := make([]float32, 384)
		for i := range v {
			v[i] = float32(i) / 100.0
		}
		return pgvector.NewVector(v)
	}

	// Negating closeVec gives cosine similarity -1, cosine distance 2,
	// so the raw score (1 - distance) is -1 and clamping is required.
	farVec := func() pgvector.Vector {
		v := make([]float32, 384)
		for i := range v {
			v[i] = -float32(i) / 100.0
		}
		return pgvector.NewVector(v)
	}

	createNote := func(title string, id uuid.UUID) *note.Note {
		titleV, _ := note.NewTitle(title)
		content, _ := note.NewContent("content")
		metadata, _ := note.NewMetadata(nil)
		n := note.NewNote(titleV, content, note.MustType("star"), metadata, note.WithID(id))
		if err := noteRepo.Save(ctx, n); err != nil {
			t.Fatalf("Save note failed: %v", err)
		}
		return n
	}

	createNote("Source one", source1ID)
	createNote("Source two", source2ID)
	createNote("Close small", targetCloseSmallID)
	createNote("Far middle", targetFarID)
	createNote("Close large", targetCloseLargeID)
	createNote("Old model", oldModelID)
	createNote("No embedding", noEmbeddingID)

	// Both sources and the two close targets share the same vector.
	for _, id := range []uuid.UUID{source1ID, source2ID, targetCloseSmallID, targetCloseLargeID} {
		if err := currentRepo.Upsert(ctx, id, closeVec()); err != nil {
			t.Fatalf("Upsert close note %v failed: %v", id, err)
		}
	}

	// Far target and the cross-model note use the distant vector.
	if err := currentRepo.Upsert(ctx, targetFarID, farVec()); err != nil {
		t.Fatalf("Upsert far target failed: %v", err)
	}
	if err := oldRepo.Upsert(ctx, oldModelID, farVec()); err != nil {
		t.Fatalf("Upsert old-model note failed: %v", err)
	}

	// Three candidates per source, but limit is 2, so the per-source limit must cut
	// off targetFarID (lowest similarity) — not the largest note_id.
	batch, err := currentRepo.FindSimilarNotesBatch(ctx, []uuid.UUID{source1ID, source2ID, noEmbeddingID}, 2)
	if err != nil {
		t.Fatalf("FindSimilarNotesBatch failed: %v", err)
	}

	// 1. Non-zero score in [0, 1] and one result per requested id.
	for _, id := range []uuid.UUID{source1ID, source2ID} {
		similar, ok := batch[id]
		if !ok {
			t.Errorf("source %v: missing from batch results", id)
			continue
		}
		if len(similar) == 0 {
			t.Errorf("source %v: expected at least one similar note, got none", id)
			continue
		}
		foundPositive := false
		for _, s := range similar {
			if s.Score < 0 || s.Score > 1 {
				t.Errorf("source %v: score %v out of [0, 1]", id, s.Score)
			}
			if s.Score > 0 {
				foundPositive = true
			}
		}
		if !foundPositive {
			t.Errorf("source %v: expected at least one positive score", id)
		}
	}

	// 2. Limit is per source, not global — and selection is by similarity, not
	// by note_id. With three candidates and limit=2, each source gets the two
	// closest targets; the far one is discarded even though its UUID is smaller
	// than targetCloseLargeID's.
	for _, id := range []uuid.UUID{source1ID, source2ID} {
		similar, ok := batch[id]
		if !ok || len(similar) != 2 {
			t.Errorf("source %v: expected 2 results, got %d (ok=%v)", id, len(similar), ok)
			continue
		}

		expected := map[uuid.UUID]bool{
			targetCloseSmallID: true,
			targetCloseLargeID: true,
		}
		for _, s := range similar {
			if !expected[s.NoteID] {
				t.Errorf("source %v: unexpected note %v", id, s.NoteID)
			}
			if s.NoteID == targetFarID {
				t.Errorf("source %v: far note %v selected over a closer neighbour", id, s.NoteID)
			}
		}

		// Both close targets share the source vector; the e2.note_id tiebreaker
		// makes the smaller UUID come first.
		if similar[0].NoteID != targetCloseSmallID {
			t.Errorf("source %v: expected first result to be %v, got %v", id, targetCloseSmallID, similar[0].NoteID)
		}
		if similar[0].Score != 1.0 {
			t.Errorf("source %v: expected close score 1.0, got %v", id, similar[0].Score)
		}
		if similar[1].NoteID != targetCloseLargeID {
			t.Errorf("source %v: expected second result to be %v, got %v", id, targetCloseLargeID, similar[1].NoteID)
		}
		if similar[1].Score != 1.0 {
			t.Errorf("source %v: expected close score 1.0, got %v", id, similar[1].Score)
		}
	}

	// 3. Model filtering works: old-model note must not leak in.
	for _, similar := range batch {
		for _, s := range similar {
			if s.NoteID == oldModelID {
				t.Errorf("old-model note %v leaked into current-model batch", oldModelID)
			}
		}
	}

	// 4. A note without an embedding returns an empty slice, not an error.
	empty, ok := batch[noEmbeddingID]
	if ok && len(empty) != 0 {
		t.Errorf("expected no results for no-embedding note, got %d", len(empty))
	}

	// 5. Self-exclusion: without `e1.note_id != e2.note_id` the source itself
	// would be its own closest match (identical vector, score 1.0). A separate
	// call with a wide limit is required — at limit=2 the source's own note_id
	// sorts past the window and the defect is unobservable.
	wide, err := currentRepo.FindSimilarNotesBatch(ctx, []uuid.UUID{source1ID, source2ID}, 10)
	if err != nil {
		t.Fatalf("FindSimilarNotesBatch wide failed: %v", err)
	}
	for _, id := range []uuid.UUID{source1ID, source2ID} {
		for _, s := range wide[id] {
			if s.NoteID == id {
				t.Errorf("source %v: self-match leaked into results", id)
			}
		}
	}

	// 6. Score clamping: the far target's raw score is -1 (cosine distance 2),
	// so the wide query must return it clamped to 0.0 — and it must still be
	// present, ranked last by distance.
	for _, id := range []uuid.UUID{source1ID, source2ID} {
		similar := wide[id]
		if len(similar) == 0 {
			continue
		}
		last := similar[len(similar)-1]
		if last.NoteID != targetFarID {
			t.Errorf("source %v: expected far note %v ranked last, got %v", id, targetFarID, last.NoteID)
		}
		if last.Score != 0.0 {
			t.Errorf("source %v: expected far score 0.0 after clamping, got %v", id, last.Score)
		}
	}
}

// TestFindSimilarNotesBatch_MatchesSinglePath is an equivalence oracle: the
// batch query must pick the same neighbours as the correct single-note query
// (ORDER BY score DESC). UUID order is deliberately opposite to similarity
// order — the nearest target has the largest UUID — so a batch path that
// sorts by note_id instead of distance fails immediately.
func TestFindSimilarNotesBatch_MatchesSinglePath(t *testing.T) {
	db, cleanup := testutil.SetupTestVectorDB(t)
	defer cleanup()

	db.Exec("CREATE EXTENSION IF NOT EXISTS vector")
	if err := db.AutoMigrate(&UserModel{}, &NoteModel{}, &NoteEmbeddingModel{}); err != nil {
		t.Fatalf("failed to migrate models: %v", err)
	}

	repo := NewEmbeddingRepository(db, "paraphrase-multilingual-MiniLM-L12-v2")
	noteRepo := NewNoteRepository(db, nil)
	ctx := context.Background()

	sourceID := uuid.MustParse("a0000000-0000-0000-0000-000000000001")
	// UUID order: far < mid < near — the inverse of similarity order.
	farID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	midID := uuid.MustParse("20000000-0000-0000-0000-000000000001")
	nearID := uuid.MustParse("f0000000-0000-0000-0000-000000000001")

	// Unit-ish vectors along different axes give a strict distance order:
	// near ≈ source, mid orthogonal-ish, far opposite.
	vecAt := func(idx int, val float32) pgvector.Vector {
		v := make([]float32, 384)
		v[idx] = val
		return pgvector.NewVector(v)
	}
	sourceVec := vecAt(0, 1.0)
	nearVecRaw := make([]float32, 384)
	nearVecRaw[0] = 0.9
	nearVecRaw[1] = 0.1
	nearVec := pgvector.NewVector(nearVecRaw)
	midVecRaw := make([]float32, 384)
	midVecRaw[0] = 0.5
	midVecRaw[1] = 0.5
	midVec := pgvector.NewVector(midVecRaw)
	farVec := vecAt(0, -1.0)

	createNote := func(title string, id uuid.UUID) {
		titleV, _ := note.NewTitle(title)
		content, _ := note.NewContent("content")
		metadata, _ := note.NewMetadata(nil)
		n := note.NewNote(titleV, content, note.MustType("star"), metadata, note.WithID(id))
		if err := noteRepo.Save(ctx, n); err != nil {
			t.Fatalf("Save note failed: %v", err)
		}
	}
	createNote("Source", sourceID)
	createNote("Far (smallest UUID)", farID)
	createNote("Mid", midID)
	createNote("Near (largest UUID)", nearID)

	for id, v := range map[uuid.UUID]pgvector.Vector{
		sourceID: sourceVec,
		farID:    farVec,
		midID:    midVec,
		nearID:   nearVec,
	} {
		if err := repo.Upsert(ctx, id, v); err != nil {
			t.Fatalf("Upsert %v failed: %v", id, err)
		}
	}

	const limit = 2
	single, err := repo.FindSimilarNotes(ctx, sourceID, limit)
	if err != nil {
		t.Fatalf("FindSimilarNotes failed: %v", err)
	}
	batch, err := repo.FindSimilarNotesBatch(ctx, []uuid.UUID{sourceID}, limit)
	if err != nil {
		t.Fatalf("FindSimilarNotesBatch failed: %v", err)
	}

	got := batch[sourceID]
	if len(got) != len(single) {
		t.Fatalf("batch returned %d neighbours, single path %d", len(got), len(single))
	}
	for i := range single {
		if got[i].NoteID != single[i].NoteID {
			t.Errorf("neighbour %d: batch=%v single=%v", i, got[i].NoteID, single[i].NoteID)
		}
	}

	// And the concrete expectation: nearest (largest UUID) first, mid second,
	// far (smallest UUID) cut off by the limit.
	if len(got) != 2 || got[0].NoteID != nearID || got[1].NoteID != midID {
		t.Errorf("expected [near %v, mid %v], got %v", nearID, midID, got)
	}
}
