# Task 2: Validation Tool MVP Implementation Specification

**Agent**: coder
**Date**: 2025-10-15
**Iteration**: 3
**Status**: Design Complete (Ready for Implementation)

---

## Objective

Develop `meta-cc validate-api` command to automatically validate API consistency according to guidelines from Iteration 2.

**MVP Scope**: 3 core checks (naming, parameter ordering, description format)
**Deferred**: Schema validation, standard parameter checking, auto-fix mode

---

## Tool Specification

### Command

```bash
meta-cc validate-api [OPTIONS]
```

### Options

```
--file <path>       Path to tools.go (default: cmd/mcp-server/tools.go)
--fast              Run fast checks only (MVP mode, default)
--full              Run all checks (deferred to future iteration)
--fix               Auto-fix safe violations (deferred to future iteration)
--quiet             Suppress output except errors
--json              Output results as JSON
```

### Exit Codes

```
0 - All checks passed
1 - Violations found
2 - Invalid usage or error
```

---

## MVP Checks (3 Core Validations)

### Check 1: Naming Pattern Validation

**Purpose**: Verify tool names follow prefix conventions

**Logic**:
```python
def validate_naming(tool_name: str) -> Result:
    """
    Check if tool name uses standard prefix and follows conventions.
    """
    valid_prefixes = ["query_", "get_", "list_", "cleanup_"]

    # Check prefix
    if not any(tool_name.startswith(p) for p in valid_prefixes):
        return FAIL(
            tool=tool_name,
            check="naming_pattern",
            message=f"Tool name must use standard prefix ({', '.join(valid_prefixes)})",
            severity="ERROR"
        )

    # Check snake_case
    if not is_snake_case(tool_name):
        return FAIL(
            tool=tool_name,
            check="naming_format",
            message="Tool name must use snake_case",
            severity="ERROR"
        )

    # Check length
    if len(tool_name) > 40:
        return WARN(
            tool=tool_name,
            check="naming_length",
            message=f"Tool name exceeds 40 characters ({len(tool_name)} chars)",
            severity="WARNING"
        )

    return PASS(tool=tool_name, check="naming")

def is_snake_case(name: str) -> bool:
    """Check if string is snake_case."""
    return name.islower() and '_' in name and not ' ' in name
```

**Example Output**:
```
✗ get_session_stats: Should use query_* prefix (returns filtered data)
  Suggestion: Rename to query_session_stats
  Reference: api-naming-convention.md (Section 2.1)

✓ query_tools: Naming follows convention
✓ list_capabilities: Naming follows convention
```

---

### Check 2: Parameter Ordering Validation

**Purpose**: Verify parameters follow tier-based ordering

**Logic**:
```python
def validate_parameter_order(tool: Tool) -> Result:
    """
    Validate parameter ordering follows tier system.
    """
    # Extract parameters (excluding standard params)
    tool_params = get_tool_specific_params(tool)
    required_params = tool.InputSchema.get("required", [])

    # Categorize parameters by tier
    tiers = categorize_params(tool_params, required_params)

    # Build expected order
    expected_order = (
        tiers[1] +  # Required
        tiers[2] +  # Filtering
        tiers[3] +  # Range
        tiers[4]    # Output control
    )

    # Get actual order
    actual_order = list(tool_params.keys())

    # Compare
    if expected_order != actual_order:
        return FAIL(
            tool=tool.Name,
            check="parameter_ordering",
            message="Parameter ordering violates tier system",
            expected=expected_order,
            actual=actual_order,
            severity="ERROR"
        )

    return PASS(tool=tool.Name, check="parameter_ordering")

def categorize_params(params: dict, required: list) -> dict:
    """
    Categorize parameters into tiers.
    """
    tiers = {1: [], 2: [], 3: [], 4: []}

    for param_name in params.keys():
        if param_name in required:
            tiers[1].append(param_name)
        elif is_filtering_param(param_name):
            tiers[2].append(param_name)
        elif is_range_param(param_name):
            tiers[3].append(param_name)
        elif is_output_param(param_name):
            tiers[4].append(param_name)
        else:
            # Unknown category - log warning
            print(f"Warning: Cannot categorize parameter '{param_name}'")

    return tiers

def is_filtering_param(name: str) -> bool:
    """
    Determine if parameter is filtering type.
    Common patterns: tool, status, pattern, filter, where, etc.
    """
    filtering_patterns = [
        "tool", "status", "pattern", "filter", "where",
        "type", "category", "target", "include_", "exclude_"
    ]
    return any(p in name for p in filtering_patterns)

def is_range_param(name: str) -> bool:
    """
    Determine if parameter is range type.
    Patterns: min_*, max_*, start_*, end_*, threshold, window
    """
    range_patterns = ["min_", "max_", "start_", "end_", "threshold", "window"]
    return any(name.startswith(p) or name == p for p in range_patterns)

def is_output_param(name: str) -> bool:
    """
    Determine if parameter is output control type.
    Patterns: limit, offset, page, cursor
    """
    output_patterns = ["limit", "offset", "page", "cursor"]
    return name in output_patterns
```

