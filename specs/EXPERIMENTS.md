# Operating points (empirical)

These constants are not guesses — they were established empirically in the Python
reference and **re-verified live in Go** (`go run ./cmd/bench` → `ratelimit_rate
0.00`, PASS). The full original experiment log (E-001…E-008) lives in the Python
repo; this is the carried-over summary that justifies the numbers in
`internal/config` and `internal/httpx`.

## Search pacing (the big one)
- The DuckDuckGo HTML rate-limit is **pacing-driven and short-lived** (recovers in
  minutes, not a persistent ban). User-Agent is a non-factor at safe spacing.
- **Per-host `8s` interval + `2s` jitter → `ratelimit_rate == 0`** at our volume.
  This is the config default. Stability over throughput, by design — a consistent
  slow pace is acceptable; throughput is explicitly not a goal.
- Confound check (done first): a rate-limit lasts minutes, so spaced experiments
  are attributable to code. Keep this in mind before any retuning.

## Retry / cache
- Retryable statuses: `202, 429, 500, 502, 503, 504`; exponential backoff
  (base 1s, cap 8s, +jitter), honoring integer `Retry-After`. 4 attempts.
- Response cache TTL `900s` (a hit skips the network).

## Fetch
- Static extraction (go-trafilatura) succeeds on typical doc/article pages.
- `401/403/451` are a distinct **`BLOCKED`** outcome (site-side authorization, e.g.
  Wikipedia's TLS-layer block) — not our instability; we report, never evade.
- **Dynamic escalation:** "thin static extraction" is the trigger to render with a
  headless browser. Over-triggering only wastes a render, never correctness.

## Synthesis / GPU
- On a 12 GB GPU, the single card **serializes** LLM synthesis — concurrency buys no
  throughput, only latency. Hence the server-mode research cap of **2**.
- `qwen2.5:14b` was the reference synthesis model (perfect quality + fastest at
  ~2s). The LLM endpoint is swappable by base URL; **default synthesis is
  extractive (no LLM/GPU)**.

## The gate
`cmd/bench` runs the query set at the operating point; it **passes only when no
request was rate-limited**. If you ever pace differently, keep the gate's
expectation aligned with the strategy.
