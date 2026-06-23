# Plan 66–70: JSONL Streaming Reader with Early Image Filtering

**Status**: Completed (see internal/parser/streaming_reader.go)
**Proposal**: [docs/proposals/proposal-jsonl-streaming-reader.md](../proposals/proposal-jsonl-streaming-reader.md)

---

## Overview

Five phases replacing all `bufio.Scanner`-based JSONL readers with `bufio.Reader.ReadBytes`-based streaming reads, adding early image-data truncation to eliminate the 4 MB / 10 MB hard line-length limits that cause session files containing screenshots to fail entirely.

| Phase | Scope | Key deliverable |
|---|---|---|
| 66 | Core infrastructure — `StreamingReader`, migrate `internal/parser/reader.go` | `internal/parser/streaming_reader.go` implemented and tested; `ParseEntries` no longer uses `bufio.Scanner` |
| 67 | MCP query executor — migrate `cmd/mcp-server/query_executor.go` (2 Scanner sites) | `processFile` and `processFileWithTimeRange` use streaming reader |
| 68 | stage2 and filters — migrate `internal/query/jq/stage2_executor.go`, `internal/query/stage2_executor.go`, `internal/mcp/filters/filters.go` | All three stage2/filter paths use streaming reader |
| 69 | Remaining handlers — migrate `cmd/mcp-server/handlers_query.go`, `internal/query/files/file_inspector.go`, `cmd/mcp-server/handlers_stage1.go` | All JSONL-reading paths migrated; `countLines` uses byte-count approach |
| 70 | Cleanup — rename `MaxScannerLineBytes` → `LargeLineWarnBytes`; add lint guard; final verification | Zero `bufio.NewScanner` calls on JSONL paths; lint rule enforced |

**Execution order constraint**: Phase 66 must complete first — it provides `StreamingReader` used by all subsequent phases. Phases 67, 68, and 69 each depend on Phase 66 and may proceed in any order among themselves. Phase 70 runs last.

```
Phase 66 (core infrastructure)
    ├── Phase 67 (query_executor.go)
    ├── Phase 68 (stage2 + filters)
    └── Phase 69 (handlers + file_inspector + countLines)
                  └── Phase 70 (cleanup + lint guard)
```

**Pre-condition**: Phase 65 merged. Phases 66–70 are independent of each other's targets but share the dependency on Phase 66.

---

## Phase 66: Core Infrastructure — StreamingReader

**Goal**: Create `internal/parser/streaming_reader.go` with the `ReadLineFiltered` function and `FilterStrategy` type. Migrate `internal/parser/reader.go` to use it. This is the sole provider of the new I/O primitive for all subsequent phases.

**Pre-condition**: None beyond a clean build.

**Estimated LOC**: ~180 lines added (new file + tests), ~30 lines modified (`reader.go`). Total additions ≤ 200 lines.

**Dependencies**: None.

---

### Stage 66.1 — Implement `streaming_reader.go` with tests (TDD)

**Goal**: Write tests first, then implement `ReadLineFiltered` and `stripImageData` (the byte-level truncation function for strategy B).

**Files**:
- `internal/parser/streaming_reader_test.go` — tests (written first)
- `internal/parser/streaming_reader.go` — implementation

**Test cases** (all must fail before implementation, pass after):

