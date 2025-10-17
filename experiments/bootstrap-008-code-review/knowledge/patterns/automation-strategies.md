# Automation Strategies for Go Code Review

**Domain**: Code Review
**Type**: Best Practices
**Created**: Iteration 2
**Status**: Validated
**Source**: Observed patterns across parser/, analyzer/, query/, validation/ modules

---

## Overview

This document outlines automation strategies to supplement manual code review, based on patterns observed during comprehensive review of the meta-cc codebase.

**Purpose**: Automate detection of common issues to:
1. Reduce manual review time (target: 5x speedup)
2. Catch issues earlier (pre-commit vs post-review)
3. Ensure consistency (automated checks don't miss patterns)
4. Free reviewers to focus on architecture and logic

---

## Strategy 1: golangci-lint Integration

### Rationale

Observed issues that golangci-lint can catch automatically:
- **Magic numbers** (found in 8+ locations across parser/, query/)
- **Variable shadowing** (QUERY-001: turn shadowing)
- **Error handling** (PARSER-003: deferred Close() without error check)
- **Unused code** (potential in private helpers)
- **Inefficient operations** (string concatenation in loops)

### Implementation

**1. Install golangci-lint**:
```bash
# macOS
brew install golangci-lint

# Linux
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Verify
golangci-lint --version
```

**2. Create `.golangci.yml` configuration**:
```yaml
# .golangci.yml
run:
  timeout: 5m
  tests: true
  modules-download-mode: readonly

linters:
  enable:
    - errcheck       # Catches unchecked errors (PARSER-003)
    - govet          # Go vet checks
    - staticcheck    # Advanced static analysis
    - gosimple       # Simplification suggestions
    - ineffassign    # Inefficient assignments
    - unused         # Unused code detection
    - gofmt          # Format checking
    - goimports      # Import organization
    - misspell       # Spelling errors
    - goconst        # Repeated constants (magic numbers)
    - godox          # TODO/FIXME/BUG comments
    - gosec          # Security issues
    - revive         # Fast linter (replaces golint)
    - stylecheck     # Go style guide enforcement

linters-settings:
  goconst:
    min-len: 2
    min-occurrences: 3  # Flag constants used 3+ times

  errcheck:
    check-type-assertions: true
    check-blank: true  # Flag _ = err

  revive:
    rules:
      - name: var-naming
        severity: warning
      - name: exported
        severity: error
      - name: error-return
        severity: error
      - name: error-strings
        severity: warning

  gosec:
    severity: medium
    excludes:
      - G104  # Duplicates errcheck
```

**3. Add to Makefile**:
```makefile
.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: lint-fix
lint-fix:
	golangci-lint run --fix ./...
```

**4. Expected Issues Caught**:
- **PARSER-003**: Deferred Close() without error check
- **QUERY-003**: Magic number 100 (goconst)
- **QUERY-001**: Variable shadowing (govet)
- **QUERY-007**: Unchecked nil parameter (staticcheck)
- **Magic numbers**: All hardcoded 100, 16, 3, 5 values (goconst)

**Impact**: Automates 20-30% of issues found in manual review

---

## Strategy 2: gosec Security Scanning

### Rationale

Security issues found during review:
- **PARSER-003**: Resource leak (deferred Close() error)
- **VALIDATION-002**: Regex injection risk
- **Potential**: File path traversal (file operations)
- **Potential**: Command injection (if adding Bash execution helpers)

### Implementation

**1. gosec is included in golangci-lint**, but can run standalone:
```bash
# Install standalone
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run
gosec ./...
```

**2. Configuration** (in `.golangci.yml` or separate `.gosec.json`):
```yaml
# Already enabled in golangci-lint config above
gosec:
  severity: medium
  confidence: medium
  excludes:
    - G104  # Duplicates errcheck
  includes:
    - G101  # Look for hardcoded credentials
    - G102  # Bind to all interfaces
    - G103  # Audit unsafe blocks
    - G201  # SQL injection
    - G202  # SQL string concatenation
    - G203  # HTML template auto-escaping
    - G301  # File permissions
    - G302  # File permissions
    - G303  # File permissions
    - G304  # File path injection
    - G305  # File traversal
    - G306  # File permissions
    - G401  # Weak crypto
    - G402  # TLS InsecureSkipVerify
    - G403  # Weak random
    - G404  # Weak random
    - G501  # Import blacklist
    - G502  # Import blacklist
    - G503  # Import blacklist
    - G504  # Import blacklist
    - G505  # Import blacklist
    - G506  # Import blacklist
```

**3. Add to CI/CD**:
```makefile
.PHONY: security
security:
	gosec -fmt=json -out=security-report.json ./...

.PHONY: security-check
security-check:
	gosec -quiet ./...
```

**Expected Issues Caught**:
- **VALIDATION-002**: Regex injection (G201/G202)
- **File operations**: Path traversal risks (G304/G305)
- **Resource leaks**: File handle management (G307)

**Impact**: Automates 10-15% of security-related issues

---

## Strategy 3: Pre-Commit Hooks

### Rationale

Catch issues BEFORE commit, reducing review iterations:
- **Format violations** (gofmt)
- **Import organization** (goimports)
- **Linting errors** (golangci-lint)
- **Test failures** (go test)
- **Security issues** (gosec critical only)

### Implementation

**1. Install pre-commit framework**:
```bash
# macOS
brew install pre-commit

# Linux
pip install pre-commit

# Verify
pre-commit --version
```

**2. Create `.pre-commit-config.yaml`**:
```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/golangci/golangci-lint
    rev: v1.54.2
    hooks:
      - id: golangci-lint
        args: [--fast]  # Fast mode for pre-commit

  - repo: https://github.com/dnephin/pre-commit-golang
    rev: v0.5.1
    hooks:
      - id: go-fmt
        name: gofmt
        description: Format Go code

      - id: go-imports
        name: goimports
        description: Organize Go imports

      - id: go-vet
        name: go vet
        description: Run go vet

      - id: go-test-repo
        name: go test
        description: Run tests
        args: [-short]  # Only short tests in pre-commit

      - id: go-mod-tidy
        name: go mod tidy
        description: Clean go.mod and go.sum

  - repo: local
    hooks:
      - id: gosec-critical
        name: gosec (critical only)
        entry: gosec -severity=high -quiet
        language: system
        pass_filenames: false
        always_run: true
```

**3. Install hooks**:
```bash
# In project root
pre-commit install

# Test on all files
pre-commit run --all-files

# Update hooks
pre-commit autoupdate
```

**4. Add to developer onboarding**:
```bash
# scripts/install-hooks.sh
#!/bin/bash
set -e

echo "Installing pre-commit hooks..."
pre-commit install

echo "Running pre-commit on all files..."
pre-commit run --all-files

echo "✓ Pre-commit hooks installed successfully"
echo "Hooks will run automatically on git commit"
```

**Expected Benefits**:
- Catches format/lint issues before commit (saves review time)
- Prevents broken tests from being committed
- Reduces commit→review→fix→resubmit cycles

**Impact**: 30-40% reduction in review iterations

---

## Strategy 4: Test Coverage Enforcement

### Rationale

**CRITICAL finding**: validation/ module has 32.5% test coverage (target: 80%)

Issues that comprehensive tests would catch:
- **VALIDATION-005**: Broken order validation logic
- **VALIDATION-006**: Random parameter order from Go maps
- **VALIDATION-008**: splitLines doesn't split
- **QUERY-002**: Silent error returns
- **Edge cases**: Empty inputs, nil parameters, invalid timestamps

### Implementation

**1. Add coverage checking to Makefile**:
```makefile
.PHONY: test-coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

.PHONY: test-coverage-html
test-coverage-html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html  # macOS
	# Or: xdg-open coverage.html  # Linux

.PHONY: test-coverage-check
test-coverage-check:
	@echo "Checking test coverage..."
	@go test -coverprofile=coverage.out ./... > /dev/null
	@COVERAGE=$$(go tool cover -func=coverage.out | tail -1 | awk '{print $$3}' | sed 's/%//'); \
	if [ "$$(echo "$$COVERAGE < 80" | bc)" -eq 1 ]; then \
		echo "FAIL: Coverage $$COVERAGE% is below 80% target"; \
		exit 1; \
	else \
		echo "PASS: Coverage $$COVERAGE% meets 80% target"; \
	fi
```

**2. Add per-package coverage check**:
```makefile
.PHONY: test-coverage-by-package
test-coverage-by-package:
	@for pkg in $$(go list ./internal/...); do \
		coverage=$$(go test -coverprofile=/tmp/coverage.out $$pkg 2>/dev/null | grep coverage | awk '{print $$5}' | sed 's/%//'); \
		if [ ! -z "$$coverage" ]; then \
			printf "%-50s %6s%%\n" "$$pkg" "$$coverage"; \
		fi; \
	done
```

**3. Add to CI/CD pipeline** (GitHub Actions example):
```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run tests with coverage
        run: |
          go test -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out > coverage.txt

      - name: Check coverage threshold
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
          echo "Coverage: ${COVERAGE}%"
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "ERROR: Coverage ${COVERAGE}% is below 80% target"
            exit 1
          fi

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

**4. Enforce in code review checklist**:
- [ ] New code has ≥80% test coverage
- [ ] Edge cases are tested (nil, empty, invalid inputs)
- [ ] Error paths are tested (all error returns exercised)
- [ ] `make test-coverage-check` passes

**Expected Impact**:
- Catches logic errors like VALIDATION-005, VALIDATION-008
- Prevents regressions when refactoring
- Validates edge case handling

**Impact**: Would have caught 6-8 issues found in validation/ review (40%+ of issues)

---

## Strategy 5: Static Analysis for Common Patterns

### Rationale

Patterns observed across multiple modules:
- **O(n*m) iteration**: Found in analyzer/, query/ (ANALYZER-016, QUERY-005)
- **Error return ambiguity**: Return 0 on error (parser/, query/)
- **Code duplication**: buildContextBefore/After, similar iteration patterns
- **Missing godoc**: 30+ functions lack documentation

### Implementation

**1. Custom linter using go/ast**:

Create `tools/lint/itercheck/itercheck.go`:
```go
// Package itercheck detects O(n*m) nested iterations
package itercheck

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "itercheck",
	Doc:  "detects O(n*m) nested iterations over same data",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := func(node ast.Node) bool {
		rangeStmt, ok := node.(*ast.RangeStmt)
		if !ok {
			return true
		}

		// Check for nested range over same collection
		ast.Inspect(rangeStmt.Body, func(inner ast.Node) bool {
			innerRange, ok := inner.(*ast.RangeStmt)
			if !ok {
				return true
			}

			// Heuristic: if both range over similar names, flag
			if isSimilarCollection(rangeStmt.X, innerRange.X) {
				pass.Reportf(innerRange.Pos(),
					"nested iteration may have O(n*m) complexity")
			}
			return true
		})

		return true
	}

	for _, file := range pass.Files {
		ast.Inspect(file, inspect)
	}

	return nil, nil
}

