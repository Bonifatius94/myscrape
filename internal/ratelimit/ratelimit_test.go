package ratelimit

import (
	"testing"
	"time"
)

// harness wires an injected clock + sleeper; sleeping advances the clock, like
// real time passing.
type harness struct {
	now   time.Time
	slept []time.Duration
}

func (h *harness) clock() time.Time { return h.now }
func (h *harness) sleep(d time.Duration) {
	h.slept = append(h.slept, d)
	h.now = h.now.Add(d)
}

func TestFirstRequestDoesNotWait(t *testing.T) {
	h := &harness{now: time.Unix(0, 0)}
	l := New(8*time.Second, 2*time.Second, h.clock, h.sleep, func() float64 { return 0 })

	l.Acquire("a.com")

	if len(h.slept) != 0 {
		t.Fatalf("first request should not wait, slept=%v", h.slept)
	}
}

func TestSecondRequestWaitsMinInterval(t *testing.T) {
	h := &harness{now: time.Unix(0, 0)}
	l := New(8*time.Second, 2*time.Second, h.clock, h.sleep, func() float64 { return 0 }) // no jitter

	l.Acquire("a.com")
	l.Acquire("a.com")

	if len(h.slept) != 1 || h.slept[0] != 8*time.Second {
		t.Fatalf("want one 8s wait, slept=%v", h.slept)
	}
}

func TestJitterIsAdded(t *testing.T) {
	h := &harness{now: time.Unix(0, 0)}
	l := New(8*time.Second, 2*time.Second, h.clock, h.sleep, func() float64 { return 0.5 })

	l.Acquire("a.com")
	l.Acquire("a.com")

	if len(h.slept) != 1 || h.slept[0] != 9*time.Second { // 8 + 0.5*2
		t.Fatalf("want 9s (8 + jitter), slept=%v", h.slept)
	}
}

func TestHostsAreIndependent(t *testing.T) {
	h := &harness{now: time.Unix(0, 0)}
	l := New(8*time.Second, 2*time.Second, h.clock, h.sleep, func() float64 { return 0 })

	l.Acquire("a.com")
	l.Acquire("b.com") // different host: must not wait on a.com's slot

	if len(h.slept) != 0 {
		t.Fatalf("distinct hosts are independent, slept=%v", h.slept)
	}
}
