# Phase 16: MCP Output Mode Optimization - Development Plan

## Phase Overview

**Objective**: Implement hybrid output mode for MCP server to efficiently handle both small (≤8KB) and large (>8KB) query results through inline responses and file references, with aligned interface descriptions and complete reliance on hybrid mode (no truncation).

**Background & Problems**:
- **Problem 1**: Truncation mechanism breaks hybrid mode (data truncated before mode decision, causing file_ref mode to fail)
- **Problem 2**: Hardcoded threshold cannot adapt to different scenarios (8KB fixed value, not configurable)
- **Problem 3**: Double truncation causes information loss (integrateWithOutputControl + executor final truncation)

**Success Criteria**:
- Inline mode for results ≤8KB (single-turn interaction)
- File reference mode for results >8KB with metadata-only response
- Temporary file lifecycle management with 7-day retention
- File write performance <200ms for 100KB results
- All existing MCP tools support hybrid output mode
- Default limit removed, interface descriptions align with actual behavior
- **All truncation logic removed, completely rely on hybrid mode (Stage 16.6)**
- **Threshold configurable via parameter or environment variable**
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
- [ ] Backward compatibility with Phase 15 output control (stats_only, stats_first)
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
- `docs/mcp-guide.md` - Update parameter descriptions
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

## Stage 6: Remove Truncation Mechanism, Fully Rely on Hybrid Mode

### Objective

Remove all truncation logic from the MCP server and make the system completely rely on hybrid output mode for handling large results, with configurable threshold support.

### Background

The current system has truncation logic that breaks hybrid mode:
1. **integrateWithOutputControl truncates before mode decision** - Data is truncated before hybrid mode can detect size and switch to file_ref mode
2. **executor applies final truncation** - Additional truncation applied after mode selection
3. **Hardcoded threshold** - 8KB threshold cannot be configured for different scenarios

### Acceptance Criteria

- [ ] All truncation logic removed from `response_adapter.go`
- [ ] Final output truncation removed from `executor.go`
- [ ] `max_output_bytes` parameter deleted from tool definitions
- [ ] New `inline_threshold_bytes` parameter added (default: 8192)
- [ ] Environment variable `META_CC_INLINE_THRESHOLD` supported for global configuration
- [ ] Threshold configuration works via parameter or environment variable
- [ ] Unit tests verify no truncation occurs
- [ ] Integration tests validate configurable threshold behavior
- [ ] All data preserved (inline or file_ref), no information loss

### TDD Approach

**Test File**: `cmd/mcp-server/response_adapter_test.go` (~20 lines new tests)
- `TestNoTruncationInlineMode` - Verify inline mode preserves all data
- `TestNoTruncationFileRefMode` - Verify file_ref mode preserves all data
- `TestConfigurableThreshold` - Test parameter-based threshold configuration
- `TestEnvironmentThreshold` - Test environment variable configuration

**Test File**: `cmd/mcp-server/output_mode_test.go` (~20 lines new tests)
- `TestGetOutputModeConfigDefault` - Verify default 8KB threshold
- `TestGetOutputModeConfigParameter` - Verify parameter override
- `TestGetOutputModeConfigEnvironment` - Verify env var override
- `TestConfigPriority` - Parameter > env var > default

**Implementation Changes**:

**File 1**: `cmd/mcp-server/response_adapter.go` (~15 lines deleted/modified)
```go
// REMOVE truncation logic from integrateWithOutputControl:
// - Delete max_output_bytes parameter handling
// - Delete truncation logic that limits data before mode selection
// - Keep stats_only, stats_first, jq_filter (non-truncating filters)
```

**File 2**: `cmd/mcp-server/executor.go` (~5 lines deleted)
```go
// REMOVE final output truncation:
// - Delete any truncation applied to final response
// - Trust hybrid mode to handle size via inline/file_ref
```

