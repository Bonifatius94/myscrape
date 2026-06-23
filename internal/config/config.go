// Package config holds the runtime settings, overridable via MYSCRAPE_* env vars.
// Defaults mirror the Python reference (see ../../specs in the Python repo). This
// is the Phase-1 subset; stability knobs (pacing/retry/cache) arrive with httpx.
package config

import (
	"os"
	"strconv"
	"time"
)

// DefaultUserAgent is a current-browser UA — we don't advertise as a bot.
const DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
	"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

// Settings is the frozen operating config for one process.
type Settings struct {
	SearchProvider    string // "roundrobin" or a single engine name
	MarginaliaAPIKey  string // free "public" key works
	TavilyAPIKey      string
	RequestTimeout    time.Duration
	UserAgent         string
	MCPTransport      string // "stdio" | "http"
	MCPHost           string
	MCPPort           int
	ResearchSynthesis string // "simple" (GPU-free, default) | "llm"
}

// FromEnv builds Settings, overriding each field from its MYSCRAPE_* env var.
func FromEnv() Settings {
	return Settings{
		SearchProvider:    env("MYSCRAPE_SEARCH_PROVIDER", "roundrobin"),
		MarginaliaAPIKey:  env("MYSCRAPE_MARGINALIA_API_KEY", "public"),
		TavilyAPIKey:      env("MYSCRAPE_TAVILY_API_KEY", ""),
		RequestTimeout:    time.Duration(envInt("MYSCRAPE_REQUEST_TIMEOUT", 20)) * time.Second,
		UserAgent:         env("MYSCRAPE_USER_AGENT", DefaultUserAgent),
		MCPTransport:      env("MYSCRAPE_MCP_TRANSPORT", "stdio"),
		MCPHost:           env("MYSCRAPE_MCP_HOST", "127.0.0.1"),
		MCPPort:           envInt("MYSCRAPE_MCP_PORT", 8000),
		ResearchSynthesis: env("MYSCRAPE_RESEARCH_SYNTHESIS", "simple"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
