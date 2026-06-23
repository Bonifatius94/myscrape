# myscrape (Go)

A self-contained **web-research MCP server** for local agents — `web_search`,
`web_fetch`, `web_research` — shipped as a compiled binary (no interpreter needed).

This is the **Go port** of the Python reference implementation. The port follows
the spec, behavior, and test suite of the original; see `specs/PORTING_GO.md` in the
Python repo for the module map and the empirically-tuned operating points that
carry over verbatim. The Python version is kept only as the differential-test
oracle/fallback until this reaches parity, then retired.

## Status — Phase 1 (scaffold)

Working spine: MCP server over **stdio** and **streamable-HTTP** (official
`modelcontextprotocol/go-sdk`), with `web_search` wired through a real provider
(Marginalia, no key needed). `web_fetch` / `web_research` are registered but
stubbed. Default synthesis is GPU-free (`simple`) by design — no LLM, no browser.

```bash
go build -o myscrape ./cmd/myscrape

# stdio (default)
./myscrape

# streamable-HTTP
MYSCRAPE_MCP_TRANSPORT=http MYSCRAPE_MCP_PORT=8000 ./myscrape   # -> http://127.0.0.1:8000/mcp
```

## Layout

```
cmd/myscrape        entrypoint (stdio | http)
internal/config     MYSCRAPE_* settings
internal/httpx      HTTP primitive (Doer seam; stability layers land here next)
internal/search     Provider interface + engine adapters (Marginalia so far)
internal/mcpserver  tool registration + handlers
internal/fetch      (Phase 1: static fetch via go-trafilatura — TODO)
internal/research   (Phase 1: chunk + BM25 + extractive synthesis — TODO)
```

## Next (Phase 1 remaining)

httpx stability layers (cache / per-host rate-limit / retry+backoff) → remaining
search providers + round-robin/circuit-breaker → static fetch (go-trafilatura) →
chunk + BM25 + extractive synthesis → wire `web_fetch` / `web_research`.
