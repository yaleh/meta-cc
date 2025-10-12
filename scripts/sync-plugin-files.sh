#!/bin/bash
# Prepare plugin files for release packaging

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DIST_DIR="$PROJECT_ROOT/dist"
CAPABILITIES_DIR="$PROJECT_ROOT/capabilities"

echo "Preparing plugin files for release packaging..."

# Verify source directories exist
if [ ! -f "$PROJECT_ROOT/.claude/commands/meta.md" ]; then
    echo "ERROR: .claude/commands/meta.md not found"
    exit 1
fi

if [ ! -d "$CAPABILITIES_DIR/commands" ]; then
    echo "ERROR: $CAPABILITIES_DIR/commands directory not found"
    exit 1
fi

# Create dist directories
mkdir -p "$DIST_DIR/commands" "$DIST_DIR/agents"

# Copy ONLY the unified meta command (capabilities are distributed separately)
echo "  Copying unified meta command from .claude/commands/..."
cp "$PROJECT_ROOT/.claude/commands/meta.md" "$DIST_DIR/commands/"

# Copy agents
echo "  Copying agents from .claude/agents/..."
if ls "$PROJECT_ROOT/.claude/agents/"*.md 1> /dev/null 2>&1; then
    cp "$PROJECT_ROOT/.claude/agents/"*.md "$DIST_DIR/agents/"
fi

if [ -d "$CAPABILITIES_DIR/agents" ] && ls "$CAPABILITIES_DIR/agents/"*.md 1> /dev/null 2>&1; then
    echo "  Copying agents from $CAPABILITIES_DIR/agents/..."
    cp "$CAPABILITIES_DIR/agents/"*.md "$DIST_DIR/agents/"
fi

# Count files
CMD_COUNT=$(find "$DIST_DIR/commands" -name "*.md" | wc -l)
AGENT_COUNT=$(find "$DIST_DIR/agents" -name "*.md" | wc -l)

echo "✓ Plugin files synced to $DIST_DIR/"
echo "✓ Total: $CMD_COUNT command file (unified meta.md), $AGENT_COUNT agent files"
echo "  Note: 13 capability files distributed separately in capabilities-latest.tar.gz"
