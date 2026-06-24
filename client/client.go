// Package client provides an HTTP client with timeouts and retry/backoff for
// calling upstream services reliably.
package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/imoudgil/servkit/config"
)

// HTTP is a resilient HTTP client wrapper.
type HTTP struct {
	inner   *http.Client
	retries int
}

// New builds an HTTP client from service configuration.
func New(cfg config.Service) *HTTP {
	return &HTTP{
		inner: &http.Client{Timeout: cfg.ClientTimeout},
		retries: cfg.ClientRetries,
	}
}

// Do executes req with retry on transient failures (5xx and network errors).
func (c *HTTP) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	req = req.Clone(ctx)

	var lastErr error
	for attempt := 0; attempt <= c.retries; attempt++ {
		if attempt > 0 {
			if err := sleep(ctx, backoff(attempt)); err != nil {
				return nil, err
			}
		}
		resp, err := c.inner.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode < 500 {
			return resp, nil
		}
		lastErr = fmt.Errorf("upstream status %d", resp.StatusCode)
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
	return nil, fmt.Errorf("request failed after %d retries: %w", c.retries, lastErr)
}

// GetJSON performs GET and returns the raw response for callers to decode.
func (c *HTTP) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(ctx, req)
}

func backoff(attempt int) time.Duration {
	// 100ms, 200ms, 400ms, ...
	d := time.Duration(100*(1<<(attempt-1))) * time.Millisecond
	if d > 2*time.Second {
		return 2 * time.Second
	}
	return d
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
