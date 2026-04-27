// Package logger configuration helpers
package logger

import (
	"os"
	"path/filepath"
)

// DefaultConfig returns default logger configuration
func DefaultConfig() Config {
	// Get project root (assuming binary runs from backend dir)
	logDir := "../logs/backend"
	
	// If logs dir doesn't exist relative to binary, use absolute path
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		// Try to create logs directory
		_ = os.MkdirAll(logDir, 0755)
	}
	
	return Config{
		Level:      INFO,
		JSONFormat: false,
		LogFile:    filepath.Join(logDir, "app.log"),
	}
}

// ProductionConfig returns production logger configuration with JSON format
func ProductionConfig() Config {
	cfg := DefaultConfig()
	cfg.JSONFormat = true
	cfg.Level = WARN
	return cfg
}

// DevelopmentConfig returns development logger configuration
func DevelopmentConfig() Config {
	cfg := DefaultConfig()
	cfg.JSONFormat = false
	cfg.Level = DEBUG
	return cfg
}
