#!/usr/bin/env bash
#
# check-debug.sh - Debug statement detection
#
# PURPOSE:
#   Detects debug statements left in production code.
#   Prevents accidental commits of debugging code.
#
# HISTORICAL IMPACT:
#   - Debug statements: ~5% of CI failures (2-3/50 samples)
#   - Common patterns: fmt.Print*, log.Print*, console.log, TODO/FIXME
#   - Issue: Noisy logs, performance impact, security leaks
#
# USAGE:
#   ./scripts/check-debug.sh
#   make check-debug
#
# EXIT CODES:
#   0 - No debug statements found
#   1 - Debug statements detected
#
# DEPENDENCIES:
#   - grep
#

set -euo pipefail

# Colors for output
readonly RED='\033[0;31m'
readonly YELLOW='\033[1;33m'
readonly GREEN='\033[0;32m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# Warning counter (not blocking, but informative)
WARNINGS=0
ERRORS=0

echo -e "${BLUE}=== Debug Statement Detection ===${NC}"
echo ""

# ==============================================================================
# Check 1/4: Find debug print statements in Go code
# ==============================================================================
echo -e "${BLUE}[1/4] Checking for fmt.Print* debug statements...${NC}"

# Find fmt.Print* (but exclude fmt.Fprintf to stderr/stdout which is legitimate)
DEBUG_PRINTS=$(grep -rn \
    --include="*.go" \
    --exclude-dir="vendor" \
    --exclude-dir=".git" \
    --exclude-dir="build" \
    --exclude-dir="dist" \
    -E "fmt\.Print(ln)?\(" . 2>/dev/null | \
    grep -v "fmt\.Fprintf" | \
    grep -v "fmt\.Fprint\(" | \
    grep -v "// allow" | \
    grep -v "//nolint" || true)

if [ -n "$DEBUG_PRINTS" ]; then
    echo -e "${RED}❌ Found fmt.Print* statements:${NC}"
    echo "$DEBUG_PRINTS" | sed 's/^/  /'
    echo ""
    ERRORS=$((ERRORS + 1))
else
    echo "✓ No fmt.Print* statements found"
fi
echo ""

# ==============================================================================
# Check 2/4: Find log.Print* statements (outside logging setup)
# ==============================================================================
echo -e "${BLUE}[2/4] Checking for log.Print* debug statements...${NC}"

# Find log.Print* in non-logging files
DEBUG_LOGS=$(find . -type f -name "*.go" \
    ! -path "./vendor/*" \
    ! -path "./.git/*" \
    ! -path "./build/*" \
    ! -path "./dist/*" \
    ! -name "*log*.go" \
    -exec grep -Hn "log\.Print" {} \; 2>/dev/null | \
    grep -v "// allow" | \
    grep -v "//nolint" || true)

if [ -n "$DEBUG_LOGS" ]; then
    echo -e "${YELLOW}⚠️  Found log.Print* statements:${NC}"
    echo "$DEBUG_LOGS" | sed 's/^/  /'
    echo ""
    WARNINGS=$((WARNINGS + 1))
else
    echo "✓ No suspicious log.Print* statements found"
fi
echo ""

# ==============================================================================
# Check 3/4: Find TODO/FIXME/HACK comments
# ==============================================================================
echo -e "${BLUE}[3/4] Checking for TODO/FIXME/HACK comments...${NC}"

# Count TODO/FIXME/HACK comments (informational only)
TODO_COUNT=$(grep -r \
    --include="*.go" \
    --include="*.sh" \
    --exclude-dir="vendor" \
    --exclude-dir=".git" \
    --exclude-dir="build" \
    --exclude-dir="dist" \
    -E "// (TODO|FIXME|HACK)" . 2>/dev/null | wc -l || echo "0")

TODO_COUNT=$(echo "$TODO_COUNT" | tr -d ' ')

if [ "$TODO_COUNT" -gt 0 ]; then
    echo -e "${BLUE}ℹ️  Found $TODO_COUNT TODO/FIXME/HACK comments${NC}"

    # Show first 5 as examples
    TODO_EXAMPLES=$(grep -rn \
        --include="*.go" \
        --include="*.sh" \
        --exclude-dir="vendor" \
        --exclude-dir=".git" \
        --exclude-dir="build" \
        --exclude-dir="dist" \
        -E "// (TODO|FIXME|HACK)" . 2>/dev/null | head -n 5 || true)

    if [ -n "$TODO_EXAMPLES" ]; then
        echo ""
        echo "Examples:"
        echo "$TODO_EXAMPLES" | sed 's/^/  /'
    fi

    echo ""
    echo "Note: These are informational only and don't block commits."
else
    echo "✓ No TODO/FIXME/HACK comments found"
fi
echo ""

# ==============================================================================
# Check 4/4: Find spew/pp/dumper debug packages
# ==============================================================================
echo -e "${BLUE}[4/4] Checking for debug package imports...${NC}"

# Find common debug packages: spew, pp, go-spew, go-dumper
DEBUG_IMPORTS=$(grep -rn \
    --include="*.go" \
    --exclude-dir="vendor" \
    --exclude-dir=".git" \
    --exclude-dir="build" \
    --exclude-dir="dist" \
    -E "\"(github.com/davecgh/go-spew|github.com/k0kubun/pp|github.com/sanity-io/litter)\"" . 2>/dev/null | \
    grep -v "// allow" | \
    grep -v "//nolint" || true)

if [ -n "$DEBUG_IMPORTS" ]; then
    echo -e "${RED}❌ Found debug package imports:${NC}"
    echo "$DEBUG_IMPORTS" | sed 's/^/  /'
    echo ""
    ERRORS=$((ERRORS + 1))
else
    echo "✓ No debug package imports found"
fi
echo ""

# ==============================================================================
# Summary
# ==============================================================================
if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}✅ No debug statements found${NC}"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    echo -e "${YELLOW}⚠️  Found $WARNINGS warnings (non-blocking)${NC}"
    echo ""
    echo "These warnings are informational only."
    echo "Review them before committing if appropriate."
    exit 0
else
    echo -e "${RED}❌ ERROR: Found $ERRORS blocking issue(s)${NC}"
    if [ $WARNINGS -gt 0 ]; then
        echo -e "${YELLOW}Also found $WARNINGS warnings (non-blocking)${NC}"
    fi
    echo ""
    echo -e "${YELLOW}How to fix:${NC}"
    echo "  1. Remove fmt.Print* statements (use proper logging)"
    echo "  2. Remove debug package imports (spew, pp, etc.)"
    echo "  3. For intentional debug code, add: // allow or //nolint"
    echo ""
    echo "Recommended replacements:"
    echo "  • fmt.Println() → log.Printf() or slog.Debug()"
    echo "  • spew.Dump()   → json.MarshalIndent() for structured output"
    echo ""
    exit 1
fi
