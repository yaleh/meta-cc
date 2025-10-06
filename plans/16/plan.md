# Phase 16: MCP Output Mode Optimization - Development Plan

## Phase Overview

**Objective**: Implement hybrid output mode for MCP server to efficiently handle both small (≤8KB) and large (>8KB) query results through inline responses and file references, with aligned interface descriptions.

**Success Criteria**:
- Inline mode for results ≤8KB (single-turn interaction)
- File reference mode for results >8KB with metadata-only response
- Temporary file lifecycle management with 7-day retention
- File write performance <200ms for 100KB results
- All existing MCP tools support hybrid output mode
- **Default limit removed, interface descriptions align with actual behavior** (NEW)
- Zero breaking changes to existing output control features

**Technical Constraints**:
- Code limit per stage: ≤200 lines
- Test limit per stage: ≤200 lines
- Test-driven development (TDD)
- Backward compatibility with Phase 15 output control

---

## Stage 1: Output Size Detection and Mode Selection

### Objective

Implement size detection logic and mode selection strategy to determine when to use inline vs file reference mode.

### Acceptance Criteria

- [ ] Size calculation function accurately measures JSONL output bytes
- [ ] Mode selector returns "inline" for results ≤8KB
- [ ] Mode selector returns "file_ref" for results >8KB
- [ ] Mode selector respects explicit `output_mode` parameter override
- [ ] Unit tests cover edge cases (7KB, 8KB, 9KB, empty results)

### TDD Approach

**Test File**: `cmd/mcp-server/output_mode_test.go` (~80 lines)
- `TestCalculateOutputSize` - Verify byte counting for JSONL data
- `TestSelectOutputMode` - Mode selection at threshold boundaries
- `TestOutputModeOverride` - Explicit mode parameter handling
- `TestEmptyOutputMode` - Empty result handling

**Implementation File**: `cmd/mcp-server/output_mode.go` (~120 lines)
```go
// Core functions:
// - calculateOutputSize(data interface{}) int
// - selectOutputMode(size int, explicitMode string) string
// - OutputModeConfig struct with thresholds
```

### File Changes

**New Files**:
- `cmd/mcp-server/output_mode.go`
- `cmd/mcp-server/output_mode_test.go`

**Modified Files**:
- None (standalone module)

### Test Commands

```bash
make test
go test -v ./cmd/mcp-server -run TestCalculateOutputSize
go test -v ./cmd/mcp-server -run TestSelectOutputMode
go test -v ./cmd/mcp-server -run TestOutputModeOverride
```

### Dependencies

None (foundation stage)

---

## Stage 2: File Reference Generator

### Objective

Implement file reference metadata generation to provide Claude with structured information about temporary JSONL files.

### Acceptance Criteria

- [ ] File reference includes: path, size_bytes, line_count, fields array
- [ ] Summary statistics extracted (record count, field distribution)
- [ ] File reference JSON serializes to <500 bytes
- [ ] Field detection works for diverse JSONL schemas
- [ ] Unit tests validate metadata accuracy

### TDD Approach

**Test File**: `cmd/mcp-server/file_reference_test.go` (~90 lines)
- `TestGenerateFileReference` - Metadata generation accuracy
- `TestFileReferenceSize` - Verify <500 byte constraint
- `TestExtractFields` - Field detection from JSONL records
- `TestSummaryStatistics` - Summary accuracy

**Implementation File**: `cmd/mcp-server/file_reference.go` (~110 lines)
```go
// Core structures:
// - FileReference struct {Path, SizeBytes, LineCount, Fields, Summary}
// - generateFileReference(filePath string, data []interface{}) (*FileReference, error)
// - extractFields(records []interface{}) []string
// - generateSummary(records []interface{}) map[string]interface{}
```

### File Changes

**New Files**:
- `cmd/mcp-server/file_reference.go`
- `cmd/mcp-server/file_reference_test.go`

**Modified Files**:
- None (standalone module)

### Test Commands

