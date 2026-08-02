package leakybucket

import (
	"sync"
	"time"
)

// BlockingBucket 计算式漏桶（阻塞虚拟队列）：不建内存队列，按 perRequest
// 间隔排定放行时刻，请求阻塞等待，保证匀速输出。
// 对应笔记 §4.2（uber-go/ratelimit 思路）。
//
// 注意：本实现是教学代码，Take 在 mu 临界区内 time.Sleep（持锁 sleep，
// 反模式）——高并发下所有请求的 goroutine 会卡在 b.mu.Lock() 上排队，
// 数量随并发只增不减。生产实现应把 sleep 挪出临界区（如 uber 的 atomic
// 版：CAS 更新 state 后在锁外 Sleep）。行为仍被 TestBlockingSerialDrain
// 验证：放行时刻两两间隔 ≈ perRequest。
type BlockingBucket struct {
	mu         sync.Mutex
	perRequest time.Duration // 放行间隔 = time.Second / rate
	last       time.Time     // 上次放行时刻
}

// NewBlockingBucket 新建阻塞式虚拟队列漏桶，rate 为每秒放行数。
func NewBlockingBucket(rate int) *BlockingBucket {
	return &BlockingBucket{perRequest: time.Second / time.Duration(rate), last: time.Now()}
}

// Take 阻塞直到放行时刻再返回。
func (b *BlockingBucket) Take() {
	b.mu.Lock()
	defer b.mu.Unlock()
	next := b.last.Add(b.perRequest)
	if now := time.Now(); now.Before(next) {
		time.Sleep(next.Sub(now)) // 未到放行时刻：持锁阻塞等待（反模式，见注释）
		b.last = next
	} else {
		b.last = now
	}
}
