# Bootstrap-004: Refactoring Guide - Iteration Prompts

**Created**: 2025-10-18
**Purpose**: Executable prompts for each iteration following BAIME v2.0 framework

---

## How to Use This Document

Each iteration has a **complete, self-contained prompt** that can be executed independently by Claude Code. The prompt includes:

1. **Context**: What has been accomplished so far
2. **Objectives**: Clear goals for this iteration
3. **Tasks**: Step-by-step execution checklist
4. **Value Assessment**: How to calculate V_instance and V_meta
5. **Outputs**: What artifacts to create

**Execution Protocol**:
1. Read the prompt for the current iteration
2. Execute all tasks systematically
3. Calculate value functions honestly
4. Document findings in `iterations/iteration-N.md`
5. Update README.md with current status
6. Assess convergence before proceeding

---

## Iteration 0: Baseline Establishment

### Context

**Experiment**: Bootstrap-004: Refactoring Guide
**Status**: Starting iteration 0 (baseline)
**Methodology**: BAIME v2.0

**Goal**: Establish baseline metrics for `internal/query/` package and calculate initial value functions.

### Objectives

1. Analyze `internal/query/` package complexity and quality
2. Establish baseline metrics (complexity, coverage, duplication, static analysis)
3. Identify and prioritize refactoring targets
4. Calculate V_instance(s₀) and V_meta(s₀)
5. Document baseline state comprehensively

### Tasks

#### Task 1: Code Complexity Analysis

Run complexity analysis tools on `internal/query/` package:

```bash
# Cyclomatic complexity (functions with complexity >10)
gocyclo -over 10 internal/query/

# Detailed complexity for all functions
gocyclo internal/query/ > data/complexity-baseline.txt

# Code duplication (threshold: 15 tokens)
dupl -threshold 15 internal/query/ > data/duplication-baseline.txt

# Static analysis issues
staticcheck ./internal/query/... > data/staticcheck-baseline.txt 2>&1 || true
go vet ./internal/query/... > data/vet-baseline.txt 2>&1 || true
```

**Document**:
- Total functions analyzed
- Functions with complexity >10 (count and list)
- Average cyclomatic complexity
- Duplication percentage
- Static analysis issue count by severity

#### Task 2: Test Coverage Analysis

Measure current test coverage:

```bash
# Generate coverage report
go test -coverprofile=data/coverage-baseline.out ./internal/query/...

# Coverage summary
go tool cover -func=data/coverage-baseline.out > data/coverage-baseline-summary.txt

# Overall coverage percentage
go tool cover -func=data/coverage-baseline.out | grep total
```

**Document**:
- Overall test coverage percentage
- Per-file coverage breakdown
- Functions without tests
- Critical paths without coverage

#### Task 3: Module Structure Analysis

Analyze module cohesion and coupling:

```bash
# List all files in internal/query/
ls -lh internal/query/

# Count lines of code
find internal/query/ -name "*.go" -exec wc -l {} + | sort -n

# Analyze imports (coupling)
grep -r "import" internal/query/ > data/imports-baseline.txt
```

**Document**:
- File count and sizes
- Total lines of code
- Import patterns (high coupling indicators)
- Module organization assessment

#### Task 4: Code Smell Identification

Identify code smells systematically:

**Manual inspection for**:
- Long functions (>50 lines)
- Large parameter lists (>5 parameters)
- Deep nesting (>3 levels)
- Magic numbers/strings
- Unclear naming
- Missing error handling
- Lack of comments on complex logic

**Document**:
- Code smell catalog (type, location, severity)
- Prioritized list of smells (high → low impact)

#### Task 5: Refactoring Target Prioritization

Create prioritized refactoring target list:

**Criteria**:
- High complexity + low test coverage → highest priority
- High duplication → medium-high priority
- Static analysis issues → medium priority
- Naming clarity issues → low-medium priority

**Output**: Prioritized list with estimated effort and impact

#### Task 6: Calculate Baseline Value Functions

##### V_instance(s₀) Calculation

**V_code_quality(s₀)**:
```
complexity_reduction = 0.0 (baseline)
duplication_reduction = 0.0 (baseline)
static_analysis_improvement = 0.0 (baseline)
naming_clarity = [subjective assessment 0.0-1.0]

V_code_quality(s₀) = 0.4×0.0 + 0.3×0.0 + 0.2×0.0 + 0.1×naming_clarity
                   = 0.1×naming_clarity
```

**V_maintainability(s₀)**:
```
test_coverage = current_coverage / 0.85
module_cohesion = [assessment based on coupling analysis 0.0-1.0]
documentation_quality = [documented_funcs / total_funcs] × clarity_factor
code_organization = [subjective assessment 0.0-1.0]

V_maintainability(s₀) = 0.4×test_coverage + 0.3×module_cohesion +
                        0.2×documentation_quality + 0.1×code_organization
```

**V_safety(s₀)**:
```
test_pass_rate = 1.0 (all tests currently pass)
behavior_preservation = 1.0 (baseline, no changes yet)
incremental_discipline = N/A (no refactoring yet)

V_safety(s₀) = 1.0 (perfect safety at baseline)
```

**V_effort(s₀)**:
```
V_effort(s₀) = 0.0 (no refactoring completed yet)
```

**V_instance(s₀)**:
```
V_instance(s₀) = 0.3×V_code_quality(s₀) + 0.3×V_maintainability(s₀) +
                 0.2×V_safety(s₀) + 0.2×V_effort(s₀)
```

**Expected**: V_instance(s₀) ≈ 0.35-0.45

