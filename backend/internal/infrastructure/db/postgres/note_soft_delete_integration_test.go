//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/testutil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// NOTE-DELETE-1 integration tests against the real PostgreSQL schema:
// soft delete, restore with links, and the 90-day purge horizon.

type softDeleteFixture struct {
	db   *gorm.DB
	repo *NoteRepository
	ctx  context.Context
}

func setupSoftDelete(t *testing.T) (*softDeleteFixture, func()) {
	t.Helper()
	db, cleanup := testutil.SetupTestDB(t)
	require.NoError(t, db.AutoMigrate(&UserModel{}, &NoteModel{}, &LinkModel{}, &NoteKeywordModel{}, &TagModel{}, &NoteTagModel{}))
	require.NoError(t, testutil.TruncateTables(db))
	return &softDeleteFixture{db: db, repo: NewNoteRepository(db, nil), ctx: context.Background()}, cleanup
}

func (f *softDeleteFixture) createNote(t *testing.T, title string) uuid.UUID {
	t.Helper()
	m := &NoteModel{Title: title, Content: "content"}
	require.NoError(t, f.db.Create(m).Error)
	return m.ID
}

func (f *softDeleteFixture) createLink(t *testing.T, source, target uuid.UUID, weight float64) uuid.UUID {
	t.Helper()
	m := &LinkModel{ID: uuid.New(), SourceNoteID: source, TargetNoteID: target, LinkType: "reference", Weight: weight, SourceType: "user"}
	require.NoError(t, f.db.Create(m).Error)
	return m.ID
}

func (f *softDeleteFixture) linkState(t *testing.T, id uuid.UUID) (deleted bool, via *uuid.UUID, weight float64) {
	t.Helper()
	var m LinkModel
	require.NoError(t, f.db.Unscoped().First(&m, "id = ?", id).Error)
	return m.DeletedAt != nil, m.DeletedViaNoteID, m.Weight
}

func TestNoteSoftDelete_DeleteKeepsRowAndMarksLinks(t *testing.T) {
	f, cleanup := setupSoftDelete(t)
	defer cleanup()

	a := f.createNote(t, "a")
	b := f.createNote(t, "b")
	l := f.createLink(t, a, b, 1.5)

	require.NoError(t, f.repo.Delete(f.ctx, a))

	// Row is still there — only deleted_at changed.
	var count int64
	require.NoError(t, f.db.Unscoped().Model(&NoteModel{}).Where("id = ?", a).Count(&count).Error)
	require.EqualValues(t, 1, count)

	// The note is invisible to the regular read paths.
	found, err := f.repo.FindByID(f.ctx, a)
	require.NoError(t, err)
	require.Nil(t, found)

	all, err := f.repo.FindAll(f.ctx)
	require.NoError(t, err)
	require.Len(t, all, 1)
	require.Equal(t, b, all[0].ID())

	// The link is soft-deleted and carries the marker.
	deleted, via, _ := f.linkState(t, l)
	require.True(t, deleted)
	require.NotNil(t, via)
	require.Equal(t, a, *via)
}

func TestNoteSoftDelete_RestoreBringsNoteAndLinksBack(t *testing.T) {
	f, cleanup := setupSoftDelete(t)
	defer cleanup()

	a := f.createNote(t, "a")
	b := f.createNote(t, "b")
	l := f.createLink(t, a, b, 1.5)

	require.NoError(t, f.repo.Delete(f.ctx, a))
	require.NoError(t, f.repo.Restore(f.ctx, a))

	found, err := f.repo.FindByID(f.ctx, a)
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, "a", found.Title().String())

	// The link is back byte-for-byte, marker cleared.
	deleted, via, weight := f.linkState(t, l)
	require.False(t, deleted)
	require.Nil(t, via)
	require.Equal(t, 1.5, weight)
}

