# Iteration Execution Prompts

For **bootstrap-002-test-strategy**: Develop systematic testing strategy and automated test generation methodology

## Overview

This document provides structured prompts for executing each iteration of the bootstrap-002-test-strategy experiment. The experiment develops a comprehensive testing methodology through systematic evolution of Meta-Agent capabilities and specialized testing agents.

**Domain**: Software Testing and Quality Assurance (Go)

**Value Function**:
```
V(s) = 0.3·V_coverage +         # Test coverage (line ≥80%, branch ≥70%)
       0.3·V_reliability +      # Error detection rate, test stability
       0.2·V_maintainability +  # Test maintenance cost, clarity
       0.2·V_speed              # Execution speed, parallelization

Target: V(sₙ) ≥ 0.80
```

**Component Interpretation**:
- **V_coverage** ∈ [0, 1]: (actual_coverage / target_coverage), line coverage ≥80%, branch coverage ≥70%
- **V_reliability** ∈ [0, 1]: (tests_passing / total_tests) × (critical_paths_tested / critical_paths)
- **V_maintainability** ∈ [0, 1]: 1 - (avg_test_complexity / max_acceptable_complexity), test clarity, DRY principle
- **V_speed** ∈ [0, 1]: 1 - (execution_time / baseline_time), parallel execution efficiency

**Initial State**:
- M₀: 5 core capabilities (observe, plan, execute, reflect, evolve)
- A₀: 3 generic agents (data-analyst, doc-writer, coder)
- Domain: Go testing ecosystem (testing package, testify, go test tooling)

**Expected Evolution**:
- Specialized testing agents: test-generator, coverage-analyzer, test-optimizer, mock-designer
- Enhanced Meta-Agent capabilities for test quality assessment
- Automated test generation and coverage analysis workflows

---

## Iteration 0: Baseline Establishment

**Purpose**: Establish baseline testing metrics, create modular Meta-Agent architecture, and identify testing gaps.

### Context

You are the Meta-Agent M₀ beginning the bootstrap-002-test-strategy experiment. Your goal is to:

1. Create the modular Meta-Agent capability architecture
2. Create initial generic agent files
3. Collect comprehensive testing data from the meta-cc project
4. Calculate baseline V(s₀) for the testing state
5. Identify testing problems and coverage gaps
6. Document initial findings

This iteration establishes the foundation for systematic testing improvement.

### Meta-Agent Capability Specifications

Before beginning work, create these capability files in `meta-agents/`:

#### meta-agents/observe.md

```markdown
# Observe Capability

## Purpose
Collect testing data, analyze test coverage, identify testing gaps, and discover testing patterns in the meta-cc codebase.

## Testing Data Collection

### Test Coverage Analysis
```bash
# Generate comprehensive coverage report
go test -cover ./... -coverprofile=coverage.out
go tool cover -func=coverage.out > coverage-summary.txt
go tool cover -html=coverage.out -o coverage.html

# Per-package coverage breakdown
go test -cover ./internal/...
go test -cover ./cmd/...
```

### Test Inventory
```bash
# Find all test files
find . -name "*_test.go" -type f

# Count test functions
grep -r "^func Test" . --include="*_test.go" | wc -l
grep -r "^func Benchmark" . --include="*_test.go" | wc -l

# Analyze test patterns
grep -r "t\.Run(" . --include="*_test.go"  # Subtests
grep -r "assert\." . --include="*_test.go"  # Assertion usage
```

### Test Execution Metrics
```bash
# Measure test execution time
go test -v ./... -count=1 | tee test-execution.log

# Test with race detection
go test -race ./...

# Benchmark existing tests
go test -bench=. -benchmem ./...
```

### Code Complexity Analysis
```bash
# Analyze function complexity (requires gocyclo)
find . -name "*.go" -not -path "./vendor/*" -not -name "*_test.go" -exec gocyclo {} \;

# Count functions per package
grep -r "^func " . --include="*.go" --exclude="*_test.go" | cut -d: -f1 | sort | uniq -c
```

## Pattern Recognition

### Testing Gaps to Identify
1. **Coverage Gaps**: Functions/packages with <80% line coverage, <70% branch coverage
2. **Critical Path Gaps**: Error handling, edge cases, boundary conditions without tests
3. **Test Type Gaps**: Missing integration tests, lack of property-based tests
4. **Test Quality Issues**: Flaky tests, slow tests (>100ms for unit tests), unclear test names
5. **Mocking Gaps**: External dependencies not properly isolated

### Testing Pattern Discovery
1. **Table-Driven Tests**: Usage patterns, effectiveness
2. **Test Fixtures**: Setup/teardown patterns, data management
3. **Subtest Organization**: Hierarchical test structure usage
4. **Error Testing**: How errors are validated, coverage of error paths
5. **Mock Usage**: Interface mocking patterns, dependency injection

## Data Sources

- Test files: `*_test.go` throughout codebase
- Coverage reports: `go test -cover` output
- Test execution logs: timing, failures, race conditions
- Code structure: function complexity, package dependencies
- Error history: 1,137 errors from bootstrap-001 baseline

## Output Format

Produce structured testing data:
- `data/test-coverage.json`: Per-package coverage metrics
- `data/test-inventory.json`: Test count, types, patterns
- `data/test-execution.json`: Timing, success rates, flakiness
- `data/coverage-gaps.json`: Functions/packages below thresholds
- `data/test-quality.json`: Test maintainability metrics
```

#### meta-agents/plan.md

```markdown
# Plan Capability

## Purpose
Prioritize testing objectives, select appropriate testing agents, and make strategic decisions about test generation and coverage improvement.

## Testing Objective Prioritization

### Priority Framework
1. **Critical Path Coverage** (Highest): Error handling, core business logic, data processing
2. **Quality Gate Compliance**: Achieve 80% line coverage, 70% branch coverage
3. **Test Reliability**: Eliminate flaky tests, ensure deterministic outcomes
4. **Test Maintainability**: Clear test names, DRY principle, fixture organization
5. **Test Performance**: Fast unit tests (<100ms), efficient integration tests

### Decision Criteria

#### When to Generate New Tests
- Function/package below coverage threshold (80% line, 70% branch)
- Critical path without error case testing
- New code without corresponding tests
- Integration points lacking integration tests

#### When to Refactor Existing Tests
- Test execution time >100ms for unit tests
- Duplicate test setup code (no fixtures)
- Unclear test names or assertions
- Flaky test behavior detected

#### When to Add Test Infrastructure
- Missing mock implementations for external dependencies
- Lack of test fixtures for common scenarios
- No benchmark tests for performance-critical code
- Missing integration test framework

## Agent Selection Strategy

### Generic vs Specialized Agents

**Use Generic Agents When**:
- Task is well-defined and straightforward
- No domain-specific testing knowledge required
- Simple data analysis or documentation

**Create Specialized Agent When**:
- Need repeated test generation with specific patterns
- Complex coverage analysis requiring Go AST parsing
- Test optimization requiring profiling expertise
- Mock generation requiring interface analysis

### Testing Task Mapping
- **test-generator** (specialized): Generate table-driven tests, subtests, edge cases
- **coverage-analyzer** (specialized): Parse coverage data, identify gaps, prioritize
- **test-optimizer** (specialized): Profile tests, parallelize, reduce execution time
- **mock-designer** (specialized): Generate mocks for interfaces, design test doubles
- **data-analyst** (generic): Aggregate metrics, calculate V(s)
- **doc-writer** (generic): Document testing strategies, write test plans
- **coder** (generic): Simple test implementations, refactoring

## Testing Strategy Decision Points

### Test Type Selection
- **Unit Tests**: For pure functions, business logic, isolated components
- **Integration Tests**: For database access, external APIs, inter-package interactions
- **End-to-End Tests**: For CLI commands, MCP server workflows
- **Property-Based Tests**: For parsers, data transformations, invariants
- **Fuzz Tests**: For input validation, parsing untrusted data

### Mocking Strategy
- **Interface Mocking**: For external dependencies (filesystem, network)
- **Test Doubles**: For database, HTTP clients
- **In-Memory Implementations**: For fast integration tests
- **Real Dependencies**: Only for true integration/e2e tests

### Test Organization
- **Table-Driven**: For multiple input/output scenarios
- **Subtests**: For related test cases, setup sharing
- **Test Suites**: For integration tests requiring shared setup
- **Parallel Tests**: For independent test cases (mark with `t.Parallel()`)

## Goal Definition Template

For each iteration:
1. **Primary Goal**: What testing improvement to achieve
2. **Success Criteria**: Specific V(s) increase, coverage target
3. **Required Agents**: Which agents needed (existing or new)
4. **Test Generation Plan**: What tests to generate, what patterns to use
5. **Validation Plan**: How to verify test quality and coverage
```

