# API Consistency Checking Methodology

**Version**: 1.0
**Created**: 2025-10-15 (Iteration 2, bootstrap-006-api-design)
**Scope**: meta-cc MCP Server API (16 tools)
**Purpose**: Define validation and quality assurance processes for API consistency

---

## Executive Summary

This document establishes a comprehensive methodology for validating and maintaining API consistency across all meta-cc MCP tools. The methodology combines manual review processes, automated checking strategies, and quality gates to ensure consistent naming, parameter ordering, documentation, and behavior.

**Key Components**:
1. **Validation Checklist**: Comprehensive review criteria for all consistency dimensions
2. **Manual Review Process**: Step-by-step guide for human validation
3. **Automated Checking**: Specification for tooling and automation
4. **Quality Gates**: Integration points (pre-commit, CI, documentation review)
5. **Remediation Guide**: How to fix consistency violations

**Target**: V_consistency ≥ 0.85 (sustained through systematic validation)

---

## Consistency Dimensions

### 1. Naming Consistency

**What**: Tool names follow established prefix conventions

**Validation Criteria**:
- Tool name uses one of: `query_*`, `get_*`, `list_*`, `cleanup_*`, `create_*`, `update_*`, `delete_*`, `analyze_*`
- Prefix matches tool function per `api-naming-convention.md` decision tree
- Name uses snake_case (no camelCase or PascalCase)
- Name is concise (≤30 characters preferred)
- Name is descriptive and unambiguous

**Reference**: `data/api-naming-convention.md`

---

### 2. Parameter Ordering Consistency

**What**: Parameters ordered according to tier-based system

**Validation Criteria**:
- Required parameters (Tier 1) appear first
- Filtering parameters (Tier 2) follow required
- Range parameters (Tier 3) follow filtering
- Output control parameters (Tier 4) follow range
- Standard parameters (Tier 5) added via MergeParameters
- Within-tier ordering follows convention (alphabetical or start/min before end/max)

**Reference**: `data/api-parameter-convention.md`

---

### 3. Parameter Naming Consistency

**What**: Parameter names follow naming conventions

**Validation Criteria**:
- All parameters use snake_case
- No camelCase or PascalCase
- Consistent naming for common concepts:
  - `limit` (not `max_results` or `count`)
  - `pattern` (not `regex` or `search`)
  - `min_*` / `max_*` for ranges
  - `start_*` / `end_*` for positional ranges

---

### 4. Description Format Consistency

**What**: Tool descriptions follow template

**Validation Criteria**:
- Format: `"<Action> <object>. Default scope: <project/session/none>."`
- Length ≤100 characters
- Includes "Default scope:" suffix
- Uses active voice and imperative form
- Focuses on "what" not "how"

**Reference**: `cmd/mcp-server/tools.go` (lines 3-13)

---

### 5. Standard Parameter Consistency

**What**: All query tools support standard parameters

**Validation Criteria**:
- Uses `MergeParameters()` function
- Includes all 6 standard parameters:
  - `scope` (string)
  - `jq_filter` (string)
  - `stats_only` (boolean)
  - `stats_first` (boolean)
  - `inline_threshold_bytes` (number)
  - `output_format` (string)

---

### 6. Response Format Consistency

**What**: All query tools use hybrid output mode

**Validation Criteria**:
- Supports inline mode (data ≤8KB)
- Supports file_ref mode (data >8KB)
- Respects `inline_threshold_bytes` parameter
- Returns consistent response structure:
  ```json
  {
    "mode": "inline" | "file_ref",
    "data": [...] | null,
    "file_ref": {...} | null
  }
  ```

---

### 7. Schema Consistency

**What**: Tool schemas are well-formed and consistent

**Validation Criteria**:
- Schema type is "object"
- Properties correctly typed (string, number, boolean)
- Required parameters listed in `required` array
- Descriptions present for all parameters
- No deprecated parameters (unless marked in description)

---

## Validation Checklist

### Pre-Submission Checklist (For New/Modified Tools)

Use this checklist before committing changes to tool definitions:

#### Naming
- [ ] Tool name follows prefix convention (query_*, get_*, list_*, cleanup_*)
- [ ] Prefix matches function per decision tree
- [ ] Name uses snake_case
- [ ] Name is unique (no duplicates)
- [ ] Name is ≤30 characters