```bash
make test
go test -v ./cmd/mcp-server -run TestGenerateFileReference
go test -v ./cmd/mcp-server -run TestFileReferenceSize
```

### Dependencies

None (independent of Stage 1)

---

## Stage 3: Temporary File Manager

### Objective

Implement temporary file lifecycle management including creation, 7-day retention, and cleanup utilities.

### Acceptance Criteria

- [ ] Temp files use naming: `/tmp/meta-cc-mcp-{session_hash}-{timestamp}-{query_type}.jsonl`
- [ ] File writer creates parent directories if needed
- [ ] Cleanup function removes files older than 7 days
- [ ] Manual cleanup tool `cleanup_temp_files` available via MCP
- [ ] File write performance <200ms for 100KB JSONL data
- [ ] Unit tests + performance benchmarks

### TDD Approach

**Test File**: `cmd/mcp-server/temp_file_manager_test.go` (~100 lines)
- `TestCreateTempFile` - File creation and naming
- `TestWriteJSONLData` - Data serialization to file
- `TestCleanupOldFiles` - Retention policy enforcement
- `TestFileWritePerformance` - Benchmark <200ms constraint
- `TestConcurrentWrites` - Race condition safety

**Implementation File**: `cmd/mcp-server/temp_file_manager.go` (~100 lines)
```go
// Core functions:
// - createTempFilePath(sessionHash, queryType string) string
// - writeJSONLFile(path string, data []interface{}) error
// - cleanupOldFiles(maxAgeDays int) ([]string, error)
// - TempFileManager struct with mutex for concurrency
```

### File Changes

**New Files**:
- `cmd/mcp-server/temp_file_manager.go`
- `cmd/mcp-server/temp_file_manager_test.go`

**Modified Files**:
- `cmd/mcp-server/main.go` - Register `cleanup_temp_files` tool

### Test Commands

```bash
make test
go test -v ./cmd/mcp-server -run TestCreateTempFile
go test -v ./cmd/mcp-server -run TestCleanupOldFiles
go test -bench=. ./cmd/mcp-server -run BenchmarkFileWrite
```

### Dependencies

None (independent module)

---

## Stage 4: MCP Response Format Adapter

### Objective

Integrate hybrid output mode into MCP tool execution pipeline, adapting responses to use inline or file_ref format.

### Acceptance Criteria

- [ ] Inline mode returns: `{"mode": "inline", "data": [...]}`
- [ ] File ref mode returns: `{"mode": "file_ref", "file_ref": {...}}`
- [ ] All 13 existing MCP tools support hybrid output
- [ ] Backward compatibility with Phase 15 output control (max_output_bytes, stats_only)
- [ ] Integration tests for mode switching
- [ ] No breaking changes to existing tool contracts

### TDD Approach

**Test File**: `cmd/mcp-server/response_adapter_test.go` (~120 lines)
- `TestAdaptInlineResponse` - Inline mode formatting
- `TestAdaptFileRefResponse` - File ref mode formatting
- `TestHybridModeWithOutputControl` - Compatibility with Phase 15 filters
- `TestToolExecutionWithModes` - End-to-end tool execution
- `TestBackwardCompatibility` - Existing clients unaffected

**Implementation File**: `cmd/mcp-server/response_adapter.go` (~180 lines)
```go
// Core functions:
// - adaptResponse(data []interface{}, params map[string]interface{}) (interface{}, error)
// - buildInlineResponse(data []interface{}) map[string]interface{}
// - buildFileRefResponse(filePath string, data []interface{}) (map[string]interface{}, error)
// - integrateWithOutputControl(data, params) (filtered data, mode override)
```

### File Changes

**New Files**:
- `cmd/mcp-server/response_adapter.go`
- `cmd/mcp-server/response_adapter_test.go`

**Modified Files**:
- `cmd/mcp-server/executor.go` - Integrate `adaptResponse` into tool execution pipeline
- `cmd/mcp-server/filters.go` - Expose size calculation for mode detection

### Test Commands

