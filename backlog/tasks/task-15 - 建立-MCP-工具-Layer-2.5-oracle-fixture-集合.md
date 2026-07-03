---
id: TASK-15
title: 建立 MCP 工具 Layer 2.5 oracle fixture 集合
status: 'Basic: Done'
assignee: []
created_date: '2026-06-24 12:12'
updated_date: '2026-06-24 12:29'
labels:
  - 'kind:basic'
dependencies: []
ordinal: 2000
---

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
# Plan: 建立 MCP 工具 Layer 2.5 oracle fixture 集合

## Context
meta-cc's Go unit tests verify internal logic but not MCP tool behavior from Claude's calling perspective. Layer 2.5 oracle framework (BAIME Exp-D) fills this gap: fixture-based behavioral testing with Haiku as oracle, targeting ≥0.85 accuracy on Class A decisions.

## Phase 1: 调研现有 MCP 工具接口和测试结构
Read /home/yale/work/meta-cc/docs/guides/mcp.md and /home/yale/work/meta-cc/docs/guides/mcp-query-tools.md to understand the parameter schemas for query_session_signals and query_session_content. Read /home/yale/work/meta-cc/internal/ to understand output structures. List what field sets each tool returns per type parameter.
### DoD
- [ ] `test -f /home/yale/work/meta-cc/docs/guides/mcp-query-tools.md`
- [ ] `test -f /home/yale/work/meta-cc/internal/query/signals.go || find /home/yale/work/meta-cc/internal -name '*.go' | xargs grep -l 'query_session_signals' | grep -q .`

## Phase 2: 编写 λ-spec 文档
Create /home/yale/work/meta-cc/docs/testing/layer25-spec.md with:
- λ-spec for query_session_signals(type=errors): input schema, expected output fields (timestamp, tool, error), ordering rule (descending by timestamp)
- λ-spec for query_session_content(role=user, pattern=X): input schema, expected output fields (timestamp, role, content match), filter semantics
- Decision classes: Class A = "does output contain required fields?" (binary, Haiku oracle)
### DoD
- [ ] `test -f /home/yale/work/meta-cc/docs/testing/layer25-spec.md`
- [ ] `grep -q 'query_session_signals' /home/yale/work/meta-cc/docs/testing/layer25-spec.md`
- [ ] `grep -q 'query_session_content' /home/yale/work/meta-cc/docs/testing/layer25-spec.md`

## Phase 3: 创建 fixture 集合（≥10 个）
Create /home/yale/work/meta-cc/docs/testing/fixtures/ directory with fixture files:
- signals-errors-basic.md: query_session_signals(type=errors) → expected fields
- signals-tokens-basic.md: query_session_signals(type=tokens) → expected stats fields
- signals-tool-stats-basic.md: query_session_signals(type=tool_stats) → expected fields
- content-user-basic.md: query_session_content(role=user) → expected record fields
- content-assistant-pattern.md: query_session_content(role=assistant, pattern=X) → pattern matching behavior
- (5+ more covering edge cases: empty results, limit parameter, file_ref mode)
Each fixture: input parameters + expected output structure description + pass/fail criteria as assertions.
### DoD
- [ ] `find /home/yale/work/meta-cc/docs/testing/fixtures -name '*.md' | wc -l | awk '{if($1>=10) exit 0; else exit 1}'`
- [ ] `grep -rl 'query_session_signals' /home/yale/work/meta-cc/docs/testing/fixtures | grep -q .`
- [ ] `grep -rl 'query_session_content' /home/yale/work/meta-cc/docs/testing/fixtures | grep -q .`

## Phase 4: 编写 Haiku oracle 测试脚本框架
Create /home/yale/work/meta-cc/docs/testing/oracle-runner.md documenting:
- How to run each fixture against a live MCP session
- Haiku oracle prompt template: given tool output JSON, assess if required fields are present and ordering is correct (yes/no + confidence)
- Scoring: pass if oracle says yes, fail otherwise; report overall accuracy
- Threshold: ≥0.85 overall pass rate required
### DoD
- [ ] `test -f /home/yale/work/meta-cc/docs/testing/oracle-runner.md`
- [ ] `grep -q '0.85' /home/yale/work/meta-cc/docs/testing/oracle-runner.md`

