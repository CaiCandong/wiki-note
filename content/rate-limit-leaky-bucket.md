---
title: "漏桶算法与令牌桶对比（Go 实现）"
---

# 漏桶算法与令牌桶对比（Go 实现）


限流（Rate Limiting）入门：漏桶（Leaky Bucket）与令牌桶（Token Bucket）的原理、Go 实现与场景选型。

---

## 目录

1. 快速选型
2. 漏桶算法：定义与流程
3. 队列式漏桶（Go）
4. 计算式漏桶（Go）
5. 令牌桶算法（Go）
6. 漏桶 vs 令牌桶：场景选型
7. 生产实践要点
8. 相关笔记

---

## 1. 快速选型

| 场景 | 优先方案 |
| ---- | -------- |
| 下游有硬性速率限制（推送厂商 API、短信、第三方服务） | 漏桶 |
| 保护数据库 / 连接池等扛不住尖峰的组件 | 漏桶 |
| 用户侧 API 限流，请求不能排队 | 令牌桶 |
| 允许短时突发、闲时攒量 | 令牌桶 |
| 网络层 QoS / 流量整形 | 令牌桶 |

---

## 2. 漏桶算法：定义与流程

### 2.1 定义

把请求想象成**往桶里倒水**：无论上游怎么突发地倒，桶底只有一个固定大小的漏口，水以**恒定速率**流出；桶满后溢出的水被丢弃（拒绝请求）。

- **桶容量（capacity）**：允许积压的请求数，相当于队列长度，满了即拒绝新请求
- **流出速率（rate）**：每秒固定放行数量，由漏口大小决定
- **核心特性**：输出绝对平滑，**不允许突发**

### 2.2 流程图

```
    突发流量（多少都行）
         │  │  │
         ▼  ▼  ▼
      ┌───────────┐
      │   漏 桶    │ ← capacity：积压队列
      │           │
      └─────┬─────┘
            │ ← rate：恒定速率流出
            ▼
        下游服务
```

### 2.3 优缺点

| 维度 | 说明 |
| ---- | ---- |
| ✅ 输出平滑 | 下游永远看到恒定速率，不抖动 |
| ✅ 保护下游 | 扛不住尖峰的组件（数据库、第三方 API）最合适 |
| ❌ 不能突发 | 系统有余力也被限死在固定速率，浪费容量 |
| ❌ 积压延迟 | 队列实现下请求排队等待，延迟不可控 |

### 2.4 两种实现

| 实现 | 思路 | 特点 |
| ---- | ---- | ---- |
| 队列式 | 请求进 FIFO 队列，定时器按固定速率出队执行 | 直观、贴合"桶"的比喻，有排队延迟 |
| 计算式 | 无队列，按流逝时间积分模拟水量 | 无排队延迟、并发友好；限流即拒绝，不积压 |

---

## 3. 队列式漏桶（Go）

buffered channel 当桶，后台 goroutine 按固定间隔出桶执行：

```go
// 队列式漏桶：请求（job）进桶排队，恒定速率出桶执行
type LeakyBucket struct {
    queue chan func() // 桶：capacity 即最大积压
    rate  time.Duration // 漏口：每个请求的放行间隔
}

func NewLeakyBucket(capacity int, rate time.Duration) *LeakyBucket {
    lb := &LeakyBucket{queue: make(chan func(), capacity), rate: rate}
    go lb.drain() // 后台匀速漏水
    return lb
}

// drain：定时器驱动，恒定速率取出一个请求执行
func (lb *LeakyBucket) drain() {
    ticker := time.NewTicker(lb.rate)
    defer ticker.Stop()
    for range ticker.C {
        select {
        case job := <-lb.queue:
            job() // 以固定速率执行
        default: // 桶空，空转
        }
    }
}

// Submit：入桶排队；桶满即拒绝（溢出丢弃）
func (lb *LeakyBucket) Submit(job func()) bool {
    select {
    case lb.queue <- job:
        return true
    default:
        return false
    }
}
```

要点：

- 出桶速率 = `rate`，与上游来速无关，输出恒定
- 桶满（channel 满）时 `Submit` 走 `default` 分支直接拒绝，不阻塞调用方
- 缺点：job 排队等待执行，积压越多延迟越大；生产代码还需在关闭时停止 `drain` goroutine

---

## 4. 计算式漏桶（Go）

无队列，基于「流逝的时间在漏水」做积分，O(1) 判断。请求来了直接判定放行或拒绝，不排队，适合高并发纯限流场景：

