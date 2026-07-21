package email

import "knowledge-graph/internal/config"

// FromConfig builds a Sender from application configuration.
// It chooses an SMTP sender when SMTP host is configured; otherwise it falls
// back to the console logger (development mode).
func FromConfig(cfg *config.Config) Sender {
	if cfg.SMTPHost != "" {
		return NewSMTP(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPFrom)
	}
	return NewConsole()
}
