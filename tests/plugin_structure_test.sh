#!/bin/bash
# Test suite for Stage 20.1: Plugin Structure
# Tests plugin.json validation and .claude/ directory structure

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
pass() {
    echo -e "${GREEN}✓${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

fail() {
    echo -e "${RED}✗${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

test_start() {
    TESTS_RUN=$((TESTS_RUN + 1))
    echo -n "Testing: $1... "
}

# Test 1: Verify plugin.json exists
test_plugin_json_exists() {
    test_start "plugin.json exists"
    if [ -f "plugin.json" ]; then
        pass "plugin.json exists"
    else
        fail "plugin.json not found"
    fi
}

# Test 2: Validate JSON syntax
test_plugin_json_valid() {
    test_start "plugin.json is valid JSON"
    if jq empty plugin.json 2>/dev/null; then
        pass "plugin.json is valid JSON"
    else
        fail "plugin.json has invalid JSON syntax"
    fi
}

# Test 3: Check SemVer version format
test_plugin_version_semver() {
    test_start "version follows SemVer 2.0"
    VERSION=$(jq -r '.version' plugin.json 2>/dev/null)
    if echo "$VERSION" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$'; then
        pass "version is valid SemVer: $VERSION"
    else
        fail "version is not valid SemVer: $VERSION"
    fi
}

# Test 4: Verify all required fields
test_required_fields() {
    test_start "all required fields present"
    REQUIRED_FIELDS=("name" "version" "description" "author" "platforms" "binaries" "integration" "install")
    MISSING_FIELDS=""

    for field in "${REQUIRED_FIELDS[@]}"; do
        if ! jq -e ".$field" plugin.json >/dev/null 2>&1; then
            MISSING_FIELDS="$MISSING_FIELDS $field"
        fi
    done

    if [ -z "$MISSING_FIELDS" ]; then
        pass "all required fields present"
    else
        fail "missing fields:$MISSING_FIELDS"
    fi
}

# Test 5: Validate .claude/ directory structure
test_directory_structure() {
    test_start ".claude/ directory structure"
    MISSING_DIRS=""

    if [ ! -d ".claude/commands" ]; then
        MISSING_DIRS="$MISSING_DIRS .claude/commands"
    fi

    if [ ! -d ".claude/agents" ]; then
        MISSING_DIRS="$MISSING_DIRS .claude/agents"
    fi

    if [ ! -d ".claude/lib" ]; then
        MISSING_DIRS="$MISSING_DIRS .claude/lib"
    fi

    if [ -z "$MISSING_DIRS" ]; then
        pass ".claude/ directory structure is correct"
    else
        fail "missing directories:$MISSING_DIRS"
    fi
}

# Test 6: Validate MCP config template exists and is valid JSON
test_mcp_config_template() {
    test_start "MCP config template"
    if [ -f ".claude/lib/mcp-config.json" ]; then
        if jq empty .claude/lib/mcp-config.json 2>/dev/null; then
            pass "MCP config template is valid JSON"
        else
            fail "MCP config template has invalid JSON syntax"
        fi
    else
        fail "MCP config template not found"
    fi
}

# Test 7: Verify platforms list is complete
test_platforms_list() {
    test_start "platforms list completeness"
    EXPECTED_PLATFORMS=("linux-amd64" "linux-arm64" "darwin-amd64" "darwin-arm64" "windows-amd64")
    PLATFORMS=$(jq -r '.platforms[]' plugin.json 2>/dev/null)

    MISSING_PLATFORMS=""
    for platform in "${EXPECTED_PLATFORMS[@]}"; do
        if ! echo "$PLATFORMS" | grep -q "^$platform$"; then
            MISSING_PLATFORMS="$MISSING_PLATFORMS $platform"
        fi
    done

    if [ -z "$MISSING_PLATFORMS" ]; then
        pass "all expected platforms present"
    else
        fail "missing platforms:$MISSING_PLATFORMS"
    fi
}

# Test 8: Verify binaries list
test_binaries_list() {
    test_start "binaries list"
    BINARIES=$(jq -r '.binaries[]' plugin.json 2>/dev/null)

    if echo "$BINARIES" | grep -q "bin/meta-cc-mcp"; then
        pass "binaries list is correct"
    else
        fail "binaries list is incomplete or incorrect"
    fi
}

# Run all tests
main() {
    echo "Running Stage 20.1 Plugin Structure Tests"
    echo "=========================================="
    echo ""

    # Change to repository root
    cd "$(dirname "$0")/.."

    test_plugin_json_exists
    test_plugin_json_valid
    test_plugin_version_semver
    test_required_fields
    test_directory_structure
    test_mcp_config_template
    test_platforms_list
    test_binaries_list

    echo ""
    echo "=========================================="
    echo "Tests run: $TESTS_RUN"
    echo "Tests passed: $TESTS_PASSED"
    echo "Tests failed: $TESTS_FAILED"
    echo ""

    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}All tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}Some tests failed.${NC}"
        exit 1
    fi
}

main "$@"
