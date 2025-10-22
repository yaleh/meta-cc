# Phase 23: Query Library Extraction - Implementation Plan

## Overview

**Goal**: Extract reusable query functions under `internal/query` so CLI (`cmd/query_*`) and MCP (`cmd/mcp-server`) share the same execution path without spawning the `meta-cc` binary.

**Timeline**: ~4 days (TDD-oriented, three iterations)

**Estimated Effort**: ~550 lines (implementation ≤320, tests ≤230). Each stage keeps source/tests ≤200 lines as required by project-planner constraints.

**Status**: Planned

---

## Objectives

1. Provide a stable `internal/query` package that exposes high-level helpers (`RunToolsQuery`, `RunUserMessagesQuery`, `RunSessionStats`) wrapping `SessionPipeline` (`pkg/pipeline/session.go`) and existing filters.
2. Refactor CLI commands (`cmd/query_tools.go`, `cmd/query_messages.go`, `cmd/stats_aggregate.go`, etc.) to use the shared helpers without changing user-facing flags.
3. Replace MCP executor subprocess invocation (`cmd/mcp-server/executor.go:90-153`) with direct library calls while preserving jq filtering (`cmd/mcp-server/jq_filter.go`) and hybrid output behavior.
4. Add unit/integration tests that compare CLI and MCP results against the new library to prevent regressions.

---

## Dependencies

- Phase 14 pipeline abstractions (`pkg/pipeline/session.go`, `internal/parser`) remain stable.
- Phase 15 standard parameters (`cmd/mcp-server/tools.go`) already normalize tool inputs.
- Phase 16 hybrid output mode (`cmd/mcp-server/output_mode.go`, `temp_file_manager.go`) should remain untouched; new code must call into these facilities.
- Existing fixtures under `tests/fixtures/` cover tool calls and message sessions; reuse them for library tests.

---

## Acceptance Criteria

- ✅ CLI query commands compile against `internal/query` with no behavior changes on golden fixtures.
- ✅ MCP executor no longer uses `exec.Command` for query tools; all tool responses flow through shared helpers and pass existing integration tests.
- ✅ Library unit tests achieve ≥85% coverage over the new package and verify pagination/filter combinations (limit/offset, jq, presets).
- ✅ Documentation references (developer guide, `docs/core/plan.md`) updated to mention the new package and migration guidance.

---

## Iteration Structure

The iteration is divided into three stages, each delivering runnable code and tests before moving to the next one.

### Stage 23.1 — Library Skeleton & CLI Integration

**Objective**: Create `internal/query` with minimal helpers and rewire CLI commands to consume them.

**Implementation Scope** (~180 lines code + ~100 lines tests):
- Add `internal/query/options.go` defining `Options`, `Pagination`, `OutputConfig` structs (subset mirroring current CLI globals).
- Implement `internal/query/tools.go`, `internal/query/messages.go` that:
  - Instantiate `SessionPipeline` (`pkg/pipeline/session.go`)
  - Apply existing filters by invoking `internal/filter` helpers
  - Return strongly typed results (`[]parser.ToolCall`, `[]cmd.UserMessage`)
- Update `cmd/query_tools.go` and `cmd/query_messages.go` to call the new helpers while keeping flag parsing untouched.

**Tests** (~100 lines):
- `internal/query/tools_test.go`: table tests verifying status/tool filters, pagination (`limit`, `offset`), and sorting call through.
- `internal/query/messages_test.go`: tests for regex pattern filtering and contextual data.
- Adjust existing CLI tests (`cmd/query_tools_test.go`, `cmd/query_messages_test.go`) to stub the new package if necessary and confirm outputs are unchanged (snapshot comparison or equality checks on fixtures).

**Acceptance**:
- CLI `meta-cc query tools --status error --limit 5` returns identical output compared to commit baseline when run against `tests/fixtures/sample-session.jsonl`.
- All unit tests pass (`go test ./internal/query ./cmd/...`).

**Dependencies**: None beyond existing packages; ensures library API compiles before MCP refactor.

### Stage 23.2 — MCP Executor Refactor

**Objective**: Switch MCP query execution to the new library and centralize jq/output helpers.

**Implementation Scope** (~140 lines code + ~80 lines tests):
- Move `ApplyJQFilter` and associated helpers from `cmd/mcp-server/jq_filter.go` into `internal/query/transform/jq.go` (exported for both CLI and MCP if needed) without changing behavior.
- Update `cmd/mcp-server/executor.go` to:
  - Build `query.Options` from incoming tool args (`scope`, `limit`, `offset`, `stats_only`, `stats_first`, `jq_filter`).
  - Invoke `query.RunTools`, `query.RunUserMessages`, etc. instead of spawning `meta-cc`.
  - Preserve `toolPipelineConfig` handling for content truncation by invoking existing helpers in-place after library call.
- Ensure hybrid output mode (`output_mode.go`) receives the same data structure it expects today.

**Tests** (~80 lines):
- Extend `cmd/mcp-server/executor_test.go` with integration cases for `query_tools` and `query_user_messages`, asserting:
  - No temporary executable call occurs (mock or inspect `ToolExecutor` internals).
  - jq filters applied via the relocated helper produce expected slices.
- Add tests for `internal/query/transform/jq` verifying default expression (`.[]`) and quoted-expression error message remain intact.

**Acceptance**:
- `go test ./cmd/mcp-server` passes with refactored executor.
- Manual or scripted invocation of `meta-cc-mcp` for `query_tools` returns identical JSONL compared to baseline.

**Dependencies**: Stage 23.1 completed (library API available).

### Stage 23.3 — Regression Harness & Documentation

**Objective**: Lock in behavior with regression tests and update developer docs.

**Implementation Scope** (~120 lines code/tests + docs):
- Introduce a golden-test harness `tests/integration/query_library_compare_test.go` (or update an existing integration test) that runs both CLI and library calls against fixtures to ensure outputs stay identical across commands.
- Add Makefile target (if helpful) or script under `test-scripts/` to automate comparison runs for future changes.
- Update documentation:
  - `docs/development/query-library.md` (new) summarizing the shared helpers and usage patterns.
  - Amend `docs/core/plan.md` Phase 23 status and mention new package.
  - Note CLI internals in `README.md` or developer guide to reflect the refactor.

**Tests** (~70 lines):
- Regression test verifying `RunToolsQuery` output matches CLI JSONL on fixtures.
- Additional assertions ensuring `QueryOptions` default to project scope and that streaming/pagination flags behave identically.

**Acceptance**:
- All regression tests pass locally (`make test`, `make test-all`).
- Documentation PR checklist satisfied; reviewers can follow migration instructions.
- Internal playbook records the new library entry points for Phase 24.

**Dependencies**: Stages 23.1 and 23.2 complete; regression harness relies on final APIs.

---

## Exit Criteria

- ✅ Code merged: `internal/query` package, CLI & MCP refactors, regression harness.
- ✅ Tests: unit + regression suite green on CI, coverage maintained ≥80%.
- ✅ Docs: Plan updates plus new developer guidance committed.
- ✅ Ready for Phase 24: `QueryOptions` structure documented and consumed by both entry points.

---

## Post-Iteration Follow-ups

1. Schedule Phase 24 kick-off meeting to confirm parameter simplification strategy leveraging the new `QueryOptions`.
2. Monitor runtime metrics (CLI latency, MCP tool duration) for regressions; if necessary, add lightweight instrumentation hooks in `internal/query`.
3. Prepare release notes highlighting that MCP now uses the shared library (should be transparent to users but worth noting for maintainers).
