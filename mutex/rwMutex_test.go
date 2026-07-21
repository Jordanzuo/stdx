package mutex

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestRWMutex_SetData(t *testing.T) {
	m := NewRWMutex(0)
	m.SetData(42)

	m.ReadData(func(v int) {
		if v != 42 {
			t.Fatalf("got %d, want 42", v)
		}
	})
}

func TestRWMutex_ReadData_concurrent(t *testing.T) {
	const n = 100
	m := NewRWMutex(&player{Score: 0})
	var reads atomic.Int64
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			m.ReadData(func(p *player) {
				reads.Add(1)
				_ = p.Score
			})
		}()
	}

	wg.Wait()

	if reads.Load() != n {
		t.Fatalf("got %d reads, want %d", reads.Load(), n)
	}
}

func TestRWMutex_WriteData_concurrent(t *testing.T) {
	const n = 100
	m := NewRWMutex(&player{Score: 0})
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			m.WriteData(func(p *player) {
				p.Score++
			})
		}()
	}

	wg.Wait()

	m.ReadData(func(p *player) {
		if p.Score != n {
			t.Fatalf("got score %d, want %d", p.Score, n)
		}
	})
}

func TestRWMutex_ReadWrite_concurrent(t *testing.T) {
	const writers = 50
	const readers = 50
	m := NewRWMutex(&player{Score: 0})
	var wg sync.WaitGroup
	wg.Add(writers + readers)

	for i := 0; i < writers; i++ {
		go func() {
			defer wg.Done()
			m.WriteData(func(p *player) {
				p.Score++
			})
		}()
	}

	for i := 0; i < readers; i++ {
		go func() {
			defer wg.Done()
			m.ReadData(func(p *player) {
				_ = p.Score
			})
		}()
	}

	wg.Wait()

	m.ReadData(func(p *player) {
		if p.Score != writers {
			t.Fatalf("got score %d, want %d", p.Score, writers)
		}
	})
}