func isSimilarCollection(outer, inner ast.Expr) bool {
	// Implementation: compare collection names
	// ...
}
```

**2. Add to golangci-lint**:
```yaml
# .golangci.yml
linters-settings:
  custom:
    itercheck:
      path: ./tools/lint/itercheck
      description: Detects O(n*m) nested iterations
      original-url: github.com/yaleh/meta-cc/tools/lint/itercheck
```

**3. Or use existing linters**:
- **gocyclo**: Detects high cyclomatic complexity
- **gocognit**: Cognitive complexity metric
- **dupl**: Code duplication detection
- **godot**: godoc comment formatting

**Expected Issues Caught**:
- **QUERY-005**: O(n*m) in calculateSequenceTimeSpan
- **ANALYZER-016**: O(n*m) in DetectFileChurn
- **Code duplication**: Similar iteration patterns

**Impact**: Automates 15-20% of performance and maintainability issues

---

## Strategy 6: Continuous Integration Quality Gates

### Rationale

Ensure all automation runs on every PR, blocking merge if issues found.

### Implementation

**GitHub Actions workflow** (`.github/workflows/ci.yml`):
```yaml
name: CI

on:
  pull_request:
    branches: [main, develop]
  push:
    branches: [main, develop]

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: v1.54
          args: --timeout=5m

  security:
    name: Security Scan
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run Gosec
        uses: securego/gosec@master
        with:
          args: '-severity=medium ./...'

  test:
    name: Tests and Coverage
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Check coverage
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "ERROR: Coverage ${COVERAGE}% < 80%"
            exit 1
          fi

      - name: Upload coverage
        uses: codecov/codecov-action@v3

  build:
    name: Build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Build
        run: make build

      - name: Verify no changes
        run: git diff --exit-code  # Fail if go.mod changed
