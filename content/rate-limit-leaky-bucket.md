---
title: "漏桶算法全面指南（Go 实现）"
---

# 漏桶算法全面指南（Go 实现）

> **版本** 2026-08 · **定位**：后端通用 · 限流算法入门与 Go 实现

漏桶（Leaky Bucket）限流算法：原理、Go 实现、分布式扩展与生产选型。

---

## 目录

1. 快速选型
2. 漏桶算法：定义与流程
3. 队列式漏桶（Go）
4. 计算式漏桶（Go）
5. 分布式漏桶
6. 生产实践要点
7. 相关笔记

---

## 1. 快速选型

| 场景 | 结论 |
| ---- | ---- |
| 下游有硬性速率限制（推送厂商 API、短信、第三方服务） | ✅ 漏桶：恒定速率输出，超限会被封禁或降权 |
| 保护数据库 / 连接池等扛不住尖峰的组件 | ✅ 漏桶：接受排队，求平滑 |
| 网络层 QoS / 流量整形（Traffic Shaping） | ✅ 漏桶：恒定比特率输出的经典工具 |
| 用户侧 API 限流，请求不能排队 | ❌ 不合适：队列式积压延迟不可控，计算式又做不到平滑 |
| 允许短时突发、闲时攒量 | ❌ 不合适：漏桶不允许突发 |

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

**典型业务**：下游有硬性速率上限的场景（推送厂商 API、短信、第三方服务）；保护扛不住尖峰的组件（数据库、连接池）。

### 2.4 两种实现

队列式（请求进 FIFO 排队、定时器匀速出队）与计算式（时间积分模拟水量、无队列）两类实现，Go 代码见 §3、§4。总览：

| 实现 | 队列 | 调用方 | 输出形状 | 代表 |
| ---- | ---- | ------ | -------- | ---- |
| 队列式：内存队列 + ticker（§3） | 有界内存队列 | 非阻塞，满即拒 | 恒定速率 | 本文代码 |
| 计算式：非阻塞判定（§4.1） | 无 | 非阻塞，立即判定 | 按需放行 | 本文代码 |
| 计算式：阻塞虚拟队列（§4.2） | 时间轴排队 | 阻塞等待 | 恒定速率（可配轻微突发） | `uber-go/ratelimit` |

### 2.5 常见误区

- **漏桶限的是放行速率，不是并发度**：rate = 100/s 表示每秒放行 100 个请求，不代表同时只有 100 个在执行——执行耗时由 worker 决定（见 §3.2 的 job 要点）
- **漏桶 ≠ QPS 上限控制**：QPS 控制回答"每秒能进多少"，漏桶回答"每秒流出多少且平滑"——保护对象不同
- **漏桶不是防爆武器**：它平滑的是速率；瞬时超量在队列式下积压、计算式下直接拒绝，峰值会被丢弃，不会凭空消失
- **起源**：漏桶最早是 ATM 网络的流量整形（Traffic Shaping）工具，为恒定比特率（CBR）业务保证平滑输出；应用层限流是它的经典迁移场景

---

## 3. 队列式漏桶（Go）

buffered channel 当桶，后台 goroutine 按固定间隔出桶执行。请求不阻塞调用方：放得进桶就排队，桶满立即拒绝。

### 3.1 核心实现

完整实现（含生命周期管理，动机见 §3.3）：

```go
// 队列式漏桶：请求（job）进桶排队，恒定速率出桶执行
type LeakyBucket struct {
    queue chan func()    // 桶：capacity 即最大积压
    rate  time.Duration  // 漏口：每个请求的放行间隔
    stop  chan struct{}  // 关闭信号（§3.3）
    once  sync.Once      // Stop 只生效一次，重复调用不 panic
    wg    sync.WaitGroup // 等待 drain 退出，优雅关闭用
}

func NewLeakyBucket(capacity int, rate time.Duration) *LeakyBucket {
    lb := &LeakyBucket{queue: make(chan func(), capacity), rate: rate, stop: make(chan struct{})}
    lb.wg.Add(1)
    go func() {
        defer lb.wg.Done()
        lb.drain()
    }()
    return lb
}

// drain：定时器驱动，恒定速率取出一个请求执行；收到关闭信号退出
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

// Submit：入桶排队；桶满即拒绝（溢出丢弃）
func (lb *LeakyBucket) Submit(job func()) bool {
    select {
    case lb.queue <- job:
        return true
    default:
        return false
    }
}

// Stop：关闭出桶，等 drain 退出后返回；队列里剩余的 job 不再执行
func (lb *LeakyBucket) Stop() {
    lb.once.Do(func() { close(lb.stop) })
    lb.wg.Wait()
}
```