| Test | Scenario | Expected |
|------|----------|---------|
| `TestStripImageData_PlainLine` | No `"type":"image"` present | Line returned unchanged |
| `TestStripImageData_TextToolResult` | `tool_result` with text content only | Line returned unchanged |
| `TestStripImageData_SingleImage` | Single `image` block with 1 MB base64 | `source.data` replaced with `"<binary-omitted>"`; result passes `json.Valid()` |
| `TestStripImageData_MultipleImages` | Two `image` blocks in same line | Both `source.data` values replaced; occurrence count = 2 |
| `TestStripImageData_NonBase64Image` | `"type":"image"` without `"type":"base64"` | Detection condition not met; line unchanged |
| `TestStripImageData_AfterStrip_ValidJSON` | Real-world image JSONL line | `json.Valid()` passes on result |
| `TestStripImageData_InvalidAfterStrip_Fallback` | Malformed JSON that becomes invalid after substitution | `json.Valid()` fails; original bytes and `false` returned |
| `TestReadLineFiltered_StrategyB_LargeImage` | 5 MB image line via `bufio.Reader` | Returns truncated bytes, `false` (not skipped), no error |
| `TestReadLineFiltered_StrategyA_ImageLine` | Image line with `StrategySkipImage` | Returns `nil`, `true` (skipped), no error |
| `TestReadLineFiltered_StrategyA_TextLine` | Non-image line with `StrategySkipImage` | Returns line bytes, `false`, no error |
| `TestReadLineFiltered_EmptyLine` | Empty line (`\n` only) | Returns empty slice, `false`, no error |
| `TestReadLineFiltered_NoTrailingNewline` | Last line in file without `\n` | Returns bytes, `false`, `io.EOF` |
| `TestReadLineFiltered_NormalLineAfterLargeLine` | Large image line followed by normal line | Both readable; second entry has correct UUID |

**Public API** (defined in `streaming_reader.go`):

```go
// FilterStrategy controls how image blocks are handled during streaming read.
type FilterStrategy int

const (
    StrategyDefault    FilterStrategy = iota // Strategy B: truncate source.data in-place
    StrategySkipImage                        // Strategy A: skip entire line if "type":"image" present
)

// ReadLineFiltered reads one line from r using ReadBytes('\n'), then applies
// the given FilterStrategy. Returns (line, skipped, error).
// skipped=true means the line was intentionally omitted (Strategy A).
// On EOF with remaining data, returns the data with io.EOF.
func ReadLineFiltered(r *bufio.Reader, strategy FilterStrategy) ([]byte, bool, error)
```

**Internal helper** (unexported, tested via `ReadLineFiltered`):

```go
// stripImageData replaces "source.data" base64 values in image blocks with
// "<binary-omitted>". Loops until no "type":"base64" remains. Returns
// (processedLine, valid) where valid=false triggers caller to skip the line.
func stripImageData(line []byte) ([]byte, bool)
```

Estimated: ~150 lines (implementation ~80, tests ~70)

Run `make dev` after writing tests (expect failures), then implement until `make commit` passes.

---

### Stage 66.2 — Migrate `internal/parser/reader.go`

**Goal**: Replace `bufio.Scanner` + `scanner.Buffer(buf, MaxScannerLineBytes)` in `ParseEntries` with `ReadLineFiltered(r, StrategyDefault)`.

**Files**:
- `internal/parser/reader.go` — replace Scanner loop with `bufio.Reader` + `ReadLineFiltered`

**Procedure**:
1. Change `bufio.NewScanner(f)` → `bufio.NewReader(f)`
2. Replace `for scanner.Scan()` loop with `for { line, skipped, err := ReadLineFiltered(r, StrategyDefault); ... }`
3. Handle `io.EOF` to terminate the loop
4. Preserve all existing error-handling semantics for JSON parse failures
5. Remove the `buf` pre-allocation and `scanner.Buffer()` call
6. Remove the `MaxScannerLineBytes` import from `internal/parser/aliases.go` usage (will be done in Phase 70; for now, just stop passing it to `scanner.Buffer`)

**Regression tests** (existing tests must continue to pass):
- All tests in `internal/parser/reader_test.go`
- New test `TestParseEntries_LargeImageLine_NotSkipped` added to `reader_test.go` confirming a 5 MB image line no longer causes `ParseEntries` to return an error, and the subsequent normal entry is still present.

Estimated: ~30 lines modified

Run `make commit` after this stage.

---

## Phase 67: MCP Query Executor Migration

**Goal**: Migrate both Scanner sites in `cmd/mcp-server/query_executor.go` to use `ReadLineFiltered`. After this phase, session files containing screenshots are fully readable by the stage1 query path.

**Pre-condition**: Phase 66 complete.

**Estimated LOC**: ~50 lines modified. Well within ≤200 lines per stage.

**Dependencies**: Phase 66.

---

### Stage 67.1 — Migrate `processFileWithTimeRange` and `processFile`

**Goal**: Replace `bufio.Scanner` in both functions with `bufio.Reader` + `ReadLineFiltered(r, StrategyDefault)`.

