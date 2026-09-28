//go:build integration
// +build integration

package outbox

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/infrastructure/db/postgres"
	infevents "knowledge-graph/internal/infrastructure/events"
	"knowledge-graph/internal/testutil"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type fixture struct {
	db       *gorm.DB
	noteRepo *NoteRepository
	linkRepo *LinkRepository
	ctx      context.Context
}

func setup(t *testing.T) (*fixture, func()) {
	t.Helper()
	db, cleanup := testutil.SetupTestDB(t)
	require.NoError(t, db.AutoMigrate(
		&postgres.UserModel{}, &postgres.NoteModel{}, &postgres.LinkModel{},
		&postgres.LinkSuppressionModel{}, &Model{}))
	require.NoError(t, db.Exec("TRUNCATE TABLE notes, links, link_suppressions, users, graph_outbox RESTART IDENTITY CASCADE").Error)
	return &fixture{
		db:       db,
		noteRepo: NewNoteRepository(postgres.NewNoteRepository(db, nil), db),
		linkRepo: NewLinkRepository(postgres.NewLinkRepository(db), db),
		ctx:      context.Background(),
	}, cleanup
}

func (f *fixture) createUser(t *testing.T) uuid.UUID {
	t.Helper()
	m := &postgres.UserModel{Login: "u-" + uuid.NewString()[:8], PasswordHash: "x"}
	require.NoError(t, f.db.Create(m).Error)
	return m.ID
}

func (f *fixture) mustNote(t *testing.T, title string, creator *uuid.UUID) *note.Note {
	t.Helper()
	tv, err := note.NewTitle(title)
	require.NoError(t, err)
	cv, err := note.NewContent("body of " + title)
	require.NoError(t, err)
	md, err := note.NewMetadata(nil)
	require.NoError(t, err)
	if creator != nil {
		return note.NewNoteWithCreator(tv, cv, note.MustType("star"), md, *creator)
	}
	return note.NewNote(tv, cv, note.MustType("star"), md)
}

func mustLink(t *testing.T, source, target uuid.UUID, creator *uuid.UUID) *link.Link {
	t.Helper()
	lt, err := link.NewLinkType("related")
	require.NoError(t, err)
	w, err := link.NewWeight(1.0)
	require.NoError(t, err)
	md, err := link.NewMetadata(map[string]interface{}{})
	require.NoError(t, err)
	if creator != nil {
		return link.NewLinkWithCreator(source, target, *creator, lt, w, md)
	}
	return link.NewLink(source, target, lt, w, md)
}

func (f *fixture) outboxRows(t *testing.T) []Model {
	t.Helper()
	var rows []Model
	require.NoError(t, f.db.Order("id").Find(&rows).Error)
	return rows
}

func (f *fixture) unsent(t *testing.T) []Model {
	t.Helper()
	var rows []Model
	require.NoError(t, f.db.Where("sent_at IS NULL").Order("id").Find(&rows).Error)
	return rows
}

func notePayloadOf(t *testing.T, row Model) (noteID, userID string) {
	t.Helper()
	var p infevents.NoteEventPayload
	require.NoError(t, json.Unmarshal(row.Payload, &p))
	return p.NoteID, p.UserID
}

func linkPayloadOf(t *testing.T, row Model) (source, target, userID string) {
	t.Helper()
	var p infevents.LinkEventPayload
	require.NoError(t, json.Unmarshal(row.Payload, &p))
	return p.SourceNoteID, p.TargetNoteID, p.UserID
}

// Criterion 1: every write leaves an outbox row with the event type, entity
// and owner that the relayer needs.
func TestOutbox_NoteWritesRecordEvents(t *testing.T) {
	f, cleanup := setup(t)
	defer cleanup()

	user := f.createUser(t)
	n := f.mustNote(t, "created", &user)
	require.NoError(t, f.noteRepo.Save(f.ctx, n))

	rows := f.unsent(t)
	require.Len(t, rows, 1)
	assert.Equal(t, EventNoteCreated, rows[0].EventType)
	assert.Equal(t, n.ID(), rows[0].EntityID)
	id, uid := notePayloadOf(t, rows[0])
	assert.Equal(t, n.ID().String(), id)
	assert.Equal(t, user.String(), uid)

	// Save on an existing row is an update.
	n2 := f.mustNote(t, "renamed", &user)
	require.NoError(t, f.noteRepo.Save(f.ctx, note.ReconstructNoteWithCreator(n.ID(), n2.Title(), n2.Content(),
		note.MustType(n2.Type()), n2.Metadata(), &user, n.CreatedAt(), n2.UpdatedAt())))

	rows = f.unsent(t)
	require.Len(t, rows, 2)
	assert.Equal(t, EventNoteUpdated, rows[1].EventType)

	// Delete and restore.
	require.NoError(t, f.noteRepo.Delete(f.ctx, n.ID()))
	require.NoError(t, f.noteRepo.Restore(f.ctx, n.ID()))

	rows = f.unsent(t)
	require.Len(t, rows, 4)
	assert.Equal(t, EventNoteDeleted, rows[2].EventType)
	assert.Equal(t, EventNoteUpdated, rows[3].EventType)
}

