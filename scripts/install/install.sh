#!/bin/bash
# meta-cc installer (enhanced)
# Supports platform detection, MCP configuration merging, and verification

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Error handling
trap 'error_exit "Installation failed at line $LINENO"' ERR

error_exit() {
    echo -e "${RED}ERROR: $1${NC}" >&2
    exit 1
}

info() {
    echo -e "${GREEN}✓${NC} $1"
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Platform and architecture detection
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux*)   PLATFORM="linux" ;;
        darwin*)  PLATFORM="darwin" ;;
        mingw*|msys*|cygwin*) PLATFORM="windows" ;;
        *) error_exit "Unsupported OS: $OS" ;;
    esac

    case "$ARCH" in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) error_exit "Unsupported architecture: $ARCH" ;;
    esac

    PLATFORM_ARCH="${PLATFORM}-${ARCH}"
}

# Install binaries
install_binaries() {
    INSTALL_DIR="${INSTALL_DIR:-${HOME}/.local/bin}"
    mkdir -p "$INSTALL_DIR"

    # Select correct binary for platform
    if [ "$PLATFORM" = "windows" ]; then
        BINARY_EXT=".exe"
    else
        BINARY_EXT=""
    fi

    # Check if binary exists in bin/ directory
    if [ ! -f "bin/meta-cc-mcp${BINARY_EXT}" ]; then
        error_exit "meta-cc-mcp binary not found in bin/"
    fi

    # Copy binary
    cp "bin/meta-cc-mcp${BINARY_EXT}" "$INSTALL_DIR/" || error_exit "Failed to copy meta-cc-mcp binary"

    # Set executable permissions (not needed on Windows)
    if [ "$PLATFORM" != "windows" ]; then
        chmod +x "$INSTALL_DIR/meta-cc-mcp"
    fi

    info "Binary installed to $INSTALL_DIR"
}

# Install Claude plugin manifest to data directory
install_plugin_manifest() {
    PLUGIN_DATA_DIR="${PLUGIN_DATA_DIR:-${HOME}/.local/share/meta-cc}"
    PLUGIN_MANIFEST_DIR="${PLUGIN_DATA_DIR}/.claude-plugin"

    if [ ! -d ".claude-plugin" ]; then
        warn ".claude-plugin directory not found, skipping plugin manifest installation"
        return
    fi

    mkdir -p "$PLUGIN_MANIFEST_DIR"
    cp .claude-plugin/plugin.json "$PLUGIN_MANIFEST_DIR/plugin.json" || error_exit "Failed to copy plugin.json"
    if [ -f ".claude-plugin/marketplace.json" ]; then
        cp .claude-plugin/marketplace.json "$PLUGIN_MANIFEST_DIR/marketplace.json" || error_exit "Failed to copy marketplace.json"
    else
        warn "marketplace.json not found, skipping marketplace metadata"
    fi

    info "Plugin manifest installed to $PLUGIN_MANIFEST_DIR"
}

