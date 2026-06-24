package search

import (
	"testing"

	"github.com/Bonifatius94/myscrape/internal/config"
)

func TestBuildRoundRobinDefault(t *testing.T) {
	p, err := Build(nil, config.Settings{SearchProvider: "roundrobin", MarginaliaAPIKey: "public"})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	rr, ok := p.(*RoundRobin)
	if !ok {
		t.Fatalf("want *RoundRobin, got %T", p)
	}
	if len(rr.providers) != 2 { // ddg + marginalia, no keys
		t.Fatalf("want 2 base engines, got %d", len(rr.providers))
	}
}

func TestBuildComposesKeyedEngines(t *testing.T) {
	p, _ := Build(nil, config.Settings{
		SearchProvider: "roundrobin",
		TavilyAPIKey:   "t", ExaAPIKey: "e", SerpAPIKey: "sa",
		SerperAPIKey: "sp", MojeekAPIKey: "m",
		GoogleAPIKey: "g", GoogleCSEID: "cx",
	})
	rr := p.(*RoundRobin)
	if len(rr.providers) != 8 { // ddg + marginalia + 6 keyed (tavily/exa/serpapi/serper/mojeek/google_cse)
		t.Fatalf("want 8 engines, got %d", len(rr.providers))
	}
}

func TestBuildSingleEngine(t *testing.T) {
	for _, name := range []string{"ddg", "marginalia", "tavily", "exa", "serpapi", "serper", "mojeek", "google_cse"} {
		p, err := Build(nil, config.Settings{SearchProvider: name})
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if p.Name() != name {
			t.Errorf("want %q, got %q", name, p.Name())
		}
	}
}

func TestBuildUnknownEngineErrors(t *testing.T) {
	if _, err := Build(nil, config.Settings{SearchProvider: "altavista"}); err == nil {
		t.Fatal("unknown engine should error")
	}
}
