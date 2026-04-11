# 手机端轻量 Agent 客户端 PRD（V1）

## 1. 文档信息

**文档名称**：手机端轻量 Agent 客户端 PRD（V1）
**版本**：v1.0 Draft
**状态**：可进入方案设计 / 技术设计
**目标读者**：产品、客户端、服务端、集成适配层开发

---

## 2. 背景

当前使用 Codex / Claude Code 一类 coding agent 时，桌面端适合长时间深度工作，但手机端体验明显不足，尤其是在以下场景：

* 用户不在电脑前，只能通过 SSH 对话框远程观察和输入
* 纯终端字符流把输入、状态、工具调用、确认、日志混在一起，滚屏严重，理解当前运行状态成本高
* agent 在运行中会不时进入等待人工介入的状态，用户需要尽快收到提醒并进行轻量接力
* 用户往往已经有一个较明确的大方向，但小细节仍需要间歇式推进，不需要手机端承担完整开发环境能力

Codex 桌面端、Claude Code 一类产品的交互形态，本质上已经不是传统聊天，也不是纯终端，而是面向 coding agent 优化过的**工作信息流**：状态、工具调用、文件变更、用户确认、运行结果都以可理解的事件方式出现。

因此需要一个**手机端轻量控制层**，用于在碎片时间里：

* 观察当前 agent/session 的运行状态
* 在需要人工介入时收到通知
* 偶尔补充一条消息推进任务
* 偶尔完成一次事项确认

该产品**不负责解决 agent 为什么停下来**，也不负责替 Codex / Claude Code 提升满负荷持续运行能力。该问题属于底层 agent/runtime 的问题域。

本产品只解决：**当 agent 正在运行、需要观察、需要被唤醒、需要轻量接力时，手机端如何低摩擦地承接。**

产品形态上，应优先承接 coding agent 的工作信息流，而不是复刻 SSH 终端或做一个泛聊天机器人。

---

## 3. 产品定位

### 3.1 产品定义

一个面向 Codex / Claude Code 的**手机端轻量观察与接力前端**。

它消费的是结构化 agent 工作信息流，而不是原始终端文本流；它发送的是轻量控制意图，而不是模拟用户在手机上操作完整开发环境。

### 3.2 非目标

本产品 V1 不做以下事情：

* 不做完整手机 IDE
* 不做完整代码编辑器
* 不做复杂 diff 审阅流
* 不做任务自动规划器
* 不做二次 agent，不替底层 agent 做推理或决策
* 不解决“agent 如何持续工作一天”的底层问题
* 不实现全量 SSH 终端体验替代
* 不把 WebRTC 做成完整远程终端主界面
* 不把产品做成纯对话式聊天客户端

### 3.3 核心价值

* 把黑盒运行态变成可观察、可理解的 agent 工作信息流
* 把“必须回到电脑”变成“手机上先接一手”
* 把高成本接入变成低延迟接入
* 在不增加手机负担的前提下，提高用户与 agent 的协作连续性

一句话定位：

> 桌面负责深度工作，手机负责不中断地接力。

---

## 4. 目标用户与场景

### 4.1 目标用户

* 高频使用 Codex / Claude Code 的开发者
* 已经在桌面端形成 agent 工作流，希望在离开电脑时仍可低成本接力的用户
* 有明确大方向，但依旧需要在小节点上手动推进的个人开发者或小团队成员

### 4.2 典型场景

#### 场景 A：观察运行状态

用户离开电脑后，希望在手机上随时看到当前 session 是否还在运行、正在做什么、是否已经停住。

#### 场景 B：收到提醒后快速接力

agent 进入需要用户确认或补充说明的节点，手机收到通知，用户打开后快速处理并让任务继续。

#### 场景 C：碎片时间补一句话

用户在电梯、通勤、排队等碎片时间，补充一句自然语言指令，让任务继续推进。

#### 场景 D：切换不同运行中的任务

用户同时有多个 session 或 provider 任务，希望在手机端快速切换当前观察对象。

