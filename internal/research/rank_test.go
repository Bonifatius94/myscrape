package research

import (
	"strings"
	"testing"
)

func TestBM25RanksRelevantAboveIrrelevant(t *testing.T) {
	passages := []Passage{
		{Text: "python asyncio taskgroup runs tasks concurrently", SourceURL: "a"},
		{Text: "banana bread recipe with flour and sugar", SourceURL: "b"},
		{Text: "the asyncio event loop in python schedules coroutines", SourceURL: "c"},
	}
	top := RankPassages("python asyncio", passages, 2)

	if len(top) != 2 {
		t.Fatalf("want top 2, got %d", len(top))
	}
	for _, p := range top {
		if strings.Contains(p.Text, "banana") {
			t.Errorf("irrelevant passage ranked into top: %q", p.Text)
		}
	}
}

func TestRankEmptyAndTopKClamp(t *testing.T) {
	if got := RankPassages("q", nil, 5); len(got) != 0 {
		t.Fatalf("nil passages -> empty, got %d", len(got))
	}
	one := []Passage{{Text: "q here", SourceURL: "a"}}
	if got := RankPassages("q", one, 5); len(got) != 1 {
		t.Fatalf("topK clamps to len, got %d", len(got))
	}
}
