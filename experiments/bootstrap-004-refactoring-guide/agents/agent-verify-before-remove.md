# Agent: Verify Before Remove

**Version**: 1.0
**Source**: Bootstrap-004, Pattern 1
**Success Rate**: Saved 2-4 hours debugging in meta-cc Iteration 1

---

## Role

Verify code is unused before removing it, preventing costly mistakes from removing actively-used code.

## When to Use

- Before removing any code (functions, files, blocks)
- When you believe code is unused or redundant
- Before major refactoring that involves deletions
- When cleanup recommendations suggest removal

## Input Schema

```yaml
target_code:
  file: string          # Required: Path to file containing code
  function: string      # Optional: Specific function name
  block: string         # Optional: Specific block description

scope: string           # Required: "file" | "package" | "project"
  # file: Check only within file
  # package: Check within package/module
  # project: Check entire codebase

verification_methods:
  static_analysis: boolean   # Default: true
  test_coverage: boolean     # Default: true
  reference_search: boolean  # Default: true
  runtime_check: boolean     # Default: false (only if applicable)
```

## Execution Process

### Step 1: Identify Candidate Code
```bash
# Note the exact location
File: internal/cache/old_cache.go
Function: CacheManagerOld
Reason: "Believed to be replaced by new cache implementation"
```

### Step 2: Choose Verification Scope
- **file**: Quick check, may miss cross-file references
- **package**: Balanced, catches most usage
- **project**: Comprehensive, recommended for public APIs

### Step 3: Run Static Analyzer
```bash
# Go
staticcheck ./path/to/scope/...
go vet ./path/to/scope/...

# Python
pylint path/to/module.py
mypy path/to/module.py

# JavaScript/TypeScript
eslint path/to/file.js
tsc --noEmit
```

**Expected Output**:
- No warnings = Likely unused
- Warnings about unused exports = Confirm unused
- No output about target = Potentially used (check further)

### Step 4: Check Test Coverage
```bash
# Go
go test -cover ./path/to/scope

# Python
pytest --cov=module

# JavaScript
jest --coverage
```

**Analysis**:
- 0% coverage + no tests = Likely unused BUT be cautious
- >0% coverage = DEFINITELY used (tests depend on it)

### Step 5: Search for References
```bash
# Using ripgrep (recommended)
rg "FunctionName" --type go

# Using grep
grep -r "FunctionName" . --include="*.go"

# IDE "Find Usages" (supplement, don't rely solely)
```

**Check for**:
- Direct calls: `FunctionName(`
- Interface implementations: `implements CacheManager`
- Type assertions: `x.(CacheManager)`
- Reflection usage: `reflect.TypeOf(CacheManager{})`

### Step 6: Verify Runtime Usage (If Applicable)
```bash
# Check logs for function invocations
grep "CacheManagerOld" /var/log/app.log

# Check API analytics
curl /api/analytics?function=CacheManagerOld

# Check reflection/dynamic usage (Go example)
rg "reflect.*CacheManagerOld"
```

### Step 7: Document Verification Results

**Template**:
```markdown
## Verification Report: [Function/File Name]

**Date**: [YYYY-MM-DD]
**Verifier**: [Agent/Person]

### Target Code
- File: internal/cache/old_cache.go
- Function: CacheManagerOld
- Reason: Replaced by new implementation

### Verification Results

#### Static Analysis
- Tool: staticcheck v0.4.6
- Command: `staticcheck ./internal/cache/...`
- Result: ✅ No warnings (not flagged as unused)
  OR
- Result: ⚠️ Warning U1000: CacheManagerOld is unused

#### Test Coverage
- Tool: go test -cover
- Coverage: 0% (no tests reference this function)
- Conclusion: No test dependencies

#### Reference Search
- Tool: ripgrep v13.0.0
- Pattern: `CacheManagerOld`
- Results:
  - internal/cache/old_cache.go:45 (definition)
  - internal/handlers/api.go:123 (usage found!)
- Conclusion: ❌ FUNCTION IS USED

#### Runtime Check
- Method: Log analysis (past 7 days)
- Result: 0 invocations logged
- Note: May not be conclusive (logs may be incomplete)

### Recommendation
**KEEP** - Function is actively used in internal/handlers/api.go

### Evidence
- Static analysis: Clean (but doesn't detect cross-package usage)
- Test coverage: 0% (but not definitive)
- Reference search: 1 active usage found ← DECISIVE
- Runtime: No logs (inconclusive)

**Decision**: Do NOT remove. Update mental model.
```

### Step 8: If Removing, Verify Tests Still Pass

```bash
# Save baseline
go test ./... > baseline.txt

# Make removal
# (delete code)

# Test after removal
go test ./... > after.txt

# Compare
diff baseline.txt after.txt

# Expected: No difference (all tests still pass)
```

## Output Schema

```yaml
verification_result:
  status: string  # "SAFE_TO_REMOVE" | "IN_USE" | "UNCERTAIN"
  confidence: number  # 0.0-1.0

evidence:
  static_analysis:
    tool: string
    warnings: number
    details: string

  test_coverage:
    percentage: number
    dependent_tests: number

  reference_search:
    tool: string
    matches: number
    locations: [string]

  runtime_check:
    method: string
    invocations: number
    period: string

recommendation:
  action: string  # "REMOVE" | "KEEP" | "INVESTIGATE_FURTHER"
  reason: string
  next_steps: [string]

report:
  markdown: string  # Full verification report
```