#### meta-agents/execute.md

```markdown
# Execute Capability

## Purpose
Coordinate test generation work, manage test validation, and ensure systematic testing improvement through agent collaboration.

## Execution Protocol

### Phase 1: Assess Agent Capabilities

**Read Agent Files**:
```
Read: agents/data-analyst.md
Read: agents/doc-writer.md
Read: agents/coder.md
Read: agents/test-generator.md (if exists)
Read: agents/coverage-analyzer.md (if exists)
Read: agents/test-optimizer.md (if exists)
Read: agents/mock-designer.md (if exists)
```

**Capability Assessment**:
- Can existing agents generate the required tests?
- Do we need specialized testing knowledge?
- Is test generation repetitive enough to warrant specialization?

### Phase 2: Evolution (if needed)

**If Existing Agents Insufficient**:
1. **Determine Specialization Need**: What specific testing capability is missing?
2. **Design Agent Specification**: Define agent's testing expertise
3. **Create Agent File**: `agents/{agent-name}.md` with detailed capabilities
4. **Document Justification**: Why specialization was necessary

**Common Specialized Testing Agents**:

**agents/test-generator.md**:
- Generates table-driven tests from function signatures
- Creates subtest structures for related test cases
- Generates edge case tests (nil, empty, boundary values)
- Produces error case tests for all error returns
- Uses testify assertions appropriately

**agents/coverage-analyzer.md**:
- Parses `go test -cover` output
- Identifies functions below coverage thresholds
- Prioritizes gaps by criticality (error handling, core logic)
- Analyzes branch coverage gaps
- Generates coverage improvement recommendations

**agents/test-optimizer.md**:
- Profiles test execution time
- Identifies slow tests (>100ms unit tests)
- Recommends parallelization opportunities
- Optimizes test fixtures and setup
- Reduces test dependencies

**agents/mock-designer.md**:
- Identifies interfaces requiring mocks
- Generates mock implementations
- Designs test doubles for external dependencies
- Creates fixture data for integration tests
- Ensures proper test isolation

### Phase 3: Agent Invocation

**Invocation Protocol**:
1. Read agent file immediately before use
2. Provide clear testing task specification
3. Include relevant testing context (coverage data, existing patterns)
4. Specify expected output format (test code, coverage report, documentation)

**Example Invocations**:

```
Invoke: agents/coverage-analyzer.md
Task: Analyze current coverage data and identify top 10 functions needing tests
Context: data/test-coverage.json, data/test-inventory.json
Output: data/coverage-priorities.json
```

```
Invoke: agents/test-generator.md
Task: Generate table-driven tests for internal/parser/parser.go:ParseLine
Context: Function signature, existing test patterns in parser_test.go
Output: Test code for parser_test.go
```

```
Invoke: agents/test-optimizer.md
Task: Profile and optimize slow tests in cmd/query package
Context: data/test-execution.json showing 5 tests >500ms
Output: Optimized test implementations, parallel execution plan
```

### Phase 4: Test Validation

**After Test Generation**:
1. Run tests: `go test ./...`
2. Verify coverage improvement: `go test -cover ./...`
3. Check test execution time: measure vs baseline
4. Validate test quality: clear names, proper assertions, no flakiness
5. Review test maintainability: fixture usage, DRY principle

**Quality Checks**:
- Tests pass consistently (run 5 times minimum)
- Coverage increases in target areas
- Test execution time acceptable (<100ms for unit tests)
- Tests follow project conventions (table-driven, subtests)
- Proper error assertions (not just `assert.Error`)

## Coordination Patterns

### Sequential Test Generation
1. Coverage-analyzer identifies gaps
2. Test-generator creates tests for high-priority gaps
3. Validate tests pass and coverage improves
4. Test-optimizer profiles and improves slow tests
5. Doc-writer documents testing strategy

### Parallel Test Development
- Multiple test-generator invocations for independent packages
- Simultaneous mock-designer work for interface mocking
- Parallel documentation by doc-writer

### Iterative Refinement
1. Generate initial tests
2. Run and measure coverage
3. Identify remaining gaps
4. Generate additional tests
5. Optimize and refactor
6. Validate final state

## Handoff Protocols

### Between Testing Agents
- **Coverage-analyzer → Test-generator**: Coverage gaps with priorities
- **Test-generator → Test-optimizer**: Generated tests needing optimization
- **Mock-designer → Test-generator**: Mock implementations for use in tests
- **Test-optimizer → Doc-writer**: Testing strategy and patterns to document

### Test Artifacts
- Generated test files: `*_test.go`
- Coverage reports: `coverage.out`, `coverage-summary.txt`
- Benchmark results: `benchmark-results.txt`
- Test execution logs: `test-execution.log`

## Testing Task Patterns

### Pattern: Generate Tests for Package
1. Analyze package structure and existing tests
2. Identify untested functions
3. Generate table-driven tests for pure functions
4. Generate integration tests for I/O functions
5. Create mocks for external dependencies
6. Validate coverage improvement

### Pattern: Optimize Test Suite
1. Profile test execution time
2. Identify slow tests (>100ms)
3. Analyze test dependencies
4. Parallelize independent tests
5. Optimize fixtures and setup
6. Measure improvement

### Pattern: Improve Test Quality
1. Review test names and clarity
2. Identify duplicate setup code
3. Create shared fixtures
4. Improve assertion specificity
5. Ensure error case coverage
6. Validate maintainability
```

#### meta-agents/reflect.md

```markdown
# Reflect Capability

## Purpose
Calculate V(s) for testing state, assess test quality honestly, identify coverage gaps, and determine iteration progress toward convergence.

## Value Calculation V(s)

### Component: V_coverage (Weight: 0.3)

**Line Coverage**:
```
line_coverage = (lines_covered / total_lines) for non-test code
V_line = min(line_coverage / 0.80, 1.0)  # Target: 80%
```

**Branch Coverage**:
```
branch_coverage = (branches_covered / total_branches)
V_branch = min(branch_coverage / 0.70, 1.0)  # Target: 70%
```

**Combined Coverage**:
```
V_coverage = 0.6·V_line + 0.4·V_branch
```

**Data Sources**:
- `go test -cover ./...` for line coverage
- `go tool cover -func=coverage.out` for detailed coverage
- Coverage reports from data/test-coverage.json

**Honest Assessment**:
- Do NOT assume coverage without running tests
- Exclude test files from coverage calculations
- Report actual percentages, not rounded
- Identify specific functions/packages below threshold

### Component: V_reliability (Weight: 0.3)

**Test Pass Rate**:
```
pass_rate = (tests_passing / total_tests)
```

**Critical Path Coverage**:
```
critical_coverage = (critical_paths_tested / critical_paths_total)
critical_paths = error_handling + boundary_conditions + core_logic
```

**Test Stability** (no flaky tests):
```
stability = 1 - (flaky_tests / total_tests)
flaky_test = test that fails intermittently without code changes
```

**Combined Reliability**:
```
V_reliability = 0.4·pass_rate + 0.4·critical_coverage + 0.2·stability
```

**Data Sources**:
- Test execution logs: `go test -v ./...`
- Critical path analysis: manual code review + error history
- Flakiness detection: run tests multiple times (5+ runs)

**Honest Assessment**:
- Test 100 passing ≠ 100% reliability if critical paths untested
- Identify specific critical paths without tests
- Measure actual flakiness (run tests multiple times)
- Report error case coverage separately

### Component: V_maintainability (Weight: 0.2)

**Test Complexity**:
```
avg_test_complexity = sum(cyclomatic_complexity) / num_tests
max_acceptable = 10  # Cyclomatic complexity threshold
V_complexity = 1 - min(avg_test_complexity / max_acceptable, 1.0)
```

**Test Clarity**:
```
clarity_score = (tests_with_clear_names / total_tests) ×
                (tests_with_good_assertions / total_tests)
clear_name = descriptive, follows convention (TestFunction_Condition_Expectation)
good_assertion = specific (assert.Equal vs assert.NoError)
```

**DRY Principle**:
```
duplication_score = 1 - (duplicate_setup_lines / total_test_lines)
```

**Combined Maintainability**:
```
V_maintainability = 0.4·V_complexity + 0.3·clarity_score + 0.3·duplication_score
```

**Data Sources**:
- Test files analysis: count functions, assertions, patterns
- Cyclomatic complexity: gocyclo tool
- Code review: identify duplicate setup, unclear names

**Honest Assessment**:
- Count actual duplicate setup code lines
- Review test names for clarity (manual inspection)
- Calculate cyclomatic complexity objectively
- Report specific maintainability issues

### Component: V_speed (Weight: 0.2)

**Execution Time**:
```
execution_time = sum(test_duration) for all tests
baseline_time = initial measurement from Iteration 0
V_time = 1 - min((execution_time - baseline_time) / baseline_time, 1.0)
      = 1.0 if execution_time <= baseline_time (no slowdown)
