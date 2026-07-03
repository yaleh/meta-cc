---
id: TASK-17
title: OCA Codify：tool_stats 驱动高频 jq 模式提炼为 MCP 工具
status: 'Basic: Backlog'
assignee: []
created_date: '2026-06-24 12:13'
updated_date: '2026-06-24 12:14'
labels:
  - 'kind:basic'
dependencies: []
ordinal: 9000
---

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
# Plan: OCA Codify — tool_stats 驱动高频 jq 模式提炼为 MCP 工具

## Context
execute_stage2_query exposes raw jq — powerful but requires users to know JSONL schema. OCA Codify principle: when a jq pattern recurs ≥5 times across sessions, it signals a compression opportunity (MDL). This task runs the Observe phase (tool_stats analysis) and one Codify step (add a parameter or new tool for the top pattern).

## Phase 1: Observe — 运行 tool_stats 分析并生成报告
Use MCP tool query_session_signals(type=tool_stats) to get call frequency data for execute_stage2_query across the project's session history. Then use get_session_directory + inspect_session_files + execute_stage2_query to extract the actual filter/transform arguments from execute_stage2_query calls. Write findings to docs/tasks/oca-codify-patterns-report.md including:
- Total execute_stage2_query call count
- Top-5 most-repeated filter patterns (with occurrence count)
- Recommendation: which pattern to codify first
### DoD
- [ ] `test -f /home/yale/work/meta-cc/docs/tasks/oca-codify-patterns-report.md`
- [ ] `grep -q 'execute_stage2_query' /home/yale/work/meta-cc/docs/tasks/oca-codify-patterns-report.md`
- [ ] `grep -q 'Top-5\|top-5\|top 5' /home/yale/work/meta-cc/docs/tasks/oca-codify-patterns-report.md`

## Phase 2: Design — 设计新参数或工具的接口
Based on Phase 1 report, design the interface for codifying the top pattern. Options:
- If pattern is time-window filtering: add `since` / `until` params to get_timeline
- If pattern is error aggregation: add aggregation mode to query_session_signals
- If pattern is GCL event extraction: add type=gcl_gates to query_session_signals
Write the interface design as an ADR (Architecture Decision Record) at docs/architecture/adr/ADR-007-codify-top-jq-pattern.md following the existing ADR format in docs/architecture/adr/template.md.
### DoD
- [ ] `test -f /home/yale/work/meta-cc/docs/architecture/adr/ADR-007-codify-top-jq-pattern.md`
- [ ] `grep -q 'Decision\|decision' /home/yale/work/meta-cc/docs/architecture/adr/ADR-007-codify-top-jq-pattern.md`

## Phase 3: Codify — 实现新参数或工具
Implement the designed interface in Go (cmd/mcp-server/ and internal/). Follow TDD: write test first, then implementation. Key files:
- cmd/mcp-server/handlers_stage1.go — add new handler or extend existing handler
- cmd/mcp-server/tools.go — register new parameter in tool schema
- internal/ — add business logic as needed
Run `make dev` after implementation.
### DoD
- [ ] `make -C /home/yale/work/meta-cc dev`
- [ ] `grep -rq 'Test.*Codify\|Test.*OCA\|Test.*Since\|Test.*Until\|Test.*Timeline\|Test.*gcl\|Test.*Aggregate' /home/yale/work/meta-cc/cmd/mcp-server`

## Phase 4: 更新 CLAUDE.md 决策树
Update the "Which MCP query tool should I use?" decision tree in /home/yale/work/meta-cc/CLAUDE.md to include the new parameter/tool and when to use it instead of execute_stage2_query. Run `make commit` to validate.
### DoD
- [ ] `make -C /home/yale/work/meta-cc commit`
- [ ] `grep -q 'execute_stage2_query' /home/yale/work/meta-cc/CLAUDE.md`

## Constraints
- Phase 3 code changes: maximum 200 lines
- Phase 1 (Observe) must complete before Phase 2 (Design) — the design depends on actual patterns found
- If no pattern occurs ≥5 times, document this as "insufficient data" and propose the most promising candidate anyway
- Do not add more than 1 new tool or 2 new parameters in Phase 3

## Acceptance Gate
- [ ] `test -f /home/yale/work/meta-cc/docs/tasks/oca-codify-patterns-report.md`
- [ ] `test -f /home/yale/work/meta-cc/docs/architecture/adr/ADR-007-codify-top-jq-pattern.md`
- [ ] `make -C /home/yale/work/meta-cc commit`
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
cap:propose=approved
<!-- SECTION:NOTES:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 test -f /home/yale/work/meta-cc/docs/tasks/oca-codify-patterns-report.md
- [ ] #2 find /home/yale/work/meta-cc/docs/architecture/adr -name '*codify*' | grep -q .
- [ ] #3 make -C /home/yale/work/meta-cc commit
<!-- DOD:END -->
