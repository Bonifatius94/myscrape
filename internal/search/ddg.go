package search

import (
	"bytes"
	"context"
	"net/url"
	"strings"

	"golang.org/x/net/html"

	"github.com/Bonifatius94/myscrape-go/internal/httpx"
)

// DDG scrapes DuckDuckGo's HTML endpoint — the one scraper we keep (no key). It's
// rate-limit-prone, but the round-robin circuit breaker routes around it when it
// cools. Result links are DDG redirects (/l/?uddg=...) that we decode back.
const ddgEndpoint = "https://html.duckduckgo.com/html/"

type DDG struct {
	http httpx.Doer
}

func NewDDG(h httpx.Doer) *DDG { return &DDG{http: h} }
func (d *DDG) Name() string    { return "ddg" }

func (d *DDG) Search(ctx context.Context, query string, maxResults int) ([]Result, error) {
	u := ddgEndpoint + "?" + url.Values{"q": {query}}.Encode()
	data, err := d.http.Get(ctx, u, nil)
	if err != nil {
		return nil, err
	}
	results, err := parseDDG(data)
	if err != nil {
		return nil, err
	}
	return capTo(results, maxResults), nil
}

func parseDDG(body []byte) ([]Result, error) {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	var out []Result
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			class := attrOf(n, "class")
			switch {
			case strings.Contains(class, "result__a"):
				if href := resolveDDGURL(attrOf(n, "href")); href != "" {
					out = append(out, newResult(len(out)+1, textOf(n), href, "", ""))
				}
			case strings.Contains(class, "result__snippet") && len(out) > 0:
				out[len(out)-1].Snippet = textOf(n)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return out, nil
}

// resolveDDGURL turns a DDG result href into the real target URL.
func resolveDDGURL(href string) string {
	if href == "" {
		return ""
	}
	if i := strings.Index(href, "uddg="); i >= 0 {
		raw := href[i+len("uddg="):]
		if amp := strings.IndexByte(raw, '&'); amp >= 0 {
			raw = raw[:amp]
		}
		if dec, err := url.QueryUnescape(raw); err == nil {
			return dec
		}
	}
	if strings.HasPrefix(href, "//") {
		return "https:" + href
	}
	return href
}

func attrOf(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func textOf(n *html.Node) string {
	var sb strings.Builder
	var f func(*html.Node)
	f = func(x *html.Node) {
		if x.Type == html.TextNode {
			sb.WriteString(x.Data)
		}
		for c := x.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return strings.TrimSpace(sb.String())
}
