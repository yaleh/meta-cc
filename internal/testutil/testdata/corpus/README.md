# DIR-099 malformed-session corpus

The shared corrupt-corpus fixture behind the cross-tool tolerance gate
(`internal/mcp/executor/corpus_tolerance_gate_test.go`). One corpus, every
session-corpus-consuming MCP tool.

DIR-018 established skip-and-report tolerance in the analysis `loadData` path.
DIR-032/DIR-034 then rewrote the `ListSessions` path behind `query_sessions`,
the behavior was not propagated, and a single 0-byte session file
(`8eda8f4e-2c74-4176-ba6b-8c45e890df42.jsonl`) hard-crashed `query_sessions`
twice in the 2026-07-30 dogfooding session. DIR-094 re-unified the contract and
pinned it for the analysis tools and `query_sessions`. This corpus exists so
the same contract is checked against *every* corpus-consuming tool from a
single table, with a completeness check that a newly added tool cannot skip.

The files are read through `go:embed` (`internal/testutil/corpus.go`) rather
than a relative path, so any package can import the fixture without resolving
a path relative to `testutil`'s own directory.

| File | Shape | Why it exists |
| --- | --- | --- |
| `control-session.jsonl` | A real 10-entry Claude session (user → assistant → Grep → Read → Bash error → Bash fix), carrying a `TODO` marker and one error→fix pair | The control. It is deliberately *not* empty of signal: the TODO and the error→fix pair make `get_tech_debt`, `analyze_errors` and `analyze_bugs` produce distinguishable output, so "the tool still returned results" is a real assertion and not merely "the call returned". Installed under its own session ID, `6a32f273-191a-49c8-a5fc-a5dcba08531a` |
| `corrupt-empty.jsonl` | 0 bytes | The exact shape of the 2026-07-30 failure: a session file Claude Code created but never wrote a message into. It parses with **no error** and zero entries, so a loader that only checks for errors drops it silently — the silent-tolerance case DIR-094 closed. Installed under the reported session ID `8eda8f4e-2c74-4176-ba6b-8c45e890df42` so the gate reproduces the reported shape verbatim |
| `corrupt-truncated.jsonl` | A JSONL line cut off mid-object (`{"type":"user",...,"mess`) | The DIR-018 case: a genuine parse **error**, reached without a 0-byte file. Guards the error-handling half of the contract |
| `corrupt-wrong-shape.jsonl` | Valid JSON whose top level is an array (`[{"type":"user"},…]`) rather than a session object | Right JSON, wrong shape. Reaches the parse-error path by a different route than truncation, so a rewrite that special-cases "truncated line" is still caught |

These are data fixtures, not Go test files: they are embedded with `go:embed`,
so `make check-fixtures` (which verifies `testutil.LoadFixture` references)
does not apply to them.

**Observability note:** an MCP tool response only shows this behavior after the
installed plugin binary is refreshed — a stale install keeps serving the old
tolerance. See `docs/guides/troubleshooting.md` ("Malformed session file
tolerated but no warning appears").
