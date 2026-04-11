# B+ 架构方案设计

## 1. 文档信息

**文档名称**：手机端轻量 Agent 客户端 B+ 架构方案设计
**版本**：v1.1 Draft
**日期**：2026-04-11
**对应 PRD**：`docs/prd.md`
**目标读者**：产品、客户端、服务端、daemon、provider adapter、runtime backend 开发

---

## 2. 设计结论

B+ 优化部分定位为 V1.1 演进，V1 主链路保持不变：

```text
WebRTC/Data Transport
  + signaling / room coordination
  + TURN fallback
  + remote daemon / session gateway
  + provider adapter
  + Session Runtime Backend
```

一句话定义：

> WebRTC/Data Transport 是实时传输层，daemon/session gateway 是结构化状态层，Session Runtime Backend 是会话运行与恢复层，手机 App 是可视化控制层。

这里的 WebRTC/Data Transport、signaling / room coordination、TURN fallback 等增强能力属于 V1.1 演进，不表示 V1 已切换当前主链路。

V1.1 不要求一开始自建完整 WebRTC 基础设施。M0 优先走本地局域网验证；M1/M2 可以使用 LiveKit Cloud 等 PaaS 托管 signaling、NAT 穿透和 TURN fallback；生产阶段再根据成本、稳定性和数据边界决定继续 PaaS、自建 LiveKit，或回到自研 signaling + coturn。

本方案不做 WebRTC 远程终端，不做纯聊天客户端，也不把 CLI/tmux 作为长期唯一运行形态。即使进入 V1.1，runtime 默认仍通过 tmux/pty CLI backend 托管当前 CLI 形态的 Codex；未来如果 Codex、Claude Code 或其他 provider 演进为 server/native session 模式，可以切换到 `ProviderServerBackend` 或 `CloudAgentBackend`，而不影响 App、协议层和信息流体验。

---

## 3. 核心设计原则

### 3.1 信息流优先

手机端主体验是 coding agent 工作信息流，不是聊天流，也不是终端字符流。

信息流应优先表达：

* 当前 session 状态
* 当前步骤摘要
* 工具调用
* 文件变化
* 命令和测试结果
* attention / confirmation
* 阶段性总结
* 必要时的 terminal peek

### 3.2 分层解耦

各层只依赖稳定接口，不依赖彼此内部实现：

* App 不绑定 provider 输出格式
* App 不关心 session 是 CLI、server 还是 cloud runtime
* WebRTC 不承担业务语义，只传输协议消息
* daemon 不做二次 agent，只做状态标准化和轻控制
* provider adapter 屏蔽 provider 差异
* runtime backend 屏蔽会话运行方式差异

### 3.3 V1.1 可落地

V1.1 只支持 Codex，但架构要为多 provider 和多 runtime 留接口。

V1.1 默认 runtime 是 tmux/pty CLI backend，因为当前本地/远端 coding agent 大量仍以 CLI 方式运行。但所有上层协议都不能写死 tmux。

### 3.4 前台实时，后台 push-only

移动端不承诺后台长连接实时在线。

V1.1 的实时能力只在 App 前台打开 session 时提供：

* App 前台：建立 WebRTC DataChannel，实时接收事件和发送控制消息
* App 后台/锁屏：关闭或挂起实时连接，依赖 push notification 唤醒用户
* App 重新打开：使用 `reconnect.resume` 和 cursor 补拉缺失事件

低延迟接力指的是“用户打开 App 处理 attention 时的低延迟”，不是“手机后台持续低延迟在线”。

### 3.5 单用户，多读单写

V1 不做多人协作，也不做复杂多端同时控制。

V1 约束：

* 单用户拥有 daemon
* 多设备可以观察同一 session
* 同一 session 同一时间只有一个 active controller
* 写操作必须幂等
* confirmation 只能被 resolve 一次
* 后续设备响应已处理 attention 时返回 `already_resolved`

这样可以避免 V1 过早进入协同编辑、冲突合并和多主一致性问题。

### 3.6 安全边界前置

daemon 运行在用户服务器或开发机上，天然接近代码、终端输出、文件系统和 agent 输入能力。安全不是 V2 增强项，而是 V1 必做边界。

V1 必须默认启用：

* workspace allowlist
* 设备绑定和撤销
* pairing token 短期有效
* 高风险操作确认
* terminal peek 限制和审计
* daemon 普通用户权限运行
* signaling 不保存 agent 工作内容

---

## 4. 总体架构

```text
Mobile App
  -> Transport Layer
       -> M0 Local LAN WebRTC / local transport
       -> PaaS WebRTC provider, e.g. LiveKit Cloud
       -> Self-hosted LiveKit / custom signaling + coturn
  -> Remote Daemon / Session Gateway
  -> Provider Adapter
  -> Session Runtime Backend
       -> TmuxCliBackend / PtyCliBackend
       -> ProviderServerBackend
       -> CloudAgentBackend
       -> CustomRunnerBackend
```

从架构形态看，这是 C/S 架构，但不是传统单 server 架构。系统里有三个主要运行端：

```text
Mobile App：用户手机端客户端
Cloud Control Plane / PaaS Transport：账号、配对、room/signaling、push、TURN credential
Remote Daemon / Session Gateway：运行在用户服务器/开发机上的 agent-side server
```

其中实时数据面是 `Mobile App <-> Remote Daemon`。如果使用 LiveKit 等 PaaS，App 和 daemon 都是 room participant，Agent Work Protocol 通过 LiveKit data API 承载；如果自建，则可以直接使用 WebRTC DataChannel。无论底层 transport 如何变化，App 和 daemon 之间传输的业务语义仍是 Agent Work Protocol。

### 4.1 Mobile App

职责：

* 登录和设备绑定
* daemon 列表和在线状态
* session 列表
* session 详情页
* RunState 展示
* AgentEvent 信息流展示
* message.send
* confirmation.respond
* terminal.peek
* push notification 打开后的 session 定位
* 断线后的 cursor resume

不做：

* 完整 IDE
* 完整终端
* 复杂 diff 审阅
* 本地仓库管理
* 二次 agent 推理

### 4.2 Signaling Service

职责：

* 用户登录态校验
* 手机设备注册
* daemon 注册
* 设备配对
* daemon 在线状态维护
* WebRTC offer / answer / ICE candidate 交换
* TURN 配置下发
* push notification 路由

不做：

* 默认中转 agent 工作流数据
* 保存完整终端输出
* 承担 session runtime
* 对 agent 工作内容做智能分析

### 4.3 WebRTC DataChannel

职责：

* 低延迟传输结构化 agent 工作协议
* 传输用户轻量控制消息
* 传输心跳和连接状态
* 支持断线恢复的 cursor 协议
* 支持小窗口 terminal peek

不做：

* 远程桌面
* 完整终端流
* 大文件传输
* 长历史日志回放

### 4.3.1 Transport Provider 策略

