#!/bin/bash
# check-temp-files.sh - Detect temporary files that should not be committed
#
# Part of: Build Quality Gates (BAIME Experiment)
# Iteration: 1 (P0)
# Purpose: Prevent commit of temporary test/debug files
# Historical Impact: Catches 28% of commit errors

set -euo pipefail

# Colors
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m'

echo "Checking for temporary files..."

ERRORS=0

# ============================================================================
# Check 1: Root directory .go files (except main.go)
# ============================================================================
echo "  [1/4] Checking root directory for temporary .go files..."

TEMP_GO=$(find . -maxdepth 1 -name "*.go" ! -name "main.go" -type f 2>/dev/null || true)

if [ -n "$TEMP_GO" ]; then
    echo -e "${RED}❌ ERROR: Temporary .go files in project root:${NC}"
    echo "$TEMP_GO" | sed 's/^/  - /'
    echo ""
    echo "These files should be:"
    echo "  1. Moved to appropriate package directories (e.g., cmd/, internal/)"
    echo "  2. Or deleted if they are debug/test scripts"
    echo ""
    ((ERRORS++)) || true
fi

# ============================================================================
# Check 2: Common temporary file patterns
# ============================================================================
echo "  [2/4] Checking for test/debug script patterns..."

TEMP_SCRIPTS=$(find . -type f \( \
    -name "test_*.go" -o \
    -name "debug_*.go" -o \
    -name "tmp_*.go" -o \
    -name "scratch_*.go" -o \
    -name "experiment_*.go" \
\) ! -path "./vendor/*" ! -path "./.git/*" ! -path "*/temp_file_manager*.go" 2>/dev/null || true)

if [ -n "$TEMP_SCRIPTS" ]; then
    echo -e "${RED}❌ ERROR: Temporary test/debug scripts found:${NC}"
    echo "$TEMP_SCRIPTS" | sed 's/^/  - /'
    echo ""
    echo "Action: Delete these temporary files before committing"
    echo ""
    ((ERRORS++)) || true
fi

# ============================================================================
# Check 3: Editor temporary files
# ============================================================================
echo "  [3/4] Checking for editor temporary files..."

EDITOR_TEMP=$(find . -type f \( \
    -name "*~" -o \
    -name "*.swp" -o \
    -name ".*.swp" -o \
    -name "*.swo" -o \
    -name "#*#" \
\) ! -path "./.git/*" 2>/dev/null | head -10 || true)

if [ -n "$EDITOR_TEMP" ]; then
    echo -e "${YELLOW}⚠️  WARNING: Editor temporary files found:${NC}"
    echo "$EDITOR_TEMP" | sed 's/^/  - /'
    echo ""
    echo "These files should be in .gitignore"
    echo "(Not blocking, but recommended to clean up)"
    echo ""
fi

# ============================================================================
# Check 4: Compiled binaries in root
# ============================================================================
echo "  [4/4] Checking for compiled binaries..."

BINARIES=$(find . -maxdepth 1 -type f \( \
    -name "meta-cc" -o \
    -name "meta-cc-mcp" -o \
    -name "*.exe" \
\) 2>/dev/null || true)

if [ -n "$BINARIES" ]; then
    echo -e "${YELLOW}⚠️  WARNING: Compiled binaries in root directory:${NC}"
    echo "$BINARIES" | sed 's/^/  - /'
    echo ""
    echo "These should be in .gitignore or build/"
    echo "(Not blocking, but verify they are not accidentally staged)"
    echo ""
fi

# ============================================================================
# Summary
# ============================================================================
echo ""
if [ "$ERRORS" -eq 0 ]; then
    echo -e "${GREEN}✅ No temporary files found${NC}"
    exit 0
else
    echo -e "${RED}❌ Found $ERRORS temporary file issue(s)${NC}"
    echo ""
    echo "Quick fix:"
    echo "  # Remove temporary .go files"
    echo "  find . -maxdepth 2 -name 'test_*.go' -o -name 'debug_*.go' | xargs rm -f"
    echo ""
    echo "  # Update .gitignore"
    echo "  echo 'test_*.go' >> .gitignore"
    echo "  echo 'debug_*.go' >> .gitignore"
    exit 1
fi
