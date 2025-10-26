#!/bin/bash
# Pre-release validation script
#
# Purpose: Validate all release requirements locally BEFORE creating git tag
# Usage: ./scripts/pre-release-check.sh <version>
# Example: ./scripts/pre-release-check.sh v2.0.3
#
# This script prevents the 3 common release failures:
# 1. Linting errors (golangci-lint)
# 2. Version mismatch (marketplace.json vs tag)
# 3. Smoke test failures (binary structure)

set -e

VERSION=$1

if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v2.0.3"
    exit 1
fi

# Remove 'v' prefix for version comparison
VERSION_NUM=${VERSION#v}

echo "========================================="
echo "Pre-Release Validation"
echo "========================================="
echo "Target version: $VERSION ($VERSION_NUM)"
echo ""

# ==================================================================
# VALIDATION TRACKING
# ==================================================================

TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=()
WARNINGS=()

check_result() {
    local check_name="$1"
    local result="$2"
    local error_msg="$3"

    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

    if [ "$result" = "pass" ]; then
        echo "  ✓ $check_name"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
    elif [ "$result" = "warn" ]; then
        echo "  ⚠ $check_name"
        WARNINGS+=("$check_name: $error_msg")
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
    else
        echo "  ✗ $check_name"
        if [ -n "$error_msg" ]; then
            echo "    Error: $error_msg"
        fi
        FAILED_CHECKS+=("$check_name: $error_msg")
    fi
}

# ==================================================================
# CHECK 1: Git Repository Status
# ==================================================================

echo "Check 1: Git Repository Status"
echo "-------------------------------"

# Check 1.1: Working directory is clean
if [ -z "$(git status --porcelain)" ]; then
    check_result "Working directory is clean" "pass"
else
    check_result "Working directory is clean" "fail" "Uncommitted changes detected. Run 'git status'"
fi

# Check 1.2: On correct branch
BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ "$BRANCH" = "main" ]; then
    check_result "On main branch" "pass"
else
    check_result "On main branch" "warn" "Current branch: $BRANCH (usually release from main)"
fi

# Check 1.3: Tag doesn't already exist
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    check_result "Tag doesn't exist yet" "fail" "Tag $VERSION already exists. Delete with: git tag -d $VERSION"
else
    check_result "Tag doesn't exist yet" "pass"
fi

echo ""

# ==================================================================
# CHECK 2: Version Consistency
# ==================================================================

echo "Check 2: Version Consistency"
echo "----------------------------"

# Check 2.1: marketplace.json exists and is valid JSON
if [ -f ".claude-plugin/marketplace.json" ]; then
    if jq . .claude-plugin/marketplace.json >/dev/null 2>&1; then
        check_result "marketplace.json is valid JSON" "pass"
    else
        check_result "marketplace.json is valid JSON" "fail" "JSON syntax error"
    fi
else
    check_result "marketplace.json exists" "fail" "File not found: .claude-plugin/marketplace.json"
fi

# Check 2.2: marketplace.json version matches target
MARKETPLACE_VERSION=$(jq -r '.plugins[0].version' .claude-plugin/marketplace.json 2>/dev/null || echo "UNKNOWN")
if [ "$MARKETPLACE_VERSION" = "$VERSION_NUM" ]; then
    check_result "marketplace.json version matches ($VERSION_NUM)" "pass"
else
    check_result "marketplace.json version matches" "fail" "marketplace.json has '$MARKETPLACE_VERSION' but target is '$VERSION_NUM'. Run: ./scripts/bump-version.sh $VERSION"
fi

echo ""

# ==================================================================
# CHECK 3: Code Quality (Linting)
# ==================================================================

echo "Check 3: Code Quality (Linting)"
echo "--------------------------------"

# Check 3.1: gofmt
UNFORMATTED=$(gofmt -l . 2>/dev/null | grep -v vendor || true)
if [ -z "$UNFORMATTED" ]; then
    check_result "Code is formatted (gofmt)" "pass"
else
    check_result "Code is formatted (gofmt)" "fail" "Run 'make fmt' to fix: $(echo $UNFORMATTED | head -1)"
fi

# Check 3.2: go vet
if go vet ./... >/dev/null 2>&1; then
    check_result "go vet passes" "pass"
else
    check_result "go vet passes" "fail" "Run 'go vet ./...' for details"
fi

