package email

import (
	"testing"

	"knowledge-graph/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestFromConfig_ReturnsSMTP(t *testing.T) {
	cfg := &config.Config{
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		SMTPUser:     "user",
		SMTPPassword: "pass",
		SMTPFrom:     "noreply@example.com",
	}

	sender := FromConfig(cfg)
	assert.IsType(t, &SMTPSender{}, sender)
}

func TestFromConfig_ReturnsConsole(t *testing.T) {
	cfg := &config.Config{}

	sender := FromConfig(cfg)
	assert.IsType(t, &ConsoleSender{}, sender)
}
