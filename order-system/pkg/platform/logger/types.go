package logger

import (
	"context"
	"time"
)

// Level represents the logging level
type Level int

const (
	// Debug level for detailed troubleshooting
	Debug Level = iota
	// Info level for general operational entries
	Info
	// Warn level for warning messages
	Warn
	// Error level for error messages
	Error
)

// String returns the string representation of the level
func (l Level) String() string {
	switch l {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Field represents a log field
type Field struct {
	Key   string
	Value interface{}
}

// Entry represents a log entry
type Entry struct {
	Level     Level
	Message   string
	Time      time.Time
	Fields    []Field
	TraceID   string
	SpanID    string
	Component string
	Error     error
}

// Logger defines the logging interface
type Logger interface {
	// Debug logs a debug message
	Debug(ctx context.Context, msg string, fields ...Field)
	// Info logs an info message
	Info(ctx context.Context, msg string, fields ...Field)
	// Warn logs a warning message
	Warn(ctx context.Context, msg string, fields ...Field)
	// Error logs an error message
	Error(ctx context.Context, msg string, err error, fields ...Field)
	// WithComponent returns a new logger with the component field set
	WithComponent(component string) Logger
	// WithFields returns a new logger with the given fields added
	WithFields(fields ...Field) Logger
}
