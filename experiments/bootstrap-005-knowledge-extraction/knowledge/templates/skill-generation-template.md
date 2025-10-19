# SKILL.md Generation Template

**Purpose**: Standardized template for generating Claude Code skill documents from BAIME experiments.

**Version**: 1.0
**Created**: 2025-10-19

---

## Template Structure

```markdown
---
name: [Skill Name]
description: [One-liner summary, max 400 chars]
allowed-tools: [Read, Write, Edit, Bash, Grep, Glob]
---

# [Skill Name]

**[Value proposition in one sentence]**

> [Memorable quote or principle]

---

## When to Use This Skill

Use this skill when:
- 🎯 **[Use case 1]**: [Description]
- 🔧 **[Use case 2]**: [Description]
- 🧪 **[Use case 3]**: [Description]
- 📊 **[Use case 4]**: [Description]
- 🔄 **[Use case 5]**: [Description]
- 📈 **[Use case 6]**: [Description]

**Don't use when**:
- ❌ [Anti-pattern 1]
- ❌ [Anti-pattern 2]
- ❌ [Anti-pattern 3]
- ❌ [Anti-pattern 4]

---

## Prerequisites

Before using this skill, ensure you have:

### Tools
- **[Tool 1]**: [Purpose]
  - Install: `[command]`
  - Purpose: [Usage]
- **[Tool 2]**: [Purpose]
  - Install: `[command]`
  - Purpose: [Usage]

### Concepts
- **[Concept 1]**: [Definition]
- **[Concept 2]**: [Definition]

### Background Knowledge
- [Background 1]
- [Background 2]

---

## Quick Start ([X] minutes)

**Context**: [Where to run this, prerequisites]

### Step 1: [Phase Name] ([X] min)

```bash
# [Commands]
```

**Decision Point**: [Key decision or output]

### Step 2: [Phase Name] ([X] min)

[Instructions]

**Goal**: [Outcome]

### Step 3: [Phase Name] ([X] min)

[Instructions]

**Discipline**: [Key practice]

---

## [N] [Artifact Type] (Patterns/Templates/Tools)

### 1. [Name] ([Purpose])

**Use for**: [When to use]

**Effectiveness**: [Quantified impact]

**Pattern/Process**:
1. [Step 1]
2. [Step 2]
[...]

**Example**: See [examples/example-name.md](examples/example-name.md)

### 2. [Name]

[Similar structure]

[...]

---

## [N] [Artifact Type] (Templates/Automation)

### 1. [Name]

**Purpose**: [What it does]

**Phases/Usage**:
[Description]

**Result**: [Quantified outcome]

See [templates/template-name.md](templates/template-name.md) or [scripts/script-name.sh](scripts/script-name.sh)

---

## Core Principles

### 1. [Principle Name]

**Principle**: [Statement]

**Evidence**: [Validation from experiment]

### 2. [Principle Name]

[Similar structure]

[...]

---

## Success Metrics (Validated)

**Instance Metrics** ([N] applications):
- **[Metric 1]**: [Value] ([Detail])
- **[Metric 2]**: [Value] ([Detail])
- **[Metric 3]**: [Value] ([Detail])
- **[Metric 4]**: [Value] ([Detail])

**Meta Metrics** (methodology quality):
- **[Metric 1]**: [Value] ([Detail])
- **[Metric 2]**: [Value] ([Detail])
- **[Metric 3]**: [Value] ([Detail])
- **[Metric 4]**: [Value] ([Detail])

---

## Transferability

**Language Independence**: [Percentage]%
- [Description of applicability]
- **Adaptation**: [Guidance for other languages]

**Codebase Generality**: [Percentage]%
- [Description of applicability]
- [Examples of codebase types]

**Domain Independence**: [Percentage]%
- [Description of applicability]
- [Examples of domains]

---

## Limitations and Gaps

**Known Limitations**:
1. **[Limitation 1]**: [Description]
2. **[Limitation 2]**: [Description]
[...]

**Trade-offs**:
- **[Trade-off 1]**: [Description]
- **[Trade-off 2]**: [Description]
[...]

---

## Related Skills

- **[Skill 1]**: [Relationship]
- **[Skill 2]**: [Relationship]
- **[Skill 3]**: [Relationship]
[...]

---

## Quick Reference

**[Category 1]**:
- [Item]: [Value]
- [Item]: [Value]

**[Category 2]**:
- [Item]: [Value]
- [Item]: [Value]

**[Category 3]**:
- [Item]: [Value]
- [Item]: [Value]

---

**Version**: [X.Y] ([Description])
**Created**: [Date]
**Source**: [Experiment path]
**Validation**: [Summary of validation]
```

