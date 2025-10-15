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
