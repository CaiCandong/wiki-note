# RabbitMQ 技术笔记系列

> 本目录收录 RabbitMQ 相关文章。

## 如何使用本系列

| 你的目标 | 建议 |
| -------- | ---- |
| 按学习顺序阅读 | 下表序号 1 → 6 |

**阅读顺序**：核心概念 → Exchange 路由 → 可靠投递 → 消费语义 → 延迟/优先级 → 集群 HA。

**选型与路径**：与 Redis 延迟队列、Feed 扇出的对比见 [messaging/README.md](../messaging/README.md)。

## 文章目录

| 序号 | 文件 | 主题 |
| ---- | ---- | ---- |
| 1 | [core-concepts.md](./core-concepts.md) | Broker、Connection、Channel、vhost、Exchange、Queue、Binding |
| 2 | [exchanges-routing.md](./exchanges-routing.md) | direct / fanout / topic / headers 与 Routing Key 设计 |
| 3 | [reliable-publishing.md](./reliable-publishing.md) | Publisher Confirm、事务、Mandatory、持久化 |
| 4 | [consumer-semantics.md](./consumer-semantics.md) | ACK、Prefetch、NACK、DLX、幂等 |
| 5 | [delay-and-priority.md](./delay-and-priority.md) | TTL、DLX 延迟拓扑、优先级、插件选型 |
| 6 | [cluster-ha.md](./cluster-ha.md) | 集群、Quorum Queue、分区、迁移 |

## 书写约定（系列内）

- 概念小节：**定义 → 流程图（Mermaid 或 ASCII）→ 要点 → 优缺点 → 典型业务**
- 保留英文术语：`Exchange`、`Routing Key`、`DLX`、`Prefetch`、`ACK`、`NACK` 等
- 表格前后保留空行