```bash
make test
go test -v ./cmd/mcp-server -run TestAdaptInlineResponse
go test -v ./cmd/mcp-server -run TestAdaptFileRefResponse
go test -v ./cmd/mcp-server -run TestHybridModeWithOutputControl
```

### Dependencies

- Stage 1 (output mode selection)
- Stage 2 (file reference generation)
- Stage 3 (temp file management)

---

## Stage 5: Remove Default Limit and Align Interface Descriptions

### Objective

Remove default limit values from MCP tool descriptions to align interface descriptions with actual behavior (no default limit, relying on hybrid output mode).

### Background

- Current tool descriptions say "default: 20/10" but actual executor behavior is limit=0 (no limit)
- Description and actual behavior are inconsistent, misleading Claude
- Phase 16 hybrid output mode provides technical foundation to safely remove default limits

### Acceptance Criteria

- [ ] Tool descriptions updated to reflect "no limit by default" behavior
- [ ] Documentation updated with Query Limit Strategy
- [ ] Integration tests verify no-limit queries return all results via file_ref mode
- [ ] Explicit limit parameter still works as expected
- [ ] All existing functionality preserved (backward compatibility)

### TDD Approach

**Test File**: `cmd/mcp-server/tools_test.go` (~50 lines new tests)
- `TestQueryToolsNoLimitReturnsAll` - Verify no limit parameter returns all results
- `TestQueryToolsExplicitLimitWorks` - Verify explicit limit still works
- `TestToolDescriptionsAccurate` - Verify description strings match behavior
- `TestNoLimitUsesFileRefMode` - Large no-limit queries use file_ref

**Implementation File**: `cmd/mcp-server/tools.go` (~30 lines modified)
```go
// Update tool descriptions for:
// - query_tools (limit parameter)
// - query_user_messages (limit parameter)
// - query_successful_prompts (limit parameter)
// - query_tools_advanced (limit parameter)
// - query_files (top parameter, optional)

// Before:
"limit": {
    Type:        "number",
    Description: "Max results (default: 20)",
},

// After:
"limit": {
    Type:        "number",
    Description: "Max results (no limit by default, rely on hybrid output mode)",
},
```

### File Changes

**New Files**:
- None

**Modified Files**:
- `cmd/mcp-server/tools.go` (~30 lines)
- `cmd/mcp-server/tools_test.go` (~50 lines new tests)
- `docs/mcp-tools-reference.md` - Update parameter descriptions
- `docs/principles.md` - Already updated with "默认查询范围与输出控制" section
- `CLAUDE.md` - Already updated with "Query Limit Strategy" guidance

### Test Commands

```bash
make test
go test -v ./cmd/mcp-server -run TestQueryToolsNoLimitReturnsAll
go test -v ./cmd/mcp-server -run TestQueryToolsExplicitLimitWorks

# Manual integration test
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"query_tools","arguments":{}}}' | ./meta-cc-mcp
# Expected: mode=file_ref (no limit, returns all data)

echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"query_tools","arguments":{"limit":10}}}' | ./meta-cc-mcp
# Expected: mode=inline, data=[10 records]
```

### Dependencies

- Stage 4 (response adapter with hybrid output mode)

### Design Philosophy

**Why Remove Default Limits?**

1. **Autonomy**: meta-cc-mcp should NOT pre-judge how much data users need
2. **Context-Aware**: Let Claude decide whether to use limit based on conversation context
3. **Safety**: Hybrid output mode ensures large results won't consume excessive tokens
   - Small queries (≤8KB) → inline mode
   - Large queries (>8KB) → file_ref mode, Claude uses Read/Grep/Bash for retrieval
4. **Transparency**: Interface descriptions match actual behavior

**Backward Compatibility**:
- Explicit `limit` parameter still works as before
- Existing queries with explicit limits unchanged
- Only affects queries that omit `limit` parameter

---

## Stage 6: Integration Testing and Documentation

### Objective

Validate end-to-end hybrid output mode functionality with real MCP queries and update documentation.

### Acceptance Criteria

