#!/bin/bash
# Codex plugin-manager E2E test for meta-cc (DIR-026).
#
# Verifies the SUPPORTED Codex installation path against the real, installed
# Codex CLI (not a mock):
#
#   Preferred:  codex plugin marketplace add <bundle>
#               codex plugin add meta-cc@meta-cc-marketplace
#   Fallback:   codex mcp add meta-cc -- <absolute-path-to-meta-cc-mcp>
#
# Every codex invocation in this script is pinned to an isolated CODEX_HOME
# created under a throwaway temp directory. Nothing here ever touches the
# operator's real ~/.codex or a running Codex session.
#
# Tested against: codex-cli 0.145.x (Codex CLI `plugin`/`mcp` subcommand
# surface as of that release). If the installed `codex` binary lacks the
# `plugin` or `mcp` subcommands this test SKIPS with an explicit reason
# rather than silently passing.
#
# Usage: ./tests/e2e/codex-plugin-manager-e2e.sh <bundle_dir> [mcp_binary]
#   bundle_dir  - a release-bundle directory built by `make bundle-release`
#                 (must contain bin/, commands/, skills/, lib/,
#                 .claude-plugin/, .codex-plugin/, .codex-mcp.json)
#   mcp_binary  - path to a built meta-cc-mcp binary used for the minimal
#                 `codex mcp add` fallback scenario (default: ./bin/meta-cc-mcp)

set -euo pipefail

BUNDLE_DIR_ARG="${1:?usage: $0 <bundle_dir> [mcp_binary]}"
MCP_BINARY_ARG="${2:-./bin/meta-cc-mcp}"
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

fail() {
    echo -e "${RED}FAILED:${NC} $1" >&2
    exit 1
}

pass() {
    echo -e "  ${GREEN}PASS${NC} - $1"
}

skip_all() {
    echo -e "${YELLOW}SKIP:${NC} $1"
    echo "(This is a soft skip, not a failure: the plugin-manager path cannot"
    echo " be exercised without a Codex CLI build that exposes 'codex plugin'"
    echo " and 'codex mcp'. Tested/expected range: codex-cli >= 0.145.)"
    exit 0
}

require_file() {
    [ -f "$1" ] || fail "missing file: $1"
}

require_cmd() {
    command -v "$1" >/dev/null 2>&1 || fail "$1 is required"
}

require_cmd jq
require_cmd python3

if [ ! -d "$BUNDLE_DIR_ARG" ]; then
    fail "bundle directory not found: $BUNDLE_DIR_ARG"
fi
BUNDLE_DIR="$(cd "$BUNDLE_DIR_ARG" && pwd)"

if [ ! -f "$MCP_BINARY_ARG" ]; then
    fail "mcp binary not found: $MCP_BINARY_ARG"
fi
MCP_BINARY="$(cd "$(dirname "$MCP_BINARY_ARG")" && pwd)/$(basename "$MCP_BINARY_ARG")"

echo "=========================================="
echo "Codex Plugin-Manager E2E Test"
echo "=========================================="
echo "Bundle:     $BUNDLE_DIR"
echo "MCP binary: $MCP_BINARY"
echo "=========================================="
echo ""

# ------------------------------------------------------------------
# Safety net: every codex invocation below is via run_codex(), which
# forces CODEX_HOME to one of our own temp dirs. This guards against a
# future edit accidentally calling the bare `codex` binary against the
# operator's real home.
# ------------------------------------------------------------------
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

