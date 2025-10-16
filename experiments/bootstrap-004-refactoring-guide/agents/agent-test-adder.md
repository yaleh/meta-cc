# Agent: Incremental Test Addition

**Version**: 1.0
**Source**: Bootstrap-004, Pattern 4
**Success Rate**: 0% → 32.5% coverage improvement in meta-cc Iteration 2

---

## Role

Systematically add tests to low-coverage packages (<50%), focusing on exported functions and measuring improvement incrementally.

## When to Use

- Package has low test coverage (<50%)
- Need to improve testing systematically
- Before major refactoring (need safety net)
- Test improvement sprints or TDD adoption

## Input Schema

```yaml
target_package:
  path: string                  # Required: Package to test
  current_coverage: number      # Optional: Current % (will measure if not provided)

target_metrics:
  target_coverage: number       # Default: 0.75 (75%)
  min_improvement: number       # Default: 0.10 (10 percentage points)

test_strategy:
  focus: string                 # "exported" | "all" | "critical"
                                # exported: Only exported functions (recommended)
                                # all: All functions
                                # critical: Functions with highest risk/complexity

  test_types: [string]          # ["success", "failure", "edge"]
                                # success: Happy path tests
                                # failure: Error condition tests
                                # edge: Boundary and edge cases

  table_driven: boolean         # Default: true (use table-driven tests)
```

## Execution Process

### Step 1: Identify Low-Coverage Packages

```bash
# Go
go test -cover ./... | grep -E "coverage: [0-4][0-9]%"

# Example output:
# ok    internal/validation  0.005s  coverage: 0.0% of statements
# ok    internal/cache       0.003s  coverage: 42.3% of statements

# Python
pytest --cov=. --cov-report=term-missing | grep -E "[0-4][0-9]%"

# JavaScript
jest --coverage | grep -E "[0-4][0-9]%"
```

**Selection Criteria**:
- Coverage <50% → High priority
- Coverage 50-80% → Medium priority (skip if time-constrained)
- Coverage >80% → Low priority (adequate)

### Step 2: Select Target Package

**Prioritization**:
```python
def prioritize_packages(packages):
    scored = []
    for pkg in packages:
        score = (
            0.4 * (1 - pkg.coverage) +        # Lower coverage = higher score
            0.3 * pkg.complexity +            # Higher complexity = needs more tests
            0.3 * pkg.change_frequency        # Frequently changed = needs tests
        )
        scored.append((pkg, score))

    return sorted(scored, key=lambda x: x[1], reverse=True)
```

**Example**:
```yaml
selected_package:
  path: "internal/validation"
  current_coverage: 0.0%
  reason: "No tests, medium complexity, 12 exported functions"
```

### Step 3: List Exported Functions

```bash
# Go: List exported functions (start with capital letter)
grep "^func [A-Z]" internal/validation/*.go

# Example output:
# description.go:func ValidateToolDescription(desc string) error
# ordering.go:func OrderParameters(params []Parameter) []Parameter
# naming.go:func ValidateToolName(name string) error

# Python: List public functions (no leading underscore)
grep "^def [^_]" mypackage/*.py

# JavaScript: List exported functions
grep "^export function" src/*.js
```

**Output**:
```yaml
exported_functions:
  - file: "description.go"
    function: "ValidateToolDescription"
    signature: "func ValidateToolDescription(desc string) error"

  - file: "ordering.go"
    function: "OrderParameters"
    signature: "func OrderParameters(params []Parameter) []Parameter"

  # ... 10 more functions
```

### Step 4: Create Test File

**Naming Convention**:
```bash
# Go
internal/validation/description_test.go

# Python
tests/test_description.py

# JavaScript
src/__tests__/description.test.js
```

**Template** (Go example):
```go
// internal/validation/description_test.go
package validation

import "testing"

// Test functions will be added below
```

### Step 5: Write Test for First Function (Success Case)

**Start with happy path**:

```go
func TestValidateToolDescription(t *testing.T) {
    // Arrange
    validDesc := "Query tool calls with filters. Default scope: project."

    // Act
    err := ValidateToolDescription(validDesc)

    // Assert
    if err != nil {
        t.Errorf("Expected no error, got: %v", err)
    }
}
```

### Step 6: Run Test, Verify Passes

```bash
go test ./internal/validation -v

# Expected output:
# === RUN   TestValidateToolDescription
# --- PASS: TestValidateToolDescription (0.00s)
# PASS
# ok      internal/validation    0.005s
```

**If test fails**: Fix implementation or test, then proceed.

### Step 7: Add Failure Case Test

```go
func TestValidateToolDescription_TooLong(t *testing.T) {
    // Arrange
    longDesc := "This is a very long description that exceeds the maximum " +
                "allowed length of 100 characters for tool descriptions in MCP"

    // Act
    err := ValidateToolDescription(longDesc)

    // Assert
    if err == nil {
        t.Error("Expected error for too-long description, got nil")
    }
}
```

