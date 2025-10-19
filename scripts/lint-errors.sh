#!/bin/bash
#
# lint-errors.sh - Simple error linting for meta-cc
#
# Checks Go error handling patterns for consistency and best practices.
# Detects:
#   1. fmt.Errorf without %w wrapping
#   2. Short error messages (<20 chars)
#   3. Missing mcerrors import in files with errors
#   4. Direct errors.New usage (should use sentinel errors)
#
# Usage:
#   ./scripts/lint-errors.sh [directory]
#
# Exit codes:
#   0 - No issues found
#   1 - Issues found (warnings)
#   2 - Error running script
#

set -euo pipefail

# Default to current directory if not specified
TARGET_DIR="${1:-.}"

# Colors for output
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Counters
TOTAL_ISSUES=0
WARNINGS=0
ERRORS=0

# Output file for detailed results
REPORT_FILE=$(mktemp)
trap "rm -f $REPORT_FILE" EXIT

echo "==================================================================" | tee "$REPORT_FILE"
echo "Error Linter - meta-cc" | tee -a "$REPORT_FILE"
echo "==================================================================" | tee -a "$REPORT_FILE"
echo "Target: $TARGET_DIR" | tee -a "$REPORT_FILE"
echo "" | tee -a "$REPORT_FILE"

#
# Check 1: fmt.Errorf without %w wrapping
#
echo "Check 1: fmt.Errorf without %w wrapping..." | tee -a "$REPORT_FILE"
echo "------------------------------------------------------------------" | tee -a "$REPORT_FILE"

# Find fmt.Errorf calls that don't use %w
# Pattern: fmt.Errorf with string literal but no %w
BARE_ERRORF=$(grep -rn 'fmt\.Errorf(".*"[^)]' "$TARGET_DIR" --include="*.go" 2>/dev/null | grep -v '%w' | grep -v '_test.go' || true)

if [ -n "$BARE_ERRORF" ]; then
    echo "$BARE_ERRORF" | while IFS= read -r line; do
        echo "${YELLOW}WARNING${NC}: $line" | tee -a "$REPORT_FILE"
        echo "  → Suggestion: Add %w and wrap with sentinel error" | tee -a "$REPORT_FILE"
        ((WARNINGS++)) || true
        ((TOTAL_ISSUES++)) || true
    done
else
    echo "${GREEN}✓${NC} No bare fmt.Errorf calls found" | tee -a "$REPORT_FILE"
fi
echo "" | tee -a "$REPORT_FILE"

#
# Check 2: Short error messages (<20 chars)
#
echo "Check 2: Short error messages (<20 chars)..." | tee -a "$REPORT_FILE"
echo "------------------------------------------------------------------" | tee -a "$REPORT_FILE"

# Find error messages that are too short (lacking context)
# Pattern: fmt.Errorf or errors.New with message < 20 chars
SHORT_ERRORS=$(grep -rn 'fmt\.Errorf\|errors\.New' "$TARGET_DIR" --include="*.go" 2>/dev/null | \
    grep -E '(fmt\.Errorf|errors\.New)\(".\{1,19\}"' | \
    grep -v '_test.go' || true)

if [ -n "$SHORT_ERRORS" ]; then
    echo "$SHORT_ERRORS" | while IFS= read -r line; do
        echo "${YELLOW}WARNING${NC}: $line" | tee -a "$REPORT_FILE"
        echo "  → Suggestion: Add operation context and details" | tee -a "$REPORT_FILE"
        ((WARNINGS++)) || true
        ((TOTAL_ISSUES++)) || true
    done
else
    echo "${GREEN}✓${NC} No short error messages found" | tee -a "$REPORT_FILE"
fi
echo "" | tee -a "$REPORT_FILE"

#
# Check 3: Missing mcerrors import
#
echo "Check 3: Missing mcerrors import in files with errors..." | tee -a "$REPORT_FILE"
echo "------------------------------------------------------------------" | tee -a "$REPORT_FILE"

