package research

import "testing"

func TestChunkPacksParagraphsToBudget(t *testing.T) {
	// three paragraphs of 2,2,1 words; budget 3 -> two passages.
	text := "alpha beta\ngamma delta\nepsilon"
	got := ChunkText(text, "http://a", 3)
	if len(got) != 2 {
		t.Fatalf("want 2 passages, got %d: %+v", len(got), got)
	}
	for _, p := range got {
		if p.SourceURL != "http://a" {
			t.Errorf("passage not tied to source: %+v", p)
		}
	}
}

func TestChunkLongParagraphStaysWhole(t *testing.T) {
	text := "one two three four five six seven" // 7 words, budget 3
	got := ChunkText(text, "u", 3)
	if len(got) != 1 || got[0].Text != text {
		t.Fatalf("a single long paragraph should stay whole, got %+v", got)
	}
}

func TestChunkEmptyInput(t *testing.T) {
	if got := ChunkText("   \n  \n", "u", 10); len(got) != 0 {
		t.Fatalf("want no passages, got %d", len(got))
	}
}
