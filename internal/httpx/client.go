// Package httpx is the one HTTP primitive the search/fetch layers share.
//
// Phase 1 is a plain net/http wrapper (GET + JSON POST, custom headers). The
// stability layers from the Python reference — Tier-0 cache, Tier-1 per-host rate
// limiter, Tier-3 retry/backoff honoring Retry-After — wrap this next (see
// specs/PORTING_GO.md). The Doer interface is the seam they slot behind, so
// providers never change.
package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Doer is the HTTP seam used by every provider/fetcher. Mockable in tests.
type Doer interface {
	Get(ctx context.Context, url string, headers map[string]string) ([]byte, error)
	PostJSON(ctx context.Context, url string, body any, headers map[string]string) ([]byte, error)
}

// HTTPError carries a non-2xx status so callers can map it (e.g. 429 -> RATELIMITED).
type HTTPError struct {
	Status int
	Body   string
}

func (e *HTTPError) Error() string { return fmt.Sprintf("http %d", e.Status) }

// Client is the default Doer over net/http.
type Client struct {
	hc        *http.Client
	userAgent string
}

// New returns a Client with the given timeout and User-Agent.
func New(timeout time.Duration, userAgent string) *Client {
	return &Client{hc: &http.Client{Timeout: timeout}, userAgent: userAgent}
}

// Get performs a GET and returns the body, or an *HTTPError on non-2xx.
func (c *Client) Get(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req, headers)
}

// PostJSON marshals body and POSTs it as application/json.
func (c *Client) PostJSON(ctx context.Context, url string, body any, headers map[string]string) ([]byte, error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, headers)
}

func (c *Client) do(req *http.Request, headers map[string]string) ([]byte, error) {
	req.Header.Set("User-Agent", c.userAgent)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, &HTTPError{Status: resp.StatusCode, Body: string(data)}
	}
	return data, nil
}
