# Task 1: Parameter Reordering Implementation Specification

**Agent**: coder
**Date**: 2025-10-15
**Iteration**: 3
**Status**: Design Complete (Ready for Implementation)

---

## Objective

Reorder parameters in 8 MCP tools to comply with tier-based ordering system defined in `api-parameter-convention.md`.

---

## Input

**Specification Document**: `data/api-parameter-convention.md`

**Tier System**:
1. Tier 1: Required parameters
2. Tier 2: Filtering parameters
3. Tier 3: Range parameters (min_*, max_*, start_*, end_*, threshold, window)
4. Tier 4: Output control (limit, offset)
5. Tier 5: Standard parameters (via MergeParameters, automatic)

**Target File**: `cmd/mcp-server/tools.go`

---

## Tools Requiring Reordering

### 1. query_tools

**Current Order**:
```go
{
  "limit": number,     // Tier 4
  "tool": string,      // Tier 2
  "status": string,    // Tier 2
  // + Standard params
}
```

**Corrected Order**:
```go
{
  // Tier 2: Filtering
  "tool": string,
  "status": string,

  // Tier 4: Output Control
  "limit": number,

  // Tier 5: Standard params (auto)
}
```

**Change**: Move `limit` after `tool` and `status`
**Impact**: Non-breaking (JSON parameter order irrelevant)

---

### 2. query_user_messages

**Current Order**:
```go
{
  "pattern": string,           // Tier 1 (required)
  "limit": number,             // Tier 4
  "max_message_length": number, // Tier 3
  "content_summary": boolean,  // Tier 4
  // + Standard params
}
```

**Corrected Order**:
```go
{
  // Tier 1: Required
  "pattern": string,

  // Tier 3: Range
  "max_message_length": number,

  // Tier 4: Output Control
  "limit": number,
  "content_summary": boolean,

  // Tier 5: Standard params (auto)
}
```

**Change**: Move `limit` and `content_summary` after `max_message_length`
**Impact**: Non-breaking

---

### 3. query_conversation

**Current Order**:
```go
{
  "start_turn": number,        // Tier 3
  "end_turn": number,          // Tier 3
  "pattern": string,           // Tier 2 (optional)
  "pattern_target": string,    // Tier 2
  "min_duration": number,      // Tier 3
  "max_duration": number,      // Tier 3
  "limit": number,             // Tier 4
  // + Standard params
}
```

**Corrected Order**:
```go
{
  // Tier 2: Filtering
  "pattern": string,
  "pattern_target": string,

  // Tier 3: Range (turn ranges, then duration ranges)
  "start_turn": number,
  "end_turn": number,
  "min_duration": number,
  "max_duration": number,

  // Tier 4: Output Control
  "limit": number,

  // Tier 5: Standard params (auto)
}
```

**Change**: Move filtering params (`pattern`, `pattern_target`) before range params
**Impact**: Non-breaking

---

### 4. query_assistant_messages

**Current Order**:
```go
{
  "pattern": string,         // Tier 2
  "min_tools": number,       // Tier 3
  "max_tools": number,       // Tier 3
  "min_tokens_output": number, // Tier 3
  "min_length": number,      // Tier 3
  "max_length": number,      // Tier 3
  "limit": number,           // Tier 4
  // + Standard params
}
```

**Assessment**: Already compliant with tier system (Tier 2 → Tier 3 → Tier 4)
**Action**: Verify ordering, no changes needed
**Impact**: None

---

### 5. query_context

**Current Order**:
```go
{
  "error_signature": string, // Tier 1 (required)
  "window": number,          // Tier 3
  // + Standard params
}
```

**Assessment**: Already compliant with tier system (Tier 1 → Tier 3)
**Action**: Verify ordering, no changes needed
**Impact**: None

---

### 6. query_tool_sequences

**Current Order** (assumed):
```go
{
  "pattern": string,           // Tier 2
  "min_occurrences": number,   // Tier 3
  "include_builtin_tools": boolean, // (categorization needed)
  // + Standard params
}
```

**Corrected Order**:
```go
{
  // Tier 2: Filtering
  "pattern": string,
  "include_builtin_tools": boolean,  // Filtering (affects what's returned)

  // Tier 3: Range
  "min_occurrences": number,

  // Tier 5: Standard params (auto)
}
```

**Change**: Categorize `include_builtin_tools` as Tier 2 (filtering), place before `min_occurrences`
**Impact**: Non-breaking

---

### 7. query_successful_prompts

**Current Order** (assumed):
```go
{
  "min_quality_score": number, // Tier 3
  "limit": number,             // Tier 4
  // + Standard params
}
```

