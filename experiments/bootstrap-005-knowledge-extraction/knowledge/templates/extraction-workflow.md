# Knowledge Extraction Workflow Template

**Purpose**: Systematic process for extracting knowledge from BAIME experiment artifacts into Claude Code skills and knowledge base entries.

**Version**: 1.0
**Created**: 2025-10-19
**Source**: Bootstrap-005, Iteration 1

---

## Overview

This template provides a step-by-step process for extracting patterns, principles, templates, and automation scripts from completed BAIME experiments and transforming them into reusable Claude Code skills.

**Time Estimate**: 60-90 minutes for complete extraction (vs 390 min ad-hoc)

---

## Phase 1: Planning (10-15 minutes)

### Step 1: Create Extraction Inventory

**What**: Structured catalog of all extractable knowledge in experiment

**How**:
1. Read `results.md` thoroughly (focus on Patterns, Principles, Templates sections)
2. Scan `iterations/*.md` for evolution trajectory and key decisions
3. List all knowledge assets in `data/extraction-inventory.json`

**Template** (JSON):
```json
{
  "experiment": "experiment-name",
  "patterns": [
    {
      "name": "Pattern Name",
      "source_file": "results.md",
      "source_lines": "309-325",
      "validated": true,
      "applications_count": 5
    }
  ],
  "principles": [
    {
      "name": "Principle Name",
      "source_file": "results.md",
      "source_lines": "366-380",
      "evidence": "Validation evidence description"
    }
  ],
  "templates": [
    {
      "filename": "template-name.md",
      "location": "knowledge/templates/",
      "lines": 276,
      "purpose": "Template purpose"
    }
  ],
  "scripts": [
    {
      "filename": "script-name.sh",
      "location": "scripts/",
      "lines": 91,
      "purpose": "Script purpose"
    }
  ],
  "examples": [
    {
      "source": "iteration-2.md",
      "narrative": "Extract Method refactoring",
      "suitable_for_walkthrough": true
    }
  ],
  "gaps": [
    "Missing X",
    "Incomplete Y"
  ]
}
```

**Output**: `data/extraction-inventory.json`

---

## Phase 2: Skill Generation (20-30 minutes)

### Step 2: Create SKILL.md with Frontmatter

**What**: Main skill document with frontmatter, overview, and quick start

**Reference**: Study existing skills (`.claude/skills/testing-strategy/SKILL.md`, `.claude/skills/error-recovery/SKILL.md`)

**Frontmatter Template**:
```markdown
---
name: [Skill Name]
description: [1-2 sentence summary, max 400 chars. Include: methodology, when to use, what it provides, validation metrics]
allowed-tools: [Read, Write, Edit, Bash, Grep, Glob] # Default set
---
```

**Required Sections**:
1. **Title** (H1): Same as `name` in frontmatter
2. **Tagline**: One-sentence value proposition
3. **When to Use This Skill**: Bullet list of use cases
4. **Don't use when**: Bullet list of anti-patterns
5. **Prerequisites**: Tools, concepts, background knowledge
6. **Quick Start**: Time-boxed walkthrough (30-45 min target)
7. **Patterns/Templates Overview**: High-level catalog
8. **Core Principles**: 6-10 principles with evidence
9. **Success Metrics**: Validated metrics (instance + meta)
10. **Transferability**: Language independence, codebase generality, domain independence
11. **Limitations and Gaps**: Known limitations, trade-offs
12. **Related Skills**: Cross-references
13. **Quick Reference**: Cheat sheet (thresholds, times, commands)

**Output**: `.claude/skills/[skill-name]/SKILL.md`

---

### Step 3: Copy Templates

**What**: Transfer templates verbatim from experiment to skill

**How**:
```bash
cp experiments/[experiment-name]/knowledge/templates/*.md .claude/skills/[skill-name]/templates/
```

**Verification**:
- ✅ File sizes match (byte-for-byte copy)
- ✅ All templates referenced in SKILL.md exist
- ✅ No broken internal links

**Output**: `.claude/skills/[skill-name]/templates/`

---

### Step 4: Create Reference Documentation

**What**: Pattern catalog and methodology documentation

**Structure**:
- `reference/patterns.md`: All patterns with code examples, validation data
- `reference/principles.md` (optional): Detailed principle explanations
- `reference/methodology.md` (optional): Deep-dive into process

**Pattern Template**:
```markdown
### Pattern: [Name]

**Context**: [When to use]

**Problem**: [What issue it solves]

**Solution**: [Step-by-step]

**Example**:
```[language]
// Before
[code]

