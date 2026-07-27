#!/bin/bash
# Codex-focused E2E test for meta-cc.
# Verifies the Codex package/install surface and the MCP server's Codex
# transcript discovery path against a real JSON-RPC tool call.
#
# Usage: ./tests/e2e/codex-e2e.sh [binary_path]

set -euo pipefail

BINARY="${1:-./bin/meta-cc-mcp}"
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

fail() {
    echo -e "${RED}FAILED:${NC} $1" >&2
    exit 1
}

pass() {
    echo -e "  ${GREEN}PASS${NC} - $1"
}

require_file() {
    [ -f "$1" ] || fail "missing file: $1"
}

require_dir() {
    [ -d "$1" ] || fail "missing directory: $1"
}

get_json_response() {
    local input="$1"
    echo "$input" | grep -E '^\s*\{' | grep '"jsonrpc"' | head -1 || true
}

send_request() {
    local request="$1"
    local raw_output
    if command -v timeout >/dev/null 2>&1; then
        raw_output=$(printf '%s\n' "$request" | timeout 8 "$BINARY" 2>&1 || true)
    elif command -v gtimeout >/dev/null 2>&1; then
        raw_output=$(printf '%s\n' "$request" | gtimeout 8 "$BINARY" 2>&1 || true)
    elif command -v python3 >/dev/null 2>&1; then
        raw_output=$(REQUEST_PAYLOAD="$request" python3 - 8 "$BINARY" <<'PY' 2>&1 || true
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
    else
        raw_output=$(printf '%s\n' "$request" | "$BINARY" 2>&1 || true)
    fi
    local response
    response=$(get_json_response "$raw_output")
    if [ -z "$response" ]; then
        # shellcheck disable=SC2001  # ${var//...} cannot prefix every line; sed is required here
        echo "$raw_output" | sed 's/^/RAW: /' >&2
    fi
    echo "$response"
}

if [ ! -f "$BINARY" ]; then
    fail "binary not found: $BINARY"
fi

if ! command -v jq >/dev/null 2>&1; then
    fail "jq is required"
fi

if ! python3 - <<'PY' >/dev/null 2>&1
import sqlite3
PY
then
    fail "python3 with sqlite3 module is required"
fi

echo "=========================================="
echo "Codex E2E Test"
echo "=========================================="
echo "Binary:      $BINARY"
echo "Project dir: $PROJECT_DIR"
echo "=========================================="
echo ""

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

CODEX_HOME="$TMP_DIR/codex-home"
CLAUDE_DIR="$TMP_DIR/claude-home"
INSTALL_DIR="$TMP_DIR/bin"
PLUGIN_DATA_DIR="$TMP_DIR/plugin-data"
SESSION_ID="codex-e2e-session"
UNIQUE_MESSAGE="codex-e2e-message-$RANDOM-$(date +%s)"
UNIQUE_ERROR="codex-e2e-error-$RANDOM-$(date +%s)"
ROLLOUT_DIR="$CODEX_HOME/rollouts/2026/06/14"
ROLLOUT_FILE="$ROLLOUT_DIR/$SESSION_ID.jsonl"

echo -e "${BLUE}Test 1: Install Codex skills package into temp CODEX_HOME${NC}"
rm -rf "$TMP_DIR/skills-package"
bash "$PROJECT_DIR/scripts/release/build-skills-package.sh" \
    --version v0.0.0-codex-e2e \
    --output "$TMP_DIR/skills-package" >/dev/null

tar -xzf "$TMP_DIR/skills-package/meta-cc-skills-v0.0.0-codex-e2e.tar.gz" -C "$TMP_DIR"
INSTALL_CODEX=1 INSTALL_CLAUDE=0 CODEX_HOME="$CODEX_HOME" CLAUDE_DIR="$CLAUDE_DIR" \
    bash "$TMP_DIR/meta-cc-skills-v0.0.0-codex-e2e/install-skills.sh" >/dev/null

for skill in prompt-find prompt-list prompt-show; do
    require_file "$CODEX_HOME/skills/$skill/SKILL.md"
done
if find "$CLAUDE_DIR" -type f 2>/dev/null | grep -q .; then
    fail "INSTALL_CLAUDE=0 wrote files under CLAUDE_DIR"
fi
pass "Codex skills installed under temp CODEX_HOME without Claude writes"
echo ""

echo -e "${BLUE}Test 2: Install full archive Codex plugin files into temp CODEX_HOME${NC}"
FULL_PKG="$TMP_DIR/full-package"
mkdir -p "$FULL_PKG/bin" "$FULL_PKG/commands" "$FULL_PKG/skills" "$FULL_PKG/lib" \
    "$FULL_PKG/.claude-plugin" "$FULL_PKG/.codex-plugin"
cp "$BINARY" "$FULL_PKG/bin/meta-cc-mcp"
cp -r "$PROJECT_DIR/plugin-src/commands/." "$FULL_PKG/commands/"
cp -r "$PROJECT_DIR/plugin-src/skills/." "$FULL_PKG/skills/"
cat > "$FULL_PKG/lib/mcp-config.json" <<'EOF'
{
  "mcpServers": {
    "meta-cc": {
      "command": "meta-cc-mcp",
      "args": []
    }
  }
}
EOF
cp "$PROJECT_DIR/plugin-src/.claude-plugin/plugin.json" "$FULL_PKG/.claude-plugin/plugin.json"
cp "$PROJECT_DIR/.claude-plugin/marketplace.json" "$FULL_PKG/.claude-plugin/marketplace.json"
cp "$PROJECT_DIR/plugin-src/.mcp.json" "$FULL_PKG/.mcp.json"
cp "$PROJECT_DIR/plugin-src/.codex-plugin/plugin.json" "$FULL_PKG/.codex-plugin/plugin.json"
cp "$PROJECT_DIR/plugin-src/.codex-mcp.json" "$FULL_PKG/.codex-mcp.json"
cp "$PROJECT_DIR/scripts/install/install.sh" "$FULL_PKG/install.sh"

(
    cd "$FULL_PKG"
    INSTALL_DIR="$INSTALL_DIR" CLAUDE_DIR="$CLAUDE_DIR" CODEX_HOME="$CODEX_HOME" \
        PLUGIN_DATA_DIR="$PLUGIN_DATA_DIR" \
        bash ./install.sh >/dev/null
)

require_file "$INSTALL_DIR/meta-cc-mcp"
require_file "$CODEX_HOME/plugins/meta-cc/.codex-plugin/plugin.json"
require_file "$CODEX_HOME/plugins/meta-cc/.codex-mcp.json"
jq -e '.skills == "./skills/" and .mcpServers == "./.codex-mcp.json"' \
    "$CODEX_HOME/plugins/meta-cc/.codex-plugin/plugin.json" >/dev/null
jq -e '.mcpServers["meta-cc"].command == "./bin/meta-cc-mcp"' \
    "$CODEX_HOME/plugins/meta-cc/.codex-mcp.json" >/dev/null
pass "Full archive installs Codex plugin manifest and MCP template"
echo ""

echo -e "${BLUE}Test 3: Query Codex transcript through real MCP JSON-RPC${NC}"
mkdir -p "$ROLLOUT_DIR"
cat > "$ROLLOUT_FILE" <<EOF
{"timestamp":"2026-06-14T06:00:00Z","type":"session_meta","payload":{"id":"$SESSION_ID","cwd":"$PROJECT_DIR","model":"gpt-5"}}
{"timestamp":"2026-06-14T06:00:01Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"$UNIQUE_MESSAGE"}]}}
{"timestamp":"2026-06-14T06:00:02Z","type":"response_item","payload":{"type":"message","role":"assistant","content":[{"type":"output_text","text":"ack"}]}}
{"timestamp":"2026-06-14T06:00:03Z","type":"response_item","payload":{"type":"function_call","name":"exec_command","call_id":"call_codex_e2e_1","arguments":"{\"cmd\":\"go test ./internal/session\",\"workdir\":\"$PROJECT_DIR\"}"}}
{"timestamp":"2026-06-14T06:00:04Z","type":"response_item","payload":{"type":"function_call_output","call_id":"call_codex_e2e_1","output":"ok"}}
{"timestamp":"2026-06-14T06:00:05Z","type":"response_item","payload":{"type":"custom_tool_call","name":"apply_patch","call_id":"call_codex_e2e_2","input":"*** Begin Patch\n*** End Patch"}}
{"timestamp":"2026-06-14T06:00:06Z","type":"response_item","payload":{"type":"custom_tool_call_output","call_id":"call_codex_e2e_2","status":"failed","output":"$UNIQUE_ERROR"}}
{"timestamp":"2026-06-14T06:00:07Z","type":"event_msg","payload":{"type":"token_count","info":{"last_token_usage":{"input_tokens":10,"cached_input_tokens":2,"output_tokens":3,"reasoning_output_tokens":1,"total_tokens":13},"total_token_usage":{"input_tokens":100,"cached_input_tokens":20,"output_tokens":30,"reasoning_output_tokens":10,"total_tokens":130},"model_context_window":200000},"rate_limits":{"limit_id":"codex-e2e-limit"}}}
EOF

CODEX_HOME="$CODEX_HOME" SESSION_ID="$SESSION_ID" ROLLOUT_FILE="$ROLLOUT_FILE" PROJECT_DIR="$PROJECT_DIR" python3 - <<'PY'
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
            "codex e2e",
            "gpt-5",
            "openai",
            130,
            "cli",
            1781416800,
        ),
    )
    conn.commit()
