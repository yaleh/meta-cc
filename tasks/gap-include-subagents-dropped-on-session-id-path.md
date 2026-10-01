---
id: gap-include-subagents-dropped-on-session-id-path
title: "include_subagents is unreachable when session_id is given: subagent
  transcripts are never searched"
status: todo
labels:
  - gap
  - defect
  - mcp
  - query
parent: null
children: []
extra:
  schema: execution
  scope:
    owner_repo: meta-cc
  acceptance: make commit
---
## Proposal

**Defect.** `include_subagents` is structurally unreachable whenever the caller
supplies `session_id`. `query_session_content` therefore returns
`total_records: 0` for needles that provably exist in that session's subagent
transcripts, and `include_subagents: true` / `false` return byte-identical
results. A caller who trusts the parameter (it exists, it defaults to `true`) is
silently told "no data".

Root cause — one dropped argument, not a missing corpus:

- `internal/mcp/executor/provider_query.go:192` — `dispatchProviderQuery` calls
  `ExecuteQueryForSession(providerName, sessionID, jqFilter, limit, workingDir, tr)`
  and **does not pass `includeSubagents`**. The sibling branch (`:194`) does pass it.
- `internal/mcp/executor/provider_query.go:72` — `ExecuteQueryForSession`'s
  signature has no such parameter and its body never mentions subagents; its own
  doc comment describes it as "a direct file lookup, no directory scan".

The enumeration that would satisfy the flag already exists and matches the
on-disk layout — `internal/mcp/query/query.go:434 GetQueryFiles(scope,
workingDir, includeSubagents)` derives `<projectDir>/<uuid>/subagents/` from
`<projectDir>/<uuid>.jsonl` — it is simply never reached on the `session_id`
path. This is why the failure is silent: the flag is accepted, defaults to true,
and is discarded before any resolver sees it.

## Evidence

Reproduced independently on HEAD `b6bc856` in this workspace, with a needle that
exists only in a subagent transcript:

    needle    0a4463dc-5039-4011-99da-b00b967fbfc0
    parent    <hash>/7a5d362c-056b-4553-88c8-47cdaabb02be.jsonl          grep -c = 0
    subagent  <hash>/7a5d362c-.../subagents/agent-ad81c407898530c85.jsonl grep -c = 2

    query_session_content{role:"all", contains:<needle>, session_id:"7a5d362c-...", include_subagents:true}
      -> {"data":[],"pagination":{...,"total_records":0}}
    query_session_content{role:"all", contains:<needle>, session_id:"7a5d362c-...", include_subagents:false}
      -> identical

Cross-checked from a second workspace/session (quay, 2026-10-01) against session
`e8dd723c-4416-4a8c-b7ea-7ab5eb25e4ce` with 5 subagent files: same
identical-zero result.

## Plan

1. Thread `includeSubagents` into `ExecuteQueryForSession` and its claude branch.
2. When true, resolve the session's subagent transcripts
   (`<projectDir>/<uuid>/subagents/*.jsonl`) by reusing the derivation
   `GetQueryFiles` already uses, rather than adding a second one.
   Note `projectHashDirName` (`internal/locator/args.go:137`) already walks the
   three-level subagent depth, so the boundary helper does not obstruct this.
3. Confirm every path that reaches this branch honours the flag: both the
   `role:"all"` conversation-flow path and the `role:"assistant"` path route
   through `dispatchProviderQuery`.

## Acceptance Criteria

- [ ] A **generated fixture** (never a real session needle) proves the flag can
      take false: create `<tmp-project-hash>/<sid>.jsonl` containing an assistant
      text `FIXTURE-MAIN-<nonce>` and `<sid>/subagents/agent-fix.jsonl` containing
      `FIXTURE-SUB-<nonce>` (with `sessionId=<sid>`, `isSidechain:true`), then
      assert `query_session_content{role:"all", contains:"FIXTURE-SUB-<nonce>",
      session_id:"<sid>", include_subagents:true}` returns `total_records >= 1`.
- [ ] The same query with `include_subagents:false` returns `total_records == 0`;
      the two arms are no longer byte-identical.
- [ ] The fixture is rebuilt per run with a fresh nonce, so a leftover from a
      previous run cannot satisfy the assertion.
- [ ] The fixture test passes a physical `working_dir`, not a symlink alias
      (an alias would trip gap-path-normalization-bypassed-in-query-sessions-filter first).

## DoD

- make commit passes
- changes landed on main
- the fixture assertion is a committed test that fails on pre-fix code and passes after

## Touches

- internal/mcp/executor/provider_query.go
- internal/mcp/query/query.go (only if the subagent derivation is shared)
- internal/mcp/executor/ (new fixture-based regression test)