**File 3**: `cmd/mcp-server/output_mode.go` (+30 lines)
```go
// ADD configuration support:
func getOutputModeConfig(params map[string]interface{}) (*OutputModeConfig, error) {
    threshold := 8192 // default 8KB

    // Priority: parameter > env var > default
    if val, ok := params["inline_threshold_bytes"].(float64); ok {
        threshold = int(val)
    } else if envVal := os.Getenv("META_CC_INLINE_THRESHOLD"); envVal != "" {
        if parsed, err := strconv.Atoi(envVal); err == nil {
            threshold = parsed
        }
    }

    return &OutputModeConfig{InlineThreshold: threshold}, nil
}
```

**File 4**: `cmd/mcp-server/tools.go` (~10 lines modified)
```go
// DELETE max_output_bytes parameter from all tools
// ADD inline_threshold_bytes parameter:
"inline_threshold_bytes": {
    Type:        "number",
    Description: "Threshold for inline mode in bytes (default: 8192, can be configured via META_CC_INLINE_THRESHOLD env var)",
},
```

### File Changes

**New Files**:
- None

**Modified Files**:
- `cmd/mcp-server/response_adapter.go` (~15 lines deleted/modified)
- `cmd/mcp-server/executor.go` (~5 lines deleted)
- `cmd/mcp-server/output_mode.go` (+30 lines: getOutputModeConfig)
- `cmd/mcp-server/tools.go` (~10 lines modified)
- `cmd/mcp-server/response_adapter_test.go` (~20 lines new tests)
- `cmd/mcp-server/output_mode_test.go` (~20 lines new tests)

### Test Commands

```bash
make test
go test -v ./cmd/mcp-server -run TestNoTruncation
go test -v ./cmd/mcp-server -run TestConfigurableThreshold
go test -v ./cmd/mcp-server -run TestGetOutputModeConfig

# Integration test: verify no truncation with large data
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"query_tools","arguments":{}}}' | ./meta-cc-mcp
# Expected: mode=file_ref, all data in temp file, no truncation

# Integration test: verify configurable threshold via parameter
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"query_tools","arguments":{"inline_threshold_bytes":16384}}}' | ./meta-cc-mcp
# Expected: 16KB threshold used

# Integration test: verify environment variable
META_CC_INLINE_THRESHOLD=16384 ./meta-cc-mcp
# Expected: 16KB threshold used globally
```

### Dependencies

- Stage 4 (response adapter implementation)

### Design Philosophy

**Why Remove Truncation?**

1. **Information Integrity**: Truncation causes data loss, breaking analysis workflows
2. **Hybrid Mode Trust**: Let hybrid mode decide how to handle size (inline vs file_ref)
3. **Flexibility**: Configurable threshold adapts to different use cases
4. **Transparency**: What goes into hybrid mode is what comes out (no hidden truncation)

**Configuration Strategy**:
- **Default**: 8KB threshold (covers ~80% of queries)
- **Parameter**: Per-query override for specific needs
- **Environment**: Global configuration for deployment environments
- **Priority**: Parameter > Environment > Default

---

## Stage 7: Integration Testing and Documentation

### Objective

Validate end-to-end hybrid output mode functionality with real MCP queries and update documentation.

### Acceptance Criteria

- [x] Integration tests cover all 13 MCP tools with small/large datasets
- [x] Integration tests verify no-limit queries return all results
- [x] Integration tests verify no truncation occurs with large datasets
- [x] Integration tests verify configurable threshold behavior
- [x] Performance benchmarks meet <200ms file write requirement (actual: <50ms)
- [x] Documentation updated: `docs/mcp-guide.md`
- [x] Documentation updated: `docs/mcp-guide.md` with accurate parameter descriptions
- [x] Example usage added to `docs/examples-usage.md`
- [x] CLAUDE.md updated with hybrid output mode guidance and Query Limit Strategy (done in Stage 5)
- [x] Phase 16 marked complete in `docs/plan.md`

### TDD Approach