Transport Layer 可以分阶段实现，避免 V1 一开始承担完整 WebRTC 基础设施复杂度。

#### M0：本地局域网验证

目标：

* 验证 Agent Work Protocol
* 验证 daemon/runtime/backend
* 验证 App/测试客户端的信息流体验

策略：

* daemon 和测试客户端在同一局域网
* 可以先不接入 TURN
* 可以使用临时 WebSocket transport 或本地 WebRTC/Pion 直连
* 如果 Flutter App 尚未启动，可先用 Go/Flutter 最小测试客户端
* 不验证复杂 NAT，不验证生产级 signaling

M0 的结论只用于判断业务协议和 daemon 是否成立，不用于证明公网移动网络可用。

M0 推荐顺序：

1. 先用本地 WebSocket transport 跑通 Agent Work Protocol、daemon、Event Store 和 Runtime Backend。
2. 再用 Flutter + LiveKit/Pion 最小客户端验证实时 data channel。
3. 最后再引入 PaaS room/token 流程。

#### M1/M2：PaaS Transport

推荐优先使用 LiveKit Cloud 等 PaaS。

PaaS 承担：

* room/signaling
* NAT traversal
* TURN fallback
* Flutter SDK
* 连接质量基础能力

我们仍然承担：

* daemon
* Provider Adapter
* Runtime Backend
* Agent Work Protocol
* Event Store / cursor resume
* attention / notification 业务判断
* 安全和审计

LiveKit data API 可用于承载 Agent Work Protocol。根据官方文档，LiveKit 的数据能力覆盖 text streams、RPC、data packets 等模式；data packets 支持 guaranteed 或 lossy 的消息投递选择。业务上仍需保留 Event Store 和 cursor，因为实时传输层不等于离线可靠队列。

LiveKit 接入形态：

```text
Mobile App participant
  <-> LiveKit room
  <-> Daemon participant
```

control plane 负责签发 LiveKit room token，App 和 daemon 使用 token 加入同一个 room。Agent Work Protocol 消息通过 LiveKit data API 发送，daemon/event store 仍然负责可靠业务状态。

#### 生产阶段：继续 PaaS 或自建

生产阶段再根据以下因素决策：

* PaaS 成本
* 中国网络可用性
* TURN fallback 比例
* 数据合规边界
* SDK 稳定性
* 是否需要自定义连接策略

可选路径：

* 继续使用 LiveKit Cloud
* 自建 LiveKit
* 自研 signaling + coturn

设计要求：

* Agent Work Protocol 不绑定 LiveKit
* Transport Provider 通过接口抽象
* App/daemon 可在不同 transport provider 之间切换

### 4.4 Remote Daemon / Session Gateway

职责：

* 本机 session 管理
* provider adapter 管理
* runtime backend 管理
* AgentEvent 标准化
* RunState 标准化
* message / confirmation 注入
* event store
* attention detection
* notification dispatch
* reconnect resume
* 本机权限和设备绑定

不做：

* 二次任务规划
* 二次工具调度
* 代替 Codex/Claude Code 推理
* 复杂多人协同

### 4.5 Provider Adapter

职责：

* 将 provider 的原始事件、输出、状态转成统一 Agent Work Protocol
* 将统一控制消息转成 provider 可接受的输入或 API 调用
* 屏蔽 Codex、Claude Code、自研 agent 的事件格式差异
* 暴露 provider capabilities

V1 默认：

* `CodexAdapter`

后续：

* `ClaudeCodeAdapter`
* `OpenCodeAdapter`
* `CustomAgentAdapter`

### 4.6 Session Runtime Backend

职责：

* 创建 session
* 恢复 session
* 停止 session
* 发送用户消息
* 响应确认
* 获取 snapshot
* 读取事件流
* 查询 capabilities

V1 默认：

* `TmuxCliBackend`
* `PtyCliBackend`

后续扩展：

* `ProviderServerBackend`
* `CloudAgentBackend`
* `CustomRunnerBackend`

---

## 5. 技术栈建议

### 5.1 推荐结论

V1.1 推荐技术栈：

| 模块 | 推荐语言 / 技术 | 选择理由 |
|---|---|---|
| Mobile App | Flutter + Dart | 更适合 Java/Go 背景，强类型、工程结构统一，适合信息流 UI 和跨端一致性 |
| Realtime SDK | LiveKit Flutter SDK；自建阶段可评估 `flutter_webrtc` | PaaS 阶段优先使用 LiveKit SDK，降低 WebRTC 原生细节复杂度 |
| Cloud Control Plane | Go | 用户熟悉 Go/Java，Go 适合 signaling token、room token、push 路由和轻量 API |
| Transport PaaS | LiveKit Cloud 优先验证 | 托管 room/signaling/TURN，降低 M1/M2 验证成本 |
| Remote Daemon / Session Gateway | Go | 静态二进制分发方便，适合长期运行、进程管理、pty/tmux、WebRTC、并发和网络 |
| WebRTC Go 实现 | Pion WebRTC 或 LiveKit Go SDK | M0 可用 Pion/本地 transport；PaaS 阶段优先接 LiveKit |
| Runtime Backend | Go interface | 与 daemon 同进程，便于管理进程、pty、事件流和本地状态 |
| Protocol Schema | JSON Schema / protobuf + 代码生成 | 先用 JSON 降低调试成本，再生成 Dart models 和 Go structs；后续可切 protobuf |
| Event Store | SQLite | daemon 本地轻量持久化，便于 cursor resume 和离线事件补拉 |
| Cloud DB | PostgreSQL | 存用户、设备、daemon registry、pairing、push token |
| Cloud Cache | Redis | daemon 在线状态、短期 pairing token、signaling 临时状态 |
| TURN | PaaS 内置 TURN；自建阶段用 coturn | 初期不自建 TURN，生产自建时再引入 coturn |

### 5.2 为什么 daemon 推荐 Go

Remote Daemon 是最靠近系统能力的一层，需要做：

* 本机进程生命周期管理
* pty/tmux 交互
* 长连接和 WebRTC DataChannel
* 本地 SQLite event store
* 文件系统和 workspace allowlist
* 单 binary 安装、升级、运行
* Linux/macOS 服务器环境适配

Go 在这些点上比 Node.js 更适合作为 daemon：

* 更容易发布单个静态二进制
* 并发模型简单稳定
* Pion WebRTC 成熟
* pty/process 管理生态足够
* 长期运行内存和部署可控

Rust 也适合 daemon，但 V1 开发成本更高；Node.js 适合快速原型，但做本机 daemon、pty、WebRTC、长期进程管理时打包和稳定性会更麻烦。

### 5.3 为什么 App 推荐 Flutter + Dart

Mobile App 的核心是信息流 UI 和轻量交互，不是重图形或本地复杂计算。

考虑到主要开发背景是 Java/Go，Flutter 比 React Native 更适合作为 V1 移动端技术栈。

