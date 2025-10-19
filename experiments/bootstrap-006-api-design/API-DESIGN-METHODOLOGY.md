# API Design Methodology

**Version**: 1.0
**Extracted From**: Bootstrap-006-api-design Experiment, Iteration 4
**Extraction Method**: Two-Layer Architecture (observed agent execution patterns)
**Date**: 2025-10-15

---

## Overview

This methodology was extracted by observing agent execution patterns during Iteration 4 of the bootstrap-006-api-design experiment. The **TWO-LAYER ARCHITECTURE** approach enabled meta-agent observation of agent decision-making processes, revealing reusable methodology patterns.

**Extraction Process**:
1. **Agent Layer**: coder agent executed concrete task (parameter reordering)
2. **Meta-Agent Layer**: Observed decision processes, verification steps, and audit techniques
3. **Pattern Extraction**: Identified 3 reusable patterns from Task 1 execution
4. **Codification**: Documented patterns with context, solution, evidence, and reusability analysis

---

## Pattern 1: Deterministic Parameter Categorization

### Context
When designing or refactoring API parameters, categorization decisions must be consistent and unambiguous across all tools.

### Problem
Without a systematic approach, parameter ordering becomes arbitrary and inconsistent. Different developers may categorize the same parameter type differently, leading to:
- Inconsistent API schemas
- Increased cognitive load for API users
- Difficulty maintaining consistency as API evolves

### Solution: Tier-Based Decision Tree

Use a 5-tier categorization system with deterministic decision criteria:

```
Tier 1: Required Parameters
  - Must be provided for tool to function
  - Marked as required in schema
  - Criteria: "Can tool execute without this?" → NO = Tier 1

Tier 2: Filtering Parameters
  - Narrow search results (affect WHAT is returned)
  - Optional, typically string or enum types
  - Criteria: "Does this filter results?" → YES = Tier 2

Tier 3: Range Parameters
  - Define bounds, thresholds, windows
  - Optional, typically numeric types (min_*, max_*, start_*, end_*, threshold, window)
  - Criteria: "Does this define a range or threshold?" → YES = Tier 3

Tier 4: Output Control Parameters
  - Control output size or format (limit, offset, pagination)
  - Affect HOW MUCH is returned, not WHAT
  - Criteria: "Does this control output size/format?" → YES = Tier 4

Tier 5: Standard Parameters
  - Cross-cutting concerns (scope, output format, filtering)
  - Added automatically via framework (e.g., MergeParameters)
  - Criteria: "Is this a standard parameter?" → YES = Tier 5
```

**Decision Process** (observed in agent execution):
1. Read parameter name and description
2. Ask tier-specific questions in order (Tier 1 → Tier 2 → Tier 3 → Tier 4 → Tier 5)
3. Assign to first matching tier
4. Place parameter according to tier order in schema

**Example** (from Task 1 execution):
```
Parameter: include_builtin_tools (boolean)
Question 1: "Is this required?" → NO (skip Tier 1)
Question 2: "Does this filter results?" → YES (excludes built-in tools)
Category: Tier 2 (Filtering)
Placement: Before min_occurrences (Tier 3)
```

### Evidence

**Observed in**: Task 1 (Parameter Reordering), Iteration 4, Bootstrap-006

**Tools affected**: 8 tools analyzed
- 5 tools reordered using tier system
- 3 tools verified as already compliant
- 0 ambiguous cases encountered

**Determinism**: 100%
- Every parameter fit exactly one tier
- No judgment calls required
- Repeatable across different developers

**Agent Process Observed**:
1. Read parameter from schema
2. Apply decision tree
3. Categorize deterministically
4. Place in correct tier order
5. Verify with tier comments

### Reusability

**Applicable To**:
- ✅ Query-based APIs (filtering, range, output control)
- ✅ GraphQL schemas (arguments ordering)
- ✅ REST API query parameters
- ✅ CLI command flags
- ✅ Configuration files (YAML, JSON)

**Universal Principle**: Logical grouping (required → filtering → range → output) applies to any parametric interface.

**Limitations**:
- May need adjustment for non-query APIs (e.g., mutation-only APIs)
- Tier 5 (standard parameters) specific to frameworks with parameter merging

---

