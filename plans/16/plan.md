# Phase 16: MCP Output Mode Optimization - Development Plan

## Phase Overview

**Objective**: Implement hybrid output mode for MCP server to efficiently handle both small (≤8KB) and large (>8KB) query results through inline responses and file references.

**Success Criteria**:
- Inline mode for results ≤8KB (single-turn interaction)
- File reference mode for results >8KB with metadata-only response
- Temporary file lifecycle management with 7-day retention
- File write performance <200ms for 100KB results
- All existing MCP tools support hybrid output mode
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

## Stage 5: Integration Testing and Documentation

### Objective

Validate end-to-end hybrid output mode functionality with real MCP queries and update documentation.

### Acceptance Criteria

- [ ] Integration tests cover all 13 MCP tools with small/large datasets
- [ ] Performance benchmarks meet <200ms file write requirement
- [ ] Documentation updated: `docs/mcp-output-modes.md`
- [ ] Example usage added to `docs/examples-usage.md`
- [ ] CLAUDE.md updated with hybrid output mode guidance
- [ ] Phase 16 marked complete in `docs/plan.md`

### TDD Approach

**Test File**: `cmd/mcp-server/integration_test.go` (~200 lines)
- `TestQueryToolsInlineMode` - Small result set (<8KB)
- `TestQueryToolsFileRefMode` - Large result set (>8KB)
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
- `docs/examples-usage.md` - Add hybrid output examples
- `docs/plan.md` - Mark Phase 16 complete
- `CLAUDE.md` - Add "Using MCP Hybrid Output" section
- `README.md` - Update feature list

### Test Commands

```bash
make all
go test -v ./cmd/mcp-server -run TestQueryToolsInlineMode
go test -v ./cmd/mcp-server -run TestQueryToolsFileRefMode
go test -bench=. ./cmd/mcp-server -run BenchmarkLargeQueryFileWrite
make test-coverage
```

### Dependencies

- Stage 4 (response adapter integration)
- All previous stages

---

## Phase-Level Integration

### Cross-Stage Integration Points

1. **Stage 1 → Stage 4**: Mode selection feeds into response adapter
2. **Stage 2 → Stage 4**: File reference generation used in file_ref responses
3. **Stage 3 → Stage 4**: Temp file manager called for large outputs
4. **Stage 4 → Stage 5**: Full pipeline tested end-to-end

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

3. **Cleanup Flow**:
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
6. **Performance Characteristics**: Benchmarks, best practices
7. **Troubleshooting**: Common issues, file permission errors

### Updates to Existing Docs

**`docs/examples-usage.md`**:
- Add section: "Working with Large MCP Query Results"
- Example: Using file_ref mode with Read/Grep tools

**`CLAUDE.md`**:
- Add section: "MCP Hybrid Output Mode"
- Guidance for Claude on when to use Read vs inline data

**`docs/plan.md`**:
- Mark Phase 16 complete
- Add link to `docs/mcp-output-modes.md`

---

## Testing Strategy

### Unit Test Coverage

- **Target**: ≥85% code coverage for new modules
- **Critical paths**: Mode selection, file write, cleanup logic
- **Edge cases**: Empty results, 8KB boundary, concurrent writes

### Integration Test Coverage

- **All 13 MCP tools**: Test with small + large datasets
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
- Optimize file reference metadata size

---

## Appendix: File Structure After Phase 16

```
meta-cc/
├── cmd/mcp-server/
│   ├── main.go                      # (Modified) Register cleanup tool
│   ├── executor.go                  # (Modified) Integrate response adapter
│   ├── filters.go                   # (Modified) Expose size calculation
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
│   ├── examples-usage.md            # (Modified) Add hybrid examples
│   └── plan.md                      # (Modified) Mark Phase 16 complete
├── plans/16/
│   └── plan.md                      # (This document)
└── CLAUDE.md                        # (Modified) Add hybrid output guidance
```

---

**Plan Version**: 1.0
**Created**: 2025-10-06
**Estimated Effort**: 3-4 days (assuming 1 stage per day + 1 day integration)
**Dependencies**: Phase 15 (output control) must be complete