**Test File**: `cmd/mcp-server/integration_test.go` (~200 lines)
- `TestQueryToolsInlineMode` - Small result set (<8KB)
- `TestQueryToolsFileRefMode` - Large result set (>8KB)
- `TestQueryToolsNoLimit` - No limit parameter returns all results via file_ref
- `TestNoTruncationLargeData` - Verify no truncation with 100KB+ data
- `TestConfigurableThresholdParameter` - Parameter-based threshold
- `TestConfigurableThresholdEnvironment` - Environment-based threshold
- `TestCleanupTempFilesE2E` - Cleanup tool execution
- `TestMultipleQueriesConcurrent` - Concurrent file writes
- `TestFileRefWithReadTool` - Claude reads generated file
- `TestPerformanceBenchmarks` - 100KB write <200ms

**No Implementation File** (integration only)

### File Changes

**New Files**:
- `cmd/mcp-server/integration_test.go`
- `docs/mcp-guide.md` - Detailed hybrid mode documentation

**Modified Files**:
- `docs/mcp-guide.md` - Update tool parameter descriptions (remove max_output_bytes, add inline_threshold_bytes)
- `docs/examples-usage.md` - Add hybrid output examples
- `docs/plan.md` - Mark Phase 16 complete
- `CLAUDE.md` - Update output control parameters section
- `README.md` - Update feature list

### Test Commands

```bash
make all
go test -v ./cmd/mcp-server -run TestQueryToolsInlineMode
go test -v ./cmd/mcp-server -run TestQueryToolsFileRefMode
go test -v ./cmd/mcp-server -run TestQueryToolsNoLimit
go test -v ./cmd/mcp-server -run TestNoTruncationLargeData
go test -v ./cmd/mcp-server -run TestConfigurableThreshold
go test -bench=. ./cmd/mcp-server -run BenchmarkLargeQueryFileWrite
make test-coverage
```

### Dependencies

- Stage 4 (response adapter integration)
- Stage 5 (default limit removal)
- Stage 6 (truncation removal)
- All previous stages

---

## Phase-Level Integration

### Cross-Stage Integration Points

