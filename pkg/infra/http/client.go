package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"order-system/pkg/infra/config"
)

// defaultClient represents the default HTTP client implementation
type defaultClient struct {
	client  *http.Client
	config  *config.Config
	baseURL string
}

// NewClient creates a new HTTP client
func NewClient(cfg *config.Config, baseURL string) Client {
	client := &http.Client{
		Timeout: cfg.HTTP.RequestTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			MaxConnsPerHost:     100,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return &defaultClient{
		client:  client,
		config:  cfg,
		baseURL: baseURL,
	}
}

// Get performs a GET request
func (c *defaultClient) Get(ctx context.Context, url string, opt *RequestOption) (*Response, error) {
	return c.do(ctx, http.MethodGet, url, nil, opt)
}

// Post performs a POST request
func (c *defaultClient) Post(ctx context.Context, url string, body []byte, opt *RequestOption) (*Response, error) {
	return c.do(ctx, http.MethodPost, url, body, opt)
}

// Put performs a PUT request
func (c *defaultClient) Put(ctx context.Context, url string, body []byte, opt *RequestOption) (*Response, error) {
	return c.do(ctx, http.MethodPut, url, body, opt)
}

// Delete performs a DELETE request
func (c *defaultClient) Delete(ctx context.Context, url string, opt *RequestOption) (*Response, error) {
	return c.do(ctx, http.MethodDelete, url, nil, opt)
}

// do performs the HTTP request with retries
func (c *defaultClient) do(ctx context.Context, method, url string, body []byte, opt *RequestOption) (*Response, error) {
	if opt == nil {
		opt = &RequestOption{
			Timeout:       c.config.HTTP.RequestTimeout,
			RetryCount:    3,
			RetryInterval: time.Second,
			MaxBodySize:   c.config.HTTP.MaxRequestSize,
		}
	}

	var resp *Response
	var lastErr error

	for i := 0; i <= opt.RetryCount; i++ {
		resp, lastErr = c.doRequest(ctx, method, url, body, opt)
		if lastErr == nil {
			return resp, nil
		}

		// Check if we should retry
		if !c.shouldRetry(lastErr) || i == opt.RetryCount {
			break
		}

		// Wait before retrying
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(opt.RetryInterval):
			continue
		}
	}

	return nil, lastErr
}

// doRequest performs a single HTTP request
func (c *defaultClient) doRequest(ctx context.Context, method, url string, body []byte, opt *RequestOption) (*Response, error) {
	fullURL := c.baseURL + url
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bytes.NewReader(body))
	if err != nil {
		return nil, &Error{
			Message: "failed to create request",
			Cause:   err,
		}
	}

	// Add headers
	for k, v := range opt.Headers {
		req.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, &Error{
			Message: "request failed",
			Cause:   err,
		}
	}
	defer resp.Body.Close()

	// Check response size
	if resp.ContentLength > opt.MaxBodySize {
		return nil, &Error{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("response body too large: %d bytes", resp.ContentLength),
		}
	}

	// Read response body
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, opt.MaxBodySize))
	if err != nil {
		return nil, &Error{
			StatusCode: resp.StatusCode,
			Message:    "failed to read response body",
			Cause:      err,
		}
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header,
		Duration:   time.Since(start),
	}, nil
}

// shouldRetry determines if a request should be retried
func (c *defaultClient) shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a network error
	if _, ok := err.(*Error); !ok {
		return true
	}

	// Check if it's a server error (5xx)
	if httpErr, ok := err.(*Error); ok && httpErr.StatusCode >= 500 {
		return true
	}

	return false
}
