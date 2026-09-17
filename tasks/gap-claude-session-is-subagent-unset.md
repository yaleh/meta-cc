---
id: gap-claude-session-is-subagent-unset
title: Claude provider never sets Session.IsSubagent, so query_sessions can
  never emit is_subagent on that path
status: ready
labels:
  - gap
  - defect
parent: null
children: []
extra:
  schema: execution
---
## Proposal

**Defect.** On the Claude path nothing ever sets `Session.IsSubagent`, so the
`is_subagent` metadata field is unreachable there — and because it is declared
`omitempty` and emitted only inside an `if`, its absence is indistinguishable
from a positive "this session is not a subagent".

The field and its emission both already exist:

- `internal/conversation/types.go:43` declares
  `IsSubagent bool` with json tag `is_subagent,omitempty`.
- `internal/mcp/executor/query_sessions_handler.go:336` emits it as
  `if s.IsSubagent { entry["is_subagent"] = true }`.

What is missing is the producer. `sessionFromEntries`
(`internal/provider/claude/provider.go:257`) builds the `conversation.Session`
with ID, Provider, Title, CWD, Model, CreatedAt, TokenUsage and Extensions only,
and `grep -rn IsSubagent internal/provider/claude/ internal/locator/` returns
**0 hits**. So the Claude path never reaches the handler's emit branch.

**Consequence.** A consumer cannot distinguish a subagent thread from a
top-level session on the Claude path even after the enumeration gap is closed —
the record simply lacks the field. Meanwhile `docs/guides/mcp-query-tools.md:109`
lists `is_subagent` among the metadata `query_sessions` returns, and does not
scope it to Codex, unlike the Codex-only *filters* it explicitly marks as failing
with an actionable error when `provider` is claude. So the guide describes
behaviour this path cannot produce.

This is distinct from the enumeration defect and has a different fix. Closing
enumeration makes subagent transcripts *reachable*; it does not make them
*attributable*. Both defects surface as "no subagent identity in query_sessions
output", which is why they are filed separately rather than as one item: a fix to
either alone leaves the other symptom intact and would look like a failed fix.

<!-- dedup-ref -->
Related: DIR-030 (Codex session discovery and filtering, where these metadata
fields are populated today) and DIR-024 (provider-aware session discovery). Both
are Codex-side; this one is the Claude-side producer that was never written.

## Plan

1. Populate `IsSubagent` in `sessionFromEntries` for a transcript whose path lies
   under a `<uuid>/subagents/` directory. The path is already in hand — the
   function receives `file` as its first argument.
2. Populate the linkage a consumer needs to attribute the record:
   `ParentThreadID` from the parent session uuid encoded in that path, so a
   listed subagent can be traced to the session that spawned it.
3. Decide the emit shape deliberately. `omitempty` plus `if` conflates "false"
   with "unknown"; if a consumer must tell a top-level session from a subagent,
   emitting `is_subagent: false` explicitly on the Claude path is the honest
   choice. Match what the Codex path already does rather than inventing a third
   convention for the same field.
4. Add tests over a fixture transcript inside `<uuid>/subagents/`, asserting
   `is_subagent: true` and the parent link in the handler's record, plus a
   top-level control asserting the field is not set true.
5. Reconcile `docs/guides/mcp-query-tools.md:109` with the outcome: either the
   Claude path now emits the field and the sentence stands as written, or the
   sentence must scope `is_subagent` to the providers that populate it.

## Acceptance Criteria

- [ ] `go test ./internal/provider/claude/ -run TestSessionFromEntriesSetsIsSubagent`
      exits 0, asserting `IsSubagent` is true for a fixture transcript under
      `<uuid>/subagents/` and false for a top-level control.
- [ ] The same test asserts `ParentThreadID` resolves to the fixture's parent
      session uuid.
- [ ] A handler-level test asserts the emitted record carries
      `is_subagent: true` for that fixture, and that a top-level session's record
      does not carry it as true.
- [ ] `docs/guides/mcp-query-tools.md` no longer describes `is_subagent` in a way
      the Claude path cannot produce, verified by reading the sentence at
      line 109 against the shipped behaviour.
- [ ] `make commit` passes.

## DoD

- A subagent transcript on the Claude path is attributable in real output: the
  record `query_sessions` returns for it carries `is_subagent: true` and a parent
  link, observed against a real subagent transcript on this host — not merely a
  green unit test over a synthetic fixture.
- The guide and the shipped behaviour agree on which providers populate
  `is_subagent`, so a reader cannot be misled in either direction.
- `make commit` passes and the change is landed on main.

## Touches

- `tasks/gap-claude-session-is-subagent-unset.md`
- `internal/provider/claude/provider.go`
- `internal/provider/claude/provider_test.go`
- `internal/mcp/executor/query_sessions_handler.go`
- `internal/conversation/types.go`
- `docs/guides/mcp-query-tools.md`