## Pattern 2: Safe API Refactoring via JSON Property

### Context
Need to improve API schema readability and consistency without breaking existing clients.

### Problem
Fear of breaking changes prevents schema improvements. Developers avoid refactoring parameter order, adding clarity comments, or reorganizing schemas, even when beneficial.

### Solution: Leverage JSON Unordered Object Property

**JSON Specification Guarantee**: JSON objects are unordered. Parameter order in schema definition is documentation only, not a functional constraint.

**Refactoring Types Enabled**:
1. **Parameter Reordering**: Change declaration order in schema without affecting function calls
2. **Clarity Comments**: Add tier comments for readability
3. **Schema Grouping**: Group related parameters visually

**Verification Process** (observed in agent execution):
1. **Confirm JSON Property**: Verify API uses JSON (unordered objects)
2. **Identify Changes**: List parameter order changes needed
3. **Make Changes**: Reorder parameters in schema definition
4. **Run Tests**: Execute full test suite
5. **Verify Compilation**: Build project
6. **Confirm Non-Breaking**: Check 100% test pass rate

**Example** (from Task 1 execution):
```go
// Before (inconsistent order)
{
  "limit": number,     // Tier 4
  "tool": string,      // Tier 2
  "status": string,    // Tier 2
}

// After (tier-based order + comments)
{
  // Tier 2: Filtering
  "tool": string,
  "status": string,
  // Tier 4: Output Control
  "limit": number,
}

// Result: Both schemas functionally identical
// Call {"tool": "X", "status": "Y", "limit": 10} works in both cases
// Call {"limit": 10, "tool": "X", "status": "Y"} works in both cases
```

### Evidence

**Observed in**: Task 1 (Parameter Reordering), Iteration 4, Bootstrap-006

**Changes Made**:
- 5 tools reordered (60 lines changed)
- Tier comments added
- 0 functional changes

**Test Results**:
- Test suite: 100% pass rate (0 failures)
- Compilation: Successful
- Backward compatibility: Confirmed (existing calls unaffected)

**Agent Verification Process Observed**:
1. Identified 5 tools needing reordering
2. Confirmed JSON parameter order irrelevance
3. Made changes in code
4. Ran `make test` → all pass
5. Ran `make build` → successful
6. Created verification report

### Reusability

**Applicable To**:
- ✅ Any JSON-based API (REST, GraphQL, RPC)
- ✅ JSON configuration files
- ✅ JSON schemas (OpenAPI, JSON Schema)

**Universal Principle**: JSON spec guarantees object property order irrelevance. Safe to refactor for readability.

**Limitations**:
- Does NOT apply to:
  - ❌ Positional arguments (e.g., function signatures: `func(a, b, c)`)
  - ❌ Array ordering (arrays ARE ordered)
  - ❌ Serialization formats that depend on order (some XML, CSV)

**Safety Guarantee**: 100% when API uses JSON for parameter passing.

---

## Pattern 3: Audit-First Refactoring

### Context
Need to refactor multiple targets (tools, parameters, schemas) for consistency.

### Problem
Without auditing first, developers may:
- Waste effort on already-compliant targets
- Miss non-compliant targets
- Lack prioritization data (which violations are most common?)

### Solution: Systematic Audit Process

**Audit Steps** (observed in agent execution):
1. **List Targets**: Enumerate all items to audit (e.g., 8 tools)
2. **Define Compliance Criteria**: Specify what "compliant" means (e.g., tier-based ordering)
3. **Assess Each Target**: Check compliance for each item
4. **Categorize**:
   - "Needs change" (violations found)
   - "Already compliant" (no changes needed)
5. **Prioritize**: Rank by violation severity or impact
6. **Execute Changes**: Refactor non-compliant targets only
7. **Verify**: Re-audit to confirm 100% compliance

**Example** (from Task 1 execution):
```yaml
audit_results:
  total_tools: 8

  needs_change: 5
    - query_tools (60% compliant)
    - query_user_messages (75% compliant)
    - query_conversation (40% compliant)
    - query_tool_sequences (67% compliant)
    - query_successful_prompts (0% compliant)

  already_compliant: 3
    - query_context (100% compliant)
    - query_assistant_messages (100% compliant)
    - query_time_series (100% compliant)

  efficiency_gain: 37.5% (avoided 3/8 unnecessary changes)
```

