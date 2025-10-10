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

# Test 3: Required fields present (marketplace schema)
test_required_fields() {
    # Marketplace-level required fields
    REQUIRED_FIELDS="name owner plugins"
    for field in $REQUIRED_FIELDS; do
        if jq -e ".$field" .claude-plugin/marketplace.json >/dev/null 2>&1; then
            pass "Required field present: $field"
        else
            fail "Missing required field: $field"
        fi
    done

    # Owner required fields
    OWNER_FIELDS="name email"
    for field in $OWNER_FIELDS; do
        if jq -e ".owner.$field" .claude-plugin/marketplace.json >/dev/null 2>&1; then
            pass "Required owner field present: $field"
        else
            fail "Missing required owner field: $field"
        fi
    done

    # Plugin required fields (first plugin)
    PLUGIN_FIELDS="name source"
    for field in $PLUGIN_FIELDS; do
        if jq -e ".plugins[0].$field" .claude-plugin/marketplace.json >/dev/null 2>&1; then
            pass "Required plugin field present: $field"
        else
            fail "Missing required plugin field: $field"
        fi
    done
}

# Test 4: Version sync with plugin.json
test_version_sync() {
    MARKETPLACE_VERSION=$(jq -r '.plugins[0].version' .claude-plugin/marketplace.json)
    PLUGIN_VERSION=$(jq -r '.version' plugin.json)

    if [ "$MARKETPLACE_VERSION" = "$PLUGIN_VERSION" ]; then
        pass "Version synchronized: $MARKETPLACE_VERSION"
    else
        fail "Version mismatch: marketplace=$MARKETPLACE_VERSION, plugin=$PLUGIN_VERSION"
    fi
}

# Test 5: Plugin source configuration
test_plugin_source() {
    SOURCE_TYPE=$(jq -r '.plugins[0].source.type' .claude-plugin/marketplace.json)
    SOURCE_URL=$(jq -r '.plugins[0].source.url' .claude-plugin/marketplace.json)

    if [ "$SOURCE_TYPE" = "github" ]; then
        pass "Plugin source type correct: $SOURCE_TYPE"
    else
        fail "Plugin source type incorrect: got $SOURCE_TYPE, expected github"
    fi

    if [ "$SOURCE_URL" = "https://github.com/yaleh/meta-cc" ]; then
        pass "Plugin source URL correct: $SOURCE_URL"
    else
        fail "Plugin source URL incorrect: $SOURCE_URL"
    fi
}

# Run all tests
echo "Running marketplace metadata validation tests..."
test_marketplace_json_exists
test_marketplace_json_valid
test_required_fields
test_version_sync
test_plugin_source

echo ""
echo -e "${GREEN}All marketplace metadata tests passed!${NC}"
