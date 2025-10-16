# Agent: Builder Extraction

**Version**: 1.0
**Source**: Bootstrap-004, Pattern 2
**Success Rate**: 75-line reduction (18.9%) in meta-cc Iteration 2

---

## Role

Extract helper functions from repetitive structure definitions (API schemas, forms, configs), reducing duplication while preserving behavioral equivalence.

## When to Use

- Files with repetitive structure definitions
- Duplication ≥15% of file size
- API schemas, configuration objects, form definitions
- Copy-paste patterns across multiple definitions

## Input Schema

```yaml
target_file:
  path: string              # Required: File with duplication
  type: string              # "api_schema" | "form_definition" | "config" | "other"

analysis:
  duplication_threshold: number  # Default: 0.15 (15%)
  min_occurrences: number        # Default: 3 (pattern must occur ≥3 times)

refactoring_options:
  incremental: boolean      # Default: true (refactor one-by-one)
  test_after_each: boolean  # Default: true
  allow_exceptions: boolean # Default: true (leave special cases unchanged)
```

## Execution Process

### Step 1: Identify Duplication via Analysis

```bash
# Look for repeated structures
grep -A 5 "InputSchema:" tools.go | less

# Count pattern occurrences
grep -c "\"scope\":" tools.go    # e.g., 15 occurrences

# Calculate duplication percentage
total_lines=$(wc -l < tools.go)
duplicate_lines=$(grep -E "(scope|jq_filter|stats_only)" tools.go | wc -l)
duplication_pct=$(echo "scale=2; $duplicate_lines / $total_lines * 100" | bc)
echo "Duplication: ${duplication_pct}%"
```

**Decision**:
- If duplication < 15%: Skip (not worth refactoring)
- If duplication 15-30%: Consider (medium priority)
- If duplication > 30%: High priority

### Step 2: Categorize Duplication

**Common Parameters** (≥50% of definitions):
```go
// These appear in majority of tools
"scope"      // 15/15 tools (100%)
"jq_filter"  // 15/15 tools (100%)
"stats_only" // 15/15 tools (100%)
```

**Optional Parameters** (20-50% of definitions):
```go
// These appear in some tools
"limit"      // 8/15 tools (53%)
"offset"     // 3/15 tools (20%)
```

**Unique Parameters** (<20% of definitions):
```go
// Tool-specific
"error_signature"  // 1/15 tools (query_context only)
"window"           // 1/15 tools (query_context only)
```

**Categorization Output**:
```yaml
common_parameters:
  - name: "scope"
    occurrences: 15
    percentage: 100%
  - name: "jq_filter"
    occurrences: 15
    percentage: 100%
  - name: "stats_only"
    occurrences: 15
    percentage: 100%

optional_parameters:
  - name: "limit"
    occurrences: 8
    percentage: 53%

unique_parameters:
  - name: "error_signature"
    occurrences: 1
    percentage: 7%
```

### Step 3: Extract Smallest Reusable Unit First

**Create helper for common parameters**:

```go
// Before: Inline definitions (repeated 15 times)
{
    Name: "query_tools",
    InputSchema: ToolSchema{
        Type: "object",
        Properties: map[string]Property{
            "tool": {...},
            "status": {...},
            "scope": {                    // ← Repeated
                Type: "string",
                Description: "Query scope: 'project' (default) or 'session'",
            },
            "jq_filter": {                // ← Repeated
                Type: "string",
                Description: "jq expression for filtering (default: '.[]')",
            },
            "stats_only": {               // ← Repeated
                Type: "boolean",
                Description: "Return only statistics (default: false)",
            },
        },
    },
}

// After: Extract common parameters
func StandardToolParameters() map[string]Property {
    return map[string]Property{
        "scope": {
            Type:        "string",
            Description: "Query scope: 'project' (default) or 'session'",
        },
        "jq_filter": {
            Type:        "string",
            Description: "jq expression for filtering (default: '.[]')",
        },
        "stats_only": {
            Type:        "boolean",
            Description: "Return only statistics (default: false)",
        },
        // ... other common parameters
    }
}
```

### Step 4: Create Merge Helper Function

```go
func MergeParameters(specific map[string]Property) map[string]Property {
    result := make(map[string]Property)

    // Add standard parameters first
    for k, v := range StandardToolParameters() {
        result[k] = v
    }

    // Override/add specific parameters
    for k, v := range specific {
        result[k] = v  // Specific parameters can override standard
    }

    return result
}
```

### Step 5: Create Schema Builder Helper

```go
func buildToolSchema(properties map[string]Property, required ...string) ToolSchema {
    schema := ToolSchema{
        Type:       "object",
        Properties: MergeParameters(properties),  // Merge common + specific
    }

    if len(required) > 0 {
        schema.Required = required
    }

    return schema
}
```

