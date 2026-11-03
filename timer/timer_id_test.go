package timer

import (
	"sync"
	"testing"
)

func TestNewGlobalId(t *testing.T) {
	setGlobalId(0)
	if got := newGlobalId(); got != 1 {
		t.Fatalf("first id = %d, want 1", got)
	}
	if got := globalId.Load(); got != 1 {
		t.Fatalf("globalId = %d, want 1", got)
	}
	if got := newGlobalId(); got != 2 {
		t.Fatalf("second id = %d, want 2", got)
	}
}

func TestNewGlobalIdConcurrent(t *testing.T) {
	setGlobalId(0)
	const count = 100
	ids := make(chan int64, count)
	var wg sync.WaitGroup
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			ids <- newGlobalId()
		}()
	}
	wg.Wait()
	close(ids)

	seen := make(map[int64]struct{}, count)
	for id := range ids {
		if _, exists := seen[id]; exists {
			t.Fatalf("duplicate id %d", id)
		}
		seen[id] = struct{}{}
	}
	if len(seen) != count {
		t.Fatalf("id count = %d, want %d", len(seen), count)
	}
}
