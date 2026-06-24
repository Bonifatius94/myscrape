package research

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Bonifatius94/myscrape/internal/search"
)

// blockingSearcher signals each entry and blocks until release is closed, so a
// test can observe how many Research calls are in flight at once.
type blockingSearcher struct {
	entered chan struct{}
	release chan struct{}
}

func (b *blockingSearcher) Search(_ context.Context, _ string, _ int) ([]search.Result, error) {
	b.entered <- struct{}{}
	<-b.release
	return nil, nil
}

func TestResearchConcurrencyCap(t *testing.T) {
	bs := &blockingSearcher{entered: make(chan struct{}, 5), release: make(chan struct{})}
	wr := NewWebResearcher(bs, &fakeFetcher{byURL: map[string]string{}}, nil, "simple", 2)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = wr.Research(context.Background(), "q", "quick", "")
		}()
	}

	// Exactly the cap (2) should reach Search; the rest block on the semaphore.
	<-bs.entered
	<-bs.entered
	select {
	case <-bs.entered:
		t.Fatal("a 3rd research entered Search before a slot freed — cap not enforced")
	case <-time.After(100 * time.Millisecond):
	}

	close(bs.release) // let everyone proceed
	wg.Wait()
}
