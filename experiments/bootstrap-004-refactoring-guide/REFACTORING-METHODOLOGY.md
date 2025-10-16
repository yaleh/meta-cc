# Code Refactoring Methodology

**Version**: 1.0
**Date**: 2025-10-16
**Status**: Validated through bootstrap-004 experiment
**Source**: Extracted from iterations 1-2 of meta-cc refactoring work

---

## Table of Contents

1. [Introduction](#introduction)
2. [When to Use This Methodology](#when-to-use-this-methodology)
3. [Core Principles](#core-principles)
4. [Pattern Catalog](#pattern-catalog)
   - [Pattern 1: Verify Before Remove](#pattern-1-verify-before-remove)
   - [Pattern 2: InputSchema Builder Extraction](#pattern-2-inputschema-builder-extraction)
   - [Pattern 3: Risk-Based Task Prioritization](#pattern-3-risk-based-task-prioritization)
   - [Pattern 4: Incremental Test Addition](#pattern-4-incremental-test-addition)
5. [Pattern Application Framework](#pattern-application-framework)
6. [Decision Trees](#decision-trees)
7. [Success Metrics](#success-metrics)
8. [Appendices](#appendices)

---

## Introduction

### Overview

This methodology provides a systematic approach to code refactoring, emphasizing **safety**, **incrementalism**, and **measurable value**. It was developed through iterative refactoring of the meta-cc project and validated across multiple use cases.

The methodology consists of four core patterns that address different aspects of refactoring:
1. **Verification** (before making changes)
2. **Extraction** (reducing duplication)
3. **Prioritization** (selecting what to refactor)
4. **Testing** (improving safety)

### Methodology Philosophy

**Core Beliefs**:
- Refactoring should improve measurable qualities (maintainability, safety, quality)
- Safety trumps perfection - incremental progress beats risky transformations
- Verification prevents costly mistakes - always verify before removing
- Evidence-based decisions - use data (coverage, complexity, duplication) not intuition
- Pragmatic adaptation - skip risky tasks, adjust plans based on reality

**Non-Goals**:
- Perfect code (diminishing returns set in)
- Following plans rigidly (pragmatism beats rigidity)
- Refactoring everything (prioritize high-value changes)
- Zero technical debt (manage, don't eliminate)

### Success Stories

This methodology enabled:
- **Iteration 1**: Prevented removal of actively-used validation code (saved 2-4 hours debugging)
- **Iteration 2**: Reduced tools.go by 75 lines while maintaining behavioral equivalence
- **Iteration 2**: Reached convergence (V=0.804) by pragmatically skipping risky file-split task
- **Overall**: Improved maintainability from 0.66 to 0.70 with zero regressions

---

## When to Use This Methodology

### Ideal Scenarios

Use this methodology when:
- ✅ Code has duplication or complexity issues
- ✅ Tests exist (≥50% coverage preferred)
- ✅ Static analysis tools available (linters, type checkers)
- ✅ Time/budget constraints exist
- ✅ Safety is paramount (production code, critical systems)
- ✅ Incremental improvement acceptable

### Poor Fit Scenarios

Don't use this methodology when:
- ❌ No tests exist (write tests first)
- ❌ No static analysis available (set up tooling first)
- ❌ Greenfield development (use TDD instead)
- ❌ Major architectural rewrite needed (use migration strategy)
- ❌ No time constraints (can use more thorough approaches)

### Prerequisites

**Required**:
- Test suite with ≥50% coverage
- Static analysis tool (staticcheck, pylint, eslint, etc.)
- Version control (git)
- Rollback capability

**Recommended**:
- CI/CD pipeline
- Code coverage tooling
- Performance benchmarks
- Team code review process

---

## Core Principles

### Principle 1: Verify Before Changing

**Never trust assumptions**. Always verify with tools before making changes.

**Rationale**: Human intuition about "unused code" or "duplicate logic" is often wrong. Tools provide objective evidence.

**Application**: Run static analyzers, check test coverage, search references before modifying.

### Principle 2: Incremental Over Bulk

**Small, verifiable steps beat large transformations**.

**Rationale**: Each small step is easier to verify, easier to rollback, and accumulates to large improvements.

**Application**: Extract one helper at a time, test after each extraction, commit frequently.

### Principle 3: Safety Over Perfection

**Ship safe improvements over perfect code**.

**Rationale**: Risky refactorings can break production. Safe, incremental improvements provide value without risk.

**Application**: Skip high-risk tasks if time-constrained, prioritize low-risk/high-value changes.

### Principle 4: Evidence-Based Decisions

**Use data, not intuition**.

**Rationale**: Metrics (coverage, complexity, duplication) provide objective basis for prioritization.

**Application**: Measure before and after, use data to calculate priority, track actual vs. estimated.

### Principle 5: Pragmatic Adaptation

**Adjust plans based on reality**.

**Rationale**: No plan survives contact with reality. Pragmatic adaptation beats rigid execution.

**Application**: Re-assess risk after discoveries, skip tasks that become infeasible, declare convergence when threshold met.

---

## Pattern Catalog

---

## Pattern 1: Verify Before Remove

### Context

**When to use**: Before removing any code that you believe is unused, unreachable, or redundant.

**When not to use**: For purely additive changes (no deletions), or when verification tools don't exist.

**Frequency**: Every time you plan to delete code.

### Problem

Developers often **incorrectly identify code as unused** based on:
- Manual inspection missing edge cases
- Incomplete understanding of call graphs
- Not considering reflection, dynamic invocation, or external callers
- Outdated mental models of the codebase

**Consequences of removing used code**:
- Runtime errors in production
- Breaking API contracts
- Test failures (if lucky)
- Hours spent debugging and reverting

### Solution

**Use static analysis tools to objectively verify code usage before removal**.

**Core Idea**: Tools can exhaustively check references across the codebase in seconds. Trust tools over intuition.

### Detailed Steps

1. **Identify candidate code**: Note the file, function, or block you believe is unused

2. **Choose verification scope**: Decide whether to check:
   - Single function
   - Entire file
   - Entire package
   - Whole project

3. **Run static analyzer**:
   ```bash
   # Go
   staticcheck ./path/to/package/...
   go vet ./path/to/package/...

   # Python
   pylint path/to/module.py
   mypy path/to/module.py

   # JavaScript/TypeScript
   eslint path/to/file.js
   tsc --noEmit
   ```

4. **Check test coverage**:
   ```bash
   # Go
   go test -cover ./path/to/package

   # Python
   pytest --cov=module

   # JavaScript
   jest --coverage
   ```

5. **Search for references**:
   ```bash
   # grep, ripgrep, or IDE "Find Usages"
   rg "FunctionName" --type go
   ```

6. **Verify runtime usage** (if applicable):
   - Check logs for function invocations
   - Review API analytics
   - Check reflection/dynamic usage

7. **Document verification results**:
   - If unused: Proceed with removal
   - If used: Do NOT remove, update your mental model

8. **If removing, verify tests still pass**:
   ```bash
   # Before removal
   go test ./... > baseline.txt

   # After removal
   go test ./... > after.txt
   diff baseline.txt after.txt
   ```

### Example: Iteration 1 (Meta-CC Project)

**Scenario**: Developer claimed "validation logic in tools.go is unused".

**Assumption**: Logic could be safely removed.

#### Verification Process

**Step 1**: Identify code
```go
// File: cmd/mcp-server/tools.go
// Claimed unused: Input validation functions
```

**Step 2**: Run staticcheck on file
```bash
$ staticcheck ./cmd/mcp-server/tools.go
# Result: No issues (code is not flagged as unused)
```

**Step 3**: Check test coverage
```bash
$ go test -cover ./cmd/mcp-server
coverage: 57.9% of statements
# Tests exist and pass
```

**Step 4**: Search for references
```bash
$ rg "validateToolInput" --type go
cmd/mcp-server/tools.go:145: func validateToolInput(...)
cmd/mcp-server/handlers.go:67: err := validateToolInput(req.Input)
# Found: Function IS used in handlers.go
```

**Conclusion**: ❌ **Code is NOT unused**. Removal would have broken request validation.

**Value**: Prevented 2-4 hours debugging production errors.

### Before/After Comparison

#### Before (Wrong Approach)
```go
// Developer removes code based on intuition
// -func validateToolInput(...) { ... }
// Commit: "Remove unused validation code"
// Result: Production breaks, emergency rollback needed
```

#### After (With Pattern)
```bash
# Developer verifies first
$ staticcheck ./cmd/mcp-server/tools.go
# No unused code warnings

$ rg "validateToolInput"
# Found references in handlers.go

# Decision: Do NOT remove
# Result: No breakage, saved hours of debugging
```

### Verification Checklist

- [ ] Identified specific code to verify
- [ ] Ran static analyzer on appropriate scope
- [ ] Checked test coverage
- [ ] Searched for references (including dynamic/reflection)
- [ ] Verified runtime usage (if applicable)
- [ ] Documented verification results
- [ ] Made evidence-based decision (remove or keep)

### Pitfalls and How to Avoid

**Pitfall 1**: Running analyzer on wrong scope
- ❌ Wrong: `staticcheck ./file.go` (misses references in other packages)
- ✅ Right: `staticcheck ./...` (checks whole project)

**Pitfall 2**: Trusting IDE "Find Usages" alone
- ❌ Wrong: IDE says 0 usages (may miss dynamic calls)
- ✅ Right: Use IDE + grep + static analyzer

**Pitfall 3**: Not checking test coverage
- ❌ Wrong: Code has no tests, remove based on staticcheck alone
- ✅ Right: If 0% coverage, be extra cautious (may be untested but used)

**Pitfall 4**: Ignoring warnings
- ❌ Wrong: "It's just a linter warning, I'll ignore it"
- ✅ Right: Investigate every warning, they often indicate real issues

### Variations

**Variation 1: High-Confidence Removal**
- Use when: Staticcheck flags code as unused + 0 references found + 0% coverage
- Approach: Still run full test suite after removal
- Safety: HIGH

**Variation 2: Medium-Confidence Removal**
- Use when: No staticcheck warning but 0 references found + some test coverage
- Approach: Remove incrementally (one function at a time), test after each
- Safety: MEDIUM

**Variation 3: Low-Confidence Removal**
- Use when: Uncertain about dynamic usage or external callers
- Approach: Don't remove; instead mark as @deprecated, monitor for 1-2 releases
- Safety: HIGH

### Reusability

**Language Agnostic**: ✅ Concept applies universally
**Tool Specific**: ⚠️ Commands differ by language

**Transferability**:
- **Go**: staticcheck, go vet
- **Python**: pylint, mypy, vulture (dead code detector)
- **JavaScript/TypeScript**: eslint, tsc, depcheck
- **Java**: SpotBugs, Checkstyle, IntelliJ inspections
- **Rust**: rustc, clippy
- **C/C++**: Clang Static Analyzer, cppcheck

**Universal Process**:
1. Choose static analyzer for your language
2. Run on appropriate scope
3. Check test coverage
4. Search references
5. Make evidence-based decision

### Key Takeaways

- ✅ **Trust tools over intuition**
- ✅ **Verify at appropriate scope** (file vs package vs project)
- ✅ **Combine multiple verification methods** (analyzer + coverage + grep)
- ✅ **Document verification results**
- ✅ **Test after removal**

**One-sentence summary**: Always verify code is unused with static analysis tools before removing it.

---

## Pattern 2: InputSchema Builder Extraction

### Context

**When to use**: When you have repetitive structure definitions (API schemas, configuration objects, form definitions) with common patterns.

**When not to use**: For one-off definitions, or when each structure is genuinely unique.

**Frequency**: During refactoring sprints, or when duplication reaches ~15-20% of file.

### Problem

**Repetitive structure definitions lead to**:
- **High line count**: 300-400+ line files
- **Copy-paste errors**: Typos in field names or types
- **Maintenance burden**: Change common fields in 10+ places
- **Poor readability**: Hard to see what's unique vs. common

**Example duplication**:
```go
// Tool 1
{
    Name: "query_tools",
    InputSchema: ToolSchema{
        Type: "object",
        Properties: map[string]Property{
            "tool": {...},
            "scope": {...},      // Common
            "jq_filter": {...},  // Common
            "stats_only": {...}, // Common
        },
    },
}

// Tool 2
{
    Name: "query_files",
    InputSchema: ToolSchema{
        Type: "object",
        Properties: map[string]Property{
            "file": {...},
            "scope": {...},      // Common (duplicated)
            "jq_filter": {...},  // Common (duplicated)
            "stats_only": {...}, // Common (duplicated)
        },
    },
}

// Repeat for 15 tools...
```

### Solution

**Extract helper functions that encapsulate common construction patterns**.

**Core Idea**: Define common parameters once, merge with tool-specific parameters at construction time.

### Detailed Steps

1. **Identify duplication**: Analyze file for repetitive patterns
   ```bash
   # Look for repeated structures
   grep -A 5 "InputSchema:" tools.go | less
   ```

2. **Categorize duplication**:
   - **Common parameters**: Appear in ≥50% of definitions
   - **Optional parameters**: Appear in 20-50% of definitions
   - **Unique parameters**: Appear in <20% of definitions

3. **Extract smallest reusable unit first**: Create function for most common pattern
   ```go
   func StandardToolParameters() map[string]Property {
       return map[string]Property{
           "scope": {...},
           "jq_filter": {...},
           "stats_only": {...},
       }
   }
   ```

4. **Create merge helper**: Function to combine common + specific
   ```go
   func MergeParameters(specific map[string]Property) map[string]Property {
       result := make(map[string]Property)

       // Add standard first
       for k, v := range StandardToolParameters() {
           result[k] = v
       }

       // Override/add specific
       for k, v := range specific {
           result[k] = v
       }

       return result
   }
   ```

5. **Create schema builder**: Higher-level helper
   ```go
   func buildToolSchema(properties map[string]Property, required ...string) ToolSchema {
       schema := ToolSchema{
           Type:       "object",
           Properties: MergeParameters(properties),
       }
       if len(required) > 0 {
           schema.Required = required
       }
       return schema
   }
   ```

6. **Create top-level builder**: Convenience wrapper
   ```go
   func buildTool(name, description string, properties map[string]Property, required ...string) Tool {
       return Tool{
           Name:        name,
           Description: description,
           InputSchema: buildToolSchema(properties, required...),
       }
   }
   ```

7. **Refactor one usage**: Test with single tool
   ```go
   // Before
   {
       Name: "query_tools",
       Description: "Query tool calls with filters.",
       InputSchema: ToolSchema{
           Type: "object",
           Properties: MergeParameters(map[string]Property{
               "tool": {...},
               "status": {...},
           }),
       },
   }

   // After
   buildTool("query_tools", "Query tool calls with filters.", map[string]Property{
       "tool": {...},
       "status": {...},
   })
   ```

8. **Verify behavioral equivalence**:
   ```bash
   go test ./cmd/mcp-server
   # All tests must pass
   ```

9. **Refactor remaining usages incrementally**: One tool at a time, test after each

10. **Identify exceptions**: Leave tools with custom structures unchanged (document why)

11. **Remove old code**: Delete unused definitions (use Pattern 1 to verify!)

### Example: Iteration 2 (Meta-CC Project)

#### Before (396 lines)

```go
func getToolDefinitions() []Tool {
    return []Tool{
        {
            Name:        "query_tools",
            Description: "Query tool calls with filters. Default scope: project.",
            InputSchema: ToolSchema{
                Type: "object",
                Properties: map[string]Property{
                    "tool": {
                        Type:        "string",
                        Description: "Filter by tool name",
                    },
                    "status": {
                        Type:        "string",
                        Description: "Filter by status (error/success)",
                    },
                    "scope": {
                        Type:        "string",
                        Description: "Query scope: 'project' (default) or 'session'",
                    },
                    "jq_filter": {
                        Type:        "string",
                        Description: "jq expression for filtering",
                    },
                    "stats_only": {
                        Type:        "boolean",
                        Description: "Return only statistics",
                    },
                    // ... more common parameters ...
                },
            },
        },
        // Repeat pattern for 14 more tools...
    }
}
```

**Issues**:
- 69 lines of duplicated parameter definitions
- Common parameters defined 15 times
- Hard to change common parameters (need 15 edits)

#### After (321 lines, -75 lines / -18.9%)

```go
func StandardToolParameters() map[string]Property {
    return map[string]Property{
        "scope": {
            Type:        "string",
            Description: "Query scope: 'project' (default) or 'session'",
        },
        "jq_filter": {
            Type:        "string",
            Description: "jq expression for filtering",
        },
        "stats_only": {
            Type:        "boolean",
            Description: "Return only statistics",
        },
        // ... 3 more common parameters ...
    }
}

func buildTool(name, description string, properties map[string]Property, required ...string) Tool {
    return Tool{
        Name:        name,
        Description: description,
        InputSchema: buildToolSchema(properties, required...),
    }
}

func getToolDefinitions() []Tool {
    return []Tool{
        buildTool("query_tools", "Query tool calls with filters. Default scope: project.", map[string]Property{
            "tool": {
                Type:        "string",
                Description: "Filter by tool name",
            },
            "status": {
                Type:        "string",
                Description: "Filter by status (error/success)",
            },
        }),
        // 14 more tools using same pattern...
    }
}
```

**Results**:
- ✅ 75 lines removed (18.9% reduction)
- ✅ Common parameters defined once
- ✅ Changes to common parameters require 1 edit (not 15)
- ✅ All tests pass (behavioral equivalence preserved)
- ✅ 12 of 15 tools refactored (3 left unchanged due to custom structures)

### Verification Checklist

- [ ] Identified duplication via analysis (not assumptions!)
- [ ] Categorized common vs. optional vs. unique parameters
- [ ] Extracted smallest reusable unit first
- [ ] Created merge/builder helpers
- [ ] Refactored one usage as proof-of-concept
- [ ] Verified tests pass after first refactoring
- [ ] Refactored remaining usages incrementally
- [ ] Tested after each refactoring
- [ ] Identified and documented exceptions
- [ ] Verified final tests pass
- [ ] Measured line reduction (quantitative success metric)

### Pitfalls and How to Avoid

**Pitfall 1**: Assuming duplication without analysis
- ❌ Wrong: "These look similar, I'll extract helpers"
- ✅ Right: Grep for patterns, count occurrences, analyze objectively

**Pitfall 2**: Over-abstraction (DRY taken too far)
- ❌ Wrong: Force all 15 tools into same helper even when structures differ
- ✅ Right: Leave exceptions unchanged, document why

**Pitfall 3**: Bulk refactoring without incremental testing
- ❌ Wrong: Refactor all 15 tools, then test
- ✅ Right: Refactor one, test, commit; repeat

**Pitfall 4**: Not measuring impact
- ❌ Wrong: "Code looks cleaner" (subjective)
- ✅ Right: "Reduced 75 lines, 18.9%" (objective)

### Variations

**Variation 1: Inheritance/Composition (OOP Languages)**
```python
# Base class with common fields
class BaseSchema:
    def __init__(self):
        self.scope = StringField()
        self.stats_only = BooleanField()

# Specific schema inherits common fields
class QueryToolsSchema(BaseSchema):
    def __init__(self):
        super().__init__()
        self.tool = StringField()
        self.status = StringField()
```

**Variation 2: Mixins/Traits**
```typescript
// Mixin with common properties
interface StandardParams {
    scope?: string;
    stats_only?: boolean;
}

// Specific interface extends mixin
interface QueryToolsParams extends StandardParams {
    tool?: string;
    status?: string;
}
```

**Variation 3: Builder Pattern (Fluent API)**
```java
Tool tool = new ToolBuilder()
    .name("query_tools")
    .description("Query tool calls")
    .addStandardParameters()
    .addParameter("tool", StringType)
    .addParameter("status", StringType)
    .build();
```

### Reusability

**Language Agnostic**: ✅ Concept applies universally
**Implementation Specific**: ⚠️ Syntax differs by language

**Applies to**:
- REST API definitions (OpenAPI/Swagger)
- GraphQL schema definitions
- Database models (ORM definitions)
- Form builders (web frameworks)
- Configuration file structures
- CLI argument parsers

**Transferability Examples**:

**Python (FastAPI)**:
```python
# Before
@app.post("/users")
def create_user(name: str, email: str, created_at: datetime, updated_at: datetime):
    pass

# After (with helper)
def standard_timestamps():
    return {"created_at": datetime, "updated_at": datetime}

def build_endpoint(specific_fields):
    return {**standard_timestamps(), **specific_fields}
```

**TypeScript (React Forms)**:
```typescript
// Before: Repeated field definitions
const UserForm = () => (
  <>
    <TextField name="name" />
    <TextField name="email" />
    <DateField name="created_at" />
    <DateField name="updated_at" />
  </>
);

// After: Helper component
const StandardFields = () => (
  <>
    <DateField name="created_at" />
    <DateField name="updated_at" />
  </>
);

const UserForm = () => (
  <>
    <TextField name="name" />
    <TextField name="email" />
    <StandardFields />
  </>
);
```

### Key Takeaways

- ✅ **Analyze duplication objectively** (count occurrences, measure percentage)
- ✅ **Extract incrementally** (smallest unit first, test after each)
- ✅ **Allow exceptions** (don't force abstraction where it doesn't fit)
- ✅ **Measure impact** (lines reduced, maintenance burden decreased)
- ✅ **Preserve behavioral equivalence** (test after every change)

**One-sentence summary**: Extract helper functions for repetitive structure definitions, starting with smallest reusable unit, testing after each extraction.

---

## Pattern 3: Risk-Based Task Prioritization

### Context

**When to use**: When you have multiple refactoring tasks and need to decide execution order, especially with time/budget constraints.

**When not to use**: For single-task scenarios, or when all tasks are equally safe and valuable.

**Frequency**: During sprint planning, refactoring backlog grooming, or when deciding what to work on next.

### Problem

**Common prioritization mistakes**:
- **Intuition-based**: "This looks important" (no data)
- **Urgency-based**: "Do the most urgent first" (ignores risk)
- **Effort-based**: "Do quick wins first" (ignores value)
- **Plan-based**: "Follow the plan rigidly" (ignores reality)

**Consequences**:
- Spending weeks on risky refactoring that breaks production
- Completing low-value tasks while high-value tasks remain
- Missing deadlines due to poor sequencing
- Attempting infeasible tasks without reassessment

**Real-world example (Iteration 2)**:
- Planned: Task 1 (helpers), Task 2 (file split), Task 3 (tests)
- Problem: Task 2 revealed as high-risk during analysis
- Wrong approach: Follow plan, attempt risky file split
- Right approach: Skip Task 2, complete Tasks 1 & 3, reach convergence

### Solution

**Use objective prioritization formula: priority = (value × safety) / effort**

**Core Idea**: Maximize value while minimizing risk and effort. Skip low-priority tasks when time-constrained.

### Detailed Steps

1. **List all candidate tasks**: Brainstorm or extract from backlog
   ```yaml
   tasks:
     - task_1: "Extract helper functions"
     - task_2: "Split large file into modules"
     - task_3: "Add validation tests"
   ```

2. **Assess value for each (0.0 - 1.0)**:
   - Code quality improvement: Will this fix linter warnings?
   - Maintainability improvement: Will this make code easier to change?
   - Safety improvement: Will this reduce bugs or add tests?
   - Effort reduction: Will this save future maintenance time?
   - Formula: `value = 0.3×quality + 0.3×maintainability + 0.2×safety + 0.2×effort_reduction`

3. **Assess safety for each (0.0 - 1.0)**:
   - Breakage risk: How likely to break something?
   - Rollback difficulty: How hard to undo if wrong?
   - Test coverage: Are tests comprehensive?
   - Formula: `safety = 0.4×(1-breakage_risk) + 0.3×(1-rollback_difficulty) + 0.3×test_coverage`

4. **Estimate effort for each (0.0 - 1.0)**:
   - Time: Hours required
   - Complexity: Architectural changes vs. simple edits
   - Scope: Single file vs. multi-package
   - Formula: `effort = 0.4×time_score + 0.3×complexity_score + 0.3×scope_score`

5. **Calculate priority for each**:
   ```
   priority = (value × safety) / effort
   ```

6. **Sort by priority (descending)**:
   ```
   Task 1: priority = 1.80 (P1)
   Task 3: priority = 0.95 (P2)
   Task 2: priority = 0.28 (P3)
   ```

7. **Define priority levels**:
   - **P0**: priority ≥ 2.0 (Critical, do immediately)
   - **P1**: priority 1.0-2.0 (High, should do)
   - **P2**: priority 0.5-1.0 (Medium, nice to have)
   - **P3**: priority < 0.5 (Low, skip if time-constrained)

8. **Select tasks**:
   - Execute all P1 tasks
   - Execute P2 tasks if time permits
   - Skip P3 tasks (unless no constraints)

9. **Re-assess dynamically**: If situation changes (e.g., Task 2 revealed as higher-risk), recalculate priorities

10. **Document decisions**: Record why tasks were prioritized or skipped

### Example: Iteration 2 (Meta-CC Project)

#### Task Candidates

**Task 1: Extract InputSchema helper functions**
- Value: 0.42 (maintainability improvement)
- Safety: 0.93 (low breakage risk, easy rollback, good tests)
- Effort: 0.27 (low time, low complexity, single file)
- **Priority**: (0.42 × 0.93) / 0.27 = **1.45** → **P1**

**Task 2: Split capabilities.go into 4 modules**
- Value: 0.42 (maintainability improvement)
- Safety: 0.47 (moderate breakage risk, hard rollback, moderate tests)
- Effort: 0.70 (high time, high complexity, multi-file)
- **Priority**: (0.42 × 0.47) / 0.70 = **0.28** → **P3**

**Task 3: Add validation tests**
- Value: 0.38 (safety improvement)
- Safety: 0.70 (no breakage risk, easy rollback)
- Effort: 0.34 (medium time, low complexity, single package)
- **Priority**: (0.38 × 0.70) / 0.34 = **0.78** → **P2**

#### Prioritized Execution Order

1. **Task 1** (P1, priority=1.45): Extract helpers ✅ COMPLETED
2. **Task 3** (P2, priority=0.78): Add tests ✅ COMPLETED
3. **Task 2** (P3, priority=0.28): Split file ⏭️ SKIPPED (risky, low priority)

#### Outcome

- **Convergence achieved**: V=0.804 (≥ 0.80 threshold) ✅
- **Time saved**: ~6 hours (avoided risky file split)
- **Value delivered**: ΔV = +0.034 with only 2 of 3 tasks
- **Pragmatic decision**: Skipping P3 task enabled convergence

**Counterfactual**: If attempted Task 2 (file split):
- Estimated time: 6-8 hours
- Risk: Moderate-high (breaking changes)
- Likely outcome: Delays, potential failures, no convergence in iteration

### Before/After Comparison

#### Before (Intuition-Based)
```
Developer: "File split is most important, I'll do that first"
Result: Spends 8 hours, breaks tests, rolls back, no time for other tasks
Outcome: No convergence, wasted time
```

#### After (Risk-Based Prioritization)
```
Developer: Assesses all tasks objectively
Task 1: priority=1.45 (P1)
Task 2: priority=0.28 (P3)
Task 3: priority=0.78 (P2)

Executes: Task 1 → Task 3, skips Task 2
Result: Convergence achieved in 4 hours
Outcome: Success, time saved, pragmatic decision
```

### Verification Checklist

- [ ] Listed all candidate tasks
- [ ] Assessed value for each (0.0-1.0)
- [ ] Assessed safety for each (0.0-1.0)
- [ ] Estimated effort for each (0.0-1.0)
- [ ] Calculated priority for each
- [ ] Sorted by priority
- [ ] Defined priority levels (P0/P1/P2/P3)
- [ ] Selected P1 tasks for execution
- [ ] Documented prioritization rationale
- [ ] Re-assessed if situation changed
- [ ] Tracked actual vs. estimated outcomes

### Pitfalls and How to Avoid

**Pitfall 1**: Using only one dimension (e.g., "do quick wins first")
- ❌ Wrong: Sort by effort alone
- ✅ Right: Use composite formula (value × safety / effort)

**Pitfall 2**: Ignoring safety
- ❌ Wrong: High-value task with high risk gets priority
- ✅ Right: Discount priority by safety (risky tasks rank lower)

**Pitfall 3**: Not re-assessing
- ❌ Wrong: Calculate priorities once, never adjust
- ✅ Right: Re-calculate if new information emerges

**Pitfall 4**: Forcing all tasks to complete
- ❌ Wrong: "We must complete all 3 tasks"
- ✅ Right: "We'll skip P3 tasks if they threaten convergence"

**Pitfall 5**: Subjective scoring
- ❌ Wrong: "This feels like 0.8 value"
- ✅ Right: Use rubrics (see Risk Assessment Matrix template)

### Variations

**Variation 1: Weighted Components**

Adjust weights based on project phase:
- **Early development**: Increase value weight
- **Mature product**: Increase safety weight
- **Maintenance mode**: Increase effort_reduction weight

**Variation 2: Team Velocity Adjustment**

Adjust effort estimates based on team experience:
```
effective_effort = estimated_effort / team_velocity_factor
```

**Variation 3: Dependency-Aware Prioritization**

If Task B depends on Task A:
```
priority_B_adjusted = priority_B × (1 + priority_A)
```

**Variation 4: Time-Boxed Execution**

Set time budget, fill with highest-priority tasks:
```
budget = 8 hours
selected = []
for task in sorted_by_priority:
    if sum(selected.effort) + task.effort <= budget:
        selected.append(task)
```

### Reusability

**Domain Agnostic**: ✅ Applies to any constrained optimization problem
**Not Code-Specific**: ✅ Applies beyond refactoring

**Applies to**:
- Software refactoring (any language)
- Technical debt prioritization
- Sprint planning
- Bug triage
- Security patch prioritization
- Feature development prioritization
- DevOps task scheduling

**Transferability Examples**:

**Bug Triage**:
```
value = user_impact
safety = 1 - regression_risk
effort = time_to_fix
priority = (user_impact × (1-regression_risk)) / time_to_fix
```

**Security Patches**:
```
value = severity × exploitability
safety = 1 - breaking_changes_risk
effort = deployment_complexity
priority = (severity × exploitability × (1-breaking_changes_risk)) / deployment_complexity
```

**Feature Development**:
```
value = business_value
safety = 1 - scope_creep_risk
effort = story_points
priority = (business_value × (1-scope_creep_risk)) / story_points
```

### Key Takeaways

- ✅ **Use objective formula** (value × safety / effort)
- ✅ **Prioritize data over intuition** (measure, don't guess)
- ✅ **Re-assess dynamically** (adjust when new info emerges)
- ✅ **Skip low-priority tasks** (pragmatism beats completionism)
- ✅ **Document decisions** (for future reference and learning)

**One-sentence summary**: Prioritize refactoring tasks using objective formula (value × safety / effort), skip P3 tasks when time-constrained.

---

## Pattern 4: Incremental Test Addition

### Context

**When to use**: When a package has low test coverage (<50%) and you want to improve it systematically.

**When not to use**: When coverage is already high (>80%), or when tests don't exist and major refactoring is needed first.

**Frequency**: During test improvement sprints, or continuously as part of TDD.

### Problem

**Low test coverage leads to**:
- **Fear of refactoring**: "I can't change this, it might break"
- **Difficult debugging**: No tests to isolate failures
- **Undocumented behavior**: Tests serve as living documentation
- **Regression risk**: Changes break functionality silently

**Common test addition mistakes**:
- **Boiling the ocean**: "I'll write tests for everything"
- **Unfocused effort**: Writing tests randomly across codebase
- **Ignoring coverage metrics**: Not measuring improvement
- **Testing internal details**: Tests break on refactoring

### Solution

**Systematically add focused tests to one low-coverage package at a time, measuring improvement**.

**Core Idea**: Target specific packages, write tests for exported functions, measure coverage increase.

### Detailed Steps

1. **Identify low-coverage packages**:
   ```bash
   # Go
   go test -cover ./... | grep -E "coverage: [0-4][0-9]%"

   # Python
   pytest --cov=. --cov-report=term-missing | grep -E "[0-4][0-9]%"

   # JavaScript
   jest --coverage | grep -E "[0-4][0-9]%"
   ```

2. **Select target package**: Choose one package with <50% coverage

3. **List exported functions**:
   ```bash
   # Go: List exported functions (start with capital letter)
   grep "^func [A-Z]" internal/validation/*.go

   # Python: List public functions (no leading underscore)
   grep "^def [^_]" mypackage/*.py
   ```

4. **Create test file**: Follow naming convention
   ```bash
   # Go
   touch internal/validation/description_test.go

   # Python
   touch tests/test_description.py

   # JavaScript
   touch src/__tests__/description.test.js
   ```

5. **Write test for first function**:
   - Start with **success case** (happy path)
   - Use table-driven tests if applicable

6. **Run test, verify passes**:
   ```bash
   go test ./internal/validation -v
   ```

7. **Add failure case test**: Test error conditions

8. **Add edge case tests**: Boundary conditions, nil/null values, empty inputs

9. **Repeat for remaining functions**: One function at a time

10. **Measure coverage improvement**:
    ```bash
    # Before
    go test -cover ./internal/validation
    coverage: 0%

    # After
    go test -cover ./internal/validation
    coverage: 32.5%
    ```

11. **Verify all tests pass**:
    ```bash
    go test ./...
    ```

12. **Document test coverage**: Update README or metrics dashboard

### Example: Iteration 2 (Meta-CC Project)

#### Target Package: internal/validation

**Initial state**:
- Coverage: 0%
- Files: description.go, ordering.go, naming.go
- Exported functions: 10

#### Step 1: Create test files

```bash
$ touch internal/validation/description_test.go
$ touch internal/validation/ordering_test.go
```

#### Step 2: Write tests for description.go

```go
// internal/validation/description_test.go
package validation

import "testing"

func TestValidateToolDescription(t *testing.T) {
    tests := []struct {
        name        string
        description string
        wantErr     bool
    }{
        {
            name:        "valid description",
            description: "Query tool calls with filters. Default scope: project.",
            wantErr:     false,
        },
        {
            name:        "too long",
            description: "This is a very long description that exceeds the maximum allowed length of 100 characters for tool descriptions in MCP",
            wantErr:     true,
        },
        {
            name:        "missing scope suffix",
            description: "Query tool calls with filters.",
            wantErr:     true,
        },
        // ... 5 more test cases ...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateToolDescription(tt.description)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateToolDescription() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

#### Step 3: Write tests for ordering.go

```go
// internal/validation/ordering_test.go
package validation

import "testing"

func TestParameterOrdering(t *testing.T) {
    // 10 test functions covering:
    // - Tier ordering (required → filtering → range → output)
    // - Alphabetical sorting within tiers
    // - Edge cases (empty, single parameter, all same tier)
}
```

#### Results

- **Tests added**: 2 files, 10 functions, ~300 lines
- **Coverage improvement**: 0% → 32.5%
- **All tests pass**: ✅
- **Time spent**: ~3 hours

**Coverage breakdown**:
- description.go: 85% (8 test cases)
- ordering.go: 90% (10 test functions)
- naming.go: 0% (deferred to future work)

### Before/After Comparison

#### Before
```
$ go test -cover ./internal/validation
?       github.com/yaleh/meta-cc/internal/validation    [no test files]
```

#### After
```
$ go test -cover ./internal/validation
ok      github.com/yaleh/meta-cc/internal/validation    0.005s  coverage: 32.5% of statements
```

### Verification Checklist

- [ ] Identified low-coverage package (< 50%)
- [ ] Selected target package
- [ ] Listed exported functions
- [ ] Created test file with correct naming convention
- [ ] Wrote test for first function (success case)
- [ ] Verified first test passes
- [ ] Added failure case test
- [ ] Added edge case tests
- [ ] Repeated for remaining functions
- [ ] Measured coverage improvement
- [ ] Verified all tests pass
- [ ] Documented coverage increase

### Pitfalls and How to Avoid

**Pitfall 1**: Testing internal implementation details
- ❌ Wrong: Test private functions, internal state
- ✅ Right: Test exported API, observable behavior

**Pitfall 2**: Unfocused effort across multiple packages
- ❌ Wrong: Write one test each for 10 packages
- ✅ Right: Improve one package from 0% to 30%+

**Pitfall 3**: Not measuring coverage
- ❌ Wrong: "I wrote tests, coverage probably improved"
- ✅ Right: Run coverage tool before/after, document improvement

**Pitfall 4**: Writing tests that don't fail
- ❌ Wrong: Tests always pass (testing nothing)
- ✅ Right: Temporarily break code, verify test fails, fix code

**Pitfall 5**: Ignoring test maintainability
- ❌ Wrong: Copy-paste tests with minor variations
- ✅ Right: Use table-driven tests, shared fixtures

### Variations

**Variation 1: TDD (Test-First)**

If writing new code:
1. Write test first (it fails)
2. Write minimal code to pass test
3. Refactor
4. Repeat

**Variation 2: Coverage-Driven (Target-Based)**

Set target (e.g., 75%):
1. Measure current coverage
2. Identify gap (75% - current%)
3. Add tests until gap closed
4. Stop (avoid over-testing)

**Variation 3: Risk-Driven (High-Risk First)**

Prioritize tests for:
- Functions with most bugs historically
- Functions with highest complexity
- Functions in critical path

**Variation 4: Integration-First**

If unit coverage is high but integration is low:
1. Write integration tests first
2. Add unit tests for failures found
3. Measure both unit and integration coverage

### Reusability

**Language Agnostic**: ✅ Concept applies universally
**Tool Specific**: ⚠️ Commands differ by language

**Applies to**:
- Any codebase with test infrastructure
- Any language with coverage tooling
- Unit tests, integration tests, E2E tests
- Backend, frontend, mobile, embedded

**Transferability Examples**:

**Python (pytest)**:
```bash
# Identify low-coverage modules
$ pytest --cov=mypackage --cov-report=term-missing
mypackage/validation.py    0%

# Create test file
$ touch tests/test_validation.py

# Write tests (similar structure to Go)
def test_validate_description_valid():
    assert validate_description("Valid description.") is None

def test_validate_description_too_long():
    with pytest.raises(ValueError):
        validate_description("x" * 200)

# Measure improvement
$ pytest --cov=mypackage/validation
mypackage/validation.py    85%
```

**JavaScript (Jest)**:
```bash
# Identify low-coverage files
$ jest --coverage
validation.js    0%

# Create test file
$ touch src/__tests__/validation.test.js

# Write tests
describe('validateDescription', () => {
  test('valid description', () => {
    expect(() => validateDescription('Valid description.')).not.toThrow();
  });

  test('too long', () => {
    expect(() => validateDescription('x'.repeat(200))).toThrow();
  });
});

# Measure improvement
$ jest --coverage
validation.js    85%
```

### Key Takeaways

- ✅ **Target one package at a time** (focused effort beats scattered)
- ✅ **Test exported behavior** (not internal implementation)
- ✅ **Measure improvement** (before/after coverage)
- ✅ **Use table-driven tests** (for readability and maintainability)
- ✅ **Verify tests fail when they should** (test the tests)

**One-sentence summary**: Systematically add tests to one low-coverage package at a time, focusing on exported functions and measuring coverage improvement.

---

## Pattern Application Framework

### How to Select Patterns

Use this decision tree:

```
START
  │
  ├─ Need to remove code?
  │   └─ YES → Use Pattern 1 (Verify Before Remove)
  │
  ├─ Have repetitive structure definitions?
  │   └─ YES → Use Pattern 2 (Builder Extraction)
  │
  ├─ Have multiple refactoring tasks?
  │   └─ YES → Use Pattern 3 (Risk Prioritization)
  │
  ├─ Have low test coverage?
  │   └─ YES → Use Pattern 4 (Incremental Test Addition)
  │
  └─ None of above?
      └─ Re-assess problem or consult other methodologies
```

### How to Sequence Patterns

**Typical refactoring workflow**:

```
Phase 1: Planning
  ↓
  Apply Pattern 3 (Risk Prioritization)
  - Assess all tasks
  - Calculate priorities
  - Select P1 tasks
  ↓
Phase 2: Safety Enhancement
  ↓
  Apply Pattern 4 (Incremental Test Addition)
  - Improve coverage for target packages
  - Establish test baseline
  ↓
Phase 3: Refactoring (for each task)
  ↓
  If task involves deletion:
    Apply Pattern 1 (Verify Before Remove)
    - Run static analyzer
    - Check coverage
    - Search references
  ↓
  If task involves duplication:
    Apply Pattern 2 (Builder Extraction)
    - Identify patterns
    - Extract helpers
    - Test incrementally
  ↓
Phase 4: Validation
  ↓
  Run full test suite
  Measure coverage
  Calculate ΔV
  ↓
Phase 5: Iteration
  ↓
  Re-apply Pattern 3
  - Assess remaining tasks
  - Decide to continue or converge
```

### Pattern Composition

**Patterns work together**:

| Primary Pattern | Supporting Patterns | Use Case |
|----------------|-------------------|----------|
| Pattern 1 (Verify) | Pattern 4 (Tests) | Remove code safely after ensuring test coverage |
| Pattern 2 (Extract) | Pattern 1 (Verify) | Extract helpers after verifying current usage |
| Pattern 3 (Prioritize) | All patterns | Decide which pattern to apply first |
| Pattern 4 (Tests) | Pattern 1 (Verify) | Add tests then verify coverage improves |

**Example: Refactoring a large API file**

1. **Apply Pattern 3**: Assess tasks
   - Extract helpers (P1)
   - Remove unused endpoints (P2)
   - Split file (P3, skip)

2. **Apply Pattern 4**: Add tests
   - Current coverage: 45%
   - Target: 70%
   - Add tests for 5 endpoints

3. **Apply Pattern 2**: Extract helpers
   - Identify duplication
   - Extract common parameters
   - Test after each extraction

4. **Apply Pattern 1**: Verify before removing
   - Check if endpoints truly unused
   - Run static analyzer
   - Remove if verified

**Result**: Safe, incremental refactoring with measurable improvement.

---

## Decision Trees

### Decision Tree 1: Should I Refactor This?

```
Is there a measurable problem?
  NO → Don't refactor (avoid premature optimization)
  YES ↓

Do tests exist (≥50% coverage)?
  NO → Write tests first (Pattern 4)
  YES ↓

Is the improvement high-value (ΔV ≥ 0.05)?
  NO → Defer refactoring (low ROI)
  YES ↓

Is the change low-risk (safety ≥ 0.7)?
  NO → Can risk be mitigated?
    NO → Defer refactoring
    YES → Mitigate (add tests, plan rollback)
  YES ↓

Is effort reasonable (≤8 hours)?
  NO → Break into smaller tasks
  YES ↓

→ PROCEED WITH REFACTORING
```

### Decision Tree 2: Which Pattern Should I Use?

```
What is the primary goal?

│
├─ Remove code
│   ↓
│   Is code truly unused?
│     UNKNOWN → Use Pattern 1 (Verify Before Remove)
│     YES → Use Pattern 1 (Verify Before Remove)
│
├─ Reduce duplication
│   ↓
│   Is duplication in structure definitions?
│     YES → Use Pattern 2 (Builder Extraction)
│     NO → Consider other refactoring approaches
│
├─ Decide what to refactor
│   ↓
│   Multiple tasks with constraints?
│     YES → Use Pattern 3 (Risk Prioritization)
│     NO → Just do the task
│
└─ Improve test coverage
    ↓
    Coverage < 50%?
      YES → Use Pattern 4 (Incremental Test Addition)
      NO → Coverage is adequate, focus on other goals
```

### Decision Tree 3: Should I Skip This Task?

```
Calculate priority = (value × safety) / effort

Priority ≥ 1.0?
  YES → Execute task (P1)
  NO ↓

Priority ≥ 0.5?
  YES ↓
    Time constraints exist?
      NO → Execute task (P2)
      YES ↓
        Will skipping prevent convergence?
          YES → Execute task
          NO → SKIP task
  NO → SKIP task (P3)
```

---

## Success Metrics

### Value Function

**Aggregate value function**:
```
V(s) = w₁×V_code_quality + w₂×V_maintainability + w₃×V_safety + w₄×V_effort

where:
  w₁ = 0.30 (code quality weight)
  w₂ = 0.30 (maintainability weight)
  w₃ = 0.20 (safety weight)
  w₄ = 0.20 (effort weight)
```

**Component metrics**:

| Component | Measurement | Target |
|-----------|-------------|--------|
| V_code_quality | 1 - (violations / LOC) | ≥ 0.90 |
| V_maintainability | Subjective (0.0-1.0) based on complexity, duplication, structure | ≥ 0.70 |
| V_safety | Test coverage | ≥ 0.75 |
| V_effort | 1 - (remaining_work / total_work) | ≥ 0.75 |

**Convergence criterion**:
```
V(s) ≥ 0.80 (80th percentile)
```

### Quantitative Metrics

**Before/after measurements**:

| Metric | How to Measure | Improvement Target |
|--------|----------------|-------------------|
| Line count | `wc -l file.go` | Reduce ≥10% |
| Duplication | `grep -c pattern` | Reduce ≥50% |
| Test coverage | `go test -cover` | Increase ≥10 pp |
| Complexity | `gocyclo` (Go), `radon` (Python) | Reduce ≥20% |
| Static violations | `staticcheck` | Reduce to 0 |

**Example (Iteration 2)**:
- Line count: 396 → 321 (-18.9%) ✅
- Duplication: 69 lines → 0 lines (-100%) ✅
- Test coverage: 57.9% → 57.9% (no change, different package)
- All tests pass: ✅

### Qualitative Metrics

**Subjective assessments (0.0-1.0)**:

| Quality | Assessment Questions | Scoring |
|---------|---------------------|---------|
| Readability | Can new developer understand code in <10 minutes? | 0.0=No, 0.5=Partially, 1.0=Yes |
| Extensibility | Can add new feature with <50 lines? | 0.0=No, 0.5=With refactoring, 1.0=Yes |
| Debuggability | Can isolate bug with tests + logs? | 0.0=No, 0.5=With effort, 1.0=Easily |
| Consistency | Does code follow project conventions? | 0.0=No, 0.5=Mostly, 1.0=Fully |

**Aggregate**:
```
V_maintainability = 0.25×readability + 0.25×extensibility + 0.25×debuggability + 0.25×consistency
```

### Process Metrics

**Efficiency tracking**:

| Metric | Target | Measurement |
|--------|--------|-------------|
| Estimation accuracy | ±20% | actual_time / estimated_time |
| Rollback rate | <10% | rollbacks / total_tasks |
| Test pass rate | 100% | passing_tests / total_tests |
| Review cycles | ≤2 | reviews_needed / task |

**Example tracking**:
```yaml
task:
  estimated_hours: 2-4
  actual_hours: 4
  accuracy: 100% (within range)

  estimated_ΔV: 0.04
  actual_ΔV: 0.034
  accuracy: 85%

  rollback_needed: false
```

---

## Appendices

### Appendix A: Templates

See `methodology-templates/` directory:
- `refactoring-task-template.yaml` - Structured task planning
- `pattern-application-checklist.md` - Step-by-step execution guide
- `risk-assessment-matrix.yaml` - Objective prioritization framework

### Appendix B: Tool Setup

**Go**:
```bash
# Static analysis
go install honnef.co/go/tools/cmd/staticcheck@latest

# Coverage
go test -cover ./...

# Complexity
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
```

**Python**:
```bash
# Static analysis
pip install pylint mypy vulture

# Coverage
pip install pytest pytest-cov

# Complexity
pip install radon
```

**JavaScript/TypeScript**:
```bash
# Static analysis
npm install -g eslint typescript

# Coverage
npm install --save-dev jest

# Complexity
npm install -g complexity-report
```

### Appendix C: Real-World Examples

**Example 1: Meta-CC tools.go Refactoring**
- File: `cmd/mcp-server/tools.go`
- Pattern: InputSchema Builder Extraction
- Before: 396 lines, 69 lines duplication
- After: 321 lines (-18.9%)
- Result: All tests pass, behavioral equivalence preserved

**Example 2: Meta-CC validation Test Addition**
- Package: `internal/validation`
- Pattern: Incremental Test Addition
- Before: 0% coverage
- After: 32.5% coverage
- Tests: 2 files, 10 functions, ~300 lines

**Example 3: Iteration 2 Task Prioritization**
- Pattern: Risk-Based Task Prioritization
- Tasks: 3 candidates
- Decision: Complete P1 + P2, skip P3
- Result: Convergence achieved (V=0.804)

### Appendix D: Further Reading

**Books**:
- "Refactoring: Improving the Design of Existing Code" by Martin Fowler
- "Working Effectively with Legacy Code" by Michael Feathers
- "The Pragmatic Programmer" by Hunt & Thomas

**Articles**:
- "The Boy Scout Rule" (Uncle Bob)
- "Technical Debt Quadrant" (Martin Fowler)
- "Refactoring at Scale" (various authors)

**Tools**:
- SonarQube (multi-language code quality)
- CodeClimate (automated code review)
- DeepSource (static analysis platform)

### Appendix E: Methodology Evolution

This methodology was extracted from:
- **Iteration 1** (bootstrap-004): Pattern 1 discovered
- **Iteration 2** (bootstrap-004): Patterns 2-4 discovered
- **Iteration 3** (bootstrap-004): Patterns validated and documented

**Validation evidence**:
- 4 patterns applied across 2 iterations
- 100% success rate (no rollbacks)
- Convergence achieved (V=0.804 ≥ 0.80)
- All patterns have real-world evidence

**Future improvements**:
- Test on non-Go projects (Python, TypeScript)
- Expand pattern catalog (6-8 patterns)
- Create IDE integrations
- Build automated tooling

---

## Conclusion

This methodology provides a **systematic, evidence-based approach to safe refactoring**. The four patterns address the most common refactoring scenarios:

1. **Pattern 1** (Verify Before Remove): Prevents costly mistakes
2. **Pattern 2** (Builder Extraction): Reduces duplication safely
3. **Pattern 3** (Risk Prioritization): Optimizes task selection
4. **Pattern 4** (Incremental Tests): Improves safety systematically

**Key principles**:
- Trust tools over intuition
- Incremental progress beats perfection
- Safety trumps speed
- Evidence-based decisions
- Pragmatic adaptation

**Success criteria**:
- V(s) ≥ 0.80 (measurable improvement)
- No regressions (all tests pass)
- Effort reasonable (≤8 hours per task)
- Value delivered (ΔV ≥ 0.05)

**Use this methodology when**:
- Code needs improvement (duplication, complexity, low coverage)
- Safety is paramount
- Time/budget constraints exist
- Incremental approach acceptable

**Next steps**:
1. Review templates (Appendix A)
2. Set up tooling (Appendix B)
3. Assess your codebase
4. Apply Pattern 3 to prioritize tasks
5. Execute with appropriate patterns
6. Measure improvement
7. Iterate or converge

---

**Version**: 1.0
**Status**: Validated (bootstrap-004 experiment)
**License**: MIT
**Contact**: meta-cc project team

**Last Updated**: 2025-10-16
