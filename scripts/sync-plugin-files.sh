#!/bin/bash
# Prepare plugin files for release packaging
# Usage:
#   ./sync-plugin-files.sh          - Sync files
#   ./sync-plugin-files.sh --verify - Verify sync (don't modify files)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DIST_DIR="$PROJECT_ROOT/dist"
CAPABILITIES_DIR="$PROJECT_ROOT/capabilities"

# Parse arguments
VERIFY_MODE=false
if [ "$1" = "--verify" ]; then
    VERIFY_MODE=true
    echo "=== Plugin File Sync Verification ==="
    echo ""
else
    echo "Preparing plugin files for release packaging..."
fi

if [ "$VERIFY_MODE" = true ]; then
    # VERIFY MODE: Check that sync was done correctly
    echo "[1/3] Verifying dist/ structure..."
    if [ ! -d "$DIST_DIR/commands" ] || [ ! -d "$DIST_DIR/agents" ]; then
        echo "❌ ERROR: Plugin file sync failed - dist/ directory not created"
        exit 1
    fi
    echo "✓ dist/ structure verified"
    echo ""

    echo "[2/3] Checking file count..."
    DIST_CMD_COUNT=$(find "$DIST_DIR/commands" -name "*.md" 2>/dev/null | wc -l)
    EXPECTED_COUNT=1

    if [ "$DIST_CMD_COUNT" -ne "$EXPECTED_COUNT" ]; then
        echo "❌ ERROR: Command file count mismatch: expected $EXPECTED_COUNT, got $DIST_CMD_COUNT"
        exit 1
    fi
    echo "✓ File count verified: $DIST_CMD_COUNT command file(s)"
    echo ""

    echo "[3/3] Verifying file content..."
    if [ ! -f "$DIST_DIR/commands/meta.md" ]; then
        echo "❌ ERROR: meta.md not found in dist/commands/"
        exit 1
    fi
    echo "✓ meta.md verified"
    echo ""

    echo "✅ Plugin file sync verification passed"
else
    # SYNC MODE: Perform the sync
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

    # Copy skills directory with all supporting files
    echo "  Copying skills from .claude/skills/..."
    if [ -d "$PROJECT_ROOT/.claude/skills" ]; then
        cp -r "$PROJECT_ROOT/.claude/skills/"* "$DIST_DIR/skills/"
        SKILL_COUNT=$(find "$DIST_DIR/skills" -name "SKILL.md" 2>/dev/null | wc -l)
        SKILL_FILES=$(find "$DIST_DIR/skills" -type f 2>/dev/null | wc -l)
        echo "    ✓ Copied $SKILL_COUNT skills ($SKILL_FILES total files)"
    fi

    # Count files
    CMD_COUNT=$(find "$DIST_DIR/commands" -name "*.md" 2>/dev/null | wc -l)
    AGENT_COUNT=$(find "$DIST_DIR/agents" -name "*.md" 2>/dev/null | wc -l)
    SKILL_COUNT=$(find "$DIST_DIR/skills" -name "SKILL.md" 2>/dev/null | wc -l)

    echo "✓ Plugin files synced to $DIST_DIR/"
    echo "✓ Total: $CMD_COUNT command, $AGENT_COUNT agents, $SKILL_COUNT skills"
    echo "  Note: 13 capability files distributed separately in capabilities-latest.tar.gz"
fi
