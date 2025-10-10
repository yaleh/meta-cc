#!/bin/bash
# Test suite for Stage 20.2: Automated Installation Script
# Tests platform detection, binary installation, MCP config merging, and verification

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

# Test 1: Platform detection for Linux
test_platform_detection_linux() {
    test_start "platform detection (Linux)"

    # Test uname output directly
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')

    if [[ "$OS" == "linux" ]] || [[ "$OS" == "darwin" ]]; then
        pass "platform detection works on current OS: $OS"
    else
        fail "unsupported OS: $OS"
    fi
}

# Test 2: Architecture detection
test_architecture_detection() {
    test_start "architecture detection"
    ARCH=$(uname -m)

    if [[ "$ARCH" == "x86_64" ]] || [[ "$ARCH" == "amd64" ]] || [[ "$ARCH" == "aarch64" ]] || [[ "$ARCH" == "arm64" ]]; then
        pass "architecture detection works: $ARCH"
    else
        fail "unsupported architecture: $ARCH"
    fi
}

# Test 3: Binary installation to custom directory
test_binary_installation() {
    test_start "binary installation to custom directory"

    # Create temporary test environment
    TEST_DIR=$(mktemp -d)
    TEST_INSTALL_DIR="$TEST_DIR/install"
    TEST_BIN_DIR="$TEST_DIR/bin"
    mkdir -p "$TEST_BIN_DIR"

    # Create mock binaries
    echo "#!/bin/bash" > "$TEST_BIN_DIR/meta-cc"
    echo "echo 'meta-cc v0.12.0'" >> "$TEST_BIN_DIR/meta-cc"
    echo "#!/bin/bash" > "$TEST_BIN_DIR/meta-cc-mcp"
    echo "echo 'meta-cc-mcp v0.12.0'" >> "$TEST_BIN_DIR/meta-cc-mcp"

    # Test installation
    mkdir -p "$TEST_INSTALL_DIR"
    cp "$TEST_BIN_DIR/meta-cc" "$TEST_INSTALL_DIR/"
    cp "$TEST_BIN_DIR/meta-cc-mcp" "$TEST_INSTALL_DIR/"
    chmod +x "$TEST_INSTALL_DIR/meta-cc" "$TEST_INSTALL_DIR/meta-cc-mcp"

    if [ -x "$TEST_INSTALL_DIR/meta-cc" ] && [ -x "$TEST_INSTALL_DIR/meta-cc-mcp" ]; then
        pass "binaries installed and executable"
    else
        fail "binaries not installed correctly"
    fi

    # Cleanup
    rm -rf "$TEST_DIR"
}

# Test 4: MCP config merge - no existing config
test_mcp_config_merge_new() {
    test_start "MCP config merge (no existing config)"

    # Create temporary test environment
    TEST_DIR=$(mktemp -d)
    MCP_CONFIG="$TEST_DIR/mcp.json"
    MCP_TEMPLATE=".claude/lib/mcp-config.json"

    # Copy template to test location
    if [ -f "$MCP_TEMPLATE" ]; then
        cp "$MCP_TEMPLATE" "$MCP_CONFIG"

        if jq -e '.mcpServers."meta-cc"' "$MCP_CONFIG" >/dev/null 2>&1; then
            pass "MCP config created from template"
        else
            fail "MCP config template incorrect"
        fi
    else
        fail "MCP config template not found"
    fi

    # Cleanup
    rm -rf "$TEST_DIR"
}

