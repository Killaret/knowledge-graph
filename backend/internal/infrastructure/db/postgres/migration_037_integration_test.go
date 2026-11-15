//go:build integration

package postgres

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/testutil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// Migration037TestSuite — LINK-TYPES-1: migration merges `reference`/`custom`
// into `related`, keeps the pair's maximum weight, soft-deletes absorbed rows.
type Migration037TestSuite struct {
	suite.Suite
	db       *gorm.DB
	noteRepo *NoteRepository
	cleanup  func()
	ctx      context.Context
}

func (s *Migration037TestSuite) SetupSuite() {
	s.db, s.cleanup = testutil.SetupTestDB(s.T())
	s.ctx = context.Background()

	models := []interface{}{
		&NoteModel{},
		&LinkModel{},
		&LinkSuppressionModel{},
		&NoteKeywordModel{},
		&UserModel{},
		&TagModel{},
		&NoteTagModel{},
	}
	s.Require().NoError(s.db.AutoMigrate(models...))
	s.noteRepo = NewNoteRepository(s.db, nil)
}

func (s *Migration037TestSuite) TearDownSuite() {
	s.cleanup()
}

func (s *Migration037TestSuite) SetupTest() {
	s.Require().NoError(testutil.TruncateTables(s.db))
}

func (s *Migration037TestSuite) makeNote(title string) uuid.UUID {
	t, _ := note.NewTitle(title)
	c, _ := note.NewContent("content")
	m, _ := note.NewMetadata(map[string]interface{}{})
	n := note.NewNote(t, c, note.MustType("star"), m)
	s.Require().NoError(s.noteRepo.Save(s.ctx, n))
	return n.ID()
}

// insertLink bypasses the repository so legacy types can be seeded directly.
func (s *Migration037TestSuite) insertLink(src, dst uuid.UUID, linkType string, weight float64, deleted bool) uuid.UUID {
	return s.insertLinkFull(src, dst, linkType, "user", weight, deleted, uuid.Nil)
}

// insertLinkFull seeds a link with an explicit source_type and an optional
// deleted_via_note_id marker (the path note restore uses to revive links).
func (s *Migration037TestSuite) insertLinkFull(src, dst uuid.UUID, linkType, sourceType string, weight float64, deleted bool, via uuid.UUID) uuid.UUID {
	id := uuid.New()
	var deletedAt *time.Time
	var viaID *uuid.UUID
	if deleted {
		now := time.Now()
		deletedAt = &now
	}
	if via != uuid.Nil {
		viaID = &via
	}
	s.Require().NoError(s.db.Exec(
		`INSERT INTO links (id, source_note_id, target_note_id, link_type, weight, source_type, created_at, updated_at, deleted_at, deleted_via_note_id)
		 VALUES (?, ?, ?, ?, ?, ?, now(), now(), ?, ?)`,
		id, src, dst, linkType, weight, sourceType, deletedAt, viaID,
	).Error)
	return id
}

type linkRow struct {
	ID         uuid.UUID
	LinkType   string
	Weight     float64
	DeletedAt  *time.Time
	SourceType string
	Metadata   []byte
}

func (s *Migration037TestSuite) pairLinks(src, dst uuid.UUID) []linkRow {
	var rows []linkRow
	s.Require().NoError(s.db.Raw(
		`SELECT id, link_type, weight, deleted_at, source_type, metadata FROM links
		 WHERE source_note_id = ? AND target_note_id = ? ORDER BY link_type`,
		src, dst,
	).Scan(&rows).Error)
	return rows
}

// pairLinksUnordered returns every row of the pair in either direction.
func (s *Migration037TestSuite) pairLinksUnordered(a, b uuid.UUID) []linkRow {
	var rows []linkRow
	s.Require().NoError(s.db.Raw(
		`SELECT id, link_type, weight, deleted_at, source_type, metadata FROM links
		 WHERE (source_note_id = ? AND target_note_id = ?)
		    OR (source_note_id = ? AND target_note_id = ?)
		 ORDER BY source_note_id, link_type`,
		a, b, b, a,
	).Scan(&rows).Error)
	return rows
}

func liveLinks(rows []linkRow) []linkRow {
	var out []linkRow
	for _, r := range rows {
		if r.DeletedAt == nil {
			out = append(out, r)
		}
	}
	return out
}

func (s *Migration037TestSuite) runMigrationFile(name string) {
	sqlBytes, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "migrations", name))
	s.Require().NoError(err)
	s.Require().NoError(s.db.Exec(string(sqlBytes)).Error)
}

