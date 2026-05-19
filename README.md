# wiki-note

个人技术笔记仓库。所有文档以中文撰写，最终用于粘贴到 Wiki 渲染。书写规范、与的同步约定见 [CLAUDE.md](./CLAUDE.md)。

## 文档目录

### 根目录

| 文件                                                             | 内容                                                                          | 原文                                                                |
| -------------------------------------------------------------- | --------------------------------------------------------------------------- | ------------------------------------------------------------------- |
| [cache-strategies.md](./cache-strategies.md)                   | 缓存策略技术指南：Cache-Aside / Read-Through / Write-Through / Write-Behind / Refresh-Ahead 概念、选型、实现要点与业务落地 |                |
| [es-highlight.md](./es-highlight.md)                           | Elasticsearch 高亮（Highlight）入门指南：原理、三种高亮器、生产实践                              |                |
| [redis-zset-delay-queue.md](./redis-zset-delay-queue.md)       | 使用 Redis ZSET 实现延迟队列：架构、生产者/消费者、租约 + Reaper、与 MQ/DB 的选型决策                  |                |
| [redis-zset-running-recovery.md](./redis-zset-running-recovery.md) | Redis ZSET 消费队列：如何最简化解决 Running 卡死（ZSET + Lease + Watchdog 模式）              |                |
| [es-multilingual-search.md](./es-multilingual-search.md)         | Elasticsearch 多语言搜索：方案 A 文档模型、按语言分索引、标题分层查询与高亮对齐                    |                |
| [feed-stream-push-pull.md](./feed-stream-push-pull.md)           | Feed 流推拉模式：Fan-out、推/拉/混合、时间分区拉、Inbox/Outbox 与生产实践                         |                |

### [rabbitmq/](./rabbitmq/) — RabbitMQ 系列

系列索引见 [rabbitmq/README.md](./rabbitmq/README.md)。

| 文件 | 内容 | 原文 |
| ---- | ---- | -------- |
| [rabbitmq/core-concepts.md](./rabbitmq/core-concepts.md) | AMQP 模型、Broker、Connection、Channel、Exchange、Queue |  |
| [rabbitmq/exchanges-routing.md](./rabbitmq/exchanges-routing.md) | Exchange 四种类型、Routing Key 规范与排障 |  |
| [rabbitmq/reliable-publishing.md](./rabbitmq/reliable-publishing.md) | Publisher Confirm、事务、Mandatory、持久化与 Outbox |  |
| [rabbitmq/consumer-semantics.md](./rabbitmq/consumer-semantics.md) | 手动 ACK、Prefetch、DLX、幂等与优雅停机 |  |
| [rabbitmq/delay-and-priority.md](./rabbitmq/delay-and-priority.md) | TTL+DLX 延迟、优先级队列、与 Redis 选型 |  |
| [rabbitmq/cluster-ha.md](./rabbitmq/cluster-ha.md) | 集群、Quorum Queue、网络分区、迁移 checklist |  |

