// Package comet contains the application service behind the comet:remind
// worker task (COMET-1 stage C).
package comet

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/domain/notification"
	"knowledge-graph/internal/domain/user"

	"github.com/google/uuid"
)

// NoteReader is the narrow read the reminder service needs.
type NoteReader interface {
	FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error)
}

// UserReader resolves the recipient's email address.
type UserReader interface {
	FindByID(ctx context.Context, id uuid.UUID) (*user.User, error)
}

// ReminderEmailSender sends a comet reminder email. Nil means SMTP is not
// configured — delivery is then in-app only.
type ReminderEmailSender interface {
	SendCometReminder(ctx context.Context, to, noteTitle string, dueAt time.Time) error
}

// ReminderService delivers a due comet reminder. Staleness is checked at fire
// time: a task carrying a remind_at that no longer matches the note is a no-op,
// which makes rescheduling a plain re-enqueue — no task cancellation needed.
type ReminderService struct {
	notes  NoteReader
	notifs notification.Repository
	users  UserReader
	email  ReminderEmailSender
}

// NewReminderService builds the service. emailSender may be nil.
func NewReminderService(notes NoteReader, notifs notification.Repository, users UserReader, emailSender ReminderEmailSender) *ReminderService {
	return &ReminderService{notes: notes, notifs: notifs, users: users, email: emailSender}
}

// ReminderDedupeKey is exported for tests and delivery logging.
func ReminderDedupeKey(noteID uuid.UUID, remindAt time.Time) string {
	return fmt.Sprintf("comet:%s:%d", noteID.String(), remindAt.Unix())
}

// HandleReminder delivers the reminder if the task is still fresh.
// expectedRemindAt is the moment the task was scheduled for; it must match
// note.RemindAt() exactly, otherwise the note was rescheduled or cleared.
func (s *ReminderService) HandleReminder(ctx context.Context, noteID uuid.UUID, expectedRemindAt time.Time) error {
	n, err := s.notes.FindByID(ctx, noteID)
	if errors.Is(err, note.ErrNoteNotFound) {
		log.Printf("[comet:remind] note %s not found/deleted — skipping", noteID)
		return nil
	}
	if err != nil {
		return fmt.Errorf("loading note %s: %w", noteID, err)
	}
	if n.DoneAt() != nil {
		log.Printf("[comet:remind] note %s is done — skipping", noteID)
		return nil
	}
	remindAt := n.RemindAt()
	if remindAt == nil || !remindAt.Equal(expectedRemindAt) {
		log.Printf("[comet:remind] note %s remind_at changed — stale task, skipping", noteID)
		return nil
	}
	if n.CreatorID() == nil {
		log.Printf("[comet:remind] note %s has no owner — skipping", noteID)
		return nil
	}

	ownerID := *n.CreatorID()
	dueLabel := ""
	if n.DueAt() != nil {
		dueLabel = n.DueAt().Format(time.RFC3339)
	}
	noteIDCopy := noteID
	nf := notification.NewNotification(
		ownerID, &noteIDCopy, notification.TypeCometReminder,
		fmt.Sprintf("Comet: %s", n.Title()),
		fmt.Sprintf("Due at %s", dueLabel),
		ReminderDedupeKey(noteID, expectedRemindAt),
	)
	created, err := s.notifs.CreateIfAbsent(ctx, nf)
	if err != nil {
		return fmt.Errorf("storing notification for note %s: %w", noteID, err)
	}
	if !created {
		log.Printf("[comet:remind] note %s already notified — dedupe", noteID)
		return nil
	}
	log.Printf("[comet:remind] notification delivered for note %s to user %s", noteID, ownerID)

	if s.email == nil || s.users == nil {
		return nil
	}
	usr, err := s.users.FindByID(ctx, ownerID)
	if err != nil || usr == nil {
		log.Printf("[comet:remind] no user record for %s — email skipped (%v)", ownerID, err)
		return nil
	}
	var due time.Time
	if n.DueAt() != nil {
		due = *n.DueAt()
	}
	if err := s.email.SendCometReminder(ctx, usr.Email(), n.Title().String(), due); err != nil {
		// Delivery fact is already stored in-app; email failure must not retry
		// the task into a duplicate — log and swallow.
		log.Printf("[comet:remind] email to %s failed: %v", usr.Email(), err)
		return nil
	}
	log.Printf("[comet:remind] email sent to %s for note %s", usr.Email(), noteID)
	return nil
}
