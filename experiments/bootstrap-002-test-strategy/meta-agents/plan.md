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