# Check 3.3: golangci-lint (if installed)
if command -v golangci-lint >/dev/null 2>&1; then
    if golangci-lint run ./... >/dev/null 2>&1; then
        check_result "golangci-lint passes" "pass"
    else
        # Check if this is the known Go version mismatch issue
        LINT_ERROR=$(golangci-lint run ./... 2>&1 | head -5)
        if echo "$LINT_ERROR" | grep -q "requires newer Go version go1.24.*application built with go1.23"; then
            check_result "golangci-lint passes" "warn" "golangci-lint version issue (built with go1.23, checking go1.24 code) - will be verified in CI"
        else
            check_result "golangci-lint passes" "fail" "Run 'golangci-lint run ./...' for details"
        fi
    fi
else
    check_result "golangci-lint passes" "warn" "golangci-lint not installed (skipping)"
fi

echo ""

# ==================================================================
# CHECK 4: Tests
# ==================================================================

echo "Check 4: Tests"
echo "--------------"

# Check 4.1: Unit tests pass
if go test -short ./... >/dev/null 2>&1; then
    check_result "Unit tests pass (short mode)" "pass"
else
    check_result "Unit tests pass (short mode)" "fail" "Run 'go test -short ./...' for details"
fi

# Check 4.2: Test coverage (if coverage goal exists)
if go test -short -coverprofile=coverage.tmp ./... >/dev/null 2>&1; then
    COVERAGE=$(go tool cover -func=coverage.tmp 2>/dev/null | tail -1 | awk '{print $3}' | sed 's/%//' || echo "0")
    rm -f coverage.tmp
    if [ "${COVERAGE%.*}" -ge 80 ]; then
        check_result "Test coverage ≥80% ($COVERAGE%)" "pass"
    else
        check_result "Test coverage ≥80%" "warn" "Current coverage: $COVERAGE% (goal: 80%)"
    fi
else
    check_result "Test coverage check" "warn" "Could not calculate coverage"
fi

echo ""

# ==================================================================
# CHECK 5: Build Validation
# ==================================================================

echo "Check 5: Build Validation"
echo "-------------------------"

# Check 5.1: MCP server builds
if go build -o /tmp/meta-cc-mcp-test ./cmd/mcp-server >/dev/null 2>&1; then
    check_result "MCP server builds" "pass"
    rm -f /tmp/meta-cc-mcp-test
else
    check_result "MCP server builds" "fail" "Run 'go build ./cmd/mcp-server' for details"
fi

# Check 5.2: Dependencies are tidy
cp go.mod go.mod.bak 2>/dev/null || true
cp go.sum go.sum.bak 2>/dev/null || true
go mod tidy >/dev/null 2>&1
if diff -q go.mod go.mod.bak >/dev/null 2>&1 && diff -q go.sum go.sum.bak >/dev/null 2>&1; then
    check_result "go.mod and go.sum are tidy" "pass"
else
    check_result "go.mod and go.sum are tidy" "fail" "Run 'go mod tidy' and commit changes"
fi
mv go.mod.bak go.mod 2>/dev/null || true
mv go.sum.bak go.sum 2>/dev/null || true

echo ""

# ==================================================================
# CHECK 6: Plugin Structure
# ==================================================================

echo "Check 6: Plugin Structure"
echo "-------------------------"

# Check 6.1: Required directories exist
REQUIRED_DIRS=(".claude" ".claude/commands" ".claude/agents" ".claude/skills" "lib")
for dir in "${REQUIRED_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        check_result "Directory exists: $dir" "pass"
    else
        check_result "Directory exists: $dir" "fail" "Required directory missing"
    fi
done

# Check 6.2: Skills structure
SKILL_COUNT=$(find .claude/skills -name "SKILL.md" 2>/dev/null | wc -l)
EXPECTED_SKILLS=$(jq -r '.plugins[0].skills | length' .claude-plugin/marketplace.json 2>/dev/null || echo 0)
if [ "$SKILL_COUNT" -eq "$EXPECTED_SKILLS" ]; then
    check_result "Skills count matches marketplace.json ($EXPECTED_SKILLS)" "pass"
else
    check_result "Skills count matches marketplace.json" "warn" "Found $SKILL_COUNT skills, marketplace.json declares $EXPECTED_SKILLS"
fi

# Check 6.3: Agents structure
AGENT_COUNT=$(find .claude/agents -name "*.md" 2>/dev/null | wc -l)
EXPECTED_AGENTS=$(jq -r '.plugins[0].agents | length' .claude-plugin/marketplace.json 2>/dev/null || echo 0)
if [ "$AGENT_COUNT" -eq "$EXPECTED_AGENTS" ]; then
    check_result "Agents count matches marketplace.json ($EXPECTED_AGENTS)" "pass"
else
    check_result "Agents count matches marketplace.json" "warn" "Found $AGENT_COUNT agents, marketplace.json declares $EXPECTED_AGENTS"
