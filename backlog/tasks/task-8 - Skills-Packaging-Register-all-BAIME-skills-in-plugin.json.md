---
id: TASK-8
title: 'Skills Packaging: Register all BAIME skills in plugin.json'
status: 'Basic: Done'
assignee: []
created_date: '2026-06-23 06:25'
updated_date: '2026-06-23 07:04'
labels:
  - 'kind:basic'
dependencies: []
references:
  - docs/proposals/skills-packaging-proposal.md
  - plugin-src/.claude-plugin/plugin.json
  - .claude-plugin/marketplace.json
modified_files:
  - plugin-src/.claude-plugin/plugin.json
priority: low
ordinal: 1000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The `plugin-src/.claude-plugin/plugin.json` (and mirrored in `.claude-plugin/marketplace.json`) declares only `commands` (prompt-find, prompt-list, prompt-show) but omits the corresponding `skills` array. Three skill directories already exist under `plugin-src/skills/` with valid `SKILL.md` entry points, but they are not registered in the plugin manifest.

See `docs/proposals/skills-packaging-proposal.md` for the original proposal context (which also covers BAIME skills in `.claude/skills/` — but for this repo the immediate gap is the three prompt-library skills in `plugin-src/skills/`).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 plugin-src/.claude-plugin/plugin.json contains a `skills` array with 3 entries
- [ ] #2 All three entries resolve to existing SKILL.md files under plugin-src/skills/
- [ ] #3 `jq . plugin-src/.claude-plugin/plugin.json` exits 0 (valid JSON)
- [ ] #4 `go test ./...` passes without errors
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Proposal

### Background

The `plugin-src/.claude-plugin/plugin.json` manifest currently declares only `commands`:

```json
"commands": [
  "./commands/prompt-find.md",
  "./commands/prompt-list.md",
  "./commands/prompt-show.md"
]
```

Three skills already exist under `plugin-src/skills/` with valid `SKILL.md` entry points:
- `plugin-src/skills/prompt-find/SKILL.md`
- `plugin-src/skills/prompt-list/SKILL.md`
- `plugin-src/skills/prompt-show/SKILL.md`

The same omission exists in `.claude-plugin/marketplace.json` (the local marketplace entry that references `./plugin-src` as source).

**User experience impact**: When users install the plugin via the Claude Code marketplace, the skills are never discovered by Claude Code because they are absent from the manifest. Users must manually invoke commands instead of being able to rely on Claude auto-selecting the appropriate skill.

### Goals

1. Register all three existing skills in `plugin-src/.claude-plugin/plugin.json` under a `skills` array.
2. Ensure the `marketplace.json` entry also reflects the skills (it inherits from the plugin source, so updating `plugin.json` is sufficient).
3. No new files need to be created — only existing manifests need updating.

### Proposed Approach

Add a `skills` array to `plugin-src/.claude-plugin/plugin.json`:

```json
"skills": [
  "./skills/prompt-find/SKILL.md",
  "./skills/prompt-list/SKILL.md",
  "./skills/prompt-show/SKILL.md"
]
```

No changes needed to `marketplace.json` (it references the plugin source directory; the plugin.json there is the authoritative manifest).

### Trade-offs

| Concern | Decision |
|---|---|
| Version bump needed? | No — adding a `skills` array is additive and non-breaking; a patch bump is optional but not required for correctness. |
| New plugin release needed? | Only if the team wants to publish the fix; the local marketplace will pick it up immediately after the JSON change. |
| Risk of breaking existing users? | None — adding a `skills` section does not affect commands or MCP server behaviour. |

---

## TDD Implementation Plan

### Phase A — Update plugin manifest to register all skills

**Goal**: Add `skills` array to `plugin-src/.claude-plugin/plugin.json` listing all three `SKILL.md` entry points.

**Tests (write / verify first)**:

1. JSON validity check:
   ```bash
   jq . plugin-src/.claude-plugin/plugin.json > /dev/null && echo "PASS"
   ```

2. Skills count assertion:
   ```bash
   COUNT=$(jq '.skills | length' plugin-src/.claude-plugin/plugin.json)
   [ "$COUNT" -eq 3 ] && echo "PASS: $COUNT skills" || echo "FAIL: expected 3, got $COUNT"
   ```

3. All SKILL.md files have a corresponding manifest entry:
   ```bash
   # Every skills/ subdirectory must have an entry in plugin.json
   for skill_dir in plugin-src/skills/*/; do
     skill_name=$(basename "$skill_dir")
     entry="./skills/${skill_name}/SKILL.md"
     jq -e --arg e "$entry" '.skills[] | select(. == $e)' plugin-src/.claude-plugin/plugin.json > /dev/null \
       && echo "PASS: $entry" || echo "FAIL: $entry not registered"
   done
   ```

**Implementation**: Edit `plugin-src/.claude-plugin/plugin.json` to add `skills` array after `commands`.

**DoD**:
- `jq . plugin-src/.claude-plugin/plugin.json > /dev/null` exits 0
- `jq '.skills | length' plugin-src/.claude-plugin/plugin.json` returns `3`
- All three skill paths resolve to existing files (`plugin-src/skills/prompt-find/SKILL.md`, etc.)
- `go test ./...` passes (no Go code changed, just JSON)

**Acceptance Gate**:
```bash
go test ./...
jq . plugin-src/.claude-plugin/plugin.json > /dev/null
jq '.skills | length' plugin-src/.claude-plugin/plugin.json  # expect 3
```
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
claimed: 2026-06-23T07:05:00Z

Phase A ✓ 2026-06-23T00:00:00Z: added skills array to plugin.json

Completed: 2026-06-23T07:06:00Z
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
## Execution Summary\n\n**Result:** Done\n**Commit:** 2832936\n\n### Changes\n- Added `skills` array to `plugin-src/.claude-plugin/plugin.json` with 3 entries\n- All DoD checks passed: valid JSON, 3 skills registered, all SKILL.md paths verified, go test ./... green
<!-- SECTION:FINAL_SUMMARY:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 JSON validity: `jq . plugin-src/.claude-plugin/plugin.json > /dev/null` exits 0
- [ ] #2 Skills count: `jq '.skills | length' plugin-src/.claude-plugin/plugin.json` returns 3
- [ ] #3 All skill paths verified: each `./skills/*/SKILL.md` entry exists on disk
- [ ] #4 Go tests green: `go test ./...` passes
<!-- DOD:END -->
