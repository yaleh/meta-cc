# Bootstrap-002: Test Strategy Development - Iteration Execution Prompts

**Framework**: BAIME (Bootstrapped AI Methodology Engineering)
**Experiment**: Bootstrap-002 Test Strategy Development
**Meta-Agent**: M₀ (5 capabilities: observe, plan, execute, reflect, evolve)

---

## Iteration Execution Protocol

λ(iteration_n, s_{n-1}) → (s_n, V(s_n), convergence_status):

```
iteration_cycle :: (M_{n-1}, A_{n-1}, s_{n-1}) → (M_n, A_n, s_n, V(s_n))
iteration_cycle(M, A, s) =
  pre_execution(s_{n-1}) →
  meta_agent_coordination(M) →
  observe_phase(A) →
  codify_phase(A) →
  automate_phase(A) →
  evaluate_phase(V) →
  convergence_check(V, M, A) →
  if converged then finalize else plan_next_iteration
```

---

## Pre-Execution Protocol

**Before each iteration**:

```
pre_execution :: State_{n-1} → Context
pre_execution(s) =
  read(iteration-{n-1}.md) ∧
  extract(V_instance, V_meta, gaps, learnings) ∧
  load(meta-agents/meta-agent-m0.md) ∧
  load(agents/*.md | if exists) ∧
  identify(focus_areas, priorities)
```

**Checklist**:
- [ ] Read previous iteration file (`iteration-{n-1}.md`)
- [ ] Review previous V_instance and V_meta scores
- [ ] Identify gaps from previous iteration
- [ ] Load Meta-Agent M₀ definition
- [ ] Load any existing agent definitions
- [ ] Determine focus for this iteration

---

## Meta-Agent M₀ Coordination

**Capabilities** (5 modular, always re-read from files):

1. **observe**: Pattern observation and data collection
2. **plan**: Iteration planning and objective setting
3. **execute**: Agent orchestration and task coordination
4. **reflect**: Value assessment and gap identification
5. **evolve**: System evolution decisions

**Coordination Pattern**:

```
meta_agent_protocol :: Capability → Execution
meta_agent_protocol(cap) =
  read(meta-agents/observe.md | plan.md | execute.md | reflect.md | evolve.md) ∧
  apply(guidance) ∧
  coordinate(agents) ∧
  ¬assume ∧ ¬cache
```

**Critical**: Always read capability files fresh, never cache instructions.

---

## Phase 1: OBSERVE (Data Collection)

**Objective**: Gather empirical data about current test state and coverage gaps

### Observation Tasks

```
observe_phase :: Agents → Observations
observe_phase(A) = sequential_execution(
  coverage_measurement,
  gap_identification,
  pattern_analysis,
  quality_assessment
)
```

### Specific Actions

**1. Measure Current Test Coverage**

```bash
# Run coverage analysis
cd /home/yale/work/meta-cc
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out > data/coverage-summary-iteration-{n}.txt

# Analyze per-package
go test -coverprofile=coverage.out ./internal/parser
go test -coverprofile=coverage.out ./internal/analyzer
go test -coverprofile=coverage.out ./internal/query
go test -coverprofile=coverage.out ./cmd/mcp-server
```

**Save outputs** to `data/coverage-*.txt`

**2. Identify High-Value Test Targets**

```bash
# Files with high change frequency (need robust tests)
meta-cc query-files --threshold 10 > data/high-change-files-iteration-{n}.jsonl

# Files with error patterns (need defensive tests)
meta-cc query-tools --status error --tool Edit > data/error-prone-files-iteration-{n}.jsonl

# Testing-related conversations (understand pain points)
meta-cc query-user-messages --pattern "test|coverage|mock|fixture" \
  > data/testing-conversations-iteration-{n}.jsonl
```

**3. Analyze Existing Test Patterns**

```bash
# Find existing test files
find . -name "*_test.go" -type f > data/existing-tests-iteration-{n}.txt

# Analyze test structure
grep -r "func Test" --include="*_test.go" | wc -l  # test count
grep -r "t.Run" --include="*_test.go" | wc -l      # subtests
grep -r "httptest" --include="*_test.go" | wc -l   # HTTP mocks
```

