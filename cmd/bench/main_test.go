package main

import "testing"

func TestSummarize(t *testing.T) {
	r := summarize([]outcome{
		{ok: true}, {ok: true}, {rateLimited: true}, {}, // 2 ok, 1 rate-limited, 1 failed
	})
	if r.Total != 4 || r.OK != 2 || r.RateLimited != 1 || r.Failed != 1 {
		t.Fatalf("counts wrong: %+v", r)
	}
	if r.RateLimitRate != 0.25 || r.SuccessRate != 0.5 {
		t.Fatalf("rates wrong: %+v", r)
	}
}

func TestGate(t *testing.T) {
	clean := summarize([]outcome{{ok: true}, {ok: true}})
	if !clean.passed() {
		t.Error("zero rate-limited should pass")
	}
	limited := summarize([]outcome{{ok: true}, {rateLimited: true}})
	if limited.passed() {
		t.Error("any rate-limit should fail the gate")
	}
	if summarize(nil).passed() {
		t.Error("empty run should not pass")
	}
}
