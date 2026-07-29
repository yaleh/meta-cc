#!/usr/bin/env bats

# Upgrade/restart semantics for the user-scope Codex plugin refresh.
#
# Models the REAL Codex CLI 0.146.0 behavior proven by a fresh-context audit:
# `codex plugin add` materializes new version B and REMOVES cached version A
# *before* any post-install verification of B. The fake Codex double reproduces
# this destructive replacement, so the script assumes nothing survives and does
# NO automatic rollback: on verification failure it exits non-zero with an
# actionable error naming the artifact plus an explicit manual recovery command.

WORKTREE_ROOT="$(cd "$(dirname "$BATS_TEST_FILENAME")/../.." && pwd)"
REFRESH="$WORKTREE_ROOT/scripts/install/refresh-codex-plugin.sh"

# make_source <version> [mcp_version]: builds a user-scope source tree whose
# manifests declare <version>. The fake MCP binary self-reports [mcp_version]
# (default <version>) over MCP `initialize` (the real serverInfo.version); pass
# a differing second argument to simulate a binary/metadata disagreement.
make_source() {
    local version="$1"
    local mcp_version="${2:-$1}"
    printf '{"name":"meta-cc","version":"%s"}\n' "$version" > "$SOURCE/.codex-plugin/plugin.json"
    printf '{"plugins":[{"name":"meta-cc","version":"%s","source":"."}]}\n' "$version" > "$SOURCE/.claude-plugin/marketplace.json"
    printf '{"mcpServers":{"meta-cc":{"command":"./bin/meta-cc-mcp","cwd":"."}}}\n' > "$SOURCE/.codex-mcp.json"
    for skill in prompt-find prompt-list prompt-show; do
        printf '# %s\n' "$skill" > "$SOURCE/skills/$skill/SKILL.md"
    done
    cat > "$SOURCE/bin/meta-cc-mcp" <<EOF
#!/bin/sh
# Fake MCP server: answer one \`initialize\` by self-reporting result.serverInfo.version.
while IFS= read -r line; do
    case "\$line" in
        *initialize*)
            printf '{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2024-11-05","serverInfo":{"name":"meta-cc-mcp","version":"%s"}}}\n' "$mcp_version"
            exit 0
            ;;
    esac
done
EOF
    chmod +x "$SOURCE/bin/meta-cc-mcp"
}

setup() {
    export CODEX_HOME="$BATS_TEST_TMPDIR/codex-home"
    export SOURCE="$BATS_TEST_TMPDIR/source"
    export CODEX_BIN="$BATS_TEST_TMPDIR/bin/codex"
    mkdir -p "$CODEX_HOME" "$SOURCE/.codex-plugin" "$SOURCE/.claude-plugin" \
        "$SOURCE/skills/prompt-find" "$SOURCE/skills/prompt-list" "$SOURCE/skills/prompt-show" \
        "$SOURCE/bin" "$(dirname "$CODEX_BIN")"
    cat > "$CODEX_HOME/config.toml" <<EOF
[marketplaces.meta-cc-marketplace]
source_type = "local"
source = "$SOURCE"

[plugins."meta-cc@meta-cc-marketplace"]
enabled = true
EOF
    make_source 2.0.0
}

