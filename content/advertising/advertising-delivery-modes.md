---
title: "投放模式：Web2Web / Web2App / App2App（深度链接与兜底链路）"
---

# 投放模式：Web2Web / Web2App / App2App（深度链接与兜底链路）

> **版本** 2026-08 · **定位**：产品/运营向 · 投放转化链路的组织形态

投放模式（Delivery Mode）指广告点击后「转化链路」的组织形态：落地到网页、引导进 App，还是直接唤起 App 内页。本文讲透 Web2Web、Web2App（含直投）、App2App 与背后的 Deeplink（深度链接）技术。

---

## 目录

1. 快速选型
2. 模式总览与链路图
3. Web2Web
4. Web2App：两段式与直投
5. App2App：应用直达
6. 深度链接（Deeplink）技术基础
7. 失败兜底与「三链直达」
8. 归因与监测要点
9. 参考链接
10. 相关笔记

---

## 1. 快速选型

| 场景 | 推荐模式 | 说明 |
| ---- | ---- | ---- |
| 纯网页业务（官网、H5 电商、线索表单） | Web2Web | 全链路在浏览器完成，无 App 依赖 |
| App 拉新下载，需要可控的引导文案/可测试落地页 | Web2App 两段式（H5 中转） | 可收集 UTM/邮箱做端到端归因，落地页可小时级迭代 |
| App 拉新，链路越短越好 | Web2App 直投 / 直达商店 | 点击直达商店或唤起，归因依赖 MMP（SKAN 等） |
| App 已装用户召回、促活（电商内页直达） | App2App（应用直达） | 点击直接唤起 App 内指定页，转化路径最短 |
| 拉新 + 促活混合 | App2App + 场景还原（Deferred Deeplink） | 未装用户下载后首次打开仍还原目标页 |

---

## 2. 模式总览与链路图

### 2.1 定义

投放模式的本质差异在三点：**转化目标**（网页还是 App）、**是否经过中转页**（H5 引导页）、**能否直达内页**（Deeplink 唤起）。由此分出三种基本模式，直投是 Web2App 的短链路变体：

```
广告点击
  ├─▶ [Web2Web]       H5 落地页 ──▶ 网页转化（表单 / 下单）
  ├─▶ [Web2App 两段式] H5 引导页 ──▶ 商店 / 渠道包下载 ──▶ 打开 App
  ├─▶ [Web2App 直投]   Deeplink 直接唤起 App（已装）或跳下载（未装）
  └─▶ [App2App 应用直达] 已装 → 唤起 App 内指定页
                       未装 → 商店下载 → 首次打开场景还原到指定页
```

### 2.2 模式对照

| 模式 | 链路 | 转化目标 | 优势 | 劣势 |
| ---- | ---- | ---- | ---- | ---- |
| **Web2Web** | 广告 → H5 落地页 | 网页行为 | 最简单、无 App 依赖、任意投放平台 | 无法复用 App 内能力（支付/推送/登录态），浏览器转化率相对低 |
| **Web2App 两段式** | 广告 → H5 引导页 → 下载/唤起 | App 下载或激活 | 可端到端归因（UTM/token 透传）、落地页可快速迭代文案与定价 | 链路长、中转页流失大 |
| **Web2App 直投** | 广告 → Deeplink 唤起 / 商店下载 | App 下载或激活 | 链路短，直达商店 | 归因依赖商店回传（SKAN/Install Referrer），落地页无法做文案测试 |
| **App2App（应用直达）** | 广告 → 唤起 App 内指定页 | App 内行为（下单、加购） | 转化路径最短（6 步 → 2 步），体验最好 | 依赖深度链接体系；未装用户需场景还原兜底 |

---

## 3. Web2Web

广告点击后落到 H5 网页落地页，全链路在浏览器内完成。

- **典型业务**：官网投放、H5 电商、线索收集（表单留资）、内容推广
- **优点**：链路最简单；不依赖 App 生态；任何媒体都能投
- **缺点**：网页转化率整体低于 App 内体验；无法利用 App 的支付、推送、登录态等能力；对电商类广告主而言，H5 的支付转化通常弱于 App 内
- **要点**：落地页加载速度与表单长度直接决定 CVR（对照 [advertising-basics.md](./advertising-basics.md) §5 的 CVR 诊断）

---