// Pair with live reference + custom + related collapses to one related row
// carrying the group's maximum weight; absorbed rows are soft-deleted.
func (s *Migration037TestSuite) TestMergeIntoExistingRelated() {
	a := s.makeNote("A1")
	b := s.makeNote("B1")

	s.insertLink(a, b, "reference", 0.4, false)
	s.insertLink(a, b, "custom", 0.9, false)
	s.insertLink(a, b, "related", 0.5, false)

	s.runMigrationFile("037_link_types_merge_related.up.sql")

	rows := s.pairLinks(a, b)
	s.Require().Len(rows, 3, "rows stay; absorbed ones are soft-deleted")

	var live []linkRow
	for _, r := range rows {
		if r.DeletedAt == nil {
			live = append(live, r)
		} else {
			s.NotEqual("related", r.LinkType, "absorbed rows keep their legacy type")
		}
	}
	s.Require().Len(live, 1)
	s.Equal("related", live[0].LinkType)
	s.InDelta(0.9, live[0].Weight, 0.0001, "survivor keeps the pair's maximum weight")
}

// Pair without a related row: heaviest legacy row converts, rest soft-delete.
func (s *Migration037TestSuite) TestMergeWithoutRelated() {
	a := s.makeNote("A2")
	b := s.makeNote("B2")

	s.insertLink(a, b, "reference", 0.3, false)
	s.insertLink(a, b, "custom", 0.8, false)

	s.runMigrationFile("037_link_types_merge_related.up.sql")

	rows := s.pairLinks(a, b)
	s.Require().Len(rows, 2)

	var live []linkRow
	for _, r := range rows {
		if r.DeletedAt == nil {
			live = append(live, r)
		}
	}
	s.Require().Len(live, 1)
	s.Equal("related", live[0].LinkType)
	s.InDelta(0.8, live[0].Weight, 0.0001)
}

// A soft-deleted related row is revived instead of converting a legacy row,
// so the pair keeps a single live link.
func (s *Migration037TestSuite) TestRevivesSoftDeletedRelated() {
	a := s.makeNote("A3")
	b := s.makeNote("B3")

	s.insertLink(a, b, "reference", 0.6, false)
	s.insertLink(a, b, "related", 0.2, true)

	s.runMigrationFile("037_link_types_merge_related.up.sql")

	rows := s.pairLinks(a, b)
	s.Require().Len(rows, 2)

	var live []linkRow
	for _, r := range rows {
		if r.DeletedAt == nil {
			live = append(live, r)
		}
	}
	s.Require().Len(live, 1)
	s.Equal("related", live[0].LinkType)
	s.InDelta(0.6, live[0].Weight, 0.0001, "revived row takes the group maximum")
}

// dependency/parent/child rows are untouched by the merge.
func (s *Migration037TestSuite) TestNonGenericTypesUntouched() {
	a := s.makeNote("A4")
	b := s.makeNote("B4")

	s.insertLink(a, b, "dependency", 0.7, false)
	s.insertLink(a, b, "reference", 0.5, false)

	s.runMigrationFile("037_link_types_merge_related.up.sql")

	rows := s.pairLinks(a, b)
	s.Require().Len(rows, 2)

	byType := map[string]linkRow{}
	for _, r := range rows {
		byType[r.LinkType] = r
	}
	s.Require().Contains(byType, "dependency")
	s.Nil(byType["dependency"].DeletedAt)
	s.InDelta(0.7, byType["dependency"].Weight, 0.0001)
	s.Require().Contains(byType, "related")
	s.Nil(byType["related"].DeletedAt)
}

// Column default becomes `related` after the migration.
func (s *Migration037TestSuite) TestColumnDefaultIsRelated() {
	s.runMigrationFile("037_link_types_merge_related.up.sql")

	var def string
	s.Require().NoError(s.db.Raw(
		`SELECT column_default FROM information_schema.columns
		 WHERE table_name = 'links' AND column_name = 'link_type'`,
	).Scan(&def).Error)
	s.Contains(def, "related")
}

// Down migration revives absorbed rows with their original types.
func (s *Migration037TestSuite) TestDownMigrationRestoresTypes() {
	a := s.makeNote("A6")
	b := s.makeNote("B6")

	s.insertLink(a, b, "reference", 0.4, false)
	s.insertLink(a, b, "custom", 0.9, false)

	s.runMigrationFile("037_link_types_merge_related.up.sql")
	s.runMigrationFile("037_link_types_merge_related.down.sql")

	rows := s.pairLinks(a, b)
	var live []linkRow
	for _, r := range rows {
		if r.DeletedAt == nil {
			live = append(live, r)
		}
	}
	s.Require().Len(live, 2, "both original rows restored with their legacy types")
	types := map[string]bool{}
	for _, r := range live {
		types[r.LinkType] = true
	}
	s.True(types["reference"] && types["custom"], "down restores the original pair")
}

