#!/usr/bin/env bash
# check-worktree-cache-env.sh — verify shared Go build/test caches across worktrees (DIR-091).
#
# Reports the resolved GOCACHE / GOMODCACHE values. PASS when both resolve to
# shared, non-worktree-local paths under $HOME. FAIL fail-closed on per-worktree
# isolation (cache inside a /tmp/meta-cc-worktrees path), an unset HOME, or
# explicit GOFLAGS/GOCACHE/GOMODCACHE overrides pointing into a worktree. WARN
# (non-fatal) when the shared cache directory is empty or cold.
#
# The workspace root is resolved from this script's own location
# (scripts/checks/ -> ../..), so the screen works regardless of the caller's
# cwd and in scratch copies used by the standalone test harness.
#
# Usage:
#   check-worktree-cache-env.sh        (no arguments)
#
# Exit codes: 0 = PASS (caches shared under $HOME); 1 = FAIL (per-worktree
# isolation, unset HOME, or worktree-directed override); 2 = environment error
# (no go toolchain).

set -u

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

if ! command -v go >/dev/null 2>&1; then
  echo "FAIL check-worktree-cache-env: go toolchain not found in PATH" >&2
  exit 2
fi

# ── Resolve Go cache environment ──────────────────────────────────────────────
GOCACHE="$(go env GOCACHE 2>/dev/null || true)"
GOMODCACHE="$(go env GOMODCACHE 2>/dev/null || true)"
GOPATH_VAL="$(go env GOPATH 2>/dev/null || true)"
GOFLAGS_VAL="$(go env GOFLAGS 2>/dev/null || true)"
HOME_VAL="${HOME:-}"

if [ -z "$HOME_VAL" ]; then
  echo "FAIL check-worktree-cache-env: HOME is unset or empty — cannot verify shared cache roots" >&2
  exit 1
fi

# ── Report resolved values ───────────────────────────────────────────────────
echo "check-worktree-cache-env: GOCACHE=$GOCACHE"
echo "check-worktree-cache-env: GOMODCACHE=$GOMODCACHE"
echo "check-worktree-cache-env: GOPATH=$GOPATH_VAL"
echo "check-worktree-cache-env: HOME=$HOME_VAL"
if [ -n "$GOFLAGS_VAL" ]; then
  echo "check-worktree-cache-env: GOFLAGS=$GOFLAGS_VAL"
fi

# ── Check GOCACHE is under $HOME ──────────────────────────────────────────────
if [ -z "$GOCACHE" ]; then
  echo "FAIL check-worktree-cache-env: GOCACHE is empty — go env returned no value" >&2
  exit 1
fi

# Normalise to absolute path (resolve any symlinks).
GOCACHE_ABS="$(cd "$GOCACHE" 2>/dev/null && pwd || echo "$GOCACHE")"

if [[ "$GOCACHE_ABS" == /tmp/* ]]; then
  echo "FAIL check-worktree-cache-env: GOCACHE resolves to a /tmp path ($GOCACHE_ABS) — per-worktree isolation detected" >&2
  exit 1
fi

if [[ "$GOCACHE_ABS" != "$HOME_VAL"/* ]]; then
  echo "FAIL check-worktree-cache-env: GOCACHE ($GOCACHE_ABS) is not under HOME ($HOME_VAL) — cache may be isolated or ephemeral" >&2
  exit 1
fi

# ── Check GOMODCACHE is under $HOME ───────────────────────────────────────────
if [ -z "$GOMODCACHE" ]; then
  echo "FAIL check-worktree-cache-env: GOMODCACHE is empty — go env returned no value" >&2
  exit 1
fi

GOMODCACHE_ABS="$(cd "$GOMODCACHE" 2>/dev/null && pwd || echo "$GOMODCACHE")"

if [[ "$GOMODCACHE_ABS" == /tmp/* ]]; then
  echo "FAIL check-worktree-cache-env: GOMODCACHE resolves to a /tmp path ($GOMODCACHE_ABS) — per-worktree isolation detected" >&2
  exit 1
fi

if [[ "$GOMODCACHE_ABS" != "$HOME_VAL"/* ]]; then
  echo "FAIL check-worktree-cache-env: GOMODCACHE ($GOMODCACHE_ABS) is not under HOME ($HOME_VAL) — cache may be isolated or ephemeral" >&2
  exit 1
fi

# ── Check GOFLAGS for per-worktree overrides ─────────────────────────────────
if [ -n "$GOFLAGS_VAL" ]; then
  if [[ "$GOFLAGS_VAL" == *"/tmp/meta-cc-worktrees"* ]]; then
    echo "FAIL check-worktree-cache-env: GOFLAGS contains a worktree path ($GOFLAGS_VAL) — cache overrides may isolate worktrees" >&2
    exit 1
  fi
fi

# ── WARN if cache directories are empty or cold (non-fatal) ───────────────────
WARNINGS=0

# Check GOCACHE warmth: report file count and total size
if [ -d "$GOCACHE_ABS" ]; then
  CACHE_FILE_COUNT="$(find "$GOCACHE_ABS" -type f 2>/dev/null | wc -l)"
  CACHE_SIZE="$(du -sh "$GOCACHE_ABS" 2>/dev/null | cut -f1 || echo "0")"
  echo "check-worktree-cache-env: GOCACHE warmth — ${CACHE_FILE_COUNT} files, ${CACHE_SIZE}"

  if [ "$CACHE_FILE_COUNT" -eq 0 ]; then
    echo "WARN check-worktree-cache-env: GOCACHE ($GOCACHE_ABS) is empty — first build in this environment will be cold" >&2
    WARNINGS=$((WARNINGS + 1))
  fi
else
  echo "WARN check-worktree-cache-env: GOCACHE directory ($GOCACHE_ABS) does not exist yet — first build will create it" >&2
  WARNINGS=$((WARNINGS + 1))
fi

# Check GOMODCACHE warmth
if [ -d "$GOMODCACHE_ABS" ]; then
  MOD_FILE_COUNT="$(find "$GOMODCACHE_ABS" -type f 2>/dev/null | wc -l)"
  echo "check-worktree-cache-env: GOMODCACHE warmth — ${MOD_FILE_COUNT} files"
else
  echo "WARN check-worktree-cache-env: GOMODCACHE directory ($GOMODCACHE_ABS) does not exist yet" >&2
  WARNINGS=$((WARNINGS + 1))
fi

# ── Result ────────────────────────────────────────────────────────────────────
if [ "$WARNINGS" -gt 0 ]; then
  echo "PASS check-worktree-cache-env: caches are shared under HOME (GOCACHE=$GOCACHE_ABS, GOMODCACHE=$GOMODCACHE_ABS) with $WARNINGS warmth warning(s)"
else
  echo "PASS check-worktree-cache-env: caches are shared under HOME (GOCACHE=$GOCACHE_ABS, GOMODCACHE=$GOMODCACHE_ABS) and warm"
fi
exit 0
