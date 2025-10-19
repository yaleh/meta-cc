# Extract Knowledge Capability

**Purpose**: Parse BAIME experiment artifacts and identify extractable knowledge (patterns, principles, templates, examples, scripts).

**Version**: 1.0
**Created**: 2025-10-19
**Source**: Bootstrap-005, Iteration 1

---

## Inputs

- **Experiment directory path**: Location of completed BAIME experiment
- **Source files**:
  - `results.md`: Final analysis with patterns, principles, metrics
  - `iterations/*.md`: Evolution trajectory, decisions, learnings
  - `knowledge/templates/`: Reusable process templates
  - `scripts/`: Automation scripts

---

## Process

### Step 1: Read results.md

**Objective**: Extract patterns, principles, and validation metrics

**Actions**:
1. **Locate "Patterns" section**:
   ```bash
   grep -n "## Patterns\|### Pattern:" results.md
   ```
   - Read each pattern description
   - Note: Name, context, problem, solution, validation data
   - Record: Source line numbers for reference

2. **Locate "Principles" section**:
   ```bash
   grep -n "## Principles\|### Principle:" results.md
   ```
   - Read each principle statement
   - Note: Name, rationale, evidence
   - Record: Source line numbers

3. **Extract validation metrics**:
   - Search for: V_instance, V_meta scores
   - Look for: Success rates, applications count, quantified impact
   - Record: All numeric metrics for accuracy verification

**Output**: List of patterns (N count), principles (M count), metrics

---

### Step 2: Scan iterations/*.md

**Objective**: Identify evolution trajectory, applied patterns, examples

**Actions**:
1. **List all iteration files**:
   ```bash
   ls -1 experiments/[experiment]/iterations/iteration-*.md
   ```

2. **For each iteration file**:
   - **Scan for pattern applications**:
     ```bash
     grep -n "Pattern:" iteration-X.md
     ```
   - **Identify detailed execution narratives** (potential examples):
     - Look for: Step-by-step walkthroughs (≥5 steps)
     - Look for: Code examples, commands, timing data
     - Look for: Quantified outcomes (before/after metrics)

   - **Note key decisions and learnings**:
     - Look for: "What Worked Well", "What Didn't Work", "Lessons Learned"
     - Extract: Insights that inform methodology refinement

**Output**: List of example candidates (K count), pattern application evidence

---

### Step 3: Inventory templates

**Objective**: Catalog reusable templates for copying

**Actions**:
1. **List template files**:
   ```bash
   ls -lh experiments/[experiment]/knowledge/templates/*.md
   ```

2. **For each template**:
   - Record: Filename, line count, file size
   - Read: First 10 lines to understand purpose
   - Note: Whether template is workflow/checklist/protocol

**Output**: Template inventory (T count, total lines)

---

### Step 4: Inventory scripts

**Objective**: Catalog automation scripts for copying

**Actions**:
1. **List script files**:
   ```bash
   ls -lh experiments/[experiment]/scripts/*.sh
   ```

2. **For each script**:
   - Record: Filename, line count, file size
   - Read: Usage comment or first 20 lines
   - Note: Purpose, dependencies, generalizability

**Output**: Script inventory (S count, total lines)

---

### Step 5: Identify cross-cutting patterns vs domain-specific

**Objective**: Classify knowledge by transferability

**Actions**:
1. **For each pattern**:
   - Check description for: "Universal", "Language-independent", "Domain-agnostic"
   - Assess: Does it apply to Go only? Testing only? Or broadly?
   - Classify: Universal / Language-family / Domain-specific

2. **For each principle**:
   - Assess: Is it fundamental (e.g., TDD) or context-specific (e.g., Go cyclomatic complexity ≤8)?
   - Classify: Universal / Domain-specific / Project-specific

**Output**: Transferability classification for each item

---

### Step 6: Create extraction inventory

**Objective**: Structured catalog to guide transformation work