fi

echo ""

# ==================================================================
# CHECK 7: Local Smoke Tests (Optional)
# ==================================================================

echo "Check 7: Local Smoke Tests (Optional)"
echo "--------------------------------------"

# Only run if smoke test script exists
if [ -f "scripts/smoke-tests.sh" ]; then
    echo "  Building test package for local smoke tests..."

    # Build MCP binary
    if go build -o build/meta-cc-mcp-linux-amd64 ./cmd/mcp-server 2>/dev/null; then
        # Create minimal test package
        mkdir -p build/test-package/meta-cc-plugin-linux-amd64/{bin,.claude-plugin,commands,agents,skills,lib}
        cp build/meta-cc-mcp-linux-amd64 build/test-package/meta-cc-plugin-linux-amd64/bin/meta-cc-mcp
        chmod +x build/test-package/meta-cc-plugin-linux-amd64/bin/meta-cc-mcp
        cp -r .claude-plugin/* build/test-package/meta-cc-plugin-linux-amd64/.claude-plugin/
        cp -r .claude/commands/*.md build/test-package/meta-cc-plugin-linux-amd64/commands/ 2>/dev/null || true
        cp -r .claude/agents/*.md build/test-package/meta-cc-plugin-linux-amd64/agents/ 2>/dev/null || true
        cp -r .claude/skills/* build/test-package/meta-cc-plugin-linux-amd64/skills/ 2>/dev/null || true
        cp -r lib/* build/test-package/meta-cc-plugin-linux-amd64/lib/ 2>/dev/null || true
        cp scripts/install.sh build/test-package/meta-cc-plugin-linux-amd64/
        cp scripts/uninstall.sh build/test-package/meta-cc-plugin-linux-amd64/ 2>/dev/null || true
        cp README.md build/test-package/meta-cc-plugin-linux-amd64/
        cp LICENSE build/test-package/meta-cc-plugin-linux-amd64/

        # Create package
        cd build/test-package
        tar -czf ../meta-cc-test-package.tar.gz meta-cc-plugin-linux-amd64
        cd ../..

        # Run smoke tests
        if bash scripts/smoke-tests.sh "$VERSION" "linux-amd64" "build/meta-cc-test-package.tar.gz" >/dev/null 2>&1; then
            check_result "Local smoke tests pass" "pass"
        else
            check_result "Local smoke tests pass" "fail" "Run 'bash scripts/smoke-tests.sh $VERSION linux-amd64 build/meta-cc-test-package.tar.gz' for details"
        fi

        # Cleanup
        rm -rf build/test-package build/meta-cc-test-package.tar.gz
    else
        check_result "Local smoke tests" "warn" "Could not build test binary"
    fi
else
    check_result "Local smoke tests" "warn" "Smoke test script not found (skipping)"
fi

echo ""

# ==================================================================
# RESULTS SUMMARY
# ==================================================================

echo "========================================="
echo "Pre-Release Validation Results"
echo "========================================="
echo "Total checks:  $TOTAL_CHECKS"
echo "Passed:        $PASSED_CHECKS"
echo "Failed:        $((TOTAL_CHECKS - PASSED_CHECKS))"
echo ""

if [ ${#WARNINGS[@]} -gt 0 ]; then
    echo "Warnings:"
    for warning in "${WARNINGS[@]}"; do
        echo "  ⚠ $warning"
    done
    echo ""
fi

if [ ${#FAILED_CHECKS[@]} -gt 0 ]; then
    echo "Failed Checks:"
    for failure in "${FAILED_CHECKS[@]}"; do
        echo "  ✗ $failure"
    done
    echo ""
    echo "❌ PRE-RELEASE VALIDATION FAILED"
    echo ""
    echo "Action Required:"
    echo "  1. Fix the issues listed above"
    echo "  2. Re-run this script: ./scripts/pre-release-check.sh $VERSION"
    echo "  3. Once all checks pass, proceed with release"
    exit 1
else
    echo "✓ ALL PRE-RELEASE CHECKS PASSED"
    echo ""
    echo "Next steps to create release:"
    echo "  1. Create and push tag:"
    echo "     git tag -a $VERSION -m \"Release $VERSION\""
    echo "     git push origin $VERSION"
    echo ""
    echo "  2. Or use the automated release script:"
    echo "     ./scripts/release.sh $VERSION"
    echo ""
    echo "The release workflow will automatically:"
    echo "  - Build binaries for all platforms"
    echo "  - Run smoke tests"
    echo "  - Create GitHub release"
    echo "  - Upload release artifacts"
    exit 0
fi