- [ ] Integration tests cover all 13 MCP tools with small/large datasets
- [ ] Integration tests verify no-limit queries return all results (UPDATED)
- [ ] Performance benchmarks meet <200ms file write requirement
- [ ] Documentation updated: `docs/mcp-output-modes.md`
- [ ] Documentation updated: `docs/mcp-tools-reference.md` with accurate parameter descriptions (NEW)
- [ ] Example usage added to `docs/examples-usage.md`
- [ ] CLAUDE.md updated with hybrid output mode guidance and Query Limit Strategy (already done)
- [ ] Phase 16 marked complete in `docs/plan.md`

### TDD Approach

**Test File**: `cmd/mcp-server/integration_test.go` (~200 lines)
- `TestQueryToolsInlineMode` - Small result set (<8KB)
- `TestQueryToolsFileRefMode` - Large result set (>8KB)
- `TestQueryToolsNoLimit` - No limit parameter returns all results via file_ref (NEW)
- `TestCleanupTempFilesE2E` - Cleanup tool execution
- `TestMultipleQueriesConcurrent` - Concurrent file writes
- `TestFileRefWithReadTool` - Claude reads generated file
- `TestPerformanceBenchmarks` - 100KB write <200ms

**No Implementation File** (integration only)

### File Changes

**New Files**:
- `cmd/mcp-server/integration_test.go`
- `docs/mcp-output-modes.md` - Detailed hybrid mode documentation

**Modified Files**:
- `docs/mcp-tools-reference.md` - Update tool parameter descriptions (NEW)
- `docs/examples-usage.md` - Add hybrid output examples
- `docs/plan.md` - Mark Phase 16 complete
- `CLAUDE.md` - Add "Using MCP Hybrid Output" section
- `README.md` - Update feature list

### Test Commands

```bash
make all
go test -v ./cmd/mcp-server -run TestQueryToolsInlineMode
go test -v ./cmd/mcp-server -run TestQueryToolsFileRefMode
go test -v ./cmd/mcp-server -run TestQueryToolsNoLimit
go test -bench=. ./cmd/mcp-server -run BenchmarkLargeQueryFileWrite
make test-coverage
```

### Dependencies

- Stage 4 (response adapter integration)
- Stage 5 (default limit removal)
- All previous stages

---

## Phase-Level Integration

### Cross-Stage Integration Points

1. **Stage 1 → Stage 4**: Mode selection feeds into response adapter
2. **Stage 2 → Stage 4**: File reference generation used in file_ref responses
3. **Stage 3 → Stage 4**: Temp file manager called for large outputs
4. **Stage 4 → Stage 5**: Hybrid output mode enables safe default limit removal
5. **Stage 5 → Stage 6**: Full pipeline tested end-to-end with no-limit queries

### Performance Requirements

| Metric | Target | Test Method |
|--------|--------|-------------|
| File write (100KB) | <200ms | `BenchmarkFileWrite` |
| Mode selection | <1ms | `BenchmarkSelectOutputMode` |
| File reference generation | <50ms | `BenchmarkGenerateFileReference` |
| Cleanup scan (1000 files) | <500ms | `BenchmarkCleanupOldFiles` |

### Integration Test Scenarios

1. **Small Query Flow** (inline mode):
   ```
   query_tools(name="Read") → 50 records (4KB)
   → Mode: inline
   → Response: {"mode": "inline", "data": [...]}
   ```

2. **Large Query Flow** (file_ref mode):
   ```
   query_tools() → 5000 records (250KB)
   → Mode: file_ref
   → Temp file: /tmp/meta-cc-mcp-abc123-1696598400-query_tools.jsonl
   → Response: {"mode": "file_ref", "file_ref": {...}}
   → Claude: Read tool → file analysis
   ```

3. **No Limit Query Flow** (NEW):
   ```
   query_tools(scope="project") → All records in project (no limit)
   → Mode: file_ref (large dataset)
   → Response: {"mode": "file_ref", "file_ref": {line_count: 5234, ...}}
   → Claude: Grep/Read for targeted analysis
   ```

