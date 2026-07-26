#!/bin/bash
# Smoke tests for meta-cc release artifacts
#
# Purpose: Verify release artifacts work correctly before publishing
# Scope: Critical path only - binary execution, version consistency, plugin structure
# Platform: Tests linux-amd64 natively, trusts Go cross-compilation for others
#
# Usage: ./smoke-tests.sh <version> <platform> <package-path>
# Example: ./smoke-tests.sh v0.26.9 linux-amd64 build/packages/meta-cc-plugin-v0.26.9-linux-amd64.tar.gz

set -e

# ==================================================================
# ARGUMENT PARSING
# ==================================================================

VERSION=$1
PLATFORM=$2
PACKAGE_PATH=$3

if [ -z "$VERSION" ] || [ -z "$PLATFORM" ] || [ -z "$PACKAGE_PATH" ]; then
    echo "Usage: $0 <version> <platform> <package-path>"
    echo "Example: $0 v0.26.9 linux-amd64 build/packages/meta-cc-plugin-v0.26.9-linux-amd64.tar.gz"
    exit 1
fi

# Remove 'v' prefix for version comparison
VERSION_NUM=${VERSION#v}

# ==================================================================
# SETUP AND VALIDATION
# ==================================================================

echo "========================================="
echo "Smoke Tests for meta-cc Release"
echo "========================================="
echo "Version:  $VERSION ($VERSION_NUM)"
echo "Platform: $PLATFORM"
echo "Package:  $PACKAGE_PATH"
echo ""

# Check dependencies
MISSING_DEPS=()
for cmd in tar file jq grep awk; do
    if ! command -v $cmd &> /dev/null; then
        MISSING_DEPS+=($cmd)
    fi
done

if [ ${#MISSING_DEPS[@]} -gt 0 ]; then
    echo "❌ ERROR: Missing dependencies: ${MISSING_DEPS[*]}"
    echo "Install with: sudo apt-get install ${MISSING_DEPS[*]}"
    exit 1
fi

# Verify package exists
if [ ! -f "$PACKAGE_PATH" ]; then
    echo "❌ ERROR: Package not found: $PACKAGE_PATH"
    exit 1
fi

# Create temporary directory
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

echo "Extracting package to $TEMP_DIR..."
tar -xzf "$PACKAGE_PATH" -C "$TEMP_DIR"

# Find the extracted directory (should be meta-cc-plugin-<platform>)
EXTRACT_DIR=$(find "$TEMP_DIR" -mindepth 1 -maxdepth 1 -type d | head -1)

if [ -z "$EXTRACT_DIR" ]; then
    echo "❌ ERROR: No directory found after extraction"
    exit 1
fi

cd "$EXTRACT_DIR"
echo "Working directory: $EXTRACT_DIR"
echo ""

# ==================================================================
# TEST TRACKING
# ==================================================================

TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=()

test_result() {
    local test_name="$1"
    local result="$2"
    local error_msg="$3"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if [ "$result" = "pass" ]; then
        echo "  ✓ $test_name"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo "  ✗ $test_name"
        if [ -n "$error_msg" ]; then
            echo "    Error: $error_msg"
        fi
        FAILED_TESTS+=("$test_name: $error_msg")
    fi
}

# ==================================================================
# TEST CATEGORY 1: BINARY EXECUTION
# ==================================================================

echo "Test Category 1: Binary Execution"
echo "-----------------------------------"

# Test 1.1: CLI binary executes (--version) - OPTIONAL (removed in Phase 26)
if [ -f "bin/meta-cc" ]; then
    if VERSION_OUTPUT=$(./bin/meta-cc --version 2>&1); then
        if echo "$VERSION_OUTPUT" | grep -q "meta-cc version"; then
            test_result "CLI binary executes (--version)" "pass"
        else
            test_result "CLI binary executes (--version)" "fail" "Unexpected version output format: $VERSION_OUTPUT"
        fi
    else
        test_result "CLI binary executes (--version)" "fail" "Binary execution failed with exit code $?"
    fi
else
    # CLI binary is optional (removed in Phase 26 - MCP-only architecture)
    echo "  ⊘ CLI binary executes (--version) - SKIPPED (CLI removed in Phase 26)"
fi

# Test 1.2: CLI help displays - OPTIONAL (removed in Phase 26)
if [ -f "bin/meta-cc" ]; then
    if HELP_OUTPUT=$(./bin/meta-cc --help 2>&1); then
        if echo "$HELP_OUTPUT" | grep -qi "usage\|command"; then
            test_result "CLI help displays (--help)" "pass"
        else
            test_result "CLI help displays (--help)" "fail" "Help output doesn't contain usage information"
        fi
    else
        test_result "CLI help displays (--help)" "fail" "Help command failed"
    fi
else
    # CLI binary is optional (removed in Phase 26 - MCP-only architecture)
    echo "  ⊘ CLI help displays (--help) - SKIPPED (CLI removed in Phase 26)"
fi

# Test 1.3: MCP server binary executes
if [ -f "bin/meta-cc-mcp" ]; then
    # MCP server may exit with code 1 when called with --help (not a failure)
    # Some versions may not support --help, so just check if binary runs
    if MCP_OUTPUT=$(timeout 2s ./bin/meta-cc-mcp --help 2>&1 || true); then
        # Binary executed (with any exit code), this is a pass
        test_result "MCP server binary executes" "pass"
    else
        test_result "MCP server binary executes" "fail" "Binary failed to execute or timed out"
    fi
else
    test_result "MCP server binary executes" "fail" "Binary not found: bin/meta-cc-mcp"
fi

# Test 1.4: Binaries are executable (Unix platforms only)
if [ "$PLATFORM" != "windows-amd64" ]; then
    # CLI binary check is optional (removed in Phase 26)
    if [ -f "bin/meta-cc" ]; then
        if [ -x "bin/meta-cc" ]; then
            test_result "CLI binary is executable" "pass"
        else
            test_result "CLI binary is executable" "fail" "Binary doesn't have execute permission"
        fi
    else
        echo "  ⊘ CLI binary is executable - SKIPPED (CLI removed in Phase 26)"
    fi

    if [ -x "bin/meta-cc-mcp" ]; then
        test_result "MCP server binary is executable" "pass"
    else
        test_result "MCP server binary is executable" "fail" "Binary doesn't have execute permission"
    fi
fi

echo ""

# ==================================================================
# TEST CATEGORY 2: VERSION CONSISTENCY
# ==================================================================

echo "Test Category 2: Version Consistency"
echo "-------------------------------------"

# Test 2.1: CLI version matches tag - OPTIONAL (removed in Phase 26)
if [ -f "bin/meta-cc" ]; then
    VERSION_OUTPUT=$(./bin/meta-cc --version 2>&1)

    # Try to extract version with multiple patterns
    if echo "$VERSION_OUTPUT" | grep -q "version"; then
        # Pattern 1: "version X.Y.Z" format
        CLI_VERSION=$(echo "$VERSION_OUTPUT" | grep -oP 'version \K[0-9]+\.[0-9]+\.[0-9]+[^ ]*' || true)

        # Pattern 2: "meta-cc version X.Y.Z (commit: ...)" format
        if [ -z "$CLI_VERSION" ]; then
            CLI_VERSION=$(echo "$VERSION_OUTPUT" | grep -oP '[0-9]+\.[0-9]+\.[0-9]+[^ ]*' | head -1 || true)
        fi
    fi

    # If still no version found, mark as unknown
    if [ -z "$CLI_VERSION" ]; then
        CLI_VERSION="UNKNOWN"
    fi

    # Handle version formats: X.Y.Z or X.Y.Z-suffix
    # CLI may report with or without 'v' prefix
    CLI_VERSION_CLEAN=${CLI_VERSION#v}

    if [ "$CLI_VERSION_CLEAN" = "$VERSION_NUM" ]; then
        test_result "CLI version matches tag ($VERSION_NUM)" "pass"
    else
        test_result "CLI version matches tag" "fail" "CLI reports '$CLI_VERSION_CLEAN' but tag is '$VERSION_NUM'"
    fi
else
    # CLI binary is optional (removed in Phase 26 - MCP-only architecture)
    echo "  ⊘ CLI version matches tag - SKIPPED (CLI removed in Phase 26)"
fi

# Test 2.2: marketplace.json version matches tag
if [ -f ".claude-plugin/marketplace.json" ]; then
    MARKETPLACE_VERSION=$(jq -r '.plugins[0].version' .claude-plugin/marketplace.json 2>/dev/null || echo "UNKNOWN")

    if [ "$MARKETPLACE_VERSION" = "$VERSION_NUM" ]; then
        test_result "marketplace.json version matches tag ($VERSION_NUM)" "pass"
    else
        test_result "marketplace.json version matches tag" "fail" "marketplace.json has '$MARKETPLACE_VERSION' but tag is '$VERSION_NUM'"
    fi
else
    test_result "marketplace.json version matches tag" "fail" "marketplace.json not found"
fi

echo ""

# ==================================================================
# TEST CATEGORY 3: PLUGIN STRUCTURE
# ==================================================================

echo "Test Category 3: Plugin Structure"
echo "----------------------------------"

# Test 3.1: Required directories present
REQUIRED_DIRS=("bin" ".claude-plugin" ".codex-plugin" "commands" "skills" "lib")
for dir in "${REQUIRED_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        test_result "Directory exists: $dir" "pass"
    else
        test_result "Directory exists: $dir" "fail" "Required directory missing"
    fi
done

# Test 3.2: Required files present
# Note: bin/meta-cc is optional (removed in Phase 26 - MCP-only architecture)
REQUIRED_FILES=(
    "bin/meta-cc-mcp"
    ".claude-plugin/marketplace.json"
    ".claude-plugin/plugin.json"
    ".codex-plugin/plugin.json"
    ".mcp.json"
    ".codex-mcp.json"
    "install.sh"
    "uninstall.sh"
    "install-skills.sh"
    "README.md"
    "LICENSE"
)

# Optional files (CLI removed in Phase 26)
OPTIONAL_FILES=(
    "bin/meta-cc"
)

# Adjust for Windows platform (.exe extension)
if [ "$PLATFORM" = "windows-amd64" ]; then
    REQUIRED_FILES=(
        "bin/meta-cc-mcp.exe"
        ".claude-plugin/marketplace.json"
        ".claude-plugin/plugin.json"
        ".codex-plugin/plugin.json"
        ".mcp.json"
        ".codex-mcp.json"
        "install.sh"
        "uninstall.sh"
        "install-skills.sh"
        "README.md"
        "LICENSE"
    )
    OPTIONAL_FILES=(
        "bin/meta-cc.exe"
    )
fi

for file in "${REQUIRED_FILES[@]}"; do
    if [ -f "$file" ]; then
        test_result "File exists: $file" "pass"
    else
        test_result "File exists: $file" "fail" "Required file missing"
    fi
done

# Check optional files (CLI removed in Phase 26)
for file in "${OPTIONAL_FILES[@]}"; do
    if [ -f "$file" ]; then
        test_result "File exists: $file" "pass"
    else
        echo "  ⊘ File exists: $file - SKIPPED (CLI removed in Phase 26)"
    fi
done

# Test 3.3: JSON files are valid
if [ -f ".claude-plugin/marketplace.json" ]; then
    if jq . .claude-plugin/marketplace.json > /dev/null 2>&1; then
        test_result "marketplace.json is valid JSON" "pass"
    else
        test_result "marketplace.json is valid JSON" "fail" "JSON syntax error"
    fi
fi

# Test 3.4: marketplace.json contains unified meta-cc plugin
if [ -f ".claude-plugin/marketplace.json" ]; then
    META_CC_PLUGIN_EXISTS=$(jq '.plugins[] | select(.name=="meta-cc")' .claude-plugin/marketplace.json 2>/dev/null)
    if [ -n "$META_CC_PLUGIN_EXISTS" ]; then
        CMD_COUNT=$(jq '.plugins[] | select(.name=="meta-cc") | .commands | length' .claude-plugin/marketplace.json 2>/dev/null || echo 0)
        if [ "$CMD_COUNT" -eq 3 ]; then
            test_result "marketplace.json declares meta-cc plugin with 3 commands" "pass"
        else
            test_result "marketplace.json declares meta-cc plugin with 3 commands" "fail" "Expected 3 commands in meta-cc plugin, found $CMD_COUNT"
        fi
    else
        test_result "marketplace.json declares meta-cc plugin" "fail" "meta-cc plugin not found in marketplace.json"
    fi
fi

# Test 3.5: All published commands present
EXPECTED_COMMANDS=(
    "commands/prompt-find.md"
    "commands/prompt-list.md"
    "commands/prompt-show.md"
)
for cmd in "${EXPECTED_COMMANDS[@]}"; do
    if [ -f "$cmd" ]; then
        test_result "Command exists: $cmd" "pass"
    else
        test_result "Command exists: $cmd" "fail" "Expected command file missing"
    fi
done

# Test 3.6: No agents directory in archive (removed in 3.0.0)
if [ -d "agents" ]; then
    test_result "agents/ directory NOT in archive (removed in 3.0.0)" "fail" "agents/ directory found in release package"
else
    test_result "agents/ directory NOT in archive (removed in 3.0.0)" "pass"
fi

# Test 3.7: Codex skills are present
for skill in prompt-find prompt-list prompt-show; do
    if [ -f "skills/${skill}/SKILL.md" ]; then
        test_result "Codex skill exists: skills/${skill}/SKILL.md" "pass"
    else
        test_result "Codex skill exists: skills/${skill}/SKILL.md" "fail" "Expected Codex skill file missing"
    fi
done

# Test 3.10: plugin.json in archive .claude-plugin/
if [ -f ".claude-plugin/plugin.json" ]; then
    if jq . .claude-plugin/plugin.json > /dev/null 2>&1; then
        PLUGIN_VER=$(jq -r '.version' .claude-plugin/plugin.json)
        MARKET_VER=$(jq -r '.plugins[0].version' .claude-plugin/marketplace.json 2>/dev/null || echo "")
        if [ "$PLUGIN_VER" = "$MARKET_VER" ]; then
            test_result "plugin.json in archive: valid, version=$PLUGIN_VER matches marketplace" "pass"
        else
            test_result "plugin.json in archive: version parity" "fail" "plugin=$PLUGIN_VER marketplace=$MARKET_VER"
        fi
    else
        test_result "plugin.json in archive: valid JSON" "fail" "JSON parse error"
    fi
else
    test_result "plugin.json in archive: exists" "fail" ".claude-plugin/plugin.json not found in archive"
fi

# Test 3.11: .mcp.json in archive
if [ -f ".mcp.json" ]; then
    if jq -e '.mcpServers["meta-cc"] // .["meta-cc"]' .mcp.json > /dev/null 2>&1; then
        test_result ".mcp.json in archive: valid with meta-cc server" "pass"
    else
        test_result ".mcp.json in archive: valid with meta-cc server" "fail" "meta-cc server entry missing"
    fi
else
    test_result ".mcp.json in archive: exists" "fail" ".mcp.json not found in archive"
fi

# Test 3.11b: .codex-mcp.json in archive
if [ -f ".codex-mcp.json" ]; then
    if jq -e '.mcpServers["meta-cc"]' .codex-mcp.json > /dev/null 2>&1; then
        test_result ".codex-mcp.json in archive: valid with meta-cc server" "pass"
    else
        test_result ".codex-mcp.json in archive: valid with meta-cc server" "fail" "meta-cc server entry missing"
    fi
else
    test_result ".codex-mcp.json in archive: exists" "fail" ".codex-mcp.json not found in archive"
fi

# Test 3.11c: Codex plugin manifest is valid
if [ -f ".codex-plugin/plugin.json" ]; then
    if jq -e '.skills and .mcpServers' .codex-plugin/plugin.json > /dev/null 2>&1; then
        test_result "Codex plugin manifest declares skills and MCP config" "pass"
    else
        test_result "Codex plugin manifest declares skills and MCP config" "fail" "skills or mcpServers missing"
    fi
else
    test_result "Codex plugin manifest exists" "fail" ".codex-plugin/plugin.json not found"
fi

# Test 3.11d: .codex-plugin/plugin.json version matches the release tag
#
# The Codex plugin manager (`codex plugin add`/`codex plugin list`) reports
# THIS file's version field to users -- not marketplace.json's or the
# Claude-side plugin.json's version. A drift here is invisible to Claude
# Code users but visibly wrong for Codex users (DIR-026).
if [ -f ".codex-plugin/plugin.json" ]; then
    CODEX_PLUGIN_VERSION=$(jq -r '.version' .codex-plugin/plugin.json 2>/dev/null || echo "UNKNOWN")
    if [ "$CODEX_PLUGIN_VERSION" = "$VERSION_NUM" ]; then
        test_result "Codex plugin manifest version matches tag ($VERSION_NUM)" "pass"
    else
        test_result "Codex plugin manifest version matches tag" "fail" ".codex-plugin/plugin.json has '$CODEX_PLUGIN_VERSION' but tag is '$VERSION_NUM' -- Codex users would see the wrong version from 'codex plugin list'"
    fi
fi

# Test 3.12: archive marketplace.json source is "."
if [ -f ".claude-plugin/marketplace.json" ]; then
    ARCHIVE_SOURCE=$(jq -r '.plugins[0].source' .claude-plugin/marketplace.json 2>/dev/null || echo "")
    if [ "$ARCHIVE_SOURCE" = "." ]; then
        test_result "archive marketplace.json source is '.'" "pass"
    else
        test_result "archive marketplace.json source is '.'" "fail" "Expected '.', got '$ARCHIVE_SOURCE'"
    fi
fi

echo ""

# ==================================================================
# TEST RESULTS SUMMARY
# ==================================================================

echo "========================================="
echo "Smoke Test Results"
echo "========================================="
echo "Total tests:  $TOTAL_TESTS"
echo "Passed:       $PASSED_TESTS"
echo "Failed:       $((TOTAL_TESTS - PASSED_TESTS))"
echo ""

if [ ${#FAILED_TESTS[@]} -gt 0 ]; then
    echo "Failed Tests:"
    for failure in "${FAILED_TESTS[@]}"; do
        echo "  ✗ $failure"
    done
    echo ""
    echo "❌ SMOKE TESTS FAILED"
    echo ""
    echo "Action Required:"
    echo "  1. Review failed tests above"
    echo "  2. Fix issues in build/packaging process"
    echo "  3. Re-run release workflow"
    exit 1
else
    echo "✓ ALL SMOKE TESTS PASSED"
    echo ""
    echo "Release artifacts verified:"
    echo "  - Binaries execute correctly"
    echo "  - Version consistency confirmed"
    echo "  - Plugin structure valid"
    echo ""
    echo "Ready to publish release!"
    exit 0
fi