run_codex() {
    local home="$1"
    shift
    case "$home" in
        "$TMP_DIR"/*) ;;
        *) fail "refusing to run codex with CODEX_HOME outside the test tmpdir: $home" ;;
    esac
    CODEX_HOME="$home" "$@"
}

if ! command -v codex >/dev/null 2>&1; then
    skip_all "codex CLI not found on PATH"
fi

CODEX_VERSION_STRING=$(codex --version 2>&1 | head -1 || true)
echo "Detected: $CODEX_VERSION_STRING"

DETECT_HOME="$TMP_DIR/detect-home"
mkdir -p "$DETECT_HOME"
for probe in "plugin marketplace add --help" "plugin add --help" "plugin list --help" \
             "plugin remove --help" "mcp add --help" "mcp list --help" "mcp remove --help"; do
    # shellcheck disable=SC2086  # intentional word-splitting of the probe subcommand
    if ! run_codex "$DETECT_HOME" codex $probe >/dev/null 2>&1; then
        skip_all "installed codex CLI ($CODEX_VERSION_STRING) does not support 'codex $probe' — plugin-manager/minimal-MCP path requires Codex CLI 0.145+"
    fi
done
echo "Codex CLI supports the full plugin/mcp command surface this test needs."
echo ""

send_jsonrpc() {
    local binary="$1"
    local request="$2"
    local home_env="$3"
    local raw_output
    raw_output=$(HOME="$TMP_DIR/fake-home" META_CC_CODEX_ROOT="$home_env" \
        REQUEST_PAYLOAD="$request" python3 - 8 "$binary" <<'PY' 2>&1 || true
import os
import subprocess
import sys

seconds = int(sys.argv[1])
binary = sys.argv[2]
payload = os.environ["REQUEST_PAYLOAD"] + "\n"

try:
    proc = subprocess.run(
        [binary],
        input=payload,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        timeout=seconds,
        check=False,
    )
    print(proc.stdout, end="")
except subprocess.TimeoutExpired as exc:
    if exc.stdout:
        print(exc.stdout, end="")
    print("__META_CC_TIMEOUT__")
PY
)
    echo "$raw_output" | grep -E '^\s*\{' | grep '"jsonrpc"' | head -1 || true
}

# ==================================================================
# Scenario A: preferred plugin-manager path
# ==================================================================
echo -e "${BLUE}Scenario A: preferred install via codex plugin marketplace/add${NC}"

CODEX_HOME_A="$TMP_DIR/codex-home-a"
mkdir -p "$CODEX_HOME_A"

EXPECTED_VERSION=$(jq -r '.version' "$BUNDLE_DIR/.codex-plugin/plugin.json")
CLAUDE_MANIFEST_VERSION=$(jq -r '.version' "$BUNDLE_DIR/.claude-plugin/plugin.json")
MARKETPLACE_VERSION=$(jq -r '.plugins[0].version' "$BUNDLE_DIR/.claude-plugin/marketplace.json")
if [ "$EXPECTED_VERSION" != "$CLAUDE_MANIFEST_VERSION" ] || [ "$EXPECTED_VERSION" != "$MARKETPLACE_VERSION" ]; then
    fail "bundle manifests disagree on version: .codex-plugin/plugin.json=$EXPECTED_VERSION .claude-plugin/plugin.json=$CLAUDE_MANIFEST_VERSION marketplace.json=$MARKETPLACE_VERSION"
fi

ADD_MKT_OUT=$(run_codex "$CODEX_HOME_A" codex plugin marketplace add "$BUNDLE_DIR" --json 2>/dev/null) \
    || fail "codex plugin marketplace add failed"
echo "$ADD_MKT_OUT" | jq -e '.marketplaceName == "meta-cc-marketplace"' >/dev/null \
    || fail "unexpected marketplace add output: $ADD_MKT_OUT"
pass "codex plugin marketplace add registered meta-cc-marketplace from the local bundle"

ADD_PLUGIN_OUT=$(run_codex "$CODEX_HOME_A" codex plugin add meta-cc@meta-cc-marketplace --json 2>/dev/null) \
    || fail "codex plugin add failed"
INSTALLED_VERSION=$(echo "$ADD_PLUGIN_OUT" | jq -r '.version')
INSTALLED_PATH=$(echo "$ADD_PLUGIN_OUT" | jq -r '.installedPath')
[ -n "$INSTALLED_PATH" ] && [ -d "$INSTALLED_PATH" ] || fail "codex plugin add did not report a valid installedPath: $ADD_PLUGIN_OUT"
if [ "$INSTALLED_VERSION" != "$EXPECTED_VERSION" ]; then
    fail "installed plugin version ($INSTALLED_VERSION) does not match bundle manifest version ($EXPECTED_VERSION)"
fi
pass "codex plugin add installed meta-cc@meta-cc-marketplace version $INSTALLED_VERSION"

LIST_OUT=$(run_codex "$CODEX_HOME_A" codex plugin list --json 2>/dev/null) || fail "codex plugin list failed"
echo "$LIST_OUT" | jq -e '.installed[] | select(.pluginId == "meta-cc@meta-cc-marketplace" and .installed == true and .enabled == true)' >/dev/null \
    || fail "meta-cc@meta-cc-marketplace is not reported as installed+enabled: $LIST_OUT"
pass "plugin is installed and enabled"

for skill in prompt-find prompt-list prompt-show; do
    require_file "$INSTALLED_PATH/skills/$skill/SKILL.md"
done
pass "installed plugin exposes all 3 skills (SKILL.md discoverable under the installed path)"

MCP_LIST_OUT=$(run_codex "$CODEX_HOME_A" codex mcp list --json 2>/dev/null) || fail "codex mcp list failed"
MCP_ENTRY_COUNT=$(echo "$MCP_LIST_OUT" | jq '[.[] | select(.name == "meta-cc")] | length')
if [ "$MCP_ENTRY_COUNT" != "1" ]; then
    fail "expected exactly one active meta-cc MCP registration after plugin install, found $MCP_ENTRY_COUNT: $MCP_LIST_OUT"
fi
pass "exactly one active meta-cc MCP registration (no duplicates) after plugin install"

MCP_COMMAND=$(echo "$MCP_LIST_OUT" | jq -r '.[] | select(.name == "meta-cc") | .transport.command')
RELATIVE_COMMAND="${MCP_COMMAND#./}"
RESOLVED_BIN="$INSTALLED_PATH/$RELATIVE_COMMAND"
[ -x "$RESOLVED_BIN" ] || fail "MCP command resolved from the installed plugin is not an executable file: $RESOLVED_BIN"
pass "MCP command '$MCP_COMMAND' resolves to an executable artifact under the installed plugin path"

# Live tools/list against the ACTUAL installed artifact (not the source binary).
TOOLS_RESPONSE=$(send_jsonrpc "$RESOLVED_BIN" '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' "$CODEX_HOME_A")
[ -n "$TOOLS_RESPONSE" ] || fail "no JSON-RPC response from installed artifact for tools/list"
TOOL_COUNT=$(echo "$TOOLS_RESPONSE" | jq -e '.result.tools | length')
[ "$TOOL_COUNT" -gt 0 ] || fail "installed artifact returned zero tools for tools/list"
pass "live tools/list against the installed artifact returned $TOOL_COUNT tools"

# Live Codex-flavored query against the installed artifact: seed a synthetic
# Codex rollout + thread index under CODEX_HOME_A and confirm meta-cc-mcp
# (as installed by the plugin manager) can read it back.
UNIQUE_MESSAGE="codex-plugin-e2e-message-$RANDOM-$(date +%s)"
SESSION_ID="codex-plugin-e2e-session"
ROLLOUT_DIR="$CODEX_HOME_A/rollouts/2026/07/26"
ROLLOUT_FILE="$ROLLOUT_DIR/$SESSION_ID.jsonl"
mkdir -p "$ROLLOUT_DIR"
cat > "$ROLLOUT_FILE" <<EOF
{"timestamp":"2026-07-26T06:00:00Z","type":"session_meta","payload":{"id":"$SESSION_ID","cwd":"$PROJECT_DIR","model":"gpt-5"}}
{"timestamp":"2026-07-26T06:00:01Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"$UNIQUE_MESSAGE"}]}}
EOF

CODEX_HOME="$CODEX_HOME_A" SESSION_ID="$SESSION_ID" ROLLOUT_FILE="$ROLLOUT_FILE" PROJECT_DIR="$PROJECT_DIR" python3 - <<'PY'
import os
import sqlite3

db_path = os.path.join(os.environ["CODEX_HOME"], "state_5.sqlite")
conn = sqlite3.connect(db_path)
try:
    conn.execute(
        """
        CREATE TABLE threads (
            id TEXT PRIMARY KEY,
            rollout_path TEXT,
            cwd TEXT,
            title TEXT,
            model TEXT,
            model_provider TEXT,
            tokens_used INTEGER,
            source TEXT,
            created_at INTEGER
        )
        """
    )
    conn.execute(
        """
        INSERT INTO threads(id, rollout_path, cwd, title, model, model_provider, tokens_used, source, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        """,
        (
            os.environ["SESSION_ID"],
            os.environ["ROLLOUT_FILE"],
            os.environ["PROJECT_DIR"],
            "codex plugin-manager e2e",
            "gpt-5",
            "openai",
            0,
            "cli",
            1785045600,
        ),
    )
    conn.commit()
finally:
    conn.close()
PY

QUERY_REQUEST=$(jq -nc --arg cwd "$PROJECT_DIR" --arg pattern "$UNIQUE_MESSAGE" \
    '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"query_session_content","arguments":{"role":"user","provider":"codex","scope":"project","working_dir":$cwd,"pattern":$pattern,"limit":5}}}')
QUERY_RESPONSE=$(send_jsonrpc "$RESOLVED_BIN" "$QUERY_REQUEST" "$CODEX_HOME_A")
[ -n "$QUERY_RESPONSE" ] || fail "no JSON-RPC response from installed artifact for query_session_content"
echo "$QUERY_RESPONSE" | jq -e '.result.content[0].text | contains("'"$UNIQUE_MESSAGE"'")' >/dev/null \
    || fail "installed artifact did not return the synthetic Codex rollout message via query_session_content"
pass "live Codex-domain query (query_session_content, provider=codex) succeeded against the installed artifact"

REMOVE_OUT=$(run_codex "$CODEX_HOME_A" codex plugin remove meta-cc@meta-cc-marketplace 2>&1) \
    || fail "codex plugin remove failed: $REMOVE_OUT"
AFTER_REMOVE=$(run_codex "$CODEX_HOME_A" codex plugin list --json 2>/dev/null)
echo "$AFTER_REMOVE" | jq -e '.installed | length == 0' >/dev/null \
    || fail "plugin still reported installed after codex plugin remove: $AFTER_REMOVE"
pass "codex plugin remove uninstalled meta-cc, confined to the isolated CODEX_HOME"
echo ""

# ==================================================================
# Scenario B: minimal MCP-only fallback (codex mcp add)
# ==================================================================
echo -e "${BLUE}Scenario B: minimal fallback via codex mcp add${NC}"

CODEX_HOME_B="$TMP_DIR/codex-home-b"
mkdir -p "$CODEX_HOME_B"

run_codex "$CODEX_HOME_B" codex mcp add meta-cc -- "$MCP_BINARY" >/dev/null 2>&1 \
    || fail "codex mcp add failed"

MCP_LIST_B=$(run_codex "$CODEX_HOME_B" codex mcp list --json 2>/dev/null) || fail "codex mcp list failed (scenario B)"
MCP_ENTRY_COUNT_B=$(echo "$MCP_LIST_B" | jq '[.[] | select(.name == "meta-cc")] | length')
[ "$MCP_ENTRY_COUNT_B" = "1" ] || fail "expected exactly one meta-cc MCP registration after 'codex mcp add', found $MCP_ENTRY_COUNT_B"
MCP_COMMAND_B=$(echo "$MCP_LIST_B" | jq -r '.[] | select(.name == "meta-cc") | .transport.command')
[ "$MCP_COMMAND_B" = "$MCP_BINARY" ] || fail "codex mcp list reported unexpected command: $MCP_COMMAND_B (expected $MCP_BINARY)"
pass "codex mcp add registered exactly one meta-cc MCP server pointing at the built binary"

TOOLS_RESPONSE_B=$(send_jsonrpc "$MCP_BINARY" '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' "$CODEX_HOME_B")
[ -n "$TOOLS_RESPONSE_B" ] || fail "no JSON-RPC response from minimal-install binary for tools/list"
TOOL_COUNT_B=$(echo "$TOOLS_RESPONSE_B" | jq -e '.result.tools | length')
[ "$TOOL_COUNT_B" -gt 0 ] || fail "minimal-install binary returned zero tools for tools/list"
pass "live tools/list against the minimal codex-mcp-add install returned $TOOL_COUNT_B tools"

run_codex "$CODEX_HOME_B" codex mcp remove meta-cc >/dev/null 2>&1 || fail "codex mcp remove failed"
MCP_LIST_B_AFTER=$(run_codex "$CODEX_HOME_B" codex mcp list --json 2>/dev/null)
[ "$(echo "$MCP_LIST_B_AFTER" | jq 'length')" = "0" ] || fail "meta-cc MCP entry still present after codex mcp remove"
pass "codex mcp remove cleaned up the minimal install, confined to the isolated CODEX_HOME"
echo ""

echo "=========================================="
echo "Codex Plugin-Manager E2E Test Complete"
echo "=========================================="
