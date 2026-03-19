package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value any
}

// F is a helper to create a Field
func F(key string, value any) Field {
	return Field{Key: key, Value: value}
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
	Fields    map[string]any
}

// Logger provides structured logging
type Logger struct {
	enabled bool
	entries []LogEntry
}

// New creates a new Logger
func New(enabled bool) *Logger {
	return &Logger{
		enabled: enabled,
		entries: make([]LogEntry, 0),
	}
}

// IsEnabled returns whether logging is enabled
func (l *Logger) IsEnabled() bool {
	return l.enabled
}

// Info logs an info level message
func (l *Logger) Info(msg string, fields ...Field) {
	if !l.enabled {
		return
	}
	l.log("INFO", msg, fields)
}

// Debug logs a debug level message
func (l *Logger) Debug(msg string, fields ...Field) {
	if !l.enabled {
		return
	}
	l.log("DEBUG", msg, fields)
}

// Warn logs a warning level message
func (l *Logger) Warn(msg string, fields ...Field) {
	if !l.enabled {
		return
	}
	l.log("WARN", msg, fields)
}

// Error logs an error level message
func (l *Logger) Error(msg string, err error, fields ...Field) {
	if !l.enabled {
		return
	}
	if err != nil {
		fields = append(fields, Field{Key: "error", Value: err.Error()})
	}
	l.log("ERROR", msg, fields)
}

func (l *Logger) log(level, msg string, fields []Field) {
	fieldMap := make(map[string]any)
	for _, f := range fields {
		fieldMap[f.Key] = f.Value
	}

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
		Fields:    fieldMap,
	}

	l.entries = append(l.entries, entry)
}

// WriteToFile writes all log entries to a file in human-readable format
func (l *Logger) WriteToFile(path string) error {
	if !l.enabled || len(l.entries) == 0 {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	var sb strings.Builder
	for _, entry := range l.entries {
		// Format: 2024-01-15 10:30:45 [INFO] Message - key=value, key2=value2
		timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
		sb.WriteString(fmt.Sprintf("%s [%s] %s", timestamp, entry.Level, entry.Message))

		if len(entry.Fields) > 0 {
			var fieldStrs []string
			for k, v := range entry.Fields {
				fieldStrs = append(fieldStrs, fmt.Sprintf("%s=%v", k, v))
			}
			sb.WriteString(" - " + strings.Join(fieldStrs, ", "))
		}

		sb.WriteString("\n")
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

// GetEntries returns all log entries (useful for testing)
func (l *Logger) GetEntries() []LogEntry {
	return l.entries
}
