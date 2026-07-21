package email

import (
	"context"
	"log"
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
