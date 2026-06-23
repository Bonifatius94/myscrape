package search

import "testing"

func TestParseExa(t *testing.T) {
	got, err := parseExa([]byte(`{"results":[
		{"title":"A","url":"https://a.com/x","text":"snippet","publishedDate":"2024-01-02"},
		{"title":"drop","url":"","text":"no url"}]}`))
	if err != nil || len(got) != 1 {
		t.Fatalf("got %d err %v", len(got), err)
	}
	if got[0].Site != "a.com" || got[0].Snippet != "snippet" || got[0].Published != "2024-01-02" {
		t.Errorf("%+v", got[0])
	}
}

func TestParseSerpAPI(t *testing.T) {
	got, err := parseSerpAPI([]byte(`{"organic_results":[
		{"title":"A","link":"https://a.com","snippet":"s","date":"x"},
		{"title":"d","link":"","snippet":"drop"}]}`))
	if err != nil || len(got) != 1 || got[0].URL != "https://a.com" {
		t.Fatalf("got %+v err %v", got, err)
	}
}

func TestParseSerper(t *testing.T) {
	got, err := parseSerper([]byte(`{"organic":[{"title":"A","link":"https://a.com","snippet":"s"}]}`))
	if err != nil || len(got) != 1 || got[0].Snippet != "s" {
		t.Fatalf("got %+v err %v", got, err)
	}
}

func TestParseMojeek(t *testing.T) {
	got, err := parseMojeek([]byte(`{"response":{"results":[{"title":"A","url":"https://a.com","desc":"d"}]}}`))
	if err != nil || len(got) != 1 || got[0].Snippet != "d" {
		t.Fatalf("got %+v err %v", got, err)
	}
}

func TestParseGoogleCSE(t *testing.T) {
	got, err := parseGoogleCSE([]byte(`{"items":[{"title":"A","link":"https://a.com","snippet":"s"}]}`))
	if err != nil || len(got) != 1 || got[0].URL != "https://a.com" {
		t.Fatalf("got %+v err %v", got, err)
	}
}

func TestParseEmptyPayloads(t *testing.T) {
	for _, fn := range []func([]byte) ([]Result, error){
		parseExa, parseSerpAPI, parseSerper, parseMojeek, parseGoogleCSE,
	} {
		if got, err := fn([]byte(`{}`)); err != nil || len(got) != 0 {
			t.Errorf("empty payload should yield no results, got %d err %v", len(got), err)
		}
	}
}
