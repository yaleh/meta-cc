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
mkdir -p "$DIST_DIR/commands" "$DIST_DIR/agents" "$DIST_DIR/skills"

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

# Copy skills directory with all supporting files (exclude archived)
echo "  Copying skills from .claude/skills/..."
if [ -d "$PROJECT_ROOT/.claude/skills" ]; then
    # Copy each skill directory individually, excluding archived
    for skill_dir in "$PROJECT_ROOT/.claude/skills/"*/; do
        skill_name=$(basename "$skill_dir")
        if [ "$skill_name" != "archived" ]; then
            cp -r "$skill_dir" "$DIST_DIR/skills/"
        fi
    done
    SKILL_COUNT=$(find "$DIST_DIR/skills" -name "SKILL.md" 2>/dev/null | wc -l)
    SKILL_FILES=$(find "$DIST_DIR/skills" -type f 2>/dev/null | wc -l)
    echo "    ✓ Copied $SKILL_COUNT skills ($SKILL_FILES total files, excluded archived)"
fi

# Count files
CMD_COUNT=$(find "$DIST_DIR/commands" -name "*.md" 2>/dev/null | wc -l)
AGENT_COUNT=$(find "$DIST_DIR/agents" -name "*.md" 2>/dev/null | wc -l)
SKILL_COUNT=$(find "$DIST_DIR/skills" -name "SKILL.md" 2>/dev/null | wc -l)

echo "✓ Plugin files synced to $DIST_DIR/"
echo "✓ Total: $CMD_COUNT command, $AGENT_COUNT agents, $SKILL_COUNT skills"
echo "  Note: 13 capability files distributed separately in capabilities-latest.tar.gz"
