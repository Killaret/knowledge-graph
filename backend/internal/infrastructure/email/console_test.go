package email

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsoleSender_SendPasswordReset(t *testing.T) {
	sender := NewConsole()
	err := sender.SendPasswordReset(context.Background(), "user@example.com", "http://example.com/reset?token=abc")
	assert.NoError(t, err)
}