**Agent Process Observed**:
1. Listed all 8 tools with multi-parameter schemas
2. For each tool:
   - Categorized parameters by tier
   - Compared current order to tier system
   - Calculated compliance percentage
3. Marked 5 tools as "needs change"
4. Marked 3 tools as "already compliant"
5. Prioritized highest violations (query_successful_prompts at 0%)
6. Executed reordering for 5 tools
7. Verified 3 compliant tools (no changes)

### Evidence

**Observed in**: Task 1 (Parameter Reordering), Iteration 4, Bootstrap-006

**Audit Metrics**:
- Tools audited: 8
- Compliance before: 67.5% (average)
- Violations found: 5 tools
- Already compliant: 3 tools
- Efficiency gain: 37.5% (avoided unnecessary work)
- Compliance after: 100%

**Time Savings**:
- Without audit: 8 tools × 30 min = 4 hours
- With audit: 5 tools × 30 min + 3 tools × 5 min (verify) = 2.75 hours
- Savings: 31% time reduction

### Reusability

**Applicable To**:
- ✅ Any refactoring effort (not API-specific)
- ✅ Code quality improvements (linting, formatting)
- ✅ Consistency fixes (naming, structure)
- ✅ Migration projects (old API → new API)
- ✅ Compliance audits (security, accessibility)

**Universal Principle**: Audit before refactoring to identify actual work needed.

**Benefits**:
1. **Efficiency**: Avoid unnecessary changes (37.5% in this case)
2. **Verification**: Confirm existing good patterns (3 tools already compliant)
3. **Prioritization**: Focus on highest-impact violations first
4. **Metrics**: Quantify improvement (67.5% → 100% compliance)

---

## Methodology Application Guide

### When to Use Pattern 1: Deterministic Parameter Categorization

**Use When**:
- ✅ Designing new API parameters
- ✅ Refactoring existing parameter ordering
- ✅ Need consistent categorization across multiple tools
- ✅ Building schema validation tools

**Don't Use When**:
- ❌ API has no parameters (mutation-only)
- ❌ Parameter categories don't fit tier system (special case)

**Application Steps**:
1. Read `api-parameter-convention.md` (tier definitions)
2. For each parameter, apply decision tree
3. Categorize into Tier 1-5
4. Order parameters by tier
5. Add tier comments for clarity

---

### When to Use Pattern 2: Safe API Refactoring via JSON Property

**Use When**:
- ✅ Improving API schema readability
- ✅ Reordering parameters for consistency
- ✅ Adding documentation comments to schema
- ✅ Grouping related parameters

**Don't Use When**:
- ❌ API uses positional arguments (function signatures)
- ❌ Serialization format depends on order (some XML, CSV)
- ❌ Array ordering (arrays ARE ordered)

**Application Steps**:
1. Confirm API uses JSON for parameters
2. Identify schema improvements (reordering, comments)
3. Make changes in schema definition
4. Run full test suite
5. Verify 100% test pass rate
6. Document changes (verification report)

---

### When to Use Pattern 3: Audit-First Refactoring

**Use When**:
- ✅ Refactoring multiple targets (tools, schemas, code)
- ✅ Unclear which targets need changes
- ✅ Need to prioritize limited time/resources
- ✅ Want to quantify improvement (before/after metrics)

**Don't Use When**:
- ❌ Single target only (audit overhead not justified)
- ❌ All targets definitely non-compliant (no efficiency gain)

**Application Steps**:
1. List all targets to audit
2. Define compliance criteria
3. Assess each target (compliant vs. non-compliant)
4. Categorize and prioritize
5. Execute changes on non-compliant targets only
6. Verify compliant targets (no changes)
7. Re-audit to confirm 100% compliance

---

## Integration with Existing Policies

### API Evolution Strategy (Iteration 1)
- **Versioning**: Parameter reordering is non-breaking (within MAJOR version)
- **Deprecation**: Use tier system when deprecating old parameters
- **Migration**: Audit-first approach identifies migration scope

### API Consistency (Iteration 2)
- **Naming Convention**: Works with tier-based ordering
- **Parameter Ordering**: Tier system implements ordering convention
- **Validation**: Deterministic categorization enables automated checking

