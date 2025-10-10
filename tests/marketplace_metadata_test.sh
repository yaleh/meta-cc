#!/bin/bash
# meta-cc marketplace metadata validation tests

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

pass() {
    echo -e "${GREEN}✓${NC} $1"
}

fail() {
    echo -e "${RED}✗${NC} $1"
    exit 1
}

# Test 1: marketplace.json exists
test_marketplace_json_exists() {
    if [ -f .claude-plugin/marketplace.json ]; then
        pass "marketplace.json exists"
    else
        fail "marketplace.json not found"
    fi
}

# Test 2: Valid JSON syntax
test_marketplace_json_valid() {
    if jq empty .claude-plugin/marketplace.json 2>/dev/null; then
        pass "marketplace.json is valid JSON"
    else
        fail "marketplace.json has invalid JSON syntax"
    fi
}

# Test 3: Required fields present
test_required_fields() {
    REQUIRED_FIELDS="name version description repository assets installation"
    for field in $REQUIRED_FIELDS; do
        if jq -e ".$field" .claude-plugin/marketplace.json >/dev/null 2>&1; then
            pass "Required field present: $field"
        else
            fail "Missing required field: $field"
        fi
    done
}

# Test 4: Version sync with plugin.json
test_version_sync() {
    MARKETPLACE_VERSION=$(jq -r '.version' .claude-plugin/marketplace.json)
    PLUGIN_VERSION=$(jq -r '.version' plugin.json)

    if [ "$MARKETPLACE_VERSION" = "$PLUGIN_VERSION" ]; then
        pass "Version synchronized: $MARKETPLACE_VERSION"
    else
        fail "Version mismatch: marketplace=$MARKETPLACE_VERSION, plugin=$PLUGIN_VERSION"
    fi
}

# Test 5: Component counts accuracy
test_component_counts() {
    SLASH_COMMANDS=$(jq -r '.components.slash_commands' .claude-plugin/marketplace.json)
    SUBAGENTS=$(jq -r '.components.subagents' .claude-plugin/marketplace.json)
    MCP_TOOLS=$(jq -r '.components.mcp_tools' .claude-plugin/marketplace.json)

    if [ "$SLASH_COMMANDS" = "10" ] && [ "$SUBAGENTS" = "3" ] && [ "$MCP_TOOLS" = "14" ]; then
        pass "Component counts correct: $SLASH_COMMANDS commands, $SUBAGENTS subagents, $MCP_TOOLS tools"
    else
        fail "Component counts incorrect: got $SLASH_COMMANDS/$SUBAGENTS/$MCP_TOOLS, expected 10/3/14"
    fi
}

# Run all tests
echo "Running marketplace metadata validation tests..."
test_marketplace_json_exists
test_marketplace_json_valid
test_required_fields
test_version_sync
test_component_counts

echo ""
echo -e "${GREEN}All marketplace metadata tests passed!${NC}"
