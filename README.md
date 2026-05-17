# wiki-note

个人技术笔记仓库。所有文档以中文撰写，最终用于粘贴到 Wiki 渲染。书写规范、与的同步约定见 [CLAUDE.md](./CLAUDE.md)。

## 文档目录


| 文件                                                             | 内容                                                                          | 原文                                                                |
| -------------------------------------------------------------- | --------------------------------------------------------------------------- | ------------------------------------------------------------------- |
| [cache-strategies.md](./cache-strategies.md)                   | 缓存策略技术指南：Cache-Aside / Read-Through / Write-Through / Write-Behind / Refresh-Ahead 概念、选型、实现要点与业务落地 |                |
| [es-highlight.md](./es-highlight.md)                           | Elasticsearch 高亮（Highlight）入门指南：原理、三种高亮器、生产实践                              |                |
| [redis-zset-delay-queue.md](./redis-zset-delay-queue.md)       | 使用 Redis ZSET 实现延迟队列：架构、生产者/消费者、租约 + Reaper、与 MQ/DB 的选型决策                  |                |
| [redis-zset-running-recovery.md](./redis-zset-running-recovery.md) | Redis ZSET 消费队列：如何最简化解决 Running 卡死（ZSET + Lease + Watchdog 模式）              |                |