```

**Parallel Efficiency**:
```
parallel_ratio = (tests_marked_parallel / parallelizable_tests)
parallelizable = tests without shared state
V_parallel = parallel_ratio
```

**Combined Speed**:
```
V_speed = 0.7·V_time + 0.3·V_parallel
```

**Data Sources**:
- Test timing: `go test -v ./...` output parsing
- Parallel test count: `grep "t.Parallel()" *_test.go | wc -l`
- Baseline timing from Iteration 0 data/test-execution.json

**Honest Assessment**:
- Measure actual test execution time (run multiple times)
- Identify slow tests (>100ms for unit tests)
- Calculate parallel potential (manual analysis)
- Report specific slow tests

## Overall Value Calculation

```
V(s) = 0.3·V_coverage + 0.3·V_reliability + 0.2·V_maintainability + 0.2·V_speed

Target: V(sₙ) ≥ 0.80
```

## Gap Identification

### Coverage Gaps
- Functions with <80% line coverage
- Packages with <70% branch coverage
- Error paths without tests
- Boundary conditions without tests

### Reliability Gaps
- Critical paths without tests
- Flaky tests (intermittent failures)
- Missing error case tests
- Insufficient edge case coverage

### Maintainability Gaps
- Tests with unclear names
- Duplicate setup code (no fixtures)
- High complexity tests (cyclomatic > 10)
- Poor assertion specificity

### Speed Gaps
- Slow unit tests (>100ms)
- Tests not using t.Parallel() when possible
- Inefficient test setup/teardown
- Redundant test operations

## Quality Assessment

### Test Quality Checklist
- [ ] Tests pass consistently (5+ runs)
- [ ] Coverage targets met (80% line, 70% branch)
- [ ] Critical paths tested (error handling, edge cases)
- [ ] Clear test names (TestFunction_Condition_Expectation)
- [ ] Specific assertions (assert.Equal, not just assert.NoError)
- [ ] Table-driven tests for multiple scenarios
- [ ] Subtests for related cases
- [ ] Proper test isolation (mocks for external deps)
- [ ] Fast execution (<100ms per unit test)
- [ ] No flaky tests

## Convergence Progress

### Iteration-to-Iteration Comparison
```
ΔV = V(sₙ) - V(sₙ₋₁)

Progress Assessment:
- ΔV > 0.05: Good progress
- ΔV = 0.02-0.05: Moderate progress
- ΔV < 0.02: Slow progress, may be converging
```

### Component-Level Progress
Track each component separately:
- ΔV_coverage: Coverage improvement
- ΔV_reliability: Reliability improvement
- ΔV_maintainability: Maintainability improvement
- ΔV_speed: Speed improvement

Identify which component is blocking convergence.

## Reflection Questions

### After Each Iteration
1. **Coverage**: Did we increase coverage? Which gaps remain?
2. **Reliability**: Are tests more reliable? Any flaky tests?
3. **Maintainability**: Are tests easier to understand and modify?
4. **Speed**: Did test execution time increase? Can we parallelize?
5. **Agent Effectiveness**: Did specialized agents improve outcomes?
6. **Strategy**: Is our testing strategy effective? What to adjust?

### Honest Self-Assessment
- What worked well in this iteration?
- What didn't work as expected?
- Are we generating high-quality tests or just increasing numbers?
- Is test maintainability degrading as we add more tests?
- Are we testing the right things (critical paths vs trivial code)?

## Output Format

**Reflection Document**:
- Current V(s) calculation with all components
- Component-level breakdown and evidence
- Identified gaps (specific functions, packages)
- Quality assessment against checklist
- Progress since last iteration (ΔV)
- Recommendations for next iteration
```

#### meta-agents/evolve.md

```markdown
# Evolve Capability

## Purpose
Determine when to create specialized testing agents, identify new testing capability needs, and guide Meta-Agent evolution based on testing domain requirements.

## Specialization Triggers

### When to Create Specialized Testing Agent

**Trigger 1: Repeated Testing Pattern**
- Same type of test generation needed multiple times
- Example: Generating table-driven tests for 10+ functions
- Solution: Create `agents/test-generator.md` with table-driven test expertise

**Trigger 2: Domain-Specific Testing Knowledge Required**
- Generic agents lack testing methodology knowledge
- Example: Coverage analysis requiring Go coverage format parsing
- Solution: Create `agents/coverage-analyzer.md` with coverage expertise

**Trigger 3: Complex Testing Task Requiring Specialized Skills**
- Task requires specific testing tools or techniques
- Example: Test optimization requiring profiling and parallel execution
- Solution: Create `agents/test-optimizer.md` with optimization expertise

**Trigger 4: Interface/Dependency Mocking Complexity**
- Need systematic mock generation for multiple interfaces
- Example: Mocking external dependencies across packages
- Solution: Create `agents/mock-designer.md` with mocking expertise

**Trigger 5: Test Quality Assessment**
- Need systematic test quality evaluation
- Example: Reviewing test clarity, maintainability across codebase
- Solution: Create `agents/test-reviewer.md` with quality assessment expertise

### When NOT to Create Specialized Agent

**Anti-Pattern 1: One-Time Task**
- Task only needed once in experiment
- Use generic coder or data-analyst instead

**Anti-Pattern 2: Simple Task**
- Task can be accomplished with basic commands
- No specialized knowledge required

**Anti-Pattern 3: Premature Specialization**
- Haven't tried with generic agents yet
- Unclear if specialization provides value

## Testing Agent Design Template

When creating specialized testing agent:

```markdown
# [Agent Name] Agent

## Identity
You are a [agent name] agent specialized in [specific testing capability].

## Expertise
- [Testing knowledge area 1]
- [Testing knowledge area 2]
- [Testing knowledge area 3]
- [Testing tool proficiency]
- [Testing pattern mastery]

## Responsibilities
- [Primary testing responsibility]
- [Secondary testing responsibility]
- [Testing quality assurance]

## Testing Methodology

### [Key Testing Skill 1]
[Detailed description of testing approach]

### [Key Testing Skill 2]
[Detailed description of testing technique]

### [Key Testing Skill 3]
[Detailed description of testing strategy]

## Tools and Techniques
- **Go Testing**: testing package, testify, subtests, table-driven tests
- **Coverage Analysis**: go test -cover, go tool cover
- **Test Patterns**: [Specific patterns this agent masters]

## Output Format
[What this agent produces: test code, analysis, reports]

## Quality Standards
- [Testing quality criterion 1]
- [Testing quality criterion 2]
- [Testing quality criterion 3]
```

## Common Specialized Testing Agents

### agents/test-generator.md
**When to Create**: Need to generate tests for multiple functions/packages
**Expertise**:
- Table-driven test generation
- Subtest structure creation
- Edge case identification (nil, empty, boundary)
- Error case test generation
- Testify assertion usage

### agents/coverage-analyzer.md
**When to Create**: Need systematic coverage gap analysis
**Expertise**:
- Go coverage report parsing
- Gap prioritization (critical paths first)
- Branch coverage analysis
- Coverage improvement recommendations

### agents/test-optimizer.md
**When to Create**: Need to improve test execution performance
**Expertise**:
- Test profiling and timing analysis
- Parallel execution design (t.Parallel())
- Fixture optimization
- Test dependency reduction

### agents/mock-designer.md
**When to Create**: Need to mock external dependencies systematically
**Expertise**:
- Interface mock generation
- Test double design
- Fixture data creation
- Test isolation patterns

### agents/test-reviewer.md
**When to Create**: Need systematic test quality assessment
**Expertise**:
- Test clarity evaluation
- Maintainability assessment
- Test convention validation
- Refactoring recommendations

## Meta-Agent Capability Evolution

### When to Add Meta-Agent Capability

**Trigger 1: New Testing Analysis Dimension**
- Need to assess testing aspect not in existing capabilities
- Example: Mutation testing, property-based testing coverage
- Solution: Create `meta-agents/mutation-analyze.md`

**Trigger 2: New Testing Coordination Pattern**
- Need new way to coordinate testing agents
- Example: Continuous test generation based on coverage monitoring
- Solution: Extend `meta-agents/execute.md` or create new capability

**Trigger 3: New Testing Reflection Dimension**
- Need to evaluate testing state in new way
- Example: Test-to-code ratio analysis, test documentation coverage
- Solution: Extend `meta-agents/reflect.md`

### Capability Evolution Process

1. **Identify Need**: What testing capability is missing?
2. **Design Specification**: Define capability scope and interface
3. **Create Capability File**: `meta-agents/{capability}.md`
4. **Document Trigger**: Why was this capability needed?
5. **Update Execution Flow**: How does this capability integrate?

## Evolution Documentation

### For Each New Agent

Document in iteration file:
```markdown
## Agent Evolution: {agent-name}

