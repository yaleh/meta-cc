#!/bin/bash
# Prepare plugin files for release packaging
# Usage:
#   ./sync-plugin-files.sh          - Sync files
#   ./sync-plugin-files.sh --verify - Verify sync (don't modify files)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DIST_DIR="$PROJECT_ROOT/dist"

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
    echo "[1/4] Verifying dist/ structure..."
    if [ ! -d "$DIST_DIR/commands" ]; then
        echo "❌ ERROR: Plugin file sync failed - dist/commands/ directory not created"
        exit 1
    fi
    echo "✓ dist/ structure verified"
    echo ""

    echo "[2/4] Checking file count..."
    DIST_CMD_COUNT=$(find "$DIST_DIR/commands" -name "*.md" 2>/dev/null | wc -l)
    EXPECTED_COUNT=3

    if [ "$DIST_CMD_COUNT" -ne "$EXPECTED_COUNT" ]; then
        echo "❌ ERROR: Command file count mismatch: expected $EXPECTED_COUNT, got $DIST_CMD_COUNT"
        exit 1
    fi
    echo "✓ File count verified: $DIST_CMD_COUNT command file(s)"
    echo ""

    echo "[3/4] Verifying command file content..."
    for cmd in prompt-find prompt-list prompt-show; do
        if [ ! -f "$DIST_DIR/commands/${cmd}.md" ]; then
            echo "❌ ERROR: ${cmd}.md not found in dist/commands/"
            exit 1
        fi
    done
    echo "✓ All 3 commands verified"
    echo ""

    echo "[4/4] Verifying Codex plugin source files..."
    for path in \
        "plugin-src/.codex-plugin/plugin.json" \
        "plugin-src/.codex-mcp.json" \
        "plugin-src/skills/prompt-find/SKILL.md" \
        "plugin-src/skills/prompt-list/SKILL.md" \
        "plugin-src/skills/prompt-show/SKILL.md"; do
        if [ ! -f "$PROJECT_ROOT/$path" ]; then
            echo "❌ ERROR: Codex plugin file missing: $path"
            exit 1
        fi
    done
    jq . "$PROJECT_ROOT/plugin-src/.codex-plugin/plugin.json" >/dev/null
    jq -e '.mcpServers["meta-cc"]' "$PROJECT_ROOT/plugin-src/.codex-mcp.json" >/dev/null
    echo "✓ Codex plugin source files verified"
    echo ""

    echo "[5/5] Verifying version consistency across plugin manifests..."
    PLUGIN_VERSION=$(jq -r '.version' "$PROJECT_ROOT/plugin-src/.claude-plugin/plugin.json")
    MARKETPLACE_VERSION=$(jq -r '.plugins[0].version' "$PROJECT_ROOT/.claude-plugin/marketplace.json")
    if [ "$PLUGIN_VERSION" != "$MARKETPLACE_VERSION" ]; then
        echo "❌ ERROR: Version mismatch between plugin manifests:"
        echo "  plugin-src/.claude-plugin/plugin.json: $PLUGIN_VERSION"
        echo "  .claude-plugin/marketplace.json:       $MARKETPLACE_VERSION"
        echo ""
        echo "Run 'scripts/release/bump-plugin-version.sh' to sync both files."
        exit 1
    fi
    echo "✓ Version consistent: $PLUGIN_VERSION"
    echo ""

    echo "✅ Plugin file sync verification passed"
else
    # SYNC MODE: Perform the sync
    # Create dist directories (clean commands to remove stale files)
    mkdir -p "$DIST_DIR/commands"
    rm -f "$DIST_DIR/commands/"*.md 2>/dev/null || true

    # Copy published commands (source: plugin-src/commands/)
    echo "  Copying published commands from plugin-src/commands/..."
    PUBLISHED_COMMANDS="prompt-find prompt-list prompt-show"
    for cmd in $PUBLISHED_COMMANDS; do
        if [ -f "$PROJECT_ROOT/plugin-src/commands/${cmd}.md" ]; then
            cp "$PROJECT_ROOT/plugin-src/commands/${cmd}.md" "$DIST_DIR/commands/"
        else
            echo "  WARNING: Expected command not found: plugin-src/commands/${cmd}.md"
        fi
    done

    # Count files
    CMD_COUNT=$(find "$DIST_DIR/commands" -name "*.md" 2>/dev/null | wc -l)

    echo "✓ Plugin files synced to $DIST_DIR/"
    echo "✓ Total: $CMD_COUNT command(s)"
fi
