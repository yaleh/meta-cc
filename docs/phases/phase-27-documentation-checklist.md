# Phase 27 Documentation Checklist

**Purpose:** Complete Phase 27 by creating the remaining documentation files.

**Status:** 95% Complete (documentation pending)

**Estimated Time:** 2.5-4.5 hours

---

## Critical Documentation Tasks

### 1. Two-Stage Query Guide ⚠️ PENDING
**File:** `docs/guides/two-stage-query-guide.md`
**Estimate:** 1-2 hours
**Lines:** ~200

**Required Sections:**
- [ ] Overview (What is two-stage query? Why?)
- [ ] Architecture diagram (Stage 1 → Stage 2 flow)
- [ ] Stage 1: File Selection
  - [ ] get_session_directory examples
  - [ ] inspect_session_files examples
- [ ] Stage 2: Query Execution
  - [ ] execute_stage2_query examples
  - [ ] jq filter syntax
  - [ ] sort, transform, limit parameters
- [ ] Complete Workflow Example
- [ ] Performance Comparison (old vs. new)
- [ ] Best Practices
- [ ] Common Patterns

**Reference Materials:**
- `docs/analysis/stage2-feasibility-summary.md` - Architecture decisions
- `internal/query/stage2_executor.go` - Implementation details
- `internal/query/file_inspector.go` - Inspection logic
- `cmd/mcp-server/handlers_stage1.go` - Stage 1 tools
- `cmd/mcp-server/handlers_stage2.go` - Stage 2 tool

**Example Structure:**
```markdown
# Two-Stage Query Architecture

## Overview

The two-stage query architecture separates file selection (Stage 1) from query execution (Stage 2), providing 100-600x performance improvements.

## Why Two-Stage?

**Problem:** Legacy `query` tool processed all 660 files every time (2-5 seconds)
**Solution:** Users select relevant files first (Stage 1), then query only those files (Stage 2)

## Stage 1: File Selection

### get_session_directory

Returns list of all session files with metadata.

**Example:**
```javascript
get_session_directory({scope: "project"})
```

[Continue with examples...]
```

---

### 2. Query Examples Library ⚠️ PENDING
**File:** `docs/examples/two-stage-query-examples.md`
**Estimate:** 1-2 hours
**Lines:** ~300

**Required Examples (5+):**
- [ ] Example 1: Find all errors in recent sessions
- [ ] Example 2: Analyze token usage over time
- [ ] Example 3: Track conversation flow for specific topic
- [ ] Example 4: Performance comparison (old vs. new)
- [ ] Example 5: Complex query with filter + sort + transform
- [ ] Example 6: Inspect file metadata before querying
- [ ] Example 7: Use convenience tools vs. two-stage

**Example Template:**
```markdown
## Example 1: Find All Errors in Recent Sessions

### Objective
Find all tool execution errors from the last 10 sessions.

### Step 1: Get Session Directory
```javascript
get_session_directory({scope: "project"})
```

**Output:**
```json
{
  "directory": "/home/user/.claude/projects/...",
  "file_count": 660,
  "newest_file": "2025-10-26T12:00:00Z",
  "oldest_file": "2025-10-01T10:00:00Z"
}
```

### Step 2: Inspect Recent Files
```javascript
inspect_session_files({
  scope: "project",
  files: ["file1.jsonl", "file2.jsonl"],
  include_samples: true,
  sample_size: 3
})
```

[Continue with Step 3: Execute Query...]

### Performance
- Old method: 2.5 seconds
- New method: 19ms
- **Speedup: 131x**
```

---

### 3. Update MCP Guide ⚠️ PENDING
**File:** `docs/guides/mcp.md`
**Estimate:** 30 minutes
**Lines:** ~40

**Required Updates:**
- [ ] Update tool count (15 → 16 tools)
- [ ] Add "Two-Stage Query Architecture" section
- [ ] Document 3 new tools:
  - [ ] get_session_directory
  - [ ] inspect_session_files
  - [ ] execute_stage2_query
- [ ] Add workflow diagram
- [ ] Add performance comparison
- [ ] Add migration notes (query/query_raw deleted)
- [ ] Link to two-stage-query-guide.md
- [ ] Link to two-stage-query-examples.md

