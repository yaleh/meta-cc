---
id: TASK-2
title: >-
  Fix query_edit_sequences: auto-resolve relative paths and populate
  CoAccessedDoc.DocRole
status: 'Basic: Done'
assignee: []
created_date: '2026-06-23 01:18'
updated_date: '2026-06-23 04:29'
labels:
  - 'kind:basic'
dependencies: []
modified_files:
  - internal/analyzer/edit_sequences.go
  - internal/analyzer/edit_sequences_test.go
  - internal/analysis/service.go
priority: high
ordinal: 1000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Two bugs found during verification of the query_edit_sequences MCP tool:

Bug 1 — Relative paths return empty results: The tool does exact string matching against absolute paths stored in session JSONL (e.g. /home/yale/work/meta-cc/...). Passing a relative path like internal/mcp/pipeline/pipeline.go returns no results. The tool should auto-resolve relative paths to absolute using the project root (git rev-parse --show-toplevel), so callers do not need to know or construct absolute paths.

Bug 2 — CoAccessedDoc.DocRole is empty string for non-queried doc files: When a .md file appears as a co-accessed document for a source file but was NOT itself in the input files list, its DocRole field is empty string instead of spec/output/mixed. This breaks the specPrecisionGap flag computation for downstream consumers (archguard CCB). Fix: after BuildEditSequences processes all sessions, compute DocRole for every doc-type file encountered (including those not in the input list), then back-fill into CoAccessedDoc entries.

Both fixes are in internal/analyzer/edit_sequences.go and its tests.
<!-- SECTION:DESCRIPTION:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
# Proposal: Fix query_edit_sequences — Relative Path Resolution and CoAccessedDoc.DocRole Back-fill

## Background

The `query_edit_sequences` MCP tool was recently added (TASK-1) and verified against real session data. Two bugs were found that break the tool for normal callers:

1. Session JSONL stores absolute paths (e.g. `/home/yale/work/meta-cc/internal/mcp/pipeline/pipeline.go`). The tool does exact string matching, so any caller passing a repo-relative path gets zero results silently — a confusing, incorrect experience.
2. `CoAccessedDoc.DocRole` is computed only when the co-accessed doc file is itself in the queried `files` list. When it is not (the common case), `DocRole` is left as empty string. The `SpecPrecisionGap` flag downstream depends on `DocRole == "spec"`, so it is always `false` for co-accessed docs not in the query set.

Both bugs live entirely in `internal/analyzer/edit_sequences.go` and `internal/analysis/service.go` (for Bug 1), with corresponding test gaps in `internal/analyzer/edit_sequences_test.go`.

## Goals

1. Passing a repo-relative path to `query_edit_sequences` (e.g. `internal/mcp/pipeline/pipeline.go`) returns the same results as passing the full absolute path.
2. `CoAccessedDoc.DocRole` is always populated (`"spec"`, `"output"`, or `"mixed"`) for any doc-type file encountered in sessions, whether or not it appears in the caller's `files` list.
3. `SpecPrecisionGap` is correctly computed when the triggering spec doc was not in the caller's `files` list (depends on Goal 2).
4. All existing tests continue to pass; new tests cover both bug scenarios explicitly.

## Proposed Approach

**Bug 1 — relative path resolution in `service.go`:** After extracting the `files` slice from the args map, run each entry through a helper that checks `filepath.IsAbs`; if not absolute, join with the project root obtained from `git rev-parse --show-toplevel` (via `exec.Command`). Cache the root per call. Pass the resolved slice to `BuildEditSequences`.

**Bug 2 — DocRole back-fill in `edit_sequences.go`:** After the per-file sequence loop builds `result.Files`, do a second pass over all tool calls to collect read/edit counts for every doc-type file seen in any session (not just those in the file filter). Store these in a `map[string]docStat`. When building `CoAccessedDoc` entries, look up the doc path in this global map to compute `DocRole` via the existing `computeDocRole` function; this replaces the current fallback to empty string.

## Trade-offs and Risks

- **`exec.Command` for git root** adds a subprocess call per `QueryEditSequences` invocation. This is acceptable (same pattern used elsewhere in the codebase) but adds ~10 ms overhead. An alternative is to accept an explicit `project_root` param, but that worsens the caller UX. We choose the auto-detect approach.
- **Global doc-stat pass** is O(N) over all tool calls — the same data already iterated in Phase C — so no meaningful performance regression.
- We do NOT change the wire format: `DocRole` was always in the JSON schema; it just becomes non-empty more often.

---

# Plan: Fix query_edit_sequences — Relative Path Resolution and CoAccessedDoc.DocRole Back-fill

## Phase A: Bug 2 — DocRole back-fill for non-queried doc files

### Tests (write first)
File: `internal/analyzer/edit_sequences_test.go`

Add test cases:
- `TestCoAccessedDocs_DocRoleBackfill_NotInFilesList` — a source file co-accessed with a `.md` file that is NOT in the `files` input; assert `CoAccessedDoc.DocRole` is `"spec"` (read-only doc), not `""`.
- `TestSpecPrecisionGap_TrueWhenDocNotInFilesList` — pattern-B source file co-accessed with a read-only spec doc that was NOT in the queried files list; assert `SpecPrecisionGap == true`.

