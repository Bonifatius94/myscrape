# myscrape (Go)

[![CI](https://github.com/Bonifatius94/myscrape/actions/workflows/ci.yml/badge.svg)](https://github.com/Bonifatius94/myscrape/actions/workflows/ci.yml)

A self-contained **web-research MCP server** for local agents — `web_search`,
`web_fetch`, `web_research` — shipped as a compiled binary (no interpreter needed).

This is the **Go port** of the Python reference implementation. The port follows
the spec, behavior, and test suite of the original; see `specs/PORTING_GO.md` in the
Python repo for the module map and the empirically-tuned operating points that
carry over verbatim. The Python version is kept only as the differential-test
oracle/fallback until this reaches parity, then retired.

## Status — Phase 1 complete (functional parity, GPU-free core)

All three tools work over **stdio** and **streamable-HTTP** (official
`modelcontextprotocol/go-sdk`):

- **`web_search`** — round-robin over many engines (DDG + Marginalia always on;
  Tavily/Exa/SerpApi/Serper/Mojeek/Google CSE join when their key is set), with a
  per-engine circuit breaker.
- **`web_fetch`** — static fetch + main-content extraction (go-trafilatura).
- **`web_research`** — search → fetch → chunk → BM25 rank → synthesize. Default
  synthesis is **extractive** (no LLM/GPU); set `synthesis=llm` (and an LLM
  endpoint) to opt into model-written answers.

The HTTP stack carries the Python reference's stability behavior verbatim: Tier-0
cache, Tier-1 per-host pacing (8s/2s), Tier-3 retry/backoff with Retry-After.

```bash
go build -o myscrape ./cmd/myscrape
./myscrape                                                     # stdio
MYSCRAPE_MCP_TRANSPORT=http MYSCRAPE_MCP_PORT=8000 ./myscrape  # -> http://127.0.0.1:8000/mcp

# Docker (static image, ~30 MB, GPU-free)
docker compose up -d --build
```

Provider keys and the optional LLM endpoint are configured via `MYSCRAPE_*` env
vars — see [`.env.example`](.env.example).

## Docs

- [`specs/SPEC.md`](specs/SPEC.md) — tool contracts + error taxonomy
- [`specs/PROVIDERS.md`](specs/PROVIDERS.md) — search engines + how to add one
- [`specs/EXPERIMENTS.md`](specs/EXPERIMENTS.md) — the empirical operating points
- [`specs/DEVLOG.md`](specs/DEVLOG.md) — Go port decisions (G-001…)
- [`docs/INTEGRATION.md`](docs/INTEGRATION.md) — consuming myscrape from another product
- [`CONTRIBUTING.md`](CONTRIBUTING.md) — workflow, hooks, conventions

## Layout

```
cmd/myscrape        entrypoint (stdio | http)
internal/config     MYSCRAPE_* settings
internal/httpx      HTTP primitive: cache -> rate-limit -> retry (the Doer seam)
internal/cache      Tier-0 TTL cache
internal/ratelimit  Tier-1 per-host pacer
internal/retry      Tier-3 backoff (Retryable, Retry-After)
internal/search     Provider interface, engine adapters, round-robin/breaker
internal/fetch      static fetch + go-trafilatura extraction
internal/research   chunk, BM25 rank, extractive + LLM synthesis, pipeline
internal/llm        OpenAI-compatible chat client (opt-in synthesis)
internal/mcpserver  tool registration + handlers
```

## Beyond the core

- **Dynamic (JS) fetch** — `internal/fetch` escalates thin/JS pages to a headless
  Chrome render via `chromedp` (`MYSCRAPE_DYNAMIC_ENABLED`, needs Chrome on PATH).
- **`cmd/bench`** — live stability gate (`go run ./cmd/bench`); passes when
  `ratelimit_rate == 0` at the 8s/2s operating point.
- **CI/release** — GitHub Actions runs the offline gate + Docker build; tagging
  `v*` builds cross-OS binaries and publishes a multi-arch image to GHCR.
