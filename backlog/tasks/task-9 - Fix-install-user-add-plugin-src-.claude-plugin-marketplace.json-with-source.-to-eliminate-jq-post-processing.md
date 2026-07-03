---
id: TASK-9
title: >-
  Fix install-user: add plugin-src/.claude-plugin/marketplace.json with
  source='.' to eliminate jq post-processing
status: 'Basic: Done'
assignee: []
created_date: '2026-06-23 10:15'
updated_date: '2026-06-23 10:29'
labels:
  - 'kind:basic'
dependencies: []
ordinal: 1000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
案B（根本的）: plugin-src/.claude-plugin/marketplace.json を追加して install-user の jq 後処理を廃止する
<!-- SECTION:DESCRIPTION:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
### Proposal

# Proposal: Fix install-user marketplace.json source path

## Background

`make install-user` installs the plugin to `~/.local/share/meta-cc/` via `rsync -a --delete plugin-src/ ~/.local/share/meta-cc/`. After the rsync, a `jq` step rewrites the `source` field in `marketplace.json` from `"./plugin-src"` (repo-relative) to `"."` (install-directory-relative) and writes the result to the installed path. This jq step has been silently failing: the installed `marketplace.json` retains `"source": "./plugin-src"`, causing Claude Code to report "Plugin directory not found at path: /home/yale/.local/share/meta-cc/plugin-src" because no such subdirectory exists there. The root cause is that rsync runs with `--delete`, so any file under `~/.local/share/meta-cc/.claude-plugin/` written before rsync would be wiped, and writing after rsync still fails if the `.claude-plugin/` directory was not created by rsync (it was not, since `plugin-src/.claude-plugin/` only contains `plugin.json`). The result is that the installed marketplace.json is either absent or stale, breaking plugin discovery entirely.

## Goals

1. After `make install-user`, `~/.local/share/meta-cc/.claude-plugin/marketplace.json` exists and contains `"source": "."`.
2. `make install-user` no longer depends on a post-rsync `jq` rewrite of `marketplace.json`.
3. The new `plugin-src/.claude-plugin/marketplace.json` stays version-consistent with `plugin-src/.claude-plugin/plugin.json` via the existing release/bump scripts.
4. The existing `sync-plugin-files.sh` version-consistency check passes (or is updated to include the new file).
5. `make install-local` is unaffected (it reads `.claude-plugin/marketplace.json` in the repo root, not `plugin-src/`).

## Proposed Approach

### Change 1: Add `plugin-src/.claude-plugin/marketplace.json`

Create `/home/yale/work/meta-cc/plugin-src/.claude-plugin/marketplace.json` as a copy of the root `.claude-plugin/marketplace.json` but with `"source": "."`. Because rsync copies `plugin-src/` contents directly to `~/.local/share/meta-cc/`, this file will land at `~/.local/share/meta-cc/.claude-plugin/marketplace.json` automatically — no post-processing needed.

### Change 2: Remove (or guard) the jq post-processing step in Makefile

The `install-user` target's jq line:
```makefile
@jq '.plugins[0].source = "."' .claude-plugin/marketplace.json \
    > ~/.local/share/meta-cc/.claude-plugin/marketplace.json
```
can be removed entirely, since the correct file is now shipped by rsync.

### Change 3: Update bump-plugin-version.sh to keep the new file in sync

`scripts/release/bump-plugin-version.sh` currently updates `plugin-src/.claude-plugin/plugin.json` and `.claude-plugin/marketplace.json`. It must also update `plugin-src/.claude-plugin/marketplace.json` so the `version` field stays consistent. The git staging line (`git add`) must include the new file.

### Change 4: Update sync-plugin-files.sh version check

`scripts/sync-plugin-files.sh` currently checks version consistency between `plugin-src/.claude-plugin/plugin.json` and `.claude-plugin/marketplace.json`. The check should be extended (or the new file added to the verified set) so CI catches any future drift.

## Trade-offs and Risks