func TestOutbox_LinkWritesRecordEvents(t *testing.T) {
	f, cleanup := setup(t)
	defer cleanup()

	user := f.createUser(t)
	src := f.mustNote(t, "src", &user)
	dst := f.mustNote(t, "dst", &user)
	require.NoError(t, f.noteRepo.Save(f.ctx, src))
	require.NoError(t, f.noteRepo.Save(f.ctx, dst))
	require.NoError(t, f.db.Exec("TRUNCATE TABLE graph_outbox").Error)

	l := mustLink(t, src.ID(), dst.ID(), &user)
	require.NoError(t, f.linkRepo.Save(f.ctx, l))

	rows := f.unsent(t)
	require.Len(t, rows, 1)
	assert.Equal(t, EventLinkCreated, rows[0].EventType)
	s, tgt, uid := linkPayloadOf(t, rows[0])
	assert.Equal(t, src.ID().String(), s)
	assert.Equal(t, dst.ID().String(), tgt)
	assert.Equal(t, user.String(), uid)

	l.UpdateWeight(mustWeight(t, 0.7))
	require.NoError(t, f.linkRepo.Update(f.ctx, l))

	require.NoError(t, f.linkRepo.Delete(f.ctx, l.ID()))

	rows = f.unsent(t)
	require.Len(t, rows, 3)
	assert.Equal(t, EventLinkUpdated, rows[1].EventType)
	assert.Equal(t, EventLinkDeleted, rows[2].EventType)
	assert.Equal(t, l.ID(), rows[2].EntityID)
}

func mustWeight(t *testing.T, v float64) link.Weight {
	t.Helper()
	w, err := link.NewWeight(v)
	require.NoError(t, err)
	return w
}

// Gamma links carry no creator — the event's owner resolves to the source
// note's creator, like the manual publish did in stage A.
func TestOutbox_GammaLinkOwnerFallsBackToSourceNoteCreator(t *testing.T) {
	f, cleanup := setup(t)
	defer cleanup()

	user := f.createUser(t)
	src := f.mustNote(t, "src", &user)
	dst := f.mustNote(t, "dst", &user)
	require.NoError(t, f.noteRepo.Save(f.ctx, src))
	require.NoError(t, f.noteRepo.Save(f.ctx, dst))
	require.NoError(t, f.db.Exec("TRUNCATE TABLE graph_outbox").Error)

	gamma := mustLink(t, src.ID(), dst.ID(), nil) // no creator
	require.NoError(t, f.linkRepo.Save(f.ctx, gamma))

	rows := f.unsent(t)
	require.Len(t, rows, 1)
	_, _, uid := linkPayloadOf(t, rows[0])
	assert.Equal(t, user.String(), uid)
}

// Criterion 1, rollback half: a failed write commits nothing — no row in the
// table, no event to deliver.
func TestOutbox_FailedWriteLeavesNoEvent(t *testing.T) {
	f, cleanup := setup(t)
	defer cleanup()

	// Restore of a live note fails inside the transaction.
	err := f.noteRepo.Restore(f.ctx, uuid.New())
	require.ErrorIs(t, err, note.ErrNoteNotFound)
	assert.Empty(t, f.outboxRows(t))

	// A note whose creator does not exist violates the FK inside the tx.
	ghost := uuid.New()
	bad := f.mustNote(t, "ghost", &ghost)
	require.Error(t, f.noteRepo.Save(f.ctx, bad))
	assert.Empty(t, f.outboxRows(t))
	assert.Empty(t, f.unsent(t))
}

