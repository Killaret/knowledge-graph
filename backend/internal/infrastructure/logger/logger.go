// Package logger provides structured logging for the backend application
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Context   string                 `json:"context"`
	File      string                 `json:"file,omitempty"`
	Line      int                    `json:"line,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// Logger provides structured logging capabilities
type Logger struct {
	context    string
	level      LogLevel
	output     io.Writer
	jsonOutput bool
	mu         sync.Mutex
}

// Config holds logger configuration
type Config struct {
	Level      LogLevel
	Output     io.Writer
	JSONFormat bool
	LogFile    string
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// Initialize sets up the default logger
func Initialize(cfg Config) error {
	var err error
	once.Do(func() {
		defaultLogger, err = New(cfg)
	})
	return err
}

// New creates a new logger instance
func New(cfg Config) (*Logger, error) {
	var output io.Writer = cfg.Output
	
	if cfg.LogFile != "" {
		// Ensure log directory exists
		dir := filepath.Dir(cfg.LogFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
		
		file, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		
		// Write to both file and stdout
		output = io.MultiWriter(os.Stdout, file)
	}
	
	if output == nil {
		output = os.Stdout
	}
	
	return &Logger{
		context:    "app",
		level:      cfg.Level,
		output:     output,
		jsonOutput: cfg.JSONFormat,
	}, nil
}

// WithContext creates a new logger with a specific context
func (l *Logger) WithContext(context string) *Logger {
	return &Logger{
		context:    context,
		level:      l.level,
		output:     l.output,
		jsonOutput: l.jsonOutput,
	}
}

// getCallerInfo returns file and line number of the caller
func getCallerInfo() (string, int) {
	_, file, line, ok := runtime.Caller(3) // Skip 3 frames to get to actual caller
	if !ok {
		return "unknown", 0
	}
	// Get just the filename
	parts := strings.Split(file, "/")
	return parts[len(parts)-1], line
}

// log writes a log entry
func (l *Logger) log(level LogLevel, message string, data map[string]interface{}) {
	if level < l.level {
		return
	}
	
	l.mu.Lock()
	defer l.mu.Unlock()
	
	file, line := getCallerInfo()
	
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level.String(),
		Message:   message,
		Context:   l.context,
		File:      file,
		Line:      line,
		Data:      data,
	}
	
	if l.jsonOutput {
		jsonData, _ := json.Marshal(entry)
		fmt.Fprintln(l.output, string(jsonData))
	} else {
		dataStr := ""
		if len(data) > 0 {
			jsonData, _ := json.Marshal(data)
			dataStr = string(jsonData)
		}
		fmt.Fprintf(l.output, "[%s] [%s] [%s] %s:%d - %s %s\n",
			entry.Timestamp, entry.Level, entry.Context,
			entry.File, entry.Line, entry.Message, dataStr)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(message string, data ...map[string]interface{}) {
	var d map[string]interface{}
	if len(data) > 0 {
		d = data[0]
	}
	l.log(DEBUG, message, d)
}

// Info logs an info message
func (l *Logger) Info(message string, data ...map[string]interface{}) {
	var d map[string]interface{}
	if len(data) > 0 {
		d = data[0]
	}
	l.log(INFO, message, d)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, data ...map[string]interface{}) {
	var d map[string]interface{}
	if len(data) > 0 {
		d = data[0]
	}
	l.log(WARN, message, d)
}

// Error logs an error message
func (l *Logger) Error(message string, err error, data ...map[string]interface{}) {
	d := map[string]interface{}{}
	if len(data) > 0 && data[0] != nil {
		for k, v := range data[0] {
			d[k] = v
		}
	}
	if err != nil {
		d["error"] = err.Error()
	}
	l.log(ERROR, message, d)
}

// Package-level functions for default logger

// WithContext creates a new logger with context from default logger
func WithContext(context string) *Logger {
	if defaultLogger == nil {
		// Initialize with defaults if not initialized
		_ = Initialize(Config{Level: INFO, Output: os.Stdout})
	}
	return defaultLogger.WithContext(context)
}

// Debug uses default logger
func Debug(message string, data ...map[string]interface{}) {
	if defaultLogger == nil {
		_ = Initialize(Config{Level: INFO, Output: os.Stdout})
	}
	defaultLogger.Debug(message, data...)
}

// Info uses default logger
func Info(message string, data ...map[string]interface{}) {
	if defaultLogger == nil {
		_ = Initialize(Config{Level: INFO, Output: os.Stdout})
	}
	defaultLogger.Info(message, data...)
}

// Warn uses default logger
func Warn(message string, data ...map[string]interface{}) {
	if defaultLogger == nil {
		_ = Initialize(Config{Level: INFO, Output: os.Stdout})
	}
	defaultLogger.Warn(message, data...)
}

// Error uses default logger
func Error(message string, err error, data ...map[string]interface{}) {
	if defaultLogger == nil {
		_ = Initialize(Config{Level: INFO, Output: os.Stdout})
	}
	defaultLogger.Error(message, err, data...)
}

// SetOutput sets the output writer for the default logger
func SetOutput(w io.Writer) {
	if defaultLogger != nil {
		defaultLogger.output = w
	}
}

// NewStandardLogger creates a standard library compatible logger
func NewStandardLogger(context string) *log.Logger {
	return log.New(&standardLoggerAdapter{context: context}, "", 0)
}

type standardLoggerAdapter struct {
	context string
}

func (a *standardLoggerAdapter) Write(p []byte) (n int, err error) {
	msg := strings.TrimSpace(string(p))
	WithContext(a.context).Info(msg)
	return len(p), nil
}
