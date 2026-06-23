package retry

import (
	"errors"
	"testing"
	"time"
)

func opts(slept *[]time.Duration) Options {
	return Options{
		Attempts: 4,
		Base:     1 * time.Second,
		Max:      8 * time.Second,
		Jitter:   0, // deterministic
		Sleep:    func(d time.Duration) { *slept = append(*slept, d) },
		Rand:     func() float64 { return 0 },
	}
}

func TestSucceedsFirstTryNoSleep(t *testing.T) {
	var slept []time.Duration
	calls := 0
	err := Do(func() error { calls++; return nil }, opts(&slept))
	if err != nil || calls != 1 || len(slept) != 0 {
		t.Fatalf("err=%v calls=%d slept=%v", err, calls, slept)
	}
}

func TestRetriesThenSucceeds(t *testing.T) {
	var slept []time.Duration
	calls := 0
	err := Do(func() error {
		calls++
		if calls < 3 {
			return &Retryable{Err: errors.New("429")}
		}
		return nil
	}, opts(&slept))
	if err != nil || calls != 3 {
		t.Fatalf("err=%v calls=%d", err, calls)
	}
	// exponential base*2^attempt: 1s, 2s
	if len(slept) != 2 || slept[0] != 1*time.Second || slept[1] != 2*time.Second {
		t.Fatalf("slept=%v", slept)
	}
}

func TestNonRetryableStopsImmediately(t *testing.T) {
	var slept []time.Duration
	calls := 0
	sentinel := errors.New("boom")
	err := Do(func() error { calls++; return sentinel }, opts(&slept))
	if !errors.Is(err, sentinel) || calls != 1 || len(slept) != 0 {
		t.Fatalf("err=%v calls=%d slept=%v", err, calls, slept)
	}
}

func TestExhaustsAttempts(t *testing.T) {
	var slept []time.Duration
	calls := 0
	err := Do(func() error { calls++; return &Retryable{Err: errors.New("503")} }, opts(&slept))
	if err == nil || calls != 4 || len(slept) != 3 { // 4 tries, 3 waits
		t.Fatalf("err=%v calls=%d slept=%v", err, calls, slept)
	}
}

func TestBackoffCapsAtMax(t *testing.T) {
	var slept []time.Duration
	o := opts(&slept)
	Do(func() error { return &Retryable{Err: errors.New("x")} }, o)
	// 1s, 2s, 4s — none exceeds Max=8s
	if slept[2] != 4*time.Second {
		t.Fatalf("slept=%v", slept)
	}
}

func TestRetryAfterIsHonored(t *testing.T) {
	var slept []time.Duration
	calls := 0
	err := Do(func() error {
		calls++
		if calls == 1 {
			return &Retryable{Err: errors.New("429"), RetryAfter: 5 * time.Second, HasRetryAfter: true}
		}
		return nil
	}, opts(&slept))
	if err != nil || len(slept) != 1 || slept[0] != 5*time.Second {
		t.Fatalf("err=%v slept=%v", err, slept)
	}
}
