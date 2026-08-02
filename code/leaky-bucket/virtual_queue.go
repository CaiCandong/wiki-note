package leakybucket

import (
	"sync"
	"time"
)

// VirtualQueuePipeline 虚拟队列版推送管线（笔记 §4.2/§5.4 思路）：
// 不建内存队列、把请求在时间轴上排队——每个 worker 循环：
//
//	漏桶 Take（时间轴排队）→ 从持久化队列拉一条 → 发送 → 循环
//
// 关键语义：
//   - 消息不丢：先 Take 后拉取——等待放行期间消息留在持久化队列里，
//     进程崩溃不丢（等待中的请求只占 goroutine，不持有消息）
//   - 放行节奏全局一致：多 pod 时所有 worker 共享同一个 Redis 漏桶
//     （DistributedBucket），Lua 原子保证总放行速率 = rate
//   - worker 数 = 并发在途 ≈ rate × 单条耗时（Little's Law）
//
// fetch 是"从持久化队列拉一条消息"的注入点：演示/测试用内存 channel
// 模拟 MQ，生产注入 RabbitMQ 消费（prefetch=1 语义天然配合）。
type VirtualQueuePipeline struct {
	bucket    rateLimiter            // 虚拟队列漏桶（单机 or Redis 分布式）
	fetch     func() (string, bool)  // 从持久化队列拉一条消息；空返回 false
	sent      chan string            // 发送出口（模拟厂商 API 调用成功）
	sendDelay time.Duration          // 模拟单条发送耗时
	stop      chan struct{}
	once      sync.Once
	wg        sync.WaitGroup
}

// NewVirtualQueuePipeline 新建虚拟队列版推送管线：workers 个 worker
// 并发在时间轴上排队放行；bucket 传共享的分布式漏桶（多 pod）或单机
// 漏桶（演示）。
func NewVirtualQueuePipeline(bucket rateLimiter, workers int, sendDelay time.Duration, fetch func() (string, bool)) *VirtualQueuePipeline {
	p := &VirtualQueuePipeline{
		bucket:    bucket,
		fetch:     fetch,
		sent:      make(chan string, 4096),
		sendDelay: sendDelay,
		stop:      make(chan struct{}),
	}
	p.wg.Add(workers)
	for i := 0; i < workers; i++ {
		go p.worker()
	}
	return p
}

// worker 单个发送协程：时间轴排队等放行，放行后拉取消息立即发送。
func (p *VirtualQueuePipeline) worker() {
	defer p.wg.Done()
	for {
		select {
		case <-p.stop:
			return
		default:
		}
		// 时间轴排队：等放行时刻（Redis Lua 或单机计算式），
		// 等待期间不持有任何消息——消息还在持久化队列里
		if err := p.bucket.Take(); err != nil {
			continue // 被拒：等下一轮放行
		}
		msg, ok := p.fetch()
		if !ok {
			continue // 持久化队列空：空转一个放行 slot
		}
		if p.sendDelay > 0 {
			time.Sleep(p.sendDelay) // 模拟厂商单条耗时
		}
		select {
		case p.sent <- msg:
		case <-p.stop:
			return
		}
	}
}

// Sent 返回发送出口；管线 Stop 后关闭。
func (p *VirtualQueuePipeline) Sent() <-chan string { return p.sent }

// Stop 关闭管线：等全部 worker 退出后关闭发送出口。
func (p *VirtualQueuePipeline) Stop() {
	p.once.Do(func() {
		close(p.stop)
		p.wg.Wait()
		close(p.sent)
	})
}
