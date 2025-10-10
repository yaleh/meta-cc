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

# Test 5: Plugin source and author configuration
test_plugin_metadata() {
    # Check source (should be object with source=github and repo=yaleh/meta-cc)
    SOURCE_TYPE=$(jq -r '.plugins[0].source.source' .claude-plugin/marketplace.json)
    SOURCE_REPO=$(jq -r '.plugins[0].source.repo' .claude-plugin/marketplace.json)
    if [ "$SOURCE_TYPE" = "github" ] && [ "$SOURCE_REPO" = "yaleh/meta-cc" ]; then
        pass "Plugin source correct: github repo $SOURCE_REPO"
    else
        fail "Plugin source incorrect: got source=$SOURCE_TYPE, repo=$SOURCE_REPO"
    fi

    # Check author (should be object with name and email)
    AUTHOR_NAME=$(jq -r '.plugins[0].author.name' .claude-plugin/marketplace.json)
    AUTHOR_EMAIL=$(jq -r '.plugins[0].author.email' .claude-plugin/marketplace.json)

    if [ "$AUTHOR_NAME" != "null" ] && [ "$AUTHOR_EMAIL" != "null" ]; then
        pass "Plugin author object correct: $AUTHOR_NAME <$AUTHOR_EMAIL>"
    else
        fail "Plugin author format incorrect: expected object with name and email"
    fi
}

# Run all tests
echo "Running marketplace metadata validation tests..."
test_marketplace_json_exists
test_marketplace_json_valid
test_required_fields
test_version_sync
test_plugin_metadata

echo ""
echo -e "${GREEN}All marketplace metadata tests passed!${NC}"
