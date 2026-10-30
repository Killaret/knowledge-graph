// Package email provides a Sender abstraction for dispatching transactional emails.
package email

import (
	"knowledge-graph/internal/auth"
)

// Sender is the port used by the application layer to send emails.
// It is an alias to the auth-level port so the interface layer never depends
// on an infrastructure package.
type Sender = auth.EmailSender

// Compile-time interface assertions.
