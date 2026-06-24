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
	ExaAPIKey         string
	SerpAPIKey        string
	SerperAPIKey      string
	MojeekAPIKey      string
	GoogleAPIKey      string
	GoogleCSEID       string
	RequestTimeout    time.Duration
	UserAgent         string
	MCPTransport      string // "stdio" | "http"
	MCPHost           string
	MCPPort           int
	ResearchSynthesis string // "simple" (GPU-free, default) | "llm"
	// Server-mode cap on concurrent web_research (bounds GPU contention in llm mode).
	MaxConcurrentResearch int

	// LLM (OpenAI-compatible) — only used when synthesis=llm. Swappable by base URL.
	LLMBaseURL string
	LLMModel   string
	LLMAPIKey  string
	LLMTimeout time.Duration

	// Stability knobs (Python reference defaults). Per-host pacing is the proven
	// 8s + 2s point; retry/backoff and cache TTL match the reference.
	RequestMinInterval time.Duration
	RequestJitter      time.Duration
	CacheTTL           time.Duration
	HTTPAttempts       int
	HTTPBaseDelay      time.Duration
	HTTPMaxDelay       time.Duration
	HTTPRetryJitter    time.Duration
}

// FromEnv builds Settings, overriding each field from its MYSCRAPE_* env var.
func FromEnv() Settings {
	return Settings{
		SearchProvider:        env("MYSCRAPE_SEARCH_PROVIDER", "roundrobin"),
		MarginaliaAPIKey:      env("MYSCRAPE_MARGINALIA_API_KEY", "public"),
		TavilyAPIKey:          env("MYSCRAPE_TAVILY_API_KEY", ""),
		ExaAPIKey:             env("MYSCRAPE_EXA_API_KEY", ""),
		SerpAPIKey:            env("MYSCRAPE_SERPAPI_API_KEY", ""),
		SerperAPIKey:          env("MYSCRAPE_SERPER_API_KEY", ""),
		MojeekAPIKey:          env("MYSCRAPE_MOJEEK_API_KEY", ""),
		GoogleAPIKey:          env("MYSCRAPE_GOOGLE_API_KEY", ""),
		GoogleCSEID:           env("MYSCRAPE_GOOGLE_CSE_ID", ""),
		RequestTimeout:        envSeconds("MYSCRAPE_REQUEST_TIMEOUT", 20),
		UserAgent:             env("MYSCRAPE_USER_AGENT", DefaultUserAgent),
		MCPTransport:          env("MYSCRAPE_MCP_TRANSPORT", "stdio"),
		MCPHost:               env("MYSCRAPE_MCP_HOST", "127.0.0.1"),
		MCPPort:               envInt("MYSCRAPE_MCP_PORT", 8000),
		ResearchSynthesis:     env("MYSCRAPE_RESEARCH_SYNTHESIS", "simple"),
		MaxConcurrentResearch: envInt("MYSCRAPE_MAX_CONCURRENT_RESEARCH", 2),
		LLMBaseURL:            env("MYSCRAPE_LLM_BASE_URL", "http://localhost:11434/v1"),
		LLMModel:              env("MYSCRAPE_LLM_MODEL", "qwen2.5:14b"),
		LLMAPIKey:             env("MYSCRAPE_LLM_API_KEY", ""),
		LLMTimeout:            envSeconds("MYSCRAPE_LLM_TIMEOUT", 120),
		RequestMinInterval:    envSeconds("MYSCRAPE_REQUEST_MIN_INTERVAL", 8),
		RequestJitter:         envSeconds("MYSCRAPE_REQUEST_JITTER", 2),
		CacheTTL:              envSeconds("MYSCRAPE_CACHE_TTL", 900),
		HTTPAttempts:          envInt("MYSCRAPE_HTTP_ATTEMPTS", 4),
		HTTPBaseDelay:         envSeconds("MYSCRAPE_HTTP_BASE_DELAY", 1),
		HTTPMaxDelay:          envSeconds("MYSCRAPE_HTTP_MAX_DELAY", 8),
		HTTPRetryJitter:       envSeconds("MYSCRAPE_HTTP_JITTER", 0.5),
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

// envSeconds reads a fractional-seconds value into a Duration.
func envSeconds(key string, defSeconds float64) time.Duration {
	secs := defSeconds
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			secs = f
		}
	}
	return time.Duration(secs * float64(time.Second))
}
