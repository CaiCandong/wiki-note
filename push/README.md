# Push 系统入门系列

> 本目录收录移动端 Push 从基础概念到规模化治理的实战笔记。全球时区与 Campaign 规则见根目录 [push-global-timezone-delivery.md](../push-global-timezone-delivery.md)。

## 如何使用本系列

| 你的目标 | 建议 |
| -------- | ---- |
| 按学习顺序阅读 | 下表序号 0 → 6 |
| 全球 Push / 时区 | 先读 [push-global-timezone-delivery.md](../push-global-timezone-delivery.md) |

**主线参考**：

- [转转 Push 架构演进](https://developer.aliyun.com/article/1673182) — 从 0 搭建
- [腾讯新闻 Push 重构](https://cloud.tencent.com/developer/article/2598882) — 规模化治理

## 阅读顺序

| 序号 | 文件 | 主题 |
| ---- | ---- | ---- |
| 0 | [push-fundamentals.md](./push-fundamentals.md) | 三方模型、Token、V0 最小发送 |
| 1 | [push-campaign-admin.md](./push-campaign-admin.md) | Campaign 后台、人群包 |
| 2 | [push-active-users-zset.md](./push-active-users-zset.md) | Redis ZSET 活跃集、离线拆包 |
| 3 | [push-mq-fanout-provider.md](./push-mq-fanout-provider.md) | MQ 扇出、Provider、深链 |
| 4 | [push-multi-vendor-token.md](./push-multi-vendor-token.md) | Token 服务、多厂商路由 |
| 5 | [push-ab-monitor.md](./push-ab-monitor.md) | AB、监控、防打扰 |
| 6 | [push-link-governance.md](./push-link-governance.md) | 链路治理、优先级、故障转移 |

## 关联主题

| 主题 | 入口 |
| ---- | ---- |
| 全球 Push 概念 | [push-global-timezone-delivery.md](../push-global-timezone-delivery.md) |
| Redis ZSET | [redis/redis-zset-delay-queue.md](../redis/redis-zset-delay-queue.md) |
| MQ 可靠投递 | [rabbitmq/reliable-publishing.md](../rabbitmq/reliable-publishing.md) |
| 选型与学习路径 | [messaging/README.md](../messaging/README.md) 路径 D |
