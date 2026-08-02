package leakybucket

import (
	"context"
	"sync"
	"time"
)

// rateLimiter 漏桶放行接口：阻塞到放行时刻；返回错误表示被拒绝
// （ErrLimited）或等待被取消。
type rateLimiter interface {
	Take() error
}

// blockingAdapter 适配单机计算式漏桶（§4.2，单 pod 演示用）。
type blockingAdapter struct{ b *BlockingBucket }

func (a blockingAdapter) Take() error { a.b.Take(); return nil }

// distAdapter 适配分布式漏桶（§5.1 Redis Lua，生产多 pod 用）：
// 所有 pod 的 dispatcher 调同一个 key 的 Take，Lua 原子保证全局
// 总放行速率 = rate，不随 pod 数翻倍。
type distAdapter struct{ b *DistributedBucket }

func (a distAdapter) Take() error { return a.b.Take(context.Background()) }

// PushPipeline 推送场景组合（笔记 §5.4）：消息不能丢，但厂商 API
// 有硬性速率限制——漏桶不做「入口拒绝」，而做「出口调度」：
//
//	生产者 → [积压队列] → 漏桶（匀速放行） → worker pool → 发送厂商 API
//
// 关键参数关系（Little's Law）：厂商单条耗时 sendDelay 时，维持每秒
// rate 的吞吐需要 workers ≈ rate × sendDelay 个并发发送协程。
// 例：rate=500/s、sendDelay=50ms → workers ≈ 25。
//
// 结构：dispatcher 等漏桶放行时刻 → 从积压队列取一条 → 交给 worker
// 异步发送。放行路径绝不同步发送，否则被 sendDelay 拖垮放行节奏。
//
// 生产多 pod 形态：漏桶注入 DistributedBucket（Redis Lua）——所有 pod
// 共享同一放行节奏（总速率 = rate，不超限），每个 pod 自己的 worker
// pool 发送；单 pod 的 worker 上限不够时按 pod 数扩容（每 pod
// workers = 总在途 ÷ pod 数）。queue 演示用内存 channel，生产换成
// 持久化队列（RabbitMQ / Redis List）。
type PushPipeline struct {
	queue     chan string   // 积压队列：容量为允许积压上限
	bucket    rateLimiter   // 漏桶：放行节奏（单机或 Redis 分布式）
	sendJob   chan string   // 放行后待发送队列
	sent      chan string   // 发送完成出口（模拟厂商 API 调用成功）
	sendDelay time.Duration // 模拟单条发送耗时
	stop      chan struct{}
	once      sync.Once
	wg        sync.WaitGroup
}

// NewPushPipeline 新建推送管线（单机漏桶）：rate 为每秒放行条数。
func NewPushPipeline(rate, queueCap, workers int, sendDelay time.Duration) *PushPipeline {
	return NewPushPipelineWithBucket(blockingAdapter{NewBlockingBucket(rate)}, queueCap, workers, sendDelay)
}

// NewPushPipelineWithBucket 新建推送管线，漏桶实现可注入：
// 生产多 pod 场景传分布式漏桶（distAdapter 包 DistributedBucket，
// Redis Lua），保证多 pod 总放行速率 = rate。
func NewPushPipelineWithBucket(bucket rateLimiter, queueCap, workers int, sendDelay time.Duration) *PushPipeline {
	p := &PushPipeline{
		queue:     make(chan string, queueCap),
		bucket:    bucket,
		sendJob:   make(chan string, queueCap),
		sent:      make(chan string, queueCap),
		sendDelay: sendDelay,
		stop:      make(chan struct{}),
	}
	p.wg.Add(1 + workers)
	go p.run()
	for i := 0; i < workers; i++ {
		go p.worker()
	}
	return p
}

// run 漏桶调度器：等放行时刻，从积压队列取一条交给 worker 发送；
// 积压队列空则空转一次放行 slot；被拒（ErrLimited）时消息留在
// 积压队列，等下一轮放行（推送场景不丢）。
func (p *PushPipeline) run() {
	defer p.wg.Done()
	for {
		select {
		case <-p.stop:
			return
		default:
		}
		if err := p.bucket.Take(); err != nil {
			continue // 被拒或取消：本轮不放行，消息留在队列
		}
		select {
		case <-p.stop:
			return
		case msg := <-p.queue:
			select {
			case p.sendJob <- msg:
			case <-p.stop:
				return
			}
		default: // 积压队列空：空转
		}
	}
}

// worker 并发发送：单条耗时 sendDelay，发送完成送入出口。
func (p *PushPipeline) worker() {
	defer p.wg.Done()
	for {
		select {
		case <-p.stop:
			return
		case msg := <-p.sendJob:
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
}

// Submit 入队；队列满时阻塞等待而非拒绝（推送场景消息不能丢）。
func (p *PushPipeline) Submit(msg string) {
	select {
	case p.queue <- msg:
	case <-p.stop: // 管线已关闭，消息不再发送
	}
}

// Sent 返回发送出口；管线 Stop 后关闭。
func (p *PushPipeline) Sent() <-chan string { return p.sent }

// Stop 关闭管线：等调度器与全部 worker 退出后关闭发送出口。
func (p *PushPipeline) Stop() {
	p.once.Do(func() {
		close(p.stop)
		p.wg.Wait()
		close(p.sent)
	})
}
