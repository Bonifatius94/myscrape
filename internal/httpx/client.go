// Package httpx is the one HTTP primitive the search/fetch layers share. It
// composes the three stability layers around net/http:
//
//   - Tier-0 cache  — a hit skips the network entirely;
//   - Tier-1 limiter — per-host pacing before every request;
//   - Tier-3 retry   — transient 202/429/5xx are retried, honoring Retry-After.
//
// The cache, limiter, and retry sleep/rng are injected, so the whole thing is
// testable offline against httptest.
package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Bonifatius94/myscrape-go/internal/cache"
	"github.com/Bonifatius94/myscrape-go/internal/ratelimit"
	"github.com/Bonifatius94/myscrape-go/internal/retry"
)

// retryableStatuses are transient — retried with backoff (Python reference parity).
var retryableStatuses = map[int]bool{202: true, 429: true, 500: true, 502: true, 503: true, 504: true}

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

// Config builds a Client. Now/Sleep/Rand default to the real implementations when
// nil, and are injected in tests for determinism.
type Config struct {
	Timeout     time.Duration
	UserAgent   string
	CacheTTL    time.Duration
	MinInterval time.Duration
	Jitter      time.Duration
	Attempts    int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	RetryJitter time.Duration
	Now         func() time.Time
	Sleep       func(time.Duration)
	Rand        func() float64
}

// Client is the default Doer: cache -> rate limit -> retry over net/http.
type Client struct {
	hc        *http.Client
	userAgent string
	cache     *cache.Memory
	cacheTTL  time.Duration
	limiter   *ratelimit.Limiter
	retryOpts retry.Options
}

// New builds a Client from cfg.
func New(cfg Config) *Client {
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	sleep := cfg.Sleep
	if sleep == nil {
		sleep = time.Sleep
	}
	rng := cfg.Rand
	if rng == nil {
		rng = rand.Float64
	}
	return &Client{
		hc:        &http.Client{Timeout: cfg.Timeout},
		userAgent: cfg.UserAgent,
		cache:     cache.NewMemory(now),
		cacheTTL:  cfg.CacheTTL,
		limiter:   ratelimit.New(cfg.MinInterval, cfg.Jitter, now, sleep, rng),
		retryOpts: retry.Options{
			Attempts: cfg.Attempts,
			Base:     cfg.BaseDelay,
			Max:      cfg.MaxDelay,
			Jitter:   cfg.RetryJitter,
			Sleep:    sleep,
			Rand:     rng,
		},
	}
}

// Get performs a cached, paced, retried GET.
func (c *Client) Get(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	return c.send(ctx, http.MethodGet, url, nil, headers)
}

// PostJSON marshals body and POSTs it as application/json (same stability stack).
func (c *Client) PostJSON(ctx context.Context, url string, body any, headers map[string]string) ([]byte, error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return c.send(ctx, http.MethodPost, url, buf, headers)
}

func (c *Client) send(ctx context.Context, method, rawURL string, body []byte, headers map[string]string) ([]byte, error) {
	// Cache key folds in the body so different POSTs to one URL don't collide.
	key := method + " " + rawURL
	if body != nil {
		key += "|" + string(body)
	}
	if v, ok := c.cache.Get(key); ok {
		return []byte(v), nil
	}

	host := hostOf(rawURL)
	var out []byte
	err := retry.Do(func() error {
		c.limiter.Acquire(host)

		var rdr io.Reader
		if body != nil {
			rdr = bytes.NewReader(body)
		}
		req, err := http.NewRequestWithContext(ctx, method, rawURL, rdr)
		if err != nil {
			return err // request build error: not retryable
		}
		req.Header.Set("User-Agent", c.userAgent)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := c.hc.Do(req)
		if err != nil {
			return err // network error: not retryable (matches the Python reference)
		}
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if retryableStatuses[resp.StatusCode] {
			d, ok := parseRetryAfter(resp.Header.Get("Retry-After"))
			return &retry.Retryable{
				Err:           &HTTPError{Status: resp.StatusCode, Body: string(data)},
				RetryAfter:    d,
				HasRetryAfter: ok,
			}
		}
		if resp.StatusCode >= 400 {
			return &HTTPError{Status: resp.StatusCode, Body: string(data)}
		}
		out = data
		return nil
	}, c.retryOpts)
	if err != nil {
		return nil, err
	}

	c.cache.Set(key, string(out), c.cacheTTL)
	return out, nil
}

// parseRetryAfter parses the integer-seconds form of a Retry-After header.
func parseRetryAfter(v string) (time.Duration, bool) {
	if v == "" {
		return 0, false
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, false // HTTP-date form: fall back to computed backoff
	}
	return time.Duration(n) * time.Second, true
}

func hostOf(raw string) string {
	if u, err := url.Parse(raw); err == nil {
		return u.Host
	}
	return ""
}
