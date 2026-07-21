// Package email provides a Sender abstraction for dispatching transactional emails.
package email

import "context"

// Sender is the port used by the application layer to send emails.
type Sender interface {
	// SendPasswordReset sends a password-reset email to the given address.
	// The resetLink is a fully qualified URL that the user should follow.
	SendPasswordReset(ctx context.Context, to, resetLink string) error
}

// Compile-time interface assertions.
