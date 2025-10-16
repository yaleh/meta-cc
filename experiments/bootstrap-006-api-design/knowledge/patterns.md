# API Design Patterns

**Version**: 1.0
**Source**: Bootstrap-006 API Design Methodology
**Status**: Validated through meta-cc experiment

---

## Overview

These patterns represent proven solutions to API design and consistency challenges. Each pattern addresses a specific problem, provides a structured solution, and references the agent that automates its execution.

**Pattern Classification**:
- **Design Patterns**: Structure and organization (Pattern 1, 2)
- **Process Patterns**: Workflow and methodology (Pattern 3)
- **Automation Patterns**: Tools and enforcement (Pattern 4, 5)
- **Documentation Patterns**: Learning and adoption (Pattern 6)

---

## Pattern 1: Deterministic Parameter Categorization

### Context
When designing or refactoring API parameters, categorization decisions must be consistent and unambiguous across all tools.

### Problem
- Without systematic approach, parameter ordering becomes arbitrary
- Different developers categorize same parameter types differently
- Inconsistent API schemas increase cognitive load
- Difficult to maintain consistency as API evolves

### Solution
**Use 5-tier decision tree for parameter categorization**:

```
Tier 1: Required Parameters
  - Must be provided for tool to function

Tier 2: Filtering Parameters
  - Narrow search results (affect WHAT is returned)

Tier 3: Range Parameters
  - Define bounds, thresholds, windows

Tier 4: Output Control Parameters
  - Control output size or format (HOW MUCH)

Tier 5: Standard Parameters
  - Cross-cutting concerns (added automatically)
```

**Decision Process**:
```
For each parameter:
  Q1: "Required?" → YES = Tier 1
  Q2: "Filters WHAT?" → YES = Tier 2
  Q3: "Range/threshold?" → YES = Tier 3
  Q4: "Output control?" → YES = Tier 4
  Q5: "Standard param?" → YES = Tier 5
```

### Implementation
**Agent**: `agent-parameter-categorizer.md`

**Usage**:
```bash
/subagent @agents/agent-parameter-categorizer.md \
  target_api.file="cmd/mcp-server/tools.go"
```

### Consequences

**Benefits**:
- 100% determinism (no ambiguous cases)
- Consistent across all tools
- Predictable parameter ordering
- Easy to validate automatically

**Tradeoffs**:
- Requires initial convention definition
- Some parameters may not fit neatly (rare)
- New developers need to learn tier system

### Evidence (Bootstrap-006 Iteration 4)
- **Parameters categorized**: 60
- **Ambiguous cases**: 0
- **Determinism rate**: 100%
- **Tools reordered**: 5
- **Compliance**: 67.5% → 100%

### Transferability
- ✅ **Query-based APIs**: Filtering, range, output control
- ✅ **GraphQL schemas**: Arguments ordering
- ✅ **REST APIs**: Query parameters
- ✅ **CLI tools**: Command flags
- ✅ **Configuration files**: YAML/JSON structure

---

## Pattern 2: Safe API Refactoring via JSON Property

### Context
Need to improve API schema readability and consistency without breaking existing clients.

### Problem
- Fear of breaking changes prevents schema improvements
- Developers avoid refactoring parameter order
- Schema readability degrades over time
- Unclear if changes are safe or breaking

### Solution
**Leverage JSON spec guarantee: Object properties are unordered**

**Refactoring Types Enabled**:
1. Parameter reordering (declaration order is documentation only)
2. Clarity comments (add tier comments)
3. Schema grouping (visual organization)

**Verification Protocol**:
```
1. Confirm JSON Property: API uses JSON (unordered objects)
2. Identify Changes: List parameter order changes
3. Make Changes: Reorder parameters in schema
4. Run Tests: Execute full test suite (100% pass expected)
5. Verify Compilation: Build project successfully
6. Confirm Non-Breaking: Old and new orders both work
```

### Implementation
**Agent**: `agent-schema-refactorer.md`

**Usage**:
```bash
/subagent @agents/agent-schema-refactorer.md \
  target_schema.file="cmd/mcp-server/tools.go" \
  refactoring_goal.type="reorder"
```

### Consequences

**Benefits**:
- 100% backward compatibility (if JSON used)
- Improves readability (tier comments, grouping)
- Safe to refactor (no client impact)
- Easy to verify (test suite confirms)

**Tradeoffs**:
- Only applies to JSON-based APIs
- Doesn't work for positional arguments
- Requires thorough testing to confirm