// Criterion 1, mechanism level: a write and its outbox row share one
// transaction — rolling it back removes both.
func TestOutbox_RollbackRemovesWriteAndEvent(t *testing.T) {
	f, cleanup := setup(t)
	defer cleanup()

	inner := postgres.NewNoteRepository(f.db, nil)
	user := f.createUser(t)
	n := f.mustNote(t, "doomed", &user)

	tx := f.db.Begin()
	require.NoError(t, tx.Error)
	txCtx := postgres.ContextWithTx(f.ctx, tx)
	created, err := inner.SaveReturning(txCtx, n)
	require.NoError(t, err)
	require.True(t, created)
	require.NoError(t, insertEvents(tx, []Model{noteEvent(EventNoteCreated, n.ID(), n.CreatorID())}))
	require.NoError(t, tx.Rollback().Error)

	var noteCount, outboxCount int64
	require.NoError(t, f.db.Model(&postgres.NoteModel{}).Where("id = ?", n.ID()).Count(&noteCount).Error)
	require.NoError(t, f.db.Model(&Model{}).Count(&outboxCount).Error)
	assert.Zero(t, noteCount)
	assert.Zero(t, outboxCount)
}

// Criterion 2: a committed write whose process dies before the relay keeps
// its row; a relayer started later delivers it.
func TestOutbox_RelayDeliversAfterCrash(t *testing.T) {
	f, cleanup := setup(t)
	defer cleanup()

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	publisher := infevents.NewPublisher(rdb, "graph:events")

	sub := rdb.Subscribe(f.ctx, "graph:events")
	defer sub.Close()
	for {
		msg, err := sub.ReceiveTimeout(f.ctx, time.Second)
		require.NoError(t, err)
		if _, ok := msg.(*redis.Subscription); ok {
			break
		}
	}

	// Write committed; the "process" never relayed it.
	n := f.mustNote(t, "pending", nil)
	require.NoError(t, f.noteRepo.Save(f.ctx, n))
	require.Len(t, f.unsent(t), 1)

	relayer := NewRelayer(f.db, publisher, 0, 100)
	sent, err := relayer.Flush(f.ctx)
	require.NoError(t, err)
	require.Equal(t, 1, sent)
	assert.Empty(t, f.unsent(t))

	msg, err := sub.ReceiveTimeout(f.ctx, 2*time.Second)
	require.NoError(t, err)
	pubMsg, ok := msg.(*redis.Message)
	require.True(t, ok)
	var ev infevents.Event
	require.NoError(t, json.Unmarshal([]byte(pubMsg.Payload), &ev))
	assert.Equal(t, EventNoteCreated, ev.Event)
	var p infevents.NoteEventPayload
	require.NoError(t, json.Unmarshal(ev.Payload, &p))
	assert.Equal(t, n.ID().String(), p.NoteID)
}

// A publish failure inside the relay transaction leaves the row unsent — the
// next pass retries it (at-least-once).
func TestOutbox_RelayRetriesAfterPublishFailure(t *testing.T) {
	f, cleanup := setup(t)
	defer cleanup()

	n := f.mustNote(t, "retry", nil)
	require.NoError(t, f.noteRepo.Save(f.ctx, n))

	// Dead Redis: publish fails, the row must stay unsent.
	dead := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	relayer := NewRelayer(f.db, infevents.NewPublisher(dead, "graph:events"), 0, 100)
	sent, err := relayer.FlushOnce(f.ctx)
	require.Error(t, err)
	assert.Zero(t, sent)
	assert.Len(t, f.unsent(t), 1)

	// Redis recovers: the same row goes out on the next pass.
	mr := miniredis.RunT(t)
	live := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	relayer = NewRelayer(f.db, infevents.NewPublisher(live, "graph:events"), 0, 100)
	sent, err = relayer.FlushOnce(f.ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, sent)
	assert.Empty(t, f.unsent(t))
}

