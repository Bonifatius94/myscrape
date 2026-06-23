package search

import "testing"

const ddgHTML = `<html><body>
<div class="result results_links_deep web-result">
  <h2 class="result__title">
    <a class="result__a" href="//duckduckgo.com/l/?uddg=https%3A%2F%2Fexample.com%2Fasync&amp;rut=abc">Async Guide</a>
  </h2>
  <a class="result__snippet" href="x">A guide to python async.</a>
</div>
<div class="result web-result">
  <a class="result__a" href="https://direct.example.org/page">Direct Title</a>
  <a class="result__snippet">Second snippet.</a>
</div>
</body></html>`

func TestParseDDG(t *testing.T) {
	got, err := parseDDG([]byte(ddgHTML))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 results, got %d: %+v", len(got), got)
	}
	// redirect href is decoded back to the real URL
	if got[0].URL != "https://example.com/async" {
		t.Errorf("uddg not decoded: %q", got[0].URL)
	}
	if got[0].Title != "Async Guide" || got[0].Snippet != "A guide to python async." {
		t.Errorf("result[0] = %+v", got[0])
	}
	// direct href passes through
	if got[1].URL != "https://direct.example.org/page" || got[1].Snippet != "Second snippet." {
		t.Errorf("result[1] = %+v", got[1])
	}
}

func TestResolveDDGURL(t *testing.T) {
	cases := map[string]string{
		"//duckduckgo.com/l/?uddg=https%3A%2F%2Fa.com%2Fx&rut=z": "https://a.com/x",
		"https://b.com/direct": "https://b.com/direct",
		"//cdn.example.com/p":  "https://cdn.example.com/p",
		"":                     "",
	}
	for in, want := range cases {
		if got := resolveDDGURL(in); got != want {
			t.Errorf("resolveDDGURL(%q) = %q, want %q", in, got, want)
		}
	}
}
