// Package fetch turns a URL into clean, readable main content. Phase 1 is static
// only: GET via the shared httpx stack, then extract with go-trafilatura. The
// dynamic (chromedp) escalation path for JS pages is a Phase-2 module.
package fetch

import (
	"context"
	"errors"

	"github.com/Bonifatius94/myscrape-go/internal/httpx"
)

// ErrEmpty means no main content could be extracted from the page.
var ErrEmpty = errors.New("no main content extracted")

// Result is the cleaned content of one fetched page.
type Result struct {
	URL        string
	Title      string
	Content    string // extracted main text
	WordCount  int
	FetchedVia string // "static" (Phase 1)
}

// Fetcher fetches and extracts a single URL.
type Fetcher struct {
	http httpx.Doer
}

// NewFetcher builds a Fetcher over the shared HTTP client.
func NewFetcher(h httpx.Doer) *Fetcher {
	return &Fetcher{http: h}
}

// Fetch GETs rawURL and extracts its main content. HTTP errors propagate (the
// caller maps them to BLOCKED/UNREACHABLE/etc.); a page with no extractable
// content returns ErrEmpty.
func (f *Fetcher) Fetch(ctx context.Context, rawURL string) (*Result, error) {
	body, err := f.http.Get(ctx, rawURL, nil)
	if err != nil {
		return nil, err
	}
	return Extract(rawURL, body)
}
