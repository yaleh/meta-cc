#!/usr/bin/env bash
# Dependency Health Check Script
# Runs all dependency health checks locally
# Pattern: CI/CD Automation Integration (Pattern 5)
# Source: iteration-2-automation-pattern.yaml

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
CHECKS_PASSED=0
CHECKS_FAILED=0
CHECKS_WARNING=0

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}Dependency Health Check${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""

# Function to print section header
print_header() {
    echo ""
    echo -e "${BLUE}-----------------------------------${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}-----------------------------------${NC}"
}

# Function to print success
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
    ((CHECKS_PASSED++))
}

# Function to print error
print_error() {
    echo -e "${RED}✗ $1${NC}"
    ((CHECKS_FAILED++))
}

# Function to print warning
print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
    ((CHECKS_WARNING++))
}

# Check 1: Vulnerability Scan
print_header "1. Vulnerability Scan (govulncheck)"

if ! command -v govulncheck &> /dev/null; then
    print_error "govulncheck not installed"
    echo "Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"
    ((CHECKS_FAILED++))
else
    if govulncheck ./... > /tmp/govulncheck-output.txt 2>&1; then
        print_success "No vulnerabilities found"
    else
        print_error "Vulnerabilities detected"
        echo ""
        cat /tmp/govulncheck-output.txt
        echo ""
    fi
fi

# Check 2: License Compliance
print_header "2. License Compliance (go-licenses)"

if ! command -v go-licenses &> /dev/null; then
    print_error "go-licenses not installed"
    echo "Install with: go install github.com/google/go-licenses@latest"
else
    if go-licenses csv ./... > /tmp/licenses.csv 2>/dev/null; then
        # Check for prohibited licenses
        PROHIBITED="GPL-2.0|GPL-3.0|AGPL-3.0|SSPL|Commons-Clause"

        if grep -qE "($PROHIBITED)" /tmp/licenses.csv 2>/dev/null; then
            print_error "Prohibited licenses found"
            echo ""
            grep -E "($PROHIBITED)" /tmp/licenses.csv
            echo ""
        else
            DEP_COUNT=$(wc -l < /tmp/licenses.csv)
            print_success "All $DEP_COUNT dependencies compliant"
        fi
    else
        print_error "License check failed"
    fi
fi

# Check 3: Dependency Freshness
print_header "3. Dependency Freshness"

if go list -m -u all > /tmp/dependency-versions.txt 2>&1; then
    OUTDATED_COUNT=$(grep -c '\[' /tmp/dependency-versions.txt || echo "0")

    if [ "$OUTDATED_COUNT" -eq 0 ]; then
        print_success "All dependencies up to date"
    else
        print_warning "$OUTDATED_COUNT outdated dependencies"
        echo ""
        echo "Outdated dependencies:"
        grep '\[' /tmp/dependency-versions.txt || true
        echo ""
        echo "Run 'scripts/update-deps.sh' to update"
    fi
else
    print_error "Dependency freshness check failed"
fi

# Check 4: Go Module Tidy
print_header "4. Go Module Tidy Check"

# Check if go.mod and go.sum are tidy
if go mod tidy -diff > /tmp/go-mod-diff.txt 2>&1; then
    print_success "go.mod and go.sum are tidy"
else
    if [ -s /tmp/go-mod-diff.txt ]; then
        print_warning "go.mod or go.sum needs tidying"
        echo ""
        echo "Differences:"
        cat /tmp/go-mod-diff.txt
        echo ""
        echo "Run 'go mod tidy' to fix"
    else
        print_success "go.mod and go.sum are tidy"
    fi
fi

# Check 5: Test Suite
print_header "5. Test Suite (optional)"

echo "Skipping tests (use 'go test ./...' to run manually)"

# Summary
echo ""
echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}Summary${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""
echo -e "Passed:   ${GREEN}$CHECKS_PASSED${NC}"
echo -e "Warnings: ${YELLOW}$CHECKS_WARNING${NC}"
echo -e "Failed:   ${RED}$CHECKS_FAILED${NC}"
echo ""

if [ $CHECKS_FAILED -gt 0 ]; then
    echo -e "${RED}Dependency health check FAILED${NC}"
    exit 1
elif [ $CHECKS_WARNING -gt 0 ]; then
    echo -e "${YELLOW}Dependency health check passed with WARNINGS${NC}"
    exit 0
else
    echo -e "${GREEN}All dependency health checks PASSED${NC}"
    exit 0
fi