**Example Output**:
```
✗ query_tools: Parameter ordering incorrect
  Expected: tool, status, limit
  Actual:   limit, tool, status

  Tier-based ordering:
    Tier 2 (Filtering): tool, status
    Tier 4 (Output):    limit

  Reference: api-parameter-convention.md (Section 2)

✓ query_context: Parameter ordering correct
```

---

### Check 3: Description Format Validation

**Purpose**: Verify tool descriptions follow template

**Logic**:
```python
import re

def validate_description(tool: Tool) -> Result:
    """
    Validate tool description follows template format.
    Template: "<Action> <object>. Default scope: <project|session|none>."
    """
    desc = tool.Description

    # Check template pattern
    pattern = r'^[A-Z].*\. Default scope: (project|session|none)\.$'
    if not re.match(pattern, desc):
        return FAIL(
            tool=tool.Name,
            check="description_format",
            message="Description must match template",
            template="'<Action> <object>. Default scope: <X>.'",
            actual=desc,
            severity="ERROR"
        )

    # Check length
    if len(desc) > 100:
        return WARN(
            tool=tool.Name,
            check="description_length",
            message=f"Description exceeds 100 characters ({len(desc)} chars)",
            severity="WARNING"
        )

    # Check for "Default scope:" presence
    if "Default scope:" not in desc:
        return FAIL(
            tool=tool.Name,
            check="description_scope",
            message="Description must include 'Default scope:' suffix",
            severity="ERROR"
        )

    return PASS(tool=tool.Name, check="description")
```

**Example Output**:
```
✓ query_tools: Description follows template
✗ query_files: Description missing "Default scope:" suffix
  Actual: "File operation stats (returns array)."
  Expected: "File operation stats. Default scope: project."

  Reference: api-consistency-methodology.md (Section 4)
```

---

## Implementation Architecture

### File Structure

```
cmd/
  validate-api.go          # Main command implementation
internal/
  validation/
    validator.go           # Core validation logic
    naming.go              # Check 1: Naming validation
    ordering.go            # Check 2: Parameter ordering validation
    description.go         # Check 3: Description validation
    parser.go              # Tool definition parser (AST or regex)
    reporter.go            # Result formatting and output
    types.go               # Type definitions (Tool, Result, etc.)
```

### Core Types

```go
// types.go

package validation

// Tool represents a parsed MCP tool definition
type Tool struct {
    Name        string
    Description string
    InputSchema struct {
        Type       string
        Properties map[string]interface{}
        Required   []string
    }
}

// Result represents a validation result
type Result struct {
    Tool     string
    Check    string
    Status   string // "PASS", "FAIL", "WARN"
    Message  string
    Severity string // "ERROR", "WARNING", "INFO"
    Details  map[string]interface{} // Expected, Actual, etc.
}

// Report aggregates all validation results
type Report struct {
    TotalTools    int
    ChecksRun     int
    Passed        int
    Failed        int
    Warnings      int
    Results       []Result
    Summary       string
}
```

### Main Command Flow

