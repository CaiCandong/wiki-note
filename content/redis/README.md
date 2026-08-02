---
title: "Redis 实践系列"
---

# Redis 实践系列

> 本目录收录 Redis 分布式锁、延迟队列与租约恢复等实践。

## 如何使用本系列

| 你的目标 | 建议 |
| -------- | ---- |
| 按学习顺序阅读 | 下表序号 1 → 3 |

**阅读顺序**：分布式锁心智模型与演进 → ZSET 延迟队列全貌 → Running 卡死与 Watchdog 恢复。

## 文章目录

| 序号 | 文件 | 主题 |
| ---- | ---- | ---- |
| 1 | [redis-distributed-lock.md](./redis-distributed-lock.md) | SETNX、单实例锁、Redlock、看门狗 |
| 2 | [redis-zset-delay-queue.md](./redis-zset-delay-queue.md) | ZSET 延迟队列、租约、Reaper、与 MQ 选型 |
| 3 | [redis-zset-running-recovery.md](./redis-zset-running-recovery.md) | processing 租约、Lease、Watchdog 回收 |

## 关联主题

- RabbitMQ 延迟拓扑：[rabbitmq/delay-and-priority.md](../rabbitmq/delay-and-priority.md)
- 消息选型总览：[messaging/README.md](../messaging/README.md)
- 缓存与秒杀：[cache-strategies.md](../cache-strategies.md)
