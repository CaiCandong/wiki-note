# Elasticsearch 高亮（Highlight）入门指南


在做搜索系统时，我们常需要把命中的关键词在结果里标出来，例如搜索「苹果手机」返回 `<em>苹果手机</em>性能非常强`。这就是 Elasticsearch 的 **Highlight（高亮）**——搜索体验里非常重要的一环：提升可读性、帮助用户快速定位命中内容。

---

## 一、什么是高亮

高亮本质上是：**将搜索命中的关键词，在返回结果中自动标记出来**。


| 原文                       | 搜索词  | 高亮结果                                |
| ------------------------ | ---- | ----------------------------------- |
| Elasticsearch 是一个分布式搜索引擎 | 搜索引擎 | Elasticsearch 是一个分布式`<em>搜索引擎</em>` |


默认使用 `<em></em>` 包裹命中内容，前端可用 CSS 自定义样式（如 `em { color: red; }`）。

---

## 二、最简示例

**索引 mapping：**

```json
PUT article
{
  "mappings": {
    "properties": {
      "title": { "type": "text" },
      "content": { "type": "text" }
    }
  }
}
```

**查询并开启高亮：**

```json
GET article/_search
{
  "query": { "match": { "content": "全文检索" } },
  "highlight": {
    "fields": { "content": {} }
  }
}
```

**返回片段：**

```json
{
  "highlight": {
    "content": [
      "Elasticsearch 是一个强大的<em>全文检索</em>引擎"
    ]
  }
}
```

---

## 三、核心配置


| 参数                       | 作用                              |
| ------------------------ | ------------------------------- |
| `fields`                 | 指定哪些字段需要高亮（如 `title`、`content`） |
| `pre_tags` / `post_tags` | 自定义高亮标签，默认 `<em></em>`          |
| `fragment_size`          | 每个高亮片段的最大长度（如 50）               |
| `number_of_fragments`    | 返回多少个高亮片段（如 3）                  |


**自定义标签示例：**

```json
"highlight": {
  "pre_tags": ["<span style='color:red'>"],
  "post_tags": ["</span>"],
  "fields": { "content": {} }
}
```

---

## 四、底层原理：Offset（字符偏移量）

高亮**不是**简单的字符串替换，而是依赖分词后的 **offset**：


| term | start_offset | end_offset |
| ---- | ------------ | ---------- |
| 苹果   | 0            | 2          |
| 手机   | 2            | 4          |


流程：找到命中的 term → 根据 offset 定位 → 插入 HTML 标签。

---

## 五、三种高亮器


| 类型                   | 配置                  | 特点                                | 适用场景             |
| -------------------- | ------------------- | --------------------------------- | ---------------- |
| **unified**（默认推荐）    | `"type": "unified"` | 官方推荐；支持 phrase/fuzzy/wildcard；精度高 | 大多数业务            |
| **plain**            | `"type": "plain"`   | 实现简单；会 re-analyze 文本，大文本性能差       | 小文本、老版本兼容        |
| **fvh**（Fast Vector） | `"type": "fvh"`     | 大文本性能最强                           | PDF、wiki、长文章、知识库 |


**FVH 前置条件：** 字段需开启 `term_vector: with_positions_offsets`：

```json
"content": {
  "type": "text",
  "term_vector": "with_positions_offsets"
}
```

---

## 六、常见问题

### 1. 中文高亮不完整

搜索「苹果手机」，却只高亮「苹果」——通常是 **索引与查询分词器不一致**（如索引 `ik_max_word`、查询 `ik_smart`）。

### 2. 高亮错位

如出现 `<em>苹果手</em>机`，多为 **offset 计算异常**，常见于自定义 tokenizer、char_filter、emoji、html_strip。

### 3. 未命中词也被高亮

搜索「苹果」却高亮「水果」——常与 **synonym、fuzzy、wildcard** 有关，unified highlighter 会做扩展匹配。

---

## 七、生产环境最佳实践

1. **小文本用 unified**：标题、摘要、评论等，默认即可。
2. **长文本用 FVH**：配合 `term_vector: with_positions_offsets`。
3. **不要高亮超大字段**：10MB HTML 易导致 CPU/内存飙升；建议建 `content_snippet` 专用于高亮。
4. **控制 fragment_size**：推荐 50～300，避免设为 10000。
5. **限制高亮字段**：避免 `"*": {}` 高亮所有字段。
6. **使用 highlight_query**：当主查询是复杂 `bool + filter + should` 时，单独指定高亮逻辑，避免错误高亮：

```json
"highlight_query": {
  "match_phrase": { "content": "苹果手机" }
}
```

---

## 八、生产级完整示例

```json
GET article/_search
{
  "query": { "match": { "content": "Elasticsearch 高亮" } },
  "highlight": {
    "pre_tags": ["<mark>"],
    "post_tags": ["</mark>"],
    "fields": {
      "content": {
        "type": "unified",
        "fragment_size": 100,
        "number_of_fragments": 2
      }
    }
  }
}
```

---

## 九、大厂为何常自研高亮

ES Highlight 存在 CPU 高、长文本慢、中文易碎、synonym/fuzzy 易异常等问题。很多大型系统采用：

- **ES 负责召回**：返回命中的 term、offset、positions
- **应用层自研高亮**：更灵活、更稳定，可实现「完整匹配 > 前缀匹配 > 单词匹配」等复杂规则

---

## 十、总结与推荐方案

高亮本质是 **「基于 Query 的文本片段生成系统」**，需理解：分词（Analyzer）、Offset、Term Vector、Fragment、Passage Scoring。


| 场景   | 推荐方案              |
| ---- | ----------------- |
| 小文本  | unified           |
| 长文本  | fvh + term_vector |
| 超大系统 | 应用层自研高亮           |


---

## 深入学习路径

1. Analyzer
2. Offset
3. unified highlighter
4. term_vector
5. 中文分词
6. fragment 原理
7. highlight_query
8. Lucene 高亮源码

掌握以上主题后，即可系统性地处理搜索引擎中的高亮工程问题。

---

**相关笔记**：[es-multilingual-search.md](./es-multilingual-search.md) — 指定语言下的索引拆分、`dis_max` 分层排序与 `highlight_query` 对齐（短剧标题 + 标签场景）。