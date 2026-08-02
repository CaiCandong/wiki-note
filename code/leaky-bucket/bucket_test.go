package leakybucket

import (
	"testing"
	"time"

	"go.uber.org/goleak"
)

// TestQueueConstantRate 验证笔记 §3 的核心特性：出桶速率 = rate，
// 与上游来速无关，输出恒定（允许 5 个 job 同时入桶，但执行间隔
// 严格等于 rate）。
func TestQueueConstantRate(t *testing.T) {
	const rate = 40 * time.Millisecond
	lb := NewLeakyBucket(10, rate)
	defer lb.Stop()

	const n = 5
	executed := make(chan time.Time, n)
	for i := 0; i < n; i++ {
		if !lb.Submit(func() { executed <- time.Now() }) {
			t.Fatal("submit should not be rejected, bucket not full")
		}
	}

	times := make([]time.Time, 0, n)
	for i := 0; i < n; i++ {
		select {
		case tt := <-executed:
			times = append(times, tt)
		case <-time.After(3 * time.Second):
			t.Fatalf("timeout: only %d/%d jobs executed", len(times), n)
		}
	}

	// 相邻放行间隔 ≈ rate（ticker 精度 + 调度抖动，给 ±15ms 容差）
	for i := 1; i < len(times); i++ {
		d := times[i].Sub(times[i-1])
		if d < rate/2 || d > rate+15*time.Millisecond {
			t.Errorf("interval[%d] = %v, want ≈ %v", i, d, rate)
		}
	}
	// 总时长 ≈ (n-1) * rate
	total := times[n-1].Sub(times[0])
	want := time.Duration(n-1) * rate
	if total < want-20*time.Millisecond || total > want+20*time.Millisecond {
		t.Errorf("total span = %v, want ≈ %v", total, want)
	}
}

// TestQueueFullRejected 验证桶满即拒：容量 2，第 3 个请求被拒绝
// （溢出丢弃），且不阻塞调用方。
func TestQueueFullRejected(t *testing.T) {
	lb := NewLeakyBucket(2, 20*time.Millisecond)
	defer lb.Stop()

	noop := func() {}
	for i := 0; i < 2; i++ {
		if !lb.Submit(noop) {
			t.Fatal("first two submits should succeed")
		}
	}
	if lb.Submit(noop) {
		t.Error("third submit should be rejected: bucket full")
	}
}

// TestQueueStop 验证 Stop 语义：drain 退出后剩余 job 不再执行，
// 重复 Stop 不 panic（sync.Once 幂等）。
func TestQueueStop(t *testing.T) {
	lb := NewLeakyBucket(5, time.Hour) // rate 极慢，job 永不执行
	ran := false
	lb.Submit(func() { ran = true })

	lb.Stop()
	lb.Stop() // 幂等

	time.Sleep(10 * time.Millisecond)
	if ran {
		t.Error("job should not run after Stop")
	}
}

// TestQueueNoLeak 验证 §3.3：drain goroutine 有停止机制（stop + wg），
// Stop 后无 goroutine 泄漏。
func TestQueueNoLeak(t *testing.T) {
	defer goleak.VerifyNone(t)

	lb := NewLeakyBucket(10, 10*time.Millisecond)
	lb.Submit(func() {})
	lb.Stop()
}
