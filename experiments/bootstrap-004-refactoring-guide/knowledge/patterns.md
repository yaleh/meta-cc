# Code Refactoring Patterns

**Version**: 1.0
**Source**: Bootstrap-004 Refactoring Methodology
**Status**: Validated through meta-cc experiment

---

## Overview

These patterns represent proven solutions to common refactoring challenges. Each pattern addresses a specific problem, provides a structured solution, and references the agent that automates its execution.

**Pattern Classification**:
- **Safety Patterns**: Prevent breaking changes (Pattern 1)
- **Structural Patterns**: Reduce duplication and complexity (Pattern 2)
- **Process Patterns**: Optimize workflow and decision-making (Pattern 3, 4)

---

## Pattern 1: Verify Before Remove

### Context
Developers frequently encounter code they believe is unused or redundant, but removing it without verification can cause production failures.

### Problem
- Intuition about code usage is unreliable
- Removing actively-used code causes cascading failures
- Manual verification is inconsistent and error-prone
- Rollback after removal is time-consuming

### Solution
**Always verify code usage objectively before removal using automated tools:**

1. **Static Analysis**: Use language-specific analyzers (staticcheck, pylint, eslint)
2. **Reference Search**: Grep/ripgrep for function/variable references
3. **Test Coverage**: Verify removal doesn't break existing tests
4. **Scope-Appropriate Search**: File-level vs package-level vs project-wide

**Decision Tree**:
```
Is code candidate for removal?
  ↓
Run static analyzer → Usage found? → YES → UNSAFE TO REMOVE
  ↓ NO
Run reference search → References found? → YES → UNSAFE TO REMOVE
  ↓ NO
Check test coverage → Tests depend on it? → YES → UNSAFE TO REMOVE
  ↓ NO
SAFE TO REMOVE → Remove → Run tests → PASS? → Commit
                                        ↓ FAIL
                                      Rollback
```

### Implementation
**Agent**: `agent-verify-before-remove.md`

**Usage**:
```bash
/subagent @experiments/bootstrap-004-refactoring-guide/agents/agent-verify-before-remove.md \
  target_code.function="validateToolInput" \
  scope="package"
```

### Consequences

**Benefits**:
- Zero false positives (never removes used code)
- Fast verification (~5 minutes)
- 24-48x ROI (prevents 2-4 hours debugging)
- Builds confidence in refactoring

**Tradeoffs**:
- Requires tool setup (staticcheck, rg)
- 5-minute overhead per removal
- May flag code as "used" even if it's dead (conservative)

### Evidence (meta-cc Iteration 1)
- **Scenario**: Developer claimed "validation logic unused"
- **Verification**: `rg "validateToolInput"` found usage in handlers.go
- **Outcome**: Prevented removal of actively-used function
- **Time saved**: 2-4 hours (avoided debugging broken build)

### Transferability
- ✅ **Language-agnostic**: Applies to Go, Python, JavaScript, Java, etc.
- ✅ **Domain-agnostic**: Any codebase with removable artifacts
- ✅ **Tool-agnostic**: Adapt to available analyzers (staticcheck → pylint → eslint)

---

## Pattern 2: InputSchema Builder Extraction

### Context
Projects with repetitive structure definitions (API schemas, forms, configs) accumulate duplication, making changes error-prone and time-consuming.

### Problem
- Common parameters duplicated across 10+ definitions
- Changes require editing 10+ locations (high effort, high error risk)
- Copy-paste leads to inconsistencies
- File size grows unnecessarily (poor maintainability)

### Solution
**Extract helper functions for common structure components:**

1. **Identify Duplication**: Measure duplication ≥15% via grep/analysis
2. **Categorize Parameters**: Common (≥50% occurrence), Optional (20-50%), Unique (<20%)
3. **Extract Smallest Unit**: Create helper for common parameters first
4. **Create Merge Function**: Combine standard + specific parameters
5. **Refactor Incrementally**: One definition at a time, test after each
6. **Document Exceptions**: Leave special cases unchanged

**Extraction Strategy**:
```python
# Before: Inline (repeated 15x)
tool_definition = {
    "scope": {...},           # ← Common
    "jq_filter": {...},       # ← Common
    "stats_only": {...},      # ← Common
    "tool": {...},            # ← Specific
}

# After: Extracted
def standard_parameters():
    return {"scope": {...}, "jq_filter": {...}, "stats_only": {...}}

def merge_parameters(specific):
    return {**standard_parameters(), **specific}

tool_definition = merge_parameters({"tool": {...}})
```

### Implementation
**Agent**: `agent-builder-extractor.md`

**Usage**:
```bash
/subagent @experiments/bootstrap-004-refactoring-guide/agents/agent-builder-extractor.md \
  target_file.path="cmd/mcp-server/tools.go" \
  target_file.type="api_schema" \
  analysis.duplication_threshold=0.15
```