---

## Field Specifications

### Frontmatter

#### `name`
- **Type**: String (Title Case)
- **Example**: "Code Refactoring", "Error Recovery", "Testing Strategy"
- **Source**: Experiment name or domain
- **Validation**: Matches H1 heading exactly

#### `description`
- **Type**: String (max 400 characters)
- **Format**: [Methodology summary]. Use when [use cases]. Provides [artifacts]. Validated with [metrics].
- **Example**: "Systematic code refactoring methodology using Test-Driven Refactoring, complexity reduction patterns, and incremental safety protocols. Use when refactoring complex functions (complexity >8), improving code maintainability, reducing technical debt, or systematizing ad-hoc refactoring. Provides 8 refactoring patterns (Extract Method, Simplify Conditionals, Characterization Tests, etc.), 3 safety templates (Safety Checklist, TDD Workflow, Commit Protocol), 1 automation script (complexity checking). Validated with 2 refactorings achieving 28% complexity reduction, 100% test pass rate, 0 regressions."
- **Components**:
  1. Methodology summary (1 sentence)
  2. Use cases (keywords)
  3. Artifacts provided (counts + names)
  4. Validation metrics (quantified)
- **Validation**: ≤400 chars, includes all 4 components

#### `allowed-tools`
- **Type**: Array of strings
- **Default**: `[Read, Write, Edit, Bash, Grep, Glob]`
- **Customization**: Add if skill needs specific tools (e.g., `WebFetch` for web-based skills)
- **Validation**: Array format, valid tool names

---

### Body Sections

#### When to Use This Skill
- **Format**: Bullet list with emoji icons
- **Content**: 6 use cases (positive) + 4 anti-patterns (negative)
- **Source**: Experiment objectives, use case analysis
- **Template**: `- 🎯 **[Category]**: [Description]`

#### Prerequisites
- **Subsections**: Tools, Concepts, Background Knowledge
- **Tools**: Name, install command, purpose
- **Concepts**: Name, definition (with inline or glossary definitions)
- **Background**: Skills or knowledge assumed

#### Quick Start
- **Time**: 30-45 minutes (prominently displayed)
- **Steps**: 3-5 sequential steps with time estimates
- **Format**: Step N: [Name] (X min) → Commands/Instructions → Decision Point/Goal
- **Context**: Where to run, what to expect

#### Patterns/Templates/Tools
- **Count**: Prominently displayed in heading (e.g., "Eight Refactoring Patterns")
- **Structure**: Name, Use For, Effectiveness, Pattern/Process, Example
- **Examples**: Link to examples/ directory
- **Ordering**: Most important/frequently used first

#### Core Principles
- **Count**: 6-10 principles
- **Structure**: Principle Name → Statement → Evidence
- **Source**: results.md "Principles" section
- **Validation**: Each principle backed by experiment data

#### Success Metrics
- **Two categories**: Instance Metrics (task quality) + Meta Metrics (methodology quality)
- **Format**: Bullet list with quantified values
- **Source**: V_instance and V_meta calculations from experiment

#### Transferability
- **Three dimensions**: Language Independence, Codebase Generality, Domain Independence
- **Format**: Percentage + description + adaptation guidance
- **Source**: Reusability assessment from results.md

#### Limitations and Gaps
- **Two categories**: Known Limitations + Trade-offs
- **Purpose**: Honest assessment of skill boundaries
- **Source**: "Limitations" section from results.md, gaps from iteration reports

#### Quick Reference
- **Purpose**: Cheat sheet for frequent reference
- **Content**: Thresholds, time estimates, key commands
- **Format**: Categorized bullet lists or tables

---

## Content Sources

| Section | Primary Source | Fallback Source |
|---------|----------------|-----------------|
| Frontmatter (name, description) | Experiment objectives | Infer from domain |
| When to Use | Experiment objectives, use case analysis | Infer from patterns |
| Prerequisites | Templates (tool requirements), Concepts | Infer from examples |
| Quick Start | Iteration reports (execution), Examples | Create from patterns |
| Patterns/Templates | results.md "Patterns" section | iteration reports |
| Core Principles | results.md "Principles" section | results.md "Learnings" |
| Success Metrics | results.md "Value Functions", iteration reports | Calculate from data |
| Transferability | results.md "Reusability" section | Assess manually |
| Limitations | results.md "Limitations", iteration reports | Extract from gaps |
| Related Skills | Explicit cross-references | Infer from domain |
| Quick Reference | Templates, iteration reports (time data) | Aggregate from patterns |