```

**Branch protection rules** (GitHub settings):
```yaml
# .github/BRANCH_PROTECTION.md (documentation)
Required checks for merge to main:
- ✓ Lint (golangci-lint)
- ✓ Security Scan (gosec)
- ✓ Tests and Coverage (≥80%)
- ✓ Build (successful compilation)

Recommended:
- Require approvals: 1 reviewer
- Dismiss stale reviews: true
- Require review from code owners: true
```

---

## Automation Effectiveness Matrix

| Strategy | Issues Caught | Implementation Time | Maintenance | ROI |
|----------|---------------|---------------------|-------------|-----|
| golangci-lint | 20-30% | 2 hours | Low | Very High |
| gosec | 10-15% | 1 hour | Low | High |
| Pre-commit hooks | 30-40% reduction in iterations | 2 hours | Low | Very High |
| Test coverage enforcement | 40%+ logic errors | 4 hours (initial) | Medium | Very High |
| Custom static analysis | 15-20% | 8-16 hours | High | Medium |
| CI/CD gates | Blocks all above from merging | 4 hours | Low | Very High |

**Total estimated automation**: 50-60% of issues found in manual review

**Remaining for manual review**:
- Architecture and design decisions
- Complex logic correctness
- API design and usability
- Performance bottlenecks (beyond obvious O(n*m))
- Domain-specific patterns

---

## Prioritized Implementation Roadmap

### Phase 1: Quick Wins (Week 1)
1. Install golangci-lint with .golangci.yml config (2 hours)
2. Add `make lint` to Makefile (30 min)
3. Run golangci-lint on codebase, fix issues (4-8 hours)
4. Add to CI/CD pipeline (2 hours)

**Expected Impact**: Catch 20-30% of issues automatically

### Phase 2: Security and Pre-Commit (Week 2)
1. Configure gosec in golangci-lint (30 min)
2. Install pre-commit hooks (2 hours)
3. Document hook installation for team (1 hour)
4. Add security scan to CI/CD (1 hour)

**Expected Impact**: +10-15% security issues, 30-40% fewer review iterations

### Phase 3: Test Coverage Enforcement (Week 3)
1. Add test-coverage-check to Makefile (1 hour)
2. Measure current coverage per package (1 hour)
3. Create coverage increase plan for <80% packages (2 hours)
4. Add coverage gate to CI/CD (1 hour)
5. Write tests to reach 80% (validation/: 8-16 hours)

**Expected Impact**: +40% logic error detection, prevent regressions

### Phase 4: Advanced Analysis (Week 4+)
1. Research custom linters for project-specific patterns (4 hours)
2. Implement O(n*m) iteration detector (8 hours)
3. Add to golangci-lint custom linters (2 hours)
4. Document maintenance procedures (2 hours)

**Expected Impact**: +15-20% performance/maintainability issues

---

## Success Metrics

Track automation effectiveness:

**Before Automation** (Baseline):
- Manual review time: ~2.45 hours per 1K lines
- Issues found: 42 in 1,224 lines (3.4%)
- False positives: 0%
- Review iterations: Average 2-3 per module

**After Automation** (Target):
- Manual review time: <0.5 hours per 1K lines (5x speedup)
- Issues caught pre-review: 50-60%
- Remaining manual issues: 40-50% (architecture, complex logic)
- Review iterations: Average 1-1.5 per module (40% reduction)

**Measurement**:
1. Track issues by detection method (automated vs manual)
2. Measure review time before/after automation
3. Count review iteration cycles
4. Monitor false positive rate (target: <5%)

---

## Maintenance

**Weekly**:
- Update golangci-lint: `golangci-lint cache clean && golangci-lint --version`
- Review new issues flagged by linters
- Update .golangci.yml if new linters become available

**Monthly**:
- Update pre-commit hooks: `pre-commit autoupdate`
- Review and update CI/CD workflows
- Analyze automation effectiveness metrics

**Quarterly**:
- Review custom linter effectiveness
- Update security scan configuration
- Reassess coverage targets based on codebase growth

---

## References

- **golangci-lint**: https://golangci-lint.run/
- **gosec**: https://github.com/securego/gosec
- **pre-commit**: https://pre-commit.com/
- **Go testing**: https://go.dev/doc/code#Testing
- **Go static analysis**: https://pkg.go.dev/golang.org/x/tools/go/analysis

---

**Status**: Validated | **Iteration**: 2 | **Last Updated**: 2025-10-17
