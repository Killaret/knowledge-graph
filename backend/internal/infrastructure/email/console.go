package email

import (
	"context"
	"log"
	"time"
)

// ConsoleSender logs emails to stdout. Useful for local development and tests.
type ConsoleSender struct{}

// NewConsole creates a new console email sender.
func NewConsole() Sender {
	return &ConsoleSender{}
}

// SendPasswordReset logs the password-reset link.
func (s *ConsoleSender) SendPasswordReset(ctx context.Context, to, resetLink string) error {
	log.Printf("[EMAIL] Password reset for %s: %s", to, resetLink)
	return nil
}

// SendCometReminder logs the comet reminder (COMET-1 stage C).
func (s *ConsoleSender) SendCometReminder(ctx context.Context, to, noteTitle string, dueAt time.Time) error {
	log.Printf("[EMAIL] Comet reminder for %s: %q due %s", to, noteTitle, dueAt.Format(time.RFC3339))
	return nil
}
