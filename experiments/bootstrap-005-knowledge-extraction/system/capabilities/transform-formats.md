# Transform Formats Capability

**Purpose**: Convert experiment format to Claude Code skill format (SKILL.md + templates + reference + examples + scripts).

**Version**: 1.0
**Created**: 2025-10-19

---

## Inputs

- **Extraction inventory** (JSON): From extract-knowledge capability
- **Target format specifications**: Claude Code skill structure
- **Format references**: `.claude/skills/testing-strategy/`, `.claude/skills/error-recovery/`

---

## Process

### Step 1: Create skill directory structure

```bash
mkdir -p .claude/skills/[skill-name]/{templates,reference,examples,scripts}
```

---

### Step 2: Generate SKILL.md with frontmatter

**Use template**: `knowledge/templates/skill-generation-template.md`

**Required sections** (in order):
1. Frontmatter (name, description ≤400 chars, allowed-tools)
2. Title + tagline + quote
3. When to Use (6 use cases + 4 anti-patterns)
4. Prerequisites (Tools, Concepts, Background)
5. Quick Start (30-45 min, 3-5 steps)
6. Patterns/Templates (N items with quantified effectiveness)
7. Core Principles (6-10 with evidence)
8. Success Metrics (instance + meta)
9. Transferability (language, codebase, domain)
10. Limitations and Gaps
11. Related Skills
12. Quick Reference

**Key requirements**:
- Frontmatter `description` ≤400 chars (includes: methodology, use cases, artifacts, validation)
- All cross-references use correct filenames (create files first, link second)
- All metrics quantified (no vague "good" or "excellent")

---

### Step 3: Copy templates verbatim

```bash
cp experiments/[experiment]/knowledge/templates/*.md .claude/skills/[skill-name]/templates/
```

**Verify**: File sizes match byte-for-byte (no transcription errors)

---

### Step 4: Create reference/patterns.md

**For each pattern in inventory**:
- Extract from results.md (use source line numbers)
- Format using pattern template
- Include: Context, Problem, Solution (steps), Example (code), Validation, Transferability

**Combine** all patterns into single `reference/patterns.md` file

---

### Step 5: Extract examples from iterations

**For each example candidate in inventory**:
- Read source iteration file (use line numbers)
- Convert narrative to step-by-step walkthrough
- Use template: `knowledge/templates/extraction-workflow.md` → Walkthrough Template
- Include: Context, Steps (with time), Summary (quantified results)

**Create**: `examples/[example-name]-walkthrough.md` per example

---

### Step 6: Copy automation scripts

```bash
cp experiments/[experiment]/scripts/*.sh .claude/skills/[skill-name]/scripts/
chmod +x .claude/skills/[skill-name]/scripts/*.sh
```

**Adapt** if needed:
- Generalize hardcoded paths (use parameters)
- Add usage comments
- Test on fresh data

---

### Step 7: Generate knowledge base entries (project-level)

**Create pattern entries** (`knowledge/patterns/[pattern-name].md`):
```markdown
---
name: [Pattern Name]
category: [Category]
source_experiment: [Experiment Name]
validation_status: validated
applications: [Count]
---

[Content from pattern]
```

**Create principle entries** (`knowledge/principles/[principle-name].md`):
```markdown
---
name: [Principle Name]
source_experiment: [Experiment Name]
evidence: [Brief description]
generality: [universal|domain-specific|project-specific]
---

[Content from principle]
```

---

## Outputs

- `.claude/skills/[skill-name]/SKILL.md`: Main skill document
- `.claude/skills/[skill-name]/templates/`: Copied templates
- `.claude/skills/[skill-name]/reference/patterns.md`: Pattern catalog
- `.claude/skills/[skill-name]/examples/*.md`: Walkthroughs
- `.claude/skills/[skill-name]/scripts/*.sh`: Automation scripts
- `knowledge/patterns/*.md`: Project-level pattern entries
- `knowledge/principles/*.md`: Project-level principle entries

---

## Quality Checks

- [ ] SKILL.md frontmatter complete (name, description ≤400, allowed-tools)
- [ ] All directories created
- [ ] All templates copied (file sizes match)
- [ ] All patterns formatted consistently
- [ ] All examples have time estimates
- [ ] All scripts executable
- [ ] All cross-references use correct filenames (no broken links)
- [ ] All metrics quantified (no vague descriptions)

---

## Time Estimate

20-30 minutes for complete transformation

---

**Version**: 1.0
**Last Updated**: 2025-10-19
