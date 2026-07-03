---
id: TASK-7
title: 'MCP Tool Budget: Consolidate query_* tools (10→3)'
status: 'Basic: Done'
assignee: []
created_date: '2026-06-23 06:24'
updated_date: '2026-06-23 07:16'
labels:
  - 'kind:basic'
dependencies: []
references:
  - docs/proposals/proposal-mcp-tool-budget.md
  - internal/mcp/tools/tools.go
  - internal/mcp/executor/query_handlers.go
  - internal/mcp/executor/handlers.go
priority: high
ordinal: 1000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Replace the 10 structurally-fragmented `query_*` MCP tools with 3 semantically-grouped tools (`query_session_content`, `query_session_signals`, `query_file_activity`), reducing schema token overhead and LLM selection noise. This is a breaking API change managed via a deprecation + alias window before hard removal.
<!-- SECTION:DESCRIPTION:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 go test ./... passes with no regressions
- [ ] #2 query_session_content, query_session_signals, query_file_activity are registered and functional
- [ ] #3 All 10 old query_* tool registrations removed from tools.go
- [ ] #4 Every old tool call has a verified equivalent in the new API (migration mapping tested)
- [ ] #5 docs/guides/mcp-query-tools.md rewritten to document new API only
- [ ] #6 No references to old tool names remain in internal/ source or *_test.go files
<!-- DOD:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 go test ./... passes after each phase
- [ ] #2 query_session_content handles role=user|assistant|tool|all with correct output
- [ ] #3 query_session_signals handles type=errors|tokens|system_errors|timestamps|tool_stats with correct output
- [ ] #4 query_file_activity handles type=snapshots with correct output
- [ ] #5 All 10 old query_* tools removed from BuildTools() output in Phase D
- [ ] #6 docs/guides/mcp-query-tools.md documents new API only (no old tool names)
- [ ] #7 No test file references old tool names after Phase D
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Proposal

### Background

meta-cc currently exposes 10 `query_*` MCP tools that all perform the same fundamental operation — filter session JSONL by event type and return matching records. The only difference between them is which event type they default to. They share the same data source, pagination mechanism, session-scope parameter, and hybrid output mode.

Exposing them as 10 separate tools creates two measurable problems:

1. **Token overhead**: 10 near-identical schemas consume redundant token budget at every session start. Each tool definition is ~200–400 tokens; 10 tools = ~2,000–4,000 tokens wasted on schema duplication.
2. **Selection noise**: The LLM must choose among `query_tool_errors`, `query_system_errors`, `query_tool_blocks`, and `query_conversation_flow` for what is conceptually one query operation, increasing the probability of sub-optimal tool selection.

Session history analysis shows 62% of all tools (13 of 21) have zero recorded calls in this project's history — usage is concentrated in ~3 tools. This confirms that fragmentation is not being justified by breadth of use.

The root cause is early-phase API design: each new signal type became a new tool rather than a parameter of one flexible query tool.

### Goals

1. Reduce `query_*` tool count from 10 to 3 named tools: `query_session_content`, `query_session_signals`, `query_file_activity`.
2. Preserve all existing query capability: every old tool call has an exact new equivalent via `type` or `role` parameter.
3. Maintain backward compatibility during transition: old tool names continue to work for at least one release cycle via alias/forwarding in the handler registry.
4. Update all integration tests to use the new tool names before removing aliases.
5. Update `docs/guides/mcp-query-tools.md` to document only the new API.

### Proposed Approach

**Phase A — Add new tools (no removal yet)**

Register `query_session_content`, `query_session_signals`, and `query_file_activity` in `internal/mcp/tools/tools.go`. Implement their handlers in `internal/mcp/executor/query_handlers.go` by routing the `role`/`type` parameter to the existing per-type handler logic. All 10 old tools remain registered during this phase.

**Phase B — Implement routing handlers**

Each new handler receives the `role` or `type` parameter and dispatches to the appropriate existing filter logic. No new filtering logic is written — the handlers are thin routers over the already-tested per-type code paths.

**Phase C — Alias layer: old tool names forward to new handlers, marked deprecated**

Update the description strings of the 10 old tools in `tools.go` to include `[DEPRECATED: use query_session_content / query_session_signals / query_file_activity]`. Update their handler dispatch to forward to the new handler functions. Integration tests at this point should pass with either old or new tool names.

