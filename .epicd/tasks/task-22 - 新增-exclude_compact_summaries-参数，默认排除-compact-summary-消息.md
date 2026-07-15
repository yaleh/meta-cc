---
id: TASK-22
title: 新增 exclude_compact_summaries 参数，默认排除 compact summary 消息
assignee: []
created_date: '2026-07-14 03:53'
updated_date: '2026-07-15 02:34'
labels:
  - 'area:mcp'
  - 'area:query'
dependencies: []
priority: medium
ordinal: 13000
pipeline_id: authoring
phase: backlog
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
compact summary 消息（`isCompactSummary: true`）是 Claude Code 自动注入的上下文压缩记录。这类条目正文体积巨大——往往包含数万 token 的历史摘要——且因为内含大量历史文本，几乎能命中任意 regex pattern 搜索。当前 `query_session_content` 的 role=user 和 role=all 路径均未对其过滤，导致搜索结果充斥噪声，真正的用户输入被淹没；用户只能自行在 jq 中写 `select(.isCompactSummary != true)` 才能规避，体验很差。

新增 `exclude_compact_summaries` boolean 参数（默认 `true`），在两个层面同时生效：
1. 在 jqFilter 阶段将 compact summary 从 pattern/contains 匹配的候选集排除（role=user 和 role=all 均适用）
2. 在 `context_turns` 扩展阶段，从 `ExpandContextTurns` 加载的 session 轮次中过滤掉 compact summary，使其不出现在 context 结果中

`exclude_compact_summaries=false` 可恢复当前行为，适用于需要检索压缩摘要内容本身的场景。

本任务独立于 TASK-21（context_turns 内部自动跳过逻辑）；不假设 TASK-21 已实现。
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 role=user default excludes compact summaries: go test ./internal/mcp/executor/ -run TestHandleQueryUserMessages_ExcludeCompactSummaries passes and asserts zero entries with isCompactSummary=true in result
- [ ] #2 role=all default excludes compact summaries: go test ./internal/mcp/executor/ -run TestHandleQueryConversationFlow_ExcludeCompactSummaries passes and asserts zero entries with isCompactSummary=true in result
- [ ] #3 exclude_compact_summaries=false restores compact summaries: go test ./internal/mcp/executor/ -run TestHandleQueryUserMessages_ExcludeCompactSummaries_FalseRestores passes and asserts isCompactSummary=true entries appear in result
- [ ] #4 context_turns excludes compact summaries when flag=true: go test ./internal/mcp/filters/ -run TestExpandContextTurns_ExcludeCompactSummaries passes and asserts no isCompactSummary=true entry appears in expanded context output
- [ ] #5 parameter is in MCP schema: grep 'exclude_compact_summaries' internal/mcp/tools/tools.go exits 0
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
# Plan: Add exclude_compact_summaries parameter to query_session_content

## Phase A: jqFilter exclusion in handleQueryUserMessages and handleQueryConversationFlow

### Tests (write first)
File: `internal/mcp/executor/handlers_test.go` (append to existing)
- `TestHandleQueryUserMessages_ExcludeCompactSummaries` — default true: record with `isCompactSummary=true` must NOT appear; plain user record must appear
- `TestHandleQueryUserMessages_ExcludeCompactSummaries_FalseRestores` — false: record with `isCompactSummary=true` MUST appear
- `TestHandleQueryConversationFlow_ExcludeCompactSummaries` — default true: record with `isCompactSummary=true` must NOT appear

### Implementation
File: `internal/mcp/executor/handlers.go`
- In `handleQueryUserMessages`: add `excludeCompact := GetBoolParam(args, "exclude_compact_summaries", true)` and when true append `| select(.isCompactSummary != true)` to `jqFilter` (mirror the `excludeSystem` block at lines 60-62)
- In `handleQueryConversationFlow`: same extraction and same jq clause appended

