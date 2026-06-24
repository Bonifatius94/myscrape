package research

import (
	"context"
	"strings"
	"testing"

	"github.com/Bonifatius94/myscrape/internal/fetch"
	"github.com/Bonifatius94/myscrape/internal/search"
)

type fakeSearcher struct {
	results []search.Result
	maxSeen int
}

func (f *fakeSearcher) Search(_ context.Context, _ string, max int) ([]search.Result, error) {
	f.maxSeen = max
	if max < len(f.results) {
		return f.results[:max], nil
	}
	return f.results, nil
}

type fakeFetcher struct{ byURL map[string]string }

func (f *fakeFetcher) Fetch(_ context.Context, url string) (*fetch.Result, error) {
	md, ok := f.byURL[url]
	if !ok {
		return nil, fetch.ErrEmpty
	}
	return &fetch.Result{URL: url, Content: md, WordCount: len(strings.Fields(md))}, nil
}

func sr(rank int, title, url string) search.Result {
	return search.Result{Rank: rank, Title: title, URL: url, Site: url}
}

func TestResearchFullLoop(t *testing.T) {
	s := &fakeSearcher{results: []search.Result{sr(1, "A", "http://a"), sr(2, "B", "http://b")}}
	f := &fakeFetcher{byURL: map[string]string{
		"http://a": "python asyncio runs tasks concurrently without threads",
		"http://b": "task groups supervise child tasks and propagate errors",
	}}
	wr := NewWebResearcher(s, f, nil, "simple", 0)

	res, err := wr.Research(context.Background(), "python asyncio", "standard", "")
	if err != nil {
		t.Fatalf("research: %v", err)
	}
	if res.Coverage.Confidence != "medium" { // two sources fetched
		t.Errorf("coverage = %+v", res.Coverage)
	}
	urls := map[string]bool{}
	for _, c := range res.Citations {
		urls[c.URL] = true
		if c.URL == "http://a" && c.Title != "A" {
			t.Errorf("citation not title-enriched: %+v", c)
		}
	}
	if !urls["http://a"] || !urls["http://b"] {
		t.Errorf("expected both sources cited, got %v", urls)
	}
	if len(res.Sources) != 2 {
		t.Errorf("want 2 sources, got %d", len(res.Sources))
	}
}

func TestEffortControlsSourceCount(t *testing.T) {
	results := make([]search.Result, 10)
	for i := range results {
		results[i] = sr(i+1, "t", "http://x")
	}
	s := &fakeSearcher{results: results}
	wr := NewWebResearcher(s, &fakeFetcher{byURL: map[string]string{}}, nil, "simple", 0)

	_, _ = wr.Research(context.Background(), "q", "quick", "")
	if s.maxSeen != 3 { // quick -> 3 sources
		t.Fatalf("quick should fetch 3 sources, maxSeen=%d", s.maxSeen)
	}
}

type fakeSynth struct{ called bool }

func (f *fakeSynth) Synthesize(_ context.Context, _ string, _ []Passage) (SynthesisResult, error) {
	f.called = true
	return SynthesisResult{
		Answer:    "llm answer [1]",
		Citations: []Citation{{Marker: "[1]", URL: "http://a"}},
		Coverage:  Coverage{Confidence: "low", Note: "single source"},
	}, nil
}

func TestResearchUsesLLMSynthesizerWhenRequested(t *testing.T) {
	s := &fakeSearcher{results: []search.Result{sr(1, "A", "http://a")}}
	f := &fakeFetcher{byURL: map[string]string{"http://a": "content about async tasks"}}
	synth := &fakeSynth{}
	wr := NewWebResearcher(s, f, synth, "simple", 0)

	res, err := wr.Research(context.Background(), "q", "standard", "llm")
	if err != nil {
		t.Fatalf("research: %v", err)
	}
	if !synth.called {
		t.Fatal("expected the LLM synthesizer to be used for synthesis=llm")
	}
	if res.Answer != "llm answer [1]" {
		t.Errorf("answer = %q", res.Answer)
	}
}

func TestUnfetchableSourcesSkipped(t *testing.T) {
	s := &fakeSearcher{results: []search.Result{sr(1, "A", "http://a"), sr(2, "B", "http://b")}}
	f := &fakeFetcher{byURL: map[string]string{"http://a": "real content about async tasks"}} // b missing
	wr := NewWebResearcher(s, f, nil, "simple", 0)

	res, _ := wr.Research(context.Background(), "q", "standard", "")
	if len(res.Sources) != 1 || res.Sources[0].URL != "http://a" {
		t.Fatalf("only fetchable source should remain, got %+v", res.Sources)
	}
	if res.Coverage.Confidence != "low" { // single source
		t.Errorf("coverage = %+v", res.Coverage)
	}
}