**4. Identify Coverage Gaps**

Use data-analyst agent or manual analysis:
- Which packages have <75% coverage?
- Which critical functions are untested?
- Which error paths lack tests?

**Save** gap analysis to `data/coverage-gaps-iteration-{n}.md`

### Observation Deliverables

- [ ] Coverage summary (`data/coverage-summary-iteration-{n}.txt`)
- [ ] High-value test targets identified
- [ ] Existing test patterns documented
- [ ] Coverage gaps prioritized
- [ ] Quality issues noted (flaky tests, slow tests)

---

## Phase 2: CODIFY (Pattern Extraction & Methodology)

**Objective**: Extract patterns and document testing methodology

### Codification Tasks

```
codify_phase :: Observations → Methodology
codify_phase(obs) = pattern_extraction(
  identify_successful_patterns,
  document_test_strategies,
  create_reusable_templates,
  define_quality_criteria
)
```

### Specific Actions

**1. Extract Successful Test Patterns**

Analyze existing high-quality tests:
- What makes a good test in this codebase?
- Which testing patterns are most effective?
- How are fixtures/mocks structured?

**Document** patterns in `knowledge/test-patterns-iteration-{n}.md`

**2. Design Coverage-Driven Workflow**

Create systematic approach:
1. Identify gap (uncovered code)
2. Prioritize (value × risk)
3. Write test (unit or integration)
4. Verify coverage improvement
5. Refactor test for quality

**Document** workflow in `knowledge/coverage-workflow-iteration-{n}.md`

**3. Create Test Templates**

For common scenarios:
- Unit test template
- Integration test template (with httptest)
- Table-driven test template
- Error path test template

**Save** templates to `knowledge/test-templates-iteration-{n}.md`

**4. Define Quality Gates**

Criteria for good tests:
1. Coverage target (75-85%)
2. Execution speed (<2 min total)
3. Flaky rate (<5%)
4. Assertion clarity (readable failures)
5. Fixture reuse (DRY principle)
6. Mock appropriateness (not over-mocked)
7. Edge case coverage
8. Error path coverage
9. Integration with CI
10. Documentation quality

**Document** quality gates in `knowledge/quality-gates-iteration-{n}.md`

### Codification Deliverables

- [ ] Test patterns documented
- [ ] Coverage-driven workflow defined
- [ ] Test templates created
- [ ] Quality gates established
- [ ] Methodology draft written

---

## Phase 3: AUTOMATE (Tool Creation & CI Integration)

**Objective**: Create tools and automation for systematic testing

### Automation Tasks

```
automate_phase :: Methodology → Tools
automate_phase(method) = tool_creation(
  ci_integration,
  coverage_gates,
  test_generators,
  reporting_automation
)
```

### Specific Actions

**1. Implement Tests (Gap Closure)**

Use coder agent to write tests for prioritized gaps:
- Start with highest-value files
- Use test templates from Phase 2
- Follow coverage-driven workflow
- Aim for incremental improvement (not perfection)

**Track** progress in `data/test-implementation-iteration-{n}.yaml`:
```yaml
file: internal/parser/parser.go
coverage_before: 45%
coverage_after: 78%
tests_added: 12
patterns_used: [unit-test, table-driven, error-path]
```

**2. CI/CD Integration**

Create/update `.github/workflows/test.yml`:
```yaml
name: Test Suite

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - run: go test -coverprofile=coverage.out ./...
      - run: go tool cover -func=coverage.out
      # Add coverage gate
      - run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$COVERAGE < 75" | bc -l) )); then
            echo "Coverage $COVERAGE% below threshold 75%"
            exit 1
          fi
```

**3. Makefile Targets**

Update `Makefile`:
```makefile
.PHONY: test
test:
\tgo test -v -race -coverprofile=coverage.out ./...

.PHONY: test-coverage
test-coverage: test
\tgo tool cover -html=coverage.out -o coverage.html
\t@echo "Coverage report: coverage.html"

.PHONY: test-fast
test-fast:
\tgo test -short ./...
```

