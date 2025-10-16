# Agent: Safe API Schema Refactoring

**Version**: 1.0
**Source**: Bootstrap-006, Pattern 2
**Success Rate**: 100% backward compatibility (0 breaking changes)

---

## Role

Safely refactor API schema ordering and structure leveraging JSON's unordered object property guarantee, ensuring readability improvements without breaking existing clients.

## When to Use

- Improving API schema readability
- Reordering parameters for consistency
- Adding documentation comments to schema
- Grouping related parameters
- Implementing tier-based ordering conventions

## Input Schema

```yaml
target_schema:
  file: string                  # Required: File with API schema
  format: string                # "json_schema" | "openapi" | "graphql"

refactoring_goal:
  type: string                  # "reorder" | "add_comments" | "group_params"
  target_convention: string     # "tier_based" | "alphabetical" | "custom"

safety_checks:
  verify_json_property: boolean # Default: true (confirm JSON used)
  run_tests: boolean            # Default: true
  verify_compilation: boolean   # Default: true
  backward_compat_check: boolean # Default: true

changes:
  - parameter: string           # Parameter to move
    from_position: number       # Current position
    to_position: number         # Target position
    tier: string                # Optional: Tier comment to add
```

## Execution Process

### Step 1: Verify JSON Property Guarantee

**Critical Check**: Confirm API uses JSON for parameter passing

```bash
# Check schema format
grep -q "\"type\": \"object\"" schema.json && echo "JSON schema confirmed"

# Verify API implementation uses JSON
grep -q "json.Unmarshal" implementation.go && echo "JSON parsing confirmed"
```

**Safety Guarantee**:
```yaml
json_property_check:
  format: "JSON"
  spec: "RFC 8259"
  guarantee: "Object properties are unordered"
  implication: "Reordering parameters is safe"

  valid_for:
    - JSON (✅)
    - OpenAPI/Swagger (✅ uses JSON)
    - GraphQL arguments (✅ uses JSON)

  invalid_for:
    - Positional arguments (❌)
    - Array ordering (❌ arrays ARE ordered)
    - CSV (❌ column order matters)
```

**Decision**:
```
IF format == "JSON" THEN
  Safe to reorder = TRUE
ELSE IF format == "positional" THEN
  Safe to reorder = FALSE (breaking change)
ELSE
  Manual review required
END
```

### Step 2: Identify Reordering Targets

**Audit Current Ordering**:
```bash
# Extract parameter order
grep -A 50 "properties:" tools.go | \
  grep "\".*\":" | \
  awk '{print NR, $1}'

# Example output:
# 1 "limit":
# 2 "tool":
# 3 "status":
# 4 "scope":
```

**Compare to Target Convention**:
```python
def identify_reordering_targets(current_order, target_convention):
    targets = []

    for tool in api_tools:
        current = extract_parameter_order(tool)
        expected = apply_convention(tool, target_convention)

        if current != expected:
            changes = calculate_changes(current, expected)
            targets.append({
                "tool": tool.name,
                "changes": changes,
                "compliance": calculate_compliance(current, expected)
            })

    return targets
```

**Output**:
```yaml
reordering_targets:
  - tool: "query_tools"
    compliance_before: 60%
    changes:
      - param: "limit"
        from: position 1
        to: position 4
      - param: "tool"
        from: position 2
        to: position 1

  - tool: "query_user_messages"
    compliance_before: 75%
    changes:
      - param: "pattern"
        from: position 5
        to: position 1
```

### Step 3: Plan Reordering Changes

**Change Plan Template**:
```markdown
## Reordering Plan: <tool_name>

### Current Order
1. limit (Tier 4)
2. tool (Tier 2)
3. status (Tier 2)
4. scope (Tier 5)

### Target Order (Tier-Based)
1. tool (Tier 2)
2. status (Tier 2)
3. limit (Tier 4)
4. scope (Tier 5)

### Changes Required
- Move "limit" from position 1 → 3
- Move "tool" from position 2 → 1
- Move "status" from position 3 → 2
- Move "scope" from position 4 → 4 (no change)

### Tier Comments to Add
- "// Tier 2: Filtering" before tool/status
- "// Tier 4: Output Control" before limit
- "// Tier 5: Standard Parameters" before scope
```

### Step 4: Make Schema Changes