finally:
    conn.close()
PY

REQUEST=$(jq -nc --arg cwd "$PROJECT_DIR" --arg pattern "$UNIQUE_MESSAGE" \
    '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"query_session_content","arguments":{"role":"user","provider":"codex","scope":"project","working_dir":$cwd,"pattern":$pattern,"limit":5}}}')
# shellcheck disable=SC2097,SC2098,SC1007  # env-prefix on shell function; shellcheck false positive
RESPONSE=$(HOME="$TMP_DIR/home" CODEX_HOME="$CODEX_HOME" META_CC_CODEX_ROOT="$CODEX_HOME" META_CC_PROJECTS_ROOT= \
    send_request "$REQUEST")
[ -n "$RESPONSE" ] || fail "no JSON-RPC response for provider-backed query_session_content(role=user)"
echo "$RESPONSE" | jq -e '.result.content[0].text | contains("'"$UNIQUE_MESSAGE"'")' >/dev/null \
    || fail "provider-backed query_session_content(role=user) did not return the Codex rollout message"
pass "provider-backed query_session_content(role=user) returned data from Codex SQLite + rollout"

REQUEST=$(jq -nc --arg cwd "$PROJECT_DIR" --arg pattern "$UNIQUE_MESSAGE" \
    '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"query_session_content","arguments":{"role":"user","provider":"codex","scope":"project","working_dir":$cwd,"pattern":$pattern,"limit":5}}}')
