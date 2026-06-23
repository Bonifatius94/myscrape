package fetch

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	trafilatura "github.com/markusmobius/go-trafilatura"
)

// Extract pulls the main readable content out of a page's HTML. rawURL is used as
// the document's original URL (helps metadata/link resolution). Returns ErrEmpty
// when nothing substantive could be extracted.
func Extract(rawURL string, html []byte) (*Result, error) {
	var orig *url.URL
	if u, err := url.Parse(rawURL); err == nil {
		orig = u
	}

	res, err := trafilatura.Extract(bytes.NewReader(html), trafilatura.Options{
		OriginalURL:     orig,
		ExcludeComments: true,
		EnableFallback:  true, // jusText / Readability fallback, like the Python lib
	})
	if err != nil {
		// trafilatura errors on too-short/no content — that's EMPTY_CONTENT for us.
		return nil, fmt.Errorf("%w (%s)", ErrEmpty, err)
	}

	text := strings.TrimSpace(res.ContentText)
	if text == "" {
		return nil, ErrEmpty
	}
	return &Result{
		URL:        rawURL,
		Title:      strings.TrimSpace(res.Metadata.Title),
		Content:    text,
		WordCount:  len(strings.Fields(text)),
		FetchedVia: "static",
	}, nil
}
