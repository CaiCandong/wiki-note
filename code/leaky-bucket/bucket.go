// Package leakybucket 是笔记《漏桶算法全面指南（Go 实现）》中各实现的可运行版本与测试。
//
// 笔记：content/rate-limit-leaky-bucket.md，章节对应：
//   - LeakyBucket         §3 队列式漏桶
//   - CounterLeakyBucket  §4.1 计算式漏桶（非阻塞判定）
//   - BlockingBucket      §4.2 计算式漏桶（阻塞虚拟队列）
//   - DistributedBucket   §5.1 分布式虚拟队列（Redis Lua）
//   - ListBucket          §5.2 分布式队列式（Redis List 惰性补账）
package leakybucket

import (
	"sync"
	"time"
)

// LeakyBucket 队列式漏桶：请求（job）进桶排队，恒定速率出桶执行。
// 对应笔记 §3。
type LeakyBucket struct {
	queue chan func()    // 桶：capacity 即最大积压
	rate  time.Duration  // 漏口：每个请求的放行间隔
	stop  chan struct{}  // 关闭信号
	once  sync.Once      // Stop 只生效一次，重复调用不 panic
	wg    sync.WaitGroup // 等待 drain 退出，优雅关闭用
}

// NewLeakyBucket 新建队列式漏桶并启动后台 drain goroutine。
func NewLeakyBucket(capacity int, rate time.Duration) *LeakyBucket {
	lb := &LeakyBucket{queue: make(chan func(), capacity), rate: rate, stop: make(chan struct{})}
	lb.wg.Add(1)
	go func() {
		defer lb.wg.Done()
		lb.drain()
	}()
	return lb
}

// drain 定时器驱动，恒定速率取出一个请求执行；收到关闭信号退出。
func (lb *LeakyBucket) drain() {
	ticker := time.NewTicker(lb.rate)
	defer ticker.Stop()
	for {
		select {
		case <-lb.stop:
			return
		case <-ticker.C:
			select {
			case job := <-lb.queue:
				job() // 以固定速率执行
			default: // 桶空，空转
			}
		}
	}
}

// Submit 入桶排队；桶满即拒绝（溢出丢弃），不阻塞调用方。
func (lb *LeakyBucket) Submit(job func()) bool {
	select {
	case lb.queue <- job:
		return true
	default:
		return false
	}
}

// Stop 关闭出桶，等 drain 退出后返回；队列里剩余的 job 不再执行。
func (lb *LeakyBucket) Stop() {
	lb.once.Do(func() { close(lb.stop) })
	lb.wg.Wait()
}