# Find files with fmt.Errorf or errors.New but no mcerrors import
FILES_WITH_ERRORS=$(grep -rl 'fmt\.Errorf\|errors\.New' "$TARGET_DIR" --include="*.go" 2>/dev/null | \
    grep -v '_test.go' | \
    grep -v 'internal/errors/errors.go' || true)

MISSING_IMPORT=0
if [ -n "$FILES_WITH_ERRORS" ]; then
    for file in $FILES_WITH_ERRORS; do
        if ! grep -q 'mcerrors "github.com/yaleh/meta-cc/internal/errors"' "$file"; then
            echo "${YELLOW}INFO${NC}: $file:1: Missing mcerrors import" | tee -a "$REPORT_FILE"
            echo "  → Suggestion: Add import mcerrors \"github.com/yaleh/meta-cc/internal/errors\"" | tee -a "$REPORT_FILE"
            ((MISSING_IMPORT++)) || true
            ((TOTAL_ISSUES++)) || true
        fi
    done
fi

if [ "$MISSING_IMPORT" -eq 0 ]; then
    echo "${GREEN}✓${NC} All error files have mcerrors import" | tee -a "$REPORT_FILE"
else
    echo "${YELLOW}INFO${NC}: $MISSING_IMPORT file(s) missing mcerrors import" | tee -a "$REPORT_FILE"
fi
echo "" | tee -a "$REPORT_FILE"

#
# Check 4: Direct errors.New usage
#
echo "Check 4: Direct errors.New usage (should use sentinels)..." | tee -a "$REPORT_FILE"
echo "------------------------------------------------------------------" | tee -a "$REPORT_FILE"

# Find errors.New calls (should typically use sentinel errors instead)
ERRORS_NEW=$(grep -rn 'errors\.New(' "$TARGET_DIR" --include="*.go" 2>/dev/null | \
    grep -v '_test.go' | \
    grep -v 'internal/errors/errors.go' || true)

if [ -n "$ERRORS_NEW" ]; then
    echo "$ERRORS_NEW" | while IFS= read -r line; do
        echo "${YELLOW}INFO${NC}: $line" | tee -a "$REPORT_FILE"
        echo "  → Suggestion: Consider using sentinel error instead" | tee -a "$REPORT_FILE"
        ((TOTAL_ISSUES++)) || true
    done
else
    echo "${GREEN}✓${NC} No direct errors.New usage found" | tee -a "$REPORT_FILE"
fi
echo "" | tee -a "$REPORT_FILE"

#
# Summary
#
echo "==================================================================" | tee -a "$REPORT_FILE"
echo "Summary" | tee -a "$REPORT_FILE"
echo "==================================================================" | tee -a "$REPORT_FILE"
echo "Total issues found: $TOTAL_ISSUES" | tee -a "$REPORT_FILE"
echo "  - Warnings: $WARNINGS" | tee -a "$REPORT_FILE"
echo "  - Errors: $ERRORS" | tee -a "$REPORT_FILE"
echo "  - Info: $((TOTAL_ISSUES - WARNINGS - ERRORS))" | tee -a "$REPORT_FILE"
echo "" | tee -a "$REPORT_FILE"

if [ "$TOTAL_ISSUES" -eq 0 ]; then
    echo "${GREEN}✓ All checks passed!${NC}" | tee -a "$REPORT_FILE"
    exit 0
elif [ "$ERRORS" -gt 0 ]; then
    echo "${RED}✗ $ERRORS error(s) found${NC}" | tee -a "$REPORT_FILE"
    echo "" | tee -a "$REPORT_FILE"
    echo "See detailed report above." | tee -a "$REPORT_FILE"
    exit 1
else
    echo "${YELLOW}⚠ $TOTAL_ISSUES issue(s) found (warnings/info only)${NC}" | tee -a "$REPORT_FILE"
    echo "" | tee -a "$REPORT_FILE"
    echo "See detailed report above." | tee -a "$REPORT_FILE"
    exit 0
fi
