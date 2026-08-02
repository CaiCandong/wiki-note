package leakybucket

import (
	"math"
	"sync"
	"time"
)

// CounterLeakyBucket 计算式漏桶（非阻塞判定）：时间积分模拟水量，无队列。
// 对应笔记 §4.1。
type CounterLeakyBucket struct {
	mu       sync.Mutex
	rate     float64 // 每秒放行速率
	capacity float64 // 桶容量
	water    float64 // 当前积压水量
	last     time.Time
}

// NewCounterLeakyBucket 新建计算式漏桶。
func NewCounterLeakyBucket(rate, capacity float64) *CounterLeakyBucket {
	return &CounterLeakyBucket{rate: rate, capacity: capacity, last: time.Now()}
}

// Allow 请求来了直接判定放行或拒绝，不排队，O(1) 且延迟恒定。
func (b *CounterLeakyBucket) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	// 先按流逝时间漏水，释放容量（水最少漏到 0）
	b.water = math.Max(0, b.water-(now.Sub(b.last).Seconds()*b.rate))
	b.last = now

	if b.water+1 <= b.capacity { // 加水后不超容量：放行并加水（避免浮点边界下水位突破容量）
		b.water++
		return true
	}
	return false // 桶满：拒绝
}
