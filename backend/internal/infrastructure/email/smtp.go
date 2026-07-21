package email

import (
	"context"
	"fmt"
	"net/smtp"
)

// SMTPSender sends emails via an SMTP server.
type SMTPSender struct {
	host     string
	port     int
	user     string
	password string
	from     string
}

// NewSMTP creates a new SMTP email sender.
func NewSMTP(host string, port int, user, password, from string) Sender {
	return &SMTPSender{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		from:     from,
	}
}

// SendPasswordReset sends a password-reset email via SMTP.
func (s *SMTPSender) SendPasswordReset(ctx context.Context, to, resetLink string) error {
	subject := "Password reset request"
	body := fmt.Sprintf("Click the link to reset your password: %s", resetLink)
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	var auth smtp.Auth
	if s.user != "" && s.password != "" {
		auth = smtp.PlainAuth("", s.user, s.password, s.host)
	}

	return smtp.SendMail(addr, auth, s.from, []string{to}, msg)
}
