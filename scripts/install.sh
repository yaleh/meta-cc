#!/bin/bash
# meta-cc installer
set -e

echo "Installing meta-cc..."

# Detect installation directory
INSTALL_DIR="${HOME}/.local"
mkdir -p "${INSTALL_DIR}/bin"

# Copy binaries
cp bin/meta-cc* "${INSTALL_DIR}/bin/" 2>/dev/null || true
cp bin/meta-cc-mcp* "${INSTALL_DIR}/bin/" 2>/dev/null || true
chmod +x "${INSTALL_DIR}/bin/meta-cc"* 2>/dev/null || true
chmod +x "${INSTALL_DIR}/bin/meta-cc-mcp"* 2>/dev/null || true

# Copy Claude Code integration files
CLAUDE_DIR="${HOME}/.claude/projects/meta-cc"
mkdir -p "${CLAUDE_DIR}/commands" "${CLAUDE_DIR}/agents"
cp -r .claude/commands/* "${CLAUDE_DIR}/commands/"
cp -r .claude/agents/* "${CLAUDE_DIR}/agents/"

echo "✓ Binaries installed to ${INSTALL_DIR}/bin"
echo "✓ Claude Code files installed to ${CLAUDE_DIR}"
echo ""
echo "Next steps:"
echo "1. Add to PATH (if needed): export PATH=\"\${HOME}/.local/bin:\${PATH}\""
echo "2. Configure MCP in Claude Code settings: ~/.claude/settings.json"
echo ""
echo "For MCP setup, see: https://github.com/yaleh/meta-cc#installation"