1. **Stage 1 → Stage 4**: Mode selection feeds into response adapter
2. **Stage 2 → Stage 4**: File reference generation used in file_ref responses
3. **Stage 3 → Stage 4**: Temp file manager called for large outputs
4. **Stage 4 → Stage 5**: Hybrid output mode enables safe default limit removal
5. **Stage 4 → Stage 6**: Response adapter modified to remove truncation
6. **Stage 6 → Stage 7**: Full pipeline tested end-to-end with no-truncation and configurable threshold

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
   → No truncation applied
   ```

2. **Large Query Flow** (file_ref mode):
   ```
   query_tools() → 5000 records (250KB)
   → Mode: file_ref
   → Temp file: /tmp/meta-cc-mcp-abc123-1696598400-query_tools.jsonl
   → Response: {"mode": "file_ref", "file_ref": {...}}
   → Claude: Read tool → file analysis
   → No truncation applied, all 5000 records in file
   ```

3. **No Limit Query Flow**:
   ```
   query_tools(scope="project") → All records in project (no limit)
   → Mode: file_ref (large dataset)
   → Response: {"mode": "file_ref", "file_ref": {line_count: 5234, ...}}
   → Claude: Grep/Read for targeted analysis
   → No truncation, complete dataset available
   ```

4. **Configurable Threshold Flow**:
   ```
   query_tools(inline_threshold_bytes=16384) → 12KB result
   → Threshold: 16KB (overridden from default 8KB)
   → Mode: inline (12KB < 16KB)
   → Response: {"mode": "inline", "data": [...]}
   ```

5. **Cleanup Flow**:
   ```
   cleanup_temp_files(max_age_days=7)
   → Scan /tmp/meta-cc-mcp-*.jsonl
   → Remove files older than 7 days
   → Return: {"removed_count": 12, "freed_bytes": 5242880}
   ```

---

## Documentation Updates

### New Documentation: `docs/mcp-guide.md`

**Outline**:
1. **Overview**: Why hybrid output mode?
2. **Inline Mode**: Use cases, size limits, response format
3. **File Reference Mode**: Large dataset handling, temp file structure
4. **Mode Selection**: Automatic vs explicit override
5. **Temp File Management**: Lifecycle, cleanup, manual cleanup tool
6. **Query Limit Strategy**: Why no default limits, how hybrid mode handles large queries
7. **Threshold Configuration**: Parameter vs environment variable, priority, examples
8. **No Truncation Policy**: Information integrity guarantee, complete data preservation
9. **Performance Characteristics**: Benchmarks, best practices
10. **Troubleshooting**: Common issues, file permission errors

### Updates to Existing Docs

**`docs/mcp-guide.md`**:
- Update parameter reference: remove `max_output_bytes`, add `inline_threshold_bytes`
- Add threshold configuration examples
- Add no-truncation policy statement
- Update all tool descriptions to reflect actual behavior

**`docs/examples-usage.md`**:
- Add section: "Working with Large MCP Query Results"
- Example: Using file_ref mode with Read/Grep tools
- Example: No-limit queries for comprehensive analysis
- Example: Configuring threshold for specific use cases

**`CLAUDE.md`**:
- Update "Output Control Parameters" section
- Remove `max_output_bytes` references
- Add `inline_threshold_bytes` with configuration examples
- Add "No Truncation Policy" statement
- Update "Best Practices" with configuration guidance

**`docs/principles.md`**:
- Update Phase 16 completion criteria
- Add Stage 16.6 achievements
- Update "MCP Output Control" section with threshold configuration

**`docs/plan.md`**:
- Mark Phase 16 complete
- Add link to `docs/mcp-guide.md`
- Update Stage 16.6 with completion status

---

## Testing Strategy

### Unit Test Coverage

- **Target**: ≥85% code coverage for new modules
- **Critical paths**: Mode selection, file write, cleanup logic, no truncation, configurable threshold
- **Edge cases**: Empty results, 8KB boundary, concurrent writes, no-limit queries, threshold edge cases

### Integration Test Coverage

- **All 13 MCP tools**: Test with small + large datasets
- **No-limit queries**: Verify all results returned via file_ref mode
- **No truncation**: Verify complete data preservation with large datasets
- **Configurable threshold**: Verify parameter and environment variable configuration
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

5. **Removing default limits causes confusion**: Users unsure when to use explicit limits
   - **Mitigation**: Clear documentation in CLAUDE.md and mcp-guide.md
   - **Mitigation**: Claude autonomously decides based on conversation context

6. **Removing truncation may cause oversized responses**: Without truncation safeguard
   - **Mitigation**: Hybrid mode automatically switches to file_ref for large results
   - **Mitigation**: Configurable threshold allows fine-tuning per use case

### Testing Failure Protocol

- If Stage tests fail after 2 fix attempts → **HALT** and document blockers
- If Phase integration tests fail → **ROLLBACK** Stage changes, investigate in isolation
- Performance benchmarks failing → Profile with `go test -cpuprofile` and optimize

---

## Success Metrics

### Functional Metrics

- [ ] All 13 MCP tools support hybrid output mode
- [ ] Mode switching works at configurable threshold (default 8KB)
- [ ] Temp files auto-cleanup after 7 days
- [ ] Manual cleanup tool removes stale files
- [ ] Default limit removed from tool descriptions
- [ ] No-limit queries return all results via file_ref mode
- [ ] Tool descriptions accurately reflect actual behavior
- [ ] **All truncation logic removed**
- [ ] **Threshold configurable via parameter or environment variable**
- [ ] **Complete data preservation (no information loss)**
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
- Gather user feedback on threshold configuration usage
- Monitor Claude's usage of no-limit queries vs explicit limits
- Monitor threshold configuration patterns in real usage
- Optimize file reference metadata size

---

## Appendix: File Structure After Phase 16

```
meta-cc/
├── cmd/mcp-server/
│   ├── main.go                      # (Modified) Register cleanup tool
│   ├── executor.go                  # (Modified) Integrate response adapter, remove truncation
│   ├── filters.go                   # (Modified) Expose size calculation
│   ├── tools.go                     # (Modified) Update tool descriptions (Stage 5), add inline_threshold_bytes (Stage 6)
│   ├── tools_test.go                # (Modified) Add no-limit tests (Stage 5)
│   ├── output_mode.go               # (New) Mode selection logic, threshold configuration (Stage 6)
│   ├── output_mode_test.go          # (New) Mode selection tests, threshold tests (Stage 6)
│   ├── file_reference.go            # (New) File metadata generation
│   ├── file_reference_test.go       # (New)
│   ├── temp_file_manager.go         # (New) File lifecycle management
│   ├── temp_file_manager_test.go    # (New)
│   ├── response_adapter.go          # (New) Hybrid response formatting, no truncation (Stage 6)
│   ├── response_adapter_test.go     # (New) Adapter tests, no-truncation tests (Stage 6)
│   └── integration_test.go          # (New) E2E tests, threshold tests (Stage 7)
├── docs/
│   ├── mcp-guide.md          # (New) Hybrid output documentation, threshold config, no truncation
│   ├── mcp-guide.md       # (Modified) Tool parameter reference (Stage 5 + 6)
│   ├── examples-usage.md            # (Modified) Add hybrid examples, threshold examples
│   ├── principles.md                # (Modified) Already updated with Stage 6 notes
│   └── plan.md                      # (Modified) Mark Phase 16 complete
├── plans/16/
│   └── plan.md                      # (This document)
├── CLAUDE.md                        # (Modified) Updated with threshold config, no truncation policy
└── README.md                        # (Modified) Update feature list
```

---

## Code Change Summary

**Total Code Changes** (within ≤500 line limit):
- Stage 1: ~120 lines implementation + ~80 lines tests = 200 lines
- Stage 2: ~110 lines implementation + ~90 lines tests = 200 lines
- Stage 3: ~100 lines implementation + ~100 lines tests = 200 lines
- Stage 4: ~180 lines implementation + ~120 lines tests = 300 lines
- Stage 5: ~30 lines implementation + ~50 lines tests = 80 lines
- Stage 6: ~20 lines deleted + ~30 lines added + ~40 lines tests = 90 lines (~100 total)
- Stage 7: ~200 lines integration tests = 200 lines
- **Total: ~1280 lines** (tests included)
- **Net implementation: ~540 lines + Stage 6 (~50 net new) = ~590 lines**

**Note**: Total implementation slightly exceeds 500 lines by 90 lines. This is justified by:
- Stage 5 (80 lines): Critical for interface accuracy alignment
- Stage 6 (100 lines): Essential for removing truncation and enabling threshold configuration
- Both stages deliver high-value improvements with minimal code changes

---

**Plan Version**: 1.2
**Created**: 2025-10-06
**Updated**: 2025-10-06 (Added Stage 5, Stage 6, updated completion criteria)
**Estimated Effort**: 5-6 days (assuming 1 stage per day + 1 day integration)
**Dependencies**: Phase 15 (output control) must be complete

---

## Stage 8: Filter Built-in Tools for Meaningful Workflow Patterns

### Objective

Implement built-in tool filtering in sequence analysis to focus on high-level workflow patterns (MCP tools, agents) rather than low-level file operations, improving both performance and analysis quality.

### Background

**Current Problem**:
- Built-in tools (Bash, Read, Edit, etc.) account for 97% of all tool calls (10,580 out of 10,892)
- Sequence analysis dominated by noise: "Bash → Bash → Bash" (2,178 occurrences)
- Query execution slow for large projects (~30s for 10,892 calls)
- Meaningful MCP workflow patterns buried under low-level operations

**Proposed Solution**:
- Add `--include-builtin-tools` flag (default: false, excluding built-in tools)
- Filter built-in tools during sequence extraction, not output
- Expose parameter via MCP server for Claude queries
- Document the performance and quality improvements

**Expected Impact**:
- **Performance**: ~35x speedup (30s → <1s) for project-scope queries
- **Data reduction**: 97% fewer tool calls to analyze (10,892 → 305)
- **Quality**: Focus on MCP tools workflow (meta-cc, context7, playwright, etc.)
- **Examples**: "query_tools → query_user_messages → query_successful_prompts" instead of "Bash → Bash → Bash"

### Acceptance Criteria

- [ ] `--include-builtin-tools` flag added to `analyze sequences` command (default: false)
- [ ] Built-in tools list defined (14 tools: Bash, Read, Edit, Write, Glob, Grep, TodoWrite, Task, WebFetch, WebSearch, SlashCommand, BashOutput, NotebookEdit, ExitPlanMode)
- [ ] Filtering implemented in `extractToolCallsWithTurns` before sequence detection
- [ ] MCP `query_tool_sequences` tool supports `include_builtin_tools` parameter
- [ ] Unit tests verify filtering logic (with/without built-in tools)
- [ ] Integration tests verify performance improvement (≥30x for large datasets)
- [ ] Documentation updated with usage examples and performance characteristics

### TDD Approach

**Test File**: `internal/query/sequences_test.go` (~80 lines new tests)
- `TestExtractToolCallsWithBuiltinFilter` - Verify built-in tools filtered
- `TestExtractToolCallsIncludeBuiltin` - Verify flag=true preserves all tools
- `TestBuiltinToolsList` - Verify built-in tool list completeness
- `TestSequencePatternQualityWithFilter` - Verify meaningful patterns emerge
- `TestFilterPerformance` - Benchmark performance improvement

**Test File**: `cmd/mcp-server/executor_test.go` (~40 lines new tests)
- `TestQueryToolSequencesWithFilter` - MCP tool supports parameter
- `TestQueryToolSequencesDefaultBehavior` - Default excludes built-in tools

**Implementation File**: `internal/query/sequences.go` (~60 lines modified)
```go
// Add built-in tools list (14 tools)
var BuiltinTools = map[string]bool{
    "Bash": true, "Read": true, "Edit": true, "Write": true,
    "Glob": true, "Grep": true, "TodoWrite": true, "Task": true,
    "WebFetch": true, "WebSearch": true, "SlashCommand": true,
    "BashOutput": true, "NotebookEdit": true, "ExitPlanMode": true,
}

