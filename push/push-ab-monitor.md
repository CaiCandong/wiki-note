# AB 实验、监控与防打扰

> **版本** 2026-05 · **定位**：Push 入门系列 · 实现篇 D（阶段 5）

> 小流量试投与择优全量、发送/到达/点击监控维度、暂停预览与 DND；智能投放时机与多通道补偿。前置：[push-multi-vendor-token.md](./push-multi-vendor-token.md)。

---

## 如何使用本文档

| 你的目标 | 建议阅读 |
| -------- | -------- |
| AB 文案择优 | 一 |
| 监控看板字段 | 二 |
| 产品化能力 | 三 |
| 与全球 Push 衔接 | 四 |

---

## 一、AB 实验

**问题**：21:00 推送点击率高——是文案好还是流量高峰？

**流程**（转转）：

1. 准备多版文案（A/B/C）
2. **小流量**（如 1%～5%）各发一版
3. 统计窗口内**点击率**（或到达后点击/到达）
4. **择优**全量其余用户

**依赖**：

- 每条 Push 带 `campaign_id`, `variant_id`, `message_id`
- 客户端点击上报 `message_id`
- 发送/到达/点击三条链路可对齐

---

## 二、监控维度

| 维度 | 示例 |
| ---- | ---- |
| 产品线 | App A / App B |
| 客户端 | iOS / Android |
| 通道 | apns / huawei / xiaomi / fcm |
| 指标 | 发送量、到达量、点击量、CTR |
| 对比 | 模板 A vs B、日环比 |

### 2.1 两类耗时（腾讯新闻）

| 指标 | 定义 |
| ---- | ---- |
| **内部链路** | 审核通过 → 调厂商 API |
| **全链路** | 审核通过 → 用户收到（含厂商） |

热点优化目标：内部 P90 降、全链路对标竞品（领先 1～4 分钟量级）。

### 2.2 Grafana 字段建议

- `push_sent_total{campaign, platform, vendor}`
- `push_delivered_total`（依赖厂商回执/客户端上报）
- `push_clicked_total`
- `push_provider_latency_seconds`（histogram）

---

## 三、产品化能力清单

| 能力 | 说明 |
| ---- | ---- |
| 暂停 / 预览 | 见 [push-active-users-zset.md](./push-active-users-zset.md) |
| 防打扰 DND | 用户本地时段不发营销 Push（[push-global-timezone-delivery.md](../push-global-timezone-delivery.md) §9） |
| 频控 | Redis 滑动窗口：用户维度 N 条/天 |
| 打散 | 同一 Campaign 随机 ±N 分钟，避免齐射 |
| 优先级 | hot / normal / low（见 [push-link-governance.md](./push-link-governance.md)） |
| 补偿触达 | Push 不到 → 短信/微信（多通道，可选） |
| 厂商高优通道 | 热点任务提高 `apns-priority` / FCM `high` |

---

## 四、智能投放时机

手淘实践：按用户**最佳触达时刻**投放 vs 固定时刻 → 打开率约 +10%～20%。

与概念篇 **智能送达**、**按用户本地固定时刻** 同族；需用户活跃画像与调度 Tick 配合。

---

## 五、case 排查（简化后）

重构目标：「用户为何没收到」只查 **3 个模块**日志（触发 / 调度 / Provider），而非 20+ 微服务 join。

检查顺序：

1. Campaign 状态、频控、DND 是否拦截
2. Token 是否 active、vendor 是否匹配
3. Provider 返回码与厂商侧配额

---

## 六、相关笔记

- [push-link-governance.md](./push-link-governance.md) — 优先级、故障转移、测试
- [push-global-timezone-delivery.md](../push-global-timezone-delivery.md) — 全球调度
- [feed-stream-push-pull.md](../feed-stream-push-pull.md) — Fan-out 对比

---

*— 文档结束 —*
