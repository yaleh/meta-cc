# Corpus-Tolerance Gate

How to register a session-corpus-consuming MCP tool in the cross-tool
malformed-file tolerance gate, and why the gate is shaped the way it is.

## The contract

A session corpus is a directory of JSONL transcripts. Any of them can be
unusable: 0 bytes because Claude Code created the file and never wrote a
message into it, truncated mid-object, or valid JSON whose top level is not a
session record. When one file is unusable, a corpus-consuming tool must:

1. **not fail the whole call** — the other sessions' results still come back;
2. **not drop the file silently** — the response names what was skipped, in
   both the prose channel (`warnings`) and the machine-readable one
   (`skipped_files`).

DIR-018 established this in the analysis `loadData` path. DIR-032/DIR-034
rewrote the `ListSessions` path behind `query_sessions` without carrying the
behavior over, and a single 0-byte file hard-crashed `query_sessions` twice in
the 2026-07-30 dogfooding session. DIR-094 re-unified the contract and pinned
it for the analysis tools and `query_sessions`. DIR-099 added the gate below so
that a tool nobody wrote a test for cannot quietly lose it again.

## Where it lives

| Piece | Path |
| --- | --- |
| The gate (one table, one registration point) | `internal/mcp/executor/corpus_tolerance_gate_test.go` |
| The shared corpus fixture and its installer | `internal/testutil/corpus.go` |
| The fixture files and what each shape is for | `internal/testutil/testdata/corpus/README.md` |

The fixture is embedded with `go:embed`, so any package can import it without
resolving a path relative to `testutil`'s own directory.

## Running it

```bash
go test -short -run TestCorpusToleranceGate ./internal/mcp/executor/
```

It is `-short`-clean and runs in well under a second, so it is on the
`make commit` path (`make commit` → `test` → `go test -short ./...`). It never
reads your real session history: `testutil.IsolatedProjectsRoot` points the
ambient lookups at a `t.TempDir()`.

## The three contract kinds

Each registry entry declares one `toleranceKind`, which selects the assertion
set the gate applies. The kind is a statement about how far that tool's
tolerance actually reaches today — not an aspiration.

| Kind | Meaning | Gate asserts |
| --- | --- | --- |
| `kindEnumerationReported` | Enumerates the corpus, skips unusable files, names every one it skipped | control results survive **and** every skipped session ID appears in `warnings`/`skipped_files` |
| `kindEnumerationSilent` | Enumerates the corpus and returns the usable results, but does not say what it skipped | control results survive; the reporting half is **not** asserted, and the entry's `note` records the gap |
| `kindSchemaSpecific` | The response is not a corpus result set (an inventory, or explicit file paths were passed) | the entry's own `assertSurvives` shows the control session still reached the response and the unusable files did not fail the call |

`kindEnumerationSilent` exists because the reporting half of the contract is
genuinely not implemented on every path. Recording that in one place, per tool,
is the point: the registry is an honest inventory of where the contract holds,
not a claim that it holds everywhere.

## Registering a new corpus-consuming tool

A tool is **corpus-consuming** if any code path it can reach opens a session
transcript file it did not receive as an explicit argument — typically anything
that resolves sessions from `working_dir`, `scope: "project"`, or the ambient
projects root.

**Step 1 — add one entry to `corpusToolGates()`.** That function is the single
registration point; one entry per tool, keyed by the tool name as it appears in
`toolspkg.GetToolDefinitions()`.

```go
{
    name:     "my_new_tool",
    kind:     kindEnumerationReported,
    mustName: allCorruptShapes(),
    argsFor:  enumeratingArgs(nil),
},
```

- `name` must match the registered MCP tool name exactly. A name that is not a
  defined tool fails the completeness test.
- `argsFor` builds the arguments for a given corpus. Use `enumeratingArgs(extra)`
  for a tool that discovers its corpus from `working_dir`, or `fileArgs()` /
  `fileArgsWith(extra)` for one pointed at explicit paths.
- `mustName` lists the corruption shapes whose session IDs the response must
  contain. Use `allCorruptShapes()` rather than naming shapes individually, so
  the entry keeps naming everything when the fixture grows.
- If the tool needs a quiet response (the gate compares bodies), pass
  `"inline_threshold_bytes": 1 << 20` in `extra`: a file_ref envelope carries a
  random temp-file path that differs between corpora for reasons unrelated to
  tolerance. `query_sessions` does this.

**Step 2 — for `kindSchemaSpecific`, also supply `assertSurvives`.** It
receives the clean and dirty corpora and both raw responses, and must check
against that tool's **own** response schema that the control session survived.
The five existing helpers (`assertDirectoryAccountsForEveryFile`,
`assertMetadataListsEveryFile`, `assertInspectReportsControlIntact`,
`assertStage2KeepsControlRecords`, `assertEditSequencesUnchanged`) are the
models to copy.

**Step 3 — if the tool departs from the full contract, say so in `note`.** A
`note` is required in spirit for any entry that is not
`kindEnumerationReported` + `allCorruptShapes()`: it is where the reason a tool
tolerates silently, or is schema-specific, is recorded for the next reader.

**Step 4 — run the gate.** `TestCorpusToleranceGate_EveryToolIsClassified`
fails if a defined tool is neither registered nor declared non-corpus-consuming,
naming the tool and pointing back at this page. That test is the enforcement:
you cannot add a tool and forget this file, because the commit path goes red.

### A tool that is genuinely not corpus-consuming

Declare it in `nonCorpusTools` with a reason instead of registering it:

```go
var nonCorpusTools = map[string]string{
    "cleanup_temp_files": "removes old MCP temp files; never opens a session transcript",
}
```

The reason is required and must be non-empty: "not corpus-consuming" has to be a
checked claim, not an omission. Declaring a tool in both places is an error.

## Adding a new corruption shape

Add the file to `internal/testutil/testdata/corpus/`, then add one
`testutil.CorruptFile` entry to `CorruptFiles()` with the shape constant, the
session ID to install it under, and a `Why` explaining which layer it fails at.
Document it in the fixture's `README.md` table as well. Every entry that used
`allCorruptShapes()` picks it up automatically; entries that name shapes
individually must be revisited, which is exactly why `allCorruptShapes()` is
preferred.

The three shapes fail at deliberately different layers — 0 bytes parses with
**no error** and zero entries (the silent-loss case), truncation is a parse
error, and the wrong-shape file is valid JSON that fails as a record. A rewrite
that special-cases one route is still caught by the others.

## Why the gate has a negative control

`TestCorpusToleranceGate_RejectsAnIntolerantTool` feeds the verdict helpers the
response of a tool that *did* lose the corpus — both the whole-batch-failure
shape and the silently-shorter-result-set shape — and fails if the gate would
have accepted either. Without it, a gate that had quietly stopped checking
anything would look identical to a gate that passes.

## See Also

- [MCP Testing Quickstart](mcp-testing-quickstart.md) — running and writing MCP
  tool tests
- [Build Quality Gates](build-quality-gates.md) — what `make commit` runs
- [Troubleshooting](troubleshooting.md) — "Malformed session file tolerated but
  no warning appears" (a stale installed binary)
