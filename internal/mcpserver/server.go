// Package mcpserver registers the three myscrape tools on an MCP server.
//
// Tool logic lives in do* functions that take dependencies and return a JSON-able
// value (a result struct or a SPEC-shaped error map), mirroring the Python
// reference's do_* functions — so the logic is testable without any transport. The
// registered tools are thin wrappers that marshal the value to text content.
package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/Bonifatius94/myscrape-go/internal/fetch"
	"github.com/Bonifatius94/myscrape-go/internal/httpx"
	"github.com/Bonifatius94/myscrape-go/internal/research"
	"github.com/Bonifatius94/myscrape-go/internal/retry"
	"github.com/Bonifatius94/myscrape-go/internal/search"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// The seams the handlers depend on (so they're testable with fakes).
type searcher interface {
	Search(ctx context.Context, query string, maxResults int) ([]search.Result, error)
}
type fetcher interface {
	Fetch(ctx context.Context, url string) (*fetch.Result, error)
}
type researcher interface {
	Research(ctx context.Context, question, effort, synthesis string) (research.ResearchResult, error)
}

// Deps are the constructed dependencies the tools need.
type Deps struct {
	Search     search.Provider
	Fetcher    *fetch.Fetcher
	Researcher *research.WebResearcher
}

// New builds the MCP server with all tools registered.
func New(deps Deps) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "myscrape", Version: "0.1.0"}, nil)

	mcp.AddTool(s, &mcp.Tool{
		Name: "web_search", Description: "Raw web search. Returns a ranked list of results (no fetching).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, a searchParams) (*mcp.CallToolResult, any, error) {
		return jsonResult(doWebSearch(ctx, deps.Search, a.Query, a.MaxResults))
	})

	mcp.AddTool(s, &mcp.Tool{
		Name: "web_fetch", Description: "Fetch one URL and return its main content as clean text.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, a fetchParams) (*mcp.CallToolResult, any, error) {
		return jsonResult(doWebFetch(ctx, deps.Fetcher, a.URL, a.MaxTokens))
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "web_research",
		Description: "Research a question end-to-end: search, fetch, then synthesize a cited answer (extractive by default; synthesis=llm for model-written).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, a researchParams) (*mcp.CallToolResult, any, error) {
		return jsonResult(doWebResearch(ctx, deps.Researcher, a))
	})

	return s
}

// --- web_search ---

type searchParams struct {
	Query      string `json:"query" jsonschema:"the search query"`
	MaxResults int    `json:"max_results,omitempty" jsonschema:"max results to return (default 10)"`
}

type searchOut struct {
	Query   string          `json:"query"`
	Results []search.Result `json:"results"`
	Count   int             `json:"count"`
}

func doWebSearch(ctx context.Context, provider searcher, query string, maxResults int) any {
	if maxResults <= 0 {
		maxResults = 10
	}
	results, err := provider.Search(ctx, query, maxResults)
	if err != nil {
		if isRateLimited(err) {
			return errMap("RATELIMITED", err.Error(), true)
		}
		return errMap("SEARCH_UNAVAILABLE", err.Error(), true)
	}
	return searchOut{Query: query, Results: results, Count: len(results)}
}

// --- web_fetch ---

type fetchParams struct {
	URL       string `json:"url" jsonschema:"the URL to fetch"`
	MaxTokens int    `json:"max_tokens,omitempty" jsonschema:"truncate content to roughly N words"`
}

type fetchOut struct {
	URL        string `json:"url"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	WordCount  int    `json:"word_count"`
	FetchedVia string `json:"fetched_via"`
	Truncated  bool   `json:"truncated"`
}

func doWebFetch(ctx context.Context, f fetcher, rawURL string, maxTokens int) any {
	res, err := f.Fetch(ctx, rawURL)
	if err != nil {
		code, msg, retryable := mapFetchError(err)
		return errMap(code, msg, retryable)
	}
	content, truncated := res.Content, false
	if maxTokens > 0 {
		if words := strings.Fields(content); len(words) > maxTokens {
			content = strings.Join(words[:maxTokens], " ")
			truncated = true
		}
	}
	return fetchOut{
		URL: res.URL, Title: res.Title, Content: content,
		WordCount: res.WordCount, FetchedVia: res.FetchedVia, Truncated: truncated,
	}
}

// --- web_research ---

type researchParams struct {
	Question   string `json:"question" jsonschema:"the question to research"`
	Effort     string `json:"effort,omitempty" jsonschema:"quick | standard | deep"`
	ReturnMode string `json:"return_mode,omitempty" jsonschema:"answer | sources | both"`
	Synthesis  string `json:"synthesis,omitempty" jsonschema:"simple (extractive, GPU-free) | llm"`
}

func doWebResearch(ctx context.Context, r researcher, a researchParams) any {
	effort := a.Effort
	if effort == "" {
		effort = "standard"
	}
	res, err := r.Research(ctx, a.Question, effort, a.Synthesis)
	if err != nil {
		if errors.Is(err, research.ErrLLM) {
			return errMap("LLM_ERROR", err.Error(), true)
		}
		if isRateLimited(err) {
			return errMap("RATELIMITED", err.Error(), true)
		}
		return errMap("SEARCH_UNAVAILABLE", err.Error(), true)
	}
	mode := a.ReturnMode
	if mode == "" {
		mode = "answer"
	}
	out := map[string]any{"citations": res.Citations, "coverage": res.Coverage, "answer": nil}
	if mode == "answer" || mode == "both" {
		out["answer"] = res.Answer
	}
	if mode == "sources" || mode == "both" {
		out["sources"] = res.Sources
	}
	return out
}

// --- error mapping + helpers ---

// mapFetchError maps fetch/HTTP errors to the SPEC error taxonomy.
func mapFetchError(err error) (code, msg string, retryable bool) {
	var re *retry.Retryable
	if errors.As(err, &re) {
		return "RATELIMITED", re.Error(), true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "TIMEOUT", netErr.Error(), true
	}
	var he *httpx.HTTPError
	if errors.As(err, &he) {
		if he.Status == 401 || he.Status == 403 || he.Status == 451 {
			return "BLOCKED", fmt.Sprintf("site returned %d", he.Status), false
		}
		return "UNREACHABLE", fmt.Sprintf("HTTP %d", he.Status), false
	}
	if errors.Is(err, fetch.ErrEmpty) {
		return "EMPTY_CONTENT", "no main content could be extracted", false
	}
	return "UNREACHABLE", err.Error(), false
}

func isRateLimited(err error) bool {
	var re *retry.Retryable
	if errors.As(err, &re) {
		return true
	}
	var he *httpx.HTTPError
	return errors.As(err, &he) && (he.Status == 429 || he.Status == 503 || he.Status == 202)
}

func jsonResult(v any) (*mcp.CallToolResult, any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, nil, nil
}

func errMap(code, message string, retryable bool) map[string]any {
	return map[string]any{
		"error": map[string]any{"code": code, "message": message, "retryable": retryable},
	}
}