**4. Coverage Reporting**

Create script `scripts/coverage-report.sh`:
```bash
#!/bin/bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | tee coverage-summary.txt
go tool cover -html=coverage.out -o coverage.html
echo "Coverage report generated: coverage.html"
```

### Automation Deliverables

- [ ] Tests implemented for priority gaps
- [ ] CI/CD workflow created/updated
- [ ] Makefile test targets added
- [ ] Coverage reporting automated
- [ ] Quality gates enforced in CI

---

## Phase 4: EVALUATE (Value Calculation)

**Objective**: Calculate dual-layer value functions and assess state

### Evaluation Tasks

```
evaluate_phase :: State → (V_instance, V_meta, Gaps)
evaluate_phase(s) = independent_assessment(
  calculate_instance_value,
  calculate_meta_value,
  identify_gaps,
  compare_to_previous
) where honest_scoring ∧ concrete_evidence
```

### V_instance Calculation

**Formula**:
```
V_instance(s_n) = 0.35·V_coverage + 0.25·V_quality + 0.20·V_maintainability + 0.20·V_automation
```

**Measurement**:

1. **V_coverage** (from coverage report):
   ```
   total_coverage = (go tool cover -func=coverage.out | grep total | awk '{print $3}')

   V_coverage = {
     1.0  if coverage ≥ 85%
     0.8  if coverage ≥ 75%
     0.6  if coverage ≥ 65%
     0.4  if coverage ≥ 50%
     0.2  otherwise
   }
   ```

2. **V_quality** (test effectiveness):
   ```
   flaky_rate = (flaky_tests / total_tests)
   execution_time = (time go test ./...)

   V_quality = {
     1.0  if flaky < 5% ∧ time < 60s
     0.8  if flaky < 10% ∧ time < 120s
     0.6  if flaky < 15% ∧ time < 180s
     0.4  otherwise
   }
   ```

3. **V_maintainability** (test code quality):
   ```
   fixture_reuse = count(shared_fixtures) / count(tests)
   duplication = (duplicated_test_code / total_test_code)

   V_maintainability = {
     1.0  if fixture_reuse > 0.7 ∧ duplication < 10%
     0.8  if fixture_reuse > 0.5 ∧ duplication < 20%
     0.6  if fixture_reuse > 0.3 ∧ duplication < 30%
     0.4  otherwise
   }
   ```

4. **V_automation** (CI integration):
   ```
   V_automation = {
     1.0  if CI + coverage_gate + auto_report
     0.8  if CI + coverage_gate
     0.6  if CI only
     0.4  if manual testing
   }
   ```

**Calculate**: `V_instance(s_n) = 0.35·V_cov + 0.25·V_qual + 0.20·V_maint + 0.20·V_auto`

### V_meta Calculation

**Formula**:
```
V_meta(s_n) = 0.40·V_completeness + 0.30·V_effectiveness + 0.30·V_reusability
```

**Measurement**:

1. **V_completeness** (methodology documentation):
   ```
   checklist = [
     process_steps_documented,
     decision_criteria_defined,
     examples_provided,
     edge_cases_covered,
     failure_modes_documented,
     rationale_explained
   ]

   V_completeness = (items_completed / total_items)
   ```

2. **V_effectiveness** (practical impact):
   ```
   time_before = estimated_manual_effort
   time_after = actual_effort_with_methodology
   speedup = time_before / time_after

   V_effectiveness = {
     1.0  if speedup > 10x
     0.8  if speedup > 5x
     0.6  if speedup > 2x
     0.4  otherwise
   }
   ```

3. **V_reusability** (transferability):
   ```
   transfer_scenarios = [
     same_language (Go → Go),
     similar_language (Go → Rust),
     different_language (Go → Python)
   ]

   modification_needed = estimate(percentage)

   V_reusability = {
     1.0  if modification < 15%
     0.8  if modification < 40%
     0.6  if modification < 70%
     0.4  otherwise
   }
   ```

**Calculate**: `V_meta(s_n) = 0.40·V_comp + 0.30·V_eff + 0.30·V_reus`

### Gap Identification

