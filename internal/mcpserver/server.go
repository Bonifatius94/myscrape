// Package mcpserver registers the three myscrape tools on an MCP server.
//
// Tool logic returns JSON (as text content), mirroring the Python reference's
// SPEC-shaped dicts so existing MCP clients and .mcp.json keep working. Phase 1
// wires web_search and web_fetch; web_research is registered but stubbed.
package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Bonifatius94/myscrape-go/internal/fetch"
	"github.com/Bonifatius94/myscrape-go/internal/httpx"
	"github.com/Bonifatius94/myscrape-go/internal/retry"
	"github.com/Bonifatius94/myscrape-go/internal/search"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Deps are the constructed dependencies the tools need.
type Deps struct {
	Search  search.Provider
	Fetcher *fetch.Fetcher
}

// New builds the MCP server with all tools registered.
func New(deps Deps) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "myscrape", Version: "0.1.0"}, nil)
	registerWebSearch(s, deps.Search)
	registerWebFetch(s, deps.Fetcher)
	registerStub(s, "web_research", "Research a question end-to-end: search, fetch, then synthesize.")
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

func registerWebSearch(s *mcp.Server, provider search.Provider) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "web_search",
		Description: "Raw web search. Returns a ranked list of results (no fetching).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args searchParams) (*mcp.CallToolResult, any, error) {
		max := args.MaxResults
		if max <= 0 {
			max = 10
		}
		results, err := provider.Search(ctx, args.Query, max)
		if err != nil {
			return errorResult("SEARCH_UNAVAILABLE", err.Error(), true)
		}
		return jsonResult(searchOut{Query: args.Query, Results: results, Count: len(results)})
	})
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

func registerWebFetch(s *mcp.Server, f *fetch.Fetcher) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "web_fetch",
		Description: "Fetch one URL and return its main content as clean text.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args fetchParams) (*mcp.CallToolResult, any, error) {
		res, err := f.Fetch(ctx, args.URL)
		if err != nil {
			code, msg, retryable := mapFetchError(err)
			return errorResult(code, msg, retryable)
		}
		content, truncated := res.Content, false
		if args.MaxTokens > 0 {
			if words := strings.Fields(content); len(words) > args.MaxTokens {
				content = strings.Join(words[:args.MaxTokens], " ")
				truncated = true
			}
		}
		return jsonResult(fetchOut{
			URL: res.URL, Title: res.Title, Content: content,
			WordCount: res.WordCount, FetchedVia: res.FetchedVia, Truncated: truncated,
		})
	})
}

// mapFetchError maps fetch/HTTP errors to the SPEC error taxonomy.
func mapFetchError(err error) (code, msg string, retryable bool) {
	var re *retry.Retryable
	if errors.As(err, &re) {
		return "RATELIMITED", re.Error(), true
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

// --- stub (web_research, until ported) ---

type stubParams struct {
	Arg string `json:"arg,omitempty" jsonschema:"placeholder"`
}

func registerStub(s *mcp.Server, name, desc string) {
	mcp.AddTool(s, &mcp.Tool{Name: name, Description: desc + " (not implemented yet — Phase 1 WIP)"},
		func(ctx context.Context, _ *mcp.CallToolRequest, _ stubParams) (*mcp.CallToolResult, any, error) {
			return errorResult("NOT_IMPLEMENTED", name+" is not implemented yet in the Go port", false)
		})
}

// --- helpers ---

// jsonResult serializes v to JSON and returns it as text content (matching the
// Python server's response shape).
func jsonResult(v any) (*mcp.CallToolResult, any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, nil, nil
}

// errorResult returns a SPEC-shaped structured error as the tool result.
func errorResult(code, message string, retryable bool) (*mcp.CallToolResult, any, error) {
	return jsonResult(map[string]any{
		"error": map[string]any{"code": code, "message": message, "retryable": retryable},
	})
}
