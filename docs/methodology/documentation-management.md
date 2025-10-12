# Documentation Management Methodology for Claude Code Projects

A language-agnostic, project-independent operational guide to managing documentation in software projects using Claude Code.

**Core Documents**: This methodology emphasizes **plan.md** (roadmap) and **principles.md** (design constraints) as the foundational documents that should be created from the very beginning of any project.

---

## Quick Navigation

### For Project Startup
1. [Phase 0: Bootstrap Your Project](#phase-0-bootstrap-your-project) - Complete startup guide
2. [Essential Files and Templates](#essential-files-and-templates) - What to create

### For Active Development
3. [Commit/Merge/Release Checklists](#commit-merge-release-checklists) - Pre-action verification
4. [Synchronization Decision Matrix](#synchronization-decision-matrix) - What to update when

### For Reference
5. [Core Principles](#core-principles) - DRY, Progressive Disclosure, Task-Oriented
6. [Real-World Example: meta-cc](#real-world-example-meta-cc) - See it in practice

---

## Phase 0: Bootstrap Your Project

**Goal**: Create minimal viable documentation that enables Claude Code to assist effectively from day one.

**Time Budget**: 1-3 days for initial setup, then evolve incrementally

### Core Requirements

**✅ Must Create** (in any order):

1. **`docs/core/plan.md`** (50-200 lines) - **REQUIRED Core Document #1**
   ```markdown
   # Development Plan

   ## Vision
   What are we building? (1-2 paragraphs)

   ## Phases
   - Phase 0: Bootstrap - Status: 🚧 In Progress
   - Phase 1: [First milestone] - Status: 📋 Planned
   - Phase 2: [Second milestone] - Status: 📋 Planned

   ## Current Status
   Working on: [Current task]
   Blockers: [Any blockers]

   ## Notes
   [Keep adding notes as you develop]
   ```

2. **`docs/core/principles.md`** (30-100 lines) - **REQUIRED Core Document #2**
   ```markdown
   # Design Principles

   ## Core Constraints
   - Code limits: [e.g., ≤500 lines per phase, ≤200 per stage]
   - Test coverage: [e.g., ≥80%]
   - [Language-specific constraints]

   ## Development Methodology
   - TDD: Write tests before implementation
   - [Other methodologies you'll follow]

   ## Architecture Principles
   - [Key architectural decisions]
   - [Non-negotiable design rules]
   ```

3. **`CLAUDE.md`** (50-150 lines)
   ```markdown
   # CLAUDE.md

   Development guide for Claude Code.

   ## Quick Links
   - [docs/core/plan.md](docs/core/plan.md) - Current roadmap
   - [docs/core/principles.md](docs/core/principles.md) - Design constraints

   ## Project Goal
   [One paragraph: what problem does this solve?]

   ## Development Commands
   - Build: `[your-build-command]`
   - Test: `[your-test-command]`
   - Lint: `[your-lint-command]`

   ## Architecture Notes
   [Add notes here as patterns emerge]

   ## FAQ
   [Add Q&A as questions arise during development]
   ```

4. **`README.md`** (50-200 lines)
   ```markdown
   # Project Name

   [One-sentence description]

   ## Status
   🚧 Early development - documentation in progress

   ## Quick Start

   ```bash
   # Installation
   [installation-command]

   # Basic usage
   [basic-usage-command]
   ```

   ## Documentation
   - [Development Guide](CLAUDE.md) - For contributors
   - [Development Plan](docs/core/plan.md) - Roadmap

   ## License
   [License type]
   ```

5. **Basic files**:
   - `.gitignore` - Standard for your language/stack
   - `LICENSE` - Choose appropriate license

6. **Directory structure**:
   ```
   docs/
   ├── core/                # Core documents
   │   ├── plan.md         # Core document #1
   │   └── principles.md   # Core document #2
   ├── guides/              # Empty initially
   ├── reference/           # Empty initially
   ├── tutorials/           # Empty initially
   └── architecture/
       └── adr/             # Empty initially
   ```

**❌ Avoid**:
- Writing comprehensive documentation before code exists
- Creating files you won't use immediately
- Detailed reference docs before patterns emerge

### Optional (Add When Needed)

Create these as you develop, not upfront:

- **`docs/architecture/adr/template.md`** - When first architectural decision needed
- **`docs/architecture/adr/ADR-001-*.md`** - First ADR
- **`docs/DOCUMENTATION_MAP.md`** - When docs grow beyond 5-6 files
- **`CONTRIBUTING.md`** - When expecting external contributors
- **`CHANGELOG.md`** - When approaching first release

### Phase 0 Completion Criteria

**Exit Criteria** (checklist):
- [ ] `docs/core/plan.md` exists with Phase 0-2 outline
- [ ] `docs/core/principles.md` exists with core constraints
- [ ] `CLAUDE.md` exists with project goal and commands
- [ ] `README.md` has project description
- [ ] Code compiles and basic tests pass
- [ ] Directory structure created

**Verification** (pseudocode):
```bash
verify_phase_0() {
  assert file_exists("docs/core/plan.md")
  assert file_exists("docs/core/principles.md")
  assert file_exists("CLAUDE.md")
  assert grep_count("Phase", "docs/core/plan.md") >= 2
  assert build_succeeds()
}
```

**Common Mistakes**:
- ❌ Skipping plan.md or principles.md (these are essential!)
- ❌ Writing too much documentation before writing code
- ❌ Creating complex directory structures prematurely
- ❌ Copy-pasting templates without customization

---

## Ongoing: Continuous Improvement

**After Phase 0**, documentation evolves continuously with your project. There are no fixed phases - just continuous synchronization at Git events.

### Key Activities

**As patterns emerge** (weeks 2-4):
- Extract development workflow → `docs/guides/development.md`
- Create ADR template → `docs/architecture/adr/template.md`
- Document first architectural decision → `ADR-001`
- Expand CLAUDE.md FAQ from actual questions
- Update README.md with tested installation steps

**As project matures** (months 2-3):
- Create CONTRIBUTING.md for external contributors
- Start CHANGELOG.md for releases
- Extract reference docs (CLI, API)
- Add cookbook entries from real usage

**Quarterly maintenance**:
- Consolidate redundant content
- Archive outdated docs
- Fix broken links
- Measure documentation health

### Continuous Improvement Checklist

**Regular tasks** (per sprint/phase):
- [ ] Update plan.md with progress
- [ ] Add FAQ entries to CLAUDE.md
- [ ] Document new patterns in guides
- [ ] Create ADRs for architectural decisions
- [ ] Keep README.md current

**Health checks** (quarterly):
```bash
# Pseudocode for health verification
verify_doc_health() {
  assert word_count("README.md") <= 500
  assert word_count("CLAUDE.md") <= 400
  assert broken_link_count() == 0
  warn_if documentation_age("critical_docs") > 90_days
}
```

---

## Plans Directory: Detailed Implementation Plans

### Purpose and Structure

**Two-tier planning system**:
- **docs/core/plan.md**: High-level roadmap (50-100 lines per phase)
- **plans/N/**: Detailed implementation (2000-3000 lines per phase)

**When to create plans/N/**:
- Complex phases requiring multiple stages (≥3 stages)
- Phases needing detailed TDD guidance
- When using @agent-project-planner for systematic execution

**Directory structure**:
```
plans/
├── 8/
│   ├── plan.md              # Detailed phase plan
│   ├── README.md            # Quick reference (optional)
│   └── stage-8.12.md        # Stage-specific details (optional)
├── 13/
│   ├── plan.md
│   └── README.md
└── 16/
    └── plan.md
```

### Plan Generation Workflow

**Step 1: Create phase overview in docs/core/plan.md**
```markdown
### Phase N: [Name]
**Goal**: [What to achieve]
**Stages**:
- Stage N.1: [First stage]
- Stage N.2: [Second stage]
**Code limit**: ≤500 lines
```

**Step 2: Generate detailed plan**
```
User: "@docs/core/plan.md @docs/core/principles.md
       使用 @agent-project-planner 为 Phase N 生成详细计划"

Result: plans/N/plan.md created (~2000-3000 lines)
```

**Step 3: Execute stages sequentially**
```
User: "@plans/N/ 列出 Phase N 各 stage，
       并使用 @agent-stage-executor 逐一执行"

Flow: Stage N.1 → N.2 → N.3 → ... (serial execution)
```

### Plan Content Structure

**Typical plans/N/plan.md structure**:
```markdown
# Phase N: [Name]

## Overview
- Background & Problems
- Goals & Success Criteria
- Architecture Design
- Code Budget (~500 lines)

## Stage N.1: [Name]
- Objective
- Acceptance Criteria
- TDD Approach (tests first)
- File Changes (new/modified/deleted)
- Test Commands
- Dependencies
- Expected Code (~100-200 lines)

## Stage N.2: [Name]
[Same structure]

## Integration Testing
## Documentation Updates
## Success Metrics
```

### Stage Execution Principles

**Agent-assisted execution**:
- **@agent-stage-executor** reads plans/N/plan.md
- Follows TDD cycle automatically
- Verifies tests pass after each stage
- Reports detailed results

**Execution pattern**:
```
For each Stage X.Y:
  1. Read plans/N/plan.md Stage X.Y section
  2. Write tests (TDD first)
  3. Implement code
  4. Run tests (must pass)
  5. Report: files changed, tests passed, metrics
```

**Serial execution principle**:
- One stage at a time (not parallel)
- Wait for stage complete before next
- Accumulate results for phase summary

### Plans vs Documentation Updates

**Update triggers**:
- **During phase**: Update docs/core/plan.md (progress, status)
- **Stage complete**: Note in docs/core/plan.md (Stage N.X ✅)
- **Phase complete**: Mark docs/core/plan.md (Phase N ✅ + metrics)
- **plans/N/plan.md**: Read-only during execution (no updates)

**Anti-pattern**: Updating plans/N/plan.md during execution
- Plan is reference, not journal
- Use docs/core/plan.md for status tracking

### Plans Directory Lifecycle

**Creation**: Phase start (via @agent-project-planner)
**Usage**: During phase execution (reference only)
**Archival**: After phase complete (preserve as historical record)

**Maintenance**:
- Keep all plans/N/ directories (don't delete)
- Useful for understanding past decisions
- Reference for similar future phases

---

## Essential Files and Templates

### Minimal Files (Phase 0)

**Required immediately**:
- `README.md` - Project entry point
- `CLAUDE.md` - Development guide
- `docs/core/plan.md` - Roadmap (Core #1)
- `docs/core/principles.md` - Constraints (Core #2)
- `.gitignore` - Version control
- `LICENSE` - Legal

**Add as needed**:
- `docs/architecture/adr/template.md` - ADR template
- `docs/DOCUMENTATION_MAP.md` - Navigation
- `CONTRIBUTING.md` - Contribution guide
- `CHANGELOG.md` - Change history

### Complete Template Library

#### ADR Template (`docs/architecture/adr/template.md`)

```markdown
# ADR-NNN: [Decision Title]

**Status**: Proposed | Accepted | Deprecated | Superseded

**Date**: YYYY-MM-DD

## Context
What issue motivates this decision? What factors are at play?

## Decision
What change are we proposing or have agreed to?

## Consequences
What becomes easier or harder?

### Positive
- Benefit 1
- Benefit 2

### Negative
- Tradeoff 1
- Tradeoff 2

## Implementation
(Optional) Implementation notes or status

## Related Decisions
- [ADR-XXX](ADR-XXX-title.md)
```

#### Phase Plan Template (for `docs/core/plan.md`)

```markdown
# Development Plan

## Vision
[What are we building? Why does it matter?]

## Phases

### Phase 0: Bootstrap (Current)
**Status**: 🚧 In Progress
**Goal**: Minimal viable documentation and project structure
**Tasks**:
- [x] Create plan.md and principles.md
- [ ] Implement core feature X
- [ ] Write tests for X

### Phase 1: [Milestone Name]
**Status**: 📋 Planned
**Goal**: [What will this achieve?]
**Dependencies**: Phase 0 complete
**Tasks**:
- [ ] Task 1
- [ ] Task 2

### Phase 2: [Next Milestone]
**Status**: 📋 Planned
**Goal**: [What will this achieve?]
**Dependencies**: Phase 1 complete

## Current Status
**Working on**: Phase 0 - Core feature X
**Blockers**: None
**Last updated**: YYYY-MM-DD

## Notes
[Add notes here as you develop]
```

#### Archive Metadata Template

```markdown
---
archived_date: YYYY-MM-DD
original_path: docs/[path]/old-doc.md
reason: superseded | deprecated | obsolete
replaced_by: docs/[path]/new-doc.md
---

# [Archived] Document Title

**⚠️ This document is archived and no longer maintained.**

- **Archived**: YYYY-MM-DD
- **Reason**: [Brief explanation]
- **Replaced by**: [Link to new doc] (if applicable)

---

[Original content follows...]
```

#### Documentation Map Template

```markdown
# Documentation Map

## Quick Navigation

### For New Users
1. [README.md](../README.md) - Quick start
2. [docs/guides/getting-started.md](guides/getting-started.md) - Tutorial

### For Developers
1. [CLAUDE.md](../CLAUDE.md) - Development guide
2. [docs/core/plan.md](plan.md) - Roadmap
3. [docs/core/principles.md](principles.md) - Design constraints
4. [docs/architecture/adr/README.md](architecture/adr/README.md) - Decisions

## Document Roles

| Document | Purpose | Update Frequency |
|----------|---------|------------------|
| README.md | Quick start | Major releases |
| CLAUDE.md | Development guide | Per phase |
| docs/core/plan.md | Roadmap | Continuous |
| docs/core/principles.md | Design constraints | Rarely |
```

---

## Commit/Merge/Release Checklists

### Pre-Commit Checklist

**Before every commit with code changes**:

- [ ] **Code comments updated**?
  ```bash
  # Verification concept:
  if code_changed && complex_logic:
      assert comments_explain_why()
  ```

- [ ] **Inline docs updated** (function/class docs)?
  ```bash
  # Pseudocode:
  check_docstring_coverage(changed_files)
  ```

- [ ] **ADR needed**? (If architectural decision made)
  - ✅ Yes → Create `docs/architecture/adr/ADR-NNN-[title].md`
  - ❌ No → Proceed

**Time investment**: 5-10% of development time

---

### Pre-Merge Checklist

**Before merging to main/develop**:

- [ ] **CLAUDE.md updated**? (If workflow changed)
  - New commands → Update "Development Commands" section
  - New FAQ → Add to FAQ section

- [ ] **docs/core/plan.md updated**?
  - Mark completed tasks as done
  - Update "Current Status"
  - Add lessons learned

- [ ] **Task guides updated**? (If workflow affected)

- [ ] **Link check passed**?
  ```bash
  # Verification concept:
  find docs -name "*.md" | xargs markdown-link-check
  ```

**Time investment**: 10-15% of phase time

---

### Pre-Release Checklist

**Before creating release tag**:

- [ ] **README.md current**?
  - Installation steps tested ✅
  - Feature list accurate ✅
  - Version references correct ✅

- [ ] **CHANGELOG.md updated**?
  ```markdown
  ## [v1.0.0] - YYYY-MM-DD
  ### Added
  - Feature 1
  ### Changed
  - Change 1
  ### Fixed
  - Bug 1
  ```

- [ ] **Reference docs complete**?
  - All commands documented
  - All APIs documented

- [ ] **Migration guide** (if breaking changes)?

- [ ] **All examples tested**?

**Readiness Check** (pseudocode):
```bash
pre_release_check() {
  assert readme_word_count() <= 500
  assert changelog_has_version(release_version)
  assert all_examples_pass()
  assert broken_link_count() == 0
}
```

**Time investment**: 1-2 days for major release

---

## Synchronization Decision Matrix

### Quick Decision Rules

1. **Did I write code?** → Update inline docs, maybe ADR
2. **Did I merge a feature?** → Update CLAUDE.md and plan.md
3. **Am I releasing?** → Update README and reference docs
4. **Did architecture change?** → Create ADR (required)
5. **Did workflow change?** → Update CLAUDE.md and guides

### Complete Matrix

**7 change types × 7 document types**:

| Change Type ↓ \ Doc → | Inline Docs | CLAUDE.md | Task Guides | Reference | ADR | plan.md | README.md |
|------------------------|-------------|-----------|-------------|-----------|-----|---------|-----------|
| **New Feature** | ✅ Function docs | ✅ FAQ if new | ✅ If workflow | ✅ CLI/API | ❌ Usually not | ✅ Update status | ⚠️ If major |
| **Bug Fix** | ✅ If logic changes | ❌ Rare | ❌ Rare | ❌ No | ❌ No | ❌ No | ❌ No |
| **Refactor** | ✅ If signature | ❌ Usually not | ❌ Usually not | ⚠️ If public API | ⚠️ If major | ❌ No | ❌ No |
| **Architecture** | ✅ Yes | ✅ If workflow | ⚠️ If affects | ❌ No | ✅ **Required** | ⚠️ If blocks | ❌ Rare |
| **Workflow** | ❌ No | ✅ **Required** | ✅ **Required** | ❌ No | ❌ No | ⚠️ If timeline | ❌ No |
| **Phase Done** | ❌ No | ✅ Update FAQ | ✅ Update | ⚠️ Extract | ⚠️ Review | ✅ **Required** | ⚠️ If milestone |
| **Release** | ❌ No | ✅ Versions | ✅ Examples | ✅ **Complete** | ❌ No | ✅ Mark | ✅ **Required** |

**Legend**:
- ✅ **Required** - Must update
- ⚠️ **Conditional** - Update if condition met
- ❌ **Not needed** - Usually skip

---

## Core Principles

### 1. Separation of Concerns

Different audiences need different docs:

- **End Users** → README.md (< 500 lines)
- **Power Users** → Reference docs (complete)
- **Contributors** → CLAUDE.md, guides
- **Maintainers** → ADRs, principles.md

### 2. DRY (Don't Repeat Yourself)

Single source of truth for each concept.

**Anti-pattern**: MCP setup explained in README, CLAUDE.md, and guides.

**Correct**: One complete guide (`docs/guides/mcp-setup.md`), others link to it.

### 3. Progressive Disclosure

Start simple, link deeper:

```
README.md (Quick start, 300 lines)
    ↓
docs/guides/getting-started.md (Tutorial, 500 lines)
    ↓
docs/guides/complete-guide.md (Comprehensive, 2000 lines)
    ↓
docs/reference/api.md (Spec, no limit)
```

### 4. Task-Oriented Organization

Organize by user goals, not system internals:

- ✅ `docs/guides/plugin-development.md` (goal: develop plugin)
- ❌ `docs/parser-internals.md` (system component)

### 5. Living Documentation

Docs evolve with code:

- Code changes → Update inline docs
- Architecture decision → Create ADR
- User questions → Update FAQ
- Usage patterns → Add cookbook entries

---

## Real-World Example: meta-cc

The meta-cc project demonstrates this methodology in practice.

### Evolution Timeline

**Phase 1**: Anti-pattern
- README.md: 1909 lines (everything in one file)
- Hard to navigate
- High Claude Code context cost

**Phase 2**: Extraction (85% reduction)
- README.md: 275 lines
- Created: cli-reference.md, features.md, jsonl-reference.md
- Improved but still had redundancy

**Phase 3**: Consolidation
- Merged 4 MCP docs → 1 comprehensive guide
- Created DOCUMENTATION_MAP.md
- Archived outdated content

**Phase 4**: Optimization (54% reduction)
- CLAUDE.md: 607 → 278 lines
- Created task-specific guides:
  - plugin-development.md
  - repository-structure.md
  - git-hooks.md
  - release-process.md
- CLAUDE.md became navigation hub

### Key Metrics

| Metric | Before | After | Change |
|--------|--------|-------|--------|
| README.md | 1909 lines | 275 lines | -85% |
| CLAUDE.md | 607 lines | 278 lines | -54% |
| MCP docs | 4 docs | 1 doc | -75% |
| Total size | ~15,000 lines | ~14,000 lines | -7% |

**Key Insight**: Total size decreased slightly, but organization improved dramatically.

### Most Accessed Documents

From meta-cc's own analysis:

1. **docs/core/plan.md**: 421 accesses - Most accessed (validates this methodology!)
2. **README.md**: 170 accesses
3. **docs/core/principles.md**: 89 accesses - Second most important
4. **CLAUDE.md**: 69 accesses

**Lesson**: plan.md and principles.md are indeed the core documents. Create them first!

### Lessons Learned

1. **Consolidate early** - Don't wait for doc sprawl
2. **Task-oriented works** - "plugin-development.md" > "plugin-structure.md" + "plugin-sync.md"
3. **Navigation matters** - DOCUMENTATION_MAP.md crucial
4. **Archive, don't delete** - Historical value
5. **Measure access** - Use data to guide optimization

### Document Update Principles

**Three documents serve different update rhythms and purposes:**

#### plan.md: High-Frequency Journal

**Update triggers**:
- Starting new phase/milestone
- Completing tasks or phases
- Discovering blockers or pivoting direction
- Daily/weekly progress tracking

**Update principle**: **Write-often, read-before-coding**
- Treat as living roadmap, not static plan
- Update status continuously (like a development journal)
- Add notes about decisions, discoveries, blockers
- Mark phases ✅ when complete, document actual outcomes

**Anti-pattern**: Only updating at phase boundaries (plan becomes stale)

#### principles.md: Low-Frequency Constitution

**Update triggers**:
- Establishing initial constraints (Phase 0)
- Discovering new architectural patterns
- Codifying repeated decisions into rules
- Refining constraints based on experience

**Update principle**: **Write-rarely, read-always**
- Treat as project constitution (stable, authoritative)
- Only update when patterns solidify into principles
- Reference before implementing new features
- Expand when architecture matures (not prematurely)

**Anti-pattern**: Updating too frequently (principles should be stable)

#### CLAUDE.md: FAQ-Driven Navigation

**Update triggers**:
- Questions arise during development (add to FAQ)
- Development workflow changes (update commands)
- Documentation grows (maintain navigation links)
- New patterns emerge (simplify, link to guides)

**Update principle**: **Accumulate-then-consolidate**
- Add FAQ entries as questions come up
- Periodically consolidate into guides
- Evolve into navigation hub (links > duplication)
- Keep concise (< 400 lines)

**Anti-pattern**: Duplicating content from other docs (DRY violation)

#### Combined Workflow

**Phase start**:
```
1. Read principles.md (understand constraints)
2. Update docs/core/plan.md (add phase overview ~100 lines)
3. Generate plans/N/plan.md (@agent-project-planner)
   - Detailed TDD implementation plan
   - Stage-by-stage breakdown
   - ~2000-3000 lines comprehensive
4. Optional: Create plans/N/README.md (quick reference)
```

**During development**:
```
1. Execute stages sequentially (@agent-stage-executor)
   - Read plans/N/plan.md Stage X.Y section
   - Follow TDD cycle: tests → implement → verify
   - Report results after each stage
2. Update docs/core/plan.md (status, completed stages)
3. Add to CLAUDE.md FAQ (as questions arise)
4. Reference principles.md (ensure compliance)
```

**Phase complete**:
```
1. Mark phase complete in docs/core/plan.md (✅ + metrics)
2. Update principles.md if new patterns emerged
3. Update CLAUDE.md if workflow changed
4. Archive plans/N/ (preserve as reference)
```

**Two-level planning**:
- **docs/core/plan.md**: High-level phase overview (goal, stages, deliverables)
- **plans/N/plan.md**: Detailed implementation (TDD cycles, file changes, tests)

**Agent-assisted workflow**:
- **@agent-project-planner**: Generate detailed plans from phase overview
- **@agent-stage-executor**: Execute stages systematically with verification

**Commit patterns**:
- `docs:` prefix for documentation-only changes
- `feat(phase-N): implement X` for features
- Update docs/core/plan.md + plans/N/ together

---

## Document Size Guidelines

| Document | Target | Maximum | Rationale |
|----------|--------|---------|-----------|
| README.md | 200-300 | 500 lines | GitHub preview, quick overview |
| CLAUDE.md | 250-350 | 400 lines | Fast Claude Code context loading |
| docs/core/plan.md | No limit | No limit | Living roadmap, grows with project |
| docs/core/principles.md | 50-200 | 500 lines | Core constraints, should be stable |
| Task guides (guides/) | 300-500 | 1000 lines | Single workflow, self-contained |
| Reference docs (reference/) | No limit | No limit | Completeness over brevity |

---

## Directory Structure Reference

### Minimal Structure (Phase 0)

```
project-root/
├── README.md                    # Public entry point
├── CLAUDE.md                    # Development guide
├── LICENSE
├── .gitignore
│
└── docs/
    └── core/
        ├── plan.md              # Core document #1
        └── principles.md        # Core document #2
```

### Mature Structure (After months of development)

```
project-root/
├── README.md
├── CLAUDE.md
├── CONTRIBUTING.md
├── CHANGELOG.md
│
└── docs/
    ├── DOCUMENTATION_MAP.md     # Navigation
    │
    ├── core/                    # Core documents
    │   ├── plan.md             # Roadmap
    │   └── principles.md       # Design constraints
    │
    ├── guides/                  # Task-oriented
    │   ├── integration.md
    │   ├── plugin-development.md
    │   └── troubleshooting.md
    │
    ├── reference/               # Complete specs
    │   ├── cli.md
    │   ├── features.md
    │   └── repository-structure.md
    │
    ├── tutorials/               # Step-by-step
    │   ├── examples.md
    │   ├── cookbook.md
    │   └── installation.md
    │
    ├── architecture/            # Design docs
    │   ├── adr/
    │   │   ├── README.md
    │   │   ├── template.md
    │   │   └── ADR-001-*.md
    │   └── proposals/
    │       └── *.md
    │
    ├── methodology/             # Universal guides
    │   └── documentation-management.md
    │
    └── archive/                 # Historical
        └── [outdated-docs]
```

---

## Anti-Patterns to Avoid

### 1. The Mega-README
**Problem**: README.md with 2000+ lines

**Solution**: Extract to specialized docs, link from README

### 2. Documentation Drift
**Problem**: Docs not updated with code

**Solution**: Update docs in same commit (use checklists)

### 3. Redundant Documentation
**Problem**: Same concept in 5 places

**Solution**: Single source of truth, link to it

### 4. Premature Organization
**Problem**: Complex structure before code exists

**Solution**: Start minimal (Phase 0), grow organically

### 5. No Entry Point
**Problem**: New users don't know where to start

**Solution**: README.md and CLAUDE.md serve as entry points

### 6. Skipping Core Docs
**Problem**: No plan.md or principles.md

**Solution**: Create these in Phase 0 (they're essential!)

---

## Conclusion

Effective documentation for Claude Code projects requires:

1. **Create plan.md and principles.md immediately** (Phase 0)
2. **Start minimal** (5-6 files), grow organically
3. **Synchronize at Git events** (use checklists)
4. **Follow core principles** (DRY, Progressive Disclosure, Task-Oriented)
5. **Measure and iterate** (learn from access patterns)

### Quick Start

**For new projects**:
1. Create Phase 0 files (plan.md, principles.md, CLAUDE.md, README.md)
2. Use pre-commit checklist for every code change
3. Use synchronization matrix when unsure what to update
4. Measure access patterns and optimize

**For existing projects**:
1. Audit current docs against this methodology
2. Create plan.md and principles.md if missing
3. Consolidate redundant content
4. Archive outdated docs
5. Measure and iterate

---

## References

### From meta-cc Project

- [DOCUMENTATION_MAP.md](../DOCUMENTATION_MAP.md) - Navigation example
- [CLAUDE.md](../../CLAUDE.md) - Development entry point
- [docs/core/principles.md](../core/principles.md) - Design constraints
- [docs/core/plan.md](../core/plan.md) - Roadmap example

### External Resources

- [ADR methodology](https://adr.github.io/) - Architecture Decision Records
- [Claude Code docs](https://docs.claude.com/en/docs/claude-code/overview) - Official documentation

---

**Document Version**: 5.0
**Last Updated**: 2025-10-12
**Status**: Operational manual

**Changelog**:
- v5.0 (2025-10-12): Simplified to Phase 0 + Ongoing; removed commit distinctions; consolidated all templates to Phase 0
- v4.0 (2025-10-12): Operational manual with checklists
- v3.0 (2025-10-12): Added templates and Chinese summary
- v2.0 (2025-10-12): Added iterative workflow
- v1.0 (2025-10-12): Initial version
