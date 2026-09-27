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
	id := uuid.New()
	var deletedAt *time.Time
	if deleted {
		now := time.Now()
		deletedAt = &now
	}
	s.Require().NoError(s.db.Exec(
		`INSERT INTO links (id, source_note_id, target_note_id, link_type, weight, source_type, created_at, updated_at, deleted_at)
		 VALUES (?, ?, ?, ?, ?, 'user', now(), now(), ?)`,
		id, src, dst, linkType, weight, deletedAt,
	).Error)
	return id
}

type linkRow struct {
	ID         uuid.UUID
	LinkType   string
	Weight     float64
	DeletedAt  *time.Time
	SourceType string
}

func (s *Migration037TestSuite) pairLinks(src, dst uuid.UUID) []linkRow {
	var rows []linkRow
	s.Require().NoError(s.db.Raw(
		`SELECT id, link_type, weight, deleted_at, source_type FROM links
		 WHERE source_note_id = ? AND target_note_id = ? ORDER BY link_type`,
		src, dst,
	).Scan(&rows).Error)
	return rows
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

func TestMigration037Suite(t *testing.T) {
	suite.Run(t, new(Migration037TestSuite))
}