### 3.2 行为要点

- 出桶速率 = `rate`，与上游来速无关，输出恒定
- 桶满（channel 满）时 `Submit` 走 `default` 分支直接拒绝，不阻塞调用方
- 容量即最大积压，按下游容忍的积压时长估算：`capacity ≈ 积压时长 ÷ rate`（如容忍积压 1s、rate 100ms → 容量 10）
- 缺点：job 排队等待执行，积压越多延迟越大（延迟不可控，见 §2.3）
- **job 须远快于 `rate`**：`job()` 同步执行会阻塞 `drain` goroutine，期间 Ticker 丢 tick（缓冲为 1，不堆积、不补发），实际出桶速率降为 1/max(rate, job 时长)，队列积压、`Submit` 拒绝率上升——漏桶限的是**放行速率而非完成速率**，job 耗时不可控时应把执行异步化到 worker pool

### 3.3 生命周期：为什么需要 stop / once / wg

- **stop**：`drain` 是常驻 goroutine，不提供停止手段时服务 reload / 热更就会泄漏（内存缓慢上涨，往往运行数天后才暴露），`Stop()` 必须被调用方显式触发
- **once**：重复 `close(lb.stop)` 会 panic，`sync.Once` 保证 `Stop` 幂等（多方持有引用时谁先调谁生效）
- **wg**：`Stop` 里 `wg.Wait()` 等 `drain` 真正退出再返回，配合优雅关闭流程；队列里剩余的 job 是否排空、返回前还是返回后处理，交给调用方决策
- 漏桶最难的不是"怎么填水"，而是**所有请求看到同一个桶、同一套时间、同一个退出信号**：桶实例必须全局共享（或按 key 复用的实例池），不能写成中间件闭包里每次新建的局部变量——每实例一套 channel 会限流失效

---

## 4. 计算式漏桶（Go）

无内存队列，基于「流逝的时间在漏水」做时间积分，O(1) 判断。这类实现有两种形态：**非阻塞判定**（请求来了立即放行或拒绝）与**阻塞等待**（虚拟队列，请求睡到放行时刻再执行，`uber-go/ratelimit` 思路）。

### 4.1 非阻塞：放行判定

请求来了直接判定放行或拒绝，不排队，延迟恒定，适合高并发纯限流场景：

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

    if b.water+1 <= b.capacity { // 加水后不超容量：放行并加水（避免浮点边界下水位突破容量）
        b.water++
        return true
    }
    return false // 桶满：拒绝
}
```

要点：

- `rate` 是每秒放行数：流逝 `dt` 秒就漏掉 `dt * rate` 的水
- 无排队 → 请求要么立即放行，要么立即拒绝，延迟恒定
- 单机限流用 `sync.Mutex` 即可；多实例共享状态用 §5 的分布式实现

### 4.2 阻塞变体：虚拟队列（uber-go/ratelimit 思路）

不建内存队列、把请求在时间轴上排队：请求来了算出"下一次放行时刻"，没到就 `Sleep` 等待，同样保证匀速输出——`uber-go/ratelimit` 即此实现：

```go
// 虚拟队列：不建内存队列，按 perRequest 间隔排定放行时刻，请求阻塞等待
type BlockingBucket struct {
    mu         sync.Mutex
    perRequest time.Duration // 放行间隔 = time.Second / rate
    last       time.Time     // 上次放行时刻
}

