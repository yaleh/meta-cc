# Knowledge Extraction Validation Checklist

**Purpose**: Quality assurance checklist for extracted Claude Code skills and knowledge base entries.

**Version**: 1.0
**Created**: 2025-10-19

---

## Overview

This checklist ensures extracted knowledge is:
- **Complete**: All patterns, principles, templates, scripts extracted
- **Accurate**: Content matches source, no transcription errors
- **Usable**: Quick Start works, examples runnable, documentation clear
- **Standard-compliant**: Format, structure, naming follow conventions

**Usage**: Execute all checks after extraction, before finalizing

---

## 1. Completeness Check

### 1.1 Pattern Extraction

- [ ] **Count patterns in source**: `grep -c "^### Pattern:" experiments/[experiment]/results.md`
- [ ] **Count patterns in skill**: `grep -c "^### [0-9]\. " .claude/skills/[skill]/reference/patterns.md`
- [ ] **Match**: Extracted count / Available count = [X]/[Y] = [Z]%
- [ ] **Target**: ≥90% (acceptable: ≥80%, poor: <80%)

**Gaps** (if any):
- Missing pattern 1: [Name] - Reason: [Why not extracted]
- Missing pattern 2: [Name] - Reason: [Why not extracted]

---

### 1.2 Principle Extraction

- [ ] **Count principles in source**: `grep -c "^### Principle:" experiments/[experiment]/results.md`
- [ ] **Count principles in skill**: `grep -c "^### [0-9]\. " .claude/skills/[skill]/SKILL.md | grep -A2 "Core Principles"`
- [ ] **Match**: Extracted count / Available count = [X]/[Y] = [Z]%
- [ ] **Target**: ≥90%

**Gaps** (if any):
- Missing principle 1: [Name] - Reason: [Why not extracted]

---

### 1.3 Template Copy

- [ ] **Count templates in source**: `ls experiments/[experiment]/knowledge/templates/*.md | wc -l`
- [ ] **Count templates in skill**: `ls .claude/skills/[skill]/templates/*.md | wc -l`
- [ ] **Match**: Copied count / Available count = [X]/[Y] = [Z]%
- [ ] **Target**: 100% (all templates should be copied)

**Verification**: File sizes match byte-for-byte
```bash
diff -q experiments/[experiment]/knowledge/templates/*.md .claude/skills/[skill]/templates/
```

---

### 1.4 Example Creation

- [ ] **Count potential examples in iterations**: Review `iterations/*.md` for detailed executions
- [ ] **Count created examples**: `ls .claude/skills/[skill]/examples/*.md | wc -l`
- [ ] **Match**: Created count / Potential count = [X]/[Y]
- [ ] **Target**: ≥1 (minimum), ≥2 (good), ≥3 (excellent)

**Examples**:
- Example 1: [Name] - Source: [iteration-X.md] - Status: [Created/Not created]
- Example 2: [Name] - Source: [iteration-Y.md] - Status: [Created/Not created]

---

### 1.5 Script Copy

- [ ] **Count scripts in source**: `ls experiments/[experiment]/scripts/*.sh | wc -l`
- [ ] **Count scripts in skill**: `ls .claude/skills/[skill]/scripts/*.sh | wc -l`
- [ ] **Match**: Copied count / Available count = [X]/[Y] = [Z]%
- [ ] **Target**: 100%

**Verification**: Scripts are executable
```bash
test -x .claude/skills/[skill]/scripts/*.sh && echo "OK" || echo "FAIL: not executable"
```

---

### Completeness Score

**Formula**:
```
V_completeness = (
  0.25 × (patterns_extracted / patterns_available) +
  0.25 × (principles_extracted / principles_available) +
  0.20 × (templates_present ? 1.0 : 0.0) +
  0.15 × (examples_count ≥ 1 ? 1.0 : examples_count * 0.5) +
  0.15 × (scripts_included / scripts_available)
)
```

**Calculation**:
- Patterns: [X]/[Y] = [Z] × 0.25 = [A]
- Principles: [X]/[Y] = [Z] × 0.25 = [B]
- Templates: [Present? 1.0 : 0.0] × 0.20 = [C]
- Examples: [Count ≥1 ? 1.0 : Count*0.5] × 0.15 = [D]
- Scripts: [X]/[Y] = [Z] × 0.15 = [E]

**V_completeness** = [A] + [B] + [C] + [D] + [E] = **[Score]**

**Scoring Guide**:
- 1.0: Complete extraction (100% patterns, 100% principles, all templates, ≥2 examples, all scripts)
- 0.8: Near complete (≥90% patterns, ≥90% principles, all templates, ≥1 example, ≥80% scripts)
- 0.6: Adequate (≥70% patterns, ≥70% principles, templates present, ≥1 example, ≥60% scripts)
- 0.4: Partial (≥50% patterns, ≥50% principles, some templates, basic example, ≥40% scripts)
- 0.2: Minimal (≥30% patterns, ≥30% principles, minimal templates, no examples, ≥20% scripts)

