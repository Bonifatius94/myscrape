package search

import (
	"fmt"
	"time"

	"github.com/Bonifatius94/myscrape-go/internal/config"
	"github.com/Bonifatius94/myscrape-go/internal/httpx"
)

// engineCooldown is how long a rate-limited engine is skipped by the breaker.
const engineCooldown = 60 * time.Second

// Build returns the configured search provider. "roundrobin" (or empty) composes
// every available engine; a named engine returns just that one. Unknown -> error.
func Build(h httpx.Doer, s config.Settings) (Provider, error) {
	switch s.SearchProvider {
	case "", "roundrobin":
		return buildRoundRobin(h, s), nil
	case "ddg":
		return NewDDG(h), nil
	case "marginalia":
		return NewMarginalia(h, s.MarginaliaAPIKey), nil
	case "tavily":
		return NewTavily(h, s.TavilyAPIKey), nil
	case "exa":
		return NewExa(h, s.ExaAPIKey), nil
	case "serpapi":
		return NewSerpAPI(h, s.SerpAPIKey), nil
	case "serper":
		return NewSerper(h, s.SerperAPIKey), nil
	case "mojeek":
		return NewMojeek(h, s.MojeekAPIKey), nil
	case "google_cse":
		return NewGoogleCSE(h, s.GoogleAPIKey, s.GoogleCSEID), nil
	default:
		return nil, fmt.Errorf("unknown search provider: %q", s.SearchProvider)
	}
}

// buildRoundRobin composes the always-on no-key engines plus any key-gated engine
// whose credential is set.
func buildRoundRobin(h httpx.Doer, s config.Settings) Provider {
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
	return NewRoundRobin(engines, nil, engineCooldown)
}