##### V_meta(s₀) Calculation

**V_methodology_completeness(s₀)**:
```
Checklist completion: 0/15 items
V_methodology_completeness(s₀) ≈ 0.10-0.20 (basic observational notes only)
```

**V_methodology_effectiveness(s₀)**:
```
No methodology yet → V_methodology_effectiveness(s₀) = 0.0
```

**V_methodology_reusability(s₀)**:
```
Nothing to reuse yet → V_methodology_reusability(s₀) = 0.0
```

**V_meta(s₀)**:
```
V_meta(s₀) = 0.4×V_completeness(s₀) + 0.3×V_effectiveness(s₀) +
             0.3×V_reusability(s₀)
           ≈ 0.4×0.15 + 0.3×0.0 + 0.3×0.0
           ≈ 0.06-0.08
```

**Expected**: V_meta(s₀) ≈ 0.15-0.25 (including initial documentation setup)

#### Task 7: Document Baseline State

Create `iterations/iteration-0.md` with:

**Section 1: Baseline Metrics Summary**
- Cyclomatic complexity statistics
- Code duplication analysis
- Static analysis results
- Test coverage report
- Module structure assessment

**Section 2: Code Smell Catalog**
- Complete list of identified smells
- Severity and priority ratings
- Estimated fix effort

**Section 3: Refactoring Target List**
- Prioritized targets (high → low)
- Rationale for prioritization
- Estimated effort and impact

**Section 4: Value Function Calculations**
- V_instance(s₀) with component breakdown
- V_meta(s₀) with component breakdown
- Methodology notes

**Section 5: Next Iteration Planning**
- Selected refactoring target for Iteration 1
- Expected approach
- Success criteria

### Outputs

**Files to Create**:
- `data/complexity-baseline.txt`
- `data/duplication-baseline.txt`
- `data/staticcheck-baseline.txt`
- `data/vet-baseline.txt`
- `data/coverage-baseline.out`
- `data/coverage-baseline-summary.txt`
- `data/imports-baseline.txt`
- `iterations/iteration-0.md` (comprehensive baseline report)

**Update**:
- `README.md` (status → "Iteration 0 Complete")

### Success Criteria

- ✅ All baseline metrics collected and documented
- ✅ Refactoring targets identified and prioritized
- ✅ V_instance(s₀) and V_meta(s₀) calculated with rationale
- ✅ Clear plan for Iteration 1
- ✅ Baseline state fully documented in iteration-0.md

---

## Iteration 1: Initial Refactoring + Pattern Observation

### Context

**Previous Iteration**: Iteration 0 (baseline established)
**Baseline Values**:
- V_instance(s₀) = [value from Iteration 0]
- V_meta(s₀) = [value from Iteration 0]

**Refactoring Target**: [Highest priority target from Iteration 0]

### Objectives

1. Execute first refactoring (highest priority target)
2. Observe and document refactoring patterns
3. Begin methodology codification
4. Improve V_instance and V_meta

**BAIME Phase**: Observe (70%), Codify (20%), Automate (10%)

### Tasks

#### Task 1: Refactoring Planning

**Select Target**: [File/function from prioritized list]

**Plan Refactoring Steps**:
1. Identify specific code smells to address
2. Choose refactoring techniques (extract method, rename, simplify, etc.)
3. Plan incremental steps (small, safe transformations)
4. Define safety checkpoints (tests to run)

**Document Plan** in `artifacts/refactoring-plan-1.md`:
- Target description
- Code smells identified
- Refactoring techniques chosen
- Step-by-step plan
- Safety verification plan

#### Task 2: Pre-Refactoring Test Creation

**Apply TDD Principles** (from Bootstrap-002):
- Write tests for current behavior (if missing)
- Ensure 100% test pass rate before refactoring
- Create behavior preservation tests

```bash
# Run tests before refactoring
go test -v ./internal/query/... > data/tests-before-refactoring-1.txt
```

#### Task 3: Execute Refactoring Incrementally

**For each refactoring step**:
1. Make small change
2. Run tests immediately
3. Commit if tests pass (use git)
4. Document what was done

**Example Workflow**:
```bash
# Step 1: Extract method
# [Make code changes]
go test ./internal/query/...
git add -p
git commit -m "refactor: extract validateInput method"

# Step 2: Rename for clarity
# [Make code changes]
go test ./internal/query/...
git add -p
git commit -m "refactor: rename ambiguous variable names"
```

**Safety Protocol**:
- If any test fails → rollback immediately
- If behavior changes → investigate and fix
- Never proceed with failing tests

#### Task 4: Post-Refactoring Verification

After refactoring complete:

```bash
# Run full test suite
go test -v ./internal/query/... > data/tests-after-refactoring-1.txt

# Measure new metrics
gocyclo internal/query/ > data/complexity-iteration-1.txt
dupl -threshold 15 internal/query/ > data/duplication-iteration-1.txt
go test -coverprofile=data/coverage-iteration-1.out ./internal/query/...
go tool cover -func=data/coverage-iteration-1.out > data/coverage-iteration-1-summary.txt

# Compare with baseline
diff data/complexity-baseline.txt data/complexity-iteration-1.txt
```

#### Task 5: Pattern Observation (BAIME Observe Phase)

**Document Observed Patterns**:

**What worked well?**
- Specific refactoring techniques that were effective
- Incremental steps that felt safe
- Test strategies that caught issues early

**What was challenging?**
- Difficult refactoring scenarios
- Unexpected behavior preservation issues
- Test gaps discovered