---

## 2. Accuracy Check

### 2.1 Pattern Description Accuracy

**Method**: Sample 5 random patterns (or 20% if <25 patterns), compare to source

**Sampling**:
```bash
# List all patterns
grep "^### [0-9]\. " .claude/skills/[skill]/reference/patterns.md

# Random sample (5 patterns)
grep "^### [0-9]\. " .claude/skills/[skill]/reference/patterns.md | shuf -n 5
```

**Verification** (per pattern):
- [ ] **Pattern 1**: [Name]
  - Source: [File, lines]
  - Extracted: [File, lines]
  - Match: [Identical / Minor differences / Significant differences / Wrong]
  - Score: [1.0 / 0.8 / 0.5 / 0.0]

- [ ] **Pattern 2**: [Name] - Score: [X]
- [ ] **Pattern 3**: [Name] - Score: [X]
- [ ] **Pattern 4**: [Name] - Score: [X]
- [ ] **Pattern 5**: [Name] - Score: [X]

**Average**: Sum of scores / 5 = **[pattern_description_accuracy]**

---

### 2.2 Code Example Correctness

**Method**: Check all code blocks for syntax validity

**Automated check**:
```bash
# Extract all code blocks
grep -A20 "^```" .claude/skills/[skill]/SKILL.md | grep -v "^```" > /tmp/code-blocks.txt

# Manual inspection or syntax check (language-specific)
# For Go: go fmt <file>
# For Python: python -m py_compile <file>
```

**Manual verification**:
- [ ] **Code block 1** (line X): Syntax valid? [Yes/No]
- [ ] **Code block 2** (line Y): Syntax valid? [Yes/No]
- [ ] **Code block 3** (line Z): Syntax valid? [Yes/No]
[...]

**Score**: (correct_examples / total_examples) = [X]/[Y] = **[code_example_correctness]**

---

### 2.3 Metrics Data Accuracy

**Method**: Extract numeric metrics from skill, compare to `results.md` source

**Metrics to verify**:
- [ ] **Complexity reduction**: Skill: [X]%, Source: [Y]% - Match: [Yes/No]
- [ ] **Coverage improvement**: Skill: [X]%, Source: [Y]% - Match: [Yes/No]
- [ ] **Success rate**: Skill: [X]%, Source: [Y]% - Match: [Yes/No]
- [ ] **Time estimates**: Skill: [X] min, Source: [Y] min - Match: [Yes/No]
- [ ] **Speedup**: Skill: [X]x, Source: [Y]x - Match: [Yes/No]

**Score**: (matching_metrics / total_metrics) = [X]/[Y] = **[metrics_data_accuracy]**

---

### 2.4 Cross-Reference Validity

**Method**: Extract all internal links, check if targets exist

**Automated check**:
```bash
# Extract all markdown links
grep -o '\[.*\](.*\.md)' .claude/skills/[skill]/SKILL.md .claude/skills/[skill]/reference/*.md .claude/skills/[skill]/examples/*.md

# Check if targets exist
for link in $(grep -oh '(.*\.md)' *.md | tr -d '()'); do
  test -f "$link" && echo "OK: $link" || echo "BROKEN: $link"
done
```

**Manual verification**:
- [ ] **Link 1**: [Link text] → [Target] - Exists: [Yes/No]
- [ ] **Link 2**: [Link text] → [Target] - Exists: [Yes/No]
- [ ] **Link 3**: [Link text] → [Target] - Exists: [Yes/No]
[...]

**Score**: (valid_links / total_links) = [X]/[Y] = **[cross_reference_validity]**

---

### Accuracy Score

**Formula**:
```
V_accuracy = (
  0.35 × pattern_description_accuracy +
  0.25 × code_example_correctness +
  0.25 × metrics_data_accuracy +
  0.15 × cross_reference_validity
)
```

**Calculation**:
- Pattern descriptions: [X] × 0.35 = [A]
- Code examples: [Y] × 0.25 = [B]
- Metrics data: [Z] × 0.25 = [C]
- Cross-references: [W] × 0.15 = [D]

**V_accuracy** = [A] + [B] + [C] + [D] = **[Score]**

**Scoring Guide**:
- 1.0: Perfect accuracy (all checks 100%)
- 0.9: Near perfect (≥95% across all checks)
- 0.8: High accuracy (≥90% across all checks)
- 0.7: Good accuracy (≥80% across all checks)
- 0.5: Moderate accuracy (≥70% across all checks)
- 0.3: Low accuracy (≥50% across all checks)

---

## 3. Usability Check

