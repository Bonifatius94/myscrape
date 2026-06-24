package research

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Bonifatius94/myscrape/internal/llm"
)

// ErrLLM wraps a failure from the LLM synthesis call (maps to LLM_ERROR).
var ErrLLM = errors.New("llm synthesis failed")

var markerRe = regexp.MustCompile(`\[(\d+)\]`)

const systemPrompt = "You are a careful research assistant. Answer the question " +
	"using ONLY the numbered sources provided. Cite every claim with its source " +
	"marker like [1]. Be concise. If the sources are insufficient or conflict, say " +
	"so plainly. Do not invent facts or sources."

// LLMSynthesizer writes a cited answer with one LLM call. Citations are derived
// deterministically from the [n] markers the model emits (we don't trust it to
// emit structured JSON), so a weak local model only has to write prose.
type LLMSynthesizer struct {
	chat        llm.ChatClient
	temperature float64
	maxTokens   int
}

// NewLLMSynthesizer builds a synthesizer over the chat client.
func NewLLMSynthesizer(chat llm.ChatClient) *LLMSynthesizer {
	return &LLMSynthesizer{chat: chat, temperature: 0.2, maxTokens: 1024}
}

func (s *LLMSynthesizer) Synthesize(ctx context.Context, question string, passages []Passage) (SynthesisResult, error) {
	sources := distinctSources(passages)
	user := buildPrompt(question, sources, passages)

	answer, err := s.chat.Complete(ctx, []llm.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: user},
	}, s.temperature, s.maxTokens)
	if err != nil {
		return SynthesisResult{}, fmt.Errorf("%w: %v", ErrLLM, err)
	}
	answer = strings.TrimSpace(answer)

	used := map[int]bool{}
	for _, m := range markerRe.FindAllStringSubmatch(answer, -1) {
		if n, err := strconv.Atoi(m[1]); err == nil {
			used[n] = true
		}
	}
	nums := make([]int, 0, len(used))
	for n := range used {
		if n >= 1 && n <= len(sources) {
			nums = append(nums, n)
		}
	}
	sort.Ints(nums)
	citations := make([]Citation, 0, len(nums))
	for _, n := range nums {
		citations = append(citations, Citation{Marker: fmt.Sprintf("[%d]", n), URL: sources[n-1]})
	}
	return SynthesisResult{Answer: answer, Citations: citations, Coverage: coverageFor(len(sources))}, nil
}

func buildPrompt(question string, sources []string, passages []Passage) string {
	blocks := make([]string, 0, len(sources))
	for i, url := range sources {
		var texts []string
		for _, p := range passages {
			if p.SourceURL == url {
				texts = append(texts, p.Text)
			}
		}
		blocks = append(blocks, fmt.Sprintf("[%d] %s\n%s", i+1, url, strings.Join(texts, "\n")))
	}
	return fmt.Sprintf(
		"Question: %s\n\nSources:\n%s\n\nWrite a concise, cited answer using [n] markers that refer to the sources above.",
		question, strings.Join(blocks, "\n\n"),
	)
}