**Reusable patterns identified**:
- Common code smells → refactoring technique mappings
- Safety verification steps
- Incremental transformation sequences

**Document** in `artifacts/pattern-observations-1.md`

#### Task 6: Begin Methodology Codification (BAIME Codify Phase)

Create initial methodology draft:

**`artifacts/methodology-draft-v1.md`**:

```markdown
# Refactoring Methodology (Draft v1)

## Process Steps

1. **Identify**: [How to find refactoring targets]
2. **Plan**: [How to plan safe refactorings]
3. **Test**: [Test creation before refactoring]
4. **Execute**: [Incremental refactoring steps]
5. **Verify**: [Safety verification procedures]

## Code Smell Catalog

[Document smells encountered and how to fix them]

## Refactoring Techniques

[Document techniques used with examples]

## Safety Checklist

- [ ] All tests pass before starting
- [ ] Incremental commits after each step
- [ ] Tests run after each change
- [ ] Behavior preservation verified
- [ ] ...
```

#### Task 7: Calculate Iteration 1 Value Functions

##### V_instance(s₁) Calculation

**V_code_quality(s₁)**:
```
complexity_reduction = (baseline_complexity - current_complexity) / baseline_complexity
duplication_reduction = (baseline_duplication - current_duplication) / baseline_duplication
static_analysis_improvement = (baseline_issues - current_issues) / baseline_issues
naming_clarity = [updated subjective assessment]

V_code_quality(s₁) = 0.4×complexity_reduction + 0.3×duplication_reduction +
                     0.2×static_analysis_improvement + 0.1×naming_clarity
```

**V_maintainability(s₁)**:
```
test_coverage = new_coverage / 0.85
module_cohesion = [updated assessment]
documentation_quality = [updated assessment]
code_organization = [updated assessment]

V_maintainability(s₁) = 0.4×test_coverage + 0.3×module_cohesion +
                        0.2×documentation_quality + 0.1×code_organization
```

**V_safety(s₁)**:
```
test_pass_rate = passing_tests / total_tests (should be 1.0)
behavior_preservation = [assessment based on verification]
incremental_discipline = [assessment of refactoring process]

V_safety(s₁) = 0.5×test_pass_rate + 0.3×behavior_preservation +
               0.2×incremental_discipline
```

**V_effort(s₁)**:
```
actual_time = [time spent on refactoring]
expected_time = [estimated ad-hoc time for same work]

V_effort(s₁) = 1.0 - (actual_time / expected_time)
```

**V_instance(s₁)**:
```
V_instance(s₁) = 0.3×V_code_quality(s₁) + 0.3×V_maintainability(s₁) +
                 0.2×V_safety(s₁) + 0.2×V_effort(s₁)
```

**Expected**: V_instance(s₁) ≈ 0.50-0.55

##### V_meta(s₁) Calculation

**V_methodology_completeness(s₁)**:
```
Checklist completion: [count items documented] / 15
Process steps documented: Yes (partial)
Examples provided: Few

V_methodology_completeness(s₁) ≈ 0.30-0.40 (structured process emerging)
```

**V_methodology_effectiveness(s₁)**:
```
Too early to measure efficiency gain accurately
Estimated based on time comparison

V_methodology_effectiveness(s₁) ≈ 0.20-0.30
```

**V_methodology_reusability(s₁)**:
```
Universal patterns identified: [percentage estimation]

V_methodology_reusability(s₁) ≈ 0.40-0.50 (some transferable patterns)
```

**V_meta(s₁)**:
```
V_meta(s₁) = 0.4×V_completeness(s₁) + 0.3×V_effectiveness(s₁) +
             0.3×V_reusability(s₁)
```

**Expected**: V_meta(s₁) ≈ 0.35-0.45

#### Task 8: Document Iteration 1 Results

Create `iterations/iteration-1.md` with:

**Section 1: Refactoring Summary**
- Target description
- Refactoring steps executed
- Metrics before/after comparison
- Challenges encountered

**Section 2: Pattern Observations**
- Successful patterns
- Challenging scenarios
- Reusable insights

**Section 3: Methodology Draft**
- Initial process steps
- Code smell catalog (partial)
- Safety checklist

**Section 4: Value Function Calculations**
- V_instance(s₁) with components and rationale
- V_meta(s₁) with components and rationale
- Comparison with s₀

**Section 5: Next Iteration Planning**
- Targets for Iteration 2
- Expected improvements
- Methodology refinement areas

### Outputs

**Files to Create**:
- `artifacts/refactoring-plan-1.md`
- `artifacts/pattern-observations-1.md`
- `artifacts/methodology-draft-v1.md`
- `data/tests-before-refactoring-1.txt`
- `data/tests-after-refactoring-1.txt`
- `data/complexity-iteration-1.txt`
- `data/duplication-iteration-1.txt`
- `data/coverage-iteration-1.out`
- `data/coverage-iteration-1-summary.txt`
- `iterations/iteration-1.md`

**Code Changes**:
- Refactored target file(s) in `internal/query/`
- New/updated tests

**Update**:
- `README.md` (status → "Iteration 1 Complete")

### Success Criteria

- ✅ At least 1 significant refactoring completed
- ✅ All tests pass (100% pass rate)
- ✅ Metrics improved (complexity, coverage, or duplication)
- ✅ Patterns observed and documented
- ✅ Methodology draft created
- ✅ V_instance(s₁) > V_instance(s₀)
- ✅ V_meta(s₁) > V_meta(s₀)

---

## Iteration 2: Methodology Codification + More Refactoring

