#!/bin/bash
# meta-cc uninstaller
# Removes meta-cc binaries and Claude Code integration files

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() {
    echo -e "${GREEN}✓${NC} $1"
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

error_exit() {
    echo -e "${RED}ERROR: $1${NC}" >&2
    exit 1
}

# Detect installation directory
INSTALL_DIR="${INSTALL_DIR:-${HOME}/.local/bin}"
CLAUDE_DIR="${HOME}/.claude"

echo "Uninstalling meta-cc..."
echo ""

# Remove binaries
if [ -f "$INSTALL_DIR/meta-cc" ] || [ -f "$INSTALL_DIR/meta-cc-mcp" ]; then
    rm -f "$INSTALL_DIR/meta-cc" "$INSTALL_DIR/meta-cc-mcp" 2>/dev/null || true
    info "Binaries removed from $INSTALL_DIR"
else
    warn "No binaries found in $INSTALL_DIR"
fi

# Remove slash commands
if ls "$CLAUDE_DIR/commands/meta-"* >/dev/null 2>&1; then
    rm -f "$CLAUDE_DIR/commands/meta-"* 2>/dev/null || true
    info "Slash commands removed from $CLAUDE_DIR/commands"
else
    warn "No slash commands found in $CLAUDE_DIR/commands"
fi

# Remove subagents
if ls "$CLAUDE_DIR/agents/meta-"* >/dev/null 2>&1; then
    rm -f "$CLAUDE_DIR/agents/meta-"* 2>/dev/null || true
    info "Subagents removed from $CLAUDE_DIR/agents"
else
    warn "No subagents found in $CLAUDE_DIR/agents"
fi

echo ""
echo "Uninstallation complete!"
echo ""
echo "Note: MCP configuration at ~/.claude/mcp.json was preserved."
echo "To remove the meta-cc MCP server, manually edit ~/.claude/mcp.json"
echo "and remove the 'meta-cc' entry from mcpServers."
echo ""
