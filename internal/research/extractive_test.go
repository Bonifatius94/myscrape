package research

import (
	"strings"
	"testing"
)

func TestExtractiveConcatenatesWithCitations(t *testing.T) {
	ps := []Passage{
		{Text: "alpha content", SourceURL: "http://a"},
		{Text: "beta content", SourceURL: "http://a"},
		{Text: "gamma content", SourceURL: "http://b"},
	}
	r := ExtractiveSynthesize("q", ps)

	if len(r.Citations) != 2 { // one per distinct source, first-seen order
		t.Fatalf("want 2 citations, got %d", len(r.Citations))
	}
	if r.Citations[0].Marker != "[1]" || r.Citations[0].URL != "http://a" {
		t.Errorf("citation[0] = %+v", r.Citations[0])
	}
	if !strings.Contains(r.Answer, "[1] http://a") || !strings.Contains(r.Answer, "gamma content") {
		t.Errorf("answer missing source block or text: %q", r.Answer)
	}
	if r.Coverage.Confidence != "medium" { // 2 distinct sources
		t.Errorf("coverage = %+v", r.Coverage)
	}
}

func TestExtractiveNoPassages(t *testing.T) {
	r := ExtractiveSynthesize("q", nil)
	if len(r.Citations) != 0 || r.Coverage.Confidence != "low" || r.Answer == "" {
		t.Fatalf("empty synthesis wrong: %+v", r)
	}
}
