# Build Quality Gates - Implementation Patterns

This document captures the key patterns and practices discovered during the BAIME build-quality-gates experiment.

## Three-Layer Architecture Pattern

### P0: Critical Checks (Pre-commit)
**Purpose**: Block commits that would definitely fail CI
**Target**: <10 seconds, 50-70% error coverage
**Examples**: Temporary files, dependency issues, test fixtures

```makefile
check-workspace: check-temp-files check-fixtures check-deps
	@echo "✅ Workspace validation passed"
```

### P1: Enhanced Checks (Quality Assurance)
**Purpose**: Ensure code quality and team standards
**Target**: <30 seconds, 80-90% error coverage
**Examples**: Script validation, import formatting, debug statements

```makefile
check-quality: check-workspace check-scripts check-imports check-debug
	@echo "✅ Quality validation passed"
```

### P2: Advanced Checks (Comprehensive)
**Purpose**: Full validation for important changes
**Target**: <60 seconds, 95%+ error coverage
**Examples**: Language-specific quality, security, performance

```makefile
check-full: check-quality check-security check-performance
	@echo "✅ Comprehensive validation passed"
```

## Script Structure Pattern

### Standard Template
```bash
#!/bin/bash
# check-[category].sh - [One-line description]
#
# Part of: Build Quality Gates
# Iteration: [P0/P1/P2]
# Purpose: [What problems this prevents]
# Historical Impact: [X% of errors this catches]

set -euo pipefail

# Colors for consistent output
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m'

echo "Checking [category]..."

ERRORS=0
# ... check logic ...

# Summary
if [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}✅ All [category] checks passed${NC}"
    exit 0
else
    echo -e "${RED}❌ Found $ERRORS [category] issue(s)${NC}"
    exit 1
fi
```

## Error Message Pattern

### Clear, Actionable Messages
```
❌ ERROR: Temporary test/debug scripts found:
  - ./test_parser.go
  - ./debug_analyzer.go

Action: Delete these temporary files before committing

To fix:
  1. Delete temporary files: rm test_*.go debug_*.go
  2. Move legitimate files to appropriate packages
  3. Run again: make check-temp-files
```

### Message Components
1. **Clear problem statement** in red
2. **Specific items found** with paths
3. **Required action** clearly stated
4. **Step-by-step fix instructions**
5. **Verification command** to re-run

## Performance Optimization Patterns

### Parallel Execution
```makefile
check-parallel:
	@make check-temp-files & \
	make check-fixtures & \
	make check-deps & \
	wait
	@echo "✅ Parallel checks completed"
```

### Incremental Checking
```bash
check-incremental:
	@if [ -n "$(git status --porcelain)" ]; then
		CHANGED=$$(git diff --name-only --cached);
		echo "Checking changed files: $$CHANGED";
		# Run checks only on changed files
	else
		$(MAKE) check-workspace
	fi
```

### Caching Strategy
```bash
# Use Go test cache
go test -short ./...

# Cache expensive operations
CACHE_DIR=.cache/check-deps
if [ ! -f "$CACHE_DIR/verified" ]; then
    go mod verify
    touch "$CACHE_DIR/verified"
fi
```

## Integration Patterns

### Makefile Structure
```makefile
# =============================================================================
# Build Quality Gates - Three-Layer Architecture
# =============================================================================

# P0: Critical checks (must pass before commit)
check-workspace: check-temp-files check-fixtures check-deps
	@echo "✅ Workspace validation passed"

# P1: Enhanced checks (quality assurance)
check-quality: check-workspace check-scripts check-imports check-debug
	@echo "✅ Quality validation passed"

# P2: Advanced checks (comprehensive validation)
check-full: check-quality check-security check-performance
	@echo "✅ Comprehensive validation passed"

# =============================================================================
# Workflow Targets
# =============================================================================

# Development iteration (fastest)
dev: fmt build
	@echo "✅ Development build complete"

# Pre-commit validation (recommended)
pre-commit: check-workspace test-short
	@echo "✅ Pre-commit checks passed"

# Full validation (before important commits)
all: check-quality test-full build-all
	@echo "✅ Full validation passed"

# CI-level validation
ci: check-full test-all build-all verify
	@echo "✅ CI validation passed"
```

### CI/CD Integration Pattern
```yaml
# GitHub Actions
- name: Run quality gates
  run: make ci

# GitLab CI
script:
  - make ci

# Jenkins
sh 'make ci'
```

