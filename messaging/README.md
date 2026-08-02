# 消息、队列与异步任务 — 选型总览

> **仓库本地索引**。各专题正文见 `redis/`、`rabbitmq/` 系列及根目录单篇。

---

## 快速选型

| 场景 | 优先方案 | 详见 |
| ---- | -------- | ---- |
| 轻量延迟任务、调度点少、团队已有 Redis | Redis ZSET 延迟队列 | [redis/redis-zset-delay-queue.md](../redis/redis-zset-delay-queue.md) |
| Worker 宕机后任务卡在 running | ZSET + Lease + Watchdog | [redis/redis-zset-running-recovery.md](../redis/redis-zset-running-recovery.md) |
| 微服务解耦、路由复杂、要 Confirm/ACK/DLX | RabbitMQ | [rabbitmq/README.md](../rabbitmq/README.md) |
| 超高吞吐事件流、日志、回放、分区消费 | Kafka（本库暂无专篇，见各文对比表） | redis 延迟队列文末选型 |
| 发帖扇出、Timeline 投递 | MQ 异步扇出 + 推拉混合 | [feed-stream-push-pull.md](../feed-stream-push-pull.md) |
| 全球 Push、按用户本地时刻送达 | 概念：cohort × 时区触发；实现见 push 系列 | [push-global-timezone-delivery.md](../push-global-timezone-delivery.md) |
| Push 平台从 0 到规模化 | push 入门系列 0→6 | [push/README.md](../push/README.md) |
| 业务写 DB 与发 MQ 的一致性 | Outbox 模式 | [rabbitmq/reliable-publishing.md](../rabbitmq/reliable-publishing.md) §7.3 |

```mermaid
flowchart TD
    Q[异步/延迟需求] --> Q1{需要复杂路由<br/>与 Broker 语义?}
    Q1 -->|是| RMQ[RabbitMQ 系列]
    Q1 -->|否| Q2{QPS 与延迟规模?}
    Q2 -->|中低、已有 Redis| ZSET[Redis ZSET 队列]
    Q2 -->|极高吞吐事件流| KFK[Kafka + 外置调度]
    Q3{Fan-out 时间线?} --> FEED[Feed 推拉 + MQ]
    Q4{全球 Push 本地时刻?} --> PUSH[Push 时区桶 + MQ]
```

ASCII：`复杂路由 → RabbitMQ`；`轻量延迟 → Redis ZSET`；`事件流 → Kafka`；`Timeline → Feed 笔记 + MQ`

---

## 推荐学习路径

### 路径 A：从缓存到队列（后端通用）

1. [cache-strategies.md](../cache-strategies.md)
2. [redis/redis-zset-delay-queue.md](../redis/redis-zset-delay-queue.md)
3. [rabbitmq/core-concepts.md](../rabbitmq/core-concepts.md) → 系列 2～6

### 路径 B：社交 / Feed 向

1. [feed-stream-push-pull.md](../feed-stream-push-pull.md)
2. [cache-strategies.md](../cache-strategies.md)（Timeline 缓存）
3. [rabbitmq/reliable-publishing.md](../rabbitmq/reliable-publishing.md)（扇出 + Outbox）

### 路径 D：Push / 全球触达

1. [push-global-timezone-delivery.md](../push-global-timezone-delivery.md)（全球概念）
2. [push/push-fundamentals.md](../push/push-fundamentals.md) → [push/README.md](../push/README.md) 系列 1～6
3. [redis/redis-zset-delay-queue.md](../redis/redis-zset-delay-queue.md)（活跃集 / 调度）
4. [rabbitmq/reliable-publishing.md](../rabbitmq/reliable-publishing.md)（Campaign Outbox）

### 路径 C：搜索向

1. [elasticsearch/es-highlight.md](../elasticsearch/es-highlight.md)
2. [elasticsearch/es-multilingual-search.md](../elasticsearch/es-multilingual-search.md)

---

## Outbox 模式（横切要点）

**问题**：业务事务已提交，但 MQ 发送失败（或相反），导致 DB 与消息不一致。

**做法**（与 [rabbitmq/reliable-publishing.md](../rabbitmq/reliable-publishing.md) 一致）：

1. 同一 DB 事务内写入业务表 + **Outbox 表**（待发送消息）。
2. 定时任务或 CDC 扫描 Outbox，调用 Broker 发送。
3. **Publisher Confirm** 成功后删除或标记 Outbox 行；失败重试。
4. 消费侧 **幂等** + 去重，避免至少一次投递产生重复副作用。

---

## 系列入口

| 系列 | 目录 |
| ---- | ---- |
| Push 入门 | [push/README.md](../push/README.md) |
| Redis 队列 | [redis/README.md](../redis/README.md) |
| RabbitMQ | [rabbitmq/README.md](../rabbitmq/README.md) |
| Elasticsearch | [elasticsearch/README.md](../elasticsearch/README.md) |
