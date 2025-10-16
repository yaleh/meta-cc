# Agent: Automated Consistency Validation Tool Builder

**Version**: 1.0
**Source**: Bootstrap-006, Pattern 4
**Success Rate**: 100% accuracy (2 violations detected, 0 false positives)

---

## Role

Build automated validation tools to enforce API conventions at scale, ensuring consistency without manual checks.

## When to Use

- Need to enforce documented conventions automatically
- Manual consistency checks are error-prone
- Inconsistencies accumulate over time
- Want to prevent violations (not just detect post-hoc)
- Building quality assurance infrastructure

## Input Schema

```yaml
validation_target:
  format: string                # Required: "json_schema" | "openapi" | "graphql"
  file: string                  # Required: File to validate
  conventions: [string]         # Required: List of conventions to check

tool_architecture:
  parser: string                # "regex" | "ast" | "custom"
  validators: [string]          # List of validator names
  reporter: string              # "terminal" | "json" | "both"

validators:
  - name: string                # Required: Validator name
    description: string         # Required: What it checks
    check_type: string          # "naming" | "ordering" | "structure" | "content"
    severity: string            # "ERROR" | "WARNING"
    rule: string                # Deterministic check rule

output_config:
  formats: [string]             # ["terminal", "json"]
  include_suggestions: boolean  # Default: true
  include_references: boolean   # Default: true
```

## Execution Process

### Step 1: Design Type System

**Core Types**:
```go
// Tool represents an API tool/endpoint
type Tool struct {
    Name        string
    Description string
    Parameters  []Parameter
}

// Parameter represents a tool parameter
type Parameter struct {
    Name        string
    Type        string
    Description string
    Required    bool
}

// ValidationResult represents check outcome
type ValidationResult struct {
    ToolName  string
    CheckName string
    Passed    bool
    Message   string
    Severity  string
    Details   map[string]interface{}
}

// Report aggregates all validation results
type Report struct {
    TotalTools   int
    TotalChecks  int
    Passed       int
    Failed       int
    Warnings     int
    Results      []ValidationResult
}
```

### Step 2: Implement Parser

**Decision**: Regex (simple, fast) vs. AST (robust, complex)

**For MVP: Use Regex**
```go
func ParseTools(content string) ([]Tool, error) {
    // Regex patterns for JSON schema
    toolPattern := regexp.MustCompile(`Name:\s*"([^"]+)"`)
    descPattern := regexp.MustCompile(`Description:\s*"([^"]+)"`)
    paramPattern := regexp.MustCompile(`"([^"]+)":\s*{[^}]*Type:\s*"([^"]+)"`)

    var tools []Tool
    // ... parsing logic
    return tools, nil
}
```

**For Production: Use AST**
```go
func ParseTools(filePath string) ([]Tool, error) {
    // Parse Go AST
    fset := token.NewFileSet()
    node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
    if err != nil {
        return nil, err
    }

    // Extract tool definitions
    var tools []Tool
    ast.Inspect(node, func(n ast.Node) bool {
        // ... AST traversal logic
        return true
    })

    return tools, nil
}
```

### Step 3: Create Validators with Deterministic Rules

**Validator Pattern**:
```go
type Validator interface {
    Name() string
    Check(tool Tool) ValidationResult
}

// Example: Naming Convention Validator
type NamingValidator struct {
    Pattern *regexp.Regexp
}

func (v *NamingValidator) Check(tool Tool) ValidationResult {
    // 1. Extract relevant data
    name := tool.Name

    // 2. Apply deterministic check
    if !v.Pattern.MatchString(name) {
        return ValidationResult{
            ToolName:  tool.Name,
            CheckName: "naming_convention",
            Passed:    false,
            Message:   fmt.Sprintf("Tool name '%s' violates naming convention", name),
            Severity:  "ERROR",
            Details: map[string]interface{}{
                "suggestion": "Use snake_case format",
                "expected":   "snake_case_pattern",
                "actual":     name,
                "reference":  "docs/api-naming-convention.md",
            },
        }
    }

    // 3. Return pass
    return ValidationResult{
        ToolName:  tool.Name,
        CheckName: "naming_convention",
        Passed:    true,
        Severity:  "INFO",
    }
}
```

**Validator: Parameter Ordering**
```go
type OrderingValidator struct {
    TierSystem TierDefinitions
}