### Step 6: Create Top-Level Builder (Optional)

```go
func buildTool(name, description string, properties map[string]Property, required ...string) Tool {
    return Tool{
        Name:        name,
        Description: description,
        InputSchema: buildToolSchema(properties, required...),
    }
}
```

### Step 7: Refactor One Usage (Proof-of-Concept)

```go
// Before (60+ lines)
{
    Name:        "query_tools",
    Description: "Query tool calls with filters. Default scope: project.",
    InputSchema: ToolSchema{
        Type: "object",
        Properties: map[string]Property{
            "tool": {
                Type:        "string",
                Description: "Filter by tool name",
            },
            "status": {
                Type:        "string",
                Description: "Filter by status (error/success)",
            },
            "scope": {                          // ← Common
                Type:        "string",
                Description: "Query scope: 'project' (default) or 'session'",
            },
            "jq_filter": {                      // ← Common
                Type:        "string",
                Description: "jq expression for filtering (default: '.[]')",
            },
            "stats_only": {                     // ← Common
                Type:        "boolean",
                Description: "Return only statistics (default: false)",
            },
        },
    },
}

// After (10 lines)
buildTool("query_tools", "Query tool calls with filters. Default scope: project.", map[string]Property{
    "tool": {
        Type:        "string",
        Description: "Filter by tool name",
    },
    "status": {
        Type:        "string",
        Description: "Filter by status (error/success)",
    },
})
// StandardToolParameters() automatically merged!
```

### Step 8: Verify Tests Pass After First Refactoring

```bash
# Before refactoring
go test ./cmd/mcp-server > baseline.txt

# After refactoring one tool
go test ./cmd/mcp-server > after_one.txt

# Compare
diff baseline.txt after_one.txt
# Expected: No difference (behavioral equivalence)

# If tests pass: Continue
# If tests fail: Rollback, investigate
```

### Step 9: Refactor Remaining Usages Incrementally

**Process**: One tool at a time
```bash
# Refactor Tool 2
# Test
go test ./cmd/mcp-server

# Refactor Tool 3
# Test
go test ./cmd/mcp-server

# ... repeat for all tools
```

**Commit Strategy**:
```bash
git add tools.go
git commit -m "refactor: extract helpers for Tools 1-5"

# ... batch in groups of 3-5 tools
```

### Step 10: Identify and Document Exceptions

**Some tools may not fit the pattern**:

```go
// Exception: query_context has custom structure
{
    Name: "query_context",
    InputSchema: ToolSchema{
        // Unique structure, don't force into builder
    },
}

// Document in code
// NOTE: query_context intentionally not using buildTool() due to unique parameter dependencies
```

**Reasons for exceptions**:
- Custom parameter dependencies
- Unique validation logic
- Legacy compatibility requirements

### Step 11: Remove Old Code (Use agent-verify-before-remove)

```bash
# After all refactoring complete
# Old inline definitions are now replaced

# Verify old code is unused
/subagent @agents/agent-verify-before-remove.md \
  target_code.function="getToolDefinitionsOld" \
  scope="package"

# If SAFE_TO_REMOVE: Delete old code
```

## Output Schema

```yaml
extraction_summary:
  extracted_helpers:
    - name: "StandardToolParameters"
      lines: number
      reused_by: [string]  # List of tools using this helper

    - name: "buildToolSchema"
      lines: number
      reused_by: [string]

  refactored_usages:
    - tool: "query_tools"
      before_lines: number
      after_lines: number
      reduction_percentage: number

  exceptions:
    - tool: "query_context"
      reason: "Unique parameter dependencies"
      refactored: boolean  # false

  metrics:
    total_lines_before: number
    total_lines_after: number
    lines_reduced: number
    reduction_percentage: number

  test_results:
    status: "PASS" | "FAIL"
    baseline_pass_count: number
    after_pass_count: number
    behavioral_equivalence: boolean

quality_checks:
  duplication_before: number  # %
  duplication_after: number   # %
  duplication_reduction: number

  maintainability_improvement:
    change_impact: string  # "Common params: 1 edit (not 15)"
    future_effort: string  # "Adding new standard param: 1 line"
```

## Success Criteria

- ✅ Line reduction ≥ 10%
- ✅ Duplication reduction ≥ 50%
- ✅ All tests pass (100% behavioral equivalence)
- ✅ At least 50% of definitions refactored
- ✅ Exceptions documented with clear rationale

## Example Execution (meta-cc Iteration 2)

**Input**:
```yaml
target_file:
  path: "cmd/mcp-server/tools.go"
  type: "api_schema"

analysis:
  duplication_threshold: 0.15
```

