package research

import (
	"context"
	"testing"

	"github.com/Bonifatius94/myscrape-go/internal/llm"
)

type fakeChat struct {
	reply       string
	gotMessages []llm.Message
}

func (f *fakeChat) Complete(_ context.Context, messages []llm.Message, _ float64, _ int) (string, error) {
	f.gotMessages = messages
	return f.reply, nil
}

func TestLLMSynthesizeCitesFromMarkers(t *testing.T) {
	ps := []Passage{
		{Text: "alpha", SourceURL: "http://a"},
		{Text: "beta", SourceURL: "http://b"},
	}
	chat := &fakeChat{reply: "Async is great [1] and groups help [2]."}
	s := NewLLMSynthesizer(chat)

	r, err := s.Synthesize(context.Background(), "q", ps)
	if err != nil {
		t.Fatalf("synthesize: %v", err)
	}
	if r.Answer != "Async is great [1] and groups help [2]." {
		t.Errorf("answer = %q", r.Answer)
	}
	if len(r.Citations) != 2 || r.Citations[0].URL != "http://a" || r.Citations[1].URL != "http://b" {
		t.Errorf("citations = %+v", r.Citations)
	}
	if r.Coverage.Confidence != "medium" {
		t.Errorf("coverage = %+v", r.Coverage)
	}
	// the system+user prompt is sent
	if len(chat.gotMessages) != 2 || chat.gotMessages[0].Role != "system" {
		t.Errorf("messages = %+v", chat.gotMessages)
	}
}

func TestLLMSynthesizeIgnoresOutOfRangeMarkers(t *testing.T) {
	ps := []Passage{{Text: "alpha", SourceURL: "http://a"}}
	chat := &fakeChat{reply: "text [1] and bogus [5]."}

	r, _ := NewLLMSynthesizer(chat).Synthesize(context.Background(), "q", ps)
	if len(r.Citations) != 1 || r.Citations[0].URL != "http://a" {
		t.Fatalf("out-of-range marker should be ignored, got %+v", r.Citations)
	}
}