#### Parameters
- [ ] Parameter ordering follows tier system (Required → Filtering → Range → Output → Standard)
- [ ] Within-tier ordering is correct (alphabetical or start/min before end/max)
- [ ] All parameters use snake_case
- [ ] Standard parameters added via MergeParameters (not manually)
- [ ] Required parameters listed in `required` array

#### Description
- [ ] Description follows template: `"<Action> <object>. Default scope: <X>."`
- [ ] Description ≤100 characters
- [ ] Description includes "Default scope:" suffix
- [ ] Description uses active voice

#### Schema
- [ ] Schema type is "object"
- [ ] Properties have correct types (string/number/boolean)
- [ ] All parameters have descriptions
- [ ] Required parameters in `required` array

#### Documentation
- [ ] Tool documented in `docs/guides/mcp.md`
- [ ] Examples provided
- [ ] Edge cases documented (if applicable)

#### Testing
- [ ] Tool tested manually (basic functionality)
- [ ] Tool tested with various parameter combinations
- [ ] Error cases tested (missing required params, invalid values)

---

## Manual Review Process

### Step 1: Naming Review

**Reviewer Action**:
1. Read tool name
2. Identify tool function (what does it do?)
3. Determine expected prefix from `api-naming-convention.md` decision tree
4. Verify actual prefix matches expected
5. Check snake_case compliance
6. Check name length (≤30 chars preferred)

**Common Violations**:
- Using `get_*` for filtered queries (should be `query_*`)
- Using non-standard prefix (e.g., `retrieve_*`, `search_*`)
- Using camelCase (`getUserMessages` instead of `get_user_messages`)

**Remediation**:
- Rename tool to follow convention
- If breaking change, follow deprecation process (api-deprecation-policy.md)

---

### Step 2: Parameter Ordering Review