### Context

**Previous Iterations**:
- Iteration 0: Baseline (V_instance = [s₀], V_meta = [s₀])
- Iteration 1: Initial refactoring (V_instance = [s₁], V_meta = [s₁])

**Patterns Identified**: [Summary from Iteration 1]

### Objectives

1. Continue refactoring (2-3 more targets)
2. Codify emerging patterns into comprehensive methodology
3. Create decision frameworks and catalogs
4. Identify automation opportunities

**BAIME Phase**: Observe (30%), Codify (50%), Automate (20%)

### Tasks

#### Task 1: Execute Additional Refactorings

**Select 2-3 Targets**: [From prioritized list]

**For each target**:
1. Plan refactoring (use emerging methodology)
2. Create/verify tests
3. Execute incrementally
4. Verify safety
5. Document patterns

**Apply learnings from Iteration 1**:
- Use successful patterns
- Avoid previous pitfalls
- Refine safety procedures

#### Task 2: Identify Common Patterns

**Cross-Refactoring Analysis**:
- What patterns appear across multiple refactorings?
- Which code smells recur?
- Which refactoring techniques are most effective?

**Pattern Categories**:
1. **Code Smell Patterns**: Common issues and detection
2. **Refactoring Technique Patterns**: Effective transformations
3. **Safety Verification Patterns**: Reliable testing approaches
4. **Incremental Step Patterns**: Safe transformation sequences

#### Task 3: Create Refactoring Decision Framework

**`artifacts/refactoring-decision-tree.md`**:

```markdown
# Refactoring Decision Framework

## When to Refactor?

- Cyclomatic complexity > 10 → High priority
- Code duplication > 15 tokens → Medium priority
- Test coverage < 70% → Medium priority
- Naming unclear → Low priority

## Which Technique to Use?

### For High Complexity
- If too many responsibilities → Extract Method/Class
- If deep nesting → Guard Clauses, Early Returns
- If long parameter list → Introduce Parameter Object

### For Code Duplication
- If exact duplication → Extract Method
- If similar but not identical → Parameterize Method
- If across files → Extract to Shared Module

### For Poor Naming
- If ambiguous → Rename Variable/Function
- If too generic → Add specificity
- If misleading → Rename to reflect actual behavior

[Add more decision trees based on observations]
```

#### Task 4: Create Comprehensive Code Smell Catalog

**`artifacts/code-smell-catalog.md`**:

```markdown
# Code Smell Catalog

## Category: Complexity

### Long Function
**Detection**: Function > 50 lines
**Impact**: High - hard to understand and test
**Refactoring**: Extract Method
**Example**: [Provide example from internal/query/]

### Deep Nesting
**Detection**: Nesting > 3 levels
**Impact**: High - hard to follow logic
**Refactoring**: Guard Clauses, Extract Method
**Example**: [...]

[Document all smells encountered, with examples]
```

#### Task 5: Refine Methodology Documentation

**Update `artifacts/methodology-draft-v2.md`**:

**Additions**:
- Complete process steps with decision criteria
- Comprehensive code smell catalog
- Refactoring technique guide with examples
- Safety verification checklist (detailed)
- Risk assessment framework

**Structure**:
```markdown
# Refactoring Methodology (v2)

## 1. Identification Phase
### 1.1 Code Smell Detection
[Detailed process with tools and criteria]

### 1.2 Prioritization Framework
[Decision tree for prioritizing targets]

## 2. Planning Phase
### 2.1 Refactoring Technique Selection
[Decision framework based on smell type]

### 2.2 Incremental Step Planning
[How to break down into safe steps]

### 2.3 Risk Assessment
[How to assess and mitigate risks]

## 3. Testing Phase
### 3.1 Pre-Refactoring Test Creation
[TDD principles application]

### 3.2 Behavior Preservation Tests
[How to verify no behavior changes]

## 4. Execution Phase
### 4.1 Incremental Transformation
[Step-by-step execution protocol]

### 4.2 Safety Checkpoints
[When and how to verify safety]

## 5. Verification Phase
### 5.1 Test Verification
[All tests must pass]

### 5.2 Metrics Verification
[Confirm improvements in metrics]

### 5.3 Behavior Preservation Verification
[Manual and automated checks]

## Appendices
### A. Code Smell Catalog
### B. Refactoring Technique Reference
### C. Safety Checklist
### D. Common Pitfalls and Solutions
```

#### Task 6: Identify Automation Opportunities

**Analysis**:
- Which steps are repetitive and automatable?
- Which checks should be automated?
- What tools would improve efficiency?

**Automation Plan** (`artifacts/automation-plan.md`):
1. **Code Smell Detector**: Script to detect common smells
2. **Refactoring Safety Checker**: Pre/post refactoring verification
3. **Complexity Reporter**: Automated complexity analysis
4. **Test Coverage Enforcer**: Ensure coverage thresholds

#### Task 7: Calculate Iteration 2 Value Functions

##### V_instance(s₂) Calculation

**Updated Metrics**:
- Complexity reduction: [calculate from data]
- Duplication reduction: [calculate from data]
- Static analysis improvement: [calculate from data]
- Test coverage: [calculate from data]

**V_instance(s₂)**: [Full calculation with components]

**Expected**: V_instance(s₂) ≈ 0.60-0.65

##### V_meta(s₂) Calculation

**V_methodology_completeness(s₂)**:
```
Checklist completion: [count] / 15 items
Process steps: Comprehensive
Decision criteria: Defined
Examples: Provided

V_methodology_completeness(s₂) ≈ 0.55-0.65 (comprehensive documentation)
```

