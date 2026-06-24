package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/Bonifatius94/myscrape/internal/httpx"
)

const marginaliaEndpoint = "https://api2.marginalia-search.com/search"

// Marginalia is a Provider backed by Marginalia's independent index (public JSON
// API, free "public" key; 503 = shared-rate-limit). No markup drift, direct URLs.
type Marginalia struct {
	http   httpx.Doer
	apiKey string
}

// NewMarginalia builds the provider; an empty key falls back to "public".
func NewMarginalia(h httpx.Doer, apiKey string) *Marginalia {
	if apiKey == "" {
		apiKey = "public"
	}
	return &Marginalia{http: h, apiKey: apiKey}
}

func (m *Marginalia) Name() string { return "marginalia" }

func (m *Marginalia) Search(ctx context.Context, query string, maxResults int) ([]Result, error) {
	endpoint := fmt.Sprintf("%s?%s", marginaliaEndpoint, url.Values{
		"query": {query},
		"count": {fmt.Sprint(maxResults)},
	}.Encode())
	body, err := m.http.Get(ctx, endpoint, map[string]string{"API-Key": m.apiKey})
	if err != nil {
		return nil, err
	}
	results, err := parseMarginalia(body)
	if err != nil {
		return nil, err
	}
	if len(results) > maxResults {
		results = results[:maxResults]
	}
	return results, nil
}

// parseMarginalia turns a Marginalia API response into ranked Results.
func parseMarginalia(body []byte) ([]Result, error) {
	var payload struct {
		Results []struct {
			URL         string `json:"url"`
			Title       string `json:"title"`
			Description string `json:"description"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	out := make([]Result, 0, len(payload.Results))
	for _, r := range payload.Results {
		if r.URL == "" {
			continue
		}
		out = append(out, Result{
			Rank:    len(out) + 1,
			Title:   strings.TrimSpace(r.Title),
			URL:     r.URL,
			Site:    siteOf(r.URL),
			Snippet: strings.TrimSpace(r.Description),
		})
	}
	return out, nil
}