**Reviewer Action**:
1. List all tool-specific parameters (excluding standard params)
2. Categorize each parameter by tier:
   - Tier 1: Required (in `required` array)
   - Tier 2: Filtering (affects what's returned)
   - Tier 3: Range (min_*, max_*, threshold, window)
   - Tier 4: Output control (limit, offset)
3. Verify parameters listed in tier order (1 → 2 → 3 → 4)
4. Verify within-tier ordering (alphabetical or paired)
5. Verify MergeParameters used for standard params

**Common Violations**:
- `limit` placed before filtering params
- Range params scattered among filtering params
- Required params not listed first
- Standard params added manually (should use MergeParameters)

**Remediation**:
- Reorder parameters to match tier system
- No API impact (JSON parameter order doesn't matter)

---

### Step 3: Description Review

**Reviewer Action**:
1. Check description format matches template
2. Verify "Default scope:" suffix present
3. Count characters (≤100)
4. Verify active voice ("Query X" not "Queries X")
5. Verify focus on "what" not "how"

**Common Violations**:
- Missing "Default scope:" suffix
- Passive voice ("Used to query..." instead of "Query...")
- Too long (>100 characters)
- Too detailed (explains implementation)

**Remediation**:
- Rewrite description to match template
- Non-breaking change (documentation only)

---

### Step 4: Schema Review

**Reviewer Action**:
1. Verify schema type is "object"
2. Check all parameters have type declarations
3. Verify required params in `required` array
4. Check parameter descriptions present
5. Verify standard params via MergeParameters

**Common Violations**:
- Missing parameter types
- Required params not in `required` array
- Manual standard parameter definitions

**Remediation**:
- Add missing type declarations
- Update `required` array
- Replace manual standard params with MergeParameters

---

### Step 5: Documentation Review

**Reviewer Action**:
1. Verify tool documented in `docs/guides/mcp.md`
2. Check examples provided
3. Verify scope declaration matches implementation
4. Check parameter descriptions match schema

**Common Violations**:
- Tool not documented
- Examples outdated
- Scope mismatch (doc says "session" but tool defaults to "project")

**Remediation**:
- Add/update documentation in mcp.md
- Update examples to match current API

---

## Automated Checking

### Automation Strategy

**Philosophy**: Automate what can be deterministically validated, leave semantic judgments to humans.

**Automatable Checks** (95% coverage):
1. Naming pattern validation (regex)
2. Parameter ordering validation (tier-based)
3. snake_case enforcement
4. Description format validation
5. Schema structure validation
6. Standard parameter presence

**Human-Required Checks** (5% coverage):
1. Semantic appropriateness (does name match function?)
2. Edge case categorization (stats vs. query)
3. Documentation quality (clarity, completeness)

---

### Tool Specification: `meta-cc validate-api`

**Command**: `meta-cc validate-api [--file tools.go] [--fix]`

**Purpose**: Validate API consistency and optionally auto-fix violations

#### Check 1: Naming Pattern Validation

**Logic**:
```python
def validate_naming(tool_name):
    valid_prefixes = ["query_", "get_", "list_", "cleanup_", "create_", "update_", "delete_", "analyze_"]

    if not any(tool_name.startswith(p) for p in valid_prefixes):
        return ERROR("Tool name must use standard prefix")

    if not is_snake_case(tool_name):
        return ERROR("Tool name must use snake_case")

    if len(tool_name) > 40:
        return WARNING("Tool name >40 characters (prefer ≤30)")

    return OK
```

**Output**:
```
✗ get_session_stats: Should use query_* prefix (returns filtered data)
✓ query_tools: Naming follows convention
✓ list_capabilities: Naming follows convention
```

---

#### Check 2: Parameter Ordering Validation

**Logic**:
```python
def validate_parameter_order(tool_params, required_params):
    tiers = categorize_params(tool_params, required_params)

    expected_order = tiers[1] + tiers[2] + tiers[3] + tiers[4]
    actual_order = tool_params

    if expected_order != actual_order:
        return ERROR("Parameter ordering violates tier system",
                     expected=expected_order,
                     actual=actual_order)

    return OK

def categorize_params(params, required):
    tier1 = [p for p in params if p in required]
    tier2 = [p for p in params if is_filtering_param(p)]
    tier3 = [p for p in params if is_range_param(p)]
    tier4 = [p for p in params if is_output_param(p)]
    return {1: tier1, 2: tier2, 3: tier3, 4: tier4}
```

**Output**:
```
✗ query_tools: Parameter ordering incorrect
  Expected: tool, status, limit
  Actual:   limit, tool, status

✓ query_context: Parameter ordering correct
```

---

#### Check 3: Description Format Validation

**Logic**:
```python
def validate_description(desc):
    if not re.match(r'^[A-Z].*\. Default scope: (project|session|none)\.$', desc):
        return ERROR("Description must match template")

    if len(desc) > 100:
        return WARNING("Description >100 characters")

    if "Default scope:" not in desc:
        return ERROR("Description must include 'Default scope:' suffix")

    return OK
```

**Output**:
```
✓ query_tools: Description follows template
✗ query_files: Description missing "Default scope:" suffix
```

---

#### Check 4: Schema Structure Validation

**Logic**:
```python
def validate_schema(tool):
    if tool.InputSchema.Type != "object":
        return ERROR("Schema type must be 'object'")

    for param, props in tool.InputSchema.Properties.items():
        if not props.Type:
            return ERROR(f"Parameter '{param}' missing type")
        if not props.Description:
            return WARNING(f"Parameter '{param}' missing description")

    for req in tool.InputSchema.Required:
        if req not in tool.InputSchema.Properties:
            return ERROR(f"Required param '{req}' not in properties")

    return OK
```

---

#### Check 5: Standard Parameter Presence

**Logic**:
```python
def validate_standard_params(tool):
    if tool.Category != "query":
        return OK  # Non-query tools don't need standard params

    standard_params = ["scope", "jq_filter", "stats_only", "stats_first",
                       "inline_threshold_bytes", "output_format"]

    missing = [p for p in standard_params if p not in tool.InputSchema.Properties]

    if missing:
        return ERROR(f"Missing standard parameters: {missing}")

    return OK
```

---

### Auto-Fix Capability

**Safe Auto-Fixes**:
1. Reorder parameters (non-breaking)
2. Add standard parameters via MergeParameters
3. Fix snake_case violations (breaking, requires deprecation)
4. Fix description format (non-breaking)

**Unsafe Auto-Fixes** (require human approval):
1. Rename tools (breaking change)
2. Change parameter names (breaking change)
3. Add/remove required parameters

**Example**:
```bash
$ meta-cc validate-api --fix

Analyzing tools.go...

[WARNING] query_tools: Parameter ordering incorrect
  Auto-fix: Reorder parameters (non-breaking)
  Apply? [y/N] y
  ✓ Fixed parameter ordering

[ERROR] get_session_stats: Should use query_* prefix
  Auto-fix: Rename to query_session_stats (BREAKING)
  Apply? [y/N] n
  ✗ Skipped (manual intervention required)

Summary:
  ✓ 1 issue auto-fixed
  ✗ 1 issue requires manual fix
```

---

## Quality Gates

### Gate 1: Pre-Commit Hook

**Trigger**: Before each git commit
**Scope**: Changed tool definitions only
**Action**: Run `meta-cc validate-api --fast`

**Fast Mode Checks**:
- Naming pattern validation
- Parameter ordering validation
- Description format validation

**Behavior**:
- **Pass**: Allow commit
- **Fail**: Block commit, show errors
- **Warning**: Allow commit, show warnings

**Installation**:
```bash
./scripts/install-hooks.sh
```

**Hook Script** (`.git/hooks/pre-commit`):
```bash
#!/bin/bash

# Check if tools.go was modified
if git diff --cached --name-only | grep -q "cmd/mcp-server/tools.go"; then
    echo "Validating API consistency..."
    ./meta-cc validate-api --fast --file cmd/mcp-server/tools.go

    if [ $? -ne 0 ]; then
        echo "❌ API consistency check failed. Fix errors before committing."
        exit 1
    fi
fi

exit 0
```

---

### Gate 2: CI Pipeline Check

**Trigger**: On every pull request
**Scope**: All tool definitions
**Action**: Run `meta-cc validate-api --full`

**Full Mode Checks**:
- All naming, parameter, description, schema validations
- Cross-tool consistency (no duplicate names)
- Documentation sync check (mcp.md reflects tools.go)

**Behavior**:
- **Pass**: Allow merge
- **Fail**: Block merge, require fixes
- **Warning**: Allow merge, create issue

**CI Configuration** (`.github/workflows/api-consistency.yml`):
```yaml
name: API Consistency Check

on: [pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Build meta-cc
        run: make build
      - name: Validate API consistency
        run: ./meta-cc validate-api --full
      - name: Check documentation sync
        run: ./scripts/check-api-docs-sync.sh
```

---

### Gate 3: Documentation Review

**Trigger**: Manual review during PR
**Scope**: Documentation accuracy and completeness
**Action**: Human reviewer checks:

1. **Documentation Completeness**:
   - [ ] Tool documented in mcp.md
   - [ ] Examples provided
   - [ ] Edge cases covered

2. **Semantic Appropriateness**:
   - [ ] Tool name matches function
   - [ ] Parameters logically named
   - [ ] Description accurately describes behavior

3. **Consistency with Existing Tools**:
   - [ ] Similar tools use similar patterns
   - [ ] No duplicate functionality
   - [ ] Integrates with existing workflows

**Reviewer Checklist**: See "Manual Review Process" section above

---

## Integration with API Evolution

### Consistency + Deprecation

**Scenario**: Fixing a consistency violation requires breaking change (e.g., renaming `get_session_stats`)

**Process** (per api-deprecation-policy.md):
1. **Identify violation** (consistency check)
2. **Assess impact** (usage statistics)
3. **Design migration** (old → new mapping)
4. **Deprecate old** (12-month notice period)
5. **Add new** (correct naming)
6. **Support both** (during transition)
7. **Remove old** (after notice period)

**Consistency Methodology Role**:
- Identifies violations early (quality gates)
- Tracks deprecation status (old vs. new)
- Validates new tool follows conventions
- Ensures migration preserves consistency

---

### Consistency + Versioning

**Scenario**: Major version bump allows breaking consistency fixes

**Process** (per api-versioning-strategy.md):
1. **Accumulate violations** (track in backlog)
2. **Plan major version** (e.g., v1.0.0 → v2.0.0)
3. **Batch consistency fixes** (rename tools, reorder params)
4. **Document changes** (migration guide)
5. **Release major version** (with consistency improvements)

**Consistency Methodology Role**:
- Provides list of violations for major version
- Validates all tools in new version follow conventions
- Ensures version bump doesn't introduce new violations

---

## Common Violation Patterns

### Anti-Pattern 1: Inconsistent Prefix Usage

**Violation**: Tool uses non-standard prefix
**Example**: `retrieve_messages` instead of `query_messages`
**Fix**: Rename to standard prefix (`query_messages`)
**Impact**: Breaking change (requires deprecation)

---

### Anti-Pattern 2: Limit First Ordering

**Violation**: `limit` parameter appears before filtering params
**Example**: `{"limit": 10, "tool": "Bash"}` order
**Fix**: Move `limit` to Tier 4 (after filtering)
**Impact**: Non-breaking (JSON parameter order irrelevant)

---

### Anti-Pattern 3: Manual Standard Parameters

**Violation**: Standard params added manually instead of via MergeParameters
**Example**: Defining `scope` in tool-specific properties
**Fix**: Remove manual definition, use MergeParameters
**Impact**: Non-breaking (same result, cleaner code)

---

### Anti-Pattern 4: Incorrect Scope Declaration

**Violation**: Description says "Default scope: session" but tool defaults to "project"
**Example**: Tool description mismatch
**Fix**: Update description to match implementation OR change default scope
**Impact**: Non-breaking (documentation fix)

---

### Anti-Pattern 5: Missing Required Parameter

**Violation**: Parameter semantically required but not in `required` array
**Example**: `error_signature` not marked required in query_context
**Fix**: Add to `required` array
**Impact**: Breaking change (now enforced)

---

## Metrics & Reporting

### Consistency Dashboard

**Tracked Metrics**:
1. **Naming Consistency**: % of tools following naming convention
2. **Parameter Ordering**: % of tools following tier system
3. **Description Format**: % of tools with compliant descriptions
4. **Schema Consistency**: % of tools with valid schemas
5. **Overall Consistency**: Weighted average of above

**Dashboard Example**:
```
API Consistency Dashboard (2025-10-15)
========================================

Naming Consistency:        93% (13/14 tools) ⚠️
  - Violations: get_session_stats

Parameter Ordering:        60% (8/13 query tools) ❌
  - Violations: query_tools, query_user_messages, query_conversation, ...

Description Format:       100% (16/16 tools) ✅

Schema Consistency:       100% (16/16 tools) ✅

Overall Consistency:       88% (Target: 95%)

Top Issues:
1. get_session_stats naming violation (breaking fix required)
2. Parameter ordering inconsistency (5 tools affected, non-breaking fix)
```

---

### Trend Tracking

**Tracked Over Time**:
- Consistency score per release
- Violations introduced vs. fixed
- Time to fix violations
- Breaking vs. non-breaking violations

**Example Report**:
```
Consistency Trend (Last 6 Months)
==================================

Oct 2025: 88% (current)
Sep 2025: 85%
Aug 2025: 82%
Jul 2025: 80%
Jun 2025: 78%
May 2025: 75%

Trend: +13% improvement (May → Oct)

Violations Fixed:
  - 3 naming violations
  - 8 parameter ordering violations
  - 2 description format violations

Violations Introduced:
  - 1 new tool with incorrect ordering (later fixed)
```

---

## Remediation Guide

### Fixing Naming Violations

**Step 1: Identify Violation**
```bash
$ meta-cc validate-api
✗ get_session_stats: Should use query_* prefix
```

**Step 2: Determine Impact**
```bash
$ meta-cc query-tools tool=get_session_stats
# Shows usage count across all sessions
```

**Step 3: Design Migration**
- New name: `query_session_stats`
- Old name: `get_session_stats` (deprecated)
- Support both during transition (12 months)

**Step 4: Implement**
1. Add new tool (`query_session_stats`)
2. Mark old tool deprecated (warning in response)
3. Update documentation (migration guide)
4. Announce deprecation (changelog)

**Step 5: Monitor & Remove**
- Track usage of old name
- After 12 months, remove old tool
- Update tests and docs

---

### Fixing Parameter Ordering Violations

**Step 1: Identify Violation**
```bash
$ meta-cc validate-api
✗ query_tools: Parameter ordering incorrect
  Expected: tool, status, limit
  Actual:   limit, tool, status
```

**Step 2: Apply Fix** (non-breaking)
```bash
$ meta-cc validate-api --fix
✓ Reordered parameters in query_tools
```

**Step 3: Verify**
```bash
$ meta-cc validate-api
✓ All tools pass parameter ordering check
```

**Step 4: Update Documentation**
- Update examples in mcp.md to show preferred order
- No user-facing changes (JSON parameter order doesn't matter)

---

## Revision History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-10-15 | Initial methodology creation (Iteration 2) |

---

**Methodology Status**: ✅ Active
**Next Review**: After validation tool implementation (Iteration 3+)
**Related Documents**:
- `api-naming-convention.md` (naming rules)
- `api-parameter-convention.md` (parameter ordering rules)
- `api-deprecation-policy.md` (breaking change process)
- `api-versioning-strategy.md` (version management)