**Instance Gaps**:
- Which packages still have <75% coverage?
- Which quality criteria not met?
- Which automation missing?

**Meta Gaps**:
- What methodology elements incomplete?
- What patterns not yet extracted?
- What edge cases not documented?

**Document** all gaps and V-scores in `iteration-{n}.md`

### Evaluation Deliverables

- [ ] V_instance calculated with evidence
- [ ] V_meta calculated with evidence
- [ ] Component breakdowns documented
- [ ] Gaps identified and prioritized
- [ ] Comparison to previous iteration (ΔV)

---

## Phase 5: CONVERGE or EVOLVE

**Objective**: Decide whether to converge or continue iterating

### Convergence Check

```
convergence_check :: (V_i, V_m, M_n, A_n) → Decision
convergence_check(V_i, V_m, M, A) =
  system_stable ∧ dual_threshold ∧ diminishing_returns →
    if all_met then CONVERGE
    else CONTINUE(next_focus)
```

**Criteria**:

1. **Dual Threshold**:
   - [ ] V_instance(s_n) ≥ 0.80
   - [ ] V_meta(s_n) ≥ 0.80

2. **System Stability**:
   - [ ] M_n == M_{n-1} (Meta-Agent unchanged)
   - [ ] A_n == A_{n-1} (Agent set unchanged)

3. **Diminishing Returns** (for 2+ iterations):
   - [ ] ΔV_instance < 0.02
   - [ ] ΔV_meta < 0.02

4. **Objectives Complete**:
   - [ ] Coverage ≥75%
   - [ ] Quality gates met (≥8/10)
   - [ ] Methodology documented
   - [ ] Automation implemented

**Decision**:
```
if all_criteria_met:
  → CONVERGED: Proceed to results analysis
else:
  → CONTINUE: Plan next iteration focus
```

### Evolution Decisions

**Agent Evolution** (only if necessary):

```
agent_evolution :: (A_n, Evidence) → A_{n+1}
agent_evolution(A, E) =
  if agent_insufficiency_demonstrated(E) ∧
     alternatives_attempted ∧
     efficiency_gain > 2x then
    create_specialized_agent(rationale, expected_improvement)
  else
    maintain(A)
```

**When to create specialized agent**:
- Generic agents fail after 2+ attempts
- Specific expertise needed (e.g., coverage analysis algorithms)
- Expected efficiency gain >2x
- Pattern will be reused extensively

**Meta-Agent Evolution** (rare):

```
meta_evolution :: (M_n, Evidence) → M_{n+1}
meta_evolution(M, E) =
  if capability_gap_demonstrated(E) ∧
     lifecycle_insufficient then
    add_capability(rationale, integration_plan)
  else
    maintain(M_0)
```

**Expected**: M₀ should remain stable (has worked in all 8 previous experiments)

### Next Iteration Planning

**If not converged**, determine focus:

```
next_focus :: Gaps → Priorities
next_focus(gaps) = prioritize(
  highest_impact_gaps,
  easiest_wins,
  methodology_completeness
)
```

**Example**:
- Iteration 1: Focus on unit test coverage (foundational)
- Iteration 2: Add integration tests (higher complexity)
- Iteration 3: Refine fixtures and patterns (quality)
- Iteration 4: Complete documentation and automation (meta)

---

## Iteration Output Template

**Each iteration must produce**: `iteration-{n}.md`

