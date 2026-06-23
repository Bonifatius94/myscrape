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
		NewDDG(h),
		NewMarginalia(h, s.MarginaliaAPIKey),
	}
	if s.TavilyAPIKey != "" {
		engines = append(engines, NewTavily(h, s.TavilyAPIKey))
	}
	if s.ExaAPIKey != "" {
		engines = append(engines, NewExa(h, s.ExaAPIKey))
	}
	if s.SerpAPIKey != "" {
		engines = append(engines, NewSerpAPI(h, s.SerpAPIKey))
	}
	if s.SerperAPIKey != "" {
		engines = append(engines, NewSerper(h, s.SerperAPIKey))
	}
	if s.MojeekAPIKey != "" {
		engines = append(engines, NewMojeek(h, s.MojeekAPIKey))
	}
	if s.GoogleAPIKey != "" && s.GoogleCSEID != "" {
		engines = append(engines, NewGoogleCSE(h, s.GoogleAPIKey, s.GoogleCSEID))
	}
	// TODO(port): DDG HTML scraper.
	return NewRoundRobin(engines, nil, engineCooldown)
}