Flutter 的优势：

* Dart 的强类型和类模型更接近 Java/Go 心智
* Flutter widget 体系统一，适合构建 AgentEvent 信息流、状态卡片和 confirmation card
* iOS/Android UI 一致性强
* 列表性能和动画控制适合实时信息流
* 配合 LiveKit Flutter SDK 可以降低 WebRTC 接入复杂度

注意：

* `flutter_webrtc` 或 LiveKit Flutter SDK 的 DataChannel/data API 稳定性必须在 M0/M1 验证
* push notification、secure storage、deep link、后台状态要从一开始按平台边界设计
* 协议类型需要从 schema 生成 Dart models，不能手写漂移

### 5.4 为什么 Cloud Control Plane 推荐 Go

Cloud Control Plane 主要是业务控制面，不是高性能数据面。

它负责：

* 用户登录
* 设备注册
* daemon registry
* pairing
* WebRTC signaling
* TURN credential 下发
* push notification routing

如果使用 LiveKit Cloud，control plane 在 V1 不需要自建完整 signaling，只需要负责：

* 用户和设备身份
* daemon registry
* pairing
* LiveKit room/token 签发
* push notification routing

这些能力用 Go 实现足够轻量，也能和 daemon 共享协议结构、错误码和部分工具代码。V1 可以选择：

* `chi`：简单、标准库风格强
* `Gin`：生态成熟，上手快
* `ConnectRPC` / gRPC：后续若需要强类型 RPC 可引入

建议 V1 使用 Go + chi，保持后端语言统一。

### 5.5 协议和类型共享

建议把协议定义放在独立目录：

```text
packages/protocol
  schemas/
  generated/
    dart/
    go/
```

V1 消息编码建议：

* DataChannel 业务消息使用 JSON
* 每条消息使用统一 envelope
* schema 使用 JSON Schema 或 protobuf schema 管理
* CI 中校验 schema 兼容性

后续如果消息量上升，可以在不改变语义协议的前提下切换为 MessagePack 或 protobuf。

### 5.6 工程仓库建议

如果采用 monorepo，建议结构：

```text
apps/
  mobile/                 # Flutter + Dart
  control-plane/          # Go control plane
daemon/
  cmd/lightac-daemon/      # Go daemon entry
  internal/session/
  internal/provider/
  internal/runtime/
  internal/webrtc/
  internal/store/
packages/
  protocol/               # schema + generated types
infra/
  livekit/
  coturn/                  # 自建阶段再启用
  docker/
docs/
```

这样语言分工清楚：

* Dart/Flutter 负责移动端 UI
* Go 负责 control plane、远端 daemon、runtime backend 和本机系统能力
* LiveKit Cloud 在初期负责 transport PaaS
* coturn 只在自建 transport 阶段引入

### 5.7 可替代方案

如果团队更偏移动原生：

* App 可改为 SwiftUI first，再补 Android/Kotlin。
* 优点是 iOS 体验和 WebRTC 原生控制更强。
* 缺点是双端成本高，协议和 UI 复用弱。

如果团队后续引入成熟前端/React 经验：

* App 可重新评估 React Native + TypeScript。
* 优点是 Web 控制台和前端生态复用强。
* 缺点是对当前 Java/Go 背景不如 Flutter 直接。

V1.1 综合建议为 Flutter + Dart、Go control plane、Go daemon、LiveKit Cloud transport。

---

## 6. 核心对象模型

### 6.1 Session

```json
{
  "session_id": "sess_01H...",
  "provider_type": "codex",
  "backend_type": "tmux_cli",
  "title": "Fix auth redirect bug",
  "repo_name": "lightAC",
  "working_directory": "/Users/gongmeng/dev/code/lightAC",
  "branch_name": "main",
  "status": "running",
  "is_active": true,
  "has_attention": false,
  "last_active_at": "2026-04-10T08:12:00Z",
  "created_at": "2026-04-10T08:00:00Z"
}
```

### 6.2 RunState

```json
{
  "session_id": "sess_01H...",
  "run_state": "running",
  "reason_code": "running_tests",
  "short_status_text": "正在运行测试",
  "current_step_summary": "验证登录重定向修复是否通过回归测试",
  "requires_attention": false,
  "last_event_at": "2026-04-10T08:12:00Z"
}
```

外部五态：

* `running`
* `waiting`
* `paused`
* `completed`
* `error`

内部 reason code：

* `waiting_for_instruction`
* `waiting_for_confirmation`
* `waiting_for_permission`
* `running_command`
* `editing_files`
* `running_tests`
* `summarizing_result`
* `recovering_session`
* `provider_error`

### 6.3 AgentEvent

```json
{
  "event_id": "evt_01H...",
  "session_id": "sess_01H...",
  "cursor": "00000129",
  "event_type": "test.finished",
  "title": "测试完成",
  "body": "12 passed, 0 failed",
  "severity": "success",
  "source": "runtime_backend",
  "confidence": "high",
  "created_at": "2026-04-10T08:12:00Z",
  "metadata": {
    "command": "npm test",
    "exit_code": 0
  }
}
```

`source` 用于表示事件来源：

* `provider_native`
* `runtime_backend`
* `filesystem_watcher`
* `git_inspector`
* `cli_parser`
* `terminal_peek`
* `daemon_synthetic`

`confidence` 用于表示事件可信度：

* `high`：进程状态、退出码、provider 原生事件、文件系统事件等高置信来源
* `medium`：由多种信号推断出的事件
* `low`：主要来自 CLI 文本解析或启发式判断

V1 UI 应优先展示高置信事件。低置信事件可以弱化展示，或合并进摘要，避免误导用户。

V1 事件类型：

* `session.created`
* `session.snapshot`
* `run_state.changed`
* `agent.thinking`
* `tool.started`
* `tool.finished`
* `file.changed`
* `command.started`
* `command.finished`
* `test.started`
* `test.finished`
* `attention.requested`
* `confirmation.requested`
* `message.received`
* `summary.generated`
* `terminal.peek`
* `error.raised`

### 6.4 Attention

```json
{
  "attention_id": "att_01H...",
  "session_id": "sess_01H...",
  "type": "confirmation",
  "title": "需要确认",
  "body": "是否继续修改测试文件？",
  "actions": [
    {
      "action_id": "continue",
      "label": "继续"
    },
    {
      "action_id": "stop",
      "label": "先停下"
    }
  ],
  "created_at": "2026-04-10T08:12:00Z",
  "expires_at": null,
  "resolved_at": null,
  "resolved_by_device_id": null
}
```

V1 中 attention / confirmation 必须是一次性状态机：

```text
open -> resolved
open -> expired
open -> cancelled
```

多个设备同时响应时，只有第一个有效响应能进入 `resolved`；后续响应返回 `already_resolved`。

### 6.5 BackendCapabilities

