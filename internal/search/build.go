package search

import (
	"time"

	"github.com/Bonifatius94/myscrape-go/internal/config"
	"github.com/Bonifatius94/myscrape-go/internal/httpx"
)

// engineCooldown is how long a rate-limited engine is skipped by the breaker.
const engineCooldown = 60 * time.Second

// Build composes every available engine into a round-robin: the always-on no-key
// engines plus any key-gated engine whose credential is set. Mirrors the Python
// build_search_provider; more engines join here as they're ported.
func Build(h httpx.Doer, s config.Settings) Provider {
	engines := []Provider{
		NewMarginalia(h, s.MarginaliaAPIKey),
	}
	if s.TavilyAPIKey != "" {
		engines = append(engines, NewTavily(h, s.TavilyAPIKey))
	}
	// TODO(port): exa, serpapi, serper, mojeek, google_cse, ddg.
	return NewRoundRobin(engines, nil, engineCooldown)
}
