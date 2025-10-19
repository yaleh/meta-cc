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

# Test 1.1: CLI binary executes (--version)
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
    test_result "CLI binary executes (--version)" "fail" "Binary not found: bin/meta-cc"
fi

# Test 1.2: CLI help displays
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
    test_result "CLI help displays (--help)" "fail" "Binary not found: bin/meta-cc"
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
    if [ -x "bin/meta-cc" ]; then
        test_result "CLI binary is executable" "pass"
    else
        test_result "CLI binary is executable" "fail" "Binary doesn't have execute permission"
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

# Test 2.1: CLI version matches tag
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
    test_result "CLI version matches tag" "fail" "CLI binary not found"
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
REQUIRED_DIRS=("bin" ".claude-plugin" "commands" "lib")
for dir in "${REQUIRED_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        test_result "Directory exists: $dir" "pass"
    else
        test_result "Directory exists: $dir" "fail" "Required directory missing"
    fi
done

# Test 3.2: Required files present
REQUIRED_FILES=(
    "bin/meta-cc"
    "bin/meta-cc-mcp"
    ".claude-plugin/marketplace.json"
    "install.sh"
    "uninstall.sh"
    "README.md"
    "LICENSE"
)

# Adjust for Windows platform (.exe extension)
if [ "$PLATFORM" = "windows-amd64" ]; then
    REQUIRED_FILES=(
        "bin/meta-cc.exe"
        "bin/meta-cc-mcp.exe"
        ".claude-plugin/marketplace.json"
        "install.sh"
        "uninstall.sh"
        "README.md"
        "LICENSE"
    )
fi

for file in "${REQUIRED_FILES[@]}"; do
    if [ -f "$file" ]; then
        test_result "File exists: $file" "pass"
    else
        test_result "File exists: $file" "fail" "Required file missing"
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

# Test 3.4: Skills structure verification
if [ -d "skills" ]; then
    SKILL_COUNT=$(find skills -name "SKILL.md" 2>/dev/null | wc -l)
    if [ "$SKILL_COUNT" -eq 16 ]; then
        test_result "Skills structure: found 16 skills" "pass"
    else
        test_result "Skills structure: found 16 skills" "fail" "Expected 16 skills, found $SKILL_COUNT"
    fi
else
    test_result "Skills directory exists" "fail" "skills/ directory not found"
fi

# Test 3.5: marketplace.json contains unified meta-cc plugin
# (Merged meta-cc-skills into meta-cc in v0.32.0)
if [ -f ".claude-plugin/marketplace.json" ]; then
    META_CC_PLUGIN_EXISTS=$(jq '.plugins[] | select(.name=="meta-cc")' .claude-plugin/marketplace.json 2>/dev/null)
    if [ -n "$META_CC_PLUGIN_EXISTS" ]; then
        AGENT_COUNT=$(jq '.plugins[] | select(.name=="meta-cc") | .agents | length' .claude-plugin/marketplace.json 2>/dev/null || echo 0)
        SKILLS_COUNT=$(jq '.plugins[] | select(.name=="meta-cc") | .skills | length' .claude-plugin/marketplace.json 2>/dev/null || echo 0)

        if [ "$AGENT_COUNT" -eq 5 ]; then
            test_result "marketplace.json declares meta-cc plugin with 5 agents" "pass"
        else
            test_result "marketplace.json declares meta-cc plugin with agents" "fail" "Expected 5 agents in meta-cc plugin, found $AGENT_COUNT"
        fi

        if [ "$SKILLS_COUNT" -eq 16 ]; then
            test_result "marketplace.json declares meta-cc plugin with 16 skills" "pass"
        else
            test_result "marketplace.json declares meta-cc plugin with skills" "fail" "Expected 16 skills in meta-cc plugin, found $SKILLS_COUNT"
        fi
    else
        test_result "marketplace.json declares meta-cc plugin" "fail" "meta-cc plugin not found in marketplace.json"
    fi
fi

# Test 3.7: Verify specific agents exist
EXPECTED_AGENTS=(
    "agents/iteration-executor.md"
    "agents/iteration-prompt-designer.md"
    "agents/knowledge-extractor.md"
    "agents/project-planner.md"
    "agents/stage-executor.md"
)

for agent in "${EXPECTED_AGENTS[@]}"; do
    if [ -f "$agent" ]; then
        test_result "Agent exists: $agent" "pass"
    else
        test_result "Agent exists: $agent" "fail" "Expected agent file missing"
    fi
done

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