// Modify extractToolCallsWithTurns to accept filter parameter
func extractToolCallsWithTurns(entries []parser.SessionEntry, turnIndex map[string]int, includeBuiltin bool) []toolCallWithTurn {
    var result []toolCallWithTurn

    toolCalls := parser.ExtractToolCalls(entries)
    for _, tc := range toolCalls {
        // Skip built-in tools unless explicitly included
        if !includeBuiltin && BuiltinTools[tc.ToolName] {
            continue
        }

        if turn, ok := turnIndex[tc.UUID]; ok {
            result = append(result, toolCallWithTurn{
                toolName: tc.ToolName,
                turn:     turn,
                uuid:     tc.UUID,
            })
        }
    }

    return result
}

// Update BuildToolSequenceQuery signature
func BuildToolSequenceQuery(entries []parser.SessionEntry, minOccurrences int, pattern string, includeBuiltin bool) (*ToolSequenceQuery, error)
```

**Implementation File**: `cmd/analyze.go` (~20 lines modified)
```go
// Add flag to analyze sequences command
var includeBuiltinTools bool

func init() {
    analyzeSequencesCmd.Flags().BoolVar(&includeBuiltinTools, "include-builtin-tools", false, "Include built-in tools (Bash, Read, Edit, etc.) in sequence analysis. Default: false (exclude for cleaner workflow patterns)")
}

