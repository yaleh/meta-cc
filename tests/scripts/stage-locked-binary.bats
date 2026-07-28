#!/usr/bin/env bats
#
# Regression test for DIR-037: `make stage` must succeed even when the
# current plugin-src/bin/meta-cc-mcp is held open (mapped for execution) by
# a running process, instead of failing with ETXTBSY ("Text file busy").
#
# Safety note: this test spawns and kills only its own throwaway process (a
# trivial Go "sleeper" binary built on the fly, held open by this test in
# the background) -- it never touches any pre-existing system process.
#
# Run with: bats tests/scripts/stage-locked-binary.bats

WORKTREE_ROOT="$(cd "$(dirname "$BATS_TEST_FILENAME")/../.." && pwd)"

setup() {
    HELD_PID=""
    SLEEPER_SRC="$BATS_TEST_TMPDIR/sleeper/main.go"
    mkdir -p "$(dirname "$SLEEPER_SRC")"
    cat > "$SLEEPER_SRC" <<'EOF'
package main

import "time"

func main() {
	time.Sleep(1 * time.Hour)
}
EOF
}

teardown() {
    if [ -n "$HELD_PID" ] && kill -0 "$HELD_PID" 2>/dev/null; then
        kill "$HELD_PID" 2>/dev/null || true
        wait "$HELD_PID" 2>/dev/null || true
    fi
    rm -rf "$WORKTREE_ROOT/plugin-src/bin" "$WORKTREE_ROOT/bin"
}

@test "make stage succeeds when plugin-src/bin/meta-cc-mcp is held open by a running process" {
    mkdir -p "$WORKTREE_ROOT/plugin-src/bin"

    # Stage a throwaway long-running binary at the target path to simulate a
    # live MCP server process holding the previous binary open. This
    # process is spawned and owned entirely by this test.
    run go build -o "$WORKTREE_ROOT/plugin-src/bin/meta-cc-mcp" "$BATS_TEST_TMPDIR/sleeper/main.go"
    [ "$status" -eq 0 ]

    OLD_INODE=$(stat -c %i "$WORKTREE_ROOT/plugin-src/bin/meta-cc-mcp")

    ( exec "$WORKTREE_ROOT/plugin-src/bin/meta-cc-mcp" ) &
    HELD_PID=$!
    sleep 0.3
    kill -0 "$HELD_PID"  # sanity: our throwaway process is alive and holding the binary open

    run make -C "$WORKTREE_ROOT" stage
    [ "$status" -eq 0 ]
    [[ "$output" == *"Staged plugin-src/bin/meta-cc-mcp"* ]]

    # The throwaway process must still be alive and running the OLD (now
    # unlinked) inode -- proving `stage` used write-temp+rename semantics,
    # not an in-place truncating overwrite.
    kill -0 "$HELD_PID"
    EXE_LINK=$(readlink "/proc/$HELD_PID/exe")
    [[ "$EXE_LINK" == *"(deleted)"* ]]

    NEW_INODE=$(stat -c %i "$WORKTREE_ROOT/plugin-src/bin/meta-cc-mcp")
    [ "$OLD_INODE" != "$NEW_INODE" ]

    # A fresh spawn of the staged path now reflects the freshly built real
    # MCP server (not the throwaway sleeper): it must answer a JSON-RPC
    # "initialize" call.
    RESPONSE=$(printf '%s\n' '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' \
        | timeout 5 "$WORKTREE_ROOT/plugin-src/bin/meta-cc-mcp" 2>&1 | grep '"jsonrpc"' | head -1 || true)
    [[ "$RESPONSE" == *'"jsonrpc":"2.0"'* ]]
    [[ "$RESPONSE" == *'"serverInfo"'* ]]
}