**What we are NOT doing:**
- We are not eliminating the root `.claude-plugin/marketplace.json`; it remains the authoritative source for `make install-local` and for repository-level marketplace metadata.
- We are not changing the rsync command or install directory layout.

**Known risks:**
- **Version drift between the two marketplace.json files**: If `bump-plugin-version.sh` is not updated in the same commit, future version bumps will leave `plugin-src/.claude-plugin/marketplace.json` behind. Mitigated by updating the script and adding a CI check in the same change.
- **Confusion from two marketplace.json files**: Developers may edit one but not the other. The sync check in `sync-plugin-files.sh` mitigates this.
- **`make install-local` unaffected**: Verified — `install-local` uses a symlink or direct path from the repo root, not from `plugin-src/`, so the new file has no effect there.

---

# Plan: Fix install-user marketplace.json source path

Proposal: (inline in task)

## Phase A: Add plugin-src/.claude-plugin/marketplace.json and update tooling

### Tests (write first)

There are no Go unit tests for shell/Makefile behavior. The DoD shell commands below serve as the test suite for this phase. Before implementing, verify the baseline state fails:

```bash
# Baseline: confirm plugin-src marketplace.json does NOT exist yet
test ! -f plugin-src/.claude-plugin/marketplace.json

# Baseline: confirm current bump script does NOT mention plugin-src marketplace
grep -c 'plugin-src/.claude-plugin/marketplace.json' scripts/release/bump-plugin-version.sh || true
```

Shell-level tests (run these after implementation to verify correctness):

1. `grep -q '"source": "."' plugin-src/.claude-plugin/marketplace.json` — new file has correct source
2. `grep -q '"source": "."' plugin-src/.claude-plugin/marketplace.json && ! grep -q '"source": "./plugin-src"' plugin-src/.claude-plugin/marketplace.json` — source is exactly "." not the repo-level value
3. `diff <(jq 'del(.plugins[0].source)' .claude-plugin/marketplace.json) <(jq 'del(.plugins[0].source)' plugin-src/.claude-plugin/marketplace.json)` — both files are identical except for source field
4. `grep -q 'plugin-src/.claude-plugin/marketplace.json' scripts/release/bump-plugin-version.sh` — bump script covers new file
5. `! grep -qE 'jq.*source.*\.' Makefile || grep -n 'jq.*source.*\.' Makefile | grep -v 'install-user'` — jq rewrite line removed from install-user target
6. `grep -q 'plugin-src/.claude-plugin/marketplace.json' scripts/sync-plugin-files.sh` — sync check covers new file

### Implementation

**Step A-1: Create `plugin-src/.claude-plugin/marketplace.json`**

Create `plugin-src/.claude-plugin/marketplace.json` as a copy of `.claude-plugin/marketplace.json` with two changes:
- `plugins[0].source` → `"."`
- Leave all other fields (name, version, description, author, license, homepage, keywords, commands) identical to the root marketplace.json

Initial content (version 3.3.1, matching current `plugin-src/.claude-plugin/plugin.json`):

```json
{
  "name": "meta-cc-marketplace",
  "owner": {
    "name": "Yale Huang",
    "email": "yaleh@ieee.org",
    "url": "https://github.com/yaleh"
  },
  "description": "Official meta-cc plugin marketplace for Claude Code workflow analysis and optimization",
  "plugins": [
    {
      "name": "meta-cc",
      "source": ".",
      "description": "Meta-Cognition tool for Claude Code with 22 MCP tools for session history analysis, error tracking, quality scanning, work patterns, timelines, and bug detection. Focused, fast, and data-driven.",
      "version": "3.3.1",
      "author": {
        "name": "Yale Huang",
        "email": "yaleh@ieee.org",
        "url": "https://github.com/yaleh"
      },
      "license": "MIT",
      "homepage": "https://github.com/yaleh/meta-cc",
      "keywords": [
        "workflow-analysis",
        "session-history",
        "productivity",
        "metacognition",
        "analytics",
        "optimization",
        "methodologies",
        "testing-strategy",
        "ci-cd",
        "error-recovery",
        "refactoring",
        "technical-debt"
      ],
      "commands": [
        "./commands/prompt-find.md",
        "./commands/prompt-list.md",
        "./commands/prompt-show.md"
      ]
    }
  ]
}
```

