# query_edit_sequences: Ordered Read/Edit Timeline per File

> Status: Implemented (2026-06-23, see internal/analyzer/edit_sequences.go + internal/mcp/executor/edit_sequences_handler.go)
> Scope: New MCP tool — exposes ordered Read/Edit event sequences per file from session
>        history; enables LLM to classify AI behavioral patterns (A/B/C)
> Related: archguard `proposal-cognitive-analysis-layer.md`

---

## Background

meta-cc currently exposes file access data in two forms:

1. **Aggregate counts** via `churn.go`: `{ file, reads, edits, writes, total }` — useful
   for identifying high-churn files, but discards the temporal ordering of events
2. **Raw tool blocks** via `query_tool_blocks`: all tool calls in a flat JSONL stream —
   ordering is preserved but requires the LLM to filter, group, and sort per-file manually

Neither form supports the primary use case for the Cognitive Analysis Layer (see archguard
`proposal-cognitive-analysis-layer.md`): classifying a file as Pattern A (high reads, low
edits — conceptually dense), Pattern B (high edits, iterative — hard to converge), or
Pattern C (balanced — healthy development).

Pattern classification requires not just counts but **sequences**: whether Claude read the
file many times before a single edit (Pattern A signal), or alternated read-edit-read-edit
in short cycles (Pattern B signal). This temporal structure is in the JSONL data but is
currently discarded by the aggregation layer.

### Concrete Example from archguard Sessions

```
src/core/query/query-engine.ts  →  R=11, E=2  (Pattern A)
  Timeline:
    09:14 Read  (line range: 1-50, orientation)
    09:31 Read  (line range: 80-120, QueryEngine.findEntity)
    10:02 Read  (line range: 150-200, scope resolution)
    10:45 Read  (checking how outputScope is applied)
    11:03 Read  (checking dependency filter logic)
    14:22 Edit  (add totalPackageCount field)
    15:10 Read  (verifying the change fits the pattern)
    ...
    [4 more Reads across subsequent sessions]
    [1 more Edit in a different session]

src/plugins/golang/atlas/builders/flow-graph-builder.ts  →  R=9, E=15  (Pattern B)
  Timeline:
    08:43 Read   (initial orientation)
    08:51 Edit   (attempt 1: add followIndirectCalls)
    09:02 Read   (test failure — re-reading to understand)
    09:08 Edit   (attempt 2: fix BFS traversal)
    09:19 Read   (test failure again)
    09:25 Edit   (attempt 3: cycle guard)
    ...
```

This sequence data is present in JSONL tool_use blocks (timestamps, tool names, file paths)
but inaccessible to LLM through current tools without expensive raw block filtering.

---

## Goals

- Add `query_edit_sequences` MCP tool that returns chronologically ordered Read/Edit events
  for one or more files, grouped by file
- Include enough context per event (timestamp, tool name, content hint) to allow LLM to
  classify the access pattern without reading full old_string/new_string content
- Support multi-session aggregation (all sessions for the project, not just the current one)
- Keep output compact — designed to fit in LLM tool result inline mode (< 32KB)

---

## Non-Goals

- Returning full `old_string`/`new_string` content by default (available as an opt-in
  parameter `includeContent: true`)
- Line-level granularity (which lines were read — not available without instrumentation)
- Bash command sequences (only Read/Edit/Write tool calls are in scope)
- Real-time streaming of events

---

## Design

### Tool Signature

```go
// Tool name: query_edit_sequences
// Input:
type QueryEditSequencesInput struct {
    Files         []string `json:"files"`          // relative file paths (required)
    IncludeContent bool    `json:"include_content"` // include old_string/new_string (default: false)
    Scope         string  `json:"scope"`           // "project" | "session" (default: "project")
    LimitPerFile  int     `json:"limit_per_file"`  // max events per file (default: 50)
}
```

### Output Format

```json
{
  "files": {
    "src/core/query/query-engine.ts": {
      "sessionCount": 4,
      "totalReads": 11,
      "totalEdits": 2,
      "readEditRatio": 5.5,
      "patternHint": "A",
      "events": [
        {
          "timestamp": "2026-06-10T09:14:22Z",
          "sessionId": "abc123",
          "tool": "Read",
          "contentHint": "file_path=src/core/query/query-engine.ts"
        },
        {
          "timestamp": "2026-06-10T14:22:05Z",
          "sessionId": "abc123",
          "tool": "Edit",
          "contentHint": "old: 'return { entityCount' → new: 'return { totalPackageCount, entityCount'"
        }
      ]
    },
    "src/plugins/golang/atlas/builders/flow-graph-builder.ts": {
      "sessionCount": 2,
      "totalReads": 9,
      "totalEdits": 15,
      "readEditRatio": 0.6,
      "patternHint": "B",
      "events": [...]
    }
  },
  "summary": {
    "totalFiles": 2,
    "patternDistribution": { "A": 1, "B": 1, "C": 0 }
  }
}
```

