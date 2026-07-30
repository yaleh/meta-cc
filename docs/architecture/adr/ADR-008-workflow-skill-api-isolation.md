---
id: ADR-008
title: "Workflow Scripts Have No Claude Code Tool API (One-Way Workflow/Skill Bridge)"
status: accepted
date: 2026-07-30
---
# ADR-008: Workflow Scripts Have No Claude Code Tool API (One-Way Workflow/Skill Bridge)

## Status

Accepted

## Context

Claude Code Workflow scripts (`.claude/workflows/*.js`) execute in a Bun-compiled
JavaScript runtime that is **isolated from the Claude Code tool API**. The only
globals available to a workflow script are the workflow DSL primitives
(`phase`, `agent`, `parallel`, `log`, `args`). Calls to the conversational tool
API — `Skill()`, `Agent()`, `Bash()`, `Read()`, `Write()`, `Edit()`, and the
rest — are not bound in that runtime and throw
`ReferenceError: <tool> is not defined` immediately.

This was observed live on 2026-07-25: a `run-routines` workflow that wrapped
`Skill({skill: "quay:run-routines"})` crashed 24ms after start with
`ReferenceError: Skill is not defined` (DIR-073). The same constraint applies to
any workflow copy distributed into this repository's `.claude/workflows/`
directory: a stale copy of `run-routines.js` here still contained the `Skill()`
call and would crash identically if invoked.

The dispatch relationship between the two extension surfaces is therefore
**asymmetric (one-way)**:

- **Skills → Workflows**: a Skill (a markdown-driven conversational agent with
  the full tool API) MAY dispatch a workflow run.
- **Workflows → Skills**: a workflow script CANNOT invoke a Skill, nor any
  other tool-API surface. A workflow's only delegation mechanism is the
  workflow DSL's own `agent(...)` primitive, which dispatches a subagent
  through the workflow engine.

A second, related runtime quirk reinforces the isolation: the Workflow tool
sometimes delivers the `args` global as a JSON-encoded string rather than the
parsed object its contract promises (DIR-114), so canonical workflow copies
normalize `args` once up front
(`const $a = (typeof args === "string") ? JSON.parse(args) : args`) and never
touch bare `args` afterward. Workflow scripts also have no `import` capability,
which is why shared logic must be mirrored inline rather than imported.

Without a written rule, every new or re-synced workflow copy risked
re-introducing a `Skill()` call copied from conversational-context code, and
code review had no documented constraint to cite.

## Decision

1. **Workflow scripts MUST NOT call the Claude Code tool API.** Any occurrence
   of `Skill(`, `Agent(`, `Bash(`, `Read(`, `Write(`, `Edit(`, or other
   conversational tool calls inside a `.claude/workflows/*.js` file is a defect:
   the call will throw `ReferenceError` at runtime. The only permitted
   delegation primitive inside a workflow is the workflow DSL's `agent(...)`.

2. **The bridge is one-way; reuse happens at compile time, not runtime.**
   Functionality that both a Skill and a workflow need is shared by
   **compile-time shared JS modules / canonical source copies** (the quay
   canonical copies under the quay plugin's `workflows/` directory are the
   single source; consumer copies such as this repo's `.claude/workflows/` are
   synced byte-for-byte from them). Reuse is NEVER achieved by runtime
   delegation from a workflow to a Skill. When a workflow needs behavior that
   lives in a Skill, the workflow dispatches an `agent(...)` whose prompt
   re-states the Skill's procedure (see the canonical `run-routines.js`, which
   inlines the routine-track pipeline from `plugin/skills/routines/SKILL.md`).

3. **Normalize `args` up front.** Every workflow copy reads arguments only
   through a normalized `$a` binding
   (`const $a = (typeof args === "string") ? JSON.parse(args) : args`),
   matching the quay canonical copies; bare `args` field access past that
   point is a defect.

4. **Sync, don't hand-edit.** Consumer copies of workflows in this repository
   are re-synced from the quay canonical copies; semantic drift is fixed in the
   canonical source and propagated, never hand-patched in the consumer copy.
   Workflow lint rules that would mechanically enforce clause 1 belong to the
   quay generator and are out of scope for this repository (reported only).

## Consequences

### Positive Impacts

- **Crash class eliminated**: the synced `.claude/workflows/run-routines.js`
  and `drain-directives.js` no longer contain tool-API calls; the 24ms
  `ReferenceError` failure mode cannot recur from these copies.
- **Documented constraint**: authors and reviewers have a single citable rule
  for what a workflow script may call, instead of rediscovering the isolation
  at runtime.
- **Clear reuse path**: the one-way bridge plus compile-time module sharing
  gives an unambiguous answer to "how does a workflow reuse Skill logic" —
  inline it (or restate it in an `agent(...)` prompt), never delegate at
  runtime.
- **Canonical-source discipline**: byte-for-byte sync from the quay canonical
  copies prevents divergent hand-edits across consumer repositories.

### Negative Impacts

- **Logic duplication**: because workflow scripts cannot `import` and cannot
  call Skills, procedures shared between a Skill and a workflow exist in two
  places (Skill markdown and the workflow's `agent(...)` prompt), which can
  drift. Mitigation: the canonical copy names its source of truth in a comment
  (e.g. DIR-073's pointer to `plugin/skills/routines/SKILL.md`).
- **No mechanical enforcement in this repo**: the lint rule that would reject
  tool-API calls in workflow scripts belongs to the quay generator; here the
  constraint is enforced by review and this ADR only.

### Risks

- **Future staleness**: consumer workflow copies can drift stale again as the
  quay canonical copies evolve. Mitigation: re-sync on any workflow-related
  finding; treat any diff against the canonical copy as a defect in the
  consumer copy.
- **Copy-paste reintroduction**: conversational-context code pasted into a
  workflow will compile (it is just JS) but crash at runtime. Mitigation: this
  ADR; grep for tool-API call names as a manual check during review.

## Implementation

- [x] `.claude/workflows/run-routines.js` re-synced from the quay canonical
  copy (no `Skill()` call; dispatches the routine-track pipeline via
  `agent(...)`).
- [x] `.claude/workflows/drain-directives.js` re-synced from the quay
  canonical copy (adds the `$a` args normalization, DIR-114).
- [x] `.claude/workflows/execute-milestone.js` re-synced from the quay
  canonical copy (adds `$a` normalization, DIR-119-B composite-args handling,
  and the DIR-117-B preparation-receipt check).
- [x] Grep verifies no `Skill(`/tool-API call remains in any
  `.claude/workflows/*.js` copy (only comment references remain).
- [x] ADR-008 registered in `docs/architecture/adr/README.md`.

## Related Decisions

- [ADR-002](ADR-002-plugin-directory-structure.md) - Plugin Directory Structure Refactoring (the plugin surface whose workflow copies this constraint governs)

## Notes

Observed evidence (2026-07-25, DIR-073): a run-routines workflow wrapping
`Skill({skill:"quay:run-routines"})` crashed in 24ms with
`ReferenceError: Skill is not defined`. The workflow DSL globals are limited to
`phase`, `agent`, `parallel`, `log`, and `args`; there is no `import`
capability, so canonical logic that must also run server-side is mirrored
inline with a comment pointing at the canonical, unit-tested source (see
`composite-args.ts` mirroring in the canonical `execute-milestone.js`).