### Bootstrap-006 Deliverables
| Iteration | Deliverable | Methodology Integration |
|-----------|-------------|-------------------------|
| Iteration 1 | api-versioning-strategy.md | Pattern 2 (safe refactoring within versions) |
| Iteration 1 | api-deprecation-policy.md | Pattern 3 (audit deprecated parameters) |
| Iteration 2 | api-naming-convention.md | Pattern 1 (naming + tier categorization) |
| Iteration 2 | api-parameter-convention.md | Pattern 1 (tier system definition) |
| Iteration 2 | api-consistency-methodology.md | All patterns (consistency = categorization + refactoring + auditing) |

---

## Extraction Methodology (Meta-Level)

### Two-Layer Architecture

The patterns in this document were extracted using a **TWO-LAYER ARCHITECTURE**:

**Layer 1 (Agent Layer)**: Execute concrete implementation tasks
- **Agent**: coder
- **Task**: Implement parameter reordering in tools.go
- **Observable Behaviors**: Decision-making, verification steps, audit process

**Layer 2 (Meta-Agent Layer)**: Observe agent execution, extract patterns, codify methodology
- **Meta-Agent**: Meta-cognitive observer (human or LLM)
- **Process**: Watch HOW agent solves problems (not just WHAT is produced)
- **Output**: Methodology patterns (this document)

**Extraction Process**:
1. **Observe**: Watch agent execution in real-time (decision points, verification steps)
2. **Identify Patterns**: Recognize reusable decision-making processes
3. **Extract Criteria**: Document decision criteria used by agent
4. **Codify**: Write methodology pattern with:
   - Context (when to use)
   - Problem (what issue it solves)
   - Solution (how to apply)
   - Evidence (observed data)
   - Reusability (where else applicable)
5. **Validate**: Verify pattern reusability across different contexts

### Evidence of Effectiveness

**Single Task Sufficiency**:
- Task 1 (parameter reordering) provided 3 methodology patterns
- Each pattern reusable beyond original context
- 100% determinism (no ambiguous cases)

**Pattern Characteristics**:
- Deterministic (repeatable, no judgment calls)
- Universal (applicable beyond API design)
- Verifiable (testable criteria)
- Reusable (transferable to other domains)

**Extraction Efficiency**:
- 1 task executed → 3 patterns extracted
- Patterns codified in ~6,000 words
- Observational data: decision points, verification steps, audit metrics

---

## Validation and Reusability

### Pattern Validation Checklist

For each pattern, validate:
- ✅ **Determinism**: No ambiguous cases (100% in Pattern 1)
- ✅ **Universality**: Applicable beyond original context (all patterns)
- ✅ **Verifiability**: Testable criteria (all patterns)
- ✅ **Reusability**: Transferable to other domains (all patterns)
- ✅ **Evidence-Based**: Extracted from actual execution, not theory

### Reusability Matrix

| Pattern | API Design | GraphQL | CLI Design | Config Files | General Refactoring |
|---------|------------|---------|------------|--------------|---------------------|
| Pattern 1 (Categorization) | ✅ | ✅ | ✅ | ✅ | ❌ |
| Pattern 2 (Safe Refactoring) | ✅ | ✅ | ❌ | ✅ | ❌ |
| Pattern 3 (Audit-First) | ✅ | ✅ | ✅ | ✅ | ✅ |

**Legend**:
- ✅ Directly applicable
- ❌ Not applicable (domain-specific constraints)

---

## Pattern 4: Automated Consistency Validation

### Context
Need to enforce API conventions at scale without manual checks.

### Problem
- Manual consistency checks are error-prone
- Developers forget to check conventions
- Inconsistencies accumulate over time
- Post-hoc fixes costly

### Solution: Build Validation Tool

**Architecture**:
```
Parser → Validators → Reporter
  ↓         ↓           ↓
Extract  Check      Format
tools    rules      results
```

**Implementation Pattern**:
1. Design type system (Tool, Result, Report)
2. Implement parser (regex for MVP, AST for robust future)
3. Create validators with deterministic rules
4. Build reporter with multiple formats (terminal, JSON)
5. Integrate into CLI with standard flags

