#!/bin/bash
# Refresh a registered user-scope meta-cc plugin through Codex's supported
# plugin manager, then verify every discovery surface resolves one version.
#
# Semantics (audited against real Codex CLI 0.146.0): `codex plugin add`
# materializes the new version B and REMOVES the previously cached version A
# *before* any post-install verification of B runs. This script therefore makes
# no "keep A until B is verified" assumption and performs NO automatic rollback:
# once `plugin add` has been invoked, the old cache is already gone, so on any
# post-install verification failure we exit non-zero with an actionable error
# naming the offending artifact plus an explicit manual recovery command.

set -euo pipefail

SOURCE_ROOT="${1:-${HOME}/.local/share/meta-cc}"
CODEX_HOME="${CODEX_HOME:-${CODEX_DIR:-${HOME}/.codex}}"
CODEX_BIN="${CODEX_BIN:-codex}"
PLUGIN_ID="meta-cc@meta-cc-marketplace"
MARKETPLACE="meta-cc-marketplace"
CONFIG="$CODEX_HOME/config.toml"
CACHE_ROOT="$CODEX_HOME/plugins/cache/$MARKETPLACE/meta-cc"
export CODEX_HOME

# fail: a pre-destructive failure. `plugin add` has not run, so the previously
# cached version is still intact; the message says so and gives recovery.
fail() {
    printf 'ERROR: %s\n' "$*" >&2
    exit 1
}

# recovery_hint: the manual command a user runs to reach a consistent state
# after a post-install failure. No automatic rollback is performed.
recovery_hint() {
    printf 'No automatic rollback was performed: Codex removed the previous cache before verification, so it cannot be restored here.\n' >&2
    printf 'Recovery (manual): fix the artifact above, then re-run the supported refresh:\n' >&2
    printf '  CODEX_HOME=%s %s plugin add %s --json\n' "$CODEX_HOME" "$CODEX_BIN" "$PLUGIN_ID" >&2
    printf '  (equivalent to: make install-user-codex)\n' >&2
}

# fail_post_install: a failure after Codex has destructively replaced the cache.
# Exit non-zero with an actionable error + explicit manual recovery; no rollback.
fail_post_install() {
    printf 'ERROR: %s\n' "$*" >&2
    recovery_hint
    exit 1
}

# probe_mcp_version <binary>: ask an MCP server binary to self-report its
# version over MCP `initialize` (result.serverInfo.version). The real binary
# interleaves slog JSON log lines with the JSON-RPC response on stdout, so we
# select only the line carrying serverInfo.version. Always returns 0; an empty
# result means the binary did not self-report a usable version.
probe_mcp_version() {
    { printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}\n' \
        | timeout 10 "$1" 2>/dev/null \
        | jq -r '.result?.serverInfo?.version? // empty' 2>/dev/null \
        | head -n 1; } || true
}

# Unregistered marketplace -> nothing to refresh; safe no-op.
if [ ! -f "$CONFIG" ] || ! grep -q '^\[marketplaces\.meta-cc-marketplace\]$' "$CONFIG"; then
    printf 'Codex marketplace %s is not registered in %s; nothing to refresh.\n' "$MARKETPLACE" "$CONFIG"
    exit 0
fi

# --- Pre-flight (runs before any destructive Codex call) --------------------
[ -f "$SOURCE_ROOT/.codex-plugin/plugin.json" ] || fail "missing Codex manifest: $SOURCE_ROOT/.codex-plugin/plugin.json"
[ -f "$SOURCE_ROOT/.claude-plugin/marketplace.json" ] || fail "missing shared marketplace manifest: $SOURCE_ROOT/.claude-plugin/marketplace.json"
command -v jq >/dev/null 2>&1 || fail "jq is required to verify the Codex plugin upgrade"
command -v "$CODEX_BIN" >/dev/null 2>&1 || fail "Codex CLI not found ($CODEX_BIN). No cache changes were made. Install Codex CLI 0.146+ and re-run: make install-user-codex"

# Unsupported refresh behavior must fail visibly (no silent success, no
# destructive fallback). These probes run BEFORE `plugin add`, so the previous
# cache is still intact when they fail.
for probe in "plugin add --help" "plugin list --help" "mcp list --help"; do
    # shellcheck disable=SC2086 -- probe is intentionally split into arguments.
    if ! "$CODEX_BIN" $probe >/dev/null 2>&1; then
        fail "Codex CLI does not expose the supported plugin refresh commands ($probe). No cache changes were made; the previous plugin is untouched. Upgrade Codex CLI to 0.146+ and re-run: make install-user-codex"
    fi
done

source_version=$(jq -er '.version | select(type == "string" and length > 0)' "$SOURCE_ROOT/.codex-plugin/plugin.json") \
    || fail "invalid version in $SOURCE_ROOT/.codex-plugin/plugin.json"
marketplace_version=$(jq -er '.plugins[] | select(.name == "meta-cc") | .version' "$SOURCE_ROOT/.claude-plugin/marketplace.json") \
    || fail "meta-cc version missing from $SOURCE_ROOT/.claude-plugin/marketplace.json"
