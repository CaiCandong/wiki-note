package leakybucket

import (
	"sort"
	"sync"
	"testing"
	"time"
)

// TestBlockingSerialDrain 验证 §4.2：并发请求被严格串行放行，
// 放行时刻两两间隔 ≈ perRequest（rate=50/s → 20ms）。
// 同时演示持锁 sleep 反模式的表象：所有请求按序逐个通过。
func TestBlockingSerialDrain(t *testing.T) {
	const (
		rate     = 50 // perRequest = 20ms
		requests = 5
	)
	b := NewBlockingBucket(rate)

	done := make(chan time.Time, requests)
	var wg sync.WaitGroup
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Take()
			done <- time.Now()
		}()
	}
	wg.Wait()
	close(done)

	times := make([]time.Time, 0, requests)
	for tt := range done {
		times = append(times, tt)
	}
	sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })

	for i := 1; i < len(times); i++ {
		d := times[i].Sub(times[i-1])
		if d < 10*time.Millisecond || d > 35*time.Millisecond {
			t.Errorf("interval[%d] = %v, want ≈ 20ms (serial release)", i, d)
		}
	}
}