### Evidence (Bootstrap-006 Iteration 4)
- **Tools refactored**: 5
- **Lines changed**: 60
- **Test pass rate**: 100%
- **Breaking changes**: 0
- **Backward compatible**: ✅ Confirmed

### Transferability
- ✅ **JSON APIs**: REST, GraphQL, RPC
- ✅ **JSON configs**: Configuration files
- ✅ **OpenAPI/Swagger**: Spec files
- ❌ **Positional args**: Function signatures (NOT safe)
- ❌ **Arrays**: Ordering matters (NOT safe)

---

## Pattern 3: Audit-First Refactoring

### Context
Need to refactor multiple targets for consistency, but unclear which targets actually need changes.

### Problem
- Without auditing, waste effort on already-compliant targets
- Miss non-compliant targets
- Lack prioritization data (which violations most common?)
- No metrics to quantify improvement

### Solution
**Systematic audit process before refactoring**:

```
1. List Targets: Enumerate all items to audit
2. Define Criteria: Specify what "compliant" means
3. Assess Each Target: Check compliance
4. Categorize:
   - Already compliant (verify only, no changes)
   - Needs change (non-compliant, prioritize)
5. Prioritize: Rank by violation severity/impact
6. Execute Changes: Refactor non-compliant targets only
7. Re-Audit: Verify 100% compliance achieved
```

**Efficiency Formula**:
```
Time Saved = (Total Targets - Non-Compliant Targets) × Time Per Target
Efficiency Gain = Time Saved / Time Without Audit
```

### Implementation
**Agent**: `agent-audit-executor.md`

**Usage**:
```bash
/subagent @agents/agent-audit-executor.md \
  audit_scope.targets='["tool1", "tool2", ...]' \
  compliance_criteria.convention="tier_based_ordering"
```

### Consequences

**Benefits**:
- 37.5% efficiency gain (avoid 3/8 unnecessary changes in test case)
- Identifies actual work needed
- Prioritizes highest-impact violations
- Quantifies improvement (before/after metrics)

**Tradeoffs**:
- Audit overhead (30-60 minutes)
- Only justified for multiple targets (n ≥ 3)
- Requires clear compliance criteria

### Evidence (Bootstrap-006 Iteration 4)
- **Tools audited**: 8
- **Already compliant**: 3 (37.5%)
- **Needs change**: 5 (62.5%)
- **Time saved**: 45 minutes (18.75%)
- **Avoidance efficiency**: 37.5%
- **Compliance**: 67.5% → 100%

### Transferability
- ✅ **Any refactoring**: Code quality, consistency, compliance
- ✅ **Code linting**: Style enforcement
- ✅ **Security audits**: Vulnerability scanning
- ✅ **Accessibility**: WCAG compliance
- ✅ **Migration projects**: Old API → New API

---

## Pattern 4: Automated Consistency Validation

### Context
Need to enforce API conventions at scale without manual checks.

### Problem
- Manual consistency checks are error-prone
- Developers forget to check conventions
- Inconsistencies accumulate over time
- Post-hoc fixes are costly

### Solution
**Build automated validation tool**:

**Architecture**:
```
Parser → Validators → Reporter
  ↓         ↓           ↓
Extract  Check      Format
tools    rules      results
```

**Validator Design**:
```go
type Validator interface {
    Name() string
    Check(tool Tool) ValidationResult
}

func (v *Validator) Check(tool Tool) ValidationResult {
    // 1. Extract relevant data
    data := extract(tool)

    // 2. Apply deterministic check
    if violates(data, convention) {
        return NewFailResult(
            tool.Name,
            "check_name",
            "Clear error message",
            suggestions: "Actionable fix"
        )
    }

    // 3. Return pass
    return NewPassResult(tool.Name, "check_name")
}
```

### Implementation
**Agent**: `agent-validation-builder.md`

**Usage**:
```bash
/subagent @agents/agent-validation-builder.md \
  validation_target.file="cmd/mcp-server/tools.go" \
  validators='["naming", "ordering", "description"]'
```

### Consequences

**Benefits**:
- Automated enforcement (no manual checks)
- 100% accuracy (0 false positives in test)
- Actionable errors (specific suggestions)
- Multiple formats (terminal + JSON for CI)

**Tradeoffs**:
- Initial implementation effort (4-6 hours)
- Requires deterministic rules (no judgment calls)
- Maintenance overhead (update validators)

### Evidence (Bootstrap-006 Iteration 5)
- **Files created**: 8
- **Lines of code**: ~600
- **Validators**: 3
- **Test coverage**: 100% (naming validator)
- **Tools validated**: 16
- **Violations detected**: 2
- **False positives**: 0