func TestNoteSoftDelete_StandaloneDeletedLinkStaysDead(t *testing.T) {
	f, cleanup := setupSoftDelete(t)
	defer cleanup()

	a := f.createNote(t, "a")
	b := f.createNote(t, "b")
	l := f.createLink(t, a, b, 1)

	// Standalone link removal is a hard delete — like today.
	require.NoError(t, f.db.Exec("DELETE FROM links WHERE id = ?", l).Error)

	require.NoError(t, f.repo.Delete(f.ctx, a))
	require.NoError(t, f.repo.Restore(f.ctx, a))

	var count int64
	require.NoError(t, f.db.Unscoped().Model(&LinkModel{}).Where("id = ?", l).Count(&count).Error)
	require.EqualValues(t, 0, count, "a link removed on its own must not come back")
}

func TestNoteSoftDelete_ReviveWaitsForBothEndpoints(t *testing.T) {
	f, cleanup := setupSoftDelete(t)
	defer cleanup()

	a := f.createNote(t, "a")
	b := f.createNote(t, "b")
	l := f.createLink(t, a, b, 1)

	// Delete A (marks the link via A), then delete B — the marker stays A
	// because the link is already down.
	require.NoError(t, f.repo.Delete(f.ctx, a))
	require.NoError(t, f.repo.Delete(f.ctx, b))

	// Restoring A alone must not resurface a link whose other end is gone.
	require.NoError(t, f.repo.Restore(f.ctx, a))
	deleted, _, _ := f.linkState(t, l)
	require.True(t, deleted, "link must stay down while endpoint b is trashed")

	// Restoring B revives it — the last deleted endpoint is back.
	require.NoError(t, f.repo.Restore(f.ctx, b))
	deleted, via, _ := f.linkState(t, l)
	require.False(t, deleted)
	require.Nil(t, via)
}

func TestNoteSoftDelete_SearchExcludesTrashed(t *testing.T) {
	f, cleanup := setupSoftDelete(t)
	defer cleanup()

	// uuid.Nil scope reads public notes only — keep both public so the
	// search would see the deleted one if the filter leaked.
	a := f.createNote(t, "needle alpha")
	b := f.createNote(t, "needle beta")
	require.NoError(t, f.db.Model(&NoteModel{}).Where("id IN ?", []uuid.UUID{a, b}).Update("is_public", true).Error)

	hits, total, err := f.repo.Search(f.ctx, uuid.Nil, "needle", 10, 0)
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Len(t, hits, 2)

	require.NoError(t, f.repo.Delete(f.ctx, a))

	hits, total, err = f.repo.Search(f.ctx, uuid.Nil, "needle", 10, 0)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, hits, 1)
	require.Equal(t, b, hits[0].ID())
}

func TestNoteSoftDelete_LinkReadsHideTrashedLinks(t *testing.T) {
	f, cleanup := setupSoftDelete(t)
	defer cleanup()

	a := f.createNote(t, "a")
	b := f.createNote(t, "b")
	l := f.createLink(t, a, b, 1)

	linkRepo := NewLinkRepository(f.db)

	require.NoError(t, f.repo.Delete(f.ctx, a))

	// Every link read path must skip the trashed link.
	byID, err := linkRepo.FindByID(f.ctx, l)
	require.NoError(t, err)
	require.Nil(t, byID)

	bySource, err := linkRepo.FindBySource(f.ctx, a)
	require.NoError(t, err)
	require.Empty(t, bySource)

	bySourceIDs, err := linkRepo.FindBySourceIDs(f.ctx, []uuid.UUID{a})
	require.NoError(t, err)
	require.Empty(t, bySourceIDs[a])

	all, total, err := linkRepo.FindAllPaginated(f.ctx, 10, 0)
	require.NoError(t, err)
	require.Zero(t, total)
	require.Empty(t, all)

	// After restore the link is visible again on every path.
	require.NoError(t, f.repo.Restore(f.ctx, a))
	bySource, err = linkRepo.FindBySource(f.ctx, a)
	require.NoError(t, err)
	require.Len(t, bySource, 1)
}

func TestNoteSoftDelete_RestoreLiveNoteIsNotFound(t *testing.T) {
	f, cleanup := setupSoftDelete(t)
	defer cleanup()

	a := f.createNote(t, "a")
	require.ErrorIs(t, f.repo.Restore(f.ctx, a), note.ErrNoteNotFound)
	require.ErrorIs(t, f.repo.Restore(f.ctx, uuid.New()), note.ErrNoteNotFound)
}

