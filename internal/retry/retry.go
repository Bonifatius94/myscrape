// Package retry is the Tier-3 backoff: transient failures (202/429/5xx) are
// retried with exponential backoff + jitter, honoring an integer Retry-After.
// Whether a failure is transient is decided by the caller, which wraps it in a
// *Retryable (the httpx layer maps retryable HTTP statuses to one).
package retry

import (
	"errors"
	"time"
)

// Retryable marks an error as transient. RetryAfter (when HasRetryAfter) overrides
// the computed backoff for the next wait.
type Retryable struct {
	Err           error
	RetryAfter    time.Duration
	HasRetryAfter bool
}

func (r *Retryable) Error() string { return r.Err.Error() }
func (r *Retryable) Unwrap() error { return r.Err }

// Options configures Do. Sleep and Rand are injected for deterministic tests.
type Options struct {
	Attempts int
	Base     time.Duration
	Max      time.Duration
	Jitter   time.Duration
	Sleep    func(time.Duration)
	Rand     func() float64 // [0,1)
}

// Do calls fn up to opt.Attempts times. It retries only while fn returns an error
// that unwraps to *Retryable; any other error stops immediately. Returns the last
// error (or nil on success).
func Do(fn func() error, opt Options) error {
	var err error
	for attempt := 0; attempt < opt.Attempts; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}
		var re *Retryable
		if !errors.As(err, &re) {
			return err // non-retryable: stop now
		}
		if attempt == opt.Attempts-1 {
			break // out of attempts; return the error below
		}
		opt.Sleep(backoff(attempt, opt, re))
	}
	return err
}

func backoff(attempt int, opt Options, re *Retryable) time.Duration {
	if re.HasRetryAfter {
		return re.RetryAfter
	}
	d := opt.Base << attempt // base * 2^attempt
	if d > opt.Max {
		d = opt.Max
	}
	return d + time.Duration(opt.Rand()*float64(opt.Jitter))
}