```json
{
  "backend_type": "provider_server",
  "capabilities": {
    "persistent_session": true,
    "structured_events": true,
    "message_injection": true,
    "confirmation_response": true,
    "terminal_peek": false,
    "file_events": true,
    "test_events": true,
    "server_side_notifications": true,
    "provider_native_resume": true
  }
}
```

---

## 7. Agent Work Protocol

### 7.1 Envelope

所有 DataChannel 业务消息都使用统一 envelope。

```json
{
  "version": "1",
  "type": "event.appended",
  "id": "msg_01H...",
  "request_id": "req_01H...",
  "session_id": "sess_01H...",
  "cursor": "00000129",
  "created_at": "2026-04-10T08:12:00Z",
  "payload": {}
}
```

字段说明：

* `version`：协议版本
* `type`：消息类型
* `id`：消息 ID
* `request_id`：请求/响应关联 ID
* `session_id`：会话 ID，部分全局消息可为空
* `cursor`：事件游标，用于断线恢复
* `created_at`：消息创建时间
* `payload`：业务负载

### 7.2 消息类型

连接与能力：

* `hello`
* `hello.ack`
* `capabilities.updated`
* `heartbeat`
* `error`

Session：

* `session.list`
* `session.list.result`
* `session.create`
* `session.created`
* `session.attach`
* `session.snapshot`
* `session.stop`

事件流：

* `event.subscribe`
* `event.appended`
* `event.batch`
* `run_state.changed`

用户控制：

* `message.send`
* `message.ack`
* `confirmation.respond`
* `attention.requested`
* `attention.resolved`

终端辅助：

* `terminal.peek`
* `terminal.peek.result`

断线恢复：

* `reconnect.resume`
* `reconnect.resume.result`

### 7.3 协议版本协商

Envelope 中的 `version` 表示当前消息使用的协议主版本。V1 使用 `"1"`。

`hello` 阶段必须完成版本和能力协商：

```text
1. App 发送支持的 protocol_versions 和 client capabilities
2. daemon 返回选中的 protocol_version 和 daemon/backend capabilities
3. 双方只使用协商后的协议版本和 capability 集合
4. 无共同协议版本时连接失败，返回 protocol_version_unsupported
```

版本规则：

* patch/minor 级新增字段必须向后兼容
* 消息新增字段默认可忽略
* 删除字段、改变字段语义、改变必填性视为 major 变化
* capability 控制可选功能，不通过版本号硬编码功能开关

App 版本高于 daemon：

* App 必须按 daemon 返回的 capability 降级
* 不展示 daemon 不支持的 UI 操作
* 发送不支持的消息时 daemon 返回 `capability_not_supported`

App 版本低于 daemon：

* daemon 必须保留当前 major version 的兼容处理
* daemon 不应向旧 App 主动推送其无法解析的必需消息
* 新功能通过 capability 暴露，旧 App 可忽略

建议新增错误码：

* `protocol_version_unsupported`
* `message_type_unsupported`
* `required_capability_missing`

### 7.4 示例：hello

```json
{
  "version": "1",
  "type": "hello",
  "id": "msg_hello_1",
  "request_id": "req_hello_1",
  "created_at": "2026-04-10T08:12:00Z",
  "payload": {
    "client_type": "mobile_app",
    "device_id": "dev_phone_1",
    "protocol_versions": ["1"],
    "supported_channels": ["control", "events", "terminal", "heartbeat"],
    "last_known_cursor": {
      "sess_01H...": "00000120"
    }
  }
}
```

### 7.5 示例：session.attach

```json
{
  "version": "1",
  "type": "session.attach",
  "id": "msg_attach_1",
  "request_id": "req_attach_1",
  "session_id": "sess_01H...",
  "created_at": "2026-04-10T08:12:00Z",
  "payload": {
    "from_cursor": "00000120",
    "include_snapshot": true
  }
}
```

### 7.6 示例：message.send

```json
{
  "version": "1",
  "type": "message.send",
  "id": "msg_user_1",
  "request_id": "req_user_1",
  "session_id": "sess_01H...",
  "created_at": "2026-04-10T08:12:00Z",
  "payload": {
    "content": "这个方向可以，继续。先别重构，优先修 bug。",
    "input_mode": "free_text"
  }
}
```

### 7.7 示例：confirmation.respond

```json
{
  "version": "1",
  "type": "confirmation.respond",
  "id": "msg_confirm_1",
  "request_id": "req_confirm_1",
  "session_id": "sess_01H...",
  "created_at": "2026-04-10T08:12:00Z",
  "payload": {
    "attention_id": "att_01H...",
    "action_id": "continue",
    "comment": "继续，但不要扩大改动范围。"
  }
}
```

### 7.8 幂等与并发约束

所有 App -> daemon 的写操作必须带 `request_id`。daemon 需要记录最近处理过的 request，保证重试不会重复执行。

需要幂等处理的消息：

* `session.create`
* `message.send`
* `confirmation.respond`
* `session.stop`
* `terminal.peek`

V1 控制权规则：

* 每个 session 同一时间只有一个 active controller
* 其他设备可以 read-only observe
* active controller 断开或超时后，其他设备可以 acquire control
* `confirmation.respond` 只能成功一次
* 已完成的 confirmation 返回 `already_resolved`

建议错误码：

* `already_resolved`
* `not_active_controller`
* `request_replayed`
* `cursor_expired`
* `capability_not_supported`
* `workspace_not_allowed`

---

## 8. DataChannel 设计

建议建立多个逻辑 channel。

| Channel | 可靠性 | 顺序 | 用途 |
|---|---|---|---|
| `control` | reliable | ordered | 请求/响应、用户动作、确认 |
| `events` | reliable | ordered | AgentEvent 和 RunState |
| `terminal` | reliable | ordered | 小窗口 terminal peek |
| `heartbeat` | unreliable 或 reliable | unordered 或 ordered | 连接健康检查 |

设计约束：

* `control` 不能被 terminal peek 阻塞
* `events` 必须有 cursor
* `terminal` 必须限流和限大小
* 大型日志和文件内容不通过 DataChannel 直接传
* DataChannel 断开后以 `reconnect.resume` 恢复业务状态
* App 后台不维持 DataChannel 长连接
* App 前台只订阅当前 session 的实时事件，其他 session 只接收摘要和 attention 状态
* 高频事件需要 batch 或 debounce 后刷新 UI，避免移动端性能和电量损耗

---

## 9. Remote Daemon / Session Gateway 设计

### 9.1 模块划分

```text
Daemon Core
  - Auth Manager
  - Device Binding Manager
  - DataChannel Gateway
  - Session Manager
  - Provider Adapter Manager
  - Runtime Backend Manager
  - Event Normalizer
  - Event Store
  - Attention Detector
  - Notification Dispatcher
  - Local Config Store
```

### 9.2 Session Manager

职责：

