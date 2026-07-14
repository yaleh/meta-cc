---
id: TASK-11
title: 'B-1: query_session_content(role=tool) 默认输出附带外层上下文字段（timestamp/session_id/turn）'
status: 'Basic: Done'
assignee: []
created_date: '2026-06-23 15:43'
updated_date: '2026-06-23 16:00'
labels:
  - 'kind:basic'
dependencies: []
references:
  - docs/guides/mcp-query-tools.md
  - internal/mcp/executor/consolidated_handlers.go
priority: high
ordinal: 1000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
当前 query_session_content(role=tool) 返回的是裸 tool_use block，字段仅含 {type, id, name, input, caller}。timestamp / session_id / turn 存在于外层 JSONL record，被工具内部抽取时丢弃。Agent 写 jq transform 时自然假设这些字段随结果一起返回，导致全空输出——查询形式成功但内容为空，比报错更难察觉。期望：非 summary 模式下，每条 tool 结果默认合并外层字段，至少附带 {timestamp, session_id, turn}。
<!-- SECTION:DESCRIPTION:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
# Proposal: B-1 query_session_content(role=tool) 附带外层上下文字段

## Background

`query_session_content(role=tool)` 最终调用 `handleQueryToolBlocks`（`internal/mcp/executor/handlers.go`），其 jq filter 将外层 JSONL record 管道进 `.message.content[]` 后再筛选，导致外层字段（`timestamp`、`sessionId`、`turn`）在结果中完全丢失。

结果对象仅含 `{type, id, name, input}` 等 tool block 自身字段，Agent 在 jq transform 里引用 `.timestamp` 或 `.session_id` 时得到 null，查询成功但数据为空，比报错更难排查。

相比之下，`content_summary=true` 模式经由 `ApplyContentSummary`（`internal/mcp/filters/filters.go`）处理，输出已含 `session_id`、`timestamp` 等字段，行为一致性缺失。

修复方向：在 jq filter 层直接合并外层字段到每个 tool block，非 summary 模式即生效，不引入新的 Go 处理层。

## Goals

1. `query_session_content(role=tool, block_type=tool_use)` 每条结果包含 `{timestamp, sessionId, turn}` 来自外层 JSONL record。
2. `query_session_content(role=tool, block_type=tool_result)` 同上，同样附带外层上下文字段。
3. `content_summary=true` 模式现有行为不变。
4. 改动不影响其他 role（user/assistant/all）的输出。
5. `go test ./...` 全部通过。

## Proposed Approach

**只修改 `internal/mcp/executor/handlers.go` 的 `handleQueryToolBlocks` 函数。**

将现有 jq filter 从"管道进数组再选择"改为"在外层 record 上做对象合并"：

- `tool_use` filter 改为：
  `select(.type == "assistant") | . as $rec | .message.content[] | select(.type == "tool_use") | {timestamp: $rec.timestamp, sessionId: $rec.sessionId, turn: $rec.turn} + .`

- `tool_result` filter 改为：
  `select(.type == "user" and (.message.content | type == "array")) | . as $rec | .message.content[] | select(.type == "tool_result") | {timestamp: $rec.timestamp, sessionId: $rec.sessionId, turn: $rec.turn} + .`

jq 的 `. as $rec` 捕获外层 record，`+ .` 合并确保 tool block 自身字段优先（若字段名冲突），外层仅补充 `timestamp`/`sessionId`/`turn`。

不涉及 `filters.go`（`ApplyContentSummary` 路径已正确）、`pipeline.go`、`tools.go`（schema 无需变更）。

## Trade-offs and Risks

- **不做什么**：不引入新的 Go struct 或后处理层，保持 jq-only 实现风格一致。不改变 `content_summary=true` 的输出格式（已含这些字段）。
- **字段命名**：JSONL raw record 使用 camelCase（`sessionId`），与 `ApplyContentSummary` 输出的 snake_case（`session_id`）不同；本 fix 维持 JSONL 原始命名（`sessionId`），与其他非-summary 工具一致。
- **字段冲突风险**：tool block 自身若有同名字段（概率极低），`+ .` 保证 tool block 字段优先，外层仅补充。
- **性能**：纯 jq 层变更，无额外 Go 分配，性能影响可忽略。

