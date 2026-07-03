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
CLAUDE_DIR="${CLAUDE_DIR:-${HOME}/.claude}"
CODEX_HOME="${CODEX_HOME:-${CODEX_DIR:-${HOME}/.codex}}"

echo "Uninstalling meta-cc..."
echo ""

# Remove binaries
if [ -f "$INSTALL_DIR/meta-cc-mcp" ]; then
    rm -f "$INSTALL_DIR/meta-cc-mcp" 2>/dev/null || true
    info "MCP binary removed from $INSTALL_DIR"
else
    warn "No MCP binary found in $INSTALL_DIR"
fi

if [ -f "$INSTALL_DIR/skill-insights" ]; then
    rm -f "$INSTALL_DIR/skill-insights" 2>/dev/null || true
    info "skill-insights binary removed from $INSTALL_DIR"
else
    warn "No skill-insights binary found in $INSTALL_DIR"
fi

# Remove legacy CLI binary if present (optional - for backwards compatibility)
if [ -f "$INSTALL_DIR/meta-cc" ]; then
    rm -f "$INSTALL_DIR/meta-cc" 2>/dev/null || true
    info "Legacy CLI binary removed from $INSTALL_DIR"
fi

# Remove slash commands (explicit list)
PUBLISHED_COMMANDS="meta prompt-find prompt-list prompt-show"
CMD_REMOVED=0
for cmd in $PUBLISHED_COMMANDS; do
    if [ -f "$CLAUDE_DIR/commands/${cmd}.md" ]; then
        rm -f "$CLAUDE_DIR/commands/${cmd}.md" 2>/dev/null || true
        CMD_REMOVED=$((CMD_REMOVED + 1))
    fi
done
if [ "$CMD_REMOVED" -gt 0 ]; then
    info "Slash commands removed ($CMD_REMOVED files) from $CLAUDE_DIR/commands"
else
    warn "No slash commands found in $CLAUDE_DIR/commands"
fi

# Remove subagents (explicit list — agents are NOT prefixed with meta-)
PUBLISHED_AGENTS="iteration-executor iteration-prompt-designer knowledge-extractor project-planner stage-executor"
AGENT_REMOVED=0
for agent in $PUBLISHED_AGENTS; do
    if [ -f "$CLAUDE_DIR/agents/${agent}.md" ]; then
        rm -f "$CLAUDE_DIR/agents/${agent}.md" 2>/dev/null || true
        AGENT_REMOVED=$((AGENT_REMOVED + 1))
    fi
done
if [ "$AGENT_REMOVED" -gt 0 ]; then
    info "Agents removed ($AGENT_REMOVED files) from $CLAUDE_DIR/agents"
else
    warn "No agents found in $CLAUDE_DIR/agents"
fi

# Remove Codex skills and plugin files
CODEX_SKILLS_REMOVED=0
for skill in prompt-find prompt-list prompt-show; do
    if [ -d "$CODEX_HOME/skills/${skill}" ]; then
        rm -rf "$CODEX_HOME/skills/${skill}" 2>/dev/null || true
        CODEX_SKILLS_REMOVED=$((CODEX_SKILLS_REMOVED + 1))
    fi
done
if [ "$CODEX_SKILLS_REMOVED" -gt 0 ]; then
    info "Codex skills removed ($CODEX_SKILLS_REMOVED directories) from $CODEX_HOME/skills"
else
    warn "No Codex skills found in $CODEX_HOME/skills"
fi

if [ -d "$CODEX_HOME/plugins/meta-cc" ]; then
    rm -rf "$CODEX_HOME/plugins/meta-cc" 2>/dev/null || true
    info "Codex plugin files removed from $CODEX_HOME/plugins/meta-cc"
else
    info "No Codex plugin files found at $CODEX_HOME/plugins/meta-cc"
fi

# Remove MCP server registration from ~/.claude/mcp.json
MCP_CONFIG="${CLAUDE_DIR}/mcp.json"
if [ -f "$MCP_CONFIG" ]; then
    if command -v jq >/dev/null 2>&1; then
        if jq -e '.mcpServers["meta-cc"]' "$MCP_CONFIG" > /dev/null 2>&1; then
            jq 'del(.mcpServers["meta-cc"])' "$MCP_CONFIG" > "$MCP_CONFIG.tmp"
            mv "$MCP_CONFIG.tmp" "$MCP_CONFIG"
            info "MCP server registration removed from $MCP_CONFIG"
        else
            info "MCP server 'meta-cc' not found in $MCP_CONFIG (already removed)"
        fi
    else
        warn "jq not found — cannot auto-remove MCP config"
        warn "Manually remove 'meta-cc' from .mcpServers in $MCP_CONFIG"
    fi
else
    info "No MCP config at $MCP_CONFIG (nothing to remove)"
fi

# Remove plugin cache entry (created by Claude Code plugin system)
PLUGIN_CACHE_DIR="${CLAUDE_DIR}/plugins/cache/meta-cc-marketplace/meta-cc"
if [ -d "$PLUGIN_CACHE_DIR" ]; then
    rm -rf "$PLUGIN_CACHE_DIR" 2>/dev/null || true
    info "Plugin cache removed from $PLUGIN_CACHE_DIR"
else
    info "No plugin cache found at $PLUGIN_CACHE_DIR (already removed or not present)"
fi

echo ""
echo "Uninstallation complete!"
echo ""
echo "Note: If installed via 'make install-local', also run 'make uninstall-local'"
echo ""
