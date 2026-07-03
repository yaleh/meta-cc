---
id: TASK-10
title: >-
  Unify version source-of-truth: release.sh delegates to bump-plugin-version.sh,
  install-user reads version from plugin.json at runtime
status: 'Basic: Done'
assignee: []
created_date: '2026-06-23 11:32'
updated_date: '2026-06-23 12:02'
labels:
  - 'kind:basic'
dependencies: []
ordinal: 1000
---

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
# Proposal: Unify version source-of-truth: release.sh delegates to bump-plugin-version.sh, install-user reads version from plugin.json at runtime

## Background

The project has two independent code paths that write plugin version numbers: `release.sh` Step 2 (lines 128-165) directly edits `.claude-plugin/marketplace.json` and `plugin-src/.claude-plugin/plugin.json` using inline `jq` commands, while `bump-plugin-version.sh` separately manages all three version files including `plugin-src/.claude-plugin/marketplace.json`. Because `release.sh` omits `plugin-src/.claude-plugin/marketplace.json`, that file is left at the old version after a release. When `make install-user` subsequently `rsync`s `plugin-src/` into `~/.local/share/meta-cc/`, it installs the stale marketplace.json (e.g., version 3.3.1) while plugin.json correctly reflects the new version (e.g., 3.3.2). Claude Code detects this mismatch and raises a "plugin directory not found" error, breaking every user-scope install after a release. The root cause is duplicated, divergent version-management logic with no authoritative single source of truth.

## Goals

1. `bump-plugin-version.sh` accepts `--version X.Y.Z` and `--non-interactive` flags so it can be invoked programmatically with an exact target version, bypassing the interactive confirmation prompt and the dirty-working-tree guard.
2. `release.sh` Step 2 is replaced by a single delegation call to `bump-plugin-version.sh --version <VERSION_NUM> --non-interactive`, eliminating the inline `jq` block and making `bump-plugin-version.sh` the sole writer of all three version files.
3. The `install-user` Makefile target, after its `rsync` step, dynamically overwrites the `version` field in the installed `~/.local/share/meta-cc/.claude-plugin/marketplace.json` using the version read at runtime from `plugin-src/.claude-plugin/plugin.json`.
4. After any release, running `make install-user` produces an installed plugin where `plugin.json` and `marketplace.json` carry identical versions, so the "plugin directory not found" mismatch cannot occur.
5. The existing interactive (`patch`/`minor`/`major`) invocation of `bump-plugin-version.sh` continues to work exactly as before.

## Trade-offs and Risks

- **Not doing**: We are not collapsing the three version files into one canonical file and generating the others. That would require changes to the plugin loader contract and Claude Code internals.
- **Not doing**: We are not adding automated integration tests for the shell scripts. Verification requires a real or dry-run release cycle.
- **Risk — commit semantics in non-interactive mode**: `bump-plugin-version.sh` must skip its own `git commit` when `--non-interactive` is set, because `release.sh` owns the release commit. If this is implemented incorrectly, a spurious extra commit is created inside the release flow. The behavior must be documented in the script.
- **Risk — bypassing the dirty-tree guard**: Skipping the `git status` clean check in non-interactive mode is intentional (release.sh has already modified files), but it removes a safety net. A code comment should explain why the guard is deliberately omitted in this mode.

---

# Plan: Unify version source-of-truth

## Phase A: Add --version and --non-interactive flags to bump-plugin-version.sh
### Tests (write first)
- tests/scripts/bump-plugin-version.bats: test `--version 9.9.9 --non-interactive` sets all 3 JSON files to 9.9.9 without prompting or committing
- tests/scripts/bump-plugin-version.bats: test `--version bad-format` exits non-zero
- tests/scripts/bump-plugin-version.bats: test positional `patch` still works (backward compat)
### Implementation
- `scripts/release/bump-plugin-version.sh`: parse `--version X.Y.Z` and `--non-interactive` flags; skip `read` prompt and git clean guard when `--non-interactive`; skip git commit when `--non-interactive` (release.sh owns the commit); validate X.Y.Z format
### DoD
- [ ] `go test ./...`
- [ ] `bash scripts/release/bump-plugin-version.sh --version 9.9.9 --non-interactive 2>&1 | grep -q "updated to 9.9.9"`
- [ ] `test "$(jq -r '.version' plugin-src/.claude-plugin/plugin.json)" = "9.9.9"`
- [ ] `test "$(jq -r '.plugins[0].version' .claude-plugin/marketplace.json)" = "9.9.9"`
- [ ] `test "$(jq -r '.plugins[0].version' plugin-src/.claude-plugin/marketplace.json)" = "9.9.9"`
- [ ] `bash scripts/release/bump-plugin-version.sh --version bad-format --non-interactive; test $? -ne 0`

