# M0 Codex CLI 可解析性报告模板

## 1. 报告目标

本报告用于判断 Codex CLI 输出是否足以支撑 V1 的结构化 AgentEvent 信息流。

核心判定：

```text
high-confidence AgentEvent coverage >= 60%
```

如果低于 60%，暂停 M2，重新评估 provider-native structured output/API、额外 instrumentation 或降低 V1 信息流承诺。

---

## 2. 采样环境

| 项 | 内容 |
|---|---|
| 日期 |  |
| 操作系统 |  |
| Codex CLI 版本 |  |
| terminal/tmux 环境 |  |
| repo |  |
| branch |  |
| daemon commit / build |  |

---

## 3. Confidence 定义

### High

来源稳定、可重复、误判概率低。

示例：

* 进程启动/退出
* exit code
* daemon 注入用户消息
* 文件系统 watcher 发现文件变化
* git status 发现变更
* 明确等待用户输入
* 明确 completed/error

### Medium

由多个信号推断，通常可靠，但可能受上下文影响。

示例：

* 从命令输出和文件变化推断正在运行测试
* 从连续输出模式推断正在执行某类任务
* 从 stdout/stderr 组合推断错误类型

### Low

主要依赖 CLI 文本、关键词或启发式判断。

示例：

* 仅从自然语言片段推断“正在编辑文件”
* 仅从输出摘要推断当前步骤
* 无法稳定区分工具调用与普通说明

---

## 4. 采样任务列表

| ID | 任务 | 预期覆盖场景 | 结果 |
|---|---|---|---|
| T1 | 简单文件修改 | file.changed / completed |  |
| T2 | bug fix | editing / command / tests |  |
| T3 | 运行测试 | command / test / exit code |  |
| T4 | 等待用户确认 | waiting / confirmation |  |
| T5 | 用户补充一句话 | message.send / resume |  |
| T6 | 失败命令 | command.finished / error |  |
| T7 | 任务总结 | summary.generated / completed |  |

---

## 5. 事件覆盖记录

| EventType | Source | Confidence | 样本数 | 成功识别 | 误判 | 漏判 | 备注 |
|---|---|---|---:|---:|---:|---:|---|
| `run_state.changed` | process/runtime | high |  |  |  |  |  |
| `message.received` | daemon | high |  |  |  |  |  |
| `command.started` | cli_parser/runtime |  |  |  |  |  |  |
| `command.finished` | runtime/exit_code | high |  |  |  |  |  |
| `file.changed` | filesystem_watcher | high |  |  |  |  |  |
| `test.started` | cli_parser |  |  |  |  |  |  |
| `test.finished` | exit_code/cli_parser |  |  |  |  |  |  |
| `attention.requested` | cli_parser/runtime |  |  |  |  |  |  |
| `confirmation.requested` | cli_parser/runtime |  |  |  |  |  |  |
| `summary.generated` | cli_parser |  |  |  |  |  |  |
| `error.raised` | stderr/exit_code |  |  |  |  |  |  |
| `terminal.peek` | terminal | high |  |  |  |  | fallback |

---

## 6. Coverage 计算

建议公式：

```text
high_confidence_coverage =
  high_confidence_successfully_identified_events /
  expected_core_events
```

其中 expected_core_events 只统计 V1 核心体验需要的事件，不统计所有低价值日志。

V1 core events：

* `run_state.changed`
* `message.received`
* `command.finished`
* `file.changed`
* `attention.requested`
* `confirmation.requested`
* `summary.generated`
* `error.raised`

结果：

```text
expected_core_events:
high_confidence_successfully_identified_events:
high_confidence_coverage:
```

---

## 7. 不可解析 / 误判案例

### Case 1

原始输出：

```text

```

预期事件：

```text

```

实际结果：

```text

```

原因分析：

```text

```

处理建议：

```text

```

---

## 8. 结论

### 是否达到 60% 门槛

```text
是 / 否
```

### 是否建议进入 M2

```text
是 / 否
```

### V1 可承诺事件

```text

```

### V1 需要降级展示的事件

```text

```

### 需要 provider-native API 或额外 instrumentation 的事件

```text

```
