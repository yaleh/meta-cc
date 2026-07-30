#!/usr/bin/env bash
# build-feasibility-check.sh — build-feasibility early screen (DIR-015). Fails closed
# when the workspace cannot build, so an unbuildable task is stopped at the gate
# boundary instead of after a full subagent dispatch (~396K tokens wasted across three
# it0-checks-failed runs in the original finding).
#
# Registered in .quay/config.yml as the fixed gate `build-feasibility` (zero per-task
# args; `quay gate <task-id> --gate build-feasibility`), and prepended to the workspace
# DoD chain in it0-dod-check.sh before its TS-enforcer delegation.
#
# NOTE (quay-side shadowing, DIR-015 investigation): quay's `lifecycle_promote`
# (todo->ready) resolves gate "dod" to quay's BUILT-IN dod gate (registry.ts checks
# gateRegistry before workspace gates — built-ins always win a name collision), which
# delegates to the Provider's taskCheck. The workspace it0 gate ALSO named "dod"
# (it0-dod-check.sh, argsKey dodArgs) is shadowed and never runs during promotion;
# dodArgs is neither required nor consulted on that path. This screen is therefore
# registered under the non-shadowed name `build-feasibility`; wiring it INTO promotion
# itself requires a quay change (out of meta-cc scope, reported only).
#
# The workspace root is resolved from this script's own location (scripts/gates/ ->
# ../..), so the screen works regardless of the caller's cwd and in scratch copies
# used by the standalone test harness.
#
# Usage:
#   build-feasibility-check.sh        (no arguments)
#
# Exit codes: 0 = PASS (workspace builds); 1 = FAIL (build broken or not a Go
# workspace); 2 = environment error (no go toolchain). All non-zero exits are
# fail-closed through quay's fixed-script gate (exit 0 = ok, non-zero = fail).

set -u

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

if ! command -v go >/dev/null 2>&1; then
  echo "FAIL build-feasibility: go toolchain not found in PATH" >&2
  exit 2
fi

if [ ! -f "$WORKSPACE_ROOT/go.mod" ]; then
  echo "FAIL build-feasibility: no go.mod at workspace root ($WORKSPACE_ROOT) — not a Go workspace" >&2
  exit 1
fi

build_output="$(cd "$WORKSPACE_ROOT" && go build ./... 2>&1)"
build_rc=$?
if [ $build_rc -ne 0 ]; then
  echo "FAIL build-feasibility: 'go build ./...' failed in $WORKSPACE_ROOT (exit $build_rc):" >&2
  echo "$build_output" | tail -n 30 >&2
  exit 1
fi

echo "PASS build-feasibility: workspace builds clean (go build ./... in $WORKSPACE_ROOT)"
exit 0