## Phase B: Replace release.sh Step 2 inline jq with delegation to bump-plugin-version.sh
### Tests (write first)
- tests/scripts/release-version-update.bats: dry-run mode verifies no file modifications
- tests/scripts/release-version-update.bats: in temp git repo, all 3 version files updated
### Implementation
- `scripts/release/release.sh`: remove inline `jq` block (lines 128-165); replace with `bash scripts/release/bump-plugin-version.sh --version "$VERSION_NUM" --non-interactive`; keep DRY_RUN guard and version-parity verification
### DoD
- [ ] `go test ./...`
- [ ] `! grep -q 'jq --arg ver.*marketplace.json' scripts/release/release.sh`
- [ ] `grep -q 'bump-plugin-version.sh' scripts/release/release.sh`
- [ ] `grep -q 'bump-plugin-version.sh.*--non-interactive' scripts/release/release.sh`

## Phase C: Fix install-user to stamp version at runtime from plugin.json
### Tests (write first)
- Manual: after `make install-user`, verify installed marketplace.json version matches plugin.json
### Implementation
- `Makefile` `install-user` target: after rsync line, add jq fixup that reads version from `plugin-src/.claude-plugin/plugin.json` and stamps it into `~/.local/share/meta-cc/.claude-plugin/marketplace.json`
### DoD
- [ ] `go test ./...`
- [ ] `grep -q 'Stamped installed marketplace.json' Makefile`
- [ ] `make install-user && test "$(jq -r '.plugins[0].version' ~/.local/share/meta-cc/.claude-plugin/marketplace.json)" = "$(jq -r '.version' plugin-src/.claude-plugin/plugin.json)"`

## Constraints
- `bump-plugin-version.sh` interactive mode must remain unchanged
- No new runtime dependencies beyond `jq`
- `release.sh` dry-run mode must still work without modifying any files
- `bump-plugin-version.sh` must not commit when `--non-interactive` is set; add code comment explaining release.sh owns the release commit

## Acceptance Gate
- [ ] `make push`
- [ ] `grep -q 'bump-plugin-version.sh' scripts/release/release.sh`
- [ ] `! grep -q 'jq --arg ver.*marketplace.json' scripts/release/release.sh`
- [ ] `grep -q 'Stamped installed' Makefile`
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Proposal approved. Starting plan draft.

Plan review iteration 1: APPROVED (fixed Phase C DoD item 3 — removed natural-language prefix 'After make install-user:' and combined into a single shell command chain with &&)

claimed: 2026-06-23T11:52:00Z

Completed: 2026-06-23T12:07:00Z
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Execution Summary

**Result:** Done  
**Commit:** b826570

### Changes
- `scripts/release/bump-plugin-version.sh` — added `--version X.Y.Z` and `--non-interactive` flags; branch/clean-dir guards and git commit skipped in non-interactive mode
- `scripts/release/release.sh` — removed inline jq block (Step 2); now delegates to `bump-plugin-version.sh --version "$VERSION_NUM" --non-interactive`
- `Makefile` install-user target — added runtime jq fixup to stamp installed marketplace.json version from plugin.json after rsync
- `plugin-src/.claude-plugin/marketplace.json` — fixed version from 3.3.1 → 3.3.2 to match plugin.json
- `tests/scripts/bump-plugin-version.bats` — new bats tests for --version/--non-interactive flags
- `tests/scripts/release-version-update.bats` — new bats tests verifying delegation pattern
- `tests/e2e/mcp-e2e-simple.sh` — fixed pre-existing jq crash on null session directory

### DoD
All 7 DoD checks passed: go test, bump-plugin-version --non-interactive, inline-jq removed from release.sh, bump-plugin-version call present, Makefile Stamped message, install-user version match, make push.
<!-- SECTION:FINAL_SUMMARY:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 go test ./...
- [x] #2 bash scripts/release/bump-plugin-version.sh --version 9.9.9 --non-interactive 2>&1 | grep -q "updated to 9.9.9"
- [x] #3 ! grep -q 'jq --arg ver.*marketplace.json' scripts/release/release.sh
- [x] #4 grep -q 'bump-plugin-version.sh.*--non-interactive' scripts/release/release.sh
- [x] #5 grep -q 'Stamped installed marketplace.json' Makefile
- [x] #6 make install-user && test "$(jq -r '.plugins[0].version' ~/.local/share/meta-cc/.claude-plugin/marketplace.json)" = "$(jq -r '.version' plugin-src/.claude-plugin/plugin.json)"
- [x] #7 make push
<!-- DOD:END -->
