# API Naming Convention Guideline

**Version**: 1.0
**Created**: 2025-10-15 (Iteration 2, bootstrap-006-api-design)
**Scope**: meta-cc MCP Server API (16 tools)
**Purpose**: Establish consistent naming patterns for API tools

---

## Executive Summary

This document defines naming conventions for meta-cc MCP API tools to ensure consistency, predictability, and maintainability. The convention establishes four primary prefix categories (`query_*`, `get_*`, `list_*`, `cleanup_*`) with clear decision criteria for choosing the appropriate prefix for new tools.

**Key Principles**:
1. **Predictability**: Users should be able to guess tool names based on function
2. **Consistency**: Similar tools use similar naming patterns
3. **Scalability**: Convention supports future growth
4. **Clarity**: Names clearly communicate tool purpose

**Current State**: 13/14 data-retrieval tools follow `query_*` pattern (93% consistency)
**Target State**: 100% consistency through standardization and deprecation

---

## Prefix Categories

### 1. `query_*` Prefix

**Purpose**: Data retrieval with filtering, analysis, or complex search

**When to Use**:
- Tool retrieves session or project data
- Supports filtering parameters (pattern, status, tool, where, etc.)
- Returns multiple records or aggregated results
- Requires analysis or pattern matching
- Operates on structured data (tool calls, messages, files)

**Examples**:
- ✅ `query_tools` - Retrieves tool calls with filters
- ✅ `query_user_messages` - Searches messages with regex
- ✅ `query_file_access` - Retrieves file operation history
- ✅ `query_session_stats` - Retrieves session statistics with analysis

**Rationale**:
- "Query" implies search/filter capability (industry standard: SQL, GraphQL)
- Distinguishes from simple retrieval (`get_*`)
- Aligns with MCP's data-centric nature
- Scalable pattern (can add query_X for any data type)

**Current Usage**: 13/16 tools (81% of API)

---

### 2. `get_*` Prefix

**Purpose**: Simple, direct retrieval of a single entity or metadata

**When to Use**:
- Tool retrieves a single, specific entity
- Minimal or no filtering required
- Direct access by identifier or key
- Returns metadata or configuration
- Utility function for retrieving capability content

**Examples**:
- ✅ `get_capability` - Retrieves capability by name (key-based lookup)
- ❌ `get_session_stats` - Should be `query_session_stats` (returns aggregated data, supports filtering)

**Rationale**:
- "Get" implies simple, direct retrieval (REST pattern: GET /resource/{id})
- Reserved for utilities and metadata access
- Clear semantic: "get X" = retrieve X by identifier

**Current Usage**: 2/16 tools (12% of API)
**Corrected Usage**: 1/16 tools (6% of API, after fixing `get_session_stats`)

---

### 3. `list_*` Prefix

**Purpose**: Enumerate a collection of items (typically metadata or catalog)

**When to Use**:
- Tool returns a list of available items
- No filtering or complex search
- Purpose is discovery or enumeration
- Returns catalog, index, or directory
- Typically used for metadata navigation

**Examples**:
- ✅ `list_capabilities` - Returns index of available capabilities

**Rationale**:
- "List" implies complete enumeration (REST pattern: GET /resources)
- Distinguishes from filtered query (query_*) and direct retrieval (get_*)
- Clear semantic: "list X" = show all available X

**Current Usage**: 1/16 tools (6% of API)

---

### 4. `cleanup_*` Prefix

**Purpose**: Maintenance operations (deletion, garbage collection, optimization)

**When to Use**:
- Tool performs cleanup or maintenance
- Modifies system state (destructive operation)
- Purpose is housekeeping, not data retrieval
- Typically used for temporary file management

**Examples**:
- ✅ `cleanup_temp_files` - Removes old temporary files

**Rationale**:
- "Cleanup" clearly signals destructive operation
- Distinguishes from retrieval operations (query_*, get_*, list_*)
- Semantic warning: "cleanup X" may delete data

**Current Usage**: 1/16 tools (6% of API)

---

### 5. Reserved Prefixes (Future Use)

**Not Currently Used** - Reserved for future expansion:

#### `create_*` Prefix
**Purpose**: Create new entities
**Example**: `create_snapshot`, `create_export`

#### `update_*` Prefix
**Purpose**: Modify existing entities
**Example**: `update_session_metadata`

#### `delete_*` Prefix
**Purpose**: Remove specific entities
**Example**: `delete_session` (use `cleanup_*` for bulk removal)

#### `analyze_*` Prefix
**Purpose**: Deep analysis or computation (distinct from simple query)
**Example**: `analyze_error_patterns`, `analyze_workflow_efficiency`

---

## Decision Tree

