package search

import "testing"

func TestParseMarginalia(t *testing.T) {
	body := []byte(`{"results":[
		{"url":"https://a.com/x","title":"Async","description":"a snippet"},
		{"url":"https://docs.python.org/3","title":"Docs","description":"more"},
		{"url":"","title":"dropped","description":"no url"}
	]}`)

	got, err := parseMarginalia(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 results (empty-url dropped), got %d", len(got))
	}
	if got[0].URL != "https://a.com/x" || got[0].Site != "a.com" {
		t.Errorf("result[0] = %+v", got[0])
	}
	if got[0].Snippet != "a snippet" {
		t.Errorf("snippet = %q", got[0].Snippet)
	}
	if got[0].Rank != 1 || got[1].Rank != 2 {
		t.Errorf("ranks = %d, %d", got[0].Rank, got[1].Rank)
	}
}

func TestParseMarginaliaEmpty(t *testing.T) {
	got, err := parseMarginalia([]byte(`{"results":[]}`))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("want 0, got %d", len(got))
	}
}
