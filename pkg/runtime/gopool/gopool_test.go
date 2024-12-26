package gopool

import (
	"sync"
	"testing"
)

// TestGoPool 测试协程池
func TestGoPool(t *testing.T) {
	p := NewPool()
	p.SetLimit(10)

	num := 0
	mu := sync.Mutex{}
	for i := 0; i < 1000; i++ {
		idx := i
		p.Go(func() error {
			mu.Lock()
			defer mu.Unlock()

			num++

			if idx == 0 {
				panic("test panic")
			}

			return nil
		})
	}

	if err := p.Wait(); err != nil {
		t.Error(err)
		return
	}

	if num != 1000 {
		t.Logf("num != 1000, num: %d", num)
	}
	return
}