**Template** (JSON):
```json
{
  "experiment": "[experiment-name]",
  "extraction_date": "[YYYY-MM-DD]",
  "patterns": [
    {
      "name": "[Pattern Name]",
      "source_file": "results.md",
      "source_lines": "[start-end]",
      "validated": true,
      "applications_count": [N],
      "transferability": "[universal|language-family|domain-specific]"
    }
  ],
  "principles": [
    {
      "name": "[Principle Name]",
      "source_file": "results.md",
      "source_lines": "[start-end]",
      "evidence": "[Brief description]",
      "transferability": "[universal|domain-specific|project-specific]"
    }
  ],
  "templates": [
    {
      "filename": "[template-name.md]",
      "location": "knowledge/templates/",
      "lines": [N],
      "bytes": [N],
      "purpose": "[Description]"
    }
  ],
  "scripts": [
    {
      "filename": "[script-name.sh]",
      "location": "scripts/",
      "lines": [N],
      "bytes": [N],
      "purpose": "[Description]",
      "dependencies": ["[tool1]", "[tool2]"]
    }
  ],
  "examples": [
    {
      "source": "iteration-[N].md",
      "narrative": "[Description]",
      "lines": "[start-end]",
      "suitable_for_walkthrough": true,
      "estimated_effort": "[X] min to extract"
    }
  ],
  "metrics": {
    "v_instance": [X.XX],
    "v_meta": [X.XX],
    "success_rate": "[X]%",
    "applications": [N]
  },
  "gaps": [
    "[Missing or incomplete item 1]",
    "[Missing or incomplete item 2]"
  ],
  "time_estimate": {
    "reading": "[X] min",
    "extraction": "[Y] min",
    "transformation": "[Z] min",
    "validation": "[W] min",
    "total": "[Total] min"
  }
}
```

**Output**: `data/extraction-inventory.json`

---

## Outputs

- **Extraction inventory** (JSON): `data/extraction-inventory.json`
  - Complete catalog of extractable knowledge
  - Source references (file, line numbers)
  - Transferability classifications
  - Time estimates

---

## Quality Checks

**Before finalizing extraction inventory**:

- [ ] **All sections of results.md reviewed**: Patterns, Principles, Metrics, Learnings
- [ ] **All iteration files scanned**: Identified example candidates
- [ ] **All templates inventoried**: Count, purpose, lines documented
- [ ] **All scripts inventoried**: Count, purpose, dependencies documented
- [ ] **Transferability assessed**: Each item classified (universal/domain/project)
- [ ] **Gaps identified**: Missing or incomplete items documented
- [ ] **Time estimates**: Realistic estimates for transformation work

---

## Success Indicators

**Extraction is complete if**:
✅ Pattern count ≥ principle count (patterns are concrete, more numerous)
✅ At least 1 example candidate identified
✅ All templates and scripts inventoried (100% coverage)
✅ Time estimate ≤ 6 hours for full transformation (reasonable scope)
✅ Gaps list is honest (low scores acceptable if documented)

**Extraction needs revision if**:
❌ Pattern count = 0 (missed patterns in results.md)
❌ Principles count = 0 (missed principles section)
❌ No examples identified (iterations not thoroughly scanned)
❌ Time estimate >10 hours (scope too large, need to prioritize)

---

## Time Estimate

**Expected**: 10-15 minutes for thorough extraction inventory

**Breakdown**:
- Read results.md: 5-7 min
- Scan iterations/*.md: 3-5 min
- Inventory templates/scripts: 1-2 min
- Create JSON: 1-2 min

---

## Anti-Patterns to Avoid

❌ **Skipping iteration files**: Results.md is primary source, but iterations have critical context and examples

❌ **Not recording source line numbers**: Makes accuracy verification impossible later

❌ **Assuming completeness**: Always enumerate gaps (missing patterns, incomplete principles)

❌ **Vague time estimates**: Base on actual line counts and complexity, not optimistic guesses

---

**Version**: 1.0
**Last Updated**: 2025-10-19
**Validated**: Bootstrap-005 Iteration 0-1 (extraction inventory created successfully)
