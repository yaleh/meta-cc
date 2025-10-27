#!/bin/bash
# Simple capability type parameter E2E test
# Tests the new type parameter for list_capabilities and get_capability

set -e

BINARY="${1:-./meta-cc-mcp}"
TESTS_PASSED=0
TESTS_FAILED=0

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo "=========================================="
echo "Capability Type Parameter Test"
echo "=========================================="
echo ""

# Setup test environment
TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

mkdir -p "$TEST_DIR/capabilities/commands"
mkdir -p "$TEST_DIR/capabilities/prompts"

cat > "$TEST_DIR/capabilities/commands/test-command.md" <<'EOF'
---
name: test-command
description: Test command capability
category: test
---
Test command content.
EOF

cat > "$TEST_DIR/capabilities/prompts/test-prompt.md" <<'EOF'
---
name: test-prompt
description: Test prompt capability
category: test
---
Test prompt content.
EOF

export META_CC_CAPABILITY_SOURCES="$TEST_DIR/capabilities"

# Test 1: list_capabilities defaults to commands
echo -n "Test 1: list_capabilities() defaults to commands... "
RESPONSE=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_capabilities","arguments":{}}}' | "$BINARY" 2>&1 | grep '"jsonrpc"' | head -1)
# Extract nested JSON from .result.content[0].text and check for capability name
if command -v jq &> /dev/null; then
    if echo "$RESPONSE" | jq -r '.result.content[0].text' 2>/dev/null | grep -q 'test-command'; then
        echo -e "${GREEN}PASSED${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}FAILED${NC}"
        echo "  Response excerpt: $(echo "$RESPONSE" | cut -c1-200)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
else
    # Fallback if jq not available - just check if response contains the name anywhere
    if echo "$RESPONSE" | grep -q 'test-command'; then
        echo -e "${GREEN}PASSED${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}FAILED${NC}"
        echo "  Response excerpt: $(echo "$RESPONSE" | cut -c1-200)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
fi

# Test 2: list_capabilities(type=prompts) returns prompts
echo -n "Test 2: list_capabilities(type=prompts) returns prompts... "
RESPONSE=$(echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_capabilities","arguments":{"type":"prompts"}}}' | "$BINARY" 2>&1 | grep '"jsonrpc"' | head -1)
# Extract nested JSON from .result.content[0].text and check for capability name
if command -v jq &> /dev/null; then
    if echo "$RESPONSE" | jq -r '.result.content[0].text' 2>/dev/null | grep -q 'test-prompt'; then
        echo -e "${GREEN}PASSED${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}FAILED${NC}"
        echo "  Response excerpt: $(echo "$RESPONSE" | cut -c1-200)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
else
    # Fallback if jq not available - just check if response contains the name anywhere
    if echo "$RESPONSE" | grep -q 'test-prompt'; then
        echo -e "${GREEN}PASSED${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}FAILED${NC}"
        echo "  Response excerpt: $(echo "$RESPONSE" | cut -c1-200)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
fi

# Test 3: get_capability defaults to commands
echo -n "Test 3: get_capability(name=test-command) defaults to commands... "
RESPONSE=$(echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_capability","arguments":{"name":"test-command"}}}' | "$BINARY" 2>&1 | grep '"jsonrpc"' | head -1)
if echo "$RESPONSE" | grep -q 'Test command content'; then
    echo -e "${GREEN}PASSED${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAILED${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

# Test 4: get_capability with type=prompts
echo -n "Test 4: get_capability(name=test-prompt, type=prompts)... "
RESPONSE=$(echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_capability","arguments":{"name":"test-prompt","type":"prompts"}}}' | "$BINARY" 2>&1 | grep '"jsonrpc"' | head -1)
if echo "$RESPONSE" | grep -q 'Test prompt content'; then
    echo -e "${GREEN}PASSED${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAILED${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

# Test 5: backward compatibility with "prompts/name" format
echo -n "Test 5: get_capability(name=prompts/test-prompt) backward compat... "
RESPONSE=$(echo '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"get_capability","arguments":{"name":"prompts/test-prompt"}}}' | "$BINARY" 2>&1 | grep '"jsonrpc"' | head -1)
if echo "$RESPONSE" | grep -q 'Test prompt content'; then
    echo -e "${GREEN}PASSED${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAILED${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

# Test 6: invalid type returns error
echo -n "Test 6: get_capability with invalid type returns error... "
RESPONSE=$(echo '{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"get_capability","arguments":{"name":"test","type":"invalid"}}}' | "$BINARY" 2>&1 | grep '"jsonrpc"' | head -1)
if echo "$RESPONSE" | grep -q '"error"' && echo "$RESPONSE" | grep -qi 'invalid.*type'; then
    echo -e "${GREEN}PASSED${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAILED${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

echo ""
echo "=========================================="
echo -e "Total: $((TESTS_PASSED + TESTS_FAILED))"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Failed: $TESTS_FAILED${NC}"
echo "=========================================="

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed.${NC}"
    exit 1
fi