**Process**:
```
Step 1: Analysis
  Total lines: 396
  Common parameters: 3 (scope, jq_filter, stats_only)
  Duplication: 69 lines (17.4%) ← Above threshold

Step 2: Categorization
  Common: 3 parameters (100% occurrence)
  Optional: 2 parameters
  Unique: varies by tool

Step 3: Extract StandardToolParameters() (15 lines)

Step 4: Create MergeParameters() (10 lines)

Step 5: Create buildToolSchema() (12 lines)

Step 6: Create buildTool() (8 lines)

Step 7: Refactor query_tools (test: PASS)

Step 8-9: Refactor remaining 14 tools incrementally
  Tools 1-5: PASS
  Tools 6-10: PASS
  Tools 11-15: PASS (12 refactored, 3 exceptions)

Step 10: Document exceptions (query_context, etc.)

Step 11: Remove old inline definitions (verified unused)
```

**Output**:
```yaml
metrics:
  total_lines_before: 396
  total_lines_after: 321
  lines_reduced: 75
  reduction_percentage: 18.9%

test_results:
  status: "PASS"
  behavioral_equivalence: true

duplication_reduction: 100%  # (69 lines → 0 lines)

exceptions:
  count: 3
  tools: ["query_context", "special_tool_a", "special_tool_b"]
```

## Pitfalls and How to Avoid

### Pitfall 1: Assuming Duplication Without Analysis
- ❌ Wrong: "These look similar, I'll extract helpers"
- ✅ Right: Grep for patterns, count occurrences, calculate percentages

### Pitfall 2: Over-Abstraction (DRY Taken Too Far)
- ❌ Wrong: Force all 15 tools into same helper even when structures differ
- ✅ Right: Leave exceptions unchanged, document why

### Pitfall 3: Bulk Refactoring Without Incremental Testing
- ❌ Wrong: Refactor all 15 tools, then test once
- ✅ Right: Refactor one, test, commit; repeat

### Pitfall 4: Not Measuring Impact
- ❌ Wrong: "Code looks cleaner" (subjective)
- ✅ Right: "Reduced 75 lines, 18.9%" (objective)

## Variations

### Variation 1: Inheritance/Composition (OOP Languages)

**Python Example**:
```python
# Base class with common fields
class BaseSchema:
    def __init__(self):
        self.scope = StringField(description="Query scope")
        self.stats_only = BooleanField(description="Return only stats")

# Specific schema inherits
class QueryToolsSchema(BaseSchema):
    def __init__(self):
        super().__init__()
        self.tool = StringField(description="Filter by tool")
        self.status = StringField(description="Filter by status")
```

### Variation 2: Mixins/Traits (TypeScript)

```typescript
// Mixin with common properties
interface StandardParams {
    scope?: string;
    stats_only?: boolean;
}

// Specific interface extends
interface QueryToolsParams extends StandardParams {
    tool?: string;
    status?: string;
}
```

### Variation 3: Builder Pattern (Fluent API)

**Java Example**:
```java
Tool tool = new ToolBuilder()
    .name("query_tools")
    .description("Query tool calls")
    .addStandardParameters()       // ← Adds common params
    .addParameter("tool", StringType)
    .addParameter("status", StringType)
    .build();
```

## Language-Specific Adaptations

### Go
- Use functions returning maps
- Merge via map iteration
- Test with go test -cover

### Python
- Use class inheritance
- Test with pytest

### JavaScript/TypeScript
- Use object spread {...base, ...specific}
- Test with Jest

### Java
- Use Builder pattern
- Test with JUnit

## Usage Examples

### As Subagent
```bash
/subagent @experiments/bootstrap-004-refactoring-guide/agents/agent-builder-extractor.md \
  target_file.path="cmd/mcp-server/tools.go" \
  target_file.type="api_schema" \
  analysis.duplication_threshold=0.15
```

### As Slash Command (if registered)
```bash
/extract-builders \
  file="cmd/mcp-server/tools.go" \
  type="api_schema" \
  threshold=0.15
```

## Evidence from Bootstrap-004

**Source**: meta-cc Iteration 2
**File**: cmd/mcp-server/tools.go (396 lines)

**Results**:
- Lines reduced: 75 (18.9%)
- Duplication eliminated: 69 lines → 0 lines (100%)
- Tools refactored: 12/15 (80%)
- Exceptions: 3 (documented)
- Test pass rate: 100%
- Time spent: ~4 hours

**Maintainability improvement**:
- Before: Change common param = edit 15 places
- After: Change common param = edit 1 place
- Effort reduction: 93%

---

**Last Updated**: 2025-10-16
**Status**: Validated (meta-cc Iteration 2)
**Reusability**: Universal (any language with structure definitions)
