#!/usr/bin/env bash
# Interactive Dependency Update Script
# Pattern: Dependency Update Testing (Pattern 6)
# Source: iteration-2-testing-pattern.yaml

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}Interactive Dependency Update${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""

# Check for required tools
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: go not found${NC}"
    exit 1
fi

# Step 1: Show current outdated dependencies
echo -e "${BLUE}Step 1: Checking for outdated dependencies...${NC}"
echo ""

if ! go list -m -u all > /tmp/deps-before.txt; then
    echo -e "${RED}Error: Failed to list dependencies${NC}"
    exit 1
fi

OUTDATED=$(grep '\[' /tmp/deps-before.txt || true)
OUTDATED_COUNT=$(echo "$OUTDATED" | grep -c '\[' || echo "0")

if [ "$OUTDATED_COUNT" -eq 0 ]; then
    echo -e "${GREEN}All dependencies are up to date!${NC}"
    exit 0
fi

echo "Found $OUTDATED_COUNT outdated dependencies:"
echo ""
echo "$OUTDATED"
echo ""

# Step 2: Ask for confirmation
read -p "Do you want to update all dependencies? (y/N): " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Update cancelled"
    exit 0
fi

# Step 3: Establish baseline
echo ""
echo -e "${BLUE}Step 2: Establishing baseline (running tests)...${NC}"
echo ""

if go test ./... > /tmp/baseline-tests.txt 2>&1; then
    BASELINE_PASS=$(grep -c "^ok" /tmp/baseline-tests.txt || echo "0")
    echo -e "${GREEN}Baseline: $BASELINE_PASS test packages passed${NC}"
else
    BASELINE_PASS=$(grep -c "^ok" /tmp/baseline-tests.txt || echo "0")
    BASELINE_FAIL=$(grep -c "^FAIL" /tmp/baseline-tests.txt || echo "0")
    echo -e "${YELLOW}Warning: $BASELINE_FAIL test packages failing in baseline${NC}"
    echo "Continuing anyway..."
fi

# Step 4: Update dependencies
echo ""
echo -e "${BLUE}Step 3: Updating dependencies...${NC}"
echo ""

# Backup go.mod and go.sum
cp go.mod go.mod.backup
cp go.sum go.sum.backup
echo "Backed up go.mod and go.sum"

# Update all dependencies
if go get -u ./...; then
    echo -e "${GREEN}Dependencies updated${NC}"
else
    echo -e "${RED}Error: Failed to update dependencies${NC}"
    echo "Restoring backup..."
    mv go.mod.backup go.mod
    mv go.sum.backup go.sum
    exit 1
fi

# Tidy
if go mod tidy; then
    echo -e "${GREEN}go.mod tidied${NC}"
else
    echo -e "${RED}Error: go mod tidy failed${NC}"
    echo "Restoring backup..."
    mv go.mod.backup go.mod
    mv go.sum.backup go.sum
    exit 1
fi

# Step 5: Run tests after update
echo ""
echo -e "${BLUE}Step 4: Running tests after update...${NC}"
echo ""

if go test ./... > /tmp/after-tests.txt 2>&1; then
    AFTER_PASS=$(grep -c "^ok" /tmp/after-tests.txt || echo "0")
    echo -e "${GREEN}After update: $AFTER_PASS test packages passed${NC}"
else
    AFTER_PASS=$(grep -c "^ok" /tmp/after-tests.txt || echo "0")
    AFTER_FAIL=$(grep -c "^FAIL" /tmp/after-tests.txt || echo "0")
    echo -e "${RED}After update: $AFTER_FAIL test packages failing${NC}"
fi

# Step 6: Compare results
echo ""
echo -e "${BLUE}Step 5: Comparing test results...${NC}"
echo ""

echo "Baseline: $BASELINE_PASS packages passed"
echo "After:    $AFTER_PASS packages passed"

if [ "$AFTER_PASS" -lt "$BASELINE_PASS" ]; then
    echo ""
    echo -e "${RED}REGRESSION DETECTED!${NC}"
    echo "Tests that passed before now fail after update"
    echo ""
    echo "Test output:"
    cat /tmp/after-tests.txt
    echo ""
    read -p "Rollback to previous versions? (Y/n): " -n 1 -r
    echo ""

    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        echo "Rolling back..."
        mv go.mod.backup go.mod
        mv go.sum.backup go.sum
        echo -e "${YELLOW}Rolled back to previous dependency versions${NC}"
        exit 1
    else
        echo -e "${YELLOW}Keeping updated dependencies despite regressions${NC}"
        rm go.mod.backup go.sum.backup
    fi
else
    echo ""
    echo -e "${GREEN}No regressions detected!${NC}"
    echo "Update successful"
    rm go.mod.backup go.sum.backup

    # Show what was updated
    echo ""
    echo -e "${BLUE}Updated dependencies:${NC}"
    go list -m -u all > /tmp/deps-after.txt
    diff /tmp/deps-before.txt /tmp/deps-after.txt || true
fi

# Step 7: Run vulnerability scan
echo ""
echo -e "${BLUE}Step 6: Running vulnerability scan...${NC}"
echo ""

if command -v govulncheck &> /dev/null; then
    if govulncheck ./...; then
        echo -e "${GREEN}No vulnerabilities found${NC}"
    else
        echo -e "${RED}Vulnerabilities found${NC}"
        echo "Consider addressing these before committing"
    fi
else
    echo -e "${YELLOW}govulncheck not installed, skipping vulnerability scan${NC}"
fi

echo ""
echo -e "${GREEN}Dependency update complete!${NC}"
echo ""
echo "Next steps:"
echo "1. Review changes with: git diff go.mod go.sum"
echo "2. Run full test suite: go test ./..."
echo "3. Commit changes: git add go.mod go.sum && git commit -m 'chore: update dependencies'"
