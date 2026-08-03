---
title: "广告归因基础（模型、窗口与隐私时代方案）"
---

# 广告归因基础（模型、窗口与隐私时代方案）

> **版本** 2026-08 · **定位**：产品/运营向 · 转化功劳归属的方法论与技术

归因（Attribution）解决的核心问题：**一次转化（下载、激活、付费）的功劳算给谁**。本文讲清归因模型、归因窗口、确定性/概率匹配，以及 ATT 之后 iOS 的 SKAN 与 AEM。

---

## 目录

1. 快速选型
2. 定义与核心问题
3. 归因模型
4. 归因窗口
5. 归因方法：确定性 vs 概率
6. 归因优先级（瀑布流）
7. 隐私时代：iOS 的 SKAN 与 AEM
8. MMP 与归因不一致
9. 参考链接
10. 相关笔记

---

## 1. 快速选型

| 需求 | 结论 | 章节 |
| ---- | ---- | ---- |
| 默认效果归因口径 | Last Click（最后点击）+ 7 天点击 / 1 天浏览窗口 | §3、§4 |
| iOS 投放优化与结算 | 日常优化看 AEM，最终结算/跨渠道核对看 SKAN | §7 |
| 安卓国内投放 | 渠道包 + 厂商商店 Install Referrer + OAID | §5.2 |
| 平台报表与自建数据对不上 | 三口径对齐：平台自归因 vs MMP vs 自建（§8） | §8 |
| 品牌曝光价值评估 | 展示归因（View-through）或多触点模型 | §3、§4 |

---

## 2. 定义与核心问题

归因指判断是什么原因促使用户产生转化（下载、激活、再营销互动）。流量分两类：

- **非自然流量**：用户与媒体渠道有过互动（展示或点击）——这类转化可被归因
- **自然流量**：用户没有与任何渠道互动，直接下载/打开——严格来说不被归因

**为什么难**：用户往往先看 A 平台展示、再点 B 平台广告、最后在 C 渠道下载——功劳如何分配？归因就是定规则把转化"记到账上"，规则不同、数据不同，结果就不同。这也是投放数据争议的根源（§8）。

---

## 3. 归因模型

| 模型 | 规则 | 适用 |
| ---- | ---- | ---- |
| **Last Click（最后点击）** | 功劳 100% 给转化前最后一次点击 | **默认口径**，MMP 与平台的主模型；之前的点击算「助攻」（assist） |
| **First Click（首次点击）** | 功劳 100% 给第一次触达 | 品牌/获客渠道的价值评估 |
| **Linear（线性）** | 路径上每次点击平分 | 简单多触点 |
| **Time Decay（时间衰减）** | 越接近转化权重越高 | 转化路径短的品类 |
| **Position Based（位置归因）** | 首次与最后一次各 40%，中间 20% | 首发拉新 + 末次转化的复合场景 |
| **数据驱动（DDA/MTA）** | 用模型学各触点贡献 | 数据量大、追求精细；受隐私限制（§7）后难以实施 |

⚠️ 隐私时代（iOS ATT 之后）用户级身份数据大幅缺失，多触点模型在 iOS 流量上基本无法运作，实际退化为「活动级别统计推断」（§7）。

---

## 4. 归因窗口

归因窗口指「广告事件（点击/浏览）到转化之间允许归因的时间范围」：

| 窗口类型 | 默认 | 说明 |
| ---- | ---- | ---- |
| **点击归因窗口** | 7 天（可 1–30 天调整） | 点击后 7 天内产生的激活算该渠道的；概率模型的点击窗口最多 24 小时 |
| **浏览归因窗口（View-through）** | 1 天（0–24 小时可调） | 只展示未点击；比点击窗口短得多 |
| **转化事件窗口** | 通常 24 小时 | 激活后再发生的事件（付费等）在窗口内回传 |

规则：既有点击又有浏览时，**总是优先归因点击**；同窗口多点击时以最后一次为准（Last Click）。

---

## 5. 归因方法：确定性 vs 概率

### 5.1 总览