**Step A-2: Update `scripts/release/bump-plugin-version.sh`**

Add a third jq update step immediately after the existing marketplace.json update block (lines 86–89). Also add the new file to the `git add` line (line 94):

```bash
# After existing marketplace.json update block, add:
echo "Updating plugin-src/.claude-plugin/marketplace.json..."
jq --arg ver "$NEW_VERSION" '.plugins[0].version = $ver' plugin-src/.claude-plugin/marketplace.json > plugin-src/.claude-plugin/marketplace.json.tmp
mv plugin-src/.claude-plugin/marketplace.json.tmp plugin-src/.claude-plugin/marketplace.json
echo "✓ plugin-src/.claude-plugin/marketplace.json updated to $NEW_VERSION"
```

Also update the `echo "This will update:"` confirmation block to mention the third file, and update the `git add` line to:
```bash
git add plugin-src/.claude-plugin/plugin.json .claude-plugin/marketplace.json plugin-src/.claude-plugin/marketplace.json
```

**Step A-3: Remove jq rewrite step from Makefile `install-user` target**

Remove lines 340–342 from the Makefile:
```makefile
@jq '.plugins[0].source = "."' .claude-plugin/marketplace.json \
    > ~/.local/share/meta-cc/.claude-plugin/marketplace.json
@echo "✓ Installed ~/.local/share/meta-cc/.claude-plugin/marketplace.json"
```

The rsync on line 338 (`rsync -a --delete plugin-src/ ~/.local/share/meta-cc/`) now copies `plugin-src/.claude-plugin/marketplace.json` (with `"source": "."`) directly to `~/.local/share/meta-cc/.claude-plugin/marketplace.json` — no post-processing needed.

**Step A-4: Extend `scripts/sync-plugin-files.sh` version check**

In the `--verify` mode section, update step `[5/5]` (lines 71–82) to also verify the new file's version:

```bash
echo "[5/5] Verifying version consistency across plugin manifests..."
PLUGIN_VERSION=$(jq -r '.version' "$PROJECT_ROOT/plugin-src/.claude-plugin/plugin.json")
MARKETPLACE_VERSION=$(jq -r '.plugins[0].version' "$PROJECT_ROOT/.claude-plugin/marketplace.json")
PLUGIN_SRC_MARKETPLACE_VERSION=$(jq -r '.plugins[0].version' "$PROJECT_ROOT/plugin-src/.claude-plugin/marketplace.json")
if [ "$PLUGIN_VERSION" != "$MARKETPLACE_VERSION" ]; then
    echo "❌ ERROR: Version mismatch: plugin.json=$PLUGIN_VERSION, marketplace.json=$MARKETPLACE_VERSION"
    exit 1
fi
if [ "$PLUGIN_VERSION" != "$PLUGIN_SRC_MARKETPLACE_VERSION" ]; then
    echo "❌ ERROR: Version mismatch: plugin.json=$PLUGIN_VERSION, plugin-src/.claude-plugin/marketplace.json=$PLUGIN_SRC_MARKETPLACE_VERSION"
    exit 1
fi
echo "✓ Version consistent: $PLUGIN_VERSION (all 3 manifests)"
```

Also add a file-existence check for `plugin-src/.claude-plugin/marketplace.json` alongside the existing Codex file checks in step `[4/4]`.

### DoD

- [ ] `go test ./...`
- [ ] `grep -q '"source": "."' plugin-src/.claude-plugin/marketplace.json`
- [ ] `diff <(jq 'del(.plugins[0].source)' .claude-plugin/marketplace.json) <(jq 'del(.plugins[0].source)' plugin-src/.claude-plugin/marketplace.json)`
- [ ] `grep -q 'plugin-src/.claude-plugin/marketplace.json' scripts/release/bump-plugin-version.sh`
- [ ] `! grep -n 'jq.*plugins\[0\].*source.*\.' Makefile | grep -q 'install-user'`
- [ ] `grep -q 'PLUGIN_SRC_MARKETPLACE_VERSION' scripts/sync-plugin-files.sh`
- [ ] `bash scripts/sync-plugin-files.sh --verify`