**V_methodology_effectiveness(s₂)**:
```
Efficiency gain: [measure based on time comparisons]
Quality improvement: [measure based on metrics]

V_methodology_effectiveness(s₂) ≈ 0.50-0.60
```

**V_methodology_reusability(s₂)**:
```
Universal patterns: [percentage]
Language-agnostic components: [assessment]

V_methodology_reusability(s₂) ≈ 0.60-0.70
```

**V_meta(s₂)**: [Full calculation]

**Expected**: V_meta(s₂) ≈ 0.55-0.65

#### Task 8: Document Iteration 2 Results

Create `iterations/iteration-2.md` with all sections as in Iteration 1, plus:

**Section 6: Cross-Refactoring Analysis**
- Common patterns identified
- Decision framework rationale
- Automation opportunities

### Outputs

**Files to Create**:
- `artifacts/methodology-draft-v2.md` (updated)
- `artifacts/refactoring-decision-tree.md`
- `artifacts/code-smell-catalog.md`
- `artifacts/automation-plan.md`
- `iterations/iteration-2.md`
- [Data files for each refactoring]

**Code Changes**:
- 2-3 additional refactored modules

**Update**:
- `README.md` (status → "Iteration 2 Complete")

### Success Criteria

- ✅ 2-3 additional refactorings completed
- ✅ Common patterns identified and documented
- ✅ Decision framework created
- ✅ Code smell catalog comprehensive
- ✅ Methodology v2 significantly improved
- ✅ V_instance(s₂) > V_instance(s₁)
- ✅ V_meta(s₂) > V_meta(s₁)
- ✅ V_meta(s₂) ≥ 0.60

---

## Iteration 3: Automation Introduction

### Context

**Previous Iterations**:
- Iteration 0: Baseline
- Iteration 1: Initial refactoring, patterns observed
- Iteration 2: Methodology codification, 2-3 more refactorings

**Current State**:
- V_instance(s₂) = [value]
- V_meta(s₂) = [value]
- Methodology v2 documented
- Automation plan created

### Objectives

1. Create automation tools for repeated refactoring patterns
2. Complete remaining refactorings in `internal/query/`
3. Finalize methodology documentation
4. Plan multi-context validation

**BAIME Phase**: Observe (20%), Codify (30%), Automate (50%)

### Tasks

#### Task 1: Implement Automation Tools

##### Tool 1: Code Smell Detector

**`scripts/detect-code-smells.sh`**:

```bash
#!/bin/bash
# Automated code smell detection for Go projects

# Usage: ./scripts/detect-code-smells.sh <package>

PACKAGE=${1:-"./..."}

echo "=== Code Smell Detection Report ==="
echo

echo "1. Cyclomatic Complexity (>10)"
gocyclo -over 10 $PACKAGE

echo
echo "2. Code Duplication (>15 tokens)"
dupl -threshold 15 $PACKAGE

echo
echo "3. Static Analysis Issues"
staticcheck $PACKAGE
go vet $PACKAGE

echo
echo "4. Test Coverage"
go test -cover $PACKAGE
```

##### Tool 2: Refactoring Safety Checker

**`scripts/refactoring-safety-check.sh`**:

```bash
#!/bin/bash
# Pre/post refactoring safety verification

echo "=== Refactoring Safety Check ==="

# Run full test suite
echo "1. Running tests..."
if ! go test ./...; then
    echo "❌ Tests failed - refactoring not safe"
    exit 1
fi

# Check test coverage
echo "2. Checking coverage..."
COVERAGE=$(go test -cover ./... 2>&1 | grep -oP '\d+\.\d+%' | head -1 | tr -d '%')
if (( $(echo "$COVERAGE < 80" | bc -l) )); then
    echo "⚠️  Coverage below 80% ($COVERAGE%)"
fi

# Run linters
echo "3. Running linters..."
staticcheck ./...
go vet ./...

echo "✅ Safety check complete"
```

##### Tool 3: Complexity Reporter

**`scripts/complexity-report.sh`**:

```bash
#!/bin/bash
# Generate comprehensive complexity report

PACKAGE=${1:-"./internal/query"}
OUTPUT=${2:-"complexity-report.txt"}

echo "=== Complexity Report ===" > $OUTPUT
echo "Generated: $(date)" >> $OUTPUT
echo >> $OUTPUT

echo "Cyclomatic Complexity:" >> $OUTPUT
gocyclo $PACKAGE >> $OUTPUT

echo >> $OUTPUT
echo "Summary Statistics:" >> $OUTPUT
gocyclo $PACKAGE | awk '{sum+=$1; count++} END {print "Average:", sum/count}'
```

##### Tool 4: Refactoring Progress Tracker

**`scripts/refactoring-progress.sh`**:

```bash
#!/bin/bash
# Track refactoring progress against baseline

BASELINE_DIR="data"
CURRENT_DIR="data"

echo "=== Refactoring Progress ==="

# Complexity improvement
echo "1. Complexity Reduction:"
echo "   Baseline: $(grep -c "^" $BASELINE_DIR/complexity-baseline.txt) high-complexity functions"
echo "   Current: $(gocyclo -over 10 internal/query/ | grep -c "^") high-complexity functions"

# Coverage improvement
echo "2. Test Coverage:"
BASELINE_COV=$(grep "total" $BASELINE_DIR/coverage-baseline-summary.txt | awk '{print $NF}')
CURRENT_COV=$(go tool cover -func=data/coverage-latest.out | grep "total" | awk '{print $NF}')
echo "   Baseline: $BASELINE_COV"
echo "   Current: $CURRENT_COV"

# Duplication reduction
echo "3. Code Duplication:"
echo "   Baseline: $(grep -c "^" $BASELINE_DIR/duplication-baseline.txt) duplication issues"
echo "   Current: $(dupl -threshold 15 internal/query/ | grep -c "^") duplication issues"
```