// PurgeSentBefore removes sent rows past the retention window and keeps
// unsent and fresh ones.
func TestOutbox_PurgeSentBefore(t *testing.T) {
	f, cleanup := setup(t)
	defer cleanup()

	old := time.Now().Add(-31 * 24 * time.Hour)
	fresh := time.Now().Add(-time.Hour)
	require.NoError(t, f.db.Exec(
		`INSERT INTO graph_outbox (event_type, entity_id, payload, sent_at) VALUES
		 ('NoteDeleted', ?, '{}', ?),
		 ('NoteDeleted', ?, '{}', ?),
		 ('NoteCreated', ?, '{}', NULL)`,
		uuid.New(), old, uuid.New(), fresh, uuid.New()).Error)

	n, err := PurgeSentBefore(f.db, f.ctx, time.Now().Add(-30*24*time.Hour))
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)

	rows := f.outboxRows(t)
	require.Len(t, rows, 2)
	assert.Nil(t, rows[1].SentAt)
}

// SaveUserLink is the most frequent write path (POST /links). Its outbox row
// must carry the saved link — a manual link's creator is the event owner.
func TestOutbox_SaveUserLinkRecordsEvent(t *testing.T) {
	f, cleanup := setup(t)
	defer cleanup()

	user := f.createUser(t)
	src := f.mustNote(t, "src", &user)
	dst := f.mustNote(t, "dst", &user)
	require.NoError(t, f.noteRepo.Save(f.ctx, src))
	require.NoError(t, f.noteRepo.Save(f.ctx, dst))
	require.NoError(t, f.db.Exec("TRUNCATE TABLE graph_outbox").Error)

	l := mustLink(t, src.ID(), dst.ID(), &user)
	saved, created, err := f.linkRepo.SaveUserLink(f.ctx, l)
	require.NoError(t, err)
	require.True(t, created)

	rows := f.unsent(t)
	require.Len(t, rows, 1)
	assert.Equal(t, EventLinkCreated, rows[0].EventType)
	assert.Equal(t, saved.ID(), rows[0].EntityID)
	s, tgt, uid := linkPayloadOf(t, rows[0])
	assert.Equal(t, src.ID().String(), s)
	assert.Equal(t, dst.ID().String(), tgt)
	assert.Equal(t, user.String(), uid)
}

// DeleteAndSuppress removes the link and records the rejection in one
// transaction — the event fires on the deleted link's pair.
func TestOutbox_DeleteAndSuppressRecordsEvent(t *testing.T) {
	f, cleanup := setup(t)
	defer cleanup()

	user := f.createUser(t)
	src := f.mustNote(t, "src", &user)
	dst := f.mustNote(t, "dst", &user)
	require.NoError(t, f.noteRepo.Save(f.ctx, src))
	require.NoError(t, f.noteRepo.Save(f.ctx, dst))
	gamma := link.NewGammaLink(src.ID(), dst.ID(), mustLinkType(t), mustWeight(t, 0.6), mustLinkMetadata(t))
	require.NoError(t, f.linkRepo.Save(f.ctx, gamma))
	require.NoError(t, f.db.Exec("TRUNCATE TABLE graph_outbox").Error)

	s := link.NewSuppression(src.ID(), dst.ID(), nil, &user)
	require.NoError(t, f.linkRepo.DeleteAndSuppress(f.ctx, gamma, s))

	rows := f.unsent(t)
	require.Len(t, rows, 1)
	assert.Equal(t, EventLinkDeleted, rows[0].EventType)
	assert.Equal(t, gamma.ID(), rows[0].EntityID)
	lSrc, lTgt, uid := linkPayloadOf(t, rows[0])
	assert.Equal(t, src.ID().String(), lSrc)
	assert.Equal(t, dst.ID().String(), lTgt)
	assert.Equal(t, user.String(), uid)

	var supCount int64
	require.NoError(t, f.db.Model(&postgres.LinkSuppressionModel{}).Where("id = ?", s.ID()).Count(&supCount).Error)
	assert.Equal(t, int64(1), supCount)
}