---

## 5. 核心设计原则

### 5.1 轻量优先

手机端只承载观察、提醒、补一句话、做一次确认，不承载复杂操作。

### 5.2 信息流优先于聊天和终端

产品核心不是一个聊天框，也不是一个终端窗口，而是一个面向 coding agent 的工作信息流。

状态、步骤、工具调用、确认、结果摘要应被结构化呈现；原始终端输出只作为必要时的辅助信息，而不是主体验。

### 5.3 不做二次智能体

手机端和中间层不替 Codex / Claude Code 再做一层 agent 推理，只做状态消费和轻控制。

### 5.4 上层统一，底层分适配

Codex 与 Claude Code 如有较大差异，可分开迭代。统一的是上层 UI 和用户心智，不强求底层实现完全统一。

### 5.5 低延迟接入

用户从收到通知到完成一次操作的路径必须尽可能短。

### 5.6 分层解耦

交互层、协议层、连接层、daemon/session gateway 层、provider adapter 层、runtime backend 层应清晰解耦。

V1 先服务 Codex，但协议和 UI 不应绑定到某一个 provider 的终端文本格式，也不应绑定到底层 agent 是否以 CLI 方式运行。后续类似 Claude Code、provider-native server agent、cloud agent 或自研 runner，可以通过 adapter/backend 接入同一套手机端工作流。

---

## 6. 产品范围

### 6.1 V1 必做范围

1. 观察窗口
2. session 切换
3. 新建 session
4. 基于 WebRTC DataChannel 的结构化连接
5. 消息输入
6. 确认通知
7. 远端轻量 daemon
8. Session Runtime Backend
9. V1 默认的 tmux/pty CLI runtime backend

### 6.2 V1 可选但不承诺

* 简化版最近事件列表
* 基础 provider 标识（Codex / Claude）
* 简化的“是否活跃”状态提示
* 终端 peek，用于必要时查看少量原始输出

### 6.3 明确排除范围

* 完整命令行终端交互
* 复杂代码 diff 操作
* 代码编写/编辑主流程
* 多人协作/分享
* 项目管理系统集成
* 自动决策分流
* 本地仓库管理功能
* 将 SSH 或 WebRTC 终端作为主界面

---

## 7. 信息架构

V1 推荐聚焦以下五个核心对象：

### 7.1 Session

表示一个当前可观察、可恢复、可切换的 agent 会话。

建议字段：

* session_id
* provider_type（V1 固定为 codex）
* title
* repo_name
* working_directory
* branch_name
* last_active_at
* status
* is_active
* has_attention

说明：

* V1 首页与列表默认展示主活动 session 及少量其他 session
* 目录、Git 分支属于列表与详情页中的关键信息，应直接展示
* 其他模型配置、session 配置等不放在列表主视图中，放入 session 配置详情页，点击查看

### 7.2 RunState

表示当前运行态。

建议字段：

* run_state（running / waiting / paused / completed / error）
* short_status_text
* current_step_summary
* last_event_at
* requires_attention

### 7.3 Message

表示用户向当前 session 补充的一条自然语言消息。

建议字段：

* message_id
* session_id
* sender（user / assistant / system）
* content
* created_at

### 7.4 Notification

表示一条需要用户关注的提醒。

建议字段：

* notification_id
* session_id
* type（attention / confirmation / completed / failed）
* title
* body
* created_at
* is_read
* action_label

### 7.5 AgentEvent

表示 agent 工作信息流中的一条结构化事件。

建议字段：

* event_id
* session_id
* event_type（state / tool / file / command / test / attention / summary / terminal）
* title
* body
* severity（info / success / warning / error）
* created_at
* metadata

说明：

* AgentEvent 是手机端观察体验的核心输入
* UI 不直接消费 provider 的原始终端文本，而是优先消费 daemon 标准化后的事件
* `terminal` 类型只用于必要的原始输出片段，不作为主信息流

---

