package logger

import (
	"bytes"
	"strings"
	"sync"
	"testing"
)

func TestLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	l, _ := New(Config{Level: DEBUG, Output: &buf, JSONFormat: false})
	l.Debug("debug message", map[string]interface{}{"detail": "data"})

	if !strings.Contains(buf.String(), "debug message") {
		t.Errorf("expected debug message in log, got %q", buf.String())
	}
}

func TestLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	l, _ := New(Config{Level: WARN, Output: &buf, JSONFormat: false})
	l.Warn("warn message")

	if !strings.Contains(buf.String(), "warn message") {
		t.Errorf("expected warn message in log, got %q", buf.String())
	}
}

func TestInitialize(t *testing.T) {
	defaultLogger = nil
	once = sync.Once{}
	err := Initialize(Config{Level: INFO, Output: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if defaultLogger == nil {
		t.Fatal("expected default logger to be initialized")
	}
}

func TestPackageLevelFunctions(t *testing.T) {
	defaultLogger = nil
	once = sync.Once{}
	WithContext("test")

	var buf bytes.Buffer
	defaultLogger.output = &buf
	defaultLogger.level = INFO

	Debug("debug")
	Info("info")
	Warn("warn")
	Error("error", nil)

	if !strings.Contains(buf.String(), "info") {
		t.Errorf("expected info message, got %q", buf.String())
	}
}