// After
[code]
```

**Safety**: [Precautions]

**Metrics**: [Quantified impact]

**Transferability**: [Language independence]

**Validation**: [Evidence from experiment]
```

**Output**: `.claude/skills/[skill-name]/reference/`

---

### Step 5: Extract Examples

**What**: Convert iteration narratives into step-by-step walkthroughs

**Source**: `iterations/*.md` (look for detailed execution sections)

**Walkthrough Template**:
```markdown
# [Pattern Name]: Complete Walkthrough

**Example**: [Specific instance]

**Source**: [Experiment], [Iteration]

**Duration**: [Actual time]

**Outcome**: [Quantified results]

---

## Context
[Initial state, goals]

## Step 1: [Phase Name] (X minutes)
[Detailed steps with commands, output, decisions]

## Step 2: [Phase Name] (X minutes)
[...]

## Summary
**Total Time**: [Total]
**Breakdown**: [Per step]
**Commits**: [Count]
**Results**: [Quantified outcomes]
**Patterns Applied**: [List]
**Lessons Learned**: [Insights]
```

**Output**: `.claude/skills/[skill-name]/examples/`

---

### Step 6: Copy Automation Scripts

**What**: Transfer automation scripts with adaptation if needed

**How**:
```bash
cp experiments/[experiment-name]/scripts/*.sh .claude/skills/[skill-name]/scripts/
chmod +x .claude/skills/[skill-name]/scripts/*.sh
```

**Adaptation Guidelines**:
- Generalize hardcoded paths (use parameters or environment variables)
- Add usage comments
- Test on fresh data (not development data)

**Output**: `.claude/skills/[skill-name]/scripts/`

---

## Phase 3: Knowledge Base Entries (10-15 minutes)

### Step 7: Create Pattern Entries

**What**: Individual markdown files for each pattern in project knowledge base

**Location**: `knowledge/patterns/`

**Frontmatter Template**:
```markdown
---
name: [Pattern Name]
category: [Category] # e.g., refactoring, testing, error-handling
source_experiment: [Experiment Name]
validation_status: validated | preliminary
applications: [Count]
---
```

**Output**: `knowledge/patterns/[pattern-name].md`

---

### Step 8: Create Principle Entries

**What**: Individual markdown files for each principle in project knowledge base

**Location**: `knowledge/principles/`

**Frontmatter Template**:
```markdown
---
name: [Principle Name]
source_experiment: [Experiment Name]
evidence: [Brief description]
generality: universal | domain-specific | project-specific
---
```

**Output**: `knowledge/principles/[principle-name].md`

---

## Phase 4: Validation (10-15 minutes)

### Step 9: Completeness Check

**Checklist**:
- [ ] All patterns from inventory extracted? (Count: extracted / available)
- [ ] All principles from inventory extracted? (Count: extracted / available)
- [ ] All templates copied? (3/3, 4/4, etc.)
- [ ] At least 1 example walkthrough created?
- [ ] All scripts included? (Count: included / available)
- [ ] All cross-references valid? (Check links)

**Tool** (automated):
```bash
# Count patterns in results.md
grep -c "^### Pattern:" experiments/[experiment]/results.md

# Count patterns in skill
grep -c "^### Pattern:" .claude/skills/[skill]/reference/patterns.md
```

**Output**: `data/validation-reports/completeness-check.md`

---

### Step 10: Accuracy Check

**Checklist**:
- [ ] Pattern descriptions match source? (Sample 5 patterns, compare)
- [ ] Code examples syntactically correct? (Run linter if possible)
- [ ] Metrics data matches source? (Cross-check numbers)
- [ ] Cross-references valid? (All links resolve)

**Sampling**:
- Sample size: 5 patterns (or 20%, whichever is larger)
- Method: Random selection (use `shuf` or manual)
- Verification: Side-by-side comparison with source

**Output**: `data/validation-reports/accuracy-check.md`

---

### Step 11: Format Check

**Checklist**:
- [ ] Frontmatter complete? (All required fields present)
- [ ] Directory structure matches conventions? (Compare to reference skills)
- [ ] Markdown syntax valid? (Run `markdownlint` if available)
- [ ] Naming conventions followed? (kebab-case files, lowercase dirs)

**Automated checks**:
```bash
# Check frontmatter
head -n 5 .claude/skills/[skill]/SKILL.md | grep -E "^name:|^description:|^allowed-tools:"

# Check directory structure
ls -la .claude/skills/[skill]/

# Check naming conventions
find .claude/skills/[skill]/ -name "*[A-Z]*" # Should be empty (lowercase only)
```