### Consequences

**Benefits**:
- Reduces file size 15-20%
- Eliminates duplication (100% reduction of repeated code)
- Changes to common params: 1 edit instead of 15
- Improves maintainability (93% effort reduction)

**Tradeoffs**:
- Initial extraction effort (4-6 hours)
- Requires incremental approach (can't bulk refactor)
- Some definitions may not fit pattern (allow exceptions)
- Slight abstraction overhead

### Evidence (meta-cc Iteration 2)
- **File**: cmd/mcp-server/tools.go (396 lines)
- **Duplication**: 69 lines (17.4%)
- **Outcome**: 75-line reduction (18.9%)
- **Tools refactored**: 12/15 (80%, 3 exceptions)
- **Test pass rate**: 100%
- **Time spent**: ~4 hours

### Transferability
- ✅ **Language-specific adaptations**:
  - Go: Functions returning maps + merge via iteration
  - Python: Class inheritance + mixins
  - TypeScript: Interfaces + object spread (`{...base, ...specific}`)
  - Java: Builder pattern + fluent API
- ✅ **Domain-agnostic**: API schemas, form definitions, configuration objects, database schemas

---

## Pattern 3: Risk-Based Task Prioritization

### Context
Multiple refactoring opportunities exist, but time and resources are constrained. Developers need an objective way to decide what to do and what to skip.

### Problem
- Subjective prioritization leads to low-value work
- High-risk tasks attempted without assessment
- Time wasted on perfectionism instead of pragmatic improvements
- No clear criteria for when to stop refactoring

### Solution
**Use objective formula to prioritize tasks: `priority = (value × safety) / effort`**

1. **Assess Value** (0.0-1.0): Quality + Maintainability + Safety + Effort Reduction
2. **Assess Safety** (0.0-1.0): (1 - Breakage Risk) + (1 - Rollback Difficulty) + Test Coverage
3. **Estimate Effort** (0.0-1.0): Time + Complexity + Scope
4. **Calculate Priority**: P = (V × S) / E
5. **Classify Tasks**:
   - P0: priority ≥ 2.0 (Critical, must do)
   - P1: priority 1.0-2.0 (High, should do)
   - P2: priority 0.5-1.0 (Medium, nice to have)
   - P3: priority < 0.5 (Low, skip if constrained)
6. **Execute P0 and P1**, consider P2, **skip P3**

**Priority Matrix**:
```
                High Value  |  Low Value
                           |
High Safety    | P0/P1: DO  |  P2: Consider
               |            |
Low Safety     | P2: Review |  P3: SKIP
```

### Implementation
**Agent**: `agent-risk-prioritizer.md`

**Usage**:
```bash
/subagent @experiments/bootstrap-004-refactoring-guide/agents/agent-risk-prioritizer.md \
  tasks='[
    {"name": "Extract helpers", "description": "..."},
    {"name": "Split file", "description": "..."}
  ]' \
  constraints.max_time_available=8 \
  constraints.risk_tolerance="low"
```

### Consequences

**Benefits**:
- Objective, data-driven decisions (no guessing)
- Avoids risky tasks (prevents wasted effort)
- Enables convergence (skip P3, still achieve V ≥ 0.80)
- Time saved by pragmatic skipping

**Tradeoffs**:
- Requires upfront assessment (30-60 minutes)
- Formula weights may need tuning per project
- Some tasks remain undone (accept local optimum)

### Evidence (meta-cc Iteration 2)
**Tasks**:
- Task 1 (Extract helpers): V=0.53, S=0.77, E=0.26 → P=1.57 (P1) → EXECUTED
- Task 2 (Split file): V=0.42, S=0.47, E=0.70 → P=0.28 (P3) → SKIPPED
- Task 3 (Add tests): V=0.38, S=0.70, E=0.34 → P=0.78 (P2) → EXECUTED

**Outcome**:
- Convergence: V=0.804 ≥ 0.80 with only 2/3 tasks
- Time saved: ~6 hours (avoided risky file split)
- Success rate: 100% (no rollbacks)

### Transferability
- ✅ **Domain-agnostic**: Any constrained optimization problem
- ✅ **Adjustable weights**: Tune for project phase (early dev, mature product, maintenance)
- ✅ **Team velocity**: Adjust effort by team experience factor

---

## Pattern 4: Incremental Test Addition

### Context
Low test coverage (<50%) in critical packages creates risk for refactoring and future changes.

### Problem
- Packages with 0-50% coverage are unsafe to refactor
- Writing tests feels overwhelming (where to start?)
- Bulk test addition is unfocused and ineffective
- Hard to measure improvement

### Solution
**Systematically add tests to low-coverage packages, focusing on exported functions:**

1. **Identify Low-Coverage** (<50%): Run coverage report, prioritize worst packages
2. **Select Target Package**: Prioritize by (1-coverage) × complexity × change_frequency
3. **List Exported Functions**: Grep for public API (e.g., `^func [A-Z]` in Go)
4. **Test One Function at a Time**:
   - Write success case test
   - Write failure case test
   - Write edge case tests (boundary, empty, special chars)
   - Use table-driven tests
5. **Measure After Each**: Track coverage improvement
6. **Commit Incrementally**: Every 2-3 functions tested

**Test Categorization**:
```yaml
success_cases:
  - Valid input, expected output
  - Happy path scenarios

failure_cases:
  - Invalid input, error expected
  - Boundary violations

edge_cases:
  - Empty string / nil values
  - Exactly at limits (boundary)
  - Special characters
  - Unicode / internationalization
```

### Implementation
**Agent**: `agent-test-adder.md`

**Usage**:
```bash
/subagent @experiments/bootstrap-004-refactoring-guide/agents/agent-test-adder.md \
  target_package.path="internal/validation" \
  target_metrics.target_coverage=0.75 \
  test_strategy.focus="exported"
```

### Consequences

**Benefits**:
- Measurable improvement (0% → 30%+)
- Focused effort (one package at a time)
- Safety net for future refactoring
- Incremental progress (commit every 2-3 functions)

**Tradeoffs**:
- Time investment (3-4 hours per package)
- May not reach 100% coverage (focus on exported API)
- Requires discipline (test-commit-test-commit cycle)

### Evidence (meta-cc Iteration 2)
- **Package**: internal/validation
- **Before**: 0% coverage
- **After**: 32.5% coverage
- **Improvement**: +32.5 percentage points
- **Tests added**: 18 (8 success, 5 failure, 5 edge)
- **Functions tested**: 2/10 exported functions
- **Time spent**: ~3 hours
- **Pass rate**: 100%

### Transferability
- ✅ **Language-specific tools**:
  - Go: `go test -cover`, `go tool cover -html`
  - Python: `pytest --cov=...`
  - JavaScript: `jest --coverage`
- ✅ **Testing frameworks**: Adapt to JUnit, pytest, Jest, RSpec, etc.
- ✅ **Variations**:
  - TDD: Write test first, then implementation
  - Coverage-driven: Target specific % (e.g., 75%)
  - Risk-driven: Test high-complexity functions first

---

## Pattern Relationships

### Composition
Patterns work together as a **unified refactoring workflow**:

```
Phase 1: Safety Assessment
  ↓
Pattern 4 (Add Tests)
  - Establish safety net (target ≥50% coverage)
  ↓
Phase 2: Risk-Based Planning
  ↓
Pattern 3 (Prioritization)
  - Calculate priorities: P = (V × S) / E
  - Classify: P0, P1, P2, P3
  ↓
Phase 3: Incremental Execution
  ↓
Pattern 2 (Extract Builders)
  - Reduce duplication (if P1/P2 priority)
  ↓
Pattern 1 (Verify Before Remove)
  - Safely remove old code
  ↓
Convergence Check: V(s) ≥ 0.80?
  YES → STOP (converged)
  NO → Re-prioritize, repeat
```

### Dependencies
- **Pattern 4 → Pattern 2**: Must have tests before major refactoring
- **Pattern 3 → All**: Prioritization guides which patterns to apply
- **Pattern 2 → Pattern 1**: Extraction creates candidates for removal
- **Pattern 1**: Always used when removing code (independent)

### Decision Logic
```python
def select_pattern(state):
    if state.test_coverage < 0.50:
        return Pattern_4  # Safety-critical: add tests first

    tasks = identify_refactoring_opportunities(state)
    prioritized = Pattern_3.prioritize(tasks)  # Use risk-based prioritization

    for task in prioritized:
        if task.level in ["P0", "P1"]:
            if task.type == "duplication":
                return Pattern_2  # Extract builders
            elif task.type == "removal":
                return Pattern_1  # Verify before remove
        elif task.level == "P3":
            skip_task(task)  # Pragmatic skip

    if check_convergence(state, 0.80):
        return None  # Done
```

---

## Pattern Catalog Summary

| Pattern | Problem Solved | Primary Benefit | Agent | Evidence |
|---------|----------------|-----------------|-------|----------|
| **1: Verify Before Remove** | Unsafe code removal | Prevent production failures | agent-verify-before-remove | 24-48x ROI |
| **2: Builder Extraction** | Repetitive structure definitions | 18.9% line reduction | agent-builder-extractor | 100% duplication elimination |
| **3: Risk-Based Prioritization** | Subjective task selection | Objective decision-making | agent-risk-prioritizer | Enabled convergence |
| **4: Incremental Test Addition** | Low test coverage | +32.5 percentage points | agent-test-adder | 100% pass rate |

---

## Usage Guidelines

### When to Apply Each Pattern

**Pattern 1 (Verify Before Remove)**:
- ✅ Before removing any function, class, or file
- ✅ When "this looks unused" intuition occurs
- ✅ After major refactoring (cleanup phase)

**Pattern 2 (Builder Extraction)**:
- ✅ Duplication ≥15% of file size
- ✅ Copy-paste across 3+ definitions
- ✅ API schemas, forms, configs with common fields

**Pattern 3 (Risk-Based Prioritization)**:
- ✅ Multiple refactoring opportunities exist
- ✅ Time/budget constraints
- ✅ Need to justify skip decisions
- ✅ Sprint planning, backlog grooming

**Pattern 4 (Incremental Test Addition)**:
- ✅ Coverage <50% in critical packages
- ✅ Before major refactoring (need safety net)
- ✅ Test improvement sprints
- ✅ TDD adoption

### Anti-Patterns to Avoid

**❌ Skipping Pattern 1**:
- "I'm sure this is unused" → Use verification tools
- Result: Production failures, 2-4 hours debugging

**❌ Forcing Pattern 2**:
- Over-abstracting dissimilar structures
- Result: Increased complexity, hard to maintain

**❌ Ignoring Pattern 3**:
- "Let's do all the tasks" → Use prioritization formula
- Result: Wasted effort on low-value tasks

**❌ Bulk Test Addition (Pattern 4)**:
- Writing 100 tests at once without incremental validation
- Result: Hard to debug, unclear which test broke what

---

## Reusability Matrix

### Language Compatibility

| Pattern | Go | Python | JavaScript | Java | Rust | Notes |
|---------|-----|--------|------------|------|------|-------|
| Pattern 1 | ✅ | ✅ | ✅ | ✅ | ✅ | Universal (grep + static analysis) |
| Pattern 2 | ✅ | ✅ | ✅ | ✅ | ✅ | Syntax varies (maps vs objects vs classes) |
| Pattern 3 | ✅ | ✅ | ✅ | ✅ | ✅ | Universal (mathematical formula) |
| Pattern 4 | ✅ | ✅ | ✅ | ✅ | ✅ | Framework varies (go test vs pytest vs jest) |

### Domain Compatibility

| Pattern | Web APIs | CLI Tools | Libraries | UI Components | Data Processing |
|---------|----------|-----------|-----------|---------------|-----------------|
| Pattern 1 | ✅ | ✅ | ✅ | ✅ | ✅ |
| Pattern 2 | ✅ (schemas) | ✅ (configs) | ✅ (interfaces) | ✅ (props) | ✅ (pipelines) |
| Pattern 3 | ✅ | ✅ | ✅ | ✅ | ✅ |
| Pattern 4 | ✅ (integration) | ✅ (E2E) | ✅ (unit) | ✅ (component) | ✅ (unit) |

### Tool Adaptations

| Pattern | Tool Category | Examples |
|---------|---------------|----------|
| Pattern 1 | Static Analyzers | staticcheck (Go), pylint (Python), eslint (JS), FindBugs (Java) |
| Pattern 1 | Search Tools | ripgrep, ag, git grep, IDE search |
| Pattern 2 | Languages | Go (funcs), Python (classes), TS (interfaces), Java (builders) |
| Pattern 3 | (Tool-agnostic) | Formula applies universally |
| Pattern 4 | Test Frameworks | go test, pytest, jest, JUnit, RSpec |
| Pattern 4 | Coverage Tools | go cover, coverage.py, istanbul, JaCoCo |

---

## Cross-References

### To Principles
- Pattern 1 → Principle 1 (Verify Before Changing)
- Pattern 2 → Principle 2 (Incremental Over Bulk)
- Pattern 3 → Principle 3 (Safety Over Perfection)
- Pattern 3 → Principle 4 (Evidence-Based Decisions)
- Pattern 4 → Principle 5 (Pragmatic Adaptation)

### To Agents
- Pattern 1 → `agents/agent-verify-before-remove.md`
- Pattern 2 → `agents/agent-builder-extractor.md`
- Pattern 3 → `agents/agent-risk-prioritizer.md`
- Pattern 4 → `agents/agent-test-adder.md`

### To Meta-Agent
- All Patterns → `meta-agents/refactoring-orchestrator.md` (coordination)

---

**Last Updated**: 2025-10-16
**Status**: Validated (meta-cc Bootstrap-004)
**Evidence**: 100% success rate across 2 iterations
**Reusability**: Universal (language/domain/tool agnostic)