| 方法 | 机制 | 准确度 | 适用 |
| ---- | ---- | ---- | ---- |
| **确定性匹配** | 设备 ID / Referrer / 渠道包等强标识直接匹配 | 高，点击型设备匹配准确率可达 100% | 有确定性信号的主流渠道 |
| **概率匹配（指纹）** | 用 IP、设备型号、系统版本、语言、UA 等弱信号建模匹配 | 低，受信号质量影响；iOS 上会丢数据 | 无确定性标识时兜底（SKAN 流量、Web-to-App、跨设备） |

**归因瀑布流**：优先确定性，其次概率；概率归因仅在确定性不可用时自动启用。

### 5.2 主要确定性手段

| 手段 | 平台 | 原理 | 说明 |
| ---- | ---- | ---- | ---- |
| **设备 ID 匹配** | iOS / Android | 广告平台点击时上报设备 ID，MMP 与 App SDK 采集的设备 ID 匹配 | iOS 主手段：IDFA（ATT 受限后失效）、IDFV；安卓：GAID、OAID、Android ID |
| **Install Referrer** | Android（Google Play + 部分厂商商店） | 通过 Referrer API 获取跳转商店前的原始 URL | 安卓端归因的主要方法；同时有 referrer 和设备 ID 匹配时优先 referrer |
| **渠道包** | Android（国内） | APK 内预写渠道 ID，首次运行识别即知来源 | 国内安卓最普及的兜底手段 |
| **自归因平台（SAN/SRN）** | Meta、Google Ads、TikTok 等 | 大媒体用自己的点击/展示数据自行认领安装，通过 MMP API 回传 | 结果由平台说了算（§8 的差异来源之一） |
| **深度链接归因** | 双端 | 唤起时携带来源参数（见 [advertising-delivery-modes.md](./advertising-delivery-modes.md) §8） | 与场景还原配合 |
| **剪贴板归因** | H5 引流场景 | 复制唯一识别码到剪贴板，App 打开后读取 | 国内 H5→App 常用 |

### 5.3 概率匹配的信号

IP 地址、设备型号、操作系统版本、设备语言、User Agent。适用：iOS SKAN 流程后的激活、Web-to-App 转化、跨设备路径（桌面看广告、手机端激活）、CTV/PC 端。

---

## 6. 归因优先级（瀑布流）

```
安装 / 激活事件
    │
    ├─▶ 预装归因（OEM / 运营商预装，优先级最高）
    ├─▶ 点击归因（窗口内最近一次点击）
    │      └─ 确定性匹配优先（设备 ID / Referrer），无信号走概率
    ├─▶ 浏览归因（窗口内浏览，仅无点击时）
    └─▶ 自然流量（任何渠道都不匹配）
```

- 优先归因给**点击**，其次**展示**；优先**确定性**，其次**概率**
- 例：A 渠道提供了 IDFA（确定性）但点击时间更早，B 渠道只提供点击（概率）且时间更晚——**确定性匹配优先获胜**
- 卸载重装：重装后的转化归因到原始安装渠道而非自然量（再归因），用于唤醒沉睡用户的再营销（再互动）场景

---

## 7. 隐私时代：iOS 的 SKAN 与 AEM

iOS 14.5 引入 ATT 框架后，IDFA 默认不可获取，用户级确定性归因失效，苹果用 **SKAdNetwork（SKAN）** 替代：聚合、延迟、匿名的回传信号。Meta 则推出自家的 **AEM**。

### 7.1 SKAN 4.0 核心机制

- 用户点击广告并安装后，设备经**隐私计时器**延迟，向广告网络发送回传（postback），包含广告系列 ID、来源 App ID、转化值（conversion value）
- **多层转化窗口**：最多 3 次回调，覆盖安装后最长 35 天——回调 1（0–2 天，延迟 24–48 小时）、回调 2（3–7 天）、回调 3（8–35 天）；对比 SKAN 3 的 1–3 天窗口大幅拉长
- **数据级别（隐私阈值）**：0–3 四级，级别越高回传越精细（来源标识符 2–4 位、细粒度/粗粒度转化值）；未达隐私阈值的广告系列只回传粗粒度值
- 优点：**确定性**（苹果官方数据）；缺点：**延迟**（最长约 41 天）、聚合无法看用户级、需集中预算越过隐私阈值

### 7.2 SKAN vs AEM

