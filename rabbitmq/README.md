# RabbitMQ 技术笔记系列

> 本目录收录 RabbitMQ 相关文章。每篇 `*.md` 独立对应 Wiki 一篇文档；同步方式与仓库根目录一致，见 [CLAUDE.md](../CLAUDE.md)。

## 如何使用本系列

| 你的目标 | 建议 |
| -------- | ---- |
| 按学习顺序阅读 | 下表序号 1 → 6 |
| 改 → 同步本地 | ` docs +fetch --api-version v2 --doc "<wiki-url>"` 后覆写对应 `.md` |
| 改本地 → 同步 | ` docs +update --api-version v2 --doc "<wiki-url>" --command overwrite --doc-format markdown --content "@rabbitmq/<file>.md"` |

**阅读顺序**：核心概念 → Exchange 路由 → 可靠投递 → 消费语义 → 延迟/优先级 → 集群 HA。

## 文章目录

| 序号 | 文件 | 主题 | 原文 |
| ---- | ---- | ---- | -------- |
| 1 | [core-concepts.md](./core-concepts.md) | Broker、Connection、Channel、vhost、Exchange、Queue、Binding |  |
| 2 | [exchanges-routing.md](./exchanges-routing.md) | direct / fanout / topic / headers 与 Routing Key 设计 |  |
| 3 | [reliable-publishing.md](./reliable-publishing.md) | Publisher Confirm、事务、Mandatory、持久化 |  |
| 4 | [consumer-semantics.md](./consumer-semantics.md) | ACK、Prefetch、NACK、DLX、幂等 |  |
| 5 | [delay-and-priority.md](./delay-and-priority.md) | TTL、DLX 延迟拓扑、优先级、插件选型 |  |
| 6 | [cluster-ha.md](./cluster-ha.md) | 集群、Quorum Queue、分区、迁移 |  |

## 书写约定（系列内）

- 概念小节：**定义 → 流程图（Mermaid 或 ASCII）→ 要点 → 优缺点 → 典型业务**
- 保留英文术语：`Exchange`、`Routing Key`、`DLX`、`Prefetch`、`ACK`、`NACK` 等
- 表格前后保留空行（渲染要求）