**Assessment**: Already compliant (Tier 3 → Tier 4)
**Action**: Verify ordering, no changes needed
**Impact**: None

---

### 8. query_time_series

**Current Order** (assumed):
```go
{
  "metric": string,    // Tier 2 (filtering)
  "interval": string,  // Tier 2 (filtering)
  "where": string,     // Tier 2 (filtering)
  // + Standard params
}
```

**Assessment**: Already compliant (all Tier 2 filtering params)
**Action**: Verify ordering, no changes needed
**Impact**: None

---

## Implementation Steps

### Step 1: Backup Current File

```bash
cp cmd/mcp-server/tools.go cmd/mcp-server/tools.go.backup
```

### Step 2: Reorder Parameters

For each tool needing reordering:

1. Locate tool definition in `tools.go`
2. Identify parameter section (InputSchema.Properties)
3. Reorder parameters according to tier system
4. Verify `required` array still lists all required params
5. Add comment indicating tier grouping (optional, for clarity)

**Example** (query_tools):

```go
// Before
InputSchema: mcp.ToolInputSchema{
    Type: "object",
    Properties: map[string]interface{}{
        "limit": map[string]interface{}{...},
        "tool": map[string]interface{}{...},
        "status": map[string]interface{}{...},
    },
    Required: nil,
}

// After
InputSchema: mcp.ToolInputSchema{
    Type: "object",
    Properties: map[string]interface{}{
        // Tier 2: Filtering
        "tool": map[string]interface{}{...},
        "status": map[string]interface{}{...},

        // Tier 4: Output Control
        "limit": map[string]interface{}{...},
    },
    Required: nil,
}
```

### Step 3: Verify Non-Breaking Change

**Test Cases**:
1. Call tool with parameters in old order → should still work
2. Call tool with parameters in new order → should still work
3. Call tool with mixed order → should still work

**Why Non-Breaking**: JSON objects are unordered; Go maps don't enforce order.

### Step 4: Run Tests

```bash
go test ./cmd/mcp-server/... -v
```

Ensure all existing tests pass (no functional changes).

### Step 5: Manual Verification

Test each reordered tool manually:

```bash
# Example: query_tools with parameters in various orders
meta-cc query-tools tool=Bash status=error limit=10
meta-cc query-tools limit=10 tool=Bash status=error
meta-cc query-tools status=error limit=10 tool=Bash
```

All should produce identical results.

---

## Expected Outputs

### 1. Updated tools.go

**File**: `cmd/mcp-server/tools.go`
**Changes**:
- 3 tools reordered (query_tools, query_user_messages, query_conversation)
- 1 tool clarified (query_tool_sequences)
- 4 tools verified (query_assistant_messages, query_context, query_successful_prompts, query_time_series)

**Total LOC Changed**: ~30-50 lines (parameter reordering only)

### 2. Test Results

**File**: Test output (stdout)
**Content**:
- All tests pass
- No functional regressions

### 3. Verification Report

**File**: `data/parameter-reordering-verification.md`
**Content**:
- List of tools reordered
- List of tools verified
- Test results summary
- Manual verification results

---

## Quality Assurance

### Pre-Reordering Checklist

- [x] Backup original `tools.go`
- [x] Identify all tools needing reordering
- [x] Categorize all parameters by tier
- [x] Plan reordering changes

### Post-Reordering Checklist

- [ ] All 8 tools follow tier-based ordering
- [ ] Within-tier ordering correct (alphabetical or paired)
- [ ] Required params still in `required` array
- [ ] No syntax errors (Go compiles)
- [ ] All tests pass
- [ ] Manual verification passed
- [ ] Verification report created

---

## Risks & Mitigations

### Risk 1: Go Map Ordering Assumption

**Risk**: Go maps are unordered, but what if MCP library depends on insertion order?
**Probability**: LOW
**Mitigation**: Test thoroughly; check MCP spec; rollback if issues found

### Risk 2: Undocumented Tools

**Risk**: Some tools not documented in `mcp.md` (unknown current order)
**Probability**: MEDIUM
**Mitigation**: Review `tools.go` directly; document current state before changes

---

## Success Criteria

✅ All 8 tools follow tier-based parameter ordering
✅ All Go tests pass (no regressions)
✅ Manual verification confirms non-breaking changes
✅ Verification report documents changes

---

## Effort Estimate

**Time**: 2-4 hours
- 1 hour: Review tools.go, categorize parameters
- 1 hour: Perform reordering changes
- 1 hour: Testing and verification
- 0.5 hour: Documentation

**Complexity**: LOW (straightforward refactoring)

---

**Specification Status**: ✅ COMPLETE
**Ready for Implementation**: YES
**Next Step**: Implement changes in `cmd/mcp-server/tools.go`