### `patternHint` Classification Rule

The tool computes `patternHint` mechanically from ratios — the LLM validates and may
override based on sequence shape:

| Condition | `patternHint` |
|---|---|
| `readEditRatio >= 3.0` | `"A"` |
| `readEditRatio <= 0.8` AND `totalEdits >= 5` | `"B"` |
| otherwise | `"C"` |

The LLM uses the `events` timeline to confirm or override: a file with ratio=2.0 that has
`[Read, Read, Edit, Read, Read, Edit]` vs `[Edit, Read, Edit, Read, Edit]` carries
different cognitive signals even at the same aggregate ratio.

### `contentHint` Field

When `includeContent: false` (default), `contentHint` is a short human-readable summary
of the tool input:
- For Read: `"file_path=<path>"` (or with `limit` if present)
- For Edit: `"old: '<first 40 chars of old_string>' → new: '<first 40 chars of new_string>'"`
- For Write: `"write <byte_count> bytes to <path>"`

This gives the LLM enough signal to understand what changed without the full diff content.

When `includeContent: true`, `contentHint` is replaced with `content: { oldString, newString }`.
This mode is for deep forensic analysis only and may produce large output.

---

## Implementation

### New file: `internal/analyzer/edit_sequences.go`

```go
package analyzer

type EditEvent struct {
    Timestamp   string `json:"timestamp"`
    SessionID   string `json:"sessionId"`
    Tool        string `json:"tool"`        // "Read" | "Edit" | "Write"
    ContentHint string `json:"contentHint"`
    Content     *EditContent `json:"content,omitempty"` // only when includeContent=true
}

type EditContent struct {
    OldString string `json:"oldString,omitempty"`
    NewString string `json:"newString,omitempty"`
}

type FileEditSequence struct {
    SessionCount   int         `json:"sessionCount"`
    TotalReads     int         `json:"totalReads"`
    TotalEdits     int         `json:"totalEdits"`
    ReadEditRatio  float64     `json:"readEditRatio"`
    PatternHint    string      `json:"patternHint"` // "A" | "B" | "C"
    Events         []EditEvent `json:"events"`
}

func BuildEditSequences(entries []SessionEntry, files []string, includeContent bool, limitPerFile int) map[string]FileEditSequence
```

The function iterates over all `tool_use` blocks, filters for `Read`/`Edit`/`Write` tools,
extracts the file path from the tool `input` map, groups and sorts by timestamp, then
applies the `patternHint` classification rule.

### Handler: `internal/mcp/executor/edit_sequences_executor.go`

Registers the tool as `query_edit_sequences` under the convenience tools category.
Follows the same hybrid output mode (inline < 32KB, file_ref ≥ 32KB) as existing tools.

### Tool schema: added to `internal/mcp/tools/tools.go`

```go
{
    Name: "query_edit_sequences",
    Description: "Return chronologically ordered Read/Edit events per file from session history. Use to classify AI behavioral patterns: A (high reads, low edits — conceptually dense), B (high edits — hard to converge), C (balanced). Default scope: project.",
    InputSchema: editSequencesInputSchema,
}
```

---

## Plan

| Phase | Work |
|---|---|
| 1 | Implement `BuildEditSequences` in `internal/analyzer/edit_sequences.go` |
| 2 | Add `patternHint` classification rule with ratio thresholds |
| 3 | Implement `contentHint` generation for Read/Edit/Write |
| 4 | Register MCP handler and tool schema |
| 5 | Unit tests: Pattern A file (ratio ≥ 3), Pattern B file (ratio ≤ 0.8, edits ≥ 5), multi-session aggregation, `includeContent=true` |
| 6 | Integration test: run against archguard project sessions, verify `query-engine.ts` → Pattern A, `flow-graph-builder.ts` → Pattern B |
| 7 | Document in `docs/guides/mcp-query-tools.md` under "Behavioral Analysis" section |

---

## Test Coverage Requirements

Per project conventions (≥ 80% coverage):

- `BuildEditSequences` with empty entries → empty map
- Single file, all reads → Pattern A, ratio = Inf → clamped to `totalReads`
- Single file, reads and edits interleaved (Pattern B shape) → Pattern B
- Multi-file input → each file independently classified
- `limitPerFile` truncates events list but not counts
- `scope=session` returns only current session events
- Files not found in any session → `{ sessionCount: 0, events: [] }`
