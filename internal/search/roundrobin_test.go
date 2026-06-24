package search

import (
	"context"
	"testing"
	"time"

	"github.com/Bonifatius94/myscrape/internal/httpx"
)

type fakeProvider struct {
	name    string
	results []Result
	err     error
	calls   int
}

func (f *fakeProvider) Name() string { return f.name }
func (f *fakeProvider) Search(_ context.Context, _ string, _ int) ([]Result, error) {
	f.calls++
	return f.results, f.err
}

func res(url string) []Result { return []Result{{Rank: 1, URL: url}} }

func TestRoundRobinRotatesAcrossEngines(t *testing.T) {
	p0 := &fakeProvider{name: "p0", results: res("http://a")}
	p1 := &fakeProvider{name: "p1", results: res("http://b")}
	rr := NewRoundRobin([]Provider{p0, p1}, nil, time.Minute)

	first, _ := rr.Search(context.Background(), "q", 5)
	second, _ := rr.Search(context.Background(), "q", 5)

	if first[0].URL != "http://a" || second[0].URL != "http://b" {
		t.Fatalf("expected rotation a then b, got %q then %q", first[0].URL, second[0].URL)
	}
}

func TestCircuitBreakerCoolsRateLimitedEngine(t *testing.T) {
	now := time.Unix(0, 0)
	p0 := &fakeProvider{name: "p0", err: &httpx.HTTPError{Status: 429}}
	p1 := &fakeProvider{name: "p1", results: nil} // healthy but empty
	rr := NewRoundRobin([]Provider{p0, p1}, func() time.Time { return now }, time.Minute)

	rr.Search(context.Background(), "q", 5) // p0 rate-limited -> cooled
	rr.Search(context.Background(), "q", 5) // p0 must be skipped

	if p0.calls != 1 {
		t.Fatalf("rate-limited engine should be cooled and skipped, p0.calls=%d", p0.calls)
	}
}

func TestCooldownExpiresAndEngineRetried(t *testing.T) {
	now := time.Unix(0, 0)
	p0 := &fakeProvider{name: "p0", err: &httpx.HTTPError{Status: 429}}
	p1 := &fakeProvider{name: "p1", results: nil}
	rr := NewRoundRobin([]Provider{p0, p1}, func() time.Time { return now }, time.Minute)

	rr.Search(context.Background(), "q", 5) // cools p0 until now+60s
	now = now.Add(61 * time.Second)         // past the cooldown
	rr.Search(context.Background(), "q", 5) // p0 should be retried

	if p0.calls != 2 {
		t.Fatalf("cooled engine should be retried after cooldown, p0.calls=%d", p0.calls)
	}
}

func TestGracefulEmptyWhenAllEnginesEmpty(t *testing.T) {
	p0 := &fakeProvider{name: "p0"}
	p1 := &fakeProvider{name: "p1"}
	rr := NewRoundRobin([]Provider{p0, p1}, nil, time.Minute)

	got, err := rr.Search(context.Background(), "q", 5)
	if err != nil || len(got) != 0 {
		t.Fatalf("healthy-but-empty should be (nil, nil), got %v err=%v", got, err)
	}
}
