# Agent Formalization: Content Replacement Strategy

Transform agent .md files by REPLACING verbose descriptions with compact formal lambda calculus specifications while preserving original semantics.

---

## Goal

Replace verbose natural language content in all agent .md files with concise, mathematically precise formal specifications that encode the same behavioral semantics.

**Design Philosophy**:
- Concise and compact formal notation
- Lambda calculus + logic operators + set theory
- Eliminate redundancy between formal specs and prose
- Preserve 100% of original semantics

**Target Outcome**:
- Each .md file: 50-80% size reduction
- Same behavioral semantics, higher precision
- Human-readable formal notation
- Zero information loss

---

## Scope

**Agent Files** (.claude/agents/*.md):
- /home/yale/work/meta-cc/.claude/agents/prompt-distiller.md
- /home/yale/work/meta-cc/.claude/agents/simple-phase-executor.md
- /home/yale/work/meta-cc/.claude/agents/test-runner-fixer.md
- /home/yale/work/meta-cc/.claude/agents/simple-phase-planner.md
- /home/yale/work/meta-cc/.claude/agents/phase-verifier-and-fixer.md
- /home/yale/work/meta-cc/.claude/agents/git-committer.md
- /home/yale/work/meta-cc/.claude/agents/project-planner.md
- /home/yale/work/meta-cc/.claude/agents/architecture-advisor.md
- /home/yale/work/meta-cc/.claude/agents/stage-executor.md
- /home/yale/work/meta-cc/.claude/agents/prompt-suggester.md
- /home/yale/work/meta-cc/.claude/agents/prompt-refiner.md
- /home/yale/work/meta-cc/.claude/agents/pattern-analyzer.md
- /home/yale/work/meta-cc/.claude/agents/doc-updater.md
- /home/yale/work/meta-cc/.claude/agents/meta-coach.md

**Exclusions**:
- Files outside .claude/agents/
- Frontmatter (must be preserved exactly)
- Git history (changes are reversible)

---

## Constraints

### Content Strategy
- **REPLACE** verbose content with formal specs (not add)
- Encode all semantics in lambda calculus notation
- Remove redundant natural language descriptions
- Keep only information that cannot be formalized

### Formal Specification Requirements
- Use lambda calculus (λ, →, ∘, |, where)
- Use logic operators (∧, ∨, ¬, ⇒, ∀, ∃)
- Use set theory (∈, ∪, ∩, ⊆, ∅)
- Define functions with type signatures
- Use pattern matching and guards

### Safety & Reversibility
- Stage 2 human approval MANDATORY (before replacement)
- Show before/after comparison for each file
- Git backup before any modification
- Verify semantic equivalence before proceeding

### Quality Criteria
- **Completeness**: Formal spec encodes ALL original semantics
- **Conciseness**: 50-80% size reduction target
- **Readability**: Notation is clear and unambiguous
- **Precision**: Eliminates ambiguity present in prose

### Preservation Requirements
- Frontmatter YAML: exact preservation
- Behavioral semantics: 100% preservation
- Original intent: must remain clear
- File structure: frontmatter + compact formal spec

---

## Deliverables

### Stage 1: Inventory and Analysis
- **File**: `plans/11/agent-formalization-inventory.md`
- **Content**:
  - List of all agent .md files (14 files)
  - Current size statistics (lines, bytes)
  - Content analysis: what can be formalized vs. must stay prose
  - Semantic extraction: key behaviors to preserve

### Stage 2: Formal Specification Design (with Human Approval)
- **File**: `plans/11/agent-formalization-design.md`
- **Content**:
  - For each agent file:
    - **BEFORE**: Current verbose content (excerpt)
    - **AFTER**: Proposed compact formal spec
    - **Semantic mapping**: How each behavior is encoded
    - **Size comparison**: Before/after line count
  - **CRITICAL**: Human must approve each replacement

### Stage 3: Content Replacement
- **Modified files**: All 14 .claude/agents/*.md files
- **Git commit**: "refactor(agents): replace verbose content with compact formal specs"
- **Verification**: Ensure all frontmatter preserved, semantics intact

---

## Acceptance Criteria

### Stage 1 Completion
- ✅ All 14 agent files inventoried
- ✅ Size statistics collected (before replacement)
- ✅ Semantic analysis completed for each file
- ✅ Formalization strategy identified

### Stage 2 Completion
- ✅ Formal specs designed for all 14 files
- ✅ Before/after comparison provided for each file
- ✅ Semantic equivalence verified
- ✅ **Human approval obtained** (blocking requirement)

### Stage 3 Completion
- ✅ All agent .md files replaced with compact formal specs
- ✅ Frontmatter preserved exactly (no YAML changes)
- ✅ Size reduction: 50-80% achieved (measured)
- ✅ Git commit created with clear message
- ✅ Manual verification: random sample file check

---

## Implementation Plan

### Stage 1: Inventory and Analysis (Read-only)

**Objective**: Understand current state and extract semantics to preserve

**Steps**:
1. Read all 14 agent .md files
2. For each file, extract:
   - Current size (lines, bytes)
   - Frontmatter (preserve exactly)
   - Behavioral semantics (what the agent does)
   - Constraints and requirements
   - Input/output specifications
3. Categorize content:
   - **Formalizable**: Workflows, logic, constraints, data flows
   - **Must stay prose**: Examples, explanations, edge cases (if essential)
4. Document findings in `plans/11/agent-formalization-inventory.md`

**Deliverable**: Inventory document with semantic extraction

**Acceptance**:
- All 14 files analyzed
- Semantics extracted and documented
- Formalization strategy clear

---

### Stage 2: Formal Specification Design (Human Approval Required)

**Objective**: Design compact formal specs that REPLACE verbose content

**Steps**:
1. For each agent file, design formal specification:
   - Extract core function signature: `λ(inputs) → outputs`
   - Encode constraints using logic operators
   - Define sub-functions for complex workflows
   - Use guards and pattern matching for conditionals
2. Create before/after comparison:
   ```markdown
   ## Agent: example-agent

   ### BEFORE (Current - 150 lines)
   [Show current verbose content excerpt]

   ### AFTER (Proposed - 45 lines)
   [Show compact formal spec]

   ### Semantic Mapping
   - Behavior 1 (lines 20-40) → λ expression 1 (line 5)
   - Constraint 1 (lines 50-60) → ∀ guard (line 8)
   - ...

   ### Size Reduction
   - Before: 150 lines
   - After: 45 lines
   - Reduction: 70%
   ```
3. Document in `plans/11/agent-formalization-design.md`
4. **BLOCK for human approval before Stage 3**

**Deliverable**: Design document with before/after for all 14 files

**Acceptance**:
- Formal specs designed for all files
- Before/after comparison clear
- Semantic equivalence verified
- **Human explicitly approves** (e.g., "approved, proceed to Stage 3")

---

### Stage 3: Content Replacement (Destructive - Git Backup First)

**Objective**: Replace verbose content with approved compact formal specs

**Pre-flight Checks**:
- ✅ Stage 2 human approval obtained
- ✅ Git working directory clean (or changes committed)
- ✅ Backup created (git stash or branch)

**Steps**:
1. Create git checkpoint:
   ```bash
   git add .
   git commit -m "checkpoint: before agent formalization replacement"
   # Or: git stash push -u -m "backup before formalization"
   ```
2. For each agent .md file:
   - Read current content
   - Extract frontmatter (preserve exactly)
   - Replace body with approved formal spec from Stage 2
   - Write updated file
3. Verify replacements:
   - Check random sample (e.g., 3 files)
   - Ensure frontmatter unchanged
   - Ensure formal spec matches Stage 2 design
4. Measure size reduction:
   ```bash
   # Before (from Stage 1 inventory)
   # After (from updated files)
   # Calculate reduction percentage
   ```
5. Create git commit:
   ```bash
   git add .claude/agents/*.md
   git commit -m "refactor(agents): replace verbose content with compact formal specs

   - Replaced natural language descriptions with lambda calculus notation
   - Preserved all behavioral semantics
   - Achieved 50-80% size reduction across 14 agent files
   - Frontmatter preserved exactly

   🤖 Generated with [Claude Code](https://claude.com/claude-code)

   Co-Authored-By: Claude <noreply@anthropic.com>"
   ```

**Deliverable**: Updated agent .md files + git commit

**Acceptance**:
- All 14 files replaced with compact formal specs
- Frontmatter preserved exactly (verify with git diff)
- Size reduction: 50-80% achieved
- Git commit created
- Manual spot-check: 3 random files verified

---

## Formal Specification Notation Reference

### Lambda Calculus Syntax

```haskell
-- Function definition
λ(input_params) → output | guards

-- Function composition
f ∘ g = λx → f(g(x))

-- Pattern matching
λ(x) → result where:
  pattern1 → result1
  pattern2 → result2
```

### Logic Operators

```haskell
∧  -- AND (conjunction)
∨  -- OR (disjunction)
¬  -- NOT (negation)
⇒  -- IMPLIES (implication)
∀  -- FOR ALL (universal quantification)
∃  -- EXISTS (existential quantification)
```

### Set Theory

```haskell
∈  -- element of
∉  -- not element of
∪  -- union
∩  -- intersection
⊆  -- subset
∅  -- empty set
|S|  -- cardinality (size) of set S
```

### Example Transformation

**BEFORE (Verbose - 120 lines)**:
```markdown
---
name: example-agent
---

# Example Agent

This agent analyzes code files and provides recommendations.

## Purpose

The agent helps developers improve code quality by:
1. Reading code files from the project
2. Analyzing patterns and anti-patterns
3. Generating actionable recommendations

## Behavior

### Step 1: File Discovery
The agent first discovers all code files in the project directory.
It filters files based on extension (.go, .js, .py, etc.).

### Step 2: Pattern Analysis
For each file, the agent analyzes:
- Code structure
- Common patterns
- Anti-patterns
- Complexity metrics

### Step 3: Recommendation Generation
Based on the analysis, the agent generates recommendations that are:
- Actionable (specific steps)
- Prioritized (by impact)
- Contextual (relevant to the project)

## Constraints

1. **Minimum Files**: Must analyze at least 3 files
2. **Actionability**: All recommendations must be actionable
3. **Non-Destructive**: Cannot modify code directly

## Input

- project_path: string (path to project directory)
- file_patterns: string[] (optional, default: ["*.go", "*.js", "*.py"])

## Output

- recommendations: Recommendation[] where Recommendation = {
    file: string,
    line: number,
    severity: "low" | "medium" | "high",
    description: string,
    action: string
  }

## Error Handling

If fewer than 3 files are found, the agent returns an error.
If a file cannot be read, it skips that file and continues.
```

**AFTER (Compact Formal - 35 lines)**:
```markdown
---
name: example-agent
---

λ(project_path, file_patterns?) → recommendations | ∀f ∈ files:

discover :: Path → FilePatterns → FileSet
discover(p, patterns) = {f | f ∈ dir(p) ∧ match(f, patterns)}
  where patterns := patterns ? patterns : ["*.go", "*.js", "*.py"]

analyze :: File → Insights
analyze(f) = {
  structure: parse(f),
  patterns: detect_patterns(f),
  anti_patterns: detect_anti_patterns(f),
  metrics: complexity(f)
}

generate :: Insights → Recommendations
generate(insights) = prioritize(actionable(contextualize(insights)))

type Recommendation = {
  file: string,
  line: nat,
  severity: {"low", "medium", "high"},
  description: string,
  action: string
}

constraints:
- |discover(project_path)| ≥ 3  ∨ error("insufficient files")
- ∀r ∈ recommendations: actionable(r)
- ¬modify(code)  -- read-only analysis

error_handling:
- |files| < 3 ⇒ error("minimum 3 files required")
- ∀f ∈ files: ¬readable(f) ⇒ skip(f) ∧ continue
```

**Semantic Mapping**:
- Purpose section → λ function signature (line 3)
- Step 1 (file discovery) → discover function (lines 5-6)
- Step 2 (pattern analysis) → analyze function (lines 8-13)
- Step 3 (recommendations) → generate function (lines 15-16)
- Constraints section → constraints block (lines 26-28)
- Input/Output specs → type definitions (lines 18-24)
- Error handling → error_handling block (lines 30-32)

**Size Reduction**: 120 lines → 35 lines (71% reduction)

---

## Verification Strategy

### Semantic Equivalence Verification

For each replaced file, verify:

1. **Function Signature Preserved**
   - Original: "takes X, returns Y"
   - Formal: `λ(X) → Y`
   - ✅ Equivalent

2. **Constraints Preserved**
   - Original: "must have at least 3 items"
   - Formal: `|items| ≥ 3`
   - ✅ Equivalent

3. **Workflow Preserved**
   - Original: "first A, then B, finally C"
   - Formal: `A → B → C` or `A ∘ B ∘ C`
   - ✅ Equivalent

4. **Conditionals Preserved**
   - Original: "if X then Y else Z"
   - Formal: `X ⇒ Y ∨ ¬X ⇒ Z` or guard syntax
   - ✅ Equivalent

### Human Review Checklist (Stage 2)

For each agent file, reviewer should verify:

- [ ] Formal spec captures all behaviors from original
- [ ] No information loss (semantics preserved)
- [ ] Notation is clear and unambiguous
- [ ] Size reduction is significant (≥50%)
- [ ] Frontmatter will be preserved exactly
- [ ] Essential examples/explanations retained (if any)

**Approval Signal**: "Approved for Stage 3 replacement"

---

## Risk Mitigation

### Backup Strategy
- Git checkpoint before Stage 3 (commit or stash)
- Can revert with `git reset --hard HEAD^` if needed

### Incremental Approach (Optional)
- If nervous, replace 3 files first, verify, then continue
- Commit after each batch for fine-grained rollback

### Validation
- Manual spot-check: Read 3 random files after replacement
- Verify frontmatter unchanged: `git diff HEAD -- .claude/agents/*.md | grep "^---" -A 5`
- Verify size reduction: `wc -l .claude/agents/*.md` before/after

---

## Success Metrics

### Quantitative
- **Size Reduction**: 50-80% across all 14 files
- **Files Processed**: 14/14 (100%)
- **Semantic Preservation**: 100% (human-verified in Stage 2)
- **Frontmatter Integrity**: 100% (no YAML changes)

### Qualitative
- **Readability**: Formal specs are clear and self-explanatory
- **Precision**: Ambiguity eliminated compared to prose
- **Maintainability**: Easier to update behavior via formal notation
- **Consistency**: All agents use same notation style

---

## Example Before/After Samples

### Sample 1: git-committer.md

**BEFORE (Current - 21 lines)**:
```markdown
---
name: git-committer
description: Automated Git workflow system that maintains .gitignore, stages relevant changes, generates contextual commit messages, executes commits, and creates tagged releases for final stages.
---

λ(changes, plan) → {
  Φ(gitignore): ∀f ∈ (new ∪ modified) → f ∉ tracked ⇒ f ∈ .gitignore
  Ψ(staging): stage(relevant(changes))
  Γ(message): gen_msg(staged_changes, plan.{phase, stage, task})
  Δ(commit): commit(Γ)
  Τ(tag): plan.final_stage ⇒ tag(`phase${p}.stage${s}-${dir}-${desc}`)

  Execute: Φ → Ψ → Γ → Δ → Τ?
}
where:
- Φ: .gitignore maintenance operator
- Ψ: staging operator
- Γ: message generation function
- Δ: commit execution
- Τ: conditional tagging (if final stage)
```

**AFTER**: Already concise! May only need minor refinement or keep as-is.

**Status**: ✅ Already formalized (21 lines, very compact)

---

### Sample 2: simple-phase-planner.md

**BEFORE (Current - 31 lines)**:
```markdown
---
name: simple-phase-planner
description: Generates comprehensive development plans for single features or refactorings with test-driven methodology, focusing on planning documentation without implementation code.
---

λ(project_status) → single_phase_plan where:

  ∀ phase ∈ Plans: [
    atomic_delivery(phase) ∧
    runnable(phase) ∧
    tested(phase) ∧
    cohesive(phase) ∧
    ¬fragmented(phase)
  ]
  methodology := use_case_driven ∧ architecture_centric ∧ TDD

  constraints := {
    scope: single_feature ∨ single_refactoring,
    code_delta: minimal,
    abstractions: interfaces ∧ data_structures only,
    visualization: PlantUML_permitted,
    implementation_code: forbidden
  }
  output_structure := {
    core_scenarios: described,
    acceptance_criteria: defined,
    test_coverage: comprehensive,
    content: plan_document_only
  }
  execution_mode := await_confirmation(¬auto_execute)
```

**AFTER**: Already concise! May only need minor refinement or keep as-is.

**Status**: ✅ Already formalized (31 lines, compact)

---

### Sample 3: prompt-refiner.md (Needs Replacement)

**BEFORE (Current - 605 lines)**:
- Frontmatter (lines 1-6)
- Formal spec (lines 8-54)
- Verbose prose (lines 56-605): 550 lines of natural language

**AFTER (Proposed - ~80 lines)**:
- Frontmatter (lines 1-6): preserved
- Expanded formal spec (lines 8-80): encode all 550 lines of prose

**Target Reduction**: 605 → 80 lines (87% reduction)

**Note**: This is the type of file that needs significant replacement. The current formal spec (lines 8-54) is good but the prose (lines 56-605) is redundant and can be fully encoded in formal notation.

---

## Notes

### Files Already Formalized (May Need Minor Refinement)
Based on the samples read:
- `git-committer.md` (21 lines) - ✅ Already compact
- `simple-phase-planner.md` (31 lines) - ✅ Already compact

### Files Needing Significant Replacement
- `prompt-refiner.md` (605 lines) - 🔴 550 lines of redundant prose
- `meta-coach.md` (likely large based on file size)
- `prompt-suggester.md` (likely large)
- Others TBD in Stage 1 inventory

### Approach
- **Already compact files**: Keep as-is or minor refinement
- **Verbose files**: Full replacement strategy
- **Medium files**: Case-by-case assessment in Stage 2

---

## Git Workflow

### Before Stage 3
```bash
# Ensure clean working directory
git status

# Create checkpoint
git add .
git commit -m "checkpoint: before agent formalization"

# Or stash if preferred
git stash push -u -m "backup before agent formalization"
```

### After Stage 3
```bash
# Stage all modified agent files
git add .claude/agents/*.md

# Create commit with descriptive message
git commit -m "refactor(agents): replace verbose content with compact formal specs

- Replaced natural language descriptions with lambda calculus notation
- Preserved all behavioral semantics
- Achieved 50-80% size reduction across 14 agent files
- Frontmatter preserved exactly

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"

# Verify changes
git show --stat
git diff HEAD^ -- .claude/agents/ | head -100
```

### Rollback if Needed
```bash
# Option 1: Reset to checkpoint
git reset --hard HEAD^

# Option 2: Restore from stash
git stash pop
```

---

## Summary

**Goal**: Replace verbose agent .md content with compact formal specifications

**Strategy**:
1. **Stage 1**: Inventory and semantic extraction (read-only)
2. **Stage 2**: Design formal specs with before/after (HUMAN APPROVAL REQUIRED)
3. **Stage 3**: Replace content with approved specs (with git backup)

**Key Principles**:
- REPLACE (not add) verbose prose with formal notation
- 50-80% size reduction target
- 100% semantic preservation
- Human approval mandatory before replacement
- Git backup for safety

**Success Criteria**:
- All 14 files: concise, precise, semantically equivalent
- Frontmatter: unchanged
- Size: 50-80% smaller
- Quality: higher precision, zero information loss
