package fetch

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

// ChromeRenderer renders JS pages with a headless Chrome via chromedp. It needs a
// Chrome/Chromium binary on PATH; if absent, Render returns an error and the
// Fetcher falls back to the static result. This is the one unstable dependency —
// isolated behind the Renderer interface and never exercised in unit tests.
type ChromeRenderer struct {
	timeout time.Duration
}

// NewChromeRenderer builds a renderer with the given per-page timeout.
func NewChromeRenderer(timeout time.Duration) *ChromeRenderer {
	return &ChromeRenderer{timeout: timeout}
}

// Render navigates to url in a headless browser and returns the rendered HTML.
func (c *ChromeRenderer) Render(ctx context.Context, url string) (string, error) {
	browserCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()
	runCtx, timeoutCancel := context.WithTimeout(browserCtx, c.timeout)
	defer timeoutCancel()

	var rendered string
	err := chromedp.Run(runCtx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.OuterHTML("html", &rendered),
	)
	return rendered, err
}
