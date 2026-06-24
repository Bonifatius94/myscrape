# Dev log — Go port

Decisions for the Go implementation. Behavior, operating points, and the original
research are inherited from the Python reference (its full DEVLOG — D-001…D-018,
F-001…F-005, E-001…E-008, CP-1…CP-34 — is the archive); `specs/EXPERIMENTS.md` is
the carried-over evidence. New decisions go here, newest on top. Format per the
`log-decision` skill.

## Decisions

### G-007 — Dynamic fetch via chromedp; module on Go 1.26
Headless rendering for thin/JS pages uses **chromedp** (pure-Go CDP, no Node),
behind the `Renderer` interface so it's mocked in tests and never run in CI. It
requires **Go 1.26**, so the module/CI/Dockerfile track 1.26. The static Docker
image sets `MYSCRAPE_DYNAMIC_ENABLED=false` (no Chrome inside); enable it on a
Chrome-equipped host. A render error falls back to the static result.

### G-006 — web_fetch returns markdown
go-trafilatura yields a content node + plain text; we render the node to markdown
(`html-to-markdown/v2`) to preserve headings/lists/links, falling back to plain
text. Matches the reference's markdown output.

### G-005 — Synthesis is extractive by default (GPU-free)
`web_research` defaults to **extractive** synthesis (concatenate BM25-ranked
passages, deterministic citations) — no LLM, no GPU. `synthesis=llm` is opt-in and
talks to any OpenAI-compatible endpoint (`internal/llm`). This decouples the
shippable binary from the GPU (which other workloads need); the server-mode cap
(default 2) bounds GPU contention when llm is enabled.

### G-004 — Stability layers carried verbatim behind the Doer seam
`internal/httpx` composes Tier-0 cache → Tier-1 per-host pacer (8s/2s) → Tier-3
retry/backoff (Retry-After aware) over `net/http`. Clock/sleep/rng are injected, so
the whole stack is deterministic offline. The round-robin circuit breaker cools a
rate-limited engine and a healthy-but-empty result never surfaces its error. These
match the reference exactly and are re-verified by `cmd/bench`.

### G-003 — Official modelcontextprotocol/go-sdk for the MCP server
Chosen over community SDKs for spec compliance; its `mcp` package ships both
`StdioTransport` and `StreamableHTTPHandler` (verified), covering the local and
shared-service transports. Tool logic lives in testable `do*` functions; the
registered tools are thin wrappers.

### G-002 — go-trafilatura for extraction
A faithful Go port of the reference's extractor (trafilatura), Readability fallback
included. Validated against the same fixtures; de-risks the one hard dependency.

### G-001 — Rewrite in Go, shipped as compiled artifacts
The product is delivered as compiled binaries (no interpreter on customer
machines). Go fits an I/O-bound MCP service and a multi-file deliverable. The
behavior is pinned by the reference's spec + tests, so this is a translation, not a
redesign. The Python repo remains the differential-test oracle until parity, then
is archived.

## Verification status

All three tools verified live: `web_search`/`web_fetch` (round-robin, markdown),
`web_research` extractive and `synthesis=llm` (qwen2.5:14b). `cmd/bench` PASS
(`ratelimit_rate 0.00`). Offline gate (gofmt + vet + tests) green in CI on 1.26.
