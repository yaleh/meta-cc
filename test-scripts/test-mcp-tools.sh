#!/bin/bash
# Test script for MCP Server tools (Stage 8.8)
# Usage: ./test-scripts/test-mcp-tools.sh

set -e

echo "=========================================="
echo "MCP Server Tools Test Suite"
echo "Stage 8.8 - Phase 8 Enhanced Tools"
echo "=========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counter
TESTS_RUN=0
TESTS_PASSED=0

# Function to run test
run_test() {
    local test_name="$1"
    local test_cmd="$2"
    local expected="$3"

    TESTS_RUN=$((TESTS_RUN + 1))
    echo -e "${BLUE}Test $TESTS_RUN: $test_name${NC}"

    result=$(eval "$test_cmd" 2>&1)

    if echo "$result" | grep -q "$expected"; then
        echo -e "${GREEN}✓ PASSED${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗ FAILED${NC}"
        echo "Expected: $expected"
        echo "Got: $result"
    fi
    echo ""
}

# Test 1: List all tools
echo "=== Phase 1: Tool Discovery ==="
run_test "List all MCP tools" \
    "echo '{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/list\"}' | ./meta-cc mcp | jq -r '.result.tools | length'" \
    "5"

run_test "Verify tool names" \
    "echo '{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/list\"}' | ./meta-cc mcp | jq -r '.result.tools[].name' | sort" \
    "query_user_messages"

# Test 2: extract_tools (Phase 8 enhanced)
echo "=== Phase 2: extract_tools (with pagination) ==="
run_test "extract_tools with limit=5" \
    "echo '{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"tools/call\",\"params\":{\"name\":\"extract_tools\",\"arguments\":{\"limit\":5}}}' | ./meta-cc mcp | jq -r '.result.content[0].text | fromjson | length'" \
    "5"

run_test "extract_tools with default limit" \
    "echo '{\"jsonrpc\":\"2.0\",\"id\":3,\"method\":\"tools/call\",\"params\":{\"name\":\"extract_tools\",\"arguments\":{}}}' | ./meta-cc mcp | jq -r '.result.content[0].text | fromjson | length <= 100'" \
    "true"

# Test 3: query_tools (Phase 8 new)
echo "=== Phase 3: query_tools (flexible filtering) ==="
run_test "query_tools with tool filter" \
    "echo '{\"jsonrpc\":\"2.0\",\"id\":4,\"method\":\"tools/call\",\"params\":{\"name\":\"query_tools\",\"arguments\":{\"tool\":\"Read\",\"limit\":3}}}' | ./meta-cc mcp | jq -r '.result.content[0].text | fromjson | .[0].ToolName'" \
    "Read"

run_test "query_tools with limit" \
    "echo '{\"jsonrpc\":\"2.0\",\"id\":5,\"method\":\"tools/call\",\"params\":{\"name\":\"query_tools\",\"arguments\":{\"limit\":10}}}' | ./meta-cc mcp | jq -r '.result.content[0].text | fromjson | length <= 10'" \
    "true"

# Test 4: query_user_messages (Phase 8 new)
echo "=== Phase 4: query_user_messages (regex search) ==="
run_test "query_user_messages with pattern" \
    "echo '{\"jsonrpc\":\"2.0\",\"id\":6,\"method\":\"tools/call\",\"params\":{\"name\":\"query_user_messages\",\"arguments\":{\"pattern\":\"Stage\",\"limit\":3}}}' | ./meta-cc mcp | jq -r '.result.content[0].text | fromjson | length <= 3'" \
    "true"

run_test "query_user_messages missing pattern (error)" \
    "echo '{\"jsonrpc\":\"2.0\",\"id\":7,\"method\":\"tools/call\",\"params\":{\"name\":\"query_user_messages\",\"arguments\":{}}}' | ./meta-cc mcp | jq -r '.error.code'" \
    "-32603"

# Test 5: Existing tools (backward compatibility)
echo "=== Phase 5: Existing Tools (backward compatibility) ==="
run_test "get_session_stats" \
    "echo '{\"jsonrpc\":\"2.0\",\"id\":8,\"method\":\"tools/call\",\"params\":{\"name\":\"get_session_stats\",\"arguments\":{}}}' | ./meta-cc mcp | jq -r '.result.content[0].text | fromjson | .TurnCount > 0'" \
    "true"

run_test "analyze_errors" \
    "echo '{\"jsonrpc\":\"2.0\",\"id\":9,\"method\":\"tools/call\",\"params\":{\"name\":\"analyze_errors\",\"arguments\":{}}}' | ./meta-cc mcp | jq -r '.result.content[0].text | fromjson | type'" \
    "array"

# Test 6: Output formats
echo "=== Phase 6: Output Format Support ==="
run_test "extract_tools with md format" \
    "echo '{\"jsonrpc\":\"2.0\",\"id\":10,\"method\":\"tools/call\",\"params\":{\"name\":\"extract_tools\",\"arguments\":{\"output_format\":\"md\",\"limit\":3}}}' | ./meta-cc mcp | jq -r '.result.content[0].text'" \
    "##"

# Summary
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo -e "Total Tests: $TESTS_RUN"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Failed: $((TESTS_RUN - TESTS_PASSED))${NC}"

if [ $TESTS_PASSED -eq $TESTS_RUN ]; then
    echo -e "\n${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}✗ Some tests failed${NC}"
    exit 1
fi
