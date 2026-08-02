---
title: wiki-note 技术笔记
description: 个人技术笔记：缓存、消息队列、Redis、Elasticsearch、Push 与限流
---

个人技术笔记仓库。正文以中文撰写，覆盖缓存策略、消息队列、Redis 实践、Elasticsearch、Feed 流、Push 系统与限流算法等主题。

## 主题地图

```mermaid
flowchart TB
    subgraph foundation["基础层"]
        CACHE[cache-strategies]
        RATE[rate-limit-leaky-bucket]
    end
    subgraph messaging["消息与队列"]
        MSG[messaging 选型总览]
        REDIS[redis 系列]
        RMQ[rabbitmq 系列]
    end
    subgraph search["搜索"]
        ES[elasticsearch 系列]
    end
    subgraph product["业务架构"]
        FEED[feed-stream-push-pull]
        PUSH_G[push-global-timezone-delivery]
        PUSH_S[push 系列]
    end
    subgraph ai["AI 工具"]
        AIT[ai-tools 系列]
    end
    CACHE --> REDIS
    CACHE --> FEED
    CACHE --> RATE
    MSG --> REDIS
    MSG --> RMQ
    MSG --> FEED
    MSG --> PUSH_G
    MSG --> PUSH_S
    RMQ --> FEED
    RMQ --> PUSH_S
    REDIS --> PUSH_G
    REDIS --> PUSH_S
    ES --> FEED
```

## 主题域

| 主题域 | 入口 | 说明 |
| ------ | ---- | ---- |
| 缓存 | [cache-strategies.md](./cache-strategies.md) | Cache-Aside 等读写策略 |
| 限流 | [rate-limit-leaky-bucket.md](./rate-limit-leaky-bucket.md) | 漏桶 / 令牌桶原理与 Go 实现 |
| 消息 / 队列 / 延迟 | [messaging/README.md](./messaging/README.md) | Redis vs RabbitMQ vs Kafka 选型与学习路径 |
| Redis 实践 | [redis/README.md](./redis/README.md) | 分布式锁、ZSET 延迟队列、租约恢复 |
| RabbitMQ | [rabbitmq/README.md](./rabbitmq/README.md) | AMQP 全系列（6 篇） |
| Elasticsearch | [elasticsearch/README.md](./elasticsearch/README.md) | 高亮、多语言搜索 |
| Feed 流 | [feed-stream-push-pull.md](./feed-stream-push-pull.md) | 推拉模式、Fan-out、Inbox/Outbox |
| Push | [push/README.md](./push/README.md) | 入门系列 7 篇 + [全球化概念](./push-global-timezone-delivery.md) |
| AI 工具 | [ai-tools/README.md](./ai-tools/README.md) | Agent 代码理解、IDE 插件选型 |

## 推荐学习路径

| 路径 | 顺序 |
| ---- | ---- |
| **A · 后端通用** | cache-strategies → redis 系列 1→3 → rabbitmq 系列 1→6 |
| **B · Feed / 社交** | feed-stream-push-pull → cache-strategies → rabbitmq/reliable-publishing |
| **D · Push / 触达** | push-global-timezone-delivery → push 系列 0→6 → redis/redis-zset-delay-queue → rabbitmq/reliable-publishing |
| **C · 搜索** | elasticsearch/es-highlight → es-multilingual-search |
