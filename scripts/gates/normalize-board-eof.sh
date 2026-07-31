#!/usr/bin/env bash
# normalize-board-eof.sh — Ensure board files (tasks/** and docs/architecture/adr/**)
# end with a trailing newline, so that pre-commit's end-of-file-fixer hook passes on
# the first try instead of mutating them and aborting the commit.
#
# Modes:
#   (default) Fix mode — appends \n to every board file whose last byte ≠ 0x0a.
#              Prints each file touched. Exit 0.
#   --check    Check mode — lists files that lack trailing newline, exits 1 if any.
#
# Usage:
#   scripts/gates/normalize-board-eof.sh          # fix
#   scripts/gates/normalize-board-eof.sh --check  # check only
set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$HERE/../.." && pwd)"
cd "$ROOT" || { echo "ERROR: cannot cd to repo root ($ROOT)" >&2; exit 1; }

MODE="${1:-fix}"  # "fix" (default) or "--check"

# Collect board files under the two globs.
# Use null-delimited find for safe handling of filenames with spaces.
board_files=()
while IFS= read -r -d '' f; do
  board_files+=("$f")
done < <(find tasks docs/architecture/adr -maxdepth 3 -type f -name '*.md' -print0 2>/dev/null || true)

if [ ${#board_files[@]} -eq 0 ]; then
  echo "normalize-board-eof: no board files found under tasks/ or docs/architecture/adr/."
  exit 0
fi

dirty=()
for f in "${board_files[@]}"; do
  # Check if the last byte is NOT 0x0a (newline)
  if [ -s "$f" ]; then
    last_byte=$(tail -c 1 "$f" | xxd -p)
    if [ "$last_byte" != "0a" ]; then
      dirty+=("$f")
    fi
  else
    # Empty file: append a newline
    dirty+=("$f")
  fi
done

if [ ${#dirty[@]} -eq 0 ]; then
  echo "normalize-board-eof: clean — all board files end with trailing newline."
  exit 0
fi

if [ "$MODE" = "--check" ]; then
  echo "normalize-board-eof: FAIL — ${#dirty[@]} board file(s) lack trailing newline:"
  for f in "${dirty[@]}"; do
    echo "  $f"
  done
  exit 1
fi

# Fix mode: append \n to each dirty file (idempotent — \n only, never rewrites body)
for f in "${dirty[@]}"; do
  echo "" >> "$f"
  echo "normalize-board-eof: fixed  $f"
done

echo "normalize-board-eof: appended trailing newline to ${#dirty[@]} file(s)."