// DeleteBySource emits one LinkDeleted per removed row — the note deletion
// path relies on it to tell every subscriber which edges vanished.
func TestOutbox_DeleteBySourceRecordsEvents(t *testing.T) {
	f, cleanup := setup(t)
	defer cleanup()

	user := f.createUser(t)
	src := f.mustNote(t, "src", &user)
	a := f.mustNote(t, "a", &user)
	b := f.mustNote(t, "b", &user)
	require.NoError(t, f.noteRepo.Save(f.ctx, src))
	require.NoError(t, f.noteRepo.Save(f.ctx, a))
	require.NoError(t, f.noteRepo.Save(f.ctx, b))

	l1 := mustLink(t, src.ID(), a.ID(), &user)
	l2 := mustLink(t, src.ID(), b.ID(), &user)
	other := mustLink(t, a.ID(), b.ID(), &user)
	require.NoError(t, f.linkRepo.Save(f.ctx, l1))
	require.NoError(t, f.linkRepo.Save(f.ctx, l2))
	require.NoError(t, f.linkRepo.Save(f.ctx, other))
	require.NoError(t, f.db.Exec("TRUNCATE TABLE graph_outbox").Error)

	require.NoError(t, f.linkRepo.DeleteBySource(f.ctx, src.ID()))

	rows := f.unsent(t)
	require.Len(t, rows, 2)
	targets := map[string]bool{}
	for _, r := range rows {
		assert.Equal(t, EventLinkDeleted, r.EventType)
		_, tgt, uid := linkPayloadOf(t, r)
		targets[tgt] = true
		assert.Equal(t, user.String(), uid)
	}
	assert.True(t, targets[a.ID().String()])
	assert.True(t, targets[b.ID().String()])

	// The unrelated link survives untouched — and silent.
	var cnt int64
	require.NoError(t, f.db.Model(&postgres.LinkModel{}).Where("id = ? AND deleted_at IS NULL", other.ID()).Count(&cnt).Error)
	assert.Equal(t, int64(1), cnt)
}

// DeleteBySourceType wipes gamma links on regeneration — every wiped row must
// still publish LinkDeleted so subscribers drop it.
func TestOutbox_DeleteBySourceTypeRecordsEvents(t *testing.T) {
	f, cleanup := setup(t)
	defer cleanup()

	user := f.createUser(t)
	src := f.mustNote(t, "src", &user)
	dst := f.mustNote(t, "dst", &user)
	require.NoError(t, f.noteRepo.Save(f.ctx, src))
	require.NoError(t, f.noteRepo.Save(f.ctx, dst))

	gamma := link.NewGammaLink(src.ID(), dst.ID(), mustLinkType(t), mustWeight(t, 0.6), mustLinkMetadata(t))
	require.NoError(t, f.linkRepo.Save(f.ctx, gamma))
	manual := mustLink(t, dst.ID(), src.ID(), &user)
	require.NoError(t, f.linkRepo.Save(f.ctx, manual))
	require.NoError(t, f.db.Exec("TRUNCATE TABLE graph_outbox").Error)

	count, err := f.linkRepo.DeleteBySourceType(f.ctx, "gamma")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	rows := f.unsent(t)
	require.Len(t, rows, 1)
	assert.Equal(t, EventLinkDeleted, rows[0].EventType)
	assert.Equal(t, gamma.ID(), rows[0].EntityID)
	// The gamma link has no creator; the owner falls back to the source note.
	_, _, uid := linkPayloadOf(t, rows[0])
	assert.Equal(t, user.String(), uid)
}

// PurgeDeletedBefore hard-deletes notes past retention — each purged row still
// needs a NoteDeleted event for subscribers that cached it.
func TestOutbox_PurgeDeletedBeforeRecordsEvents(t *testing.T) {
	f, cleanup := setup(t)
	defer cleanup()

	user := f.createUser(t)
	n := f.mustNote(t, "purge me", &user)
	require.NoError(t, f.noteRepo.Save(f.ctx, n))
	require.NoError(t, f.noteRepo.Delete(f.ctx, n.ID()))
	require.NoError(t, f.db.Exec("TRUNCATE TABLE graph_outbox").Error)

	purged, err := f.noteRepo.PurgeDeletedBefore(f.ctx, time.Now().Add(time.Hour))
	require.NoError(t, err)
	assert.Equal(t, int64(1), purged)

	rows := f.unsent(t)
	require.Len(t, rows, 1)
	assert.Equal(t, EventNoteDeleted, rows[0].EventType)
	assert.Equal(t, n.ID(), rows[0].EntityID)
	id, uid := notePayloadOf(t, rows[0])
	assert.Equal(t, n.ID().String(), id)
	assert.Equal(t, user.String(), uid)
}

func mustLinkType(t *testing.T) link.LinkType {
	t.Helper()
	lt, err := link.NewLinkType("related")
	require.NoError(t, err)
	return lt
}

func mustLinkMetadata(t *testing.T) link.Metadata {
	t.Helper()
	md, err := link.NewMetadata(map[string]interface{}{})
	require.NoError(t, err)
	return md
}
