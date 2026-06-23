package search

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/Bonifatius94/myscrape-go/internal/httpx"
)

// This file ports the remaining key-gated JSON API providers. Each is the same
// shape: build a request, parse a JSON list into ranked Results, cap to max.

// cap trims results to at most maxResults.
func capTo(results []Result, maxResults int) []Result {
	if len(results) > maxResults {
		return results[:maxResults]
	}
	return results
}

// --- Exa (semantic, POST x-api-key) ---

const exaEndpoint = "https://api.exa.ai/search"

type Exa struct {
	http   httpx.Doer
	apiKey string
}

func NewExa(h httpx.Doer, apiKey string) *Exa { return &Exa{http: h, apiKey: apiKey} }
func (e *Exa) Name() string                   { return "exa" }

func (e *Exa) Search(ctx context.Context, query string, maxResults int) ([]Result, error) {
	body := map[string]any{
		"query":      query,
		"numResults": min(maxResults, 25),
		"type":       "auto",
		"contents":   map[string]any{"text": map[string]any{"maxCharacters": 300}},
	}
	data, err := e.http.PostJSON(ctx, exaEndpoint, body, map[string]string{"x-api-key": e.apiKey})
	if err != nil {
		return nil, err
	}
	results, err := parseExa(data)
	if err != nil {
		return nil, err
	}
	return capTo(results, maxResults), nil
}

func parseExa(body []byte) ([]Result, error) {
	var p struct {
		Results []struct {
			Title         string `json:"title"`
			URL           string `json:"url"`
			Text          string `json:"text"`
			PublishedDate string `json:"publishedDate"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, err
	}
	var out []Result
	for _, r := range p.Results {
		if r.URL == "" {
			continue
		}
		out = append(out, newResult(len(out)+1, r.Title, r.URL, r.Text, r.PublishedDate))
	}
	return out, nil
}

// --- SerpApi (real Google, GET, key in query) ---

const serpapiEndpoint = "https://serpapi.com/search.json"

type SerpAPI struct {
	http   httpx.Doer
	apiKey string
}

func NewSerpAPI(h httpx.Doer, apiKey string) *SerpAPI { return &SerpAPI{http: h, apiKey: apiKey} }
func (s *SerpAPI) Name() string                       { return "serpapi" }

func (s *SerpAPI) Search(ctx context.Context, query string, maxResults int) ([]Result, error) {
	u := serpapiEndpoint + "?" + url.Values{
		"q": {query}, "api_key": {s.apiKey}, "engine": {"google"},
		"num": {strconv.Itoa(min(maxResults, 20))},
	}.Encode()
	data, err := s.http.Get(ctx, u, nil)
	if err != nil {
		return nil, err
	}
	results, err := parseSerpAPI(data)
	if err != nil {
		return nil, err
	}
	return capTo(results, maxResults), nil
}

func parseSerpAPI(body []byte) ([]Result, error) {
	var p struct {
		Organic []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
			Date    string `json:"date"`
		} `json:"organic_results"`
	}
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, err
	}
	var out []Result
	for _, r := range p.Organic {
		if r.Link == "" {
			continue
		}
		out = append(out, newResult(len(out)+1, r.Title, r.Link, r.Snippet, r.Date))
	}
	return out, nil
}

// --- Serper (Google, POST X-API-KEY) ---

const serperEndpoint = "https://google.serper.dev/search"

type Serper struct {
	http   httpx.Doer
	apiKey string
}

func NewSerper(h httpx.Doer, apiKey string) *Serper { return &Serper{http: h, apiKey: apiKey} }
func (s *Serper) Name() string                      { return "serper" }

func (s *Serper) Search(ctx context.Context, query string, maxResults int) ([]Result, error) {
	body := map[string]any{"q": query, "num": min(maxResults, 20)}
	data, err := s.http.PostJSON(ctx, serperEndpoint, body, map[string]string{"X-API-KEY": s.apiKey})
	if err != nil {
		return nil, err
	}
	results, err := parseSerper(data)
	if err != nil {
		return nil, err
	}
	return capTo(results, maxResults), nil
}

func parseSerper(body []byte) ([]Result, error) {
	var p struct {
		Organic []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
			Date    string `json:"date"`
		} `json:"organic"`
	}
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, err
	}
	var out []Result
	for _, r := range p.Organic {
		if r.Link == "" {
			continue
		}
		out = append(out, newResult(len(out)+1, r.Title, r.Link, r.Snippet, r.Date))
	}
	return out, nil
}

// --- Mojeek (independent index, GET) ---

const mojeekEndpoint = "https://api.mojeek.com/search"

type Mojeek struct {
	http   httpx.Doer
	apiKey string
}

func NewMojeek(h httpx.Doer, apiKey string) *Mojeek { return &Mojeek{http: h, apiKey: apiKey} }
func (m *Mojeek) Name() string                      { return "mojeek" }

func (m *Mojeek) Search(ctx context.Context, query string, maxResults int) ([]Result, error) {
	u := mojeekEndpoint + "?" + url.Values{
		"q": {query}, "api_key": {m.apiKey}, "fmt": {"json"},
		"t": {strconv.Itoa(min(maxResults, 20))},
	}.Encode()
	data, err := m.http.Get(ctx, u, nil)
	if err != nil {
		return nil, err
	}
	results, err := parseMojeek(data)
	if err != nil {
		return nil, err
	}
	return capTo(results, maxResults), nil
}

func parseMojeek(body []byte) ([]Result, error) {
	var p struct {
		Response struct {
			Results []struct {
				Title string `json:"title"`
				URL   string `json:"url"`
				Desc  string `json:"desc"`
			} `json:"results"`
		} `json:"response"`
	}
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, err
	}
	var out []Result
	for _, r := range p.Response.Results {
		if r.URL == "" {
			continue
		}
		out = append(out, newResult(len(out)+1, r.Title, r.URL, r.Desc, ""))
	}
	return out, nil
}

// --- Google Programmable Search (CSE, GET, key + cx) ---

const googleCSEEndpoint = "https://www.googleapis.com/customsearch/v1"

type GoogleCSE struct {
	http   httpx.Doer
	apiKey string
	cseID  string
}

func NewGoogleCSE(h httpx.Doer, apiKey, cseID string) *GoogleCSE {
	return &GoogleCSE{http: h, apiKey: apiKey, cseID: cseID}
}
func (g *GoogleCSE) Name() string { return "google_cse" }

func (g *GoogleCSE) Search(ctx context.Context, query string, maxResults int) ([]Result, error) {
	u := googleCSEEndpoint + "?" + url.Values{
		"key": {g.apiKey}, "cx": {g.cseID}, "q": {query},
		"num": {strconv.Itoa(min(maxResults, 10))}, // CSE caps a request at 10
	}.Encode()
	data, err := g.http.Get(ctx, u, nil)
	if err != nil {
		return nil, err
	}
	results, err := parseGoogleCSE(data)
	if err != nil {
		return nil, err
	}
	return capTo(results, maxResults), nil
}

func parseGoogleCSE(body []byte) ([]Result, error) {
	var p struct {
		Items []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"items"`
	}
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, err
	}
	var out []Result
	for _, r := range p.Items {
		if r.Link == "" {
			continue
		}
		out = append(out, newResult(len(out)+1, r.Title, r.Link, r.Snippet, ""))
	}
	return out, nil
}