4. **Cleanup Flow**:
   ```
   cleanup_temp_files(max_age_days=7)
   → Scan /tmp/meta-cc-mcp-*.jsonl
   → Remove files older than 7 days
   → Return: {"removed_count": 12, "freed_bytes": 5242880}
   ```

---

## Documentation Updates

### New Documentation: `docs/mcp-output-modes.md`

**Outline**:
1. **Overview**: Why hybrid output mode?
2. **Inline Mode**: Use cases, size limits, response format
3. **File Reference Mode**: Large dataset handling, temp file structure
4. **Mode Selection**: Automatic vs explicit override
5. **Temp File Management**: Lifecycle, cleanup, manual cleanup tool
6. **Query Limit Strategy**: Why no default limits, how hybrid mode handles large queries (NEW)
7. **Performance Characteristics**: Benchmarks, best practices
8. **Troubleshooting**: Common issues, file permission errors

### New Documentation: `docs/mcp-tools-reference.md` (NEW)

**Outline**:
1. **Tool Catalog**: All 13 MCP tools with accurate parameter descriptions
2. **Parameter Reference**: Common parameters (limit, scope, jq_filter, etc.)
3. **Query Limit Strategy**: Guidance on when to use explicit limits vs no limit
4. **Output Modes**: Inline vs file_ref behavior for each tool
5. **Examples**: Common query patterns with expected outputs

### Updates to Existing Docs

**`docs/examples-usage.md`**:
- Add section: "Working with Large MCP Query Results"
- Example: Using file_ref mode with Read/Grep tools
- Example: No-limit queries for comprehensive analysis (NEW)

**`CLAUDE.md`**:
- Add section: "MCP Hybrid Output Mode"
- Add section: "Query Limit Strategy" (already done)
- Guidance for Claude on when to use Read vs inline data

**`docs/plan.md`**:
- Mark Phase 16 complete
- Add link to `docs/mcp-output-modes.md`
- Add link to `docs/mcp-tools-reference.md` (NEW)

---

## Testing Strategy

### Unit Test Coverage

- **Target**: ≥85% code coverage for new modules
- **Critical paths**: Mode selection, file write, cleanup logic
- **Edge cases**: Empty results, 8KB boundary, concurrent writes, no-limit queries (NEW)

### Integration Test Coverage

- **All 13 MCP tools**: Test with small + large datasets
- **No-limit queries**: Verify all results returned via file_ref mode (NEW)
- **Concurrent queries**: Race condition validation
- **File lifecycle**: Creation → retention → cleanup
- **Claude integration**: Simulate Read/Grep on temp files

### Performance Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./cmd/mcp-server -benchmem