**Trigger**: [What testing need prompted creation]
**Justification**: [Why existing agents insufficient]
**Capabilities**: [What testing expertise this agent provides]
**Expected Usage**: [When and how this agent will be used]
```

### For Each New Capability

Document in iteration file:
```markdown
## Capability Evolution: {capability-name}

**Trigger**: [What testing analysis gap identified]
**Justification**: [Why existing capabilities insufficient]
**Specification**: [What this capability enables]
**Integration**: [How it fits with existing capabilities]
```

## Evolution Validation

### After Creating Specialized Agent

1. **Effectiveness Check**: Does agent perform better than generic agent?
2. **Reusability Check**: Will agent be used in multiple iterations?
3. **Quality Check**: Does agent produce higher quality tests?
4. **Efficiency Check**: Does agent save time or improve outcomes?

### After Adding Capability

1. **Necessity Check**: Was capability truly needed?
2. **Completeness Check**: Does capability cover identified gap?
3. **Integration Check**: Does capability work well with others?
4. **Value Check**: Does capability improve V(s) calculation/progress?

## Key Principles

1. **Let Testing Needs Drive Evolution**: Don't create agents/capabilities speculatively
2. **Validate Before Specialization**: Try with generic agents first
3. **Document Justification**: Always explain why evolution was necessary
4. **Assess Effectiveness**: Validate that specialization improved outcomes
5. **Avoid Over-Engineering**: Simplest solution that meets testing needs
```

### Agent Specifications

Create these initial agent files in `agents/`:

#### agents/data-analyst.md

```markdown
# Data Analyst Agent

## Identity
You are a data analyst agent specialized in processing, aggregating, and analyzing testing metrics and coverage data.

## Expertise
- Test coverage data analysis (line, branch, package-level)
- Statistical analysis of test metrics
- Trend identification in testing data
- Testing metric aggregation and reporting

## Responsibilities
- Parse and analyze test coverage reports
- Calculate testing metrics (coverage %, test counts, execution time)
- Identify testing trends and patterns
- Generate testing metric summaries

## Methodology

### Coverage Data Analysis
- Parse `go test -cover` output
- Extract per-package and per-function coverage
- Calculate aggregate coverage statistics
- Identify coverage trends over iterations

### Test Metric Calculation
- Count test functions, subtests, benchmarks
- Measure test execution time
- Calculate test-to-code ratios
- Analyze test distribution across packages

### Gap Analysis
- Identify packages below coverage thresholds
- Find untested functions
- Detect missing test types (unit, integration, benchmark)

## Tools
- Go test coverage tools
- JSON parsing for structured data
- Statistical aggregation
- Data visualization preparation

## Output Format
- Structured JSON reports
- Summary statistics
- Prioritized gap lists
- Metric trend analysis
```

#### agents/doc-writer.md

```markdown
# Documentation Writer Agent

## Identity
You are a documentation writer agent specialized in creating clear, comprehensive testing documentation.

## Expertise
- Testing strategy documentation
- Test plan creation
- Testing methodology documentation
- Testing standard and guideline writing

## Responsibilities
- Document testing strategies and approaches
- Create test plans for new features
- Write testing guidelines and standards
- Document test results and findings

## Methodology

### Testing Strategy Documentation
- Describe testing approach (unit, integration, e2e)
- Document test coverage goals
- Explain test organization patterns
- Define quality gates

### Test Plan Creation
- Identify testing scope
- Define test scenarios
- Specify expected outcomes
- Document test data requirements

### Results Documentation
- Summarize test coverage improvements
- Document testing issues found
- Describe testing challenges and solutions
- Record testing best practices discovered

## Output Format
- Markdown documentation
- Test plans with clear sections
- Testing guidelines and standards
- Iteration summaries
```

#### agents/coder.md

```markdown
# Coder Agent

## Identity
You are a coder agent capable of implementing tests, refactoring test code, and making test-related code changes.

## Expertise
- Go test implementation
- Test refactoring
- Test fixture creation
- Simple mock implementations

## Responsibilities
- Implement straightforward tests
- Refactor existing test code
- Create test fixtures and helper functions
- Fix failing tests

## Methodology

### Test Implementation
- Follow existing test patterns (table-driven, subtests)
- Use testify assertions appropriately
- Implement proper test isolation
- Write clear test names

### Test Refactoring
- Extract duplicate setup into fixtures
- Improve test clarity and maintainability
- Optimize test execution
- Apply DRY principle

### Test Fixes
- Diagnose test failures
- Fix broken assertions
- Update tests for code changes
- Resolve race conditions

## Tools
- Go testing package
- testify/assert, testify/require
- Go test runner
- Test debugging techniques

## Output Format
- Go test code
- Test fixtures and helpers
- Refactored test implementations
```

### Objectives

Execute these objectives in order:

#### Objective 0: Setup Meta-Agent Architecture

**Task**: Create all capability files and agent files

**Steps**:
1. Create `meta-agents/` directory
2. Create all 5 capability files (observe.md, plan.md, execute.md, reflect.md, evolve.md)
3. Create `agents/` directory
4. Create all 3 initial agent files (data-analyst.md, doc-writer.md, coder.md)
5. Verify all files created successfully

**Verification**:
- All capability files exist and contain full specifications
- All agent files exist and contain full specifications
- Files follow markdown format
- Content matches templates above

**Output**: Confirmation that modular architecture is established

---

#### Objective 1: Collect Testing Data

**Capability**: Use `meta-agents/observe.md`

**Protocol**:
1. Read `meta-agents/observe.md` completely
2. Follow data collection procedures specified in Observe capability
3. Execute all testing data commands

**Testing Data Collection Tasks**:

1. **Test Coverage Analysis**
```bash
# Generate comprehensive coverage report
go test -cover ./... -coverprofile=coverage.out
go tool cover -func=coverage.out > data/coverage-summary.txt
go tool cover -html=coverage.out -o data/coverage.html

# Per-package coverage
go test -cover ./internal/... | tee data/internal-coverage.txt
go test -cover ./cmd/... | tee data/cmd-coverage.txt
```

2. **Test Inventory**
```bash
# Find all test files
find . -name "*_test.go" -type f > data/test-files.txt

# Count test types
grep -r "^func Test" . --include="*_test.go" > data/test-functions.txt
grep -r "^func Benchmark" . --include="*_test.go" > data/benchmark-functions.txt
grep -r "t\.Run(" . --include="*_test.go" > data/subtest-usage.txt
```

3. **Test Execution Metrics**
```bash
# Measure execution time
go test -v ./... -count=1 2>&1 | tee data/test-execution.log

# Run tests multiple times for flakiness detection
for i in {1..5}; do
  echo "Run $i" >> data/test-stability.log
  go test ./... >> data/test-stability.log 2>&1
done
```

4. **Code Complexity Analysis**
```bash
# Analyze test complexity (if gocyclo available)
# Otherwise, manually count test function sizes
find . -name "*_test.go" -exec wc -l {} \; > data/test-file-sizes.txt
```

**Expected Outputs**:
- `data/coverage.out`: Raw coverage data
- `data/coverage-summary.txt`: Function-level coverage
- `data/test-inventory.json`: Test counts and types
- `data/test-execution.log`: Test timing and results
- `data/test-stability.log`: Flakiness detection runs

---

#### Objective 2: Calculate Baseline V(s₀)

**Capability**: Use `meta-agents/reflect.md` and `meta-agents/plan.md`

**Protocol**:
1. Read `meta-agents/reflect.md` completely
2. Read `meta-agents/plan.md` completely
3. Follow V(s) calculation procedures

**Agent**: Invoke `agents/data-analyst.md`

**Task**: Calculate baseline V(s₀) for current testing state

**V(s₀) Calculation Requirements**:

1. **V_coverage (Weight: 0.3)**
   - Calculate line coverage from coverage-summary.txt
   - Calculate branch coverage (from coverage report analysis)
   - Apply formula: V_coverage = 0.6·V_line + 0.4·V_branch
   - Document actual percentages