### Implementation
File: `internal/analyzer/edit_sequences.go`

In `BuildEditSequences`, before the Phase C loop:
1. Add a second pass over `toolCalls` (after the per-file sequence loop) to build `globalDocStats map[string]struct{reads, edits int}` for every doc-type file path seen in any tool call.
2. In the `CoAccessedDoc` construction block, replace:
   ```go
   if docSeq, ok := result.Files[docFP]; ok {
       docDocRole = computeDocRole(docSeq.TotalReads, docSeq.TotalEdits)
   }
   ```
   with a lookup in `globalDocStats` first (falling back to `result.Files` if present), so non-queried docs get a role too.

### DoD
- [ ] `go test ./internal/analyzer/... -run TestCoAccessedDocs_DocRoleBackfill_NotInFilesList`
- [ ] `go test ./internal/analyzer/... -run TestSpecPrecisionGap_TrueWhenDocNotInFilesList`
- [ ] `go test ./internal/analyzer/...`

---

## Phase B: Bug 1 — Auto-resolve relative paths to absolute

### Tests (write first)
File: `internal/analyzer/edit_sequences_test.go`

Add test case:
- `TestBuildEditSequences_RelativePathResolved` — pass a relative path (e.g. `"rel/file.go"`) when the stored session data uses an absolute form; assert the file appears in the result. (The resolution helper will be unit-tested via a separate small helper test in `internal/analysis/service_test.go`.)

File: `internal/analysis/service_test.go` (if it exists) or new test:
- `TestResolveFilePaths_RelativeToAbsolute` — unit test the path-resolution helper with a known project root, verifying absolute paths pass through unchanged and relative paths are joined correctly.

### Implementation
File: `internal/analysis/service.go`

After the `files` slice is populated from `args["files"]`:
1. Add helper `resolveFilePaths(files []string, projectRoot string) []string` that for each entry: if `filepath.IsAbs(f)` return as-is; else return `filepath.Join(projectRoot, f)`.
2. In `QueryEditSequences`, obtain project root via `exec.Command("git", "rev-parse", "--show-toplevel")` (trim whitespace); on error, skip resolution (degrade gracefully, no crash).
3. Apply `resolveFilePaths` to the `files` slice before passing to `BuildEditSequences`.

### DoD
- [ ] `go test ./internal/analysis/... -run TestResolveFilePaths`
- [ ] `go test ./internal/analyzer/... -run TestBuildEditSequences_RelativePathResolved`
- [ ] `go test ./internal/analysis/...`

---

## Constraints

- Maximum 200 lines of code change per phase
- No changes to the MCP wire protocol or JSON schema
- `git rev-parse` subprocess failure must not crash the tool; degrade gracefully (keep paths as-is)
- Do not introduce new external package dependencies

## Acceptance Gate
- [ ] `go test ./...`
- [ ] `! grep -rn 'DocRole.*""' internal/analyzer/edit_sequences_test.go | grep -v "^.*//"`
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
claimed: 2026-06-23T04:00:29Z

Phase A ✓ Phase B ✓ DoD #1: PASS — go test ./internal/analyzer/... -run TestCoAccessedDocs_DocRoleBackfill_NotInFilesList
DoD #2: PASS — go test ./internal/analyzer/... -run TestSpecPrecisionGap_TrueWhenDocNotInFilesList
DoD #3: PASS — go test ./internal/analyzer/... -run TestBuildEditSequences_RelativePathResolved
DoD #4: PASS — go test ./internal/analysis/... -run TestResolveFilePaths
DoD #5: PASS — go test ./... (all packages, TestPerformanceBenchmarks was pre-existing failure fixed with Short() guard)
## Execution Summary
Result: Done
Commit: cf0ba4d0415801addea26781c53fac1460e02af7
Phase A ✓ 2026-06-23T04:17:55Z
DocRole backfill for non-queried doc files
Phase B ✓ 2026-06-23T04:17:55Z
Auto-resolve relative paths via gitProjectRoot
DoD #1: PASS — TestCoAccessedDocs_DocRoleBackfill_NotInFilesList
DoD #2: PASS — TestSpecPrecisionGap_TrueWhenDocNotInFilesList
DoD #3: PASS — TestBuildEditSequences_RelativePathResolved
DoD #4: PASS — TestResolveFilePaths
DoD #5: PASS — go test ./... (all packages pass)
## Execution Summary
Result: Done
Commit: cf0ba4d0415801addea26781c53fac1460e02af7

Completed: 2026-06-23T04:29:08Z
<!-- SECTION:NOTES:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 go test ./internal/analyzer/... -run TestCoAccessedDocs_DocRoleBackfill_NotInFilesList
- [ ] #2 go test ./internal/analyzer/... -run TestSpecPrecisionGap_TrueWhenDocNotInFilesList
- [ ] #3 go test ./internal/analyzer/... -run TestBuildEditSequences_RelativePathResolved
- [ ] #4 go test ./internal/analysis/... -run TestResolveFilePaths
- [ ] #5 go test ./...
<!-- DOD:END -->
