package research

import (
	"fmt"
	"strings"
)

// ExtractiveSynthesize builds a cited answer from passages without an LLM: each
// distinct source becomes a numbered, cited block. GPU-free (Python D-017).
func ExtractiveSynthesize(question string, passages []Passage) SynthesisResult {
	sources := distinctSources(passages)
	if len(sources) == 0 {
		return SynthesisResult{
			Answer:   "No usable sources were found for this question.",
			Coverage: coverageFor(0),
		}
	}

	blocks := make([]string, 0, len(sources))
	citations := make([]Citation, 0, len(sources))
	for i, url := range sources {
		var texts []string
		for _, p := range passages {
			if p.SourceURL == url {
				texts = append(texts, p.Text)
			}
		}
		blocks = append(blocks, fmt.Sprintf("[%d] %s\n%s", i+1, url, strings.Join(texts, "\n\n")))
		citations = append(citations, Citation{Marker: fmt.Sprintf("[%d]", i+1), URL: url})
	}
	return SynthesisResult{
		Answer:    strings.Join(blocks, "\n\n"),
		Citations: citations,
		Coverage:  coverageFor(len(sources)),
	}
}

func distinctSources(passages []Passage) []string {
	seen := make(map[string]bool)
	var order []string
	for _, p := range passages {
		if !seen[p.SourceURL] {
			seen[p.SourceURL] = true
			order = append(order, p.SourceURL)
		}
	}
	return order
}

func coverageFor(n int) Coverage {
	switch {
	case n >= 3:
		return Coverage{Confidence: "high", Note: fmt.Sprintf("%d corroborating sources", n)}
	case n == 2:
		return Coverage{Confidence: "medium", Note: "two sources"}
	case n == 1:
		return Coverage{Confidence: "low", Note: "single source"}
	default:
		return Coverage{Confidence: "low", Note: "no sources found"}
	}
}