# Install Claude Code integration files
install_claude_files() {
    CLAUDE_DIR="${CLAUDE_DIR:-${HOME}/.claude}"
    mkdir -p "$CLAUDE_DIR/commands" "$CLAUDE_DIR/agents" "$CLAUDE_DIR/skills"

    # Check if commands directory exists (required)
    if [ ! -d "commands" ]; then
        error_exit "commands directory not found"
    fi

    # Copy slash commands
    cp commands/* "$CLAUDE_DIR/commands/" 2>/dev/null || error_exit "Failed to copy slash commands"

    # Copy subagents if directory exists (removed in 3.0.0+)
    if [ -d "agents" ] && [ "$(ls -A agents 2>/dev/null)" ]; then
        cp agents/* "$CLAUDE_DIR/agents/" 2>/dev/null || warn "No agents to copy"
    fi

    # Copy skills if directory exists
    if [ -d "skills" ]; then
        cp -r skills/* "$CLAUDE_DIR/skills/" 2>/dev/null || warn "No skills to copy"
        SKILL_COUNT=$(find "$CLAUDE_DIR/skills" -name "SKILL.md" 2>/dev/null | wc -l)
        if [ "$SKILL_COUNT" -gt 0 ]; then
            info "Installed $SKILL_COUNT skills"
        fi
    fi

    info "Claude Code files installed to $CLAUDE_DIR"
}

install_codex_files() {
    CODEX_HOME="${CODEX_HOME:-${CODEX_DIR:-${HOME}/.codex}}"
    mkdir -p "$CODEX_HOME/skills" "$CODEX_HOME/plugins/meta-cc"

    if [ -d "skills" ]; then
        cp -r skills/* "$CODEX_HOME/skills/" 2>/dev/null || warn "No Codex skills to copy"
        SKILL_COUNT=$(find "$CODEX_HOME/skills" -name "SKILL.md" 2>/dev/null | wc -l)
        if [ "$SKILL_COUNT" -gt 0 ]; then
            info "Installed $SKILL_COUNT Codex skills"
        fi
    else
        warn "skills directory not found, skipping Codex skills"
    fi

    if [ -d ".codex-plugin" ]; then
        cp -r .codex-plugin "$CODEX_HOME/plugins/meta-cc/"
        info "Codex plugin manifest installed to $CODEX_HOME/plugins/meta-cc/.codex-plugin"
    else
        warn ".codex-plugin directory not found, skipping Codex plugin manifest"
    fi

    if [ -f ".codex-mcp.json" ]; then
        cp ".codex-mcp.json" "$CODEX_HOME/plugins/meta-cc/.codex-mcp.json"
        info "Codex MCP template installed to $CODEX_HOME/plugins/meta-cc/.codex-mcp.json"
    fi

    info "Codex files installed to $CODEX_HOME"
}

# Merge MCP configuration
merge_mcp_config() {
    CLAUDE_DIR="${CLAUDE_DIR:-${HOME}/.claude}"
    MCP_CONFIG="${CLAUDE_DIR}/mcp.json"
    mkdir -p "$CLAUDE_DIR"

    # For manual installs, prefer lib/mcp-config.json (PATH-based, suitable for manual install).
    # .mcp.json uses ${CLAUDE_PLUGIN_ROOT}/bin/meta-cc-mcp which only works in plugin-managed
    # environments. Fall back to .mcp.json only if lib/mcp-config.json is not present.
    if [ -f "lib/mcp-config.json" ]; then
        MCP_TEMPLATE="lib/mcp-config.json"
    elif [ -f ".mcp.json" ]; then
        MCP_TEMPLATE=".mcp.json"
    else
        warn "No MCP config template found (lib/mcp-config.json or .mcp.json), skipping MCP configuration"
        return
    fi

    if [ ! -f "$MCP_CONFIG" ]; then
        # No existing config, copy template
        cp "$MCP_TEMPLATE" "$MCP_CONFIG"
        info "MCP configuration created at $MCP_CONFIG"
    else
        # Merge with existing config using jq
        if command -v jq >/dev/null 2>&1; then
            TEMP_CONFIG=$(mktemp)
            jq -s '.[0] * .[1]' "$MCP_CONFIG" "$MCP_TEMPLATE" > "$TEMP_CONFIG"
            mv "$TEMP_CONFIG" "$MCP_CONFIG"
            info "MCP configuration merged (existing servers preserved)"
        else
            warn "jq not found, skipping MCP config merge"
            echo ""
            echo "Please manually add meta-cc to $MCP_CONFIG:"
            echo ""
            cat "$MCP_TEMPLATE"
            echo ""
        fi
    fi
}

# Verify installation
verify_installation() {
    INSTALL_DIR="${INSTALL_DIR:-${HOME}/.local/bin}"

    # Check binary exists
    if [ ! -f "$INSTALL_DIR/meta-cc-mcp" ]; then
        error_exit "meta-cc-mcp binary not found at $INSTALL_DIR/meta-cc-mcp"
    fi

    # Check binary is executable
    if [ ! -x "$INSTALL_DIR/meta-cc-mcp" ]; then
        error_exit "meta-cc-mcp binary is not executable"
    fi

    # Skip version test (binary doesn't support --version yet)
    # This prevents installation from hanging on systems where --version is tested
    # TODO: Re-enable when --version flag is added to meta-cc-mcp
    info "Binary installed successfully"
}

# Main installation flow
main() {
    echo "Installing meta-cc..."
    echo ""

    detect_platform
    info "Detected platform: $PLATFORM_ARCH"

    install_binaries
    install_plugin_manifest
    install_claude_files
    install_codex_files
    merge_mcp_config
    verify_installation

    echo ""
    echo "Installation complete! 🎉"
    echo ""
    echo "Next steps:"
    echo "1. Add to PATH (if needed): export PATH=\"\$HOME/.local/bin:\$PATH\""
    echo "2. Restart Claude Code and/or Codex to load the integrations"
    echo "3. If needed, configure MCP manually with the bundled .mcp.json or .codex-mcp.json"
    echo ""
}

main "$@"
