package research

import (
	"math"
	"regexp"
	"sort"
	"strings"
)

var tokenRe = regexp.MustCompile(`[a-z0-9]+`)

func tokenize(s string) []string { return tokenRe.FindAllString(strings.ToLower(s), -1) }

// RankPassages scores passages against query with BM25 (k1=1.5, b=0.75) and
// returns the top-k, ties keeping original order. nil/empty -> nil.
func RankPassages(query string, passages []Passage, topK int) []Passage {
	if len(passages) == 0 {
		return nil
	}
	qTerms := tokenize(query)

	docs := make([]map[string]int, len(passages))
	lengths := make([]int, len(passages))
	df := map[string]int{}
	total := 0
	for i, p := range passages {
		tf := map[string]int{}
		toks := tokenize(p.Text)
		for _, t := range toks {
			tf[t]++
		}
		docs[i] = tf
		lengths[i] = len(toks)
		total += len(toks)
		for t := range tf {
			df[t]++
		}
	}

	n := float64(len(passages))
	avgdl := float64(total) / n
	if avgdl == 0 {
		avgdl = 1
	}
	const k1, b = 1.5, 0.75

	type scored struct {
		idx   int
		score float64
	}
	scores := make([]scored, len(passages))
	for i := range passages {
		var s float64
		dl := float64(lengths[i])
		for _, t := range qTerms {
			f := float64(docs[i][t])
			if f == 0 {
				continue
			}
			idf := math.Log(1 + (n-float64(df[t])+0.5)/(float64(df[t])+0.5))
			s += idf * (f * (k1 + 1)) / (f + k1*(1-b+b*dl/avgdl))
		}
		scores[i] = scored{i, s}
	}
	sort.SliceStable(scores, func(a, b int) bool { return scores[a].score > scores[b].score })

	if topK > len(scores) {
		topK = len(scores)
	}
	out := make([]Passage, 0, topK)
	for i := 0; i < topK; i++ {
		out = append(out, passages[scores[i].idx])
	}
	return out
}
