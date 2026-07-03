# Documentation Session Signals: File Type Classification and Co-Access Pairs

> Status: Implemented (fileType/docRole on EditEvent, CoAccessedDocs/DocVoid/SpecPrecisionGap on FileEditSequence, globalDocStats backfill — all in internal/analyzer/edit_sequences.go; verified 2026-06-23)
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
