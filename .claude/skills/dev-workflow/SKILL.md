---
name: dev-workflow
description: How to build the myscrape-go codebase. Use whenever implementing, fixing, refactoring, or extending Go code in this repo — covers the stable-tools rule, the TDD loop, the pre-commit gate, and commit discipline.
---

# myscrape-go dev workflow

Operating rules for this repository. Follow them by default.

## Tooling: stable only
- Standard Go toolchain: **gofmt, go vet, go test, go build**. Go **1.26+**.
- Use stable, widely-used libraries. Isolate anything unstable (HTTP, the headless
  browser) behind an interface and **mock it in tests** — see the `Doer` seam
  (`internal/httpx`) and the `Renderer` seam (`internal/fetch`). Tests run offline;
  no network, no GPU.

## TDD loop (mandatory for logic)
1. 🔴 Write a failing test first; run it; confirm it fails for the right reason
   (in Go, a missing symbol = a compile failure = red).
2. 🟢 Minimum code to pass; run; confirm green.
3. ♻️ Refactor; tests stay green.

Inject clock/sleep/rng so time-dependent code (rate limiter, backoff) is
deterministic — see `internal/ratelimit` and `internal/retry`.

## The gate
Enable the hook once: `git config core.hooksPath .githooks`. It runs **gofmt + go
vet + go test** on every commit (CI mirrors it). Don't bypass it. For stability
work also run `go run ./cmd/bench` (live; keep `ratelimit_rate == 0`).

## Behavior is inherited
Operating points (8s/2s pacing, retryable statuses, BM25, circuit breaker, the
error taxonomy, effort knobs) mirror the Python reference and are evidence-backed
(`specs/EXPERIMENTS.md`). Don't "optimize" them without re-running the gate.

## Docs & decisions
- Specs live in **`specs/`**, user docs in **`docs/`** — never the repo root.
- Record every non-trivial decision in **`specs/DEVLOG.md`** (see `log-decision`).

## Checkpoints & commits
Work in small checkpoints; **commit after each** (commit, don't push unless asked).
A checkpoint is green when gofmt is clean, vet passes, and tests pass. Imperative
subject + a body explaining the *why*.