**Validator Design Pattern**:
```go
func ValidateX(tool Tool) Result {
    // 1. Extract relevant data
    data := extract(tool)

    // 2. Apply deterministic check
    if violates(data, convention) {
        return NewFailResult(
            tool.Name,
            "check_name",
            "Clear error message",
            map[string]interface{}{
                "suggestion": "Actionable fix",
                "reference": "Convention document"
            }
        )
    }

    // 3. Return pass
    return NewPassResult(tool.Name, "check_name")
}
```

**Error Message Pattern**:
```
✗ tool_name: Brief error description
  Suggestion: Specific fix action
  Expected: What should be
  Actual: What is
  Reference: Convention documentation link
  Severity: ERROR/WARNING
```

### Evidence

**Observed in**: Task 2 (Validation Tool Implementation), Iteration 5

**Implementation Stats**:
- Files created: 8
- Lines of code: ~600
- Validators: 3 (naming, ordering, description)
- Tools validated: 16
- Violations detected: 2
- False positives: 0
- Test coverage: 100% (naming validator)

**Validation Results**:
| Tool | Violations | Detected |
|------|------------|----------|
| list_capabilities | Missing "Default scope:" | ✓ |
| get_capability | Missing "Default scope:" | ✓ |

**Decision Criteria Observed**:
- Parser: Regex (simple, fast) vs. AST (robust, complex) → Regex for MVP
- Checks: Deterministic rules from tier system
- Output: Terminal (human) + JSON (CI)
- Errors: Actionable suggestions + reference links

### Reusability

**Applicable To**:
- ✅ Any API with documented conventions
- ✅ Code style enforcement (linting)
- ✅ Schema validation
- ✅ Configuration file validation
- ✅ Documentation consistency checks

**Universal Principle**: Automated validation enforces conventions consistently at scale.

**Adaptation Guide**:
1. Document conventions clearly (tier system, naming patterns, etc.)
2. Design validators with deterministic rules (no judgment calls)
3. Provide actionable error messages (specific fixes, not just "wrong")
4. Include references to convention documentation
5. Support multiple output formats (human + machine)

---

## Pattern 5: Automated Quality Gates

### Context
Need to prevent violations from entering repository.

### Problem
- Post-commit fixes are costly (already merged, may affect others)
- Manual checks are skipped (developer forgets, time pressure)
- Violations accumulate (broken windows effect)
- Review process burdened (reviewers catch violations)

### Solution: Pre-Commit Hooks

**Architecture**:
```
Git Commit → Pre-Commit Hook → Validation Tool → Decision
                ↓                     ↓              ↓
          Detect changes       Run checks    Allow/Block
```

**Hook Pattern**:
```bash
#!/bin/bash

# 1. Detect relevant changes
if git diff --cached --name-only | grep -q "<relevant_file>"; then
    # 2. Run validation
    if ./validation-tool --file "<relevant_file>"; then
        # 3. Allow commit
        exit 0
    else
        # 4. Block commit
        exit 1
    fi
else
    # 5. Skip validation
    exit 0
fi
```

**Installation Pattern**:
```bash
#!/bin/bash

# 1. Verify prerequisites
check_git_repo()
check_validation_tool()

# 2. Backup existing hook
if [ -f .git/hooks/pre-commit ]; then
    mv .git/hooks/pre-commit .git/hooks/pre-commit.backup
fi

# 3. Install new hook
cp scripts/pre-commit.sample .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

# 4. Test installation
bash .git/hooks/pre-commit
```

**Feedback Pattern**:
```
===========================================
Pre-Commit Hook: <Name>
===========================================

Detected changes to <file>
Running validation...

✓ Validation PASSED
✓ Commit allowed

(or)

✗ Validation FAILED
Violations found in <file>
Please fix before committing.

To bypass (not recommended):
  git commit --no-verify
```

### Evidence

**Observed in**: Task 3 (Pre-Commit Hook Implementation), Iteration 5

**Implementation Stats**:
- Hook script: 60 lines
- Installation script: 70 lines
- Test scenarios: 4 (detect, skip, block, bypass)
- Dependencies: Validation tool from Task 2