[ "$source_version" = "$marketplace_version" ] \
    || fail "source manifests disagree before refresh: Codex=$source_version marketplace=$marketplace_version. No cache changes were made."

# --- Destructive refresh ----------------------------------------------------
# `codex plugin add` materializes B and removes A before returning. From this
# point on the previous cache is gone; any failure below cannot be rolled back
# automatically and must surface the manual recovery command.
add_output=$("$CODEX_BIN" plugin add "$PLUGIN_ID" --json) \
    || fail_post_install "Codex plugin refresh (plugin add) failed after the previous cache may already have been removed."
installed_version=$(jq -er '.version' <<<"$add_output") || fail_post_install "Codex refresh returned no installed version: $add_output"
installed_path=$(jq -er '.installedPath' <<<"$add_output") || fail_post_install "Codex refresh returned no installed path: $add_output"
[ "$installed_version" = "$source_version" ] \
    || fail_post_install "Codex selected version $installed_version but source is $source_version (discovery/source disagreement)."
[ "$installed_path" = "$CACHE_ROOT/$source_version" ] \
    || fail_post_install "Codex selected unexpected cache path $installed_path (expected $CACHE_ROOT/$source_version)."

# --- Post-install verification: every surface must resolve source_version ---
list_output=$("$CODEX_BIN" plugin list --json) \
    || fail_post_install "Codex could not report discovery metadata (plugin list) after refresh."
resolved_version=$(jq -er --arg id "$PLUGIN_ID" '.installed[] | select(.pluginId == $id and .installed == true and .enabled == true) | .version' <<<"$list_output") \
    || fail_post_install "discovery metadata: $PLUGIN_ID is not reported installed and enabled after refresh."
[ "$resolved_version" = "$source_version" ] \
    || fail_post_install "discovery metadata resolves $resolved_version but source/cache are $source_version."

[ -f "$installed_path/.codex-plugin/plugin.json" ] \
    || fail_post_install "cache artifact missing: $installed_path/.codex-plugin/plugin.json"
[ "$(jq -r .version "$installed_path/.codex-plugin/plugin.json")" = "$source_version" ] \
    || fail_post_install "cache Codex manifest version does not match $source_version ($installed_path)."
[ -f "$installed_path/.claude-plugin/marketplace.json" ] \
    || fail_post_install "cache artifact missing: $installed_path/.claude-plugin/marketplace.json"
[ "$(jq -r '.plugins[] | select(.name == "meta-cc") | .version' "$installed_path/.claude-plugin/marketplace.json")" = "$source_version" ] \
    || fail_post_install "cache marketplace manifest version does not match $source_version ($installed_path)."
for skill in prompt-find prompt-list prompt-show; do
    [ -f "$installed_path/skills/$skill/SKILL.md" ] || fail_post_install "cache is incomplete: missing skills/$skill/SKILL.md under $installed_path"
done

mcp_output=$("$CODEX_BIN" mcp list --json) \
    || fail_post_install "Codex could not report the meta-cc MCP registration (mcp list) after refresh."
mcp_count=$(jq '[.[] | select(.name == "meta-cc")] | length' <<<"$mcp_output")
[ "$mcp_count" = 1 ] || fail_post_install "expected exactly one meta-cc MCP registration, found $mcp_count."
mcp_cwd=$(jq -er '.[] | select(.name == "meta-cc") | .transport.cwd' <<<"$mcp_output") \
    || fail_post_install "meta-cc MCP registration has no cwd."
mcp_command=$(jq -er '.[] | select(.name == "meta-cc") | .transport.command' <<<"$mcp_output") \
    || fail_post_install "meta-cc MCP registration has no command."
[ "$(cd "$mcp_cwd" && pwd -P)" = "$(cd "$installed_path" && pwd -P)" ] \
    || fail_post_install "meta-cc MCP cwd ($mcp_cwd) does not resolve to the selected cache $installed_path."
case "$mcp_command" in
    /*) mcp_binary="$mcp_command" ;;
    *) mcp_binary="$mcp_cwd/${mcp_command#./}" ;;
esac
[ -x "$mcp_binary" ] || fail_post_install "MCP binary is missing or not executable: $mcp_binary"

# The MCP binary must SELF-REPORT the same version as source/cache metadata --
# executability alone is not sufficient. Query it over MCP `initialize`.
mcp_reported=$(probe_mcp_version "$mcp_binary")
[ -n "$mcp_reported" ] \
    || fail_post_install "MCP binary $mcp_binary did not self-report a version over MCP initialize."
[ "$mcp_reported" = "$source_version" ] \
    || fail_post_install "MCP binary self-reports version $mcp_reported but source/cache metadata is $source_version."

printf 'Codex plugin %s is aligned at version %s (%s).\n' "$PLUGIN_ID" "$source_version" "$installed_path"
printf 'Discovery metadata, cache manifests, source manifests, and MCP binary all report %s.\n' "$source_version"
printf 'Start a new Codex session to load the upgraded skills and MCP server; running sessions do not hot-reload plugins.\n'