### Naming a New Tool

```
START
│
├─ Does the tool RETRIEVE data?
│  │
│  YES
│  ├─ Does it support FILTERING or SEARCH?
│  │  │
│  │  YES → Use `query_*`
│  │  │    Examples: query_errors, query_prompts
│  │  │
│  │  NO → Does it retrieve by IDENTIFIER?
│  │     │
│  │     YES → Use `get_*`
│  │     │    Examples: get_capability, get_session_metadata
│  │     │
│  │     NO → Does it ENUMERATE a catalog?
│  │        │
│  │        YES → Use `list_*`
│  │             Examples: list_capabilities, list_sessions
│  │
│  NO
│  ├─ Does the tool MODIFY or DELETE data?
│  │  │
│  │  YES → Is it MAINTENANCE/CLEANUP?
│  │     │
│  │     YES → Use `cleanup_*`
│  │     │    Examples: cleanup_temp_files, cleanup_old_sessions
│  │     │
│  │     NO → Use `create_*`, `update_*`, or `delete_*`
│  │          Examples: create_export, delete_session
│  │
│  NO → SPECIAL CASE (requires judgment)
│       Examples: `analyze_*` for complex computation
│
END
```

---

## Edge Cases

### Case 1: Stats vs. Query

**Question**: Is "get session statistics" a query or a get?

**Analysis**:
- Statistics are **aggregated data** (not a single entity)
- Supports **filtering** via standard parameters (scope, jq_filter)
- Returns **multiple metrics** or dimensions
- **Conclusion**: Use `query_session_stats`

**Current Violation**: `get_session_stats` should be `query_session_stats`

### Case 2: Utility vs. Query

**Question**: Is "get capability content" a query or a get?

**Analysis**:
- Retrieves **single entity** by name (identifier)
- **No filtering** or complex search
- Direct lookup operation
- **Conclusion**: Use `get_capability` ✅ (correctly named)

### Case 3: List vs. Query

**Question**: When to use `list_*` vs. `query_*`?

**Guideline**:
- Use `list_*` when purpose is **discovery/enumeration** (show all available X)
- Use `query_*` when purpose is **filtered retrieval** (find X matching criteria)

**Examples**:
- ✅ `list_capabilities` - Enumerate all capabilities (discovery)
- ✅ `query_tools` - Find tools matching filters (search)

### Case 4: Cleanup vs. Delete

**Question**: When to use `cleanup_*` vs. `delete_*`?

**Guideline**:
- Use `cleanup_*` for **bulk/automatic maintenance** (remove old/temporary items)
- Use `delete_*` for **targeted removal** (remove specific item by ID)

**Examples**:
- ✅ `cleanup_temp_files` - Remove files older than N days (bulk)
- ✅ `delete_session` - Remove specific session by ID (targeted)

---

## Examples

### Good Names

1. **`query_tools`**
   ✅ Retrieves tool calls with filtering
   ✅ Clearly indicates search capability
   ✅ Follows `query_*` pattern

2. **`query_user_messages`**
   ✅ Searches messages with regex
   ✅ Complex filtering (pattern matching)
   ✅ Returns multiple records

3. **`get_capability`**
   ✅ Direct retrieval by name
   ✅ No filtering needed
   ✅ Single entity lookup

4. **`list_capabilities`**
   ✅ Enumerates all available items
   ✅ Discovery/catalog purpose
   ✅ No filtering

5. **`cleanup_temp_files`**
   ✅ Maintenance operation
   ✅ Bulk deletion
   ✅ Clear destructive intent

### Bad Names (Anti-Patterns)

1. **❌ `get_session_stats`**
   **Issue**: Stats are aggregated data, not single entity
   **Fix**: `query_session_stats` (supports filtering, returns metrics)

2. **❌ `retrieve_tools`**
   **Issue**: Non-standard verb (use `query`, `get`, or `list`)
   **Fix**: `query_tools` (filtering capability)

3. **❌ `search_messages`**
   **Issue**: Inconsistent with `query_*` pattern
   **Fix**: `query_messages` or `query_user_messages`

4. **❌ `tools_query`**
   **Issue**: Noun-verb order (should be verb-noun)
   **Fix**: `query_tools`

5. **❌ `delete_old_files`**
   **Issue**: Should use `cleanup_*` for bulk operations
   **Fix**: `cleanup_temp_files` (bulk maintenance)

---

## Handling Current Outlier

### Issue: `get_session_stats`

**Problem**: Breaks naming convention (only non-`query_*` data-retrieval tool)

**Analysis**:
- **Current name**: `get_session_stats`
- **Function**: Retrieves session statistics with filtering
- **Parameters**: Supports `scope`, `jq_filter`, `stats_only` (filtering)
- **Returns**: Aggregated metrics (multiple data points)
- **Conclusion**: Should be `query_session_stats`