**Output**: `data/validation-reports/format-check.md`

---

### Step 12: Usability Check

**Checklist**:
- [ ] Quick Start executable in stated time? (Test with fresh perspective)
- [ ] Examples runnable? (Execute code examples if possible)
- [ ] Documentation clarity? (All terms defined, prerequisites listed, steps numbered, outcomes stated)

**Test**:
- Time Quick Start execution (target: ≤30-45 min)
- Identify friction points (unclear steps, missing context, broken examples)
- Document improvements needed

**Output**: `data/validation-reports/usability-check.md`

---

## Phase 5: Finalization (5-10 minutes)

### Step 13: Calculate V_instance

**Formula**:
```
V_instance = 0.3×V_completeness + 0.3×V_accuracy + 0.2×V_usability + 0.2×V_format
```

**Data Collection**:
- V_completeness: (extracted_count / available_count) per category
- V_accuracy: (correct_samples / total_samples)
- V_usability: Quick Start time score + examples runnable + clarity checks
- V_format: Frontmatter complete + structure correct + syntax valid + naming conventions

**Output**: `data/baseline-value-calculation.md`

---

### Step 14: Document Problems and Gaps

**Categories**:
1. **Time-consuming steps**: Which steps took longer than expected?
2. **Error-prone steps**: Where did errors occur?
3. **Automation opportunities**: What could be automated?
4. **Systematization needs**: What needs better process definition?

**Template**:
```markdown
## Problems Encountered

### Time-Consuming
1. [Problem]: [Description]
   - Time: [Actual] vs [Expected]
   - Root cause: [Analysis]
   - Opportunity: [Potential improvement]

### Error-Prone
[Similar format]

### Automation Opportunities
[List]

### Systematization Needs
[List]
```

**Output**: Included in iteration report

---

## Best Practices

**Do**:
✅ Create extraction inventory FIRST (provides roadmap)
✅ Study reference skills for format (testing-strategy, error-recovery)
✅ Copy templates verbatim (preserve accuracy)
✅ Sample randomly for accuracy checks (avoid cherry-picking)
✅ Test Quick Start with fresh perspective (avoid author bias)
✅ Document all gaps honestly (low scores are acceptable if accurate)

**Don't**:
❌ Skip inventory step (leads to ad-hoc wandering)
❌ Write links before creating targets (causes broken links)
❌ Inflate scores (honest assessment critical for methodology improvement)
❌ Assume behavior (verify with source material)
❌ Create files before understanding naming conventions (causes rework)

---

## Anti-Patterns to Avoid

1. **Skipping Inventory**: Extracting without plan → incomplete extraction
2. **Premature Linking**: Writing cross-references before creating files → broken links
3. **Optimistic Estimation**: Assuming completeness without counting → inflated scores
4. **Author Bias**: Testing usability yourself → missing real friction points
5. **Format Deviation**: Ignoring reference skills → inconsistent structure

---

## Time Estimates by Phase

| Phase | Expected Time | Critical Path |
|-------|---------------|---------------|
| Planning (Inventory) | 10-15 min | Yes (guides all work) |
| SKILL.md Creation | 20-30 min | Yes (core artifact) |
| Templates Copy | <5 min | No (simple copy) |
| Reference Docs | 15-20 min | Yes (patterns are key) |
| Examples | 15-20 min per example | Partial (1 required, 2+ better) |
| Scripts Copy | <5 min | No (simple copy) |
| Knowledge Base | 10-15 min | No (project-level, not skill) |
| Validation | 10-15 min | Yes (quality assurance) |
| Finalization | 5-10 min | Yes (metrics calculation) |
| **Total** | **60-90 min** | |

---

## Success Indicators

**You're on track if**:
✅ Extraction inventory created in <15 minutes
✅ SKILL.md frontmatter matches reference skills exactly
✅ All templates copied without errors (file sizes match)
✅ At least 1 example walkthrough created
✅ Validation finds <5 issues

**You're off track if**:
❌ Inventory took >20 minutes (too detailed, simplify)
❌ SKILL.md frontmatter different from references (restart with template)
❌ Cross-references broken (created links before targets)
❌ No examples created (examples critical for usability)
❌ Validation finds >10 issues (extraction quality poor)

---

**Version**: 1.0
**Last Updated**: 2025-10-19
**Validated**: Bootstrap-005 Iteration 0 (baseline), Iteration 1 (systematization)