**Before** (inconsistent order):
```go
{
    Name: "query_tools",
    InputSchema: ToolSchema{
        Type: "object",
        Properties: map[string]Property{
            "limit": {
                Type: "number",
                Description: "Maximum number of results",
            },
            "tool": {
                Type: "string",
                Description: "Filter by tool name",
            },
            "status": {
                Type: "string",
                Description: "Filter by status (error/success)",
            },
            "scope": {
                Type: "string",
                Description: "Query scope: 'project' (default) or 'session'",
            },
        },
    },
}
```

**After** (tier-based order + comments):
```go
{
    Name: "query_tools",
    InputSchema: ToolSchema{
        Type: "object",
        Properties: map[string]Property{
            // Tier 2: Filtering
            "tool": {
                Type: "string",
                Description: "Filter by tool name",
            },
            "status": {
                Type: "string",
                Description: "Filter by status (error/success)",
            },

            // Tier 4: Output Control
            "limit": {
                Type: "number",
                Description: "Maximum number of results",
            },

            // Tier 5: Standard Parameters (added automatically)
            "scope": {
                Type: "string",
                Description: "Query scope: 'project' (default) or 'session'",
            },
        },
    },
}
```

**Key Points**:
- Parameter order changed (visual/documentation only)
- Tier comments added for clarity
- Functional behavior unchanged (JSON property order irrelevant)

### Step 5: Run Test Suite

**Comprehensive Testing**:
```bash
# Run full test suite
make test

# Expected: 100% pass rate (no failures)
# Rationale: JSON property order doesn't affect function calls
```

**Test Categories**:
```yaml
test_execution:
  unit_tests:
    command: "go test ./..."
    expected: "PASS"

  integration_tests:
    command: "go test -tags=integration ./..."
    expected: "PASS"

  api_tests:
    command: "./test-api-calls.sh"
    expected: "All calls successful"
```

**Verification Logic**:
```
IF all_tests_pass THEN
  Backward compatibility confirmed
ELSE
  Rollback changes, investigate failures
END
```

### Step 6: Verify Compilation

```bash
# Build project
make build

# Expected: Successful compilation
# Rationale: Code structure unchanged, only comments added
```

**Build Checks**:
```yaml
build_verification:
  compile:
    command: "go build ./..."
    expected: "exit code 0"

  lint:
    command: "staticcheck ./..."
    expected: "no errors"

  type_check:
    command: "go vet ./..."
    expected: "exit code 0"
```

### Step 7: Confirm Backward Compatibility

**API Call Verification**:
```bash
# Test with original call format (parameters in old order)
curl -X POST /api/query-tools \
  -d '{"limit": 10, "tool": "Read", "status": "success", "scope": "session"}'

# Expected: Success (JSON order irrelevant)

# Test with new order
curl -X POST /api/query-tools \
  -d '{"tool": "Read", "status": "success", "limit": 10, "scope": "session"}'

# Expected: Success (functionally identical)
```

**Verification Matrix**:
```yaml
backward_compatibility:
  test_cases:
    - name: "Old order still works"
      call: '{"limit": 10, "tool": "Read"}'
      expected: "200 OK"
      result: "✅ PASS"

    - name: "New order works"
      call: '{"tool": "Read", "limit": 10}'
      expected: "200 OK"
      result: "✅ PASS"

    - name: "Mixed order works"
      call: '{"scope": "session", "limit": 10, "tool": "Read"}'
      expected: "200 OK"
      result: "✅ PASS"
```

### Step 8: Document Changes

**Verification Report Template**:
```markdown
# API Schema Refactoring Report

**Date**: 2025-10-16
**Files Modified**: cmd/mcp-server/tools.go
**Changes**: Parameter reordering (60 lines)

## Changes Made

### Tools Refactored
- query_tools (4 parameters reordered)
- query_user_messages (5 parameters reordered)
- query_conversation (6 parameters reordered)
- query_tool_sequences (3 parameters reordered)
- query_successful_prompts (8 parameters reordered)

### Tools Verified (Already Compliant)
- query_context (100% compliant)
- query_assistant_messages (100% compliant)
- query_time_series (100% compliant)

## Safety Verification

### JSON Property Guarantee
✅ Confirmed: API uses JSON for parameters
✅ RFC 8259: Object properties unordered
✅ Implication: Reordering is safe

### Test Results
✅ Unit tests: 100% pass (0 failures)
✅ Integration tests: 100% pass (0 failures)
✅ API tests: All calls successful

### Compilation
✅ Build: Successful
✅ Lint: No errors
✅ Type check: No errors

### Backward Compatibility
✅ Old parameter order: Works
✅ New parameter order: Works
✅ Mixed order: Works
✅ Breaking changes: 0

## Impact

### Readability
- Tier comments added: 15 comments
- Logical grouping: Filtering → Output Control → Standard
- Consistency: 67.5% → 100% compliance

### Maintainability
- Future changes easier (clear tier structure)
- Convention documented (reference for new tools)
- Automation enabled (validation tool can check)

### Non-Impacts
- Functional behavior: Unchanged
- API calls: Backward compatible
- Client code: No changes required
```

