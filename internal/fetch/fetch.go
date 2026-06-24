// Package fetch turns a URL into clean, readable main content. Static fetch (GET
// via the shared httpx stack + go-trafilatura extraction) is the default; for
// thin/JS-rendered pages it can escalate to a headless-browser render (chromedp)
// behind the Renderer seam, then re-extract.
package fetch

import (
	"context"
	"errors"

	"github.com/Bonifatius94/myscrape/internal/httpx"
)

// ErrEmpty means no main content could be extracted from the page.
var ErrEmpty = errors.New("no main content extracted")

// Result is the cleaned content of one fetched page.
type Result struct {
	URL        string
	Title      string
	Content    string // extracted main content (markdown)
	WordCount  int
	FetchedVia string // "static" | "dynamic"
}

// Renderer renders a URL with a real browser and returns the resulting HTML. It's
// the unstable seam (chromedp); tests mock it.
type Renderer interface {
	Render(ctx context.Context, url string) (string, error)
}

// Fetcher fetches and extracts a single URL.
type Fetcher struct {
	http      httpx.Doer
	renderer  Renderer // optional; nil -> static only
	threshold int      // escalate when static word count is below this
}

// NewFetcher builds a static-only Fetcher.
func NewFetcher(h httpx.Doer) *Fetcher {
	return &Fetcher{http: h}
}

// NewFetcherWithRenderer builds a Fetcher that escalates thin/empty static
// extractions to a browser render.
func NewFetcherWithRenderer(h httpx.Doer, r Renderer, threshold int) *Fetcher {
	return &Fetcher{http: h, renderer: r, threshold: threshold}
}

// Fetch GETs rawURL and extracts its main content, escalating to a render when the
// static result is thin and a renderer is configured. HTTP errors propagate; a
// page with no extractable content (even after render) returns ErrEmpty.
func (f *Fetcher) Fetch(ctx context.Context, rawURL string) (*Result, error) {
	body, err := f.http.Get(ctx, rawURL, nil)
	if err != nil {
		return nil, err
	}
	staticRes, staticErr := Extract(rawURL, body)

	if f.renderer != nil && f.shouldEscalate(staticRes) {
		if rendered, rerr := f.renderer.Render(ctx, rawURL); rerr == nil {
			if dynRes, derr := Extract(rawURL, []byte(rendered)); derr == nil {
				if staticErr != nil || dynRes.WordCount > staticRes.WordCount {
					dynRes.FetchedVia = "dynamic"
					return dynRes, nil
				}
			}
		}
	}
	return staticRes, staticErr
}

// shouldEscalate reports whether the static extraction is too thin to trust.
func (f *Fetcher) shouldEscalate(res *Result) bool {
	return res == nil || res.WordCount < f.threshold
}
