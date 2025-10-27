#!/bin/bash
# E2E test for capability type parameter
# Tests list_capabilities and get_capability with type parameter
# Usage: ./tests/e2e/capability-type-e2e.sh [binary_path]

set -e

BINARY="${1:-./meta-cc-mcp}"
TESTS_PASSED=0
TESTS_FAILED=0

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
    exit 1
fi

# Helper: Extract JSON-RPC response (filter out log lines)
get_json_response() {
    local input="$1"
    # Look for lines starting with { and containing "jsonrpc"
    echo "$input" | grep -E '^\s*\{' | grep '"jsonrpc"' | head -1
}

# Helper: Send JSON-RPC request
send_request() {
    local request="$1"
    local timeout="${2:-5}"
    # Start fresh server for each request, capture output, extract JSON response
    local raw_output=$(echo "$request" | timeout "$timeout" "$BINARY" 2>&1)
    local response=$(get_json_response "$raw_output")
    if [ -z "$response" ]; then
        echo '{"error":{"code":-32603,"message":"No JSON-RPC response"}}'
    else
        echo "$response"
    fi
}

# Test function
test_capability_call() {
    local test_name="$1"
    local tool_name="$2"
    local arguments="$3"
    local validation_cmd="$4"

    echo -n "  $test_name... "

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

    # Run validation command
    if ! eval "$validation_cmd" <<< "$response"; then
        echo -e "${RED}FAILED${NC}"
        echo "      Response: $(echo "$response" | jq -c .)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi

    echo -e "${GREEN}PASSED${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

# Header
echo "=========================================="
echo "Capability Type Parameter E2E Test"
echo "=========================================="
echo "Binary: $BINARY"
echo "Time: $(date)"
echo ""

# Setup: Create test capability directories and files
echo -e "${BLUE}Setting up test environment...${NC}"
TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

mkdir -p "$TEST_DIR/capabilities/commands"
mkdir -p "$TEST_DIR/capabilities/prompts"

# Create test command capability
cat > "$TEST_DIR/capabilities/commands/test-command.md" <<'EOF'
---
name: test-command
description: Test command capability
category: test
---

This is a test command capability.
EOF

# Create test prompt capabilities
cat > "$TEST_DIR/capabilities/prompts/test-prompt.md" <<'EOF'
---
name: test-prompt
description: Test prompt capability
category: test
---

This is a test prompt capability.
EOF

cat > "$TEST_DIR/capabilities/prompts/meta-prompt-search.md" <<'EOF'
---
name: meta-prompt-search
description: Search historical prompts
category: internal
---

This is the search capability.
EOF

export META_CC_CAPABILITY_SOURCES="$TEST_DIR/capabilities"
echo "  Created test capabilities in: $TEST_DIR/capabilities"
echo ""

# Test Suite
echo -e "${BLUE}Test Suite 1: list_capabilities with type parameter${NC}"

test_capability_call \
    "list_capabilities() defaults to commands" \
    "list_capabilities" \
    '{}' \
    'jq -e ".result.content[0].text" | jq -e ".capabilities[] | select(.name == \"test-command\")" >/dev/null'

test_capability_call \
    "list_capabilities(type=commands) returns commands" \
    "list_capabilities" \
    '{"type": "commands"}' \
    'jq -e ".result.content[0].text" | jq -e ".capabilities[] | select(.name == \"test-command\")" >/dev/null'

test_capability_call \
    "list_capabilities(type=prompts) returns prompts" \
    "list_capabilities" \
    '{"type": "prompts"}' \
    'jq -e ".result.content[0].text" | jq -e ".capabilities[] | select(.name == \"test-prompt\")" >/dev/null'

test_capability_call \
    "list_capabilities(type=prompts) includes meta-prompt-search" \
    "list_capabilities" \
    '{"type": "prompts"}' \
    'jq -e ".result.content[0].text" | jq -e ".capabilities[] | select(.name == \"meta-prompt-search\")" >/dev/null'

echo ""
echo -e "${BLUE}Test Suite 2: get_capability with type parameter${NC}"

test_capability_call \
    "get_capability(name=test-command) defaults to commands" \
    "get_capability" \
    '{"name": "test-command"}' \
    'jq -e ".result.content[0].text" | grep -q "test command capability"'

test_capability_call \
    "get_capability(name=test-command, type=commands) returns command" \
    "get_capability" \
    '{"name": "test-command", "type": "commands"}' \
    'jq -e ".result.content[0].text" | grep -q "test command capability"'

test_capability_call \
    "get_capability(name=test-prompt, type=prompts) returns prompt" \
    "get_capability" \
    '{"name": "test-prompt", "type": "prompts"}' \
    'jq -e ".result.content[0].text" | grep -q "test prompt capability"'

test_capability_call \
    "get_capability(name=meta-prompt-search, type=prompts) returns internal" \
    "get_capability" \
    '{"name": "meta-prompt-search", "type": "prompts"}' \
    'jq -e ".result.content[0].text" | grep -q "Search historical prompts"'

echo ""
echo -e "${BLUE}Test Suite 3: Backward compatibility${NC}"

test_capability_call \
    "get_capability(name=prompts/test-prompt) auto-parses type" \
    "get_capability" \
    '{"name": "prompts/test-prompt"}' \
    'jq -e ".result.content[0].text" | grep -q "test prompt capability"'

test_capability_call \
    "get_capability(name=prompts/meta-prompt-search) auto-parses" \
    "get_capability" \
    '{"name": "prompts/meta-prompt-search"}' \
    'jq -e ".result.content[0].text" | grep -q "Search historical prompts"'

echo ""
echo -e "${BLUE}Test Suite 4: Error handling${NC}"

test_capability_call \
    "get_capability with invalid type returns error" \
    "get_capability" \
    '{"name": "test", "type": "invalid"}' \
    'jq -e ".error.message" | grep -qi "invalid.*type"'

test_capability_call \
    "get_capability with non-existent capability returns error" \
    "get_capability" \
    '{"name": "non-existent", "type": "prompts"}' \
    'jq -e ".error.message" | grep -qi "not found"'

# Summary
echo ""
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo -e "Total: $((TESTS_PASSED + TESTS_FAILED))"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Failed: $TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed.${NC}"
    exit 1
fi