## 8. 用户故事

### 8.1 观察相关

* 作为用户，我希望打开手机后就能看到当前哪个 session 在跑，以及它现在是否还活着。
* 作为用户，我希望不用翻聊天记录，也能快速知道 agent 正在做什么。

### 8.2 通知相关

* 作为用户，我希望在 agent 需要我时收到通知，而不是频繁主动检查。
* 作为用户，我希望通知里直接告诉我这是哪个任务、为什么找我、我点进去后能立刻处理。

### 8.3 输入相关

* 作为用户，我希望在手机上只补一句简短指令，就能继续推进当前任务。
* 作为用户，我不希望手机端承担复杂输入负担。

### 8.4 session 相关

* 作为用户，我希望快速切换不同任务/不同 provider 的会话。
* 作为用户，我希望回到上次正在观察的 session，而不是每次重新找。

---

## 9. 核心功能设计

## 9.1 首页 / 观察窗口

### 目标

让用户在 3 秒内理解“现在哪个任务在跑、状态如何、是否需要我介入”。

### 页面结构

#### 主卡片

展示当前主活动 session：

* session 名称
* 当前状态
* 最近一条状态摘要
* 最近活跃时间
* 目录
* Git 分支
* 是否需要关注

#### 下方少量 session 列表

展示若干其他 session，形态类似轻量订单列表：

* session 标题
* 当前状态
* 最近活跃时间
* 是否待处理

### 交互

* 点击主卡片进入当前 session 详情
* 点击列表项切换到其他 session
* 支持进入“新建 session”流程
* 如有待处理通知，首页显式展示

### 设计要求

* 状态文案短
* 层级清晰
* 不依赖长文本滚动
* 主卡片突出“当前主活动 session”，列表承担切换入口

---

## 9.2 Session 列表 / 切换

### 目标

支持用户在多个运行中的 session 之间快速切换，并能新建 session。

### 展示内容

* session 标题
* 当前状态
* 最近活跃时间
* 是否待处理

### 交互

* 单击切换当前观察 session
* 支持进入新建 session
* 默认优先展示最近活跃、需要关注的 session

### 说明

* 首页默认突出主活动 session
* 其他 session 以列表形式展示，便于快速切换
* V1 不在列表页直接暴露复杂 session 配置

---

## 9.3 新建 Session

### 目标

让用户在手机端以最低输入负担创建一个新的 Codex session。

### 最小必填字段

* 目录（必选）
* 分支（可选）
* 首条消息（可选，但建议填写）

### 推荐交互

#### 主路径：目录选择

目录不应以手打路径为主，而应优先通过选择完成。优先级建议如下：

* 最近使用的目录
* 收藏/固定目录
* 当前已有 session 的目录
* 常用 repo / workspace 列表

#### 次路径：分支选择

* 默认当前分支
* 可选已有分支
* 可手动输入新分支名

#### 最后一步：填写首条消息

* 用于启动新 session 的初始任务说明
* 支持自由文本输入

### 高级入口

提供“通过 SSH / 命令方式创建”的高级入口，作为熟悉终端用户的 fallback，但不应作为默认主路径。

### 设计原则

* 目录是核心，但目录路径不适合手机端手打
* 主流程应以选择为主，输入为辅
* 高级方式保留灵活性，但不能破坏轻量主体验

---

## 9.4 Session 详情页

### 目标

提供一个“上面看状态、下面发一句话”的轻量交互页面。

### 页面结构

#### 顶部

* session 名称
* 当前状态标签
* 最近活跃时间
* 目录
* Git 分支
* session 配置入口（点击查看模型配置等信息）

#### 中部

* 简化状态流 / 最近事件
* 当前步骤摘要
* 是否需要确认

#### 底部

* 消息输入框
* 发送按钮
* 如需要确认，则展示确认按钮区域

### 交互原则

* 上半区用于观察
* 下半区用于轻输入
* 不展示复杂控制台字符流作为核心信息
* session 配置不占主界面信息密度，按需点击查看