func (b *BlockingBucket) Take() {
    b.mu.Lock()
    defer b.mu.Unlock()
    next := b.last.Add(b.perRequest)
    if now := time.Now(); now.Before(next) {
        time.Sleep(next.Sub(now)) // 未到放行时刻：持锁阻塞等待（反模式，见要点）
        b.last = next
    } else {
        b.last = now
    }
}
```

要点：

- 无内存队列、无后台 goroutine，输出同样平滑
- `uber-go/ratelimit` 引入 `maxSlack`（默认 10 个间隔）：空闲期攒下的时间信用允许之后**临时轻微突发**，等于给漏桶开了一个小口（可用 `WithoutSlack` 禁用）
- **持锁 sleep 是反模式**：`Take()` 在 `mu` 临界区内 `time.Sleep`，放行严格串行——高并发下所有请求的 goroutine 卡在 `b.mu.Lock()` 上排队，每个请求占一个 goroutine，数量随并发只增不减（goroutine 爆炸），内存与调度压力持续上涨。正确做法是把 sleep 挪出临界区：uber 的 atomic 版即 CAS 更新 `state` 后在锁外 `Sleep`
- 局限：等待无上限，没有"最大排队数"概念，超限不拒绝而是继续排队；生产需自行加信号量计数（`sem := make(chan struct{}, maxQueue)` 非阻塞入队，满即拒）——分布式版（§5.1）用 `capacity` 参数天然限住排队数
- 与 §4.1 区别：§4.1 立即判定、不排队（放行即放行，输出不平滑）；§4.2 阻塞等待、严格匀速（漏桶语义最纯正的形态）

## 5. 分布式漏桶

单机实现（§3、§4）把队列、`last` 都存在进程内存里；分布式要求多实例共享同一份限流状态：状态放 Redis、判定用 Lua 原子化。分布式漏桶有三种形态：虚拟队列（§5.1）、Redis List 队列（§5.2）、消息队列（§5.3）。

### 5.1 虚拟队列式：Redis Lua

单机版的 `last` 换成一个 Redis key、多实例共享，用 Lua 脚本保证「读 last → 计算 → 写 last」原子性；时间戳取 Redis `TIME` 命令的**服务器时间**，避免各实例时钟漂移让空闲判定失真：

```lua
-- 分布式虚拟队列漏桶：所有实例共享同一个 key
-- KEYS[1] = 限流 key（值为上次放行时刻 last，毫秒）
-- ARGV[1] = perRequest（放行间隔毫秒，= 1000 / rate）
-- ARGV[2] = capacity（允许排队的最大请求数）
-- 返回 {1, waitMs} 放行（waitMs 为需等待时长）；{0, waitMs} 拒绝（waitMs 可作为 Retry-After）
local t = redis.call('TIME')
local now = tonumber(t[1]) * 1000 + math.floor(tonumber(t[2]) / 1000) -- Redis 服务器时间（毫秒）
local last = tonumber(redis.call('GET', KEYS[1]) or '0')
local perRequest = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])

if last == 0 or now >= last + perRequest then -- 首次，或空闲超一个间隔：立即放行并校准 last（未做 maxSlack，不积累空闲信用）
    redis.call('SET', KEYS[1], now)
    return {1, 0}
end

local wait = last + perRequest - now -- 轮到我要等多久
if wait > capacity * perRequest then -- 前面已排满 capacity 人：拒绝
    return {0, wait}
end

redis.call('SET', KEYS[1], last + perRequest) -- 时间轴上占一个位置
return {1, wait}
```

Go 侧调用（`github.com/redis/go-redis/v9`）：

```go
// Take：原子入队判定；放行但需等待时阻塞到放行时刻，拒绝返回 ErrLimited
func (b *DistributedBucket) Take(ctx context.Context) error {
    res, err := b.rdb.Eval(ctx, leakyBucketLua, []string{b.key},
        b.perRequest.Milliseconds(), b.capacity).Result()
    if err != nil {
        return err
    }
    vals := res.([]interface{})
    if vals[0].(int64) != 1 {
        return ErrLimited // HTTP 层返回 429，Retry-After 取 vals[1] 的毫秒数
    }
    if wait := time.Duration(vals[1].(int64)) * time.Millisecond; wait > 0 {
        select {
        case <-ctx.Done():
            return ctx.Err() // 调用方取消，不再等
        case <-time.After(wait):
        }
    }
    return nil
}
```

要点：

- **与单机版同一套语义**：`last` 递进一个 `perRequest` 就是"时间轴上占位"，`capacity × perRequest` 即最大积压时长——与 §3.2 的容量估算公式互相印证
- **天然限排队数**：`capacity` 参数就是虚拟队列的排队上限，超限立即拒绝——正好补上 §4.2 单机版"等待无上限"的短板
- **拒绝语义**：拒绝时返回等待时长，HTTP 层回 `Retry-After` 让客户端退避（呼应 §6 第 3 条）

### 5.2 队列式：Redis List 桶 + 匀速出队

桶 = Redis List（`LPUSH` 入队、`RPOP` 出队、`LLEN` 查容量）。匀速出队不能靠每个实例各跑一个 ticker——多个出队者会竞争加速。两种做法：

- **单执行者**：只在一个实例上跑出队 goroutine，其余实例只入队；简单但有单点，需配合主备切换
- **惰性补账**（推荐）：不跑后台出队者，出队由请求触发——Lua 按流逝时间算出"本应出掉的数量"一次性清出，所有实例共享 `last` 出队时刻：

```lua
-- 惰性补账：入队前先把「按匀速该出的队」出掉
-- KEYS[1] = 桶 list，KEYS[2] = last 出队时刻（毫秒）
-- ARGV[1] = perRequest（毫秒），ARGV[2] = capacity，ARGV[3] = job id
local t = redis.call('TIME')
local now = tonumber(t[1]) * 1000 + math.floor(tonumber(t[2]) / 1000)
local last = tonumber(redis.call('GET', KEYS[2]) or '0')
local perRequest = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])