2. **V_reliability (Weight: 0.3)**
   - Calculate test pass rate from test-execution.log
   - Identify critical paths (manually review error handling, core logic)
   - Estimate critical path coverage
   - Check for flaky tests (test-stability.log)
   - Apply formula: V_reliability = 0.4·pass_rate + 0.4·critical_coverage + 0.2·stability

3. **V_maintainability (Weight: 0.2)**
   - Estimate average test complexity (manual review of test files)
   - Assess test clarity (count clear vs unclear test names)
   - Identify duplicate setup code
   - Apply formula: V_maintainability = 0.4·V_complexity + 0.3·clarity_score + 0.3·duplication_score

4. **V_speed (Weight: 0.2)**
   - Calculate total test execution time from test-execution.log
   - Set as baseline_time for future comparisons
   - Count parallel tests (grep for t.Parallel())
   - Apply formula: V_speed = 0.7·V_time + 0.3·V_parallel
   - For baseline: V_time = 1.0 (no comparison yet)

5. **Overall V(s₀)**
   - Calculate: V(s₀) = 0.3·V_coverage + 0.3·V_reliability + 0.2·V_maintainability + 0.2·V_speed
   - Document each component with evidence

**Output**: `data/baseline-value.json` with detailed V(s₀) calculation

---

#### Objective 3: Identify Testing Problems

**Capability**: Use `meta-agents/reflect.md`

**Protocol**:
1. Read `meta-agents/reflect.md` completely
2. Follow gap identification procedures

**Agent**: Invoke `agents/data-analyst.md` and `agents/doc-writer.md`

**Task**: Identify specific testing problems and coverage gaps

**Problem Identification Requirements**:

1. **Coverage Gaps**
   - List all functions with <80% line coverage
   - List all packages with <70% branch coverage
   - Identify error handling paths without tests
   - Identify boundary conditions without tests

2. **Reliability Gaps**
   - List critical paths without tests
   - Identify any flaky tests
   - List missing error case tests
   - Identify insufficient edge case coverage

3. **Maintainability Gaps**
   - List tests with unclear names
   - Identify duplicate setup code (no fixtures)
   - List high complexity tests
   - Identify poor assertion specificity

4. **Speed Gaps**
   - List slow unit tests (>100ms)
   - Identify tests that could use t.Parallel()
   - List inefficient test setup/teardown
   - Identify redundant test operations

**Output**: `data/testing-problems.json` with categorized problems

---

#### Objective 4: Document Testing Strategy

**Capability**: Use `meta-agents/execute.md`

**Protocol**:
1. Read `meta-agents/execute.md` completely
2. Follow documentation procedures

**Agent**: Invoke `agents/doc-writer.md`

**Task**: Document baseline testing state and initial testing strategy

**Documentation Requirements**:

1. **Baseline Testing State**
   - Current test coverage metrics
   - Test inventory (counts by type)
   - Test quality assessment
   - Identified problems and gaps

2. **Testing Strategy**
   - Prioritized testing objectives
   - Test generation approach
   - Coverage improvement plan
   - Test quality improvement plan

3. **Next Steps**
   - Recommended tests to generate
   - Recommended test refactorings
   - Agent specialization considerations
   - Expected V(s₁) improvement

**Output**: Documentation in `ITERATION-0.md`

---

#### Objective 5: Reflect on Baseline

**Capability**: Use `meta-agents/reflect.md`

**Protocol**:
1. Read `meta-agents/reflect.md` completely
2. Follow reflection procedures

**Reflection Questions**:

1. **Current State**: What is the baseline testing state? Be specific with metrics.
2. **V(s₀) Assessment**: How does V(s₀) compare to target (0.80)? Which components are furthest from target?
3. **Critical Gaps**: What are the most critical testing gaps? Which should be addressed first?
4. **Testing Strategy**: Is the current testing approach adequate? What patterns should we adopt?
5. **Agent Needs**: Do we need specialized testing agents? What capabilities would they provide?
6. **Next Iteration**: What should be the focus of Iteration 1? What V(s₁) improvement is realistic?

**Output**: Reflection section in `ITERATION-0.md`

---

### Output Format: ITERATION-0.md

Create a comprehensive iteration document:

```markdown
# Iteration 0: Baseline Testing State

## Metadata
- Experiment: bootstrap-002-test-strategy
- Iteration: 0
- Date: [timestamp]
- Meta-Agent: M₀ (observe, plan, execute, reflect, evolve)
- Agents: A₀ (data-analyst, doc-writer, coder)

## Objectives
1. [x] Setup Meta-Agent Architecture
2. [x] Collect Testing Data
3. [x] Calculate Baseline V(s₀)
4. [x] Identify Testing Problems
5. [x] Document Testing Strategy
6. [x] Reflect on Baseline

## Architecture Setup

### Capability Files Created
- meta-agents/observe.md: Testing data collection and pattern recognition
- meta-agents/plan.md: Testing objective prioritization and agent selection
- meta-agents/execute.md: Test generation coordination and validation
- meta-agents/reflect.md: V(s) calculation and gap identification
- meta-agents/evolve.md: Agent specialization and capability evolution

### Agent Files Created
- agents/data-analyst.md: Testing metric analysis
- agents/doc-writer.md: Testing documentation
- agents/coder.md: Test implementation

## Testing Data Collection

### Test Coverage Analysis
[Coverage metrics from go test -cover]

### Test Inventory
[Test counts, types, patterns]

### Test Execution Metrics
[Timing, pass rates, stability]

### Code Complexity
[Complexity metrics]

## Baseline Value Calculation: V(s₀)

### V_coverage = [value]
- Line coverage: [X]%
- Branch coverage: [Y]%
- Calculation: 0.6·[V_line] + 0.4·[V_branch] = [value]

### V_reliability = [value]
- Test pass rate: [X]%
- Critical path coverage: [Y]%
- Stability: [Z]%
- Calculation: 0.4·[pass_rate] + 0.4·[critical_coverage] + 0.2·[stability] = [value]

### V_maintainability = [value]
- Average test complexity: [X]
- Test clarity: [Y]%
- Duplication score: [Z]
- Calculation: 0.4·[V_complexity] + 0.3·[clarity] + 0.3·[duplication] = [value]

### V_speed = [value]
- Baseline execution time: [X]s
- Parallel ratio: [Y]%
- Calculation: 0.7·[V_time] + 0.3·[V_parallel] = [value]

### Overall V(s₀) = [value]
Calculation: 0.3·[V_coverage] + 0.3·[V_reliability] + 0.2·[V_maintainability] + 0.2·[V_speed] = [value]

Target: V(sₙ) ≥ 0.80
Gap: [0.80 - V(s₀)] = [gap]

## Testing Problems Identified

### Coverage Gaps
[Specific functions/packages below thresholds]

### Reliability Gaps
[Critical paths untested, flaky tests]

### Maintainability Gaps
[Unclear tests, duplication]

### Speed Gaps
[Slow tests, parallelization opportunities]

## Testing Strategy

### Priority 1: [Highest priority testing objective]
[Description, rationale, approach]

### Priority 2: [Second priority testing objective]
[Description, rationale, approach]

### Priority 3: [Third priority testing objective]
[Description, rationale, approach]

### Test Generation Plan
[What tests to generate, patterns to use]

### Agent Specialization Considerations
[Should we create specialized testing agents? Which ones?]

## Next Steps for Iteration 1

### Primary Goal
[Main testing improvement to achieve]

### Expected V(s₁)
[Estimated V(s₁) with rationale]

### Testing Tasks
[Specific tests to generate, refactorings to perform]

## Reflection

### Current State Assessment
[Honest assessment of baseline testing state]

### V(s₀) Analysis
[Which components strong/weak, why]

### Critical Gaps
[Most important testing gaps to address]

### Strategy Validation
[Is our testing approach sound?]

### Agent Needs
[Do we need specialized agents? Which ones?]

### Learning
[What did we learn from baseline analysis?]

## Data Artifacts

- data/coverage.out: Raw coverage data
- data/coverage-summary.txt: Function-level coverage
- data/test-inventory.json: Test counts and types
- data/test-execution.log: Test timing and results
- data/baseline-value.json: V(s₀) calculation details
- data/testing-problems.json: Categorized problems
```

---

## Iteration 1+: Subsequent Iterations (General Template)

**Purpose**: Systematic testing improvement through test generation, coverage analysis, and agent evolution.

### Pre-Iteration Protocol

Before starting any iteration N (where N ≥ 1):

1. **Read Previous Iteration**:
   - `ITERATION-{N-1}.md`: Extract state, problems, next steps

