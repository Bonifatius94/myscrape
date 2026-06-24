# Search providers

More independent engines ⇒ more resilience. The round-robin rotates the starting
engine per query and a per-engine circuit breaker cools any engine that
rate-limits, so losing one never stops search. Strategy: scrape only where it's
durable (DDG), prefer clean JSON APIs everywhere else.

## Engines

| Provider | Module | Env var(s) | Type | Note |
|----------|--------|-----------|------|------|
| DuckDuckGo | `search/ddg.go` | — | HTML scrape | always on; rate-limit-prone (breaker covers it) |
| Marginalia | `search/marginalia.go` | `MYSCRAPE_MARGINALIA_API_KEY` (`public`) | JSON API | always on; independent index |
| Tavily | `search/tavily.go` | `MYSCRAPE_TAVILY_API_KEY` | JSON API | free 1,000/mo, no card |
| Exa | `search/providers.go` | `MYSCRAPE_EXA_API_KEY` | JSON API | semantic/neural |
| SerpApi | `search/providers.go` | `MYSCRAPE_SERPAPI_API_KEY` | JSON API | real Google, 100/mo |
| Serper | `search/providers.go` | `MYSCRAPE_SERPER_API_KEY` | JSON API | Google, 2,500/mo |
| Mojeek | `search/providers.go` | `MYSCRAPE_MOJEEK_API_KEY` | JSON API | independent index |
| Google CSE | `search/providers.go` | `MYSCRAPE_GOOGLE_API_KEY` + `MYSCRAPE_GOOGLE_CSE_ID` | JSON API | needs both; 100/day |

A key-gated engine joins the round-robin the moment its key is set (Google needs
both vars). `MYSCRAPE_SEARCH_PROVIDER` selects a single named engine, or
`roundrobin` (default) for the composite.

## Adding one

Copy a provider (e.g. `search/tavily.go`) + its parser test, register it in
`search.Build`, and add the config key in `internal/config`. ~30–40 lines.

## Not carried from the Python reference

- **SearXNG** — out of scope (we don't run a SearXNG instance).
- **Brave API** — no longer free (removed its free tier in early 2026).
- Scraping Bing/Brave-web/Mojeek-HTML/Yahoo/Ecosia — blocked from a residential IP;
  the JSON APIs above replace them.