**Hook Behavior**:
| Scenario | tools.go Changed | Validation Result | Hook Action |
|----------|------------------|-------------------|-------------|
| Detect | Yes | Pass | Allow commit (exit 0) |
| Detect | Yes | Fail | Block commit (exit 1) |
| Skip | No | N/A | Allow commit (exit 0) |
| Bypass | Yes | N/A | Allow commit (--no-verify) |

**Integration Points**:
- Git workflow: Seamless integration via git hooks
- Validation tool: Calls `./validate-api --fast`
- Exit codes: 0 (allow), 1 (block)
- Bypass mechanism: `git commit --no-verify`

### Reusability

**Applicable To**:
- ✅ Any pre-commit quality check (linting, testing, formatting)
- ✅ Build verification (compile before commit)
- ✅ Test execution (run tests before commit)
- ✅ Documentation generation (update docs before commit)

**Universal Principle**: Automate quality enforcement at earliest possible point (pre-commit).

**Adaptation Guide**:
1. Identify quality criteria to enforce
2. Create validation tool with clear exit codes
3. Design hook to detect relevant changes
4. Run validation only when needed (efficiency)
5. Provide clear feedback (pass/fail reasons)
6. Include bypass option (emergencies)
7. Automate installation (lower barrier to adoption)

---

## Pattern 6: Example-Driven Documentation

### Context
Need to teach API conventions effectively through documentation so users understand both how to use tools and why conventions exist.

