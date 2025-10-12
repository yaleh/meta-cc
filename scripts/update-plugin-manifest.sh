#!/bin/bash
# Auto-update plugin.json with current command and agent files

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
PLUGIN_JSON="$PROJECT_ROOT/.claude-plugin/plugin.json"

echo "Updating plugin manifest..."

# Collect command files
COMMANDS=$(find "$PROJECT_ROOT/.claude/commands" -name "*.md" -type f | sort | \
  sed "s|$PROJECT_ROOT/|./|" | jq -R . | jq -s .)

# Collect agent files
AGENTS=$(find "$PROJECT_ROOT/.claude/agents" -name "*.md" -type f 2>/dev/null | sort | \
  sed "s|$PROJECT_ROOT/|./|" | jq -R . | jq -s . || echo "[]")

# Update plugin.json
jq --argjson commands "$COMMANDS" \
   --argjson agents "$AGENTS" \
   '.commands = $commands | .agents = $agents' \
   "$PLUGIN_JSON" > "$PLUGIN_JSON.tmp"

mv "$PLUGIN_JSON.tmp" "$PLUGIN_JSON"

echo "✓ Updated plugin.json:"
echo "  Commands: $(echo "$COMMANDS" | jq 'length')"
echo "  Agents: $(echo "$AGENTS" | jq 'length')"
