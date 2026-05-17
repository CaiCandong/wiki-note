# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 仓库性质

这是一个**个人技术笔记仓库**，没有源码、构建、测试。每个 `*.md` 是一篇独立笔记，**正本在 Wiki**，本地 `*.md` 是文档的同步副本。

仓库当前**没有 commit、没有 remote**。不要主动执行 `git commit` / `git push`。

## 与文档的关联关系

所有原文统一存放在 **「我的文档库」**（个人 Wiki 空间）下。新增 / 同步时不要落到其他空间或个人云空间。

每篇本地 `*.md` 顶部 blockquote 必须保留：

```
```

`README.md` 表格的最后一列也维护 URL。新增笔记时同步更新两处。

## 同步：必须用 

**禁止手抄、复制粘贴内容**。所有同步操作通过 ` docs +fetch --api-version v2` / `+update` 完成，避免中英文标点、表格空行、代码块格式漂移。

| 场景               | 命令                                                                                               |
| ---------------- | ------------------------------------------------------------------------------------------------ |
| 把内容拉到本地（重写本地）  | ` docs +fetch --api-version v2 --doc "<wiki-url-or-token>"` 后转写为 Markdown                 |
| 把本地修改推回（覆盖整篇） | ` docs +update --api-version v2 --doc "<wiki-url-or-token>" --command overwrite --content-file <file>` |
| 仅追加段落            | ` docs +update --api-version v2 --doc "<wiki-url-or-token>" --command append --content '<p>...</p>'`   |
| 局部精修（推荐 XML）     | ` docs +update --api-version v2 ... --command str_replace / block_insert_after / block_replace`       |

工作流：

1. **改 → 同步到本地**：`docs +fetch` 取最新 → 按本仓书写规范转成 Markdown → 覆写本地文件
2. **改本地 → 同步到**：本地修订完成 → `docs +update` 推回 → 复核页面渲染
3. **新增一篇笔记**：先 `docs +create --api-version v2` 在建文档拿到 wiki URL → 落地本地 `<slug>.md` → 写入 URL 与  → 更新 README

注意：

- v2 API 必须带 `--api-version v2`；不要用旧版本接口
- 推回前，确认目标 doc 的 `` 与 blockquote 标注一致，避免误覆盖到别的文档
- `+update` 是高风险写操作，遵循  的 `--yes` 确认协议；可先用 `--dry-run` 预览
- 同步前后，侧的画板 token（`<whiteboard token="...">`）会变化，本地 Markdown 用 ```` ```mermaid ```` 代码块承载图源，不保留 token

## 渲染目标：文档

所有笔记的最终呈现形式是粘贴/同步到 Wiki，不是 GitHub。书写格式必须照顾的渲染特性：

- **表格前后必须空行**。在表格紧贴标题或段落时会解析失败。编辑既有表格时保持上下空行。
- Mermaid 图块在中**不一定能渲染**。复杂时序/流程图保留 Mermaid 的同时，可在前/后配 ASCII 阶梯图或一句节点顺序说明。
- 中英文混排时使用括号注释：`Cache-Aside（旁路缓存）`、`Highlight（高亮）`。
- URL 在内会被自动转链；本地 Markdown 写裸 URL 即可（`
## 书写规范

参考已有文档，新笔记沿用：

- 顶部 `# 标题` + blockquote 副标题，副标题包含 **原文 URL** 和 ****
- 章节超过 5 节的笔记，开头放一张「目录」或「如何使用本文档」导航
- 概念/策略小节统一模板：**定义 → 流程图 → 读写要点 → 优缺点 → 典型业务**
- 选型类内容先给「快速选型」表，再展开理论
- 关键技术名词保留英文原文：`TTL`、`Cache-Aside`、`Loader`、`Write-Through`、`ZSET`、`lease` 等不翻译
- 文件名使用小写连字符：`cache-strategies.md`、`redis-zset-delay-queue.md`

## 编辑注意

- **本地 vs 是双向同步关系**，不是单向导出。任何一侧变更都要走  同步到另一侧，不要让两边 drift。
- 不要引入构建工具、Lint、front matter、文档站脚手架（Hugo / Docusaurus 等）。
- 新增 / 删除笔记后同步更新 `README.md` 表格。