### Step 8: Add Edge Case Tests

**Edge cases to consider**:
- Empty string
- Nil values (if applicable)
- Boundary values (exactly at limit)
- Special characters
- Unicode/internationalization

```go
func TestValidateToolDescription_EdgeCases(t *testing.T) {
    tests := []struct {
        name        string
        description string
        wantErr     bool
    }{
        {
            name:        "empty string",
            description: "",
            wantErr:     true,
        },
        {
            name:        "exactly 100 characters",
            description: strings.Repeat("a", 100),
            wantErr:     false,
        },
        {
            name:        "101 characters (boundary)",
            description: strings.Repeat("a", 101),
            wantErr:     true,
        },
        {
            name:        "special characters",
            description: "Query tools with filters: scope, status! Default: project.",
            wantErr:     false,
        },
        {
            name:        "unicode",
            description: "查询工具调用 with filters. Default scope: project.",
            wantErr:     false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateToolDescription(tt.description)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateToolDescription() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Step 9: Repeat for Remaining Functions

**Process**: One function at a time

```
Function 1: ValidateToolDescription
  ✅ Success case test
  ✅ Failure case test
  ✅ Edge case tests
  ✅ Run: go test -v

Function 2: OrderParameters
  ✅ Success case test
  ✅ Failure case test
  ✅ Edge case tests
  ✅ Run: go test -v

