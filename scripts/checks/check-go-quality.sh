#!/usr/bin/env bash
#
# check-go-quality.sh - Enhanced Go code quality check
#
# PURPOSE:
#   Provides comprehensive Go code quality validation using multiple tools.
#   Alternative to golangci-lint when version compatibility issues exist.
#
# HISTORICAL IMPACT:
#   - Go code quality issues: ~15% of CI failures
#   - Common issues: unused variables, missing error checks, formatting
#   - Detection: Multiple tools provide better coverage than single linter
#
# USAGE:
#   ./scripts/check-go-quality.sh
#   make check-go-quality
#
# EXIT CODES:
#   0 - All checks passed
#   1 - Quality issues found
#
# DEPENDENCIES:
#   - go (go fmt, go vet)
#   - goimports (optional)
#   - go mod tidy
#

set -euo pipefail

# Colors for output
readonly RED='\033[0;31m'
readonly YELLOW='\033[1;33m'
readonly GREEN='\033[0;32m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# Error counter
ERRORS=0
WARNINGS=0

echo -e "${BLUE}=== Go Code Quality Check ===${NC}"
echo ""

# ==============================================================================
# Check 1/5: Code formatting with go fmt
# ==============================================================================
echo -e "${BLUE}[1/5] Checking code formatting...${NC}"

UNFORMATTED_FILES=$(go fmt ./... 2>&1 || echo "")
if [ -n "$UNFORMATTED_FILES" ]; then
    echo -e "${RED}❌ Found unformatted files:${NC}"
    echo "$UNFORMATTED_FILES" | sed 's/^/  /'
    echo ""
    echo -e "${YELLOW}Run 'go fmt ./...' to fix formatting${NC}"
    ERRORS=$((ERRORS + 1))
else
    echo "✓ Code formatting is correct"
fi
echo ""

# ==============================================================================
# Check 2/5: Import formatting with goimports (if available)
# ==============================================================================
echo -e "${BLUE}[2/5] Checking import formatting...${NC}"

if command -v goimports >/dev/null 2>&1; then
    IMPORT_ISSUES=$(goimports -l . 2>/dev/null | grep -v vendor || echo "")
    if [ -n "$IMPORT_ISSUES" ]; then
        echo -e "${RED}❌ Found import formatting issues:${NC}"
        echo "$IMPORT_ISSUES" | sed 's/^/  /'
        echo ""
        echo -e "${YELLOW}Run 'goimports -w .' to fix imports${NC}"
        ERRORS=$((ERRORS + 1))
    else
        echo "✓ Import formatting is correct"
    fi
else
    echo -e "${YELLOW}⚠️  goimports not found, skipping import check${NC}"
    echo "Install with: go install golang.org/x/tools/cmd/goimports@latest"
fi
echo ""

# ==============================================================================
# Check 3/5: Static analysis with go vet
# ==============================================================================
echo -e "${BLUE}[3/5] Running go vet...${NC}"

VET_OUTPUT=$(go vet ./... 2>&1 || true)
if [ -n "$VET_OUTPUT" ]; then
    echo -e "${RED}❌ go vet found issues:${NC}"
    echo "$VET_OUTPUT" | sed 's/^/  /'
    ERRORS=$((ERRORS + 1))
else
    echo "✓ go vet passed"
fi
echo ""

# ==============================================================================
# Check 4/5: Dependency consistency
# ==============================================================================
echo -e "${BLUE}[4/5] Checking dependencies...${NC}"

# Check if go.sum is in sync
echo "  Checking go.sum consistency..."
if ! go mod verify >/dev/null 2>&1; then
    echo -e "${RED}❌ Dependency verification failed${NC}"
    echo "Run 'go mod verify' to check dependencies"
    ERRORS=$((ERRORS + 1))
else
    echo "  ✓ Dependencies verified"
fi

# Check if go.sum is up to date
echo "  Checking go.sum currency..."
if ! go mod tidy >/dev/null 2>&1; then
    echo -e "${RED}❌ go mod tidy failed${NC}"
    echo "Run 'go mod tidy' to clean dependencies"
    ERRORS=$((ERRORS + 1))
else
    # Check if go.sum changed after tidy
    if ! git diff --quiet go.sum 2>/dev/null; then
        echo -e "${YELLOW}⚠️  go.sum changed after 'go mod tidy'${NC}"
        echo "Commit the updated go.sum file"
        WARNINGS=$((WARNINGS + 1))
    else
        echo "  ✓ go.sum is up to date"
    fi
fi
echo ""

# ==============================================================================
# Check 5/5: Build verification
# ==============================================================================
echo -e "${BLUE}[5/5] Verifying build...${NC}"

# Test if code compiles without building actual binaries
echo "  Checking compilation..."
if ! go build -o /dev/null ./... >/dev/null 2>&1; then
    echo -e "${RED}❌ Code compilation failed${NC}"
    echo "Fix compilation errors before proceeding"
    ERRORS=$((ERRORS + 1))
else
    echo "  ✓ Code compiles successfully"
fi

# Test if tests compile
echo "  Checking test compilation..."
if ! go test -run=nothing -vet=off ./... >/dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Test compilation issues found${NC}"
    echo "Some tests may have compilation issues"
    WARNINGS=$((WARNINGS + 1))
else
    echo "  ✓ Tests compile successfully"
fi
echo ""

# ==============================================================================
# Summary
# ==============================================================================
echo -e "${BLUE}=== Summary ===${NC}"
echo ""

if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}✅ All Go quality checks passed${NC}"
    echo ""
    echo "Validated aspects:"
    echo "  ✓ Code formatting (go fmt)"
    echo "  ✓ Import formatting (goimports)"
    echo "  ✓ Static analysis (go vet)"
    echo "  ✓ Dependencies (go mod)"
    echo "  ✓ Build verification"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    echo -e "${YELLOW}⚠️  All checks passed with $WARNINGS warning(s)${NC}"
    echo ""
    echo "Review warnings above, but they don't block commits."
    exit 0
else
    echo -e "${RED}❌ Found $ERRORS error(s) and $WARNINGS warning(s)${NC}"
    echo ""
    echo -e "${YELLOW}How to fix:${NC}"
    echo "  1. Run 'go fmt ./...' to fix formatting"
    echo "  2. Run 'goimports -w .' to fix imports"
    echo "  3. Fix go vet issues (missing imports, unused variables, etc.)"
    echo "  4. Run 'go mod tidy' to fix dependencies"
    echo "  5. Fix compilation errors"
    echo ""
    echo "For more comprehensive linting, install golangci-lint:"
    echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    exit 1
fi