| 维度 | AEM（Meta） | SKAN（苹果） |
| ---- | ---- | ---- |
| 归属 | Meta 自家平台 | 苹果官方，跨全网 iOS 渠道 |
| 实时性 | 近乎实时，利于优化 | 延迟 1–3 天（最长约 41 天） |
| 归因模型 | 1 天/7 天点击、展示归因、再营销 | 逻辑固定，由苹果决定 |
| 数据性质 | **建模推算**，波动较大 | 确定性（硬数据），受隐私阈值限制 |

**实操结论**：日常优化看 AEM（实时可调出价、关停素材），最终结算与跨渠道核对看 SKAN；配合 MMP 做统一视图与去重。

---

## 8. MMP 与归因不一致

**MMP（Mobile Measurement Partner）**：第三方归因平台（AppsFlyer、Adjust、Singular、Branch 等），提供归因、安装跟踪、事件追踪、深度链接、反作弊与报告，是投放数据的"运营层"。

**选型要点**：渠道覆盖（是否对接你的投放平台）、定价、功能深度（W2A 能力如 Adjust）、隐私合规（SKAN/AEM 支持）。

**为什么报表数字总对不上**（三大口径）：

| 口径 | 来源 | 特点 |
| ---- | ---- | ---- |
| 平台自归因（SAN 认领） | Meta / Google / TikTok 自己算 | 大媒体说了算，通常对自己有利 |
| MMP 归因 | 第三方 SDK + 二次匹配 | 相对中立，统一视图 |
| 自建统计 | 自有服务端 | 只认自己定义的转化 |

差异来源：归因窗口与模型不同、概率建模误差（iOS 上部分激活流入自然量）、SKAN 延迟导致结算时点不同、反作弊剔除规则不同、暗投（dark spend，无法识别来源的渠道消耗）。

**对齐建议**：确认统一归因窗口与模型；以 SKAN 为准核跨渠道总量；用调整系数校准 AEM/建模数据与 MMP 的偏差；单渠道投放可用**基线增量法（baseline uplift）**计算真实增量。

---

## 9. 参考链接

以下为本文撰写时参考的公开资料：

- [SKAdNetwork 4 原理（Adjust）](https://www.help.adjust.com/zh/article/how-skadnetwork-4-works)
- [什么是 SKAdNetwork（Tenjin）](https://docs.tenjin.com/docs/zh/introduction-to-skadnetwork)
- [SKAN vs Probabilistic vs SSOT: Finding Your Attribution Happy Place（Aarki）](https://www.aarki.com/zh/insights/skan-vs-probabilistic-vs-ssot/)
- [Modeling attribution on iOS: what works, what doesn't（RevenueCat）](https://www.revenuecat.com/blog/growth/ios-attribution-guide-skan-aem-probabilistic/)
- [Meta 中的 AEM 是什么（小红书）](https://www.xiaohongshu.com/discovery/item/696ef192000000002202d60b)
- [AppsFlyer 归因模型](https://support.appsflyer.com/hc/zh-cn/articles/207447053-AppsFlyer%E5%BD%92%E5%9B%A0%E6%A8%A1%E5%9E%8B)
- [概率建模归因 - 合作渠道指南（AppsFlyer）](https://support.appsflyer.com/hc/zh-cn/articles/40188089592209-%E6%A6%82%E7%8E%87%E5%BB%BA%E6%A8%A1%E5%BD%92%E5%9B%A0-%E5%90%88%E4%BD%9C%E6%B8%A0%E9%81%93%E6%8C%87%E5%8D%97)
- [Adjust 的归因方法](https://www.help.adjust.com/zh/article/attribution-methods)
- [APP 用户归因技术解读与应用（Xinstall）](https://blog.csdn.net/Xinstall/article/details/150421666)
- [海外 MMP 归因基础知识（AMZ123）](https://www.amz123.com/t/u2M0mqDB)

---

## 10. 相关笔记

- [advertising-delivery-modes.md](./advertising-delivery-modes.md) — 投放模式中的归因差异：直投依赖 SKAN/Referrer，Web2App 可用 token 透传端到端归因
- [advertising-basics.md](./advertising-basics.md) — 本系列基础篇：ROAS/ROI 指标与冷启动考核
