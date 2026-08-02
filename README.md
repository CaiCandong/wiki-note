# wiki-note

个人技术笔记仓库。正文以中文撰写，笔记唯一正本为 [content/](./content/) 下的 `*.md`，纳入 Git 版本管理；由 **Quartz** 构建为 GitHub Pages 静态站：**https://caicandong.github.io/wiki-note/**（含全文搜索、关系图谱、反向链接）。书写规范见 [CLAUDE.md](./CLAUDE.md)。

---

## 如何使用

| 角色 | 建议 |
| ---- | ---- |
| **读者（网页）** | 访问 https://caicandong.github.io/wiki-note/，支持全文搜索、Mermaid 渲染、反向链接与图谱 |
| **读者（源码）** | 从下方系列入口进入，GitHub 原生渲染 Markdown 与 Mermaid |
| **作者** | 编辑 [content/](./content/) 下 `.md`（需带 `title` front matter）→ 提交推送 → Actions 自动构建部署 |

本地预览与构建：

```bash
npx quartz build --serve   # 本地预览 http://localhost:8080
npx quartz build           # 构建产物到 public/（已 gitignore）
```

---

## 内容导航

| 系列 | 入口 |
| ---- | ---- |
| 主题地图 · 学习路径 · 根目录单篇 | [content/index.md](./content/index.md)（站点首页） |
| 消息 / 队列 / 延迟选型 | [content/messaging/README.md](./content/messaging/README.md) |
| Push | [content/push/README.md](./content/push/README.md) |
| Redis 实践 | [content/redis/README.md](./content/redis/README.md) |
| RabbitMQ | [content/rabbitmq/README.md](./content/rabbitmq/README.md) |
| Elasticsearch | [content/elasticsearch/README.md](./content/elasticsearch/README.md) |
| AI 工具 | [content/ai-tools/README.md](./content/ai-tools/README.md) |
| 示例代码（可运行 Go 实现） | [code/leaky-bucket/](./code/leaky-bucket/) — 漏桶限流全实现 + 测试 |

---

## 维护

- **导航分工**：主题地图与学习路径归 [content/index.md](./content/index.md)，选型归 [content/messaging/README.md](./content/messaging/README.md)，本文档只做落地页与系列入口
- **示例代码**：`code/leaky-bucket/` 为笔记配套的可运行 Go 实现（独立 go module），改动需 `go test ./...` 通过，不入 Quartz 构建
- **新增 / 删除 / 移动笔记**：同步更新对应系列 `README.md` 与 [content/index.md](./content/index.md)；书写规范见 [CLAUDE.md](./CLAUDE.md)
- **站点部署**：推送 `main` 后 GitHub Actions 自动构建发布到 https://caicandong.github.io/wiki-note/（首次需在仓库 Settings → Pages → Source 选择 **GitHub Actions**）