---

# Plan: TDD Implementation

## Phase 1 — 修改 jq filter，合并外层上下文字段

### Tests（write first）

文件：`internal/mcp/executor/handlers_test.go`（已存在，追加新测试）

新增测试用例：

- `TestHandleQueryToolBlocks_ToolUse_IncludesTimestamp`
  — 构造含 `timestamp`/`sessionId`/`turn` 的 assistant JSONL record，调用 `handleQueryToolBlocks`，断言每条结果含 `timestamp` 且不为 nil。

- `TestHandleQueryToolBlocks_ToolUse_IncludesSessionId`
  — 同上，断言结果含 `sessionId`（camelCase）。

- `TestHandleQueryToolBlocks_ToolUse_IncludesTurn`
  — 同上，断言结果含 `turn`。

- `TestHandleQueryToolBlocks_ToolResult_IncludesTimestamp`
  — 构造含 `timestamp`/`sessionId`/`turn` 的 user JSONL record（content 为 tool_result 数组），断言结果含 `timestamp`。

- `TestHandleQueryToolBlocks_ToolUse_PreservesToolFields`
  — 断言原有 tool block 字段（`type`、`id`、`name`、`input`）仍然存在，未被外层字段覆盖。

### Implementation

文件：`internal/mcp/executor/handlers.go`，函数 `handleQueryToolBlocks`（第 158-176 行）

修改 `jqFilter` 赋值部分：

```
tool_use:
  旧: select(.type == "assistant") | .message.content[] | select(.type == "tool_use")
  新: select(.type == "assistant") | . as $rec | .message.content[] | select(.type == "tool_use") | {timestamp: $rec.timestamp, sessionId: $rec.sessionId, turn: $rec.turn} + .

tool_result:
  旧: select(.type == "user" and ...) | .message.content[] | select(.type == "tool_result")
  新: select(.type == "user" and ...) | . as $rec | .message.content[] | select(.type == "tool_result") | {timestamp: $rec.timestamp, sessionId: $rec.sessionId, turn: $rec.turn} + .
```

估计变更：~4 行（2 行 jq 字符串替换）

### DoD

```sh
go test ./...
go test ./internal/mcp/executor/... -v -run TestHandleQueryToolBlocks
go vet ./internal/mcp/executor/...
```

## Phase 2 — 更新 MCP 工具 schema 描述，反映新输出字段

### Tests（write first）

文件：`internal/mcp/tools/tools_test.go`（已存在，追加新测试）

新增测试用例：

- `TestQuerySessionContentSchema_ToolRole_DescribesContextFields`
  — 读取 `query_session_content` 的 `role` 字段 description，断言含有 "timestamp" 或 "sessionId" 关键词（验证 schema 描述已更新）。

### Implementation

文件：`internal/mcp/tools/tools.go`，`BuildTool("query_session_content", ...)` 中 `"block_type"` property 的 `Description` 字段（第 257 行附近）

将 `Description` 从：
```
"When role=tool: 'tool_use' or 'tool_result' (default: 'tool_use')"
```
改为：
```
"When role=tool: 'tool_use' or 'tool_result' (default: 'tool_use'). Each result includes outer context fields: timestamp, sessionId, turn."
```

估计变更：~2 行

### DoD

```sh
go test ./...
go test ./internal/mcp/tools/... -v -run TestQuerySessionContentSchema
go vet ./internal/mcp/tools/...
```
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Proposal + Plan APPROVED. Ready for implementation.

claimed: 2026-06-23T15:50:00Z

Completed: 2026-06-23T16:05:00Z

Fixed handleQueryToolBlocks jq filters to capture outer record with `. as $rec` and merge {timestamp, sessionId, turn} before tool block fields, for both tool_use and tool_result paths. Updated block_type schema description. 5 new tests in handlers_test.go + 1 in tools_test.go. All 37 test packages pass.
<!-- SECTION:NOTES:END -->
