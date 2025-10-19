# Bootstrap-004: Refactoring Guide - Iteration Prompts

**Experiment**: Develop systematic code refactoring methodology through observation of agent refactoring patterns

**Domain**: Code Refactoring Methodology

**Meta-Objective**: Create transferable refactoring methodology applicable across codebases and languages

**Instance Objective**: Refactor `internal/query/` package (~500 lines) to improve code quality, reduce cyclomatic complexity by 30%, achieve 85% test coverage

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Value Functions](#value-functions)
3. [Iteration 0: Baseline Establishment](#iteration-0-baseline-establishment)
4. [Iteration N: Subsequent Iterations](#iteration-n-subsequent-iterations)
5. [Knowledge Organization](#knowledge-organization)
6. [Results Analysis Template](#results-analysis-template)
7. [Execution Guidance](#execution-guidance)

---

## Architecture Overview

### Meta-Agent System (Modular Capabilities)

Create separate capability files in `experiments/bootstrap-004-refactoring-guide/capabilities/`:

**Lifecycle Capabilities**:
- `collect-refactoring-data.md`: Extract code metrics, complexity data, test coverage, smell patterns
- `analyze-refactoring-gaps.md`: Identify quality gaps, refactoring opportunities, risk assessment
- `plan-refactoring-strategy.md`: Design safe refactoring sequences, incremental steps, rollback plans
- `execute-refactoring.md`: Apply refactoring transformations, maintain test discipline
- `evaluate-refactoring-quality.md`: Calculate dual-layer value functions, assess methodology quality
- `check-convergence.md`: Validate methodology convergence against thresholds
- `evolve-refactoring-system.md`: Evidence-based agent/capability evolution

### Agent System (Specialized Executors)

Create agent files in `experiments/bootstrap-004-refactoring-guide/agents/`:

**Refactoring Agents**:
- `code-smell-detector.md`: Identify refactoring targets using static analysis
- `complexity-analyzer.md`: Measure cyclomatic complexity, identify hotspots
- `refactoring-planner.md`: Plan safe, incremental refactoring sequences
- `safety-checker.md`: Verify behavior preservation, test discipline
- `impact-analyzer.md`: Assess change impact, dependency analysis
- `test-maintainer.md`: Ensure test coverage and quality during refactoring

**Note**: Start with generic meta-agent in Iteration 0. Only create specialized agents when evidence demonstrates need (>5x performance gap or systematic capability deficiency).

### Modular Organization Principle

- **One file per component**: Each capability/agent in separate file
- **Clear interfaces**: Document inputs, outputs, preconditions
- **Independent evolution**: Track changes per component
- **Reusability**: Design for cross-project transferability

---

## Value Functions

### Instance Value Function: Refactoring Quality

**Dual-Layer Evaluation**: Instance layer measures task-specific quality. Meta layer assesses methodology quality. Both must meet thresholds for convergence.

```
V_instance(s) = 0.3·V_code_quality + 0.3·V_maintainability + 0.2·V_safety + 0.2·V_effort
```

#### V_code_quality (0.3 weight)

**Definition**: Code quality improvement achieved through refactoring

**Components**:
- Cyclomatic complexity reduction: (baseline - current) / baseline × target_weight
  - Target: 30% reduction
  - Measurement: `gocyclo -over 10 internal/query/`
- Code duplication elimination: duplicated_blocks_removed / total_duplicated_blocks
  - Measurement: `dupl -threshold 50 internal/query/`
- Static analysis improvements: warnings_fixed / total_warnings
  - Measurement: `staticcheck ./internal/query/...`, `go vet ./internal/query/...`

**Scoring Rubric**:
- **1.0**: ≥30% complexity reduction, zero duplication, zero static warnings
- **0.8**: 20-29% complexity reduction, <5% duplication, <3 warnings
- **0.6**: 10-19% complexity reduction, <10% duplication, <10 warnings
- **0.4**: 5-9% complexity reduction, <20% duplication, <20 warnings
- **0.2**: 1-4% complexity reduction, some progress on duplication/warnings
- **0.0**: No measurable improvement

**Evidence Requirements**:
- Before/after complexity reports with specific metrics
- Duplication analysis showing removed blocks
- Static analysis output showing warning reduction
- Concrete numbers for all measurements

#### V_maintainability (0.3 weight)

**Definition**: Long-term code maintainability and understandability

**Components**:
- Test coverage: current_coverage / target_coverage (85%)
  - Measurement: `go test -cover ./internal/query/...`
- Module cohesion: functions_properly_organized / total_functions
  - Assessment: Single responsibility, clear boundaries, logical grouping
- Documentation quality: documented_functions / public_functions
  - Requirement: All public APIs with GoDoc comments explaining purpose, params, returns

**Scoring Rubric**:
- **1.0**: ≥85% coverage, perfect cohesion, complete documentation
- **0.8**: 75-84% coverage, good cohesion, >90% documentation
- **0.6**: 65-74% coverage, acceptable cohesion, >75% documentation
- **0.4**: 55-64% coverage, some cohesion issues, >50% documentation
- **0.2**: 45-54% coverage, poor cohesion, <50% documentation
- **0.0**: <45% coverage or significant maintainability issues

**Evidence Requirements**:
- Coverage report with line-by-line breakdown
- Module structure analysis showing organization
- Documentation coverage scan
- Specific examples of cohesion improvements

#### V_safety (0.2 weight)

**Definition**: Refactoring safety and behavior preservation

**Components**:
- Test pass rate: passing_tests / total_tests (must be 100%)
- Behavior preservation: refactoring_steps_verified / total_steps
  - Every step must maintain all tests passing
- Incremental discipline: commits_with_passing_tests / total_commits
  - Every commit should be safe to deploy
- Rollback capability: git_discipline_score (clean history, revertible changes)

**Scoring Rubric**:
- **1.0**: 100% tests pass, all steps verified, perfect git discipline
- **0.8**: 100% tests pass, >95% steps verified, good git discipline
- **0.6**: 100% tests pass, >90% steps verified, acceptable git discipline
- **0.4**: 98-99% tests pass, some verification gaps
- **0.2**: 95-97% tests pass, significant verification gaps
- **0.0**: <95% tests pass or broken tests after refactoring

**Evidence Requirements**:
- Test execution logs showing 100% pass rate
- Git commit history demonstrating incremental safety
- Documentation of verification process per step
- Evidence of rollback capability (clean git history)

#### V_effort (0.2 weight)

**Definition**: Refactoring efficiency and time investment

**Components**:
- Time efficiency: baseline_time / actual_time (target: 5-10x speedup)
  - Baseline: Ad-hoc refactoring estimate
  - Actual: Measured time with methodology
- Automation utilization: automated_checks / total_checks
  - Leverage tooling for detection and verification
- Rework minimization: clean_refactorings / total_refactorings
  - Minimize steps requiring rollback or rework

**Scoring Rubric**:
- **1.0**: ≥10x speedup, >90% automation, minimal rework
- **0.8**: 7-9x speedup, 75-90% automation, <10% rework
- **0.6**: 5-6x speedup, 60-75% automation, <20% rework
- **0.4**: 3-4x speedup, 40-60% automation, <30% rework
- **0.2**: 2x speedup, <40% automation, significant rework
- **0.0**: No speedup or methodology slower than ad-hoc

**Evidence Requirements**:
- Time tracking data comparing methodology vs baseline estimate
- List of automated tools used and coverage
- Count of rollbacks or rework steps with justification
- Concrete metrics for all efficiency claims

**Instance Value Calculation Protocol**:
1. Collect quantitative evidence for each component
2. Calculate component scores using rubrics
3. Weight and sum: V_instance = Σ(weight_i × score_i)
4. Document evidence trail for all scores
5. Challenge optimistic scores with disconfirming evidence

**Convergence Threshold**: V_instance ≥ 0.75 (sustained for 2 iterations)

---

### Meta Value Function: Methodology Quality

```
V_meta(s) = 0.4·V_completeness + 0.3·V_effectiveness + 0.3·V_reusability
```

#### V_completeness (0.4 weight)

**Definition**: Coverage of refactoring methodology lifecycle

**Assessment Rubric**:

**Detection Phase** (0.25):
- **Exceptional (1.0)**: Comprehensive smell taxonomy (8+ categories), automated detection tooling, prioritization framework, validated patterns
- **Strong (0.75)**: Good taxonomy (5-7 categories), semi-automated detection, basic prioritization
- **Acceptable (0.5)**: Basic taxonomy (3-4 categories), manual detection, ad-hoc prioritization
- **Weak (0.25)**: Minimal taxonomy (1-2 categories), inconsistent detection
- **Missing (0.0)**: No systematic detection approach

**Planning Phase** (0.25):
- **Exceptional (1.0)**: Comprehensive refactoring patterns (10+ types), safety protocols, incremental sequencing, rollback strategies, validated plans
- **Strong (0.75)**: Good patterns (6-9 types), safety guidelines, basic sequencing
- **Acceptable (0.5)**: Basic patterns (3-5 types), some safety consideration
- **Weak (0.25)**: Minimal patterns (1-2 types), limited safety planning
- **Missing (0.0)**: No systematic planning approach

**Execution Phase** (0.25):
- **Exceptional (1.0)**: Detailed transformation recipes, TDD integration, continuous verification, git discipline protocols, automation support
- **Strong (0.75)**: Good transformation guidance, test requirements, verification steps
- **Acceptable (0.5)**: Basic transformation steps, some test coverage
- **Weak (0.25)**: Minimal guidance, inconsistent testing
- **Missing (0.0)**: No systematic execution approach

**Verification Phase** (0.25):
- **Exceptional (1.0)**: Multi-layer validation (tests, metrics, behavior), automated regression detection, quality gates, rollback triggers
- **Strong (0.75)**: Good validation coverage, some automation, clear criteria
- **Acceptable (0.5)**: Basic validation, manual checks, informal criteria
- **Weak (0.25)**: Minimal validation, inconsistent checks
- **Missing (0.0)**: No systematic verification approach

**Evidence Requirements**:
- Document each phase with concrete artifacts
- Provide examples demonstrating methodology application
- Show progression across iterations
- Enumerate gaps explicitly

**Convergence Threshold**: V_completeness ≥ 0.60

---

#### V_effectiveness (0.3 weight)

**Definition**: Demonstrated methodology impact on refactoring outcomes

**Assessment Rubric**:

**Quality Improvement** (0.33):
- **Exceptional (1.0)**: Consistent quality gains (≥3 examples), quantified improvements, before/after metrics
- **Strong (0.75)**: Good quality gains (2 examples), measurable improvements
- **Acceptable (0.5)**: Some quality gains (1 example), qualitative improvements
- **Weak (0.25)**: Unclear quality impact
- **Missing (0.0)**: No demonstrated quality improvement

**Safety Record** (0.33):
- **Exceptional (1.0)**: Zero breaking changes, 100% test pass rate, clean rollback capability, documented verification
- **Strong (0.75)**: Minimal issues (<5% incidents), strong test discipline
- **Acceptable (0.5)**: Some issues (5-10% incidents), acceptable test coverage
- **Weak (0.25)**: Frequent issues (>10% incidents), poor test discipline
- **Missing (0.0)**: No safety tracking or major breakages

**Efficiency Gains** (0.33):
- **Exceptional (1.0)**: ≥10x speedup demonstrated, high automation, minimal rework
- **Strong (0.75)**: 5-9x speedup, good automation
- **Acceptable (0.5)**: 3-4x speedup, some automation
- **Weak (0.25)**: <3x speedup, minimal automation
- **Missing (0.0)**: No efficiency improvement

**Evidence Requirements**:
- Quantitative metrics for all effectiveness claims
- Concrete examples with before/after comparisons
- Time tracking data or effort estimates
- Safety incident log (if any)

**Convergence Threshold**: V_effectiveness ≥ 0.60

---

#### V_reusability (0.3 weight)

**Definition**: Methodology transferability across contexts

**Assessment Rubric**:

**Language Independence** (0.33):
- **Exceptional (1.0)**: Principles apply to 5+ languages (Go, Python, JavaScript, Rust, Java), language-agnostic patterns documented
- **Strong (0.75)**: Principles apply to 3-4 languages, mostly transferable
- **Acceptable (0.5)**: Principles apply to 2 languages, some language-specific adaptations needed
- **Weak (0.25)**: Mostly language-specific, limited transferability
- **Missing (0.0)**: Completely language-specific

**Codebase Generality** (0.33):
- **Exceptional (1.0)**: Patterns apply to diverse codebases (CLI, library, web service, embedded), codebase-agnostic principles
- **Strong (0.75)**: Patterns apply to 2-3 codebase types, good generality
- **Acceptable (0.5)**: Patterns apply to 1-2 codebase types, some adaptations needed
- **Weak (0.25)**: Mostly specific to one codebase type
- **Missing (0.0)**: Completely codebase-specific

**Abstraction Quality** (0.33):
- **Exceptional (1.0)**: Universal principles extracted, minimal context-specific details, clear adaptation guidelines
- **Strong (0.75)**: Good principles, some context-specific elements, basic adaptation guidance
- **Acceptable (0.5)**: Mixed principles and specifics, limited adaptation guidance
- **Weak (0.25)**: Mostly context-specific, unclear principles
- **Missing (0.0)**: No abstraction, purely instance-specific

**Evidence Requirements**:
- Explicit analysis of transferability dimensions
- Examples showing application to different contexts
- Clear separation of universal principles vs. context-specific details
- Adaptation guidelines for different scenarios

**Convergence Threshold**: V_reusability ≥ 0.60

---

### Meta Value Calculation Protocol

1. **Independent Assessment**: Evaluate each component using rubrics WITHOUT reference to instance scores
2. **Evidence Grounding**: Support every score with concrete artifacts and examples
3. **Gap Enumeration**: Explicitly list missing elements or weaknesses
4. **Challenge High Scores**: Seek disconfirming evidence for scores ≥0.8
5. **Avoid Common Biases**:
   - Pattern existence ≠ high completeness (need coverage depth)
   - Single success ≠ high effectiveness (need consistent results)
   - Theoretical applicability ≠ high reusability (need demonstrated transfer)

**Convergence Threshold**: V_meta ≥ 0.70 (sustained for 2 iterations)

---

## Iteration 0: Baseline Establishment

**Context**: First iteration to establish baseline measurements and initial system state

**Expected Baseline**: V_instance ≈ 0.15-0.25, V_meta ≈ 0.10-0.20 (low baseline is normal and acceptable)

**Important**: This iteration focuses on honest measurement and data collection, NOT on achieving high scores. Low initial values are expected and indicate accurate assessment.

---

### System Setup

**Create Modular Architecture**:

1. **Create capability files** in `experiments/bootstrap-004-refactoring-guide/capabilities/`:
   - `collect-refactoring-data.md`: Data collection procedures
   - `evaluate-refactoring-quality.md`: Value function calculation
   - (Add others as needed, avoid premature specialization)

2. **Create initial agent file** in `experiments/bootstrap-004-refactoring-guide/agents/`:
   - `meta-agent.md`: Generic refactoring meta-agent
   - (Do NOT create specialized agents yet - wait for evidence of need)

3. **Create data directory**: `experiments/bootstrap-004-refactoring-guide/data/iteration-0/`

4. **Create knowledge directory**: `experiments/bootstrap-004-refactoring-guide/knowledge/`

**Agent Specialization Principle**: Start with generic meta-agent. Only create specialized agents when retrospective evidence shows:
- Performance gap >5x between ideal and generic agent
- Systematic capability deficiency across multiple iterations
- Clear specialization benefits that justify overhead

---

### Objectives

Execute these steps sequentially:

#### Step 1: Collect Baseline Code Metrics

**Goal**: Establish quantitative baseline for `internal/query/` package

**Tasks**:
1. **Cyclomatic Complexity**:
   ```bash
   gocyclo -over 1 internal/query/ > data/iteration-0/complexity-baseline.txt
   gocyclo -avg internal/query/ >> data/iteration-0/complexity-baseline.txt
   ```
   - Document: Total functions, average complexity, functions >10 complexity

2. **Code Duplication**:
   ```bash
   dupl -threshold 15 internal/query/ > data/iteration-0/duplication-baseline.txt
   ```
   - Document: Number of duplicate blocks, total duplicated lines

3. **Static Analysis**:
   ```bash
   staticcheck ./internal/query/... > data/iteration-0/staticcheck-baseline.txt 2>&1
   go vet ./internal/query/... > data/iteration-0/govet-baseline.txt 2>&1
   ```
   - Document: Warning counts by category

4. **Test Coverage**:
   ```bash
   go test -cover ./internal/query/... > data/iteration-0/coverage-baseline.txt
   go test -coverprofile=data/iteration-0/coverage.out ./internal/query/...
   go tool cover -func=data/iteration-0/coverage.out >> data/iteration-0/coverage-baseline.txt
   ```
   - Document: Overall percentage, per-file coverage, uncovered functions

5. **File Statistics**:
   ```bash
   find internal/query/ -name "*.go" -exec wc -l {} + > data/iteration-0/file-stats.txt
   ```
   - Document: Total lines, files count, distribution

**Save Results**: `data/iteration-0/baseline-metrics.md` with summary analysis

---

#### Step 2: Identify Code Smells

**Goal**: Catalog refactoring opportunities using manual inspection

**Tasks**:
1. Read each file in `internal/query/`:
   - `query.go`, `engine.go`, `filters.go`, `time_series.go`, etc.

2. Document code smells by category:
   - **Long Functions**: Functions >50 lines
   - **High Complexity**: Functions >10 cyclomatic complexity
   - **Duplicated Logic**: Similar code blocks across files
   - **Poor Naming**: Unclear variable/function names
   - **Missing Tests**: Uncovered functions
   - **God Objects**: Files/structs with too many responsibilities
   - **Primitive Obsession**: Over-reliance on primitive types
   - **Feature Envy**: Functions accessing other structs' data excessively

3. Prioritize smells:
   - High priority: High complexity + low coverage
   - Medium priority: Duplication + maintainability impact
   - Low priority: Cosmetic issues

**Save Results**: `data/iteration-0/code-smells.md` with categorized list

---

#### Step 3: Attempt Initial Refactoring

**Goal**: Refactor 1-2 functions using ad-hoc approach to establish baseline effort

**Tasks**:
1. Select high-priority smell (e.g., high complexity function)
2. Write tests if missing (TDD approach)
3. Apply refactoring (e.g., extract method, simplify conditionals)
4. Verify tests pass
5. Measure time taken

**Track**:
- Time spent (minutes)
- Steps taken
- Issues encountered
- Rollbacks needed

**Save Results**: `data/iteration-0/refactoring-log.md` with detailed notes

---

#### Step 4: Calculate Baseline Value Functions

**Goal**: Establish honest baseline V_instance and V_meta scores

**V_instance Calculation**:

1. **V_code_quality**:
   - Complexity reduction: (baseline - current) / baseline
   - Duplication elimination: blocks_removed / total_blocks
   - Static warnings: warnings_fixed / total_warnings
   - Expected: ~0.05-0.15 (minimal progress)

2. **V_maintainability**:
   - Test coverage: current / 85% target
   - Module cohesion: subjective assessment
   - Documentation: documented / public_functions
   - Expected: ~0.10-0.25 (partial coverage)

3. **V_safety**:
   - Test pass rate: passing / total
   - Behavior preservation: verified_steps / total_steps
   - Incremental discipline: safe_commits / total_commits
   - Expected: 0.20-0.40 (basic safety)

4. **V_effort**:
   - Time efficiency: baseline_estimate / actual_time (likely <1.0)
   - Automation: minimal at baseline
   - Rework: likely some rollbacks
   - Expected: 0.10-0.20 (inefficient)

**Calculate**: V_instance = 0.3·V_code_quality + 0.3·V_maintainability + 0.2·V_safety + 0.2·V_effort

**Expected V_instance**: 0.15-0.25

---

**V_meta Calculation**:

1. **V_completeness**:
   - Detection: Basic manual inspection (0.25-0.40)
   - Planning: Ad-hoc, no systematic approach (0.20-0.35)
   - Execution: Informal steps (0.25-0.40)
   - Verification: Basic testing (0.30-0.50)
   - Expected: 0.25-0.40

2. **V_effectiveness**:
   - Quality improvement: Minimal demonstrable gains (0.20-0.40)
   - Safety record: Basic test discipline (0.30-0.50)
   - Efficiency: No speedup yet (0.10-0.30)
   - Expected: 0.20-0.40

3. **V_reusability**:
   - Language independence: Unclear at baseline (0.20-0.40)
   - Codebase generality: Instance-specific (0.20-0.40)
   - Abstraction quality: No principles extracted (0.10-0.30)
   - Expected: 0.15-0.35

**Calculate**: V_meta = 0.4·V_completeness + 0.3·V_effectiveness + 0.3·V_reusability

**Expected V_meta**: 0.10-0.20

---

**Save Results**: `data/iteration-0/value-functions.md` with detailed calculations and evidence

**Critical**: Use rubrics rigorously. Low scores are expected and honest. Do NOT inflate scores to appear successful.

---

#### Step 5: Document Initial Problems

**Goal**: Identify gaps and inefficiencies to address in subsequent iterations

**Analyze**:
1. **Detection Gaps**: What smells were missed? What tools could help?
2. **Planning Gaps**: Was refactoring sequence optimal? What risks were overlooked?
3. **Execution Gaps**: What steps were unclear? Where did inefficiency occur?
4. **Verification Gaps**: What safety checks were missing? What could be automated?

**Document**:
- Specific problems encountered
- Hypotheses for improvement (do NOT commit to solutions yet)
- Data needed to validate hypotheses

**Save Results**: `data/iteration-0/problems-identified.md`

---

### Baseline Iteration Summary

**Deliverables**:
1. `data/iteration-0/baseline-metrics.md`: Quantitative baseline
2. `data/iteration-0/code-smells.md`: Identified refactoring targets
3. `data/iteration-0/refactoring-log.md`: Initial refactoring attempt
4. `data/iteration-0/value-functions.md`: V_instance and V_meta calculations
5. `data/iteration-0/problems-identified.md`: Gap analysis

**Expected Outcomes**:
- V_instance ≈ 0.15-0.25 (low, honest assessment)
- V_meta ≈ 0.10-0.20 (minimal methodology maturity)
- Clear problem identification
- Data-driven foundation for iteration 1

**Do NOT**:
- Inflate scores to appear successful
- Design full methodology prematurely
- Create specialized agents without evidence
- Plan evolution without retrospective validation

---

## Iteration N: Subsequent Iterations

**Context**: Execute after Iteration 0. Each iteration follows Observe-Codify-Automate-Evaluate (OCA) cycle.

**Important**: This template applies to iterations 1-N. Adapt based on previous iteration state.

---

### Pre-Iteration Setup

**Read Previous Iteration State**:

1. **Load Previous Scores**:
   - Read `data/iteration-(N-1)/value-functions.md`
   - Extract V_instance(s_{N-1}), V_meta(s_{N-1})
   - Note: These are starting points, NOT targets to beat artificially

2. **Load Identified Problems**:
   - Read `data/iteration-(N-1)/problems-identified.md`
   - Prioritize problems by impact on value functions

3. **Load System State**:
   - Review current capabilities in `capabilities/`
   - Review current agents in `agents/`
   - Assess what exists vs. what's needed

4. **Create Iteration Directory**:
   ```bash
   mkdir -p data/iteration-N/
   ```

---

### Capability Reading Protocol

**Before Starting Iteration**:
- Read ALL capability files in `capabilities/` to understand available tools
- Understand lifecycle coverage and gaps

**Before Using Specific Capability**:
- Re-read the specific capability file to ensure correct usage
- Follow capability instructions precisely

**Rationale**: Prevents capability misuse and ensures consistent execution

---

### Phase 1: Observe (Data Collection)

**Goal**: Gather empirical data about refactoring patterns and methodology gaps

**Capability**: Use `capabilities/collect-refactoring-data.md`

---

#### Task 1: Collect Code Metrics

**Execute**:
1. Run complexity analysis:
   ```bash
   gocyclo -over 1 internal/query/ > data/iteration-N/complexity-current.txt
   gocyclo -avg internal/query/ >> data/iteration-N/complexity-current.txt
   ```

2. Run duplication analysis:
   ```bash
   dupl -threshold 15 internal/query/ > data/iteration-N/duplication-current.txt
   ```

3. Run static analysis:
   ```bash
   staticcheck ./internal/query/... > data/iteration-N/staticcheck-current.txt 2>&1
   go vet ./internal/query/... > data/iteration-N/govet-current.txt 2>&1
   ```

4. Run coverage analysis:
   ```bash
   go test -cover ./internal/query/... > data/iteration-N/coverage-current.txt
   go test -coverprofile=data/iteration-N/coverage.out ./internal/query/...
   go tool cover -func=data/iteration-N/coverage.out >> data/iteration-N/coverage-current.txt
   ```

**Compare with Baseline**:
- Calculate deltas for each metric
- Identify improvements and regressions

**Save**: `data/iteration-N/metrics-comparison.md`

---

#### Task 2: Analyze Refactoring Session

**If refactoring occurred in previous iteration**:

1. **Review Refactoring Log**:
   - Read `data/iteration-(N-1)/refactoring-log.md`
   - Extract patterns: common steps, decision points, verification checks

2. **Categorize Refactorings**:
   - Extract method: Count, success rate, time spent
   - Simplify conditionals: Count, complexity reduction
   - Remove duplication: Count, lines saved
   - Rename: Count, clarity improvement
   - (Add categories as observed)

3. **Identify Pain Points**:
   - Steps requiring manual effort
   - Verification steps that were missed
   - Time sinks
   - Uncertainty points

**Save**: `data/iteration-N/refactoring-patterns.md`

---

#### Task 3: Query Meta-CC Data

**Use meta-cc MCP tools** to analyze session patterns:

1. **File Access Patterns**:
   ```
   query_file_access(file="internal/query/query.go")
   ```
   - Identify frequently edited files
   - Track read-edit patterns

2. **Error Patterns**:
   ```
   query_tools(status="error")
   ```
   - Identify tool failures during refactoring
   - Common error signatures

3. **Tool Sequences**:
   ```
   query_tool_sequences(min_occurrences=2)
   ```
   - Identify workflow patterns (Read → Edit → Bash test cycle)
   - Detect inefficient sequences

**Save**: `data/iteration-N/session-analysis.md`

---

#### Task 4: Document Observations

**Synthesize data** into actionable observations:

1. **What worked well?**
   - Effective refactoring patterns
   - Successful safety protocols
   - Efficient tool usage

2. **What didn't work?**
   - Missed smells
   - Inefficient steps
   - Verification gaps

3. **What's missing?**
   - Undetected patterns
   - Uncodified knowledge
   - Unautomated steps

**Save**: `data/iteration-N/observations.md`

---

### Phase 2: Codify (Strategy Formation)

**Goal**: Translate observations into methodology improvements

**Capability**: Use `capabilities/analyze-refactoring-gaps.md` and `capabilities/plan-refactoring-strategy.md`

---

#### Task 1: Gap Analysis

**Analyze V_instance Gaps**:

1. **V_code_quality gaps**:
   - Which complexity hotspots remain?
   - What duplication persists?
   - Which static warnings are unaddressed?

2. **V_maintainability gaps**:
   - Which functions lack tests?
   - Which modules have poor cohesion?
   - Which APIs lack documentation?

3. **V_safety gaps**:
   - Were any tests broken during refactoring?
   - Were any steps unverified?
   - Is git history clean and revertible?

4. **V_effort gaps**:
   - What manual steps consume time?
   - What could be automated?
   - Where does rework occur?

**Prioritize Gaps**:
- High impact on V_instance
- Addressable in current iteration
- Data-supported (not speculative)

**Save**: `data/iteration-N/gap-analysis.md`

---

#### Task 2: Strategy Formation

**Design iteration strategy** addressing top gaps:

1. **Detection Strategy**:
   - What smells to target?
   - What tools to use?
   - What prioritization criteria?

2. **Refactoring Strategy**:
   - What refactoring patterns to apply?
   - What sequence to follow?
   - What safety protocols to enforce?

3. **Verification Strategy**:
   - What tests to write/maintain?
   - What metrics to track?
   - What quality gates to enforce?

4. **Automation Strategy**:
   - What steps to automate?
   - What scripts/tools to create?
   - What manual oversight to retain?

**Save**: `data/iteration-N/iteration-strategy.md`

---

#### Task 3: Codify Patterns

**Extract reusable patterns** from observations:

**Pattern Template**:
```markdown
## Pattern: [Name]

**Context**: When to apply this pattern
**Problem**: What problem it solves
**Solution**: Step-by-step refactoring procedure
**Example**: Before/after code example
**Safety**: Verification steps
**Metrics**: Expected impact on V_instance components
```

**Save Patterns**: `knowledge/patterns/[pattern-name].md`

**Update Pattern Index**: `knowledge/patterns/INDEX.md`

---

### Phase 3: Automate (Execution)

**Goal**: Apply refactoring methodology and capture execution data

**Capability**: Use `capabilities/execute-refactoring.md`

**Agent**: Use appropriate agent(s) from `agents/` (start with meta-agent, invoke specialized agents only if they exist and are needed)

---

#### Task 1: Prepare Refactoring Session

**Setup**:
1. Create clean git branch:
   ```bash
   git checkout -b iteration-N-refactoring
   ```

2. Document refactoring plan in `data/iteration-N/refactoring-plan.md`:
   - Target files/functions
   - Refactoring sequence
   - Safety checkpoints
   - Expected metrics impact

3. Start refactoring log: `data/iteration-N/refactoring-log.md`

---

#### Task 2: Execute Refactoring (TDD Cycle)

**For each refactoring target**:

1. **Write/Update Tests** (if missing):
   - Create test file: `internal/query/*_test.go`
   - Cover current behavior (characterization tests)
   - Ensure tests pass before refactoring
   - Document in log: test creation time, coverage added

2. **Apply Refactoring**:
   - Follow codified pattern from `knowledge/patterns/`
   - Make incremental changes (small commits)
   - Document in log: refactoring type, steps taken, time spent

3. **Verify Safety**:
   - Run tests: `go test ./internal/query/...`
   - Must pass 100%
   - Run coverage: `go test -cover ./internal/query/...`
   - Track coverage change
   - Document in log: verification results

4. **Commit Incrementally**:
   ```bash
   git add [files]
   git commit -m "refactor: [description] - [pattern-name]"
   ```
   - Every commit must have passing tests
   - Keep commits small and focused
   - Document in log: commit hash, files changed

5. **Measure Impact**:
   - Run metrics after refactoring
   - Compare with pre-refactoring state
   - Document in log: metric deltas

**Continue** until iteration time budget exhausted or refactoring targets completed

---

#### Task 3: Document Execution

**Update Refactoring Log** with:
- Total refactorings completed
- Patterns applied (count by type)
- Time breakdown (test writing, refactoring, verification)
- Issues encountered
- Rollbacks/rework needed
- Final metrics comparison

**Save**: `data/iteration-N/refactoring-log.md`

---

### Phase 4: Evaluate

**Goal**: Calculate V_instance and V_meta, assess progress

**Capability**: Use `capabilities/evaluate-refactoring-quality.md`

---

#### Task 1: Calculate V_instance

**Follow Instance Value Function rubrics** (see [Value Functions](#value-functions) section):

1. **V_code_quality**:
   - Complexity reduction: (baseline - current) / baseline
   - Duplication elimination: removed_blocks / total_blocks
   - Static warnings: fixed_warnings / total_warnings
   - Apply rubric scoring
   - Document evidence

2. **V_maintainability**:
   - Test coverage: current_coverage / 85%
   - Module cohesion: assess organization quality
   - Documentation: documented / public_functions
   - Apply rubric scoring
   - Document evidence

3. **V_safety**:
   - Test pass rate: passing / total (must be 100%)
   - Behavior preservation: verified_steps / total_steps
   - Incremental discipline: safe_commits / total_commits
   - Apply rubric scoring
   - Document evidence

4. **V_effort**:
   - Time efficiency: estimate_baseline / actual_time
   - Automation: automated_checks / total_checks
   - Rework: clean_refactorings / total_refactorings
   - Apply rubric scoring
   - Document evidence

**Calculate**: V_instance(s_N) = 0.3·V_code_quality + 0.3·V_maintainability + 0.2·V_safety + 0.2·V_effort

**Save**: `data/iteration-N/value-instance.md` with detailed evidence

---

#### Task 2: Calculate V_meta

**Follow Meta Value Function rubrics** (see [Value Functions](#value-functions) section):

**IMPORTANT**: Evaluate V_meta INDEPENDENTLY of V_instance. Meta-layer assesses methodology quality, not task quality.

1. **V_completeness**:
   - Detection phase: Assess taxonomy, tools, prioritization (0.25 weight)
   - Planning phase: Assess patterns, safety protocols, sequencing (0.25 weight)
   - Execution phase: Assess transformation recipes, TDD, verification (0.25 weight)
   - Verification phase: Assess validation, automation, quality gates (0.25 weight)
   - Apply rubric scoring for each phase
   - Document evidence (artifacts in `knowledge/`)

2. **V_effectiveness**:
   - Quality improvement: Demonstrated gains with examples (0.33 weight)
   - Safety record: Test discipline, zero breakages (0.33 weight)
   - Efficiency gains: Measured speedup (0.33 weight)
   - Apply rubric scoring
   - Document evidence (metrics, logs)

3. **V_reusability**:
   - Language independence: How many languages would patterns apply to? (0.33 weight)
   - Codebase generality: How many codebase types? (0.33 weight)
   - Abstraction quality: Universal principles vs. context-specific details (0.33 weight)
   - Apply rubric scoring
   - Document evidence (transferability analysis)

**Calculate**: V_meta(s_N) = 0.4·V_completeness + 0.3·V_effectiveness + 0.3·V_reusability

**Save**: `data/iteration-N/value-meta.md` with detailed evidence

---

#### Task 3: Compare with Previous Iteration

**Delta Analysis**:
- ΔV_instance = V_instance(s_N) - V_instance(s_{N-1})
- ΔV_meta = V_meta(s_N) - V_meta(s_{N-1})

**Interpret**:
- Positive delta: Improvement (expected)
- Negative delta: Regression (investigate causes)
- Near-zero delta: Plateau (check for convergence or need for new approach)

**Save**: `data/iteration-N/value-comparison.md`

---

### Phase 5: Convergence Check

**Goal**: Determine if methodology has converged or requires further iteration

**Capability**: Use `capabilities/check-convergence.md`

---

#### Convergence Criteria

**Both layers must meet thresholds**:
1. **Instance Layer**: V_instance(s_N) ≥ 0.75
2. **Meta Layer**: V_meta(s_N) ≥ 0.70
3. **Stability**: Thresholds sustained for 2 consecutive iterations
4. **Diminishing Returns**: ΔV < 0.05 for both layers

**Evaluate**:

1. **Instance Convergence**:
   - Is V_instance ≥ 0.75? YES/NO
   - Evidence: [specific scores and components]
   - If NO: What gaps remain?

2. **Meta Convergence**:
   - Is V_meta ≥ 0.70? YES/NO
   - Evidence: [specific rubric assessments]
   - If NO: What methodology gaps remain?

3. **Stability Check**:
   - Is this the 2nd consecutive iteration above thresholds? YES/NO
   - If NO: Continue iterating

4. **Diminishing Returns Check**:
   - Is ΔV_instance < 0.05 AND ΔV_meta < 0.05? YES/NO
   - If YES and below thresholds: Need new approach

**Convergence Decision**:
- **CONVERGED**: All criteria met → Proceed to Results Analysis
- **NOT CONVERGED**: Continue to next iteration → Identify gaps for iteration N+1

**Save**: `data/iteration-N/convergence-check.md`

---

### Phase 6: System Evolution (If Needed)

**Goal**: Evidence-based evolution of capabilities and agents

**Capability**: Use `capabilities/evolve-refactoring-system.md`

**IMPORTANT**: Only evolve system when retrospective evidence demonstrates necessity. Do NOT evolve based on:
- Pattern matching or similarity to other experiments
- Anticipatory design or theoretical completeness
- Premature optimization

---

#### When to Evolve

**Evidence-Based Triggers**:

1. **Create New Capability**:
   - Retrospective evidence: Repeated manual workflow across 2+ iterations
   - Gap analysis: Missing lifecycle phase coverage
   - Attempted alternatives: Manual approach attempted, proven inefficient
   - Quantifiable benefit: Clear improvement in V_meta or V_instance

2. **Create Specialized Agent**:
   - Performance gap: >5x difference between ideal and generic agent
   - Systematic deficiency: Generic agent fails consistently at specific task
   - Attempted workarounds: Tried compensating with capabilities, insufficient
   - Clear specialization ROI: Demonstrated need outweighs overhead

3. **Modify Existing Component**:
   - Retrospective evidence: Component used incorrectly or incompletely 2+ times
   - Gap analysis: Component missing critical functionality
   - Usage data: Component invoked but inadequate
   - Improvement necessity: Modification required for convergence

**Anti-Patterns** (DO NOT evolve for these reasons):
- "This looks like testing methodology needs test-executor agent" (pattern matching)
- "Refactoring should have smell-detector, planner, executor" (theoretical completeness)
- "Let's create automation-builder capability proactively" (anticipatory design)

---

#### Evolution Process

**If evidence supports evolution**:

1. **Document Justification**:
   - Create `data/iteration-N/evolution-justification.md`
   - List specific evidence from retrospectives
   - Show attempted alternatives
   - Quantify expected improvement

2. **Create/Modify Component**:
   - **New Capability**: Create `capabilities/[name].md` with clear interface
   - **New Agent**: Create `agents/[name].md` with specialization scope
   - **Modify Component**: Edit existing file, track changes

3. **Update System Documentation**:
   - Update `ARCHITECTURE.md` with new component
   - Update capability/agent index files

4. **Validate Evolution**:
   - Test new component in current iteration
   - Measure impact on V_meta (completeness/effectiveness)
   - Compare with pre-evolution state

**Save**: `data/iteration-N/evolution-log.md` documenting changes and validation

---

### Iteration Summary

**Deliverables** (each iteration):
1. `data/iteration-N/observations.md`: Empirical observations
2. `data/iteration-N/gap-analysis.md`: V_instance and V_meta gaps
3. `data/iteration-N/iteration-strategy.md`: Iteration plan
4. `data/iteration-N/refactoring-log.md`: Execution log
5. `data/iteration-N/value-instance.md`: V_instance calculation with evidence
6. `data/iteration-N/value-meta.md`: V_meta calculation with evidence
7. `data/iteration-N/convergence-check.md`: Convergence assessment
8. `knowledge/patterns/*.md`: Extracted patterns (if any)
9. `data/iteration-N/evolution-log.md`: System evolution (if any)

**Expected Trajectory**:
- Iterations 1-2: Rapid V_instance improvement, moderate V_meta improvement
- Iterations 3-4: Slowing V_instance gains, continued V_meta improvement
- Iterations 5-7: Convergence approach, diminishing returns

**Do NOT**:
- Inflate scores to show artificial progress
- Evolve system without retrospective evidence
- Skip convergence checks
- Compromise honesty for appearance of success

---

## Knowledge Organization

**Goal**: Separate permanent knowledge from ephemeral iteration data

### Directory Structure

```
experiments/bootstrap-004-refactoring-guide/
├── capabilities/           # Lifecycle capabilities (modular)
│   ├── collect-refactoring-data.md
│   ├── analyze-refactoring-gaps.md
│   ├── plan-refactoring-strategy.md
│   ├── execute-refactoring.md
│   ├── evaluate-refactoring-quality.md
│   ├── check-convergence.md
│   └── evolve-refactoring-system.md
├── agents/                 # Specialized agents (evidence-based)
│   ├── meta-agent.md       # Generic agent (start here)
│   └── [specialized-agents].md  # Only if evidence supports
├── data/                   # Ephemeral iteration data
│   ├── iteration-0/
│   ├── iteration-1/
│   └── iteration-N/
├── knowledge/              # Permanent methodology knowledge
│   ├── patterns/           # Refactoring patterns
│   │   ├── INDEX.md
│   │   ├── extract-method.md
│   │   ├── simplify-conditionals.md
│   │   ├── remove-duplication.md
│   │   └── [other-patterns].md
│   ├── principles/         # Universal refactoring principles
│   │   ├── INDEX.md
│   │   ├── safety-first.md
│   │   ├── incremental-refactoring.md
│   │   ├── test-coverage-discipline.md
│   │   └── behavior-preservation.md
│   ├── templates/          # Reusable templates
│   │   ├── refactoring-checklist.md
│   │   ├── safety-verification.md
│   │   └── pattern-template.md
│   ├── best-practices/     # Context-specific practices
│   │   ├── go-refactoring.md
│   │   ├── tdd-workflow.md
│   │   └── git-discipline.md
│   └── methodology/        # Project-wide reusable methodology
│       ├── refactoring-methodology.md
│       └── transferability-analysis.md
└── ARCHITECTURE.md         # System architecture documentation
```

---

### Knowledge Extraction Guidelines

**When to Extract Knowledge**:

1. **Patterns** (`knowledge/patterns/`):
   - Refactoring technique applied successfully 2+ times
   - Clear before/after improvement
   - Generalizable beyond specific instance
   - Extract: Name, context, steps, example, safety checks

2. **Principles** (`knowledge/principles/`):
   - Universal truth discovered through iteration
   - Applies across multiple patterns
   - Not context-specific
   - Extract: Principle statement, rationale, evidence, examples

3. **Templates** (`knowledge/templates/`):
   - Reusable structure used 3+ times
   - Standardizes workflow
   - Reduces cognitive load
   - Extract: Template structure, usage instructions, example

4. **Best Practices** (`knowledge/best-practices/`):
   - Context-specific guidance (e.g., Go-specific, TDD-specific)
   - Proven effective in domain
   - Not universally applicable
   - Extract: Practice description, context, rationale, examples

5. **Methodology** (`knowledge/methodology/`):
   - Comprehensive refactoring methodology (at convergence)
   - Integrates patterns, principles, practices
   - Designed for reuse across projects
   - Extract: Full lifecycle guide, adaptation instructions

---

### Knowledge Index Maintenance

**Update INDEX.md files** in each knowledge subdirectory:

**Index Template**:
```markdown
# [Category] Index

## Overview
[Brief description of category]

## Contents

### [Item Name]
- **File**: [filename].md
- **Source**: Iteration [N]
- **Validation**: [Applied in X refactorings, Y% success rate]
- **Transferability**: [Universal / Language-specific / Domain-specific]
- **Description**: [One-line summary]

[Repeat for each item]

## Cross-References
- Related patterns: [links]
- Related principles: [links]
- Related practices: [links]
```

---

### Dual Output Principle

**Local Knowledge** (experiment-specific):
- `data/iteration-N/*`: Ephemeral iteration data
- `capabilities/*`, `agents/*`: System components for this experiment

**Project Methodology** (reusable across projects):
- `knowledge/methodology/refactoring-methodology.md`: Transferable methodology
- `knowledge/patterns/*`: Reusable patterns
- `knowledge/principles/*`: Universal principles

**At convergence**: Extract project-wide methodology to `/docs/methodology/refactoring.md` (or similar) for use beyond this experiment.

---

## Results Analysis Template

**Context**: Use this template when convergence is achieved

**Goal**: Comprehensive analysis of experiment outcomes and methodology validation

---

### 1. Convergence Summary

**Final Scores**:
- V_instance(final) = [score]
- V_meta(final) = [score]
- Iterations to convergence: [N]

**Convergence Evidence**:
- Instance threshold (≥0.75): Met in iterations [X, Y]
- Meta threshold (≥0.70): Met in iterations [X, Y]
- Stability: Sustained for [N] iterations
- Diminishing returns: ΔV < 0.05 confirmed

---

### 2. Trajectory Analysis

**V_instance Trajectory**:
```
Iteration 0: [score]  (Baseline)
Iteration 1: [score]  (Δ = +[delta])
Iteration 2: [score]  (Δ = +[delta])
...
Iteration N: [score]  (Δ = +[delta])
```

**V_meta Trajectory**:
```
Iteration 0: [score]  (Baseline)
Iteration 1: [score]  (Δ = +[delta])
...
Iteration N: [score]  (Δ = +[delta])
```

**Visualization**: Create trajectory plots if helpful

**Analysis**:
- Rate of convergence (fast/moderate/slow)
- Inflection points (where progress accelerated/decelerated)
- Correlation between V_instance and V_meta improvements

---

### 3. Instance Task Results

**Refactoring Outcomes**:

1. **Code Quality**:
   - Baseline complexity: [average]
   - Final complexity: [average]
   - Reduction: [%]
   - Duplication: [baseline blocks] → [final blocks] ([%] reduction)
   - Static warnings: [baseline count] → [final count] ([%] reduction)

2. **Maintainability**:
   - Test coverage: [baseline %] → [final %]
   - Module cohesion: [before/after assessment]
   - Documentation: [baseline %] → [final %]

3. **Safety**:
   - Test pass rate: [%] (should be 100%)
   - Behavior preservation: [verified steps / total steps]
   - Rollbacks needed: [count]

4. **Efficiency**:
   - Total refactoring time: [hours]
   - Baseline estimate: [hours]
   - Speedup: [X]x
   - Automation level: [%]

**Deliverables**:
- Refactored code in `internal/query/`
- Enhanced test suite with [%] coverage
- [N] commits with clean git history

---

### 4. Methodology Outputs

**Knowledge Artifacts**:

1. **Patterns** (`knowledge/patterns/`):
   - [Pattern 1]: [brief description, application count]
   - [Pattern 2]: [brief description, application count]
   - Total: [N] patterns extracted

2. **Principles** (`knowledge/principles/`):
   - [Principle 1]: [brief description]
   - [Principle 2]: [brief description]
   - Total: [N] principles discovered

3. **Templates** (`knowledge/templates/`):
   - [Template 1]: [brief description, usage count]
   - Total: [N] templates created

4. **Best Practices** (`knowledge/best-practices/`):
   - [Practice 1]: [brief description]
   - Total: [N] practices documented

5. **Comprehensive Methodology** (`knowledge/methodology/`):
   - `refactoring-methodology.md`: Full lifecycle guide
   - Coverage: Detection → Planning → Execution → Verification

---

### 5. Transferability Tests

**Reusability Assessment**:

1. **Language Independence**:
   - Patterns applicable to: [Go, Python, JavaScript, Rust, Java, etc.]
   - Language-agnostic principles: [list]
   - Language-specific adaptations: [list]
   - Score: [% transferable across languages]

2. **Codebase Generality**:
   - Applicable to: [CLI tools, libraries, web services, embedded systems, etc.]
   - Universal patterns: [list]
   - Context-specific patterns: [list]
   - Score: [% transferable across codebase types]

3. **Domain Generality**:
   - Applicable beyond meta-cc: [YES/NO]
   - Domains tested: [list if any]
   - Required adaptations: [list]

**Transferability Evidence**:
- Concrete examples of how patterns apply to other contexts
- Adaptation guidelines for different scenarios
- Limitations and constraints

---

### 6. Methodology Validation

**Effectiveness Evidence**:

1. **Quality Improvement**:
   - Demonstrated in [N] refactoring sessions
   - Average quality gain: [%]
   - Consistent improvements: [YES/NO]

2. **Safety Record**:
   - Zero breaking changes: [YES/NO]
   - Test discipline: [% commits with passing tests]
   - Rollback capability: [demonstrated/not-needed]

3. **Efficiency Gains**:
   - Measured speedup: [X]x over ad-hoc
   - Automation achieved: [%]
   - Rework minimized: [% clean refactorings]

**Validation Confidence**: [High/Medium/Low] based on evidence quality

---

### 7. System Evolution Summary

**Architecture Evolution**:

1. **Capabilities Created**:
   - Initial: [list baseline capabilities]
   - Added: [list capabilities added during iterations]
   - Final count: [N]
   - Justification: [evidence-based / premature]

2. **Agents Created**:
   - Initial: meta-agent (generic)
   - Specialized agents added: [list]
   - Justification for specialization: [evidence summary]
   - Performance gain: [X]x (if applicable)

3. **Evolution Triggers**:
   - Iteration [X]: [component added] - Reason: [evidence summary]
   - Iteration [Y]: [component modified] - Reason: [evidence summary]

**Evolution Assessment**:
- Evidence-driven: [YES/NO]
- Necessary changes only: [YES/NO]
- Avoided premature optimization: [YES/NO]

---

### 8. Learnings and Insights

**Key Discoveries**:

1. **About Refactoring Domain**:
   - [Insight 1]: [description]
   - [Insight 2]: [description]
   - [Insight N]: [description]

2. **About BAIME Methodology**:
   - What worked well: [list]
   - What didn't work: [list]
   - Surprises: [list]

3. **About Agent Coordination**:
   - Generic vs. specialized agents: [findings]
   - Capability modularity: [findings]
   - Evolution timing: [findings]

**Unexpected Outcomes**:
- [List any unexpected results, positive or negative]

---

### 9. Comparison with Reference Experiments

**Bootstrap-002 (Test Strategy)**:
- Iterations: 6 vs. [N] (this experiment)
- Final V_instance: 0.80 vs. [score]
- Final V_meta: 0.80 vs. [score]
- Similarities: [list]
- Differences: [list]

**Bootstrap-003 (Error Recovery)**:
- Iterations: 3 vs. [N] (this experiment)
- Rapid convergence factors: [analysis]
- Similarities: [list]
- Differences: [list]

**Insights from Comparison**:
- Domain complexity impact on iteration count
- Baseline quality impact on convergence rate
- Transferable learnings

---

### 10. Recommendations

**For Future Refactoring Projects**:
1. [Recommendation 1]
2. [Recommendation 2]
3. [Recommendation N]

**For BAIME Methodology**:
1. [Improvement 1]
2. [Improvement 2]
3. [Improvement N]

**For Meta-CC Development**:
1. [Application 1]
2. [Application 2]
3. [Application N]

---

### 11. Knowledge Catalog

**Permanent Artifacts** (for reuse):

1. **Methodology Document**:
   - Location: `knowledge/methodology/refactoring-methodology.md`
   - Completeness: [%]
   - Validation status: [validated/needs-testing]

2. **Pattern Library**:
   - Location: `knowledge/patterns/`
   - Count: [N] patterns
   - Validation: [% applied successfully in practice]

3. **Principle Set**:
   - Location: `knowledge/principles/`
   - Count: [N] principles
   - Universality: [% transferable]

4. **Project-Wide Methodology**:
   - Location: `/docs/methodology/refactoring.md` (if extracted)
   - Reusability: [cross-project/meta-cc-specific]

**Ephemeral Data** (not for reuse):
- `data/iteration-*/*`: Iteration-specific logs and calculations
- `experiments/bootstrap-004-refactoring-guide/`: Experiment workspace

---

## Execution Guidance

### Perspective and Embodiment

**You are the Meta-Agent** for refactoring methodology development:
- Your task: Develop systematic refactoring methodology through observation and codification
- Your domain: Code refactoring, quality improvement, TDD, incremental change
- Your output: Transferable methodology applicable across codebases and languages
- Your constraint: Evidence-based evolution, honest assessment, rigorous convergence

**Embody Refactoring Expertise**:
- Understand code smells deeply (long methods, high complexity, duplication, poor naming, etc.)
- Apply refactoring patterns systematically (extract method, simplify conditionals, remove duplication, etc.)
- Enforce safety protocols (TDD, incremental commits, behavior preservation, test discipline)
- Measure quality rigorously (cyclomatic complexity, duplication, coverage, static analysis)

---

### Dual-Layer Evaluation Protocol

**Critical**: Instance and Meta layers are INDEPENDENT. Evaluate separately.

**Instance Layer** (Refactoring Quality):
- Measures: How well did we refactor `internal/query/`?
- Evidence: Code metrics, test coverage, complexity reduction, safety record
- Scoring: Apply V_instance rubrics rigorously
- Threshold: V_instance ≥ 0.75

**Meta Layer** (Methodology Quality):
- Measures: How good is the refactoring methodology we developed?
- Evidence: Completeness of lifecycle coverage, demonstrated effectiveness, transferability
- Scoring: Apply V_meta rubrics INDEPENDENTLY of instance scores
- Threshold: V_meta ≥ 0.70

**Independence Validation**:
- Good instance outcome ≠ good methodology (could be luck or manual effort)
- Good methodology ≠ good instance outcome (could be hard problem or limited time)
- Both must be evaluated on their own merits with separate evidence

---

### Honest Assessment Protocol

**Systematic Bias Avoidance**:

1. **Seek Disconfirming Evidence**:
   - Before assigning high score: What evidence contradicts this?
   - Example: "V_completeness seems high, but detection phase lacks automation"
   - Action: Lower score or improve detection phase

2. **Enumerate Gaps Explicitly**:
   - List missing elements in each rubric category
   - Don't gloss over weaknesses
   - Example: "Execution phase lacks transformation recipes for 4 refactoring types"

3. **Ground Scores in Concrete Evidence**:
   - Every score must cite specific artifacts
   - Example: "V_effectiveness = 0.80 because: (1) complexity reduced 28% in 3 functions [see data/iteration-3/metrics], (2) zero test failures across 45 commits [see git log], (3) 7x speedup measured [see time logs]"
   - Avoid: "V_effectiveness seems good" (no evidence)

4. **Challenge High Scores**:
   - Scores ≥0.8 require exceptional evidence
   - Ask: "What would make this a 1.0? Why isn't it there?"
   - Downgrade if evidence is weak

5. **Avoid Anti-Patterns**:
   - ❌ "We created many patterns, so V_completeness is high" → Need coverage depth, not just existence
   - ❌ "Methodology seems comprehensive" → Need rubric-by-rubric assessment with evidence
   - ❌ "V_instance improved, so V_meta must be good" → Layers are independent
   - ❌ "Patterns look transferable theoretically" → Need demonstrated transfer or explicit analysis

---

### Convergence Rigor

**Both Thresholds Required**:
- V_instance ≥ 0.75 AND V_meta ≥ 0.70
- Not one or the other; both must meet threshold
- Sustained for 2 iterations (stability)

**Stability Validation**:
- Scores must remain above threshold for 2 consecutive iterations
- Prevents premature convergence on temporary gains
- Example: If iteration 4 meets threshold but iteration 5 drops below, NOT converged

**Diminishing Returns**:
- If ΔV < 0.05 for both layers AND below threshold: stuck
- Action: Investigate methodology gaps, consider evolution, or reassess approach

**Do NOT**:
- Declare convergence based on single iteration
- Inflate scores to reach threshold
- Compromise rigor for appearance of success

---

### Honesty and Authenticity

**Discover, Don't Assume**:
- Let patterns emerge from data, don't impose them
- Extract principles from observations, don't theorize them
- Evolve system based on retrospective evidence, not anticipatory design

**No Token Limits**:
- Complete all analysis thoroughly
- Document all evidence comprehensively
- Don't summarize or skip for brevity

**Humble Uncertainty**:
- If unclear, state uncertainty explicitly
- If evidence is weak, acknowledge it
- If gaps exist, enumerate them

**Rigorous Grounding**:
- Every claim backed by data
- Every score backed by rubric application
- Every evolution backed by retrospective evidence

---

### Key Principles Summary

1. **Modular Architecture**: Separate files for capabilities and agents, clear interfaces
2. **Evidence-Based Evolution**: Only evolve system when retrospective data demonstrates necessity
3. **Dual-Layer Independence**: Evaluate V_instance and V_meta separately with distinct evidence
4. **Honest Calculation**: Apply rubrics rigorously, challenge high scores, enumerate gaps
5. **Low Baseline Acceptance**: Iteration 0 scores of 0.10-0.25 are normal and expected
6. **Convergence Discipline**: Both thresholds + 2 iteration stability required
7. **Knowledge Extraction**: Separate permanent methodology from ephemeral data
8. **Refactoring Safety**: TDD, incremental commits, behavior preservation, 100% test pass rate
9. **Metrics-Driven**: Ground all assessments in quantitative metrics (complexity, coverage, duplication, static analysis)
10. **Transferability Focus**: Design methodology for reuse across languages and codebases

---

### Common Pitfalls to Avoid

**Premature Optimization**:
- ❌ Creating specialized agents before evidence of need
- ❌ Building automation before workflow is proven
- ❌ Designing comprehensive taxonomy before observing patterns

**Score Inflation**:
- ❌ Assigning high scores without rigorous evidence
- ❌ Inflating V_meta because V_instance improved
- ❌ Ignoring gaps to reach convergence threshold

**Pattern Matching**:
- ❌ "Testing methodology had these agents, so refactoring should too"
- ❌ "Other experiments converged in N iterations, so we should too"
- ❌ Copying structure from reference experiments without evidence

**Convergence Shortcuts**:
- ❌ Declaring convergence after single iteration above threshold
- ❌ Weakening rubrics to achieve convergence
- ❌ Ignoring stability requirement

**Safety Compromise**:
- ❌ Accepting <100% test pass rate
- ❌ Skipping incremental verification
- ❌ Making large changes without safety checkpoints

---

## Appendix: Quick Reference

### Value Function Quick Reference

**V_instance Components**:
- V_code_quality (0.3): Complexity reduction, duplication elimination, static analysis
- V_maintainability (0.3): Test coverage, module cohesion, documentation
- V_safety (0.2): Test pass rate, behavior preservation, incremental discipline
- V_effort (0.2): Time efficiency, automation, rework minimization

**V_meta Components**:
- V_completeness (0.4): Detection, Planning, Execution, Verification phases
- V_effectiveness (0.3): Quality improvement, safety record, efficiency gains
- V_reusability (0.3): Language independence, codebase generality, abstraction quality

### Iteration Checklist

**Every Iteration**:
- [ ] Read previous iteration state (scores, problems, system state)
- [ ] Observe: Collect metrics, analyze patterns, query meta-cc data
- [ ] Codify: Gap analysis, strategy formation, pattern extraction
- [ ] Automate: Execute refactoring with TDD cycle, document thoroughly
- [ ] Evaluate: Calculate V_instance and V_meta with evidence
- [ ] Convergence: Check thresholds and stability
- [ ] Evolve: Only if retrospective evidence supports it
- [ ] Save: All deliverables in `data/iteration-N/`

### Convergence Checklist

- [ ] V_instance ≥ 0.75
- [ ] V_meta ≥ 0.70
- [ ] Sustained for 2 consecutive iterations
- [ ] Diminishing returns: ΔV < 0.05 for both layers
- [ ] All rubrics applied rigorously with evidence
- [ ] Gaps enumerated explicitly
- [ ] High scores challenged with disconfirming evidence

---

**End of Iteration Prompts**

**Next Steps**:
1. Begin with [Iteration 0: Baseline Establishment](#iteration-0-baseline-establishment)
2. Follow baseline objectives sequentially
3. Calculate honest baseline scores (expect V_instance ≈ 0.15-0.25, V_meta ≈ 0.10-0.20)
4. Use baseline as foundation for iteration 1
5. Iterate until convergence criteria met
6. Complete [Results Analysis](#results-analysis-template) upon convergence

**Remember**: Low baseline scores are expected and indicate honest assessment. Success is measured by rigorous methodology development, not by inflated scores.
