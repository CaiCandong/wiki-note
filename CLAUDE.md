# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 仓库性质

这是一个**个人技术笔记仓库**，没有源码、构建、测试。每篇笔记是 `content/` 下的一个 `*.md`，**本地 Markdown 即唯一正本**，纳入 **Git** 版本管理（`origin` 为 GitHub 备份）。站点由 **Quartz 5** 构建，推送 `main` 后 GitHub Actions 自动部署到 **https://caicandong.github.io/wiki-note/**。

## 目录结构

| 路径 | 说明 |
| ---- | ---- |
| `content/` | 全部笔记（根目录单篇 + `redis/`、`rabbitmq/`、`push/`、`elasticsearch/`、`ai-tools/`、`messaging/` 系列） |
| `content/index.md` | 站点首页（主题地图 + 主题域导航） |
| `README.md`（根） | GitHub 仓库落地页（链接指向 `content/`） |
| `quartz.config.yaml` | 站点配置（标题、baseUrl、插件开关） |
| `quartz/`、`components/`、`plugins/`、`layouts/`、`static/` | Quartz 生成器源码（勿手动修改） |

## 新增 / 修改一篇笔记

1. 在 `content/` 下建 `<slug>.md`（小写连字符；系列文放对应系列子目录）
2. **顶部必须带 front matter**：`title` 与正文首个 `# H1` 一致（示例见下）
3. 系列笔记更新系列内 `README.md`（阅读顺序表）；涉及主题地图时同步更新 `content/index.md` 与根 `README.md`
4. 本地预览：`npx quartz build --serve` → http://localhost:8080
5. 提交推送 → Actions 自动构建部署

```markdown
---
title: "漏桶算法与令牌桶对比（Go 实现）"
---

# 漏桶算法与令牌桶对比（Go 实现）
```

## 渲染目标：GitHub + Quartz 站点双端

笔记在 GitHub 与 Quartz 站点双端呈现，书写格式需两端兼容：

- **表格前后保留空行**，保证各渲染器解析稳定
- Mermaid 图块双端原生渲染；复杂时序/流程图保留 Mermaid 的同时可配一句节点顺序说明
- 中英文混排时使用括号注释：`Cache-Aside（旁路缓存）`、`Highlight（高亮）`
- 交叉引用使用**相对路径** Markdown 链接（`[缓存策略](./cache-strategies.md)`），Quartz 自动解析为站点内可点链接

## 书写规范

参考已有文档，新笔记沿用：

- 顶部 `# 标题` + front matter `title` + 一句话副标题
- 章节超过 5 节的笔记，开头放一张「目录」或「如何使用本文档」导航
- 概念/策略小节统一模板：**定义 → 流程图 → 读写要点 → 优缺点 → 典型业务**
- 选型类内容先给「快速选型」表，再展开理论
- 关键技术名词保留英文原文：`TTL`、`Cache-Aside`、`Loader`、`Write-Through`、`ZSET`、`lease` 等不翻译
- 文件名使用小写连字符：`cache-strategies.md`、`redis/redis-zset-delay-queue.md`
- 系列子目录：`rabbitmq/`、`redis/`、`elasticsearch/`；每系列一个 `README.md`（阅读顺序 + 互链）
- 跨系列选型与路径：[`content/messaging/README.md`](./content/messaging/README.md)
- 概念文末尾可加 **相关笔记** 互链

## 编辑注意

- 不要引入其他构建工具、Lint、文档站脚手架（Quartz 已就位，勿叠加）
- 新增 / 删除 / 移动笔记后同步更新根 `README.md`、系列 `README.md`、`content/index.md`
- 提交前检查：`npx quartz build` 无构建错误；`public/`、`node_modules/`、`.quartz/` 不入库（已 gitignore）
