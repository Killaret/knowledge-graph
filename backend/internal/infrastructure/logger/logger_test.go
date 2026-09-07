package logger

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestLogLevelString(t *testing.T) {
	tests := []struct {
		level LogLevel
		want  string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{LogLevel(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.want {
			t.Errorf("level %d: got %q, want %q", tt.level, got, tt.want)
		}
	}
}

func TestLogger_New(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{Level: INFO, Output: &buf, JSONFormat: false}
	l, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l == nil {
		t.Fatal("expected non-nil logger")
	}

	l.Info("hello")
	if !strings.Contains(buf.String(), "hello") {
		t.Errorf("expected log to contain 'hello', got %q", buf.String())
	}
}

func TestLogger_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	l, _ := New(Config{Level: INFO, Output: &buf, JSONFormat: true})
	l.Info("json test", map[string]interface{}{"key": "value"})

	if !strings.Contains(buf.String(), `"message":"json test"`) {
		t.Errorf("expected JSON message, got %q", buf.String())
	}
}

func TestLogger_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l, _ := New(Config{Level: WARN, Output: &buf, JSONFormat: false})

	l.Debug("debug message")
	l.Info("info message")
	if strings.Contains(buf.String(), "debug message") || strings.Contains(buf.String(), "info message") {
		t.Error("debug/info messages should be filtered out")
	}

	l.Warn("warning message")
	if !strings.Contains(buf.String(), "warning message") {
		t.Error("warn message should be logged")
	}
}

func TestLogger_WithContext(t *testing.T) {
	var buf bytes.Buffer
	l, _ := New(Config{Level: INFO, Output: &buf, JSONFormat: false})
	ctxLogger := l.WithContext("worker")
	ctxLogger.Info("ctx message")

	if !strings.Contains(buf.String(), "[worker]") {
		t.Errorf("expected context in log, got %q", buf.String())
	}
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	l, _ := New(Config{Level: ERROR, Output: &buf, JSONFormat: false})
	l.Error("failed", errors.New("boom"))

	if !strings.Contains(buf.String(), "boom") {
		t.Errorf("expected error in log, got %q", buf.String())
	}
}