// Case A (review): a live manual `reference` must win over a live gamma
// `related` — the user's link stays a user link, the model weight moves to
// metadata.gamma (SaveUserLink promotion semantics).
func (s *Migration037TestSuite) TestManualReferenceBeatsGammaRelated() {
	a := s.makeNote("A7")
	b := s.makeNote("B7")

	s.insertLinkFull(a, b, "related", "gamma", 0.62, false, uuid.Nil)
	s.insertLink(a, b, "reference", 0.9, false)

	s.runMigrationFile("037_link_types_merge_related.up.sql")

	rows := s.pairLinks(a, b)
	live := liveLinks(rows)
	s.Require().Len(live, 1)
	s.Equal("related", live[0].LinkType)
	s.Equal("user", live[0].SourceType, "the manual link wins — a user link must not become a model one")
	s.InDelta(0.9, live[0].Weight, 0.0001)
	s.Contains(string(live[0].Metadata), `"gamma"`, "model weight preserved in metadata.gamma")
	s.Contains(string(live[0].Metadata), "0.62")
}

// Case B (review): a `dependency` row on the same pair is not part of the
// merge and must not donate its weight.
func (s *Migration037TestSuite) TestDependencyWeightDoesNotContaminate() {
	a := s.makeNote("A8")
	b := s.makeNote("B8")

	s.insertLink(a, b, "related", 0.3, false)
	s.insertLink(a, b, "reference", 0.5, false)
	s.insertLink(a, b, "dependency", 0.95, false)

	s.runMigrationFile("037_link_types_merge_related.up.sql")

	rows := s.pairLinks(a, b)
	byType := map[string]linkRow{}
	for _, r := range rows {
		byType[r.LinkType] = r
	}
	s.Require().Contains(byType, "dependency")
	s.Nil(byType["dependency"].DeletedAt)
	s.InDelta(0.95, byType["dependency"].Weight, 0.0001, "dependency untouched")

	live := liveLinks(rows)
	var generic []linkRow
	for _, r := range live {
		if r.LinkType == "related" {
			generic = append(generic, r)
		}
	}
	s.Require().Len(generic, 1)
	s.InDelta(0.5, generic[0].Weight, 0.0001, "max over the generic group only — never the dependency weight")
}

// Case C (review): a deleted non-generic row (`parent`, deleted via a note)
// must not donate its weight either.
func (s *Migration037TestSuite) TestDeletedParentWeightDoesNotContaminate() {
	a := s.makeNote("A9")
	b := s.makeNote("B9")

	s.insertLink(a, b, "related", 0.2, false)
	s.insertLink(a, b, "custom", 0.4, false)
	s.insertLinkFull(a, b, "parent", "user", 1.0, true, a)

	s.runMigrationFile("037_link_types_merge_related.up.sql")

	rows := s.pairLinks(a, b)
	byType := map[string]linkRow{}
	for _, r := range rows {
		byType[r.LinkType] = r
	}
	s.Require().Contains(byType, "parent")
	s.InDelta(1.0, byType["parent"].Weight, 0.0001, "parent untouched")

	live := liveLinks(rows)
	var generic []linkRow
	for _, r := range live {
		if r.LinkType == "related" {
			generic = append(generic, r)
		}
	}
	s.Require().Len(generic, 1)
	s.InDelta(0.4, generic[0].Weight, 0.0001, "max over the generic live rows only")
}

// Case D (review): a rejected (soft-deleted) gamma `related` must stay
// deleted; the live manual `reference` survives as the pair's link.
func (s *Migration037TestSuite) TestRejectedGammaStaysDeleted() {
	a := s.makeNote("A10")
	b := s.makeNote("B10")

	rejected := s.insertLinkFull(a, b, "related", "gamma", 0.9, true, uuid.Nil)
	s.insertLink(a, b, "reference", 0.5, false)

	s.runMigrationFile("037_link_types_merge_related.up.sql")

	rows := s.pairLinks(a, b)
	live := liveLinks(rows)
	s.Require().Len(live, 1)
	s.Equal("user", live[0].SourceType)
	s.Equal("related", live[0].LinkType)
	s.InDelta(0.5, live[0].Weight, 0.0001)

	for _, r := range rows {
		if r.ID == rejected {
			s.NotNil(r.DeletedAt, "rejected gamma link must not be revived")
		}
	}
}