**Files**:
- `cmd/mcp-server/query_executor.go` — two Scanner sites replaced

**Procedure**:
1. For `processFileWithTimeRange`: replace `bufio.NewScanner` + `scanner.Buffer` → `bufio.NewReader`; replace `for scanner.Scan()` loop; adjust `scanner.Err()` check → loop `io.EOF` termination.
2. For `processFile`: same pattern.
3. Both functions currently `return results, error` on `ErrTooLong` (caller logs warning and continues). After migration, this path is eliminated — large lines are handled transparently. Preserve the `results` accumulation behaviour (partial results on JSON unmarshal error remain as-is).

**Tests** (in `cmd/mcp-server/query_executor_test.go` or equivalent):
- `TestProcessFile_LargeImageLine_NoError` — file with a 5 MB image line; function returns entries without error
- `TestProcessFileWithTimeRange_LargeImageLine_NoError` — same for `processFileWithTimeRange`
- Existing tests must continue to pass

Estimated: ~50 lines modified, ~30 lines added in tests

Run `make commit` after this stage.

---

## Phase 68: stage2 and Filters Migration

**Goal**: Migrate `readJSONLFile` in both stage2 executors and `loadTurnsForSession` in `filters.go` to use streaming reader.

**Pre-condition**: Phase 66 complete.

**Estimated LOC**: ~90 lines across 3 files. All within ≤500 lines for the phase, split into two stages to stay ≤200 each.

**Dependencies**: Phase 66.

---

### Stage 68.1 — Migrate `internal/query/jq/stage2_executor.go` and `internal/query/stage2_executor.go`

**Goal**: Replace Scanner in `readJSONLFile` in both packages. Note: per proposal ISSUE-4, these currently return `nil, error` on `ErrTooLong`, causing entire stage2 queries to fail. After migration, files with large image lines will be processed successfully.

**Files**:
- `internal/query/jq/stage2_executor.go` — `readJSONLFile` Scanner → streaming reader
- `internal/query/stage2_executor.go` — `readJSONLFile` Scanner → streaming reader

**Procedure**:
1. Both `readJSONLFile` functions currently open the file, create a Scanner with 10 MB buffer, loop `scanner.Scan()`, unmarshal each line into `json.RawMessage`, and return `[]json.RawMessage`.
2. Replace Scanner with `bufio.NewReader(f)` + `ReadLineFiltered(r, StrategyDefault)`.
3. The returned `[]json.RawMessage` for image lines will contain the truncated JSON (with `"<binary-omitted>"` placeholder). This is the correct and expected behaviour per Proposal.
4. Update `io.EOF` handling for loop termination.

