#!/usr/bin/env bats
#
# Tests for release.sh Step 2 delegation to bump-plugin-version.sh
#
# Run with: bats tests/scripts/release-version-update.bats
#

WORKTREE_ROOT="$(cd "$(dirname "$BATS_TEST_FILENAME")/../.." && pwd)"
RELEASE_SH="$WORKTREE_ROOT/scripts/release/release.sh"

@test "release.sh does NOT contain inline jq marketplace.json update pattern" {
    run grep -q 'jq --arg ver.*marketplace.json' "$RELEASE_SH"
    [ "$status" -ne 0 ]
}

@test "release.sh DOES call bump-plugin-version.sh with --non-interactive" {
    run grep -q 'bump-plugin-version.sh.*--non-interactive' "$RELEASE_SH"
    [ "$status" -eq 0 ]
}
