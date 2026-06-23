package search

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Bonifatius94/myscrape-go/internal/httpx"
)

const tavilyEndpoint = "https://api.tavily.com/search"

// Tavily is a Provider backed by the Tavily API (free 1,000/mo, no card). POST with
// a Bearer key; clean results with a short content snippet per hit.
type Tavily struct {
	http   httpx.Doer
	apiKey string
}

// NewTavily builds the provider.
func NewTavily(h httpx.Doer, apiKey string) *Tavily {
	return &Tavily{http: h, apiKey: apiKey}
}

func (t *Tavily) Name() string { return "tavily" }

func (t *Tavily) Search(ctx context.Context, query string, maxResults int) ([]Result, error) {
	body := map[string]any{
		"query":        query,
		"max_results":  min(maxResults, 20),
		"search_depth": "basic",
	}
	data, err := t.http.PostJSON(ctx, tavilyEndpoint, body,
		map[string]string{"Authorization": "Bearer " + t.apiKey})
	if err != nil {
		return nil, err
	}
	results, err := parseTavily(data)
	if err != nil {
		return nil, err
	}
	if len(results) > maxResults {
		results = results[:maxResults]
	}
	return results, nil
}

func parseTavily(body []byte) ([]Result, error) {
	var payload struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
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
			Snippet: strings.TrimSpace(r.Content),
		})
	}
	return out, nil
}
