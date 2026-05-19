# RabbitMQ 核心概念

> **版本** 2026-05 · 适用于架构设计、技术评审与实现参考

> Broker、Connection、Channel、Virtual Host、Exchange、Queue、Binding 与消息在 AMQP 模型中的流转。

---

## 如何使用本文档

| 你的目标 | 建议阅读 |
| -------- | -------- |
| 5 分钟建立心智模型 | 一、AMQP 模型总览 |
| 对接客户端前必读 | 二、Connection 与 Channel |
| 排障「消息去哪了」 | 三、Exchange / Queue / Binding |
| 上线前 checklist | 五、生产实践 |

**核心原则**：RabbitMQ 是**协议级消息中间件**；正确性依赖 Exchange 路由、持久化、Confirm 与 ACK 的组合，而不是「发到队列就完事」。

---

## 一、AMQP 模型总览

### 1.1 定义

**AMQP 0-9-1**（Advanced Message Queuing Protocol）把消息系统抽象为：**生产者**把消息发到 **Exchange**，Exchange 按 **Binding** 规则把消息路由到 **Queue**，**消费者**从 Queue 拉取或被动接收投递。

RabbitMQ 是 AMQP 的开源实现，并扩展了 **Publisher Confirm**、**Consumer ACK**、**TTL**、**DLX** 等能力。

### 1.2 组件关系

```mermaid
flowchart LR
    P[Producer 生产者] -->|publish| EX[Exchange]
    EX -->|binding| Q1[Queue A]
    EX -->|binding| Q2[Queue B]
    Q1 --> C1[Consumer 1]
    Q2 --> C2[Consumer 2]
    subgraph broker["Broker 节点"]
        EX
        Q1
        Q2
    end
```

ASCII 等价：`Producer → Exchange → (Binding) → Queue → Consumer`

### 1.3 与「直接发到队列」的区别

| 维度 | 直接 Queue | 经 Exchange |
| ---- | ---------- | ----------- |
| 路由 | 单队列 | 多队列、模式匹配、广播 |
| 解耦 | 生产者需知队列名 | 生产者只知 Exchange + Routing Key |
| 典型 | 简单脚本 | 微服务、事件总线 |

**读写要点**：应用代码里 `basicPublish(exchange, routingKey, body)`；队列名通常不出现在 publish 参数里（除非发到 default exchange）。

### 1.4 优缺点

| 优点 | 缺点 |
| ---- | ---- |
| 路由灵活、生态成熟 | 运维与调参比 Redis List 重 |
| 持久化、集群、插件丰富 | 极端超高吞吐场景可能不如 Kafka |
| 语义完整（ACK、DLX） | 错误配置易导致堆积或丢消息 |

### 1.5 典型业务

- 订单创建后通知库存、积分、风控等多个下游
- 异步解耦：HTTP 接口只投递，慢任务由 Worker 消费
- 跨语言系统集成（Java 生产、Go 消费）

---

## 二、Broker 与 Virtual Host

### 2.1 Broker

**Broker** 指 RabbitMQ 服务进程（单节点或集群中的一个 Erlang 节点）。对外提供 AMQP（5672）与管理 HTTP API（15672）。

### 2.2 Virtual Host（vhost）

**vhost** 是逻辑隔离单元：Exchange、Queue、权限、策略均挂在某个 vhost 下。默认 `"/"`。

| 要点 | 说明 |
| ---- | ---- |
| 隔离 | 不同环境 / 租户用不同 vhost |
| 权限 | 用户需被授予 vhost 上的 configure / write / read |
| 连接 URI | `amqp://user:pass@host:5672/my_vhost` |

### 2.3 优缺点与业务

- **优点**：一套集群多环境，无需多套物理集群
- **缺点**：误配权限可能跨 vhost 不可见，排障要先确认 vhost
- **典型**：`prod`、`staging` 各一 vhost；多租户 SaaS 每租户一 vhost

---

## 三、Connection 与 Channel

### 3.1 定义

- **Connection**：TCP 连接 + 认证，重量级，应**长连接、少创建**。
- **Channel**：在 Connection 上的轻量级逻辑会话，实际 `publish` / `consume` 都在 Channel 上执行。

### 3.2 流程（客户端）

```mermaid
sequenceDiagram
    participant App as 应用
    participant Conn as Connection
    participant Ch as Channel
    participant Br as Broker
    App->>Conn: 建立 TCP + 认证
    App->>Ch: connection.createChannel()
    App->>Br: declare exchange / queue
    App->>Br: basicPublish / basicConsume
```

### 3.3 读写要点

| 实践 | 原因 |
| ---- | ---- |
| 每线程/协程独立 Channel | Channel 非线程安全 |
| 连接池复用 Connection | 避免频繁握手 |
| 捕获 Channel 关闭异常 | 网络闪断会关 Channel，需重建 |
| `channelMax` 注意上限 | 单连接 Channel 数有限 |

