package httpx

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func testClient() *Client {
	return New(Config{
		Timeout:     5 * time.Second,
		UserAgent:   "test",
		CacheTTL:    900 * time.Second,
		MinInterval: 0,
		Jitter:      0,
		Attempts:    4,
		BaseDelay:   time.Millisecond,
		MaxDelay:    time.Millisecond,
		RetryJitter: 0,
		Sleep:       func(time.Duration) {}, // don't actually wait in tests
		Rand:        func() float64 { return 0 },
	})
}

func TestGetCachesSecondCall(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&hits, 1)
		_, _ = w.Write([]byte("body"))
	}))
	defer srv.Close()

	c := testClient()
	b1, err1 := c.Get(context.Background(), srv.URL, nil)
	b2, err2 := c.Get(context.Background(), srv.URL, nil)
	if err1 != nil || err2 != nil || string(b1) != "body" || string(b2) != "body" {
		t.Fatalf("b1=%q b2=%q err1=%v err2=%v", b1, b2, err1, err2)
	}
	if got := atomic.LoadInt32(&hits); got != 1 {
		t.Fatalf("cache should prevent a 2nd network hit, hits=%d", got)
	}
}

func TestRetriesOn503ThenSucceeds(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&hits, 1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	c := testClient()
	b, err := c.Get(context.Background(), srv.URL, nil)
	if err != nil || string(b) != "ok" {
		t.Fatalf("b=%q err=%v", b, err)
	}
	if got := atomic.LoadInt32(&hits); got != 3 {
		t.Fatalf("want 3 attempts (2x503 then 200), got %d", got)
	}
}

func TestNonRetryableStatusReturnsHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := testClient()
	_, err := c.Get(context.Background(), srv.URL, nil)
	var he *HTTPError
	if !errors.As(err, &he) || he.Status != 404 {
		t.Fatalf("want *HTTPError 404, got %v", err)
	}
}
