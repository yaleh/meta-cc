# API Parameter Ordering Convention

**Version**: 1.0
**Created**: 2025-10-15 (Iteration 2, bootstrap-006-api-design)
**Scope**: meta-cc MCP Server API (16 tools)
**Purpose**: Establish consistent parameter ordering for API tools

---

## Executive Summary

This document defines parameter ordering conventions for meta-cc MCP API tools to improve consistency, readability, and maintainability. The convention establishes a tier-based ordering system that prioritizes required parameters, followed by filtering, range, and output control parameters.

**Key Principles**:
1. **Required First**: Required parameters appear before optional parameters
2. **Logical Grouping**: Parameters grouped by purpose (filtering, range, output)
3. **Predictability**: Consistent ordering across all tools
4. **Backward Compatible**: JSON parameter order doesn't affect function calls

**Current State**: ~60% parameter ordering consistency (tool-specific params)
**Target State**: 100% consistency through tier-based ordering

---

## Tier-Based Ordering System

### Overview

All parameters are categorized into 5 tiers, ordered from highest to lowest priority:

```
Tier 1: Required Parameters
Tier 2: Filtering Parameters
Tier 3: Range Parameters
Tier 4: Output Control Parameters
Tier 5: Standard Parameters (via MergeParameters)
```

**Ordering Rule**: Parameters within each tier are listed before parameters in lower tiers.

---

### Tier 1: Required Parameters

**Purpose**: Parameters that must be provided for the tool to function

**Characteristics**:
- Marked as `required: true` in schema
- Tool cannot execute without these parameters
- Typically identifiers, patterns, or essential filters

**Examples**:
- `pattern` (query_user_messages, query_assistant_messages)
- `error_signature` (query_context)
- `file` (query_file_access)
- `where` (query_tools_advanced)
- `name` (get_capability)

**Ordering within Tier 1**:
- **Single required param**: Place first
- **Multiple required params**: Alphabetical order