---

## 9.5 消息输入

### 目标

支持用户用一条自然语言消息补充指导或继续推进。

### 输入特征

* 单条短消息优先
* 不追求复杂 prompt 编排
* 不承担大量上下文整理任务

### 典型输入示例

* 继续按照刚才思路推进
* 先别重构，优先修 bug
* 这个方向可以，继续
* 先停一下，我晚点回来看

---

## 9.6 确认通知

### 目标

当 session 需要用户介入时，及时提醒并提供最短处理路径。

### 触发条件（V1）

* 需要用户确认
* 运行结束
* 运行失败

### 通知要求

* 明确是哪个 session
* 明确为什么提醒
* 点开后能直接进入对应 session

### 设计要求

* 不做噪声通知
* 不对所有状态变化都通知
* 仅保留三类核心通知：需要介入 / 已完成 / 已失败

---

## 10. 交互流程

## 10.1 观察流程

1. 用户打开 App
2. 首页展示当前重点 session
3. 用户查看当前状态摘要
4. 如无事发生，退出即可

## 10.2 通知接力流程

1. agent/runtime 产生需要 attention 的事件
2. 手机收到通知
3. 用户点击进入对应 session
4. 查看当前状态/事项
5. 完成一次确认或补充一句话
6. session 继续运行

## 10.3 切换流程

1. 用户打开 session 列表
2. 选择另一运行中 session
3. 进入详情页继续观察或输入

---

## 11. 底层架构建议

### 11.1 架构原则

* 上层 UI 统一
* 底层 provider 分适配
* 结构化协议优先，不以终端字符流作为主数据源
* 中间层只做状态标准化、事件转译与轻控制，不做智能决策
* 通信层、状态层、provider 层解耦

V1 主架构采用：

```text
Mobile App
  -> Signaling Service
  -> WebRTC DataChannel
  -> Remote Daemon / Session Gateway
  -> Provider Adapter
  -> Session Runtime Backend
       -> Tmux/Pty CLI Runtime
       -> Provider-native Server Runtime
       -> Cloud Agent Runtime
```

其中：

* WebRTC DataChannel 是实时传输层
* signaling service 负责设备配对、会话发现和连接协商
* TURN fallback 只作为直连失败时的兜底链路
* remote daemon / session gateway 是结构化状态层与会话网关
* provider adapter 负责屏蔽不同 agent/provider 的事件与控制接口差异
* Session Runtime Backend 是会话运行与恢复层
* tmux/pty CLI runtime 是 V1 默认 backend，不是唯一 backend
* 手机 App 是可视化控制层

### 11.2 分层职责

#### Mobile App

* 展示 Session / RunState / AgentEvent / Notification
* 支持轻量消息输入和确认操作
* 支持 session 切换、新建和恢复
* 必要时提供 terminal peek，但不把终端作为主界面

#### Signaling Service

* 用户登录与设备配对
* 远端 daemon 注册和在线状态维护
* WebRTC offer / answer / ICE candidate 交换
* 推送通知路由

说明：

signaling service 不承载 agent 工作流主体数据，也不默认中转终端内容。

#### WebRTC DataChannel

* 传输结构化 agent 工作协议
* 支持低延迟状态更新、用户输入、确认响应和心跳
* 支持断线后通过 session cursor 恢复事件流

DataChannel 中传输的核心不是 ANSI 终端字符流，而是结构化消息，例如：

* `session.snapshot`
* `run_state.changed`
* `event.appended`
* `attention.requested`
* `message.send`
* `confirmation.respond`
* `terminal.peek`
* `heartbeat`
* `reconnect.resume`

#### Remote Daemon / Session Gateway

* 管理本机可用 session
* 通过 provider adapter 调用不同 runtime backend
* 启动、恢复、停止 session
* 解析或接收 provider 输出并转译为 AgentEvent
* 标准化 Session / RunState / Notification
* 接收手机端轻量控制消息并注入到底层 agent
* 记录事件游标，支持断线恢复

