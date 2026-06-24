package fetch

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type stubDoer struct {
	body []byte
	err  error
}

func (s stubDoer) Get(context.Context, string, map[string]string) ([]byte, error) {
	return s.body, s.err
}
func (s stubDoer) PostJSON(context.Context, string, any, map[string]string) ([]byte, error) {
	return s.body, s.err
}

type fakeRenderer struct {
	html   string
	err    error
	called bool
}

func (r *fakeRenderer) Render(context.Context, string) (string, error) {
	r.called = true
	return r.html, r.err
}

const thinHTML = "<html><body><p>x</p></body></html>" // extracts to nothing

func TestFetchEscalatesWhenStaticThin(t *testing.T) {
	rend := &fakeRenderer{html: sampleHTML}
	f := NewFetcherWithRenderer(stubDoer{body: []byte(thinHTML)}, rend, 50)

	res, err := f.Fetch(context.Background(), "u")
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if !rend.called {
		t.Error("renderer should be invoked for a thin static page")
	}
	if res.FetchedVia != "dynamic" {
		t.Errorf("fetchedVia = %q, want dynamic", res.FetchedVia)
	}
	if !strings.Contains(res.Content, "TaskGroups") {
		t.Errorf("dynamic content missing: %q", res.Content)
	}
}

func TestFetchNoEscalationWhenStaticRich(t *testing.T) {
	rend := &fakeRenderer{html: "<html><body><p>junk</p></body></html>"}
	f := NewFetcherWithRenderer(stubDoer{body: []byte(sampleHTML)}, rend, 20)

	res, err := f.Fetch(context.Background(), "u")
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if rend.called {
		t.Error("renderer must NOT be called when static content is rich")
	}
	if res.FetchedVia != "static" {
		t.Errorf("fetchedVia = %q, want static", res.FetchedVia)
	}
}

func TestFetchRendererErrorFallsBackToStatic(t *testing.T) {
	// Truly empty static -> ErrEmpty; render fails -> the static error stands.
	rend := &fakeRenderer{err: errors.New("no chrome available")}
	f := NewFetcherWithRenderer(stubDoer{body: []byte("<html><body></body></html>")}, rend, 50)

	_, err := f.Fetch(context.Background(), "u")
	if !errors.Is(err, ErrEmpty) {
		t.Fatalf("want ErrEmpty fallback, got %v", err)
	}
}
