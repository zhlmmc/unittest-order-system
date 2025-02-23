package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"order-system/pkg/infra/config"
)

// defaultLogger implements the Logger interface
type defaultLogger struct {
	mu        sync.Mutex
	out       io.Writer
	level     Level
	component string
	fields    []Field
}

// New creates a new logger
func New(cfg *config.Config) (Logger, error) {
	level, err := parseLevel(cfg.Logger.Level)
	if err != nil {
		return nil, err
	}

	var out io.Writer
	switch cfg.Logger.Output {
	case "stdout":
		out = os.Stdout
	case "stderr":
		out = os.Stderr
	default:
		file, err := os.OpenFile(cfg.Logger.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		out = file
	}

	return &defaultLogger{
		out:   out,
		level: level,
	}, nil
}

// parseLevel parses the log level string
func parseLevel(level string) (Level, error) {
	switch level {
	case "debug":
		return Debug, nil
	case "info":
		return Info, nil
	case "warn":
		return Warn, nil
	case "error":
		return Error, nil
	default:
		return Info, fmt.Errorf("invalid log level: %s", level)
	}
}

// Debug implements Logger.Debug
func (l *defaultLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	if l.level <= Debug {
		l.log(ctx, Debug, msg, nil, fields...)
	}
}

// Info implements Logger.Info
func (l *defaultLogger) Info(ctx context.Context, msg string, fields ...Field) {
	if l.level <= Info {
		l.log(ctx, Info, msg, nil, fields...)
	}
}

// Warn implements Logger.Warn
func (l *defaultLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	if l.level <= Warn {
		l.log(ctx, Warn, msg, nil, fields...)
	}
}

// Error implements Logger.Error
func (l *defaultLogger) Error(ctx context.Context, msg string, err error, fields ...Field) {
	if l.level <= Error {
		l.log(ctx, Error, msg, err, fields...)
	}
}

// WithComponent implements Logger.WithComponent
func (l *defaultLogger) WithComponent(component string) Logger {
	return &defaultLogger{
		out:       l.out,
		level:     l.level,
		component: component,
		fields:    l.fields,
	}
}

// WithFields implements Logger.WithFields
func (l *defaultLogger) WithFields(fields ...Field) Logger {
	return &defaultLogger{
		out:       l.out,
		level:     l.level,
		component: l.component,
		fields:    append(l.fields, fields...),
	}
}

// log writes a log entry
func (l *defaultLogger) log(ctx context.Context, level Level, msg string, err error, fields ...Field) {
	entry := Entry{
		Level:     level,
		Message:   msg,
		Time:      time.Now(),
		Component: l.component,
		Error:     err,
		Fields:    append(l.fields, fields...),
	}

	// Add trace information if available
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		entry.TraceID = traceID
	}
	if spanID, ok := ctx.Value("span_id").(string); ok {
		entry.SpanID = spanID
	}

	// Convert entry to JSON
	data, err := json.Marshal(map[string]interface{}{
		"level":     entry.Level.String(),
		"time":      entry.Time.Format(time.RFC3339),
		"msg":       entry.Message,
		"component": entry.Component,
		"trace_id":  entry.TraceID,
		"span_id":   entry.SpanID,
		"fields":    fieldsToMap(entry.Fields),
		"error":     errorToString(entry.Error),
	})
	if err != nil {
		// If JSON marshaling fails, write a simple error message
		fmt.Fprintf(l.out, "failed to marshal log entry: %v\n", err)
		return
	}

	// Write the log entry
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out.Write(append(data, '\n'))
}

// fieldsToMap converts Fields to a map
func fieldsToMap(fields []Field) map[string]interface{} {
	result := make(map[string]interface{}, len(fields))
	for _, f := range fields {
		result[f.Key] = f.Value
	}
	return result
}

// errorToString converts an error to a string
func errorToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
