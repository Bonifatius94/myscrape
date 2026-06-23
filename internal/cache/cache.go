// Package cache is the Tier-0 response cache: a hit skips the network entirely.
package cache

import (
	"sync"
	"time"
)

// Clock returns the current time; injected so expiry is deterministic in tests.
type Clock func() time.Time

type entry struct {
	value   string
	expires time.Time
}

// Memory is a goroutine-safe in-memory TTL cache.
type Memory struct {
	mu    sync.Mutex
	now   Clock
	items map[string]entry
}

// NewMemory builds a cache that reads time from now.
func NewMemory(now Clock) *Memory {
	return &Memory{now: now, items: make(map[string]entry)}
}

// Get returns the cached value and true if present and unexpired.
func (c *Memory) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.items[key]
	if !ok {
		return "", false
	}
	if !c.now().Before(e.expires) { // now >= expires -> stale
		delete(c.items, key)
		return "", false
	}
	return e.value, true
}

// Set stores value under key for the given TTL.
func (c *Memory) Set(key, value string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = entry{value: value, expires: c.now().Add(ttl)}
}