func (v *OrderingValidator) Check(tool Tool) ValidationResult {
    // Categorize parameters by tier
    categorized := categorizeParameters(tool.Parameters, v.TierSystem)

    // Get expected order (tier-based)
    expectedOrder := sortByTier(categorized)

    // Get actual order
    actualOrder := tool.Parameters

    // Compare
    for i := range expectedOrder {
        if i < len(actualOrder) && expectedOrder[i].Name != actualOrder[i].Name {
            return ValidationResult{
                ToolName:  tool.Name,
                CheckName: "parameter_ordering",
                Passed:    false,
                Message:   fmt.Sprintf("Parameters not in tier order"),
                Severity:  "ERROR",
                Details: map[string]interface{}{
                    "suggestion": "Reorder parameters by tier (1→2→3→4→5)",
                    "expected":   paramNames(expectedOrder),
                    "actual":     paramNames(actualOrder),
                    "reference":  "docs/api-parameter-convention.md",
                },
            }
        }
    }

    return ValidationResult{
        ToolName:  tool.Name,
        CheckName: "parameter_ordering",
        Passed:    true,
        Severity:  "INFO",
    }
}
```

**Validator: Description Format**
```go
type DescriptionValidator struct {
    RequiredPattern string  // e.g., "Default scope: <scope>"
}

func (v *DescriptionValidator) Check(tool Tool) ValidationResult {
    desc := tool.Description

    // Check if pattern present
    if !strings.Contains(desc, v.RequiredPattern) {
        return ValidationResult{
            ToolName:  tool.Name,
            CheckName: "description_format",
            Passed:    false,
            Message:   fmt.Sprintf("Missing required pattern: '%s'", v.RequiredPattern),
            Severity:  "WARNING",
            Details: map[string]interface{}{
                "suggestion": fmt.Sprintf("Add '%s' to description", v.RequiredPattern),
                "expected":   "Description with scope declaration",
                "actual":     desc,
                "reference":  "docs/api-consistency-methodology.md",
            },
        }
    }

    return ValidationResult{
        ToolName:  tool.Name,
        CheckName: "description_format",
        Passed:    true,
        Severity:  "INFO",
    }
}
```

### Step 4: Build Reporter with Multiple Formats

**Terminal Reporter** (Human-Readable):
```go
func ReportTerminal(report Report) {
    fmt.Println("===========================================")
    fmt.Printf("API Validation Report\n")
    fmt.Println("===========================================\n")

    fmt.Printf("Tools validated: %d\n", report.TotalTools)
    fmt.Printf("Checks performed: %d\n", report.TotalChecks)
    fmt.Printf("✓ Passed: %d\n", report.Passed)
    fmt.Printf("✗ Failed: %d\n", report.Failed)
    fmt.Printf("⚠ Warnings: %d\n\n", report.Warnings)

    // Group by tool
    for _, result := range report.Results {
        if !result.Passed {
            fmt.Printf("✗ %s: %s\n", result.ToolName, result.Message)

            if suggestion, ok := result.Details["suggestion"]; ok {
                fmt.Printf("  Suggestion: %s\n", suggestion)
            }
            if expected, ok := result.Details["expected"]; ok {
                fmt.Printf("  Expected: %v\n", expected)
            }
            if actual, ok := result.Details["actual"]; ok {
                fmt.Printf("  Actual: %v\n", actual)
            }
            if ref, ok := result.Details["reference"]; ok {
                fmt.Printf("  Reference: %s\n", ref)
            }
            fmt.Printf("  Severity: %s\n\n", result.Severity)
        }
    }

    // Exit code
    if report.Failed > 0 {
        os.Exit(1)
    }
}
```

**JSON Reporter** (Machine-Readable, for CI):
```go
func ReportJSON(report Report) {
    output, err := json.MarshalIndent(report, "", "  ")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(string(output))

    // Exit code
    if report.Failed > 0 {
        os.Exit(1)
    }
}
```

**Example Terminal Output**:
```
===========================================
API Validation Report
===========================================

Tools validated: 16
Checks performed: 48
✓ Passed: 46
✗ Failed: 2
⚠ Warnings: 0