```markdown
# Iteration {n}: {Title}

**Date**: YYYY-MM-DD
**Duration**: X hours
**Status**: {In Progress | Completed}

## Pre-Execution Context

- Previous V_instance: {value}
- Previous V_meta: {value}
- Focus areas: {list}
- Gaps addressed: {list}

## Observe Phase

### Data Collected
- Coverage: {percentage}
- Test count: {number}
- Gaps identified: {list}

### Observations
{findings}

## Codify Phase

### Patterns Extracted
{patterns}

### Methodology Updates
{methodology changes}

## Automate Phase

### Tests Implemented
- Files: {list}
- Coverage improvement: {before → after}

### Tools Created
{tools}

## Evaluate Phase

### V_instance Calculation
- V_coverage: {value} ({evidence})
- V_quality: {value} ({evidence})
- V_maintainability: {value} ({evidence})
- V_automation: {value} ({evidence})
- **V_instance(s_{n})**: {total}

### V_meta Calculation
- V_completeness: {value} ({evidence})
- V_effectiveness: {value} ({evidence})
- V_reusability: {value} ({evidence})
- **V_meta(s_{n})**: {total}

### Gaps Identified
**Instance gaps**: {list}
**Meta gaps**: {list}

## Convergence Check

- [ ] V_instance ≥ 0.80: {yes/no} ({value})
- [ ] V_meta ≥ 0.80: {yes/no} ({value})
- [ ] M_n == M_{n-1}: {yes/no}
- [ ] A_n == A_{n-1}: {yes/no}
- [ ] ΔV_instance < 0.02: {yes/no} ({delta})
- [ ] ΔV_meta < 0.02: {yes/no} ({delta})

**Status**: {CONVERGED | CONTINUE}

## Evolution Decisions

### Agent Evolution
{decisions and rationale}

### Meta-Agent Evolution
{decisions and rationale}

## Next Iteration Plan

**Focus**: {areas}
**Priorities**: {list}
**Expected ΔV**: {estimate}

## Artifacts Created

- Data: {list files in data/}
- Knowledge: {list files in knowledge/}
- Agents: {list files in agents/}
- Code: {tests written}

## Reflections

### What Worked
{successes}

### What Didn't Work
{failures}

### Learnings
{insights}
```

---

## Constraints and Principles

### From BAIME Framework

**Do**:
- ✅ Read capability files fresh every time (¬cache)
- ✅ Calculate V(s) honestly based on actual state
- ✅ Let agent specialization emerge from data
- ✅ Complete all phases thoroughly
- ✅ Document all decisions and evidence

**Don't**:
- ❌ Predetermine agent evolution path
- ❌ Force convergence at target iteration
- ❌ Inflate value metrics to meet targets
- ❌ Skip phases due to perceived constraints
- ❌ Assume capabilities without reading files

### Context Management

```
context_pressure :: State → Strategy
context_pressure(s) =
  if usage(s) > 0.80 then
    serialize_to(knowledge/) ∧ split_session
  else if usage(s) > 0.50 then
    reference_compression ∧ link_files
  else
    standard_protocol
```

**If context >80%**:
- Save iteration state to `knowledge/iteration-{n}-state.md`
- Continue in new session
- Load via file references

### Honest Assessment

**Critical principles**:
1. Seek disconfirming evidence (not just confirming)
2. Weight evidence objectively
3. Revise estimates if evidence contradicts
4. Document uncertainty
5. Avoid:
   - Inflating V-scores
   - Selective evidence
   - Moving goalposts
   - Cherry-picking metrics

---

## Quick Reference

### Per-Iteration Checklist

- [ ] Pre-execution: Read previous iteration, load M₀ and agents
- [ ] Observe: Collect coverage data, identify gaps, analyze patterns
- [ ] Codify: Extract patterns, document methodology, create templates
- [ ] Automate: Write tests, integrate CI, create tools
- [ ] Evaluate: Calculate V_instance and V_meta with evidence
- [ ] Check convergence: Assess all criteria
- [ ] Decide: Converge or plan next iteration
- [ ] Document: Create iteration-{n}.md with complete analysis
- [ ] Save artifacts: Data, knowledge, code

### Expected Timeline

| Iteration | Focus | Duration | Expected ΔV |
|-----------|-------|----------|-------------|
| 0 | Baseline | 2h | N/A |
| 1 | Unit tests | 2h | +0.15 |
| 2 | Integration tests | 2.5h | +0.12 |
| 3 | Quality & patterns | 2h | +0.08 |
| 4 | Documentation & automation | 2h | +0.05 |
| 5 | Convergence verification | 1.5h | <0.02 |

**Total**: 12 hours, 5-6 iterations (estimated)

---

**Ready to execute**: Start with Iteration 0 (baseline establishment)
**Next step**: Run observation commands and establish baseline metrics
