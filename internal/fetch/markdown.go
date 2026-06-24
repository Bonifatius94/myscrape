package fetch

import (
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"golang.org/x/net/html"
)

// htmlToMarkdown converts an HTML fragment to markdown, returning "" on error.
func htmlToMarkdown(htmlStr string) string {
	if strings.TrimSpace(htmlStr) == "" {
		return ""
	}
	md, err := htmltomarkdown.ConvertString(htmlStr)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(md)
}

// nodeToMarkdown renders an html.Node back to HTML and converts it to markdown.
func nodeToMarkdown(n *html.Node) string {
	if n == nil {
		return ""
	}
	var sb strings.Builder
	if err := html.Render(&sb, n); err != nil {
		return ""
	}
	return htmlToMarkdown(sb.String())
}
