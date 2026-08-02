# Elasticsearch 技术笔记系列

> 本目录收录 ES 搜索与高亮相关文章。

## 如何使用本系列

| 你的目标 | 建议 |
| -------- | ---- |
| 按学习顺序阅读 | 下表序号 1 → 2 |

**阅读顺序**：高亮基础 → 多语言搜索（含 `highlight_query` 与索引拆分）。

## 文章目录

| 序号 | 文件 | 主题 |
| ---- | ---- | ---- |
| 1 | [es-highlight.md](./es-highlight.md) | Highlight 原理、三种 highlighter、生产实践 |
| 2 | [es-multilingual-search.md](./es-multilingual-search.md) | 按语言分索引、`dis_max`、高亮对齐 |

## 关联主题

- 缓存层：[cache-strategies.md](../cache-strategies.md)（搜索结果缓存、旁路失效）
- 消息与索引同步：[messaging/README.md](../messaging/README.md)（CDC / 异步入库场景）