## Success Criteria

- ✅ All verification methods executed
- ✅ Evidence documented for each method
- ✅ Clear recommendation provided (REMOVE/KEEP/INVESTIGATE)
- ✅ If REMOVE: Tests pass after removal
- ✅ Confidence level ≥ 0.80 for SAFE_TO_REMOVE

## Example Execution

**Scenario**: Remove `validateToolInput` function from tools.go

```bash
# Input
target_code:
  file: "cmd/mcp-server/tools.go"
  function: "validateToolInput"
scope: "project"

# Step 1: Identify
# Function at line 145, claimed "unused validation logic"

# Step 2: Scope
# Using "project" (check entire codebase)

# Step 3: Static Analysis
$ staticcheck ./cmd/mcp-server/tools.go
# Result: No issues (code is NOT flagged as unused)

# Step 4: Test Coverage
$ go test -cover ./cmd/mcp-server
# coverage: 57.9% of statements
# Result: Tests exist and pass

# Step 5: Reference Search
$ rg "validateToolInput" --type go
cmd/mcp-server/tools.go:145: func validateToolInput(...)
cmd/mcp-server/handlers.go:67: err := validateToolInput(req.Input)
# Result: Found usage in handlers.go

# Conclusion
verification_result:
  status: "IN_USE"
  confidence: 1.0

recommendation:
  action: "KEEP"
  reason: "Function actively used in handlers.go line 67"

# Value: Prevented 2-4 hours debugging production errors
```

## Pitfalls and How to Avoid

### Pitfall 1: Running Analyzer on Wrong Scope
- ❌ Wrong: `staticcheck ./file.go` (misses cross-package references)
- ✅ Right: `staticcheck ./...` (checks whole project)

### Pitfall 2: Trusting IDE "Find Usages" Alone
- ❌ Wrong: IDE says 0 usages (may miss dynamic calls, reflection)
- ✅ Right: Use IDE + grep + static analyzer

### Pitfall 3: Not Checking Test Coverage
- ❌ Wrong: 0% coverage = unused (incorrect assumption)
- ✅ Right: 0% coverage = no test dependencies (but may be used in prod)

### Pitfall 4: Ignoring Warnings
- ❌ Wrong: "It's just a linter warning, I'll ignore it"
- ✅ Right: Investigate every warning, they indicate real issues

## Variations

### Variation 1: High-Confidence Removal
**Use when**: Staticcheck flags as unused + 0 references + 0% coverage
**Approach**: Remove, but still run full test suite
**Safety**: HIGH

### Variation 2: Medium-Confidence Removal
**Use when**: No staticcheck warning but 0 references + some coverage
**Approach**: Remove incrementally (one function at a time), test after each
**Safety**: MEDIUM

### Variation 3: Low-Confidence Removal
**Use when**: Uncertain about dynamic usage or external callers
**Approach**: Mark as @deprecated, monitor for 1-2 releases, then remove
**Safety**: HIGH

## Language-Specific Adaptations

### Go
```bash
staticcheck ./...
go vet ./...
go test -cover ./...
rg "FunctionName" --type go
```

### Python
```bash
pylint path/to/module.py
mypy path/to/module.py
vulture path/to/module.py  # Dead code detector
pytest --cov=module
rg "function_name" --type py
```

### JavaScript/TypeScript
```bash
eslint path/to/file.js
tsc --noEmit
depcheck  # Unused dependencies
jest --coverage
rg "functionName" --type js
```

### Java
```bash
# Use IDE inspections (IntelliJ)
# Or: SpotBugs, Checkstyle
mvn spotbugs:check
rg "methodName" --type java
```

## Usage Examples

### As Subagent
```bash
/subagent @experiments/bootstrap-004-refactoring-guide/agents/agent-verify-before-remove.md \
  target_code.file="internal/cache/old_cache.go" \
  target_code.function="CacheManagerOld" \
  scope="project"
```

### As Slash Command (if registered)
```bash
/verify-before-remove \
  file="internal/cache/old_cache.go" \
  function="CacheManagerOld" \
  scope="project"
```

### Programmatic (if API exists)
```python
result = verify_before_remove(
    target_code={
        "file": "internal/cache/old_cache.go",
        "function": "CacheManagerOld"
    },
    scope="project"
)

if result.recommendation.action == "REMOVE":
    # Safe to proceed with removal
    remove_code(target_code)
else:
    print(f"Cannot remove: {result.recommendation.reason}")
```

## Evidence from Bootstrap-004

**Source**: meta-cc Iteration 1
**Scenario**: Developer claimed "validation logic in tools.go is unused"
**Verification**: Prevented incorrect removal
**Value**: Saved 2-4 hours debugging

**Metrics**:
- False negative rate: 0% (caught all usage)
- Time to verify: ~5 minutes
- Time saved: 2-4 hours (prevented debugging)
- ROI: 24-48x

---

**Last Updated**: 2025-10-16
**Status**: Validated (meta-cc Iteration 1)
**Reusability**: Universal (any language with static analysis tools)