# ... repeat for all 10 functions
```

**Commit Strategy**:
```bash
# After each 2-3 functions
git add internal/validation/*_test.go
git commit -m "test: add tests for ValidateToolDescription and OrderParameters"
```

### Step 10: Measure Coverage Improvement

```bash
# Before adding tests
go test -cover ./internal/validation
# coverage: 0.0% of statements

# After adding tests
go test -cover ./internal/validation
# coverage: 32.5% of statements

# Improvement: +32.5 percentage points
```

**Detailed Coverage Report**:
```bash
go test -coverprofile=coverage.out ./internal/validation
go tool cover -func=coverage.out

# Output:
# internal/validation/description.go:25:  ValidateToolDescription  85.7%
# internal/validation/ordering.go:42:    OrderParameters          90.0%
# internal/validation/naming.go:18:      ValidateToolName         0.0%
# ...
# total:                                  (statements)             32.5%
```

### Step 11: Verify All Tests Pass

```bash
# Run full test suite
go test ./...

# Expected: All tests pass
# ok    internal/validation  0.008s  coverage: 32.5%
# ok    internal/parser      0.012s  coverage: 78.3%
# ...
```

**If any tests fail**: Investigate and fix before proceeding.

### Step 12: Document Test Coverage

**Update README or docs**:
```markdown
## Test Coverage

| Package | Coverage | Status |
|---------|----------|--------|
| internal/validation | 32.5% | ⬆️ Improved from 0% |
| internal/parser | 78.3% | ✅ Good |
| internal/cache | 42.3% | ⚠️ Needs improvement |
```

**Metrics Dashboard** (if exists):
```json
{
  "package": "internal/validation",
  "coverage_before": 0.0,
  "coverage_after": 32.5,
  "improvement": 32.5,
  "tests_added": 10,
  "lines_tested": 45,
  "date": "2025-10-16"
}
```

## Output Schema

```yaml
test_results:
  package: string
  coverage_before: number
  coverage_after: number
  improvement: number        # Percentage points
  improvement_percentage: number  # Relative improvement

tests_added:
  count: number
  by_type:
    success_cases: number
    failure_cases: number
    edge_cases: number

  by_file:
    - file: string
      tests_added: number
      coverage: number

functions_tested:
  total: number
  covered: [string]
  not_covered: [string]

test_status:
  all_pass: boolean
  pass_count: number
  fail_count: number

time_spent:
  total_hours: number
  per_function_avg: number

quality_metrics:
  table_driven_percentage: number
  edge_cases_per_function: number
```

## Success Criteria

- ✅ Coverage improvement ≥ 10 percentage points
- ✅ All tests pass (100% pass rate)
- ✅ At least 50% of exported functions tested
- ✅ Each function has success + failure + edge cases
- ✅ Table-driven tests used where applicable

## Example Execution (meta-cc Iteration 2)

**Input**:
```yaml
target_package:
  path: "internal/validation"
  current_coverage: 0.0%

target_metrics:
  target_coverage: 0.75
  min_improvement: 0.10

test_strategy:
  focus: "exported"
  test_types: ["success", "failure", "edge"]
  table_driven: true
```

**Process**:
```
Step 1: Identify low-coverage packages
  Found: internal/validation (0% coverage)

Step 2: Select target package
  Selected: internal/validation (10 exported functions)

Step 3: List exported functions
  description.go: ValidateToolDescription
  ordering.go: OrderParameters
  naming.go: ValidateToolName
  ... (7 more)

Step 4: Create test files
  Created: description_test.go, ordering_test.go

Step 5-8: Write tests for description.go
  Success case: ✅
  Failure case: ✅
  Edge cases (5): ✅

Step 9: Repeat for ordering.go
  10 test functions: ✅

Step 10: Measure coverage
  Before: 0%
  After: 32.5%
  Improvement: +32.5 percentage points

Step 11: Verify all tests pass
  Result: ✅ All pass
```

**Output**:
```yaml
coverage_before: 0.0%
coverage_after: 32.5%
improvement: 32.5 percentage points

tests_added:
  count: 18
  success_cases: 8
  failure_cases: 5
  edge_cases: 5

time_spent: 3 hours

files:
  - file: "description_test.go"
    tests: 8
    coverage: 85%

  - file: "ordering_test.go"
    tests: 10
    coverage: 90%
```

## Pitfalls and How to Avoid

### Pitfall 1: Testing Internal Implementation
- ❌ Wrong: Test private functions, internal state
- ✅ Right: Test exported API, observable behavior

### Pitfall 2: Unfocused Effort
- ❌ Wrong: Write one test each for 10 packages
- ✅ Right: Improve one package from 0% to 30%+

### Pitfall 3: Not Measuring Coverage
- ❌ Wrong: "I wrote tests, coverage probably improved"
- ✅ Right: Measure before/after, document improvement

### Pitfall 4: Writing Tests That Don't Fail
- ❌ Wrong: Tests always pass (testing nothing)
- ✅ Right: Temporarily break code, verify test fails, fix

### Pitfall 5: Ignoring Maintainability
- ❌ Wrong: Copy-paste tests with minor variations
- ✅ Right: Use table-driven tests, shared fixtures

## Variations

### Variation 1: TDD (Test-First)
**When**: Writing new code
```
1. Write test first (it fails)
2. Write minimal code to pass test
3. Refactor
4. Repeat
```

### Variation 2: Coverage-Driven (Target-Based)
**When**: Specific coverage target (e.g., 75%)
```
1. Measure current coverage
2. Identify gap (75% - current%)
3. Add tests until gap closed
4. Stop (avoid over-testing)
```

### Variation 3: Risk-Driven (High-Risk First)
**When**: Limited time, prioritize critical functions
```
Prioritize tests for:
- Functions with most bugs historically
- Functions with highest complexity
- Functions in critical path
```

### Variation 4: Integration-First
**When**: Unit coverage high, integration low
```
1. Write integration tests first
2. Add unit tests for failures found
3. Measure both unit and integration coverage
```

## Language-Specific Adaptations

### Go
```bash
# Identify low-coverage
go test -cover ./... | grep -E "[0-4][0-9]%"

# List exported functions
grep "^func [A-Z]" internal/validation/*.go

# Create test file
touch internal/validation/description_test.go

# Run tests
go test -v ./internal/validation

# Measure coverage
go test -cover ./internal/validation
go test -coverprofile=coverage.out ./internal/validation
go tool cover -html=coverage.out
```

### Python (pytest)
```bash
# Identify low-coverage
pytest --cov=. --cov-report=term-missing | grep -E "[0-4][0-9]%"

# List public functions
grep "^def [^_]" mypackage/*.py

# Create test file
touch tests/test_validation.py

# Run tests
pytest tests/test_validation.py -v

# Measure coverage
pytest --cov=mypackage/validation --cov-report=term-missing
```

### JavaScript (Jest)
```bash
# Identify low-coverage
jest --coverage | grep -E "[0-4][0-9]%"

# List exported functions
grep "^export function" src/*.js

# Create test file
touch src/__tests__/validation.test.js

# Run tests
jest validation.test.js

# Measure coverage
jest --coverage --collectCoverageFrom=src/validation.js
```

## Usage Examples

### As Subagent
```bash
/subagent @experiments/bootstrap-004-refactoring-guide/agents/agent-test-adder.md \
  target_package.path="internal/validation" \
  target_metrics.target_coverage=0.75 \
  test_strategy.focus="exported"
```

### As Slash Command (if registered)
```bash
/add-tests \
  package="internal/validation" \
  target=75 \
  focus="exported"
```

## Evidence from Bootstrap-004

**Source**: meta-cc Iteration 2

**Target Package**: internal/validation
- Initial coverage: 0%
- Final coverage: 32.5%
- Improvement: +32.5 percentage points

**Metrics**:
- Tests added: 18 (2 files, 10 functions)
- Lines of test code: ~300
- Time spent: ~3 hours
- Test pass rate: 100%

**Coverage Breakdown**:
- description.go: 85% (8 test cases)
- ordering.go: 90% (10 test functions)
- naming.go: 0% (deferred to future work)

**Quality**:
- Table-driven tests: 100%
- Edge cases per function: avg 2.5
- All tests pass: ✅

---

**Last Updated**: 2025-10-16
**Status**: Validated (meta-cc Iteration 2)
**Reusability**: Universal (any language with test coverage tools)
