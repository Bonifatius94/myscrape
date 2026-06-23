package cache

import (
	"testing"
	"time"
)

func TestHitAndMiss(t *testing.T) {
	now := time.Unix(1000, 0)
	c := NewMemory(func() time.Time { return now })

	if _, ok := c.Get("k"); ok {
		t.Fatal("want miss on empty cache")
	}
	c.Set("k", "v", 10*time.Second)
	if v, ok := c.Get("k"); !ok || v != "v" {
		t.Fatalf("want hit v, got %q ok=%v", v, ok)
	}
}

func TestExpiry(t *testing.T) {
	cur := time.Unix(1000, 0)
	c := NewMemory(func() time.Time { return cur })

	c.Set("k", "v", 10*time.Second)
	cur = cur.Add(9 * time.Second)
	if _, ok := c.Get("k"); !ok {
		t.Fatal("should still be live before TTL")
	}
	cur = cur.Add(2 * time.Second) // now 11s > 10s TTL
	if _, ok := c.Get("k"); ok {
		t.Fatal("should be expired after TTL")
	}
}