**Impact**:
- **Users affected**: All users of `get_session_stats`
- **Breaking change**: YES (tool name change)
- **Severity**: 🟡 MODERATE (functional equivalent exists)

### Deprecation Strategy

Per `api-deprecation-policy.md`, breaking changes require:
1. **12-month deprecation period**
2. **Deprecation warnings** (documentation + runtime)
3. **Migration guide**
4. **Dual support** (both names functional during transition)

**Proposed Timeline**:

**Phase 1: Announce (Month 0)**
- Add `query_session_stats` as new tool (identical to `get_session_stats`)
- Mark `get_session_stats` as deprecated in documentation
- Add deprecation warning to `get_session_stats` response
- Publish migration guide

**Phase 2: Migrate (Months 1-12)**
- Encourage users to switch to `query_session_stats`
- Both tools remain functional
- Track usage (how many users still use old name)
- Provide migration support

**Phase 3: Remove (Month 13+)**
- Remove `get_session_stats` from API
- Return error: "Tool deprecated. Use query_session_stats instead."
- Document removal in changelog

### Migration Guide

**For Users**:

```markdown
## Migrating from get_session_stats to query_session_stats

### What Changed?
- Old name: `get_session_stats`
- New name: `query_session_stats`
- Function: Identical (no parameter or behavior changes)

### Migration Steps

1. **Find Usage**:
   ```bash
   grep -r "get_session_stats" .claude/
   ```

2. **Replace**:
   ```bash
   sed -i 's/get_session_stats/query_session_stats/g' .claude/**/*.md
   ```

3. **Test**:
   Verify `query_session_stats` works as expected

### Why the Change?
To improve API consistency:
- 13/14 data-retrieval tools use `query_*` prefix
- `get_*` reserved for simple identifier-based retrieval
- `query_*` better reflects filtering capability

### Support
Deprecation period: 12 months (until 2026-10-15)
Both names will work during transition period.
```

---

## Validation Checklist

When adding a new tool, verify:

- [ ] **Prefix follows convention** (query_*, get_*, list_*, cleanup_*)
- [ ] **Decision tree consulted** (correct prefix chosen)
- [ ] **Name uses verb-noun order** (e.g., query_tools, not tools_query)
- [ ] **Name uses snake_case** (no camelCase)
- [ ] **Name is concise** (≤30 characters preferred)
- [ ] **Name is descriptive** (purpose clear from name)
- [ ] **Name doesn't duplicate existing tool** (uniqueness check)
- [ ] **Edge cases considered** (stats vs. query, list vs. query)

---

## Industry Alignment

### REST API Patterns

| HTTP Method | Action | meta-cc Prefix Equivalent |
|-------------|--------|---------------------------|
| GET /resources | List all | `query_*` or `list_*` |
| GET /resources/{id} | Get one | `get_*` |
| POST /resources | Create | `create_*` (future) |
| PUT /resources/{id} | Update | `update_*` (future) |
| DELETE /resources/{id} | Delete | `delete_*` or `cleanup_*` |

### GraphQL Patterns

| GraphQL | meta-cc Equivalent |
|---------|-------------------|
| query { users(filter: ...) } | `query_users` |
| query { user(id: ...) } | `get_user` |
| mutation { createUser } | `create_user` |

### MCP Best Practices

- **Tool names are actions** (verbs, not nouns)
- **Consistent prefixes** aid tool discovery
- **Clear semantics** reduce user confusion

---

## Metrics

### Current Consistency

```yaml
naming_consistency_before:
  query_pattern: 13/14 data-retrieval tools (93%)
  outliers: 1 (get_session_stats)
  score: 0.85

naming_consistency_after:
  query_pattern: 14/14 data-retrieval tools (100%)
  outliers: 0
  score: 0.93

improvement: +0.08
```

### Expected Impact

**V_consistency Contribution**:
- Naming consistency: 0.85 → 0.93 (+0.08)
- Overall V_consistency: 0.72 → ~0.76 (+0.04)
- Weighted ΔV: +0.04 × 0.30 = +0.012

**Combined with other consistency improvements** (parameter ordering, methodology):
- Target V_consistency: 0.85
- Total improvement: +0.13

---

## Revision History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-10-15 | Initial guideline creation (Iteration 2) |

---

**Guideline Status**: ✅ Active
**Next Review**: After get_session_stats deprecation (2026-10-15)
**Related Documents**:
- `api-parameter-convention.md` (parameter ordering)
- `api-deprecation-policy.md` (deprecation process)
- `api-consistency-methodology.md` (validation)