* 创建 session
* 查询 session
* attach/detach session
* 更新 session metadata
* 维护 active session
* 维护 session 与 provider/backend 的关系

### 9.3 Event Normalizer

职责：

* 将 backend/provider 原始输出转成 AgentEvent
* 生成短摘要
* 更新 RunState
* 提取 attention / confirmation
* 降级处理不可识别输出

降级策略：

* 能识别则生成结构化事件
* 不确定则生成 `agent.thinking` 或 `summary.generated`
* 原始文本只进入 terminal peek 或低优先级 log metadata

### 9.4 Event Store

V1 可以使用本地轻量存储。

建议数据：

* session metadata
* run state
* agent event
* message log
* attention state
* device binding
* backend capabilities
* cursor index

保留策略：

* 默认保留最近 50 个 session
* 每个 session 默认最多保留 10000 条 AgentEvent
* 全局事件上限默认 200000 条
* terminal peek 不长期保留，默认只保留最近一次短窗口
* 单条 terminal peek 默认最多 200 行或 64 KB
* 超过保留上限后按 session last_active_at 和 event cursor FIFO 淘汰
* 用户可清理本地事件记录

量级预估：

* 普通 30 分钟 coding session 预计 300-3000 条 AgentEvent
* 高频工具/测试循环 session 可能达到 5000-10000 条 AgentEvent
* V1 不保存完整终端流，因此 SQLite 压力主要来自结构化事件写入

SQLite 策略：

* 使用 WAL mode
* event insert 走批量事务
* cursor/session_id 建索引
* 定期 vacuum 或增量 vacuum
* 磁盘占用超过阈值时触发自动清理

默认磁盘阈值：

* warning：本地 event store 超过 500 MB
* cleanup：本地 event store 超过 1 GB
* hard cap：本地 event store 超过 2 GB 时停止保留低优先级事件，只保留 attention、run_state、message 和 summary

### 9.5 Attention Detector

职责：

* 判断 agent 是否需要用户介入
* 判断是否需要推送
* 去重和抑制噪声通知

V1 触发：

* confirmation requested
* waiting for instruction
* completed
* failed

不通知：

* 普通工具调用
* 高频状态变化
* 无需用户动作的中间日志

### 9.6 Workspace / Repo Discovery

workspace/repo 列表来自 daemon 本地 workspace allowlist。

规则：

* 用户先在 daemon 本地配置 allowlist root
* daemon 只扫描 allowlist root 下的一级或二级目录
* 识别规则优先 `.git`、常见项目文件和最近 session 记录
* 大目录扫描必须异步和可取消
* App 请求列表时分页返回
* 手动输入路径必须 canonicalize 后校验是否落在 allowlist 内
* 禁止通过软链接跳出 allowlist

建议分页：

```json
{
  "items": [],
  "next_cursor": "repo_page_2",
  "has_more": true
}
```

V1 可以先支持：

* 最近使用目录
* allowlist 下 repo 列表
* 手动输入路径并校验

---

## 10. Session Runtime Backend 设计

### 10.1 统一接口

```text
SessionRuntimeBackend
  createSession(input) -> SessionHandle
  attachSession(session_id) -> SessionHandle
  resumeSession(session_id, cursor) -> EventBatch
  sendMessage(session_id, message) -> Ack
  respondConfirmation(session_id, response) -> Ack
  getSnapshot(session_id) -> SessionSnapshot
  streamEvents(session_id, cursor) -> EventStream
  peekTerminal(session_id, options) -> TerminalPeek
  stopSession(session_id) -> Ack
  getCapabilities() -> BackendCapabilities
```

### 10.2 TmuxCliBackend

适用：

* 当前 Codex CLI
* Claude Code CLI
* 其他 terminal-based coding agent

职责：

* 创建 tmux session
* 启动 CLI agent
* 注入用户输入
* 读取 pty/tmux 输出
* 维护进程状态
* 支持 attach/detach/recover
* 提供 terminal peek

限制：

* 结构化事件依赖 daemon 转译
* provider 输出格式变化会影响解析
* terminal peek 需要严格限流

### 10.3 PtyCliBackend

适用：

* 不想依赖 tmux 的轻量运行场景
* daemon 自己管理 pty 生命周期

差异：

* 持久化能力弱于 tmux
* 更容易嵌入 daemon
* 需要 daemon 自己处理进程恢复

### 10.4 ProviderServerBackend

适用：

* provider 提供原生 server/session API
* provider 自带事件流
* provider 自带会话恢复

职责：

* 调用 provider session API
* 订阅 provider event stream
* 转发 message / confirmation
* 映射 provider-native resume

特点：

* 可能不支持 terminal peek
* 结构化事件更可靠
* 持久化由 provider 承担

### 10.5 CloudAgentBackend

适用：

* agent 运行在云端 task/run/thread
* 手机端只是观察和控制云端任务

特点：

* 本地 daemon 可以退化为 gateway
* session 持久化由云端承担
* 需要更严格的权限和审计

---

## 11. Provider Adapter 设计

### 11.1 CodexAdapter V1 职责

* 根据 runtime backend 启动 Codex
* 将 Codex 输出转成 AgentEvent
* 识别 waiting / confirmation / completed / error
* 将用户消息注入 Codex
* 生成 session snapshot
* 暴露 Codex adapter capabilities

### 11.2 转译分层

强结构事件：

* run state changed
* confirmation requested
* completed
* error

弱结构事件：

* current step summary
* command started/finished
* file changed
* tests started/finished

兜底事件：

* terminal peek
* raw snippet metadata
* generic summary

### 11.3 Provider 差异处理

Adapter 对上只暴露统一协议，对下可以适配不同 provider：

```text
CodexAdapter
  -> TmuxCliBackend
  -> ProviderServerBackend

ClaudeCodeAdapter
  -> TmuxCliBackend
  -> ProviderServerBackend

CustomAgentAdapter
  -> CustomRunnerBackend
```

### 11.4 Claude Code 纸面适配验证

V1 只实现 CodexAdapter，但 Agent Work Protocol 不能过拟合 Codex。设计冻结前，需要用 Claude Code 的典型工作流做纸面适配验证。

需要验证的 Claude Code 场景：

* 工具调用
* 文件读取
* 文件编辑
* shell command
* 测试执行
* permission request
* 用户确认
* 错误恢复
* 任务总结

映射要求：

| Claude Code 行为 | AgentEvent / RunState 映射 |
|---|---|
| tool use start/end | `tool.started` / `tool.finished` |
| file read/edit | `file.changed` 或 `tool.finished` metadata |
| shell command | `command.started` / `command.finished` |
| tests | `test.started` / `test.finished` |
| permission request | `attention.requested` + `confirmation.requested` |
| waiting for user | `run_state.changed` + `waiting_for_instruction` |
| task done | `run_state.changed` + `completed` |
| error | `error.raised` + `provider_error` |
| summary | `summary.generated` |