---

## Validation Checklist

Before finalizing SKILL.md:

### Frontmatter
- [ ] `name` is Title Case?
- [ ] `name` matches H1 heading exactly?
- [ ] `description` ≤400 characters?
- [ ] `description` includes: methodology, use cases, artifacts, validation metrics?
- [ ] `allowed-tools` is valid array?

### Structure
- [ ] All required sections present?
- [ ] Section order matches template?
- [ ] Heading levels consistent (H2 for main sections, H3 for subsections)?

### Content
- [ ] Quick Start has time estimate?
- [ ] Patterns have quantified effectiveness?
- [ ] Principles have evidence?
- [ ] Metrics are quantified (not vague)?
- [ ] Examples are linked (not inline)?

### Links
- [ ] All cross-references valid?
- [ ] Examples exist?
- [ ] Templates exist?
- [ ] Scripts exist?

### Completeness
- [ ] All patterns from experiment included?
- [ ] All principles from experiment included?
- [ ] All templates referenced?
- [ ] All scripts referenced?

---

## Common Issues

### Issue 1: Description Too Long
**Problem**: Description exceeds 400 character limit

**Fix**:
- Remove redundant words ("very", "really", "quite")
- Use abbreviations (e.g., "TDD" instead of "Test-Driven Development")
- Condense artifact lists (e.g., "8 patterns (Extract Method, etc.)" instead of listing all)
- Prioritize most important information

**Example**:
❌ "This comprehensive and detailed code refactoring methodology provides systematic approaches using Test-Driven Development principles..."
✅ "Systematic code refactoring using TDD, complexity reduction patterns, incremental safety protocols..."

---

### Issue 2: Broken Links
**Problem**: Links point to non-existent files

**Fix**:
- Create files BEFORE linking
- Use exact filenames (case-sensitive)
- Verify links after creation: `find . -name "*.md" -exec grep -l "broken-link.md" {} \;`

---

### Issue 3: Vague Metrics
**Problem**: "Good success rate" instead of "100% (5/5)"

**Fix**:
- Always quantify: percentage, ratio, absolute numbers
- Include denominator: "2/2", "5/5", "10/10"
- Provide context: "100% (5/5 applications successful, 0 failures)"

---

### Issue 4: Missing Prerequisites
**Problem**: Skill assumes knowledge without stating

**Fix**:
- List ALL tools needed (with install commands)
- Define ALL concepts (inline or glossary)
- State ALL background knowledge assumptions

---

## Time Estimates

| Task | Time Estimate |
|------|---------------|
| Draft frontmatter | 5 min |
| Write "When to Use" | 5 min |
| Write "Prerequisites" | 5 min |
| Write "Quick Start" | 10 min |
| Document patterns/templates | 15 min |
| Document principles | 5 min |
| Extract success metrics | 5 min |
| Assess transferability | 5 min |
| Document limitations | 5 min |
| Create quick reference | 5 min |
| Validate and fix links | 5 min |
| **Total** | **65-75 min** |

**Optimization**: Draft all sections first, then validate and fix links at end

---

## Example: Well-Formatted SKILL.md Header

```markdown
---
name: Code Refactoring
description: Systematic code refactoring methodology using Test-Driven Refactoring, complexity reduction patterns, and incremental safety protocols. Use when refactoring complex functions (complexity >8), improving code maintainability, reducing technical debt, or systematizing ad-hoc refactoring. Provides 8 refactoring patterns (Extract Method, Simplify Conditionals, Characterization Tests, etc.), 3 safety templates (Safety Checklist, TDD Workflow, Commit Protocol), 1 automation script (complexity checking). Validated with 2 refactorings achieving 28% complexity reduction, 100% test pass rate, 0 regressions.
allowed-tools: [Read, Write, Edit, Bash, Grep, Glob]
---

# Code Refactoring

**Transform risky, ad-hoc refactoring into safe, systematic complexity reduction with TDD discipline.**

> Tests are your safety net. Incremental commits are your rollback points. Complexity metrics are your guide.

---
```

**Why this is good**:
✅ Frontmatter complete (name, description, allowed-tools)
✅ Description ≤400 chars (exact: 398 chars)
✅ Description includes all 4 components
✅ H1 matches frontmatter `name` exactly
✅ Value proposition clear and memorable
✅ Quote reinforces core principles

---

**Version**: 1.0
**Last Updated**: 2025-10-19
**Validated**: Bootstrap-005 Iteration 0-1 (code-refactoring SKILL.md successfully generated)
