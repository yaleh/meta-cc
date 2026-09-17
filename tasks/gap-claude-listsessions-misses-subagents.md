---
id: gap-claude-listsessions-misses-subagents
title: Claude provider ListSessions never enumerates subagent transcripts
  (query_sessions omits them; session_id cannot address them)
status: todo
labels:
  - gap
  - defect
parent: null
children: []
extra:
  schema: execution
---
## Proposal

**Defect.** The Claude provider's session listing enumerates only top-level
transcripts and never the subagent transcripts that sit beside them, so every
MCP surface built on that listing is structurally blind to subagent history.

`Provider.ListSessions` (`internal/provider/claude/provider.go:95`) resolves its
file set from `SessionLocator.AllSessionsFromProject`
(`internal/locator/args.go:138`). Neither function references subagents at all —
`grep -rn subagent internal/provider/claude/provider.go` returns 0 hits — the
walk covers `<projectDir>/*.jsonl` and stops there.

Subagent transcripts live one level deeper, at
`<projectDir>/<session-uuid>/subagents/agent-<agentId>.jsonl`, and meta-cc
already knows how to read them. `GetQueryFiles`
(`internal/mcp/query/query.go:430`) implements exactly that scan, and its own doc
comment states the contract: `scope=project, includeSubagents=true → top-level +
all */subagents/*.jsonl`, with the walk fixed at two levels deep so it does not
pick up `tool-results/`.

So the two paths disagree about what the corpus IS:

| path | enumerates subagent transcripts? |
|---|---|
| `GetQueryFiles` -> content/signal queries (`executor.go:178`) | yes |
| `AllSessionsFromProject` -> `ListSessions` -> `query_sessions` | no |

**Measured on this host, 2026-09-17.** The project corpus held 13 top-level
`.jsonl` files and 2 subagent `.jsonl` files. `query_sessions` with
`scope: "project"` and `include_subagents: true` returned exactly 13 records;
both subagent ids (`a16670a1c4453c992`, `ad81c407898530c85`) were absent from
the output. `query_session_signals` with `session_id: "a16670a1c4453c992"`
failed with `session file not found for ID "a16670a1c4453c992"`, because
`findSessionFile` resolves the same top-level-only set.

**Consequence.** The documented discovery workflow — "use `query_sessions` to
discover which session to target before querying it with `session_id`" — cannot
reach a subagent on the Claude path at all. Subagent history is readable today
only by hand-assembling the path from `get_session_directory` and going through
the two-stage tools, which is not the advertised route.

<!-- dedup-ref -->
Related but distinct: DIR-034 and DIR-039 (the Codex `ListSessionsPage` cursor
API is not used by `query_sessions`), DIR-094 and DIR-099 (malformed-file
tolerance across corpus-enumerating tools). Those concern which pages are
fetched and how bad files are tolerated; this one concerns which files are in
scope at all.

## Plan

1. Extend the Claude enumeration to include subagent transcripts. Pick the seam
   deliberately: either `AllSessionsFromProject` grows the subagent walk, or
   `ListSessions` composes the same discovery `GetQueryFiles` already performs.
   Prefer ONE implementation shared by both callers over a second copy of the
   rule — this repository has removed single-source drift repeatedly, and two
   copies of "what is the corpus" is exactly that defect class.
2. Preserve the two-levels-deep rule and the `tool-results/` avoidance proven in
   `GetQueryFiles`, so the wider enumeration does not start listing non-session
   files.
3. Make `findSessionFile` resolve a subagent id, so `session_id=<agentId>` works
   on the query, timeline, and analysis tools for a subagent the listing just
   returned.
4. Cover it with a fixture: one top-level session plus one
   `<uuid>/subagents/agent-*.jsonl`, asserting the subagent appears in
   `ListSessions` and is reachable by id.
5. Verify against the live corpus rather than the fixture alone: 13 top-level +
   2 subagent files must list as 15, with both agent ids present.

## Acceptance Criteria

- [ ] `go test ./internal/provider/claude/ -run TestListSessionsIncludesSubagents`
      exits 0, against a fixture containing one `<uuid>/subagents/agent-*.jsonl`.
- [ ] `go test ./internal/locator/` exits 0, with a case covering the subagent
      directory level and confirming `tool-results/` is still excluded.
- [ ] Against the live corpus, `query_sessions({scope:"project",
      include_subagents:true})` returns 15 records, verified by grepping its
      output for both `a16670a1c4453c992` and `ad81c407898530c85`.
- [ ] `query_session_signals({session_id:"a16670a1c4453c992"})` returns records
      rather than `session file not found`.
- [ ] `make commit` passes.

## DoD

- The two real subagent transcripts on this host are discoverable through the
  documented route end to end: `query_sessions` lists them, and the id it returns
  can be passed to another query tool to read that subagent's records — not
  merely that a unit test over a synthetic fixture is green.
- Exactly one implementation of the subagent-scan rule exists, shared by the
  listing and by `GetQueryFiles`; a third caller cannot reintroduce the split.
- `make commit` passes and the change is landed on main.

## Touches

- `tasks/gap-claude-listsessions-misses-subagents.md`
- `internal/provider/claude/provider.go`
- `internal/provider/claude/provider_test.go`
- `internal/locator/args.go`
- `internal/locator/args_test.go`
- `internal/mcp/query/query.go`
