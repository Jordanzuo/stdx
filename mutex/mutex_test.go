package mutex

import (
	"sync"
	"testing"
)

func TestMutex_SetData(t *testing.T) {
	m := NewMutex(0)
	m.SetData(42)

	m.WriteData(func(v int) {
		if v != 42 {
			t.Fatalf("got %d, want 42", v)
		}
	})
}

func TestMutex_WriteData_pointerConcurrent(t *testing.T) {
	const n = 100
	m := NewMutex(&player{Score: 0})
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

	m.WriteData(func(p *player) {
		if p.Score != n {
			t.Fatalf("got score %d, want %d", p.Score, n)
		}
	})
}

func TestMutex_SetData_concurrentReplace(t *testing.T) {
	const n = 100
	m := NewMutex(0)
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(v int) {
			defer wg.Done()
			m.SetData(v + 1)
		}(i)
	}

	wg.Wait()

	m.WriteData(func(v int) {
		if v < 1 || v > n {
			t.Fatalf("got %d, want value in [1, %d]", v, n)
		}
	})
}

func TestMutex_WriteData_mapConcurrent(t *testing.T) {
	const n = 100
	m := NewMutex(make(map[int]int))
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(v int) {
			defer wg.Done()
			m.WriteData(func(values map[int]int) {
				values[v] = v
			})
		}(i)
	}

	wg.Wait()

	m.WriteData(func(values map[int]int) {
		if len(values) != n {
			t.Fatalf("got len %d, want %d", len(values), n)
		}
	})
}

func TestMutex_LockUnlock_valueConcurrent(t *testing.T) {
	const n = 100
	m := NewMutex(player{Score: 0})
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			p := m.Lock()
			p.Score++
			m.SetDataAfterLock(p)
			m.Unlock()
		}()
	}

	wg.Wait()

	p := m.Lock()
	score := p.Score
	m.Unlock()
	if score != n {
		t.Fatalf("got score %d, want %d", score, n)
	}
}

type player struct {
	Score int
}