if last > 0 then
    local drained = math.floor((now - last) / perRequest) -- 流逝时间应出掉的数量
    for i = 1, drained do
        redis.call('RPOP', KEYS[1])
    end
    redis.call('SET', KEYS[2], last + drained * perRequest)
else
    redis.call('SET', KEYS[2], now)
end

if redis.call('LLEN', KEYS[1]) >= capacity then
    return 0 -- 桶满：拒绝
end
redis.call('LPUSH', KEYS[1], ARGV[3])
return 1
```

要点：惰性补账只保证**桶里积压的队形符合匀速**，出掉的 job 仍需有人执行——所以"队列式 + 多实例"的工程形态通常直接上消息队列（§5.3），Redis List 版更适合有独立调度器的场景（调度器定时 `RPOP` + 派发）。

### 5.3 消息队列：天然分布式漏桶

队列即分布式 FIFO 桶，消费者的消费速率即出桶速率，生产者完全不用感知限流：

| 漏桶概念 | 消息队列对应 |
| -------- | ------------ |
| 桶容量（最大积压） | 队列长度上限（RabbitMQ `x-max-length`） |
| 溢出丢弃 | `x-overflow: reject-publish`（或 `drop-head`） |
| 恒定速率出桶 | 消费者按固定间隔消费（prefetch=1 + 处理耗时 ≈ 放行间隔） |

- 请求不能静默丢弃的场景（短信、推送）：配合死信队列（DLX）收容溢出消息
- 消费节奏控制见 [rabbitmq/consumer-semantics.md](./rabbitmq/consumer-semantics.md) 的 Prefetch 语义

### 5.4 场景组合：推送限速（消息不能丢，漏桶 + 积压队列）

推送（短信、APNs、厂商推送）场景有个矛盾：厂商 API 有硬性速率限制（超限封禁），漏桶正好匹配；但推送消息是业务消息**不能丢**，而漏桶的"溢出丢弃"（§3 桶满即拒、§5.3 `reject-publish`）会丢消息。解法是把漏桶从**入口拒绝**改成**出口调度**：

```
生产者 ──▶ [积压队列（持久化）] ──▶ 漏桶（匀速放行） ──▶ 厂商 API
   （只管入队，不丢）       （吸收峰值）       （控制发送节奏）
