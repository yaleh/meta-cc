#!/usr/bin/env bash
# it0-dod-check.sh — DoD meta-enforcer, charter M25-dod-meta-enforcer (DIR-017 Step 1). Thin
# wrapper delegating to it0-dod-check.ts — mirrors it0-backlog-projection-check.sh's exact
# wrapper shape (usage/arg check, node availability check, delegate, propagate exit code), the
# same `it0-*-check.sh` wraps `it0-*-check.ts` convention as every existing pair in `scripts/`.
#
# Given a milestone id, a charter file path, and an ABSORB-entry text file, runs all DoD clauses
# (0-9, including Clause 8 — task canonical-lifecycle-record, DIR-014 item 6 /
# M40-dir014-task-canonical-lifecycle-record — and Clause 9 — needs-human split-or-commit, DIR-026)
# defined in inherited-core.md's "Definition of Done" section. See it0-dod-check.ts's own header
# comment for the full per-clause implementation.
#
# DIR-015: a build-feasibility early screen (`go build ./...` via build-feasibility-check.sh)
# runs BEFORE the TS-enforcer delegation and fails closed (exit 1) when the workspace does
# not build — an unbuildable workspace can never satisfy DoD, so fail fast at the gate
# boundary. The same screen is registered standalone as the fixed gate `build-feasibility`
# in .quay/config.yml.
#
# Usage:
#   it0-dod-check.sh <milestone-id> <charter-file> <absorb-entry-file>
#
# Exit codes: 0 = PASS (all clauses satisfied); 1 = FAIL (clause violation or undeclared self-
# exemption); 2 = usage/environment error.

set -u

if [ "$#" -lt 3 ]; then
  echo "Usage: $0 <milestone-id> <charter-file> <absorb-entry-file>" >&2
  exit 2
fi

if ! command -v node >/dev/null 2>&1; then
  echo "ERROR: node required" >&2
  exit 2
fi

# DIR-015: build-feasibility early screen — fail closed before delegating to the TS
# enforcer when the workspace does not build. The screen prints its own reason.
"$(dirname "$0")/build-feasibility-check.sh" || exit 1

node "$(dirname "$0")/it0-dod-check.ts" "$1" "$2" "$3"
exit $?
