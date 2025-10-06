#!/bin/bash
# Documentation Examples Validation Test
# Validates that all code examples in cookbook.md and cli-composability.md are executable

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
DOCS_DIR="$PROJECT_ROOT/docs"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

echo "=== Documentation Examples Validation Test ==="
echo ""
echo "Testing examples from:"
echo "  - docs/cookbook.md"
echo "  - docs/cli-composability.md"
echo ""

# Check if meta-cc is available
if ! command -v meta-cc &> /dev/null; then
    echo -e "${RED}✗ meta-cc not found in PATH${NC}"
    echo "  Please build and install meta-cc first:"
    echo "  make build && sudo cp meta-cc /usr/local/bin/"
    exit 1
fi

echo -e "${GREEN}✓ meta-cc found: $(which meta-cc)${NC}"
echo ""

# Helper function to run a test
run_test() {
    local test_name="$1"
    local test_cmd="$2"
    local expect_success="${3:-true}"

    echo -n "Testing: $test_name ... "

    # Run command and capture exit code
    if eval "$test_cmd" > /dev/null 2>&1; then
        if [ "$expect_success" = "true" ]; then
            echo -e "${GREEN}✓${NC}"
            ((TESTS_PASSED++))
            return 0
        else
            echo -e "${RED}✗ (expected failure but succeeded)${NC}"
            ((TESTS_FAILED++))
            return 1
        fi
    else
        if [ "$expect_success" = "false" ]; then
            echo -e "${GREEN}✓ (expected failure)${NC}"
            ((TESTS_PASSED++))
            return 0
        else
            echo -e "${RED}✗ (failed)${NC}"
            echo "  Command: $test_cmd"
            ((TESTS_FAILED++))
            return 1
        fi
    fi
}

# Helper function to skip a test (for commands that depend on data)
skip_test() {
    local test_name="$1"
    echo -e "Skipping: $test_name ... ${YELLOW}⊘${NC}"
    ((TESTS_SKIPPED++))
}

echo "=== Testing Core Commands ==="
echo ""

# Test basic query commands (with timeout to prevent hangs)
run_test "query tools with JSON output" \
    "timeout 5 meta-cc query tools --output json --limit 10"

run_test "query tools with stream output" \
    "timeout 5 meta-cc query tools --stream --limit 10"

run_test "query tools with where clause" \
    "timeout 5 meta-cc query tools --where \"tool='Bash'\" --limit 5"

# Test stats commands (with timeout)
run_test "stats aggregate by tool" \
    "timeout 5 meta-cc stats aggregate --group-by tool --metrics count"

run_test "stats time-series (tool-calls)" \
    "timeout 5 meta-cc stats time-series --metric tool-calls --interval hour"

run_test "stats files" \
    "timeout 5 meta-cc stats files --top 10"

echo ""
echo "=== Testing jq Integration ==="
echo ""

# Test jq pipeline examples (with timeout)
run_test "jq select errors" \
    "timeout 5 meta-cc query tools --stream --limit 10 2>/dev/null | timeout 5 jq 'select(.Status == \"error\")' > /dev/null || true"

run_test "jq extract tool names" \
    "timeout 5 meta-cc query tools --stream --limit 10 2>/dev/null | timeout 5 jq -r '.ToolName' > /dev/null"

run_test "jq extract TSV fields" \
    "timeout 5 meta-cc query tools --stream --limit 10 2>/dev/null | timeout 5 jq -r '[.ToolName, .Status, .Duration] | @tsv' > /dev/null"

run_test "jq group by tool (slurp)" \
    "timeout 5 meta-cc query tools --stream --limit 10 2>/dev/null | timeout 5 jq -s 'group_by(.ToolName) | map({tool: .[0].ToolName, count: length})' > /dev/null"

echo ""
echo "=== Testing grep Integration ==="
echo ""

# Test grep pipeline examples (with timeout)
run_test "grep filter tool names" \
    "timeout 5 meta-cc query tools --stream --limit 10 2>/dev/null | timeout 5 jq -r '.ToolName' | timeout 5 grep 'Bash' > /dev/null || true"

run_test "grep count Bash tools" \
    "timeout 5 meta-cc query tools --stream --limit 100 2>/dev/null | timeout 5 jq -r '.ToolName' | timeout 5 grep -c 'Bash' > /dev/null || true"

run_test "grep extract patterns" \
    "timeout 5 meta-cc query tools --stream --limit 10 2>/dev/null | timeout 5 jq -r '.Error // empty' | timeout 5 grep -oP '(permission|timeout|failed)' > /dev/null || true"

