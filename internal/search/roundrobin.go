package search

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Bonifatius94/myscrape/internal/httpx"
	"github.com/Bonifatius94/myscrape/internal/retry"
)

// RoundRobin composes independent engines: it rotates the starting engine per
// query (spreading load) and a per-engine circuit breaker cools any engine that
// rate-limits, skipping it until the cooldown elapses. A healthy-but-empty result
// is returned as empty (not an error) so a cooling engine's failure never surfaces.
type RoundRobin struct {
	mu          sync.Mutex
	providers   []Provider
	now         func() time.Time
	cooldown    time.Duration
	cooledUntil map[string]time.Time
	cursor      int
}

// NewRoundRobin builds the composite. A nil now defaults to time.Now.
func NewRoundRobin(providers []Provider, now func() time.Time, cooldown time.Duration) *RoundRobin {
	if now == nil {
		now = time.Now
	}
	return &RoundRobin{
		providers:   providers,
		now:         now,
		cooldown:    cooldown,
		cooledUntil: make(map[string]time.Time),
	}
}

func (r *RoundRobin) Name() string { return "roundrobin" }

func (r *RoundRobin) Search(ctx context.Context, query string, maxResults int) ([]Result, error) {
	var lastErr error
	anyResponded := false
	for _, p := range r.nextOrder() {
		if r.isCooled(p.Name()) {
			continue
		}
		results, err := p.Search(ctx, query, maxResults)
		if err != nil {
			if isTransient(err) {
				r.cool(p.Name())
			}
			lastErr = err
			continue
		}
		anyResponded = true
		if len(results) > 0 {
			return results, nil
		}
	}
	if anyResponded {
		return nil, nil // an engine answered, just with nothing
	}
	return nil, lastErr // every engine errored or was cooled
}

// nextOrder returns the providers rotated to start at the cursor, advancing it.
func (r *RoundRobin) nextOrder() []Provider {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := len(r.providers)
	if n == 0 {
		return nil
	}
	start := r.cursor % n
	r.cursor = (r.cursor + 1) % n
	order := make([]Provider, 0, n)
	order = append(order, r.providers[start:]...)
	order = append(order, r.providers[:start]...)
	return order
}

func (r *RoundRobin) isCooled(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	until, ok := r.cooledUntil[name]
	return ok && r.now().Before(until)
}

func (r *RoundRobin) cool(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cooledUntil[name] = r.now().Add(r.cooldown)
}

// isTransient reports whether an error is a rate-limit/transient signal worth
// cooling the engine for (retry-exhausted Retryable, or a 429/503/202 status).
func isTransient(err error) bool {
	var re *retry.Retryable
	if errors.As(err, &re) {
		return true
	}
	var he *httpx.HTTPError
	if errors.As(err, &he) {
		return he.Status == 429 || he.Status == 503 || he.Status == 202
	}
	return false
}