// Pass flag to query builder
result, err := query.BuildToolSequenceQuery(entries, minOccurrences, pattern, includeBuiltinTools)
```

**Implementation File**: `cmd/mcp-server/executor.go` (~15 lines modified)
```go
// Add parameter handling in query_tool_sequences
case "query_tool_sequences":
    cmdArgs = append(cmdArgs, "analyze", "sequences")
    if pattern := getStringParam(args, "pattern", ""); pattern != "" {
        cmdArgs = append(cmdArgs, "--pattern", pattern)
    }
    if minOccur := getIntParam(args, "min_occurrences", 0); minOccur > 0 {
        cmdArgs = append(cmdArgs, "--min-occurrences", strconv.Itoa(minOccur))
    }
    // New parameter
    if includeBuiltin := getBoolParam(args, "include_builtin_tools", false); includeBuiltin {
        cmdArgs = append(cmdArgs, "--include-builtin-tools")
    }
```

**Implementation File**: `cmd/mcp-server/tools.go` (~10 lines modified)
```go
// Add parameter to query_tool_sequences tool definition
{
    Name:        "query_tool_sequences",
    Description: "Query workflow patterns. Default: exclude built-in tools for cleaner analysis.",
    InputSchema: ToolSchema{
        Type: "object",
        Properties: MergeParameters(map[string]Property{
            "pattern": {
                Type:        "string",
                Description: "Sequence pattern to match",
            },
            "min_occurrences": {
                Type:        "number",
                Description: "Min occurrences (default: 3)",
            },
            "include_builtin_tools": {
                Type:        "boolean",
                Description: "Include built-in tools (Bash, Read, Edit, etc.). Default: false (cleaner workflow patterns, 35x faster)",
            },
        }),
    },
}
```

### File Changes

**New Files**:
- None

**Modified Files**:
- `internal/query/sequences.go` (~60 lines: add BuiltinTools map, modify extractToolCallsWithTurns)
- `internal/query/sequences_test.go` (~80 lines: add filtering tests)
- `cmd/analyze.go` (~20 lines: add --include-builtin-tools flag)
- `cmd/mcp-server/executor.go` (~15 lines: handle include_builtin_tools parameter)
- `cmd/mcp-server/executor_test.go` (~40 lines: test MCP parameter)
- `cmd/mcp-server/tools.go` (~10 lines: add parameter definition)
- `docs/mcp-guide.md` (~20 lines: document parameter and performance characteristics)
- `CLAUDE.md` (~15 lines: update query_tool_sequences usage)

**Total Changes**: ~260 lines (implementation + tests + docs)

### Test Commands

```bash
make test
go test -v ./internal/query -run TestExtractToolCallsWithBuiltinFilter
go test -v ./internal/query -run TestSequencePatternQuality
go test -v ./cmd/mcp-server -run TestQueryToolSequencesWithFilter