## Tool Chain Management Patterns

### Version Consistency
```bash
# Pin versions in configuration
.golangci.yml:  version: "1.64.8"
.tool-versions: golangci-lint 1.64.8
```

### Docker-based Toolchains
```dockerfile
FROM golang:1.21.0
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
RUN go install golang.org/x/tools/cmd/goimports@latest
```

### Cross-Platform Compatibility
```bash
# Use portable tools
find . -name "*.go" # instead of platform-specific tools
grep -r "TODO" .     # instead of IDE-specific search
```

## Quality Metrics Patterns

### V_instance Calculation
```bash
V_instance=$(echo "scale=3;
  0.4 * (1 - $ci_failure_rate) +
  0.3 * (1 - $avg_iterations/$baseline_iterations) +
  0.2 * ($baseline_time/$detection_time/10) +
  0.1 * $error_coverage" | bc)
```

### Metrics Collection
```bash
# Automated metrics collection
collect_metrics() {
    local ci_failure_rate=$(get_ci_failure_rate)
    local detection_time=$(measure_detection_time)
    local error_coverage=$(calculate_error_coverage)
    # Calculate and store metrics
}
```

### Trend Monitoring
```python
# Plot quality trends over time
def plot_metrics_trend(metrics_data):
    # Visualize V_instance and V_meta improvement
    # Show convergence toward targets
    pass
```

## Error Handling Patterns

### Graceful Degradation
```bash
# Continue checking even if one check fails
ERRORS=0
check_temp_files || ERRORS=$((ERRORS + 1))
check_fixtures   || ERRORS=$((ERRORS + 1))

if [ $ERRORS -gt 0 ]; then
    echo "Found $ERRORS issues"
    exit 1
fi
```

### Tool Availability
```bash
# Handle missing optional tools
if command -v goimports >/dev/null; then
    goimports -l .
else
    echo "⚠️ goimports not available, skipping import check"
fi
```

### Clear Exit Codes
```bash
# 0: Success
# 1: Errors found
# 2: Configuration issues
# 3: Tool not available
```

## Team Adoption Patterns

### Gradual Enforcement
```bash
# Start with warnings
if [ "${ENFORCE_QUALITY:-false}" = "true" ]; then
    make check-workspace-strict
else
    make check-workspace-warning
fi
```

### Easy Fix Commands
```bash
# Provide one-command fixes
fix-imports:
	@echo "Fixing imports..."
	@goimports -w .
	@echo "✅ Imports fixed"

fix-temp-files:
	@echo "Removing temporary files..."
	@rm -f test_*.go debug_*.go
	@echo "✅ Temporary files removed"
```

### Documentation Integration
```bash
# Link to documentation in error messages
echo "See: docs/guides/build-quality-gates.md#temporary-files"
```

## Maintenance Patterns

### Regular Updates
```bash
# Monthly tool updates
update-quality-tools:
	@echo "Updating quality gate tools..."
	@go install -a github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@make check-full && echo "✅ Tools updated successfully"
```

### Performance Monitoring
```bash
# Benchmark performance regularly
benchmark-quality-gates:
	@for i in {1..10}; do
		time make check-full 2>&1 | grep real
	done
```

### Feedback Collection
```bash
# Collect team feedback
collect-quality-feedback:
	@echo "Please share your experience with quality gates:"
	@echo "1. What's working well?"
	@echo "2. What's frustrating?"
	@echo "3. Suggested improvements?"
```

## Anti-Patterns to Avoid

### ❌ Don't Do This
```bash
# Too strict - blocks legitimate work
if [ -n "$(git status --porcelain)" ]; then
    echo "Working directory must be clean"
    exit 1
fi

# Too slow - developers won't use it
make check-slow-heavy-analysis  # Takes 5+ minutes

# Unclear errors - developers don't know how to fix
echo "❌ Code quality issues found"
exit 1
```

### ✅ Do This Instead
```bash
# Flexible - allows legitimate work
if [ -n "$(find . -name "*.tmp")" ]; then
    echo "❌ Temporary files found"
    echo "Remove: find . -name '*.tmp' -delete"
fi

# Fast - developers actually use it
make check-quick-essentials  # <30 seconds

# Clear errors - developers can fix immediately
echo "❌ Import formatting issues in:"
echo "  - internal/parser.go"
echo "Fix: goimports -w ."
```
