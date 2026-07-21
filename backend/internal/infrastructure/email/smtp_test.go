package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSMTP(t *testing.T) {
	sender := NewSMTP("smtp.example.com", 587, "user@example.com", "password", "noreply@example.com")
	assert.NotNil(t, sender)
}