```go
// cmd/validate-api.go

package main

import (
    "flag"
    "fmt"
    "os"

    "github.com/yaleh/meta-cc/internal/validation"
)

func main() {
    // Parse flags
    filePath := flag.String("file", "cmd/mcp-server/tools.go", "Path to tools.go")
    fast := flag.Bool("fast", true, "Run fast checks only")
    quiet := flag.Bool("quiet", false, "Suppress output except errors")
    jsonOutput := flag.Bool("json", false, "Output as JSON")
    flag.Parse()

    // Parse tools.go
    tools, err := validation.ParseTools(*filePath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error parsing tools.go: %v\n", err)
        os.Exit(2)
    }

    // Run validations
    validator := validation.NewValidator(*fast)
    report := validator.Validate(tools)

    // Output results
    reporter := validation.NewReporter(*quiet, *jsonOutput)
    reporter.Print(report)

    // Exit with appropriate code
    if report.Failed > 0 {
        os.Exit(1)
    }
    os.Exit(0)
}
```

---

## Parsing Strategy

### Option 1: Regex-Based Parsing (Simpler, MVP)

**Approach**: Extract tool definitions using regex patterns

**Pros**:
- Simple implementation
- No AST dependency
- Fast execution

**Cons**:
- Fragile (breaks if tools.go structure changes)
- Limited accuracy

**Example**:
```go
func parseToolsRegex(filePath string) ([]Tool, error) {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return nil, err
    }

    // Regex to match tool definitions
    toolPattern := `Name:\s*"([^"]+)",\s*Description:\s*"([^"]+)"`
    paramPattern := `"([^"]+)":\s*map\[string\]interface\{\}`

    // Extract tools and parameters
    // ...

    return tools, nil
}
```

### Option 2: AST-Based Parsing (Better, More Complex)

**Approach**: Use Go's `go/parser` and `go/ast` packages

**Pros**:
- Accurate parsing
- Robust against formatting changes
- Can extract all metadata

**Cons**:
- More complex implementation
- Slower execution

**Recommendation**: Use regex for MVP, plan AST for future enhancement

---

## Testing

### Unit Tests

**File**: `internal/validation/naming_test.go`

```go
func TestValidateNaming(t *testing.T) {
    tests := []struct {
        name     string
        toolName string
        wantPass bool
    }{
        {"valid query prefix", "query_tools", true},
        {"valid get prefix", "get_capability", true},
        {"invalid prefix", "retrieve_tools", false},
        {"not snake_case", "queryTools", false},
        {"too long", "query_very_long_tool_name_exceeding_limit", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := validateNaming(tt.toolName)
            if (result.Status == "PASS") != tt.wantPass {
                t.Errorf("validateNaming(%q) = %v, want pass=%v",
                    tt.toolName, result.Status, tt.wantPass)
            }
        })
    }
}
```

### Integration Tests

**File**: `cmd/validate-api_test.go`

```go
func TestValidateAPICommand(t *testing.T) {
    // Create temporary tools.go with violations
    tmpFile := createTestToolsFile(t)
    defer os.Remove(tmpFile)

    // Run command
    cmd := exec.Command("./meta-cc", "validate-api", "--file", tmpFile)
    output, err := cmd.CombinedOutput()

    // Verify exit code
    if err == nil {
        t.Error("Expected non-zero exit code for file with violations")
    }

    // Verify output contains expected violations
    if !strings.Contains(string(output), "naming_pattern") {
        t.Error("Expected naming violation in output")
    }
}
```

---

## Example Output

### Terminal Output (Default)

```
API Consistency Validation
==========================

Analyzing cmd/mcp-server/tools.go...
Found 16 tools

Running checks (MVP mode):
  ✓ Naming pattern validation
  ✓ Parameter ordering validation
  ✓ Description format validation

Results:
--------

✗ get_session_stats: Naming pattern violation
  Tool name should use query_* prefix (returns filtered data)
  Suggestion: Rename to query_session_stats
  Reference: api-naming-convention.md (Section 2.1)
  Severity: ERROR

✗ query_tools: Parameter ordering violation
  Expected: tool, status, limit
  Actual:   limit, tool, status
  Reference: api-parameter-convention.md (Section 2)
  Severity: ERROR