✗ list_capabilities: Missing required pattern: 'Default scope:'
  Suggestion: Add 'Default scope: <scope>' to description
  Expected: Description with scope declaration
  Actual: List all available capabilities from configured sources. Returns compact capability index.
  Reference: docs/api-consistency-methodology.md
  Severity: WARNING

✗ get_capability: Missing required pattern: 'Default scope:'
  Suggestion: Add 'Default scope: <scope>' to description
  Expected: Description with scope declaration
  Actual: Retrieve complete capability content by name from configured sources.
  Reference: docs/api-consistency-methodology.md
  Severity: WARNING
```

### Step 5: Integrate into CLI with Standard Flags

**CLI Structure**:
```bash
validate-api [options] <file>

Options:
  --check <name>      Run specific check (naming, ordering, description)
  --format <format>   Output format (terminal, json)
  --severity <level>  Minimum severity to report (ERROR, WARNING)
  --fast              Skip slow checks
  --help              Show help
```

**CLI Implementation**:
```go
func main() {
    // Parse flags
    checkFlag := flag.String("check", "all", "Check to run")
    formatFlag := flag.String("format", "terminal", "Output format")
    severityFlag := flag.String("severity", "ERROR", "Minimum severity")
    fastFlag := flag.Bool("fast", false, "Skip slow checks")
    flag.Parse()

    // Validate file argument
    if flag.NArg() < 1 {
        log.Fatal("Usage: validate-api [options] <file>")
    }
    filePath := flag.Arg(0)

    // Parse tools
    tools, err := ParseTools(filePath)
    if err != nil {
        log.Fatal(err)
    }

    // Create validators
    validators := createValidators(*checkFlag, *fastFlag)

    // Run validation
    report := validate(tools, validators)

    // Filter by severity
    report = filterBySeverity(report, *severityFlag)

    // Output report
    if *formatFlag == "json" {
        ReportJSON(report)
    } else {
        ReportTerminal(report)
    }
}
```

### Step 6: Add Test Suite

**Unit Tests for Validators**:
```go
func TestNamingValidator(t *testing.T) {
    validator := &NamingValidator{
        Pattern: regexp.MustCompile(`^[a-z][a-z0-9_]*$`),
    }

    tests := []struct {
        name     string
        toolName string
        wantPass bool
    }{
        {"valid snake_case", "query_tools", true},
        {"invalid camelCase", "queryTools", false},
        {"invalid PascalCase", "QueryTools", false},
        {"valid with numbers", "query_tools_v2", true},
        {"invalid hyphen", "query-tools", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tool := Tool{Name: tt.toolName}
            result := validator.Check(tool)

            if result.Passed != tt.wantPass {
                t.Errorf("Expected passed=%v, got passed=%v", tt.wantPass, result.Passed)
            }
        })
    }
}
```

**Integration Tests**:
```go
func TestValidateAPI(t *testing.T) {
    // Create test file
    testFile := createTestFile(t, `
        Tool{Name: "query_tools", Description: "Query tools. Default scope: project."}
        Tool{Name: "InvalidName", Description: "Bad name"}
    `)
    defer os.Remove(testFile)

    // Run validation
    report := runValidation(testFile)

    // Assert
    if report.TotalTools != 2 {
        t.Errorf("Expected 2 tools, got %d", report.TotalTools)
    }

    if report.Failed != 1 {
        t.Errorf("Expected 1 failure (naming), got %d", report.Failed)
    }
}
```

### Step 7: Document Usage

**README for Validation Tool**:
```markdown
# validate-api

Validates API tools against documented conventions.

## Installation

```bash
go install ./cmd/validate-api
```

## Usage

```bash
# Validate all checks
validate-api cmd/mcp-server/tools.go

# Run specific check
validate-api --check naming cmd/mcp-server/tools.go

# JSON output (for CI)
validate-api --format json cmd/mcp-server/tools.go

# Fast mode (skip slow checks)
validate-api --fast cmd/mcp-server/tools.go
```

## Checks

1. **Naming Convention**: snake_case tool names
2. **Parameter Ordering**: Tier-based ordering (1→2→3→4→5)
3. **Description Format**: Required patterns (e.g., "Default scope:")

