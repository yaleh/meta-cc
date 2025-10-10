#!/bin/bash
# Sync plugin files from .claude/ to root for release

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Syncing plugin files from .claude/ to root directories..."

# Verify source directories exist
if [ ! -d "$PROJECT_ROOT/.claude/commands" ]; then
    echo "ERROR: .claude/commands directory not found"
    exit 1
fi

if [ ! -d "$PROJECT_ROOT/.claude/agents" ]; then
    echo "ERROR: .claude/agents directory not found"
    exit 1
fi

# Remove old root directories if they exist
rm -rf "$PROJECT_ROOT/commands" "$PROJECT_ROOT/agents"

# Copy from .claude/ to root
cp -r "$PROJECT_ROOT/.claude/commands" "$PROJECT_ROOT/commands"
cp -r "$PROJECT_ROOT/.claude/agents" "$PROJECT_ROOT/agents"

echo "✓ Synced commands/ and agents/ from .claude/"

# Count files for verification
CMD_COUNT=$(find "$PROJECT_ROOT/commands" -name "*.md" | wc -l)
AGENT_COUNT=$(find "$PROJECT_ROOT/agents" -name "*.md" | wc -l)

echo "✓ Synced $CMD_COUNT command files and $AGENT_COUNT agent files"
echo "Files ready for plugin packaging"
