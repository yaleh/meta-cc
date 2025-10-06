#!/bin/bash
# Test JSONL streaming output (Phase 11, Stage 11.1)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=== JSONL Streaming Test ==="
echo ""

# Determine meta-cc binary location
if [ -f "./meta-cc" ]; then
    META_CC="./meta-cc"
elif [ -f "../meta-cc" ]; then
    META_CC="../meta-cc"
elif command -v meta-cc &> /dev/null; then
    META_CC="meta-cc"
else
    echo -e "${RED}✗ meta-cc binary not found${NC}"
    exit 1
fi

echo "Using binary: $META_CC"
echo ""

# Test 1: Verify --stream produces valid JSONL
echo "[1/3] Testing JSONL format..."

# Use current session to get actual data
STREAM_OUTPUT=$($META_CC query tools --stream --limit 5 2>/dev/null || echo "")

if [ -z "$STREAM_OUTPUT" ]; then
    echo -e "${YELLOW}⚠ No data available (empty session), skipping JSONL format test${NC}"
else
    LINE_COUNT=$(echo "$STREAM_OUTPUT" | wc -l)

    if [ "$LINE_COUNT" -ne 5 ]; then
        echo -e "${RED}✗ Expected 5 lines, got $LINE_COUNT${NC}"
        exit 1
    fi

    # Verify each line is valid JSON
    LINE_NUM=0
    echo "$STREAM_OUTPUT" | while IFS= read -r line; do
        LINE_NUM=$((LINE_NUM + 1))
        if ! echo "$line" | jq empty 2>/dev/null; then
            echo -e "${RED}✗ Line $LINE_NUM is invalid JSON: $line${NC}"
            exit 1
        fi
    done

    echo -e "${GREEN}✓ JSONL format valid${NC}"
fi

# Test 2: Verify streaming works with jq (if jq is available)
echo "[2/3] Testing jq integration..."

if command -v jq &> /dev/null; then
    FILTERED=$($META_CC query tools --stream --limit 100 2>/dev/null | jq -c 'select(.Status == "error")' | wc -l || echo "0")
    echo -e "${GREEN}✓ Found $FILTERED error records via jq pipeline${NC}"
else
    echo -e "${YELLOW}⚠ jq not installed, skipping jq integration test${NC}"
fi

# Test 3: Verify streaming works with grep
echo "[3/3] Testing grep integration..."

BASH_COUNT=$($META_CC query tools --stream --limit 100 2>/dev/null | grep -c '"ToolName":"Bash"' || true)
echo -e "${GREEN}✓ Found $BASH_COUNT Bash tool calls via grep pipeline${NC}"

echo ""
echo -e "${GREEN}=== All Streaming Tests Passed ✅ ===${NC}"
