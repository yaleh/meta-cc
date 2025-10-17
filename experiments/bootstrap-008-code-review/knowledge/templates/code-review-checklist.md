# Go Code Review Checklist

**Purpose**: Systematic code review template based on observed patterns
**Domain**: Go code review
**Created**: Iteration 2
**Source**: Taxonomy from iterations 1-2, validated across 4 modules (parser, analyzer, query, validation)

---

## How to Use This Checklist

1. **Read module completely** before checking individual items
2. **Check each category** systematically (don't skip sections)
3. **Document findings** with issue ID, severity, location, recommendation
4. **Assign severity** based on impact (Critical > High > Medium > Low)
5. **Provide code examples** for medium+ severity issues

**Time estimate**: 15-30 minutes per 100 lines (after template internalization)

---

## 1. Correctness (HIGHEST PRIORITY)

### 1.1 Logic and Edge Cases
- [ ] **Nil pointer checks**: Are nil parameters validated?
  - Example: `BuildFileAccessQuery(entries, path)` - check `entries != nil`
- [ ] **Empty input handling**: Does code handle empty slices/strings/maps?
  - Example: `BuildContextQuery([], sig, 3)` - returns empty occurrences?
- [ ] **Boundary conditions**: Off-by-one errors, slice bounds, integer overflow?
  - Example: `timeline[0]` and `timeline[len-1]` access - check `len >= 2`
- [ ] **Loop invariants**: Do loops terminate? Are indices correct?
  - Example: `for i := 0; i <= len(tools)-seqLen; i++` - correct bound?

### 1.2 Error Handling
- [ ] **Error return patterns**: Do functions return explicit errors vs silent failures?
  - ❌ ANTI-PATTERN: `return 0` on error (ambiguous with valid 0)
  - ✅ CORRECT: `return 0, fmt.Errorf("failed to parse: %w", err)`
- [ ] **Error wrapping**: Are errors wrapped with context using `%w`?
  - Example: `fmt.Errorf("failed to read %q: %w", path, err)`
- [ ] **Deferred operations**: Are deferred `Close()`, `Unlock()` errors checked?
  - ❌ ANTI-PATTERN: `defer file.Close()` (ignores error)
  - ✅ CORRECT: Named return + error check in defer
- [ ] **Error propagation**: Are all errors from called functions checked?
  - Check for `_` = err patterns (should be rare and justified)

### 1.3 Type Safety and Conversions
- [ ] **Type assertions**: Are type assertions safe-checked `val, ok := x.(Type)`?
- [ ] **Interface conversions**: Nil interface checks before method calls?
- [ ] **Integer overflow**: Converting int64 → int, float64 → int safe on 32-bit?
  - Example: `int((last - first) / 60)` - cap to MaxInt?

### 1.4 Concurrency (if applicable)
- [ ] **Race conditions**: Are shared variables protected (mutex, channels)?
- [ ] **Goroutine leaks**: Are goroutines cleaned up (context cancellation)?
- [ ] **Channel deadlocks**: Are sends/receives balanced?

---

## 2. Performance

### 2.1 Algorithm Efficiency
- [ ] **O(n*m) iterations**: Nested loops over same/related data?
  - ❌ ANTI-PATTERN: `for _, occ := range occurrences { for _, tc := range toolCalls { ... }}`
  - ✅ FIX: Build index map once, then lookup `O(n+m)` instead of `O(n*m)`
- [ ] **Multi-pass algorithms**: Can single-pass replace multiple iterations?
  - Example: Build turn index + extract tool calls in one pass
- [ ] **Sorting necessity**: Is sorting actually needed or can you use heap/select?
  - Example: Top-N selection doesn't need full sort

### 2.2 Memory Allocation
- [ ] **String concatenation in loops**: Use `strings.Builder` instead?
  - ❌ ANTI-PATTERN: `str += part` in loop
  - ✅ FIX: `builder.WriteString(part)`
- [ ] **Unnecessary copying**: Are large structs passed by pointer?
- [ ] **Slice preallocation**: Known capacity? `make([]T, 0, cap)` instead of `[]T{}`?

### 2.3 I/O Optimization
- [ ] **Buffered I/O**: Are file reads buffered (`bufio.Reader`)?
- [ ] **Repeated file operations**: Can you batch or cache?

---

## 3. Maintainability

### 3.1 Complexity
- [ ] **Function length**: Functions >50 lines? Consider splitting
- [ ] **Cyclomatic complexity**: >10 branches? Refactor to reduce complexity
- [ ] **Nesting depth**: >3 levels? Extract to functions or early returns

### 3.2 Code Duplication
- [ ] **Repeated logic**: Similar code blocks (>5 lines) in multiple places?
  - Example: `buildContextBefore`/`buildContextAfter` nearly identical
  - ✅ FIX: Extract common logic to `buildContext(direction string)`
- [ ] **Copy-paste patterns**: Search for similar variable names, structure

### 3.3 Coupling and Cohesion
- [ ] **God objects**: Does one file/struct handle too many concerns?
  - Example: workflow.go (479 lines) handles sequences, churn, idle detection
  - ✅ FIX: Split to sequences.go, churn.go, idle.go
- [ ] **Circular dependencies**: Do packages import each other?
- [ ] **Interface segregation**: Are interfaces small and focused?

---

## 4. Readability

### 4.1 Naming
- [ ] **Variable names**: Descriptive for scope (short in small scopes, long in large)?
  - ✅ GOOD: `i`, `err`, `ok` in small scopes
  - ✅ GOOD: `errorSignature`, `toolCallWithTurn` in larger scopes
  - ❌ BAD: `x`, `tmp`, `data` in large scopes
- [ ] **Function names**: Verb-based, describe action?
  - ✅ GOOD: `BuildContextQuery`, `parseTimestamp`, `isValidToolName`
  - ❌ BAD: `process`, `handle`, `do` (too vague)
- [ ] **Package names**: Short, lowercase, single-word?
  - ✅ GOOD: `query`, `validation`, `parser`
  - ❌ BAD: `queryutils`, `validationhelpers`

### 4.2 Structure and Organization
- [ ] **Logical grouping**: Related functions together, separated by comments?
- [ ] **Function ordering**: Public before private, callers before callees?
- [ ] **Early returns**: Used to reduce nesting?
  - ✅ GOOD: `if err != nil { return err }` at top
  - ❌ BAD: Deep nesting with `if err == nil { ... }`

### 4.3 Comments and Documentation
- [ ] **godoc on exported functions**: All exported functions documented?
  - Format: `// FunctionName <verb> <description>. <details>.`
  - Example: `// BuildContextQuery builds a context query for error signature.`
- [ ] **godoc on private helpers**: Complex private functions documented?
  - **Required if**: >10 lines, non-obvious logic, complex algorithm
- [ ] **Comment accuracy**: Do comments match code? Outdated comments removed?
- [ ] **TODO/FIXME**: Are they tracked in issue tracker? Justified if inline?

---

## 5. Go Best Practices and Idioms

### 5.1 Idiomatic Patterns
- [ ] **Error wrapping**: Use `fmt.Errorf("...: %w", err)` not `fmt.Sprintf`
- [ ] **Early returns**: Reduce nesting with early validation
- [ ] **Accept interfaces, return structs**: Function params are interfaces?
- [ ] **Zero values**: Leverage zero values (empty slice, nil map, false bool)?

### 5.2 Naming Conventions
- [ ] **Getters**: No `Get` prefix (use `Balance()` not `GetBalance()`)?
- [ ] **Interfaces**: `-er` suffix (Reader, Writer, Closer)?
- [ ] **Acronyms**: All caps or all lowercase (HTTP, http, not Http)?
- [ ] **Receiver names**: Consistent, short (1-2 letters), not `this`/`self`?

### 5.3 Anti-Patterns to Flag
- [ ] **Naked returns**: In long functions (>10 lines)?
- [ ] **Defer in loops**: Accumulates resources (use manual cleanup)?
- [ ] **Goroutine without context**: Missing cancellation mechanism?
- [ ] **Exported mutable globals**: Can callers modify (break encapsulation)?
  - ❌ BAD: `var BuiltinTools = map[string]bool{...}`
  - ✅ FIX: Unexport and provide `IsBuiltinTool(name string) bool`

---

## 6. Security

### 6.1 Input Validation
- [ ] **Untrusted input**: Are file paths, user input, external data validated?
  - Example: File path traversal (`.` and `..` in paths)
  - Example: Command injection (shell metacharacters)
- [ ] **Resource exhaustion**: Are slice/map sizes bounded to prevent OOM?
- [ ] **Regex complexity**: ReDoS risk from complex regex on untrusted input?

### 6.2 Data Handling
- [ ] **Credential handling**: No hardcoded secrets, API keys, passwords?
- [ ] **Sensitive data logging**: Passwords, tokens not logged?
- [ ] **Error messages**: Don't leak sensitive info (file paths, internal state)?

### 6.3 Dependencies
- [ ] **Vulnerable packages**: Run `go list -m all | nancy` or `govulncheck`?
- [ ] **Least privilege**: Only necessary permissions for file/network ops?

---

## 7. Testing

### 7.1 Test Coverage
- [ ] **Coverage ≥80%**: Run `go test -coverprofile=coverage.out ./...`
  - Check: `go tool cover -func=coverage.out | tail -1`
- [ ] **Edge cases tested**: Empty, nil, zero, max values?
- [ ] **Error paths tested**: All error returns exercised?

### 7.2 Test Quality
- [ ] **Table-driven tests**: Multiple cases in one test function?
- [ ] **Test names descriptive**: `TestFoo_EmptyInput_ReturnsError` format?
- [ ] **Test independence**: Tests don't depend on order, share state?
- [ ] **Flaky tests**: No race conditions, timing dependencies?

### 7.3 Missing Tests (Flag if NO tests exist)
- [ ] **Complex algorithms**: Regex parsing, bracket matching, sorting?
- [ ] **Output formatting**: Terminal/JSON output generation?
- [ ] **Integration points**: File I/O, external dependencies?

---

## 8. Cross-Cutting Patterns (Check After Individual Review)

### 8.1 Consistency
- [ ] **Error wrapping style**: Consistent across codebase?
- [ ] **Naming conventions**: Variables, functions follow same patterns?
- [ ] **Code structure**: Similar problems solved similarly?

### 8.2 Hard-Coded Constants (Flag for extraction)
- [ ] **Magic numbers**: Literal numbers without named constants?
  - Example: `100`, `16`, `3`, `5` → Extract to `const`
- [ ] **String lists**: Tool names, parameter names in code vs config?
  - Example: `[]string{"file_path", "notebook_path", "path"}` → package var
- [ ] **Repeated patterns**: Maps created multiple times with same keys?

### 8.3 Missing Abstractions
- [ ] **Registry opportunities**: Hard-coded lists that could be registries?
  - Example: Validation checks hard-coded in Validate() function
- [ ] **Configuration**: Should hard-coded values be configurable?

---

## Severity Assignment Guide

Use this rubric to assign severity levels:

### Critical
- **Security vulnerabilities** (injection, auth bypass)
- **Data corruption risks** (lost writes, race conditions)
- **Crashes or panics** (nil pointer, out of bounds)
- **Broken core functionality** (feature doesn't work at all)
  - Example: VALIDATION-005 (order validation doesn't check order)

### High
- **Bugs affecting correctness** (wrong results, silent failures)
  - Example: QUERY-002 (returns 0 on error, ambiguous)
- **Significant performance issues** (>2x slower, O(n²) instead of O(n))
  - Example: QUERY-005 (O(n*m) nested iteration)
- **Resource leaks** (file handles, goroutines, memory)
- **Missing critical validation** (nil checks, bounds checks)
  - Example: QUERY-007 (missing entries != nil check)
- **Missing tests for critical code** (regex parsing, formatting)

### Medium
- **Code smells** (duplication, high complexity)
- **Readability issues** (unclear names, missing docs)
- **Non-idiomatic Go** (violates conventions)
- **Minor correctness issues** (edge cases not handled)
- **Moderate performance issues** (<2x slower)

### Low
- **Naming improvements** (could be clearer)
- **Comment enhancements** (add context)
- **Style inconsistencies** (minor deviations)
- **Optimization opportunities** (no current bottleneck)

---

## Issue Documentation Template

For each issue found, document:

```yaml
id: MODULE-NNN
severity: [critical|high|medium|low]
category: [correctness|performance|maintainability|readability|go_idioms|security|testing]
file: path/to/file.go
line: 123
function: functionName
description: One-line summary
explanation: |
  Detailed explanation of the issue.
  Why it's a problem and what could go wrong.
recommendation: Specific actionable fix
example: |
  # Optional code example showing the fix
  func fixed() {
      // Better approach
  }
```

---

## Post-Review Actions

After completing checklist:

1. **Categorize issues** by severity and category
2. **Count issues** (total, by severity, by category)
3. **Identify patterns** (repeated issues indicate systematic problems)
4. **Prioritize fixes** (critical → high → medium → low)
5. **Generate summary** (overview, critical issues, recommendations)
6. **Update taxonomy** (add new issue types if discovered)

---

## Automation Checks (Run Before Manual Review)

**ALWAYS run these automated checks first** (see automation-strategies.md):

```bash
# Format and imports
go fmt ./...
goimports -w .

# Linting
golangci-lint run ./...

# Security
gosec ./...

# Tests and coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

**Only do manual review AFTER** fixing automated issues. Saves time and focuses manual review on architecture/logic.

---

## References

- **Taxonomy**: knowledge/patterns/initial-issue-taxonomy.md
- **Automation**: knowledge/patterns/automation-strategies.md
- **Go Code Review Comments**: https://go.dev/wiki/CodeReviewComments
- **Effective Go**: https://go.dev/doc/effective_go

---

**Status**: Validated | **Iteration**: 2 | **Last Updated**: 2025-10-17