# shellcheck disable=SC2097,SC2098,SC1007  # env-prefix on shell function; shellcheck false positive
RESPONSE=$(HOME="$TMP_DIR/home" CODEX_HOME="$CODEX_HOME" META_CC_CODEX_ROOT="$CODEX_HOME" META_CC_PROJECTS_ROOT= \
    send_request "$REQUEST")
[ -n "$RESPONSE" ] || fail "no JSON-RPC response for query_session_content(role=user)"
echo "$RESPONSE" | jq -e '.result.content[0].text | contains("'"$UNIQUE_MESSAGE"'")' >/dev/null \
    || fail "query_session_content(role=user) did not return the Codex transcript message"
pass "query_session_content(role=user) returned data from the Codex transcript"

REQUEST=$(jq -nc --arg cwd "$PROJECT_DIR" \
    '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"query_session_signals","arguments":{"type":"tool_stats","provider":"codex","scope":"project","working_dir":$cwd,"tool":"exec_command","limit":5}}}')
# shellcheck disable=SC2097,SC2098,SC1007  # env-prefix on shell function; shellcheck false positive
RESPONSE=$(HOME="$TMP_DIR/home" CODEX_HOME="$CODEX_HOME" META_CC_CODEX_ROOT="$CODEX_HOME" META_CC_PROJECTS_ROOT= \
    send_request "$REQUEST")
[ -n "$RESPONSE" ] || fail "no JSON-RPC response for query_session_signals(type=tool_stats)"
echo "$RESPONSE" | jq -e '.result.content[0].text | contains("exec_command")' >/dev/null \
    || fail "query_session_signals(type=tool_stats) did not return Codex function_call tool data"
pass "query_session_signals(type=tool_stats) returned Codex function_call data"

REQUEST=$(jq -nc --arg cwd "$PROJECT_DIR" --arg error "$UNIQUE_ERROR" \
    '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"query_session_signals","arguments":{"type":"errors","provider":"codex","scope":"project","working_dir":$cwd,"limit":5}}}')
# shellcheck disable=SC2097,SC2098,SC1007  # env-prefix on shell function; shellcheck false positive
RESPONSE=$(HOME="$TMP_DIR/home" CODEX_HOME="$CODEX_HOME" META_CC_CODEX_ROOT="$CODEX_HOME" META_CC_PROJECTS_ROOT= \
    send_request "$REQUEST")