#### Task 2: Complete Remaining Refactorings

**Using Automation Tools**:
1. Run `detect-code-smells.sh` to find remaining issues
2. Plan refactorings for remaining targets
3. Use `refactoring-safety-check.sh` before and after each refactoring
4. Track progress with `refactoring-progress.sh`

**Goal**: Complete refactoring of entire `internal/query/` package

#### Task 3: Finalize Methodology Documentation

**Create `artifacts/methodology-final.md`**:

**Complete all sections**:
- Process steps (all 5 phases fully documented)
- Decision frameworks (comprehensive)
- Code smell catalog (complete with examples)
- Refactoring technique reference (all techniques with examples)
- Safety checklist (detailed)
- Automation tools usage guide
- Common pitfalls and solutions
- Cross-language adaptation notes

**Quality checklist**:
- [ ] All 15 completeness checklist items addressed
- [ ] Examples provided for all techniques
- [ ] Edge cases documented
- [ ] Failure modes and solutions described
- [ ] Tool usage explained
- [ ] Cross-language notes included

#### Task 4: Plan Multi-Context Validation

**Select Validation Context**: Choose different package (e.g., `internal/parser/`)

**Validation Plan** (`artifacts/validation-plan.md`):
1. Apply methodology to new context
2. Measure adaptation effort (% of methodology needing modification)
3. Track efficiency (time vs ad-hoc estimation)
4. Document transferability challenges
5. Refine methodology for universality

#### Task 5: Calculate Iteration 3 Value Functions

##### V_instance(s₃) Calculation

**Metrics**: [Full refactoring completion assessment]

**Expected**: V_instance(s₃) ≈ 0.70-0.75

##### V_meta(s₃) Calculation

**V_methodology_completeness(s₃)**:
```
Checklist: 13-15 / 15 items complete
Documentation: Comprehensive and final

V_methodology_completeness(s₃) ≈ 0.70-0.80
```

**V_methodology_effectiveness(s₃)**:
```
Efficiency gain: Measured with automation (estimate 5x+)
Quality improvement: Significant

V_methodology_effectiveness(s₃) ≈ 0.65-0.75
```

**V_methodology_reusability(s₃)**:
```
Universal patterns: High percentage
Validation plan: Ready

V_methodology_reusability(s₃) ≈ 0.70-0.80
```

**V_meta(s₃)**: [Full calculation]

**Expected**: V_meta(s₃) ≈ 0.70-0.75

#### Task 6: Document Iteration 3 Results

Create `iterations/iteration-3.md` with all sections, plus:

**Section 7: Automation Implementation**
- Tools created and usage
- Efficiency improvements measured
- Automation impact on workflow

**Section 8: Methodology Finalization**
- Complete documentation overview
- Completeness checklist verification

### Outputs

**Files to Create**:
- `scripts/detect-code-smells.sh`
- `scripts/refactoring-safety-check.sh`
- `scripts/complexity-report.sh`
- `scripts/refactoring-progress.sh`
- `artifacts/methodology-final.md`
- `artifacts/validation-plan.md`
- `iterations/iteration-3.md`

**Code Changes**:
- Complete refactoring of `internal/query/`

**Update**:
- `README.md` (status → "Iteration 3 Complete, Preparing Validation")

### Success Criteria

- ✅ 4 automation tools implemented and tested
- ✅ All refactoring targets in `internal/query/` completed
- ✅ Methodology documentation finalized (15/15 checklist items)
- ✅ Multi-context validation plan ready
- ✅ V_instance(s₃) ≥ 0.70
- ✅ V_meta(s₃) ≥ 0.70

---

## Iteration 4: Multi-Context Validation

### Context

**Previous Iterations**:
- Iterations 0-3: Methodology developed, `internal/query/` refactored
- Current: V_instance(s₃) ≈ 0.70-0.75, V_meta(s₃) ≈ 0.70-0.75

**Validation Context**: [Selected package, e.g., `internal/parser/`]

### Objectives

1. Apply methodology to new context (different package)
2. Measure transferability and adaptation effort
3. Refine methodology for universality
4. Assess convergence possibility

**BAIME Phase**: Validation (70%), Refinement (30%)

### Tasks

#### Task 1: Apply Methodology to New Context

**Select Context**: `internal/parser/` (or another appropriate package)

**Full Methodology Application**:
1. **Identification Phase**: Use code smell detector
2. **Planning Phase**: Use decision framework
3. **Testing Phase**: Create behavior tests
4. **Execution Phase**: Incremental refactoring
5. **Verification Phase**: Safety checks

**Track**:
- Which methodology steps transfer directly (0% modification)
- Which steps need minor adaptation (<20% modification)
- Which steps need significant adaptation (>20% modification)
- Time spent vs expected (efficiency measurement)

#### Task 2: Measure Transferability

**Adaptation Effort Analysis**:

```
Components requiring adaptation:
1. [Component A]: [% modification needed], [reason]
2. [Component B]: [% modification needed], [reason]
...

Overall adaptation effort = Σ(component_weight × modification_%)
```

**Calculate V_methodology_reusability**:
```
reusability = 1.0 - (adaptation_effort / 100)

Target: >80% reusability (score ≥ 0.80)
```