#### Session Runtime Backend

Session Runtime Backend 负责会话运行、消息注入、事件读取与恢复能力。它是抽象层，不等同于 tmux。

V1 默认实现：

* `TmuxCliBackend`：通过 tmux/pty/shell 托管 Codex CLI

后续可扩展实现：

* `PtyCliBackend`：daemon 直接管理 pty 进程生命周期
* `ProviderServerBackend`：对接 provider-native session API
* `CloudAgentBackend`：对接云端托管 agent task/run
* `CustomRunnerBackend`：对接自研 runner 或企业内部 agent

统一能力包括：

* 创建 session
* 恢复 session
* 发送用户消息
* 响应确认
* 获取 session 快照
* 流式读取 AgentEvent
* 停止 session
* 查询 backend capabilities

tmux 的角色只是 CLI backend 的会话持久化基础设施，不是手机端 UI 的数据模型，也不是长期架构唯一形态。

#### Backend Capability Negotiation

不同 runtime backend 的能力不同，daemon/session gateway 需要在 session attach 或创建时返回 capability 信息，供 App 和协议层决定功能开关。

建议能力字段：

* `persistent_session`
* `structured_events`
* `message_injection`
* `confirmation_response`
* `terminal_peek`
* `file_events`
* `test_events`
* `server_side_notifications`
* `provider_native_resume`

示例：

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

例如 CLI backend 通常支持 `terminal_peek`，但结构化事件需要 daemon 转译；server-native backend 可能不支持 `terminal_peek`，但能提供更可靠的 provider 原生事件。

### 11.3 Provider 适配层

V1 建议先只实现：

* Codex Adapter

Claude Adapter 不纳入 V1 范围，可在后续版本接入。

Provider Adapter 负责把不同 agent 运行形态转成统一协议：

* 标准化 session 信息
* 标准化 run state
* 标准化消息输入接口
* 标准化通知事件
* 标准化 AgentEvent 信息流
* 屏蔽 provider 的终端格式、命令格式和事件格式差异

后续类似 Claude Code、其他 terminal-based coding agent、provider-native server agent、cloud agent 或自研 runner，只要能实现 adapter/backend，就可以接入同一套移动端观察与接力体验。

### 11.4 不建议承担的职责

* 不做二次任务规划
* 不做二次工具调度
* 不做 agent 自主推理
* 不做复杂流程仲裁
* 不把 daemon 做成新的 agent
* 不把 WebRTC DataChannel 做成纯终端流管道

---

## 12. 状态模型（V1 建议）

V1 运行状态建议保持简单：

* `running`：正在运行
* `waiting`：等待用户/等待进一步输入/等待底层恢复
* `paused`：已暂停
* `completed`：已完成
* `error`：运行异常

说明：
手机端 UI 对外仍保持简单五态，但 daemon 内部应保留 `reason_code`，用于通知准确性、事件归类和后续扩展。

建议 V1 内部 reason_code 示例：

* `waiting_for_instruction`
* `waiting_for_confirmation`
* `waiting_for_permission`
* `running_command`
* `editing_files`
* `running_tests`
* `summarizing_result`
* `recovering_session`
* `provider_error`

这样既不增加手机端理解负担，也避免架构层过度依赖模糊文本。

---

## 13. 通知策略

### 13.1 必须通知

* 当前活跃 session 进入 waiting 且需要用户介入
* 当前活跃 session 完成
* 当前活跃 session 异常中断

### 13.2 可以不通知

* 普通状态变更
* 高频事件流更新
* 无需用户动作的中间步骤

### 13.3 通知体验目标

* 少而关键
* 一眼能懂
* 一点就进

---

## 14. 设计风格建议

* 强状态感，弱装饰感
* 首页信息密度适中
* 优先卡片式和标签式表达
* 不在手机端暴露复杂工程细节
* 让用户感受到“我能随时接住”，而不是“我必须在手机上完整工作”

---

