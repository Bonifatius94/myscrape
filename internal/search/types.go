// Package search provides a swappable Provider interface and engine adapters.
package search

import (
	"context"
	"net/url"
)

// Result is one search hit. Published is an ISO date string when available.
type Result struct {
	Rank      int    `json:"rank"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Site      string `json:"site"`
	Snippet   string `json:"snippet"`
	Published string `json:"published,omitempty"`
}

// Provider is a swappable search backend. Implementations must not get
// rate-limited at target volume — that's what the stability benchmark verifies.
type Provider interface {
	Name() string
	Search(ctx context.Context, query string, maxResults int) ([]Result, error)
}

// siteOf returns the host of a URL, or "" if it can't be parsed.
func siteOf(raw string) string {
	if u, err := url.Parse(raw); err == nil {
		return u.Host
	}
	return ""
}
