package leakybucket

import (
	"testing"
	"time"
)

// TestCounterBurstThenRefill 验证 §4.1：容量内可突发放行（水位不满
// 立即放行），满后拒绝，随时间流逝漏水恢复。
func TestCounterBurstThenRefill(t *testing.T) {
	b := NewCounterLeakyBucket(10, 5) // rate=10/s，capacity=5

	for i := 0; i < 5; i++ {
		if !b.Allow() {
			t.Fatalf("allow #%d should pass: bucket has capacity", i+1)
		}
	}
	if b.Allow() {
		t.Error("6th allow should be rejected: bucket full")
	}

	time.Sleep(100 * time.Millisecond) // 漏掉 1 个（10/s × 0.1s）
	if !b.Allow() {
		t.Error("allow after 100ms should pass: 1 unit leaked out")
	}
}

// TestCounterRate 验证放行速率 ≈ rate：rate=100/s 时 1 秒内放行
// 约 100 个（±20% 容差）。
func TestCounterRate(t *testing.T) {
	b := NewCounterLeakyBucket(100, 10)

	var allowed int
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if b.Allow() {
			allowed++
		}
		time.Sleep(2 * time.Millisecond) // 每秒 500 次尝试，远高于 rate
	}

	if allowed < 80 || allowed > 120 {
		t.Errorf("allowed = %d in 1s, want ≈ 100 (rate=100/s)", allowed)
	}
}
