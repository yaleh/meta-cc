#!/bin/bash
# Test quality checks
# Detects common test quality issues

set -e

echo "=== Test Quality Check ==="
echo ""

# Check 1: Hardcoded project hashes
echo "[1/2] Checking for hardcoded project hashes in tests..."
PROBLEM_FILES=$(grep -rn -E 'projectHash[[:space:]]*:=[[:space:]]*"[/-]' --include="*_test.go" cmd/ pkg/ 2>/dev/null | \
    grep -v writeSessionFixture || true)

if [ -n "$PROBLEM_FILES" ]; then
    echo "⚠️  WARNING: Hardcoded project hash detected:"
    echo "$PROBLEM_FILES" | sed 's/^/  /'
    echo "   Consider using writeSessionFixture() for better cross-platform compatibility."
    echo "   (Non-blocking warning)"
else
    echo "✓ No problematic project hashes found"
fi
echo ""

# Check 2: os.UserHomeDir() usage in tests
echo "[2/2] Checking for os.UserHomeDir() in tests..."
FILES=$(grep -rl 'os\.UserHomeDir' --include="*_test.go" . 2>/dev/null || true)

if [ -n "$FILES" ]; then
    echo "⚠️  WARNING: os.UserHomeDir() found in tests:"
    echo "$FILES" | sed 's/^/  - /'
    echo "   Consider using META_CC_PROJECTS_ROOT instead."
else
    echo "✓ No os.UserHomeDir() usage in tests"
fi
echo ""

echo "✅ Test quality check passed"
