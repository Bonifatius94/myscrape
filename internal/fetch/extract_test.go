package fetch

import (
	"errors"
	"strings"
	"testing"
)

const sampleHTML = `<html><head><title>Async Guide</title></head><body>
<nav>home about contact menu junk navigation links</nav>
<article>
<h1>A Guide to Python Asyncio</h1>
<p>Python asyncio lets a single thread run many I/O-bound tasks concurrently by
cooperatively switching between them whenever one would otherwise block waiting on
the network or the filesystem. This avoids the overhead and complexity of threads.</p>
<p>The event loop is the core of asyncio. It schedules coroutines, runs callbacks,
and drives the whole cooperative-multitasking machine forward one step at a time as
awaited operations complete and resume the coroutines that were waiting on them.</p>
<p>TaskGroups, introduced in Python 3.11, supervise a set of child tasks. If any
child raises, the group cancels its siblings and propagates the error, which makes
structured concurrency far easier to reason about than a bare gather call.</p>
<p>In practice you await coroutines, wrap concurrent work in tasks, and let the
event loop interleave them. The result is dramatically higher throughput for
workloads dominated by waiting rather than by computation.</p>
</article>
<footer>copyright 2026 some company all rights reserved footer junk</footer>
</body></html>`

func TestExtractMainContent(t *testing.T) {
	res, err := Extract("https://example.com/guide", []byte(sampleHTML))
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if res.Title == "" {
		t.Errorf("want a title, got empty")
	}
	if !strings.Contains(res.Content, "TaskGroups") {
		t.Errorf("content missing main text; got: %q", res.Content)
	}
	if strings.Contains(res.Content, "navigation links") || strings.Contains(res.Content, "footer junk") {
		t.Errorf("boilerplate not stripped: %q", res.Content)
	}
	if res.WordCount == 0 {
		t.Errorf("want non-zero word count")
	}
	if res.FetchedVia != "static" {
		t.Errorf("fetchedVia = %q, want static", res.FetchedVia)
	}
}

func TestExtractEmptyReturnsErrEmpty(t *testing.T) {
	_, err := Extract("https://example.com", []byte("<html><body></body></html>"))
	if !errors.Is(err, ErrEmpty) {
		t.Fatalf("want ErrEmpty, got %v", err)
	}
}
