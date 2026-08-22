package email

import (
	"context"
	"errors"
	"net/smtp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSMTPSender_SendPasswordReset(t *testing.T) {
	called := false
	sender := &SMTPSender{
		host:     "smtp.example.com",
		port:     587,
		user:     "user",
		password: "pass",
		from:     "noreply@example.com",
		sendMailFunc: func(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
			called = true
			assert.Equal(t, "smtp.example.com:587", addr)
			assert.Equal(t, "noreply@example.com", from)
			assert.Equal(t, []string{"user@example.com"}, to)
			assert.Contains(t, string(msg), "Password reset request")
			return nil
		},
	}

	err := sender.SendPasswordReset(context.Background(), "user@example.com", "http://reset")
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestSMTPSender_SendPasswordReset_Error(t *testing.T) {
	sender := &SMTPSender{
		host: "smtp.example.com",
		port: 587,
		from: "noreply@example.com",
		sendMailFunc: func(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
			return errors.New("smtp error")
		},
	}

	err := sender.SendPasswordReset(context.Background(), "user@example.com", "http://reset")
	assert.Error(t, err)
}

func TestNewSMTP(t *testing.T) {
	sender := NewSMTP("host", 123, "user", "pass", "from").(*SMTPSender)
	assert.Equal(t, "host", sender.host)
	assert.Equal(t, 123, sender.port)
	assert.Equal(t, "user", sender.user)
	assert.Equal(t, "pass", sender.password)
	assert.Equal(t, "from", sender.from)
}
