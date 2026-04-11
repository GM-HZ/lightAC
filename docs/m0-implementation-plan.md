# M0 技术探测实施计划

## 1. 目标

M0 不是完整产品开发，而是验证 B+ 架构的地基是否成立。

M0 必须回答 4 个问题：

1. Codex CLI 是否能被稳定托管、输入和恢复。
2. Codex CLI 输出能否产生足够多 high-confidence AgentEvent。
3. Agent Work Protocol 是否能在本地 transport 上跑通。
4. Flutter + LiveKit 是否适合作为后续实时 transport。

---

## 2. 默认决策

M0 默认采用以下决策，避免继续空转：

| 项 | M0 决策 |
|---|---|
| transport | 本地 WebSocket |
| runtime backend | TmuxCliBackend |
| OS | macOS first |
| client | 最小 CLI/Web 测试客户端 |
| cloud control plane | 不做 |
| Flutter + LiveKit | M0 后半段 spike |
| storage | SQLite |
| daemon language | Go |

说明：

* WebSocket 只是 M0 本地验证 transport，不代表生产 transport。
* TmuxCliBackend 是 V1 默认方向，因为 daemon 崩溃后 tmux session 仍可保留。
* M0 先不做账号、配对、push、TURN、LiveKit room token。

---

## 3. 建议目录结构

M0 先创建最小目录，不铺开完整 monorepo。

```text
daemon/
  cmd/lightac-daemon/
  internal/protocol/
  internal/store/
  internal/runtime/tmux/
  internal/session/
  internal/transport/ws/
  internal/normalizer/
tools/
  m0-client/
docs/
  m0-implementation-plan.md
  m0-codex-cli-coverage-template.md
  m0-codex-cli-coverage-report.md
```

后续进入 M1/M2 再增加：

```text
apps/mobile/
apps/control-plane/
packages/protocol/
infra/
```

---

## 4. M0-A：Go Daemon Skeleton

### 目标

建立 daemon 最小骨架，能启动、写本地事件、通过本地 WebSocket 与测试客户端交换协议消息。

### 范围

* Go module 初始化
* daemon 启动入口
* 本地配置加载
* WebSocket server
* Agent Work Protocol envelope
* `hello`
* `heartbeat`
* SQLite event store
* mock session
* mock `event.appended`

### 验收

* `lightac-daemon` 可在 macOS 本地启动
* 测试客户端能连接 daemon
* 客户端发送 `hello`，daemon 返回 `hello.ack`
* 客户端发送 `heartbeat`，daemon 返回 heartbeat ack
* daemon 能写入并读取 mock AgentEvent
* 客户端能收到 mock `event.appended`

---

## 5. M0-B：TmuxCliBackend Spike

### 目标

验证 daemon 能通过 tmux 托管 CLI agent session。

### 范围

* 创建 tmux session
* 启动 Codex CLI 或可替代 shell command
* 注入用户输入
* 读取 tmux pane 输出
* terminal peek
* daemon 重启后重新 attach existing tmux session
* session metadata 写入 SQLite

### 验收

* daemon 能创建 tmux session
* daemon 能向 tmux session 注入输入
* daemon 能读取 pane 输出
* daemon 停止后，tmux session 不被杀掉
* daemon 重启后能发现并重新 attach session
* terminal peek 能返回最近输出窗口

---

## 6. M0-C：Codex CLI 可解析性报告

### 目标

验证 Codex CLI 输出是否能支撑结构化 AgentEvent 信息流。

### 范围

使用典型 coding agent 任务采样：

* 简单文件修改
* bug fix
* 运行测试
* 等待用户确认
* 用户补充一句话
* 任务完成总结
* 错误/失败场景

### 输出

产出：

```text
docs/m0-codex-cli-coverage-report.md
```

报告必须包含：

* 采样任务列表
* 可识别事件清单
* high / medium / low confidence 分类
* high-confidence coverage
* 不可解析或误判案例
* 是否达到 60% 门槛
* 是否建议进入 M2

### 验收

* high-confidence AgentEvent coverage >= 60%
* 至少稳定识别：
  * running
  * waiting 或需要用户输入
  * completed
  * error
  * command finished / exit code
  * message injected
* 如果低于 60%，暂停 M2，重新评估 provider-native structured output/API 路线

---

## 7. M0-D：Minimal Protocol Client

### 目标

验证测试客户端能使用 Agent Work Protocol 完成最小 session 观察和接力。

### 范围

测试客户端可用 Go CLI、简单 Web 页面或其他最小实现。

支持：

* `hello`
* `session.list`
* `session.attach`
* `event.subscribe`
* `message.send`
* `terminal.peek`
* `reconnect.resume`

### 验收

* 客户端能列出 session
* 客户端能 attach session
* 客户端能看到事件流
* 客户端能发送一条消息到 tmux session
* 客户端断开重连后能从 cursor 补拉事件

---

## 8. M0-E：Flutter + LiveKit Spike

### 目标

验证 Flutter + LiveKit data API 是否适合作为后续实时 transport。

### 范围

* Flutter 最小 App
* Go daemon LiveKit participant
* LiveKit Cloud room
* room token 签发可先用本地脚本
* 通过 LiveKit data API 发送 Agent Work Protocol envelope
* `hello`
* `heartbeat`
* `event.appended`
* `message.send`

### 验收

* Flutter App 能加入 LiveKit room
* daemon 能加入同一 room
* 双方能交换 envelope
* App 前台能稳定收发消息
* App 后台/恢复行为有基础观察记录
* 明确 LiveKit data API 是否满足 M1/M2 需求

---

## 9. M0 完成标准

M0 完成时必须有：

* Go daemon skeleton
* 本地 WebSocket transport
* SQLite event store
* TmuxCliBackend spike
* 最小 protocol client
* Codex CLI coverage report
* Flutter + LiveKit spike 结论

M0 通过条件：

* TmuxCliBackend 可创建、输入、读取和恢复 session
* Agent Work Protocol 本地跑通
* high-confidence AgentEvent coverage >= 60%
* reconnect.resume 本地跑通
* Flutter + LiveKit data API 没有发现阻塞性问题

---

## 10. M0 不做

* 完整 Flutter App
* 完整 cloud control plane
* 账号系统
* 设备配对
* push notification
* TURN 生产验证
* 多 provider
* 多设备同时控制
* 完整终端
* 完整安全审计 UI

---

## 11. 下一步

实施顺序：

1. 创建 Go module 和 daemon skeleton。
2. 实现 protocol envelope 和本地 WebSocket transport。
3. 实现 SQLite Event Store。
4. 实现 mock session/event。
5. 实现 TmuxCliBackend spike。
6. 开始 Codex CLI 输出采样和 coverage report。