## 4. Web2App：两段式与直投

Web2App 指广告最终目标是把用户引导进 App（下载或唤起），分为两种形态：

### 4.1 两段式（H5 中转）

广告点击后先打开 H5 引导页，再通过引导按钮进入商店或唤起 App。

- **优点**：Web 端可捕获 UTM、click_id 并存入 cookie/服务端，实现端到端归因（广告点击 → 订阅/付费全链路可追溯）；可以小时级迭代落地页文案、定价、试用策略，不必等 App 发版；可先收集邮箱/手机号，作为跨端身份拼接主键，解决「安装后失联」
- **缺点**：链路长，引导页与商店跳转之间流失大
- **适用**：归因要求高、需要快速测试落地页的买量场景

### 4.2 直投（Direct-to-app）

点击广告后直接跳转 App Store / Google Play 下载页（海外语境），或直接通过 Deeplink 唤起已装 App。

- **优点**：链路最短，转化路径从「点击→搜索→筛选→找到内容」6 步缩短为「点击→直达」2 步
- **缺点**：UTM 参数无法跨越商店跳转，归因主要依赖 SKAN 回传和 MMP postback，渠道收入归因不清晰；无法在落地页做文案/定价测试
- **适用**：链路简单、依赖 MMP 归因的买量场景

---

## 5. App2App：应用直达

用户点击 App 媒体（信息流/激励位）上的广告后，**直接到达广告主 App 内指定的落地页**（而非首页），这就是各家平台的「应用直达」/「一键唤起」（腾讯体系称「应用直达」，巨量体系常见「App2App」叫法）。

- **已安装用户**：直接唤起 App 并跳转到指定页面（商品详情、结算页、活动页）——转化路径最短，电商类广告主效果最明显
- **未安装用户**：引导去商店下载，安装完成后**首次打开仍跳转到指定页面**（场景还原，见 §6.4）
- **投放配置**：需将 App 的 Scheme 格式写入直达页地址（形如 `example://showProduct?Id=330885`，不支持 http(s) 地址）；iOS 投放同时配置 Universal Link，让流量端优先用通用链接唤起
- **适用**：已装用户召回/促活、电商内页直达、跨 App 导流

---

## 6. 深度链接（Deeplink）技术基础

应用直达的核心是深度链接：把用户路由到 App 内具体页面，而不是停留在网页或 App 首页。

### 6.1 技术对比

| 技术 | 机制 | 优缺点 | 适用 |
| ---- | ---- | ---- | ---- |
| **URL Scheme**（`myapp://product/123`） | 自定义协议拉起 App | 实现简单、双端通用；但 iOS 弹窗确认易流失、可被拦截/伪造、微信内被屏蔽、易丢归因信号 | 简单场景；作为兜底手段 |
| **Universal Link**（iOS 9+） | 标准 HTTPS 域名 + 服务器部署 `apple-app-site-association` | 免确认弹窗、体验自然、来源可验证、不丢归因信号——iOS 最稳定唤起方式 | iOS 投放主链（Meta AEM 归因必须） |
| **App Links**（Android 6+） | HTTPS + 数字资产验证 | 类似 Universal Link 的无缝跳转 | Android 主链 |
| **Deferred Deeplink**（延迟深度链接/场景还原） | 未安装时先记录点击意图，安装后首次打开还原目标页并保留参数 | 拉新投放的核心能力；需第三方 SDK（AppsFlyer/Adjust/Branch/openinstall 等） | 买量拉新、Web2App |

### 6.2 关键差异：标准链接 vs 延迟深度链接

- **标准深度链接**：仅 App **已安装**时生效；未安装会报错或跳网页——用户被送去商店后链接上下文丢失，安装完落到首页，必须重新寻找内容
- **延迟深度链接（Deferred Deeplink）**：未安装用户点击 → 系统记录点击意图与参数（渠道、活动 ID、邀请人）→ 用户去商店下载 → 安装后首次打开直接进入原目标页面并还原上下文

**结论**：投放广告必须上延迟深度链接，普通 Deeplink 只能覆盖已装用户。

### 6.3 特殊环境：微信

微信内置浏览器屏蔽 Scheme 直接跳转，需使用**微信开放标签**（`wx-open-launch-app`），配合服务号认证、开放平台移动应用关联、签名与 Web SDK/App SDK 回调等全套配置；开放标签失效时需其他能力兜底。

