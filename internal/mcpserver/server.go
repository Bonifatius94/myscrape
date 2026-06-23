// Package mcpserver registers the three myscrape tools on an MCP server.
//
// Tool logic returns JSON (as text content), mirroring the Python reference's
// SPEC-shaped dicts so existing MCP clients and .mcp.json keep working. Phase 1
// wires web_search; web_fetch and web_research are registered but stubbed.
package mcpserver

import (
	"context"
	"encoding/json"

	"github.com/Bonifatius94/myscrape-go/internal/search"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Deps are the constructed dependencies the tools need.
type Deps struct {
	Search search.Provider
}

// New builds the MCP server with all tools registered.
func New(deps Deps) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "myscrape", Version: "0.1.0"}, nil)
	registerWebSearch(s, deps.Search)
	registerStub(s, "web_fetch", "Fetch one URL and return its main content as clean markdown.")
	registerStub(s, "web_research", "Research a question end-to-end: search, fetch, then synthesize.")
	return s
}

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
			// TODO(phase1): map to the SPEC structured-error taxonomy
			// (RATELIMITED/SEARCH_UNAVAILABLE/...) instead of a tool error.
			return errorResult("SEARCH_UNAVAILABLE", err.Error())
		}
		return jsonResult(searchOut{Query: args.Query, Results: results, Count: len(results)})
	})
}

type stubParams struct {
	// Accept anything; the stub just reports it's not implemented yet.
	Arg string `json:"arg,omitempty" jsonschema:"placeholder"`
}

func registerStub(s *mcp.Server, name, desc string) {
	mcp.AddTool(s, &mcp.Tool{Name: name, Description: desc + " (not implemented yet — Phase 1 WIP)"},
		func(ctx context.Context, _ *mcp.CallToolRequest, _ stubParams) (*mcp.CallToolResult, any, error) {
			return errorResult("NOT_IMPLEMENTED", name+" is not implemented yet in the Go port")
		})
}

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
func errorResult(code, message string) (*mcp.CallToolResult, any, error) {
	return jsonResult(map[string]any{
		"error": map[string]any{"code": code, "message": message, "retryable": false},
	})
}
