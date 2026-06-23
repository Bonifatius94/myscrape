// Command myscrape is the MCP server entrypoint (stdio or streamable-http).
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/Bonifatius94/myscrape-go/internal/config"
	"github.com/Bonifatius94/myscrape-go/internal/httpx"
	"github.com/Bonifatius94/myscrape-go/internal/mcpserver"
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

	// Phase 1: a single no-key provider. Round-robin over all engines comes next.
	provider := search.NewMarginalia(client, cfg.MarginaliaAPIKey)
	server := mcpserver.New(mcpserver.Deps{Search: provider})

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
