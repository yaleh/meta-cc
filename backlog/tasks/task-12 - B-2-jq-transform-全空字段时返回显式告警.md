---
id: TASK-12
title: 'B-2: jq transform 全空字段时返回显式告警'
status: 'Basic: Done'
assignee: []
created_date: '2026-06-23 15:43'
updated_date: '2026-06-23 15:57'
labels:
  - 'kind:basic'
dependencies: []
references:
  - docs/guides/two-stage-query-guide.md
  - internal/mcp/executor/stage2.go
priority: high
ordinal: 2000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
当 execute_stage2_query 或 query_session_content 等工具应用 jq transform 后，所有结果记录的所有字段均为 null/空字符串时，当前实现静默返回空行数据，不给任何提示。Agent 无法区分"真的没数据"和"字段路径写错了"，往往误信空结果或直接放弃工具改用 Bash。期望：检测到 transform 输出全空字段时，在响应中追加 warning，建议用户检查字段路径，并提示先用 inspect_session_files(include_samples=true) 确认实际结构。
<!-- SECTION:DESCRIPTION:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
# Proposal: B-2 jq transform 全空字段时返回显式告警

## Background

当 Agent 使用 `execute_stage2_query` 或 `query_session_content` 的 `transform` 参数时，若字段路径写错（例如 `.nonexistent_field`），jq 会对每条记录返回 `null`，形成全为 `null` 的结果列表。当前实现静默返回这些空行，响应的 `warnings` 数组为空。Agent 看到结果列表非空（记录数量正常），却无法判断数据是否有效，往往误以为"数据本来就是空的"或直接放弃工具改用 Bash。

两条受影响路径的警告机制已经存在但未覆盖此场景：`query_session_content` 等工具通过 `pipeline.BuildResponse` → `InjectWarnings` 注入警告；`execute_stage2_query` 则通过 `mcquery.HandleExecuteStage2Query` 直接序列化结果，返回的 JSON 对象中无 `warnings` 字段。

## Goals

1. 当 transform 参数存在且输出结果中每条记录的每个字段值均为 `null` 或空字符串时，在响应的 `warnings` 数组中追加明确告警。
2. 告警内容引导用户：（a）检查字段路径是否正确；（b）调用 `inspect_session_files(include_samples=true)` 确认实际数据结构。
3. 覆盖两条路径：`execute_stage2_query`（engine 层）和 `query_session_content`（query/executor 层）。
4. 无 transform、或结果为空、或结果中存在非空字段时，不触发告警（零误报）。

## Proposed Approach

**Path 1 — `execute_stage2_query`**

修改 `internal/query/engine/stage2.go`：在 `ExecuteStage2Query` 函数内，`streamFilesWithJQ` 返回 `results` 后，当 `query.Transform != ""` 时，调用新增的 `detectAllNullTransform(results) bool` 辅助函数检测是否全空。若为真，在返回的 `Stage2Result` 里追加 `Warnings []string` 字段，并在 `Stage2Result` struct 中增加该字段。

然后在 `internal/mcp/query/stage.go` 的 `HandleExecuteStage2Query` 中，把 `result.Warnings` 合并到最终 JSON 响应的 `warnings` 键。

**Path 2 — `query_session_content`（以及其他 consolidated 工具）**

修改 `internal/mcp/query/query.go`：在 `QueryExecutor.RunQuery` / `RunQueryWithTimeRange` 里，当 `transform != ""` 时，检测 `StreamFiles` / `StreamFilesWithTimeRange` 返回的 `QueryResult.Entries`，若全空则 `append` 到 `result.Warnings`。该 `QueryResult.Warnings` 已经通过 `pipeline.BuildResponse` → `InjectWarnings` 注入到响应，无需额外改动。

**辅助函数**

新增 `isAllNullOrEmpty(entries []interface{}) bool`，逻辑：遍历所有 entry，若为 `map[string]interface{}`，检查所有 value 是否均为 `nil` 或 `""`；若为其他类型（`string`, `bool`, `float64` 等标量），视为非空直接返回 false；空 entries 返回 false。

## Trade-offs and Risks

- **不处理**：transform 输出为标量（如 `.message.content` 返回字符串列表）的情况，标量非空可直接判断，不额外告警。
- **不处理**：transform 输出为数组类型的情况，留待后续需求。
- **风险**：`detectAllNullTransform` 在记录数量大时有遍历开销，但属于线性 O(n)，可接受。
- **不改变**：无 transform 时的行为，保持向后兼容。

