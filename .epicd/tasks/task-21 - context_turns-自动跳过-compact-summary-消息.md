---
id: TASK-21
title: context_turns 自动跳过 compact summary 消息
assignee: []
created_date: '2026-07-14 03:52'
updated_date: '2026-07-15 02:51'
labels:
  - 'area:mcp'
  - 'area:query'
dependencies: []
priority: medium
ordinal: 12000
pipeline_id: execution
phase: done
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
compact summary 是 Claude Code 自动注入的上下文压缩边界消息（`isCompactSummary: true`），在会话的 JSONL 文件中以 user 类型条目出现，正文为数千字乃至数万 token 的历史摘要。

当 `context_turns > 0` 时，`ExpandContextTurns`（`internal/mcp/filters/filters.go`）会将该会话文件里的所有条目按顺序排列，计算前后 N 个位置的窗口，再将窗口内所有条目拼入结果。compact summary 条目因体积极大（单条可达数万 token），且对理解上下文没有实际帮助，被当作普通轮次拉入会导致结果体积暴增。

修复目标：在 `ExpandContextTurns` 中，构建 session 轮次序列（`uuidToIndex`）和最终输出时，自动跳过 `isCompactSummary=true` 的条目——它们既不计入轮次计数，也不出现在返回结果中。被用户直接匹配的条目本身如果是 compact summary，同样被跳过（不输出）。

**Non-Goals**
- 不新增任何用户可见的参数或开关；行为对 compact summary 的跳过是无条件的（TASK-22 新增 exclude_compact_summaries 参数是独立任务）
- 不修改 parser 层或其他 filter 函数
- 不改变非 compact summary 条目的任何现有行为
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 ExpandContextTurns 跳过 compact summary 不出现在结果中：TestExpandContextTurns_SkipsCompactSummary 通过，断言 result 不含任何 isCompactSummary=true 的条目；verify: go test ./internal/mcp/filters/ -run TestExpandContextTurns_SkipsCompactSummary
- [ ] #2 非 compact summary 条目行为不变：现有 TestExpandContextTurns_Basic / WindowClampAtStart / WindowClampAtEnd / OverlappingWindows / MultipleSessionOrder 全部继续通过；verify: cd /home/yale/work/meta-cc && go test ./internal/mcp/filters/ -run TestExpandContextTurns
- [ ] #3 被匹配条目本身是 compact summary 时也不输出：TestExpandContextTurns_MatchedCompactSummarySkipped 通过，rawData 含一条 compact summary，result 为空；verify: cd /home/yale/work/meta-cc && go test ./internal/mcp/filters/ -run TestExpandContextTurns_MatchedCompactSummarySkipped
- [ ] #4 全套测试通过：verify: cd /home/yale/work/meta-cc && make commit
- [ ] #5 compact summary 不成为 context 锚点（不占有效轮次）——session=[u0,u1(compact),u2,u3,u4]，匹配 u2，N=1：由于 u1 不进 uuidToIndex，u2 物理 idx=2，window=[turns[1],turns[2],turns[3]]，emit 跳过 compact turns[1]，result 含 [u2,u3]（2条而非4条）；TestExpandContextTurns_CompactSummaryNotCountedAsTurn 断言 len==2 且不含 u1；verify: go test ./internal/mcp/filters/ -run TestExpandContextTurns_CompactSummaryNotCountedAsTurn
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
# Plan: context_turns 自动跳过 compact summary 消息

## Phase A: 新增测试（TDD — 先写测试，令其 Red）

### Tests (write first)
File: `internal/mcp/filters/filters_test.go`

新增4个测试函数（追加至文件末尾）：

1. **TestExpandContextTurns_SkipsCompactSummary**  
   session=[u0, u1(isCompactSummary=true), u2, u3, u4]，匹配 u2，N=2  
   断言 result 中不含任何 isCompactSummary=true 的条目