func TestNoteSoftDelete_PurgeHorizon(t *testing.T) {
	f, cleanup := setupSoftDelete(t)
	defer cleanup()

	old := f.createNote(t, "old")
	fresh := f.createNote(t, "fresh")
	b := f.createNote(t, "b")
	oldLink := f.createLink(t, old, b, 1)
	liveLink := f.createLink(t, fresh, b, 1)

	require.NoError(t, f.repo.Delete(f.ctx, old))
	require.NoError(t, f.repo.Delete(f.ctx, fresh))

	// Backdate: old was trashed 91 days ago, fresh 89 days ago.
	require.NoError(t, f.db.Exec("UPDATE notes SET deleted_at = ? WHERE id = ?", time.Now().Add(-91*24*time.Hour), old).Error)
	require.NoError(t, f.db.Exec("UPDATE notes SET deleted_at = ? WHERE id = ?", time.Now().Add(-89*24*time.Hour), fresh).Error)

	purged, err := f.repo.PurgeDeletedBefore(f.ctx, time.Now().Add(-90*24*time.Hour))
	require.NoError(t, err)
	require.EqualValues(t, 1, purged)

	var oldCount, freshCount, oldLinkCount, liveLinkCount int64
	require.NoError(t, f.db.Unscoped().Model(&NoteModel{}).Where("id = ?", old).Count(&oldCount).Error)
	require.NoError(t, f.db.Unscoped().Model(&NoteModel{}).Where("id = ?", fresh).Count(&freshCount).Error)
	require.NoError(t, f.db.Unscoped().Model(&LinkModel{}).Where("id = ?", oldLink).Count(&oldLinkCount).Error)
	require.NoError(t, f.db.Unscoped().Model(&LinkModel{}).Where("id = ?", liveLink).Count(&liveLinkCount).Error)

	require.EqualValues(t, 0, oldCount, "91-day-old trash is gone")
	require.EqualValues(t, 0, oldLinkCount, "its links went through the FK cascade")
	require.EqualValues(t, 1, freshCount, "89-day-old trash still rests")
	require.EqualValues(t, 1, liveLinkCount)
}

// The restore route's access check runs before the handler can report 404 —
// it needs a lookup that still sees the trashed row.
func TestNoteSoftDelete_FindByIDIncludingDeletedSeesTrash(t *testing.T) {
	f, cleanup := setupSoftDelete(t)
	defer cleanup()

	a := f.createNote(t, "a")
	require.NoError(t, f.repo.Delete(f.ctx, a))

	live, err := f.repo.FindByID(f.ctx, a)
	require.NoError(t, err)
	require.Nil(t, live, "regular read must not see the trash")

	trashed, err := f.repo.FindByIDIncludingDeleted(f.ctx, a)
	require.NoError(t, err)
	require.NotNil(t, trashed, "including-deleted lookup must find the trashed row")
	require.Equal(t, a, trashed.ID())
}

// Recommendations were computed before the delete and note_recommendations
// has no trigger on notes.deleted_at — the read must filter the join itself.
func TestNoteSoftDelete_RecommendationsHideTrashedTarget(t *testing.T) {
	f, cleanup := setupSoftDelete(t)
	defer cleanup()
	require.NoError(t, f.db.AutoMigrate(&RecommendationModel{}))

	a := f.createNote(t, "a")
	live := f.createNote(t, "live candidate")
	trashed := f.createNote(t, "trashed candidate")

	recRepo := NewRecommendationRepository(f.db)
	require.NoError(t, recRepo.SaveBatch(f.ctx, a, map[uuid.UUID]float64{live: 0.9, trashed: 0.8}))

	require.NoError(t, f.repo.Delete(f.ctx, trashed))

	recs, err := recRepo.GetRecommendations(f.ctx, a, 10)
	require.NoError(t, err)
	require.Len(t, recs, 1)
	require.Equal(t, live, recs[0].RecommendedNoteID, "the trashed candidate must not be suggested")
}