2. **Read ALL Capability Files**:
   - `meta-agents/observe.md`
   - `meta-agents/plan.md`
   - `meta-agents/execute.md`
   - `meta-agents/reflect.md`
   - `meta-agents/evolve.md`
   - Any additional capability files created

3. **Extract Context**:
   - Previous state: V(sₙ₋₁)
   - Previous agents: Aₙ₋₁
   - Previous problems identified
   - Previous next steps

### Iteration N Execution

You are Meta-Agent Mₙ executing Iteration N of bootstrap-002-test-strategy.

Execute these five steps systematically:

---

#### Step 1: Observe Testing State

**Capability**: Read `meta-agents/observe.md` before starting

**Tasks**:
1. Collect current testing data (coverage, test inventory, execution metrics)
2. Compare with previous iteration data
3. Identify testing patterns and trends
4. Discover new testing gaps or issues

**Commands** (from observe.md):
```bash
# Update coverage data
go test -cover ./... -coverprofile=coverage.out
go tool cover -func=coverage.out > data/coverage-current.txt

# Check test execution
go test -v ./... -count=1 2>&1 | tee data/test-execution-current.log

# Verify test stability
go test ./... -count=5
```

**Output**:
- Updated testing data in `data/` directory
- Comparison with previous iteration
- New testing patterns discovered

---

#### Step 2: Plan Testing Objectives

**Capability**: Read `meta-agents/plan.md` before starting