[ -n "$RESPONSE" ] || fail "no JSON-RPC response for query_session_signals(type=errors)"
echo "$RESPONSE" | jq -e '.result.content[0].text | contains("'"$UNIQUE_ERROR"'")' >/dev/null \
    || fail "query_session_signals(type=errors) did not return Codex failed tool output"
pass "query_session_signals(type=errors) returned Codex failed tool output"

REQUEST=$(jq -nc --arg cwd "$PROJECT_DIR" \
    '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"query_session_signals","arguments":{"type":"tokens","provider":"codex","scope":"project","working_dir":$cwd,"limit":5}}}')
# shellcheck disable=SC2097,SC2098,SC1007  # env-prefix on shell function; shellcheck false positive
RESPONSE=$(HOME="$TMP_DIR/home" CODEX_HOME="$CODEX_HOME" META_CC_CODEX_ROOT="$CODEX_HOME" META_CC_PROJECTS_ROOT= \
    send_request "$REQUEST")
[ -n "$RESPONSE" ] || fail "no JSON-RPC response for query_session_signals(type=tokens)"
echo "$RESPONSE" | jq -e '.result.content[0].text | fromjson | .data[] | select(.message.usage.input_tokens == 10 and .message.usage.output_tokens == 3)' >/dev/null \
    || fail "query_session_signals(type=tokens) did not return Codex token_count usage"
pass "query_session_signals(type=tokens) returned Codex token_count usage"

REQUEST=$(jq -nc --arg cwd "$PROJECT_DIR" \
    '{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"get_work_patterns","arguments":{"provider":"codex","scope":"project","working_dir":$cwd}}}')
# shellcheck disable=SC2097,SC2098,SC1007  # env-prefix on shell function; shellcheck false positive
RESPONSE=$(HOME="$TMP_DIR/home" CODEX_HOME="$CODEX_HOME" META_CC_CODEX_ROOT="$CODEX_HOME" META_CC_PROJECTS_ROOT= \
    send_request "$REQUEST")
[ -n "$RESPONSE" ] || fail "no JSON-RPC response for get_work_patterns"
echo "$RESPONSE" | jq -e '.result.content[0].text | fromjson | .tool_frequency[] | select(.tool_name == "exec_command" and .count == 1)' >/dev/null \
    || fail "get_work_patterns did not count Codex exec_command tool usage"
echo "$RESPONSE" | jq -e '.result.content[0].text | fromjson | .tool_frequency[] | select(.tool_name == "apply_patch" and .count == 1)' >/dev/null \
    || fail "get_work_patterns did not count Codex apply_patch tool usage"
pass "get_work_patterns counted Codex tool usage"
echo ""

echo -e "${BLUE}Test 7: get_session_directory(provider=codex) resolves only Codex rollout files (DIR-024)${NC}"
REQUEST=$(jq -nc --arg cwd "$PROJECT_DIR" \
    '{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"get_session_directory","arguments":{"scope":"project","provider":"codex","working_dir":$cwd}}}')
# shellcheck disable=SC2097,SC2098,SC1007  # env-prefix on shell function; shellcheck false positive
RESPONSE=$(HOME="$TMP_DIR/home" CODEX_HOME="$CODEX_HOME" META_CC_CODEX_ROOT="$CODEX_HOME" META_CC_PROJECTS_ROOT= \
    send_request "$REQUEST")
[ -n "$RESPONSE" ] || fail "no JSON-RPC response for get_session_directory(provider=codex)"
echo "$RESPONSE" | jq -e --arg rollout "$ROLLOUT_FILE" \
    '.result.content[0].text | fromjson | .provider == "codex" and (.files | index($rollout) != null)' >/dev/null \
    || fail "get_session_directory(provider=codex) did not return the Codex rollout file"
echo "$RESPONSE" | jq -e '.result.content[0].text | fromjson | .files | all(test("\\.claude/") | not)' >/dev/null \
    || fail "get_session_directory(provider=codex) leaked a Claude session path"
pass "get_session_directory(provider=codex) returned only the Codex rollout file"

REQUEST=$(jq -nc --arg cwd "$PROJECT_DIR" \
    '{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"get_session_directory","arguments":{"scope":"project","provider":"bogus","working_dir":$cwd}}}')