## Exit Codes

- `0`: All checks passed
- `1`: One or more checks failed

## Integration

### Local Development
```bash
make validate
```

### Pre-Commit Hook
```bash
./scripts/install-hooks.sh
# Runs validation automatically on commit
```

### CI/CD
```yaml
- name: Validate API
  run: validate-api --format json cmd/mcp-server/tools.go
```
```

### Step 8: Provide Example Output (Passing and Failing)

**Passing Example**:
```
===========================================
API Validation Report
===========================================

Tools validated: 16
Checks performed: 48
✓ Passed: 48
✗ Failed: 0
⚠ Warnings: 0

All checks passed! ✓
```

**Failing Example**:
```
===========================================
API Validation Report
===========================================

Tools validated: 16
Checks performed: 48
✓ Passed: 46
✗ Failed: 2
⚠ Warnings: 0

✗ list_capabilities: Missing required pattern: 'Default scope:'
  Suggestion: Add 'Default scope: <scope>' to description
  Reference: docs/api-consistency-methodology.md
  Severity: WARNING

✗ get_capability: Missing required pattern: 'Default scope:'
  Suggestion: Add 'Default scope: <scope>' to description
  Reference: docs/api-consistency-methodology.md
  Severity: WARNING

Validation failed. Please fix the errors above.
```

### Step 9: Integrate into Development Workflow

**Makefile Target**:
```makefile
.PHONY: validate
validate:
	@echo "Validating API consistency..."
	@go run ./cmd/validate-api cmd/mcp-server/tools.go

.PHONY: validate-fast
validate-fast:
	@echo "Running fast validation checks..."
	@go run ./cmd/validate-api --fast cmd/mcp-server/tools.go
```

**Pre-Commit Hook Integration** (see agent-quality-gate-installer):
```bash
#!/bin/bash
# .git/hooks/pre-commit

if git diff --cached --name-only | grep -q "cmd/mcp-server/tools.go"; then
    echo "Validating API consistency..."
    ./validate-api --fast cmd/mcp-server/tools.go
    if [ $? -ne 0 ]; then
        echo "API validation failed. Fix errors or use --no-verify to skip."
        exit 1
    fi
fi
```

### Step 10: Continuous Improvement

**Add New Validators**:
```go
// Add new validator
type DeprecationValidator struct{}

func (v *DeprecationValidator) Check(tool Tool) ValidationResult {
    if strings.Contains(tool.Description, "DEPRECATED") {
        return ValidationResult{
            ToolName:  tool.Name,
            CheckName: "deprecation_check",
            Passed:    false,
            Message:   "Tool is deprecated",
            Severity:  "WARNING",
        }
    }
    return ValidationResult{Passed: true}
}

// Register in validator factory
func createValidators(checkFlag string) []Validator {
    validators := []Validator{
        &NamingValidator{...},
        &OrderingValidator{...},
        &DescriptionValidator{...},
        &DeprecationValidator{},  // New validator
    }
    return validators
}
```

## Output Schema

```yaml
validation_tool:
  implementation:
    files_created: [string]
    lines_of_code: number
    validators_implemented: number
    test_coverage: number

  validation_results:
    tools_validated: number
    checks_performed: number
    passed: number
    failed: number
    warnings: number

  detected_violations:
    - tool: string
      check: string
      message: string
      severity: string
      details: map[string]interface{}

  integration:
    cli_command: string
    makefile_target: string
    pre_commit_hook: boolean
    ci_integration: boolean

quality_metrics:
  false_positives: number
  false_negatives: number
  accuracy: number  # (TP + TN) / Total
```

## Success Criteria

- ✅ Deterministic validators (no ambiguity)
- ✅ 0 false positives
- ✅ Actionable error messages (specific suggestions)
- ✅ Multiple output formats (terminal + JSON)
- ✅ CLI integration with standard flags
- ✅ Test coverage ≥80%
- ✅ Documentation (usage, examples, integration)

## Example Execution (Bootstrap-006 Iteration 5)

**Input**:
```yaml
validation_target:
  format: "json_schema"
  file: "cmd/mcp-server/tools.go"
  conventions:
    - "naming_convention"
    - "parameter_ordering"
    - "description_format"