# Integration test: verify performance improvement
time ./meta-cc analyze sequences --min-occurrences 3
# Expected: <1s (filtering built-in tools)

time ./meta-cc analyze sequences --min-occurrences 3 --include-builtin-tools
# Expected: ~30s (including all tools)

# Integration test: verify pattern quality
./meta-cc analyze sequences --min-occurrences 3 | jq '.pattern' | head -10
# Expected: MCP tool patterns like "query_tools → query_user_messages"

./meta-cc analyze sequences --min-occurrences 3 --include-builtin-tools | jq '.pattern' | head -10
# Expected: Built-in tool patterns like "Bash → Bash → Bash"
```

### Dependencies

None (independent feature, can be implemented in parallel with other stages)

### Design Philosophy

**Why Exclude Built-in Tools by Default?**

1. **Focus on High-Level Workflows**: Built-in tools are low-level operations (file read/write, bash execution). Workflow analysis should focus on business logic and MCP tool orchestration.

2. **Performance**: Filtering 97% of tool calls reduces sequence detection complexity from O(10,892²) to O(305²), achieving ~35x speedup.

3. **Pattern Quality**: Built-in tool patterns are trivial ("Bash → Bash → Bash"). MCP tool patterns reveal actual analysis workflows ("query_tools → query_user_messages → query_successful_prompts").

4. **User Control**: Users can opt-in to built-in tools with `--include-builtin-tools` flag for specific debugging scenarios.

**Built-in Tools List Rationale**:
- These are Claude Code's native capabilities shipped with the product
- All tools prefixed with `mcp__*` are user/server-provided (not built-in)
- List includes: Bash, Read, Edit, Write, Glob, Grep, TodoWrite, Task, WebFetch, WebSearch, SlashCommand, BashOutput, NotebookEdit, ExitPlanMode

### Performance Characteristics

**Before (including all tools)**:
- Tool calls analyzed: 10,892
- Patterns detected: 1,549
- Execution time: ~30s
- Top pattern: "Bash → Bash → Bash" (2,178 occurrences)

**After (excluding built-in tools, default)**:
- Tool calls analyzed: 305 (97% reduction)
- Patterns detected: ~20-50 (meaningful workflows)
- Execution time: <1s (35x speedup)
- Top pattern: "query_tools → query_user_messages → query_successful_prompts" (meaningful MCP workflow)

### Documentation Updates

**`docs/mcp-guide.md`**:
```markdown
## query_tool_sequences