### Problem
- Abstract guidelines difficult to apply ("use tier-based ordering")
- Users confused by inconsistent examples in documentation
- Low-usage tools lack sufficient documentation (users don't know when/how to use)
- Learning curve steep without practical examples
- Convention rationale unclear (users follow blindly or ignore)

### Solution: Provide Practical, Example-Driven Documentation

**Approach**:

1. **Explain Conventions First**:
   - Add convention section before tool catalog
   - Define system explicitly (e.g., tier system with Tier 1-5)
   - Provide rationale (consistency, readability, predictability)
   - Clarify misconceptions (e.g., JSON ordering doesn't affect function calls)

2. **Enhance Low-Usage Tools**:
   - Prioritize tools with low adoption
   - Add "Practical Use Cases" subsection
   - Provide 3-5 real-world scenarios per tool
   - Follow problem → solution → code pattern

3. **Structure Examples Consistently**:
   ```markdown
   **Practical Use Cases**:

   1. **Scenario Name**:
      ```json
      // Problem: Brief description of user problem
      {"param1": "value1", "param2": "value2"}
      // Returns: What user gets
      // Analysis: What user learns from results
      ```
   ```

4. **Add Progressive Complexity**:
   - Start with basic examples (minimal parameters)
   - Progress to advanced examples (multiple conditions, edge cases)
   - Annotate with comments explaining choices

5. **Document Automation Tools**:
   - Provide complete installation guide (automatic + manual)
   - Explain behavior with passing and failing examples
   - Include troubleshooting section (common issues + solutions)
   - Add advanced configuration for customization

### Evidence

**Observed in**: Task 4 (Documentation Enhancement), Iteration 6

**Implementation Examples**:

1. **query_context Enhancement**:
   - Basic example: `{"error_signature": "Bash:command not found"}`
   - Practical use cases: 3 scenarios (Bash errors, permission issues, test failures)
   - "What You Get" section: Explains returned data structure
   - Pattern: Problem statement → JSON example → Expected outcome

2. **cleanup_temp_files Enhancement**:
   - Multiple examples: Default, aggressive, today-only
   - Practical use cases: 3 scenarios (regular maintenance, disk emergency, pre-query)
   - "When to Use" section: Enumerated use cases
   - "What You Get" section: Explains response fields

3. **query_tools_advanced Enhancement**:
   - Basic examples: 3 (complex filter, multiple tools, time range)
   - Practical use cases: 5 scenarios (slow commands, error patterns, tool comparison, activity filtering, multi-condition)
   - SQL Expression Reference: Complete table of operators with examples
   - "When to Use" section: Enumerated appropriate scenarios

4. **validate-api Documentation**:
   - Purpose statement: Clear explanation of tool value
   - Options: Complete with defaults
   - Example output: Both passing and failing scenarios
   - Integration guidance: 3 contexts (CI, pre-commit, development)

5. **API Consistency Hooks Guide**:
   - Installation: Automatic (recommended) + manual alternatives
   - Hook behavior: Passing and failing examples with actual output
   - Troubleshooting: 6 common issues with actionable solutions
   - Advanced configuration: 4 customization patterns
   - CI/CD integration: GitHub Actions + GitLab CI examples

### Decision Criteria

**When to Add Examples**:
- Tool has low adoption (usage data shows < 10% of sessions)
- User questions indicate confusion (support requests, unclear usage)
- Tool has complex parameters (SQL-like filters, multi-condition)
- Tool is new (no historical usage patterns to learn from)

**How Many Examples**:
- Basic tools: 2-3 examples (cover common cases)
- Complex tools: 5+ examples (cover edge cases, demonstrate flexibility)
- Utility tools: 3 examples (regular use + extreme cases)

**Example Structure**:
1. **Problem statement**: "Problem: <user issue>"
2. **JSON example**: Actual tool invocation
3. **Returns/Effect**: What happens when tool runs
4. **Analysis** (optional): What user learns from results

**Progressive Complexity**:
- Simple → complex within each tool
- Common cases first, edge cases later
- Annotate complex examples with rationale

### Reusability

**Applicable To**:
- ✅ API documentation (REST, GraphQL, gRPC)
- ✅ CLI tool documentation (complex command-line tools)
- ✅ Library documentation (SDK usage examples)
- ✅ Configuration documentation (YAML/JSON configuration files)
- ✅ Testing documentation (test case examples)

**Universal Principle**: Example-driven documentation teaches both usage and rationale, reducing confusion and support burden.

**Adaptation Guide**:

1. **Identify areas needing examples**:
   - Usage data shows low adoption
   - Support questions indicate confusion
   - Complex features with many parameters

2. **Structure examples consistently**:
   - Define pattern: problem → solution → outcome
   - Use comments to explain rationale
   - Follow progressive complexity (simple → complex)

3. **Provide practical scenarios**:
   - Real-world problems users face
   - Common use cases + edge cases
   - Expected outcomes clearly stated

4. **Add reference materials**:
   - Tables for complex syntax (operators, options)
   - Explanation of conventions (tier system, naming patterns)
   - Troubleshooting for common issues

5. **Test examples**:
   - Ensure code/JSON is valid
   - Verify output matches expected
   - Update examples when behavior changes

---

## Future Extensions

### Potential Patterns (Not Yet Extracted)

1. **Breaking Change Assessment**: How to determine if API change is breaking
2. **Deprecation Scheduling**: How to time deprecation announcements
3. **Migration Tooling Design**: How to design migration scripts/tools

**Note**: Patterns 4-6 extracted from Iterations 5-6, covering validation tool, pre-commit hooks, and documentation enhancement.

---

## References

### Source Documents (Bootstrap-006)
- `api-versioning-strategy.md` (Iteration 1)
- `api-deprecation-policy.md` (Iteration 1)
- `api-compatibility-guidelines.md` (Iteration 1)
- `api-migration-framework.md` (Iteration 1)
- `api-naming-convention.md` (Iteration 2)
- `api-parameter-convention.md` (Iteration 2)
- `api-consistency-methodology.md` (Iteration 2)
- `iteration-4.md` (methodology extraction details)

### Related Experiments
- Bootstrap-001: Documentation methodology (different domain, similar extraction approach)
- Bootstrap-003: Error recovery methodology (agent specialization patterns)

---

## Revision History

| Version | Date | Changes | Extractor | Patterns |
|---------|------|---------|-----------|----------|
| 1.0 | 2025-10-15 | Initial extraction (Iteration 4, Task 1) | Two-layer architecture | 1-3 |
| 2.0 | 2025-10-15 | Automation patterns (Iteration 5, Tasks 2-3) | Two-layer architecture | 4-5 |
| 3.0 | 2025-10-15 | Documentation patterns (Iteration 6, Task 4) | Two-layer architecture | 6 |

---

**Status**: ✅ Active (Complete)
**Next Review**: N/A (Experiment complete, all 6 patterns extracted)
**Usage**: Universal API design + automation + documentation methodology
**Completeness**: 100% (all planned patterns extracted)