// Case E (review): a legacy row deleted via a deleted note must not come
// back as a second live legacy link when the note is restored. The
// migration detaches deleted_via_note_id, so Restore revives nothing extra.
func (s *Migration037TestSuite) TestNoteDeletedRowStaysDetachedOnRestore() {
	a := s.makeNote("A11")
	b := s.makeNote("B11")

	s.insertLink(a, b, "related", 0.5, false)
	s.insertLinkFull(a, b, "reference", "user", 0.7, true, b)

	s.runMigrationFile("037_link_types_merge_related.up.sql")

	// Simulate note restore: revive everything marked deleted_via_note_id = b.
	s.Require().NoError(s.db.Exec(
		`UPDATE links SET deleted_at = NULL, deleted_via_note_id = NULL
		 WHERE deleted_via_note_id = ?`, b,
	).Error)

	live := liveLinks(s.pairLinks(a, b))
	s.Require().Len(live, 1, "restore must not resurrect the absorbed legacy row")
	s.Equal("related", live[0].LinkType)
	s.Equal("user", live[0].SourceType)
}

// Case F (review): related A→B + reference B→A are the same pair — exactly
// one live merged row survives, keeping the manual link's direction.
func (s *Migration037TestSuite) TestReverseDirectionPairCollapses() {
	a := s.makeNote("A12")
	b := s.makeNote("B12")

	s.insertLink(a, b, "related", 0.3, false)
	s.insertLink(b, a, "reference", 0.8, false)

	s.runMigrationFile("037_link_types_merge_related.up.sql")

	live := liveLinks(s.pairLinksUnordered(a, b))
	s.Require().Len(live, 1, "one edge per unordered pair regardless of direction")
	s.Equal("related", live[0].LinkType)
	s.InDelta(0.8, live[0].Weight, 0.0001)

	var src, dst uuid.UUID
	s.Require().NoError(s.db.Raw(
		`SELECT source_note_id, target_note_id FROM links WHERE id = ?`, live[0].ID,
	).Row().Scan(&src, &dst))
	s.Equal(b, src, "survivor keeps the manual reference's direction B→A")
	s.Equal(a, dst)
}

// Case A′ (tail): the gamma `related` is HEAVIER than the manual
// `reference`. Per SaveUserLink the promoted user row keeps the USER's
// weight; the model's weight moves to metadata.gamma. Mutation "drop
// (source_type='user') from the survivor ordering" turns this red — the
// gamma row wins and the pair becomes a model link.
func (s *Migration037TestSuite) TestA2UserWeightSurvivesOverHeavierGamma() {
	a := s.makeNote("A13")
	b := s.makeNote("B13")

	s.insertLinkFull(a, b, "related", "gamma", 0.95, false, uuid.Nil)
	s.insertLink(a, b, "reference", 0.5, false)

	s.runMigrationFile("037_link_types_merge_related.up.sql")

	live := liveLinks(s.pairLinks(a, b))
	s.Require().Len(live, 1)
	s.Equal("related", live[0].LinkType)
	s.Equal("user", live[0].SourceType, "the user's link wins regardless of weight")
	s.InDelta(0.5, live[0].Weight, 0.0001, "SaveUserLink promotion keeps the user's weight, not the model's")
	s.Contains(string(live[0].Metadata), `"gamma"`)
	s.Contains(string(live[0].Metadata), "0.95", "the model's weight survives in metadata.gamma")
}

// Case D′ (tail): a deleted merged row heavier than the live one must not
// resurrect. The pair needs a live legacy row to enter the merge loop, so
// both rows are `reference`. Mutation "drop (deleted_at IS NULL) from the
// survivor ordering" turns this red — the dead row wins and is revived.
func (s *Migration037TestSuite) TestD2LiveRowBeatsHeavierDeleted() {
	a := s.makeNote("A14")
	b := s.makeNote("B14")

	s.insertLink(a, b, "reference", 0.3, false)
	dead := s.insertLink(b, a, "reference", 0.9, true)

	s.runMigrationFile("037_link_types_merge_related.up.sql")

	rows := s.pairLinksUnordered(a, b)
	live := liveLinks(rows)
	s.Require().Len(live, 1)
	s.Equal("related", live[0].LinkType)
	s.InDelta(0.3, live[0].Weight, 0.0001, "the live row survives; the heavier deleted one stays dead")
	for _, r := range rows {
		if r.ID == dead {
			s.NotNil(r.DeletedAt, "the heavier deleted row must not be revived")
		}
	}
}

