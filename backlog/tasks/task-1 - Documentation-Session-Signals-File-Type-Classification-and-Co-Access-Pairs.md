---
id: TASK-1
title: 'Documentation Session Signals: File Type Classification and Co-Access Pairs'
status: 'Basic: Done'
assignee: []
created_date: '2026-06-22 23:52'
updated_date: '2026-06-23 00:41'
labels:
  - 'kind:basic'
dependencies: []
ordinal: 1000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
# Documentation Session Signals: File Type Classification and Co-Access Pairs

> Status: Draft (rev 1)
> Scope: Extend `query_edit_sequences` to classify files by type (source/doc/config),
>        compute per-file doc role (spec/output/mixed), surface session co-access pairs,
>        and emit `doc_void` / `specPrecisionGap` boolean flags — all mechanically, no LLM
> Branch: `feat/doc-session-signals` (future)
> Depends on: `proposal-edit-sequence-tool.md`
> Consumed by: archguard `proposal-doc-code-sync-analysis.md` (CCB assembly)

---

## Background

`query_edit_sequences` (see `proposal-edit-sequence-tool.md`) returns an ordered
Read/Edit timeline per file and classifies source files as Pattern A/B/C. It currently
treats all files uniformly — `.md` documentation files, `.ts` source files, and `.json`
config files are all mixed into the same event stream without distinction.

Empirical analysis of archguard's session history shows that documentation files are not
noise: **17.3% of all file touches are `.md` files** (42 files, 129 touches), and they
play two structurally different roles:

**Role REF (Specification Reference)** — heavily read, rarely edited. Claude consults
these before and during implementation as an external memory source.

```
plan-73-81-format-encoding-experiment.md   R=12, E=6
plan-59-66-intrinsic-dimension-...md       R= 4, E=1
docs/dev-guide/architecture.md             R= 3, E=1
```

**Role OUTPUT (Generated Artifact)** — heavily edited, rarely read. Claude writes these
as deliverables; they are not consulted as input.

```
experiments/format-encoding/REPORT.md      R= 2, E=17
proposal-intrinsic-dimension-...md         R= 0, E= 5
```

Knowing which `.md` files were co-accessed alongside a source file in the same session
unlocks two mechanical signals that the current tool cannot produce:

- **`doc_void`**: a Pattern B source file had no spec doc co-accessed in any session →
  the LLM navigated high-iteration work with no written specification to consult
- **`specPrecisionGap`**: a Pattern B source file had a spec doc co-accessed, yet
  iteration remained high → the spec exists but was insufficiently precise

Both signals are fully mechanical (boolean logic on computed fields). They are emitted as
flags for the archguard CCB assembler to interpret and convert into natural-language
guidance.

---

## Goals

- Add `fileType` and `docRole` fields to `EditEvent` (extension matching + ratio
  thresholds, no LLM)
- Add `CoAccessedDocs` to `FileEditSequence`: doc files co-accessed in the same sessions
  as this source file, with per-doc read counts and role classification
- Compute `DocVoid` and `SpecPrecisionGap` boolean flags as derived fields on
  `FileEditSequence`, mechanically from the above
- Expose all new fields through the existing `query_edit_sequences` MCP tool response
  (no new tool needed)

---

## Non-Goals