✗ query_user_messages: Parameter ordering violation
  Expected: pattern, max_message_length, limit, content_summary
  Actual:   pattern, limit, max_message_length, content_summary
  Reference: api-parameter-convention.md (Section 2)
  Severity: ERROR

✓ list_capabilities: All checks passed
✓ get_capability: All checks passed
✓ cleanup_temp_files: All checks passed
... (13 more tools)

Summary:
--------
Total tools:     16
Checks run:      48 (3 checks × 16 tools)
Passed:          42
Failed:          3
Warnings:        3

Overall Status: FAILED (3 violations found)

Exit code: 1
```

### JSON Output (--json)

```json
{
  "total_tools": 16,
  "checks_run": 48,
  "passed": 42,
  "failed": 3,
  "warnings": 3,
  "results": [
    {
      "tool": "get_session_stats",
      "check": "naming_pattern",
      "status": "FAIL",
      "message": "Tool name should use query_* prefix",
      "severity": "ERROR",
      "details": {
        "suggestion": "query_session_stats",
        "reference": "api-naming-convention.md"
      }
    },
    ...
  ],
  "summary": "FAILED (3 violations found)"
}
```

---

## Documentation

### CLI Reference Entry

**File**: `docs/reference/cli.md` (add section)

```markdown
## meta-cc validate-api

**Purpose**: Validate API consistency according to established conventions

**Usage**:
```bash
meta-cc validate-api [OPTIONS]
```

**Options**:
- `--file <path>` - Path to tools.go (default: cmd/mcp-server/tools.go)
- `--fast` - Run fast checks only (MVP mode, default)
- `--quiet` - Suppress output except errors
- `--json` - Output results as JSON

**Exit Codes**:
- `0` - All checks passed
- `1` - Violations found
- `2` - Invalid usage or error

**Checks Performed** (MVP):
1. Naming pattern validation (query_*, get_*, list_*, cleanup_*)
2. Parameter ordering validation (tier-based system)
3. Description format validation (template compliance)

**Example**:
```bash
# Validate current API
meta-cc validate-api

# Validate specific file
meta-cc validate-api --file path/to/tools.go

# JSON output for CI integration
meta-cc validate-api --json
```

**References**:
- Naming conventions: `data/api-naming-convention.md`
- Parameter ordering: `data/api-parameter-convention.md`
- Validation methodology: `data/api-consistency-methodology.md`
```

---

## Integration

### CI Integration

**File**: `.github/workflows/api-consistency.yml`

```yaml
name: API Consistency Check

on: [pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Build meta-cc
        run: make build

      - name: Validate API consistency
        run: ./meta-cc validate-api --file cmd/mcp-server/tools.go

      - name: Upload validation report
        if: failure()
        uses: actions/upload-artifact@v3
        with:
          name: validation-report
          path: validation-report.json
```

---

## Success Criteria

✅ Tool correctly parses tools.go (16 tools extracted)
✅ Naming pattern check identifies `get_session_stats` violation
✅ Parameter ordering check identifies 3 violations (query_tools, query_user_messages, query_conversation)
✅ Description format check verifies all tools (16/16 pass)
✅ Exit code 0 for passing tools, 1 for violations
✅ Actionable error messages with references
✅ Tests pass (unit + integration)
✅ Documentation complete (CLI reference)

---

## Deferred to Future Iterations

**Not in MVP**:
1. Schema structure validation (Check 4)
2. Standard parameter presence check (Check 5)
3. Auto-fix mode (`--fix`)
4. Full mode (`--full`) with semantic checks
5. AST-based parsing (using regex for MVP)

**Rationale**: Focus on core validation to enable quality gates (pre-commit hook)

---

## Effort Estimate

**Time**: 8-12 hours
- 2 hours: Parser implementation (regex-based)
- 3 hours: Check implementations (naming, ordering, description)
- 2 hours: Reporter and output formatting
- 2 hours: Testing (unit + integration)
- 1 hour: Documentation (CLI reference, integration guide)

**Complexity**: MODERATE (parsing + validation logic + testing)

---

**Specification Status**: ✅ COMPLETE
**Ready for Implementation**: YES
**Next Step**: Implement in `cmd/validate-api.go` and `internal/validation/`
