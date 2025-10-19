# Capability: Collect Refactoring Data

## Purpose
Extract quantitative code metrics, complexity data, test coverage, and code smell patterns to establish baseline and track progress.

## When to Use
- Beginning of each iteration to measure current state
- After refactoring to measure impact
- When calculating value function components

## Inputs
- Target package/directory path
- Previous iteration metrics (if exists)

## Outputs
- Cyclomatic complexity report
- Code duplication analysis
- Static analysis results
- Test coverage metrics
- File statistics

## Procedure

### 1. Cyclomatic Complexity Analysis

```bash
# Detailed complexity per function
gocyclo -over 1 [target_path] > data/iteration-N/complexity-current.txt

# Average complexity
gocyclo -avg [target_path] >> data/iteration-N/complexity-current.txt
```

**Metrics to extract**:
- Total functions analyzed
- Average complexity score
- Functions with complexity >10 (high complexity)
- Functions with complexity >15 (very high complexity)
- Maximum complexity value and function name

### 2. Code Duplication Detection

```bash
# Find duplicated code blocks (threshold: 15 tokens)
dupl -threshold 15 [target_path] > data/iteration-N/duplication-current.txt
```

**Metrics to extract**:
- Number of duplicate block pairs
- Total lines duplicated
- Files with most duplication
- Largest duplicate block size

### 3. Static Analysis

```bash
# Staticcheck analysis
staticcheck ./[target_path]/... > data/iteration-N/staticcheck-current.txt 2>&1

# Go vet analysis
go vet ./[target_path]/... > data/iteration-N/govet-current.txt 2>&1
```

**Metrics to extract**:
- Total warnings/issues count
- Issues by category (e.g., unused variables, suspicious constructs)
- High priority issues
- Files with most issues

### 4. Test Coverage

```bash
# Overall coverage percentage
go test -cover ./[target_path]/... > data/iteration-N/coverage-current.txt

# Detailed coverage profile
go test -coverprofile=data/iteration-N/coverage.out ./[target_path]/...

# Function-level coverage breakdown
go tool cover -func=data/iteration-N/coverage.out >> data/iteration-N/coverage-current.txt
```

**Metrics to extract**:
- Overall coverage percentage
- Per-file coverage percentages
- Uncovered functions list
- Functions with partial coverage

### 5. File Statistics

```bash
# Line counts per file
find [target_path] -name "*.go" -exec wc -l {} + > data/iteration-N/file-stats.txt
```

**Metrics to extract**:
- Total lines of code
- Number of files
- Average lines per file
- Largest files

## Comparison Protocol

When comparing with baseline or previous iteration:

1. **Calculate deltas** for each metric:
   - Complexity reduction: (baseline_avg - current_avg) / baseline_avg × 100%
   - Duplication reduction: (baseline_blocks - current_blocks) / baseline_blocks × 100%
   - Warning reduction: (baseline_warnings - current_warnings) / baseline_warnings × 100%
   - Coverage improvement: current_coverage - baseline_coverage

2. **Identify regressions**:
   - Any metric worse than baseline
   - Document cause if identified

3. **Document trends**:
   - Consistent improvement direction
   - Acceleration or deceleration

## Output Format

Create `data/iteration-N/metrics-comparison.md` with:

```markdown
# Metrics Comparison - Iteration N

## Cyclomatic Complexity
- Baseline average: [value]
- Current average: [value]
- Change: [±%]
- Functions >10: [baseline count] → [current count]

## Code Duplication
- Baseline blocks: [count]
- Current blocks: [count]
- Change: [±%]
- Lines saved: [count]

## Static Analysis
- Baseline warnings: [count]
- Current warnings: [count]
- Change: [±%]

## Test Coverage
- Baseline: [%]
- Current: [%]
- Change: [±%]
- Uncovered functions: [count]

## Summary
[Brief narrative analysis of trends]
```

## Success Criteria

- All metrics collected without errors
- Data files created with complete output
- Comparison with baseline calculated (if not iteration 0)
- Metrics ready for value function calculation

## Notes

- Run from project root directory
- Ensure Go toolchain available
- Install required tools: `gocyclo`, `dupl`, `staticcheck`
- Capture both stdout and stderr for error detection