### 3.4 优缺点

| 优点 | 缺点 |
| ---- | ---- |
| 多路复用，省 TCP | Channel 泄漏会导致 broker 资源耗尽 |
| 符合 AMQP 标准 | 异步回调模型学习曲线 |

### 3.5 典型业务

Web 应用：进程启动时建 1 个 Connection，请求线程从池取 Channel 发消息。

---

## 四、Exchange、Queue、Binding

### 4.1 Exchange

**Exchange** 只负责接收消息并路由，**不存储**消息（无绑定或未匹配时可能丢弃或 return）。

| 类型 | 路由依据 | 详见 |
| ---- | -------- | ---- |
| direct | Routing Key 精确匹配 | [exchanges-routing.md](./exchanges-routing.md) |
| fanout | 忽略 Routing Key，广播 | 同上 |
| topic | Routing Key 模式匹配 | 同上 |
| headers | 消息头匹配 | 同上 |

### 4.2 Queue

**Queue** 存储消息，供消费者消费。

| 属性 | 说明 |
| ---- | ---- |
| durable | 队列元数据持久化，重启后仍在 |
| exclusive | 仅当前连接可用，连接断则删 |
| autoDelete | 无消费者时自动删除 |
| arguments | `x-message-ttl`、`x-dead-letter-exchange` 等 |

### 4.3 Binding

**Binding** 是 Exchange 到 Queue 的链接，含可选 **Routing Key**（及 headers 匹配规则）。

```
Exchange "orders" --[binding key: created]--> Queue "order.created"
```

### 4.4 消息在 Broker 内的状态

| 状态 | 含义 |
| ---- | ---- |
| ready | 在队列中等待投递 |
| unacked | 已投递给消费者，等待 ACK |
| 无消费者 | 消息堆积在 ready（受 max-length 等限制） |

### 4.5 典型业务

- 订单 Exchange `order.topic`，Routing Key `order.created` → 库存队列
- 同一事件 fanout 到审计、搜索索引、通知三个队列

---

## 五、Message 与投递语义预览

### 5.1 Message 结构

| 部分 | 内容 |
| ---- | ---- |
| payload | 业务 body（建议 JSON + schema 版本） |
| properties | `content_type`、`delivery_mode`、`message_id`、`timestamp` |
| headers | 自定义元数据；headers exchange 路由用 |

**delivery_mode**：`1` 非持久化，`2` 持久化（需队列 durable + 消息 persistent 才落盘）。

### 5.2 与系列其他篇的关系

| 主题 | 文档 |
| ---- | ---- |
| Exchange 类型详解 | [exchanges-routing.md](./exchanges-routing.md) |
| 生产者可靠投递 | [reliable-publishing.md](./reliable-publishing.md) |
| 消费 ACK / DLX | [consumer-semantics.md](./consumer-semantics.md) |
| TTL / 延迟 | [delay-and-priority.md](./delay-and-priority.md) |
| 集群 HA | [cluster-ha.md](./cluster-ha.md) |

---

## 六、生产实践

### 6.1 声明与发布分离

| 项 | 建议 |
| ---- | ---- |
| 拓扑声明 | 部署脚本 / 基础设施即代码（Terraform、Operator）统一 declare |
| 应用 | 假设 Exchange/Queue 已存在，或启动时 idempotent declare |
| 命名 | `{业务}.{事件}` 如 `order.payment.completed` |

### 6.2 上线 checklist

- [ ] vhost、用户权限最小化
- [ ] 队列 durable + 消息 persistent（需持久时）
- [ ] 已配置 DLX 与 max-length（防无限堆积）
- [ ] 监控：队列深度、publish/deliver 速率、unacked 数量
- [ ] 连接/Channel 泄漏检测

### 6.3 常见误区

| 误区 | 正确理解 |
| ---- | -------- |
| 「发到 RabbitMQ 就不会丢」 | 需 Confirm + 持久化 + 正确 ACK |
| 「一个 Connection 全局共用一个 Channel」 | 多线程必须多 Channel |
| 「队列名写在 publish 里」 | 标准模型是 Exchange + Routing Key |

---

## 七、附录

### 7.1 术语表

| 术语 | 说明 |
| ---- | ---- |
| AMQP | 高级消息队列协议 |
| Routing Key | 生产者指定，Exchange 用于路由的字符串 |
| Consumer Tag | `basicConsume` 时客户端标识 |
| Prefetch | 单消费者未 ACK 消息上限，见消费篇 |

### 7.2 参考

- [RabbitMQ Documentation — AMQP 0-9-1 Model](https://www.rabbitmq.com/tutorials/amqp-concepts.html)
- [RabbitMQ — Connections and Channels](https://www.rabbitmq.com/docs/connections)
