# Validate Artifacts Capability

**Purpose**: Quality assurance for extracted Claude Code skills and knowledge base entries.

**Version**: 1.0
**Created**: 2025-10-19

---

## Inputs

- **Transformed artifacts**: `.claude/skills/[skill-name]/`, `knowledge/`
- **Source material**: `experiments/[experiment]/`
- **Validation criteria**: Completeness, accuracy, format, usability

---

## Process

### Step 1: Completeness check

**Use checklist**: `knowledge/templates/validation-checklist.md` → Section 1

**Automated counts**:
```bash
# Patterns
grep -c "^### Pattern:" experiments/[experiment]/results.md  # Available
grep -c "^### [0-9]\. " .claude/skills/[skill]/reference/patterns.md  # Extracted

# Templates
ls experiments/[experiment]/knowledge/templates/*.md | wc -l  # Available
ls .claude/skills/[skill]/templates/*.md | wc -l  # Copied

# Examples
ls .claude/skills/[skill]/examples/*.md | wc -l  # Created

# Scripts
ls experiments/[experiment]/scripts/*.sh | wc -l  # Available
ls .claude/skills/[skill]/scripts/*.sh | wc -l  # Copied
```

**Calculate**: V_completeness score using formula from validation checklist

---

### Step 2: Accuracy check

**Sample patterns** (5 random or 20% of total):
```bash
grep "^### [0-9]\. " .claude/skills/[skill]/reference/patterns.md | shuf -n 5
```

**Compare** each sampled pattern to source in `results.md`:
- Identical → 1.0
- Minor differences → 0.8
- Significant differences → 0.5
- Wrong → 0.0

**Check code examples**: Syntax valid? (manual inspection or linter)

**Check metrics**: Numbers match source? (cross-reference with results.md)

**Check links**: All resolve?
```bash
for link in $(grep -oh '(.*\.md)' .claude/skills/[skill]/*.md | tr -d '()'); do
  test -f ".claude/skills/[skill]/$link" && echo "OK: $link" || echo "BROKEN: $link"
done
```

**Calculate**: V_accuracy score using formula from validation checklist

---

### Step 3: Format check

**Frontmatter**:
```bash
head -n 5 .claude/skills/[skill]/SKILL.md | grep -E "^name:|^description:|^allowed-tools:"
```
- All 3 fields present?
- Description ≤400 chars?

**Directory structure**:
```bash
ls -la .claude/skills/[skill]/
```
- Matches reference skills (testing-strategy)?

**Markdown syntax** (if markdownlint available):
```bash
markdownlint .claude/skills/[skill]/*.md .claude/skills/[skill]/**/*.md
```

**Naming conventions**:
```bash
find .claude/skills/[skill]/ -name "*[A-Z]*"  # Should be empty (lowercase/kebab-case)
```

**Calculate**: V_format score using formula from validation checklist

---

### Step 4: Usability check

**Quick Start execution**:
- Execute from scratch (fresh terminal)
- Time completion
- Document friction points

**Examples runnability**:
- Attempt to execute code examples
- Note: Success / Fail / N/A (conceptual)

**Documentation clarity**:
- All terms defined?
- Prerequisites listed?
- Steps numbered?
- Outcomes stated?

**Calculate**: V_usability score using formula from validation checklist

---

### Step 5: Calculate overall V_instance

**Formula**:
```
V_instance = 0.3×V_completeness + 0.3×V_accuracy + 0.2×V_usability + 0.2×V_format
```

**Threshold**: ≥0.85 (convergence target)

---

### Step 6: Generate validation report

**Template**:
```markdown
# Validation Report: [Skill Name]

**Date**: [YYYY-MM-DD]
**Validator**: [Name/Agent]

## Summary

- **V_instance**: [X.XX] (Target: ≥0.85)
- **Status**: [PASS/FAIL]

## Component Scores

- **V_completeness**: [X.XX] ([Tier])
- **V_accuracy**: [X.XX] ([Tier])
- **V_usability**: [X.XX] ([Tier])
- **V_format**: [X.XX] ([Tier])

## Issues Found

### Critical (High Priority)
1. [Issue description] - Impact: [X] - Location: [File:line]

### High (Medium Priority)
1. [Issue description] - Impact: [X] - Location: [File:line]

### Medium (Low Priority)
1. [Issue description] - Impact: [X] - Location: [File:line]

## Remediation Checklist

- [ ] Fix issue 1: [Description]
- [ ] Fix issue 2: [Description]

## Recommendation

[APPROVE for use / REVISE and revalidate]
```

---

## Outputs

- **Validation report**: `data/validation-reports/[skill-name]-validation.md`
- **V_instance score**: [X.XX]
- **Pass/fail status**: [PASS if ≥0.85, FAIL if <0.85]
- **Remediation list**: Issues to fix, prioritized by impact

---

## Quality Checks

- [ ] All validation checks executed
- [ ] All scores calculated using formulas (not estimated)
- [ ] All issues documented with specific locations
- [ ] Evidence for scores provided (counts, samples, test results)
- [ ] Honest assessment (no score inflation)

---

## Time Estimate

10-15 minutes for validation (excluding Quick Start execution which takes 30-45 min)

---

**Version**: 1.0
**Last Updated**: 2025-10-19
