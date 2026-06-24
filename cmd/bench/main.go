// Command bench is the live stability gate: it runs a set of queries through the
// configured search provider at the operating point and reports the rate-limit
// rate. The gate passes when ratelimit_rate == 0 (the Python reference's pass
// condition). Exit code is non-zero if the gate fails.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Bonifatius94/myscrape/internal/config"
	"github.com/Bonifatius94/myscrape/internal/httpx"
	"github.com/Bonifatius94/myscrape/internal/retry"
	"github.com/Bonifatius94/myscrape/internal/search"
)

var queries = []string{
	"python asyncio taskgroup",
	"go context cancellation pattern",
	"rust ownership borrow checker",
	"sqlite write-ahead logging",
	"http retry backoff jitter",
	"bm25 ranking formula",
}

type outcome struct {
	query       string
	dur         time.Duration
	rateLimited bool
	ok          bool
	results     int
}

type report struct {
	Total         int
	OK            int
	RateLimited   int
	Failed        int
	RateLimitRate float64
	SuccessRate   float64
}

func summarize(outcomes []outcome) report {
	r := report{Total: len(outcomes)}
	for _, o := range outcomes {
		switch {
		case o.rateLimited:
			r.RateLimited++
		case o.ok:
			r.OK++
		default:
			r.Failed++
		}
	}
	if r.Total > 0 {
		r.RateLimitRate = float64(r.RateLimited) / float64(r.Total)
		r.SuccessRate = float64(r.OK) / float64(r.Total)
	}
	return r
}

// passed is the gate: no request was rate-limited.
func (r report) passed() bool { return r.Total > 0 && r.RateLimited == 0 }

func isRateLimited(err error) bool {
	var re *retry.Retryable
	if errors.As(err, &re) {
		return true
	}
	var he *httpx.HTTPError
	return errors.As(err, &he) && (he.Status == 429 || he.Status == 503 || he.Status == 202)
}

func main() {
	config.LoadDotEnv(".env")
	cfg := config.FromEnv()
	client := httpx.New(httpx.Config{
		Timeout: cfg.RequestTimeout, UserAgent: cfg.UserAgent, CacheTTL: cfg.CacheTTL,
		MinInterval: cfg.RequestMinInterval, Jitter: cfg.RequestJitter,
		Attempts: cfg.HTTPAttempts, BaseDelay: cfg.HTTPBaseDelay,
		MaxDelay: cfg.HTTPMaxDelay, RetryJitter: cfg.HTTPRetryJitter,
	})
	provider, err := search.Build(client, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	fmt.Printf("bench: %d queries via %q (min-interval %s)\n", len(queries), provider.Name(), cfg.RequestMinInterval)
	var outcomes []outcome
	for _, q := range queries {
		t0 := time.Now()
		res, err := provider.Search(context.Background(), q, 5)
		o := outcome{query: q, dur: time.Since(t0)}
		if err != nil {
			o.rateLimited = isRateLimited(err)
		} else {
			o.ok = true
			o.results = len(res)
		}
		status := "ok"
		if o.rateLimited {
			status = "RATE-LIMITED"
		} else if !o.ok {
			status = "FAILED"
		}
		fmt.Printf("  %-34s %5s  %4d results  %s\n", q, status, o.results, o.dur.Round(time.Millisecond))
		outcomes = append(outcomes, o)
	}

	r := summarize(outcomes)
	fmt.Printf("\n%d total | ok %d | rate-limited %d | failed %d\n", r.Total, r.OK, r.RateLimited, r.Failed)
	fmt.Printf("ratelimit_rate %.2f | success_rate %.2f | gate: %s\n",
		r.RateLimitRate, r.SuccessRate, gateLabel(r.passed()))
	if !r.passed() {
		os.Exit(1)
	}
}

func gateLabel(pass bool) string {
	if pass {
		return "PASS"
	}
	return "FAIL"
}