## 15. 成功指标

### 15.1 产品指标

* 用户从收到通知到进入 session 的时间
* 用户从进入 session 到完成一次处理的时间
* 手机端消息发送后的继续运行成功率
* 活跃用户每周通过手机端完成的接力次数

### 15.2 体验指标

* 用户认为“比 SSH 对话框明显更轻松”的主观反馈
* 用户是否愿意把手机端作为日常观察入口
* 通知是否被认为有用而非打扰

---

## 16. V1 / V2 规划建议

## V1

聚焦“可观察、可通知、可补一句话、可切 session”，并直接采用结构化连接主架构：

* WebRTC DataChannel
* signaling service
* TURN fallback
* remote daemon
* Session Runtime Backend
* tmux/pty CLI runtime backend 作为 V1 默认实现

## V2 可考虑

* 更丰富的状态时间线
* 更好的事件摘要
* 更清晰的确认卡片
* 更好的 provider 差异适配
* 与桌面端 session 的更强连续性
* 更丰富的 session 自动化与批量任务管理

---

## 17. 风险与约束

### 17.1 Provider 能力差异

Codex 与 Claude Code 对 session、event、notification、恢复能力的暴露程度可能不同，需在适配层处理。

### 17.2 手机端过载风险

如果把过多桌面端细节直接搬到手机，会破坏“轻量”定位。

### 17.3 通知噪声风险

如果通知策略不收敛，会迅速让用户关闭通知。

### 17.4 状态抽象过细风险

V1 若过早引入过多状态类型，会增加 UI 和接入复杂度，不利于快速验证。

### 17.5 连接复杂度风险

WebRTC DataChannel + signaling + TURN fallback 的工程复杂度高于 SSH/Mosh。V1 需要限制协议范围，只传输结构化状态、轻量控制消息、心跳与必要的 terminal peek，避免把连接层做成完整远程终端。

### 17.6 Daemon 过重风险

daemon 是结构化状态层，不是新的 agent runtime。V1 应只承担 session 管理、状态转译、消息注入、事件记录和断线恢复，不承担智能规划、工具调度或复杂流程仲裁。

---

## 18. 已确认的 V1 边界

1. **Provider 范围**
   V1 先只支持 Codex。

2. **首页形态**
   首页采用“一个主活动 session 卡片 + 下方少量 session 列表”的结构。

3. **Session 能力**
   V1 支持切换 session 与新建 session。

4. **首页与详情展示信息**
   目录与 Git 分支直接展示；其他模型配置、session 配置放入配置详情中，点击查看。

5. **新建 Session 方式**
   新建 session 最少字段为：目录（必选）+ 分支（可选）+ 首条消息（可选）。目录优先通过 daemon 返回的最近/收藏/可选列表选择，手动输入作为兜底；SSH / 命令方式仅作为高级 fallback。

6. **通知范围**
   V1 保留三类通知：需要介入 / 已完成 / 已失败。

7. **输入方式**
   V1 先只支持自由文本输入，不做快捷操作按钮。

8. **Session 展示策略**
   首页突出主活动 session，其他 session 以列表形态展示，类似轻量订单列表。

9. **主连接策略**
   V1 主连接策略采用 WebRTC DataChannel + signaling + TURN fallback + remote daemon/session gateway + Session Runtime Backend。

10. **信息流策略**
   手机端主体验消费结构化 AgentEvent，不直接消费原始终端字符流。terminal peek 仅作为必要时的辅助能力。

11. **Runtime 策略**
   V1 默认 runtime backend 为 tmux/pty CLI backend，用于托管当前 CLI 形态的 Codex。架构上保留 provider-native server runtime、cloud agent runtime 和 custom runner runtime 的扩展口。

## 19. 一句话结论

这个产品不是手机上的完整 agent IDE，也不是二次智能体，更不是 WebRTC 远程终端。

它是一个**面向 coding agent 工作信息流的轻量观察与接力前端**：让用户在不回到电脑的情况下，也能看见 agent、被 agent 唤醒、并用很低的成本接一手。