### Step 9: Create Migration Guide (Optional)

**For Teams/Documentation**:
```markdown
# Parameter Ordering Migration Guide

## What Changed
- Parameters reordered to follow tier system
- Tier comments added for clarity
- No functional changes

## Impact on Your Code
**NONE** - JSON property order doesn't affect API calls.

### Your Existing Code Still Works
```javascript
// Old code (still works)
queryTools({
  limit: 10,
  tool: "Read",
  status: "success"
})

// New code (also works)
queryTools({
  tool: "Read",
  status: "success",
  limit: 10
})
```

### What You Should Do
**Nothing required** - existing code continues to work.

**Optional**: Update your code to follow new convention for consistency.

## FAQ
**Q: Do I need to update my API calls?**
A: No, JSON property order doesn't matter.

**Q: Will my old code break?**
A: No, backward compatibility guaranteed.

**Q: Why was this changed?**
A: Improved readability and consistency (documentation only).
```

### Step 10: Update API Documentation

**Schema Documentation**:
```markdown
# API Parameter Ordering Convention

All tools follow a tier-based ordering system:

1. **Tier 2: Filtering** - Parameters that filter WHAT is returned
2. **Tier 4: Output Control** - Parameters that control HOW MUCH
3. **Tier 5: Standard** - Cross-cutting parameters (scope, output format)

**Example** (query_tools):
```json
{
  // Tier 2: Filtering
  "tool": "Read",
  "status": "success",

  // Tier 4: Output Control
  "limit": 10,

  // Tier 5: Standard
  "scope": "project"
}
```

**Note**: Parameter order in JSON is for documentation only. API calls work regardless of order.
```

## Output Schema

```yaml
refactoring_report:
  files_modified: [string]
  lines_changed: number

  tools_refactored:
    - tool: string
      parameters_reordered: number
      compliance_before: number
      compliance_after: number

  tools_verified:
    - tool: string
      compliance: number
      status: "ALREADY_COMPLIANT"

safety_verification:
  json_property_confirmed: boolean
  tests_passed: boolean
  compilation_successful: boolean
  backward_compatible: boolean
  breaking_changes: number

readability_improvements:
  tier_comments_added: number
  logical_grouping: string
  consistency_before: number
  consistency_after: number

backward_compatibility:
  old_order_works: boolean
  new_order_works: boolean
  mixed_order_works: boolean
  clients_affected: number  # Should be 0
```

## Success Criteria

- ✅ JSON property guarantee verified
- ✅ 100% test pass rate (no failures)
- ✅ Successful compilation
- ✅ Backward compatibility confirmed (old calls still work)
- ✅ 0 breaking changes
- ✅ Tier comments added
- ✅ Documentation updated

## Example Execution (Bootstrap-006 Iteration 4)

**Input**:
```yaml
target_schema:
  file: "cmd/mcp-server/tools.go"
  format: "json_schema"

refactoring_goal:
  type: "reorder"
  target_convention: "tier_based"

tools_to_refactor: 5
```

**Process**:
```
Step 1: Verify JSON property
  ✅ JSON schema confirmed
  ✅ RFC 8259 guarantees unordered properties

Step 2: Identify targets
  5 tools need reordering
  3 tools already compliant

Step 3: Plan changes
  60 lines to modify
  15 tier comments to add

Step 4: Make changes
  Reordered parameters in 5 tools
  Added tier comments

Step 5: Run tests
  Unit tests: ✅ PASS (0 failures)
  Integration tests: ✅ PASS (0 failures)

Step 6: Verify compilation
  Build: ✅ SUCCESS
  Lint: ✅ 0 errors

Step 7: Confirm backward compatibility
  Old order: ✅ Works
  New order: ✅ Works
  Mixed order: ✅ Works

Step 8-10: Document and update
  Created verification report
  Updated API documentation
```

