//go:build integration
// +build integration

package postgres

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// LINKS-1 regression: the 027 closure view used array_agg over uuid[] paths,
// which fails with "cannot accumulate arrays of different dimensionality" as
// soon as a note pair has paths of different lengths. The 032 migration
// recreates the view; this test applies that exact SQL and exercises the case.
func TestClosureView_RefreshesWithDivergentPathLengths(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	if err := db.AutoMigrate(&UserModel{}, &NoteModel{}, &LinkModel{}); err != nil {
		t.Fatalf("failed to migrate models: %v", err)
	}

	upSQL, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "migrations", "032_fix_note_links_closure_path.up.sql"))
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(upSQL)).Error)

	ctx := context.Background()
	noteRepo := NewNoteRepository(db, nil)
	linkRepo := NewLinkRepository(db)

	// Chain A -> B -> C plus the shortcut A -> C: the pair (A, C) has a
	// length-1 path and a length-2 path — the shape that broke array_agg.
	var notes []*note.Note
	for _, title := range []string{"A", "B", "C"} {
		tv, _ := note.NewTitle(title)
		cv, _ := note.NewContent("content " + title)
		md, _ := note.NewMetadata(nil)
		n := note.NewNote(tv, cv, note.MustType("star"), md)
		require.NoError(t, noteRepo.Save(ctx, n))
		notes = append(notes, n)
	}

	lt, _ := link.NewLinkType("related")
	w, _ := link.NewWeight(0.9)
	md, _ := link.NewMetadata(nil)
	require.NoError(t, linkRepo.Save(ctx, link.NewLink(notes[0].ID(), notes[1].ID(), lt, w, md)))
	require.NoError(t, linkRepo.Save(ctx, link.NewLink(notes[1].ID(), notes[2].ID(), lt, w, md)))
	require.NoError(t, linkRepo.Save(ctx, link.NewLink(notes[0].ID(), notes[2].ID(), lt, w, md)))

	require.NoError(t, db.Exec("REFRESH MATERIALIZED VIEW CONCURRENTLY note_links_closure").Error)

	var rows []struct {
		AncestorID   string
		DescendantID string
		Distance     int
		Path         string
	}
	require.NoError(t, db.Raw(
		"SELECT ancestor_id::text, descendant_id::text, distance, path::text FROM note_links_closure WHERE ancestor_id = ?",
		notes[0].ID()).Scan(&rows).Error)

	byTarget := map[string]int{}
	paths := map[string]string{}
	for _, r := range rows {
		byTarget[r.DescendantID] = r.Distance
		paths[r.DescendantID] = r.Path
	}
	assert.Equal(t, 1, byTarget[notes[1].ID().String()])
	assert.Equal(t, 1, byTarget[notes[2].ID().String()], "min distance for A->C")
	assert.Contains(t, paths[notes[2].ID().String()], notes[0].ID().String(), "path column must be a uuid[] literal")
}
