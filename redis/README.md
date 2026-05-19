# Redis 队列与延迟任务系列

> 本目录收录基于 Redis 的延迟队列、租约与卡死恢复实践。每篇 `*.md` 独立对应 Wiki 一篇文档；同步约定见 [CLAUDE.md](../CLAUDE.md)。

## 如何使用本系列

| 你的目标 | 建议 |
| -------- | ---- |
| 按学习顺序阅读 | 下表序号 1 → 2 |
| 改本地 → 同步 | 在仓库根目录：` docs +update --api-version v2 --doc "<wiki-url>" --command overwrite --doc-format markdown --content "@redis/<file>.md"` |

**阅读顺序**：ZSET 延迟队列全貌 → Running 卡死与 Watchdog 恢复。

## 文章目录

| 序号 | 文件 | 主题 | 原文 |
| ---- | ---- | ---- | -------- |
| 1 | [redis-zset-delay-queue.md](./redis-zset-delay-queue.md) | ZSET 延迟队列、租约、Reaper、与 MQ 选型 |  |
| 2 | [redis-zset-running-recovery.md](./redis-zset-running-recovery.md) | processing 租约、Lease、Watchdog 回收 |  |

## 关联主题

- RabbitMQ 延迟拓扑：[rabbitmq/delay-and-priority.md](../rabbitmq/delay-and-priority.md)
- 消息选型总览：[messaging/README.md](../messaging/README.md)