**Rationale**:
- Required params are conceptually primary (define the query)
- Placing first improves clarity (users see what's essential)
- Early error detection (missing required param fails fast)
- Industry standard (GraphQL, SQL required clauses first)

---

### Tier 2: Filtering Parameters

**Purpose**: Parameters that filter or refine the query results

**Characteristics**:
- Optional parameters
- Affect which records are returned (not how many)
- Typically string or enum types
- Semantically narrow the search space

**Examples**:
- `tool` (query_tools) - Filter by tool name
- `status` (query_tools) - Filter by error/success
- `pattern_target` (query_conversation) - Filter user/assistant/any
- `min_quality_score` (query_successful_prompts) - Quality threshold

**Ordering within Tier 2**:
- **Single-value filters first** (`tool`, `status`)
- **Multi-value filters** (`where` clauses) - if optional
- **Alphabetical within groups**

**Rationale**:
- Filtering is conceptually "narrowing the search"
- Placed after required (defines query) but before range (refines further)
- Logical flow: What to search (T1) → What to match (T2) → What range (T3) → How much (T4)

---

### Tier 3: Range Parameters

**Purpose**: Parameters that define ranges, thresholds, or bounds

**Characteristics**:
- Optional parameters
- Numeric types (typically)
- Define minimum, maximum, or threshold values
- Affect query boundaries

**Examples**:
- `start_turn`, `end_turn` (query_conversation)
- `min_duration`, `max_duration` (query_conversation)
- `min_tools`, `max_tools` (query_assistant_messages)
- `min_length`, `max_length` (query_assistant_messages)
- `min_tokens_output` (query_assistant_messages)
- `threshold` (query_files)
- `window` (query_context)
- `min_occurrences` (query_tool_sequences)

**Ordering within Tier 3**:
- **Positional ranges first** (`start_turn` before `end_turn`)
- **Duration ranges** (`min_duration` before `max_duration`)
- **Size ranges** (`min_length` before `max_length`)
- **Within groups**: `start/min` before `end/max`
- **Alphabetical** if no natural pairing

**Rationale**:
- Range parameters refine the query further
- Placing after filtering maintains logical flow
- Start/min before end/max is natural reading order
- Industry pattern (SQL BETWEEN, GraphQL ranges)

---

### Tier 4: Output Control Parameters

**Purpose**: Parameters that control how much or what format is returned

**Characteristics**:
- Optional parameters
- Affect output size/format, not query logic
- Typically `limit`, `offset`, pagination controls

**Examples**:
- `limit` (query_tools, query_user_messages, etc.)

**Ordering within Tier 4**:
- `limit` first
- `offset` (if added in future)
- `page_size`, `page_number` (if pagination added)

**Rationale**:
- Output control is conceptually separate from query logic
- Affects result presentation, not what's selected
- Placing last emphasizes "how much" vs. "what"
- Industry standard (SQL LIMIT, GraphQL pagination last)

---

### Tier 5: Standard Parameters

**Purpose**: Parameters common to all query tools (via MergeParameters)

**Included via `MergeParameters()` function**:
- `scope` (string) - "project" or "session"
- `jq_filter` (string) - jq expression
- `stats_only` (boolean) - Return only stats
- `stats_first` (boolean) - Stats first, then details
- `inline_threshold_bytes` (number) - Hybrid mode threshold
- `output_format` (string) - "jsonl" or "tsv"

**Ordering within Tier 5**:
- **Fixed order** (defined in StandardToolParameters function)
- No manual reordering needed (automatic via MergeParameters)

**Rationale**:
- Standard parameters are added programmatically
- Consistent across all tools automatically
- Placing last avoids clutter in tool-specific params

---

## Ordering Examples

### Example 1: query_tools (Current → Corrected)

**Current Order** (inconsistent):
```go
{
  "limit": "number",        // Tier 4 (output control)
  "tool": "string",         // Tier 2 (filtering)
  "status": "string",       // Tier 2 (filtering)
  // + Standard params (Tier 5)
}
```

**Corrected Order** (tier-based):
```go
{
  // Tier 2: Filtering
  "tool": "string",
  "status": "string",

  // Tier 4: Output Control
  "limit": "number",

  // Tier 5: Standard params (auto-merged)
}
```

**Rationale**: Filtering params (tool, status) define what to query. Limit controls output size.

---

### Example 2: query_user_messages (Current → Corrected)

**Current Order**:
```go
{
  "pattern": "string",           // Tier 1 (required)
  "limit": "number",             // Tier 4 (output control)
  "max_message_length": "number", // Tier 3 (range)
  "content_summary": "boolean",  // Tier 4 (output control, deprecated)
  // + Standard params (Tier 5)
}
```

**Corrected Order**:
```go
{
  // Tier 1: Required
  "pattern": "string",

  // Tier 3: Range
  "max_message_length": "number",

  // Tier 4: Output Control
  "limit": "number",
  "content_summary": "boolean",  // (deprecated, but still here)

  // Tier 5: Standard params (auto-merged)
}
```

**Rationale**: Required pattern first, range second, output control last.

---

### Example 3: query_assistant_messages (Current → Corrected)

**Current Order**:
```go
{
  "pattern": "string",         // Tier 2 (filtering, optional)
  "min_tools": "number",       // Tier 3 (range)
  "max_tools": "number",       // Tier 3 (range)
  "min_tokens_output": "number", // Tier 3 (range)
  "min_length": "number",      // Tier 3 (range)
  "max_length": "number",      // Tier 3 (range)
  "limit": "number",           // Tier 4 (output control)
  // + Standard params (Tier 5)
}
```

**Corrected Order**:
```go
{
  // Tier 2: Filtering
  "pattern": "string",

  // Tier 3: Range (grouped by type)
  "min_tools": "number",
  "max_tools": "number",
  "min_tokens_output": "number",
  "min_length": "number",
  "max_length": "number",

  // Tier 4: Output Control
  "limit": "number",

  // Tier 5: Standard params (auto-merged)
}
```

**Rationale**: Pattern filter first, ranges grouped (tools, tokens, length), limit last.

---

### Example 4: query_conversation (Current → Corrected)

**Current Order**:
```go
{
  "start_turn": "number",      // Tier 3 (range)
  "end_turn": "number",        // Tier 3 (range)
  "pattern": "string",         // Tier 2 (filtering, optional)
  "pattern_target": "string",  // Tier 2 (filtering)
  "min_duration": "number",    // Tier 3 (range)
  "max_duration": "number",    // Tier 3 (range)
  "limit": "number",           // Tier 4 (output control)
  // + Standard params (Tier 5)
}
```

**Corrected Order**:
```go
{
  // Tier 2: Filtering
  "pattern": "string",
  "pattern_target": "string",

  // Tier 3: Range (grouped: turn ranges, then duration ranges)
  "start_turn": "number",
  "end_turn": "number",
  "min_duration": "number",
  "max_duration": "number",

  // Tier 4: Output Control
  "limit": "number",

  // Tier 5: Standard params (auto-merged)
}
```

**Rationale**: Filtering first (what to match), turn range, duration range, output control.

---

### Example 5: query_context (Current → Corrected)

**Current Order**:
```go
{
  "error_signature": "string", // Tier 1 (required)
  "window": "number",          // Tier 3 (range)
  // + Standard params (Tier 5)
}
```

**Corrected Order**:
```go
{
  // Tier 1: Required
  "error_signature": "string",

  // Tier 3: Range
  "window": "number",

  // Tier 5: Standard params (auto-merged)
}
```

**Rationale**: Already correct! Required param first, range second.

---

## Decision Tree

### Ordering a New Parameter

```
START
│
├─ Is the parameter REQUIRED?
│  │
│  YES → Tier 1 (Required Parameters)
│  │     Place at beginning
│  │
│  NO
│  ├─ Does it FILTER query results?
│  │  │
│  │  YES → Tier 2 (Filtering Parameters)
│  │  │     Examples: tool, status, pattern (if optional)
│  │  │
│  │  NO
│  │  ├─ Does it define a RANGE or THRESHOLD?
│  │     │
│  │     YES → Tier 3 (Range Parameters)
│  │     │     Examples: min_*, max_*, start_*, end_*, window
│  │     │
│  │     NO
│  │     ├─ Does it control OUTPUT size/format?
│  │        │
│  │        YES → Tier 4 (Output Control)
│  │        │     Examples: limit, offset
│  │        │
│  │        NO → Is it a STANDARD parameter?
│  │           │
│  │           YES → Tier 5 (auto-merged, don't add manually)
│  │           │
│  │           NO → SPECIAL CASE (consult maintainer)
│
END
```

---

## Validation Checklist

When adding or reordering parameters:

- [ ] **Required params in Tier 1** (required: true in schema)
- [ ] **Filtering params in Tier 2** (affects what's returned)
- [ ] **Range params in Tier 3** (min/max/threshold/window)
- [ ] **Output control in Tier 4** (limit, offset, pagination)
- [ ] **Standard params via MergeParameters** (don't add manually)
- [ ] **Within-tier ordering correct** (alphabetical or start/min before end/max)
- [ ] **No gaps in tiers** (Tier 1 → Tier 2 → Tier 3 → Tier 4 → Tier 5)

---

## Backward Compatibility

### Why Parameter Order Doesn't Matter (JSON)

**JSON objects are unordered**:
- Parameter order in schema definition is **documentation only**
- Function calls use **named parameters** (key-value pairs)
- Changing parameter order is **non-breaking**

**Example**:
```json
// Both calls are equivalent
{"tool": "Bash", "status": "error", "limit": 10}
{"limit": 10, "tool": "Bash", "status": "error"}
```

**Implication**: Reordering parameters in tool definitions is **safe**.

### Migration Strategy

**Phase 1: Documentation Update**
- Update parameter ordering in `cmd/mcp-server/tools.go`
- Update documentation in `docs/guides/mcp.md`
- No API changes (backward compatible)

**Phase 2: Example Updates**
- Update code examples to use new ordering
- Update documentation examples
- Communicate preference for new ordering

**Phase 3: Consistency Enforcement**
- Add linter to check parameter ordering
- Pre-commit hook to validate new tools
- CI check for consistency violations

**No User Impact**: Existing code continues to work unchanged.

---

## Rationale

### Why Tier-Based Ordering?

1. **Cognitive Load**: Predictable ordering reduces mental effort
2. **Scannability**: Users can quickly find parameters
3. **Consistency**: Same pattern across all tools
4. **Maintainability**: Clear rules for adding new parameters

### Industry Patterns

**SQL**:
```sql
SELECT * FROM table
WHERE condition     -- Filtering (Tier 2)
  AND value > min   -- Range (Tier 3)
LIMIT 10;           -- Output control (Tier 4)
```

**GraphQL**:
```graphql
query {
  users(
    filter: "active"     # Filtering (Tier 2)
    minAge: 18           # Range (Tier 3)
    maxAge: 65           # Range (Tier 3)
    limit: 10            # Output control (Tier 4)
  )
}
```

**REST Query Params**:
```
/api/users?status=active&minAge=18&maxAge=65&limit=10
           ^Tier 2      ^Tier 3  ^Tier 3  ^Tier 4
```

### Alignment with meta-cc API

- **Standard params last** (Tier 5) aligns with MergeParameters implementation
- **Limit last** (Tier 4) matches CLI argument patterns
- **Required first** (Tier 1) matches schema validation order

---

## Metrics

### Current Consistency

```yaml
parameter_ordering_before:
  standard_params: 100% (via MergeParameters)
  tool_specific_params: ~60% (inconsistent)
  overall_score: 0.80

parameter_ordering_after:
  standard_params: 100% (via MergeParameters)
  tool_specific_params: 100% (tier-based)
  overall_score: 1.00

improvement: +0.20
```

### Expected Impact

**V_consistency Contribution**:
- Parameter ordering: 0.80 → 1.00 (+0.20)
- Overall V_consistency component weight: ~20%
- V_consistency improvement: 0.72 → ~0.76 (+0.04)
- Weighted ΔV: +0.04 × 0.30 = +0.012

**Combined with naming convention**:
- Total V_consistency improvement: 0.72 → 0.80+ (+0.08+)

---

## Future Considerations

### Pagination Parameters

If pagination is added in the future:

**Tier 4: Output Control**
```go
{
  "limit": "number",      // Page size
  "offset": "number",     // Start position
  "page": "number",       // Page number (alternative to offset)
  "cursor": "string",     // Cursor-based pagination
}
```

**Ordering**: `limit`, `offset`, `page`, `cursor` (most common first)

### Sorting Parameters

If sorting is added:

**Tier 4: Output Control** (after limit)
```go
{
  "limit": "number",
  "sort_by": "string",    // Field to sort by
  "sort_order": "string", // "asc" or "desc"
}
```

---

## Reordering Summary

### Tools Requiring Reordering

**Priority 1: High Usage Tools**
1. `query_tools` - Move limit after filtering
2. `query_user_messages` - Move limit after range
3. `query_assistant_messages` - Already compliant ✅

**Priority 2: Moderate Usage Tools**
4. `query_conversation` - Move filtering before range
5. `query_context` - Already compliant ✅
6. `query_tool_sequences` - Verify ordering

**Priority 3: Low Usage Tools**
7-13. Other query_* tools - Apply tier-based ordering

---

## Revision History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-10-15 | Initial convention creation (Iteration 2) |

---

**Convention Status**: ✅ Active
**Next Review**: After reordering implementation (Iteration 3+)
**Related Documents**:
- `api-naming-convention.md` (naming patterns)
- `api-consistency-methodology.md` (validation)
- `cmd/mcp-server/tools.go` (implementation)
