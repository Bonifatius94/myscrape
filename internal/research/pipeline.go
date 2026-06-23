package research

import (
	"context"
	"sync"

	"github.com/Bonifatius94/myscrape-go/internal/fetch"
	"github.com/Bonifatius94/myscrape-go/internal/search"
)

// Searcher, Fetcher, and Synthesizer are the seams the pipeline depends on (so
// it's tested with fakes, no network).
type Searcher interface {
	Search(ctx context.Context, query string, maxResults int) ([]search.Result, error)
}
type Fetcher interface {
	Fetch(ctx context.Context, url string) (*fetch.Result, error)
}
type Synthesizer interface {
	Synthesize(ctx context.Context, question string, passages []Passage) (SynthesisResult, error)
}

// effort -> (max sources fetched, max passages kept).
var effortLevels = map[string][2]int{
	"quick":    {3, 6},
	"standard": {5, 10},
	"deep":     {8, 16},
}

const passageWords = 120

// WebResearcher runs the cooked loop: search -> fetch -> chunk -> rank ->
// synthesize. Synthesis is "simple" (extractive, GPU-free) or "llm" (model-written)
// when an LLM synthesizer is configured.
type WebResearcher struct {
	search           Searcher
	fetch            Fetcher
	llm              Synthesizer // optional; nil -> always extractive
	defaultSynthesis string
}

// NewWebResearcher wires the pipeline. llm may be nil (extractive only);
// defaultSynthesis is the mode used when a call doesn't specify one.
func NewWebResearcher(s Searcher, f Fetcher, llm Synthesizer, defaultSynthesis string) *WebResearcher {
	if defaultSynthesis == "" {
		defaultSynthesis = "simple"
	}
	return &WebResearcher{search: s, fetch: f, llm: llm, defaultSynthesis: defaultSynthesis}
}

// Research executes the loop for the given effort/synthesis and returns a cited
// result. synthesis "" falls back to the server default.
func (w *WebResearcher) Research(ctx context.Context, question, effort, synthesis string) (ResearchResult, error) {
	lvl, ok := effortLevels[effort]
	if !ok {
		lvl = effortLevels["standard"]
	}
	maxSources, maxPassages := lvl[0], lvl[1]

	results, err := w.search.Search(ctx, question, maxSources)
	if err != nil {
		return ResearchResult{}, err
	}
	titleByURL := make(map[string]string, len(results))
	for _, r := range results {
		titleByURL[r.URL] = r.Title
	}

	// Fetch sources concurrently; a blocked/empty source is dropped, not fatal.
	pages := make([]*fetch.Result, len(results))
	var wg sync.WaitGroup
	for i, r := range results {
		wg.Add(1)
		go func(i int, url string) {
			defer wg.Done()
			if page, err := w.fetch.Fetch(ctx, url); err == nil {
				pages[i] = page // distinct index per goroutine: no race
			}
		}(i, r.URL)
	}
	wg.Wait()

	var passages []Passage
	for i, r := range results {
		if pages[i] != nil {
			// Key passages by the search URL so titles map cleanly.
			passages = append(passages, ChunkText(pages[i].Content, r.URL, passageWords)...)
		}
	}

	top := RankPassages(question, passages, maxPassages)

	mode := synthesis
	if mode == "" {
		mode = w.defaultSynthesis
	}
	var synth SynthesisResult
	if mode == "llm" && w.llm != nil {
		synth, err = w.llm.Synthesize(ctx, question, top)
		if err != nil {
			return ResearchResult{}, err
		}
	} else {
		synth = ExtractiveSynthesize(question, top)
	}

	citations := make([]Citation, len(synth.Citations))
	for i, c := range synth.Citations {
		citations[i] = Citation{Marker: c.Marker, URL: c.URL, Title: titleByURL[c.URL]}
	}
	var sources []Source
	seen := make(map[string]bool)
	for _, c := range citations {
		if !seen[c.URL] {
			seen[c.URL] = true
			sources = append(sources, Source{URL: c.URL, Title: titleByURL[c.URL]})
		}
	}

	return ResearchResult{
		Answer:    synth.Answer,
		Citations: citations,
		Sources:   sources,
		Coverage:  synth.Coverage,
	}, nil
}