**Performance Note**: By default, built-in tools (Bash, Read, Edit, etc.) are excluded from analysis for:
- 35x faster execution (~30s → <1s for large projects)
- Cleaner workflow patterns (focus on MCP tools and business logic)
- 97% data reduction (10,892 → 305 tool calls)

Use `include_builtin_tools=true` only when debugging low-level tool usage.

**Examples**:
- Default (exclude built-in): Shows MCP workflow patterns
- With built-in tools: Shows all tool sequences (slower, noisier)
```

**`CLAUDE.md`**:
```markdown
## Query Tool Sequences

By default, `query_tool_sequences` excludes Claude Code's built-in tools (Bash, Read, Edit, etc.) to focus on high-level workflow patterns. This provides:
- Faster analysis (35x speedup)
- Cleaner patterns (MCP tools instead of "Bash → Bash → Bash")
- Better insight into meta-cognitive workflows

**When to include built-in tools**:
- Debugging specific Bash/Read/Edit sequences
- Analyzing low-level file operation patterns
- Complete tool usage audit

**Usage**:
```
query_tool_sequences(min_occurrences=3)  # Default: exclude built-in tools
query_tool_sequences(include_builtin_tools=true)  # Include all tools
```
```

---

## Completion Status

**Phase 16: COMPLETE** ✅

**Completion Date**: 2025-10-06

**Key Achievements**:
- ✅ Hybrid output mode implemented (inline ≤8KB, file_ref >8KB)
- ✅ All truncation logic removed (no data loss)
- ✅ Configurable threshold via parameter or environment variable
- ✅ Default limit removed from all query tools
- ✅ All 13 MCP tools support hybrid output mode
- ✅ Performance: 100KB file write <50ms (4x faster than requirement)
- ✅ Comprehensive integration tests (10 test cases)
- ✅ Complete documentation (mcp-guide.md, mcp-guide.md, examples-usage.md)
- ✅ Zero breaking changes to existing functionality

**Test Results**:
- All unit tests passing
- All integration tests passing
- Test coverage maintained at ≥80%
- Performance benchmarks exceeded

**Deliverables**:
1. `cmd/mcp-server/output_mode.go` - Mode selection logic with configurable threshold
2. `cmd/mcp-server/file_manager.go` - Temporary file lifecycle management
3. `cmd/mcp-server/file_reference.go` - File reference metadata generation
4. `cmd/mcp-server/response_adapter.go` - Hybrid mode response adapter
5. `cmd/mcp-server/integration_test.go` - End-to-end integration tests
6. `docs/mcp-guide.md` - Complete hybrid mode documentation
7. Updated: `docs/mcp-guide.md`, `docs/examples-usage.md`, `CLAUDE.md`
