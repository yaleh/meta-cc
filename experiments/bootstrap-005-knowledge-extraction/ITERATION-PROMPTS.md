# ITERATION-PROMPTS.md

**Experiment**: Bootstrap-005: Knowledge Extraction Methodology
**Domain**: Knowledge Engineering - Systematic extraction and transformation of BAIME experiment artifacts into reusable Claude Code skills and knowledge base entries
**Version**: 1.0
**Created**: 2025-10-19

---

## Table of Contents

1. [Experiment Overview](#experiment-overview)
2. [Architecture Specification](#architecture-specification)
3. [Value Functions](#value-functions)
4. [Baseline Iteration (Iteration 0)](#baseline-iteration-iteration-0)
5. [Subsequent Iterations (Iteration N)](#subsequent-iterations-iteration-n)
6. [Knowledge Organization](#knowledge-organization)
7. [Results Analysis](#results-analysis)
8. [Execution Guidance](#execution-guidance)

---

## Experiment Overview

### Domain Analysis

**Domain**: Knowledge Engineering for BAIME Experiments

**Core Concepts**:
- **Knowledge Extraction**: Identifying patterns, principles, templates, and automation scripts from experiment artifacts
- **Format Transformation**: Converting experiment-specific formats to standardized skill/knowledge base formats
- **Quality Preservation**: Maintaining accuracy and completeness during transformation
- **Generalization**: Creating methodology that works across different experiment types and domains

**Data Sources**:
1. **Primary Target**: `experiments/bootstrap-004-refactoring-guide/` (Iteration 0 baseline)
2. **Validation Targets**: `experiments/bootstrap-002-testing-strategy/`, `experiments/bootstrap-003-error-recovery/`
3. **Format References**: `.claude/skills/testing-strategy/`, `.claude/skills/error-recovery/`
4. **Target Structures**: `knowledge/patterns/`, `knowledge/principles/`, `knowledge/best-practices/`

**Instance Task**: Extract knowledge from Bootstrap-004 experiment and transform into `.claude/skills/code-refactoring/` skill + knowledge base entries

**Methodology Task**: Develop systematic, generalizable, efficient knowledge extraction workflow with automation tools

### Success Criteria

**Instance Success** (V_instance ≥ 0.85):
- Complete extraction: All patterns, principles, templates, examples, scripts
- Accurate transformation: Content matches source, code examples correct
- Usable output: Quick Start works, examples runnable, documentation clear
- Standard compliance: Frontmatter complete, structure matches conventions

**Methodology Success** (V_meta ≥ 0.75):
- Generalizable: Works on Bootstrap-002, Bootstrap-003 without modification
- Efficient: ≥2x speedup vs manual baseline
- Automated: ≥60% automation rate

**Convergence**: Both thresholds met + validated on ≥2 experiments + stable across final 2 iterations

---

## Architecture Specification

### Meta-Agent System (Modular Capabilities)

**Purpose**: Knowledge extraction lifecycle management - orchestrate extraction, transformation, validation

**Capabilities** (stored in `system/capabilities/`):

1. **`extract-knowledge.md`**: Parse experiment artifacts and identify extractable knowledge
   - Parse `results.md` for patterns, principles, metrics
   - Parse `iterations/*.md` for evolution trajectory, decisions
   - Parse `knowledge/templates/` for reusable templates
   - Parse `scripts/` for automation tools
   - Identify cross-cutting patterns vs domain-specific patterns
   - Output: Structured extraction inventory (JSON/markdown)

2. **`transform-formats.md`**: Convert experiment format to skill/knowledge base format
   - Generate SKILL.md with frontmatter (name, description, location, timestamps)
   - Create `templates/` directory with extracted templates
   - Create `reference/` directory with methodology documentation
   - Create `examples/` directory with walkthroughs
   - Copy `scripts/` with adaptation for skill context
   - Generate `knowledge/patterns/*.md` with frontmatter
   - Generate `knowledge/principles/*.md` with frontmatter
   - Output: Transformed artifacts ready for validation

3. **`validate-artifacts.md`**: Quality assurance for extracted knowledge
   - Completeness check: All expected artifacts present
   - Accuracy check: Content matches source material
   - Format check: Frontmatter valid, markdown syntax correct, links work
   - Usability check: Examples executable, Quick Start testable
   - Standard compliance: Directory structure matches conventions
   - Output: Validation report with pass/fail + remediation list

**Reading Protocol**:
- Read ALL capabilities before starting iteration
- Read specific capability again immediately before using it
- Capabilities are prescriptive - follow exactly

### Agent System (Specialized Executors)

**Initial Setup** (Iteration 0): Use generic meta-agent only (no specialization)

**Specialization Triggers** (evidence-based):
- Performance gap >5x between generic agent and specialized requirement
- Retrospective evidence from previous iterations shows repeated failures
- Attempted alternatives with generic agent documented

**Potential Specialization** (only create if triggered):
- **PatternExtractor Agent**: IF generic agent struggles to identify patterns consistently
- **FormatTransformer Agent**: IF transformation rules become complex and error-prone
- **QualityValidator Agent**: IF validation requires domain-specific expertise

**Evolution Protocol**:
1. Document performance issues with generic agent
2. Attempt process improvements first
3. Only create specialized agent if process improvements fail
4. Validate improvement quantitatively (before/after metrics)

### Modular Principle

**File Organization**:
```
experiments/bootstrap-005-knowledge-extraction/
├── ITERATION-PROMPTS.md           # This file
├── system/
│   ├── capabilities/              # Modular capabilities
│   │   ├── extract-knowledge.md
│   │   ├── transform-formats.md
│   │   └── validate-artifacts.md
│   └── agents/                    # Specialized agents (create only if needed)
│       └── .gitkeep
├── iterations/
│   ├── iteration-0.md             # Baseline
│   ├── iteration-1.md             # Systematization
│   ├── iteration-2.md             # Automation
│   └── iteration-3.md             # Generalization
├── data/
│   ├── extraction-inventory.json  # Identified knowledge to extract
│   ├── transformation-log.json    # Transformation decisions
│   └── validation-reports/        # Quality checks per iteration
├── knowledge/
│   ├── templates/                 # Reusable templates
│   ├── patterns/                  # Extracted patterns (meta-level)
│   └── principles/                # Extracted principles (meta-level)
├── scripts/
│   ├── extract-patterns.sh        # Pattern extraction automation
│   ├── generate-frontmatter.sh    # Frontmatter generation
│   └── validate-skill.sh          # Validation automation
└── results.md                     # Final analysis
```

**Separation of Concerns**:
- Capabilities = reusable logic (domain-independent where possible)
- Agents = execution context (only create when specialized expertise needed)
- Data = iteration-specific artifacts
- Knowledge = permanent, reusable knowledge
- Scripts = automation tools

---

## Value Functions

### V_instance: Knowledge Transformation Quality

**Formula**:
```
V_instance = 0.3 × V_completeness + 0.3 × V_accuracy + 0.2 × V_usability + 0.2 × V_format
```

**Component Rubrics**:

#### 1. V_completeness (Weight: 0.3)

**Definition**: Extraction coverage - percentage of available knowledge successfully extracted

**Measurement**:
```
V_completeness = (
  0.25 × (patterns_extracted / patterns_available) +
  0.25 × (principles_extracted / principles_available) +
  0.20 × (templates_present ? 1.0 : 0.0) +
  0.15 × (examples_count ≥ 1 ? 1.0 : examples_count * 0.5) +
  0.15 × (scripts_included / scripts_available)
)
```

**Data Collection**:
1. Count available patterns in `results.md` → `patterns_available`
2. Count extracted patterns in `.claude/skills/code-refactoring/` → `patterns_extracted`
3. Count available principles in `results.md` → `principles_available`
4. Count extracted principles in `knowledge/principles/` → `principles_extracted`
5. Check templates directory exists and populated → `templates_present`
6. Count example walkthroughs in `examples/` → `examples_count`
7. Count automation scripts in source → `scripts_available`
8. Count automation scripts in skill → `scripts_included`

**Scoring Guide**:
- 1.0: Complete extraction (100% patterns, 100% principles, all templates, ≥2 examples, all scripts)
- 0.8: Near complete (≥90% patterns, ≥90% principles, all templates, ≥1 example, ≥80% scripts)
- 0.6: Adequate (≥70% patterns, ≥70% principles, templates present, ≥1 example, ≥60% scripts)
- 0.4: Partial (≥50% patterns, ≥50% principles, some templates, basic example, ≥40% scripts)
- 0.2: Minimal (≥30% patterns, ≥30% principles, minimal templates, no examples, ≥20% scripts)
- 0.0: Incomplete (<30% extraction across categories)

**Honest Assessment Protocol**:
- Enumerate missing patterns explicitly
- Document reasons for exclusions
- Challenge high scores: "What am I missing?"

#### 2. V_accuracy (Weight: 0.3)

**Definition**: Transformation correctness - fidelity to source material

**Measurement**:
```
V_accuracy = (
  0.35 × pattern_description_accuracy +
  0.25 × code_example_correctness +
  0.25 × metrics_data_accuracy +
  0.15 × cross_reference_validity
)
```

**Data Collection**:

**Pattern Description Accuracy** (manual verification):
- Sample 5 random patterns
- Compare extracted description vs source in `results.md`
- Score: Identical=1.0, Minor differences=0.8, Significant differences=0.5, Wrong=0.0
- Average scores → `pattern_description_accuracy`

**Code Example Correctness**:
- Run syntax check on all code blocks: `grep -E '```' *.md | syntax-check`
- Compare code examples to source material
- Score: (correct_examples / total_examples)

**Metrics Data Accuracy**:
- Extract numeric metrics from skill (speedup, coverage, etc.)
- Compare to `results.md` source
- Score: (matching_metrics / total_metrics)

**Cross-Reference Validity**:
- Extract all internal links: `grep -o '\[.*\](.*\.md)' *.md`
- Check if targets exist
- Score: (valid_links / total_links)

**Scoring Guide**:
- 1.0: Perfect accuracy (all checks 100%)
- 0.9: Near perfect (≥95% across all checks)
- 0.8: High accuracy (≥90% across all checks)
- 0.7: Good accuracy (≥80% across all checks)
- 0.5: Moderate accuracy (≥70% across all checks)
- 0.3: Low accuracy (≥50% across all checks)
- 0.0: Poor accuracy (<50% accuracy)

**Honest Assessment Protocol**:
- Sample randomly (avoid cherry-picking)
- Document all discrepancies
- Investigate root causes of inaccuracies

#### 3. V_usability (Weight: 0.2)

**Definition**: Practical utility for end users

**Measurement**:
```
V_usability = (
  0.40 × quick_start_works +
  0.35 × examples_runnable +
  0.25 × documentation_clarity
)
```

**Data Collection**:

**Quick Start Works**:
- Test: Can new user start using skill in 30 minutes?
- Steps: (1) Read SKILL.md, (2) Follow Quick Start, (3) Run first example
- Time: Measure actual time taken
- Score: 1.0 if ≤30 min, 0.75 if ≤45 min, 0.5 if ≤60 min, 0.25 if ≤90 min, 0.0 if >90 min

**Examples Runnable**:
- Test: Execute all code examples
- Success criteria: Examples run without errors OR errors are intentional (documented)
- Score: (successful_examples / total_examples)

**Documentation Clarity**:
- Test: Can user understand without external context?
- Checks: (1) All terms defined, (2) Prerequisites listed, (3) Steps numbered, (4) Expected outcomes stated
- Score: 1.0 if all checks pass, 0.75 if 3/4, 0.5 if 2/4, 0.25 if 1/4, 0.0 if 0/4

**Scoring Guide**:
- 1.0: Excellent usability (Quick Start ≤30 min, all examples work, all clarity checks pass)
- 0.8: Good usability (Quick Start ≤45 min, ≥90% examples work, 3/4 clarity checks)
- 0.6: Adequate usability (Quick Start ≤60 min, ≥80% examples work, 2/4 clarity checks)
- 0.4: Poor usability (Quick Start ≤90 min, ≥60% examples work, 1/4 clarity checks)
- 0.2: Very poor usability (Quick Start >90 min, <60% examples work, 0/4 clarity checks)

**Honest Assessment Protocol**:
- Test with fresh perspective (avoid author bias)
- Document friction points
- Time actual user workflows

#### 4. V_format (Weight: 0.2)

**Definition**: Compliance with standards and conventions

**Measurement**:
```
V_format = (
  0.30 × frontmatter_complete +
  0.30 × directory_structure_correct +
  0.25 × markdown_syntax_valid +
  0.15 × naming_conventions_followed
)
```

**Data Collection**:

**Frontmatter Complete**:
- Required fields: `name`, `description`, `location`, `created`, `updated`
- Check: All fields present in SKILL.md and knowledge/*.md
- Score: (present_fields / required_fields)

**Directory Structure Correct**:
- Expected: `SKILL.md`, `templates/`, `reference/`, `examples/`, `scripts/`
- Check: Compare to existing skills (testing-strategy, error-recovery)
- Score: 1.0 if perfect match, 0.8 if minor differences, 0.5 if significant differences, 0.0 if wrong

**Markdown Syntax Valid**:
- Check: Run markdown linter: `markdownlint *.md`
- Score: 1.0 if 0 errors, 0.9 if ≤5 errors, 0.7 if ≤10 errors, 0.5 if ≤20 errors, 0.0 if >20 errors

**Naming Conventions Followed**:
- Files: kebab-case for filenames
- Directories: lowercase, singular nouns
- Headers: Title Case for H1, Sentence case for H2+
- Score: 1.0 if all conventions followed, 0.75 if 1 violation, 0.5 if 2-3 violations, 0.0 if >3 violations

**Scoring Guide**:
- 1.0: Perfect compliance (all checks 100%)
- 0.9: Near perfect (≥95% compliance)
- 0.8: High compliance (≥90% compliance)
- 0.7: Good compliance (≥80% compliance)
- 0.5: Moderate compliance (≥70% compliance)
- 0.3: Low compliance (≥50% compliance)
- 0.0: Non-compliant (<50% compliance)

**Honest Assessment Protocol**:
- Use automated checks where possible
- Document all deviations from standards
- Justify intentional deviations

---

### V_meta: Extraction Methodology Quality

**Formula**:
```
V_meta = 0.4 × V_generality + 0.3 × V_efficiency + 0.3 × V_automation
```

**Component Rubrics**:

#### 1. V_generality (Weight: 0.4)

**Definition**: Methodology applicability across different experiments and domains

**Measurement**:
```
V_generality = (
  0.30 × bootstrap_002_success +
  0.30 × bootstrap_003_success +
  0.25 × domain_independence +
  0.15 × experiment_type_flexibility
)
```

**Data Collection**:

**Bootstrap-002 Success** (testing-strategy):
- Apply methodology to extract knowledge from Bootstrap-002
- Measure: V_instance score achieved
- Score: 1.0 if V_instance ≥0.85, 0.8 if ≥0.75, 0.6 if ≥0.65, 0.4 if ≥0.50, 0.0 if <0.50
- Document: Adaptations required (none=best, minor=good, major=poor)

**Bootstrap-003 Success** (error-recovery):
- Apply methodology to extract knowledge from Bootstrap-003
- Measure: V_instance score achieved
- Score: Same as Bootstrap-002

**Domain Independence**:
- Test: Do extraction rules depend on specific domain (refactoring, testing, errors)?
- Review: Extraction rules, transformation templates, validation criteria
- Score: 1.0 if fully domain-independent, 0.7 if minor domain assumptions, 0.4 if significant domain coupling, 0.0 if domain-specific

**Experiment Type Flexibility**:
- Test: Does methodology work for both retrospective (historical data) and prospective (live deployment) experiments?
- Bootstrap-003 = retrospective, Bootstrap-004 = prospective
- Score: 1.0 if works for both, 0.5 if works for one, 0.0 if neither

**Scoring Guide**:
- 1.0: Highly generalizable (works on both validation experiments with V_instance ≥0.85, fully domain-independent, works for both experiment types)
- 0.8: Good generalizability (works on both with V_instance ≥0.75, minor domain assumptions, works for both types)
- 0.6: Moderate generalizability (works on 1 validation experiment with V_instance ≥0.75, some domain coupling, works for one type)
- 0.4: Limited generalizability (works on 1 validation experiment with V_instance ≥0.65, significant domain coupling)
- 0.2: Poor generalizability (struggles on validation experiments, highly domain-specific)
- 0.0: Not generalizable (fails on validation experiments)

**Honest Assessment Protocol**:
- Test on genuinely different experiments (not just similar domains)
- Document all adaptations required
- Seek counter-examples: "Where does this fail?"

#### 2. V_efficiency (Weight: 0.3)

**Definition**: Time savings compared to manual baseline

**Measurement**:
```
V_efficiency = min(1.0, (speedup - 1.0) / (target_speedup - 1.0))
where target_speedup = 2.0 (≥2x speedup)
```

**Data Collection**:

**Baseline Time** (Iteration 0):
- Measure: Total time for manual extraction (Bootstrap-004 → code-refactoring skill)
- Include: Reading source, identifying patterns, writing SKILL.md, creating examples, validating
- Document: Time per phase (reading=X, extraction=Y, transformation=Z, validation=W)
- Record: `baseline_time_hours`

**Methodology Time** (Iteration 1+):
- Measure: Total time using systematic methodology + automation
- Include: All phases with new process
- Record: `methodology_time_hours`

**Speedup Calculation**:
```
speedup = baseline_time_hours / methodology_time_hours
efficiency_score = min(1.0, (speedup - 1.0) / (2.0 - 1.0))
```

**Examples**:
- Speedup 2.0x → efficiency_score = 1.0 (meets target)
- Speedup 1.5x → efficiency_score = 0.5 (50% of target)
- Speedup 3.0x → efficiency_score = 1.0 (exceeds target, capped)

**Scoring Guide**:
- 1.0: Excellent efficiency (≥2x speedup)
- 0.8: Good efficiency (1.8x speedup)
- 0.6: Moderate efficiency (1.6x speedup)
- 0.4: Low efficiency (1.4x speedup)
- 0.2: Minimal efficiency (1.2x speedup)
- 0.0: No improvement (≤1x speedup)

**Honest Assessment Protocol**:
- Track time rigorously (no rounding up)
- Include all overhead (setup, debugging, validation)
- Account for learning curve (methodology familiarity)

#### 3. V_automation (Weight: 0.3)

**Definition**: Degree of tool support and automation

**Measurement**:
```
V_automation = (
  0.50 × automation_rate +
  0.30 × tool_coverage +
  0.20 × tool_reliability
)
```

**Data Collection**:

**Automation Rate**:
- List all methodology steps (from `extract-knowledge.md`, `transform-formats.md`, `validate-artifacts.md`)
- Classify: Automated (tool does it), Manual (human does it)
- Count: `automated_steps`, `manual_steps`
- Calculate: `automation_rate = automated_steps / (automated_steps + manual_steps)`

**Tool Coverage**:
- Identify automation opportunities: Pattern extraction, frontmatter generation, validation, link checking, syntax checking
- Count: `implemented_tools`, `total_opportunities`
- Calculate: `tool_coverage = implemented_tools / total_opportunities`

**Tool Reliability**:
- Test: Run all automation tools on Bootstrap-004
- Measure: Success rate (successful_runs / total_runs)
- Check: False positives, false negatives in validation tools
- Score: 1.0 if 100% reliable, 0.9 if ≥95%, 0.8 if ≥90%, 0.6 if ≥80%, 0.4 if ≥70%, 0.0 if <70%

**Scoring Guide**:
- 1.0: Highly automated (≥60% automation rate, ≥80% tool coverage, 100% reliability)
- 0.8: Good automation (≥50% automation rate, ≥70% tool coverage, ≥90% reliability)
- 0.6: Moderate automation (≥40% automation rate, ≥60% tool coverage, ≥80% reliability)
- 0.4: Limited automation (≥30% automation rate, ≥50% tool coverage, ≥70% reliability)
- 0.2: Minimal automation (≥20% automation rate, ≥40% tool coverage, ≥60% reliability)
- 0.0: No automation (<20% automation rate, <40% tool coverage, <60% reliability)

**Honest Assessment Protocol**:
- Count ALL steps (don't skip "obvious" manual steps)
- Test tool reliability on fresh data (not development data)
- Document tool limitations explicitly

---

### Convergence Criteria

**Dual-Layer Thresholds**:
- V_instance ≥ 0.85 (high-quality extraction)
- V_meta ≥ 0.75 (efficient, generalizable methodology)

**Validation Requirements**:
- Methodology tested on ≥2 different experiments (Bootstrap-002, Bootstrap-003)
- Both validation tests achieve V_instance ≥ 0.75

**Stability Requirements**:
- Final 2 iterations show stable scores (Δ < 0.05 in both V_instance and V_meta)
- No major issues discovered in validation phase

**Declaration**:
When all criteria met, declare convergence in iteration report:
```
## Convergence Assessment

**Status**: CONVERGED ✅

**Evidence**:
- V_instance = 0.87 (≥0.85) ✅
- V_meta = 0.78 (≥0.75) ✅
- Validation: Bootstrap-002 (V_instance=0.81), Bootstrap-003 (V_instance=0.79) ✅
- Stability: Iteration N-1 (0.86, 0.77), Iteration N (0.87, 0.78), Δ=(0.01, 0.01) ✅
```

---

## Baseline Iteration (Iteration 0)

### Context

**Purpose**: Establish baseline through manual knowledge extraction from Bootstrap-004

**Expected Outcome**: Low baseline is ACCEPTABLE and EXPECTED. This is the unoptimized starting point.

**Baseline Hypothesis**:
- V_instance: 0.20-0.40 (manual process, no systematic methodology, likely incomplete)
- V_meta: N/A (no methodology exists yet, only measuring manual process time)
- Time: 6-10 hours (reading, extraction, transformation, validation all manual)

### System Setup

**Create Modular Architecture**:

1. **Create capability files** (even if not used in Iteration 0 - creates structure):

```bash
mkdir -p system/capabilities
cat > system/capabilities/extract-knowledge.md << 'EOF'
# Extract Knowledge Capability

**Purpose**: Parse experiment artifacts and identify extractable knowledge

## Inputs
- Experiment directory path
- Source files: results.md, iterations/*.md, knowledge/templates/, scripts/

## Process
1. Parse results.md for patterns, principles, metrics
2. Parse iterations/*.md for evolution, decisions
3. Identify templates in knowledge/templates/
4. Identify scripts in scripts/
5. Classify knowledge: patterns vs principles vs best practices
6. Identify cross-cutting patterns vs domain-specific

## Outputs
- Extraction inventory (JSON): List of identified knowledge items
- Classification: Pattern/Principle/Template/Script/Example
- Source references: File and line numbers

## Quality Checks
- All sections of results.md reviewed
- All iteration files reviewed
- Templates inventoried
- Scripts inventoried
EOF

cat > system/capabilities/transform-formats.md << 'EOF'
# Transform Formats Capability

**Purpose**: Convert experiment format to skill/knowledge base format

## Inputs
- Extraction inventory (from extract-knowledge)
- Target format specifications (SKILL.md frontmatter, directory structure)
- Format references: .claude/skills/testing-strategy/, .claude/skills/error-recovery/

## Process
1. Generate SKILL.md with frontmatter (name, description, location, timestamps)
2. Create templates/ directory and copy templates
3. Create reference/ directory and convert methodology docs
4. Create examples/ directory and extract walkthroughs
5. Create scripts/ directory and adapt automation tools
6. Generate knowledge/patterns/*.md with frontmatter
7. Generate knowledge/principles/*.md with frontmatter

## Outputs
- .claude/skills/code-refactoring/ directory structure
- knowledge/patterns/ entries
- knowledge/principles/ entries

## Quality Checks
- All directories created
- Frontmatter complete
- Content matches source
- Links valid
EOF

cat > system/capabilities/validate-artifacts.md << 'EOF'
# Validate Artifacts Capability

**Purpose**: Quality assurance for extracted knowledge

## Inputs
- Transformed artifacts (.claude/skills/code-refactoring/, knowledge/)
- Source material (experiments/bootstrap-004-refactoring-guide/)
- Validation criteria (completeness, accuracy, format, usability)

## Process
1. Completeness check: All expected artifacts present
2. Accuracy check: Sample patterns and compare to source
3. Format check: Frontmatter valid, markdown syntax, links work
4. Usability check: Quick Start testable, examples runnable
5. Standard compliance: Directory structure matches conventions

## Outputs
- Validation report (pass/fail per check)
- Remediation list (issues to fix)
- Quality scores (for V_instance calculation)

## Quality Checks
- All validation criteria applied
- Evidence documented
- Scoring honest (not inflated)
EOF
```

2. **Create agent directory** (empty initially):
```bash
mkdir -p system/agents
touch system/agents/.gitkeep
```

3. **Create data directories**:
```bash
mkdir -p data/validation-reports
mkdir -p knowledge/templates
mkdir -p knowledge/patterns
mkdir -p knowledge/principles
mkdir -p scripts
```

### Objectives

**Sequential Steps**:

1. **Manual Knowledge Extraction** (4-6 hours estimated):
   - Read `experiments/bootstrap-004-refactoring-guide/results.md` thoroughly
   - Identify all patterns mentioned (list in extraction-inventory.json)
   - Identify all principles mentioned (list in extraction-inventory.json)
   - Identify templates in knowledge/templates/ (list files)
   - Identify scripts in scripts/ (list files)
   - Identify examples/walkthroughs in iterations/ (extract excerpts)
   - Document: `data/extraction-inventory.json`

2. **Manual Format Transformation** (3-4 hours estimated):
   - Study format references: `.claude/skills/testing-strategy/SKILL.md`, `.claude/skills/error-recovery/SKILL.md`
   - Create `.claude/skills/code-refactoring/SKILL.md` with frontmatter
   - Copy templates to `.claude/skills/code-refactoring/templates/`
   - Write reference documentation in `reference/`
   - Extract examples to `examples/`
   - Copy scripts to `scripts/` (adapt if needed)
   - Create `knowledge/patterns/` entries
   - Create `knowledge/principles/` entries

3. **Manual Validation** (1-2 hours estimated):
   - Check completeness: All patterns extracted?
   - Check accuracy: Sample 5 patterns, compare to source
   - Check format: Frontmatter complete? Markdown valid?
   - Check usability: Quick Start testable?
   - Document: `data/validation-reports/iteration-0.md`

4. **Baseline Value Calculation**:
   - Calculate V_completeness: Count patterns/principles/templates/scripts
   - Calculate V_accuracy: Manual verification of samples
   - Calculate V_usability: Test Quick Start, check examples
   - Calculate V_format: Check frontmatter, structure, markdown
   - Calculate V_instance = weighted average
   - Record baseline time: Total hours spent
   - Document: In iteration-0.md

5. **Problem Identification**:
   - What was time-consuming? (reading, writing, formatting, validation)
   - What was error-prone? (format inconsistencies, missed patterns, incorrect transformations)
   - What could be automated? (pattern extraction, frontmatter generation, validation checks)
   - What needs systematization? (extraction process, transformation rules, quality criteria)
   - Document: In iteration-0.md "Problems and Gaps"

### Data Collection

**Track rigorously**:
- Total time: X hours (be honest, include all time)
- Time breakdown: Reading=A, Extraction=B, Transformation=C, Validation=D
- Patterns identified: N patterns
- Patterns extracted: M patterns (M ≤ N)
- Principles identified: P principles
- Principles extracted: Q principles (Q ≤ P)
- Templates: T templates
- Scripts: S scripts
- Examples: E examples

**Document in**: `iterations/iteration-0.md`

### Expected Baseline Values

**Honest Expectations** (low baseline is GOOD - shows room for improvement):

```
V_completeness: 0.40-0.60 (likely miss some patterns, incomplete extraction)
V_accuracy: 0.60-0.80 (manual process prone to transcription errors)
V_usability: 0.30-0.50 (first attempt, may miss usability issues)
V_format: 0.50-0.70 (format compliance but likely some deviations)

V_instance: 0.20-0.40 (weighted average)
```

**Why Low Baseline is Expected**:
- No systematic process → inconsistent extraction
- Manual identification → easy to miss patterns
- First-time formatting → likely deviations from standards
- No validation tools → errors undetected

**Why Low Baseline is Acceptable**:
- Provides clear improvement opportunity
- Realistic starting point
- Honest assessment foundation

### Constraints

- Use ONLY manual process (no automation, no systematic methodology)
- Track time honestly (include breaks, debugging, rework)
- Document all issues encountered
- Do NOT inflate baseline scores (honest assessment critical)

### Deliverables

1. `iterations/iteration-0.md`: Complete iteration report with baseline values
2. `data/extraction-inventory.json`: List of identified knowledge
3. `.claude/skills/code-refactoring/`: Extracted skill (even if incomplete)
4. `knowledge/patterns/`: Initial pattern entries (even if incomplete)
5. `knowledge/principles/`: Initial principle entries (even if incomplete)
6. `data/validation-reports/iteration-0.md`: Validation results

---

## Subsequent Iterations (Iteration N)

### Context Extraction

**Before Starting Each Iteration**:

1. **Read Previous Iteration Report**:
   - File: `iterations/iteration-{N-1}.md`
   - Extract: System state, value scores (V_instance, V_meta components), identified problems
   - Note: Evolution decisions, what worked, what didn't

2. **Read Current System State**:
   - Capabilities: `system/capabilities/*.md` (what processes exist)
   - Agents: `system/agents/*.md` (what specialized agents exist, if any)
   - Scripts: `scripts/*.sh` (what automation exists)
   - Knowledge: `knowledge/` (what has been extracted and organized)

3. **Identify Gaps**:
   - Compare previous V_instance to threshold (0.85)
   - Compare previous V_meta to threshold (0.75)
   - Review problems list from previous iteration
   - Prioritize: Biggest gaps first

### Capability Reading Protocol

**Critical Protocol** (follow exactly):

1. **Read ALL Capabilities at Iteration Start**:
   ```bash
   # At the very beginning of iteration, read ALL:
   cat system/capabilities/extract-knowledge.md
   cat system/capabilities/transform-formats.md
   cat system/capabilities/validate-artifacts.md
   # + any new capabilities created in previous iterations
   ```

2. **Re-read Specific Capability Before Use**:
   ```bash
   # Immediately before using a capability, read it again:
   cat system/capabilities/extract-knowledge.md
   # Then follow its instructions exactly
   ```

3. **Why This Protocol**:
   - Capabilities evolve across iterations
   - Re-reading ensures using latest version
   - Prevents using stale/incorrect process

### Iteration Lifecycle

**Phase 1: Data Collection**

**Objective**: Gather data about current methodology performance

**Process**:
1. If testing on new experiment (Bootstrap-002, Bootstrap-003):
   - Apply current methodology to new experiment
   - Track time per phase
   - Track issues encountered
   - Calculate V_instance for new experiment

2. If improving current extraction (Bootstrap-004):
   - Identify specific component to improve (extraction, transformation, validation)
   - Apply improved process
   - Measure impact on V_instance components

3. Document:
   - Time spent: X hours
   - V_instance achieved: Y
   - V_meta components: Automation rate, tool coverage, etc.
   - Issues: List of problems encountered

**Phase 2: Strategy Formation**

**Objective**: Analyze gaps and design improvements

**Process**:
1. **Gap Analysis**:
   ```
   Current V_instance: X
   Target V_instance: 0.85
   Gap: 0.85 - X = Y

   Component gaps:
   - V_completeness: Current A, Target 0.85, Gap B
   - V_accuracy: Current C, Target 0.85, Gap D
   - V_usability: Current E, Target 0.85, Gap F
   - V_format: Current G, Target 0.85, Gap H

   Prioritization: Address largest weighted gap first
   ```

2. **Root Cause Analysis**:
   - Why is completeness low? (missed patterns, incomplete scanning)
   - Why is accuracy low? (transcription errors, format inconsistencies)
   - Why is usability low? (unclear docs, broken examples)
   - Why is format low? (missing frontmatter, wrong structure)

3. **Improvement Hypothesis**:
   - IF we systematize extraction (checklist, templates) THEN V_completeness improves
   - IF we automate frontmatter generation THEN V_accuracy improves
   - IF we validate Quick Start THEN V_usability improves
   - IF we automate format checks THEN V_format improves

4. **Design Improvements**:
   - Update capabilities: Add steps, clarify process
   - Create automation scripts: Pattern extractor, validator
   - Create templates: Extraction checklist, transformation template

**Phase 3: Execution**

**Objective**: Implement improvements and measure impact

**Process**:
1. **Implement Changes**:
   - Update capability files: Edit `system/capabilities/*.md`
   - Create/update scripts: Write `scripts/*.sh`
   - Create/update templates: Write `knowledge/templates/*.md`

2. **Apply Improved Methodology**:
   - Follow updated capabilities exactly
   - Use new automation tools
   - Track time spent
   - Document issues

3. **Measure Impact**:
   - Re-calculate V_instance components
   - Calculate V_meta components (automation rate, efficiency)
   - Compare to previous iteration

**Phase 4: Evaluation**

**Objective**: Assess improvement and validate methodology

**Process**:
1. **Value Function Calculation**:
   - Calculate all V_instance components (use rubrics from "Value Functions" section)
   - Calculate all V_meta components (use rubrics from "Value Functions" section)
   - Calculate overall V_instance and V_meta
   - Compare to previous iteration: ΔV_instance, ΔV_meta

2. **Hypothesis Validation**:
   - Did improvements increase V_instance as predicted?
   - Which improvements worked? Which didn't?
   - Were there unexpected side effects?

3. **Evidence Documentation**:
   - Quantitative: Before/after scores, time measurements
   - Qualitative: What felt easier? What was still hard?
   - Concrete examples: Show specific improvements

**Phase 5: Convergence Check**

**Objective**: Determine if methodology has converged

**Process**:
1. **Threshold Check**:
   ```
   V_instance ≥ 0.85? [YES/NO]
   V_meta ≥ 0.75? [YES/NO]
   Validated on ≥2 experiments? [YES/NO]
   Stability (Δ < 0.05 over last 2 iterations)? [YES/NO]
   ```

2. **If All YES → CONVERGED**:
   - Document convergence evidence
   - Proceed to Results Analysis
   - Stop iterations

3. **If Any NO → Continue**:
   - Identify largest remaining gap
   - Design next iteration improvements
   - Document plan for iteration N+1

### Evolution Guidance

**Evidence-Based Evolution** (CRITICAL):

**When to Evolve System** (add capabilities, agents, tools):

✅ **Valid Triggers**:
1. **Retrospective Evidence**: Previous iterations show repeated failures in specific area
   - Example: "Iteration 1 and 2 both missed 30% of patterns → Need systematic extraction checklist"

2. **Gap Analysis**: Quantified gap in value function with identified root cause
   - Example: "V_accuracy = 0.65 (gap: 0.20), root cause: manual frontmatter generation errors → Automate frontmatter"

3. **Attempted Alternatives**: Tried simpler solutions first, they failed
   - Example: "Tried manual validation checklist (Iteration 1), still missed errors → Need automated validator"

4. **Necessity Demonstrated**: Evolution addresses specific, measured problem
   - Example: "Pattern extraction takes 4 hours manually, automation could reduce to 1 hour → Build pattern extractor"

❌ **Invalid Triggers** (DO NOT evolve based on these):
1. **Pattern Matching**: "Other experiments use specialized agents → We should too" ❌
2. **Anticipatory Design**: "We might need X in the future → Build X now" ❌
3. **Theoretical Completeness**: "A complete system should have Y → Add Y" ❌
4. **Complexity Bias**: "More components = better" ❌

**Evolution Protocol**:

1. **Document Evidence**:
   ```
   Problem: [Specific issue from previous iterations]
   Evidence: [Quantitative data showing issue]
   Root Cause: [Analysis of why issue occurs]
   Attempted Solutions: [What was tried, why it failed]
   Proposed Evolution: [What to add/change]
   Expected Impact: [Predicted improvement in V_instance or V_meta]
   ```

2. **Implement Evolution**:
   - Make changes (new capability, new script, updated process)
   - Document changes in iteration report
   - Test on current data

3. **Validate Evolution**:
   - Measure actual impact (before/after scores)
   - Compare to predicted impact
   - Document: Did it work? By how much?

**Agent Specialization Protocol** (only create specialized agents when necessary):

**Trigger**: Generic meta-agent shows >5x performance gap vs specialized requirement

**Evidence Required**:
- Attempted with generic meta-agent across ≥2 iterations
- Documented failures (what went wrong, why)
- Quantified performance gap (time, accuracy, completeness)
- Identified specialized expertise needed

**Example**:
```
Problem: Pattern extraction with generic agent takes 4 hours, error rate 30%
Evidence: Iteration 1 (4.5h, 28% errors), Iteration 2 (4.2h, 32% errors)
Root Cause: Requires domain expertise to identify extractable patterns
Specialized Need: Pattern recognition across different experiment formats
Proposed Agent: PatternExtractor (specialized in identifying reusable patterns)
Expected Improvement: 2 hours, 10% error rate
```

**Only Then**: Create `system/agents/pattern-extractor.md`

### Key Principles

**Throughout All Iterations**:

1. **Honest Calculation**: No inflated scores, seek disconfirming evidence
2. **Dual-Layer Focus**: Track both V_instance (quality) and V_meta (methodology) in every iteration
3. **Justified Evolution**: Every change backed by evidence from previous iterations
4. **Rigorous Convergence**: Don't declare convergence prematurely

**Anti-Patterns to Avoid**:
- "Good enough" without measuring (calculate V_instance/V_meta honestly)
- Adding complexity without evidence (only evolve when triggered)
- Premature convergence (validate on multiple experiments)
- Ignoring V_meta (methodology quality matters as much as instance quality)

### Deliverables (Each Iteration)

1. **Iteration Report**: `iterations/iteration-N.md`
   - Context summary
   - Objectives
   - Execution log
   - Data collection
   - Value calculations (V_instance, V_meta with all components)
   - Evaluation
   - Problems and gaps
   - Next iteration plan OR convergence declaration

2. **Updated System**:
   - Updated capabilities: `system/capabilities/*.md` (if evolved)
   - Updated agents: `system/agents/*.md` (if created/evolved)
   - Updated scripts: `scripts/*.sh` (if created/improved)
   - Updated knowledge: `knowledge/` (improved extraction)

3. **Validation Data**:
   - Validation report: `data/validation-reports/iteration-N.md`
   - If testing on new experiment: Results for Bootstrap-002 or Bootstrap-003

---

## Knowledge Organization

### Directory Structure

```
experiments/bootstrap-005-knowledge-extraction/
├── knowledge/
│   ├── templates/              # Reusable templates (permanent)
│   │   ├── extraction-checklist.md
│   │   ├── pattern-template.md
│   │   ├── principle-template.md
│   │   └── skill-frontmatter-template.md
│   ├── patterns/               # Extracted patterns (permanent, meta-level)
│   │   ├── pattern-extraction-heuristics.md
│   │   ├── format-transformation-rules.md
│   │   └── validation-strategies.md
│   ├── principles/             # Extracted principles (permanent, meta-level)
│   │   ├── evidence-based-extraction.md
│   │   ├── format-standardization.md
│   │   └── quality-preservation.md
│   └── best-practices/         # Context-specific practices (permanent)
│       ├── frontmatter-generation.md
│       ├── link-validation.md
│       └── example-extraction.md
└── data/                       # Ephemeral iteration data
    ├── extraction-inventory.json
    ├── transformation-log.json
    └── validation-reports/
```

### Knowledge Index

**Purpose**: Track extracted knowledge and its sources

**File**: `knowledge/INDEX.md`

**Format**:
```markdown
# Knowledge Index

## Patterns

### Pattern: Extraction Heuristics
- **File**: `patterns/pattern-extraction-heuristics.md`
- **Source**: Iteration 2, experiments/bootstrap-004-refactoring-guide/results.md
- **Domain**: Cross-cutting (applies to all BAIME experiments)
- **Validation**: Tested on Bootstrap-002, Bootstrap-003
- **Status**: Validated ✅

### Pattern: Format Transformation Rules
- **File**: `patterns/format-transformation-rules.md`
- **Source**: Iteration 1, comparison of .claude/skills/testing-strategy/ and error-recovery/
- **Domain**: Claude Code skills (specific to .claude/skills/ format)
- **Validation**: Applied to Bootstrap-004 extraction
- **Status**: Validated ✅

## Principles

### Principle: Evidence-Based Extraction
- **File**: `principles/evidence-based-extraction.md`
- **Source**: Iteration 0 reflection, manual extraction issues
- **Domain**: Universal (applies to all knowledge extraction)
- **Validation**: Applied in all subsequent iterations
- **Status**: Validated ✅

## Templates

### Template: Extraction Checklist
- **File**: `templates/extraction-checklist.md`
- **Source**: Iteration 1, systematization of extraction process
- **Usage**: Used in extract-knowledge capability
- **Status**: Active

### Template: SKILL.md Frontmatter
- **File**: `templates/skill-frontmatter-template.md`
- **Source**: Iteration 1, analysis of existing skills
- **Usage**: Used in transform-formats capability
- **Status**: Active

## Best Practices

### Best Practice: Frontmatter Generation
- **File**: `best-practices/frontmatter-generation.md`
- **Source**: Iteration 2, automation development
- **Context**: Claude Code skills frontmatter
- **Status**: Active
```

### Dual Output

**Local Knowledge** (experiment-specific):
- Stored in: `experiments/bootstrap-005-knowledge-extraction/knowledge/`
- Purpose: Document this experiment's findings
- Lifetime: Permanent (part of experiment record)

**Project Methodology** (reusable across projects):
- Stored in: `/home/yale/work/meta-cc/knowledge/` (project root)
- Purpose: Reusable patterns, principles, templates for future experiments
- Lifetime: Permanent (part of meta-cc knowledge base)
- Updated: At convergence, copy validated knowledge from experiment to project

### Organization Principle

**Separate Ephemeral from Permanent**:

**Ephemeral** (data/):
- Iteration-specific artifacts
- Intermediate processing results
- Validation reports per iteration
- Can be regenerated
- Example: `data/validation-reports/iteration-2.md`

**Permanent** (knowledge/):
- Validated patterns, principles, templates
- Reusable across iterations and experiments
- Cannot be easily regenerated
- Example: `knowledge/patterns/extraction-heuristics.md`

**Decision Rule**: If it's reusable in future iterations/experiments → knowledge/. If it's specific to one iteration → data/.

---

## Results Analysis

### Context

**When**: After convergence is declared (all criteria met)

**Purpose**: Comprehensive analysis of experiment outcomes and methodology validation

### Analysis Dimensions

#### 1. System Output

**Instance-Level Artifacts**:
- `.claude/skills/code-refactoring/`: Complete skill with SKILL.md, templates/, reference/, examples/, scripts/
- `knowledge/patterns/`: Pattern entries extracted from Bootstrap-004
- `knowledge/principles/`: Principle entries extracted from Bootstrap-004
- Quality metrics: V_instance = X (≥0.85)

**Methodology Artifacts**:
- `system/capabilities/`: Extraction, transformation, validation capabilities
- `system/agents/`: Specialized agents (if any created)
- `scripts/`: Automation tools (pattern extractor, frontmatter generator, validator)
- `knowledge/templates/`: Reusable templates for future extractions
- Efficiency metrics: V_meta = Y (≥0.75)

#### 2. Convergence Validation

**Threshold Compliance**:
```
✅ V_instance = [X] ≥ 0.85
✅ V_meta = [Y] ≥ 0.75
✅ Validation: Bootstrap-002 (V_instance=[A]), Bootstrap-003 (V_instance=[B]) both ≥0.75
✅ Stability: Iteration [N-1] vs [N], Δ = [C] < 0.05
```

**Evidence Summary**:
- Completeness: [X]% patterns extracted, [Y]% principles extracted
- Accuracy: [Z]% verification success rate
- Usability: Quick Start completed in [T] minutes, [E]% examples runnable
- Format: [F]% compliance rate
- Generality: Works on [G] different experiments
- Efficiency: [H]x speedup
- Automation: [I]% automation rate

#### 3. Trajectory Analysis

**Iteration Evolution**:
```
Iteration 0 (Baseline):
- V_instance = 0.35, V_meta = N/A
- Time: 8 hours
- Issues: Incomplete extraction, format inconsistencies, manual validation

Iteration 1 (Systematization):
- V_instance = 0.62, V_meta = 0.45
- Time: 6 hours
- Improvements: Extraction checklist, transformation templates, format validation
- Issues: Still manual, pattern recognition inconsistent

Iteration 2 (Automation):
- V_instance = 0.78, V_meta = 0.68
- Time: 4 hours
- Improvements: Pattern extractor script, frontmatter generator, automated validator
- Issues: Generalization untested

Iteration 3 (Generalization):
- V_instance = 0.87, V_meta = 0.78
- Time: 3.5 hours (Bootstrap-004), 4 hours (Bootstrap-002), 4.5 hours (Bootstrap-003)
- Improvements: Domain-independent extraction rules, experiment type flexibility
- Convergence: ✅
```

**Trajectory Visualization**:
```
V_instance:  0.35 → 0.62 → 0.78 → 0.87 (Δ: +0.27, +0.16, +0.09)
V_meta:       N/A → 0.45 → 0.68 → 0.78 (Δ:  N/A, +0.23, +0.10)
Time:         8h  →  6h  →  4h  → 3.5h (Efficiency: 2.3x)
```

**Key Transitions**:
1. Iteration 0→1: Systematization (+0.27 V_instance, -25% time)
2. Iteration 1→2: Automation (+0.16 V_instance, +0.23 V_meta, -33% time)
3. Iteration 2→3: Generalization (+0.09 V_instance, +0.10 V_meta, -12% time, validated on 2 experiments)

#### 4. Domain Results

**Knowledge Extraction Domain**:

**Instance-Level**:
- Patterns extracted: [N] patterns from Bootstrap-004
- Principles extracted: [M] principles from Bootstrap-004
- Templates created: [T] reusable templates
- Scripts created: [S] automation scripts
- Examples created: [E] walkthroughs
- Skill completeness: [X]% (all expected components present)

**Methodology-Level**:
- Generality: Tested on 3 experiments (Bootstrap-002, Bootstrap-003, Bootstrap-004)
- Domain independence: Works for refactoring, testing, error recovery (different domains)
- Experiment type flexibility: Works for retrospective (Bootstrap-003) and prospective (Bootstrap-004)
- Efficiency: 2.3x speedup (8h → 3.5h)
- Automation: 65% automation rate ([A] automated steps / [B] total steps)

**Knowledge Catalog**:
1. **Pattern**: Extraction heuristics (how to identify extractable patterns)
2. **Pattern**: Format transformation rules (experiment → skill → knowledge base)
3. **Pattern**: Validation strategies (completeness, accuracy, format, usability checks)
4. **Principle**: Evidence-based extraction (only extract what's documented)
5. **Principle**: Format standardization (consistent frontmatter, structure, naming)
6. **Principle**: Quality preservation (accuracy checks during transformation)
7. **Template**: Extraction checklist (systematic pattern/principle identification)
8. **Template**: SKILL.md frontmatter (standardized metadata)
9. **Template**: Pattern/Principle markdown (knowledge base format)
10. **Best Practice**: Frontmatter generation (automated metadata creation)
11. **Best Practice**: Link validation (automated cross-reference checking)
12. **Best Practice**: Example extraction (identifying runnable code examples)

#### 5. Reusability Tests

**Transferability Assessment**:

**Test 1: Different Domain (Bootstrap-002 - Testing Strategy)**:
- Domain: Software testing (different from refactoring)
- Result: V_instance = [X] (≥0.75 required for validation)
- Adaptations required: [List any modifications needed]
- Conclusion: [High/Moderate/Low] transferability to different domains

**Test 2: Different Experiment Type (Bootstrap-003 - Error Recovery)**:
- Type: Retrospective validation (historical data) vs Bootstrap-004 prospective
- Result: V_instance = [Y] (≥0.75 required for validation)
- Adaptations required: [List any modifications needed]
- Conclusion: [High/Moderate/Low] transferability to different experiment types

**Generalization Score**:
```
Domain Independence: [0.0-1.0] (based on adaptations needed)
Experiment Type Flexibility: [0.0-1.0] (retrospective vs prospective)
Overall Generality: V_generality = [Z]
```

#### 6. Methodology Validation

**Empirical Validation**:
- Iterations to convergence: [N] iterations (expected: 3-4)
- Total time: [T] hours (baseline: 8h, final: 3.5h, overhead: [O]h for methodology development)
- Speedup: [S]x (meets ≥2x target: ✅/❌)
- Automation rate: [A]% (meets ≥60% target: ✅/❌)

**Retrospective Validation**:
- Applied to previous experiments: Bootstrap-002, Bootstrap-003
- Success rate: [X]/2 experiments achieved V_instance ≥0.75
- Identified improvements: [List improvements to methodology from validation]

**Quality Metrics**:
- False positive rate: [X]% (automated tools incorrectly flagging issues)
- False negative rate: [Y]% (automated tools missing real issues)
- Manual verification required: [Z]% of automated checks

**Confidence Assessment**:
```
Confidence = (1 - false_positive_rate) × (1 - false_negative_rate) × validation_success_rate
Confidence = (1 - [X]) × (1 - [Y]) × ([S]/2) = [C]

Target: ≥0.75 confidence
Result: [PASS/FAIL]
```

#### 7. Learnings

**Patterns Discovered**:
1. **Pattern**: [Name and brief description]
   - Evidence: [Where observed]
   - Generality: [Universal/Domain-specific/Project-specific]
   - Validation: [How validated]

2. **Pattern**: [Name and brief description]
   - Evidence: [Where observed]
   - Generality: [Universal/Domain-specific/Project-specific]
   - Validation: [How validated]

[Continue for all patterns]

**Principles Extracted**:
1. **Principle**: [Name and brief description]
   - Rationale: [Why it matters]
   - Application: [How to apply]
   - Validation: [Evidence it works]

2. **Principle**: [Name and brief description]
   - Rationale: [Why it matters]
   - Application: [How to apply]
   - Validation: [Evidence it works]

[Continue for all principles]

**Methodology Insights**:
- What worked well: [List successful approaches]
- What didn't work: [List failed approaches and why]
- Unexpected discoveries: [List surprises]
- Evolution decisions: [List system evolutions and their justifications]
- Anti-patterns avoided: [List temptations resisted]

**Meta-Learnings** (about BAIME itself):
- Value function design: [Insights about V_instance and V_meta]
- Convergence criteria: [Were thresholds appropriate?]
- Iteration structure: [Insights about lifecycle phases]
- Evolution protocol: [Insights about evidence-based evolution]

#### 8. Knowledge Catalog

**Extracted Knowledge Assets**:

**Patterns** (`knowledge/patterns/`):
1. `pattern-extraction-heuristics.md`: How to identify extractable patterns in experiment artifacts
2. `format-transformation-rules.md`: Rules for converting experiment format to skill/knowledge base format
3. `validation-strategies.md`: Strategies for validating extraction quality
[List all pattern files with brief descriptions]

**Principles** (`knowledge/principles/`):
1. `evidence-based-extraction.md`: Only extract knowledge that's documented and validated
2. `format-standardization.md`: Consistent frontmatter, structure, naming across artifacts
3. `quality-preservation.md`: Maintain accuracy during transformation
[List all principle files with brief descriptions]

**Templates** (`knowledge/templates/`):
1. `extraction-checklist.md`: Systematic checklist for pattern/principle identification
2. `skill-frontmatter-template.md`: Standardized SKILL.md frontmatter
3. `pattern-template.md`: Template for knowledge/patterns/*.md files
4. `principle-template.md`: Template for knowledge/principles/*.md files
[List all template files with brief descriptions]

**Best Practices** (`knowledge/best-practices/`):
1. `frontmatter-generation.md`: How to automate metadata creation
2. `link-validation.md`: How to validate cross-references
3. `example-extraction.md`: How to identify and extract runnable examples
[List all best practice files with brief descriptions]

**Automation Tools** (`scripts/`):
1. `extract-patterns.sh`: Automated pattern extraction from experiment artifacts
2. `generate-frontmatter.sh`: Automated frontmatter generation for SKILL.md and knowledge/*.md
3. `validate-skill.sh`: Automated quality validation (completeness, accuracy, format)
[List all script files with brief descriptions]

**Cross-Reference Map**:
- Pattern → Template: [Map which templates implement which patterns]
- Principle → Best Practice: [Map which best practices embody which principles]
- Capability → Scripts: [Map which scripts automate which capabilities]

**Reusability Index**:
```
Universal (any domain, any project): [N] assets
Domain-specific (knowledge extraction): [M] assets
Project-specific (meta-cc only): [P] assets

Reusability score: (Universal × 1.0 + Domain × 0.5 + Project × 0.2) / Total
```

### Visualization

**Trajectory Graphs**:
```
# Create visualizations (if tools available)
# Value function evolution over iterations
# Time efficiency over iterations
# Automation rate over iterations
```

**Evolution Diagram**:
```
# System evolution across iterations
# Capabilities added/modified
# Agents created (if any)
# Scripts developed
```

### Deliverables

1. **Results Report**: `results.md` with all analysis dimensions
2. **Knowledge Catalog**: `knowledge/INDEX.md` with complete asset inventory
3. **Visualizations**: Graphs of trajectory, evolution (if created)
4. **Reusability Assessment**: Validation results for Bootstrap-002, Bootstrap-003
5. **Meta-Learnings**: Insights about BAIME framework itself

---

## Execution Guidance

### Perspective

**Embodiment**: You are the meta-agent for knowledge extraction

**Your Role**:
- Orchestrate the extraction, transformation, validation lifecycle
- Follow capabilities exactly (read before use)
- Calculate value functions honestly
- Evolve system based on evidence only

**Your Constraints**:
- No token limits - complete, thorough analysis
- Honest assessment - low scores are acceptable if accurate
- Evidence-based evolution - no changes without justification
- Dual-layer focus - track both V_instance and V_meta

### Rigor

**Honest Dual-Layer Calculation**:

**V_instance** (every iteration):
1. Calculate V_completeness: Count patterns/principles/templates/scripts, use formula
2. Calculate V_accuracy: Sample patterns, compare to source, use formula
3. Calculate V_usability: Test Quick Start, check examples, use formula
4. Calculate V_format: Check frontmatter, structure, markdown, use formula
5. Calculate V_instance: Weighted average (0.3, 0.3, 0.2, 0.2)
6. Document evidence for each score

**V_meta** (every iteration after 0):
1. Calculate V_generality: Test on validation experiments, check domain independence
2. Calculate V_efficiency: Measure time, calculate speedup
3. Calculate V_automation: Count automated steps, calculate rate
4. Calculate V_meta: Weighted average (0.4, 0.3, 0.3)
5. Document evidence for each score

**No Shortcuts**: Calculate all components, every iteration

### Thoroughness

**No Token Limits**:
- Complete extraction (don't stop at "first few" patterns)
- Thorough validation (check all patterns, not just samples)
- Comprehensive documentation (full iteration reports)
- Detailed evidence (show your work)

**Complete Analysis**:
- Read all source material (entire results.md, all iterations/*.md)
- Extract all knowledge (patterns, principles, templates, scripts, examples)
- Validate thoroughly (completeness, accuracy, format, usability)
- Document everything (decisions, issues, measurements)

### Authenticity

**Discover, Don't Assume**:
- Don't assume you know what patterns exist - read the source
- Don't assume the methodology will work - validate it
- Don't assume high scores - calculate honestly
- Don't assume convergence - check criteria rigorously

**Data-Driven**:
- Base all decisions on measurement (not intuition)
- Base all evolution on evidence (not theory)
- Base all scores on data (not optimism)
- Base convergence on thresholds (not "feels done")

### Evaluation Protocol

**Independent Dual-Layer Assessment**:

**Instance Layer** (task quality):
- Measure: Quality of extracted skill and knowledge base entries
- Against: V_instance components (completeness, accuracy, usability, format)
- Question: "Is this a high-quality skill that users can actually use?"

**Meta Layer** (methodology quality):
- Measure: Quality of extraction methodology itself
- Against: V_meta components (generality, efficiency, automation)
- Question: "Can this methodology be reused for other experiments?"

**Convergence**:
- Both layers must meet thresholds
- Both layers must be stable
- Validation experiments must succeed
- Only then: convergence

### Honest Assessment

**Systematic Bias Avoidance**:

**Seek Disconfirming Evidence**:
- "What patterns did I miss?" (completeness)
- "What did I get wrong?" (accuracy)
- "Where will users struggle?" (usability)
- "What standards did I violate?" (format)
- "Where does this methodology fail?" (generality)

**Enumerate Gaps Explicitly**:
- List missing patterns (not just count extracted)
- List errors found (not just pass/fail)
- List usability issues (not just "works")
- List format deviations (not just "mostly correct")

**Ground Scores in Concrete Evidence**:
- V_completeness = (15 extracted / 20 available) = 0.75 ✅
- V_completeness = "most patterns extracted" = ❌ (not concrete)
- V_accuracy = (4 correct / 5 sampled) = 0.80 ✅
- V_accuracy = "looks good" = ❌ (not measured)

**Challenge High Scores**:
- If V_instance > 0.90: "What am I missing? This seems too good."
- If V_meta > 0.85: "Is this really this generalizable? Test on harder experiment."
- If all components high: "Check for measurement bias. Re-validate."

**Anti-Patterns to Avoid**:

❌ **Optimistic Estimation**:
- "Probably extracted most patterns" → Count exactly
- "Examples probably work" → Test them
- "Should be generalizable" → Validate it

❌ **Premature Convergence**:
- "Close enough to threshold" → Must meet exactly
- "Stable enough" → Check Δ < 0.05 rigorously
- "Validated on one experiment" → Must be ≥2 experiments

❌ **Unjustified Evolution**:
- "Let's add a specialized agent" → Show evidence of need (>5x gap)
- "We need more automation" → Show which manual steps are bottlenecks
- "This capability would be nice" → Show gap it addresses

❌ **Single-Layer Focus**:
- "V_instance is high, we're done" → Check V_meta too
- "Methodology is efficient" → Check instance quality too
- "Either/or" → Both layers required

### Success Indicators

**You're Doing It Right If**:
✅ Every iteration has calculated V_instance and V_meta (after Iteration 0)
✅ Scores have concrete evidence (counts, measurements, test results)
✅ Low scores are documented honestly (not hidden or rationalized)
✅ Evolution decisions cite specific evidence from previous iterations
✅ Convergence declaration includes all criteria (thresholds, validation, stability)
✅ Knowledge organization separates permanent (knowledge/) from ephemeral (data/)
✅ Capabilities are read before use (following protocol)

**You're Doing It Wrong If**:
❌ Scores are "estimated" or "approximate" (must be calculated)
❌ High scores without evidence ("looks good")
❌ Evolution without justification ("seems like we should")
❌ Convergence without validation (only tested on Bootstrap-004)
❌ Missing V_meta (only tracking V_instance)
❌ Mixing permanent and ephemeral knowledge (all in one directory)

---

## Appendix: Quick Reference

### Iteration Checklist

**Every Iteration**:
- [ ] Read ALL capabilities at start
- [ ] Read specific capability before use
- [ ] Execute lifecycle phases (collect, strategize, execute, evaluate, converge)
- [ ] Calculate V_instance (all 4 components)
- [ ] Calculate V_meta (all 3 components, except Iteration 0)
- [ ] Document evidence for all scores
- [ ] Identify gaps and root causes
- [ ] Design improvements (if continuing)
- [ ] Check convergence criteria (if potentially done)
- [ ] Write iteration report

### Value Function Quick Reference

**V_instance = 0.3×Completeness + 0.3×Accuracy + 0.2×Usability + 0.2×Format**

**V_meta = 0.4×Generality + 0.3×Efficiency + 0.3×Automation**

**Convergence**: V_instance ≥ 0.85, V_meta ≥ 0.75, Validated ≥2 experiments, Stable (Δ<0.05)

### File Locations

- Capabilities: `system/capabilities/*.md`
- Agents: `system/agents/*.md` (create only if needed)
- Iteration reports: `iterations/iteration-N.md`
- Data: `data/` (ephemeral)
- Knowledge: `knowledge/` (permanent)
- Scripts: `scripts/*.sh`
- Results: `results.md`

### Key Commands

```bash
# Read all capabilities
cat system/capabilities/*.md

# Calculate completeness
grep -c "Pattern:" knowledge/patterns/  # Extracted
grep -c "Pattern:" ../bootstrap-004-refactoring-guide/results.md  # Available

# Validate markdown
markdownlint .claude/skills/code-refactoring/*.md

# Test examples
cd .claude/skills/code-refactoring/examples && bash example-1.sh

# Run automation
bash scripts/extract-patterns.sh ../bootstrap-004-refactoring-guide/
bash scripts/validate-skill.sh .claude/skills/code-refactoring/
```

---

**END OF ITERATION-PROMPTS.md**
