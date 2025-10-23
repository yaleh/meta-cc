#!/bin/bash
# check-[category].sh - [One-line description]
#
# Part of: Build Quality Gates
# Iteration: [P0/P1/P2]
# Purpose: [What problems this prevents]
# Historical Impact: [X% of errors this catches]
#
# shellcheck disable=SC2078,SC1073,SC1072,SC1123
# Note: This is a template file with placeholder syntax, not meant to be executed as-is

set -euo pipefail

# Colors for consistent output
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "Checking [category]..."

ERRORS=0
WARNINGS=0

# ============================================================================
# Check 1: [Specific check name]
# ============================================================================
echo "  [1/N] Checking [specific pattern]..."

# Your validation logic here
if [ condition ]; then
    echo -e "${RED}❌ ERROR: [Clear problem description]${NC}"
    echo "[Detailed explanation of what was found]"
    echo ""
    echo "To fix:"
    echo "  1. [Specific action step]"
    echo "  2. [Specific action step]"
    echo "  3. [Verification step]"
    echo ""
    ((ERRORS++)) || true
elif [ warning_condition ]; then
    echo -e "${YELLOW}⚠️  WARNING: [Warning description]${NC}"
    echo "[Optional improvement suggestion]"
    echo ""
    ((WARNINGS++)) || true
else
    echo -e "${GREEN}✓${NC} [Check passed]"
fi

# ============================================================================
# Continue with more checks...
# ============================================================================

# ============================================================================
# Summary
# ============================================================================
echo ""
if [ $ERRORS -eq 0 ]; then
    if [ $WARNINGS -eq 0 ]; then
        echo -e "${GREEN}✅ All [category] checks passed${NC}"
    else
        echo -e "${YELLOW}⚠️  All critical checks passed, $WARNINGS warning(s)${NC}"
    fi
    exit 0
else
    echo -e "${RED}❌ Found $ERRORS [category] error(s), $WARNINGS warning(s)${NC}"
    echo "Please fix errors before committing"
    exit 1
fi
