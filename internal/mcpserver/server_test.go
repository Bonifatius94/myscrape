package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Bonifatius94/myscrape/internal/fetch"
	"github.com/Bonifatius94/myscrape/internal/httpx"
	"github.com/Bonifatius94/myscrape/internal/research"
	"github.com/Bonifatius94/myscrape/internal/retry"
	"github.com/Bonifatius94/myscrape/internal/search"
)

type stubSearch struct {
	results []search.Result
	err     error
}

func (s stubSearch) Search(context.Context, string, int) ([]search.Result, error) {
	return s.results, s.err
}

type stubFetch struct {
	res *fetch.Result
	err error
}

func (f stubFetch) Fetch(context.Context, string) (*fetch.Result, error) { return f.res, f.err }

type stubResearch struct {
	res research.ResearchResult
	err error
}

func (r stubResearch) Research(context.Context, string, string, string) (research.ResearchResult, error) {
	return r.res, r.err
}

type timeoutErr struct{}

func (timeoutErr) Error() string   { return "i/o timeout" }
func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return true }

func errCode(out any) string {
	m, ok := out.(map[string]any)
	if !ok {
		return ""
	}
	e, ok := m["error"].(map[string]any)
	if !ok {
		return ""
	}
	return e["code"].(string)
}

func TestDoWebSearch(t *testing.T) {
	out := doWebSearch(context.Background(), stubSearch{results: []search.Result{{Rank: 1, URL: "http://a"}}}, "q", 5)
	so, ok := out.(searchOut)
	if !ok || so.Count != 1 || so.Results[0].URL != "http://a" {
		t.Fatalf("got %#v", out)
	}
}

func TestDoWebSearchRateLimited(t *testing.T) {
	out := doWebSearch(context.Background(), stubSearch{err: &retry.Retryable{Err: errors.New("429")}}, "q", 5)
	if errCode(out) != "RATELIMITED" {
		t.Fatalf("want RATELIMITED, got %v", out)
	}
}

func TestDoWebFetchTruncates(t *testing.T) {
	res := &fetch.Result{URL: "u", Title: "T", Content: "one two three four", WordCount: 4, FetchedVia: "static"}
	out := doWebFetch(context.Background(), stubFetch{res: res}, "u", 2)
	fo := out.(fetchOut)
	if !fo.Truncated || len(strings.Fields(fo.Content)) != 2 {
		t.Fatalf("got %#v", fo)
	}
}

func TestDoWebFetchErrorMapping(t *testing.T) {
	cases := map[error]string{
		&httpx.HTTPError{Status: 403}:            "BLOCKED",
		&httpx.HTTPError{Status: 500}:            "UNREACHABLE",
		fetch.ErrEmpty:                           "EMPTY_CONTENT",
		timeoutErr{}:                             "TIMEOUT",
		&retry.Retryable{Err: errors.New("429")}: "RATELIMITED",
	}
	for err, want := range cases {
		out := doWebFetch(context.Background(), stubFetch{err: err}, "u", 0)
		if got := errCode(out); got != want {
			t.Errorf("error %v -> %q, want %q", err, got, want)
		}
	}
}

func TestDoWebResearchReturnModes(t *testing.T) {
	res := research.ResearchResult{
		Answer:    "a [1]",
		Citations: []research.Citation{{Marker: "[1]", URL: "http://a"}},
		Sources:   []research.Source{{URL: "http://a"}},
		Coverage:  research.Coverage{Confidence: "low"},
	}
	out := doWebResearch(context.Background(), stubResearch{res: res}, researchParams{Question: "q", ReturnMode: "both"}).(map[string]any)
	if out["answer"] != "a [1]" || out["sources"] == nil {
		t.Fatalf("both mode wrong: %#v", out)
	}
	ansOnly := doWebResearch(context.Background(), stubResearch{res: res}, researchParams{Question: "q"}).(map[string]any)
	if _, ok := ansOnly["sources"]; ok {
		t.Errorf("answer mode should omit sources")
	}
}

func TestDoWebResearchLLMError(t *testing.T) {
	out := doWebResearch(context.Background(), stubResearch{err: fmt.Errorf("%w: boom", research.ErrLLM)}, researchParams{Question: "q"})
	if errCode(out) != "LLM_ERROR" {
		t.Fatalf("want LLM_ERROR, got %v", out)
	}
}