验收：

* 如果 Claude Code 的关键语义无法映射到现有 AgentEvent / RunState，需要在 V1 协议冻结前调整协议。
* 如果只是 provider 特有 metadata，可放入 `metadata`，不新增顶层事件类型。
* 不为了某个 provider 的 UI 细节污染通用协议。

---

## 12. 关键流程

### 12.1 设备配对

```text
1. daemon 启动，生成 daemon_id 和 pairing token
2. 用户在手机 App 登录
3. App 扫码或输入 pairing code
4. signaling service 绑定 user_id、device_id、daemon_id
5. daemon 与 App 完成连接授权
6. 后续 App 可发现该 daemon
```

### 12.2 建立 WebRTC 连接

```text
1. App 从 signaling service 获取 daemon 在线状态
2. App 发起 connect request
3. signaling service 转发 offer / answer / ICE candidate
4. App 与 daemon 建立 DataChannel
5. 双方发送 hello / hello.ack
6. daemon 返回 capabilities 和 session list
```

### 12.3 新建 Session

```text
1. App 请求 daemon 返回 workspace/repo 列表
2. 用户选择目录、分支、首条消息
3. App 发送 session.create
4. daemon 选择 provider adapter 和 runtime backend
5. backend 创建 session
6. adapter 启动 Codex
7. daemon 写入 session metadata
8. daemon 返回 session.created 和 session.snapshot
9. events channel 开始推送 AgentEvent
```

### 12.4 观察 Session

```text
1. App 发送 session.attach(session_id, from_cursor)
2. daemon 返回 session.snapshot
3. daemon 补发 from_cursor 之后的 event.batch
4. daemon 推送实时 event.appended
5. App 更新信息流和状态
```

### 12.5 用户接力

```text
1. daemon 识别 attention.requested
2. daemon 写入 Event Store
3. daemon 通过 signaling/push 通知手机
4. 用户打开 App 进入 session
5. 用户发送 message.send 或 confirmation.respond
6. daemon 调用 provider adapter
7. adapter 调用 runtime backend 注入输入或 API 调用
8. agent 继续运行
9. daemon 推送 run_state.changed
```

### 12.6 断线恢复

```text
1. App 本地保存每个 session 的 last_cursor
2. DataChannel 断开
3. App 重连后发送 reconnect.resume
4. daemon 查询 Event Store
5. daemon 返回缺失 event.batch
6. App 恢复信息流
7. App 重新进入实时订阅
```

---

## 13. 移动端信息架构

### 13.1 首页

展示：

* 主活动 session
* run state
* current step summary
* repo / branch
* has attention
* 少量其他 session

操作：

* 进入 session
* 切换 session
* 新建 session

### 13.2 Session 详情

页面结构：

```text
顶部：session title / run state / repo / branch
中部：AgentEvent 工作信息流
底部：message input / confirmation card
辅助：terminal peek
```

卡片类型：

* 状态卡
* 步骤卡
* 工具卡
* 文件卡
* 命令卡
* 测试卡
* 确认卡
* 摘要卡
* 错误卡
* terminal peek 卡

### 13.3 Terminal Peek

Terminal peek 是辅助能力。

约束：

* 默认折叠
* 只显示短窗口
* 用户主动打开
* 不持续滚屏
* 不作为主要输入界面

---

## 14. 安全设计

### 14.1 身份和绑定

* 用户登录后才能绑定 daemon
* daemon 有独立 daemon_id
* 手机设备有独立 device_id
* 支持设备撤销
* 支持 daemon 撤销
* pairing token 短期有效，且只能使用一次
* daemon 侧保留本地 trusted device 列表

### 14.2 连接安全

* WebRTC 使用 DTLS 加密
* signaling service 需要鉴权
* pairing token 短期有效
* TURN credential 短期有效
* DataChannel hello 阶段校验协议和设备身份

### 14.3 操作安全

高风险操作需要确认：

* stop session
* 新建 session
* 切换分支
* 执行危险命令
* 扩大目录权限
* 读取 terminal peek

V1 必须启用 workspace allowlist：

* daemon 只能在 allowlist 目录下创建 session
* App 只能看到 allowlist 内的 workspace/repo
* 手动输入路径必须通过 allowlist 校验
* 禁止通过 `..`、软链接或环境变量绕过 allowlist
* 不允许 App 任意请求读取文件内容

daemon 运行权限：

* 默认以普通用户权限运行
* 不要求 root
* 不默认开放公网监听端口
* 本地管理端口默认绑定 localhost
* 远程访问必须经过 device binding 和 DataChannel 鉴权

terminal peek 约束：

* 默认折叠，不自动打开
* 按 session、设备、时间窗口限流
* 限制最大字节数和最大行数
* 记录审计
* 后续可增加敏感信息遮蔽

### 14.4 审计

daemon 记录：

* 设备连接
* session 创建
* message.send
* confirmation.respond
* stop session
* terminal.peek
* notification dispatch
* acquire/release active controller
* workspace allowlist 变更

### 14.5 Push Payload 安全

Push notification payload 经过 APNs/FCM 等第三方基础设施，不应包含敏感动态内容。

V1 push payload 白名单：

* `session_id`
* `notification_id`
* `notification_type`
* `attention_type`
* `created_at`

不允许进入 push payload：

* 终端输出
* 文件路径全文
* 分支名
* repo 绝对路径
* 错误详情
* agent 生成的动态摘要
* 用户输入原文
* secret/token/API key 相关文本

展示策略：

* 系统通知文案使用固定模板，例如“有一个 Agent 任务需要确认”
* 用户点击通知打开 App 后，再通过 DataChannel 或安全 API 获取详情
* 如果必须展示 session title，V1 应默认使用用户手动设置的 title 或脱敏 title

### 14.6 Daemon 运维设计

daemon 是长期运行的 agent-side server，需要具备基础运维能力。

安装：

* macOS 使用 launchd user agent
* Linux 使用 systemd user service
* 默认以普通用户运行
* 默认不监听公网端口

自愈：

* launchd/systemd 负责异常退出后自动重启
* daemon 启动后重新加载 local config、event store 和 trusted devices
* daemon 重启后重新发现 existing tmux sessions
* daemon 崩溃不得杀掉 CLI agent session

升级：

* V1 采用优雅重启，不做热升级
* 升级前记录 active sessions
* 升级后重新 attach existing sessions
* schema migration 必须向后兼容或可回滚
* 与 App 协议不兼容时通过 hello/version negotiation 降级或拒绝连接

日志：

* 本地结构化日志
* 日志分级：debug / info / warn / error
* 默认不记录终端原文和用户输入全文
* 日志滚动和大小限制

远程诊断：

* 用户显式授权后生成诊断包
* 诊断包默认只包含版本、配置摘要、连接状态、错误码、最近非敏感日志
* terminal peek、用户输入、完整路径默认不进入诊断包

