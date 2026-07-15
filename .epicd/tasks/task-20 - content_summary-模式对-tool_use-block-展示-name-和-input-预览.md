---
id: TASK-20
title: content_summary 模式对 tool_use block 展示 name 和 input 预览
assignee: []
created_date: '2026-07-14 03:52'
updated_date: '2026-07-15 02:16'
labels:
  - 'area:mcp'
  - 'area:query'
dependencies: []
priority: high
ordinal: 11000
pipeline_id: authoring
phase: backlog
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Background

`query_session_content(role=tool, block_type=tool_use, content_summary=true)` 的 `content_preview` 字段恒为空。根本原因：`ApplyContentSummary`（`internal/mcp/filters/filters.go:103`）调用 `extractContentString`，该函数只处理 string 类型的 `message.content` 或 flat `content` 字段。但 `handleQueryToolBlocks` 返回的 tool_use 记录是扁平结构 `{type, name, input, timestamp, sessionId, turn}`，无 string content 字段，extractContentString 始终返回空字符串。

结果：有 5000+ 条 tool_use 记录的项目中，summary 模式无法区分任何记录，实际失效。

## Goals

1. `content_preview` 对 tool_use records 格式为 `<name> <input_json_truncated>`，例如 `mcp__manda__Dispatch {"id": "p-1",...}`
2. 截断长度由现有 `preview_length` 参数控制（默认 100）
3. tool_result 及其他 block 类型的 preview 行为不变

## Approach

在 `ApplyContentSummary`（`filters.go`）的记录循环中，先检查 `msgMap["type"] == "tool_use"`，若是则调用新的 `buildToolUsePreview(name, input, previewLength)` 构造 preview（name 拼 JSON marshal 的 input，截断到 previewLength）；否则走原有 `extractContentString` 路径。

`extractContentString` 本身不改动，tool_result 的 content 已是字符串，仍走原路径。

## Non-Goals

- 不改变 role=user / role=assistant 的 preview 逻辑
- 不新增参数（复用 `preview_length`）
- 不改变 tool_result block 的 preview 逻辑
- 不改变 summary 输出的其他字段（session_id、turn_sequence、timestamp、uuid）
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 preview 格式为 '<name> <input_json_truncated>'
- [ ] #2 preview 长度受 preview_length 参数控制
- [ ] #3 tool_result block 的 preview 行为不受影响
- [ ] #4 go test ./internal/mcp/filters/... -run TestApplyContentSummary_ToolUse_NonEmpty 通过；返回记录的 content_preview 字段非空字符串
- [ ] #5 content_preview 匹配 Go regexp ^[A-Za-z_][\w:]* \{；即 name（字母数字下划线冒号）后跟空格和 JSON 对象开头；验证：TestApplyContentSummary_ToolUse_Format
- [ ] #6 preview_length=20 时 content_preview 的 rune 长度 ≤ 23（20 字符 + "..."）；验证：TestApplyContentSummary_ToolUse_PreviewLength
- [ ] #7 go test ./internal/mcp/filters/... -run TestApplyContentSummary_ToolResult 通过；tool_result 记录的 content_preview 与改动前行为一致（有 content 时非空，无 content 时为空）
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
# Plan: content_summary 模式对 tool_use block 展示 name 和 input 预览

## Phase A: TDD — 先写测试，再实现

### Tests (write first)
- 文件: `internal/mcp/filters/filters_test.go`
- 新增（写入后运行 `go test ./internal/mcp/filters/... -run TestApplyContentSummary_ToolUse` 应红灯）：
  - `TestApplyContentSummary_ToolUse_NonEmpty` — 构造 tool_use record，断言 content_preview 非空
  - `TestApplyContentSummary_ToolUse_Format` — 断言 content_preview 以 `name + " {"` 开头（regexp）
  - `TestApplyContentSummary_ToolUse_PreviewLength` — preview_length=20，断言 rune 长度 ≤ 23
  - `TestApplyContentSummary_ToolResult_Unchanged` — tool_result record，断言 preview 仍走原 extractContentString 路径（有 content 字段时非空）

### Implementation
- 文件: `internal/mcp/filters/filters.go`
- 在 `ApplyContentSummary`（line 110 循环内），提取 preview 逻辑改为：
  - 若 `msgMap["type"] == "tool_use"`：调用 `buildToolUsePreview(msgMap, previewLength)`
  - 否则：走原有 `extractContentString` 路径
- 新增私有函数 `buildToolUsePreview(msgMap map[string]interface{}, previewLength int) string`：
  1. 提取 `name` string 和 `input` interface{}
  2. `json.Marshal(input)` → inputJSON（marshal 失败时退回空字符串）
  3. 拼接 `name + " " + string(inputJSON)`，按 rune 截断到 previewLength，超长追加 "..."

### DoD
- [ ] `go test ./internal/mcp/filters/... -run 'TestApplyContentSummary_ToolUse|TestApplyContentSummary_ToolResult_Unchanged'`
- [ ] `make commit`

## Constraints
- 不改 `extractContentString` 本身
- 不改 summary 输出的其他字段（session_id、turn_sequence、timestamp、uuid）
- 不新增 MCP 参数
- json.Marshal 失败时 buildToolUsePreview 返回空字符串（退化到之前行为），不 panic

## Acceptance Gate
- [ ] `make commit`
- [ ] `go test ./internal/mcp/filters/... -run TestApplyContentSummary`
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
authoring/draft self-review: Approved after 2 round(s)

authoring/refining review: APPROVED after 2 iteration(s)

Phase A done: TDD complete. 4 new tests (NonEmpty/Format/PreviewLength/ToolResult_Unchanged) all pass. buildToolUsePreview() added to filters.go. make commit passed all hooks. Committed ec998f0.
<!-- SECTION:NOTES:END -->
