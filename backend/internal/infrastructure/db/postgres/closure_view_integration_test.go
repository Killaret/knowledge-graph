//go:build integration
// +build integration

package postgres

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// LINKS-1 regression: the 027 closure view used array_agg over uuid[] paths,
// which fails with "cannot accumulate arrays of different dimensionality" as
// soon as a note pair has paths of different lengths. The 032 migration fixed
// the aggregate; 033 bound the recursion to the consumers' max depth (5) with
// UNION-deduped levels and dropped the stored path column — the path between
// two notes is computed on demand by the graph service. Tests apply the
// migrations in the same order production does.
func TestClosureView_RefreshesWithDivergentPathLengths(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	if err := db.AutoMigrate(&UserModel{}, &NoteModel{}, &LinkModel{}); err != nil {
		t.Fatalf("failed to migrate models: %v", err)
	}

	for _, name := range []string{
		"032_fix_note_links_closure_path.up.sql",
		"033_bounded_note_links_closure.up.sql",
	} {
		upSQL, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "migrations", name))
		require.NoError(t, err)
		require.NoError(t, db.Exec(string(upSQL)).Error)
	}

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
		Weight       float64
	}
	require.NoError(t, db.Raw(
		"SELECT ancestor_id::text, descendant_id::text, distance, weight FROM note_links_closure WHERE ancestor_id = ?",
		notes[0].ID()).Scan(&rows).Error)

	byTarget := map[string]int{}
	weights := map[string]float64{}
	for _, r := range rows {
		byTarget[r.DescendantID] = r.Distance
		weights[r.DescendantID] = r.Weight
	}
	assert.Equal(t, 1, byTarget[notes[1].ID().String()])
	assert.Equal(t, 1, byTarget[notes[2].ID().String()], "min distance for A->C")
	assert.Equal(t, 0.9, weights[notes[2].ID().String()], "weight of the shortest A->C path")
}

// LINKS-1 rework boundary: on a dense graph (1 000 notes, ~3 outgoing links
// each — the shape gamma links create) REFRESH must stay in seconds. The
// pre-033 CTE enumerated every simple path up to depth 10 and needed ~9
// minutes on a far smaller seed, so reverting the migration turns this red.
// The bound is 20 s, not tighter: a testcontainer on a loaded Windows host
// measured ~7–9 s; the pre-033 CTE is two orders of magnitude slower, so the
// margin still catches the regression without flaking on slow hardware.
func TestClosureView_RefreshBoundedOnDenseGraph(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	if err := db.AutoMigrate(&UserModel{}, &NoteModel{}, &LinkModel{}); err != nil {
		t.Fatalf("failed to migrate models: %v", err)
	}

	upSQL, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "migrations", "033_bounded_note_links_closure.up.sql"))
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(upSQL)).Error)

	require.NoError(t, db.Exec(
		"INSERT INTO notes (id, title, type, is_public) "+
			"SELECT gen_random_uuid(), 'note-' || g, 'star', true FROM generate_series(1, 1000) g",
	).Error)

	// Deterministic degree-3 ring-ish graph: i -> i+1, i->i+7, i->i+31.
	require.NoError(t, db.Exec(
		`INSERT INTO links (id, source_note_id, target_note_id, link_type, weight, source_type)
		 SELECT gen_random_uuid(), s.id, t.id, 'related', 0.8, 'user'
		 FROM (
		   SELECT id, (row_number() OVER (ORDER BY id))::int - 1 AS rn FROM notes
		 ) s
		 JOIN LATERAL (
		   SELECT DISTINCT t.id
		   FROM (
		     SELECT id, (row_number() OVER (ORDER BY id))::int - 1 AS rn FROM notes
		   ) t
		   WHERE t.rn IN ((s.rn + 1) % 1000, (s.rn + 7) % 1000, (s.rn + 31) % 1000)
		     AND t.id <> s.id
		 ) t ON true
		 ON CONFLICT DO NOTHING`,
	).Error)

	var linkCount int64
	require.NoError(t, db.Raw("SELECT count(*) FROM links").Scan(&linkCount).Error)
	require.Greater(t, linkCount, int64(2500), "seed should give ~3000 links")

	start := time.Now()
	require.NoError(t, db.Exec("REFRESH MATERIALIZED VIEW CONCURRENTLY note_links_closure").Error)
	elapsed := time.Since(start)
	t.Logf("REFRESH note_links_closure on 1000 nodes / %d links took %v", linkCount, elapsed)
	assert.Less(t, elapsed, 20*time.Second, "closure refresh must stay in seconds on a dense graph")

	var pairs int64
	require.NoError(t, db.Raw("SELECT count(*) FROM note_links_closure").Scan(&pairs).Error)
	assert.Greater(t, pairs, int64(1000), "closure must actually contain rows")
	var maxDist int
	require.NoError(t, db.Raw("SELECT COALESCE(MAX(distance),0) FROM note_links_closure").Scan(&maxDist).Error)
	assert.LessOrEqual(t, maxDist, 5, "closure must not exceed the consumers' max depth")
}