---

## 15. 可靠性设计

### 15.1 Cursor 机制

每个 AgentEvent 都有单调递增 cursor。

App 保存：

* per-session last_cursor
* last snapshot time
* reconnect attempts

daemon 支持：

* event.batch
* reconnect.resume
* cursor not found fallback

cursor 规则：

* cursor 在单个 session 内单调递增
* Event Store 写入成功后才能对 App 推送
* App ack 只表示客户端收到，不表示用户已读
* cursor 过期时返回最新 snapshot 和最近事件窗口
* confirmation 状态以 daemon/event store 为准，不以 App 本地状态为准

### 15.2 心跳

heartbeat 用于：

* 检测连接存活
* 显示 daemon 在线状态
* 触发重连
* 判断延迟

心跳策略：

* App 前台使用较短心跳间隔
* App 后台停止或显著降低心跳
* heartbeat 失败不立即判定 session 失败，只判定连接失效
* session 真实状态以 daemon reconnect 后 snapshot 为准

### 15.3 TURN fallback

TURN 只作为直连失败兜底。

策略：

* 优先直连
* 失败后 fallback TURN
* 记录连接类型
* 用于诊断网络质量

### 15.4 Push 与实时连接分工

push 负责唤醒用户：

* attention requested
* completed
* failed

DataChannel 负责 App 打开后的实时控制：

* event stream
* message.send
* confirmation.respond
* terminal.peek

后台策略：

* App 后台不承诺 WebRTC 常驻
* push notification 是后台唤醒主路径
* 用户打开 App 后通过 cursor 补齐离线事件
* notification payload 只包含 session_id、通知类型和短摘要，不携带敏感终端内容

### 15.5 多设备状态一致性

V1 采用单用户、多设备读、单设备写模型。

一致性规则：

* `Session`、`RunState`、`Attention` 以 daemon/event store 为权威状态
* 每个写操作必须带 `request_id`
* daemon 对 `request_id` 做去重
* active controller 有租约，断线或超时后释放
* read-only 设备可以接收事件，但不能发送 `message.send` 或 `confirmation.respond`
* 设备切换控制权需要显式 acquire control

Active controller 租约：

```text
idle -> acquiring -> active -> renewing -> released
active -> expired
active -> revoked
```

默认参数：

* lease duration：60 秒
* renew interval：20 秒
* grace period：15 秒
* acquire timeout：5 秒

续约规则：

* active controller 前台连接时通过 heartbeat 自动续约
* App 进入后台时主动 release control
* 网络抖动时 grace period 内不立即释放
* 超过 grace period 未续约则进入 expired
* 其他设备不会自动抢占，必须显式 acquire control
* 用户手动切换设备时，新设备 acquire control，旧设备收到 revoked

冲突处理：

* 同一时间只有一个 active controller
* acquire control 以 daemon/event store 原子更新为准
* 已过期租约不能发送写操作
* read-only 设备发写操作返回 `not_active_controller`

---

## 16. 可观测性设计

系统有 Mobile App、Cloud Control Plane、Remote Daemon 三个运行端，必须能定位问题发生在哪一段。

### 16.1 核心指标

连接指标：

* WebRTC connection success rate
* direct connection rate
* TURN fallback rate
* reconnect count per session
* reconnect resume success rate
* average reconnect duration

事件链路指标：

* agent event generated -> daemon stored latency
* daemon stored -> App displayed latency
* event store write latency
* event batch size
* cursor resume missing event count

daemon 指标：

* daemon uptime
* daemon restart count
* active session count
* event store size
* event normalizer error rate
* provider adapter error rate
* runtime backend error rate

移动端体验指标：

* notification tap -> session visible duration
* session visible -> first action duration
* confirmation success rate
* message.send ack latency
* terminal peek open rate
* sessions resolved without terminal peek rate

### 16.2 V1 SLI / SLO 建议

V1 可先使用内部 dogfood SLO：

* WebRTC 前台连接成功率 >= 90%
* reconnect.resume 成功率 >= 95%
* notification tap -> session visible P95 < 3 秒
* message.send ack P95 < 1 秒
* daemon event store write P95 < 50 ms
* sessions resolved without terminal peek rate >= 60%
* high-confidence AgentEvent coverage >= 60%

这些不是对外承诺，而是判断 V1 是否具备继续产品化价值的内部门槛。

### 16.3 日志和 Trace

每次 session 需要统一 trace id：

* `session_id`
* `connection_id`
* `device_id`
* `daemon_id`
* `request_id`

日志原则：

* 默认不记录敏感正文
* 错误码结构化
* 关键状态变化可追踪
* 诊断包需用户显式授权

---

## 17. V1 落地范围

### 17.1 Mobile App

必做：

* 登录
* 设备配对
* daemon 列表
* session 列表
* session 详情
* AgentEvent 信息流
* message.send
* confirmation.respond
* push 打开指定 session
* reconnect.resume
* 前台 DataChannel 实时连接
* 后台 push-only 策略
* 当前 session 事件订阅
* 高频事件 UI batch/debounce

可选：

* terminal.peek
* 最近目录收藏

### 17.2 Signaling Service

必做：

* 用户鉴权
* daemon registry
* device registry
* pairing
* offer / answer / ICE exchange
* TURN credential 下发
* push notification routing

### 17.3 Daemon

必做：

* local config
* device binding
* DataChannel gateway
* session manager
* CodexAdapter
* TmuxCliBackend 或 PtyCliBackend
* event normalizer
* event store
* attention detector
* reconnect cursor
* request_id 幂等去重
* active controller 租约
* workspace allowlist
* terminal peek 限流和审计

### 17.4 暂不做

* 多 provider
* 多人协作
* 多设备同时写
* 完整 diff review
* 手机代码编辑器
* 完整终端
* 大文件传输
* 自动任务规划
* 云端托管 agent
* 后台 WebRTC 长连接

---

## 18. 里程碑建议

### M0：CLI 可解析性与协议适配探测

目标：

* 采样 Codex CLI 在典型任务中的输出
* 识别可高置信解析的事件类型
* 估算 high-confidence AgentEvent coverage
* 验证 message injection、waiting、completed、error 的识别方式
* 进行 Claude Code 纸面适配验证
* 本地局域网 transport 跑通最小 Agent Work Protocol

验收：

* 产出 Codex CLI 事件覆盖率报告
* 明确 high / medium / low confidence 事件清单
* high-confidence AgentEvent coverage >= 60%，否则暂停 M2 并重新评估 provider-native structured output/API 路线
* 产出 Claude Code 典型工作流到 AgentEvent / RunState 的映射表
* 明确 V1 CodexAdapter 的可承诺事件范围
* 本地 WebSocket 或局域网 WebRTC 能完成 `hello` / `heartbeat` / `event.appended`

### M1：协议和 daemon 骨架

目标：