### 3.1 Quick Start Execution Test

**Method**: Execute Quick Start from scratch, time completion

**Setup**:
- Use fresh terminal session (simulate new user)
- Have no prior context (don't assume knowledge)
- Follow steps exactly as written

**Execution**:
- [ ] **Start time**: [HH:MM]
- [ ] **Step 1**: Completable? [Yes/No] - Time: [X] min - Issues: [None/List]
- [ ] **Step 2**: Completable? [Yes/No] - Time: [X] min - Issues: [None/List]
- [ ] **Step 3**: Completable? [Yes/No] - Time: [X] min - Issues: [None/List]
- [ ] **End time**: [HH:MM]
- [ ] **Total time**: [X] min

**Friction Points**:
1. [Issue description] - Location: [Step/line] - Impact: [High/Medium/Low]
2. [Issue description] - Location: [Step/line] - Impact: [High/Medium/Low]

**Score**:
- 1.0 if ≤30 min
- 0.75 if ≤45 min
- 0.5 if ≤60 min
- 0.25 if ≤90 min
- 0.0 if >90 min

**quick_start_works** = **[Score]**

---

### 3.2 Examples Runnability Test

**Method**: Attempt to execute all code examples

**Examples**:
- [ ] **Example 1**: [Name]
  - File: [Path]
  - Executable: [Yes/No/Partial]
  - Errors: [None/List]
  - Status: [Success/Fail/N/A (conceptual only)]

- [ ] **Example 2**: [Name] - Status: [Success/Fail/N/A]
- [ ] **Example 3**: [Name] - Status: [Success/Fail/N/A]

**Score**: (successful_examples / total_examples) = [X]/[Y] = **[examples_runnable]**

**Note**: Conceptual examples (not meant to be executed) count as "Success" if clearly labeled

---

### 3.3 Documentation Clarity Test

**Method**: Check if documentation is self-contained and clear

**Checks**:
- [ ] **All terms defined**: No undefined jargon? [Yes/No]
  - Undefined terms: [List] or [None]

- [ ] **Prerequisites listed**: Tools, concepts, background stated? [Yes/No]
  - Missing prerequisites: [List] or [None]

- [ ] **Steps numbered**: Sequential, unambiguous steps? [Yes/No]
  - Unclear steps: [List] or [None]

- [ ] **Expected outcomes stated**: Each step has clear outcome? [Yes/No]
  - Missing outcomes: [List] or [None]

**Score**:
- 1.0 if all checks pass (4/4)
- 0.75 if 3/4
- 0.5 if 2/4
- 0.25 if 1/4
- 0.0 if 0/4

**documentation_clarity** = **[Score]**

---

### Usability Score

**Formula**:
```
V_usability = (
  0.40 × quick_start_works +
  0.35 × examples_runnable +
  0.25 × documentation_clarity
)
```

**Calculation**:
- Quick Start: [X] × 0.40 = [A]
- Examples: [Y] × 0.35 = [B]
- Documentation clarity: [Z] × 0.25 = [C]

**V_usability** = [A] + [B] + [C] = **[Score]**

**Scoring Guide**:
- 1.0: Excellent usability (Quick Start ≤30 min, all examples work, all clarity checks pass)
- 0.8: Good usability (Quick Start ≤45 min, ≥90% examples work, 3/4 clarity checks)
- 0.6: Adequate usability (Quick Start ≤60 min, ≥80% examples work, 2/4 clarity checks)
- 0.4: Poor usability (Quick Start ≤90 min, ≥60% examples work, 1/4 clarity checks)
- 0.2: Very poor usability (Quick Start >90 min, <60% examples work, 0/4 clarity checks)

---

## 4. Format Check

### 4.1 Frontmatter Completeness

**Required fields**:
- [ ] `name`: Present? [Yes/No] - Value: [X]
- [ ] `description`: Present? [Yes/No] - Length: [X] chars (max 400)
- [ ] `allowed-tools`: Present? [Yes/No] - Valid array? [Yes/No]

**Score**: (present_fields / required_fields) = [X]/3 = **[frontmatter_complete]**

---

### 4.2 Directory Structure Correctness

**Expected structure** (from reference skills):
```
.claude/skills/[skill-name]/
├── SKILL.md
├── templates/
├── reference/
├── examples/
└── scripts/
```

**Verification**:
- [ ] `SKILL.md` exists?
- [ ] `templates/` directory exists?
- [ ] `reference/` directory exists?
- [ ] `examples/` directory exists?
- [ ] `scripts/` directory exists (if applicable)?

**Comparison**: Compare to `testing-strategy` or `error-recovery` structure
```bash
diff -r .claude/skills/testing-strategy/ .claude/skills/[skill-name]/
```

**Score**:
- 1.0 if perfect match (all directories, same structure)
- 0.8 if minor differences (extra directories okay)
- 0.5 if significant differences (missing required directories)
- 0.0 if wrong structure

**directory_structure_correct** = **[Score]**

---

### 4.3 Markdown Syntax Validity

**Automated check** (if `markdownlint` available):
```bash
markdownlint .claude/skills/[skill-name]/*.md .claude/skills/[skill-name]/**/*.md
```

**Manual checks**:
- [ ] Heading levels consistent (H2 for main, H3 for sub)?
- [ ] Lists properly formatted (consistent indentation)?
- [ ] Code blocks have language tags?
- [ ] Links properly formatted `[text](url)`?

**Error count**: [X] errors

**Score**:
- 1.0 if 0 errors
- 0.9 if ≤5 errors
- 0.7 if ≤10 errors
- 0.5 if ≤20 errors
- 0.0 if >20 errors

**markdown_syntax_valid** = **[Score]**

---

### 4.4 Naming Conventions

**Rules**:
- Files: kebab-case (e.g., `extract-method-walkthrough.md`)
- Directories: lowercase, singular nouns (e.g., `templates/`, `examples/`)
- Headers: Title Case for H1, Sentence case for H2+

**Verification**:
```bash
# Check for uppercase in filenames (should be empty)
find .claude/skills/[skill-name]/ -name "*[A-Z]*"

# Check for non-kebab-case (should be empty)
find .claude/skills/[skill-name]/ -name "*_*" -o -name "* *"
```

**Violations**:
- [ ] File naming: [Count] violations - Files: [List or None]
- [ ] Directory naming: [Count] violations - Dirs: [List or None]
- [ ] Header casing: [Count] violations - Headers: [List or None]

**Total violations**: [X]

**Score**:
- 1.0 if 0 violations
- 0.75 if 1 violation
- 0.5 if 2-3 violations
- 0.0 if >3 violations

**naming_conventions_followed** = **[Score]**

---

### Format Score

**Formula**:
```
V_format = (
  0.30 × frontmatter_complete +
  0.30 × directory_structure_correct +
  0.25 × markdown_syntax_valid +
  0.15 × naming_conventions_followed
)
```

**Calculation**:
- Frontmatter: [X] × 0.30 = [A]
- Directory structure: [Y] × 0.30 = [B]
- Markdown syntax: [Z] × 0.25 = [C]
- Naming conventions: [W] × 0.15 = [D]

**V_format** = [A] + [B] + [C] + [D] = **[Score]**

**Scoring Guide**:
- 1.0: Perfect compliance (all checks 100%)
- 0.9: Near perfect (≥95% compliance)
- 0.8: High compliance (≥90% compliance)
- 0.7: Good compliance (≥80% compliance)
- 0.5: Moderate compliance (≥70% compliance)
- 0.3: Low compliance (≥50% compliance)

---

## Overall V_instance Calculation

**Formula**:
```
V_instance = 0.3×V_completeness + 0.3×V_accuracy + 0.2×V_usability + 0.2×V_format
```

**Calculation**:
- V_completeness: [X] × 0.3 = [A]
- V_accuracy: [Y] × 0.3 = [B]
- V_usability: [Z] × 0.2 = [C]
- V_format: [W] × 0.2 = [D]

**V_instance** = [A] + [B] + [C] + [D] = **[Score]**

**Convergence Threshold**: ≥0.85 (target)

**Interpretation**:
- ≥0.85: Strong extraction, ready for use
- 0.75-0.84: Good extraction, minor improvements needed
- 0.65-0.74: Adequate extraction, significant improvements needed
- <0.65: Weak extraction, major rework required

---

## Remediation Checklist

If V_instance < 0.85, prioritize fixes:

### High Priority (Impact >0.05)
- [ ] Missing patterns (V_completeness)
- [ ] Broken links (V_accuracy)
- [ ] Quick Start doesn't work (V_usability)

### Medium Priority (Impact 0.02-0.05)
- [ ] Missing examples (V_completeness)
- [ ] Inaccurate metrics (V_accuracy)
- [ ] Unclear documentation (V_usability)
- [ ] Frontmatter incomplete (V_format)

### Low Priority (Impact <0.02)
- [ ] Markdown syntax errors (V_format)
- [ ] Naming convention violations (V_format)

---

## Time Estimates

| Check Category | Time Estimate |
|----------------|---------------|
| Completeness | 10-15 min |
| Accuracy | 15-20 min |
| Usability | 30-45 min (includes Quick Start execution) |
| Format | 10-15 min |
| **Total** | **65-95 min** |

**Critical Path**: Usability check (Quick Start execution takes longest)

---

**Version**: 1.0
**Last Updated**: 2025-10-19
**Validated**: Bootstrap-005 Iteration 0-1 (code-refactoring skill validated successfully)
