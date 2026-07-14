---
id: TASK-14
title: 引入 GCL 前提台账标注格式到 meta-cc 工作流
status: 'Basic: Done'
assignee: []
created_date: '2026-06-24 12:12'
updated_date: '2026-06-24 12:23'
labels:
  - 'kind:basic'
dependencies: []
ordinal: 1000
---

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
# Plan: 引入 GCL 前提台账标注格式到 meta-cc 工作流

## Context
meta-cc captures session traces but has no semantic representation for gate events. Introducing a structured [E]/[C]/[H] annotation convention enables GCL baseline measurement via existing query_session_content — no new tools needed.

## Phase 1: 定义标注格式规范
Read /home/yale/work/meta-cc/CLAUDE.md and /home/yale/work/meta-cc/docs/guides/mcp.md to understand current workflow conventions. Then append a "## GCL Gate Annotation Format" section to /home/yale/work/meta-cc/docs/guides/mcp.md (or create docs/guides/gcl-annotation.md if mcp.md is too large) documenting:
- Format: `[GCL-gate] <decision-name>` header line
- Per-criterion lines: `[E|C|H] <criterion>: <evidence-description>`
- Summary line: `GCL-self-report: E=n C=n H=n`
- Example: freshnessCheck gate with 2E, 1C, 1H items
### DoD
- [ ] `grep -q 'GCL-self-report' /home/yale/work/meta-cc/docs/guides/mcp.md || grep -q 'GCL-self-report' /home/yale/work/meta-cc/docs/guides/gcl-annotation.md`
- [ ] `grep -q 'GCL-gate' /home/yale/work/meta-cc/docs/guides/mcp.md || grep -q 'GCL-gate' /home/yale/work/meta-cc/docs/guides/gcl-annotation.md`

## Phase 2: 提供示例查询调用
Add a "## Querying GCL Events" section to the same doc with a concrete query_session_content call example:
```
query_session_content({role: "assistant", pattern: "GCL-self-report"})
```
And a stage2 jq transform that extracts E/C/H counts from matched content.
### DoD
- [ ] `grep -q 'query_session_content' /home/yale/work/meta-cc/docs/guides/mcp.md || grep -q 'query_session_content' /home/yale/work/meta-cc/docs/guides/gcl-annotation.md`
- [ ] `grep -q 'GCL-self-report' /home/yale/work/meta-cc/docs/guides/mcp.md || grep -q 'GCL-self-report' /home/yale/work/meta-cc/docs/guides/gcl-annotation.md`

## Phase 3: 验证格式可被 query_session_content 解析
Manually write 3 synthetic gate event examples (as markdown fixtures in docs/guides/gcl-annotation.md or a fixtures/ subdirectory) that conform to the format. Run query_session_content with pattern "GCL-self-report" and verify it would match.
### DoD
- [ ] `grep -c '\[GCL-gate\]' /home/yale/work/meta-cc/docs/guides/gcl-annotation.md | grep -q '^[3-9]'`

## Constraints
- Do not implement new MCP tools or modify Go source code
- Annotation format must be parseable by existing query_session_content pattern matching
- Documentation only; no runtime changes

## Acceptance Gate
- [ ] `grep -q 'GCL-self-report' /home/yale/work/meta-cc/docs/guides/mcp.md || grep -q 'GCL-self-report' /home/yale/work/meta-cc/docs/guides/gcl-annotation.md`
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
cap:propose=approved

claimed: 2026-06-24T12:20:25Z

Phase 1 ✓ 2026-06-24T00:00:00Z — Created docs/guides/gcl-annotation.md with [GCL-gate]/[E]/[C]/[H] format spec and examples

Phase 2 ✓ 2026-06-24T00:00:00Z — Added Querying GCL Events section with query_session_content calls and two-stage jq transforms

Phase 3 ✓ 2026-06-24T00:00:00Z — Added 3 synthetic gate fixture examples (freshnessCheck, readyToMerge, phaseGate-implementation)

DoD #1: PASS — grep -q 'GCL-self-report' gcl-annotation.md

DoD #2: PASS — grep -q 'GCL-gate' gcl-annotation.md

DoD #3: PASS — grep -q 'query_session_content' gcl-annotation.md

DoD #4: PASS — grep -c '[GCL-gate]' gcl-annotation.md returns 6 (≥3)

## Execution Summary
Result: Done
Commit: e2095bf

WARNING: agent-summary missing

Completed: 2026-06-24T12:23:48Z
<!-- SECTION:NOTES:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 grep -q 'GCL-self-report' /home/yale/work/meta-cc/docs/guides/mcp.md || grep -q 'GCL-self-report' /home/yale/work/meta-cc/docs/guides/gcl-annotation.md
- [ ] #2 grep -q 'GCL-gate' /home/yale/work/meta-cc/docs/guides/mcp.md || grep -q 'GCL-gate' /home/yale/work/meta-cc/docs/guides/gcl-annotation.md
- [ ] #3 grep -q 'query_session_content' /home/yale/work/meta-cc/docs/guides/mcp.md || grep -q 'query_session_content' /home/yale/work/meta-cc/docs/guides/gcl-annotation.md
<!-- DOD:END -->
