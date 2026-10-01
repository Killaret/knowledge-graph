package comet

import (
	"context"
	"testing"
	"time"

	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/domain/notification"
	"knowledge-graph/internal/domain/user"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeNoteReader struct {
	note *note.Note
	err  error
}

func (f *fakeNoteReader) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	return f.note, f.err
}

type fakeNotifRepo struct {
	created []notification.Notification
	dup     bool
}

func (f *fakeNotifRepo) CreateIfAbsent(ctx context.Context, n *notification.Notification) (bool, error) {
	if f.dup {
		return false, nil
	}
	f.created = append(f.created, *n)
	return true, nil
}

func (f *fakeNotifRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*notification.Notification, error) {
	return nil, nil
}

func (f *fakeNotifRepo) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	return 0, nil
}

func (f *fakeNotifRepo) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	return nil
}

type fakeUserReader struct {
	user *user.User
	err  error
}

func (f *fakeUserReader) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return f.user, f.err
}

type fakeEmailSender struct {
	calls []string
	err   error
}

func (f *fakeEmailSender) SendCometReminder(ctx context.Context, to, noteTitle string, dueAt time.Time) error {
	f.calls = append(f.calls, to)
	return f.err
}

func cometNote(t *testing.T, owner uuid.UUID, due time.Time, remindSec int64, doneAt *time.Time) *note.Note {
	t.Helper()
	title, err := note.NewTitle("Врач")
	require.NoError(t, err)
	content, err := note.NewContent("сходить")
	require.NoError(t, err)
	typ := note.MustType("comet")
	n := note.NewNote(title, content, typ, note.Metadata{},
		note.WithCometFields(&due, &remindSec, doneAt))
	n.SetCreatorID(owner)
	return n
}

func TestHandleReminder_DeliversNotification(t *testing.T) {
	owner := uuid.New()
	due := time.Now().Add(time.Hour)
	var sec int64 = 900
	n := cometNote(t, owner, due, sec, nil)

	notifs := &fakeNotifRepo{}
	svc := NewReminderService(&fakeNoteReader{note: n}, notifs, &fakeUserReader{}, nil)

	err := svc.HandleReminder(context.Background(), n.ID(), *n.RemindAt())
	require.NoError(t, err)
	require.Len(t, notifs.created, 1)
	assert.Equal(t, owner, notifs.created[0].UserID())
	assert.Equal(t, notification.TypeCometReminder, notifs.created[0].Type())
	assert.Equal(t, ReminderDedupeKey(n.ID(), *n.RemindAt()), notifs.created[0].DedupeKey())
}

func TestHandleReminder_StaleWhenRescheduled(t *testing.T) {
	owner := uuid.New()
	due := time.Now().Add(time.Hour)
	var sec int64 = 900
	n := cometNote(t, owner, due, sec, nil)

	notifs := &fakeNotifRepo{}
	svc := NewReminderService(&fakeNoteReader{note: n}, notifs, &fakeUserReader{}, nil)

	// task scheduled for a different remind_at — e.g. the date moved
	stale := n.RemindAt().Add(time.Minute)
	require.NoError(t, svc.HandleReminder(context.Background(), n.ID(), stale))
	assert.Empty(t, notifs.created)
}

func TestHandleReminder_SkipsDone(t *testing.T) {
	owner := uuid.New()
	due := time.Now().Add(time.Hour)
	var sec int64 = 900
	done := time.Now()
	n := cometNote(t, owner, due, sec, &done)

	notifs := &fakeNotifRepo{}
	svc := NewReminderService(&fakeNoteReader{note: n}, notifs, &fakeUserReader{}, nil)

	// RemindAt() is nil for a done note — the task was scheduled for
	// due − remind_before before the note was completed.
	scheduled := due.Add(-time.Duration(sec) * time.Second)
	require.NoError(t, svc.HandleReminder(context.Background(), n.ID(), scheduled))
	assert.Empty(t, notifs.created)
}

func TestHandleReminder_SkipsDeleted(t *testing.T) {
	notifs := &fakeNotifRepo{}
	svc := NewReminderService(&fakeNoteReader{err: note.ErrNoteNotFound}, notifs, &fakeUserReader{}, nil)

	require.NoError(t, svc.HandleReminder(context.Background(), uuid.New(), time.Now()))
	assert.Empty(t, notifs.created)
}

func TestHandleReminder_DedupeSuppressesEmail(t *testing.T) {
	owner := uuid.New()
	due := time.Now().Add(time.Hour)
	var sec int64 = 900
	n := cometNote(t, owner, due, sec, nil)

	usr, err := user.NewUser(owner, "u", "u@example.com", "hash", "user", time.Now(), time.Now(), nil)
	require.NoError(t, err)

	notifs := &fakeNotifRepo{dup: true}
	email := &fakeEmailSender{}
	svc := NewReminderService(&fakeNoteReader{note: n}, notifs, &fakeUserReader{user: usr}, email)

	require.NoError(t, svc.HandleReminder(context.Background(), n.ID(), *n.RemindAt()))
	assert.Empty(t, email.calls, "duplicate delivery must not re-send email")
}

func TestHandleReminder_SendsEmailWhenConfigured(t *testing.T) {
	owner := uuid.New()
	due := time.Now().Add(time.Hour)
	var sec int64 = 900
	n := cometNote(t, owner, due, sec, nil)

	usr, err := user.NewUser(owner, "u", "u@example.com", "hash", "user", time.Now(), time.Now(), nil)
	require.NoError(t, err)

	notifs := &fakeNotifRepo{}
	email := &fakeEmailSender{}
	svc := NewReminderService(&fakeNoteReader{note: n}, notifs, &fakeUserReader{user: usr}, email)

	require.NoError(t, svc.HandleReminder(context.Background(), n.ID(), *n.RemindAt()))
	assert.Equal(t, []string{"u@example.com"}, email.calls)
}

func TestHandleReminder_NoEmailWithoutSMTP(t *testing.T) {
	owner := uuid.New()
	due := time.Now().Add(time.Hour)
	var sec int64 = 900
	n := cometNote(t, owner, due, sec, nil)

	notifs := &fakeNotifRepo{}
	// email sender nil — SMTP not configured
	svc := NewReminderService(&fakeNoteReader{note: n}, notifs, &fakeUserReader{}, nil)

	require.NoError(t, svc.HandleReminder(context.Background(), n.ID(), *n.RemindAt()))
	assert.Len(t, notifs.created, 1)
}
