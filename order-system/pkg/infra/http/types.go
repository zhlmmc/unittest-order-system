package http

import (
	"context"
	"time"
)

// RequestOption represents options for a request
type RequestOption struct {
	Timeout       time.Duration
	RetryCount    int
	RetryInterval time.Duration
	MaxBodySize   int64
	Headers       map[string]string
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Body       []byte
	Headers    map[string][]string
	Duration   time.Duration
}

// Error represents an HTTP error
type Error struct {
	StatusCode int
	Message    string
	Cause      error
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Client interface defines the HTTP client behavior
type Client interface {
	Get(ctx context.Context, url string, opt *RequestOption) (*Response, error)
	Post(ctx context.Context, url string, body []byte, opt *RequestOption) (*Response, error)
	Put(ctx context.Context, url string, body []byte, opt *RequestOption) (*Response, error)
	Delete(ctx context.Context, url string, opt *RequestOption) (*Response, error)
}