**Section Template:**
```markdown
## Two-Stage Query Architecture

**New in v2.1.0:** The MCP server now uses a two-stage query architecture for 100-600x performance improvements.

### Stage 1: File Selection

**get_session_directory**
- **Purpose:** List all session files
- **Parameters:** scope (project|session)
- **Returns:** Directory path, file count, date range

**inspect_session_files**
- **Purpose:** Examine file metadata and preview content
- **Parameters:** scope, files[], include_samples, sample_size
- **Returns:** File metadata, record counts, sample entries

### Stage 2: Query Execution

**execute_stage2_query**
- **Purpose:** Execute jq queries on selected files
- **Parameters:** scope, files[], filter, sort_by, sort_order, transform, limit
- **Returns:** Query results with metadata

### Performance

| Method | Files Processed | Execution Time |
|--------|----------------|----------------|
| Legacy (query) | All (660) | 2-5 seconds |
| Two-Stage | Selected (10) | 8-19ms |
| **Speedup** | | **100-600x** |

### Migration

**Removed Tools:**
- `query` - Use two-stage workflow
- `query_raw` - Use two-stage workflow

**Convenience Tools Preserved:**
- `query_tool_errors` - Still available
- `query_token_usage` - Still available
- [List other 8 tools...]

See [Two-Stage Query Guide](two-stage-query-guide.md) for complete documentation.
```

---

## Optional Tasks (Low Priority)

### 4. Fix E2E Tests ⚠️ OPTIONAL
**File:** `tests/e2e/mcp-e2e-simple.sh`
**Issue:** Tests 7-10, 12 have shell escaping issues
**Estimate:** 30-60 minutes

**Problem:**
- Nested JSON in bash requires complex escaping
- Tests work manually but fail in script

**Solution Approaches:**
1. Use temporary JSON files instead of inline strings
2. Simplify test assertions
3. Use jq to build JSON payloads
4. Accept current state (8/12 passing is acceptable)

**Recommendation:** Defer to Phase 28. Current E2E coverage is sufficient.

---

## Completion Criteria

**Phase 27 is complete when:**
- [x] All 5 implementation stages complete (27.1-27.5)
- [x] All unit tests passing
- [x] All integration tests passing
- [x] Basic E2E tests passing
- [x] Performance target met (< 100ms)
- [ ] Two-stage query guide created
- [ ] Query examples library created
- [ ] MCP guide updated
- [ ] All documentation reviewed

**Additional Nice-to-Haves:**
- [ ] E2E tests 100% passing (currently 67%)
- [ ] Performance benchmarks documented
- [ ] User migration guide

---

## Quick Start

To complete Phase 27 documentation:

1. **Read Reference Materials** (30 min)
   - Review `docs/analysis/stage2-feasibility-summary.md`
   - Review `internal/query/stage2_executor.go`
   - Review existing MCP guide format

2. **Create Two-Stage Guide** (1-2 hours)
   - Use template above
   - Include architecture diagram (ASCII art is fine)
   - Add 2-3 workflow examples
   - Document best practices

3. **Create Examples Library** (1-2 hours)
   - Use template above
   - Write 5-7 practical examples
   - Include expected outputs
   - Add performance comparisons

4. **Update MCP Guide** (30 min)
   - Add Phase 27 section
   - Update tool count
   - Add links to new guides
   - Document migration path

5. **Review and Test** (30 min)
   - Read all documentation
   - Verify examples work
   - Check links
   - Proofread

**Total Time:** 2.5-4.5 hours

---

## Success Metrics

**Documentation Quality:**
- [ ] Clear and concise writing
- [ ] Working code examples
- [ ] Architecture diagrams included
- [ ] Performance data included
- [ ] Migration path documented

**Completeness:**
- [ ] All 3 new tools documented
- [ ] 5+ practical examples
- [ ] Workflow diagrams
- [ ] Best practices
- [ ] Migration notes

**Usability:**
- [ ] Easy to find (linked from main docs)
- [ ] Easy to understand (clear structure)
- [ ] Easy to use (copy-paste examples)

---

## Resources

**Internal Documentation:**
- `docs/analysis/stage2-feasibility-summary.md`
- `docs/analysis/stage2-go-implementation-feasibility.md`
- `internal/query/stage2_executor.go`
- `cmd/mcp-server/handlers_stage1.go`
- `cmd/mcp-server/handlers_stage2.go`

**Test Files:**
- `internal/query/stage2_executor_test.go`
- `internal/query/file_inspector_test.go`
- `tests/e2e/mcp-e2e-simple.sh`

**Existing Guides:**
- `docs/guides/mcp.md` - Current MCP guide
- `docs/guides/unified-query-api.md` - Related query documentation
- `docs/examples/mcp-query-cookbook.md` - Example format reference

---

## Contact / Questions

If you have questions while completing documentation:
1. Review reference materials listed above
2. Check test files for usage examples
3. Run tools manually for output examples
4. Refer to this checklist for structure guidance

---

**Ready to start?** Begin with Section 1 (Two-Stage Query Guide) and work through the checklist sequentially.
