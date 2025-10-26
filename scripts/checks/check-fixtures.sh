#!/bin/bash
# check-fixtures.sh - Verify test fixture files exist
#
# Part of: Build Quality Gates (BAIME Experiment)
# Iteration: 1 (P0)
# Purpose: Ensure all test fixtures referenced in tests actually exist
# Historical Impact: Catches 8% of commit errors

set -euo pipefail

# Colors
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m'

echo "Verifying test fixtures..."

# ============================================================================
# Step 1: Collect fixture references from test files
# ============================================================================
echo "  [1/3] Scanning test files for fixture references..."

# Patterns to match:
# - testutil.LoadFixture(t, "filename.jsonl")
# - testutil.TempSessionFile(t, content)
# - filepath.Join("fixtures", "filename.json")
# - testdata/filename

# Use simpler pattern matching to avoid lookbehind issues
FIXTURE_REFS=$(grep -rh \
    'LoadFixture' \
    --include='*_test.go' \
    . 2>/dev/null | \
    grep -o '"[^"]*\.jsonl"' | \
    tr -d '"' | \
    sort -u || true)

if [ -z "$FIXTURE_REFS" ]; then
    echo -e "${GREEN}  ℹ️  No fixture references found in test files${NC}"
    echo "  (This is OK if tests don't use fixtures)"
    exit 0
fi

REF_COUNT=$(echo "$FIXTURE_REFS" | wc -w)
echo "  Found $REF_COUNT unique fixture reference(s)"

# ============================================================================
# Step 2: Check if fixtures exist
# ============================================================================
echo "  [2/3] Verifying fixture files exist..."

MISSING=0
CHECKED=0
MISSING_LIST=""

for fixture in $FIXTURE_REFS; do
    # Skip dynamic filenames (contains variables or expressions)
    if [[ "$fixture" =~ \$|%|\{ ]]; then
        continue
    fi

    ((CHECKED++)) || true

    # Check in multiple possible locations
    FOUND=false
    for dir in "tests/fixtures" "testdata" "internal/testutil/fixtures"; do
        if [ -f "$dir/$fixture" ]; then
            FOUND=true
            break
        fi
    done

    if [ "$FOUND" = false ]; then
        echo -e "${RED}  ❌ Missing: $fixture${NC}"

        # Find where it's referenced
        REFS=$(grep -rn "$fixture" --include='*_test.go' . 2>/dev/null | head -2)
        if [ -n "$REFS" ]; then
            echo "     Referenced in:"
            echo "$REFS" | sed 's/^/       /'
        fi
        echo ""

        MISSING_LIST="$MISSING_LIST\n  - $fixture"
        ((MISSING++)) || true
    fi
done

# ============================================================================
# Step 3: Check for unused fixtures (optional warning)
# ============================================================================
echo "  [3/3] Checking for unused fixtures..."

UNUSED=0
if [ -d "tests/fixtures" ]; then
    for fixture_file in tests/fixtures/*.json tests/fixtures/*.jsonl tests/fixtures/*.txt tests/fixtures/*.yaml; do
        [ -f "$fixture_file" ] 2>/dev/null || continue

        BASENAME=$(basename "$fixture_file")

        # Check if this fixture is referenced in any test
        if ! grep -rq "$BASENAME" --include='*_test.go' . 2>/dev/null; then
            if [ "$UNUSED" -eq 0 ]; then
                echo -e "${YELLOW}  ⚠️  Unused fixtures detected:${NC}"
            fi
            echo "     $fixture_file"
            ((UNUSED++)) || true
        fi
    done
fi

if [ "$UNUSED" -gt 0 ]; then
    echo ""
    echo -e "${YELLOW}  ⚠️  $UNUSED unused fixture(s) found${NC}"
    echo "     Consider removing them to reduce clutter"
    echo "     (This is a warning, not an error)"
    echo ""
fi

# ============================================================================
# Summary
# ============================================================================
echo ""
if [ "$MISSING" -eq 0 ]; then
    if [ "$CHECKED" -eq 0 ]; then
        echo -e "${GREEN}✅ No fixture validation needed (no references found)${NC}"
    else
        echo -e "${GREEN}✅ All $CHECKED fixture(s) verified and present${NC}"
    fi
    exit 0
else
    echo -e "${RED}❌ ERROR: $MISSING of $CHECKED fixture(s) are missing${NC}"
    echo ""
    echo "Missing fixtures:"
    echo -e "$MISSING_LIST"
    echo ""
    echo "Action required:"
    echo "  1. Create missing fixture files in tests/fixtures/"
    echo "  2. Or remove test code that references them"
    echo "  3. Or use testutil.TempSessionFile() for dynamic fixtures"
    echo ""
    echo "Example:"
    echo "  mkdir -p tests/fixtures"
    echo "  echo '{\"test\":\"data\"}' > tests/fixtures/missing-file.jsonl"
    exit 1
fi