2. **TestExpandContextTurns_CompactSummaryNotCountedAsTurn**  
   session=[u0, u1(compact), u2, u3, u4]，匹配 u2，N=1  
   u1 不进 uuidToIndex，u2 物理 idx=2，window 覆盖 turns[1..3]，emit 跳过 compact turns[1]  
   断言 len(result)==2，result 含 u2(context=false) 和 u3(context=true)，不含 u1

3. **TestExpandContextTurns_MatchedCompactSummarySkipped**  
   session=[u0(compact)], rawData 仅含 u0  
   断言 len(result)==0（被匹配的 compact summary 自身也被跳过）

4. **TestExpandContextTurns_AllCompactSummarySession**  
   session=[u0(compact), u1(compact), u2(compact)]，rawData 匹配 u1  
   断言 len(result)==0

### Implementation
（Phase B 完成之前，这些测试应当 Fail）

### DoD
- [ ] `go test ./internal/mcp/filters/ -run TestExpandContextTurns_SkipsCompactSummary 2>&1 | grep -E 'FAIL|undefined'`  （Red 阶段预期失败）

## Phase B: 实现 ExpandContextTurns 跳过逻辑

### Tests (write first)
（Phase A 已写，此处复用）

### Implementation
File: `internal/mcp/filters/filters.go`，函数 `ExpandContextTurns`

**修改1**（约 filters.go:249）：在构建 `uuidToIndex` 的循环中，跳过 compact summary：
```go
// 在「if uuid != "" { uuidToIndex[uuid] = i }」之前加：
isCompact, _ := obj["isCompactSummary"].(bool)
if isCompact {
    continue
}
```

**修改2**（约 filters.go:298）：在 emit 循环中，跳过 compact summary：
```go
// 在「uuid, _ := turnObj["uuid"].(string)」之后加：
isCompact, _ := turnObj["isCompactSummary"].(bool)
if isCompact {
    continue
}
```

修改2确保即便 compact summary 的物理 index 落在 windowSet 范围内，也不会被输出。

### DoD
- [ ] `cd /home/yale/work/meta-cc && go test ./internal/mcp/filters/ -run TestExpandContextTurns_SkipsCompactSummary`
- [ ] `cd /home/yale/work/meta-cc && go test ./internal/mcp/filters/ -run TestExpandContextTurns_CompactSummaryNotCountedAsTurn`
- [ ] `cd /home/yale/work/meta-cc && go test ./internal/mcp/filters/ -run TestExpandContextTurns_MatchedCompactSummarySkipped`
- [ ] `cd /home/yale/work/meta-cc && go test ./internal/mcp/filters/ -run TestExpandContextTurns_AllCompactSummarySession`
- [ ] `cd /home/yale/work/meta-cc && go test ./internal/mcp/filters/ -run TestExpandContextTurns`

## Constraints
- 每 Stage 修改不超过200行
- 不修改 parser 层或其他 filter 函数
- 不新增公开 API 参数
- 两处修改合计约10行代码（修改1约3行，修改2约3行）

## Acceptance Gate
- [ ] `cd /home/yale/work/meta-cc && make commit`
- [ ] `cd /home/yale/work/meta-cc && go test ./internal/mcp/filters/ -run TestExpandContextTurns -v 2>&1 | grep -E 'PASS|ok'`
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
authoring/draft self-review: Approved after 1 round. All 5 ACs carry executable verify commands; adversarial check confirms no scenario where all ACs pass while the goal (compact summaries absent from context_turns results) is unmet.

authoring/refining review: APPROVED after 2 iterations. Iteration 1 caught incorrect test expectation in AC#2 (len==3 was wrong given physical-index-based window computation; corrected to len==2). Iteration 2: all checklist criteria pass — goal coverage complete, TDD structure correct, Acceptance Gate executable, phase ordering valid, file paths confirmed.
<!-- SECTION:NOTES:END -->