**Tests**:
- `TestReadJSONLFile_LargeImageLine_ReturnsData` (in each package's test file) — file with 5 MB image line + normal line; function returns 2 results, no error; image entry has `"<binary-omitted>"` in data field
- Previously-failing queries (returning error due to `ErrTooLong`) now return data: update any existing tests that expected error returns on large lines
- Existing passing tests must continue to pass

Estimated: ~50 lines modified, ~30 lines in tests

Run `make dev` after tests are written (expect failures on changed expected values), then implement and run `make commit`.

---

### Stage 68.2 — Migrate `internal/mcp/filters/filters.go`

**Goal**: Replace Scanner in `loadTurnsForSession`. This function currently has a silent `scanner.Err()` bug (error not checked) — fix that as part of the migration.

**Files**:
- `internal/mcp/filters/filters.go` — `loadTurnsForSession` Scanner → streaming reader

**Procedure**:
1. Replace `bufio.NewScanner` + `scanner.Buffer(buf, 10*1024*1024)` with `bufio.NewReader`.
2. Replace `for scanner.Scan()` loop with streaming reader loop.
3. The existing silent `scanner.Err()` bug is eliminated by the new loop structure (loop exits on `io.EOF` or returns error on other read errors).
4. Remove the `10*1024*1024` literal (tracked under proposal acceptance criterion 4).

**Tests**:
- `TestLoadTurnsForSession_LargeImageLine_NoError` — session file with 5 MB image line; function returns turns without error, image turn present with truncated data
- Existing tests in filters package must pass

Estimated: ~25 lines modified, ~20 lines in tests

Run `make commit` after this stage.

---

## Phase 69: Remaining Handlers Migration

**Goal**: Migrate `cmd/mcp-server/handlers_query.go` (`loadTurnsForSession`), `internal/query/files/file_inspector.go` (`InspectFiles`), and `cmd/mcp-server/handlers_stage1.go` (`countLines`). The `countLines` function uses a special byte-counting approach rather than streaming reader.

**Pre-condition**: Phase 66 complete.

**Estimated LOC**: ~80 lines modified, ~50 lines in tests. Within ≤500 lines for the phase, split into two stages.

**Dependencies**: Phase 66.

---

### Stage 69.1 — Migrate `cmd/mcp-server/handlers_query.go` and `internal/query/files/file_inspector.go`

**Goal**: Replace Scanner in `loadTurnsForSession` (handlers side, which also has the silent `scanner.Err()` bug) and in `InspectFiles` (which calls `json.Unmarshal` per line, failing on 10 MB+ lines).

**Files**:
- `cmd/mcp-server/handlers_query.go` — `loadTurnsForSession` Scanner → streaming reader
- `internal/query/files/file_inspector.go` — `InspectFiles` Scanner → streaming reader

**Procedure for `handlers_query.go`**:
1. Same pattern as Stage 68.2: replace Scanner with `bufio.NewReader` + `ReadLineFiltered(r, StrategyDefault)`.
2. Eliminate the silent `scanner.Err()` bug.
3. Remove the `10*1024*1024` literal.

**Procedure for `file_inspector.go`**:
1. `InspectFiles` scans JSONL lines and calls `json.Unmarshal` on each. Replace Scanner (10 MB buffer) with `bufio.NewReader` + `ReadLineFiltered(r, StrategyDefault)`.
2. Image lines will be unmarshalled from the truncated JSON — the `"<binary-omitted>"` placeholder is valid JSON string content, so `json.Unmarshal` will succeed.
3. Remove the `10*1024*1024` literal.
4. Add `io.EOF` loop termination.

**Tests**:
- `TestLoadTurnsForSession_Handlers_LargeImageLine_NoError` — handler-side `loadTurnsForSession` handles 5 MB image line
- `TestInspectFiles_LargeImageLine_NoError` — `InspectFiles` on file with 5 MB image line returns correct entry count without error
- Existing tests must pass

Estimated: ~50 lines modified, ~30 lines in tests

Run `make commit` after this stage.

---

### Stage 69.2 — Fix `countLines` in `cmd/mcp-server/handlers_stage1.go`

**Goal**: Replace the default-64KB-limit Scanner in `countLines` with a byte-counting approach using `bufio.Reader` + `ReadBytes('\n')` loop. This function does not need JSON parsing — only line counting. (`bytes.Count(fileBytes, []byte("\n"))` is not used because it requires reading the entire file into memory, which is unsuitable for large session files; `ReadBytes` reads in fixed-size chunks.)

**Files**:
- `cmd/mcp-server/handlers_stage1.go` — `countLines` rewritten using `bufio.Reader.ReadBytes`

**Procedure**:
1. Replace `bufio.NewScanner` with `bufio.NewReader`.
2. Use a loop: `for { _, err := r.ReadBytes('\n'); count++; if err == io.EOF { break } else if err != nil { return 0, err } }`.
3. Handle the case where the final line has no trailing newline (do not double-count).
4. The caller currently ignores the error return from `countLines`; this is a pre-existing issue documented in the proposal. Do NOT fix the caller in this stage — it is out of scope and would require understanding all call sites. Add a `//nolint:errcheck` comment or a TODO note for a follow-up.

**Tests**:
- `TestCountLines_LargeImageLine_NoError` — file with a 5 MB line; `countLines` returns correct count and no error (was previously failing with `ErrTooLong`)
- `TestCountLines_EmptyFile` — returns 0, no error
- `TestCountLines_NoTrailingNewline` — correct count when last line lacks `\n`
- Existing tests for `countLines` (if any) must pass

Estimated: ~20 lines modified, ~20 lines in tests

Run `make commit` after this stage.

---

## Phase 70: Cleanup and Lint Guard

**Goal**: Rename `MaxScannerLineBytes` to `LargeLineWarnBytes`; add monitoring log usage; verify zero `bufio.NewScanner` calls remain on JSONL paths; add a lint guard to prevent regression.

**Pre-condition**: Phases 66–69 all complete.

**Estimated LOC**: ~30 lines modified, ~20 lines added (lint config). Well within ≤500 lines.

**Dependencies**: Phases 66, 67, 68, 69.

---

### Stage 70.1 — Rename constant and add monitoring

**Goal**: Rename `MaxScannerLineBytes` (defined in `internal/types/constants.go` and aliased in `internal/parser/aliases.go`) to `LargeLineWarnBytes`. Add debug-level monitoring in `ReadLineFiltered` for lines exceeding this threshold.

**Two-phase constant lifecycle** (per proposal `MaxScannerLineBytes` disposal decision):
- **Phase 70 (this stage)**: Rename to `LargeLineWarnBytes`; semantics change from "hard limit" to "soft monitoring threshold". The alias in `internal/parser/aliases.go` is updated in parallel.
- **Post-stabilization (follow-up task, not in this plan)**: Once all acceptance criteria are verified green across several real sessions, remove `LargeLineWarnBytes` entirely (proposal Option 1). The monitoring log call in `streaming_reader.go` switches to a hardcoded literal or is removed. Track this as a TODO comment: `// TODO(post-stabilization): remove LargeLineWarnBytes once streaming reader is proven stable`.

**Files**:
- `internal/types/constants.go` — rename `MaxScannerLineBytes` → `LargeLineWarnBytes`; update comment to reflect new semantics ("soft warning threshold for large line monitoring, not a hard limit"); add TODO comment for post-stabilization removal
- `internal/parser/aliases.go` — update alias name to `LargeLineWarnBytes`
- `internal/parser/streaming_reader.go` — add post-read check: `if len(line) > LargeLineWarnBytes { slog.Debug("large line detected", "bytes", len(line)) }`
- All files that previously referenced `MaxScannerLineBytes` or `parser.MaxScannerLineBytes` — update to new name

**Verification**: Run `grep -r "MaxScannerLineBytes" .` — expect zero results.

**Tests**:
- `TestReadLineFiltered_LargeLineMonitoring` — confirm that a line exceeding `LargeLineWarnBytes` does not cause an error (monitoring is purely observational)
- Existing constant usage tests (if any) updated to new name

Estimated: ~30 lines modified

Run `make commit` after this stage.

---

### Stage 70.2 — Add lint guard and final acceptance verification

**Goal**: Add a custom lint rule (or `grep`-based CI check via `Makefile`) preventing JSONL-reading code from using raw `bufio.NewScanner`. Run full acceptance verification.

**Files**:
- `Makefile` — add `check-no-scanner` target that runs `grep -rn "bufio\.NewScanner" internal/ cmd/mcp-server/` and fails if any results contain JSONL file paths (excluding `main.go` stdin reader which is explicitly exempt)
- `.golangci.yml` (or existing lint config) — optionally add `forbidigo` rule for `bufio\.NewScanner` with exception for `main.go`

**Acceptance verification checklist** (mirrors proposal acceptance criteria):

1. `670a30a2-f413-4fdc-b2e4-ae05779aff05.jsonl` line 262 (6.8 MB): parse without error, other lines correct — verified by integration test or manual `go run`.
2. `make commit` exits 0 — all existing tests pass.
3. Memory profile: `go test -run TestParseEntries_LargeImageLine -memprofile=mem.out ./internal/parser/` — heap delta < 1 MB compared to baseline (document result in PR description).
4. `internal/types/constants.go` contains `LargeLineWarnBytes`, zero occurrences of `MaxScannerLineBytes` in codebase.
5. `streaming_reader.go` test coverage ≥ 80% (via `make test-coverage`).
6. `internal/query/files/file_inspector.go` and `cmd/mcp-server/handlers_stage1.go` confirmed migrated (regression tested in Phases 69.1 and 69.2).

**Tests**:
- `TestNoRawScannerOnJSONLPaths` — a Go test that `exec.Command("grep", ...)` the source tree and fails if `bufio.NewScanner` appears in JSONL-reading files other than `main.go` (optional; Makefile check is sufficient)

Estimated: ~20 lines added

Run `make push` (full check) after this stage.

---

## Testing Strategy

### TDD Protocol

Every stage follows strict TDD:
1. Write failing tests first
2. Run `make dev` to confirm failures
3. Implement until `make commit` passes
4. Never proceed to the next stage until `make commit` is green

### Coverage Requirements

- `internal/parser/streaming_reader.go`: ≥ 80% line coverage
- All modified files: existing coverage must not decrease

### Test Fixture Requirements

All tests use in-memory fixtures constructed programmatically (no new files in `tests/fixtures/` required for unit tests). Integration tests that need real JSONL files use `testing.Short()` guard:

```go
func TestStreamingReader_RealWorldFile(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test - requires real session file")
    }
    // ...
}
```

### Key Test Scenarios (cross-phase)

| Scenario | Phase covered | Test name pattern |
|----------|--------------|-------------------|
| Large image line no longer causes error | 66, 67, 68, 69 | `*_LargeImageLine_NoError` |
| Normal line after large line preserved | 66 | `TestReadLineFiltered_NormalLineAfterLargeLine_Preserved` |
| Multiple image blocks all truncated | 66 | `TestStripImageData_MultipleImages` |
| Text-only tool_result unchanged | 66 | `TestStripImageData_TextToolResult` |
| Strategy A skips image lines | 66 | `TestReadLineFiltered_StrategyA_ImageLine` |
| strategy2 query returns data (was error) | 68 | `TestReadJSONLFile_LargeImageLine_ReturnsData` |
| countLines handles 64KB+ lines | 69 | `TestCountLines_LargeImageLine_NoError` |
| Invalid JSON falls back gracefully | 66 | `TestStripImageData_InvalidAfterStrip_Fallback` |

---

## Risk Register

| Risk | Phase | Mitigation |
|------|-------|-----------|
| R1: byte replacement creates invalid JSON | 66 | `json.Valid()` check after each replacement; fallback to line-skip with warning log |
| R2: GC pressure from large allocations | 66 | Early truncation ensures large slice discarded before `Unmarshal`; debug monitoring via `LargeLineWarnBytes` |
| R3: existing tests assert on `ErrTooLong` behaviour | 67–69 | Full `make commit` after each stage; update expected values explicitly |
| R4: silent `scanner.Err()` bugs re-introduced | 68, 69 | New loop structure eliminates Scanner entirely; code review checklist item |
| R5: stage2 expected-error tests need updating | 68 | Documented in Stage 68.1 procedure; update expected values from error → data |
| R6: `countLines` caller ignores error | 69 | Out of scope for this plan; documented with TODO comment |
| R7: `main.go` stdin Scanner mistakenly migrated | 70 | Explicitly excluded from lint guard; `main.go` is listed as exempt |

---

## Dependency Graph Summary

```
Phase 66 (streaming_reader.go + parser/reader.go)
    │
    ├─► Phase 67 (query_executor.go — 2 Scanner sites)
    │       Stage 67.1: processFile + processFileWithTimeRange
    │
    ├─► Phase 68 (stage2 + filters — 3 Scanner sites)
    │       Stage 68.1: jq/stage2_executor.go + stage2_executor.go
    │       Stage 68.2: filters/filters.go
    │
    └─► Phase 69 (handlers + file_inspector + countLines — 3 Scanner sites)
            Stage 69.1: handlers_query.go + file_inspector.go
            Stage 69.2: handlers_stage1.go (countLines)
                │
                └─► Phase 70 (cleanup + lint guard)
                        Stage 70.1: constant rename + monitoring
                        Stage 70.2: lint guard + acceptance verification
```

**Total affected files**: 9 JSONL-reading files (per proposal complete Scanner inventory, excluding `main.go` stdin reader which is out of scope).

**Total estimated LOC**:
- Added: ~330 lines (new file + tests across all phases)
- Modified: ~260 lines
- Phase 66 alone stays within the 500-line limit; each subsequent phase is well under the limit.
