#!/bin/bash
# check-session-locator-scope.sh - Fail the build if the raw, unscoped
# SessionLocator.FromSessionID primitive is called from outside
# internal/locator/.
#
# Background (DIR-033): locator.SessionLocator.FromSessionID(sessionID)
# resolves a session ID to a file by scanning every project-hash directory
# on disk, with NO comparison against the caller's working_dir/cwd. The
# identical defect shape — calling FromSessionID directly and forgetting
# the cwd-boundary comparison — was independently introduced and
# independently caught three separate times in this repo's history:
#   1. internal/mcp/executor/provider_query.go (DIR-030)
#   2. internal/provider/claude/provider.go's findSessionFile (DIR-032 build)
#   3. internal/analysis/service.go's loadData (DIR-032 adversarial audit)
#
# FromSessionIDScoped (internal/locator/args.go) crystallizes the fix: it
# wraps FromSessionID with the required boundary check in exactly one
# place. This script is the structural guardrail that keeps it that way —
# it fails loudly if a new unscoped `.FromSessionID(` call site appears
# anywhere outside internal/locator/ (the only place allowed to call the
# raw primitive is FromSessionIDScoped itself).
#
# Usage: bash scripts/checks/check-session-locator-scope.sh

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

echo "Checking for unscoped FromSessionID( call sites outside internal/locator/..."

# Match ".FromSessionID(" (a call), not "FromSessionIDScoped(" — the (
# right after FromSessionID excludes the Scoped variant since its next
# character is 'S', not '('. Exclude _test.go files (locator's own test
# fixtures/documentation legitimately exercise the raw primitive to prove
# FromSessionIDScoped's boundary check is necessary) and exclude
# internal/locator/ itself, the only sanctioned home for direct callers.
VIOLATIONS=$(grep -rn '\.FromSessionID(' --include='*.go' . 2>/dev/null \
    | grep -v '/internal/locator/' \
    | grep -v '_test\.go:' \
    || true)

if [ -n "$VIOLATIONS" ]; then
    echo -e "${RED}❌ ERROR: Found unscoped .FromSessionID( call site(s) outside internal/locator/:${NC}"
    echo ""
    echo "$VIOLATIONS" | sed 's/^/  /'
    echo ""
    echo "FromSessionID is a GLOBAL search across every project-hash directory"
    echo "on disk with no comparison against the caller's working_dir/cwd —"
    echo "calling it directly is the cross-project session leak class fixed"
    echo "in DIR-030/DIR-032/DIR-033. Use SessionLocator.FromSessionIDScoped"
    echo "(sessionID, workingDir) instead, which enforces the cwd boundary."
    exit 1
fi

echo -e "${GREEN}✓ No unscoped FromSessionID( call sites found outside internal/locator/${NC}"
exit 0