**Phase D — Remove old registrations and update tests + docs**

After verifying alias parity, remove the 10 old tool registrations from `tools.go`. Update all integration tests to use new tool names and parameter forms. Rewrite `docs/guides/mcp-query-tools.md`.

### Trade-offs

| Trade-off | Assessment |
|---|---|
| Breaking API change | Mitigated by alias phase (Phase C); hard removal deferred to Phase D |
| Slash command dependencies | CLAUDE.md FAQ examples reference old tool names — must be updated in Phase D |
| BAIME agent prompts | Any agent prompt using old tool names will still work during alias phase; must update before alias removal |
| Test update burden | ~10 integration test call sites to rename; mechanical, low risk |
| Complexity of routing handler | Low — parameter dispatch, not new filtering logic |

**Migration mapping** (old → new):

| Old | New |
|---|---|
| `query_tool_errors` | `query_session_signals({type:"errors"})` |
| `query_token_usage` | `query_session_signals({type:"tokens"})` |
| `query_system_errors` | `query_session_signals({type:"system_errors"})` |
| `query_timestamps` | `query_session_signals({type:"timestamps"})` |
| `query_tools` | `query_session_signals({type:"tool_stats"})` |
| `query_summaries` | `query_session_content({role:"assistant", contains:"## Summary"})` |
| `query_tool_blocks` | `query_session_content({role:"tool"})` |
| `query_user_messages` | `query_session_content({role:"user"})` |
| `query_conversation_flow` | `query_session_content({role:"all"})` |
| `query_file_snapshots` | `query_file_activity({type:"snapshots"})` |

---

## Implementation Plan

### Phase A: Define new consolidated tool schemas (no old tool removal)

**Goal**: Register `query_session_content`, `query_session_signals`, `query_file_activity` in `internal/mcp/tools/tools.go`. Old tools remain untouched.

**Tests first** (`internal/mcp/tools/tools_test.go`):
- New tools appear in the tool list returned by `BuildTools()`
- New tools have required parameters `role`/`type`/`files` with correct enum values in schema
- Old tools are still present (no regression)

**Implementation** (`internal/mcp/tools/tools.go`):
- Add `BuildTool("query_session_content", ...)` with `role` enum: `user|assistant|tool|all` and optional `contains`, `limit`, `scope`
- Add `BuildTool("query_session_signals", ...)` with `type` enum: `errors|tokens|system_errors|timestamps|tool_stats|all` and optional `limit`, `scope`
- Add `BuildTool("query_file_activity", ...)` with `type` enum: `snapshots|edits|reads|all`, optional `files []string`, `limit`, `scope`

**DoD**:
- `go test ./internal/mcp/...` passes
- Three new tool names appear in `BuildTools()` output
- No existing tool registrations removed

---

### Phase B: Implement routing handlers for new tools

**Goal**: New tools are fully functional end-to-end. Routing dispatches to existing handler logic.

**Tests first** (`internal/mcp/executor/query_handlers_test.go` or new `consolidated_handlers_test.go`):
- `query_session_content` with `role="user"` returns same results as calling `query_user_messages` handler directly
- `query_session_content` with `role="tool"` matches `query_tool_blocks` handler output
- `query_session_content` with `role="assistant"` + `contains="## Summary"` matches `query_summaries` output
- `query_session_content` with `role="all"` matches `query_conversation_flow` output
- `query_session_signals` with `type="errors"` matches `query_tool_errors` output
- `query_session_signals` with `type="tokens"` matches `query_token_usage` output
- `query_session_signals` with `type="system_errors"` matches `query_system_errors` output
- `query_session_signals` with `type="timestamps"` matches `query_timestamps` output
- `query_session_signals` with `type="tool_stats"` matches `query_tools` output
- `query_file_activity` with `type="snapshots"` matches `query_file_snapshots` output

**Implementation** (`internal/mcp/executor/query_handlers.go`):
- `HandleQuerySessionContent(params)`: switch on `params.Role`, delegate to existing per-role logic
- `HandleQuerySessionSignals(params)`: switch on `params.Type`, delegate to existing per-type logic
- `HandleQueryFileActivity(params)`: switch on `params.Type`, currently only `snapshots` wired; others are no-ops or empty

