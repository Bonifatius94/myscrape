package search

import (
	"context"
	"testing"
)

func TestParseTavily(t *testing.T) {
	body := []byte(`{"results":[
		{"title":"A","url":"https://a.com/x","content":"a snippet","score":0.9},
		{"title":"B","url":"https://docs.python.org/3","content":"more"},
		{"title":"drop","url":"","content":"no url"}
	]}`)
	got, err := parseTavily(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 (empty-url dropped), got %d", len(got))
	}
	if got[0].URL != "https://a.com/x" || got[0].Site != "a.com" || got[0].Snippet != "a snippet" {
		t.Errorf("result[0] = %+v", got[0])
	}
}

// fakeDoer records the last request so we can assert on headers/body.
type fakeDoer struct {
	body    []byte
	url     string
	headers map[string]string
	gotBody any
}

func (f *fakeDoer) Get(_ context.Context, url string, h map[string]string) ([]byte, error) {
	f.url, f.headers = url, h
	return f.body, nil
}

func (f *fakeDoer) PostJSON(_ context.Context, url string, body any, h map[string]string) ([]byte, error) {
	f.url, f.headers, f.gotBody = url, h, body
	return f.body, nil
}

func TestTavilySendsBearerKey(t *testing.T) {
	doer := &fakeDoer{body: []byte(`{"results":[{"title":"A","url":"https://a.com","content":"s"}]}`)}
	p := NewTavily(doer, "tvly-abc")

	got, err := p.Search(context.Background(), "python async", 5)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 result, got %d", len(got))
	}
	if doer.headers["Authorization"] != "Bearer tvly-abc" {
		t.Errorf("auth header = %q", doer.headers["Authorization"])
	}
}