* daemon 可启动
* App/测试客户端可连接 daemon
* hello / session.list / heartbeat 跑通
* 本地 event store 可写入和读取
* hello 阶段完成协议版本和 capability negotiation
* LiveKit Cloud room/token spike 跑通

验收：

* 本地模拟客户端能建立 DataChannel
* daemon 能返回 mock session
* cursor 机制可补拉 mock event
* App 版本高于/低于 daemon 时能按 capability 降级或明确失败
* Flutter 或测试客户端可通过 LiveKit data API 与 daemon 交换 envelope

### M2：Codex CLI session

目标：

* TmuxCliBackend 或 PtyCliBackend 可创建 Codex session
* 可注入用户消息
* 可读取输出并生成基础 AgentEvent

验收：

* 手机或测试客户端能新建 session
* 能看到 running / waiting / completed / error
* 能发送一句话让 session 继续

### M3：移动端信息流

目标：

* session 列表
* session 详情
* AgentEvent 卡片
* confirmation card
* push notification

验收：

* 用户收到 attention 通知后可进入对应 session
* 用户能完成一次 confirmation/respond
* 用户不需要看终端也能理解当前状态

### M4：断线恢复和可靠性

目标：

* reconnect.resume
* event.batch
* heartbeat
* TURN fallback
* terminal.peek 限流
* active controller 租约
* push payload 白名单
* daemon 自愈和重启恢复
* 基础可观测性指标

验收：

* 切换网络后能恢复 session 信息流
* 不丢失 attention
* terminal peek 不阻塞 control 消息
* daemon 重启后能重新发现 existing tmux sessions
* push payload 不包含动态敏感内容
* 能看到连接成功率、TURN fallback 率、reconnect.resume 成功率和 message.send ack latency

---

## 19. 关键风险

### 19.1 技术复杂度与部署成本过高

判断：

该风险真实存在。B+ 同时包含 Mobile App、Cloud Control Plane、Remote Daemon、TURN、runtime backend，明显比 SSH/Mosh 方案复杂。

但 SSH/Mosh 只能验证“手机远程进终端”，不能验证“结构化 agent 工作信息流”。因此不建议回退到 SSH 主路径，而是收窄 V1.1。

缓解：

* V1.1 只支持 Codex
* V1.1 只支持单用户
* V1.1 默认只实现 tmux/pty CLI backend
* provider/server/cloud backend 只保留接口，不实现
* TURN 使用 coturn，不自研
* control plane 只做账号、配对、signaling、push、TURN credential
* terminal peek 作为辅助能力，不做完整终端
* 以端到端 vertical slice 为第一优先级

### 19.2 Codex CLI 输出解析不稳定

判断：

这是 V1 最大产品风险之一。Codex CLI 输出格式可能变化，不同版本和不同终端环境也可能影响解析质量。如果事件流主要依赖文本解析，用户可能重新依赖 terminal peek。

缓解：

* V1 只承诺高置信事件稳定
* AgentEvent 必须带 `source` 和 `confidence`
* 优先使用进程状态、退出码、文件系统 watcher、git 状态等高置信信号
* CLI parser 生成的事件默认不标记为 high confidence
* 弱结构事件可降级为 summary/log
* terminal peek 兜底
* 后续优先接 provider 原生事件 API

### 19.3 WebRTC 移动网络复杂

判断：

该风险真实存在，尤其是 4G/5G 切换、复杂 NAT、企业防火墙、系统后台限制。B+ 不能假设手机后台维持稳定 WebRTC 长连接。

缓解：

* signaling 和 TURN fallback 前置
* heartbeat + reconnect.resume
* 不依赖长连接保存唯一状态
* 关键事件写入 daemon event store
* App 前台实时，后台 push-only
* reconnect 后以 snapshot + event.batch 恢复状态
* 记录连接类型和失败原因，用于网络诊断

### 19.4 Daemon 职责过重

判断：

该风险真实存在。daemon 同时承担连接、session、事件、通知、adapter、backend 调用，如果没有边界会变成难以调试的大泥球。

缓解：

* daemon 内部按模块拆分
* provider adapter 和 runtime backend 必须是接口边界
* daemon 崩溃不能杀掉 tmux session
* daemon 重启后能重新发现 existing sessions
* Event Store 持久化 session metadata、cursor、attention
* daemon 不提供任意 shell API
* daemon 不做二次 agent 和智能规划

### 19.5 daemon 权限过大

缓解：

* device binding
* 操作审计
* 高风险操作确认
* workspace allowlist
* terminal peek 限制
* daemon 普通用户权限运行
* 本地端口默认 localhost
* signaling 不保存 agent 工作内容
* pairing token 短期单次有效

### 19.6 移动端电池与性能

判断：

该风险真实存在。WebRTC 长连接、高频事件流和后台网络都会影响电池和稳定性。

缓解：

* App 后台不维持实时 DataChannel
* 后台依赖 push notification
* 只订阅当前 session 实时事件
* 非当前 session 只展示摘要状态
* 高频事件 batch/debounce 后刷新 UI
* terminal peek 手动触发且限流
* heartbeat 前台高频、后台降频或停止

### 19.7 同步与状态一致性

判断：

该风险真实存在，但 V1 可以通过产品约束避免复杂协同。

缓解：

* 单用户模型
* 多设备可读
* 单 session 单 active controller
* 写操作带 `request_id` 幂等
* confirmation 一次性 resolve
* cursor per-session 单调递增
* daemon/event store 是权威状态源

### 19.8 产品退化成终端

缓解：

* terminal peek 默认折叠
* App 主界面只消费 AgentEvent
* DataChannel 不传完整终端流
* 指标关注“无需看终端也能理解状态”

---

## 20. 待技术探测问题

1. Codex CLI 的启动、恢复和消息注入方式，以当前本机环境为准需要再做技术探测。
2. Codex CLI 输出中哪些状态可以高置信解析，哪些必须降级为低置信事件或 terminal peek。
3. Flutter + LiveKit data API 在目标 iOS/Android 版本上的稳定性。
4. terminal peek 在 V1 是否纳入用户可见能力，还是作为 debug-only 能力。
5. M0 本地 transport 使用临时 WebSocket 还是局域网 WebRTC/Pion 直连。
6. Claude Code 纸面适配验证是否发现当前 AgentEvent / RunState 缺失关键语义。

---

## 21. 最终结论

B+ 方案的长期价值在于：产品主体验绑定的是 Agent Work Protocol，而不是某个 provider、某种终端输出或某种 runtime。

V1 用 tmux/pty CLI backend 是为了快速托管当前 CLI 形态的 Codex；架构上真正稳定的是：

```text
Mobile App
  -> Agent Work Protocol
  -> Remote Daemon / Session Gateway
  -> Provider Adapter
  -> Session Runtime Backend
```

这使产品可以同时覆盖现在的 CLI agent 和未来的 server-native / cloud-native agent。