---

## 20. 附录：连接方案与当前决策

这一部分用于明确：V1 如何连接家里服务器、云主机或远端开发机上的 Codex。

### 20.1 背景问题

当前使用场景中，Codex 通常运行在家里服务器、云主机或远端开发机上。手机端如果要承担“观察、提醒、补一句话、做一次确认”的角色，就必须解决两个问题：

1. 如何建立低延迟、可恢复、移动网络友好的连接。
2. 如何避免把终端字符流直接暴露成产品主体验。

纯 SSH/Mosh 的问题不是不能连接，而是它天然提供的是终端文本流。终端文本适合人坐在桌面前操作，却不适合手机端做状态摘要、通知、确认卡片和任务切换。

因此，V1 不以 SSH 作为主路径，也不以 WebRTC 远程终端作为主路径，而是采用结构化 agent 工作协议。

---

### 20.2 当前主方案：方案 B+

#### 方案名称

**WebRTC DataChannel + signaling + TURN fallback + remote daemon/session gateway + Session Runtime Backend**

#### 方案描述

* 手机 App 通过 signaling service 与远端 daemon 完成配对和连接协商
* 手机 App 与远端 daemon 之间优先建立 WebRTC DataChannel
* DataChannel 传输结构化 agent 工作协议
* 直连失败时通过 TURN fallback
* 服务器侧由 daemon/session gateway 管理 session
* daemon 通过 provider adapter 调用 Session Runtime Backend
* V1 默认使用 tmux/pty CLI backend 保持 Codex CLI 会话持续运行
* 未来可替换为 provider-native server backend 或 cloud agent backend
* runtime backend 负责会话运行与恢复，不负责定义手机端 UI 形态

#### 核心判断

WebRTC 在本产品里不是“远程桌面/远程终端技术”，而是实时通信协议。

DataChannel 中传输的核心不是 ANSI 文本流，而是结构化事件和控制消息：

* session 快照
* run state 变化
* agent 工作事件
* attention 请求
* 用户消息
* 确认响应
* 心跳
* 断线恢复游标
* 必要时的 terminal peek

#### 为什么优先采用该方案

1. **符合产品真实目标**
   产品要解决的是 agent 工作可视化和移动接力，不是做一个更复杂的手机终端。

2. **适合 coding agent 信息流**
   Codex 桌面端、Claude Code 一类产品的交互都更接近 coding agent 工作信息流，而不是普通聊天流。手机端应复用这种心智：状态、步骤、工具调用、确认和结果摘要优先。

3. **避免纯文本协议上限**
   SSH/Mosh 的文本流很难自然支撑结构化通知、状态归类、确认卡片、session 列表、事件游标和断线恢复。

4. **具备产品化潜力**
   signaling + WebRTC + TURN fallback 更适合没有公网 IP、不想暴露公网端口、移动网络频繁切换的真实用户环境。

5. **保留底层持久化能力**
   V1 由 tmux/pty CLI backend 承担会话持久化，保障断线后 session 不丢失；未来如果 provider 提供 server/native session API，则由 provider 自身承担持久化。

6. **层与层解耦**
   App 不绑定 Codex 终端格式；daemon 不绑定某一种 UI；provider adapter 不影响连接层；runtime backend 不固定为 CLI/tmux。后续其他类似 coding agent 可以接入部分模块。

---

### 20.3 分层边界

#### Mobile App

主体验是 agent 工作信息流：

* 当前 session 状态
* 当前步骤摘要
* 最近事件
* attention 卡片
* 轻量输入
* session 切换
* 新建 session

不做完整终端主界面。

#### Signaling Service

负责：

* 用户登录
* 设备配对
* daemon 注册
* 会话发现
* WebRTC 连接协商
* 推送通知路由

不默认承载 agent 数据面。

#### WebRTC DataChannel

负责实时数据面：

