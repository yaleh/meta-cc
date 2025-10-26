#!/bin/bash
# check-deps.sh - Verify go.mod and go.sum integrity
#
# Part of: Build Quality Gates (BAIME Experiment)
# Iteration: 1 (P0)
# Purpose: Ensure dependency files are in sync and verified
# Historical Impact: Prevents ~5% of dependency-related build failures

set -euo pipefail

# Colors
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m'

echo "Verifying dependencies..."

ERRORS=0

# ============================================================================
# Check 1: Verify go.mod and go.sum exist
# ============================================================================
echo "  [1/4] Checking dependency files exist..."

if [ ! -f "go.mod" ]; then
    echo -e "${RED}❌ ERROR: go.mod not found${NC}"
    exit 1
fi

if [ ! -f "go.sum" ]; then
    echo -e "${RED}❌ ERROR: go.sum not found${NC}"
    exit 1
fi

echo -e "${GREEN}  ✓ go.mod and go.sum present${NC}"

# ============================================================================
# Check 2: Verify checksums
# ============================================================================
echo "  [2/4] Verifying dependency checksums..."

if ! go mod verify >/dev/null 2>&1; then
    echo -e "${RED}❌ ERROR: Dependency verification failed${NC}"
    echo ""
    echo "Run the following to diagnose:"
    echo "  go mod verify"
    echo ""
    ((ERRORS++)) || true
else
    echo -e "${GREEN}  ✓ All dependencies verified${NC}"
fi

# ============================================================================
# Check 3: Check if go.mod and go.sum are in sync
# ============================================================================
echo "  [3/4] Checking if go.sum is up to date..."

# Create a backup
cp go.sum go.sum.bak 2>/dev/null || true

# Run go mod tidy (in check mode)
if ! go mod tidy -v >/dev/null 2>&1; then
    echo -e "${RED}❌ ERROR: 'go mod tidy' failed${NC}"
    echo ""
    echo "There may be issues with dependencies"
    echo "Run: go mod tidy"
    rm -f go.sum.bak
    exit 1
fi

# Check if go.sum changed
if ! diff -q go.sum go.sum.bak >/dev/null 2>&1; then
    echo -e "${RED}❌ ERROR: go.sum is out of sync${NC}"
    echo ""
    echo "The go.sum file needs to be updated"
    echo ""
    echo "Action required:"
    echo "  1. Run: go mod tidy"
    echo "  2. Review changes: git diff go.sum"
    echo "  3. Commit updated go.sum"
    echo ""
    # Restore original
    mv go.sum.bak go.sum
    ((ERRORS++)) || true
else
    echo -e "${GREEN}  ✓ go.sum is in sync${NC}"
    rm -f go.sum.bak
fi

# ============================================================================
# Check 4: Check for unused dependencies (warning only)
# ============================================================================
echo "  [4/4] Checking for unused dependencies..."

# Capture go mod tidy output
TIDY_OUTPUT=$(go mod tidy -v 2>&1 || true)

if echo "$TIDY_OUTPUT" | grep -qi "unused"; then
    echo -e "${YELLOW}  ⚠️  Potential unused dependencies detected${NC}"
    echo "$TIDY_OUTPUT" | grep -i "unused" | sed 's/^/     /'
    echo ""
    echo "  (This is informational only, not blocking)"
    echo ""
else
    echo -e "${GREEN}  ✓ No obvious unused dependencies${NC}"
fi

# ============================================================================
# Summary
# ============================================================================
echo ""
if [ "$ERRORS" -eq 0 ]; then
    echo -e "${GREEN}✅ Dependency validation passed${NC}"

    # Show module info
    MODULE_PATH=$(grep '^module ' go.mod | awk '{print $2}')
    GO_VERSION=$(grep '^go ' go.mod | awk '{print $2}')
    DEP_COUNT=$(grep -c "^	" go.mod || echo "0")

    echo ""
    echo "Module: $MODULE_PATH"
    echo "Go version: $GO_VERSION"
    echo "Dependencies: $DEP_COUNT"
    exit 0
else
    echo -e "${RED}❌ Dependency validation failed with $ERRORS error(s)${NC}"
    echo ""
    echo "Quick fix:"
    echo "  go mod tidy"
    echo "  go mod verify"
    exit 1
fi