- Analyzing the *content* of documentation files (that is LLM reasoning, not meta-cc's job)
- Determining whether a spec document is *correct* (LLM reasoning)
- Treating test files (`.test.ts`, `_test.go`) as documentation — they are already
  handled by archguard's test analysis layer
- Accessing git history — `docFreshnessGap` requires co-change data that only archguard
  holds; it is out of scope for this proposal

---

## Design

### 1. File Type Classification

Extend `EditEvent` in `internal/analyzer/edit_sequences.go`:

```go
type EditEvent struct {
    Timestamp   string `json:"timestamp"`
    SessionID   string `json:"sessionId"`
    Tool        string `json:"tool"`
    ContentHint string `json:"contentHint"`
    FileType    string `json:"fileType"`            // "source" | "doc" | "config" | "other"
    DocRole     string `json:"docRole,omitempty"`   // "spec" | "output" | "mixed" — doc files only
}
```

`FileType` rules (extension matching, no LLM):

| Extension | FileType |
|---|---|
| `.md`, `.rst`, `.txt` | `doc` |
| `.ts`, `.go`, `.py`, `.java`, `.cpp`, `.rs`, `.kt` | `source` |
| `.json`, `.yaml`, `.toml`, `.env`, `.lock` | `config` |
| anything else | `other` |

`DocRole` rules for `doc` files (ratio thresholds on the file's own aggregate counts,
no LLM):

| Condition | DocRole |
|---|---|
| `readEditRatio >= 3.0` | `spec` |
| `readEditRatio <= 0.5` AND `totalEdits >= 3` | `output` |
| otherwise | `mixed` |

`DocRole` is set at the `FileEditSequence` level (per-file aggregate), not per-event. An
individual `EditEvent` for a doc file carries the file-level `DocRole` as a convenience
denormalization.

### 2. Session Co-Access Pairs

New fields on `FileEditSequence`: the set of **doc** files that were touched (Read or
Edit by any tool) in any session that also touched this source file.

```go
type FileEditSequence struct {
    // ... existing fields from proposal-edit-sequence-tool.md ...
    CoAccessedDocs []CoAccessedDoc `json:"coAccessedDocs,omitempty"`
    DocVoid            bool        `json:"docVoid"`
    SpecPrecisionGap   bool        `json:"specPrecisionGap"`
}

type CoAccessedDoc struct {
    FilePath      string `json:"filePath"`
    DocRole       string `json:"docRole"`       // "spec" | "output" | "mixed"
    CoAccessCount int    `json:"coAccessCount"` // number of sessions where both were touched
    TotalDocReads int    `json:"totalDocReads"` // total Read calls to this doc across all sessions
}
```

`CoAccessedDocs` is populated by grouping all events by `SessionID`, then for each
session that touched the target source file, collecting every doc-type file also touched
in that session. Results are aggregated across sessions and sorted by `CoAccessCount`
descending.

Only `fileType == "doc"` files appear in `CoAccessedDocs`. Config and other source files
are excluded.

**Example — documentation void:**

```json
{
  "filePath": "src/plugins/golang/atlas/builders/flow-graph-builder.ts",
  "totalReads": 9,
  "totalEdits": 15,
  "patternHint": "B",
  "coAccessedDocs": [],
  "docVoid": true,
  "specPrecisionGap": false
}
```

**Example — spec precision gap:**

```json
{
  "filePath": "experiments/format-encoding/lib/corpus.ts",
  "totalReads": 3,
  "totalEdits": 13,
  "patternHint": "B",
  "coAccessedDocs": [
    {
      "filePath": "docs/plans/plan-73-81-format-encoding-experiment.md",
      "docRole": "spec",
      "coAccessCount": 3,
      "totalDocReads": 12
    }
  ],
  "docVoid": false,
  "specPrecisionGap": true
}
```

### 3. `doc_void` Flag (mechanical)

```go
DocVoid = patternHint == "B"
          && len(CoAccessedDocs) == 0
          && float64(sessionReadCount) < float64(sessionEditCount) * 0.8
```

The third condition (`reads < edits × 0.8`) excludes the case where extra source-code
reads compensate for the absence of doc reads. If the LLM read the source file many more
times than it edited it, the absence of a spec doc is less critical.

### 4. `specPrecisionGap` Flag (mechanical)

```go
SpecPrecisionGap = patternHint == "B"
                   && any(CoAccessedDocs, func(d CoAccessedDoc) bool {
                          return d.DocRole == "spec"
                      })
                   && maxTotalDocReads(CoAccessedDocs) >= 3
```

`maxTotalDocReads` returns the highest `TotalDocReads` value across all spec-role docs
in `CoAccessedDocs`. The threshold of 3 filters out incidental single-session reads;
a doc consulted ≥ 3 times was genuinely load-bearing.

Both flags are computed inside `BuildEditSequences` immediately after populating
`CoAccessedDocs`. No additional pass is needed.

---

## Output Contract for Consumers

Downstream consumers (archguard CCB assembler) read these fields from the
`query_edit_sequences` response:

```
FileEditSequence.CoAccessedDocs   → which spec docs were co-accessed
FileEditSequence.DocVoid          → boolean flag (no spec, high iteration)
FileEditSequence.SpecPrecisionGap → boolean flag (spec exists, still high iteration)
EditEvent.FileType                → filter events by source/doc/config
EditEvent.DocRole                 → role of each doc event
```

The flags are sensors only. Interpretation and natural-language guidance are produced by
the LLM in archguard's CCB assembler, not here.

---

## Plan

| Phase | Work |
|---|---|
| 1 | Add `FileType` classification to `EditEvent` (extension matching) |
| 2 | Add `DocRole` classification to `FileEditSequence` for doc-type files (ratio thresholds) |
| 3 | Add `CoAccessedDocs` computation to `BuildEditSequences` (GROUP BY sessionId + filter doc files) |
| 4 | Add `DocVoid` and `SpecPrecisionGap` boolean derivations |
| 5 | Expose all new fields in `query_edit_sequences` MCP tool response |
| 6 | Unit tests (see below) |

---

## Test Coverage Requirements

Per project conventions (≥ 80% coverage):

- `fileType` for each extension bucket (source / doc / config / other)
- `docRole`: REF path (ratio ≥ 3.0), OUTPUT path (ratio ≤ 0.5, edits ≥ 3), MIXED fallback
- `coAccessedDocs`: single session with one source + one doc → one entry; two sessions
  with overlapping docs → aggregated counts; doc file touched in session without source
  file → not included
- `docVoid=true`: Pattern B source, empty coAccessedDocs, reads < edits × 0.8
- `docVoid=false`: Pattern B source, empty coAccessedDocs, but reads ≥ edits × 0.8
  (extra reads compensate)
- `specPrecisionGap=true`: Pattern B + spec doc with totalDocReads ≥ 3
- `specPrecisionGap=false`: Pattern B + spec doc with totalDocReads < 3 (threshold not met)
- `specPrecisionGap=false`: Pattern C source (ratio within balanced range) + spec doc
- Integration: run against archguard project sessions, assert
  `flow-graph-builder.ts → docVoid=true` and
  `corpus.ts → specPrecisionGap=true`
<!-- SECTION:DESCRIPTION:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
# Proposal: Documentation Session Signals — File Type Classification and Co-Access Pairs

## Background

`query_edit_sequences` returns an ordered Read/Edit timeline per file and classifies source
files as Pattern A/B/C. It currently treats all files uniformly — `.md` documentation files,
`.ts` source files, and `.json` config files are mixed into the same event stream without
distinction. Empirical analysis of archguard session history shows that 17.3% of all file
touches are `.md` files (42 files, 129 touches) and they play two structurally different roles:
spec-reference docs (heavily read, rarely edited) and output docs (heavily edited, rarely read).
Knowing which docs were co-accessed alongside a source file unlocks two fully mechanical boolean
signals — `doc_void` (Pattern B source had no spec doc co-accessed → navigated with no spec) and
`specPrecisionGap` (Pattern B source had a spec co-accessed yet iteration remained high → spec
was insufficiently precise) — that the current tool cannot produce and the archguard CCB
assembler needs.

## Goals

1. `fileType` field added to `EditEvent` classifying each file as `source`, `doc`, `config`, or
   `other` via extension matching (no LLM), with all four buckets verified by unit tests.
2. `docRole` field added to `FileEditSequence` for doc-type files as `spec`, `output`, or
   `mixed` via ratio thresholds (`readEditRatio >= 3.0` → spec; `<= 0.5` AND `edits >= 3` →
   output; otherwise mixed), verified by unit tests for all three paths.
3. `coAccessedDocs []CoAccessedDoc` added to `FileEditSequence`: doc files co-accessed in the
   same sessions as this source file, with per-doc `coAccessCount` and `totalDocReads`,
   populated by GROUP BY sessionId, verified by unit tests for aggregation and exclusion.
4. `docVoid bool` and `specPrecisionGap bool` derived boolean flags computed in
   `BuildEditSequences` from the above fields per the specified formulas, verified by unit tests
   for all true/false boundary cases.
5. All four new fields exposed in the existing `query_edit_sequences` MCP tool response with no
   new tool needed and no breaking change to existing fields.

## Proposed Approach

Extend `internal/analyzer/edit_sequences.go` in four incremental passes:
1. **FileType classification**: add `FileType` and `DocRole` string fields to `EditEvent`; add a
   `classifyFileType(path string) string` helper using a switch on file extension; call it for
   every event during `BuildEditSequences`.
2. **DocRole per sequence**: after all events for a `FileEditSequence` are collected, compute
   `readEditRatio = totalReads / totalEdits` and assign `DocRole` if `fileType == "doc"`;
   back-fill `DocRole` onto each event in the sequence as a convenience denormalization.
3. **CoAccessedDocs**: build a `sessionIndex map[string][]string` (sessionID → file paths
   touched); for each source `FileEditSequence`, iterate its sessions, collect doc-type file
   paths, aggregate `coAccessCount` and `totalDocReads`, sort by `coAccessCount` descending.
4. **Boolean flags**: immediately after populating `CoAccessedDocs`, compute `DocVoid` and
   `SpecPrecisionGap` using the formulas in the proposal; no additional pass needed.
Expose all new fields in the MCP response by updating the JSON serialization of
`FileEditSequence` and `EditEvent` (no new tool handler, no schema version bump needed).

## Trade-offs and Risks

- **Not doing**: content analysis of doc files (LLM reasoning), correctness judgement of specs,
  treatment of `.test.ts`/`_test.go` as docs, git-history-based `docFreshnessGap` (requires
  archguard co-change data — out of scope).
- **Risk**: `readEditRatio` is undefined when `totalEdits == 0`; guard with a fallback to
  `"spec"` or skip `DocRole` assignment for unedited docs.
- **Risk**: `query_edit_sequences` currently depends on `proposal-edit-sequence-tool.md` being
  implemented; this proposal assumes that prerequisite is already in place.
- **Risk**: the integration test (`flow-graph-builder.ts → docVoid=true`) depends on live
  archguard session data being available in the test environment; mock data may be needed.

---

# Plan: Documentation Session Signals — File Type Classification and Co-Access Pairs

Proposal: docs/proposals/proposal-doc-session-signals.md

Note: `query_edit_sequences` (prerequisite per `proposal-edit-sequence-tool.md`) is not yet
implemented. This plan implements both the base tool and the doc signal extensions in one task.

## Phase A: Core data structures + classifyFileType

### Tests (write first)
File: `internal/analyzer/edit_sequences_test.go`
- `TestClassifyFileType_Doc`: `.md`, `.rst`, `.txt` → `"doc"`
- `TestClassifyFileType_Source`: `.go`, `.ts`, `.py`, `.java`, `.cpp`, `.rs`, `.kt` → `"source"`
- `TestClassifyFileType_Config`: `.json`, `.yaml`, `.toml`, `.env`, `.lock` → `"config"`
- `TestClassifyFileType_Other`: `.png`, `.bin`, `""` → `"other"`
- `TestEditEventJSON`: `EditEvent` JSON round-trips with `fileType`, `docRole` fields
- `TestFileEditSequenceJSON`: `FileEditSequence` JSON includes `coAccessedDocs`, `docVoid`, `specPrecisionGap`

### Implementation
File: `internal/analyzer/edit_sequences.go`
- Add `EditEvent` struct: `Timestamp`, `SessionID`, `Tool`, `ContentHint`, `FileType`, `DocRole` (omitempty), `Content` (*EditContent, omitempty)
- Add `EditContent` struct: `OldString`, `NewString`
- Add `CoAccessedDoc` struct: `FilePath`, `DocRole`, `CoAccessCount`, `TotalDocReads`
- Add `FileEditSequence` struct: `SessionCount`, `TotalReads`, `TotalEdits`, `ReadEditRatio`, `PatternHint`, `Events []EditEvent`, `CoAccessedDocs []CoAccessedDoc`, `DocVoid bool`, `SpecPrecisionGap bool`
- Add `EditSequenceSummary` struct: `TotalFiles int`, `PatternDistribution map[string]int`
- Add `EditSequencesResult` struct: `Files map[string]FileEditSequence`, `Summary EditSequenceSummary`
- Add `classifyFileType(path string) string` helper (switch on `filepath.Ext`)
- Add empty `BuildEditSequences(entries []types.SessionEntry, files []string, includeContent bool, limitPerFile int) EditSequencesResult` stub (returns empty result)

### DoD
- [ ] `go test ./internal/analyzer/... -run TestClassifyFileType`
- [ ] `go test ./internal/analyzer/... -run TestEditEventJSON`
- [ ] `go test ./internal/analyzer/... -run TestFileEditSequenceJSON`
- [ ] `go build ./...`

---

## Phase B: BuildEditSequences — event collection + PatternHint + DocRole

### Tests (write first)
File: `internal/analyzer/edit_sequences_test.go`
- `TestBuildEditSequences_EmptyEntries`: empty input → empty result
- `TestBuildEditSequences_SingleFile`: 2 Read + 1 Edit for one file → correct counts, ratio, events sorted by timestamp
- `TestBuildEditSequences_PatternA`: readEditRatio ≥ 3.0 → patternHint `"A"`
- `TestBuildEditSequences_PatternB`: readEditRatio ≤ 0.8 AND totalEdits ≥ 5 → patternHint `"B"`
- `TestBuildEditSequences_PatternC`: ratio 1.5, edits 2 → patternHint `"C"`
- `TestBuildEditSequences_DocRoleSpec`: doc file with readEditRatio ≥ 3.0 → docRole `"spec"`
- `TestBuildEditSequences_DocRoleOutput`: doc file with ratio ≤ 0.5 AND edits ≥ 3 → docRole `"output"`
- `TestBuildEditSequences_DocRoleMixed`: doc file with ratio 1.5 → docRole `"mixed"`
- `TestBuildEditSequences_FileTypeOnEvents`: events for `.md` file have `fileType="doc"` and `docRole` denormalized
- `TestBuildEditSequences_ContentHint`: Read event → contentHint `"file_path=..."`, Edit → `"old: '...' → new: '...'"`
- `TestBuildEditSequences_LimitPerFile`: limitPerFile=3 truncates events to 3

### Implementation
File: `internal/analyzer/edit_sequences.go` — implement `BuildEditSequences`:
1. Build `uuidToSessionID map[string]string` from entries (uuid → sessionID)
2. Call `types.ExtractToolCalls(entries)`, filter to `Read`/`Edit`/`Write` tools only (via `types.FileActionType`)
3. Extract file path from `tc.Input["file_path"]` (string assertion); skip if empty
4. If `files` param non-empty, skip files not in the set
5. Group events by file path; for each event append `EditEvent{Timestamp, SessionID, Tool, ContentHint, FileType}`
6. Compute `contentHint`: Read → `"file_path=<path>"`, Edit → `"old: '<40chars>' → new: '<40chars>'"`, Write → `"write <len> bytes to <path>"`
7. After grouping: sort events by `Timestamp`; compute `TotalReads`, `TotalEdits`, `SessionCount` (unique sessionIDs); `ReadEditRatio`
8. Apply `patternHint` rule; apply `docRole` for `fileType=="doc"` files; denormalize `FileType`+`DocRole` onto each event
9. Apply `limitPerFile` truncation; populate `EditSequenceSummary`
10. Return `EditSequencesResult`

### DoD
- [ ] `go test ./internal/analyzer/... -run TestBuildEditSequences`
- [ ] `go build ./...`

---

## Phase C: CoAccessedDocs + DocVoid + SpecPrecisionGap

### Tests (write first)
File: `internal/analyzer/edit_sequences_test.go`
- `TestCoAccessedDocs_SingleSession`: source + doc in same session → one `CoAccessedDoc` entry
- `TestCoAccessedDocs_TwoSessions`: source + same doc in 2 sessions → coAccessCount=2, totalDocReads aggregated
- `TestCoAccessedDocs_DocOnlySession`: doc touched in session without source → NOT in coAccessedDocs
- `TestCoAccessedDocs_ConfigExcluded`: `.json` file co-accessed → excluded from coAccessedDocs
- `TestCoAccessedDocs_SortedByCoAccessCount`: two co-accessed docs → sorted descending by coAccessCount
- `TestDocVoid_True`: patternHint B + empty coAccessedDocs + reads < edits×0.8 → docVoid=true
- `TestDocVoid_FalseByReads`: patternHint B + empty coAccessedDocs + reads ≥ edits×0.8 → docVoid=false
- `TestDocVoid_FalseByPattern`: patternHint C + empty coAccessedDocs → docVoid=false
- `TestSpecPrecisionGap_True`: patternHint B + spec coAccessedDoc with totalDocReads ≥ 3 → specPrecisionGap=true
- `TestSpecPrecisionGap_FalseByThreshold`: patternHint B + spec doc totalDocReads < 3 → specPrecisionGap=false
- `TestSpecPrecisionGap_FalseByPattern`: patternHint C + spec doc → specPrecisionGap=false

### Implementation
File: `internal/analyzer/edit_sequences.go` — add to `BuildEditSequences` after Phase B logic:
1. Build `sessionToFiles map[string][]string` from all tool calls (sessionID → file paths touched, deduped)
2. For each `FileEditSequence` in result:
   a. Collect the set of sessionIDs where this file was accessed
   b. For each such session, collect all doc-type files also in `sessionToFiles[sessionID]`
   c. Aggregate: `coAccessCount` = number of sessions containing both, `totalDocReads` = sum of Read events for that doc file across those sessions
   d. Set `DocRole` on each `CoAccessedDoc` from the doc file's own `FileEditSequence.DocRole`
   e. Sort by `CoAccessCount` descending
3. Compute `DocVoid`: `patternHint=="B" && len(CoAccessedDocs)==0 && float64(TotalReads) < float64(TotalEdits)*0.8`
4. Compute `SpecPrecisionGap`: `patternHint=="B" && any spec-role doc with TotalDocReads≥3`

### DoD
- [ ] `go test ./internal/analyzer/... -run TestCoAccessedDocs`
- [ ] `go test ./internal/analyzer/... -run TestDocVoid`
- [ ] `go test ./internal/analyzer/... -run TestSpecPrecisionGap`
- [ ] `go build ./...`

---

## Phase D: MCP tool registration + schema

### Tests (write first)
File: `internal/mcp/executor/edit_sequences_handler_test.go`
- `TestHandleQueryEditSequences_MissingFiles`: no `files` param → error or empty result
- `TestHandleQueryEditSequences_ValidInput`: valid files + mock entries → response contains `files` map with correct structure
- `TestQueryEditSequencesSchema`: `GetToolDefinitions()` includes `query_edit_sequences` with required `files` array param

### Implementation
New file: `internal/mcp/executor/edit_sequences_handler.go`
- `init()` that calls `registerQueryHandler("query_edit_sequences", handleQueryEditSequences)`
- `handleQueryEditSequences(e *ToolExecutor, scope string, args map[string]interface{}) (mcquery.QueryResult, error)`:
  - Extract `files []string` from args (required)
  - Extract `include_content bool`, `scope string`, `limit_per_file int` (defaults: false, "project", 50)
  - Load entries via `e.loadEntries(scope)`
  - Call `analyzer.BuildEditSequences(entries, files, includeContent, limitPerFile)`
  - Marshal result to JSON, return via hybrid output mode (inline <32KB, file_ref ≥32KB)

Update: `internal/mcp/tools/tools.go` — add to `GetToolDefinitions()`:
```go
BuildTool("query_edit_sequences",
    "Return chronologically ordered Read/Edit events per file from session history. Classifies AI behavioral patterns: A (high reads, low edits), B (high edits, iterative), C (balanced). Includes doc co-access signals: fileType, docRole, coAccessedDocs, docVoid, specPrecisionGap. Default scope: project.",
    map[string]Property{
        "files":           {Type: "array", Items: &Property{Type: "string"}, Description: "Relative file paths to query (required)"},
        "include_content": {Type: "boolean", Description: "Include full old/new string content (default: false)"},
        "scope":           {Type: "string", Description: "project or session (default: project)"},
        "limit_per_file":  {Type: "integer", Description: "Max events per file (default: 50)"},
    },
    "files",
)
```

### DoD
- [ ] `go test ./internal/mcp/... -run TestHandleQueryEditSequences`
- [ ] `go test ./internal/mcp/... -run TestQueryEditSequencesSchema`
- [ ] `go build ./...`

---

## Constraints

- `DocRole` is undefined when `totalEdits == 0`; guard: skip DocRole for unedited doc files (leave empty string)
- `ReadEditRatio` when `totalEdits == 0`: set to `float64(TotalReads)` (treat as very high ratio)
- `includeContent=true` replaces `contentHint` with `content: {oldString, newString}` on each event
- No new MCP tool — all new fields added to existing `query_edit_sequences` response shape
- No breaking changes to existing `FileEditSequence` shape (new fields are additive, omitempty where appropriate)
- The integration test case (`flow-graph-builder.ts → docVoid=true`) requires real archguard session data; unit tests use synthetic fixtures only

## Acceptance Gate

- [ ] `go test ./...`
- [ ] `go build ./cmd/mcp-server/`
- [ ] `go vet ./...`
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Proposal approved. Starting plan review.

Plan review iteration 1: APPROVED
premise-ledger:
[E] goal coverage: all 5 Goals map 1:1 to phases A-D (fileType→A, docRole→A/B, coAccessedDocs→C, docVoid/specPrecisionGap→C, MCP exposure→D)
[E] TDD structure: every phase has ### Tests (write first) before ### Implementation, in correct order
[E] TDD order: first DoD item in each phase is a go test ./internal/... per-package command; proves red→green for that phase
[E] acceptance gate: ## Acceptance Gate block present, first item is go test ./...
[E] DoD executability: all DoD and Acceptance Gate items are shell commands (go test, go build, go vet)
[E] phase ordering: A (structs) → B (BuildEditSequences logic) → C (CoAccessedDocs/flags, depends on B) → D (MCP handler, depends on A+B+C); no circular deps
[E] scope discipline: all phases directly back a proposal Goal; no gold-plating
[E] file paths: all new files (edit_sequences.go, edit_sequences_test.go, edit_sequences_handler.go, edit_sequences_handler_test.go) do not exist yet; existing tools.go confirmed present; types.SessionEntry, ExtractToolCalls, FileActionType, registerQueryHandler, BuildTool, GetToolDefinitions all confirmed in codebase
GCL-self-report: E=8 C=0 H=0

claimed: 2026-06-23T00:12:47Z

Phase A ✓ 2026-06-23T00:27:08Z
Core data structures (EditEvent, FileEditSequence, CoAccessedDoc, EditSequencesResult) + classifyFileType
DoD #1: PASS — go test ./internal/analyzer/... -run TestClassifyFileType
DoD #2: PASS — go test ./internal/analyzer/... -run TestEditEventJSON
DoD #3: PASS — go test ./internal/analyzer/... -run TestFileEditSequenceJSON

Phase B ✓ 2026-06-23T00:27:08Z
BuildEditSequences: event collection, PatternHint (A/B/C), DocRole, ContentHint, limitPerFile
DoD #4: PASS — go test ./internal/analyzer/... -run TestBuildEditSequences

Phase C ✓ 2026-06-23T00:27:08Z
CoAccessedDocs aggregation, DocVoid, SpecPrecisionGap; fixed UUID collision in test helpers
DoD #5: PASS — go test ./internal/analyzer/... -run TestCoAccessedDocs
DoD #6: PASS — go test ./internal/analyzer/... -run TestDocVoid
DoD #7: PASS — go test ./internal/analyzer/... -run TestSpecPrecisionGap

Phase D ✓ 2026-06-23T00:27:08Z
MCP tool registration (query_edit_sequences handler + schema); updated tool count tests to 22
DoD #8: PASS — go test ./internal/mcp/executor/... -run TestHandleQueryEditSequences
DoD #9: PASS — go test ./internal/mcp/executor/... -run TestQueryEditSequencesSchema

Acceptance gate: go test ./... PASS, go build ./cmd/mcp-server/ PASS, go vet ./... PASS

Completed: 2026-06-23T00:28:38Z

Re-claimed: 2026-06-23T00:29:05Z — previous agent did not commit

## Execution Summary
Result: Done
Commit: $(cd /home/yale/work/meta-cc-TASK-1 && git rev-parse HEAD)
## Execution Summary
Result: Done
Commit: 5607593c89c78824ec74d8e981b7c3c086b6cefa

Completed: 2026-06-23T00:41:03Z
<!-- SECTION:NOTES:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 go test ./internal/analyzer/... -run TestClassifyFileType
- [ ] #2 go test ./internal/analyzer/... -run TestEditEventJSON
- [ ] #3 go test ./internal/analyzer/... -run TestFileEditSequenceJSON
- [ ] #4 go build ./...
- [ ] #5 go test ./internal/analyzer/... -run TestBuildEditSequences
- [ ] #6 go build ./...
- [ ] #7 go test ./internal/analyzer/... -run TestCoAccessedDocs
- [ ] #8 go test ./internal/analyzer/... -run TestDocVoid
- [ ] #9 go test ./internal/analyzer/... -run TestSpecPrecisionGap
- [ ] #10 go build ./...
- [ ] #11 go test ./internal/mcp/... -run TestHandleQueryEditSequences
- [ ] #12 go test ./internal/mcp/... -run TestQueryEditSequencesSchema
- [ ] #13 go build ./...
- [ ] #14 go test ./...
- [ ] #15 go build ./cmd/mcp-server/
- [ ] #16 go vet ./...
<!-- DOD:END -->