// Case E′ (tail): the pair's only link is a `reference` deleted together
// with a note. The migration converts it to `related` and keeps
// deleted_via_note_id, so restoring the note brings the link back — a
// detached marker would leave the pair linkless forever.
func (s *Migration037TestSuite) TestE2ViaNoteDeletedLegacyReturnsOnRestore() {
	a := s.makeNote("A15")
	b := s.makeNote("B15")

	s.insertLinkFull(a, b, "reference", "user", 0.7, true, b)

	s.runMigrationFile("037_link_types_merge_related.up.sql")

	var linkType string
	var via *uuid.UUID
	var deletedAt *time.Time
	s.Require().NoError(s.db.Raw(
		`SELECT link_type, deleted_via_note_id, deleted_at FROM links
		 WHERE source_note_id = ? AND target_note_id = ?`, a, b,
	).Row().Scan(&linkType, &via, &deletedAt))
	s.Equal("related", linkType, "the note-deleted legacy row converts to related")
	s.NotNil(deletedAt, "the row stays in the trash until the note is restored")
	s.Require().NotNil(via, "the note-restore marker survives the migration")
	s.Equal(b, *via)

	// Simulate note restore.
	s.Require().NoError(s.db.Exec(
		`UPDATE links SET deleted_at = NULL, deleted_via_note_id = NULL
		 WHERE deleted_via_note_id = ?`, b,
	).Error)

	live := liveLinks(s.pairLinks(a, b))
	s.Require().Len(live, 1, "restore brings the pair's link back")
	s.Equal("related", live[0].LinkType)
	s.InDelta(0.7, live[0].Weight, 0.0001)
}

// The collision side of E′: a via-note-deleted `reference` on a pair that
// already has a `related` row detaches instead of retyping — a restore must
// not produce a second edge.
func (s *Migration037TestSuite) TestE2CollisionDetachesInsteadOfDuplicating() {
	a := s.makeNote("A16")
	b := s.makeNote("B16")

	s.insertLinkFull(a, b, "related", "user", 0.6, true, b)
	s.insertLinkFull(b, a, "reference", "user", 0.7, true, b)

	s.runMigrationFile("037_link_types_merge_related.up.sql")

	// Simulate note restore.
	s.Require().NoError(s.db.Exec(
		`UPDATE links SET deleted_at = NULL, deleted_via_note_id = NULL
		 WHERE deleted_via_note_id = ?`, b,
	).Error)

	live := liveLinks(s.pairLinksUnordered(a, b))
	s.Require().Len(live, 1, "restore brings back the pair's single related edge, not a doubled pair")
	s.Equal("related", live[0].LinkType)
}

// E′ adversarial: same-direction `reference` + `custom`, both deleted via
// the same note, nothing live. Retyping both would duplicate the directed
// (src,dst,'related') key and abort the migration — only the heaviest
// converts; the loser detaches.
func (s *Migration037TestSuite) TestE2SameDirectionDeletedLegacyNoKeyClash() {
	a := s.makeNote("A17")
	b := s.makeNote("B17")

	s.insertLinkFull(a, b, "reference", "user", 0.8, true, b)
	s.insertLinkFull(a, b, "custom", "user", 0.4, true, b)

	s.runMigrationFile("037_link_types_merge_related.up.sql")

	var relatedCount int
	s.Require().NoError(s.db.Raw(
		`SELECT count(*) FROM links
		 WHERE source_note_id = ? AND target_note_id = ? AND link_type = 'related'`, a, b,
	).Row().Scan(&relatedCount))
	s.Equal(1, relatedCount, "exactly one directed related — no UNIQUE clash")

	var convertedVia *uuid.UUID
	var convertedWeight float64
	s.Require().NoError(s.db.Raw(
		`SELECT deleted_via_note_id, weight FROM links
		 WHERE source_note_id = ? AND target_note_id = ? AND link_type = 'related'`, a, b,
	).Row().Scan(&convertedVia, &convertedWeight))
	s.Require().NotNil(convertedVia)
	s.Equal(b, *convertedVia)
	s.InDelta(0.8, convertedWeight, 0.0001, "the heaviest candidate converts")
}

func TestMigration037Suite(t *testing.T) {
	suite.Run(t, new(Migration037TestSuite))
}
