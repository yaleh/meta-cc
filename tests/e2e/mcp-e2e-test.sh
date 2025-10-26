#!/bin/bash
# E2E test suite for meta-cc-mcp server
# Usage: ./tests/e2e/mcp-e2e-test.sh [binary_path]
#
# This script tests the MCP server by sending JSON-RPC requests via stdin
# and validating responses from stdout.

set -e

BINARY="${1:-./meta-cc-mcp}"
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Check if binary exists
if [ ! -f "$BINARY" ]; then
    echo -e "${RED}Error: Binary not found: $BINARY${NC}"
    echo "Usage: $0 [binary_path]"
    exit 1
fi

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo -e "${RED}Error: jq is not installed${NC}"
    echo "Install with: sudo apt-get install jq (or brew install jq on macOS)"
    exit 1
fi

# Helper: Send JSON-RPC request with timeout
send_request() {
    local request="$1"
    local timeout="${2:-5}"

    # Use timeout to prevent hanging
    echo "$request" | timeout "$timeout" "$BINARY" 2>/dev/null || {
        echo '{"error":{"code":-32603,"message":"Request timeout"}}'
    }
}

# Helper: Test tool call
test_tool_call() {
    local test_name="$1"
    local tool_name="$2"
    local arguments="$3"
    local expected_pattern="$4"

    echo -n "  Testing $test_name... "

    local request=$(cat <<EOF
{
  "jsonrpc": "2.0",
  "id": $RANDOM,
  "method": "tools/call",
  "params": {
    "name": "$tool_name",
    "arguments": $arguments
  }
}
EOF
)

    local response=$(send_request "$request")

    # Check for JSON-RPC error
    if echo "$response" | jq -e '.error' >/dev/null 2>&1; then
        local error_msg=$(echo "$response" | jq -r '.error.message')
        echo -e "${RED}FAILED${NC}"
        echo "      Error: $error_msg"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi

    # Check for expected pattern
    if [ -n "$expected_pattern" ]; then
        if ! echo "$response" | grep -q "$expected_pattern"; then
            echo -e "${RED}FAILED${NC}"
            echo "      Expected pattern not found: $expected_pattern"
            echo "      Response: $(echo "$response" | jq -c .)"
            TESTS_FAILED=$((TESTS_FAILED + 1))
            return 1
        fi
    fi

    echo -e "${GREEN}PASSED${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

# Header
echo "=========================================="
echo "MCP E2E Test Suite"
echo "=========================================="
echo "Binary: $BINARY"
echo "Time: $(date)"
echo "=========================================="
echo ""

# Test 0: Server initialization
echo -e "${BLUE}=== Phase 0: Server Initialization ===${NC}"
echo -n "  Testing server response... "
INIT_RESPONSE=$(send_request '{"jsonrpc":"2.0","id":0,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}')

if echo "$INIT_RESPONSE" | jq -e '.result' >/dev/null 2>&1; then
    echo -e "${GREEN}PASSED${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAILED${NC}"
    echo "      Server did not respond to initialize"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Test 1: Tool discovery
echo -e "${BLUE}=== Phase 1: Tool Discovery ===${NC}"
TOOLS_LIST=$(send_request '{"jsonrpc":"2.0","id":1,"method":"tools/list"}')
TOOL_COUNT=$(echo "$TOOLS_LIST" | jq -r '.result.tools | length' 2>/dev/null || echo 0)

echo -n "  Testing tools/list... "
if [ "$TOOL_COUNT" -gt 0 ]; then
    echo -e "${GREEN}PASSED${NC}"
    echo "      Found $TOOL_COUNT tools"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAILED${NC}"
    echo "      No tools found"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

# List discovered tools
echo "  Available tools:"
echo "$TOOLS_LIST" | jq -r '.result.tools[].name' 2>/dev/null | sed 's/^/    - /' || echo "    (unable to list)"
echo ""

# Test 2: Legacy tools (backward compatibility)
echo -e "${BLUE}=== Phase 2: Legacy Tools (Backward Compatibility) ===${NC}"

# Check if query_user_messages exists
if echo "$TOOLS_LIST" | jq -e '.result.tools[] | select(.name == "query_user_messages")' >/dev/null 2>&1; then
    test_tool_call "query_user_messages" "query_user_messages" '{"pattern":".*","limit":5}' "result"
else
    echo -e "  ${YELLOW}SKIPPED${NC} query_user_messages (not found)"
    TESTS_SKIPPED=$((TESTS_SKIPPED + 1))
fi

# Check if query_tools exists
if echo "$TOOLS_LIST" | jq -e '.result.tools[] | select(.name == "query_tools")' >/dev/null 2>&1; then
    test_tool_call "query_tools" "query_tools" '{"limit":5}' "result"
else
    echo -e "  ${YELLOW}SKIPPED${NC} query_tools (not found)"
    TESTS_SKIPPED=$((TESTS_SKIPPED + 1))
fi

# Check if query_tool_errors exists
if echo "$TOOLS_LIST" | jq -e '.result.tools[] | select(.name == "query_tool_errors")' >/dev/null 2>&1; then
    test_tool_call "query_tool_errors" "query_tool_errors" '{"limit":5}' "result"
else
    echo -e "  ${YELLOW}SKIPPED${NC} query_tool_errors (not found)"
    TESTS_SKIPPED=$((TESTS_SKIPPED + 1))
fi
echo ""

# Test 3: Phase 27 new tools
echo -e "${BLUE}=== Phase 3: Phase 27 New Tools ===${NC}"

# Test get_session_directory
if echo "$TOOLS_LIST" | jq -e '.result.tools[] | select(.name == "get_session_directory")' >/dev/null 2>&1; then
    test_tool_call "get_session_directory (project scope)" "get_session_directory" '{"scope":"project"}' "directory"
    test_tool_call "get_session_directory (session scope)" "get_session_directory" '{"scope":"session"}' "directory"
else
    echo -e "  ${YELLOW}SKIPPED${NC} get_session_directory (not implemented yet)"
    TESTS_SKIPPED=$((TESTS_SKIPPED + 2))
fi

# Test inspect_session_files
if echo "$TOOLS_LIST" | jq -e '.result.tools[] | select(.name == "inspect_session_files")' >/dev/null 2>&1; then
    # Get session directory first
    SESSION_DIR_RESPONSE=$(send_request '{"jsonrpc":"2.0","id":100,"method":"tools/call","params":{"name":"get_session_directory","arguments":{"scope":"project"}}}')
    SESSION_DIR=$(echo "$SESSION_DIR_RESPONSE" | jq -r '.result.content[0].text' 2>/dev/null | jq -r '.directory' 2>/dev/null || echo "")

    if [ -n "$SESSION_DIR" ] && [ -d "$SESSION_DIR" ]; then
        # Get up to 3 recent files
        FILES_ARRAY=$(ls -t "$SESSION_DIR"/*.jsonl 2>/dev/null | head -3 | jq -R . | jq -s . || echo "[]")
        FILE_COUNT=$(echo "$FILES_ARRAY" | jq 'length')

        if [ "$FILE_COUNT" -gt 0 ]; then
            INSPECT_ARGS="{\"files\":$FILES_ARRAY,\"include_samples\":false}"
            test_tool_call "inspect_session_files" "inspect_session_files" "$INSPECT_ARGS" "files"
        else
            echo -e "  ${YELLOW}SKIPPED${NC} inspect_session_files (no session files found)"
            TESTS_SKIPPED=$((TESTS_SKIPPED + 1))
        fi
    else
        echo -e "  ${YELLOW}SKIPPED${NC} inspect_session_files (no session directory)"
        TESTS_SKIPPED=$((TESTS_SKIPPED + 1))
    fi
else
    echo -e "  ${YELLOW}SKIPPED${NC} inspect_session_files (not implemented yet)"
    TESTS_SKIPPED=$((TESTS_SKIPPED + 1))
fi
echo ""

# Test 4: Stage 2 query
echo -e "${BLUE}=== Phase 4: Stage 2 Query ===${NC}"

if echo "$TOOLS_LIST" | jq -e '.result.tools[] | select(.name == "execute_stage2_query")' >/dev/null 2>&1; then
    # Get session directory
    SESSION_DIR_RESPONSE=$(send_request '{"jsonrpc":"2.0","id":101,"method":"tools/call","params":{"name":"get_session_directory","arguments":{"scope":"project"}}}')
    SESSION_DIR=$(echo "$SESSION_DIR_RESPONSE" | jq -r '.result.content[0].text' 2>/dev/null | jq -r '.directory' 2>/dev/null || echo "")

    if [ -n "$SESSION_DIR" ] && [ -d "$SESSION_DIR" ]; then
        FILES_ARRAY=$(ls -t "$SESSION_DIR"/*.jsonl 2>/dev/null | head -3 | jq -R . | jq -s . || echo "[]")
        FILE_COUNT=$(echo "$FILES_ARRAY" | jq 'length')

        if [ "$FILE_COUNT" -gt 0 ]; then
            # Test basic query
            STAGE2_ARGS="{\"files\":$FILES_ARRAY,\"filter\":\"select(.type == \\\"user\\\")\",\"limit\":5}"
            test_tool_call "execute_stage2_query (basic)" "execute_stage2_query" "$STAGE2_ARGS" "results"

            # Test with sort
            STAGE2_ARGS_SORT="{\"files\":$FILES_ARRAY,\"filter\":\"select(.type == \\\"user\\\")\",\"sort\":\"sort_by(.timestamp)\",\"limit\":5}"
            test_tool_call "execute_stage2_query (with sort)" "execute_stage2_query" "$STAGE2_ARGS_SORT" "results"

            # Test with transform
            STAGE2_ARGS_TRANSFORM="{\"files\":$FILES_ARRAY,\"filter\":\"select(.type == \\\"user\\\")\",\"transform\":\"\\\"\\\\(.timestamp[:19])\\\"\",\"limit\":3}"
            test_tool_call "execute_stage2_query (with transform)" "execute_stage2_query" "$STAGE2_ARGS_TRANSFORM" "results"
        else
            echo -e "  ${YELLOW}SKIPPED${NC} execute_stage2_query (no session files)"
            TESTS_SKIPPED=$((TESTS_SKIPPED + 3))
        fi
    else
        echo -e "  ${YELLOW}SKIPPED${NC} execute_stage2_query (no session directory)"
        TESTS_SKIPPED=$((TESTS_SKIPPED + 3))
    fi
else
    echo -e "  ${YELLOW}SKIPPED${NC} execute_stage2_query (not implemented yet)"
    TESTS_SKIPPED=$((TESTS_SKIPPED + 3))
fi
echo ""

# Test 5: Error handling
echo -e "${BLUE}=== Phase 5: Error Handling ===${NC}"

# Test invalid tool name
echo -n "  Testing invalid tool name... "
INVALID_TOOL_RESPONSE=$(send_request '{"jsonrpc":"2.0","id":200,"method":"tools/call","params":{"name":"nonexistent_tool","arguments":{}}}')
if echo "$INVALID_TOOL_RESPONSE" | jq -e '.error' >/dev/null 2>&1; then
    echo -e "${GREEN}PASSED${NC}"
    echo "      Error correctly returned"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAILED${NC}"
    echo "      Should return error for invalid tool"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

# Test invalid method
echo -n "  Testing invalid method... "
INVALID_METHOD_RESPONSE=$(send_request '{"jsonrpc":"2.0","id":201,"method":"invalid/method","params":{}}')
if echo "$INVALID_METHOD_RESPONSE" | jq -e '.error' >/dev/null 2>&1; then
    echo -e "${GREEN}PASSED${NC}"
    echo "      Error correctly returned"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAILED${NC}"
    echo "      Should return error for invalid method"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Summary
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo -e "Total Tests: $((TESTS_PASSED + TESTS_FAILED + TESTS_SKIPPED))"
echo -e "  ${GREEN}Passed:${NC}  $TESTS_PASSED"
echo -e "  ${RED}Failed:${NC}  $TESTS_FAILED"
echo -e "  ${YELLOW}Skipped:${NC} $TESTS_SKIPPED"
echo "=========================================="

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "\n${GREEN}✓ All tests passed!${NC}\n"
    exit 0
else
    echo -e "\n${RED}✗ Some tests failed${NC}\n"
    exit 1
fi