#### Task 3: Document Transfer Challenges

**`artifacts/transfer-challenges.md`**:

**Challenges Encountered**:
1. **Language-specific issues**: [e.g., Go-specific idioms]
2. **Domain-specific issues**: [e.g., parser vs query engine differences]
3. **Tool limitations**: [e.g., tools not applicable to new context]

**Solutions Applied**:
- How challenges were addressed
- Methodology refinements made

#### Task 4: Refine Methodology for Universality

**Update `artifacts/methodology-final-v2.md`**:

**Additions**:
- Universal principles (language-agnostic)
- Language-specific adaptations section
- Domain-specific considerations
- Transfer guidance for new contexts

**Structure Enhancement**:
```markdown
## Universal Principles
[Core principles that transfer 100%]

## Language-Specific Adaptations
### Go
[Go-specific considerations]

### Python
[How to adapt for Python]

### Rust
[How to adapt for Rust]

### Java
[How to adapt for Java]

## Domain-Specific Considerations
[How methodology adapts across domains]

## Transfer Guide
[Step-by-step guide for applying to new contexts]
```

#### Task 5: Calculate Iteration 4 Value Functions

##### V_instance(s₄) Calculation

**Context 1 (internal/query/)**: Fully refactored
**Context 2 (validation context)**: Partially refactored

**Combined Assessment**: [How to weight both contexts]

**Expected**: V_instance(s₄) ≈ 0.80-0.85

##### V_meta(s₄) Calculation

**V_methodology_completeness(s₄)**:
```
All 15 items complete + transfer guidance

V_methodology_completeness(s₄) ≈ 0.80-0.90
```

**V_methodology_effectiveness(s₄)**:
```
Efficiency validated in new context
Quality improvements confirmed

V_methodology_effectiveness(s₄) ≈ 0.75-0.85
```

**V_methodology_reusability(s₄)**:
```
Measured transferability: [calculated %]

V_methodology_reusability(s₄) = [based on measured adaptation effort]

Target: ≥ 0.80
```

**V_meta(s₄)**: [Full calculation]

**Expected**: V_meta(s₄) ≈ 0.75-0.85

#### Task 6: Assess Convergence

**Check All Criteria**:
```
V_instance(s₄) ≥ 0.80? [Yes/No]
V_meta(s₄) ≥ 0.80? [Yes/No]
M₄ == M₃? [Yes - M₀ stable]
A₄ == A₃? [Yes/No - check agent set]
ΔV_instance < 0.02 (s₃ to s₄)? [Yes/No]
ΔV_meta < 0.02 (s₃ to s₄)? [Yes/No]
```

**Decision**:
- If ALL criteria met → CONVERGED (proceed to final documentation)
- If NOT converged → Plan Iteration 5 focusing on gaps

#### Task 7: Document Iteration 4 Results

Create `iterations/iteration-4.md` with all sections, plus:

**Section 9: Multi-Context Validation**
- Validation context description
- Methodology application results
- Transferability measurements
- Challenges and solutions

**Section 10: Convergence Assessment**
- All criteria evaluation
- Convergence decision
- Next steps (if not converged)

### Outputs

**Files to Create**:
- `artifacts/transfer-challenges.md`
- `artifacts/methodology-final-v2.md`
- `iterations/iteration-4.md`

**Code Changes**:
- Refactorings in validation context

**Update**:
- `README.md` (status → "Iteration 4 Complete" or "CONVERGED")

### Success Criteria

- ✅ Methodology applied to new context successfully
- ✅ Transferability ≥ 80% (adaptation effort ≤ 20%)
- ✅ Methodology refined for universality
- ✅ V_instance(s₄) ≥ 0.80
- ✅ V_meta(s₄) ≥ 0.80 (ideally)
- ✅ Convergence assessed with clear decision

---

## Iteration 5+: Convergence Refinement (If Needed)

### Context

**Previous Iterations**: 0-4 completed
**Current State**:
- V_instance(s₄) = [value]
- V_meta(s₄) = [value]
- Convergence gaps: [list what needs improvement]

### Objectives

**Focus on Remaining Gaps**:
- If V_instance < 0.80: Additional refactoring or metric improvements
- If V_meta < 0.80: Methodology documentation or validation gaps
- If ΔV too large: Continue refinement until stability

### Tasks

**Adaptive Based on Gaps**:

#### Scenario A: V_instance Gap

**Actions**:
1. Complete remaining high-impact refactorings
2. Improve test coverage to 85%+
3. Eliminate remaining duplication
4. Verify all metrics meet targets

#### Scenario B: V_meta Gap

**Actions**:
1. Complete missing methodology documentation
2. Add more examples and edge cases
3. Enhance cross-language adaptation guidance
4. Validate transferability in additional contexts
5. Improve automation tool coverage

#### Scenario C: Stability Gap

**Actions**:
1. Make smaller, incremental improvements
2. Focus on polishing existing work
3. Verify all deliverables are complete
4. Allow value functions to stabilize

### Value Function Calculation

**Calculate V_instance(s₅) and V_meta(s₅)** with focus on gap closure

**Convergence Check**:
```
V_instance(s₅) ≥ 0.80? [Yes/No]
V_meta(s₅) ≥ 0.80? [Yes/No]
M₅ == M₄? [Yes]
A₅ == A₄? [Yes]
ΔV_instance(s₄ to s₅) < 0.02? [Yes/No]
ΔV_meta(s₄ to s₅) < 0.02? [Yes/No]
```

