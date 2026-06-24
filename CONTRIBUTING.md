# Contributing to myscrape-go

This is the Go port of myscrape (see `README.md`). The original Python repo is the
reference for behavior and the test suite.

## Toolchain

- **Go 1.26+** (chromedp requires it).
- Standard tooling: `gofmt`, `go vet`, `go test`, `go build`.

## TDD loop (required for logic)

1. 🔴 Write a failing test first; confirm it fails for the right reason.
2. 🟢 Minimum code to pass; confirm green.
3. ♻️ Refactor; tests stay green.

Keep unstable dependencies (HTTP, the headless browser) behind an interface and
mock them in tests — `internal/fetch` (the `Renderer` seam) and `internal/httpx`
(the `Doer` seam) are the pattern. Tests must run offline; no network/GPU.

## The gate

Install the pre-commit hook once:

```bash
git config core.hooksPath .githooks
```

It runs **gofmt + go vet + go test** on every commit (mirrored by CI). Don't bypass
it. Before pushing:

```bash
gofmt -l .        # must print nothing
go vet ./...
go test ./...
go run ./cmd/bench # live stability gate (network) — keep ratelimit_rate == 0
```

## Conventions

- **Stable tools only.** Isolate anything unstable behind an interface.
- The behavior, operating points (8s/2s pacing, BM25, circuit breaker), and error
  taxonomy mirror the Python reference — keep them in sync.
- Commit per checkpoint with an imperative subject and a body explaining the *why*.

## Adding a search provider

Copy a provider in `internal/search/` (e.g. `tavily.go`), add a parser test, and
register it in `search.Build` (and a config key in `internal/config`). It joins the
round-robin when its key is set.