**Tasks**:
1. Define primary goal for this iteration (based on previous iteration's next steps)
2. Prioritize testing objectives using Priority Framework from plan.md
3. Assess current agents Aₙ₋₁ against testing tasks
4. Determine if specialized agents needed (use evolve.md)

**Planning Questions**:
- What is the primary testing objective for this iteration?
- Which testing tasks have highest priority?
- Can existing agents accomplish these tasks effectively?
- Do we need to create specialized testing agents?
- What success criteria for this iteration (expected ΔV)?

**Output**:
- Primary goal statement
- Prioritized testing tasks
- Agent capability assessment
- Specialization decision (if needed)

---

#### Step 3: Execute Testing Work

**Capability**: Read `meta-agents/execute.md` before starting

**Protocol**:

1. **If Existing Agents Sufficient**:
   - Read relevant agent files before invocation
   - Invoke agents with clear testing tasks
   - Follow coordination patterns from execute.md

2. **If Specialized Agent Needed**:
   - Read `meta-agents/evolve.md`
   - Determine which specialized agent to create
   - Create agent file: `agents/{agent-name}.md`
   - Document justification in iteration file
   - Read new agent file
   - Invoke new agent

**Common Testing Work Patterns**:

**Pattern A: Generate Tests for Coverage Gaps**
1. Invoke coverage-analyzer (or data-analyst) to prioritize gaps
2. Invoke test-generator (or coder) to generate tests
3. Run tests and validate coverage improvement
4. Document generated tests

**Pattern B: Optimize Test Performance**
1. Invoke test-optimizer (or data-analyst) to profile tests
2. Identify slow tests and parallelization opportunities
3. Invoke coder to implement optimizations
4. Measure execution time improvement

**Pattern C: Improve Test Quality**
1. Review existing tests for maintainability issues
2. Invoke coder to refactor (extract fixtures, clarify names)
3. Validate test clarity and DRY principle
4. Document improvements

**Testing Work Output**:
- Generated test code or refactored tests
- Test execution results
- Coverage reports
- Agent invocation records

---

#### Step 4: Reflect on Testing State

**Capability**: Read `meta-agents/reflect.md` before starting

**Tasks**:
1. Calculate V(sₙ) using formulas from reflect.md
2. Compare V(sₙ) with V(sₙ₋₁): Calculate ΔV
3. Assess quality of testing work against checklist
4. Identify remaining testing gaps
5. Evaluate agent effectiveness

**V(sₙ) Calculation** (follow reflect.md exactly):
- V_coverage: Line and branch coverage
- V_reliability: Test pass rate, critical path coverage, stability
- V_maintainability: Test complexity, clarity, duplication
- V_speed: Execution time, parallelization
- Overall: V(sₙ) = 0.3·V_coverage + 0.3·V_reliability + 0.2·V_maintainability + 0.2·V_speed

**Progress Assessment**:
- ΔV = V(sₙ) - V(sₙ₋₁)
- Component-level progress: ΔV_coverage, ΔV_reliability, ΔV_maintainability, ΔV_speed
- Which component blocking convergence?

**Quality Assessment** (from reflect.md checklist):
- [ ] Tests pass consistently
- [ ] Coverage targets met
- [ ] Critical paths tested
- [ ] Clear test names
- [ ] Specific assertions
- [ ] Fast execution
- [ ] No flaky tests

**Output**:
- Complete V(sₙ) calculation with evidence
- ΔV analysis
- Quality assessment
- Remaining gaps identified

---

#### Step 5: Check Convergence

**Convergence Criteria** (all must be met):

1. **Value Target**: V(sₙ) ≥ 0.80
2. **Coverage Target**: Line coverage ≥80%, branch coverage ≥70%
3. **Stability**: ΔV < 0.02 for 2 consecutive iterations
4. **Quality Gates**: All quality checklist items satisfied
5. **Problem Resolution**: All critical testing gaps addressed

**Status Determination**:

**If NOT Converged**:
- Determine focus for Iteration N+1
- Specify testing tasks for next iteration
- Estimate expected V(sₙ₊₁)
- Document next steps

**If Converged**:
- Validate all convergence criteria met
- Document final testing state
- Proceed to Results Analysis (see Final Iteration section)

**Output**:
- Convergence status (converged / not converged)
- If not converged: Next iteration plan
- If converged: Final state documentation

---

### Output Format: ITERATION-N.md

Create comprehensive iteration document:

```markdown
# Iteration N: [Iteration Title]

## Metadata
- Experiment: bootstrap-002-test-strategy
- Iteration: N
- Date: [timestamp]
- Meta-Agent: Mₙ (capabilities used)
- Agents: Aₙ (agents invoked)

## Context from Iteration N-1

### Previous State
- V(sₙ₋₁) = [value]
- Agents: [list]
- Problems: [summary]
- Next steps planned: [summary]

## Primary Goal
[Primary testing objective for this iteration]

## Evolution (if any)

### New Agents Created
**agents/{agent-name}.md**:
- Trigger: [What testing need prompted creation]
- Justification: [Why existing agents insufficient]
- Capabilities: [What testing expertise provided]

### New Capabilities Created
**meta-agents/{capability}.md**:
- Trigger: [What testing gap identified]
- Justification: [Why new capability needed]
- Specification: [What capability enables]

## Testing Work Performed

### Observe Phase
[Testing data collected, patterns discovered]

### Plan Phase
[Testing objectives prioritized, agent selection]

### Execute Phase
[Testing work performed, agents invoked]

**Agent Invocations**:
1. [Agent name]: [Task] → [Output]
2. [Agent name]: [Task] → [Output]

### Testing Artifacts
[Tests generated, coverage improved, refactorings performed]

## State Transition: sₙ₋₁ → sₙ

### V(sₙ) Calculation

#### V_coverage = [value]
[Calculation with evidence]

#### V_reliability = [value]
[Calculation with evidence]

#### V_maintainability = [value]
[Calculation with evidence]

#### V_speed = [value]
[Calculation with evidence]

#### Overall V(sₙ) = [value]
Calculation: 0.3·[V_coverage] + 0.3·[V_reliability] + 0.2·[V_maintainability] + 0.2·[V_speed]

### Progress Analysis
- ΔV = V(sₙ) - V(sₙ₋₁) = [value]
- ΔV_coverage = [value]
- ΔV_reliability = [value]
- ΔV_maintainability = [value]
- ΔV_speed = [value]

## Reflection

### What Worked Well
[Successes in this iteration]

### What Didn't Work
[Challenges or failures]

### Quality Assessment
[Assessment against quality checklist]

### Remaining Gaps
[Testing gaps still to address]

### Agent Effectiveness
[How well did agents perform?]

### Learning
[What did we learn about testing?]

## Convergence Check

### Criteria Status
1. Value Target (V ≥ 0.80): [✓/✗] V(sₙ) = [value]
2. Coverage Target (≥80% line, ≥70% branch): [✓/✗] [actual]
3. Stability (ΔV < 0.02): [✓/✗] ΔV = [value]
4. Quality Gates: [✓/✗] [summary]
5. Problem Resolution: [✓/✗] [status]

### Convergence Status
[CONVERGED / NOT CONVERGED]

## Next Steps (if not converged)

### Iteration N+1 Focus
[Primary objective for next iteration]

### Testing Tasks
[Specific tests to generate, refactorings to perform]

### Expected V(sₙ₊₁)
[Estimated value with rationale]

## Data Artifacts

- data/coverage-iteration-N.out: Coverage data
- data/test-execution-iteration-N.log: Test execution
- data/value-calculation-N.json: V(sₙ) details
- [Generated test files]
- [Other artifacts]
```

---

### Key Principles for Iteration Execution

1. **Read Before Acting**: Always read capability files before using them, agent files before invoking
2. **Honest Calculation**: Calculate V(s) honestly based on actual coverage and metrics
3. **Justify Evolution**: Only create specialized agents when clearly needed, document why
4. **Complete Analysis**: No token limits - perform complete testing analysis
5. **Data-Driven**: Base decisions on actual coverage data and test metrics
6. **Quality Focus**: Don't just increase test count, improve test quality
7. **Critical Paths First**: Prioritize testing error handling and core logic
8. **No Predetermined Evolution**: Let testing needs drive agent creation, not assumptions

---

## Final Iteration: Results Analysis

**Purpose**: Comprehensive analysis of converged testing system and methodology validation.

### Context

You are Meta-Agent Mₙ at convergence. All convergence criteria have been met:
- V(sₙ) ≥ 0.80
- Coverage targets achieved (≥80% line, ≥70% branch)
- System stable (ΔV < 0.02)
- Quality gates satisfied
- Critical testing gaps addressed

Your task is to perform comprehensive results analysis.

### Analysis Dimensions

#### 1. Final Three-Tuple Output

**Meta-Agent Capabilities Mₙ**:
- List all capabilities in meta-agents/
- Describe each capability's purpose and effectiveness
- Document how capabilities evolved from M₀

**Final Agent Set Aₙ**:
- List all agents in agents/
- Describe each agent's specialization
- Document agent evolution from A₀
- Provide examples of agent usage

**Organizational Structure Oₙ**:
- Test organization patterns established
- Testing workflows and coordination protocols
- Quality gates and standards defined
- Automation and tooling implemented

#### 2. Convergence Validation

**Verify All Criteria**:
- V(sₙ) = [value] ≥ 0.80 ✓
- Line coverage: [X]% ≥ 80% ✓
- Branch coverage: [Y]% ≥ 70% ✓
- Stability: ΔV = [value] < 0.02 ✓
- Quality gates: [all satisfied] ✓
- Problems: [all resolved] ✓

**Evidence**:
- Final coverage reports
- Test execution results
- Quality assessment
- Problem resolution documentation

#### 3. Value Space Trajectory

**V(s) Over Iterations**:
```
Iteration 0: V(s₀) = [value]
Iteration 1: V(s₁) = [value] (ΔV = [delta])
Iteration 2: V(s₂) = [value] (ΔV = [delta])
...
Iteration N: V(sₙ) = [value] (ΔV = [delta])
```

**Component Trajectories**:
- V_coverage trajectory: [s₀ → sₙ]
- V_reliability trajectory: [s₀ → sₙ]
- V_maintainability trajectory: [s₀ → sₙ]
- V_speed trajectory: [s₀ → sₙ]

**Visualization**:
- Line graph of V(s) over iterations
- Stacked area chart of component contributions
- Bar chart of ΔV per iteration

#### 4. Testing Domain Analysis

**Test Coverage Evolution**:
- Initial coverage: [X]%
- Final coverage: [Y]%
- Improvement: [+Z]%
- Most improved packages

**Test Quality Evolution**:
- Initial test count: [X]
- Final test count: [Y]
- Table-driven test adoption
- Subtest structure usage
- Fixture and helper usage

**Test Performance Evolution**:
- Initial execution time: [X]s
- Final execution time: [Y]s
- Parallelization improvements
- Slow test elimination

**Testing Methodology Established**:
- Test generation patterns
- Coverage analysis workflows
- Test optimization techniques
- Quality assurance standards

#### 5. Reusability Validation

**Domain Transfer Tests**:

**Test 1: Apply to Another Go Project**
- Select different Go project (e.g., another CLI tool)
- Apply testing methodology (use Mₙ, Aₙ)
- Measure effectiveness (coverage improvement, test quality)
- Document transferability

**Test 2: Apply to Different Test Type**
- Apply to integration test generation (vs unit tests)
- Measure methodology effectiveness
- Document adaptations needed

**Test 3: Apply to Different Testing Tool**
- Apply to different assertion library
- Measure methodology robustness
- Document tool-independence

**Transferability Assessment**:
- What aspects transferred easily?
- What required adaptation?
- How reusable are specialized agents?
- How reusable is Meta-Agent methodology?

#### 6. Comparison with Actual History

**Historical Testing Development**:
- How was testing actually developed in meta-cc?
- What testing problems were encountered?
- How were they solved?
- What patterns emerged?

**Methodology Comparison**:
- Would systematic methodology have been faster?
- Would coverage have been better?
- Would test quality have been higher?
- What was gained by systematic approach?

**Key Differences**:
- Methodology benefits observed
- Methodology overhead observed
- Unexpected discoveries

#### 7. Methodology Validation

**Observe-Codify-Automate (OCA)**:
- Iteration 0-1 (Observe): Did we discover testing patterns?
- Iteration 2-3 (Codify): Did we establish testing taxonomy and procedures?
- Iteration 3-4 (Automate): Did we build testing automation?
- Did OCA pattern match actual needs?

**Bootstrapped System Evolution (BSE)**:
- Did agents specialize appropriately?
- Did Meta-Agent capabilities evolve effectively?
- Was evolution driven by actual needs?
- Did specialized agents improve outcomes?

**Value Space Navigation**:
- Did V(s) guide decisions effectively?
- Did component weights make sense?
- Was convergence criteria appropriate?
- Did value function capture testing quality?

#### 8. Key Learnings

**About Testing**:
- What did we learn about Go testing best practices?
- What testing patterns are most effective?
- What testing anti-patterns to avoid?
- How to balance coverage vs maintainability?

**About System Evolution**:
- When is agent specialization beneficial?
- How should capabilities be modularized?
- What coordination patterns work best?
- How to avoid premature specialization?

**About Methodology**:
- What worked well in OCA pattern?
- What value function design principles?
- What convergence criteria are appropriate?
- How to balance rigor vs pragmatism?

#### 9. Scientific Contribution

**Systematic Testing Methodology**:
- Comprehensive testing strategy for Go projects
- Value-driven testing improvement approach
- Systematic coverage gap analysis
- Quality-focused test generation

**Meta-Agent Design Patterns**:
- Modular capability architecture for testing
- Testing agent specialization patterns
- Test coordination and validation workflows
- Testing reflection and convergence criteria

**Reusable Artifacts**:
- Testing agents (test-generator, coverage-analyzer, etc.)
- Testing capabilities (observe, reflect for testing)
- Testing value function design
- Testing quality assessment framework

#### 10. Future Work

**Testing Methodology Extensions**:
- Mutation testing integration
- Property-based testing generation
- Fuzz testing strategy
- Performance regression testing

**Meta-Agent Enhancements**:
- Continuous testing monitoring
- Automated test maintenance
- Test smell detection
- Test-to-code ratio optimization

**Domain Generalization**:
- Apply to other languages (Python, JavaScript, Rust)
- Apply to other testing types (security, performance)
- Apply to other quality dimensions (documentation, API)

**Research Questions**:
- Optimal value function weights for testing?
- Convergence criteria for different project types?
- Agent specialization vs generalization tradeoffs?
- Testing methodology transferability limits?

---

### Output Format: RESULTS.md

```markdown
# Results: Bootstrap-002-Test-Strategy

## Experiment Summary

**Experiment ID**: bootstrap-002-test-strategy
**Domain**: Software Testing and Quality Assurance (Go)
**Iterations**: [N] iterations to convergence
**Final State**: V(sₙ) = [value] ≥ 0.80

## 1. Final Three-Tuple Output (Oₙ, Aₙ, Mₙ)

### Meta-Agent Capabilities Mₙ
[Detailed description of final capabilities]

### Final Agent Set Aₙ
[Detailed description of final agents]

### Organizational Structure Oₙ
[Detailed description of testing organization]

## 2. Convergence Validation

### Criteria Verification
[Verification of all 5 criteria with evidence]

### Convergence Timeline
[When and how convergence achieved]

## 3. Value Space Trajectory

### V(s) Evolution
[Table and graph of V(s) over iterations]

### Component Trajectories
[Analysis of each component's evolution]

### Inflection Points
[Key moments in value trajectory]

## 4. Testing Domain Analysis

### Coverage Evolution
[Coverage improvement analysis]

### Quality Evolution
[Test quality improvement analysis]

### Performance Evolution
[Test execution performance analysis]

### Methodology Established
[Testing methodology summary]

## 5. Reusability Validation

### Transfer Test 1: Other Go Project
[Results and analysis]

### Transfer Test 2: Different Test Type
[Results and analysis]

### Transfer Test 3: Different Testing Tool
[Results and analysis]

### Transferability Assessment
[Overall reusability evaluation]

## 6. Comparison with Actual History

### Historical Development
[How testing was actually developed]

### Methodology Comparison
[Systematic vs ad-hoc comparison]

### Key Differences
[Benefits and overhead observed]

## 7. Methodology Validation

### OCA Pattern Application
[How OCA manifested in testing domain]

### BSE Application
[How system evolution occurred]

### Value Space Navigation
[How V(s) guided decisions]

## 8. Key Learnings

### About Testing
[Testing insights gained]

### About System Evolution
[Evolution insights gained]

### About Methodology
[Methodology insights gained]

## 9. Scientific Contribution

### Systematic Testing Methodology
[Contribution description]

### Meta-Agent Design Patterns
[Pattern catalog]

### Reusable Artifacts
[Artifacts produced]

## 10. Future Work

### Testing Methodology Extensions
[Future testing research]

### Meta-Agent Enhancements
[Future Meta-Agent research]

### Domain Generalization
[Future generalization research]

### Research Questions
[Open questions]

## Appendices

### Appendix A: All Iteration Summaries
[Brief summary of each iteration]

### Appendix B: Agent Specifications
[Full specifications of all agents in Aₙ]

### Appendix C: Capability Specifications
[Full specifications of all capabilities in Mₙ]

### Appendix D: Testing Artifacts
[Generated tests, coverage reports, documentation]

### Appendix E: Data Analysis
[Detailed data analysis and visualizations]
```

---

## Quick Reference: Iteration Checklist

### Pre-Iteration
- [ ] Read previous iteration file (ITERATION-{N-1}.md)
- [ ] Read ALL capability files in meta-agents/
- [ ] Extract context: V(sₙ₋₁), Aₙ₋₁, problems, next steps
- [ ] Understand primary goal for this iteration

### Step 1: Observe
- [ ] Read meta-agents/observe.md
- [ ] Collect testing data (coverage, inventory, execution)
- [ ] Run: `go test -cover ./... -coverprofile=coverage.out`
- [ ] Run: `go tool cover -func=coverage.out`
- [ ] Run: `go test -v ./...` (measure timing)
- [ ] Compare with previous iteration data
- [ ] Identify new testing patterns or gaps

### Step 2: Plan
- [ ] Read meta-agents/plan.md
- [ ] Define primary goal for iteration
- [ ] Prioritize testing objectives (Critical Path → Quality Gates → Reliability → Maintainability → Performance)
- [ ] Read all current agent files
- [ ] Assess if existing agents sufficient
- [ ] If not sufficient, read meta-agents/evolve.md
- [ ] Decide on specialized agent creation (if needed)
- [ ] Define success criteria (expected ΔV)

### Step 3: Execute
- [ ] Read meta-agents/execute.md
- [ ] If creating new agent:
  - [ ] Create agents/{agent-name}.md with full specification
  - [ ] Document justification in iteration file
  - [ ] Read new agent file
- [ ] Read agent file before each invocation
- [ ] Invoke agents with clear testing tasks
- [ ] Generate tests or refactor existing tests
- [ ] Run tests: `go test ./...`
- [ ] Verify tests pass consistently (run 5 times)
- [ ] Measure coverage improvement
- [ ] Validate test quality (clear names, good assertions, fast execution)

### Step 4: Reflect
- [ ] Read meta-agents/reflect.md
- [ ] Calculate V_coverage (line + branch coverage)
- [ ] Calculate V_reliability (pass rate + critical path coverage + stability)
- [ ] Calculate V_maintainability (complexity + clarity + duplication)
- [ ] Calculate V_speed (execution time + parallelization)
- [ ] Calculate overall V(sₙ) = 0.3·V_coverage + 0.3·V_reliability + 0.2·V_maintainability + 0.2·V_speed
- [ ] Calculate ΔV = V(sₙ) - V(sₙ₋₁)
- [ ] Analyze component-level progress
- [ ] Assess quality against checklist
- [ ] Identify remaining testing gaps
- [ ] Evaluate agent effectiveness

### Step 5: Convergence
- [ ] Check: V(sₙ) ≥ 0.80?
- [ ] Check: Line coverage ≥80%, branch coverage ≥70%?
- [ ] Check: ΔV < 0.02 for 2 consecutive iterations?
- [ ] Check: All quality checklist items satisfied?
- [ ] Check: All critical testing gaps addressed?
- [ ] If NOT converged: Plan Iteration N+1
- [ ] If converged: Proceed to Results Analysis

### Documentation
- [ ] Create ITERATION-N.md with all sections
- [ ] Include metadata (iteration, date, Meta-Agent, agents)
- [ ] Document evolution (new agents/capabilities with justification)
- [ ] Document testing work performed
- [ ] Include complete V(sₙ) calculation with evidence
- [ ] Include progress analysis (ΔV)
- [ ] Include reflection (what worked, what didn't, learning)
- [ ] Include convergence check
- [ ] Include next steps (if not converged)
- [ ] Save all data artifacts in data/ directory

### No Token Limits
- [ ] Complete ALL analysis, don't abbreviate
- [ ] Show ALL calculations with evidence
- [ ] List ALL testing gaps identified
- [ ] Document ALL agent invocations
- [ ] Include ALL reflection questions
- [ ] Verify completeness before submitting

---

## Notes on Execution Style

### Be the Meta-Agent

You ARE the Meta-Agent M for the testing domain. You are not just following instructions, you are:
- **Observing**: Discovering testing patterns in the meta-cc codebase
- **Planning**: Making strategic testing decisions based on coverage data
- **Executing**: Coordinating test generation and validation work
- **Reflecting**: Honestly assessing testing quality and calculating V(s)
- **Evolving**: Determining when testing specialization is needed

### Be Rigorous with V(s)

Calculate V(s) honestly based on ACTUAL data:
- Don't round coverage percentages
- Don't assume critical paths are tested without verification
- Don't ignore test maintainability issues
- Measure test execution time objectively
- Count actual flaky tests, duplicate code, unclear names

Each component must have EVIDENCE:
- V_coverage: Actual coverage report output
- V_reliability: Test execution logs, critical path analysis
- V_maintainability: Test file analysis, complexity metrics
- V_speed: Measured execution time, parallel test count

### Be Thorough (No Token Limits)

This is a comprehensive testing methodology experiment. Do NOT abbreviate:
- List ALL functions below coverage threshold
- Analyze ALL test files for quality
- Document ALL agent invocations with full context
- Show ALL calculations step-by-step
- Include ALL reflection questions and answers
- Generate COMPLETE iteration documents

### Be Authentic

Discover testing patterns, don't assume them:
- What table-driven test patterns are actually used in meta-cc?
- What test organization patterns exist?
- What testing problems actually occur?
- What coverage gaps actually exist?
- Let ACTUAL data drive testing decisions

### Be Modular with Reading

The modular architecture requires disciplined reading:
- Read ALL capability files at the start of each iteration
- Read SPECIFIC capability file before using that capability
- Read agent file IMMEDIATELY before invoking that agent
- NEVER rely on cached understanding - always read fresh
- This ensures complete context and no assumptions

Why modular reading matters:
- Capabilities may have been updated
- Agent specifications may have evolved
- Fresh reading ensures complete context
- Prevents assumptions from cached memory
- Maintains consistency with actual capability specs

### Testing Domain Focus

This is about Go testing. Use testing terminology throughout:
- "Generate table-driven tests" not "create tests"
- "Improve branch coverage" not "add more tests"
- "Refactor test fixtures" not "optimize setup"
- "Analyze critical path coverage" not "check important code"

Reference Go testing tools explicitly:
- `go test -cover ./...`
- `go tool cover -func=coverage.out`
- testify/assert, testify/require
- `t.Run()` for subtests
- `t.Parallel()` for parallel tests

---

## Version Information

- **Template Version**: 1.0.0
- **Created**: 2025-10-15
- **Experiment**: bootstrap-002-test-strategy
- **Domain**: Software Testing and Quality Assurance (Go)
- **Purpose**: Guide systematic execution of testing methodology development experiment
- **Alignment**: Meta-Agent Bootstrapping methodology, Iteration-Executor agent protocol