**If NOT converged**: Plan Iteration 6 (rare, based on EXPERIMENTS-OVERVIEW.md)

### Outputs

- `iterations/iteration-5.md`
- Any remaining artifacts
- Updated code/documentation

### Success Criteria

- ✅ All convergence criteria met
- ✅ V_instance(s_N) ≥ 0.80
- ✅ V_meta(s_N) ≥ 0.80
- ✅ System stable (M and A unchanged)
- ✅ Value functions stable (Δ < 0.02)
- ✅ Ready for final results report

---

## Final: Results Report

### When to Execute

Execute this section only after **FULL CONVERGENCE** achieved.

### Objectives

1. Create comprehensive results report
2. Summarize experiment outcomes
3. Document final methodology
4. Provide usage examples

### Tasks

#### Task 1: Create `results.md`

**Structure**:

```markdown
# Bootstrap-004: Refactoring Guide - Results

## Experiment Summary
- Start date: [date]
- Completion date: [date]
- Total iterations: [N]
- Total duration: [hours]

## Final Values
- V_instance(s_N) = [value]
- V_meta(s_N) = [value]
- Convergence iteration: [N]

## Instance Layer Outcomes
### Refactoring Accomplishments
- Cyclomatic complexity reduction: [X%]
- Test coverage improvement: [X%]
- Code duplication elimination: [X%]
- Static analysis issues resolved: [X]

### Code Quality Improvements
[Detailed metrics comparison]

## Meta Layer Outcomes
### Methodology Deliverable
- Complete refactoring methodology: ✅
- Decision frameworks: ✅
- Code smell catalog: ✅
- Automation tools: [X] tools created
- Multi-context validation: ✅

### Effectiveness Measurements
- Efficiency gain: [X]x speedup
- Quality improvement: [X]%
- Transferability: [X]%

## Scientific Contributions
[What this experiment taught about refactoring methodology]

## Transferability Assessment
[How methodology applies to other languages/domains]

## Lessons Learned
[Key insights from experiment]

## Usage Examples
[How to use the methodology in practice]
```

#### Task 2: Package Methodology for Reuse

**Create standalone methodology guide**: `artifacts/REFACTORING-METHODOLOGY.md`

**Audience**: Any developer wanting to refactor code systematically

**Contents**:
- Self-contained refactoring guide
- All decision frameworks
- Code smell catalog
- Tool usage instructions
- Examples and case studies

#### Task 3: Update EXPERIMENTS-OVERVIEW.md

Add Bootstrap-004 to completed experiments section with:
- Status: ✅ COMPLETED
- Values achieved
- Iterations taken
- Key results
- Transferability percentage

#### Task 4: Create Usage Examples

**`artifacts/usage-examples.md`**:

Provide 3-5 examples of applying methodology:
1. Go project refactoring
2. Simulated Python adaptation
3. Simulated Rust adaptation
4. Large codebase application
5. Legacy code refactoring

### Outputs

**Files to Create**:
- `results.md` (comprehensive final report)
- `artifacts/REFACTORING-METHODOLOGY.md` (standalone guide)
- `artifacts/usage-examples.md`

**Updates**:
- `README.md` (status → "✅ COMPLETED")
- `experiments/EXPERIMENTS-OVERVIEW.md` (add Bootstrap-004 results)

### Success Criteria

- ✅ Complete results report created
- ✅ Methodology packaged for reuse
- ✅ Usage examples provided
- ✅ Experiment properly documented in overview
- ✅ All deliverables finalized

---

## Appendix: Quick Reference

### Value Function Calculations

**V_instance(s)**:
```
V_instance(s) = 0.3·V_code_quality + 0.3·V_maintainability +
                0.2·V_safety + 0.2·V_effort

Where:
  V_code_quality = 0.4·complexity_reduction + 0.3·duplication_reduction +
                   0.2·static_improvement + 0.1·naming_clarity

  V_maintainability = 0.4·(coverage/0.85) + 0.3·cohesion +
                      0.2·doc_quality + 0.1·organization

  V_safety = 0.5·test_pass_rate + 0.3·behavior_preservation +
             0.2·incremental_discipline

  V_effort = 1.0 - (actual_time / expected_time)
```

**V_meta(s)**:
```
V_meta(s) = 0.4·V_completeness + 0.3·V_effectiveness + 0.3·V_reusability

Where:
  V_completeness = checklist_items / 15 (with adjustments)
  V_effectiveness = 0.5·efficiency_gain + 0.5·quality_improvement
  V_reusability = 1.0 - (adaptation_effort / 100)
```

### Convergence Criteria

```
CONVERGED iff ALL of the following:
  ✅ V_instance(s_N) ≥ 0.80
  ✅ V_meta(s_N) ≥ 0.80
  ✅ M_N == M_{N-1}
  ✅ A_N == A_{N-1}
  ✅ ΔV_instance < 0.02 (for 2+ iterations)
  ✅ ΔV_meta < 0.02 (for 2+ iterations)
```

### BAIME Context Allocation

- **Observe** (Iteration 1): 70% - Execute refactoring, observe patterns
- **Codify** (Iteration 2): 50% - Document methodology systematically
- **Automate** (Iteration 3): 50% - Create automation tools
- **Validate** (Iteration 4): 70% - Multi-context validation

**30/40/20/10 Overall**:
- 30% Pattern Observation
- 40% Methodology Codification
- 20% Automation
- 10% Validation

---

**Document Version**: 1.0
**Created**: 2025-10-18
**Last Updated**: 2025-10-18