```

#### 5.4.1 基础语义：不丢消息

- **漏桶位置在出队侧**：消息先全量进积压队列（RabbitMQ / Redis List），漏桶只决定"什么时候从队列取一条发送"，不再拒绝任何消息——峰值被队列吸收，发送速率恒定
- **先 Take 后拉消息**：消费端先拿到漏桶放行时刻，再从队列取消息发送——等待期间消息留在持久化队列里，进程崩溃不丢；反过来"先拉消息再等放行"会让消息在等待期间脱离持久化保护，进程崩了即丢
- **确认语义**：发送成功才 ack（MQ）或删除（Redis List）；失败走重试 / 死信队列（见 [push/push-fundamentals.md](./push/push-fundamentals.md) 的厂商限流重试）
- **消费节奏**：prefetch=1 时消费者逐条处理，天然配合漏桶出队节奏（见 [rabbitmq/consumer-semantics.md](./rabbitmq/consumer-semantics.md)）

#### 5.4.2 两种组织形态

| 形态 | 结构 | 适用 | 代码 |
| ---- | ---- | ---- | ---- |
| 调度器 + 积压队列 | 消息全量入队，dispatcher 匀速出队 | 生产端主动投递的批量场景 | `code/leaky-bucket/pipeline.go` |
| 消费者直连虚拟队列 | worker 直接 `Take` 时间轴排队，无内存队列 | 匹配 MQ 消费模型，**生产推荐** | `code/leaky-bucket/virtual_queue.go` |

#### 5.4.3 生产约束：长耗时厂商 × 多 pod

- **单条耗时长的厂商（如 50ms/条、QPS 上限 500）**：放行间隔 = 1000/500 = 2ms，但单条发送 50ms——放行路径**绝不能同步发送**，否则放行被 50ms 阻塞，实际吞吐塌到 20/s（单协程 1/50ms）。解法：漏桶放行后交给 **worker pool 异步发送**，并发在途数 = rate × 单条耗时（Little's Law：500/s × 0.05s = 25），workers 配 25 起步、留余量（重试会占在途）；worker 不足时吞吐 = workers ÷ 单条耗时 < rate，积压只增不减，漏桶速率形同虚设
- **生产多 pod 水平扩容（虚拟队列版，不建内存队列）**：单 pod 的 worker 上限有限，QPS 高 / 单条耗时长时需要多 pod 扩容，但**漏桶状态必须跨 pod 共享**——每个 pod 各跑一个单机漏桶时，总放行速率 = N × rate，实测 3 pod 各限 300/s 时 600 条 0.72s 发完（900/s），直接超限。正确做法（§4.2 虚拟队列思路）：每个 pod 的 worker 直接在**时间轴上排队**——调共享 Redis 漏桶（§5.1 Lua）的 `Take` 等放行时刻，放行后从 MQ 拉一条发送（先 Take 后拉取，等待期间消息留在持久化队列，进程崩溃不丢）；Lua 原子保证全局总放行速率 = rate，各 pod 无需任何内存队列。实测 3 pod × 5 worker 共享 300/s 时 600 条 1.90s 发完，且**节奏平稳**：放行间隔中位数 3.0ms（≈perRequest 3.3ms）、99% 间隔落在 [perRequest/2, 2×perRequest]（无突发堆叠、无大空窗）、100ms 窗口放行 22~34 条（目标 30）
- 可运行演示与测试（`code/leaky-bucket/`）：`pipeline.go`（积压队列形态：500 条瞬间灌入、rate=500、sendDelay=50ms、workers=25 全部送达约 1s 清空，workers=5 时吞吐塌到 100/s）；`virtual_queue.go`（虚拟队列形态：worker 时间轴排队，多 pod 共享 Redis 漏桶 3 × 5 worker 600 条 1.90s 送达，各 pod 单机漏桶 0.72s 超限）

### 5.5 分布式注意事项

- **时钟一致性**：时间戳统一取 Redis `TIME`（服务器时间），客户端各实例 `time.Now()` 漂移会让"空闲判定"失真（有的实例以为空闲直接放行）
- **key 清理**：限流 key 不设 TTL 会永久占内存，设大 TTL（如 `perRequest × capacity × 2`）并每次写入刷新
- **可用性**：分布式限流依赖 Redis，Redis 故障时限流失效——接入层要定好降级策略（fail-open 放行 / fail-closed 拒绝）并监控
- **性能**：Lua 单次执行微秒级，但每请求多一次 Redis RTT；请求量极高时用"本地限流 + Redis 兜底"的双层结构

---

## 6. 生产实践要点

1. **直接用现成库**：Go 生态的漏桶库不多——`uber-go/ratelimit` 是阻塞式虚拟队列（§4.2 思路）的成熟实现；分布式场景直接参考 §5 的 Lua 脚本。本文代码用于理解原理。
2. **单机 vs 分布式**：本文 §3、§4 是单机限流；多实例共享限流状态时用 §5 的分布式实现。计数器窗口类（`INCR` + 过期时间）见 [cache-strategies.md](./cache-strategies.md)。
3. **拒绝策略**：被限流时返回 HTTP 429 并带 `Retry-After` 头，让客户端退避，而不是静默丢弃（分布式下直接取 Lua 返回的等待时长）。
4. **组合使用**：限流只是保护手段之一，常与降级、重试配合（见 [cache-strategies.md](./cache-strategies.md) 的雪崩防护）。

---

## 7. 相关笔记

- [cache-strategies.md](./cache-strategies.md) — 雪崩防护、Redis INCR 限流
- [push/push-fundamentals.md](./push/push-fundamentals.md) — 厂商 API 限流重试
- [rabbitmq/consumer-semantics.md](./rabbitmq/consumer-semantics.md) — Prefetch 消费限流
