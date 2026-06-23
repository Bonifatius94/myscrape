// Command myscrape is the MCP server entrypoint (stdio or streamable-http).
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/Bonifatius94/myscrape-go/internal/config"
	"github.com/Bonifatius94/myscrape-go/internal/fetch"
	"github.com/Bonifatius94/myscrape-go/internal/httpx"
	"github.com/Bonifatius94/myscrape-go/internal/mcpserver"
	"github.com/Bonifatius94/myscrape-go/internal/research"
	"github.com/Bonifatius94/myscrape-go/internal/search"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	cfg := config.FromEnv()
	client := httpx.New(httpx.Config{
		Timeout:     cfg.RequestTimeout,
		UserAgent:   cfg.UserAgent,
		CacheTTL:    cfg.CacheTTL,
		MinInterval: cfg.RequestMinInterval,
		Jitter:      cfg.RequestJitter,
		Attempts:    cfg.HTTPAttempts,
		BaseDelay:   cfg.HTTPBaseDelay,
		MaxDelay:    cfg.HTTPMaxDelay,
		RetryJitter: cfg.HTTPRetryJitter,
	})

	// Round-robin over every available engine (key-gated ones join when their key
	// is set), with the per-engine circuit breaker.
	provider := search.Build(client, cfg)
	fetcher := fetch.NewFetcher(client)
	researcher := research.NewWebResearcher(provider, fetcher)
	server := mcpserver.New(mcpserver.Deps{
		Search:     provider,
		Fetcher:    fetcher,
		Researcher: researcher,
	})

	switch cfg.MCPTransport {
	case "http", "streamable-http":
		handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, nil)
		mux := http.NewServeMux()
		mux.Handle("/mcp", handler)
		addr := fmt.Sprintf("%s:%d", cfg.MCPHost, cfg.MCPPort)
		log.Printf("myscrape MCP serving on http://%s/mcp", addr)
		log.Fatal(http.ListenAndServe(addr, mux))
	default:
		if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			log.Fatal(err)
		}
	}
}