# shellcheck disable=SC2097,SC2098,SC1007  # env-prefix on shell function; shellcheck false positive
RESPONSE=$(HOME="$TMP_DIR/home" CODEX_HOME="$CODEX_HOME" META_CC_CODEX_ROOT="$CODEX_HOME" META_CC_PROJECTS_ROOT= \
    send_request "$REQUEST")
[ -n "$RESPONSE" ] || fail "no JSON-RPC response for get_session_directory(provider=bogus)"
echo "$RESPONSE" | jq -e '.error.message | test("invalid provider")' >/dev/null \
    || fail "get_session_directory(provider=bogus) did not fail closed with an actionable error"
pass "get_session_directory(provider=bogus) failed closed with an actionable error"
echo ""

echo -e "${BLUE}Test 8: get_session_metadata(provider=codex) returns Codex raw schema, not Claude's (DIR-024)${NC}"
REQUEST=$(jq -nc --arg cwd "$PROJECT_DIR" \
    '{"jsonrpc":"2.0","id":9,"method":"tools/call","params":{"name":"get_session_metadata","arguments":{"scope":"project","provider":"codex","working_dir":$cwd}}}')
# shellcheck disable=SC2097,SC2098,SC1007  # env-prefix on shell function; shellcheck false positive
RESPONSE=$(HOME="$TMP_DIR/home" CODEX_HOME="$CODEX_HOME" META_CC_CODEX_ROOT="$CODEX_HOME" META_CC_PROJECTS_ROOT= \
    send_request "$REQUEST")
[ -n "$RESPONSE" ] || fail "no JSON-RPC response for get_session_metadata(provider=codex)"
echo "$RESPONSE" | jq -e --arg rollout "$ROLLOUT_FILE" \
    '.result.content[0].text | fromjson | .provider == "codex" and (.files[0].path == $rollout)' >/dev/null \
    || fail "get_session_metadata(provider=codex) did not describe the Codex rollout file"
echo "$RESPONSE" | jq -e '.result.content[0].text | fromjson | .jsonl_schema | has("legacy_response_item_fields") and (has("user_message_fields") | not)' >/dev/null \
    || fail "get_session_metadata(provider=codex) did not use the Codex-specific raw schema"
pass "get_session_metadata(provider=codex) returned Codex-specific raw schema and file info"

REQUEST=$(jq -nc --arg cwd "$PROJECT_DIR" \
    '{"jsonrpc":"2.0","id":10,"method":"tools/call","params":{"name":"get_session_metadata","arguments":{"scope":"project","provider":"all","working_dir":$cwd}}}')
# shellcheck disable=SC2097,SC2098,SC1007  # env-prefix on shell function; shellcheck false positive
RESPONSE=$(HOME="$TMP_DIR/home" CODEX_HOME="$CODEX_HOME" META_CC_CODEX_ROOT="$CODEX_HOME" META_CC_PROJECTS_ROOT= \
    send_request "$REQUEST")
[ -n "$RESPONSE" ] || fail "no JSON-RPC response for get_session_metadata(provider=all)"
echo "$RESPONSE" | jq -e '.result.content[0].text | fromjson | .providers.codex.jsonl_schema | has("legacy_response_item_fields")' >/dev/null \
    || fail "get_session_metadata(provider=all) did not keep a separate Codex schema under providers.codex"
pass "get_session_metadata(provider=all) kept Claude and Codex results separate"

REQUEST=$(jq -nc --arg cwd "$PROJECT_DIR" \
    '{"jsonrpc":"2.0","id":11,"method":"tools/call","params":{"name":"get_session_metadata","arguments":{"scope":"project","provider":"bogus","working_dir":$cwd}}}')
# shellcheck disable=SC2097,SC2098,SC1007  # env-prefix on shell function; shellcheck false positive
RESPONSE=$(HOME="$TMP_DIR/home" CODEX_HOME="$CODEX_HOME" META_CC_CODEX_ROOT="$CODEX_HOME" META_CC_PROJECTS_ROOT= \
    send_request "$REQUEST")
[ -n "$RESPONSE" ] || fail "no JSON-RPC response for get_session_metadata(provider=bogus)"
echo "$RESPONSE" | jq -e '.error.message | test("invalid provider")' >/dev/null \
    || fail "get_session_metadata(provider=bogus) did not fail closed with an actionable error"
pass "get_session_metadata(provider=bogus) failed closed with an actionable error"
echo ""

echo "=========================================="
echo "Codex E2E Test Complete"
echo "=========================================="
