# myscrape — tool contracts

Three MCP tools. Two raw (`web_search`, `web_fetch`), one cooked (`web_research`).
Responses are JSON (returned as MCP text content). Errors are structured, never
exceptions: `{ "error": { "code", "message", "retryable" } }`.

## `web_search(query, max_results=10)`
Ranked results, no fetching.
```
{ query, count, results: [ { rank, title, url, site, snippet, published } ] }
```
Errors: `RATELIMITED` (retryable) · `SEARCH_UNAVAILABLE`.

## `web_fetch(url, max_tokens?)`
One URL → clean main content as **markdown**.
```
{ url, title, content, word_count, fetched_via: "static"|"dynamic", truncated }
```
Errors: `BLOCKED` (401/403/451) · `UNREACHABLE` · `EMPTY_CONTENT` · `TIMEOUT` ·
`RATELIMITED`.

## `web_research(question, effort="standard", return_mode="answer", synthesis="")`
Full loop: search → fetch top N → chunk → BM25 rank → synthesize.
- `effort`: `quick` (3 sources / 6 passages) · `standard` (5/10) · `deep` (8/16).
- `return_mode`: `answer` · `sources` · `both`.
- `synthesis`: `simple` (extractive, **no LLM/GPU**) · `llm` (model-written). Empty
  → server default (`MYSCRAPE_RESEARCH_SYNTHESIS`, default `simple`).
```
{ answer, citations: [ { marker:"[1]", url, title } ],
  sources?: [ { url, title } ],
  coverage: { confidence: "high"|"medium"|"low", note } }
```
Errors: `LLM_ERROR` (retryable) · `RATELIMITED` · `SEARCH_UNAVAILABLE`.

## Cross-cutting
- **Citations:** `[n]` markers resolve against `citations`, derived
  deterministically from the numbered sources (extractive) or the markers the model
  emits (llm). Out-of-range markers are dropped.
- **Coverage:** distinct-source count → high (≥3) / medium (2) / low (≤1).
- **Politeness:** per-host pacing + backoff are internal; the caller never tunes
  them (see `specs/EXPERIMENTS.md` for the operating point).
- **Concurrency:** in server (HTTP) mode `web_research` is capped
  (`MYSCRAPE_MAX_CONCURRENT_RESEARCH`, default 2) to bound GPU contention in llm
  mode; stdio is single-user/unbounded.