Register new handlers in `internal/mcp/executor/registry.go` or `executor.go`.

**DoD**:
- `go test ./internal/mcp/...` passes
- Each new tool call returns output equivalent to its old counterpart for each `type`/`role` value

---

### Phase C: Mark old tools deprecated, forward to new handlers

**Goal**: Old tool names still work but carry deprecation notices. Enables zero-disruption transition period.

**Tests first**:
- Calling old tool name `query_tool_errors` returns same result as `query_session_signals({type:"errors"})`
- Old tool descriptions contain the string `DEPRECATED`
- Tool count in `BuildTools()` = original count + 3 new tools (both old and new registered)

**Implementation** (`internal/mcp/tools/tools.go`):
- Update each old tool's description string to prepend `[DEPRECATED — use query_session_signals / query_session_content / query_file_activity instead] `

(`internal/mcp/executor/` dispatch):
- Old handler functions forward to new handler functions (or executor routes old names to new handlers)

**DoD**:
- `go test ./internal/mcp/...` passes
- Old tool names callable and return correct results
- All old tool descriptions contain `DEPRECATED`

---

### Phase D: Remove old tool registrations; update tests and docs

**Goal**: Hard removal of the 10 old tools. All test references updated. Docs rewritten.

**Tests first**:
- `BuildTools()` returns exactly the new set — old names `query_tool_errors`, `query_token_usage`, etc. are absent
- All integration tests use new tool names and `type`/`role` parameters

**Implementation**:
- Remove 10 old `BuildTool(...)` calls from `internal/mcp/tools/tools.go`
- Remove or consolidate old handler functions from `internal/mcp/executor/query_handlers.go`
- Update all `*_test.go` files that call old tool names
- Rewrite `docs/guides/mcp-query-tools.md` to document `query_session_content`, `query_session_signals`, `query_file_activity`
- Update `CLAUDE.md` FAQ examples and any agent prompts in `.claude/commands/`

**DoD**:
- `go test ./...` passes with zero references to old tool names in test files
- `grep -r "query_tool_errors\|query_token_usage\|query_system_errors\|query_timestamps\|query_summaries\|query_tool_blocks\|query_user_messages\|query_conversation_flow\|query_file_snapshots" internal/` returns no matches
- `docs/guides/mcp-query-tools.md` documents only new API

---

## Acceptance Gate

- `go test ./...` passes
- `BuildTools()` returns exactly 3 new query tools (old 10 are gone)
- Each old tool's query capability reachable via new tool + parameter
- `docs/guides/mcp-query-tools.md` updated
- No references to deprecated old tool names in `internal/` source or test files
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-06-23: Phase D (硬删除旧工具名) 风险经评估可接受——当前用户规模小，迁移成本低于长期维护 deprecated alias 的负担。全部 4 个 Phase 均纳入执行范围。优先级提升为 high。

claimed: 2026-06-23T06:45:00Z

Phase A ✓ 2026-06-23T00:00:00Z: Added query_session_content, query_session_signals, query_file_activity BuildTool entries to tools.go with required role/type params. All new tools appear in schema index. Tests pass.

Phase B ✓ 2026-06-23T00:00:00Z: Implemented consolidated_handlers.go with handleQuerySessionContent (routes role=user/assistant/tool/all), handleQuerySessionSignals (routes type=errors/tokens/system_errors/timestamps/tool_stats), handleQueryFileActivity (routes type=snapshots). Registered in queryHandlerRegistry via init(). All mcp tests pass.

Completed: 2026-06-23T07:40:00Z
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Execution Summary\n\n**Result:** Done\n**Commit:** 8db045c → merged 83120d6\n\n### Changes (23 files, +984 / -647)\n- tools.go: 10 old query_* registrations → 3 consolidated tools\n- consolidated_handlers.go (new): thin routers for query_session_content/signals/file_activity\n- All *_test.go updated to new tool names and role/type params\n- docs/guides/mcp-query-tools.md rewritten for new API\n- CLAUDE.md FAQ examples updated\n- Bug fix: skip default pattern injection for content_type=array\n- Total tool count: 22 → 15
<!-- SECTION:FINAL_SUMMARY:END -->
