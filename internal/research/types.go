// Package research turns a question into a cited answer: chunk fetched pages into
// passages, BM25-rank them against the question, then synthesize. Phase 1 is
// extractive (no LLM, GPU-free).
package research

// Passage is a citable unit of text tied to its source URL.
type Passage struct {
	Text      string
	SourceURL string
}

// Citation maps a [n] marker to a source URL (and title once enriched).
type Citation struct {
	Marker string `json:"marker"`
	URL    string `json:"url"`
	Title  string `json:"title,omitempty"`
}

// Coverage is the "confident vs thin" signal.
type Coverage struct {
	Confidence string `json:"confidence"` // high | medium | low
	Note       string `json:"note"`
}

// SynthesisResult is the output of a synthesizer (extractive or LLM).
type SynthesisResult struct {
	Answer    string
	Citations []Citation
	Coverage  Coverage
}

// Source is a cited source with its title.
type Source struct {
	URL   string `json:"url"`
	Title string `json:"title,omitempty"`
}

// ResearchResult is the full web_research output.
type ResearchResult struct {
	Answer    string
	Citations []Citation
	Sources   []Source
	Coverage  Coverage
}
