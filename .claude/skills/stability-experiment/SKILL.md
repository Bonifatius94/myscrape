---
name: stability-experiment
description: How to investigate a stability/rate-limiting issue in myscrape-go rigorously. Use whenever search or fetch starts failing intermittently, gets rate-limited (202/429/503), or behaves non-deterministically — and you need to attribute the cause to a real factor, not luck.
---

# Running a stability experiment

This is the method that established the operating point (`specs/EXPERIMENTS.md`).
Reuse it; don't improvise.

## Tool
- **`go run ./cmd/bench`** — runs the query set through the full httpx stack (cache
  + per-host rate limit + retry) and reports `ratelimit_rate`. The gate passes only
  when it's `0`. For isolating an endpoint, a few raw `http.Get`s at a fixed spacing
  are enough — the point is controlled, attributable observation.

## The method (in order)
1. **Check the persistence confound FIRST.** Measure how long a rate-limit lasts
   before any comparative test. If it persists for hours, back-to-back runs aren't
   attributable to code — a "fix" may just be a lapsed limit. Probe isolated
   requests over time right after a limited run; proceed only if recovery is short
   (minutes — which is what the reference found for DDG).
2. **Isolate one variable** (spacing, UA, endpoint, headers) — change exactly one.
3. **Prefer slowing down over evasion.** Stability over throughput is the rule.
   Longer intervals / more jitter before anything clever; never proxies/CAPTCHA
   (out of scope — the honest fallback is a keyed API).
4. **Validate on the full stack** (`cmd/bench`), not just an isolated probe.
5. **Keep the gate aligned with the strategy** — if you pace deliberately, the gate
   must expect the pace, not flag it.
6. **Bake the proven operating point into config defaults.**

## Documenting (required)
Summarize the finding in `specs/EXPERIMENTS.md` (Hypothesis · Setup · Observations ·
Results · Insight) and log the resulting decision via `log-decision`. Stop when the
gate passes with margin — not when it's theoretically optimal.
