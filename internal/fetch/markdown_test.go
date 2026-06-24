package fetch

import (
	"strings"
	"testing"
)

func TestHTMLToMarkdown(t *testing.T) {
	md := htmlToMarkdown("<h1>Title</h1><p>Body <strong>text</strong> here.</p><ul><li>one</li><li>two</li></ul>")
	if !strings.Contains(md, "# Title") {
		t.Errorf("heading not converted: %q", md)
	}
	if !strings.Contains(md, "Body") || !strings.Contains(md, "text") {
		t.Errorf("body text lost: %q", md)
	}
	if !strings.Contains(md, "one") || !strings.Contains(md, "two") {
		t.Errorf("list items lost: %q", md)
	}
}

func TestHTMLToMarkdownEmpty(t *testing.T) {
	if md := htmlToMarkdown(""); md != "" {
		t.Errorf("empty html -> empty md, got %q", md)
	}
}
