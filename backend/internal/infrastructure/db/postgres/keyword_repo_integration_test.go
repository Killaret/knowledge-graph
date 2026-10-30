//go:build integration

package postgres

import (
	"context"
	"testing"

	"knowledge-graph/internal/testutil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NLP-2: выборка keyword-recompute — заметки с непустым текстом и без строки
// note_keywords от текущего экстрактора. Сценарий из спеки: одна заметка со
// строкой yake-0.4.8, одна с keybert-*, одна без строк — выбираются ровно две;
// после обработки повторный запуск выбирает ноль.
func TestKeywordRepository_FindNoteIDsMissingExtractor_Integration(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	require.NoError(t, db.AutoMigrate(&UserModel{}, &NoteModel{}, &NoteKeywordModel{}))

	repo := NewKeywordRepository(db)

	oldID := uuid.New()   // строки от старого экстрактора
	freshID := uuid.New() // строки от текущего экстрактора
	bareID := uuid.New()  // без строк вообще
	emptyID := uuid.New() // пустой контент — не должен выбираться

	for _, n := range []NoteModel{
		{ID: oldID, Title: "old", Content: "has keywords from yake"},
		{ID: freshID, Title: "fresh", Content: "has keybert keywords"},
		{ID: bareID, Title: "bare", Content: "no keyword rows"},
		{ID: emptyID, Title: "empty", Content: "   "},
	} {
		require.NoError(t, db.Create(&n).Error)
	}
	require.NoError(t, db.Create(&NoteKeywordModel{
		NoteID: oldID, Keyword: "деревья", Surface: "деревья",
		Extractor: "yake-0.4.8", Weight: 0.5,
	}).Error)
	require.NoError(t, db.Create(&NoteKeywordModel{
		NoteID: freshID, Keyword: "дерево", Surface: "деревья",
		Extractor: "keybert-hybrid-0.9", Weight: 0.8,
	}).Error)

	ids, err := repo.FindNoteIDsMissingExtractor(ctx, "keybert-hybrid-0.9")
	require.NoError(t, err)
	assert.ElementsMatch(t, []uuid.UUID{oldID, bareID}, ids)

	// «Обработка»: старые строки заменены строками текущего экстрактора.
	require.NoError(t, repo.SaveAll(ctx, oldID, []NoteKeywordModel{
		{NoteID: oldID, Keyword: "дерево", Surface: "деревья",
			Extractor: "keybert-hybrid-0.9", Weight: 0.8},
	}))
	require.NoError(t, repo.SaveAll(ctx, bareID, []NoteKeywordModel{
		{NoteID: bareID, Keyword: "кот", Surface: "кот",
			Extractor: "keybert-hybrid-0.9", Weight: 0.9},
	}))

	ids, err = repo.FindNoteIDsMissingExtractor(ctx, "keybert-hybrid-0.9")
	require.NoError(t, err)
	assert.Empty(t, ids)

	// Другой экстрактор — выборка снова находит всех с текстом.
	ids, err = repo.FindNoteIDsMissingExtractor(ctx, "future-extractor-2.0")
	require.NoError(t, err)
	assert.ElementsMatch(t, []uuid.UUID{oldID, freshID, bareID}, ids)
}
