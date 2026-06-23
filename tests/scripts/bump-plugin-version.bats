#!/usr/bin/env bats
#
# Unit tests for scripts/release/bump-plugin-version.sh --version and --non-interactive flags
#
# Run with: bats tests/scripts/bump-plugin-version.bats
#

WORKTREE_ROOT="$(cd "$(dirname "$BATS_TEST_FILENAME")/../.." && pwd)"

setup() {
    # Save original files content
    export ORIG_PLUGIN_JSON=$(cat "$WORKTREE_ROOT/plugin-src/.claude-plugin/plugin.json")
    export ORIG_MARKETPLACE_JSON=$(cat "$WORKTREE_ROOT/.claude-plugin/marketplace.json")
    export ORIG_SRC_MARKETPLACE_JSON=$(cat "$WORKTREE_ROOT/plugin-src/.claude-plugin/marketplace.json")
}

teardown() {
    # Restore original files
    echo "$ORIG_PLUGIN_JSON" > "$WORKTREE_ROOT/plugin-src/.claude-plugin/plugin.json"
    echo "$ORIG_MARKETPLACE_JSON" > "$WORKTREE_ROOT/.claude-plugin/marketplace.json"
    echo "$ORIG_SRC_MARKETPLACE_JSON" > "$WORKTREE_ROOT/plugin-src/.claude-plugin/marketplace.json"
}

@test "bump-plugin-version.sh: --version 9.9.9 --non-interactive sets all 3 JSON files to 9.9.9" {
    run bash "$WORKTREE_ROOT/scripts/release/bump-plugin-version.sh" --version 9.9.9 --non-interactive

    [ "$status" -eq 0 ]
    [[ "$output" =~ "updated to 9.9.9" ]]

    PLUGIN_VER=$(jq -r '.version' "$WORKTREE_ROOT/plugin-src/.claude-plugin/plugin.json")
    MARKET_VER=$(jq -r '.plugins[0].version' "$WORKTREE_ROOT/.claude-plugin/marketplace.json")
    SRC_MARKET_VER=$(jq -r '.plugins[0].version' "$WORKTREE_ROOT/plugin-src/.claude-plugin/marketplace.json")

    [ "$PLUGIN_VER" = "9.9.9" ]
    [ "$MARKET_VER" = "9.9.9" ]
    [ "$SRC_MARKET_VER" = "9.9.9" ]
}

@test "bump-plugin-version.sh: --version bad-format --non-interactive exits non-zero" {
    run bash "$WORKTREE_ROOT/scripts/release/bump-plugin-version.sh" --version bad-format --non-interactive

    [ "$status" -ne 0 ]
}

@test "bump-plugin-version.sh: --version 9.9.9 --non-interactive does NOT create a git commit" {
    COMMIT_BEFORE=$(git -C "$WORKTREE_ROOT" rev-parse HEAD)

    run bash "$WORKTREE_ROOT/scripts/release/bump-plugin-version.sh" --version 9.9.9 --non-interactive

    [ "$status" -eq 0 ]

    COMMIT_AFTER=$(git -C "$WORKTREE_ROOT" rev-parse HEAD)
    [ "$COMMIT_BEFORE" = "$COMMIT_AFTER" ]
}
