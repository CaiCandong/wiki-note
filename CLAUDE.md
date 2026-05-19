# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 仓库性质

这是一个**个人技术笔记仓库**，没有源码、构建、测试。每个 `*.md` 是一篇独立笔记，**正本在 Wiki**，本地 `*.md` 是文档的同步副本。

本地副本纳入 **Git** 版本管理（`origin` 为 GitHub 备份）；**不要主动** `git commit` / `git push`，除非用户明确要求。

## 与文档的关联关系

所有原文统一存放在 **「我的文档库」**（个人 Wiki 空间）下。新增 / 同步时不要落到其他空间或个人云空间。

每篇本地 `*.md` 顶部 blockquote 必须保留：

```
```

`README.md` 表格的最后一列也维护 URL；**同时更新** [`docs/`](./docs/)（`path`、``、``、`series`、`status`）。提交前建议运行 `./scripts/`。

## 同步：必须用 

**禁止手抄、复制粘贴内容**。所有同步操作通过 ` docs +fetch --api-version v2` / `+update` 完成，避免中英文标点、表格空行、代码块格式漂移。

| 场景               | 命令                                                                                               |
| ---------------- | ------------------------------------------------------------------------------------------------ |
| 把内容拉到本地（重写本地）  | ` docs +fetch --api-version v2 --doc "<wiki-url-or-token>"` 后转写为 Markdown                 |
| 把本地修改推回（覆盖整篇） | ` docs +update --api-version v2 --doc "<wiki-url-or-token>" --command overwrite --doc-format markdown --content "@<file>.md"` |
| 在「我的文档库」新建 Wiki 页 | ` wiki +node-create --space-id my_library --title "标题"`（拿到 `node_token` / `obj_token`） |
| 仅追加段落            | ` docs +update --api-version v2 --doc "<wiki-url-or-token>" --command append --content '<p>...</p>'`   |
| 局部精修（推荐 XML）     | ` docs +update --api-version v2 ... --command str_replace / block_insert_after / block_replace`       |

工作流：

1. **改 → 同步到本地**：`docs +fetch` 取最新 → 按本仓书写规范转成 Markdown → 覆写本地文件
2. **改本地 → 同步到**：本地修订完成 → `docs +update` 推回 → 复核页面渲染
3. **新增一篇笔记**：优先 `wiki +node-create --space-id my_library` 建 Wiki 页 → 落地本地 `<slug>.md`（系列文放在 `rabbitmq/`、`redis/`、`elasticsearch/` 等子目录）→ blockquote 写入 ` 与 ``（即 `obj_token`）→ `docs +update` 推正文 → 更新 README、`docs/`、系列内 `README.md`（如有）

### 冲突与正本（避免 drift）

| 规则 | 说明 |
| ---- | ---- |
| **单篇单源编辑** | 同一篇文档在同一轮改动中，只改或只改本地，不要两侧并行改 |
| **默认正本** | 未特别声明时，以 **** 为准：已改 → `docs +fetch` 覆盖本地；本地已改 → `docs +update` 覆盖 |
| **两侧都已改** | 先人工决定保留哪一侧，再用 fetch 或 overwrite 对齐另一侧；禁止手抄合并正文 |
| **登记册** | `` 的 `` 必须与 blockquote 一致；`` 用于提交前校验 |
| **临时文件** | `*.feishu-body.md` 为 fetch 中间产物，已 `.gitignore`，勿提交 |

推送时在**仓库根目录**执行，`--content` 使用相对路径，例如 `@redis/redis-zset-delay-queue.md`、`@elasticsearch/es-highlight.md`。

### 认证：避免频繁 `auth login`

**默认不需要每次同步都登录。** 用户身份 token 保存在本机 Keychain，access token 过期后由 CLI 用 refresh token 自动续期（` auth status` 可查看 `expiresAt` / `refreshExpiresAt`）。

**Agent 遇到 API 失败时，按顺序处理，不要一上来就要求用户 `auth login`：**

1. **先查状态**：` auth status`（或 `auth list`）。仅当 `tokenStatus` 非 `valid`，或报错 `need_user_authorization`、refresh 失败时，才需要重新授权。
2. **区分权限 vs 登录**：`permission_violations` / 缺 scope → 增量授权 ` auth login --scope "<缺失的 scope>"`（多次 login 的 scope **会累积**，不必清空重登）。
3. **执行环境**：在 Cursor / 沙箱中跑  可能无法读写 macOS Keychain（`keychain ... operation not permitted`），导致 refresh 失败并误报要登录。同步命令应使用**完整本机权限**（非受限沙箱）；OAuth 授权建议在用户**系统终端**完成一次即可。
4. **不要擅自反复 login**：同一任务内若已 `valid`，禁止因单次网络抖动就重复发起 `auth login`。
5. **refresh 过期**：`refreshExpiresAt` 之后（通常约 7 天量级，以 `auth status` 为准）才需用户重新 `auth login`；可指定 `--domain` 或 `--scope` 做最小范围授权。

**用户侧减少登出频率：**

- 日常同步前：` auth status`，有效则直接 `docs +fetch` / `docs +update`。
- 新增能力（如首次用 Wiki / 邮箱）再用 `auth login --scope "..."` 补 scope，不要全量重登。
- 避免频繁 `auth logout`；换机或撤销授权后才需完整登录。

注意：

- v2 API 必须带 `--api-version v2`；不要用旧版本接口
- 推回前，确认目标 doc 的 `` 与 blockquote 标注一致，避免误覆盖到别的文档
- `+update` 是高风险写操作；可先用 `--dry-run` 预览；`overwrite` 会清空正文
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
- 文件名使用小写连字符：`cache-strategies.md`、`redis/redis-zset-delay-queue.md`
- 系列子目录：`rabbitmq/`、`redis/`、`elasticsearch/`；每系列一个 `README.md`（阅读顺序 + 链接）
- 跨系列选型与路径：[`messaging/README.md`](./messaging/README.md)（**仅本地索引**，无页）
- 根 `README.md` 维护主题地图与总表；[`docs/`](./docs/) 维护 machine-readable 元数据

## 编辑注意

- **本地 vs 是双向同步关系**，不是单向导出。任何一侧变更都要走  同步到另一侧，不要让两边 drift。
- 不要引入构建工具、Lint、front matter、文档站脚手架（Hugo / Docusaurus 等）。
- 新增 / 删除 / 移动笔记后同步更新 `README.md`、`docs/`、系列 `README.md`；概念文末尾可加 **相关笔记** 互链。
-  Wiki 目录树可与本地子目录**不同构**（可扁平，本地按主题分目录即可）。