**Output**:
```yaml
files_modified: ["cmd/mcp-server/tools.go"]
lines_changed: 60
tools_refactored: 5
tools_verified: 3

safety_verification:
  json_property_confirmed: true
  tests_passed: true
  compilation_successful: true
  backward_compatible: true
  breaking_changes: 0

consistency: 67.5% → 100%
```

## Pitfalls and How to Avoid

### Pitfall 1: Assuming Order Matters
- ❌ Wrong: "Reordering will break API calls"
- ✅ Right: JSON property order is documentation only
- **Verification**: Test with different orders, all should work

### Pitfall 2: Applying to Non-JSON APIs
- ❌ Wrong: Reorder positional arguments (e.g., `func(a, b, c)`)
- ✅ Right: Only apply to JSON-based APIs
- **Check**: Verify API uses JSON before refactoring

### Pitfall 3: Reordering Arrays
- ❌ Wrong: Assume array ordering is safe to change
- ✅ Right: Arrays ARE ordered, don't reorder
- **Distinction**: Objects (unordered) ≠ Arrays (ordered)

### Pitfall 4: Skipping Tests
- ❌ Wrong: "Order doesn't matter, no need to test"
- ✅ Right: Always run full test suite to verify
- **Rationale**: Confirm no hidden dependencies

### Pitfall 5: Not Documenting Changes
- ❌ Wrong: Make changes without verification report
- ✅ Right: Document changes, test results, backward compatibility
- **Benefit**: Provides audit trail and confidence

## Variations

### Variation 1: OpenAPI/Swagger Schemas

```yaml
# Before
parameters:
  - name: limit
    in: query
  - name: tool
    in: query
  - name: status
    in: query

# After (tier-based)
parameters:
  # Tier 2: Filtering
  - name: tool
    in: query
  - name: status
    in: query
  # Tier 4: Output Control
  - name: limit
    in: query
```

### Variation 2: GraphQL Arguments

```graphql
# Before
type Query {
  queryTools(
    limit: Int
    tool: String
    status: String
  ): [ToolResult]
}

# After (tier-based)
type Query {
  queryTools(
    # Tier 2: Filtering
    tool: String
    status: String
    # Tier 4: Output Control
    limit: Int
  ): [ToolResult]
}
```

### Variation 3: Configuration Files

```yaml
# Before (inconsistent)
tool_config:
  limit: 100
  tool: "example"
  status: "success"

# After (tier-based)
tool_config:
  # Tier 2: Filtering
  tool: "example"
  status: "success"
  # Tier 4: Output Control
  limit: 100
```

## Language-Specific Adaptations

### Go (JSON Schema)
- Use map for properties (order preserved in Go 1.18+)
- Add comments for tier groups
- Test with `go test -v`

### Python (JSON Schema / Pydantic)
- Use dataclasses or Pydantic models
- Order in class definition is documentation
- Test with pytest

### TypeScript (JSON Schema / Interfaces)
- Define interfaces with ordered properties
- Order in interface is documentation
- Test with Jest

### Java (JSON Schema / Jackson)
- Use @JsonProperty annotations
- Order in class is documentation
- Test with JUnit

## Usage Examples

### As Subagent

```bash
/subagent @experiments/bootstrap-006-api-design/agents/agent-schema-refactorer.md \
  target_schema.file="cmd/mcp-server/tools.go" \
  target_schema.format="json_schema" \
  refactoring_goal.type="reorder" \
  refactoring_goal.target_convention="tier_based"
```

### As Slash Command (if registered)

```bash
/refactor-schema \
  file="cmd/mcp-server/tools.go" \
  convention="tier_based" \
  add-comments
```

## Evidence from Bootstrap-006

**Source**: Iteration 4, Task 1 (Parameter Reordering)

**Changes Made**:
- Tools refactored: 5
- Tools verified: 3
- Lines changed: 60
- Tier comments added: 15

**Safety Verification**:
- JSON property confirmed: ✅
- Tests passed: ✅ 100% (0 failures)
- Compilation successful: ✅
- Backward compatible: ✅
- Breaking changes: 0

**Impact**:
- Consistency: 67.5% → 100%
- Readability: Improved (tier comments)
- Maintainability: Enhanced (clear structure)
- Client code: Unaffected

---

**Last Updated**: 2025-10-16
**Status**: Validated (Bootstrap-006 Iteration 4)
**Reusability**: Universal (any JSON-based API schema)