---

## 7. 失败兜底与「三链直达」

### 7.1 唤端失败原因

用户未安装 App、用户取消唤起确认弹窗、媒体环境限制外跳（微信/微博）、浏览器主动拦截或限制自动唤起、唤端请求发起时机过晚等。

### 7.2 兜底优先级

```
应用直达（Universal Link 优先 → URL Scheme 兜底）
    ↓ 失败
小程序落地页（微信生态）
    ↓ 失败 / 返回
H5 落地页（保证链路不断）
```

**「三链直达」**：点击广告时先执行应用直达跳转并同时加载 H5 落地页；直达失败则跳小程序落地页；用户从 App/小程序落地页返回媒体时展示 H5 落地页。实测相比「双链」（应用直达 + H5）**转化成本降低约 15%、转化率提升 1 个百分点**。

要点：广告后台需选「应用直达 + 小程序兜底」而非「应用直达 + 小程序」；Deeplink 必填且需同时配置 Universal Link；自定义 H5 落地页需取消自动执行 H5 直达功能（否则用户无法返回媒体）。

### 7.3 兜底策略清单

| 失败场景 | 兜底动作 |
| ---- | ---- |
| 未安装 App | 跳商店/下载页 + 场景还原（Deferred Deeplink） |
| iOS 唤起弹窗被取消 | 改用 Universal Link 免除弹窗 |
| 微信/微博环境限制外跳 | 微信开放标签；微博提示右上角浏览器打开 |
| 唤端时机过晚/超时 | 定时器 + 页面可见性监听（visibilitychange）检测唤起是否成功，超时自动降级商店/H5/小程序 |

---

## 8. 归因与监测要点

- **直投的归因难点**：商店跳转会丢 UTM 参数，需依赖 **SKAN（iOS）/ Install Referrer（Android）** 回传和 MMP postback；渠道收入归因不清晰
- **Web2App 的归因优势**：UTM/click_id 存于 Web 端，通过**不透明 token**（而非裸 UTM）透传给 App，App 首次打开用 token 从后端拉取归因数据；购买事件走**服务端到服务端（S2S）**上报，携带 user_id/session_id 与原始 UTM 关联
- **归因窗口**：建议 7 天点击 / 1 天浏览，以「最后付费点击」为准；用户同时触达两条路径时以最后一次付费点击归属
- **核心监测指标**：唤端率（唤起成功 UV ÷ 唤起请求 UV）、下载完成率、激活率
- **平台限制**：iOS Safari 特定重定向会丢参数；Google Ads Web to App Connect 不支持自定义 Scheme 和第三方重定向链接，必须使用 App Links / Universal Links

---

## 9. 参考链接

以下为本文撰写时参考的公开资料：

- [「应用直达」广告投放指引（腾讯广告）](https://tencentads.com/Faqlist/Detail?id=359)
- [Deferred Deeplink vs Standard Deep Link（Airbridge）](https://www.airbridge.io/en/blog/deferred-deeplink-vs-standard-deep-link)
- [从网页到 App 无缝跳转：openinstall 场景还原技术全链路解析（腾讯云）](https://cloud.tencent.cn/developer/article/2502569)
- [Create an effective deep-linking strategy using Web to App Connect（Google Ads）](https://support.google.com/google-ads/answer/16401018)
- [Deep Linking for App Growth: Universal Links and App Links（Appalize）](https://www.appalize.com/da/blog/user-acquisition/deep-linking-for-app-growth-universal-links-and-app-links-strategy)
- [你不知道的转化路径——三链直达（小红书）](https://www.xiaohongshu.com/discovery/item/68a5eb0c000000001d00d246)
- [Deeplink 唤端失败原因及优化唤起率（小红书）](https://www.xiaohongshu.com/discovery/item/66a791350000000027010783)
- [网页跳转 App 统计实现：一键拉起监测点击与安装量（xinstall）](https://www.xinstall.com/article/11449)

---

## 10. 相关笔记

- [advertising-basics.md](./advertising-basics.md) — 本系列基础篇：生态角色、计费出价、CTR/CVR 诊断
- [feed-stream-push-pull.md](../feed-stream-push-pull.md) — 信息流（Feed）是 App2App 应用直达的主要广告位形态