* 结构化事件流
* 用户控制消息
* confirmation response
* heartbeat
* reconnect resume
* terminal peek

不作为完整终端流管道。

#### Remote Daemon / Session Gateway

负责：

* session 管理
* provider adapter 调用
* 状态转译
* 消息注入
* 事件持久化
* 通知生成
* 断线恢复

不做二次 agent，不做智能规划。

#### Session Runtime Backend

负责：

* 创建、恢复、停止 session
* 发送用户消息
* 响应用户确认
* 获取 session snapshot
* 流式读取 AgentEvent
* 查询 backend capabilities

V1 默认实现：

* `TmuxCliBackend` / `PtyCliBackend`：托管 CLI 形态的 Codex

未来扩展实现：

* `ProviderServerBackend`：对接 provider-native session API
* `CloudAgentBackend`：对接云端托管 agent task/run
* `CustomRunnerBackend`：对接自研 runner 或企业内部 agent

tmux/pty 不作为手机端产品界面，也不作为长期唯一运行形态。

---

### 20.4 SSH / Mosh 的定位

SSH / Mosh 不作为 V1 主连接策略。

它可以保留为高级 fallback：

* 远端 daemon 安装或诊断入口
* 高级用户手动排障
* 连接失败时的紧急接入方式
* 开发调试时的运维通道

但产品主体验不能依赖用户直接消费 SSH 终端，否则会回到当前手机 SSH 体验差的问题。

---

### 20.5 不建议作为主路径的方案

#### 不建议 1：纯 SSH / Mosh + tmux

原因：

* 数据形态是终端文本流，不是 agent 工作信息流
* 难以稳定抽象状态、确认、通知和事件历史
* 手机端体验上限低
* 连接配置和排障成本都交给用户

#### 不建议 2：WebRTC 远程终端

原因：

* 技术投入会被终端体验吞掉
* 容易把产品带偏成远程终端 App
* 仍然没有天然解决结构化状态问题
* 与“轻量观察与接力”的核心价值不一致

#### 不建议 3：全量 relay 中转

原因：

* 服务成本高
* 带宽和长连接成本高
* 延迟更不可控
* 对当前轻量数据面不划算

TURN 只作为 fallback，不作为默认主链路。

---

### 20.6 后续演进路径

#### V1

* WebRTC DataChannel + signaling + TURN fallback
* remote daemon
* Session Runtime Backend
* tmux/pty CLI backend
* Codex Adapter
* 结构化 AgentEvent
* terminal peek 作为辅助能力
* SSH / Mosh 作为高级 fallback

#### V1.5

* 更完整的事件摘要
* 更清晰的 confirmation card
* session cursor 和离线事件补拉
* daemon 安装、升级和诊断能力
* 更稳定的目录/分支选择
* backend capability negotiation

#### V2

* Claude Adapter 或其他 provider adapter
* ProviderServerBackend / CloudAgentBackend
* 更强的桌面端 session 连续性
* 多任务队列和批量观察
* 更丰富的移动端任务发起能力
* 团队协作和权限模型

---

### 20.7 结论

V1 主链路保持不变，B+ 优化部分进入 V1.1：

**WebRTC DataChannel + signaling + TURN fallback + remote daemon/session gateway + Session Runtime Backend。**

一句话总结：

> WebRTC 是实时传输层，daemon/session gateway 是结构化状态层，Session Runtime Backend 是会话运行与恢复层，手机 App 是可视化控制层。

这里的 WebRTC DataChannel、signaling、TURN fallback 等能力属于 V1.1 增强链路，不表示 V1 已切换当前主链路。

V1.1 默认通过 tmux/pty CLI backend 托管当前 CLI 形态的 Codex；未来如果 provider 演进为 server/native session 模式，可以切换到 ProviderServerBackend 或 CloudAgentBackend，而不影响 App、协议层和信息流体验。

该方案的核心不是复刻终端，也不是绑定 CLI runtime，而是为 coding agent 建立一套适合移动场景的结构化工作协议。
