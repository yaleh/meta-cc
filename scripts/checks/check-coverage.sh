#!/bin/bash
# Check test coverage meets threshold
# Usage: ./check-coverage.sh [THRESHOLD]
#   Default threshold: 80% (matches CLAUDE.md / docs/core/principles.md coverage target)

set -e

THRESHOLD=${1:-80.0}

echo "=== Test Coverage Check ==="
echo ""
echo "Checking test coverage threshold..."

# Ensure coverage.out exists
if [ ! -f "coverage.out" ]; then
    echo "❌ ERROR: coverage.out not found"
    echo "Run 'make test-coverage' first to generate coverage report"
    exit 1
fi

# Extract total coverage percentage
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')

echo "Current coverage: ${COVERAGE}%"
echo "Required threshold: ${THRESHOLD}%"
echo ""

# Compare coverage to threshold (using awk for floating point comparison)
if awk "BEGIN {exit !($COVERAGE < $THRESHOLD)}"; then
    echo "❌ ERROR: Coverage ${COVERAGE}% is below threshold ${THRESHOLD}%"
    echo ""
    echo "Please add tests to increase coverage above ${THRESHOLD}%"
    echo ""
    echo "To check coverage locally:"
    echo "  make test-coverage"
    echo "  open coverage.html"
    exit 1
else
    echo "✅ Coverage ${COVERAGE}% meets threshold ${THRESHOLD}%"
fi