validators:
  - naming_convention
  - parameter_ordering
  - description_format
```

**Process**:
```
Step 1: Design type system
  Tool, Parameter, ValidationResult, Report

Step 2: Implement parser
  Regex-based (MVP), ~100 lines

Step 3: Create 3 validators
  NamingValidator: ~50 lines
  OrderingValidator: ~80 lines
  DescriptionValidator: ~40 lines

Step 4: Build reporter
  Terminal format: ~60 lines
  JSON format: ~20 lines

Step 5: CLI integration
  Standard flags: ~80 lines

Step 6: Test suite
  Unit tests: ~150 lines
  Coverage: 100% (naming validator)

Step 7-9: Documentation and integration
  README: ~100 lines
  Makefile targets: 2
  Pre-commit hook ready
```

**Output**:
```yaml
implementation:
  files_created: 8
  lines_of_code: ~600
  validators: 3
  test_coverage: 100% (naming validator)

validation_results:
  tools_validated: 16
  checks_performed: 48
  passed: 46
  failed: 2
  warnings: 0

violations_detected:
  - list_capabilities (missing "Default scope:")
  - get_capability (missing "Default scope:")

false_positives: 0
accuracy: 100%
```

## Pitfalls and How to Avoid

### Pitfall 1: Non-Deterministic Validators
- ❌ Wrong: Validators with judgment calls
- ✅ Right: Deterministic rules (yes/no checks)
- **Example**: "Good description" (subjective) vs. "Contains 'Default scope:'" (deterministic)

### Pitfall 2: Unclear Error Messages
- ❌ Wrong: "Description invalid"
- ✅ Right: "Missing 'Default scope:' pattern. Add 'Default scope: project' to description."
- **Benefit**: Developers know exactly how to fix

### Pitfall 3: No Test Coverage
- ❌ Wrong: Ship without tests
- ✅ Right: 80%+ test coverage, test edge cases
- **Risk**: Bugs in validators lead to false positives/negatives

### Pitfall 4: Single Output Format
- ❌ Wrong: Terminal output only
- ✅ Right: Terminal (human) + JSON (CI)
- **Integration**: CI systems need machine-readable output

### Pitfall 5: Parser Fragility
- ❌ Wrong: Regex breaks on edge cases
- ✅ Right: Use AST for production (robust)
- **Trade-off**: Regex for MVP (fast), AST for long-term (robust)

## Variations

### Variation 1: GraphQL Schema Validator

```go
type GraphQLValidator struct{}

func (v *GraphQLValidator) Check(schema GraphQLSchema) ValidationResult {
    // Check argument ordering
    // Check description format
    // Check deprecation annotations
}
```

### Variation 2: REST API Validator (OpenAPI)

```yaml
# Validate OpenAPI spec
validate-api --format openapi api-spec.yaml
```

### Variation 3: Code Style Validator

```go
// Validate Go code style
validate-api --check code-style internal/**/*.go
```

## Usage Examples

### As Subagent

```bash
/subagent @experiments/bootstrap-006-api-design/agents/agent-validation-builder.md \
  validation_target.file="cmd/mcp-server/tools.go" \
  validation_target.conventions='["naming", "ordering", "description"]' \
  tool_architecture.parser="regex" \
  tool_architecture.reporter="both"
```

### As Slash Command (if registered)

```bash
/build-validator \
  file="cmd/mcp-server/tools.go" \
  conventions="naming,ordering,description" \
  output="terminal,json"
```

## Evidence from Bootstrap-006

**Source**: Iteration 5, Task 2 (Validation Tool Implementation)

**Implementation Stats**:
- Files created: 8
- Lines of code: ~600
- Validators: 3
- Test coverage: 100% (naming validator)

**Validation Results**:
- Tools validated: 16
- Violations detected: 2
- False positives: 0
- Accuracy: 100%

**Integration**:
- CLI: ✅ Standard flags
- Makefile: ✅ 2 targets
- Pre-commit hook: ✅ Ready
- CI/CD: ✅ JSON output

---

**Last Updated**: 2025-10-16
**Status**: Validated (Bootstrap-006 Iteration 5)
**Reusability**: Universal (any API with documented conventions)
