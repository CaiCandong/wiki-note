# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 仓库性质

这是一个**个人技术笔记仓库**，没有源码、构建、测试。每个 `*.md` 是一篇独立笔记，**本地 Markdown 即唯一正本**，纳入 **Git** 版本管理（`origin` 为 GitHub 备份）。

## 新增一篇笔记

1. 本地创建 `<slug>.md`（小写连字符命名；系列文放在 `rabbitmq/`、`redis/`、`elasticsearch/` 等子目录）
2. 按下方「书写规范」撰写正文
3. 更新根 `README.md` 文档目录表；系列笔记同时更新系列内 `README.md`（如有）
4. 完成，无需其他登记操作

## 渲染目标：GitHub Markdown

所有笔记最终在 GitHub（及本机 Markdown 编辑器）中阅读。书写格式遵循通用 Markdown 规范：

- **表格前后保留空行**，保证各渲染器解析稳定
- Mermaid 图块在 GitHub 原生渲染；复杂时序/流程图同时保留 Mermaid 与一句节点顺序说明（或 ASCII 阶梯图），便于不支持 Mermaid 的场景阅读
- 中英文混排时使用括号注释：`Cache-Aside（旁路缓存）`、`Highlight（高亮）`
- 交叉引用使用标准 Markdown 链接：`[缓存策略](./cache-strategies.md)`

## 书写规范

参考已有文档，新笔记沿用：

- 顶部 `# 标题` + 一句话副标题（概述本篇内容）
- 章节超过 5 节的笔记，开头放一张「目录」或「如何使用本文档」导航
- 概念/策略小节统一模板：**定义 → 流程图 → 读写要点 → 优缺点 → 典型业务**
- 选型类内容先给「快速选型」表，再展开理论
- 关键技术名词保留英文原文：`TTL`、`Cache-Aside`、`Loader`、`Write-Through`、`ZSET`、`lease` 等不翻译
- 文件名使用小写连字符：`cache-strategies.md`、`redis/redis-zset-delay-queue.md`
- 系列子目录：`rabbitmq/`、`redis/`、`elasticsearch/`；每系列一个 `README.md`（阅读顺序 + 互链）
- 跨系列选型与路径：[`messaging/README.md`](./messaging/README.md)（仅本地索引）
- 根 `README.md` 维护主题地图与总表

## 编辑注意

- 不要引入构建工具、Lint、front matter、文档站脚手架（Hugo / Docusaurus 等）。
- 新增 / 删除 / 移动笔记后同步更新 `README.md` 与系列 `README.md`；概念文末尾可加 **相关笔记** 互链。
- 提交前检查：无遗漏更新索引、表格格式正确、交叉链接有效。
