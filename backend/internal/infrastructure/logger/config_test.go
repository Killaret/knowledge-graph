package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, INFO, cfg.Level)
	assert.Contains(t, cfg.LogFile, "app.log")
	assert.False(t, cfg.JSONFormat)
}

func TestProductionConfig(t *testing.T) {
	cfg := ProductionConfig()
	assert.True(t, cfg.JSONFormat)
	assert.Equal(t, WARN, cfg.Level)
}