## Constraints
- No Go source code changes; this is documentation and test specification only
- Fixtures are markdown specification files, not executable scripts (execution requires live MCP session)
- oracle-runner.md documents the process; actual oracle calls happen manually or via future automation

## Acceptance Gate
- [ ] `find /home/yale/work/meta-cc/docs/testing/fixtures -name '*.md' | wc -l | awk '{if($1>=10) exit 0; else exit 1}'`
- [ ] `test -f /home/yale/work/meta-cc/docs/testing/layer25-spec.md`
- [ ] `test -f /home/yale/work/meta-cc/docs/testing/oracle-runner.md`
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
cap:propose=approved

claimed: 2026-06-24T12:20:34Z

Phase 1 ✓ 2026-06-24T00:00:00Z — Read mcp-query-tools.md and mcp.md; traced consolidated_handlers.go to understand query_session_signals and query_session_content routing and output structures. DoD #1: PASS — mcp-query-tools.md exists. DoD #2: PASS — query_session_signals found in consolidated_handlers.go.

Phase 2 ✓ 2026-06-24T00:00:00Z — Created docs/testing/layer25-spec.md with λ-specs for query_session_signals (all 5 types) and query_session_content (all 4 roles), including decision classes A/B/C, parameter schemas, required output fields, and hybrid output mode spec. DoD #1: PASS — layer25-spec.md exists. DoD #2: PASS — query_session_signals present. DoD #3: PASS — query_session_content present.

Phase 3 ✓ 2026-06-24T00:00:00Z — Created 14 fixture files in docs/testing/fixtures/: 5 signals fixtures (errors, tokens, tool_stats, timestamps, system_errors, provider-all), 5 content fixtures (user-basic, user-pattern, assistant-basic, tool-blocks, all-flow), 3 edge-case fixtures (empty-results, limit-parameter, file-ref-mode). DoD #1: PASS — 14 >= 10 fixtures. DoD #2: PASS — query_session_signals in fixtures. DoD #3: PASS — query_session_content in fixtures.

Phase 4 ✓ 2026-06-24T00:00:00Z — Created docs/testing/oracle-runner.md with Haiku oracle prompt template, step-by-step fixture execution protocol, scoring formula (overall_accuracy >= 0.85), fixture index table (14 fixtures, ~57 assertions), and reporting format. DoD #1: PASS — oracle-runner.md exists. DoD #2: PASS — 0.85 threshold present.

## Execution Summary
Result: Done
Commit: 4d41046 — docs(testing): add Layer 2.5 oracle fixture suite for MCP tool behavioral testing
Files created: docs/testing/layer25-spec.md, docs/testing/oracle-runner.md, docs/testing/fixtures/ (14 fixtures)

workerLoop DoD #1: PASS — find /home/yale/work/meta-cc/docs/testing/fixtures -name '*.md' | wc -l | awk '{if($1>=10) exit 0; else exit 1}'

workerLoop DoD #2: PASS — test -f /home/yale/work/meta-cc/docs/testing/layer25-spec.md

workerLoop DoD #3: PASS — test -f /home/yale/work/meta-cc/docs/testing/oracle-runner.md

WARNING: agent-summary missing

Completed: 2026-06-24T12:29:36Z
<!-- SECTION:NOTES:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 find /home/yale/work/meta-cc/docs/testing/fixtures -name '*.md' | wc -l | awk '{if($1>=10) exit 0; else exit 1}'
- [ ] #2 test -f /home/yale/work/meta-cc/docs/testing/layer25-spec.md
- [ ] #3 test -f /home/yale/work/meta-cc/docs/testing/oracle-runner.md
<!-- DOD:END -->
