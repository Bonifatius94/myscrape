// Package ratelimit is the Tier-1 per-host pacer: it spaces requests to one host
// by a minimum interval (+ jitter) so we never trip a search engine's rate limit.
// The proven safe operating point is 8s + 2s jitter (Python reference E-002).
package ratelimit

import (
	"sync"
	"time"
)

// Clock returns the current time; Sleeper blocks for d. Both injected for tests.
type Clock func() time.Time
type Sleeper func(time.Duration)

// Limiter paces requests per host. Safe for concurrent use.
type Limiter struct {
	mu          sync.Mutex
	now         Clock
	sleep       Sleeper
	rand        func() float64 // [0,1)
	minInterval time.Duration
	jitterMax   time.Duration
	nextAllowed map[string]time.Time
}

// New builds a Limiter. rand returns a value in [0,1) used to scale jitter.
func New(minInterval, jitterMax time.Duration, now Clock, sleep Sleeper, rand func() float64) *Limiter {
	return &Limiter{
		now:         now,
		sleep:       sleep,
		rand:        rand,
		minInterval: minInterval,
		jitterMax:   jitterMax,
		nextAllowed: make(map[string]time.Time),
	}
}

// Acquire blocks until it's polite to make a request to host, then reserves the
// next slot. The reservation happens under lock; the wait is outside it.
func (l *Limiter) Acquire(host string) {
	l.mu.Lock()
	now := l.now()
	target := l.minInterval + time.Duration(l.rand()*float64(l.jitterMax))
	earliest := now
	if next, ok := l.nextAllowed[host]; ok && next.After(now) {
		earliest = next
	}
	l.nextAllowed[host] = earliest.Add(target)
	wait := earliest.Sub(now)
	l.mu.Unlock()

	if wait > 0 {
		l.sleep(wait)
	}
}