```go
// 计算式漏桶：时间积分模拟水量，无队列
type CounterLeakyBucket struct {
    mu       sync.Mutex
    rate     float64 // 每秒放行速率
    capacity float64 // 桶容量
    water    float64 // 当前积压水量
    last     time.Time
}

func NewCounterLeakyBucket(rate, capacity float64) *CounterLeakyBucket {
    return &CounterLeakyBucket{rate: rate, capacity: capacity, last: time.Now()}
}

func (b *CounterLeakyBucket) Allow() bool {
    b.mu.Lock()
    defer b.mu.Unlock()

    now := time.Now()
    // 先按流逝时间漏水，释放容量（水最少漏到 0）
    b.water = math.Max(0, b.water-(now.Sub(b.last).Seconds()*b.rate))
    b.last = now

    if b.water < b.capacity { // 桶未满：放行并加水
        b.water++
        return true
    }
    return false // 桶满：拒绝
}
```

要点：

- `rate` 是每秒放行数：流逝 `dt` 秒就漏掉 `dt * rate` 的水
- 无排队 → 请求要么立即放行，要么立即拒绝，延迟恒定
- 单机限流用 `sync.Mutex` 即可；多实例需换 Redis 原子操作（见 §7）

---

## 5. 令牌桶算法（Go）

与漏桶相反：按速率往桶里**放令牌**，请求拿到令牌才放行；桶里攒下的令牌就是**突发额度**（burst）：

```go
// 令牌桶：按 rate 补充令牌，burst 决定最大突发请求数
type TokenBucket struct {
    mu     sync.Mutex
    rate   float64 // 每秒补充令牌数
    burst  float64 // 桶容量 = 最大突发额度
    tokens float64 // 当前令牌数
    last   time.Time
}

func NewTokenBucket(rate, burst float64) *TokenBucket {
    return &TokenBucket{rate: rate, burst: burst, tokens: burst, last: time.Now()}
}

func (b *TokenBucket) Allow() bool {
    b.mu.Lock()
    defer b.mu.Unlock()

    now := time.Now()
    // 补充令牌，最多攒满 burst（闲时攒量）
    b.tokens = math.Min(b.burst, b.tokens+now.Sub(b.last).Seconds()*b.rate)
    b.last = now

    if b.tokens >= 1 { // 有令牌：取走一枚并放行
        b.tokens--
        return true
    }
    return false // 无令牌：拒绝
}
```

要点：

- 空闲期间令牌攒到 `burst` 上限 → 之后可以一口气放行 `burst` 个请求，**允许突发**
- 令牌桶和计算式漏桶代码结构几乎一样，差别只在「漏水」vs「补令牌」、以及放行判断方向

---

## 6. 漏桶 vs 令牌桶：场景选型

选型先回答两个问题：**下游能扛突发吗？请求能接受排队吗？**

| 维度 | 漏桶 | 令牌桶 |
| ---- | ---- | ------ |
| 输出形状 | 恒定速率，绝对平滑 | 允许突发，突发上限 = 令牌存量 |
| 请求延迟 | 可能排队等待 | 拿不到令牌立即拒绝，无排队 |
| 吞吐利用 | 低峰也慢，浪费容量 | 低峰攒令牌，利用率高 |
| 保护对象 | 下游（厂商 API / 数据库） | 系统整体 QPS |
| 典型位置 | 生产者侧、调用厂商前 | 网关、业务入口 |
| 典型产品 | nginx `limit_req` | `golang.org/x/time/rate`、Guava `RateLimiter`、Sentinel |

场景对照：

- 推送厂商 API、短信服务有硬性速率上限 → **漏桶**，超限会被封禁或降权
- 数据库写流量保护、数据迁移限速 → **漏桶**（可接受排队，求平滑）
- 用户侧 API 网关限 QPS → **令牌桶**（实时接口等不起排队）
- 秒杀 / 热点活动、允许短时突发 → **令牌桶**
- 生产系统常见组合：入口令牌桶控 QPS + 厂商侧漏桶控节奏

---

## 7. 生产实践要点

1. **直接用现成库**：Go 生态用 `golang.org/x/time/rate`（令牌桶，支持 `Allow` / `Wait` / `WaitN` 批量放行），不必重复造轮子；本文代码用于理解原理。
2. **单机 vs 分布式**：本文实现是单机限流；多实例需用 Redis 原子操作（`INCR` + 过期时间，或 Lua 脚本）做分布式限流。
3. **拒绝策略**：被限流时返回 HTTP 429 并带 `Retry-After` 头，让客户端退避，而不是静默丢弃。
4. **组合使用**：限流只是保护手段之一，常与降级、重试配合（见 cache-strategies 的雪崩防护）。

---

## 8. 相关笔记

- [cache-strategies.md](./cache-strategies.md) — 雪崩防护、Redis INCR 限流
- [push/push-fundamentals.md](./push/push-fundamentals.md) — 厂商 API 限流重试
- [rabbitmq/consumer-semantics.md](./rabbitmq/consumer-semantics.md) — Prefetch 消费限流