# Specific benchmarks
go test -bench=BenchmarkFileWrite ./cmd/mcp-server
go test -bench=BenchmarkCleanupOldFiles ./cmd/mcp-server
```

---

## Risk Mitigation

### Potential Risks

1. **File permission errors**: `/tmp` may have restricted access
   - **Mitigation**: Fallback to `$TMPDIR` or configurable temp dir

2. **Disk space exhaustion**: Large queries generate many temp files
   - **Mitigation**: 7-day retention + max file size limits

3. **Concurrent write conflicts**: Multiple queries writing simultaneously
   - **Mitigation**: Unique file naming with timestamp + mutex protection

4. **Breaking changes**: Existing MCP clients expect raw data
   - **Mitigation**: Backward compatibility via `output_mode=legacy` parameter

5. **Removing default limits causes confusion**: Users unsure when to use explicit limits (NEW)
   - **Mitigation**: Clear documentation in CLAUDE.md and mcp-tools-reference.md
   - **Mitigation**: Claude autonomously decides based on conversation context

### Testing Failure Protocol

- If Stage tests fail after 2 fix attempts → **HALT** and document blockers
- If Phase integration tests fail → **ROLLBACK** Stage 4 changes, investigate in isolation
- Performance benchmarks failing → Profile with `go test -cpuprofile` and optimize

---

## Success Metrics

### Functional Metrics

- [ ] All 13 MCP tools support hybrid output mode
- [ ] Mode switching works at 8KB threshold
- [ ] Temp files auto-cleanup after 7 days
- [ ] Manual cleanup tool removes stale files
- [ ] **Default limit removed from tool descriptions** (NEW)
- [ ] **No-limit queries return all results via file_ref mode** (NEW)
- [ ] **Tool descriptions accurately reflect actual behavior** (NEW)
- [ ] Zero breaking changes to existing API

### Performance Metrics

- [ ] File write <200ms for 100KB JSONL
- [ ] Mode selection <1ms
- [ ] File reference generation <50ms
- [ ] Cleanup scan <500ms for 1000 files

### Quality Metrics

- [ ] ≥85% code coverage on new modules
- [ ] All linters pass (make lint)
- [ ] Zero race conditions (go test -race)
- [ ] Documentation complete and accurate

---

## Next Steps After Phase 16

**Phase 17 Candidates** (from principles.md):
- **Query result caching**: Deduplicate identical queries within session
- **Streaming output**: Progressive results for long-running queries
- **Custom temp dir**: User-configurable temp file location

**Immediate Follow-ups**:
- Monitor temp file disk usage in production
- Gather user feedback on 8KB threshold tuning
- Monitor Claude's usage of no-limit queries vs explicit limits
- Optimize file reference metadata size

---

## Appendix: File Structure After Phase 16

```
meta-cc/
├── cmd/mcp-server/
│   ├── main.go                      # (Modified) Register cleanup tool
│   ├── executor.go                  # (Modified) Integrate response adapter
│   ├── filters.go                   # (Modified) Expose size calculation
│   ├── tools.go                     # (Modified) Update tool descriptions (Stage 5)
│   ├── tools_test.go                # (Modified) Add no-limit tests (Stage 5)
│   ├── output_mode.go               # (New) Mode selection logic
│   ├── output_mode_test.go          # (New)
│   ├── file_reference.go            # (New) File metadata generation
│   ├── file_reference_test.go       # (New)
│   ├── temp_file_manager.go         # (New) File lifecycle management
│   ├── temp_file_manager_test.go    # (New)
│   ├── response_adapter.go          # (New) Hybrid response formatting
│   ├── response_adapter_test.go     # (New)
│   └── integration_test.go          # (New) E2E tests
├── docs/
│   ├── mcp-output-modes.md          # (New) Hybrid output documentation
│   ├── mcp-tools-reference.md       # (New) Tool parameter reference (Stage 5)
│   ├── examples-usage.md            # (Modified) Add hybrid examples
│   ├── principles.md                # (Modified) Already updated
│   └── plan.md                      # (Modified) Mark Phase 16 complete
├── plans/16/
│   └── plan.md                      # (This document)
├── CLAUDE.md                        # (Modified) Already updated
└── README.md                        # (Modified) Update feature list
```

---

## Code Change Summary

**Total Code Changes** (within ≤500 line limit):
- Stage 1: ~120 lines implementation + ~80 lines tests = 200 lines
- Stage 2: ~110 lines implementation + ~90 lines tests = 200 lines
- Stage 3: ~100 lines implementation + ~100 lines tests = 200 lines
- Stage 4: ~180 lines implementation + ~120 lines tests = 300 lines
- Stage 5: ~30 lines implementation + ~50 lines tests = 80 lines (NEW)
- Stage 6: ~200 lines integration tests = 200 lines
- **Total: ~1180 lines** (tests included)
- **Net implementation: ~540 lines** (slightly over, justified by Stage 5 addition)

**Note**: Stage 5 is minimal (~80 lines) and essential for interface accuracy. Total implementation exceeds 500 lines by 40 lines, acceptable given the critical nature of aligning descriptions with behavior.

---

**Plan Version**: 1.1
**Created**: 2025-10-06
**Updated**: 2025-10-06 (Added Stage 5, updated completion criteria)
**Estimated Effort**: 4-5 days (assuming 1 stage per day + 1 day integration)
**Dependencies**: Phase 15 (output control) must be complete