### Transferability
- ✅ **Any API**: With documented conventions
- ✅ **Code linting**: Style enforcement
- ✅ **Schema validation**: OpenAPI, JSON Schema
- ✅ **Config validation**: YAML/JSON files
- ✅ **Documentation**: Consistency checks

---

## Pattern 5: Automated Quality Gates

### Context
Need to prevent violations from entering repository, not just detect post-commit.

### Problem
- Post-commit fixes are costly (already merged)
- Manual checks are skipped (developer forgets, time pressure)
- Violations accumulate (broken windows effect)
- Review process burdened (reviewers catch violations)

### Solution
**Install pre-commit hooks to enforce quality**:

**Hook Architecture**:
```
Git Commit → Pre-Commit Hook → Validation Tool → Decision
                ↓                     ↓              ↓
          Detect changes       Run checks    Allow/Block
```

**Hook Pattern**:
```bash
#!/bin/bash

# 1. Detect relevant changes
if git diff --cached --name-only | grep -q "tools.go"; then
    # 2. Run validation
    if ./validate-api --fast tools.go; then
        # 3. Allow commit
        exit 0
    else
        # 4. Block commit
        echo "Fix errors before committing"
        exit 1
    fi
else
    # 5. Skip validation
    exit 0
fi
```

### Implementation
**Agent**: `agent-quality-gate-installer.md`

**Usage**:
```bash
/subagent @agents/agent-quality-gate-installer.md \
  quality_gate.type="pre-commit" \
  quality_gate.validation_command="./validate-api --fast"
```

### Consequences

**Benefits**:
- 100% violation prevention (blocks before merge)
- Automatic enforcement (no manual checks)
- Fast feedback (<5 seconds)
- Bypass option (emergencies: `--no-verify`)

