---
id: gap-path-normalization-bypassed-in-query-sessions-filter
title: "query_sessions returns silently empty for a symlinked alias path:
  filter.CWD bypasses PathToHash"
status: ready
labels:
  - gap
  - defect
  - mcp
  - query
  - reliability
parent: null
children: []
extra:
  schema: execution
  scope:
    owner_repo: meta-cc
  acceptance: make commit
---
## Proposal

**Defect.** `query_sessions` returns a silently empty result when `working_dir`
(or an explicit `cwd`) is a symlinked alias of the real project path. The
physical path returns sessions; the alias returns nothing — no error, no
warning. On this host `/home/yale` → `/data/home/yale`, so
`working_dir:"/home/yale/work/meta-cc"` yields `(no output)` while
`working_dir:"/data/home/yale/work/meta-cc"` yields `{"count":35}`.

Root cause: two path-normalization disciplines coexist, and exactly one
comparison bypasses the one the corpus is keyed by.

- Corpus keys come from `PathToHash` (`internal/locator/args.go:386`), which
  resolves symlinks (`:393 filepath.EvalSymlinks`) — this is correct and should
  not change.
- 12 of the 13 production `filepath.Abs` call sites funnel their result into
  `PathToHash` before any comparison and are therefore alias-safe. Examples:
  `args.go:124→127`, `args.go:162→167`, `args.go:184→189`, `args.go:245→250`.
  The last is reached via `provider.go:106/:213 AllTranscriptsFromProject`,
  which is why `query_session_content` and `analyze_errors` stay safe even
  though they pass through an `Abs` — there, `Abs` is a waypoint, not a terminus.
- The one exception is `internal/mcp/executor/query_sessions_handler.go:90`
  (`filepath.Abs` — Clean only, no symlink resolution) whose result is written
  straight into the filter at `:101` (`filter.CWD = projectPath`) and compared,
  never passing through `PathToHash`.

Why silence rather than an error: the loud `no sessions found for project %q
(hash: %s)` diagnostic (`args.go:171`, `:193`) fires only when the project
*directory* cannot be found. With the alias the directory IS found
(`PathToHash` resolves it), so nothing errors — the filter simply matches no
session and an empty result is returned. A genuinely nonexistent path is loud
(`/nonexistent-mcc-probe-9f3a` names its hash), which is what distinguishes the
alias case from a miss and confirms it as a third, silent mode.

## Evidence

Variable-isolating probe — `working_dir` held constant as the alias, only `cwd`
varies:

    query_sessions{stats_only:true, working_dir:"/home/yale/work/meta-cc", cwd:"/data/home/yale/work/meta-cc"} -> {"count":35,...}
    query_sessions{stats_only:true, working_dir:"/home/yale/work/meta-cc", cwd:"/home/yale/work/meta-cc"}      -> (no output)

Both arms are affected: the default arm (no `cwd`; `:98-99` sets
`filter.CWD = projectPath`) and the explicit-parameter arm (`cwd` overrides
`boundaryCWD`).

Entry points checked behaviourally for alias/physical equivalence (all identical):

    query_session_content                    -> same sessionId, same 51295 bytes, record_count 2
    analyze_errors                           -> total_errors 52, by_tool / by_type identical
    get_session_directory (provider=claude)  -> 35 files, same resolved directory

## Codex open item — ADJUDICATED: same defect (fixed)

The filing left this unproven: `internal/mcp/query/stage.go:217`
(`resolveProjectPath`, `Abs`) reaches `rawfiles.NewRegistry(projectPath)` +
`rawfiles.SelectCodexFiles` through its only two callers (`stage.go:115
buildCodexDirectoryResult`, `stage.go:589 buildCodexMetadataResult`), both
Codex-only. Codex rollouts are not stored under the Claude `<hash>` layout, so
the `PathToHash` funnel does not exist on that path.

Verdict: **it IS the same defect.** `SelectCodexFiles` matches sessions with a
RAW cwd comparison (`providerrecords.FilterSessionsForScope`, `records.go:109`),
not through `PathToHash`, so a symlinked alias `working_dir` filtered every
Codex session out and surfaced the loud `no codex sessions found for project
<alias>` while the physical path returned the corpus. Behaviourally proven
against a hermetic Codex fixture (a thread whose recorded cwd is the resolved
project): both `get_session_directory` and `get_session_metadata` fail for the
alias form and match the physical form after the fix — see
`TestCodexPathTakingEntryPoints_AliasEqualsPhysical` in
`internal/mcp/executor/query_sessions_path_alias_test.go` (red before the fix,
green after).

## Plan

1. Normalize the comparison input: pass `filter.CWD` through the same resolver
   the corpus is keyed by (`PathToHash`), or compare hashes rather than raw
   paths.
2. Cover both arms (`:98-99` default; the explicit `cwd` override).
3. Adjudicate the Codex open item with a project that has codex sessions and
   record the verdict either way.

## Resolution

- `internal/mcp/executor/query_sessions_handler.go`: new `physicalProjectPath`
  helper (`filepath.Abs` + `filepath.EvalSymlinks`, falling back to the Abs'd
  path on resolution failure — matching `PathToHash`'s own fallback, so a
  nonexistent path keeps its loud, named miss). Applied to BOTH the default
  `working_dir` arm and the explicit `cwd` arm.
- `internal/mcp/query/stage.go`: `resolveProjectPath` now also resolves
  symlinks, closing the Codex discovery path adjudicated above.
- Regression is a property, not two symptoms:
  `TestPathTakingEntryPoints_AliasEqualsPhysical` asserts alias == physical for
  every path-taking entry point (query_sessions default + explicit-cwd,
  query_session_content, analyze_errors, get_session_directory);
  `TestQuerySessions_SymlinkedAliasMatchesPhysical` is the focused two-arm
  regression; `TestCodexPathTakingEntryPoints_AliasEqualsPhysical` is the Codex
  adjudication. All three were red before the fix and green after.

## Acceptance Criteria

- [x] `query_sessions` with a symlinked alias `working_dir` returns the same
      result as the physical path (default arm).
- [x] Same with an explicit `cwd` set to the alias (explicit-parameter arm).
- [x] The regression is written as a property, not as two symptoms: for every
      path-taking entry point, alias and physical forms must return the same
      result. Expectation table at the time of filing: 12 sites route through
      `PathToHash`; 1 (`query_sessions_handler.go:101`) does not.
- [x] The Codex open item above is adjudicated and its verdict recorded in this task.

## DoD

- make commit passes
- changes landed on main
- the alias/physical equivalence assertions are committed tests

## Touches

- tasks/gap-path-normalization-bypassed-in-query-sessions-filter.md
- internal/mcp/executor/query_sessions_handler.go
- internal/mcp/query/stage.go (only if the Codex open item resolves to a defect)
- internal/mcp/executor/query_sessions_path_alias_test.go (new)
