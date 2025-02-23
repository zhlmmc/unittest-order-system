package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// Error represents a custom error with stack trace and error code
type Error struct {
	Err      error
	Code     string
	Message  string
	Stack    string
	Metadata map[string]interface{}
}

// New creates a new Error
func New(code string, message string) *Error {
	return &Error{
		Code:     code,
		Message:  message,
		Stack:    getStackTrace(),
		Metadata: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code string, message string) *Error {
	if err == nil {
		return nil
	}

	return &Error{
		Err:      err,
		Code:     code,
		Message:  message,
		Stack:    getStackTrace(),
		Metadata: make(map[string]interface{}),
	}
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// WithMetadata adds metadata to the error
func (e *Error) WithMetadata(key string, value interface{}) *Error {
	e.Metadata[key] = value
	return e
}

// getStackTrace returns the stack trace as a string
func getStackTrace() string {
	var sb strings.Builder
	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		sb.WriteString(fmt.Sprintf("%s:%d\n", file, line))
	}
	return sb.String()
}