**Tradeoffs**:
- Setup overhead (30-60 minutes)
- Adds commit time (<5 seconds)
- Requires discipline (don't bypass frequently)

### Evidence (Bootstrap-006 Iteration 5)
- **Hook script**: 60 lines
- **Installation script**: 70 lines
- **Test scenarios**: 4 (all passing)
- **Average runtime**: ~2 seconds
- **Violations prevented**: 100%
- **Bypass rate**: 0%

### Transferability
- ✅ **Any quality check**: Linting, testing, formatting
- ✅ **Build verification**: Compile before commit
- ✅ **Test execution**: Run tests before commit
- ✅ **Documentation**: Generate docs before commit

---

## Pattern 6: Example-Driven Documentation

### Context
Need to teach API conventions and usage effectively so users understand both HOW to use tools and WHY conventions exist.

### Problem
- Abstract guidelines difficult to apply
- Users confused by inconsistent examples
- Low-usage tools lack sufficient documentation
- Learning curve steep without practical examples
- Convention rationale unclear

### Solution
**Provide practical, example-driven documentation**:

**Documentation Structure**:
```markdown
1. **Explain Conventions First**:
   - Define system (tier system, naming patterns)
   - Provide rationale (consistency, readability)
   - Clarify misconceptions (JSON ordering)

2. **Practical Use Cases**:
   - Problem statement (what user wants to do)
   - JSON example (actual tool invocation)
   - Returns/Effect (what happens)
   - Analysis (what user learns)

3. **Progressive Complexity**:
   - Basic examples first (minimal parameters)
   - Advanced examples later (multiple conditions)
   - Annotate with rationale

4. **Troubleshooting**:
   - 6+ common issues
   - Symptom → Cause → Fix pattern
   - Actionable solutions
```

**Example Template**:
```markdown
### tool_name

**Practical Use Cases**:

1. **Scenario Name**:
   ```json
   // Problem: [User problem description]
   {"param": "value"}
   // Returns: [What user gets]
   // Analysis: [What user learns]
   ```
```

### Implementation
**Agent**: `agent-documentation-enhancer.md`

**Usage**:
```bash
/subagent @agents/agent-documentation-enhancer.md \
  documentation_target.tools='["query_context", "cleanup_temp_files"]' \
  enhancement_strategy.add_practical_cases=true
```

### Consequences

**Benefits**:
- Improved adoption (users understand when to use)
- Reduced support burden (examples answer questions)
- Better learning (WHY + HOW, not just WHAT)
- Higher accuracy (tested examples)

**Tradeoffs**:
- Documentation effort (2-3 hours per tool)
- Maintenance overhead (update when API changes)
- Examples need testing (ensure they work)

### Evidence (Bootstrap-006 Iteration 6)
- **Tools enhanced**: 3
- **Practical examples**: 11
- **Basic examples**: 8
- **Advanced examples**: 6
- **Total examples**: 25
- **Troubleshooting items**: 6
- **Examples tested**: 25
- **Examples passing**: 25
- **Accuracy**: 100%

### Transferability
- ✅ **Any API**: REST, GraphQL, gRPC
- ✅ **CLI tools**: Complex command-line tools
- ✅ **Libraries**: SDK usage examples
- ✅ **Configurations**: YAML/JSON files
- ✅ **Testing**: Test case examples

---

## Pattern Relationships

### Composition
Patterns work together as a **unified API design workflow**:

```
Phase 1: Consistency Foundation
  ↓
Pattern 3 (Audit)
  - Identify non-compliant tools
  ↓
Pattern 1 (Categorize)
  - Apply tier-based categorization
  ↓
Pattern 2 (Refactor)
  - Reorder parameters safely
  ↓
Phase 2: Automation (Scale)
  ↓
Pattern 4 (Build Validator)
  - Automate consistency checking
  ↓
Pattern 5 (Install Quality Gate)
  - Prevent future violations
  ↓
Phase 3: Documentation (Adoption)
  ↓
Pattern 6 (Enhance Docs)
  - Improve low-usage tool adoption
  ↓
Convergence: 100% Compliance + Full Automation + Enhanced Docs
```

### Dependencies
- **Pattern 1 → Pattern 2**: Categorization informs refactoring
- **Pattern 3 → All**: Audit identifies which patterns to apply
- **Pattern 4 → Pattern 5**: Validator enables quality gate
- **Pattern 2 → Pattern 4**: Conventions need enforcement
- **Pattern 6**: Independent, applied after consistency achieved

### Decision Logic
```python
def select_patterns(state):
    patterns = []

    # Foundation: Always audit first
    if state.needs_refactoring:
        patterns.append(Pattern_3)  # Audit

    # Consistency: If non-compliant tools found
    if state.compliance < 1.0:
        patterns.append(Pattern_1)  # Categorize
        patterns.append(Pattern_2)  # Refactor

    # Automation: If no validation exists
    if not state.has_validation_tool:
        patterns.append(Pattern_4)  # Build validator

    if not state.has_quality_gate:
        patterns.append(Pattern_5)  # Install hook

    # Documentation: If low adoption exists
    if state.has_low_usage_tools:
        patterns.append(Pattern_6)  # Enhance docs

    return patterns
```

---

## Pattern Catalog Summary

| Pattern | Problem Solved | Primary Benefit | Agent | Evidence |
|---------|----------------|-----------------|-------|----------|
| **1: Deterministic Categorization** | Inconsistent parameter ordering | 100% determinism | agent-parameter-categorizer | 0 ambiguous cases |
| **2: Safe Refactoring** | Fear of breaking changes | 0 breaking changes | agent-schema-refactorer | 100% backward compat |
| **3: Audit-First** | Wasted effort on compliant targets | 37.5% efficiency gain | agent-audit-executor | 3/8 targets skipped |
| **4: Automated Validation** | Manual checks error-prone | 0 false positives | agent-validation-builder | 100% accuracy |
| **5: Quality Gates** | Post-commit violations | 100% prevention | agent-quality-gate-installer | Blocks before merge |
| **6: Example-Driven Docs** | Abstract guidelines hard to apply | Improved adoption | agent-documentation-enhancer | 25 tested examples |

---

## Usage Guidelines

### When to Apply Each Pattern

**Pattern 1 (Deterministic Categorization)**:
- ✅ Designing new API parameters
- ✅ Refactoring existing parameter ordering
- ✅ Need consistency across multiple tools
- ✅ Building schema validation

**Pattern 2 (Safe Refactoring)**:
- ✅ Improving schema readability
- ✅ Reordering parameters for consistency
- ✅ Adding tier comments
- ✅ Grouping related parameters
- ❌ NOT for positional arguments

**Pattern 3 (Audit-First)**:
- ✅ Multiple targets to refactor (n ≥ 3)
- ✅ Unclear which targets need changes
- ✅ Need prioritization data
- ✅ Want to quantify improvement
- ❌ NOT for single target (overhead not justified)

**Pattern 4 (Automated Validation)**:
- ✅ Have documented conventions
- ✅ Manual checks are error-prone
- ✅ Inconsistencies accumulate
- ✅ Want to scale enforcement

**Pattern 5 (Quality Gates)**:
- ✅ Need to prevent violations
- ✅ Manual checks skipped
- ✅ Review burden high
- ✅ Have validation tool (Pattern 4 prerequisite)

**Pattern 6 (Example-Driven Docs)**:
- ✅ Low-adoption tools
- ✅ Complex parameters
- ✅ User confusion
- ✅ Need to explain conventions

### Anti-Patterns to Avoid

**❌ Skipping Pattern 3 (Audit)**:
- Don't: Refactor all targets blindly
- Result: Wasted effort (37.5% in test case)

**❌ Forcing Pattern 1 on Non-Conforming APIs**:
- Don't: Apply tier system to all parameter types
- Result: Over-abstraction, complexity

**❌ Applying Pattern 2 to Arrays**:
- Don't: Reorder array elements (arrays ARE ordered)
- Result: Breaking changes

**❌ Pattern 4 Without Deterministic Rules**:
- Don't: Build validators with judgment calls
- Result: False positives, inconsistent enforcement

**❌ Pattern 5 Without Fast Validation**:
- Don't: Run slow checks in pre-commit hook
- Result: Slow commits (>5 seconds), developers bypass

**❌ Pattern 6 With Untested Examples**:
- Don't: Write examples without testing
- Result: Broken examples, eroded trust

---

## Reusability Matrix

### Language Compatibility

| Pattern | Go | Python | JavaScript | Java | Rust | Notes |
|---------|-----|--------|------------|------|------|-------|
| Pattern 1 | ✅ | ✅ | ✅ | ✅ | ✅ | Universal (tier system) |
| Pattern 2 | ✅ | ✅ | ✅ | ✅ | ✅ | JSON-specific |
| Pattern 3 | ✅ | ✅ | ✅ | ✅ | ✅ | Universal (audit methodology) |
| Pattern 4 | ✅ | ✅ | ✅ | ✅ | ✅ | Adapt to language tools |
| Pattern 5 | ✅ | ✅ | ✅ | ✅ | ✅ | Git hooks (language-agnostic) |
| Pattern 6 | ✅ | ✅ | ✅ | ✅ | ✅ | Universal (documentation) |

### Domain Compatibility

| Pattern | REST APIs | GraphQL | CLI Tools | Config Files | Libraries |
|---------|-----------|---------|-----------|--------------|-----------|
| Pattern 1 | ✅ | ✅ | ✅ | ✅ | ✅ |
| Pattern 2 | ✅ | ✅ | ❌ | ✅ | ❌ |
| Pattern 3 | ✅ | ✅ | ✅ | ✅ | ✅ |
| Pattern 4 | ✅ | ✅ | ✅ | ✅ | ✅ |
| Pattern 5 | ✅ | ✅ | ✅ | ✅ | ✅ |
| Pattern 6 | ✅ | ✅ | ✅ | ✅ | ✅ |

### Tool Adaptations

| Pattern | Tool Category | Examples |
|---------|---------------|----------|
| Pattern 1 | Categorization Systems | Tier-based, Alphabetical, Priority-based |
| Pattern 2 | Schema Formats | JSON, YAML, TOML (unordered mappings) |
| Pattern 3 | Audit Tools | Custom scripts, linters, static analyzers |
| Pattern 4 | Validation Frameworks | Go validators, Python linters, TypeScript checkers |
| Pattern 5 | Git Hooks | pre-commit, pre-push, commit-msg |
| Pattern 6 | Documentation Formats | Markdown, RST, AsciiDoc, HTML |

---

## Cross-References

### To Agents
- Pattern 1 → `agents/agent-parameter-categorizer.md`
- Pattern 2 → `agents/agent-schema-refactorer.md`
- Pattern 3 → `agents/agent-audit-executor.md`
- Pattern 4 → `agents/agent-validation-builder.md`
- Pattern 5 → `agents/agent-quality-gate-installer.md`
- Pattern 6 → `agents/agent-documentation-enhancer.md`

### To Meta-Agent
- All Patterns → `meta-agents/api-design-orchestrator.md` (coordination)

### To Principles
- Pattern 1 → Principle of Determinism
- Pattern 2 → Principle of Safe Transformation
- Pattern 3 → Principle of Evidence-Based Action
- Pattern 4 → Principle of Automated Enforcement
- Pattern 5 → Principle of Prevention Over Detection
- Pattern 6 → Principle of Learning Through Examples

---

**Last Updated**: 2025-10-16
**Status**: Validated (Bootstrap-006 Complete)
**Evidence**: 100% success rate across iterations 4-6
**Reusability**: Universal (language/domain/tool agnostic)