### DoD
- `go test ./internal/mcp/executor/ -run TestHandleQueryUserMessages_ExcludeCompactSummaries`
- `go test ./internal/mcp/executor/ -run TestHandleQueryUserMessages_ExcludeCompactSummaries_FalseRestores`
- `go test ./internal/mcp/executor/ -run TestHandleQueryConversationFlow_ExcludeCompactSummaries`
- `go test ./internal/mcp/executor/`

## Phase B: ExcludeCompactSummaries in PipelineConfig and ExpandContextTurns

### Tests (write first)
File: `internal/mcp/filters/filters_test.go` (append to existing)
- `TestExpandContextTurns_ExcludeCompactSummaries` — session=[u0, u1(isCompactSummary=true), u2, u3, u4]; match u2, N=2, excludeCompact=true; assert result contains zero entries with `isCompactSummary=true`; u1 must not appear even as context turn
Also update all 10 existing `ExpandContextTurns(...)` call sites in `filters_test.go` to pass the new 4th bool parameter as `false` (preserving existing test behavior)

### Implementation
File: `internal/mcp/filters/filters.go`
- Change signature to: `func ExpandContextTurns(rawData []interface{}, N int, baseDir string, excludeCompactSummaries bool) ([]interface{}, error)`
- After loading turns via `loadTurnsForSession`, when `excludeCompactSummaries=true` filter out entries where `obj["isCompactSummary"] == true` before building `uuidToIndex` and the result window

File: `internal/mcp/pipeline/pipeline.go`
- Add `ExcludeCompactSummaries bool` field to `PipelineConfig`
- Update call: `filterspkg.ExpandContextTurns(parsedData, pc.ContextTurns, baseDir, pc.ExcludeCompactSummaries)`

File: `internal/mcp/executor/executor.go`
- In `NewToolPipelineConfig`: add `ExcludeCompactSummaries: GetBoolParam(args, "exclude_compact_summaries", true)`

### DoD
- `go test ./internal/mcp/filters/ -run TestExpandContextTurns_ExcludeCompactSummaries`
- `go test ./internal/mcp/filters/`
- `go test ./internal/mcp/pipeline/`

## Phase C: Add parameter to MCP tool schema

### Tests (write first)
No named test; AC #5 is the grep verification command below.

### Implementation
File: `internal/mcp/tools/tools.go`
- Add to the `query_session_content` property map: `"exclude_compact_summaries": {Type: "boolean", Description: "Exclude compact summary messages (isCompactSummary=true) from results and context_turns. Default: true. Pass false to include them (e.g. to search the summaries themselves)."}`

### DoD
- `grep 'exclude_compact_summaries' internal/mcp/tools/tools.go`
- `go test ./internal/mcp/tools/`

## Constraints
- TASK-21 is independent; do not add any unconditional skip inside ExpandContextTurns.
- Only role=user and role=all paths are modified; role=assistant and role=tool are unchanged (compact summaries are type=user entries).
- All 10 existing ExpandContextTurns call sites in filters_test.go must be updated to pass false as the new 4th argument.
- Default is true (safe default; excludes noise by default).

## Acceptance Gate
- `make commit`
- `grep 'exclude_compact_summaries' internal/mcp/tools/tools.go`
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
authoring/draft self-review: Approved after 1 round. All criteria passed: motivation explains why (noise/pollution from compact summaries with no existing exclusion), goals are concretely verifiable (named test cases + grep), approach confirmed by codebase search (isCompactSummary is top-level field, exclude_system_messages is proven analog pattern, ExpandContextTurns loads turns from loadTurnsForSession), non-goals stated, no contradictions, adversarial AC check found no gap.

authoring/refining review: APPROVED after 2 iterations. Iteration 1 caught a file-path typo and, more importantly, identified that ExpandContextTurns has 10 existing call sites in filters_test.go that would break on a signature change — plan revised in iteration 2 to explicitly require updating all 10 existing call sites to pass false as the new 4th bool. All review criteria satisfied: goal coverage complete across 3 phases, TDD structure correct, acceptance gate is make commit, all DoD items are executable shell commands, phase ordering is independent (no forward dependencies), scope is contained to stated goals, all file paths confirmed.
<!-- SECTION:NOTES:END -->