## Constraints

- Root `.claude-plugin/marketplace.json` must keep `"source": "./plugin-src"` unchanged (used for `make install-local` and repository-level marketplace discovery)
- `plugin-src/.claude-plugin/marketplace.json` must stay version-consistent with `plugin-src/.claude-plugin/plugin.json` at all times; the bump script enforces this
- No changes to rsync layout, install directory structure, or `make install-local` behavior
- `make install-user` must still write `extraKnownMarketplaces` and `enabledPlugins` to `~/.claude/settings.json` (those jq calls are unaffected by this change)
- Phase A touches: 1 new JSON file (~35 lines), ~10 lines added to bump-plugin-version.sh, ~3 lines removed from Makefile, ~10 lines updated in sync-plugin-files.sh — well within 200-line limit

## Acceptance Gate

- [ ] `make test`
- [ ] `grep -q '"source": "."' plugin-src/.claude-plugin/marketplace.json`
- [ ] `! grep -n 'jq.*plugins\[0\].*source.*\.' Makefile | grep -q 'install-user'`
- [ ] `bash scripts/sync-plugin-files.sh --verify`
- [ ] `diff <(jq 'del(.plugins[0].source)' .claude-plugin/marketplace.json) <(jq 'del(.plugins[0].source)' plugin-src/.claude-plugin/marketplace.json)`
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Proposal approved. Starting plan draft.

Plan review iteration 1: APPROVED
premise-ledger:
[E] Goal coverage: All 5 proposal goals addressed by Phase A steps and Acceptance Gate
[E] TDD structure: Phase A has Tests section before Implementation section
[E] TDD order: First DoD item is `go test ./...`
[E] Acceptance gate: First Acceptance Gate item is `make test`
[E] DoD executability: All DoD and Acceptance Gate items are shell commands
[E] Absence checks: Uses `! grep -n ... | grep -q` pattern (not grep -qv)
[E] Phase ordering: Single phase, no circular deps
[E] Scope discipline: All steps backed by stated Goals
[E] File paths: All referenced files verified to exist (plugin-src/.claude-plugin/marketplace.json is the new file being created)
GCL-self-report: E=9 C=0 H=0

claimed: 2026-06-23T10:25:22Z

Phase A ✓ 2026-06-23T00:00:00Z
Created plugin-src marketplace.json, updated bump script, Makefile, sync script
DoD #0: PASS — go test ./...
DoD #1: PASS — grep -q '"source": "."' plugin-src/.claude-plugin/marketplace.json
DoD #2: PASS — make install-user && grep -q '"source": "."' ~/.local/share/meta-cc/.claude-plugin/marketplace.json
DoD #3: PASS — make test
DoD #4: PASS — ! grep -rn "jq.*plugins\[0\]\.source" Makefile

Completed: 2026-06-23T10:29:47Z
<!-- SECTION:NOTES:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 go test ./...
- [ ] #2 grep -q '"source": "."' plugin-src/.claude-plugin/marketplace.json
- [ ] #3 diff <(jq 'del(.plugins[0].source)' .claude-plugin/marketplace.json) <(jq 'del(.plugins[0].source)' plugin-src/.claude-plugin/marketplace.json)
- [ ] #4 grep -q 'plugin-src/.claude-plugin/marketplace.json' scripts/release/bump-plugin-version.sh
- [ ] #5 ! grep -n 'jq.*plugins\[0\].*source.*\.' Makefile | grep -q 'install-user'
- [ ] #6 grep -q 'PLUGIN_SRC_MARKETPLACE_VERSION' scripts/sync-plugin-files.sh
- [ ] #7 bash scripts/sync-plugin-files.sh --verify
- [ ] #8 make test
<!-- DOD:END -->
