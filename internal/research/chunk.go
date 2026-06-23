package research

import (
	"regexp"
	"strings"
)

var paragraphSplit = regexp.MustCompile(`\n+`)

// ChunkText packs paragraphs into passages of at most maxWords words (a single
// long paragraph stays whole), each tied to sourceURL. Empty input -> nil.
func ChunkText(text, sourceURL string, maxWords int) []Passage {
	var paras []string
	for _, p := range paragraphSplit.Split(text, -1) {
		if t := strings.TrimSpace(p); t != "" {
			paras = append(paras, t)
		}
	}

	var out []Passage
	var current []string
	curWords := 0
	flush := func() {
		if len(current) > 0 {
			out = append(out, Passage{Text: strings.Join(current, "\n\n"), SourceURL: sourceURL})
			current, curWords = nil, 0
		}
	}
	for _, p := range paras {
		w := len(strings.Fields(p))
		if len(current) > 0 && curWords+w > maxWords {
			flush()
		}
		current = append(current, p)
		curWords += w
	}
	flush()
	return out
}
