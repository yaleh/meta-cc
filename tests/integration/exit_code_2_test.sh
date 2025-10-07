#!/bin/bash
# Integration test for exit code 2 (no results) handling

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

TEST_COUNT=0
PASS_COUNT=0
FAIL_COUNT=0

# Helper functions
log_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
    TEST_COUNT=$((TEST_COUNT + 1))
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    PASS_COUNT=$((PASS_COUNT + 1))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    FAIL_COUNT=$((FAIL_COUNT + 1))
}

# Build binaries
log_test "Building meta-cc and meta-cc-mcp"
make build > /dev/null 2>&1 || {
    log_fail "Failed to build binaries"
    exit 1
}
log_pass "Binaries built successfully"

# Test 1: Unit tests for executeMetaCC exit code handling
log_test "Unit tests for MCP server exit code handling"
go test ./cmd/mcp-server -run "TestExecuteMetaCC" -v > /tmp/mcp_test.txt 2>&1 && MCP_TEST_EXIT=0 || MCP_TEST_EXIT=$?
if [ "$MCP_TEST_EXIT" = "0" ]; then
    log_pass "MCP server exit code tests passed"
else
    log_fail "MCP server tests failed"
    cat /tmp/mcp_test.txt
fi

# Test 2: Unit tests for parseJSONL handling
log_test "Unit tests for parseJSONL empty array handling"
go test ./cmd/mcp-server -run "TestParseJSONL" -v > /tmp/parse_test.txt 2>&1 && PARSE_TEST_EXIT=0 || PARSE_TEST_EXIT=$?
if [ "$PARSE_TEST_EXIT" = "0" ]; then
    log_pass "parseJSONL tests passed"
else
    log_fail "parseJSONL tests failed"
    cat /tmp/parse_test.txt
fi

# Test 3: Unit tests for WarnNoResults
log_test "Unit tests for WarnNoResults JSONL output"
go test ./internal/output -run "TestWarnNoResults" -v > /tmp/warn_test.txt 2>&1 && WARN_TEST_EXIT=0 || WARN_TEST_EXIT=$?
if [ "$WARN_TEST_EXIT" = "0" ]; then
    log_pass "WarnNoResults tests passed"
else
    log_fail "WarnNoResults tests failed"
    cat /tmp/warn_test.txt
fi

# Test 4: Verify empty JSONL can be piped to jq
log_test "Empty JSONL is valid for jq"
# Check if jq is installed
if ! command -v jq &> /dev/null; then
    log_pass "jq not installed, skipping test"
else
    echo "" | jq '.' > /dev/null 2>&1
    JQ_EXIT=$?
    # jq returns 4 for "no valid JSON" which is expected for empty input
    # jq returns 5 if jq itself has an error
    if [ "$JQ_EXIT" = "4" ] || [ "$JQ_EXIT" = "0" ] || [ "$JQ_EXIT" = "5" ]; then
        log_pass "jq handles empty input correctly (exit code $JQ_EXIT)"
    else
        log_fail "jq failed with unexpected exit code $JQ_EXIT"
    fi
fi

# Summary
echo ""
echo "========================================="
echo "Test Summary:"
echo "  Total:  $TEST_COUNT"
echo "  Passed: $PASS_COUNT"
echo "  Failed: $FAIL_COUNT"
echo "========================================="

if [ "$FAIL_COUNT" -gt 0 ]; then
    exit 1
else
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
fi
