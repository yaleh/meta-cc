# Observe Capability

## Purpose
Collect testing data, analyze test coverage, identify testing gaps, and discover testing patterns in the meta-cc codebase.

## Testing Data Collection

### Test Coverage Analysis
```bash
# Generate comprehensive coverage report
go test -cover ./... -coverprofile=coverage.out
go tool cover -func=coverage.out > coverage-summary.txt
go tool cover -html=coverage.out -o coverage.html

# Per-package coverage breakdown
go test -cover ./internal/...
go test -cover ./cmd/...
```

### Test Inventory
```bash
# Find all test files
find . -name "*_test.go" -type f

# Count test functions
grep -r "^func Test" . --include="*_test.go" | wc -l
grep -r "^func Benchmark" . --include="*_test.go" | wc -l

# Analyze test patterns
grep -r "t\.Run(" . --include="*_test.go"  # Subtests
grep -r "assert\." . --include="*_test.go"  # Assertion usage
```

### Test Execution Metrics
```bash
# Measure test execution time
go test -v ./... -count=1 | tee test-execution.log

# Test with race detection
go test -race ./...

# Benchmark existing tests
go test -bench=. -benchmem ./...
```

### Code Complexity Analysis
```bash
# Analyze function complexity (requires gocyclo)
find . -name "*.go" -not -path "./vendor/*" -not -name "*_test.go" -exec gocyclo {} \;

# Count functions per package
grep -r "^func " . --include="*.go" --exclude="*_test.go" | cut -d: -f1 | sort | uniq -c
```

## Pattern Recognition

### Testing Gaps to Identify
1. **Coverage Gaps**: Functions/packages with <80% line coverage, <70% branch coverage
2. **Critical Path Gaps**: Error handling, edge cases, boundary conditions without tests
3. **Test Type Gaps**: Missing integration tests, lack of property-based tests
4. **Test Quality Issues**: Flaky tests, slow tests (>100ms for unit tests), unclear test names
5. **Mocking Gaps**: External dependencies not properly isolated

### Testing Pattern Discovery
1. **Table-Driven Tests**: Usage patterns, effectiveness
2. **Test Fixtures**: Setup/teardown patterns, data management
3. **Subtest Organization**: Hierarchical test structure usage
4. **Error Testing**: How errors are validated, coverage of error paths
5. **Mock Usage**: Interface mocking patterns, dependency injection

## Data Sources

- Test files: `*_test.go` throughout codebase
- Coverage reports: `go test -cover` output
- Test execution logs: timing, failures, race conditions
- Code structure: function complexity, package dependencies
- Error history: 1,137 errors from bootstrap-001 baseline

## Output Format

Produce structured testing data:
- `data/test-coverage.json`: Per-package coverage metrics
- `data/test-inventory.json`: Test count, types, patterns
- `data/test-execution.json`: Timing, success rates, flakiness
- `data/coverage-gaps.json`: Functions/packages below thresholds
- `data/test-quality.json`: Test maintainability metrics