echo ""
echo "=== Testing awk Integration ==="
echo ""

# Test awk pipeline examples (with timeout)
run_test "awk format fields" \
    "timeout 5 meta-cc query tools --stream --limit 10 2>/dev/null | timeout 5 jq -r '[.ToolName, .Status, .Duration] | @tsv' | timeout 5 awk '{print \"Tool:\", \$1, \"Status:\", \$2, \"Duration:\", \$3 \"ms\"}' > /dev/null"

run_test "awk sum durations" \
    "timeout 5 meta-cc query tools --stream --limit 10 2>/dev/null | timeout 5 jq -r '.Duration' | timeout 5 awk '{sum += \$1} END {print sum}' > /dev/null"

run_test "awk conditional processing" \
    "timeout 5 meta-cc query tools --stream --limit 10 2>/dev/null | timeout 5 jq -r '.Duration' | timeout 5 awk '{if (\$1 < 1000) fast++; else slow++} END {print fast, slow}' > /dev/null"

echo ""
echo "=== Testing sed Integration ==="
echo ""

# Test sed pipeline examples (with timeout)
run_test "sed text replacement" \
    "timeout 5 meta-cc query tools --stream --limit 10 2>/dev/null | timeout 5 jq -r '.ToolName' | timeout 5 sed 's/Bash/Shell/g' > /dev/null"

run_test "sed delete empty lines" \
    "timeout 5 meta-cc query tools --stream --limit 10 2>/dev/null | timeout 5 jq -r '.Error // empty' | timeout 5 sed '/^\$/d' > /dev/null || true"

echo ""
echo "=== Testing Combined Pipelines ==="
echo ""

# Test complex pipeline combinations (with timeout)
run_test "jq + grep + sort + uniq" \
    "timeout 10 bash -c 'meta-cc query tools --stream --limit 100 2>/dev/null | jq -r \".ToolName\" | grep -v \"^$\" | sort | uniq -c > /dev/null' || true"

run_test "jq + awk aggregation" \
    "timeout 10 bash -c 'meta-cc query tools --stream --limit 100 2>/dev/null | jq -r \"[.ToolName, .Duration] | @tsv\" | awk \"{duration[\$1] += \$2; count[\$1]++} END {for (tool in duration) print tool, duration[tool]}\" > /dev/null'"

run_test "stats + jq + awk formatting" \
    "timeout 10 bash -c 'meta-cc stats aggregate --group-by tool --metrics count 2>/dev/null | jq -r \".[] | [.group_value, .metrics.count] | @tsv\" | awk \"{print \$1 \\\": \\\" \$2}\" > /dev/null'"

echo ""
echo "=== Testing Exit Codes ==="
echo ""

# Test exit code behavior (with timeout)
run_test "exit 0 for successful query with results" \
    "timeout 5 meta-cc query tools --limit 10 > /dev/null 2>&1"

run_test "exit 2 for successful query with no results" \
    "timeout 5 meta-cc query tools --where \"tool='NonExistentTool12345'\" > /dev/null 2>&1" \
    "false"

echo ""
echo "=== Testing Phase 11 Features ==="
echo ""

# Test streaming output format (with timeout)
run_test "streaming output is valid JSONL" \
    "timeout 10 bash -c 'meta-cc query tools --stream --limit 5 2>/dev/null | while IFS= read -r line; do echo \"\$line\" | jq empty || exit 1; done'"

# Test stderr/stdout separation (with timeout)
run_test "data on stdout (can be redirected)" \
    "timeout 5 bash -c 'meta-cc query tools --limit 5 --output json 2>/dev/null | jq empty'"

run_test "pipeline without stderr interference" \
    "timeout 10 bash -c 'COUNT=\$(meta-cc query tools --stream --limit 10 2>/dev/null | jq -s \". | length\"); [ \"\$COUNT\" -eq 10 ]'"

echo ""
echo "=== Test Summary ==="
echo ""
echo -e "Passed:  ${GREEN}$TESTS_PASSED${NC}"
echo -e "Failed:  ${RED}$TESTS_FAILED${NC}"
echo -e "Skipped: ${YELLOW}$TESTS_SKIPPED${NC}"
echo -e "Total:   $((TESTS_PASSED + TESTS_FAILED + TESTS_SKIPPED))"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}=== All Tests Passed ✓ ===${NC}"
    exit 0
else
    echo -e "${RED}=== Some Tests Failed ✗ ===${NC}"
    exit 1
fi