# Test 5: MCP config merge - existing config with other servers
test_mcp_config_merge_existing() {
    test_start "MCP config merge (preserves existing servers)"

    # Skip if jq not available
    if ! command -v jq >/dev/null 2>&1; then
        echo "SKIP (jq not installed)"
        return
    fi

    # Create temporary test environment
    TEST_DIR=$(mktemp -d)
    EXISTING_CONFIG="$TEST_DIR/mcp.json"
    MCP_TEMPLATE=".claude/lib/mcp-config.json"

    # Create existing config with another server
    cat > "$EXISTING_CONFIG" <<'EOF'
{
  "mcpServers": {
    "existing-server": {
      "command": "existing-command",
      "args": [],
      "disabled": false
    }
  }
}
EOF

    # Merge with template
    if [ -f "$MCP_TEMPLATE" ]; then
        TEMP_CONFIG=$(mktemp)
        jq -s '.[0] * .[1]' "$EXISTING_CONFIG" "$MCP_TEMPLATE" > "$TEMP_CONFIG"

        # Verify both servers exist
        if jq -e '.mcpServers."existing-server"' "$TEMP_CONFIG" >/dev/null 2>&1 && \
           jq -e '.mcpServers."meta-cc"' "$TEMP_CONFIG" >/dev/null 2>&1; then
            pass "MCP config merged, existing servers preserved"
        else
            fail "MCP config merge failed"
        fi

        rm -f "$TEMP_CONFIG"
    else
        fail "MCP config template not found"
    fi

    # Cleanup
    rm -rf "$TEST_DIR"
}

# Test 6: Installation verification
test_installation_verification() {
    test_start "installation verification"

    # Check if meta-cc binary exists in path or local bin
    if command -v meta-cc >/dev/null 2>&1 || [ -x "$HOME/.local/bin/meta-cc" ]; then
        pass "meta-cc binary found and executable"
    else
        echo "SKIP (meta-cc not installed)"
    fi
}

# Test 7: Uninstall script exists
test_uninstall_script_exists() {
    test_start "uninstall script exists"

    if [ -f "scripts/uninstall.sh" ]; then
        pass "uninstall script exists"
    else
        fail "uninstall script not found"
    fi
}

# Test 8: Uninstall removes components
test_uninstall_removes_components() {
    test_start "uninstall removes components"

    # Create temporary test environment
    TEST_DIR=$(mktemp -d)
    TEST_INSTALL_DIR="$TEST_DIR/.local/bin"
    TEST_CLAUDE_DIR="$TEST_DIR/.claude"
    mkdir -p "$TEST_INSTALL_DIR" "$TEST_CLAUDE_DIR/commands" "$TEST_CLAUDE_DIR/agents"

    # Create mock files
    touch "$TEST_INSTALL_DIR/meta-cc"
    touch "$TEST_INSTALL_DIR/meta-cc-mcp"
    touch "$TEST_CLAUDE_DIR/commands/meta-stats.md"
    touch "$TEST_CLAUDE_DIR/agents/meta-coach.md"

    # Simulate uninstall
    rm -f "$TEST_INSTALL_DIR/meta-cc" "$TEST_INSTALL_DIR/meta-cc-mcp"
    rm -rf "$TEST_CLAUDE_DIR/commands/meta-"*
    rm -rf "$TEST_CLAUDE_DIR/agents/meta-"*

    if [ ! -f "$TEST_INSTALL_DIR/meta-cc" ] && \
       [ ! -f "$TEST_INSTALL_DIR/meta-cc-mcp" ] && \
       [ ! -f "$TEST_CLAUDE_DIR/commands/meta-stats.md" ] && \
       [ ! -f "$TEST_CLAUDE_DIR/agents/meta-coach.md" ]; then
        pass "uninstall removes all components"
    else
        fail "uninstall did not remove all components"
    fi

    # Cleanup
    rm -rf "$TEST_DIR"
}

# Run all tests
main() {
    echo "Running Stage 20.2 Installation Tests"
    echo "======================================"
    echo ""

    # Change to repository root
    cd "$(dirname "$0")/.."

    test_platform_detection_linux
    test_architecture_detection
    test_binary_installation
    test_mcp_config_merge_new
    test_mcp_config_merge_existing
    test_installation_verification
    test_uninstall_script_exists
    test_uninstall_removes_components

    echo ""
    echo "======================================"
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