# write_fake_codex installs a Codex double whose `plugin add` models the real
# CLI's destructive replacement: materialize B, then remove every other cached
# version (old A) BEFORE returning -- i.e. before post-install verification.
# Each removed old version is journaled to $CODEX_HOME/events as proof.
write_fake_codex() {
    cat > "$CODEX_BIN" <<'EOF'
#!/bin/bash
set -euo pipefail
home="${CODEX_HOME:?}"
source_root="${SOURCE:?}"
cache_root="$home/plugins/cache/meta-cc-marketplace/meta-cc"
case "$*" in
  "plugin add --help"|"plugin list --help"|"mcp list --help") exit 0 ;;
  "plugin add meta-cc@meta-cc-marketplace --json")
    version=$(jq -r .version "$source_root/.codex-plugin/plugin.json")
    target="$cache_root/$version"
    rm -rf "$cache_root/.incoming-$version"
    mkdir -p "$cache_root/.incoming-$version"
    cp -a "$source_root/." "$cache_root/.incoming-$version/"
    mv "$cache_root/.incoming-$version" "$target"
    # Destructive replacement: the real CLI deletes the previously cached
    # version(s) here, before any post-install verification of B runs.
    for old in "$cache_root"/*; do
      [ -e "$old" ] || continue
      base=$(basename "$old")
      [ "$base" = "$version" ] && continue
      printf 'removed-old:%s\n' "$base" >> "$home/events"
      rm -rf "$old"
    done
    printf '%s\n' "$version" > "$home/selected-version"
    printf '{"version":"%s","installedPath":"%s"}\n' "$version" "$target"
    ;;
  "plugin list --json")
    version=$(cat "$home/selected-version")
    printf '{"installed":[{"pluginId":"meta-cc@meta-cc-marketplace","version":"%s","installed":true,"enabled":true}]}\n' "$version"
    ;;
  "mcp list --json")
    version=$(cat "$home/selected-version")
    printf '[{"name":"meta-cc","transport":{"command":"./bin/meta-cc-mcp","cwd":"%s"}}]\n' "$cache_root/$version"
    ;;
  *) exit 64 ;;
esac
EOF
    chmod +x "$CODEX_BIN"
}

# seed_old_cache <version> plants a previously installed version A as the
# currently selected version, giving the upgrade something to replace.
seed_old_cache() {
    local v="$1"
    local old="$CODEX_HOME/plugins/cache/meta-cc-marketplace/meta-cc/$v"
    mkdir -p "$old/skills/prompt-find"
    printf '# old %s\n' "$v" > "$old/skills/prompt-find/SKILL.md"
    printf '%s\n' "$v" > "$CODEX_HOME/selected-version"
}

@test "A->B upgrade: Codex destructively replaces A, every surface aligns at B, restart required" {
    write_fake_codex
    seed_old_cache 1.0.0
    old="$CODEX_HOME/plugins/cache/meta-cc-marketplace/meta-cc/1.0.0"

    run "$REFRESH" "$SOURCE"
    echo "$output"
    [ "$status" -eq 0 ]
    # B selected and fully materialized.
    [ "$(cat "$CODEX_HOME/selected-version")" = "2.0.0" ]
    [ -x "$CODEX_HOME/plugins/cache/meta-cc-marketplace/meta-cc/2.0.0/bin/meta-cc-mcp" ]
    # A was removed by Codex inside plugin add (before verification); the script
    # resurrected/pruned nothing itself.
    [ ! -d "$old" ]
    grep -q 'removed-old:1.0.0' "$CODEX_HOME/events"
    # All surfaces agree, so the next session has no dangling A-path skill.
    [[ "$output" == *"aligned at version 2.0.0"* ]]
    [[ "$output" == *"Start a new Codex session"* ]]
}

@test "verification failure (MCP binary self-reports wrong version) exits non-zero, no rollback" {
    write_fake_codex
    seed_old_cache 1.0.0
    make_source 2.0.0 1.5.0   # manifests say 2.0.0; MCP binary self-reports 1.5.0
    old="$CODEX_HOME/plugins/cache/meta-cc-marketplace/meta-cc/1.0.0"

    run "$REFRESH" "$SOURCE"
    echo "$output"
    [ "$status" -ne 0 ]
    # Error names the inconsistent artifact and the disagreement.
    [[ "$output" == *"MCP binary self-reports version 1.5.0"* ]]
    [[ "$output" == *"2.0.0"* ]]
    # Explicit manual recovery command is provided, and no automatic rollback
    # is claimed or attempted.
    [[ "$output" == *"Recovery"* ]]
    [[ "$output" == *"plugin add meta-cc@meta-cc-marketplace"* ]]
    [[ "$output" == *"No automatic rollback"* ]]
    [ ! -d "$old" ]   # A already destroyed by Codex; NOT restored
}

@test "verification failure (missing skill artifact) exits non-zero naming the artifact with recovery" {
    write_fake_codex
    seed_old_cache 1.0.0
    rm "$SOURCE/skills/prompt-show/SKILL.md"
    old="$CODEX_HOME/plugins/cache/meta-cc-marketplace/meta-cc/1.0.0"

    run "$REFRESH" "$SOURCE"
    echo "$output"
    [ "$status" -ne 0 ]
    [[ "$output" == *"skills/prompt-show/SKILL.md"* ]]
    [[ "$output" == *"Recovery"* ]]
    [ ! -d "$old" ]   # no rollback: A gone and not restored
}

@test "unsupported Codex plugin manager fails visibly before any destructive call" {
    seed_old_cache 1.0.0
    cat > "$CODEX_BIN" <<'EOF'
#!/bin/sh
exit 2
EOF
    chmod +x "$CODEX_BIN"

    run "$REFRESH" "$SOURCE"
    echo "$output"
    [ "$status" -ne 0 ]
    [[ "$output" == *"does not expose the supported plugin refresh commands"* ]]
    [[ "$output" == *"No cache changes were made"* ]]
    # The previous usable A is untouched because no destructive call was made.
    [ -f "$CODEX_HOME/plugins/cache/meta-cc-marketplace/meta-cc/1.0.0/skills/prompt-find/SKILL.md" ]
}

@test "unregistered isolated CODEX_HOME is a safe no-op" {
    rm "$CODEX_HOME/config.toml"
    cat > "$CODEX_BIN" <<'EOF'
#!/bin/sh
exit 99
EOF
    chmod +x "$CODEX_BIN"

    run "$REFRESH" "$SOURCE"
    [ "$status" -eq 0 ]
    [[ "$output" == *"not registered"* ]]
    [ ! -d "$CODEX_HOME/plugins" ]
}