---

# Plan: B-2 TDD Implementation

## Phase 1: engine 层检测与 Stage2Result.Warnings

### Tests（write first）

文件：`internal/query/engine/stage2_test.go`

- `TestIsAllNullOrEmpty_AllNull` — 输入全 null map，期望 true
- `TestIsAllNullOrEmpty_AllEmpty` — 输入全空字符串 map，期望 true
- `TestIsAllNullOrEmpty_Mixed` — 输入含非空字段，期望 false
- `TestIsAllNullOrEmpty_Scalar` — 输入标量字符串，期望 false
- `TestIsAllNullOrEmpty_EmptySlice` — 输入空 slice，期望 false
- `TestExecuteStage2Query_TransformAllNull_Warning` — transform=`.nonexistent` 产生全 null 结果，Stage2Result.Warnings 非空，包含 "inspect_session_files"
- `TestExecuteStage2Query_TransformValidField_NoWarning` — transform=`{type}` 产生有效结果，Stage2Result.Warnings 为空
- `TestExecuteStage2Query_NoTransform_NoWarning` — 无 transform，Stage2Result.Warnings 为空

### Implementation

文件：`internal/query/engine/stage2.go`

- `Stage2Result` struct 增加 `Warnings []string` 字段
- 新增 `isAllNullOrEmpty(entries []interface{}) bool`
- `ExecuteStage2Query`：在返回前，若 `query.Transform != ""` 且 `isAllNullOrEmpty(results)`，追加告警到 `Stage2Result.Warnings`

### DoD

```
go test ./internal/query/engine/...
go test ./...
```

---

## Phase 2: HandleExecuteStage2Query 注入 warnings 到 JSON 响应

### Tests（write first）

文件：`internal/mcp/query/stage_test.go`（新建）或现有 handler test

- `TestHandleExecuteStage2Query_TransformAllNull_WarningsInResponse` — mock JSONL + transform=`.nonexistent`，响应 JSON 包含非空 `warnings` 数组
- `TestHandleExecuteStage2Query_TransformValidField_EmptyWarnings` — transform=`{type}`，响应 JSON 的 `warnings` 为空数组

### Implementation

文件：`internal/mcp/query/stage.go`（`HandleExecuteStage2Query` 函数）

- 在序列化返回 map 时，把 `result.Warnings` 合并到 `"warnings"` 键（若为 nil 则用空 slice）

### DoD

```
go test ./internal/mcp/query/...
go test ./...
```

---

## Phase 3: query_session_content transform 路径告警

### Tests（write first）

文件：`internal/mcp/query/query_test.go`（或新建 `query_transform_warning_test.go`）

检查是否已有该文件：`internal/mcp/query/`

- `TestRunQuery_TransformAllNull_Warning` — filter=`select(.type=="user")`, transform=`.nonexistent`，QueryResult.Warnings 非空
- `TestRunQuery_TransformValidField_NoWarning` — transform=`{type}`，QueryResult.Warnings 为空（或仅保留已有文件跳过警告）
- `TestRunQueryWithTimeRange_TransformAllNull_Warning` — 同上，带 TimeRange

### Implementation

文件：`internal/mcp/query/query.go`

- `RunQuery`：在 `StreamFiles` 返回后，若 `transform != ""` 且 `isAllNullOrEmpty(result.Entries)`，追加告警
- `RunQueryWithTimeRange`：同上
- 复用 Phase 1 的 `isAllNullOrEmpty`（跨包调用或在 `internal/mcp/query/` 包内复制一份轻量版）

### DoD

```
go test ./internal/mcp/query/...
go test ./...
```
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Proposal + Plan APPROVED. Ready for implementation.

claimed: 2026-06-23T15:50:00Z

Completed: 2026-06-23T15:55:00Z

Phase 1: added Warnings []string to Stage2Result, isAllNullOrEmpty helper, warning injection in ExecuteStage2Query.
Phase 2: HandleExecuteStage2Query now includes warnings key in JSON response.
Phase 3: RunQuery and RunQueryWithTimeRange warn when transform produces all-null entries.
13 new tests. All pass.
<!-- SECTION:NOTES:END -->
