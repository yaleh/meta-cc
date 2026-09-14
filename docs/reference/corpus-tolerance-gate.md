# Corpus Tolerance Gate (DIR-099)

How meta-cc keeps every session-corpus-consuming tool tolerating malformed
session files, and what to do when you add one.

## The contract

A session corpus is whatever `ListSessions`/`GetQueryFiles` enumerate for a
project directory. Any of those files can be unusable — Claude Code writes a
zero-byte file, a session-start stub with no messages, or a write truncated
mid-JSON. DIR-018 established that such a file is **skipped and reported**,
never fatal. DIR-032/DIR-034 then rewrote the `ListSessions` path behind
`query_sessions` without that behavior, and one empty file hard-crashed the
tool. DIR-094 restored it everywhere it had been lost.

There are two contracts, because the tools genuinely work two ways:

| Contract | Tools | What must hold |
|----------|-------|----------------|
| `exclusionContract` | `query_sessions`, `get_timeline`, `analyze_errors`, `analyze_bugs`, `quality_scan`, `get_work_patterns`, `get_tech_debt` | The tool decides per file whether that whole file contributes anything. The healthy session's data must still be returned, **and** every excluded file must be named in the response `warnings`. |
| `streamingContract` | `query_session_content`, `query_session_signals`, `query_file_activity` | The tool streams records line by line and never makes a whole-file decision; a truncated line is skipped and the rest of the file still contributes. There is no excluded file to name, so the claim is only: a malformed sibling neither aborts the call nor changes the records returned. |

A tool that tolerates *silently* — dropping a whole file without naming it —
satisfies neither contract. That is the DIR-018 point: silence is
indistinguishable from "there was nothing there".

## Where the gate lives

- `tests/fixtures/malformed-corpus/` — the checked-in corpus: `healthy.jsonl`
  plus `empty.jsonl` (zero bytes), `stub.jsonl` (metadata-only entries) and
  `malformed.jsonl` (truncated mid-JSON). DIR-094 added it; DIR-099 reuses it.
- `internal/testutil/malformed_corpus.go` — seeds that corpus into a real
  project directory reachable through a `working_dir`
  (`SeedMalformedCorpusProject`), and returns the excluded names.
- `internal/mcp/executor/corpus_tolerance_test.go` — the gate:
  - `TestCorpusTolerance_CorruptSiblingsChangeNothing` runs each registered
    corpus-consuming tool **twice over the same project** — once with the
    malformed siblings, once after removing them — and requires the responses
    (minus `warnings` and the volatile `file_ref` path) to be identical. For an
    `exclusionContract` tool it also requires every excluded file to be named,
    and requires the clean corpus to produce no exclusions at all.
  - `TestCorpusTolerance_EveryRegisteredToolIsClassified` requires every name
    in `specialToolRegistry` ∪ `queryHandlerRegistry` to appear in exactly one
    of the two registration tables.
  - `TestCorpusTolerance_ClassifierRejectsAnUnclassifiedTool` is the
    non-vacuity guard for that check: a classifier that matched nothing would
    otherwise let the coverage test pass while enforcing nothing.

The gate runs in well under a second, so it is on the `make commit` path
(`go test -short ./...`) with no skip and no build tag.

## Registering a new corpus-consuming tool

Adding a tool to `specialToolRegistry`/`queryHandlerRegistry` without
registering it here makes `TestCorpusTolerance_EveryRegisteredToolIsClassified`
fail with the tool's name. That failure is the instruction; the fix is always
one of two edits in `internal/mcp/executor/corpus_tolerance_test.go`:

1. **The tool reads session corpus content** — add an entry to
   `corpusToolCases`:

   ```go
   {
       name:     "your_tool",
       contract: exclusionContract, // or streamingContract
       args:     map[string]interface{}{"type": "..."}, // the args that select the corpus-reading path
       healthyMarker: `"unique_value":1`, // a substring only healthy.jsonl can contribute
   },
   ```

   Only the arguments that select the corpus-reading path belong in `args`;
   the harness supplies `working_dir` (the seeded project) and pins
   `provider: "claude"`, so the gate does not change meaning depending on
   which agent host runs it. Add `"output_mode": "inline"` for the query
   tools: file_ref mode moves the records into a temp file, where no marker
   can be found.

2. **The tool does not read session corpus content** — add a
   `nonCorpusTools` entry with a written reason ("takes an explicit `files`
   array", "returns per-file metadata without parsing"). The reason is the
   deliverable: it records that the classification was decided, not missed.

### Choosing `healthyMarker`

Pick a substring that only `healthy.jsonl` can produce, and make it specific
enough that a zero-valued field cannot match it (`"total_errors":1`, not
`"total_errors"`). A marker that also matches an empty result keeps passing
after the healthy session stops reaching the tool, and the regression goes
silent again — the same trap DIR-094 hit with `analyze_bugs` and
`get_tech_debt`, whose fixture had to be extended before either emitted any
session-derived output at all.

If this corpus genuinely cannot exercise the tool's data path, do not invent a
marker: set `emptyCorpusReason` instead, explaining what the corpus cannot
reach and which other entry covers the same code path. An entry may leave
`healthyMarker` empty *only* by writing one of those, so "no marker" is a
reviewed statement rather than an omission. `query_file_activity` is the one
such entry today: its selector matches nothing in the fixture.

## What the gate does not cover

Codex sessions. The corpus is a Claude one, and the Codex path has its own
per-session tolerance tests (`codex_corrupt_session_tolerance_test.go`,
`setupCodexMultiSessionFixtureProject`). A tool that reads both providers needs
an entry here for its Claude path plus coverage over there for its Codex path.

## See Also

- [Testing Guide](../guides/testing.md) - Two-stage acceptance workflow
- [Repository Structure](repository-structure.md) - Where tests and fixtures live
